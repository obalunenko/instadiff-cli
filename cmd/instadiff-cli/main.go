package main

import (
	"fmt"
	"log"
	"os"

	"github.com/pkg/errors"
	"github.com/urfave/cli"

	"github.com/oleg-balunenko/insta-follow-diff/internal/config"
	"github.com/oleg-balunenko/insta-follow-diff/internal/service"
)

func main() {

	app := cli.NewApp()
	app.Name = "instadiff-cli"
	app.Usage = `a command line tool for managing instagram account followers and followings`
	app.Author = "Oleg Balunenko"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "config_path",
			Value: ".config.json",
			Usage: "Path to the config file",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:    "list-followers",
			Aliases: []string{"followers"},
			Usage:   "list your followers",
			Action:  listFollowers,
		},
		{
			Name:    "list-followings",
			Aliases: []string{"followings"},
			Usage:   "list your followings",
			Action:  listFollowings,
		},
		{
			Name:    "clean-followers",
			Aliases: []string{"clean"},
			Usage:   "Un follow not mutual followings, except of whitelisted",
			Action:  cleanFollowings,
		},
		{
			Name:    "not-mutual-followings",
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

	return nil
}

func cleanFollowings(ctx *cli.Context) error {
	svc, stop, err := serviceSetUp(ctx)
	if err != nil {
		return err
	}
	defer stop()
	count, err := svc.UnFollowAllNotMutualExceptWhitelisted()
	if err != nil {
		if errors.Cause(err) == service.ErrLimitExceed {
			log.Printf("Total unfollowed before limit exceeded: %d \n", count)
		} else {
			return err
		}
	} else {
		log.Printf("total unfollowed: %d \n", count)
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
	fmt.Printf("Not following back: %d \n", len(notMutualFollowers))
	for _, nf := range notMutualFollowers {
		fmt.Printf("%s - %s \n", nf.UserName, nf.FullName)
	}
	return nil
}
