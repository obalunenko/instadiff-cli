// Package models describes models for database communication.
package models

import (
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

// MakeUsersBatch constructs UsersBatch.
func MakeUsersBatch(bt UsersBatchType, users []User, created time.Time) UsersBatch {
	return UsersBatch{
		Users:     users,
		Type:      bt,
		CreatedAt: created,
	}
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
	// UsersBatchTypeNewFollowings represents new followings.
	UsersBatchTypeNewFollowings
	// UsersBatchTypeLostFollowings represents lost followings.
	UsersBatchTypeLostFollowings

	usersBatchTypeSentinel // should be always last. New types should be added at the end before sentinel.
)

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

//go:generate stringer -type=DiffType -trimprefix=DiffType

// DiffType marks what is the type of diff hostory is.
type DiffType uint

const (
	// DiffTypeUnknown is unknown type, to cover default value case.
	DiffTypeUnknown DiffType = iota

	// DiffTypeFollowers represents followers history.
	DiffTypeFollowers
	// DiffTypeFollowings represents followings history.
	DiffTypeFollowings

	diffTypeSentinel // should be always last. New types should be added at the end before sentinel.
)

// DiffHistory represents history of account changes.
type DiffHistory struct {
	DiffType DiffType
	History  map[time.Time][]UsersBatch
}

// Add adds user batch to the history.
func (d *DiffHistory) Add(batches ...UsersBatch) {
	if len(batches) == 0 {
		return
	}

	for i := range batches {
		batch := batches[i]
		date := batch.CreatedAt

		l, ok := d.History[date]
		if !ok {
			l = make([]UsersBatch, 0, 2)
		}

		l = append(l, batch)

		d.History[date] = l
	}
}

// Get returns all UsersBatch for specified date.
func (d *DiffHistory) Get(date time.Time) []UsersBatch {
	return d.History[date]
}

// MakeDiffHistory constructs DiffHistory.
func MakeDiffHistory(dt DiffType) DiffHistory {
	return DiffHistory{
		DiffType: dt,
		History:  make(map[time.Time][]UsersBatch),
	}
}
