package main

import (
	"github.com/urfave/cli/v2"
)

func globalFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:     cfgPath,
			Usage:    "Path to the config file",
			Required: false,
			Value:    ".config.json",
		},
		&cli.StringFlag{
			Name:     logLevel,
			Usage:    "Level of output logs",
			Required: false,
			Value:    "INFO",
		},
		&cli.StringFlag{
			Name:     username,
			Usage:    "Username (optional parameter to avoid manual input during cli run)",
			Required: false,
			Value:    "",
		},
		&cli.BoolFlag{
			Name:     incognito,
			Usage:    "Incognito removes session on application exit.",
			Required: false,
			Value:    false,
		},
	}
}

func addListFlag() *cli.BoolFlag {
	return &cli.BoolFlag{
		Name:     list,
		Usage:    "Print the full list instead of only number",
		Required: false,
		Value:    false,
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

func uploadMediaFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:     filePath,
			Usage:    "Path to the media file",
			Required: true,
			Value:    "",
		},
		&cli.BoolFlag{
			Name:     storyPhoto,
			Usage:    "If true - media will be uploaded as story photo",
			Required: false,
			Value:    false,
		},
	}
}
