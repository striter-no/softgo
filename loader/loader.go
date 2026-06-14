package textures

import (
	"image"
	"image/draw"
	"image/gif"
	"log"
	"os"

	_ "image/jpeg"
	_ "image/png"

	"github.com/striter-no/softgo/render"
	"github.com/ungerik/go3d/vec4"
)

type Animation struct {
	Frames    []render.Texture
	Delays    []int
	LoopCount int
}

func ConvertGIFToAnimation(filename string) (*Animation, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	g, err := gif.DecodeAll(file)
	if err != nil {
		return nil, err
	}

	anim := &Animation{
		Frames:    make([]render.Texture, len(g.Image)),
		Delays:    g.Delay,
		LoopCount: g.LoopCount,
	}

	bounds := image.Rect(0, 0, g.Config.Width, g.Config.Height)
	canvas := image.NewRGBA(bounds)

	for i, img := range g.Image {
		draw.Draw(canvas, img.Bounds(), img, img.Bounds().Min, draw.Over)

		anim.Frames[i] = *rgbaToTexture(canvas)

		if g.Disposal[i] == gif.DisposalBackground {
			draw.Draw(canvas, img.Bounds(), image.Transparent, image.Point{}, draw.Src)
		}
	}

	return anim, nil
}

func getImageFromFilePath(filePath string) (image.Image, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	return img, err
}

func ConvertImageToTexture(filename string) *render.Texture {
	img, err := getImageFromFilePath(filename)
	if err != nil {
		log.Printf("Failed to get image from path '%s': %v", filename, err)
		return nil
	}

	if rgba, ok := img.(*image.RGBA); ok {
		return rgbaToTexture(rgba)
	}

	b := img.Bounds()
	dst := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))

	draw.Draw(dst, dst.Bounds(), img, b.Min, draw.Src)

	return rgbaToTexture(dst)
}

func rgbaToTexture(img *image.RGBA) *render.Texture {
	bounds := img.Bounds()
	width := bounds.Max.X
	height := bounds.Max.Y

	pixels := make([]vec4.T, width*height)

	for y := range height {
		for x := range width {
			offset := (y-img.Rect.Min.Y)*img.Stride + (x-img.Rect.Min.X)*4

			r := img.Pix[offset]
			g := img.Pix[offset+1]
			b := img.Pix[offset+2]
			a := img.Pix[offset+3]

			pixels[y*width+x] = vec4.T{
				float32(r),
				float32(g),
				float32(b),
				float32(a),
			}
		}
	}

	return &render.Texture{
		Width:  width,
		Height: height,
		Pixels: pixels,
	}
}
