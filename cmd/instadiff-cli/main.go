// instadiff-cli is a command line tool for managing instagram account followers and followings.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/urfave/cli/v2"

	log "github.com/obalunenko/logger"

	"github.com/obalunenko/instadiff-cli/internal/config"
	"github.com/obalunenko/instadiff-cli/internal/service"
)

const (
	list      = "list"
	logLevel  = "log_level"
	cfgPath   = "config_path"
	incognito = "incognito"
	users     = "users"
	username  = "username"
	filePath  = "file_path"
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

func serviceSetUp(c *cli.Context) (*service.Service, error) {
	setLogger(c)

	configPath := c.String(cfgPath)

	cfgDir := filepath.Dir(configPath)

	cfg, err := config.Load(c.Context, configPath)
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	cancelCtx, cancelFunc := context.WithCancel(c.Context)

	c.Context = cancelCtx

	go func() {
		sig := <-sigChan
		// empty line for clear output.
		log.WithField(c.Context, "signal", sig.String()).Info("Signal received")
		cancelFunc()

		time.Sleep(1 * time.Second)

		log.Info(cancelCtx, "Exit")

		os.Exit(1)
	}()

	return service.New(cancelCtx, cfg, service.Params{
		SessionPath: cfgDir,
		IsIncognito: c.Bool(incognito),
		Username:    c.String(username),
	})
}

func setLogger(c *cli.Context) {
	log.Init(c.Context, log.Params{
		Level:        c.String(logLevel),
		Format:       "text",
		SentryParams: log.SentryParams{},
	})
}
