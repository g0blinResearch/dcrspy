// spy.go defines the chainMonitor and stakeMonitor, which handle the block
// connected notifications via blockConnChan.  They are separate because we
// might want to run without a wallet, just monitoring dcrd data.
//
// chappjc

package main

import (
	_ "errors"
	_ "fmt"
	"strings"
	"sync"
	"time"
	//"github.com/davecgh/go-spew/spew"
)

// for getblock, ticketfeeinfo, estimatestakediff, etc.
type chainMonitor struct {
	collector          *blockDataCollector
	dataSaver          BlockDataSaver
	blockConnectedChan chan int32
	quit               chan struct{}
	wg                 *sync.WaitGroup
}

// newChainMonitor creates a new chainMonitor
func newChainMonitor(collector *blockDataCollector,
	blockConnChan chan int32, saver BlockDataSaver,
	quit chan struct{}, wg *sync.WaitGroup) *chainMonitor {
	return &chainMonitor{
		collector:          collector,
		dataSaver:          saver,
		blockConnectedChan: blockConnChan,
		quit:               quit,
		wg:                 wg,
	}
}

// blockConnectedHandler handles block connected notifications, which trigger
// data collection and storage.
func (p *chainMonitor) blockConnectedHandler() {
	defer p.wg.Done()
out:
	for {
		select {
		case height, ok := <-p.blockConnectedChan:
			if !ok {
				log.Warnf("Block connected channel closed.")
				break out
			}
			daemonLog.Infof("Block height %v connected", height)
			//atomic.StoreInt32(&glChainHeight, height)

			// blocking call for data collection
			blockData, err := p.collector.collect()
			if err != nil {
				log.Errorf("Block data collection failed.")
				break out
			}

			if p.dataSaver != nil {
				// save data to whereever the saver wants to put it
				go p.dataSaver.Store(blockData)
				// TODO: Loop over a slice of savers (stdout, MySQL, etc.)
			}

		case _, ok := <-p.quit:
			if !ok {
				log.Infof("Got quit signal. Exiting block connected handler for BLOCK monitor.")
				break out
			}
		}
	}

}

// for getstakeinfo, etc.
type stakeMonitor struct {
	collector          *stakeInfoDataCollector
	dataSaver          StakeInfoDataSaver
	blockConnectedChan chan int32
	quit               chan struct{}
	wg                 *sync.WaitGroup
}

// newStakeMonitor creates a new stakeMonitor
func newStakeMonitor(collector *stakeInfoDataCollector,
	blockConnChan chan int32, saver StakeInfoDataSaver,
	quit chan struct{}, wg *sync.WaitGroup) *stakeMonitor {
	return &stakeMonitor{
		collector:          collector,
		dataSaver:          saver,
		blockConnectedChan: blockConnChan,
		quit:               quit,
		wg:                 wg,
	}
}

// blockConnectedHandler handles block connected notifications, which trigger
// data collection and storage.
func (p *stakeMonitor) blockConnectedHandler() {
	defer p.wg.Done()
out:
	for {
		select {
		case height, ok := <-p.blockConnectedChan:
			if !ok {
				log.Warnf("Block connected channel closed.")
				break out
			}

			// Let the wallet process the new block (too bad no wallet ntfns!)
			time.Sleep(time.Millisecond * 300)

			// Try to collect the data, retry if wallet says to
		collect:
			stakeInfo, err := p.collector.collect(uint32(height))
			if err != nil {
				log.Errorf("Stake info data collection failed: %v", err)
				if strings.Contains(err.Error(), "try again later") {
					time.Sleep(time.Millisecond * 700)
					goto collect // mmm, feel so dirty! maybe make this "cleaner" later
				}
				break out
			}

			if p.dataSaver != nil {
				// save data to whereever the saver wants to put it
				go p.dataSaver.Store(stakeInfo)
				// TODO: Loop over a slice of savers (stdout, MySQL, etc.)
			}

		case _, ok := <-p.quit:
			if !ok {
				log.Infof("Got quit signal. Exiting block connected handler for STAKE monitor.")
				break out
			}
		}
	}

}
