package api

import (
	"context"
	"errors"
	"math"
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
	SSAAFactor int

	zBuffer    []float32
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

	if s.SSAAFactor < 1 {
		s.SSAAFactor = 1
	}

	ssaaWidth := int(fw) * s.SSAAFactor
	ssaaHeight := int(fh) * s.SSAAFactor
	numSsaaPixels := ssaaWidth * ssaaHeight

	if len(s.ssaaBuffer) != numSsaaPixels {
		s.ssaaBuffer = make([]vec3.T, numSsaaPixels)
		s.zBuffer = make([]float32, numSsaaPixels)
	}
	for i := range s.ssaaBuffer {
		s.ssaaBuffer[i] = vec3.T{0.8, 0.8, 0.8}
		s.zBuffer[i] = math.MaxFloat32
	}

	checkDepth := func(x, y int, z float32) bool {
		if x < 0 || x >= ssaaWidth || y < 0 || y >= ssaaHeight {
			return false
		}
		idx := y*ssaaWidth + x
		return z < s.zBuffer[idx]
	}

	rasterPix := func(x, y int, z float32, r, g, b uint8, u, v float32, nx, ny, nz float32) {
		if x < 0 || x >= ssaaWidth || y < 0 || y >= ssaaHeight {
			return
		}

		idx := y*ssaaWidth + x
		if z >= s.zBuffer[idx] {
			return
		}

		frag := s.FragShader.PixelFunc(
			u, v,
			vec4.T{float32(r) / 255.0, float32(g) / 255.0, float32(b) / 255.0, 1.0},
			vec3.T{nx, ny, nz},
			s.FragShader,
		)

		s.zBuffer[idx] = z
		s.ssaaBuffer[y*ssaaWidth+x] = vec3.T{frag[0], frag[1], frag[2]}
	}

	for _, tbo := range s.tbos {
		vert0 := s.VertexShader.Perform(&tbo.V0, &tbo.N0, &tbo.C0, &tbo.UV0, s.VertexShader)
		vert1 := s.VertexShader.Perform(&tbo.V1, &tbo.N1, &tbo.C1, &tbo.UV1, s.VertexShader)
		vert2 := s.VertexShader.Perform(&tbo.V2, &tbo.N2, &tbo.C2, &tbo.UV2, s.VertexShader)

		w0, w1, w2 := vert0.Pos.W(), vert1.Pos.W(), vert2.Pos.W()

		if (vert0.Pos[0] < -w0 && vert1.Pos[0] < -w1 && vert2.Pos[0] < -w2) ||
			(vert0.Pos[0] > w0 && vert1.Pos[0] > w1 && vert2.Pos[0] > w2) ||
			(vert0.Pos[1] < -w0 && vert1.Pos[1] < -w1 && vert2.Pos[1] < -w2) ||
			(vert0.Pos[1] > w0 && vert1.Pos[1] > w1 && vert2.Pos[1] > w2) ||
			(vert0.Pos[2] < -w0 && vert1.Pos[2] < -w1 && vert2.Pos[2] < -w2) ||
			(vert0.Pos[2] > w0 && vert1.Pos[2] > w1 && vert2.Pos[2] > w2) {
			continue
		}

		clippedTris := render.ClipTriangle(vert0, vert1, vert2, 0.1)

		for _, tri := range clippedTris {
			ndc0 := vec2.T{tri[0].Pos[0] / tri[0].Pos.W(), tri[0].Pos[1] / tri[0].Pos.W()}
			ndc1 := vec2.T{tri[1].Pos[0] / tri[1].Pos.W(), tri[1].Pos[1] / tri[1].Pos.W()}
			ndc2 := vec2.T{tri[2].Pos[0] / tri[2].Pos.W(), tri[2].Pos[1] / tri[2].Pos.W()}

			ndcZ0 := tri[0].Pos[2] / tri[0].Pos.W()
			ndcZ1 := tri[1].Pos[2] / tri[1].Pos.W()
			ndcZ2 := tri[2].Pos[2] / tri[2].Pos.W()

			screenV0 := vec2.T{(ndc0[0] + 1.0) * 0.5 * float32(ssaaWidth), (1.0 - ndc0[1]) * 0.5 * float32(ssaaHeight)}
			screenV1 := vec2.T{(ndc1[0] + 1.0) * 0.5 * float32(ssaaWidth), (1.0 - ndc1[1]) * 0.5 * float32(ssaaHeight)}
			screenV2 := vec2.T{(ndc2[0] + 1.0) * 0.5 * float32(ssaaWidth), (1.0 - ndc2[1]) * 0.5 * float32(ssaaHeight)}

			render.RasterizeTriangle(
				screenV0, screenV1, screenV2,
				ndcZ0, ndcZ1, ndcZ2,
				tri[0].Pos.W(), tri[1].Pos.W(), tri[2].Pos.W(),
				tri[0].Color.Vec3(), tri[1].Color.Vec3(), tri[2].Color.Vec3(),
				tri[0].UV, tri[1].UV, tri[2].UV,
				tri[0].Normal, tri[1].Normal, tri[2].Normal,
				checkDepth,
				rasterPix,
			)
		}
	}

	invSamples := 1.0 / float32(s.SSAAFactor*s.SSAAFactor)
	for y := 0; y < int(fh); y++ {
		for x := 0; x < int(fw); x++ {

			if s.SSAAFactor == 1 {
				idx := y*ssaaWidth + x
				c := s.ssaaBuffer[idx]
				s.Screen.SetPixel(x, y, graphics.NewBGPixel(
					uint(c[0]*255), uint(c[1]*255), uint(c[2]*255), "",
				))
			} else {
				var avgR, avgG, avgB float32

				for sy := range s.SSAAFactor {
					for sx := range s.SSAAFactor {
						idx := (y*s.SSAAFactor+sy)*ssaaWidth + (x*s.SSAAFactor + sx)
						c := s.ssaaBuffer[idx]
						avgR += c[0]
						avgG += c[1]
						avgB += c[2]
					}
				}

				avgR *= invSamples
				avgG *= invSamples
				avgB *= invSamples

				s.Screen.SetPixel(x, y, graphics.NewBGPixel(
					uint(avgR*255), uint(avgG*255), uint(avgB*255), "",
				))
			}
		}
	}

	s.CurrentFPS = s.fpsCounter.Tick()
	s.tbos = nil
	return nil
}
