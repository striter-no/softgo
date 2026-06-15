package render

import (
	"math"

	"github.com/ungerik/go3d/vec3"
)

type Framebuffer struct {
	Width       int
	Height      int
	ColorBuffer []vec3.T
	DepthBuffer []float32
	HasColor    bool
}

func NewFramebuffer(width, height int, depthOnly bool) *Framebuffer {
	fb := &Framebuffer{
		Width:       width,
		Height:      height,
		DepthBuffer: make([]float32, width*height),
		HasColor:    !depthOnly,
	}

	if !depthOnly {
		fb.ColorBuffer = make([]vec3.T, width*height)
	}

	return fb
}

func (fb *Framebuffer) Clear(clearColor vec3.T) {
	for i := 0; i < fb.Width*fb.Height; i++ {
		if fb.HasColor {
			fb.ColorBuffer[i] = clearColor
		}
		fb.DepthBuffer[i] = math.MaxFloat32
	}
}
