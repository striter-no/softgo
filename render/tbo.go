package render

import (
	"github.com/ungerik/go3d/vec3"
	"github.com/ungerik/go3d/vec4"
)

// Triangle Buffer Object
type TBO struct {
	V0, V1, V2 vec3.T
	C0, C1, C2 vec4.T
}
