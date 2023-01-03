// coverbadger is a tool to generate coverage badge images for Markdown files using Go.
// Either enter a Markdown file that does not already exist, or a Markdown file (like your README.md)
// that you want to update with coverage badge info.
// After executing of `coverbadger` the following badge will be added
//
// !`[coverbadger-tag-do-not-edit](<badge_url>)`
//
// This tag will be replaced by the image for your coverage badge.
//
// To update a .md file badge (note: comma-separated):
// Manually set the coverage value (note: do not include %):
//
// `coverbadger -md="README.md,coverage.md" -coverage=95`
//
// [Example of usage](https://github.com/obalunenko/coverbadger/blob/master/scripts/update-readme-coverage.sh)
package main

import (
	"context"
	"flag"
	"strings"

	log "github.com/obalunenko/logger"

	"github.com/obalunenko/coverbadger/internal/coverbadge"
)

func main() {
	ctx := context.Background()

	printVersion(ctx)

	var (
		badgeStyle = flag.String(
			"style",
			"flat",
			"Badge style from list: ["+strings.Join(coverbadge.BadgeStyleNames(), ",")+"]",
		)
		updateMdFiles = flag.String(
			"md",
			"",
			"A list of markdown filenames for badge updates.",
		)
		manualCoverage = flag.Float64(
			"coverage",
			-1.0,
			"A manually inputted coverage float.",
		)
	)

	flag.Parse()

	config := coverbadge.Params{
		BadgeStyle:     *badgeStyle,
		UpdateMdFiles:  *updateMdFiles,
		ManualCoverage: *manualCoverage,
	}
	if err := coverbadge.Badger(ctx, config); err != nil {
		log.WithError(ctx, err).Fatal("badger")
	}
}
