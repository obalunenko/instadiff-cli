// Code generated by "stringer -type=UserAction -trimprefix=userAction -linecomment"; DO NOT EDIT.

package actions

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[userActionUnknown-0]
	_ = x[UserActionFollow-1]
	_ = x[UserActionUnfollow-2]
	_ = x[UserActionBlock-3]
	_ = x[UserActionUnblock-4]
	_ = x[UserActionRemove-5]
	_ = x[userActionSentinel-6]
}

const _UserAction_name = "UnknownFollowUnfollowBlockUnblockRemoveSentinel"

var _UserAction_index = [...]uint8{0, 7, 13, 21, 26, 33, 39, 47}

func (i UserAction) String() string {
	if i >= UserAction(len(_UserAction_index)-1) {
		return "UserAction(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _UserAction_name[_UserAction_index[i]:_UserAction_index[i+1]]
}
