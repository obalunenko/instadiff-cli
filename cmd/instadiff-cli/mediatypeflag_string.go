// Code generated by "stringer -type=mediaTypeFlag -trimprefix=mediaTypeFlag -linecomment"; DO NOT EDIT.

package main

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[mediaTypeUndefined-0]
	_ = x[mediaTypeStoryPhoto-1]
	_ = x[mediaTypeSentinel-2]
}

const _mediaTypeFlag_name = "undefinedstory_photosentinel"

var _mediaTypeFlag_index = [...]uint8{0, 9, 20, 28}

func (i mediaTypeFlag) String() string {
	if i >= mediaTypeFlag(len(_mediaTypeFlag_index)-1) {
		return "mediaTypeFlag(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _mediaTypeFlag_name[_mediaTypeFlag_index[i]:_mediaTypeFlag_index[i+1]]
}
