package main

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"

	"github.com/oleg-balunenko/instadiff-cli/internal/config"
	"github.com/oleg-balunenko/instadiff-cli/internal/service"
)

func main() {
	app := cli.NewApp()
	app.Name = "instadiff-cli"
	app.Usage = `a command line tool for managing instagram account followers and followings`
	app.Author = "Oleg Balunenko"
	app.Version = printVersion()
	app.Flags = []cli.Flag{
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

	app.Commands = []cli.Command{
		{
			Name:   "followers",
			Usage:  "List your followers",
			Action: listFollowers,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "list",
					Usage: "Print the full list instead of only number",
				},
			},
		},
		{
			Name:   "followings",
			Usage:  "List your followings",
			Action: listFollowings,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "list",
					Usage: "Print the full list instead of only number",
				},
			},
		},
		{
			Name:    "clean-followers",
			Aliases: []string{"clean"},
			Usage:   "Un follow not mutual followings, except of whitelisted",
			Action:  cleanFollowings,
		},
		{
			Name:    "unmutual",
			Aliases: []string{"unmutual"},
			Usage:   "List all not mutual followings",
			Action:  listNotMutual,
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

type stopFunc func()

func serviceSetUp(ctx *cli.Context) (*service.Service, stopFunc, error) {
	var err error
	configPath := ctx.GlobalString("config_path")
	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to load config")
	}
	setLogger(ctx)

	cfg.SetDebug(ctx.GlobalBool("debug"))

	svc, stop, err := service.New(cfg)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to create service")
	}
	return svc, stop, nil
}

func listFollowers(ctx *cli.Context) error {
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
	list := ctx.Bool("list")
	if list {
		for _, fu := range followers {
			fmt.Printf("%s - %d \n", fu.UserName, fu.ID)
		}
	}
	return nil
}

func listFollowings(ctx *cli.Context) error {
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
	list := ctx.Bool("list")
	if list {
		for _, fu := range followings {
			fmt.Printf("%s - %d \n", fu.UserName, fu.ID)
		}
	}

	return nil
}

func cleanFollowings(ctx *cli.Context) error {
	svc, stop, err := serviceSetUp(ctx)
	if err != nil {
		return err
	}
	defer stop()
	log.Info("Cleaning from not mutual followings...")
	count, err := svc.UnFollowAllNotMutualExceptWhitelisted()
	if err != nil {
		if errors.Cause(err) == service.ErrLimitExceed {
			log.Infof("Total unfollowed before limit exceeded: %d \n", count)
		} else {
			return err
		}
	} else {
		log.Infof("Total unfollowed: %d \n", count)
	}
	return nil
}

func listNotMutual(ctx *cli.Context) error {
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
	for _, nf := range notMutualFollowers {
		fmt.Printf("%s - %s \n", nf.UserName, nf.FullName)
	}
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
