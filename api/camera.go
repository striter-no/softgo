package api

import (
	"math"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/striter-no/softgo/api/keyboard"
	"github.com/striter-no/softgo/api/mouse"
	"github.com/ungerik/go3d/vec3"
)

type Camera struct {
	VP         mgl32.Mat4
	Projection mgl32.Mat4

	Position vec3.T
	Rotation vec3.T

	Locked    bool
	Sensivity float32
	Speed     float32

	Near float32
	Far  float32
	FOV  float32

	mouse    mouse.WindowMouse
	keyboard keyboard.WindowKeyboard
}

func NewCamera(position vec3.T, sensivity float32, speed float32, mouse mouse.WindowMouse, keyboard keyboard.WindowKeyboard, near, far, fov float32) *Camera {
	return &Camera{
		Position:   position,
		Rotation:   vec3.T{0, 0, 0},
		Sensivity:  sensivity,
		Speed:      speed,
		mouse:      mouse,
		keyboard:   keyboard,
		Locked:     false,
		Near:       near,
		Far:        far,
		FOV:        fov,
		Projection: mgl32.Ident4(),
	}
}

func (c *Camera) UpdateOnHID(aspect float32, movement bool) {
	if !c.Locked {
		return
	}

	w, h := c.mouse.GetSize()
	centerX, centerY := w/2, h/2
	posX, posY := c.mouse.GetPosition()

	dx := float32(posX - centerX)
	dy := float32(posY - centerY)

	c.mouse.MoveMouse(centerX, centerY)

	c.Rotation[1] += dx * c.Sensivity
	c.Rotation[0] -= dy * c.Sensivity

	if c.Rotation[0] > 89.0 {
		c.Rotation[0] = 89.0
	}
	if c.Rotation[0] < -89.0 {
		c.Rotation[0] = -89.0
	}

	yawRad := float64(mgl32.DegToRad(c.Rotation[1]))
	pitchRad := float64(mgl32.DegToRad(c.Rotation[0]))

	forward := mgl32.Vec3{
		float32(math.Sin(yawRad) * math.Cos(pitchRad)),
		float32(math.Sin(pitchRad)),
		float32(-math.Cos(yawRad) * math.Cos(pitchRad)),
	}.Normalize()

	worldUp := mgl32.Vec3{0, 1, 0}
	right := forward.Cross(worldUp).Normalize()

	if movement {
		if c.keyboard.IsKeyPressed(keyboard.KeyW) { // forward
			c.Position[0] += forward.X() * c.Speed
			c.Position[1] += forward.Y() * c.Speed
			c.Position[2] += forward.Z() * c.Speed
		} else if c.keyboard.IsKeyPressed(keyboard.KeyS) { // backwards
			c.Position[0] -= forward.X() * c.Speed
			c.Position[1] -= forward.Y() * c.Speed
			c.Position[2] -= forward.Z() * c.Speed
		}

		if c.keyboard.IsKeyPressed(keyboard.KeyA) { // left
			c.Position[0] -= right.X() * c.Speed
			c.Position[1] -= right.Y() * c.Speed
			c.Position[2] -= right.Z() * c.Speed
		} else if c.keyboard.IsKeyPressed(keyboard.KeyD) { // right
			c.Position[0] += right.X() * c.Speed
			c.Position[1] += right.Y() * c.Speed
			c.Position[2] += right.Z() * c.Speed
		}

		if c.keyboard.IsKeyPressed(keyboard.KeyE) { // up
			c.Position[1] += c.Speed
		} else if c.keyboard.IsKeyPressed(keyboard.KeyQ) { // down
			c.Position[1] -= c.Speed
		}

	}

	proj := mgl32.Perspective(mgl32.DegToRad(c.FOV), aspect, c.Near, c.Far)
	c.Projection = proj

	mglPos := mgl32.Vec3{c.Position[0], c.Position[1], c.Position[2]}
	target := mglPos.Add(forward)

	view := mgl32.LookAtV(mglPos, target, worldUp)

	c.VP = proj.Mul4(view)
}
