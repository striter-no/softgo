package render

import (
	"math"

	"github.com/ungerik/go3d/vec2"
	"github.com/ungerik/go3d/vec3"
)

func edgeFunction(a, b, p vec2.T) float32 {
	return (b[0]-a[0])*(p[1]-a[1]) - (b[1]-a[1])*(p[0]-a[0])
}

func RasterizeTriangle(v0, v1, v2 vec2.T, c0, c1, c2 vec3.T, drawPixel func(x, y int, r, g, b uint8)) {
	minX := int(math.Max(0, float64(min(v0[0], v1[0], v2[0]))))
	minY := int(math.Max(0, float64(min(v0[1], v1[1], v2[1]))))
	maxX := int(math.Ceil(float64(max(v0[0], v1[0], v2[0]))))
	maxY := int(math.Ceil(float64(max(v0[1], v1[1], v2[1]))))

	area := edgeFunction(v0, v1, v2)
	if area == 0 {
		return
	}

	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			p := vec2.T{float32(x) + 0.5, float32(y) + 0.5}

			w0 := edgeFunction(v1, v2, p)
			w1 := edgeFunction(v2, v0, p)
			w2 := edgeFunction(v0, v1, p)

			isInside := (w0 >= 0 && w1 >= 0 && w2 >= 0) || (w0 <= 0 && w1 <= 0 && w2 <= 0)

			if !isInside {
				continue
			}

			w0 /= area
			w1 /= area
			w2 /= area

			r := c0[0]*w0 + c1[0]*w1 + c2[0]*w2
			g := c0[1]*w0 + c1[1]*w1 + c2[1]*w2
			b := c0[2]*w0 + c1[2]*w1 + c2[2]*w2

			r = min(255, max(0, r))
			g = min(255, max(0, g))
			b = min(255, max(0, b))

			drawPixel(x, y, uint8(r), uint8(g), uint8(b))
		}
	}
}
