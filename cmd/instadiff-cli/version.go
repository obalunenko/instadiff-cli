// instadiff-cli is a command line tool for managing instagram account followers and followings.
package main

import (
	"fmt"
)

const unset = "unset"

var ( // build info
	version   = unset
	date      = unset
	commit    = unset
	goversion = unset
)

// versionInfo returns stringed version info.
func versionInfo() string {
	return fmt.Sprintf("GO-%s: %s-%s-%s \n", goversion, version, commit, date)
}
