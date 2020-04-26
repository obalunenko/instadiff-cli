package service

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/oleg-balunenko/instadiff-cli/internal/models"
)

func Test_getLostFollowers(t *testing.T) {
	type args struct {
		old []models.User
		new []models.User
	}

	var tests = []struct {
		name string
		args args
		want []models.User
	}{
		{
			name: "one lost user",
			args: args{
				old: []models.User{{ID: 1}, {ID: 2}, {ID: 10}},
				new: []models.User{{ID: 1}, {ID: 3}, {ID: 10}},
			},
			want: []models.User{{ID: 2}},
		},
		{
			name: "equal",
			args: args{
				old: []models.User{{ID: 1}, {ID: 2}, {ID: 10}},
				new: []models.User{{ID: 1}, {ID: 2}, {ID: 10}},
			},
			want: nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got := getLostFollowers(tt.args.old, tt.args.new)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_getNewFollowers(t *testing.T) {
	type args struct {
		old []models.User
		new []models.User
	}

	var tests = []struct {
		name string
		args args
		want []models.User
	}{
		{
			name: "new 1 user",
			args: args{
				old: []models.User{{ID: 11}, {ID: 2}, {ID: 12}},
				new: []models.User{{ID: 11}, {ID: 3}, {ID: 12}},
			},
			want: []models.User{{ID: 3}},
		},
		{
			name: "equal",
			args: args{
				old: []models.User{{ID: 11}, {ID: 2}, {ID: 12}},
				new: []models.User{{ID: 11}, {ID: 2}, {ID: 12}},
			},
			want: nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got := getNewFollowers(tt.args.old, tt.args.new)
			assert.Equal(t, tt.want, got)
		})
	}
}
