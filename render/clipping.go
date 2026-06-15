package render

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/ungerik/go3d/vec2"
	"github.com/ungerik/go3d/vec3"
	"github.com/ungerik/go3d/vec4"
)

type VertexOut struct {
	Pos     mgl32.Vec4
	Color   vec4.T
	UV      vec2.T
	Normal  vec3.T
	FragPos vec3.T
}

func lerp(a, b VertexOut, t float32) VertexOut {
	return VertexOut{
		Pos: mgl32.Vec4{
			a.Pos[0] + (b.Pos[0]-a.Pos[0])*t,
			a.Pos[1] + (b.Pos[1]-a.Pos[1])*t,
			a.Pos[2] + (b.Pos[2]-a.Pos[2])*t,
			a.Pos[3] + (b.Pos[3]-a.Pos[3])*t,
		},
		Color: vec4.T{
			a.Color[0] + (b.Color[0]-a.Color[0])*t,
			a.Color[1] + (b.Color[1]-a.Color[1])*t,
			a.Color[2] + (b.Color[2]-a.Color[2])*t,
			a.Color[3] + (b.Color[3]-a.Color[3])*t,
		},
		UV: vec2.T{
			a.UV[0] + (b.UV[0]-a.UV[0])*t,
			a.UV[1] + (b.UV[1]-a.UV[1])*t,
		},
		Normal: vec3.T{
			a.Normal[0] + (b.Normal[0]-a.Normal[0])*t,
			a.Normal[1] + (b.Normal[1]-a.Normal[1])*t,
			a.Normal[2] + (b.Normal[2]-a.Normal[2])*t,
		},
		FragPos: vec3.T{
			a.FragPos[0] + (b.FragPos[0]-a.FragPos[0])*t,
			a.FragPos[1] + (b.FragPos[1]-a.FragPos[1])*t,
			a.FragPos[2] + (b.FragPos[2]-a.FragPos[2])*t,
		},
	}
}

func ClipTriangle(v0, v1, v2 VertexOut, W_NEAR float32) ([2][3]VertexOut, int) {
	var out [2][3]VertexOut

	d0 := v0.Pos.W() - W_NEAR
	d1 := v1.Pos.W() - W_NEAR
	d2 := v2.Pos.W() - W_NEAR

	inCount := 0
	if d0 >= 0 {
		inCount++
	}
	if d1 >= 0 {
		inCount++
	}
	if d2 >= 0 {
		inCount++
	}

	switch inCount {
	case 0:
		return out, 0
	case 3:
		out[0] = [3]VertexOut{v0, v1, v2}
		return out, 1
	case 1:
		if d0 >= 0 {
			out[0] = [3]VertexOut{v0, lerp(v0, v1, d0/(d0-d1)), lerp(v0, v2, d0/(d0-d2))}
		} else if d1 >= 0 {
			out[0] = [3]VertexOut{v1, lerp(v1, v2, d1/(d1-d2)), lerp(v1, v0, d1/(d1-d0))}
		} else {
			out[0] = [3]VertexOut{v2, lerp(v2, v0, d2/(d2-d0)), lerp(v2, v1, d2/(d2-d1))}
		}
		return out, 1
	case 2:
		if d0 < 0 {
			n1 := lerp(v1, v0, d1/(d1-d0))
			n2 := lerp(v2, v0, d2/(d2-d0))
			out[0] = [3]VertexOut{v1, v2, n1}
			out[1] = [3]VertexOut{v2, n2, n1}
		} else if d1 < 0 {
			n1 := lerp(v2, v1, d2/(d2-d1))
			n2 := lerp(v0, v1, d0/(d0-d1))
			out[0] = [3]VertexOut{v2, v0, n1}
			out[1] = [3]VertexOut{v0, n2, n1}
		} else {
			n1 := lerp(v0, v2, d0/(d0-d2))
			n2 := lerp(v1, v2, d1/(d1-d2))
			out[0] = [3]VertexOut{v0, v1, n1}
			out[1] = [3]VertexOut{v1, n2, n1}
		}
		return out, 2
	}
	return out, 0
}
