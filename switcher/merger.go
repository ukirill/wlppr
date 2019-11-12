package switcher

import (
	"image"
	"image/draw"
	"os"
	"path/filepath"

	"github.com/ukirill/wlppr-go/internal"

	"github.com/anthonynsimon/bild/imgio"
	"github.com/anthonynsimon/bild/transform"
)

func (s baseSwitcher) mergeImage(paths []string) (fn string, err error) {
	if s.dispnum == 2 {
		fn, err = s.merge(paths[0], paths[1])
	} else {
		img := s.transform(paths[0])
		fn, err = randImageName(s.cachePath, ".png")
		imgio.Save(fn, img, imgio.PNG)
	}

	fn, err = filepath.Abs(fn)
	return
}

func (s *baseSwitcher) merge(path1, path2 string) (string, error) {
	// TODO: Optimize file operations
	img1 := s.transform(path1)
	img2 := s.transform(path2)
	r1 := image.Rect(0, 0, s.resW, s.resH)
	r2 := image.Rect(s.resW, 0, s.resW*2, s.resH)
	r := image.Rect(0, 0, s.resW*2, s.resH)
	rgba := image.NewRGBA(r)

	draw.Draw(rgba, r1, img1, image.Point{0, 0}, draw.Src)
	draw.Draw(rgba, r2, img2, image.Point{0, 0}, draw.Src)
	fn, err := randImageName(s.cachePath, ".png")
	imgio.Save(fn, rgba.SubImage(r), imgio.PNG)
	os.Remove(path1)
	os.Remove(path2)
	return fn, err
}

func (s *baseSwitcher) transform(path string) image.Image {
	img, _ := imgio.Open(path)
	trans := transform.Resize(img, s.resW, s.resH, transform.Lanczos)
	return trans.SubImage(image.Rect(0, 0, s.resW, s.resH))
}

func randImageName(basepath, fext string) (string, error) {
	fname := internal.RandStringBytes(16) + fext
	fname = filepath.Join(basepath, fname)
	return fname, nil
}
