package models

// User stores information about user.
type User struct {
	ID       int64  `bson:"id"`
	UserName string `bson:"user_name"`
	FullName string `bson:"full_name"`
}

// UsersBatch represents info about specific type of users (followers, followings).
type UsersBatch struct {
	Users []User         `bson:"users"`
	Type  UsersBatchType `bson:"batch_type"`
}

// EmptyUsersBatch represents nil batch
var EmptyUsersBatch = UsersBatch{
	Users: nil,
	Type:  UsersBatchTypeUnknown,
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

	usersBatchTypeSentinel // should be always last
)

// Valid checks if value is valid type.
func (i UsersBatchType) Valid() bool {
	return i > UsersBatchTypeUnknown && i < usersBatchTypeSentinel
}

// MakeUser creates User with passed values.
func MakeUser(id int64, username string, fullname string) User {
	return User{ID: id, UserName: username, FullName: fullname}
}

// Limits represents action limits.
type Limits struct {
	Follow   int
	UnFollow int
}
