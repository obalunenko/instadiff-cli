package main

import (
	"context"

	"github.com/urfave/cli/v2"
)

func commands(ctx context.Context) []*cli.Command {
	return []*cli.Command{
		{
			Name:   "followers",
			Usage:  "List your followers",
			Action: executeCmd(ctx, cmdListFollowers),
			Flags:  []cli.Flag{addListFlag()},
		},
		{
			Name:   "followings",
			Usage:  "List your followings",
			Action: executeCmd(ctx, cmdListFollowings),
			Flags:  []cli.Flag{addListFlag()},
		},
		{
			Name:    "clean-followers",
			Aliases: []string{"clean"},
			Usage:   "Un follow not mutual followings, except of whitelisted",
			Action:  executeCmd(ctx, cmdCleanFollowings),
		},
		{
			Name:    "remove-followers",
			Aliases: []string{"rm", "remove"},
			Usage:   "Remove a list of followers, by username.",
			Action:  executeCmd(ctx, cmdRemoveFollowers),
			Flags: []cli.Flag{
				&cli.StringSliceFlag{
					Name:        "follower",
					Category:    "",
					DefaultText: "",
					FilePath:    "",
					Usage:       "Follower to remove",
					Required:    true,
					Hidden:      false,
					HasBeenSet:  false,
					Value:       &cli.StringSlice{},
					Destination: nil,
					Aliases:     nil,
					EnvVars:     nil,
					TakesFile:   false,
				},
			},
		},
		{
			Name:   "unmutual",
			Usage:  "List all not mutual followings",
			Action: executeCmd(ctx, cmdListNotMutual),
			Flags:  []cli.Flag{addListFlag()},
		},
		{
			Name:   "bots",
			Usage:  "List all bots or business accounts (alpha)",
			Action: executeCmd(ctx, cmdListBotsAndBusiness),
			Flags:  []cli.Flag{addListFlag()},
		},
		{
			Name:   "diff",
			Usage:  "List diff followers (lost and new)",
			Action: executeCmd(ctx, cmdListDiff),
			Flags:  []cli.Flag{addListFlag()},
		},
	}
}
