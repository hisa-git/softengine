package main

import (
	"context"
	"math/rand"
	"os"
	"os/signal"

	"github.com/striter-no/softengine/api"
	"github.com/striter-no/softengine/api/shaders"
	"github.com/striter-no/softengine/entity"
	"github.com/striter-no/softengine/lights"
	"github.com/striter-no/softgo/api/assets"
	"github.com/ungerik/go3d/vec3"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	engine, err := api.NewEngine(ctx)
	if err != nil {
		panic(err)
	}

	defer engine.End()

	// Init lightning
	engine.RScreen.BackColor = vec3.T{0, 0, 0}
	engine.LightConfig.Ambient = lights.AmbientLight{Color: vec3.T{0.1, 0.1, 0.1}}
	engine.LightConfig.Directional = lights.DirectLight{
		Color:     vec3.T{1.0, 0.3, 0.2},
		Direction: vec3.T{-0.5, -1.0, -0.2},
	}

	bulb := &lights.PointLight{
		Color:     vec3.T{0.1, 0.2, 0.9},
		Position:  vec3.T{0, 0, 0},
		Intensity: 1.0,
		Constant:  1.0, Linear: 0.09, Quadratic: 0.32,
	}

	engine.NewPointLight(bulb) // bulbID :=

	// Adding shaders
	engine.UpdateShaders(
		shaders.NewBaseFragmentShader(),
		shaders.NewBaseVertexShader(),
	)

	// Adding camera
	engine.InitCamera(vec3.T{0, 0, 2}, 0.08, 0.1)
	engine.Camera.Locked = true

	// Adding object
	cubeTex, err := entity.NewModelImageTexture("./assets/textures/onigiri.jpg")
	if err != nil {
		panic(err)
	}

	mesh, err := assets.LoadOBJ("./assets/meshes/suzanne.obj")
	if err != nil {
		panic(err)
	}

	// adding LODs ---
	meshMed := api.GenerateLOD(mesh, 0.2)
	meshLow := api.GenerateLOD(mesh, 0.5)
	// ---

	cubeObj := entity.NewObject3D(
		vec3.T{0, 0, 0},
		vec3.T{0, 0, 0},
		vec3.T{1, 1, 1},
		mesh, cubeTex,
	)

	cubeObj.AddLOD(meshMed, 15.0)
	cubeObj.AddLOD(meshLow, 35.0)

	for range 100 {
		clone := cubeObj.Clone()
		clone.Position = vec3.T{
			(rand.Float32() - 0.5) * 60,
			(rand.Float32() - 0.5) * 60,
			(rand.Float32() - 0.5) * 60,
		}
		if _, err := engine.AddObject(clone); err != nil {
			panic(err)
		}
	}

	// Main loop
	for engine.IsRunning() {
		engine.UpdateHID()

		bulb.Position = engine.Camera.Position
		for _, obj := range engine.Objects {
			obj.LookAt(engine.Camera.Position, true)
		}

		engine.Camera.Speed = 20 * engine.TSystem.DeltaTime

		if err := engine.DrawObjects(); err != nil {
			panic(err)
		}
		engine.Blit()
	}
}
