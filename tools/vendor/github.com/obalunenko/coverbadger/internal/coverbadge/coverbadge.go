package coverbadge

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	log "github.com/obalunenko/logger"
)

type badge struct {
	Style          string
	ImageExtension string
}

// ErrInvalidCoverageValue returns when invalid coverage value was set.
var ErrInvalidCoverageValue = errors.New("invalid coverage value")

func (badge badge) generateBadgeBadgeURL(cov float64) (string, error) {
	const (
		bitsize   int = 64
		badgeName     = "coverage"
	)

	if cov < 0 || cov > 100 {
		return "", ErrInvalidCoverageValue
	}

	url := fmt.Sprintf(
		"https://img.shields.io/badge/%s-%s%%25-brightgreen%s?longCache=true&style=%s",
		badgeName,
		strconv.FormatFloat(cov, 'G', -1, bitsize),
		badge.ImageExtension,
		badge.Style,
	)

	return url, nil
}

var (
	regex = regexp.MustCompile(`!\[coverbadger-tag-do-not-edit]\(.*\)`)
)

func (badge badge) writeBadgeToMd(ctx context.Context, fpath string, cov float64) error {
	badgeURL, err := badge.generateBadgeBadgeURL(cov)
	if err != nil {
		return fmt.Errorf("generate badge URL: %w", err)
	}

	newImageTag := fmt.Sprintf("![coverbadger-tag-do-not-edit](%s)", badgeURL)

	filedata, err := os.ReadFile(filepath.Clean(fpath))
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}

	var markdownData string
	if len(filedata) == 0 {
		markdownData = newImageTag
	} else {
		if !regex.MatchString(string(filedata)) {
			// try to add badge to the top of Markdown
			markdownData = newImageTag + "\n\n" + string(filedata)
		} else {
			markdownData = regex.ReplaceAllString(string(filedata), newImageTag)
		}
	}

	err = os.WriteFile(fpath, []byte(markdownData), os.ModePerm)
	if err != nil {
		return fmt.Errorf("update markdown file: %w", err)
	}

	log.Info(ctx, "Wrote badge image to markdown file")

	return nil
}
