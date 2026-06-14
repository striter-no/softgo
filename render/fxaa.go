package render

import (
	"math"

	"github.com/ungerik/go3d/vec3"
)

func getLuma(color vec3.T) float32 {
	return color[0]*0.299 + color[1]*0.587 + color[2]*0.114
}

// * ApplyFXAA
// - `src`: og pixel buffer
// - `dst`: destination buffer
func ApplyFXAA(src []vec3.T, dst []vec3.T, width, height int) {
	const edgeThreshold = 0.125

	getPixel := func(x, y int) vec3.T {
		if x < 0 {
			x = 0
		}
		if x >= width {
			x = width - 1
		}
		if y < 0 {
			y = 0
		}
		if y >= height {
			y = height - 1
		}
		return src[y*width+x]
	}

	for y := range height {
		for x := range width {
			cM := getPixel(x, y)   // Center
			cN := getPixel(x, y-1) // North
			cS := getPixel(x, y+1) // South
			cW := getPixel(x-1, y) // West
			cE := getPixel(x+1, y) // East

			lM := getLuma(cM)
			lN := getLuma(cN)
			lS := getLuma(cS)
			lW := getLuma(cW)
			lE := getLuma(cE)

			lMin := min(lM, min(lN, min(lS, min(lW, lE))))
			lMax := max(lM, max(lN, max(lS, max(lW, lE))))
			contrast := lMax - lMin

			if contrast < edgeThreshold {
				dst[y*width+x] = cM
				continue
			}

			edgeHorz := math.Abs(float64(lN-lM)) + math.Abs(float64(lS-lM))
			edgeVert := math.Abs(float64(lW-lM)) + math.Abs(float64(lE-lM))

			var blendColor vec3.T
			var blendFactor float32 = 0.25

			if edgeHorz > edgeVert {
				if math.Abs(float64(lN-lM)) > math.Abs(float64(lS-lM)) {
					blendColor = cN
				} else {
					blendColor = cS
				}
			} else {
				if math.Abs(float64(lW-lM)) > math.Abs(float64(lE-lM)) {
					blendColor = cW
				} else {
					blendColor = cE
				}
			}

			// result = Center * (1 - blend) + Target * blend
			dst[y*width+x] = vec3.T{
				cM[0]*(1.0-blendFactor) + blendColor[0]*blendFactor,
				cM[1]*(1.0-blendFactor) + blendColor[1]*blendFactor,
				cM[2]*(1.0-blendFactor) + blendColor[2]*blendFactor,
			}
		}
	}
}
