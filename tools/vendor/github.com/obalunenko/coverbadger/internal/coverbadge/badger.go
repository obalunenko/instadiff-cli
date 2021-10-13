package coverbadge

import (
	"fmt"
	"strings"
)

// Params holds Badger parameters.
type Params struct {
	BadgeStyle     string
	UpdateMdFiles  string
	ManualCoverage float64
	ImgExt         string
}

// Badger updates cover badge according tp Params.
func Badger(p Params) error {
	style, err := parseBadgeStyle(p.BadgeStyle)
	if err != nil {
		return fmt.Errorf("invalid badge style flag: %w", err)
	}

	b := badge{
		Style:          style.String(),
		ImageExtension: p.ImgExt,
	}

	cov := p.ManualCoverage

	if p.UpdateMdFiles == "" {
		return fmt.Errorf("no md files passed for update")
	}

	files := strings.Split(p.UpdateMdFiles, ",")
	if len(files) < 1 {
		return fmt.Errorf("invalid files list, filenames should be separated by ',' or only one passed")
	}

	for _, f := range files {
		if err := b.writeBadgeToMd(f, cov); err != nil {
			return fmt.Errorf("write badge to markdown[%s]: %w", f, err)
		}
	}

	return nil
}
