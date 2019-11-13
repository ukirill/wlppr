package merger

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"os"

	"github.com/anthonynsimon/bild/transform"
)

type Pixel struct {
	Point image.Point
	Color color.Color
}

type ImageDim struct {
	X, Y int
}

func openDecode(filepath string) (image.Image, string, error) {
	imgFile, err := os.Open(filepath)
	if err != nil {
		return nil, "", err
	}
	defer imgFile.Close()
	img, format, err := image.Decode(imgFile)
	if err != nil {
		return nil, "", err
	}
	return img, format, nil
}

func decodePixels(img image.Image, offsetX, offsetY int) []*Pixel {
	pixels := []*Pixel{}
	for y := 0; y <= img.Bounds().Max.Y; y++ {
		for x := 0; x <= img.Bounds().Max.X; x++ {
			p := &Pixel{
				Point: image.Point{X: x + offsetX, Y: y + offsetY},
				Color: img.At(x, y),
			}
			pixels = append(pixels, p)
		}
	}
	return pixels
}

func resize(img image.Image, size *ImageDim) image.Image{
	dx := img.Bounds().Dx()
	dy := img.Bounds().Dy()
	ratio := (float64(dx)/float64(size.X), float64(dy)/float64(size.Y))
	transform.Resize(dx/ratio)
}

func newProportionalSize()

func Process(imgs []image.Image, sizes []*ImageDim) {
	for i, img := range imgs {

	}
}



func main() {
	img1, _, err := openDecode("makey.png")
	if err != nil {
		panic(err)
	}
	img2, _, err := openDecode("sample.jpg")
	if err != nil {
		panic(err)
	}
	// collect pixel data from each image
	pixels1 := decodePixels(img1, 0, 0)
	// the second image has a Y-offset of img1's max Y (appended at bottom)
	pixels2 := decodePixels(img2, 0, img1.Bounds().Max.Y)
	pixelSum := append(pixels1, pixels2...)

	// Set a new size for the new image equal to the max width
	// of bigger image and max height of two images combined
	newRect := image.Rectangle{
		Min: img1.Bounds().Min,
		Max: image.Point{
			X: img2.Bounds().Max.X,
			Y: img2.Bounds().Max.Y + img1.Bounds().Max.Y,
		},
	}
	finImage := image.NewRGBA(newRect)
	// This is the cool part, all you have to do is loop through
	// each Pixel and set the image's color on the go
	for _, px := range pixelSum {
		finImage.Set(
			px.Point.X,
			px.Point.Y,
			px.Color,
		)
	}
	draw.Draw(finImage, finImage.Bounds(), finImage, image.Point{0, 0}, draw.Src)

	// Create a new file and write to it
	out, err := os.Create("./output.png")
	if err != nil {
		panic(err)
		os.Exit(1)
	}
	err = png.Encode(out, finImage)
	if err != nil {
		panic(err)
		os.Exit(1)
	}
}