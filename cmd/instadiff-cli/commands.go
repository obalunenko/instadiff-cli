package main

import (
	"context"

	"github.com/urfave/cli/v2"
)

func commands(ctx context.Context) []*cli.Command {
	return []*cli.Command{
		{
			Name:    "list-followers",
			Aliases: []string{"followers"},
			Usage:   "List your followers",
			Action:  executeCmd(ctx, cmdListFollowers),
			Flags:   []cli.Flag{addListFlag()},
		},
		{
			Name:    "list-followings",
			Aliases: []string{"followings"},
			Usage:   "List your followings",
			Action:  executeCmd(ctx, cmdListFollowings),
			Flags:   []cli.Flag{addListFlag()},
		},
		{
			Name:    "clean-followings",
			Aliases: []string{"clean", "unfollow-unmutual", "remove-unmutual", "rm-unmutual"},
			Usage:   "Un follow not mutual followings, except of whitelisted",
			Action:  executeCmd(ctx, cmdCleanFollowings),
		},
		{
			Name:    "remove-followers",
			Aliases: []string{"rm", "remove"},
			Usage:   "Remove a list of followers, by username.",
			Action:  executeCmd(ctx, cmdRemoveFollowers),
			Flags:   []cli.Flag{addUsersFlag()},
		},
		{
			Name:    "unfollow-users",
			Aliases: []string{"unfollow", "remove-followings"},
			Usage:   "Unfollow a list of followings, by username.",
			Action:  executeCmd(ctx, cmdUnfollowUsers),
			Flags:   []cli.Flag{addUsersFlag()},
		},
		{
			Name:    "follow-users",
			Aliases: []string{"follow", "add-followings"},
			Usage:   "Follow a list of followings, by username.",
			Action:  executeCmd(ctx, cmdFollowUsers),
			Flags:   []cli.Flag{addUsersFlag()},
		},
		{
			Name:    "list-unmutual",
			Aliases: []string{"unmutual"},
			Usage:   "List all not mutual followings",
			Action:  executeCmd(ctx, cmdListNotMutual),
			Flags:   []cli.Flag{addListFlag()},
		},
		{
			Name:    "list-useless",
			Aliases: []string{"useless, bots"},
			Usage:   "List all statistic-useless accounts (bots, business accounts or mass-followers) (alpha)",
			Action:  executeCmd(ctx, cmdListUseless),
			Flags:   []cli.Flag{addListFlag()},
		},
		{
			Name:    "list-diff",
			Aliases: []string{"diff"},
			Usage:   "List diff for account (lost and new followers and followings)",
			Action:  executeCmd(ctx, cmdListDiff),
			Flags:   []cli.Flag{addListFlag()},
		},
		{
			Name:    "diff-history",
			Aliases: []string{"history"},
			Usage:   "List diff account history (lost and new followers and followings)",
			Action:  executeCmd(ctx, cmdListHistoryDiff),
		},
		{
			Name:    "upload",
			Aliases: []string{"u"},
			Usage:   "Upload media to profile",
			Action:  executeCmd(ctx, cmdUploadMedia),
			Flags:   uploadMediaFlags(),
		},
	}
}
