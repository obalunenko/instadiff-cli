package main

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	logging "github.com/sirupsen/logrus"

	"github.com/obalunenko/version"
)

func printVersion(_ context.Context) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.TabIndent)

	_, err := fmt.Fprintf(w, `
| app_name:	%s	|
| version:	%s	|
| short_commit:	%s	|
| build_date:	%s	|
| commit:	%s	|
        \   ^__^
         \  (oo)\_______
            (__)\       )\/\
                ||----w |
                ||     ||

`,
		version.GetAppName(),
		version.GetVersion(),
		version.GetShortCommit(),
		version.GetBuildDate(),
		version.GetCommit())
	if err != nil {
		logging.WithError(err).Error("print version")
	}
}
