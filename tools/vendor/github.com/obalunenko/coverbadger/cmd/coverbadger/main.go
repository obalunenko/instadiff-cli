package main

import (
	"context"
	"flag"
	"strings"

	log "github.com/sirupsen/logrus"

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
	if err := coverbadge.Badger(config); err != nil {
		log.WithError(err).Fatal("badger")
	}
}
