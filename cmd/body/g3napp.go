package body // TODO SHOULD NOT BE IN BODY?

import (
	"fmt"
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

type G3nApp struct {
	app *app.Application
	scene *core.Node
	holder ResultQueueHolder
	flyCam *flycam.FlyCam
}

//type renderLoop func(*renderer.Renderer, time.Duration)

// singleton
var g3nApp G3nApp

func StartG3nApp(width, height int, holder ResultQueueHolder, done chan<- bool) {
	if g3nApp.app != nil {
		panic("Cannot call StartG3nApp twice")
	}
	go func() {
		// TODO support initial cam location
		g3nApp = G3nApp{
			app.App(width, height, "N-Body Golang Simulation"),
			core.NewNode(),
			holder,
			&flycam.FlyCam{},
		}
		gui.Manager().Set(g3nApp.scene)
		g3nApp.flyCam = flycam.NewFlyCam(window.Get().(*window.GlfwWindow), g3nApp.scene, width, height)

		// TODO register screen resize callback

		// set the background to black
		g3nApp.app.Gls().ClearColor(0.0, 0.0, 0.0, 1.0)

		// Create and add light TODO REMOVE ONCE SUN IS FUNCTIONAL
		l := light.NewDirectional(&math32.Color{1, 1, 1}, 5.0)
		l.SetPosition(1000, 0, 0)
		g3nApp.scene.Add(l)

		//// DELETE THIS
		//geom := geometry.NewSphere(10, 20, 20)
		//mat := material.NewStandard(math32.NewColor("firebrick"))
		//mesh := graphic.NewMesh(geom, mat)
		//mesh.SetPosition(0, 0, 0)
		//g3nApp.scene.Add(mesh)

		g3nApp.app.Run(renderLoop)
		done<- true
	}()
}



func renderLoop(renderer *renderer.Renderer, deltaTime time.Duration) {
	updateSim()
	g3nApp.app.Gls().Clear(gls.DEPTH_BUFFER_BIT | gls.STENCIL_BUFFER_BIT | gls.COLOR_BUFFER_BIT)
	err := renderer.Render(g3nApp.scene, g3nApp.flyCam.Cam())
	if err != nil {
		fmt.Printf("render error: %v\n", err)
	}
}

func updateSim() {
	rq, ok := g3nApp.holder.nextComputedQueue()
	if !ok {
		return
	}
	bri := rq.queue[0]
	geom := geometry.NewSphere(float64(bri.radius), 20, 20)
	mat := material.NewStandard(math32.NewColor("firebrick"))
	mesh := graphic.NewMesh(geom, mat)
	mesh.SetPosition(bri.x, bri.y, bri.z)
	g3nApp.scene.Add(mesh)
}