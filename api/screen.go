package api

import (
	"context"
	"errors"
	"math"
	"runtime"
	"sync"
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

	numWorkers := runtime.NumCPU()
	chunkSize := len(s.ssaaBuffer) / numWorkers

	var wg sync.WaitGroup
	for i := range numWorkers {
		start := i * chunkSize
		end := start + chunkSize
		if i == numWorkers-1 {
			end = len(s.ssaaBuffer)
		}

		wg.Add(1)
		go func(start, end int) {
			defer wg.Done()
			for j := start; j < end; j++ {
				s.ssaaBuffer[j] = s.BackColor
				s.zBuffer[j] = math.MaxFloat32
			}
		}(start, end)
	}
	return nil
}

func edgeFunction(a, b, p vec2.T) float64 {
	return (float64(b[0])-float64(a[0]))*(float64(p[1])-float64(a[1])) - (float64(b[1])-float64(a[1]))*(float64(p[0])-float64(a[0]))
}

func (s *RenderScreen) fastRasterize(
	omniDir bool,
	v0, v1, v2 vec2.T,
	z0, z1, z2 float32,
	w0, w1, w2 float32,
	c0, c1, c2 vec3.T,
	uv0, uv1, uv2 vec2.T,
	n0, n1, n2 vec3.T,
	fp0, fp1, fp2 vec3.T,
	target *render.Framebuffer,
) {
	screenWidth := target.Width
	screenHeight := target.Height

	minX := int(max(0.0, min(v0[0], v1[0], v2[0])))
	minY := int(max(0.0, min(v0[1], v1[1], v2[1])))

	maxX := int(min(float32(screenWidth-1), max(v0[0], v1[0], v2[0])+0.5))
	maxY := int(min(float32(screenHeight-1), max(v0[1], v1[1], v2[1])+0.5))

	area := edgeFunction(v0, v1, v2)

	if !omniDir && area >= 0 {
		return
	}

	invArea := 1.0 / area
	invW0 := 1.0 / w0
	invW1 := 1.0 / w1
	invW2 := 1.0 / w2

	stepX0 := float64(-(v2[1] - v1[1]))
	stepY0 := float64(v2[0] - v1[0])

	stepX1 := float64(-(v0[1] - v2[1]))
	stepY1 := float64(v0[0] - v2[0])

	stepX2 := float64(-(v1[1] - v0[1]))
	stepY2 := float64(v1[0] - v0[0])

	pStart := vec2.T{float32(minX) + 0.5, float32(minY) + 0.5}
	rowBc0 := edgeFunction(v1, v2, pStart)
	rowBc1 := edgeFunction(v2, v0, pStart)
	rowBc2 := edgeFunction(v0, v1, pStart)

	for y := minY; y <= maxY; y++ {
		bc0 := rowBc0
		bc1 := rowBc1
		bc2 := rowBc2

		// Оптимизация: вычисляем базовый индекс строки один раз,
		// чтобы избежать умножения y * screenWidth на каждом пикселе
		rowIdx := y * screenWidth

		for x := minX; x <= maxX; x++ {
			if bc0 <= 0 && bc1 <= 0 && bc2 <= 0 {
				normBc0 := float32(bc0 * invArea)
				normBc1 := float32(bc1 * invArea)
				normBc2 := float32(bc2 * invArea)

				interpZ := z0*normBc0 + z1*normBc1 + z2*normBc2

				// ИНЛАЙН: Объединенная проверка глубины и границ
				// Границы (x, y) уже гарантированы циклами minX..maxX и minY..maxY
				idx := rowIdx + x

				if interpZ < target.DepthBuffer[idx] && interpZ <= 1.0 && interpZ >= -1.0 {

					zFunc := normBc0*invW0 + normBc1*invW1 + normBc2*invW2
					invZFunc := 1.0 / zFunc

					interpU := (uv0[0]*invW0*normBc0 + uv1[0]*invW1*normBc1 + uv2[0]*invW2*normBc2) * invZFunc
					interpV := (uv0[1]*invW0*normBc0 + uv1[1]*invW1*normBc1 + uv2[1]*invW2*normBc2) * invZFunc

					interpNx := (n0[0]*invW0*normBc0 + n1[0]*invW1*normBc1 + n2[0]*invW2*normBc2) * invZFunc
					interpNy := (n0[1]*invW0*normBc0 + n1[1]*invW1*normBc1 + n2[1]*invW2*normBc2) * invZFunc
					interpNz := (n0[2]*invW0*normBc0 + n1[2]*invW1*normBc1 + n2[2]*invW2*normBc2) * invZFunc

					interpFpX := float32(float32(fp0[0])*invW0*normBc0+float32(fp1[0])*invW1*normBc1+float32(fp2[0])*invW2*normBc2) * invZFunc
					interpFpY := float32(float32(fp0[1])*invW0*normBc0+float32(fp1[1])*invW1*normBc1+float32(fp2[1])*invW2*normBc2) * invZFunc
					interpFpZ := float32(float32(fp0[2])*invW0*normBc0+float32(fp1[2])*invW1*normBc1+float32(fp2[2])*invW2*normBc2) * invZFunc

					r := c0[0]*normBc0 + c1[0]*normBc1 + c2[0]*normBc2
					g := c0[1]*normBc0 + c1[1]*normBc1 + c2[1]*normBc2
					b := c0[2]*normBc0 + c1[2]*normBc1 + c2[2]*normBc2

					// ИНЛАЙН: Ручное ограничение диапазона быстрее вызова math.Min/Max для float32
					if r < 0 {
						r = 0
					} else if r > 255 {
						r = 255
					}
					if g < 0 {
						g = 0
					} else if g > 255 {
						g = 255
					}
					if b < 0 {
						b = 0
					} else if b > 255 {
						b = 255
					}

					// ИНЛАЙН: Вызов фрагментного шейдера и запись в буфер
					frag := s.FragShader.PixelFunc(
						interpU, interpV,
						vec4.T{r / 255.0, g / 255.0, b / 255.0, 1.0},
						vec3.T{interpNx, interpNy, interpNz},
						vec4.T{interpFpX, interpFpY, interpFpZ, interpZ},
						s.FragShader,
					)

					alpha := frag[3]
					if alpha >= 0.01 {
						if target.HasColor {
							if alpha < 1.0 {
								oldColor := target.ColorBuffer[idx]
								invAlpha := 1.0 - alpha // Оптимизация: вынесли за скобки
								frag[0] = frag[0]*alpha + oldColor[0]*invAlpha
								frag[1] = frag[1]*alpha + oldColor[1]*invAlpha
								frag[2] = frag[2]*alpha + oldColor[2]*invAlpha
							}
							target.ColorBuffer[idx] = vec3.T{frag[0], frag[1], frag[2]}
						}

						if alpha >= 1.0 || !target.HasColor {
							target.DepthBuffer[idx] = interpZ
						}
					}
				}
			}

			bc0 += stepX0
			bc1 += stepX1
			bc2 += stepX2
		}

		rowBc0 += stepY0
		rowBc1 += stepY1
		rowBc2 += stepY2
	}
}

func (s *RenderScreen) DrawCall(mesh []render.TBO, target *render.Framebuffer) error {
	if s.FragShader == nil || s.VertexShader == nil {
		return errors.New("No fragment or vertex shader set")
	}

	for _, tbo := range mesh {
		vert0 := s.VertexShader.Perform(tbo.V0, tbo.N0, tbo.C0, tbo.UV0, s.VertexShader)
		vert1 := s.VertexShader.Perform(tbo.V1, tbo.N1, tbo.C1, tbo.UV1, s.VertexShader)
		vert2 := s.VertexShader.Perform(tbo.V2, tbo.N2, tbo.C2, tbo.UV2, s.VertexShader)

		w0, w1, w2 := vert0.Pos.W(), vert1.Pos.W(), vert2.Pos.W()

		// Frustum culling на уровне треугольника
		if vert0.Pos[2] > w0 && vert1.Pos[2] > w1 && vert2.Pos[2] > w2 {
			continue
		}
		if vert0.Pos[2] < -w0 && vert1.Pos[2] < -w1 && vert2.Pos[2] < -w2 {
			continue
		}
		if vert0.Pos[0] < -w0 && vert1.Pos[0] < -w1 && vert2.Pos[0] < -w2 {
			continue
		}
		if vert0.Pos[0] > w0 && vert1.Pos[0] > w1 && vert2.Pos[0] > w2 {
			continue
		}

		clippedTris, count := render.ClipTriangle(vert0, vert1, vert2, 0.1)

		for i := 0; i < count; i++ {
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

			// Передаем target напрямую
			s.fastRasterize(
				tbo.OmniDir,
				screenV0, screenV1, screenV2,
				ndcZ0, ndcZ1, ndcZ2,
				tri[0].Pos.W(), tri[1].Pos.W(), tri[2].Pos.W(),
				tri[0].Color.Vec3(), tri[1].Color.Vec3(), tri[2].Color.Vec3(),
				tri[0].UV, tri[1].UV, tri[2].UV,
				tri[0].Normal, tri[1].Normal, tri[2].Normal,
				tri[0].FragPos, tri[1].FragPos, tri[2].FragPos,
				target,
			)
		}
	}

	return nil
}

func (s *RenderScreen) Present(mainFBO *render.Framebuffer) {
	fh, fw := float32(s.Screen.Height), float32(s.Screen.Width)
	invSamples := 1.0 / float32(s.SSAAFactor*s.SSAAFactor)

	fastClamp := func(mn, mx, v float32) float32 {
		return max(min(v, mx), mn)
	}

	numWorkers := runtime.NumCPU()
	chunkSize := int(fh) / numWorkers
	if chunkSize == 0 {
		chunkSize = 1
	}

	// var wg sync.WaitGroup
	// for i := 0; i < numWorkers; i++ {
	// 	startY := i * chunkSize
	// 	endY := startY + chunkSize
	// 	if i == numWorkers-1 {
	// 		endY = int(fh) // Последний воркер забирает остаток
	// 	}

	// 	wg.Add(1)
	// 	go func(startY, endY int) {
	// 		defer wg.Done()
	// 		for y := startY; y < endY; y++ {
	for y := 0; y < int(fh); y++ {
		for x := 0; x < int(fw); x++ {
			if s.SSAAFactor == 1 {
				c := mainFBO.ColorBuffer[y*mainFBO.Width+x]
				s.Screen.SetPixel(x, y, graphics.NewBGPixel(
					uint(fastClamp(0, 255, c[0]*255)),
					uint(fastClamp(0, 255, c[1]*255)),
					uint(fastClamp(0, 255, c[2]*255)), ""))
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
				s.Screen.SetPixel(x, y, graphics.NewBGPixel(
					uint(fastClamp(0, 255, avgR*invSamples*255)),
					uint(fastClamp(0, 255, avgG*invSamples*255)),
					uint(fastClamp(0, 255, avgB*invSamples*255)),
					""))
			}
		}
	}
	// 	}(startY, endY)
	// }

	// wg.Wait()
	s.CurrentFPS = s.fpsCounter.Tick()
}
