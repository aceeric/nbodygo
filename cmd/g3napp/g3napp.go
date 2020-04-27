package g3napp

import (
	"math/rand"
	"nbodygo/cmd/body"
	"nbodygo/cmd/flycam"
	"nbodygo/cmd/globals"
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

//
// G3nApp state
//
type G3nApp struct {
	// G3N managed
	app   *app.Application
	scene *core.Node
	// result queue holder provides lists of renderable objects each frame
	holder *runner.ResultQueueHolder
	// fly camera
	flyCam *flycam.FlyCam
	// G3N meshes in the scene graph - synced to the renderable objects obtained from 'holder'
	meshes map[int]*graphic.Mesh
	// each body that is a sun also creates a light source
	lightSources map[int]*light.Point
}

// singleton
var g3nApp G3nApp

//
// Starts the G3N render loop using the passed params
//
// args:
//   initialCam    - initial camera position (always looks at 0,0,0 from this vantage point)
//   width, height - screen dimensions
//   holder        - result queue holder - provides bodies to render
//   done          - channel to signal caller to indicate that the window was closed by virtue of
//                   the user pressing ESC
//
func StartG3nApp(initialCam *math32.Vector3, width, height int, holder *runner.ResultQueueHolder, done chan<- bool) {
	if g3nApp.app != nil {
		panic("Cannot call StartG3nApp twice")
	}
	go func() {
		g3nApp = G3nApp{
			app.App(width, height, "N-Body Golang Simulation"),
			core.NewNode(),
			holder,
			&flycam.FlyCam{},
			map[int]*graphic.Mesh{},
			map[int]*light.Point{},
		}
		gui.Manager().Set(g3nApp.scene)
		g3nApp.flyCam = flycam.NewFlyCam(window.Get().(*window.GlfwWindow), g3nApp.scene, width, height, *initialCam)

		// TODO register screen resize callback

		// set the background to black
		g3nApp.app.Gls().ClearColor(0.0, 0.0, 0.0, 1.0)
		g3nApp.app.Run(renderLoop) // G3N engine calls the passed function until user presses ESC
		done <- true               // user pressed ESC
	}()
}

//
// Causes the g3nApp.app.Run function to return in the go routine run by the 'StartG3nApp' function
//
func StopG3nApp() {
	g3nApp.app.IWindow.(*window.GlfwWindow).SetShouldClose(true)
}

//
// Callback - called by the G3N engine according to its hard-coded frame rate.
//
func renderLoop(renderer *renderer.Renderer, _ time.Duration) {
	updateSim()
	g3nApp.app.Gls().Clear(gls.DEPTH_BUFFER_BIT | gls.STENCIL_BUFFER_BIT | gls.COLOR_BUFFER_BIT)
	err := renderer.Render(g3nApp.scene, g3nApp.flyCam.Cam())
	if err != nil {
		// TODO PANIC?
	}
}

//
// Consumes the result queue holder to get a list of bodies and uses the list to update the
// G3N scene graph
//
func updateSim() {
	rq, ok := g3nApp.holder.Next()
	if !ok {
		return
	}
	renderedBodies := 0
	lightSources := 0
	for _, bri := range rq.Queue() {
		if !bri.Exists {
			// body no longer exists so remove from the scene graph
			if mesh, ok := g3nApp.meshes[bri.Id]; ok {
				removeMeshFromSceneGraph(mesh, bri.Id)
			}
		} else {
			var mesh *graphic.Mesh
			mesh, ok := g3nApp.meshes[bri.Id]
			if !ok {
				// add G3N representation of the body to our local list
				mesh = addBody(bri)
			} else {
				// allow a body radius to change. Note - the JMonkey API supports the ability to
				// change a sphere's radius on the fly. Didn't see an easy way to do this with G3N
				// so this is a work-around: just remove and re-add the body
				if float32(bri.Radius) != mesh.GetGeometry().BoundingBox().Max.X {
					removeMeshFromSceneGraph(mesh, bri.Id)
					mesh = addBody(bri)
				}
				// allow a body color to change if it is not a sun
				if bri.BodyColor != globals.Random && !bri.IsSun {
					mat := mesh.GetMaterial(0).(*material.Standard)
					color := mat.AmbientColor()
					if !color.Equals(xlatColor(bri.BodyColor)) {
						mat.SetColor(xlatColor(bri.BodyColor))
					}
				}
			}
			// update this body's position and if the body has a light source, also update that
			mesh.SetPosition(bri.X, bri.Y, bri.Z)
			if pl, ok := g3nApp.lightSources[bri.Id]; ok {
				pl.SetPosition(bri.X, bri.Y, bri.Z)
				lightSources++
			}
			renderedBodies++
		}
	}
}

//
// removes the passed mesh from the scene graph and if a light source is associated with the
// mesh (based on id) remove that as well
//
func removeMeshFromSceneGraph(mesh *graphic.Mesh, id int) {
	g3nApp.scene.Remove(mesh)
	delete(g3nApp.meshes, id)
	if l, ok := g3nApp.lightSources[id]; ok {
		g3nApp.scene.Remove(l)
		delete(g3nApp.lightSources, id)
	}
}

//
// Translates a sim body color to a G3N body color. These color names are compatible with the Java version.
// TODO support all G3N colors
//
func xlatColor(color globals.BodyColor) *math32.Color {
	switch color {
	case globals.Black:
		return &math32.Color{R: 0, G: 0, B: 0}
	case globals.White:
		return &math32.Color{R: 1, G: 1, B: 1}
	case globals.Darkgray:
		return &math32.Color{R: 0.663, G: 0.663, B: 0.663}
	case globals.Gray:
		return &math32.Color{R: 0.502, G: 0.502, B: 0.502}
	case globals.Lightgray:
		return &math32.Color{R: 0.827, G: 0.827, B: 0.827}
	case globals.Red:
		return &math32.Color{R: 1.000, G: 0.000, B: 0.000}
	case globals.Green:
		return &math32.Color{R: 0.000, G: 0.502, B: 0.000}
	case globals.Blue:
		return &math32.Color{R: 0.000, G: 0.000, B: 1.000}
	case globals.Yellow:
		return &math32.Color{R: 1.000, G: 1.000, B: 0.000}
	case globals.Magenta:
		return &math32.Color{R: 1.000, G: 0.000, B: 1.000}
	case globals.Cyan:
		return &math32.Color{R: 0.000, G: 1.000, B: 1.000}
	case globals.Orange:
		return &math32.Color{R: 1.000, G: 0.647, B: 0.000}
	case globals.Brown:
		return &math32.Color{R: 0.647, G: 0.165, B: 0.165}
	case globals.Pink:
		return &math32.Color{R: 1.000, G: 0.753, B: 0.796}
	case globals.Random:
		fallthrough
	default:
		rand.Seed(time.Now().UnixNano())
		return &math32.Color{R: rand.Float32(), G: rand.Float32(), B: rand.Float32()}
	}
}

//
// Converts the passed 'Renderable' into a G3N mesh, adds the mesh to the instance map of meshes, and also
// adds the mesh to the G3N scene graph
//
func addBody(bri *body.Renderable) *graphic.Mesh {
	var mesh *graphic.Mesh
	if bri.IsSun {
		geom := geometry.NewSphere(bri.Radius, 20, 20)
		mat := material.NewStandard(xlatColor(globals.White))
		mat.SetShininess(1)
		mat.SetEmissiveColor(xlatColor(globals.White))
		mesh = graphic.NewMesh(geom, mat)
		pl := light.NewPoint(xlatColor(globals.White), bri.Intensity)
		pl.SetLinearDecay(.00001)
		pl.SetQuadraticDecay(.00001)
		pl.SetPosition(bri.X, bri.Y, bri.Z)
		g3nApp.scene.Add(pl)
		g3nApp.lightSources[bri.Id] = pl
	} else {
		geom := geometry.NewSphere(bri.Radius, 20, 20)
		mat := material.NewStandard(xlatColor(bri.BodyColor))
		mesh = graphic.NewMesh(geom, mat)
	}
	g3nApp.scene.Add(mesh)
	g3nApp.meshes[bri.Id] = mesh
	return mesh
}
