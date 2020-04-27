package runner

import (
	"fmt"
	"nbodygo/cmd/body"
	"runtime"
	"time"
)

//
// The computation runner runs the n-body computation perpetually in a loop until signaled to stop. The
// runner contains a worker pool and a reference to the sim body collection.
//

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
	submitMillis, waitMillis int64
	startTime, stopTime      time.Time
	goroutines               int
	// the work pool
	wp *WorkPool
	// the bodies in the sim
	bc *body.BodyCollection
	// supports test - stop after this many iterations
	maxIteration int
	// true if running
	running bool
	// applied to the force and velocity by the body force computer as a smoothing factor
	timeScaling   float64
	timeScaleChan chan float64
	// holds computed results for the render engine
	resultQueueHolder *ResultQueueHolder
	// coefficient of restitution for elastic collision
	R       float64
	RChan   chan float64
	deletes int
	delChan chan int
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
//   timeScaling       Multiplier for force and velocity calc - determines sim "speed"
//   resultQueueHolder Holds computed results
//   bc                Collection of bodies in the simulation
//
func NewComputationRunner(workerCnt int, timeScaling float64, resultQueueHolder *ResultQueueHolder,
	bc *body.BodyCollection) *ComputationRunner {
	r := ComputationRunner{
		stop:              make(chan bool),
		workerCnt:         workerCnt,
		wp:                NewWorkPool(workerCnt, bc),
		bc:                bc,
		timeScaling:       timeScaling,
		timeScaleChan:     make(chan float64, 1),
		resultQueueHolder: resultQueueHolder,
		R:                 1,
		RChan:             make(chan float64, 1),
		delChan:           make(chan int, 1),
	}
	return &r
}

//
// Supports debugging - sets the max iterations for the runner. After this many iterations the
// runner will shut down
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

//
// Calls into the work pool contained in the struct to change the pool size. This is enqueued
// and handled by the pool the next time work is submitted to the pool
//
func (r *ComputationRunner) SetWorkers(workerCnt int) {
	r.workerCnt = workerCnt
	r.wp.SetPoolSize(workerCnt)
}

//
// Returns the time scaling factor in the runner
//
func (r *ComputationRunner) TimeScaling() float64 {
	return r.timeScaling
}

//
// Sets the time scaling factor in the runner to the passed value
//
func (r *ComputationRunner) SetTimeScaling(timeScaling float64) {
	r.timeScaleChan <- timeScaling
}

//
// If a change the the time scale has been enqueued in the channel, use
// it to update the time scale
//
func (r *ComputationRunner) updateTimeScaling() {
	select {
	case r.timeScaling = <-r.timeScaleChan:
		return
	default:
	}
}

//
// Returns the coefficient of restitution
//
func (r *ComputationRunner) CoefficientOfRestitution() float64 {
	return r.R
}

//
// Sets the coefficient of restitution in the runner to the passed value
//
func (r *ComputationRunner) SetCoefficientOfRestitution(R float64) {
	r.RChan <- R
}

//
// If a change the the coefficient of restitution has been enqueued in the channel, use
// it to update the coefficient of restitution
//
func (r *ComputationRunner) updateCoefficientOfRestitution() {
	select {
	case r.R = <-r.RChan:
		return
	default:
	}
}

//
// Sends a message to delete the passed number of bodies from the sim
//
func (r *ComputationRunner) RemoveBodies(deletes int) {
	r.delChan <- deletes
}

//
// Handles a request to remove bodies from the sim
//
func (r *ComputationRunner) processDeletes() {
	select {
	case r.deletes = <-r.delChan:
		delCnt := r.deletes
		r.deletes = 0
		if delCnt == -1 {
			r.bc.IterateOnce(func(b *body.Body) {
				b.SetNotExists()
			})
		} else {
			removedCnt, step, iter := 0, 0, 0
			if delCnt > r.bc.Count() {
				step = 1
			} else {
				step = r.bc.Count() / delCnt
			}
			shouldRemove := false
			r.bc.IterateOnce(func(b *body.Body) {
				if iter%step == 0 {
					shouldRemove = true
				}
				iter++
				if shouldRemove && !b.Pinned && b.Exists {
					b.SetNotExists()
					shouldRemove = false
					removedCnt++
					if removedCnt >= delCnt {
						return
					}
				}
			})
		}
		return
	default:
	}
}

//
// Returns the count of workers in the worker pool
//
func (r *ComputationRunner) WorkerCount() int {
	return r.workerCnt
}

//
// Main runner. A go routine that runs until stopped. In a loop: runs one computation, and monitors
// the stop channel
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
		case <-r.stop:
			r.running = false
		}
		runtime.Gosched()
	}
	r.stopTime = time.Now()
	r.stop <- true
}

//
// Runs one computation. Executes a nested loop:
//   for each body
//     for each other-body
//       compute the force on body from other-body and detect collisions
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
// In order to synchronize access to the body collection this function also kind of serves as the traffic
// cop for adds/deletes/mods while the sim is running
//
func (r *ComputationRunner) runOneComputation() {
	r.iterations++
	r.updateTimeScaling()
	r.updateCoefficientOfRestitution()
	r.processDeletes()
	r.bc.HandleGetBody()
	r.bc.HandleModBody()
	rq, ok := r.resultQueueHolder.NewResultQueue()
	if !ok {
		return
	}
	start := time.Now()
	submits := 0

	// slightly better performance this way - give each worker a slice to work on
	workers := len(r.wp.workers)
	arr := r.bc.GetArray()
	max := len(arr)
	size := max / workers
	if max < 100 {
		// for small simulations just use one worker
		size = len(arr)
	}
	for offset := 0; offset < max; offset += size {
		end := offset + size
		if offset+size > max {
			end = max
		}
		r.wp.submitSlice(arr[offset:end])
		r.submits++
		submits++
	}
	if submits != 0 {
		r.submitMillis += time.Now().Sub(start).Milliseconds()
		start = time.Now()
		r.wp.wait()
		r.waits++
		r.waitMillis += time.Now().Sub(start).Milliseconds()
	}
	/*
		// this initial approach submits to the work pool one body at a time, which
		// is how the Java app does it
		r.bc.IterateOnce(func(b *body.Body) {
			if b.Exists {
				r.wp.submit(b)
				r.submits++
				submits++
			}
		})
		if submits != 0 {
			r.submitMillis += time.Now().Sub(start).Milliseconds()
			start = time.Now()
			r.wp.wait()
			r.waits++
			r.waitMillis += time.Now().Sub(start).Milliseconds()
		}
	*/
	r.bc.ProcessMods()
	r.bc.IterateOnce(func(b *body.Body) {
		ri := b.Update(r.timeScaling, r.R)
		rq.Add(ri)
	})
	r.resultQueueHolder.Add(rq)
	r.bc.Cycle(r.R)
	r.computations++
	r.goroutines += runtime.NumGoroutine()
}
