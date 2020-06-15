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

	"github.com/oleg-balunenko/instadiff-cli/internal/config"
	"github.com/oleg-balunenko/instadiff-cli/internal/models"
	"github.com/oleg-balunenko/instadiff-cli/internal/service"
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
			Name:  "config_path",
			Value: ".config.json",
			Usage: "Path to the config file",
		},
		cli.StringFlag{
			Name:     "log_level",
			Usage:    "Level of output logs",
			Required: false,
			Hidden:   false,
			Value:    log.InfoLevel.String(),
		},
		cli.BoolFlag{
			Name:        "debug",
			Usage:       "Debug mode, where actions has no real effect",
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
		Name:  list,
		Usage: "Print the full list instead of only number",
	}
}
func serviceSetUp(ctx *cli.Context) (*service.Service, service.StopFunc, error) {
	var err error

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
		fmt.Println()
		log.Printf("signal [%s] received", sig.String())
		cancelFunc()
	}()

	cfg.SetDebug(ctx.GlobalBool("debug"))

	return service.New(cancelCtx, cfg, cfgDir)
}

func cmdListFollowers(ctx *cli.Context) error {
	svc, stop, err := serviceSetUp(ctx)
	if err != nil {
		return err
	}

	defer stop()

	followers, err := svc.GetFollowers()
	if err != nil {
		return err
	}

	fmt.Printf("Followers: %d \n", len(followers))

	printUsersList(ctx, followers)

	return nil
}

func printUsersList(ctx *cli.Context, users []models.User) {
	if ctx.Bool(list) {
		for _, us := range users {
			fmt.Printf("%s - %d \n", us.UserName, us.ID)
		}
	}
}

func cmdListFollowings(ctx *cli.Context) error {
	svc, stop, err := serviceSetUp(ctx)
	if err != nil {
		return err
	}

	defer stop()

	followings, err := svc.GetFollowings()
	if err != nil {
		return err
	}

	fmt.Printf("Followings: %d \n", len(followings))

	printUsersList(ctx, followings)

	return nil
}

func cmdCleanFollowings(ctx *cli.Context) error {
	svc, stop, err := serviceSetUp(ctx)
	if err != nil {
		return err
	}

	defer stop()

	log.Info("Cleaning from not mutual followings...")

	count, err := svc.UnFollowAllNotMutualExceptWhitelisted()

	switch {
	case err == nil:
		log.Infof("Total unfollowed: %d \n", count)
	case errors.Is(err, service.ErrLimitExceed):
		log.Infof("Total unfollowed before limit exceeded: %d \n", count)
	case errors.Is(err, service.ErrCorrupted):
		log.Infof("Total unfollowed before corrupted: %d \n", count)
	}

	return err
}

func cmdListNotMutual(ctx *cli.Context) error {
	svc, stop, err := serviceSetUp(ctx)
	if err != nil {
		return err
	}

	defer stop()

	notMutualFollowers, err := svc.GetNotMutualFollowers()
	if err != nil {
		return err
	}

	log.Infof("Not following back: %d \n", len(notMutualFollowers))

	printUsersList(ctx, notMutualFollowers)

	return nil
}

func cmdListDiff(ctx *cli.Context) error {
	svc, stop, err := serviceSetUp(ctx)
	if err != nil {
		return err
	}

	defer stop()

	diff, err := svc.GetDiffFollowers()
	if err != nil {
		return err
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
		return err
	}

	defer stop()

	bots, err := svc.GetBusinessAccountsOrBotsFromFollowers()
	if err != nil {
		return err
	}

	log.Infof("Could be blocked: %d \n", len(bots))

	printUsersList(ctx, bots)

	return nil
}

func setLogger(ctx *cli.Context) {
	formatter := log.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
		DisableSorting:  false,
		ForceColors:     true,
	}

	log.SetFormatter(&formatter)

	lvl, err := log.ParseLevel(ctx.GlobalString("log_level"))
	if err != nil {
		lvl = log.InfoLevel
	}

	log.SetLevel(lvl)
}
