package lights

import (
	"github.com/ungerik/go3d/vec3"
)

type DirectLight struct {
	Color     vec3.T
	Direction vec3.T
}

type AmbientLight struct {
	Color vec3.T
}

type PointLight struct {
	Color     vec3.T
	Position  vec3.T
	Intensity float32

	Constant  float32 // in general 1.0
	Linear    float32 // in general 0.09
	Quadratic float32 // in general 0.032
}

type SpotLight struct {
	Color     vec3.T
	Position  vec3.T
	Direction vec3.T

	Intensity float32
	Constant  float32
	Linear    float32
	Quadratic float32

	CosCutOff float32
	OuterCos  float32
}

type LightingConfig struct {
	Ambient     AmbientLight
	Directional DirectLight

	PointLights map[int]*PointLight
	SpotLights  map[int]*SpotLight
}
