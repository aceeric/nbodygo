package sim

import (
	"nbodygo/cmd/body"
	"nbodygo/cmd/g3napp"
	"nbodygo/cmd/grpcserver"
	"nbodygo/cmd/runner"
	"nbodygo/internal/pkg/math32"
	"runtime"
	"time"
)

const (
	//Default value for the number of result queues
	defaultMaxResultQueues = 10

	// If not rendering, then the amount of time to sleep between polling the result queue. Goal is to
	// ensure the computation runner is running at full speed
	noRenderSleepMs = 5

	// a default scaling factor if an override is not provided
	defaultTimeScaling = .000000001

	// default computation workers in the work pool
	defaultWorkers = 5
)

//
// N-Body Sim state. Initialized by the 'NBodySimBuilder' function
//
type NBodySim struct {
	// A list of bodies to start the simulation with
	bodies []body.SimBody

	// The number of workers to use for the computation runner
	workers int

	// The time scaling factor, which speeds or slows the sim
	scaling float64

	// The initial camera position
	initialCam math32.Vector3

	// If not nil, then the sim runner will call the function after the sim is started. The
	// function can then modify the body queue while the sim is running and exit when it is done
	simWorker SimWorker

	// If false, then don't start the rendering engine. Useful for testing/debugging since the
	// rendering engine and OpenGL can interfere with single-stepping in the IDE
	render bool

	// Screen resolution
	resolution [2]int

	// if --no-render specified, then user can also specify --runmillis indicating how long
	// to run the sim
	runMillis int
}

//
// Simulation runner
//
// - Initializes instrumentation which - depending on TODO what is Go equivalent of JVM properties - could be
//   NOP instrumentation, or Prometheus instrumentation
// - Initializes a collection to hold all the bodies in the simulation
// - Initializes a result queue holder to hold computed results
// - Initializes a computation runner and starts it, which perpetually computes the sim in a thread,
//   placing the computed results into the result queue holder
// - Initializes the G3N graphics engine and starts it - which renders the computed results from the result queue
//   perpetually in a thread
// - Starts a gRPC server to handle requests from external entities to modify various aspects of the simulation
//   while it is running (e.g. to add bodies or change sim characteristics)
// - Waits for the G3N goroutine to exit (when the user presses ESC)
// - Cleans up
//
func (sim NBodySim) Run() {
	// TODO start instrumentation
	sbc := body.NewSimBodyCollection(sim.bodies) // todo return interface vs return struct consistency
	rqh := runner.NewResultQueueHolder(defaultMaxResultQueues, true) // todo all NEWs return pointers uniformly
	simDone := make(chan bool) // to shut down the G3N engine
	if sim.render {
		g3napp.StartG3nApp(&sim.initialCam, sim.resolution[0], sim.resolution[1], &rqh, simDone)
	}
	crunner := runner.NewComputationRunner(sim.workers, sim.scaling, &rqh, sbc).Start()
	grpcserver.Start(newGrpcSimCb(sbc, crunner, &rqh))
	if sim.simWorker != nil {
		go sim.simWorker(sbc)
	}
	waitForSimEnd(sim.render, &rqh, simDone, sim.runMillis)
	grpcserver.Stop()
	crunner.Stop()
	// TODO stop instrumentation
	crunner.PrintStats()
}

//
// If rendering, then blocks on the passed 'simDone' channel and then returns. The channel is signaled by the
// 'g3napp' package when the user presses ESC. If not rendering, then loops perpetually consuming the passed
// result queue holder so the computation runner can run at max capacity. (Supports test/performance analysis)
//
// args:
//  render    - if true, waits for the graphics engine to signal the passed 'simDone' channel then returns
//  rqh       - holds queues of bodies with updated position
//  simDone   - channel that G3N will signal on
//  runMillis - if --no-render, then an amount of time to run before exiting. If -1, runs forever
//
func waitForSimEnd(render bool, rqh *runner.ResultQueueHolder, simDone chan bool, runMillis int) {
	if render {
		// wait for the user to press ESC which shuts down the G3N and then signals the simDone channel
		<-simDone
	} else {
		start := time.Now()
		for {
			rq, ok := rqh.Next()
			if ok {
				dummy := float64(0)
				for _, bri := range rq.Queue() {
					dummy += bri.Radius()
				}
			}
			time.Sleep(noRenderSleepMs)
			if runMillis != -1 {
				elapsed := int(time.Now().Sub(start).Milliseconds())
				if elapsed > runMillis {
					return
				}
			}
			runtime.Gosched()
		}
	}
}

//
// Builder pattern // TODO move to its own file?
//
type NBodySimBuilder interface {
	Bodies([]body.SimBody) NBodySimBuilder
	Workers(int) NBodySimBuilder
	Scaling(float64) NBodySimBuilder
	InitialCam(math32.Vector3) NBodySimBuilder
	SimWorker(SimWorker) NBodySimBuilder
	Render(bool) NBodySimBuilder
	Resolution([2]int) NBodySimBuilder
	VSync(bool) NBodySimBuilder
	FrameRate(int) NBodySimBuilder
	RunMillis(int) NBodySimBuilder
	Build() NBodySim
}

type nBodySimBuilder struct {
	bodies     []body.SimBody
	workers    int
	scaling    float64
	initialCam math32.Vector3
	simThread  SimWorker
	render     bool
	resolution [2]int
	vSync      bool
	frameRate  int
	runMillis  int
}

func (b nBodySimBuilder) Bodies(bodies []body.SimBody) NBodySimBuilder {
	b.bodies = bodies
	return b
}

func (b nBodySimBuilder) Workers(threads int) NBodySimBuilder {
	b.workers = threads
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

func (b nBodySimBuilder) SimWorker(simThread SimWorker) NBodySimBuilder {
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

func (b nBodySimBuilder) RunMillis(runMillis int) NBodySimBuilder {
	b.runMillis = runMillis
	return b
}

func (b nBodySimBuilder) Build() NBodySim {
	return newNBodySim(b)
}

func newNBodySim(b nBodySimBuilder) NBodySim {
	return NBodySim{
		bodies:     b.bodies,
		workers:    b.workers,
		scaling:    b.scaling,
		initialCam: b.initialCam,
		simWorker:  b.simThread,
		render:     b.render,
		resolution: b.resolution,
		runMillis:  b.runMillis,
	}
}

func NewNBodySimBuilder() NBodySimBuilder {
	// initialize a builder with reasonable defaults in case overrides are not provided
	b := nBodySimBuilder{
		bodies:     []body.SimBody{}, // no bodies
		workers:    defaultWorkers,
		scaling:    defaultTimeScaling,
		initialCam: *math32.NewVector3(100, 100, 100),
		simThread:  nil,
		render:     true,
		resolution: [2]int{2560, 1440},
		vSync:      true, // not currently used
		frameRate:  -1, // not currently used
		runMillis:  -1,
	}
	return b
}
