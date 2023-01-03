package main

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	log "github.com/obalunenko/logger"
	"github.com/obalunenko/version"
)

func printVersion(ctx context.Context) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.TabIndent)

	_, err := fmt.Fprintf(w, `
| app_name:	%s	|
| version:	%s	|
| short_commit:	%s	|
| commit:	%s	|
| build_date:	%s	|
| goversion:	%s	|
        \   ^__^
         \  (oo)\_______
            (__)\       )\/\
                ||----w |
                ||     ||

`,
		version.GetAppName(),
		version.GetVersion(),
		version.GetShortCommit(),
		version.GetCommit(),
		version.GetBuildDate(),
		version.GetGoVersion())
	if err != nil {
		log.WithError(ctx, err).Error("print version")
	}
}
