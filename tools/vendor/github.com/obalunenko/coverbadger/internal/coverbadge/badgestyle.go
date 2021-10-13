package coverbadge

import "fmt"

//go:generate stringer -type=badgeStyle -linecomment

// badgeStyle is an enumeration of badge style values.
type badgeStyle uint

const (
	badgeStyleUnknown badgeStyle = iota // unknown

	badgeStylePlastic     // plastic
	badgeStyleFlat        // flat
	badgeStyleFlatSquare  // flat-square
	badgeStyleForTheBadge // for-the-badge
	badgeStyleSocial      // social

	badgeStyleSentinel // sentinel
)

func badgeStyleDict() map[string]badgeStyle {
	res := make(map[string]badgeStyle)

	for i := badgeStyleUnknown; i < badgeStyleSentinel; i++ {
		res[i.String()] = i
	}

	return res
}

func parseBadgeStyle(v string) (badgeStyle, error) {
	style, ok := badgeStyleDict()[v]
	if !ok {
		return badgeStyleUnknown, fmt.Errorf("invalid badge stle value")
	}

	return style, nil
}

func (i badgeStyle) IsValid() bool {
	return i > badgeStyleUnknown && i < badgeStyleSentinel
}

// BadgeStyleNames returns list of valid badgeStyle names.
func BadgeStyleNames() []string {
	res := make([]string, 0, badgeStyleSentinel-1-1)

	for i := badgeStyleUnknown; i < badgeStyleSentinel; i++ {
		if i.IsValid() {
			res = append(res, i.String())
		}
	}

	return res
}
