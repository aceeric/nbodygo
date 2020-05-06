package sim

import (
	"nbodygo/cmd/body"
	"nbodygo/internal/g3n/math32"
)

//
// Builder pattern. Builds an 'nBodySim' struct
//

type nBodySimBuilder struct {
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

func NewNBodySimBuilder() *nBodySimBuilder {
	// initialize a builder with reasonable defaults in case overrides are not provided
	b := nBodySimBuilder{
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

func (sb *nBodySimBuilder) Bodies(bodies []*body.Body) *nBodySimBuilder {
	sb.bodies = bodies
	return sb
}

func (sb *nBodySimBuilder) Workers(threads int) *nBodySimBuilder {
	sb.workers = threads
	return sb
}

func (sb *nBodySimBuilder) Scaling(scaling float64) *nBodySimBuilder {
	sb.scaling = scaling
	return sb
}

func (sb *nBodySimBuilder) InitialCam(initialCam math32.Vector3) *nBodySimBuilder {
	sb.initialCam = initialCam
	return sb
}

func (sb *nBodySimBuilder) SimWorker(simThread Worker) *nBodySimBuilder {
	sb.simThread = simThread
	return sb
}

func (sb *nBodySimBuilder) Render(render bool) *nBodySimBuilder {
	sb.render = render
	return sb
}

func (sb *nBodySimBuilder) Resolution(resolution [2]int) *nBodySimBuilder {
	sb.resolution = resolution
	return sb
}

func (sb *nBodySimBuilder) VSync(vSync bool) *nBodySimBuilder {
	sb.vSync = vSync
	return sb
}

func (sb *nBodySimBuilder) FrameRate(frameRate int) *nBodySimBuilder {
	sb.frameRate = frameRate
	return sb
}

func (sb *nBodySimBuilder) RunMillis(runMillis int) *nBodySimBuilder {
	sb.runMillis = runMillis
	return sb
}

func (sb *nBodySimBuilder) Build() *nBodySim {
	return &nBodySim{
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
