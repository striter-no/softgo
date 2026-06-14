package main

import (
	"context"
	"fmt"
	"math"
	"os"
	"os/signal"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/striter-no/softgo/api"
	"github.com/striter-no/softgo/render"
	"github.com/striter-no/stg/graphics"
	"github.com/ungerik/go3d/vec3"
	"github.com/ungerik/go3d/vec4"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	s, err := api.NewRenderScreen(ctx)
	if err != nil {
		panic(err)
	}

	defer s.End()
	s.Init()

	tri := render.TBO{
		V0: vec3.T{0.0, 0.5, 0}, V1: vec3.T{0.5, -0.5, 0}, V2: vec3.T{-0.5, -0.5, 0},
		C0: vec4.T{255, 0, 0, 0}, C1: vec4.T{0, 255, 0, 0}, C2: vec4.T{0, 0, 255, 0},
	}

	s.FragShader = api.NewFragShader(fragShader)
	s.VertexShader = api.NewVertexShader(vertShader)

	var tick int64
	var mvp mgl32.Mat4
	for s.IsOpen() {
		mvp = updateMVP(tick, s)
		s.VertexShader.SetUniform("mvp", &mvp)

		s.FeedTBO(tri)
		tri.V0[2] = float32(math.Sin(float64(tick)*0.005) * 2)
		tri.V1[2] = float32(math.Sin(float64(tick)*0.005) * 2)
		tri.V2[2] = float32(math.Sin(float64(tick)*0.005) * 2)

		if err := s.Blit(); err != nil {
			panic(err)
		}

		s.Screen.SetText(0, 0, fmt.Sprintf("FPS: %f", s.CurrentFPS), graphics.NewFGPixel(255, 255, 255, ""))
		s.Screen.Blit()

		tick++
	}
}

func updateMVP(tick int64, rs *api.RenderScreen) mgl32.Mat4 {
	aspect := float32(rs.Screen.Width) / (float32(rs.Screen.Height) * 0.5)
	proj := mgl32.Perspective(mgl32.DegToRad(60), aspect, 0.1, 100.0)
	view := mgl32.LookAtV(
		mgl32.Vec3{0, 0, 1.0}, // Position
		mgl32.Vec3{0, 0, 0},   // Look AT point
		mgl32.Vec3{0, 1, 0},   // Direction?
	)

	angle := float32(tick) * 0.001
	model := mgl32.HomogRotate3DY(angle)

	return proj.Mul4(view).Mul4(model)
}

func fragShader(x float32, y float32, col vec4.T, _ *api.FragmentShader) vec4.T {
	return col
}

func vertShader(vert *vec3.T, s *api.VertexShader) mgl32.Vec4 {
	mvpAny, _ := s.GetUniform("mvp")
	mvp := mvpAny.(*mgl32.Mat4)

	v := mgl32.Vec4{vert[0], vert[1], vert[2], 1.0}
	return mvp.Mul4x1(v)
}
