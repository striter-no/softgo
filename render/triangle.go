package render

import (
	"github.com/ungerik/go3d/vec2"
	"github.com/ungerik/go3d/vec3"
)

func edgeFunction(a, b, p vec2.T) float32 {
	return (b[0]-a[0])*(p[1]-a[1]) - (b[1]-a[1])*(p[0]-a[0])
}

func RasterizeTriangle(
	v0, v1, v2 vec2.T,
	z0, z1, z2 float32,
	w0, w1, w2 float32,
	c0, c1, c2 vec3.T,
	uv0, uv1, uv2 vec2.T,
	n0, n1, n2 vec3.T,
	checkDepth func(x, y int, z float32) bool,
	drawPixel func(x, y int, z float32, r, g, b uint8, u, v float32, nx, ny, nz float32),
) {
	minX := int(max(0, min(v0[0], v1[0], v2[0])))
	minY := int(max(0, min(v0[1], v1[1], v2[1])))
	maxX := int(max(v0[0], v1[0], v2[0]) + 0.5)
	maxY := int(max(v0[1], v1[1], v2[1]) + 0.5)

	area := edgeFunction(v0, v1, v2)

	if area >= 0 {
		return
	}

	invArea := 1.0 / area
	invW0 := 1.0 / w0
	invW1 := 1.0 / w1
	invW2 := 1.0 / w2

	stepX0 := -(v2[1] - v1[1])
	stepY0 := (v2[0] - v1[0])

	stepX1 := -(v0[1] - v2[1])
	stepY1 := (v0[0] - v2[0])

	stepX2 := -(v1[1] - v0[1])
	stepY2 := (v1[0] - v0[0])

	pStart := vec2.T{float32(minX) + 0.5, float32(minY) + 0.5}
	rowBc0 := edgeFunction(v1, v2, pStart)
	rowBc1 := edgeFunction(v2, v0, pStart)
	rowBc2 := edgeFunction(v0, v1, pStart)

	for y := minY; y <= maxY; y++ {
		bc0 := rowBc0
		bc1 := rowBc1
		bc2 := rowBc2

		for x := minX; x <= maxX; x++ {
			if bc0 <= 0 && bc1 <= 0 && bc2 <= 0 {
				normBc0 := bc0 * invArea
				normBc1 := bc1 * invArea
				normBc2 := bc2 * invArea

				interpZ := z0*normBc0 + z1*normBc1 + z2*normBc2

				if checkDepth(x, y, interpZ) {

					zFunc := normBc0*invW0 + normBc1*invW1 + normBc2*invW2
					invZFunc := 1.0 / zFunc

					interpU := (uv0[0]*invW0*normBc0 + uv1[0]*invW1*normBc1 + uv2[0]*invW2*normBc2) * invZFunc
					interpV := (uv0[1]*invW0*normBc0 + uv1[1]*invW1*normBc1 + uv2[1]*invW2*normBc2) * invZFunc

					interpNx := (n0[0]*invW0*normBc0 + n1[0]*invW1*normBc1 + n2[0]*invW2*normBc2) * invZFunc
					interpNy := (n0[1]*invW0*normBc0 + n1[1]*invW1*normBc1 + n2[1]*invW2*normBc2) * invZFunc
					interpNz := (n0[2]*invW0*normBc0 + n1[2]*invW1*normBc1 + n2[2]*invW2*normBc2) * invZFunc

					r := c0[0]*normBc0 + c1[0]*normBc1 + c2[0]*normBc2
					g := c0[1]*normBc0 + c1[1]*normBc1 + c2[1]*normBc2
					b := c0[2]*normBc0 + c1[2]*normBc1 + c2[2]*normBc2

					r = min(255, max(0, r))
					g = min(255, max(0, g))
					b = min(255, max(0, b))

					drawPixel(x, y, interpZ, uint8(r), uint8(g), uint8(b), interpU, interpV, interpNx, interpNy, interpNz)
				}
			}

			bc0 += stepX0
			bc1 += stepX1
			bc2 += stepX2
		}

		rowBc0 += stepY0
		rowBc1 += stepY1
		rowBc2 += stepY2
	}
}
