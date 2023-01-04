package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"sort"
	"text/tabwriter"
	"time"

	"github.com/obalunenko/instadiff-cli/internal/utils"

	"github.com/urfave/cli/v2"

	log "github.com/obalunenko/logger"

	"github.com/obalunenko/instadiff-cli/internal/media"
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
		ctx = log.ContextWithLogger(c.Context, log.FromContext(c.Context).WithField("cmd", c.Command.Name))

		c.Context = ctx

		svc, err := serviceSetUp(c)
		if err != nil {
			return fmt.Errorf("service setup: %w", err)
		}

		defer func() {
			utils.LogError(ctx, svc.Stop(ctx), "Error occurred during the service stop")
		}()

		return f(c, svc)
	}
}

func cmdListFollowers(c *cli.Context, svc *service.Service) error {
	ctx := c.Context

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
	ctx := c.Context

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
	var f cmdWithCountFunc = func(c *cli.Context, svc *service.Service) (int, error) {
		ctx := c.Context

		log.Info(ctx, "Cleaning from not mutual followings...")

		return svc.UnFollowAllNotMutualExceptWhitelisted(ctx)
	}

	return cmdHandleCount(c, svc, f, "clean not mutual followings")
}

type cmdWithCountFunc func(c *cli.Context, svc *service.Service) (int, error)

func cmdHandleCount(c *cli.Context, svc *service.Service, f cmdWithCountFunc, operation string) error {
	ctx := c.Context

	count, err := f(c, svc)

	l := log.WithField(ctx, "operation", operation).WithField("count", count)

	switch {
	case errors.Is(err, service.ErrNoUsers):
		l.Info("There is no users to process")

		return nil
	case errors.Is(err, service.ErrCorrupted):
		l.Info("Processed before corrupted")

		return err
	case errors.Is(err, service.ErrLimitExceed):
		l.Info("Processed before limit exceeded")

		return nil
	case errors.Is(err, nil):
		l.Info("Total processed")

		return nil
	default:
		return err
	}
}

func cmdRemoveFollowers(c *cli.Context, svc *service.Service) error {
	var f cmdWithCountFunc = func(c *cli.Context, svc *service.Service) (int, error) {
		ctx := c.Context

		followers := c.StringSlice(users)

		log.WithField(ctx, "count", len(followers)).Info("Removing followers...")

		return svc.RemoveFollowersByUsername(ctx, followers)
	}

	return cmdHandleCount(c, svc, f, "remove followers")
}

func cmdUnfollowUsers(c *cli.Context, svc *service.Service) error {
	var f cmdWithCountFunc = func(c *cli.Context, svc *service.Service) (int, error) {
		ctx := c.Context

		usrs := c.StringSlice(users)

		log.WithField(ctx, "count", len(usrs)).Info("Unfollow users...")

		return svc.UnfollowUsers(ctx, usrs)
	}

	return cmdHandleCount(c, svc, f, "unfollow users")
}

func cmdFollowUsers(c *cli.Context, svc *service.Service) error {
	var f cmdWithCountFunc = func(c *cli.Context, svc *service.Service) (int, error) {
		ctx := c.Context

		usrs := c.StringSlice(users)

		log.WithField(ctx, "count", len(usrs)).Info("Following users...")

		return svc.FollowUsers(ctx, usrs)
	}

	return cmdHandleCount(c, svc, f, "follow users")
}

func cmdListNotMutual(c *cli.Context, svc *service.Service) error {
	ctx := c.Context

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
	ctx := c.Context

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
	ctx := c.Context

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
		tLayout       = "02-01-2006 15:04:05"
	)

	w := tabwriter.NewWriter(os.Stdout, minWidth, tabWidth, padding, padChar, tabwriter.TabIndent|tabwriter.Debug)

	if _, err := fmt.Fprintln(w); err != nil {
		return fmt.Errorf("write empty line: %w", err)
	}

	if _, err := fmt.Fprintf(w, "date \t lost \t new \n"); err != nil {
		return fmt.Errorf("write header list: %w", err)
	}

	const recnum = 2

	var dates = make([]time.Time, 0, len(dh.History))

	for date := range dh.History {
		d := date

		dates = append(dates, d)
	}

	sort.Slice(dates, func(i, j int) bool {
		return dates[i].After(dates[j])
	})

	for _, date := range dates {
		records := dh.History[date]

		if len(records) > recnum {
			return errors.New("wrong diff history data")
		}

		var l, n models.UsersBatch

		for i := range records {
			r := records[i]

			switch r.Type {
			case models.UsersBatchTypeLostFollowers, models.UsersBatchTypeLostFollowings:
				l = r
			case models.UsersBatchTypeNewFollowers, models.UsersBatchTypeNewFollowings:
				n = r
			default:
				return fmt.Errorf("invalid batch type[%s]", r.Type.String())
			}
		}

		if _, err := fmt.Fprintf(w, "%s \t %d \t %d \n", date.Format(tLayout), len(l.Users), len(n.Users)); err != nil {
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

func cmdListUseless(c *cli.Context, svc *service.Service) error {
	ctx := c.Context

	bots, err := svc.GetUselessFollowers(ctx)
	if err != nil {
		return fmt.Errorf("get useless followers: %w", err)
	}

	log.WithField(ctx, "count", len(bots)).Info("Could be blocked")

	return printUsersList(c, bots)
}

var errEmptyFilePath = errors.New("path is empty")

func cmdUploadMedia(c *cli.Context, svc *service.Service) error {
	ctx := c.Context

	p := c.String(filePath)

	file, err := getMediaFile(ctx, p)
	if err != nil {
		return fmt.Errorf("get media file: %w", err)
	}

	if err = svc.UploadMedia(ctx, file, getMediaType(c)); err != nil {
		return fmt.Errorf("upload media: %w", err)
	}

	return nil
}

func getMediaFile(ctx context.Context, fpath string) (io.Reader, error) {
	if fpath == "" {
		return nil, errEmptyFilePath
	}

	f, err := os.Open(path.Clean(fpath))
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}

	defer func() {
		utils.LogError(ctx, f.Close(), "Failed to close file descriptor")
	}()

	content, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(content), nil
}

//go:generate stringer -type=mediaTypeFlag -trimprefix=mediaTypeFlag -linecomment

type mediaTypeFlag uint

const (
	mediaTypeUndefined mediaTypeFlag = iota // undefined

	mediaTypeStoryPhoto // story_photo

	mediaTypeSentinel // sentinel
)

func getMediaType(c *cli.Context) media.Type {
	mt := media.TypeUndefined

	if c.Bool(mediaTypeStoryPhoto.String()) {
		mt = media.TypeStoryPhoto
	}

	return mt
}
