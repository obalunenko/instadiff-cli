package main

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	log "github.com/sirupsen/logrus"

	"github.com/obalunenko/version"
)

func printVersion(_ context.Context) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.TabIndent)

	_, err := fmt.Fprintf(w, `
| app_name:	%s	|
| version:	%s	|
| go_version:	%s	|
| commit:	%s	|
| short_commit:	%s	|
| build_date:	%s	|
        \   ^__^
         \  (oo)\_______
            (__)\       )\/\
                ||----w |
                ||     ||
`,
		version.GetAppName(),
		version.GetVersion(),
		version.GetGoVersion(),
		version.GetCommit(),
		version.GetShortCommit(),
		version.GetBuildDate(),
	)
	if err != nil {
		log.WithError(err).Error("print version")
	}
}

func versionInfo() string {
	return fmt.Sprintf("%s-%s-%s \n", version.GetVersion(), version.GetCommit(), version.GetBuildDate())
}
