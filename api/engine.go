package api

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/striter-no/softengine/api/shaders"
	"github.com/striter-no/softengine/entity"
	"github.com/striter-no/softengine/lights"
	"github.com/striter-no/softengine/sounds"
	sapi "github.com/striter-no/softgo/api"
	"github.com/striter-no/softgo/api/keyboard"
	"github.com/striter-no/softgo/api/mouse"
	"github.com/striter-no/softgo/render"
	"github.com/striter-no/stg/graphics"
	"github.com/ungerik/go3d/vec3"
	"github.com/ungerik/go3d/vec4"
)

type Engine struct {
	ctx      context.Context
	Mouse    mouse.WindowMouse
	Keyboard keyboard.WindowKeyboard

	Camera  *sapi.Camera
	RScreen *sapi.RenderScreen
	TSystem TimeSystem

	FragShader **sapi.FragmentShader
	VertShader **sapi.VertexShader

	Objects       map[int]*entity.Object3D
	incrementalID int

	lastUpdate  time.Time
	LightConfig lights.LightingConfig

	MainFBO     *render.Framebuffer
	SoundSystem *sounds.SoundSystem
}

func NewEngine(ctx context.Context) (*Engine, error) {

	Mouse, err := mouse.NewWindowMouse()
	if err != nil {
		return nil, err
	}

	Keyboard, err := keyboard.NewWindowKeyboard()
	if err != nil {
		return nil, err
	}

	s, err := sapi.NewRenderScreen(ctx)
	if err != nil {
		return nil, err
	}

	SoundSystem, err := sounds.NewSoundSystem(vec3.T{0, 0, 0})
	if err != nil {
		return nil, err
	}

	Mouse.LockCursor()
	Mouse.HideMouse()

	s.SSAAFactor = 1
	s.Init()

	return &Engine{
		ctx:        ctx,
		Mouse:      Mouse,
		Keyboard:   Keyboard,
		Objects:    make(map[int]*entity.Object3D),
		RScreen:    s,
		FragShader: &s.FragShader,
		VertShader: &s.VertexShader,
		LightConfig: lights.LightingConfig{
			PointLights: make(map[int]*lights.PointLight),
			SpotLights:  make(map[int]*lights.SpotLight),
		},
		MainFBO:     render.NewFramebuffer(s.Screen.Width*s.SSAAFactor, s.Screen.Height*s.SSAAFactor, false),
		SoundSystem: SoundSystem,
	}, nil
}

func (e *Engine) InitCamera(position vec3.T, sensitivity, speed, near, far, fov float32) {
	e.Camera = sapi.NewCamera(position, sensitivity, speed, e.Mouse, e.Keyboard, near, far, fov)
}

func (e *Engine) AddObject(obj *entity.Object3D) (int, error) {
	if obj == nil {
		return 0, errors.New("Cannot add nil object")
	}

	id := e.incrementalID
	e.Objects[id] = obj

	e.incrementalID++
	return id, nil
}

func (e *Engine) GetObject(id int) (*entity.Object3D, error) {
	if obj, ok := e.Objects[id]; ok {
		return obj, nil
	}

	return nil, errors.New("failed to get object")
}

func (e *Engine) RemoveObject(id int) {
	delete(e.Objects, id)
}

func (e *Engine) IsRunning() bool {
	return e.RScreen.IsOpen()
}

func (e *Engine) UpdateHID() {
	e.lastUpdate = time.Now()

	e.Mouse.PollEvents()
	e.Keyboard.PollEvents()

	if e.RScreen.Screen.Height == 0 {
		return
	}

	aspect := float32(e.RScreen.Screen.Width) / (float32(e.RScreen.Screen.Height))
	e.Camera.UpdateOnHID(aspect)
}

func (e *Engine) UpdateShaders(
	fragShader *sapi.FragmentShader,
	vertShader *sapi.VertexShader,
) {
	e.RScreen.FragShader = fragShader
	e.RScreen.VertexShader = vertShader
}

func (e *Engine) NewSpotLight(conf *lights.SpotLight) int {
	e.LightConfig.SpotLights[e.incrementalID] = conf
	e.incrementalID++

	return e.incrementalID - 1
}

func (e *Engine) NewPointLight(conf *lights.PointLight) int {
	e.LightConfig.PointLights[e.incrementalID] = conf
	e.incrementalID++

	return e.incrementalID - 1
}

func (e *Engine) RemovePointLigth(id int) {
	delete(e.LightConfig.PointLights, id)
}

func (e *Engine) RemoveSpotLigth(id int) {
	delete(e.LightConfig.SpotLights, id)
}

func (e *Engine) DrawObjects() error {
	e.RScreen.Clear()
	if e.MainFBO.Width != e.RScreen.Screen.Width*e.RScreen.SSAAFactor ||
		e.MainFBO.Height != e.RScreen.Screen.Height*e.RScreen.SSAAFactor {

		e.MainFBO = render.NewFramebuffer(e.RScreen.Screen.Width*e.RScreen.SSAAFactor, e.RScreen.Screen.Height*e.RScreen.SSAAFactor, false)
	}

	e.MainFBO.Clear(e.RScreen.BackColor)

	type RenderNode struct {
		Obj              *entity.Object3D
		DistanceToCamera float32
		MVP              mgl32.Mat4
	}

	var renderQueue []RenderNode

	for _, obj := range e.Objects {
		model := obj.GetModelMatrix()
		mvp := e.Camera.VP.Mul4(model)

		center := mgl32.Vec4{0, 0, 0, 1}
		clipCenter := mvp.Mul4x1(center)

		maxScale := obj.Scale[0]
		if obj.Scale[1] > maxScale {
			maxScale = obj.Scale[1]
		}
		if obj.Scale[2] > maxScale {
			maxScale = obj.Scale[2]
		}

		actualRadius := obj.BaseRadius * maxScale

		if clipCenter.W() < -actualRadius {
			continue
		}

		if clipCenter.W() > 0 {
			ndcX := clipCenter.X() / clipCenter.W()
			ndcY := clipCenter.Y() / clipCenter.W()
			ndcZ := clipCenter.Z() / clipCenter.W()

			bound := 1.5 * (1.0 + (actualRadius / clipCenter.W()))
			zbound := (1.0 + (actualRadius / clipCenter.W()))

			if ndcX < -bound || ndcX > bound || ndcY < -bound || ndcY > bound || ndcZ > zbound || ndcZ < -zbound {
				continue
			}
		}

		renderQueue = append(renderQueue, RenderNode{
			Obj:              obj,
			DistanceToCamera: clipCenter.W(),
			MVP:              mvp,
		})
	}

	sort.Slice(renderQueue, func(i, j int) bool {
		return renderQueue[i].DistanceToCamera > renderQueue[j].DistanceToCamera
	})

	for _, node := range renderQueue {
		obj := node.Obj
		activeMesh := obj.GetActiveMesh(node.DistanceToCamera)

		ctx := &shaders.ShaderContext{
			MVP:     node.MVP,
			Model:   obj.GetModelMatrix(),
			ViewPos: e.Camera.Position,

			Texture: obj.Texture.Texture,
			Color:   vec4.T{obj.Texture.BaseColor[0], obj.Texture.BaseColor[1], obj.Texture.BaseColor[2], 1},

			Lights:     e.LightConfig,
			IsStraight: !obj.CanBeLit,
		}

		(*e.VertShader).SetUniform("ctx", ctx)
		(*e.FragShader).SetUniform("ctx", ctx)
		if err := e.RScreen.DrawCall(activeMesh, e.MainFBO); err != nil {
			return err
		}
	}

	return nil
}

func (e *Engine) Blit() {

	e.RScreen.Present(e.MainFBO)

	e.RScreen.Screen.SetText(
		0, 0, fmt.Sprintf("FPS: %.1f", e.RScreen.CurrentFPS), graphics.NewFGPixel(255, 255, 255, ""),
	)

	e.RScreen.Screen.Blit()
	e.TSystem.FPS = float32(e.RScreen.CurrentFPS)
	e.TSystem.DeltaTime = float32(time.Since(e.lastUpdate).Milliseconds()) / 1000
	e.TSystem.Ticks++
}

func (e *Engine) End() {

	e.Mouse.UnlockCursor()
	e.Mouse.ShowMouse()
	e.Mouse.Close()

	e.Keyboard.Close()
	e.RScreen.End()

	e.SoundSystem.End()
}
