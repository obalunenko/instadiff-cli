package models

// UserInfo stores information about user.
type UserInfo struct {
	ID       int64
	UserName string
	FullName string
}

// MakeUserInfo creates UserInfo with passed values.
func MakeUserInfo(id int64, username string, fullname string) UserInfo {
	return UserInfo{ID: id, UserName: username, FullName: fullname}
}

// Limits represents action limits.
type Limits struct {
	Follow   int
	UnFollow int
}
