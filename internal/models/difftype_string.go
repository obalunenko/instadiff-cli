// Code generated by "stringer -type=DiffType -trimprefix=DiffType"; DO NOT EDIT.

package models

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[DiffTypeUnknown-0]
	_ = x[DiffTypeFollowers-1]
	_ = x[DiffTypeFollowings-2]
	_ = x[diffTypeSentinel-3]
}

const _DiffType_name = "UnknownFollowersFollowingsdiffTypeSentinel"

var _DiffType_index = [...]uint8{0, 7, 16, 26, 42}

func (i DiffType) String() string {
	if i >= DiffType(len(_DiffType_index)-1) {
		return "DiffType(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _DiffType_name[_DiffType_index[i]:_DiffType_index[i+1]]
}
