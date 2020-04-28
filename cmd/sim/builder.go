package sim

import (
	"nbodygo/cmd/body"
	"nbodygo/internal/g3n/math32"
)

//
// Builder pattern. Builds an 'NBodySim' struct
//

type NBodySimBuilder struct {
	bodies     []*body.Body
	workers    int
	scaling    float64
	initialCam math32.Vector3
	simThread  Worker
	render     bool
	resolution [2]int
	vSync      bool
	frameRate  int
	runMillis  int
}

func NewNBodySimBuilder() *NBodySimBuilder {
	// initialize a builder with reasonable defaults in case overrides are not provided
	b := NBodySimBuilder{
		bodies:     []*body.Body{}, // no bodies
		workers:    defaultWorkers,
		scaling:    defaultTimeScaling,
		initialCam: *math32.NewVector3(100, 100, 100),
		simThread:  nil,
		render:     true,
		resolution: [2]int{2560, 1440},
		vSync:      true, // not currently used
		frameRate:  -1,   // not currently used
		runMillis:  -1,
	}
	return &b
}

func (sb *NBodySimBuilder) Bodies(bodies []*body.Body) *NBodySimBuilder {
	sb.bodies = bodies
	return sb
}

func (sb *NBodySimBuilder) Workers(threads int) *NBodySimBuilder {
	sb.workers = threads
	return sb
}

func (sb *NBodySimBuilder) Scaling(scaling float64) *NBodySimBuilder {
	sb.scaling = scaling
	return sb
}

func (sb *NBodySimBuilder) InitialCam(initialCam math32.Vector3) *NBodySimBuilder {
	sb.initialCam = initialCam
	return sb
}

func (sb *NBodySimBuilder) SimWorker(simThread Worker) *NBodySimBuilder {
	sb.simThread = simThread
	return sb
}

func (sb *NBodySimBuilder) Render(render bool) *NBodySimBuilder {
	sb.render = render
	return sb
}

func (sb *NBodySimBuilder) Resolution(resolution [2]int) *NBodySimBuilder {
	sb.resolution = resolution
	return sb
}

func (sb *NBodySimBuilder) VSync(vSync bool) *NBodySimBuilder {
	sb.vSync = vSync
	return sb
}

func (sb *NBodySimBuilder) FrameRate(frameRate int) *NBodySimBuilder {
	sb.frameRate = frameRate
	return sb
}

func (sb *NBodySimBuilder) RunMillis(runMillis int) *NBodySimBuilder {
	sb.runMillis = runMillis
	return sb
}

func (sb *NBodySimBuilder) Build() *NBodySim {
	return &NBodySim{
		bodies:     sb.bodies,
		workers:    sb.workers,
		scaling:    sb.scaling,
		initialCam: sb.initialCam,
		simWorker:  sb.simThread,
		render:     sb.render,
		resolution: sb.resolution,
		runMillis:  sb.runMillis,
	}
}
