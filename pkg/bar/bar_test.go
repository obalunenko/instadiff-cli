package bar

import (
	"sync"
	"testing"

	"github.com/schollz/progressbar/v2"
	"github.com/stretchr/testify/assert"
)

func TestBType_Valid(t *testing.T) {
	const notExisted BType = 999

	tests := []struct {
		name string
		bt   BType
		want bool
	}{
		{
			name: "invalid - unknownType",
			bt:   BTypeUnknown,
			want: false,
		},
		{
			name: "valid - rendered type",
			bt:   BTypeRendered,
			want: true,
		},
		{
			name: "invalid - not existed",
			bt:   notExisted,
			want: false,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			got := tt.bt.Valid()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNew(t *testing.T) {
	type args struct {
		cap     int
		barType BType
	}

	tests := []struct {
		name string
		args args
		want Bar
	}{
		{
			name: "make rendered",
			args: args{
				cap:     2,
				barType: BTypeRendered,
			},
			want: &realBar{
				bar:   progressbar.New(2),
				stop:  sync.Once{},
				wg:    sync.WaitGroup{},
				bchan: make(chan struct{}),
			},
		},
		{
			name: "make void",
			args: args{
				cap:     2,
				barType: BTypeVoid,
			},
			want: &voidBar{
				stop:  sync.Once{},
				wg:    sync.WaitGroup{},
				bchan: make(chan struct{}),
			},
		},
		{
			name: "make nil",
			args: args{
				cap:     2,
				barType: BTypeUnknown,
			},
			want: nil,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			got := New(tt.args.cap, tt.args.barType)
			assert.IsType(t, tt.want, got)
		})
	}
}
