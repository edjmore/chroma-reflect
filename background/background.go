package background

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"os/user"
	"path/filepath"
	"time"
)

var path string

func init() {
	usr, err := user.Current()
	if err != nil {
		panic(fmt.Errorf("unable to get user home dir: %v", err))
	}
	path = filepath.Join(usr.HomeDir, "AppData", "Roaming", "Microsoft", "Windows", "Themes", "TranscodedWallpaper")
}

// Get the last modification time for the background image file.
func ModTime() (time.Time, error) {
	info, err := os.Stat(path)
	if err != nil {
		var m time.Time
		return m, err
	}
	return info.ModTime(), nil
}

// Get the average color of the current backgroud image.
func Color() (int, error) {
	img, err := load()
	if err != nil {
		return 0, err
	}

	return averageColor(img, img.Bounds()), nil
}

// Divide the background image into a grid, and get the average color for each cell.
func Colors() ([6][22]int, error) {
	var colors [6][22]int

	img, err := load()
	if err != nil {
		return colors, err
	}

	b := img.Bounds()
	w, h := b.Dx()/16, b.Dy()/6

	for x := 0; x < 16; x++ {
		for y := 0; y < 6; y++ {
			colors[y][x] = dominantColor(
				img,
				b.Intersect(image.Rect(x*w, y*h, (x+1)*w, (y+1)*h)),
			)
		}
	}
	return colors, nil
}

func load() (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	img, err := jpeg.Decode(f)
	if err != nil {
		img, err = png.Decode(f)
	}
	return img, err
}

func averageColor(img image.Image, b image.Rectangle) int {
	var rAvg, gAvg, bAvg uint32
	for x := b.Min.X; x < b.Max.X; x++ {
		for y := b.Min.Y; y < b.Max.Y; y++ {
			px := img.At(x, y)
			r, g, b, _ := px.RGBA()
			rAvg += r >> 8
			gAvg += g >> 8
			bAvg += b >> 8
		}
	}

	area := uint32(b.Dx() * b.Dy())
	return int((bAvg/area)<<16 | (gAvg/area)<<8 | rAvg/area)
}

func dominantColor(img image.Image, b image.Rectangle) int {
	var color, count int
	bins := make(map[int]int)

	for x := b.Min.X; x < b.Max.X; x++ {
		for y := b.Min.Y; y < b.Max.Y; y++ {
			px := img.At(x, y)
			r, g, b, _ := px.RGBA()
			bgr := int((b>>8)<<16 | (g>>8)<<8 | (r >> 8))

			bins[bgr] += 1
			if bins[bgr] > count {
				color = bgr
				count = bins[bgr]
			}
		}
	}

	return color
}
