package runner

import (
	"fmt"
	"nbodygo/cmd/body"
	"runtime"
	"time"
)

//
// Computation runner state
//
type ComputationRunner struct {
	// stops the  runner
	stop chan bool
	// number of workers to create in the worker pool
	workerCnt int
	// metrics
	iterations, computations uint
	submits, waits           uint
	submitMillis             int64
	waitMillis               int64
	startTime                time.Time
	stopTime                 time.Time
	goroutines               int
	// the work pool
	wp *WorkPool
	// the bodies in the sim
	sbc body.SimBodyCollection
	// supports test - stop after this many iterations
	maxIteration int
	// true if running
	running           bool
	// applied to the force and velocity by the body force computer
	timeScaling       float64
	// holds computed results for the render engine
	resultQueueHolder ResultQueueHolder
}

//
// Prints metrics to the console
//
func (r *ComputationRunner) PrintStats() {
	totalMillis := r.stopTime.Sub(r.startTime).Milliseconds()
	fps := float32(r.computations) / float32(totalMillis) * 1000
	fmt.Printf("Runner\n workerCnt: %v\n submits: %v\n waits: %v\n iterations: %v\n computations: %v\n"+
		" submitMillis: %v\n submitMillis/computation: %v\n waitMillis: %v\n waitMillis/computation: %v\n"+
		" frames per second: %v\n avg # goroutines: %v\n elapsed time: %v\n",
		r.workerCnt, r.submits, r.waits, r.iterations, r.computations, r.submitMillis,
		float32(r.submitMillis)/float32(r.computations), r.waitMillis, float32(r.waitMillis)/float32(r.computations),
		fps, float32(r.goroutines)/float32(r.computations), r.stopTime.Sub(r.startTime))
	r.wp.PrintStats()
}

//
// Creates a new computation runner
//
// args:
//   workerCnt         Number of workers in the pool
//   timeScaling       Multiplier for force and velocity calc
//   resultQueueHolder Holds computed results
//
func NewComputationRunner(workerCnt int, timeScaling float64, resultQueueHolder ResultQueueHolder,
	cc body.SimBodyCollection) *ComputationRunner {
	r := ComputationRunner{
		stop:              make(chan bool),
		workerCnt:         workerCnt,
		wp:                NewWorkPool(workerCnt, cc),
		sbc:               cc,
		timeScaling:       timeScaling,
		resultQueueHolder: resultQueueHolder,
	}
	return &r
}

//
// Supports debugging - sets the max iterations for the runner
//
func (r *ComputationRunner) SetMaxIterations(maxIteration int) *ComputationRunner {
	r.maxIteration = maxIteration
	return r
}

//
// Starts the runner
//
func (r *ComputationRunner) Start() *ComputationRunner {
	go r.run()
	return r
}

//
// Stops the runner
//
func (r *ComputationRunner) Stop() {
	if r.running {
		r.stop <- true
		<-r.stop
	}
	r.wp.ShutDownWorkPool()
}

// TODO this doesn't do anything yet - has to be integrated into the work pool
func (r *ComputationRunner) SetWorkers(workerCnt int) {
	r.workerCnt = workerCnt
}

//
// Main runner. A go routine that runs until stopped. Runs one computation, and monitors
// the stop channel in a loop
//
func (r *ComputationRunner) run() {
	r.startTime = time.Now()
	for r.running = true; r.running; {
		select {
		default:
			r.runOneComputation()
			if r.maxIteration > 0 {
				if r.maxIteration--; r.maxIteration == 0 {
					r.running = false
				}
			}
			runtime.Gosched()
		case <-r.stop:
			r.running = false
		}
	}
	r.stopTime = time.Now()
	r.stop <- true
}
//
// Runs one computation. Executes a nested loop:
//   for each body
//     for each other-body
//       compute the force on body from other-body
//
// Each body from the outer loop is submitted into the worker pool, and has access to the whole body
// collection . Therefore, each body is free to update its own force without thread synchronization on the
// force member fields because its the only body doing that calculation on itself. The application
// of the total final force to the body velocity and position is performed as the last step once
// the entire collection of bodies have had their force computed.
//
// So at that time, it is safe to update the velocity and position without synchronization because no
// other threads are reading the bodies. The results are stored in a queue of {@link BodyRenderInfo}
// instances which the graphics engine consumes. The graphics engine continually gets a copy of the
// body values (and only what it needs to render the visuals) so there is never thread contention
// between the graphics engine and the body position computation
//
func (r *ComputationRunner) runOneComputation() {
	r.iterations++
	rq, ok := r.resultQueueHolder.NewResultQueue()
	if !ok {
		return
	}
	start := time.Now()
	bodies := 0
	r.sbc.IterateOnce(func(c body.SimBody) {
		r.wp.submit(c)
		r.submits++
		bodies++
	})
	if bodies == 0 {
		return
	}
	r.submitMillis += time.Now().Sub(start).Milliseconds()
	start = time.Now()
	r.wp.wait()
	r.waits++
	r.waitMillis += time.Now().Sub(start).Milliseconds()

	r.sbc.ProcessMods()

	r.sbc.IterateOnce(func(c body.SimBody) {
		ri := c.Update(r.timeScaling)
		rq.AddRenderable(ri)
	})
	r.resultQueueHolder.SetComputed(rq)
	r.sbc.Cycle()
	r.computations++
	r.goroutines += runtime.NumGoroutine()
}
