package sim

import (
	"nbodygo/cmd/collection"
	"nbodygo/cmd/g3napp"
	"nbodygo/cmd/interfaces"
	"nbodygo/cmd/runner"
	"nbodygo/internal/pkg/math32"
	"time"
)

const (
	//Default value for the number of result queues
	defaultMaxResultQueues = 10

	// If no rendering, then the amount of time to sleep between polling the result queue. Goal is the
	// ensure the computation runner is running at full throttle
	noRenderSleepMs = 500 // TODO PUT BACK

	// a default scaling factor if an override is not provided
	defaultTimeScaling = .000000001

	// default computation goroutines
	defaultThreads = 5
)

type NBodySim struct {
	// A list of bodies to start the simulation with
	bodies []interfaces.SimBody

	// The number of threads to use for the computation runner
	threads int

	// The time scaling factor, which speeds or slows the sim
	scaling float64

	// The initial camera position
	initialCam math32.Vector3

	// If not null, then the sim runner will call the function after the sim is started.
	// The function is expected to then modify the body queue while the sim is running and exit when
	// it is done
	simThread func(cc interfaces.SimBodyCollection) // TODO RENAME GOROUTINE OR SOMETHING

	// If false, then don't start the rendering engine. Useful for testing/debugging since the
	// rendering engine and OpenGL can interfere with single-stepping in the IDE
	render bool

	// Screen resolution. Note - depending on the resolution specified, on a dual monitor system the OpenGL
	// subsystem may locate the sim window on a monitor of its choosing, rather than on the primary monitor
	resolution [2]int
}

func (sim NBodySim) Run() {
	// TODO start instrumentation
	bodyQueue := collection.NewSimBodyCollection(sim.bodies)
	rqh := runner.NewResultQueueHolder(defaultMaxResultQueues)
	// used to gracefully shut down the G3N engine
	simDone := make(chan bool)
	//sim.render = false
	if sim.render {
		g3napp.StartG3nApp(&sim.initialCam, sim.resolution[0], sim.resolution[1], rqh, simDone)
	}
	cr := runner.NewComputationRunner(sim.threads, sim.scaling, rqh, bodyQueue).Start()
	// TODO start gRPC server
	if sim.simThread != nil {
		go sim.simThread(bodyQueue) // TODO needs a channel to close
	}
	waitForSimEnd(sim.render, &rqh, simDone)
	// TODO signal simThread if it is running
	// TODO stop gRPC server
	cr.Stop()
	// TODO stop instrumentation
}

func waitForSimEnd(render bool, rqh *runner.ResultQueueHolder, simDone chan bool) {
	if render {
		// wait for the user to press ESC which shuts down the G3N and then signals the simDone channel
		<-simDone
	} else {
		for {
			// TODO need a way to get out of this?
			rq, ok := rqh.NextComputedQueue()
			_ = rq; _ = ok
			time.Sleep(noRenderSleepMs)
		}
	}
}

type NBodySimBuilder interface {
	Bodies([]interfaces.SimBody) NBodySimBuilder
	Threads(int) NBodySimBuilder
	Scaling(float64) NBodySimBuilder
	InitialCam(math32.Vector3) NBodySimBuilder
	SimThread(func(interfaces.SimBodyCollection)) NBodySimBuilder
	Render(bool) NBodySimBuilder
	Resolution([2]int) NBodySimBuilder
	VSync(bool) NBodySimBuilder
	FrameRate(int) NBodySimBuilder
	Build() NBodySim
}

type nBodySimBuilder struct {
	bodies     []interfaces.SimBody
	threads    int
	scaling    float64
	initialCam math32.Vector3
	simThread  func(interfaces.SimBodyCollection)
	render     bool
	resolution [2]int
	vSync      bool
	frameRate  int
}

func (b nBodySimBuilder) Bodies(bodies []interfaces.SimBody) NBodySimBuilder {
	b.bodies = bodies
	return b
}
func (b nBodySimBuilder) Threads(threads int) NBodySimBuilder {
	b.threads = threads
	return b
}

func (b nBodySimBuilder) Scaling(scaling float64) NBodySimBuilder {
	b.scaling = scaling
	return b
}

func (b nBodySimBuilder) InitialCam(initialCam math32.Vector3) NBodySimBuilder {
	b.initialCam = initialCam
	return b
}

func (b nBodySimBuilder) SimThread(simThread func(interfaces.SimBodyCollection)) NBodySimBuilder {
	b.simThread = simThread
	return b
}

func (b nBodySimBuilder) Render(render bool) NBodySimBuilder {
	b.render = render
	return b
}

func (b nBodySimBuilder) Resolution(resolution [2]int) NBodySimBuilder {
	b.resolution = resolution
	return b
}

func (b nBodySimBuilder) VSync(vSync bool) NBodySimBuilder {
	b.vSync = vSync
	return b
}

func (b nBodySimBuilder) FrameRate(frameRate int) NBodySimBuilder {
	b.frameRate = frameRate
	return b
}

func (b nBodySimBuilder) Build() NBodySim {
	return newNBodySim(b)
}

func newNBodySim(b nBodySimBuilder) NBodySim {
	return NBodySim{
		bodies:     b.bodies,
		threads:    b.threads,
		scaling:    b.scaling,
		initialCam: b.initialCam,
		simThread:  b.simThread,
		render:     b.render,
		resolution: b.resolution,
	}
}

func NewNBodySimBuilder() NBodySimBuilder {
	// initialize a builder with reasonable defaults in case overrides are not provided
	b := nBodySimBuilder{
		bodies:     []interfaces.SimBody{}, // no bodies
		threads:    defaultThreads,
		scaling:    defaultTimeScaling,
		initialCam: *math32.NewVector3(100, 100, 100),
		simThread:  nil,
		render:     true,
		resolution: [2]int{2560, 1440},
		vSync:      true, // not currently used
		frameRate:  -1, // not currently used
	}
	return b
}
