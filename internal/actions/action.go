// Package actions provides list of actions that could be performed.
package actions

//go:generate stringer -type=UserAction -trimprefix=userAction -linecomment

// UserAction represents actions can be done over users.
type UserAction uint

const (
	userActionUnknown UserAction = iota

	// UserActionFollow action.
	UserActionFollow // Follow
	// UserActionUnfollow action.
	UserActionUnfollow // Unfollow
	// UserActionBlock action.
	UserActionBlock // Block
	// UserActionUnblock action.
	UserActionUnblock // Unblock
	// UserActionRemove action.
	UserActionRemove // Remove

	userActionSentinel
)
