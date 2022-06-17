package main

import (
	"github.com/urfave/cli/v2"
)

func globalFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:     cfgPath,
			Usage:    "Path to the config file",
			Required: true,
			Value:    ".config.json",
		},
		&cli.StringFlag{
			Name:  logLevel,
			Usage: "Level of output logs",
			Value: "INFO",
		},
		&cli.BoolFlag{
			Name:  debug,
			Usage: "Debug mode, where actions has no real effect",
		},
		&cli.BoolFlag{
			Name:  incognito,
			Usage: "Incognito removes session on application exit.",
		},
	}
}

func addListFlag() *cli.BoolFlag {
	return &cli.BoolFlag{
		Name:  list,
		Usage: "Print the full list instead of only number",
	}
}

func addUsersFlag() *cli.StringSliceFlag {
	return &cli.StringSliceFlag{
		Name:     users,
		Usage:    "List of usernames for action",
		Required: true,
		Value:    &cli.StringSlice{},
	}
}
