package render

import "github.com/ungerik/go3d/vec4"

type Texture struct {
	Width, Height int
	Pixels        []vec4.T
}

// gets pixel via UV
func (t *Texture) Sample(u, v float32) vec4.T {
	if t.Width == 0 || t.Height == 0 || len(t.Pixels) == 0 {
		return vec4.T{1, 0, 1, 1}
	}

	x := int(u * float32(t.Width))
	y := int(v * float32(t.Height))

	if x < 0 {
		x = 0
	}
	if x >= t.Width {
		x = t.Width - 1
	}
	if y < 0 {
		y = 0
	}
	if y >= t.Height {
		y = t.Height - 1
	}

	return t.Pixels[y*t.Width+x]
}
