package api

import (
	"errors"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/ungerik/go3d/vec3"
)

type VertexShader struct {
	Params  map[string]any
	Perform func(tbo *vec3.T, shader *VertexShader) mgl32.Vec4
}

func NewVertexShader(pxFunc func(tbo *vec3.T, shader *VertexShader) mgl32.Vec4) *VertexShader {
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
