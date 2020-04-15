package main

import (
	"nbodygo/cmd/flycam"
	"nbodygo/internal/pkg/app"
	"nbodygo/internal/pkg/core"
	"nbodygo/internal/pkg/geometry"
	"nbodygo/internal/pkg/gls"
	"nbodygo/internal/pkg/graphic"
	"nbodygo/internal/pkg/gui"
	"nbodygo/internal/pkg/light"
	"nbodygo/internal/pkg/material"
	"nbodygo/internal/pkg/math32"
	"nbodygo/internal/pkg/renderer"
	"nbodygo/internal/pkg/window"
	"time"
)

const (
	width = 2560
	height = 1440
)

func main() {
	// Create application and scene
	a := app.App(width, height)
	scene := core.NewNode()

	// Set the scene to be managed by the gui manager
	gui.Manager().Set(scene)

	// Create a Fly Camera and attach it to the scene
	flyCam := flycam.NewFlyCam(window.Get().(*window.GlfwWindow), scene, width, height)

	// TODO
	//onResize := func(evname string, ev interface{}) {
	//	width, height := a.GetSize()
	//	a.Gls().Viewport(0, 0, int32(width), int32(height))
	//	// Update the camera's aspect ratio
	//	flyCam.cam.SetAspect(float32(width) / float32(height))
	//}
	//a.Subscribe(window.OnWindowSize, onResize)

	geom := geometry.NewSphere(10, 20, 20)
	mat := material.NewStandard(math32.NewColor("firebrick"))
	mesh := graphic.NewMesh(geom, mat)
	scene.Add(mesh)

	// Create and add light
	l := light.NewDirectional(&math32.Color{1, 1, 1}, 5.0)
	l.SetPosition(1000, 0, 0)
	scene.Add(l)

	// Set background color to black
	a.Gls().ClearColor(0.0, 0.0, 0.0, 1.0)

	// Run the application - the application will call the render function according to a hard-coded frame rate
	a.Run(func(renderer *renderer.Renderer, deltaTime time.Duration) {
		a.Gls().Clear(gls.DEPTH_BUFFER_BIT | gls.STENCIL_BUFFER_BIT | gls.COLOR_BUFFER_BIT)
		renderer.Render(scene, flyCam.Cam())
		println(deltaTime.Milliseconds())
	})
}
