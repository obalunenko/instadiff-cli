package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/urfave/cli/v2"

	log "github.com/obalunenko/logger"

	"github.com/obalunenko/instadiff-cli/internal/models"
	"github.com/obalunenko/instadiff-cli/internal/service"
)

func notFound(ctx context.Context) cli.CommandNotFoundFunc {
	return func(c *cli.Context, command string) {
		if _, err := fmt.Fprintf(
			c.App.Writer,
			"Command [%s] not supported.\nTry --help flag to see how to use it\n",
			command,
		); err != nil {
			log.WithError(ctx, err).Fatal("Failed to print not found message")
		}
	}
}

func onExit(ctx context.Context) cli.AfterFunc {
	return func(c *cli.Context) error {
		log.Info(ctx, "Exit...")

		return nil
	}
}

type cmdFunc func(c *cli.Context, svc *service.Service) error

func executeCmd(ctx context.Context, f cmdFunc) cli.ActionFunc {
	return func(c *cli.Context) error {
		c.Context = ctx

		svc, stop, err := serviceSetUp(c)
		if err != nil {
			return fmt.Errorf("service setup: %w", err)
		}

		defer func() {
			if err = stop(); err != nil {
				log.WithError(ctx, err).Warn("Error occurred during the service stop")
			}
		}()

		return f(c, svc)
	}
}

func cmdListFollowers(c *cli.Context, svc *service.Service) error {
	ctx := log.ContextWithLogger(c.Context, log.FromContext(c.Context).WithField("cmd", "list_followers"))

	followers, err := svc.GetFollowers(ctx)
	if err != nil {
		return fmt.Errorf("get followers: %w", err)
	}

	log.WithFields(ctx, log.Fields{
		"count": len(followers),
	}).Info("Followers")

	return printUsersList(c, followers)
}

func printUsersList(c *cli.Context, users []models.User) error {
	if len(users) == 0 {
		return nil
	}

	if !c.Bool(list) {
		return nil
	}

	const (
		padding  int  = 1
		minWidth int  = 0
		tabWidth int  = 0
		padChar  byte = ' '
	)

	w := tabwriter.NewWriter(os.Stdout, minWidth, tabWidth, padding, padChar, tabwriter.TabIndent|tabwriter.Debug)

	if _, err := fmt.Fprintln(w); err != nil {
		return fmt.Errorf("write empty line: %w", err)
	}

	if _, err := fmt.Fprintf(w, "username \t ID \t full name \n"); err != nil {
		return fmt.Errorf("write header list: %w", err)
	}

	for _, us := range users {
		if _, err := fmt.Fprintf(w, "%s \t %d \t %s \n", us.UserName, us.ID, us.FullName); err != nil {
			return fmt.Errorf("write user details line: %w", err)
		}
	}

	if _, err := fmt.Fprintln(w); err != nil {
		return fmt.Errorf("write empty line: %w", err)
	}

	if err := w.Flush(); err != nil {
		return fmt.Errorf("flush writer: %w", err)
	}

	return nil
}

func cmdListFollowings(c *cli.Context, svc *service.Service) error {
	ctx := log.ContextWithLogger(c.Context, log.FromContext(c.Context).WithField("cmd", "list_followings"))

	followings, err := svc.GetFollowings(ctx)
	if err != nil {
		return fmt.Errorf("get followings: %w", err)
	}

	log.WithFields(ctx, log.Fields{
		"count": len(followings),
	}).Info("Followings")

	return printUsersList(c, followings)
}

func cmdCleanFollowings(c *cli.Context, svc *service.Service) error {
	ctx := log.ContextWithLogger(c.Context, log.FromContext(c.Context).WithField("cmd", "clean_not_mutual"))

	log.Info(ctx, "Cleaning from not mutual followings...")

	count, err := svc.UnFollowAllNotMutualExceptWhitelisted(ctx)

	switch {
	case errors.Is(err, service.ErrNoUsers):
		log.Info(ctx, "There is no users to unfollow")

		return nil
	case errors.Is(err, service.ErrCorrupted):
		log.WithField(ctx, "count", count).Info("Unfollowed before corrupted")

		return fmt.Errorf("clean notmutual: %w", err)
	case errors.Is(err, service.ErrLimitExceed):
		log.WithField(ctx, "count", count).Info("Unfollowed before limit exceeded")

		return nil
	case errors.Is(err, nil):
		log.WithField(ctx, "count", count).Info("Total unfollowed")

		return nil

	default:
		return fmt.Errorf("clean notmutual: %w", err)
	}
}

func cmdRemoveFollowers(c *cli.Context, svc *service.Service) error {
	ctx := log.ContextWithLogger(c.Context, log.FromContext(c.Context).WithField("cmd", "remove_followers"))

	followers := c.StringSlice("follower")

	log.WithField(ctx, "count", len(followers)).Info("Removing followers...")

	count, err := svc.RemoveFollowersByUsername(ctx, followers)
	if err != nil {
		if errors.Is(err, service.ErrLimitExceed) {
			log.WithField(ctx, "count", count).Info("Unfollowed before limit exceeded")

			return nil
		}

		if errors.Is(err, service.ErrCorrupted) {
			log.WithField(ctx, "count", count).Info("Unfollowed before corrupted")

			return fmt.Errorf("remove followers: %w", err)
		}

		return fmt.Errorf("remove followers: %w", err)
	}

	log.WithField(ctx, "count", count).Info("Total unfollowed")

	return nil
}

func cmdListNotMutual(c *cli.Context, svc *service.Service) error {
	ctx := log.ContextWithLogger(c.Context, log.FromContext(c.Context).WithField("cmd", "list_not_mutual"))

	notMutualFollowers, err := svc.GetNotMutualFollowers(ctx)
	if err != nil {
		return fmt.Errorf("get not mutual: %w", err)
	}

	log.WithFields(ctx, log.Fields{
		"count": len(notMutualFollowers),
	}).Info("Not following back")

	return printUsersList(c, notMutualFollowers)
}

func cmdListDiff(c *cli.Context, svc *service.Service) error {
	ctx := log.ContextWithLogger(c.Context, log.FromContext(c.Context).WithField("cmd", "list_diff"))

	diffFlwrs, err := svc.GetDiffFollowers(ctx)
	if err != nil {
		return fmt.Errorf("fetch diff followers: %w", err)
	}

	diffFlwngs, err := svc.GetDiffFollowings(ctx)
	if err != nil {
		return fmt.Errorf("fetch diff followings: %w", err)
	}

	result := make([]models.UsersBatch, 0, len(diffFlwrs)+len(diffFlwngs))
	result = append(result, diffFlwrs...)
	result = append(result, diffFlwngs...)

	return printBatches(ctx, c, result)
}

func printBatches(ctx context.Context, c *cli.Context, batches []models.UsersBatch) error {
	for i := range batches {
		batch := batches[i]

		log.WithFields(ctx, log.Fields{
			"batch_type": batch.Type,
			"count":      len(batch.Users),
		}).Info("Users batch")

		if err := printUsersList(c, batch.Users); err != nil {
			return err
		}
	}

	return nil
}

func cmdListHistoryDiff(c *cli.Context, svc *service.Service) error {
	ctx := log.ContextWithLogger(c.Context, log.FromContext(c.Context).WithField("cmd", "list_history_diff"))

	diffFlwrs, err := svc.GetHistoryDiffFollowers(ctx)
	if err != nil {
		return fmt.Errorf("get hostory diff followers: %w", err)
	}

	if err = printDiffHistory(ctx, diffFlwrs); err != nil {
		return fmt.Errorf("print followers history: %w", err)
	}

	diffFlwngs, err := svc.GetHistoryDiffFollowings(ctx)
	if err != nil {
		return fmt.Errorf("get hostory diff followings: %w", err)
	}

	if err = printDiffHistory(ctx, diffFlwngs); err != nil {
		return fmt.Errorf("print followings history: %w", err)
	}

	return nil
}

func printDiffHistory(ctx context.Context, dh models.DiffHistory) error {
	ctx = log.ContextWithLogger(ctx, log.FromContext(ctx).WithField("diff_type", dh.DiffType))

	log.Info(ctx, "Diff history")

	if len(dh.History) == 0 {
		log.Info(ctx, "No data")

		return nil
	}

	const (
		padding  int  = 1
		minWidth int  = 0
		tabWidth int  = 0
		padChar  byte = ' '
	)

	w := tabwriter.NewWriter(os.Stdout, minWidth, tabWidth, padding, padChar, tabwriter.TabIndent|tabwriter.Debug)

	if _, err := fmt.Fprintln(w); err != nil {
		return fmt.Errorf("write empty line: %w", err)
	}

	if _, err := fmt.Fprintf(w, "date \t lost \t new \n"); err != nil {
		return fmt.Errorf("write header list: %w", err)
	}

	for date, records := range dh.History {
		if len(records) > 2 {
			return errors.New("wrong diff history data")
		}

		var l, n models.UsersBatch

		for i := range records {
			switch records[i].Type {
			case models.UsersBatchTypeLostFollowers, models.UsersBatchTypeLostFollowings:
				l = records[i]
			case models.UsersBatchTypeNewFollowers, models.UsersBatchTypeNewFollowings:
				n = records[i]
			}
		}

		if _, err := fmt.Fprintf(w, "%s \t %d \t %d \n", date.String(), len(l.Users), len(n.Users)); err != nil {
			return fmt.Errorf("write user details line: %w", err)
		}
	}

	if _, err := fmt.Fprintln(w); err != nil {
		return fmt.Errorf("write empty line: %w", err)
	}

	if err := w.Flush(); err != nil {
		return fmt.Errorf("flush writer: %w", err)
	}

	return nil
}

func cmdListBotsAndBusiness(c *cli.Context, svc *service.Service) error {
	ctx := log.ContextWithLogger(c.Context, log.FromContext(c.Context).WithField("cmd", "list_bots_and_business"))

	bots, err := svc.GetBusinessAccountsOrBotsFromFollowers(ctx)
	if err != nil {
		return fmt.Errorf("get business and bots: %w", err)
	}

	log.WithField(ctx, "count", len(bots)).Info("Could be blocked")

	return printUsersList(c, bots)
}
