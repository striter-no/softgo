package assets

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/striter-no/softgo/render"
	"github.com/ungerik/go3d/vec2"
	"github.com/ungerik/go3d/vec3"
	"github.com/ungerik/go3d/vec4"
)

func LoadOBJ(filepath string) ([]render.TBO, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("не удалось открыть файл %s: %w", filepath, err)
	}
	defer file.Close()

	var vertices []vec3.T
	var uvs []vec2.T
	var normals []vec3.T
	var triangles []render.TBO

	scanner := bufio.NewScanner(file)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) == 0 {
			continue
		}

		switch parts[0] {
		case "v":
			if len(parts) >= 4 {
				x, _ := strconv.ParseFloat(parts[1], 32)
				y, _ := strconv.ParseFloat(parts[2], 32)
				z, _ := strconv.ParseFloat(parts[3], 32)
				vertices = append(vertices, vec3.T{float32(x), float32(y), float32(z)})
			}
		case "vt":
			if len(parts) >= 3 {
				u, _ := strconv.ParseFloat(parts[1], 32)
				v, _ := strconv.ParseFloat(parts[2], 32)
				uvs = append(uvs, vec2.T{float32(u), float32(v)})
			}
		case "vn":
			if len(parts) >= 4 {
				nx, _ := strconv.ParseFloat(parts[1], 32)
				ny, _ := strconv.ParseFloat(parts[2], 32)
				nz, _ := strconv.ParseFloat(parts[3], 32)
				normals = append(normals, vec3.T{float32(nx), float32(ny), float32(nz)})
			}
		case "f":
			if len(parts) >= 4 {
				v1, vt1, vn1 := parseFaceVertex(parts[1])
				v2, vt2, vn2 := parseFaceVertex(parts[2])
				v3, vt3, vn3 := parseFaceVertex(parts[3])

				var uv0, uv1, uv2 vec2.T
				if vt1 > 0 && vt1 <= len(uvs) {
					uv0 = uvs[vt1-1]
				}
				if vt2 > 0 && vt2 <= len(uvs) {
					uv1 = uvs[vt2-1]
				}
				if vt3 > 0 && vt3 <= len(uvs) {
					uv2 = uvs[vt3-1]
				}

				var n0, n1, n2 vec3.T
				if vn1 > 0 && vn1 <= len(normals) {
					n0 = normals[vn1-1]
				}
				if vn2 > 0 && vn2 <= len(normals) {
					n1 = normals[vn2-1]
				}
				if vn3 > 0 && vn3 <= len(normals) {
					n2 = normals[vn3-1]
				}

				tri := render.TBO{
					V0: vertices[v1-1], V1: vertices[v2-1], V2: vertices[v3-1],
					UV0: uv0, UV1: uv1, UV2: uv2,
					N0: n0, N1: n1, N2: n2,

					C0: vec4.T{255, 255, 255, 1},
					C1: vec4.T{255, 255, 255, 1},
					C2: vec4.T{255, 255, 255, 1},
				}
				triangles = append(triangles, tri)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read from file %s: %w", filepath, err)
	}

	return triangles, nil
}

func parseFaceVertex(fv string) (vIdx, vtIdx, vnIdx int) {
	parts := strings.Split(fv, "/")

	vIdx, _ = strconv.Atoi(parts[0])

	if len(parts) > 1 && parts[1] != "" {
		vtIdx, _ = strconv.Atoi(parts[1])
	} else {
		vtIdx = 0
	}

	if len(parts) > 2 && parts[2] != "" {
		vnIdx, _ = strconv.Atoi(parts[2])
	} else {
		vnIdx = 0
	}

	return vIdx, vtIdx, vnIdx
}
