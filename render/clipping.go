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
	// Считаем "расстояние" каждой вершины до плоскости отсечения (Near Plane)
	d0 := v0.Pos.W() - W_NEAR
	d1 := v1.Pos.W() - W_NEAR
	d2 := v2.Pos.W() - W_NEAR

	// Считаем, сколько вершин находится ПЕРЕД камерой (видимы)
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
		// Все вершины позади камеры
		return nil

	case 3:
		// Все вершины перед камерой — возвращаем как есть
		return [][3]VertexOut{{v0, v1, v2}}

	case 1:
		// 1 вершина внутри, 2 снаружи (получается 1 маленький треугольник)
		if d0 >= 0 {
			return [][3]VertexOut{{v0, lerp(v0, v1, d0/(d0-d1)), lerp(v0, v2, d0/(d0-d2))}}
		} else if d1 >= 0 {
			return [][3]VertexOut{{v1, lerp(v1, v2, d1/(d1-d2)), lerp(v1, v0, d1/(d1-d0))}}
		} else {
			return [][3]VertexOut{{v2, lerp(v2, v0, d2/(d2-d0)), lerp(v2, v1, d2/(d2-d1))}}
		}

	case 2:
		// 2 вершины внутри, 1 снаружи (получается четырехугольник, разбиваем на 2 треугольника)
		// Важно: строго сохраняем круговой порядок CCW!
		if d0 < 0 { // v0 снаружи (внутри v1 и v2)
			n1 := lerp(v1, v0, d1/(d1-d0))
			n2 := lerp(v2, v0, d2/(d2-d0))
			return [][3]VertexOut{{v1, v2, n1}, {v2, n2, n1}}
		} else if d1 < 0 { // v1 снаружи (внутри v2 и v0)
			n1 := lerp(v2, v1, d2/(d2-d1))
			n2 := lerp(v0, v1, d0/(d0-d1))
			return [][3]VertexOut{{v2, v0, n1}, {v0, n2, n1}}
		} else { // v2 снаружи (внутри v0 и v1)
			n1 := lerp(v0, v2, d0/(d0-d2))
			n2 := lerp(v1, v2, d1/(d1-d2))
			return [][3]VertexOut{{v0, v1, n1}, {v1, n2, n1}}
		}
	}

	return nil
}
