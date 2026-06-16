package api

import (
	"errors"

	"github.com/striter-no/softgo/render"
	"github.com/ungerik/go3d/vec2"
	"github.com/ungerik/go3d/vec3"
	"github.com/ungerik/go3d/vec4"
)

type VertexShader struct {
	Params  map[string]any
	Perform func(pos vec3.T, norm vec3.T, color vec4.T, uv vec2.T, shader *VertexShader) render.VertexOut
}

func NewVertexShader(pxFunc func(pos vec3.T, norm vec3.T, color vec4.T, uv vec2.T, shader *VertexShader) render.VertexOut) *VertexShader {
	return &VertexShader{
		Params:  make(map[string]any, 0),
		Perform: pxFunc,
	}
}

func (s *VertexShader) SetUniform(key string, val any) {
	s.Params[key] = val
}

func (s *VertexShader) GetUniform(key string) (any, error) {
	val, exists := s.Params[key]

	if exists {
		return val, nil
	}

	return nil, errors.New("Uniform's key is not found")
}
