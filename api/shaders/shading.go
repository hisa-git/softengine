package shaders

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/striter-no/softengine/lights"
	"github.com/striter-no/softgo/render"
	"github.com/ungerik/go3d/vec3"
	"github.com/ungerik/go3d/vec4"
)

type ShaderContext struct {
	MVP     mgl32.Mat4
	Model   mgl32.Mat4
	Texture *render.Texture
	Color   vec4.T
	ViewPos vec3.T

	Lights lights.LightingConfig
}
