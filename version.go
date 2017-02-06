package main

import "fmt"

type version struct {
	Major, Minor, Patch int
	Label               string
	Nick                string
}

var ver = version{
	Major: 0,
	Minor: 8,
	Patch: 0,
	Label: "beta",
	Nick:  "Austin Powers"}

// CommitHash may be set on the build command line:
// go build -ldflags "-X main.CommitHash=`git rev-parse --short HEAD`"
var CommitHash string

const appName string = "dcrspy"

func (v *version) String() string {
	var hashStr string
	if CommitHash != "" {
		hashStr = "+" + CommitHash
	}
	if v.Label != "" {
		return fmt.Sprintf("%d.%d.%d-%s%s \"%s\"",
			v.Major, v.Minor, v.Patch, v.Label, hashStr, v.Nick)
	}
	return fmt.Sprintf("%d.%d.%d%s \"%s\"",
		v.Major, v.Minor, v.Patch, hashStr, v.Nick)
}
