package g3napp

import (
	"math/rand"
	"nbodygo/cmd/flycam"
	"nbodygo/cmd/globals"
	"nbodygo/cmd/interfaces"
	"nbodygo/cmd/runner"
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
	holder runner.ResultQueueHolder
	flyCam *flycam.FlyCam
	meshes map[int]*graphic.Mesh
	lightSources map[int]*light.Point
	initialCam *math32.Vector3
}

// singleton
var g3nApp G3nApp

func StartG3nApp(initialCam *math32.Vector3, width, height int, holder runner.ResultQueueHolder, done chan<- bool) {
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
			map[int]*graphic.Mesh{},
			map[int]*light.Point{},
			initialCam,
		}
		gui.Manager().Set(g3nApp.scene)
		g3nApp.flyCam = flycam.NewFlyCam(window.Get().(*window.GlfwWindow), g3nApp.scene, width, height, *initialCam)  // TODO test in struct initializer

		// TODO register screen resize callback

		// set the background to black
		g3nApp.app.Gls().ClearColor(0.0, 0.0, 0.0, 1.0)
		g3nApp.app.Run(renderLoop) // G3N engine calls the passed function
		done<- true
	}()
}

// causes the g3nApp.app.Run function to return, resulting in the 'done' channel passed to the
// 'StartG3nApp' function to get signalled so the caller of that function knows that the app is
// stopped.
func StopG3nApp() {
	g3nApp.app.IWindow.(*window.GlfwWindow).SetShouldClose(true)
}

// Callback - called by the G3N engine according to its hard-coded frame rate.
func renderLoop(renderer *renderer.Renderer, _ time.Duration) {
	updateSim()
	g3nApp.app.Gls().Clear(gls.DEPTH_BUFFER_BIT | gls.STENCIL_BUFFER_BIT | gls.COLOR_BUFFER_BIT)
	err := renderer.Render(g3nApp.scene, g3nApp.flyCam.Cam())
	if err != nil {
		// TODO PANIC?
		//fmt.Printf("render error: %v\n", err)
	}
}

// Consumes the result queue holder to get a list of bodies and uses the list to update the simulation
func updateSim() {
	rq, ok := g3nApp.holder.NextComputedQueue()
	if !ok {
		return
	}

	//println("got a computed queue")
	renderedBodies := 0
	lightSources := 0
	for _, bri := range rq.Queue() {
		//fmt.Printf("g3napp updateSim bri=%+v\n", bri)
		if !bri.Exists() {
			// body no longer exists so remove from the scene graph
			if mesh, ok := g3nApp.meshes[bri.Id()]; ok {
				g3nApp.scene.Remove(mesh)
				delete(g3nApp.meshes, bri.Id())
				if l, ok := g3nApp.lightSources[bri.Id()]; ok {
					g3nApp.scene.Remove(l)
				}
			}
		} else {
			var mesh *graphic.Mesh
			mesh, ok := g3nApp.meshes[bri.Id()]
			if !ok {
				// add G3N representation of the body to our local list
				mesh = addBody(bri)
			}
			// TODO how to interrogate and change radius?

			// allow a body color to change if it is not a sun
			if bri.BodyColor() != globals.Random && !bri.IsSun() {
				mat := mesh.GetMaterial(0).(*material.Standard)
				color := mat.AmbientColor()
				if !color.Equals(xlatColor(bri.BodyColor())) {
					mat.SetColor(xlatColor(bri.BodyColor()))
				}
			}
			// update this body's position and if the body has a light source, also update that
			mesh.SetPosition(bri.X32(), bri.Y32(), bri.Z32())
			if pl, ok := g3nApp.lightSources[bri.Id()]; ok {
				pl.SetPosition(bri.X32(), bri.Y32(), bri.Z32())
				lightSources++
			}
			renderedBodies++
		}
	}
	//fmt.Printf("g3napp rendered %v bodies and %v light sources\n", renderedBodies, lightSources)
}

//
// Translates a sim body color to a G3N body color. These colors are compatible with the Java version.
// TODO support all G3N colors
//
func xlatColor(color globals.BodyColor) *math32.Color {
	switch color {
	case globals.Black: return &math32.Color{0,0,0}
	case globals.White: return &math32.Color{1,1,1}
	case globals.Darkgray: return &math32.Color{0.663, 0.663, 0.663}
	case globals.Gray: return &math32.Color{0.502, 0.502, 0.502}
	case globals.Lightgray: return &math32.Color{0.827, 0.827, 0.827}
	case globals.Red: return &math32.Color{1.000, 0.000, 0.000}
	case globals.Green: return &math32.Color{0.000, 0.502, 0.000}
	case globals.Blue: return &math32.Color{0.000, 0.000, 1.000}
	case globals.Yellow: return &math32.Color{1.000, 1.000, 0.000}
	case globals.Magenta: return &math32.Color{1.000, 0.000, 1.000}
	case globals.Cyan: return &math32.Color{0.000, 1.000, 1.000}
	case globals.Orange: return &math32.Color{1.000, 0.647, 0.000}
	case globals.Brown: return &math32.Color{0.647, 0.165, 0.165}
	case globals.Pink: return &math32.Color{1.000, 0.753, 0.796}
	case globals.Random:
		fallthrough
	default:
		rand.Seed(time.Now().UnixNano())
		return &math32.Color{rand.Float32(), rand.Float32(), rand.Float32()}
	}
}

//
// Converts the passed renderable into a G3N mesh, adds the mesh to the instance map of meshes, and also
// adds the mesh to the G3N scene graph
//
func addBody(bri interfaces.Renderable) *graphic.Mesh {
	var mesh *graphic.Mesh
	if bri.IsSun() {
		geom := geometry.NewSphere(float64(bri.Radius()), 20, 20)
		mat := material.NewStandard(xlatColor(globals.White))
		mat.SetShininess(1)
		mat.SetEmissiveColor(xlatColor(globals.White))
		mesh = graphic.NewMesh(geom, mat)
		pl := light.NewPoint(xlatColor(globals.White), 100)
		pl.SetLinearDecay(.00001)
		pl.SetQuadraticDecay(.00001)
		pl.SetPosition(bri.X32(), bri.Y32(), bri.Z32())
		g3nApp.scene.Add(pl)
		g3nApp.lightSources[bri.Id()] = pl
	} else {
		geom := geometry.NewSphere(float64(bri.Radius()), 20, 20)
		mat := material.NewStandard(xlatColor(bri.BodyColor()))
		mesh = graphic.NewMesh(geom, mat)
	}
	g3nApp.scene.Add(mesh)
	g3nApp.meshes[bri.Id()] = mesh
	//fmt.Printf("g3napp.addBody added renderable: %+v\n", bri)
	return mesh
}