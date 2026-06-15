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

	FragShader   *FragmentShader
	VertexShader *VertexShader

	ssaaBuffer []vec3.T
	SSAAFactor int

	zBuffer    []float32
	CurrentFPS float64

	BackColor           vec3.T
	isResolutionChanged bool
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
		BackColor:    vec3.T{0.2, 0.2, 0.2},
	}

	return out, nil
}

func (s *RenderScreen) Init() {
	s.Screen.DisableEcho()
	s.Screen.EnterAlt()
	s.Screen.HideCursor()
}

func (s *RenderScreen) End() {
	s.Screen.ShowCursor()
	s.Screen.ExitAlt()
	s.Screen.EnableEcho()
}

func (s *RenderScreen) IsOpen() bool {
	return s.Screen.IsRunning()
}

func (s *RenderScreen) ChangedRes() bool {
	v := s.isResolutionChanged
	s.isResolutionChanged = false

	return v
}

func (s *RenderScreen) Clear() error {
	if err := s.Screen.ClearPixels(); err != nil {
		return err
	}

	if s.SSAAFactor < 1 {
		s.SSAAFactor = 1
	}

	fh, fw := float32(s.Screen.Height), float32(s.Screen.Width)
	ssaaWidth := int(fw) * s.SSAAFactor
	ssaaHeight := int(fh) * s.SSAAFactor
	numSsaaPixels := ssaaWidth * ssaaHeight

	if len(s.ssaaBuffer) != numSsaaPixels {
		s.ssaaBuffer = make([]vec3.T, numSsaaPixels)
		s.zBuffer = make([]float32, numSsaaPixels)
	}

	for i := range s.ssaaBuffer {
		s.ssaaBuffer[i] = s.BackColor
		s.zBuffer[i] = math.MaxFloat32
	}
	return nil
}

func (s *RenderScreen) DrawCall(mesh []render.TBO, target *render.Framebuffer) error {
	if s.FragShader == nil || s.VertexShader == nil {
		return errors.New("No fragment or vertex shader set")
	}

	checkDepth := func(x, y int, z float32) bool {
		if x < 0 || x >= target.Width || y < 0 || y >= target.Height {
			return false
		}
		idx := y*target.Width + x
		return z < target.DepthBuffer[idx]
	}

	rasterPix := func(x, y int, z float32, r, g, b uint8, u, v float32, nx, ny, nz float32, fpx, fpy, fpz float32) {
		if x < 0 || x >= target.Width || y < 0 || y >= target.Height {
			return
		}

		if z > 1.0 || z < -1.0 {
			return
		}

		idx := y*target.Width + x
		if z >= target.DepthBuffer[idx] {
			return
		}

		frag := s.FragShader.PixelFunc(
			u, v,
			vec4.T{float32(r) / 255.0, float32(g) / 255.0, float32(b) / 255.0, 1.0},
			vec3.T{nx, ny, nz},
			vec3.T{fpx, fpy, fpz},
			s.FragShader,
		)

		alpha := frag[3]
		if alpha < 0.01 {
			return
		}

		if target.HasColor {
			if alpha < 1.0 {
				oldColor := target.ColorBuffer[idx]
				frag[0] = frag[0]*alpha + oldColor[0]*(1.0-alpha)
				frag[1] = frag[1]*alpha + oldColor[1]*(1.0-alpha)
				frag[2] = frag[2]*alpha + oldColor[2]*(1.0-alpha)
			}
			target.ColorBuffer[idx] = vec3.T{frag[0], frag[1], frag[2]}
		}

		if alpha >= 1.0 || !target.HasColor {
			target.DepthBuffer[idx] = z
		}
	}

	for _, tbo := range mesh {
		vert0 := s.VertexShader.Perform(&tbo.V0, &tbo.N0, &tbo.C0, &tbo.UV0, s.VertexShader)
		vert1 := s.VertexShader.Perform(&tbo.V1, &tbo.N1, &tbo.C1, &tbo.UV1, s.VertexShader)
		vert2 := s.VertexShader.Perform(&tbo.V2, &tbo.N2, &tbo.C2, &tbo.UV2, s.VertexShader)

		clippedTris, count := render.ClipTriangle(vert0, vert1, vert2, 0.1)

		for i := range count {
			tri := clippedTris[i]
			w0 := tri[0].Pos.W()
			w1 := tri[1].Pos.W()
			w2 := tri[2].Pos.W()

			if (tri[0].Pos[0] < -w0 && tri[1].Pos[0] < -w1 && tri[2].Pos[0] < -w2) ||
				(tri[0].Pos[0] > w0 && tri[1].Pos[0] > w1 && tri[2].Pos[0] > w2) ||
				(tri[0].Pos[1] < -w0 && tri[1].Pos[1] < -w1 && tri[2].Pos[1] < -w2) ||
				(tri[0].Pos[1] > w0 && tri[1].Pos[1] > w1 && tri[2].Pos[1] > w2) ||
				(tri[0].Pos[2] > w0 && tri[1].Pos[2] > w1 && tri[2].Pos[2] > w2) {
				continue
			}

			ndc0 := vec2.T{tri[0].Pos[0] / tri[0].Pos.W(), tri[0].Pos[1] / tri[0].Pos.W()}
			ndc1 := vec2.T{tri[1].Pos[0] / tri[1].Pos.W(), tri[1].Pos[1] / tri[1].Pos.W()}
			ndc2 := vec2.T{tri[2].Pos[0] / tri[2].Pos.W(), tri[2].Pos[1] / tri[2].Pos.W()}

			ndcZ0 := tri[0].Pos[2] / tri[0].Pos.W()
			ndcZ1 := tri[1].Pos[2] / tri[1].Pos.W()
			ndcZ2 := tri[2].Pos[2] / tri[2].Pos.W()

			screenV0 := vec2.T{(ndc0[0] + 1.0) * 0.5 * float32(target.Width), (1.0 - ndc0[1]) * 0.5 * float32(target.Height)}
			screenV1 := vec2.T{(ndc1[0] + 1.0) * 0.5 * float32(target.Width), (1.0 - ndc1[1]) * 0.5 * float32(target.Height)}
			screenV2 := vec2.T{(ndc2[0] + 1.0) * 0.5 * float32(target.Width), (1.0 - ndc2[1]) * 0.5 * float32(target.Height)}

			render.RasterizeTriangle(
				tbo.OmniDir,
				screenV0, screenV1, screenV2,
				ndcZ0, ndcZ1, ndcZ2,
				tri[0].Pos.W(), tri[1].Pos.W(), tri[2].Pos.W(),
				tri[0].Color.Vec3(), tri[1].Color.Vec3(), tri[2].Color.Vec3(),
				tri[0].UV, tri[1].UV, tri[2].UV,
				tri[0].Normal, tri[1].Normal, tri[2].Normal,
				tri[0].FragPos, tri[1].FragPos, tri[2].FragPos,
				target.Width, target.Height,
				checkDepth,
				rasterPix,
			)
		}
	}

	return nil
}

func (s *RenderScreen) Present(mainFBO *render.Framebuffer) {
	fh, fw := float32(s.Screen.Height), float32(s.Screen.Width)
	invSamples := 1.0 / float32(s.SSAAFactor*s.SSAAFactor)

	for y := 0; y < int(fh); y++ {
		for x := 0; x < int(fw); x++ {
			if s.SSAAFactor == 1 {
				c := mainFBO.ColorBuffer[y*mainFBO.Width+x]
				s.Screen.SetPixel(x, y, graphics.NewBGPixel(uint(c[0]*255), uint(c[1]*255), uint(c[2]*255), ""))
			} else {
				var avgR, avgG, avgB float32
				for sy := 0; sy < s.SSAAFactor; sy++ {
					for sx := 0; sx < s.SSAAFactor; sx++ {
						c := mainFBO.ColorBuffer[(y*s.SSAAFactor+sy)*mainFBO.Width+(x*s.SSAAFactor+sx)]
						avgR += c[0]
						avgG += c[1]
						avgB += c[2]
					}
				}
				s.Screen.SetPixel(x, y, graphics.NewBGPixel(uint(avgR*invSamples*255), uint(avgG*invSamples*255), uint(avgB*invSamples*255), ""))
			}
		}
	}

	s.CurrentFPS = s.fpsCounter.Tick()
}
