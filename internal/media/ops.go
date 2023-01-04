package media

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"io"
	"net/http"
	"os"
	"path"

	log "github.com/obalunenko/logger"

	"github.com/obalunenko/instadiff-cli/internal/utils"

	"github.com/disintegration/imaging"
)

// AddBorders add white borders to the image.
func AddBorders(r io.Reader, mt Type) (io.Reader, error) {
	var (
		w, h int
	)

	switch mt {
	case TypeStoryPhoto:
		w = 1080
		h = 1920
	default:
		return nil, fmt.Errorf("unsupported media type[%s]", mt.String())
	}

	return addBorders(r, w, h)
}

func addBorders(r io.Reader, w, h int) (io.Reader, error) {
	img, err := decode(r)
	if err != nil {
		return nil, err
	}

	const (
		b = 100
	)

	img = resizeImage(img, w-b, h-b)
	img = resizeImage(img, w, h)

	return encode(img)
}

func decode(r io.Reader) (image.Image, error) {
	// get file type
	content, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	ct, err := getFileContentType(bytes.NewReader(content))
	if err != nil {
		return nil, fmt.Errorf("get file content type: %w", err)
	}

	log.WithFields(context.TODO(), log.Fields{
		"file_type": ct,
	}).Info("Media file")

	var img image.Image

	switch ct {
	case "application/octet-stream":
		img, err = decodeHEIF(bytes.NewReader(content))
	default:
		img, _, err = image.Decode(bytes.NewReader(content))
	}
	if err != nil {
		return nil, err
	}

	return img, nil
}

func encode(img image.Image) (io.Reader, error) {
	buf := new(bytes.Buffer)

	if err := jpeg.Encode(buf, img, nil); err != nil {
		return nil, err
	}

	return buf, nil
}

func resizeImage(src image.Image, w, h int) image.Image {
	dst := imaging.New(w, h, color.White)

	tmp := imaging.Fit(src, w, h, imaging.CatmullRom)
	iw := tmp.Rect.Bounds().Max.X - tmp.Rect.Bounds().Min.X
	ih := tmp.Rect.Bounds().Max.Y - tmp.Rect.Bounds().Min.Y

	return imaging.Overlay(dst, tmp, image.Pt(w/2-iw/2, h/2-ih/2), 1.0)
}

var errEmptyFilePath = errors.New("path is empty")

func GetMediaFile(ctx context.Context, fpath string) (io.Reader, error) {
	f, err := os.Open(path.Clean(fpath))
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}

	defer func() {
		utils.LogError(ctx, f.Close(), "Failed to close file descriptor")
	}()

	content, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	ct, err := getFileContentType(bytes.NewReader(content))
	if err != nil {
		return nil, fmt.Errorf("get file content type: %w", err)
	}

	log.WithFields(ctx, log.Fields{
		"file_type": ct,
		"file_path": fpath,
	}).Info("Media file")

	return bytes.NewReader(content), nil
}

func getFileContentType(f io.Reader) (string, error) {
	// to sniff the content type only the first
	// 512 bytes are used.
	const sniffLen = 512

	buf := make([]byte, sniffLen)

	_, err := f.Read(buf)
	if err != nil {
		return "", err
	}

	// the function that actually does the trick
	ct := http.DetectContentType(buf)

	return ct, nil
}

func decodeHEIF(r io.Reader) (image.Image, error) {
	// There is no way to implement heif decoder without C or external tools usage.

	return nil, errors.New("unsupported format")
}
