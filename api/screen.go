package api

import (
	"context"
	"errors"
	"time"

	"github.com/striter-no/softgo/render"
	"github.com/striter-no/stg/graphics"
	"github.com/ungerik/go3d/vec2"
	"github.com/ungerik/go3d/vec3"
	"github.com/ungerik/go3d/vec4"
)

type RenderScreen struct {
	ctx context.Context

	Screen     *graphics.Screen
	fpsCounter *FPSCounter

	tbos         []render.TBO
	FragShader   *FragmentShader
	VertexShader *VertexShader

	ssaaBuffer []vec3.T
	CurrentFPS float64
}

func NewRenderScreen(ctx context.Context) (*RenderScreen, error) {
	sc, err := graphics.NewScreen(graphics.NewBGPixel(10, 10, 10, " "), ctx)
	if err != nil {
		return nil, err
	}

	out := &RenderScreen{
		ctx:          ctx,
		Screen:       sc,
		FragShader:   nil,
		VertexShader: nil,
		fpsCounter:   NewFPSCounter(time.Second),
	}

	return out, nil
}

func (s *RenderScreen) Init() {
	s.Screen.EnterAlt()
	s.Screen.HideCursor()
}

func (s *RenderScreen) End() {
	s.Screen.ShowCursor()
	s.Screen.ExitAlt()
}

func (s *RenderScreen) FeedTBO(tbo render.TBO) {
	s.tbos = append(s.tbos, tbo)
}

func (s *RenderScreen) IsOpen() bool {
	return s.Screen.IsRunning()
}

func (s *RenderScreen) Blit() error {
	if s.FragShader == nil || s.VertexShader == nil {
		return errors.New("No fragment or vertex shader set")
	}

	if err := s.Screen.ClearPixels(); err != nil {
		return err
	}

	fh, fw := float32(s.Screen.Height), float32(s.Screen.Width)

	const ssaaFactor = 2 // 2x SSAA
	ssaaWidth := int(fw) * ssaaFactor
	ssaaHeight := int(fh) * ssaaFactor
	numSsaaPixels := ssaaWidth * ssaaHeight

	if len(s.ssaaBuffer) != numSsaaPixels {
		s.ssaaBuffer = make([]vec3.T, numSsaaPixels)
	}
	for i := range s.ssaaBuffer {
		s.ssaaBuffer[i] = vec3.T{0, 0, 0}
	}

	rasterPix := func(x, y int, r, g, b uint8) {
		if x < 0 || x >= ssaaWidth || y < 0 || y >= ssaaHeight {
			return
		}

		frag := s.FragShader.PixelFunc(
			float32(x)/float32(ssaaWidth),
			float32(y)/float32(ssaaHeight),
			vec4.T{float32(r) / 255.0, float32(g) / 255.0, float32(b) / 255.0, 1.0},
			s.FragShader,
		)

		s.ssaaBuffer[y*ssaaWidth+x] = vec3.T{frag[0], frag[1], frag[2]}
	}

	for _, tbo := range s.tbos {
		vert0 := render.VertexOut{Pos: s.VertexShader.Perform(&tbo.V0, s.VertexShader), Color: tbo.C0}
		vert1 := render.VertexOut{Pos: s.VertexShader.Perform(&tbo.V1, s.VertexShader), Color: tbo.C1}
		vert2 := render.VertexOut{Pos: s.VertexShader.Perform(&tbo.V2, s.VertexShader), Color: tbo.C2}

		clippedTris := render.ClipTriangle(vert0, vert1, vert2, 0.1)

		for _, tri := range clippedTris {
			ndc0 := vec2.T{tri[0].Pos[0] / tri[0].Pos.W(), tri[0].Pos[1] / tri[0].Pos.W()}
			ndc1 := vec2.T{tri[1].Pos[0] / tri[1].Pos.W(), tri[1].Pos[1] / tri[1].Pos.W()}
			ndc2 := vec2.T{tri[2].Pos[0] / tri[2].Pos.W(), tri[2].Pos[1] / tri[2].Pos.W()}

			screenV0 := vec2.T{(ndc0[0] + 1.0) * 0.5 * float32(ssaaWidth), (1.0 - ndc0[1]) * 0.5 * float32(ssaaHeight)}
			screenV1 := vec2.T{(ndc1[0] + 1.0) * 0.5 * float32(ssaaWidth), (1.0 - ndc1[1]) * 0.5 * float32(ssaaHeight)}
			screenV2 := vec2.T{(ndc2[0] + 1.0) * 0.5 * float32(ssaaWidth), (1.0 - ndc2[1]) * 0.5 * float32(ssaaHeight)}

			render.RasterizeTriangle(
				screenV0, screenV1, screenV2,
				tri[0].Color.Vec3(), tri[1].Color.Vec3(), tri[2].Color.Vec3(),
				rasterPix,
			)
		}
	}

	for y := 0; y < int(fh); y++ {
		for x := 0; x < int(fw); x++ {

			idx00 := (y*ssaaFactor)*ssaaWidth + (x * ssaaFactor)
			idx10 := (y*ssaaFactor)*ssaaWidth + (x*ssaaFactor + 1)
			idx01 := (y*ssaaFactor+1)*ssaaWidth + (x * ssaaFactor)
			idx11 := (y*ssaaFactor+1)*ssaaWidth + (x*ssaaFactor + 1)

			c00 := s.ssaaBuffer[idx00]
			c10 := s.ssaaBuffer[idx10]
			c01 := s.ssaaBuffer[idx01]
			c11 := s.ssaaBuffer[idx11]

			avgR := (c00[0] + c10[0] + c01[0] + c11[0]) * 0.25
			avgG := (c00[1] + c10[1] + c01[1] + c11[1]) * 0.25
			avgB := (c00[2] + c10[2] + c01[2] + c11[2]) * 0.25

			s.Screen.SetPixel(x, y, graphics.NewBGPixel(
				uint(avgR*255),
				uint(avgG*255),
				uint(avgB*255), "",
			))
		}
	}

	s.CurrentFPS = s.fpsCounter.Tick()
	s.tbos = nil
	return nil
}
