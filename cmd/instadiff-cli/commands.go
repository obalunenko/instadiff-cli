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
					Name:     "follower",
					Usage:    "Follower to remove",
					Required: true,
					Value:    &cli.StringSlice{},
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
			Usage:  "List diff for account (lost and new followers and followings)",
			Action: executeCmd(ctx, cmdListDiff),
			Flags:  []cli.Flag{addListFlag()},
		},
		{
			Name:    "diff-history",
			Aliases: []string{"history"},
			Usage:   "List diff account history (lost and new followers and followings)",
			Action:  executeCmd(ctx, cmdListHistoryDiff),
		},
	}
}
