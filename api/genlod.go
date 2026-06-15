package api

import (
	"math"

	"github.com/striter-no/softgo/render"
	"github.com/ungerik/go3d/vec3"
)

func GenerateLOD(original []render.TBO, gridSize float32) []render.TBO {
	var simplified []render.TBO

	snap := func(v vec3.T) vec3.T {
		return vec3.T{
			float32(math.Round(float64(v[0]/gridSize))) * gridSize,
			float32(math.Round(float64(v[1]/gridSize))) * gridSize,
			float32(math.Round(float64(v[2]/gridSize))) * gridSize,
		}
	}

	for _, tri := range original {
		t := tri

		t.V0 = snap(t.V0)
		t.V1 = snap(t.V1)
		t.V2 = snap(t.V2)

		if (t.V0[0] == t.V1[0] && t.V0[1] == t.V1[1] && t.V0[2] == t.V1[2]) ||
			(t.V1[0] == t.V2[0] && t.V1[1] == t.V2[1] && t.V1[2] == t.V2[2]) ||
			(t.V0[0] == t.V2[0] && t.V0[1] == t.V2[1] && t.V0[2] == t.V2[2]) {
			continue
		}

		simplified = append(simplified, t)
	}

	return simplified
}
