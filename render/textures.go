package render

import "github.com/ungerik/go3d/vec4"

type Texture struct {
	Width, Height int
	Pixels        []vec4.T

	Mipmaps [][][]vec4.T
	Widths  []int
	Heights []int

	MipmapsDistanceStep float32
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

func (t *Texture) SampleLod(u, v float32, distance float32) vec4.T {

	lod := max(int(distance/t.MipmapsDistanceStep), 0)
	if lod >= len(t.Mipmaps) {
		lod = len(t.Mipmaps) - 1
	}

	w := float32(t.Widths[lod])
	h := float32(t.Heights[lod])

	x := int(u*w) % int(w)
	y := int(v*h) % int(h)
	if x < 0 {
		x += int(w)
	}
	if y < 0 {
		y += int(h)
	}

	return t.Mipmaps[lod][y][x]
}

func (t *Texture) GenerateMipmaps(maxLevels int, distanceStep float32) {
	if t.Width == 0 || t.Height == 0 || len(t.Pixels) == 0 {
		return
	}

	t.MipmapsDistanceStep = distanceStep
	t.Mipmaps = make([][][]vec4.T, 0)
	t.Widths = make([]int, 0)
	t.Heights = make([]int, 0)

	level0 := make([][]vec4.T, t.Height)
	for y := 0; y < t.Height; y++ {
		level0[y] = make([]vec4.T, t.Width)
		for x := 0; x < t.Width; x++ {
			level0[y][x] = t.Pixels[y*t.Width+x]
		}
	}

	t.Mipmaps = append(t.Mipmaps, level0)
	t.Widths = append(t.Widths, t.Width)
	t.Heights = append(t.Heights, t.Height)

	currentLevel := level0
	currentW := t.Width
	currentH := t.Height

	for i := 1; (maxLevels <= 0) || (i < maxLevels); i++ {
		if currentW <= 1 && currentH <= 1 {
			break
		}

		nextLevel := generateNextMipmap(currentLevel, currentW, currentH)

		currentW /= 2
		currentH /= 2

		if currentW < 1 {
			currentW = 1
		}
		if currentH < 1 {
			currentH = 1
		}

		t.Mipmaps = append(t.Mipmaps, nextLevel)
		t.Widths = append(t.Widths, currentW)
		t.Heights = append(t.Heights, currentH)

		currentLevel = nextLevel
	}
}

func generateNextMipmap(current [][]vec4.T, w, h int) [][]vec4.T {
	newW := w / 2
	newH := h / 2

	if newW < 1 {
		newW = 1
	}
	if newH < 1 {
		newH = 1
	}

	next := make([][]vec4.T, newH)
	for y := 0; y < newH; y++ {
		next[y] = make([]vec4.T, newW)
		for x := 0; x < newW; x++ {

			origX := x * 2
			origY := y * 2

			origX1 := origX + 1
			if origX1 >= w {
				origX1 = origX
			}
			origY1 := origY + 1
			if origY1 >= h {
				origY1 = origY
			}

			c1 := current[origY][origX]
			c2 := current[origY][origX1]
			c3 := current[origY1][origX]
			c4 := current[origY1][origX1]

			next[y][x] = vec4.T{
				(c1[0] + c2[0] + c3[0] + c4[0]) * 0.25,
				(c1[1] + c2[1] + c3[1] + c4[1]) * 0.25,
				(c1[2] + c2[2] + c3[2] + c4[2]) * 0.25,
				(c1[3] + c2[3] + c3[3] + c4[3]) * 0.25,
			}
		}
	}
	return next
}
