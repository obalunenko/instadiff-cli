package media

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/obalunenko/getenv"
	"github.com/olegfedoseev/image-diff"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type file struct {
	path string
}

func Test_addBorders(t *testing.T) {
	if getenv.BoolOrDefault("CI", false) {
		t.Skip("Doesn't work on CI")
	}

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

			want := getReaderFromPath(t, tt.want.path)

			diffImageReaders(t, want, got)
		})
	}
}

func getReaderFromPath(tb testing.TB, path string) io.Reader {
	tb.Helper()

	content, err := os.ReadFile(path)
	require.NoError(tb, err)

	return bytes.NewReader(content)
}

func diffImageReaders(tb testing.TB, want, actual io.Reader) {
	tb.Helper()

	wantimg, err := decode(want)
	require.NoError(tb, err)

	actimg, err := decode(actual)
	require.NoError(tb, err)

	var eq bool

	d, percent, err := diff.CompareImages(wantimg, actimg)
	require.NoError(tb, err)

	if percent > 0.0 {
		name := strings.ReplaceAll(fmt.Sprintf("%s_diff.jpg", tb.Name()), "/", "_")

		f, err := os.Create(filepath.Join("testdata", name))
		require.NoError(tb, err)

		tb.Cleanup(func() {
			require.NoError(tb, f.Close())
		})

		r, err := encode(d)
		require.NoError(tb, err)

		buf := new(bytes.Buffer)

		_, err = buf.ReadFrom(r)
		require.NoError(tb, err)

		_, err = f.Write(buf.Bytes())
		require.NoError(tb, err)
	} else {
		eq = true
	}

	assert.True(tb, eq)
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
