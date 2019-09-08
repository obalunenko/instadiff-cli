package main

import "fmt"

const unset = "unset"

var ( // build info
	version = unset
	date    = unset
	commit  = unset
)

func printVersion() string {
	info := fmt.Sprintf("Version: %s-%s-%s \n", version, commit, date)
	return info
}
