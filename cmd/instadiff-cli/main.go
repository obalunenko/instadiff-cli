// instadiff-cli is a command line tool for managing instagram account followers and followings.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/urfave/cli/v2"

	log "github.com/obalunenko/logger"

	"github.com/obalunenko/instadiff-cli/internal/config"
	"github.com/obalunenko/instadiff-cli/internal/service"
)

const (
	list     = "list"
	debug    = "debug"
	logLevel = "log_level"
	cfgPath  = "config_path"
)

func main() {
	ctx := context.Background()

	app := cli.NewApp()
	app.Name = "instadiff-cli"
	app.Usage = `a command line tool for managing instagram account followers and followings`
	app.Authors = []*cli.Author{{
		Name:  "Oleg Balunenko",
		Email: "oleg.balunenko@gmail.com",
	}}
	app.Version = printVersion(ctx)
	app.Flags = globalFlags()
	app.Commands = commands(ctx)
	app.CommandNotFound = notFound(ctx)
	app.After = onExit(ctx)
	app.Before = printHeader(ctx)

	if err := app.Run(os.Args); err != nil {
		log.WithError(ctx, err).Fatal("Failed to run")
	}
}

func serviceSetUp(c *cli.Context) (*service.Service, service.StopFunc, error) {
	setLogger(c)

	configPath := c.String(cfgPath)

	cfgDir := filepath.Dir(configPath)

	cfg, err := config.Load(c.Context, configPath)
	if err != nil {
		return nil, nil, fmt.Errorf("load config: %w", err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	cancelCtx, cancelFunc := context.WithCancel(context.Background())

	go func() {
		sig := <-sigChan
		// empty line for clear output.
		log.WithField(c.Context, "signal", sig.String()).Info("Signal received")
		cancelFunc()
	}()

	cfg.SetDebug(c.Bool(debug))

	return service.New(cancelCtx, cfg, cfgDir)
}

func setLogger(c *cli.Context) {
	log.Init(c.Context, log.Params{
		Level:        c.String(logLevel),
		Format:       "text",
		SentryParams: log.SentryParams{},
	})
}
