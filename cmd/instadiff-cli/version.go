// instadiff-cli is a command line tool for managing instagram account followers and followings.
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
