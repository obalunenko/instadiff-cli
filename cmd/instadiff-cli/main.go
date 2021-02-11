// instadiff-cli is a command line tool for managing instagram account followers and followings.
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"

	"github.com/obalunenko/instadiff-cli/internal/config"
	"github.com/obalunenko/instadiff-cli/internal/models"
	"github.com/obalunenko/instadiff-cli/internal/service"
)

const list = "list"

func main() {
	app := cli.NewApp()
	app.Name = "instadiff-cli"
	app.Usage = `a command line tool for managing instagram account followers and followings`
	app.Author = "Oleg Balunenko"
	app.Version = versionInfo()
	app.Flags = globalFlags()
	app.Commands = commands()

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func globalFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:        "config_path",
			Usage:       "Path to the config file",
			EnvVar:      "",
			FilePath:    "",
			Required:    false,
			Hidden:      false,
			TakesFile:   false,
			Value:       ".config.json",
			Destination: nil,
		},
		cli.StringFlag{
			Name:        "log_level",
			Usage:       "Level of output logs",
			EnvVar:      "",
			FilePath:    "",
			Required:    false,
			Hidden:      false,
			TakesFile:   false,
			Value:       log.InfoLevel.String(),
			Destination: nil,
		},
		cli.BoolFlag{
			Name:        "debug",
			Usage:       "Debug mode, where actions has no real effect",
			EnvVar:      "",
			FilePath:    "",
			Required:    false,
			Hidden:      false,
			Destination: nil,
		},
	}
}

func commands() []cli.Command {
	return []cli.Command{
		{
			Name:   "followers",
			Usage:  "List your followers",
			Action: cmdListFollowers,
			Flags:  []cli.Flag{addListFlag()},
		},
		{
			Name:   "followings",
			Usage:  "List your followings",
			Action: cmdListFollowings,
			Flags:  []cli.Flag{addListFlag()},
		},
		{
			Name:    "clean-followers",
			Aliases: []string{"clean"},
			Usage:   "Un follow not mutual followings, except of whitelisted",
			Action:  cmdCleanFollowings,
		},
		{
			Name:   "unmutual",
			Usage:  "List all not mutual followings",
			Action: cmdListNotMutual,
			Flags:  []cli.Flag{addListFlag()},
		},
		{
			Name:   "bots",
			Usage:  "List all bots or business accounts (alpha)",
			Action: cmdListBotsAndBusiness,
			Flags:  []cli.Flag{addListFlag()},
		},
		{
			Name:   "diff",
			Usage:  "List diff followers (lost and new)",
			Action: cmdListDiff,
			Flags:  []cli.Flag{addListFlag()},
		},
	}
}

func addListFlag() cli.BoolFlag {
	return cli.BoolFlag{
		Name:        list,
		Usage:       "Print the full list instead of only number",
		EnvVar:      "",
		FilePath:    "",
		Required:    false,
		Hidden:      false,
		Destination: nil,
	}
}

func serviceSetUp(ctx *cli.Context) (*service.Service, service.StopFunc, error) {
	configPath := ctx.GlobalString("config_path")

	cfgDir := filepath.Dir(configPath)

	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, nil, fmt.Errorf("load config: %w", err)
	}

	setLogger(ctx)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	cancelCtx, cancelFunc := context.WithCancel(context.Background())

	go func() {
		sig := <-sigChan
		// empty line for clear output.
		log.Printf("\nsignal [%s] received", sig.String())
		cancelFunc()
	}()

	cfg.SetDebug(ctx.GlobalBool("debug"))

	return service.New(cancelCtx, cfg, cfgDir)
}

func cmdListFollowers(ctx *cli.Context) error {
	svc, stop, err := serviceSetUp(ctx)
	if err != nil {
		return fmt.Errorf("service setup: %w", err)
	}

	defer stop()

	followers, err := svc.GetFollowers()
	if err != nil {
		return fmt.Errorf("get followers: %w", err)
	}

	log.WithFields(log.Fields{
		"count": len(followers),
	}).Info("Followers")

	printUsersList(ctx, followers)

	return nil
}

func printUsersList(ctx *cli.Context, users []models.User) {
	if ctx.Bool(list) {
		for _, us := range users {
			_, _ = fmt.Fprintf(os.Stdout, "%s - %d \n", us.UserName, us.ID)
		}
	}
}

func cmdListFollowings(ctx *cli.Context) error {
	svc, stop, err := serviceSetUp(ctx)
	if err != nil {
		return fmt.Errorf("service setup: %w", err)
	}

	defer stop()

	followings, err := svc.GetFollowings()
	if err != nil {
		return fmt.Errorf("get followings: %w", err)
	}

	log.WithFields(log.Fields{
		"count": len(followings),
	}).Info("Followings")

	printUsersList(ctx, followings)

	return nil
}

func cmdCleanFollowings(ctx *cli.Context) error {
	svc, stop, err := serviceSetUp(ctx)
	if err != nil {
		return fmt.Errorf("service setup: %w", err)
	}

	defer stop()

	log.Info("Cleaning from not mutual followings...")

	count, err := svc.UnFollowAllNotMutualExceptWhitelisted()
	if err != nil {
		if errors.Is(err, service.ErrLimitExceed) {
			log.Infof("Total unfollowed before limit exceeded: %d \n", count)

			return nil
		}

		if errors.Is(err, service.ErrCorrupted) {
			log.Infof("Total unfollowed before corrupted: %d \n", count)

			return fmt.Errorf("clean notmutual: %w", err)
		}

		return fmt.Errorf("clean notmutual: %w", err)
	}

	log.Infof("Total unfollowed: %d \n", count)

	return nil
}

func cmdListNotMutual(ctx *cli.Context) error {
	svc, stop, err := serviceSetUp(ctx)
	if err != nil {
		return fmt.Errorf("service setup: %w", err)
	}

	defer stop()

	notMutualFollowers, err := svc.GetNotMutualFollowers()
	if err != nil {
		return fmt.Errorf("get not mutual: %w", err)
	}

	log.WithFields(log.Fields{
		"count": len(notMutualFollowers),
	}).Info("Not following back")

	printUsersList(ctx, notMutualFollowers)

	return nil
}

func cmdListDiff(ctx *cli.Context) error {
	svc, stop, err := serviceSetUp(ctx)
	if err != nil {
		return fmt.Errorf("service setup: %w", err)
	}

	defer stop()

	diff, err := svc.GetDiffFollowers()
	if err != nil {
		return fmt.Errorf("fet diff followers: %w", err)
	}

	for _, batch := range diff {
		log.Infof("%s: %d \n", batch.Type, len(batch.Users))

		printUsersList(ctx, batch.Users)
	}

	return nil
}

func cmdListBotsAndBusiness(ctx *cli.Context) error {
	svc, stop, err := serviceSetUp(ctx)
	if err != nil {
		return fmt.Errorf("service setup: %w", err)
	}

	defer stop()

	bots, err := svc.GetBusinessAccountsOrBotsFromFollowers()
	if err != nil {
		return fmt.Errorf("get business and bots: %w", err)
	}

	log.Infof("Could be blocked: %d \n", len(bots))

	printUsersList(ctx, bots)

	return nil
}

func setLogger(ctx *cli.Context) {
	formatter := log.TextFormatter{
		ForceColors:               true,
		DisableColors:             false,
		ForceQuote:                false,
		DisableQuote:              false,
		EnvironmentOverrideColors: false,
		DisableTimestamp:          false,
		FullTimestamp:             true,
		TimestampFormat:           "2006-01-02 15:04:05",
		DisableSorting:            false,
		SortingFunc:               nil,
		DisableLevelTruncation:    false,
		PadLevelText:              false,
		QuoteEmptyFields:          false,
		FieldMap:                  nil,
		CallerPrettyfier:          nil,
	}

	log.SetFormatter(&formatter)

	lvl, err := log.ParseLevel(ctx.GlobalString("log_level"))
	if err != nil {
		lvl = log.InfoLevel
	}

	log.SetLevel(lvl)
}
