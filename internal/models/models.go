// Package models describes models for database communication.
package models

import (
	"errors"
	"fmt"
	"time"
)

// User stores information about user.
type User struct {
	ID       int64  `bson:"id"`
	UserName string `bson:"user_name"`
	FullName string `bson:"full_name"`
}

// UsersBatch represents info about specific type of users (followers, followings).
type UsersBatch struct {
	Users     []User         `bson:"users"`
	Type      UsersBatchType `bson:"batch_type"`
	CreatedAt time.Time      `bson:"created_at"`
}

// EmptyUsersBatch represents nil batch.
var EmptyUsersBatch = UsersBatch{
	Users:     nil,
	Type:      UsersBatchTypeUnknown,
	CreatedAt: time.Time{},
}

//go:generate stringer -type=UsersBatchType -trimprefix=UsersBatchType

// UsersBatchType marks what is the type of users stored in the batch.
type UsersBatchType int

const (
	// UsersBatchTypeUnknown is unknown type, to cover default value case.
	UsersBatchTypeUnknown UsersBatchType = iota

	// UsersBatchTypeFollowers represents user's followers.
	UsersBatchTypeFollowers
	// UsersBatchTypeFollowings represents user's followings.
	UsersBatchTypeFollowings
	// UsersBatchTypeNotMutual represents users that not following back.
	UsersBatchTypeNotMutual
	// UsersBatchTypeBusinessAccounts represents business accounts.
	UsersBatchTypeBusinessAccounts
	// UsersBatchTypeLostFollowers represents lost followers.
	UsersBatchTypeLostFollowers
	// UsersBatchTypeNewFollowers represents new followers.
	UsersBatchTypeNewFollowers

	usersBatchTypeSentinel // should be always last. New types should be added at the end before sentinel.
)

// ErrInvalidUsersBatchType means that batch type not supported.
var ErrInvalidUsersBatchType = errors.New("invalid users batch type")

// MakeInvalidBatchTypeError returns ErrInvalidUsersBatchType with added bathtype info.
func MakeInvalidBatchTypeError(t UsersBatchType) error {
	return fmt.Errorf("%s: %w", t.String(), ErrInvalidUsersBatchType)
}

// Valid checks if value is valid type.
func (i UsersBatchType) Valid() bool {
	return i > UsersBatchTypeUnknown && i < usersBatchTypeSentinel
}

// MakeUser creates User with passed values.
func MakeUser(id int64, username, fullname string) User {
	return User{ID: id, UserName: username, FullName: fullname}
}

// Limits represents action limits.
type Limits struct {
	Follow   int
	UnFollow int
}
