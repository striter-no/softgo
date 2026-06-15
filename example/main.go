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

	anim, err := textures.ConvertGIFToAnimation("./assets/textures/rickroll.gif")
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

	tex := &render.Texture{
		Width: 312, Height: 312,
		Pixels: generateCheckerboard(312, 312),
	}

	tex2 := textures.ConvertImageToTexture("./assets/textures/onigiri.jpg")

	s.FragShader = api.NewFragShader(fragShader)
	s.VertexShader = api.NewVertexShader(vertShader)

	s.FragShader.SetUniform("tex", tex)

	s.SSAAFactor = 1

	var tick int64
	var ftick int
	mesh1, _ := assets.LoadOBJ("./assets/meshes/suzanne.obj")
	mesh2, _ := assets.LoadOBJ("./assets/meshes/cube.obj")
	mesh3, _ := assets.LoadOBJ("./assets/meshes/plane.obj")

	mainFBO := render.NewFramebuffer(s.Screen.Width*s.SSAAFactor, s.Screen.Height*s.SSAAFactor, false)
	oldW, oldH := s.Screen.Width, s.Screen.Height
	for s.IsOpen() {
		winMouse.PollEvents()
		winKeyboard.PollEvents()

		aspect := float32(s.Screen.Width) / (float32(s.Screen.Height))
		camera.UpdateOnHID(aspect)

		s.Clear()
		if oldW != s.Screen.Width || oldH != s.Screen.Height {
			mainFBO = render.NewFramebuffer(s.Screen.Width*s.SSAAFactor, s.Screen.Height*s.SSAAFactor, false)
		}

		mainFBO.Clear(s.BackColor)

		camPos := camera.Position
		s.FragShader.SetUniform("viewPos", &camPos)

		// first model --------------------
		angle := float32(tick) * 0.001

		modelMatrix1 := mgl32.Translate3D(-2.0, 0, 0).Mul4(mgl32.HomogRotate3DY(angle))
		mvp1 := camera.VP.Mul4(modelMatrix1)

		s.VertexShader.SetUniform("mvp", &mvp1)
		s.VertexShader.SetUniform("model", &modelMatrix1)
		s.FragShader.SetUniform("tex", tex)

		s.DrawCall(mesh1, mainFBO)

		// second model --------------------
		modelMatrix2 := mgl32.Translate3D(2.0, 0, 0).Mul4(mgl32.HomogRotate3DY(-angle * 2))
		mvp2 := camera.VP.Mul4(modelMatrix2)

		s.VertexShader.SetUniform("mvp", &mvp2)
		s.VertexShader.SetUniform("model", &modelMatrix2)
		s.FragShader.SetUniform("tex", tex2)

		s.DrawCall(mesh2, mainFBO)

		// third model --------------------
		modelMatrix3 := mgl32.Translate3D(.0, -2, .0).Mul4(mgl32.HomogRotate3DY(angle * 0.5)).Mul4(mgl32.Scale3D(5, 1, 5))
		mvp3 := camera.VP.Mul4(modelMatrix3)

		s.VertexShader.SetUniform("mvp", &mvp3)
		s.VertexShader.SetUniform("model", &modelMatrix3)

		frame := &anim.Frames[ftick%len(anim.Frames)]
		s.FragShader.SetUniform("tex", frame)

		if tick%8 == 0 {
			ftick++
		}

		s.DrawCall(mesh3, mainFBO)

		// drawing all meshes --------------------
		s.Present(mainFBO)

		s.Screen.SetText(0, 0, fmt.Sprintf("FPS: %.1f", s.CurrentFPS), graphics.NewFGPixel(255, 255, 255, ""))

		oldW, oldH = s.Screen.Width, s.Screen.Height
		s.Screen.Blit() // drawing to the terminal

		tick++
	}
}

func fragShader(u float32, v float32, col vec4.T, norm vec3.T, fragPos vec3.T, s *api.FragmentShader) vec4.T {
	texAny, _ := s.GetUniform("tex")
	tex := texAny.(*render.Texture)
	texColor := tex.Sample(u, v)

	lenN := float32(math.Sqrt(float64(norm[0]*norm[0] + norm[1]*norm[1] + norm[2]*norm[2])))
	if lenN > 0 {
		norm[0] /= lenN
		norm[1] /= lenN
		norm[2] /= lenN
	}

	lightPos := vec3.T{2.0, 2.0, -2.0}
	lightColor := vec3.T{1.0, 0.9, 0.7}

	lightDir := vec3.T{lightPos[0] - fragPos[0], lightPos[1] - fragPos[1], lightPos[2] - fragPos[2]}
	distance := float32(math.Sqrt(float64(lightDir[0]*lightDir[0] + lightDir[1]*lightDir[1] + lightDir[2]*lightDir[2])))

	if distance > 0 {
		lightDir[0] /= distance
		lightDir[1] /= distance
		lightDir[2] /= distance
	}

	diffuse := norm[0]*lightDir[0] + norm[1]*lightDir[1] + norm[2]*lightDir[2]
	if diffuse < 0 {
		diffuse = 0
	}

	attenuation := 1.0 / (1.0 + 0.09*distance + 0.032*(distance*distance))
	specular := float32(0.0)

	if diffuse > 0 {
		viewPosAny, _ := s.GetUniform("viewPos")
		viewPos := viewPosAny.(*vec3.T)

		viewDir := vec3.T{viewPos[0] - fragPos[0], viewPos[1] - fragPos[1], viewPos[2] - fragPos[2]}
		lenV := float32(math.Sqrt(float64(viewDir[0]*viewDir[0] + viewDir[1]*viewDir[1] + viewDir[2]*viewDir[2])))
		if lenV > 0 {
			viewDir[0] /= lenV
			viewDir[1] /= lenV
			viewDir[2] /= lenV
		}

		negL := vec3.T{-lightDir[0], -lightDir[1], -lightDir[2]}
		dotNL := negL[0]*norm[0] + negL[1]*norm[1] + negL[2]*norm[2]

		reflectDir := vec3.T{
			negL[0] - 2.0*dotNL*norm[0],
			negL[1] - 2.0*dotNL*norm[1],
			negL[2] - 2.0*dotNL*norm[2],
		}

		specDot := viewDir[0]*reflectDir[0] + viewDir[1]*reflectDir[1] + viewDir[2]*reflectDir[2]
		if specDot < 0 {
			specDot = 0
		}

		specularStrength := float32(0.5)
		shininess := float64(64.0)

		specVal := float32(math.Pow(float64(specDot), shininess))

		specular = specularStrength * specVal * attenuation
	}

	diffuse *= attenuation
	ambient := float32(0.1)
	intensity := ambient + diffuse + specular

	return vec4.T{
		(texColor[0] / 255.0) * lightColor[0] * intensity,
		(texColor[1] / 255.0) * lightColor[1] * intensity,
		(texColor[2] / 255.0) * lightColor[2] * intensity,
		1.0,
	}
}

func vertShader(vert *vec3.T, normal *vec3.T, color *vec4.T, uv *vec2.T, s *api.VertexShader) render.VertexOut {
	mvpAny, _ := s.GetUniform("mvp")
	mvp := mvpAny.(*mgl32.Mat4)

	modelAny, _ := s.GetUniform("model")
	model := modelAny.(*mgl32.Mat4)

	v := mgl32.Vec4{vert[0], vert[1], vert[2], 1.0}
	worldPos := model.Mul4x1(v)

	n := mgl32.Vec4{normal[0], normal[1], normal[2], 0.0}
	transformedNormal := model.Mul4x1(n)

	return render.VertexOut{
		Pos:     mvp.Mul4x1(v),
		Normal:  vec3.T{transformedNormal[0], transformedNormal[1], transformedNormal[2]},
		UV:      *uv,
		Color:   *color,
		FragPos: vec3.T{worldPos[0], worldPos[1], worldPos[2]},
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
