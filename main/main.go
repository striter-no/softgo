package main

import (
	"context"
	"fmt"
	"math"
	"os"
	"os/signal"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/striter-no/softgo/api"
	"github.com/striter-no/softgo/api/assets"
	"github.com/striter-no/softgo/api/keyboard"
	"github.com/striter-no/softgo/api/mouse"
	textures "github.com/striter-no/softgo/loader"

	"github.com/striter-no/softgo/render"
	"github.com/striter-no/stg/graphics"
	"github.com/ungerik/go3d/vec2"
	"github.com/ungerik/go3d/vec3"
	"github.com/ungerik/go3d/vec4"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	anim, err := textures.ConvertGIFToAnimation("./assets/rickroll.gif")
	if err != nil {
		panic(err)
	}

	winMouse, err := mouse.NewWindowMouse()
	if err != nil {
		panic(err)
	}

	winMouse.LockCursor()
	winMouse.HideMouse()

	defer winMouse.UnlockCursor()
	defer winMouse.ShowMouse()

	winKeyboard, err := keyboard.NewWindowKeyboard()
	if err != nil {
		panic(err)
	}

	defer winKeyboard.Close()

	camera := api.NewCamera(
		vec3.T{0, 0, 0},
		.08,
		.1,
		winMouse, winKeyboard,
	)

	camera.Locked = true

	s, err := api.NewRenderScreen(ctx)
	if err != nil {
		panic(err)
	}

	defer s.End()
	s.Init()

	mesh, err := assets.LoadOBJ("./assets/suzanne.obj")
	if err != nil {
		panic(err)
	}

	tex := &render.Texture{
		Width: 312, Height: 312,
		Pixels: generateCheckerboard(312, 312),
	}

	s.FragShader = api.NewFragShader(fragShader)
	s.VertexShader = api.NewVertexShader(vertShader)

	s.FragShader.SetUniform("tex", tex)

	s.SSAAFactor = 2

	var tick int64
	var ftick int
	for s.IsOpen() {
		winMouse.PollEvents()
		winKeyboard.PollEvents()

		aspect := float32(s.Screen.Width) / (float32(s.Screen.Height))
		camera.UpdateOnHID(aspect)

		angle := float32(tick) * 0.001
		modelMatrix := mgl32.HomogRotate3DY(angle)
		scaleMatrix := mgl32.Scale3D(2.5, 2.5, 2.5)

		s.VertexShader.SetUniform("model", &modelMatrix)

		finalMVP := camera.VP.Mul4(modelMatrix).Mul4(scaleMatrix)

		s.VertexShader.SetUniform("mvp", &finalMVP)

		if tick%7 == 0 {
			s.FragShader.SetUniform("tex", &anim.Frames[ftick%len(anim.Frames)])
			ftick++
		}

		for i := range mesh {
			s.FeedTBO(mesh[i])
		}

		if err := s.Blit(); err != nil {
			panic(err)
		}

		s.Screen.SetText(0, 0, fmt.Sprintf("FPS: %f", s.CurrentFPS), graphics.NewFGPixel(255, 255, 255, ""))
		s.Screen.Blit()

		tick++
	}
}

func fragShader(u float32, v float32, col vec4.T, norm vec3.T, s *api.FragmentShader) vec4.T {
	texAny, _ := s.GetUniform("tex")
	tex := texAny.(*render.Texture)
	texColor := tex.Sample(u, v)

	length := float32(math.Sqrt(float64(norm[0]*norm[0] + norm[1]*norm[1] + norm[2]*norm[2])))
	if length > 0 {
		norm[0] /= length
		norm[1] /= length
		norm[2] /= length
	}

	lightDir := vec3.T{0.5, 1.0, 0.5}

	lightLen := float32(math.Sqrt(float64(lightDir[0]*lightDir[0] + lightDir[1]*lightDir[1] + lightDir[2]*lightDir[2])))
	lightDir[0] /= lightLen
	lightDir[1] /= lightLen
	lightDir[2] /= lightLen

	diffuse := norm[0]*lightDir[0] + norm[1]*lightDir[1] + norm[2]*lightDir[2]

	if diffuse < 0 {
		diffuse = 0
	}

	ambient := float32(0.2)
	intensity := ambient + diffuse*0.8

	return vec4.T{
		(texColor[0] / 255.0) * intensity,
		(texColor[1] / 255.0) * intensity,
		(texColor[2] / 255.0) * intensity,
		1.0,
	}
}

func vertShader(vert *vec3.T, normal *vec3.T, color *vec4.T, uv *vec2.T, s *api.VertexShader) render.VertexOut {
	mvpAny, _ := s.GetUniform("mvp")
	mvp := mvpAny.(*mgl32.Mat4)

	modelAny, _ := s.GetUniform("model")
	model := modelAny.(*mgl32.Mat4)

	v := mgl32.Vec4{vert[0], vert[1], vert[2], 1.0}

	n := mgl32.Vec4{normal[0], normal[1], normal[2], 0.0}
	transformedNormal := model.Mul4x1(n)

	return render.VertexOut{
		Pos:    mvp.Mul4x1(v),
		Normal: vec3.T{transformedNormal[0], transformedNormal[1], transformedNormal[2]},
		UV:     *uv,
		Color:  *color,
	}
}

func generateCheckerboard(w, h int) []vec4.T {
	pixels := make([]vec4.T, w*h)
	size := 16
	for y := range h {
		for x := range w {
			if (x/size+y/size)%2 == 0 {
				pixels[y*w+x] = vec4.T{255, 0, 255, 1}
			} else {
				pixels[y*w+x] = vec4.T{0, 255, 0, 1}
			}
		}
	}
	return pixels
}
