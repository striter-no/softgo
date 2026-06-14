package api

import (
	"errors"

	"github.com/ungerik/go3d/vec3"
	"github.com/ungerik/go3d/vec4"
)

type FragmentShader struct {
	Params    map[string]any
	PixelFunc func(x float32, y float32, fragColor vec4.T, normal vec3.T, fragPos vec3.T, shader *FragmentShader) vec4.T
}

func NewFragShader(pxFunc func(x float32, y float32, fragColor vec4.T, normal vec3.T, fragPos vec3.T, shader *FragmentShader) vec4.T) *FragmentShader {
	return &FragmentShader{
		Params:    make(map[string]any, 0),
		PixelFunc: pxFunc,
	}
}

func (s *FragmentShader) SetUniform(key string, val any) {
	s.Params[key] = val
}

func (s *FragmentShader) GetUniform(key string) (any, error) {
	val, exists := s.Params[key]

	if exists {
		return val, nil
	}

	return nil, errors.New("Uniform's key is not found")
}
