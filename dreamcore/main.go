package main

import (
	"context"
	"log"

	"github.com/striter-no/softengine/api"
	"github.com/striter-no/softengine/api/shaders"
	"github.com/striter-no/softengine/entity"
	"github.com/striter-no/softengine/lights"
	"github.com/striter-no/softgo/api/assets"
	"github.com/striter-no/softgo/api/keyboard"
	"github.com/ungerik/go3d/vec3"
)

func main() {
	engine, err := api.NewEngine(context.Background())
	if err != nil {
		panic(err)
	}

	defer engine.End()

	// Init
	engine.RScreen.BackColor = vec3.T{0.8, 0.8, 1}
	engine.LightConfig.Ambient = lights.AmbientLight{Color: vec3.T{0.1, 0.1, 0.1}}
	engine.LightConfig.Directional = lights.DirectLight{
		Color:       vec3.T{.8, 0.9, 1},
		Direction:   vec3.T{-0.2, -.5, -0.2},
		CastShadows: true,
	}
	spotlight := lights.NewSpotLight(
		vec3.T{1.0, .0, 1.0},
		vec3.T{0.0, 150.0, 0.0},
		vec3.T{0.0, -1.0, -0.2},
		4.0, 1.0, 0.009, 0.00032, // attenuation
		12.5, 17.5,
		true,
	)
	engine.NewSpotLight(spotlight)

	engine.UpdateShaders(
		shaders.NewBaseFragmentShader(),
		shaders.NewBaseVertexShader(),
	)

	engine.InitCamera(vec3.T{0, 0, 2}, 0.08, 100, 0.1, 2000, 90)
	engine.Camera.Locked = true

	// Ambient

	windID := engine.SoundSystem.AddSpeaker("./assets/sounds/wind.mp3", 1, 1)
	if windID == -1 {
		log.Fatal("Failed to load sound")
	}

	engine.SoundSystem.PlayID(windID)

	// Objects

	grassTex, err := entity.NewModelImageTexture("./assets/textures/grass.jpg")
	if err != nil {
		panic(err)
	}

	// grassMesh, err := assets.LoadOBJ("./assets/meshes/plane.obj")
	generator := entity.NewTerrainGenerator(100.0, 0.2, 50)
	grassMesh := generator.Generate(150, 150)

	grassObj := entity.NewObject3D(
		vec3.T{0, 0, 0},
		vec3.T{0, 0, 0},
		vec3.T{1, 1, 1},
		grassMesh, grassTex, true,
	)

	if _, err = engine.AddObject(grassObj); err != nil {
		panic(err)
	}

	monkey, _ := assets.LoadOBJ("./assets/meshes/suzanne.obj")
	onigiriTex, _ := entity.NewModelImageTexture("./assets/textures/onigiri.jpg")

	monkeyObj := entity.NewObject3D(
		vec3.T{0, 100, 0},
		vec3.T{0, 0, 0},
		vec3.T{30, 30, 30},
		monkey, onigiriTex, true,
	)

	if _, err = engine.AddObject(monkeyObj); err != nil {
		panic(err)
	}

	// Skybox

	// skyboxTex, err := entity.NewModelImageTexture("./assets/textures/skybox.png")
	// if err != nil {
	// 	panic(err)
	// }

	// skyboxMesh, err := assets.LoadOBJ("./assets/meshes/skybox.obj")

	// skyboxObj := entity.NewObject3D(
	// 	vec3.T{0, 0, 0},
	// 	vec3.T{0, 0, 0},
	// 	vec3.T{1500, 1500, 1500},
	// 	skyboxMesh, skyboxTex, false,
	// )

	// var skyboxID int
	// if _, err = engine.AddObject(skyboxObj); err != nil {
	// 	panic(err)
	// }

	// Run
	engine.RScreen.SSAAFactor = 1
	for engine.IsRunning() {
		if engine.Keyboard.IsKeyPressed(keyboard.KeyEsc) {
			break
		}

		engine.UpdateHID()

		engine.SoundSystem.ChangeIDPosition(windID, engine.Camera.Position)
		engine.SoundSystem.UpdateListener(engine.Camera.Position)

		// skyboxObj.Position = engine.Camera.Position
		// skyboxObj.UpdateMat()
		// spotlight.Position = engine.Camera.Position

		monkeyObj.LookAt(engine.Camera.Position, true)

		engine.Camera.Speed = 200 * engine.TSystem.DeltaTime

		if err := engine.DrawObjects(); err != nil {
			panic(err)
		}
		engine.Blit()
	}
}
