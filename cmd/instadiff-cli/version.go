package main

import (
	"context"
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/urfave/cli/v2"

	log "github.com/obalunenko/logger"
	"github.com/obalunenko/version"
)

func printVersion(ctx context.Context) string {
	var buf strings.Builder

	w := tabwriter.NewWriter(&buf, 0, 0, 1, ' ', tabwriter.TabIndent)

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
		log.WithError(ctx, err).Error("print version")
	}

	if err := w.Flush(); err != nil {
		log.WithError(ctx, err).Fatal("flush")
	}

	return buf.String()
}

func printHeader(_ context.Context) cli.BeforeFunc {
	const (
		padding  int  = 1
		minWidth int  = 0
		tabWidth int  = 0
		padChar  byte = ' '
	)

	return func(c *cli.Context) error {
		w := tabwriter.NewWriter(c.App.Writer, minWidth, tabWidth, padding, padChar, tabwriter.TabIndent)

		_, err := fmt.Fprintf(w, `

██╗███╗   ██╗███████╗████████╗ █████╗ ██████╗ ██╗███████╗███████╗     ██████╗██╗     ██╗
██║████╗  ██║██╔════╝╚══██╔══╝██╔══██╗██╔══██╗██║██╔════╝██╔════╝    ██╔════╝██║     ██║
██║██╔██╗ ██║███████╗   ██║   ███████║██║  ██║██║█████╗  █████╗█████╗██║     ██║     ██║
██║██║╚██╗██║╚════██║   ██║   ██╔══██║██║  ██║██║██╔══╝  ██╔══╝╚════╝██║     ██║     ██║
██║██║ ╚████║███████║   ██║   ██║  ██║██████╔╝██║██║     ██║         ╚██████╗███████╗██║
╚═╝╚═╝  ╚═══╝╚══════╝   ╚═╝   ╚═╝  ╚═╝╚═════╝ ╚═╝╚═╝     ╚═╝          ╚═════╝╚══════╝╚═╝
                                                                                        

`)
		if err != nil {
			return fmt.Errorf("print header: %w", err)
		}

		if err = w.Flush(); err != nil {
			return fmt.Errorf("flush wirter: %w", err)
		}

		return nil
	}
}
