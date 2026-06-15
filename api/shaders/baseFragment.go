package shaders

import (
	"math"

	"github.com/striter-no/softgo/api"
	"github.com/ungerik/go3d/vec3"
	"github.com/ungerik/go3d/vec4"
)

func calculateLightContribution(
	lightDir vec3.T, // normalized
	distance float32,
	lightColor vec3.T,
	intensity float32,
	constant, linear, quadratic float32,
	norm vec3.T, // normalized
	viewDir vec3.T, // normalized
	shininess float32,
) vec3.T {
	// 1. Diffuse
	diffuse := norm[0]*lightDir[0] + norm[1]*lightDir[1] + norm[2]*lightDir[2]
	if diffuse < 0 {
		diffuse = 0
	}

	// 2. Specular
	specular := float32(0.0)
	if diffuse > 0 {
		negL := vec3.T{-lightDir[0], -lightDir[1], -lightDir[2]}
		dotNL := negL[0]*norm[0] + negL[1]*norm[1] + negL[2]*norm[2]

		reflectDir := vec3.T{
			negL[0] - 2.0*dotNL*norm[0],
			negL[1] - 2.0*dotNL*norm[1],
			negL[2] - 2.0*dotNL*norm[2],
		}

		specDot := viewDir[0]*reflectDir[0] + viewDir[1]*reflectDir[1] + viewDir[2]*reflectDir[2]
		if specDot < 0 {
			specDot = 0
		}

		specVal := float32(math.Pow(float64(specDot), float64(shininess)))
		specular = specVal
	}

	// 3. Attenuation
	attenuation := intensity / (constant + linear*distance + quadratic*(distance*distance))

	return vec3.T{
		lightColor[0] * (diffuse + specular) * attenuation,
		lightColor[1] * (diffuse + specular) * attenuation,
		lightColor[2] * (diffuse + specular) * attenuation,
	}
}

func fragShader(u float32, v float32, col vec4.T, norm vec3.T, fragPos vec3.T, s *api.FragmentShader) vec4.T {
	ctxAny, _ := s.GetUniform("ctx")
	ctx := ctxAny.(*ShaderContext)

	var texColor vec4.T
	if ctx.Texture != nil {
		texColor = ctx.Texture.Sample(u, v)
	} else {
		// texColor = ctx.Color
		texColor = col
	}

	texR := texColor[0] / 255.0
	texG := texColor[1] / 255.0
	texB := texColor[2] / 255.0
	alpha := texColor[3] / 255.0

	if ctx.IsStraight {
		return vec4.T{texR, texG, texB, alpha}
	}

	lenN := float32(math.Sqrt(float64(norm[0]*norm[0] + norm[1]*norm[1] + norm[2]*norm[2])))
	if lenN > 0 {
		norm[0] /= lenN
		norm[1] /= lenN
		norm[2] /= lenN
	}

	viewDir := vec3.T{ctx.ViewPos[0] - fragPos[0], ctx.ViewPos[1] - fragPos[1], ctx.ViewPos[2] - fragPos[2]}
	lenV := float32(math.Sqrt(float64(viewDir[0]*viewDir[0] + viewDir[1]*viewDir[1] + viewDir[2]*viewDir[2])))
	if lenV > 0 {
		viewDir[0] /= lenV
		viewDir[1] /= lenV
		viewDir[2] /= lenV
	}

	// 3. Ambient
	ambient := ctx.Lights.Ambient.Color
	resultR := texR * ambient[0]
	resultG := texG * ambient[1]
	resultB := texB * ambient[2]

	shininess := float32(64.0)

	// 4. Direct light
	dl := ctx.Lights.Directional
	lightDir := vec3.T{-dl.Direction[0], -dl.Direction[1], -dl.Direction[2]}

	contrib := calculateLightContribution(
		lightDir, 0.0,
		dl.Color, 1.0,
		1.0, 0.0, 0.0, // constant=1, linear=0, quadratic=0
		norm, viewDir, shininess,
	)
	resultR += texR * contrib[0]
	resultG += texG * contrib[1]
	resultB += texB * contrib[2]

	// 5. Point lights
	for _, pl := range ctx.Lights.PointLights {
		lightDir := vec3.T{pl.Position[0] - fragPos[0], pl.Position[1] - fragPos[1], pl.Position[2] - fragPos[2]}
		distance := float32(math.Sqrt(float64(lightDir[0]*lightDir[0] + lightDir[1]*lightDir[1] + lightDir[2]*lightDir[2])))

		if distance > 0 {
			lightDir[0] /= distance
			lightDir[1] /= distance
			lightDir[2] /= distance
		}

		contrib := calculateLightContribution(
			lightDir, distance, pl.Color, pl.Intensity,
			pl.Constant, pl.Linear, pl.Quadratic,
			norm, viewDir, shininess,
		)
		resultR += texR * contrib[0]
		resultG += texG * contrib[1]
		resultB += texB * contrib[2]
	}

	// 6. Spot lights
	for _, sl := range ctx.Lights.SpotLights {
		lightDir := vec3.T{sl.Position[0] - fragPos[0], sl.Position[1] - fragPos[1], sl.Position[2] - fragPos[2]}
		distance := float32(math.Sqrt(float64(lightDir[0]*lightDir[0] + lightDir[1]*lightDir[1] + lightDir[2]*lightDir[2])))

		if distance > 0 {
			lightDir[0] /= distance
			lightDir[1] /= distance
			lightDir[2] /= distance
		}

		theta := lightDir[0]*(-sl.Direction[0]) + lightDir[1]*(-sl.Direction[1]) + lightDir[2]*(-sl.Direction[2])

		if theta > sl.CosCutOff {
			epsilon := sl.CosCutOff - sl.OuterCos
			spotIntensity := float32(1.0)
			if epsilon > 0 {
				spotIntensity = (theta - sl.OuterCos) / epsilon
				if spotIntensity > 1.0 {
					spotIntensity = 1.0
				}
				if spotIntensity < 0.0 {
					spotIntensity = 0.0
				}
			}

			contrib := calculateLightContribution(
				lightDir, distance, sl.Color, sl.Intensity*spotIntensity,
				sl.Constant, sl.Linear, sl.Quadratic,
				norm, viewDir, shininess,
			)
			resultR += texR * contrib[0]
			resultG += texG * contrib[1]
			resultB += texB * contrib[2]
		}
	}

	if resultR > 1.0 {
		resultR = 1.0
	}
	if resultG > 1.0 {
		resultG = 1.0
	}
	if resultB > 1.0 {
		resultB = 1.0
	}

	return vec4.T{
		resultR,
		resultG,
		resultB,
		alpha,
	}
}

func NewBaseFragmentShader() *api.FragmentShader {
	return api.NewFragShader(fragShader)
}
