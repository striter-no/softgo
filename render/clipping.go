package render

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/ungerik/go3d/vec2"
	"github.com/ungerik/go3d/vec3"
	"github.com/ungerik/go3d/vec4"
)

type VertexOut struct {
	Pos    mgl32.Vec4
	Color  vec4.T
	UV     vec2.T
	Normal vec3.T
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
	}
}

func ClipTriangle(v0, v1, v2 VertexOut, W_NEAR float32) [][3]VertexOut {
	var inside []VertexOut
	var outside []VertexOut

	for _, v := range []VertexOut{v0, v1, v2} {
		if v.Pos.W() >= W_NEAR {
			inside = append(inside, v)
		} else {
			outside = append(outside, v)
		}
	}

	switch len(inside) {
	case 0:
		return nil

	case 3:
		return [][3]VertexOut{{inside[0], inside[1], inside[2]}}

	case 1:
		t0 := (W_NEAR - inside[0].Pos.W()) / (outside[0].Pos.W() - inside[0].Pos.W())
		t1 := (W_NEAR - inside[0].Pos.W()) / (outside[1].Pos.W() - inside[0].Pos.W())

		newV1 := lerp(inside[0], outside[0], t0)
		newV2 := lerp(inside[0], outside[1], t1)

		return [][3]VertexOut{{inside[0], newV1, newV2}}

	case 2:
		t0 := (W_NEAR - inside[0].Pos.W()) / (outside[0].Pos.W() - inside[0].Pos.W())
		t1 := (W_NEAR - inside[1].Pos.W()) / (outside[0].Pos.W() - inside[1].Pos.W())

		newV0 := lerp(inside[0], outside[0], t0)
		newV1 := lerp(inside[1], outside[0], t1)

		return [][3]VertexOut{
			{inside[0], inside[1], newV0},
			{inside[1], newV1, newV0},
		}
	}

	return nil
}
