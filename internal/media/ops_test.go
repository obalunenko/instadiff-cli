package media

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type file struct {
	path string
}

func Test_addBorders(t *testing.T) {
	type file struct {
		path string
	}

	type args struct {
		w     int
		h     int
		input file
	}

	tests := []struct {
		name    string
		args    args
		want    file
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "400x400_square_smaller",
			args: args{
				w: 1080,
				h: 1920,
				input: file{
					path: filepath.Join("testdata", "400x400.jpg"),
				},
			},
			want: file{
				path: filepath.Join("testdata", "400x400_square_smaller.jpg"),
			},

			wantErr: assert.NoError,
		},
		{
			name: "1200x2600_rect_bigger",
			args: args{
				w: 1080,
				h: 1920,
				input: file{
					path: filepath.Join("testdata", "1200x2600.jpg"),
				},
			},
			want: file{
				path: filepath.Join("testdata", "1200x2600_rect_bigger.jpg"),
			},
			wantErr: assert.NoError,
		},
		{
			name: "1080x1090_rect_smaller",
			args: args{
				w: 1080,
				h: 1920,
				input: file{
					path: filepath.Join("testdata", "1080x1090.jpg"),
				},
			},
			want: file{
				path: filepath.Join("testdata", "1080x1090_rect_smaller.jpg"),
			},
			wantErr: assert.NoError,
		},
		{
			name: "1080x1920_rect_exact",
			args: args{
				w: 1080,
				h: 1920,
				input: file{
					path: filepath.Join("testdata", "1080x1920.jpg"),
				},
			},
			want: file{
				path: filepath.Join("testdata", "1080x1920_rect_exact.jpg"),
			},
			wantErr: assert.NoError,
		},
		{
			name: "1440x960_heic",
			args: args{
				w: 1080,
				h: 1920,
				input: file{
					path: filepath.Join("testdata", "sample1.heic"),
				},
			},
			want: file{
				path: "",
			},
			wantErr: assert.Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := addBorders(getReaderFromPath(t, tt.args.input.path), tt.args.w, tt.args.h)
			if !tt.wantErr(t, err) {
				return
			}

			if tt.want.path == "" {
				return
			}

			dst, err := os.Create(filepath.Join("testdata", fmt.Sprintf("%s.jpg", tt.name)))
			require.NoError(t, err)

			t.Cleanup(func() {
				require.NoError(t, dst.Close())
			})

			buf := new(bytes.Buffer)

			_, err = buf.ReadFrom(got)
			require.NoError(t, err)

			content := buf.Bytes()

			_, err = dst.Write(content)
			require.NoError(t, err)

			want := getReaderFromPath(t, tt.want.path)

			diff(t, want, bytes.NewReader(content))
		})
	}
}

func getReaderFromPath(tb testing.TB, path string) io.Reader {
	tb.Helper()

	content, err := os.ReadFile(path)
	require.NoError(tb, err)

	return bytes.NewReader(content)
}

func diff(tb testing.TB, want, actual io.Reader) {
	tb.Helper()

	h1, h2 := sha256.New(), sha256.New()

	_, err := io.Copy(h1, want)
	require.NoError(tb, err)

	_, err = io.Copy(h2, actual)
	require.NoError(tb, err)

	assert.True(tb, bytes.Equal(h1.Sum(nil), h2.Sum(nil)))
}

func TestGetMediaFile(t *testing.T) {
	ctx := context.Background()

	type args struct {
		ctx   context.Context
		fpath string
	}
	tests := []struct {
		name    string
		args    args
		want    file
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "",
			args: args{
				ctx:   ctx,
				fpath: filepath.Join("testdata", "1080x1090.jpg"),
			},
			want: file{
				path: filepath.Join("testdata", "1080x1090.jpg"),
			},
			wantErr: assert.NoError,
		},
		{
			name: "",
			args: args{
				ctx:   ctx,
				fpath: filepath.Join("testdata", "sample1.heic"),
			},
			want: file{
				path: filepath.Join("testdata", "sample1.heic"),
			},
			wantErr: assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetMediaFile(tt.args.ctx, tt.args.fpath)
			if !tt.wantErr(t, err, fmt.Sprintf("GetMediaFile(%v, %v)", tt.args.ctx, tt.args.fpath)) {
				return
			}

			want := getReaderFromPath(t, tt.want.path)

			diff(t, want, got)
		})
	}
}

func Test_getFileContentType(t *testing.T) {
	tests := []struct {
		name    string
		input   file
		want    string
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "jpeg",
			input: file{
				path: filepath.Join("testdata", "1080x1090.jpg"),
			},
			want:    "image/jpeg",
			wantErr: assert.NoError,
		},
		{
			name: "jpeg converted",
			input: file{
				path: filepath.Join("testdata", "400x400_square_smaller.jpg"),
			},
			want:    "image/jpeg",
			wantErr: assert.NoError,
		},
		{
			name: "heic",
			input: file{
				path: filepath.Join("testdata", "sample1.heic"),
			},
			want:    "application/octet-stream",
			wantErr: assert.NoError,
		},
		{
			name: "png",
			input: file{
				path: filepath.Join("testdata", "tavern.png"),
			},
			want:    "image/png",
			wantErr: assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := getReaderFromPath(t, tt.input.path)

			got, err := getFileContentType(r)
			if !tt.wantErr(t, err) {
				return
			}

			assert.Equal(t, tt.want, got)
		})
	}
}
