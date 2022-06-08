package main

import (
	"github.com/urfave/cli/v2"
)

func globalFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:        cfgPath,
			Category:    "",
			DefaultText: "",
			FilePath:    "",
			Usage:       "Path to the config file",
			Required:    false,
			Hidden:      false,
			HasBeenSet:  false,
			Value:       ".config.json",
			Destination: nil,
			Aliases:     nil,
			EnvVars:     nil,
			TakesFile:   false,
		},
		&cli.StringFlag{
			Name:        logLevel,
			Category:    "",
			DefaultText: "",
			FilePath:    "",
			Usage:       "Level of output logs",
			Required:    false,
			Hidden:      false,
			HasBeenSet:  false,
			Value:       "INFO",
			Destination: nil,
			Aliases:     nil,
			EnvVars:     nil,
			TakesFile:   false,
		},
		&cli.BoolFlag{
			Name:        debug,
			Category:    "",
			DefaultText: "",
			FilePath:    "",
			Usage:       "Debug mode, where actions has no real effect",
			Required:    false,
			Hidden:      false,
			HasBeenSet:  false,
			Value:       false,
			Destination: nil,
			Aliases:     nil,
			EnvVars:     nil,
		},
	}
}

func addListFlag() *cli.BoolFlag {
	return &cli.BoolFlag{
		Name:        list,
		Category:    "",
		DefaultText: "",
		FilePath:    "",
		Usage:       "Print the full list instead of only number",
		Required:    false,
		Hidden:      false,
		HasBeenSet:  false,
		Value:       false,
		Destination: nil,
		Aliases:     nil,
		EnvVars:     nil,
	}
}
