package diff

import (
	"errors"
	"image"
	"image/color"
	"os"
)

// CompareFiles will load images from files and then compare it with CompareImages
func CompareFiles(src, dst string) (diff image.Image, percent float64, err error) {
	srcImage, err := loadImage(src)
	if err != nil {
		return nil, 0, err
	}
	dstImage, err := loadImage(dst)
	if err != nil {
		return nil, 0, err
	}
	return CompareImages(srcImage, dstImage)
}

// CompareImages will check images size and pixel by pixel difference
func CompareImages(src, dst image.Image) (diff image.Image, percent float64, err error) {
	srcBounds := src.Bounds()
	dstBounds := dst.Bounds()
	if !boundsMatch(srcBounds, dstBounds) {
		return nil, 100.0, errors.New("image sizes don't match")
	}

	diffImage := image.NewRGBA(image.Rect(0, 0, srcBounds.Max.X, srcBounds.Max.Y))

	var differentPixels float64
	for y := srcBounds.Min.Y; y < srcBounds.Max.Y; y++ {
		for x := srcBounds.Min.X; x < srcBounds.Max.X; x++ {
			r, g, b, _ := dst.At(x, y).RGBA()
			diffImage.Set(x, y, color.RGBA{uint8(r), uint8(g), uint8(b), 64})

			if !isEqualColor(src.At(x, y), dst.At(x, y)) {
				differentPixels++
				// Add red dot in diff image
				diffImage.Set(x, y, color.RGBA{255, 0, 0, 255})
			}
		}
	}

	diffPercent := differentPixels / float64(srcBounds.Max.X*srcBounds.Max.Y) * 100
	return diffImage, diffPercent, nil
}

func isEqualColor(a, b color.Color) bool {
	r1, g1, b1, a1 := a.RGBA()
	r2, g2, b2, a2 := b.RGBA()

	return r1 == r2 && g1 == g2 && b1 == b2 && a1 == a2
}

func boundsMatch(a, b image.Rectangle) bool {
	return a.Min.X == b.Min.X && a.Min.Y == b.Min.Y && a.Max.X == b.Max.X && a.Max.Y == b.Max.Y
}

func loadImage(filename string) (image.Image, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		return nil, err
	}
	return img, nil
}
