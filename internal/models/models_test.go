package models

import (
	"testing"

	"github.com/magiconair/properties/assert"
)

func TestUsersBatchType_Valid(t *testing.T) {
	const notExisted UsersBatchType = 999

	tests := []struct {
		name string
		i    UsersBatchType
		want bool
	}{
		{
			name: "valid batch type",
			i:    UsersBatchTypeBusinessAccounts,
			want: true,
		},
		{
			name: "invalid batch type",
			i:    notExisted,
			want: false,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			got := tt.i.Valid()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMakeUser(t *testing.T) {
	type args struct {
		id       int64
		username string
		fullname string
	}

	tests := []struct {
		name string
		args args
		want User
	}{
		{
			name: "make user",
			args: args{
				id:       1,
				username: "Test User",
				fullname: "Full test name",
			},
			want: User{
				ID:       1,
				UserName: "Test User",
				FullName: "Full test name",
			},
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			got := MakeUser(tt.args.id, tt.args.username, tt.args.fullname)
			assert.Equal(t, tt.want, got)
		})
	}
}
