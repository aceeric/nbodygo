package runner

import (
	"fmt"
	"nbodygo/cmd/interfaces"
	"sync"
	"time"
)

//
// Defines the worker pool state
//
type WorkPool struct {
	// dynamically updated by the 'submit' function to round-robin work into the pool
	wrkIdx      uint
	// workers
	workers     []*Worker
	// wait group - waits for all submitted work to complete
	wg          sync.WaitGroup
	// metrics
	submissions int64
	millis      int64
	// the simulation body collection
	sbc         interfaces.SimBodyCollection
}

//
// Defines the worker state
//
type Worker struct {
	// unique ID per worker
	id          uint
	// how the worker pool shuts down workers
	stop        chan bool
	// how the worker receives work
	compute     chan interfaces.SimBody
	// metrics
	invocations int
	millis      int64
}
//
// A worker goroutine that is started by the work pool. Waits for a body on its 'compute' channel and when
// it gets a body, calls the 'ForceComputer' method on the the body. Is stopped using the 'stop'
// channel
//
// args:
//   w    This worker
//   wg   Wait group to signal completion of force calculation for the body
//   sbc  The collection wrapper over the bodies in the sim
//
func worker(w *Worker, wg *sync.WaitGroup, sbc interfaces.SimBodyCollection) {
	millis := int64(0)
	invocations := 0
	for {
		select {
		case <-w.stop:
			// note that this may leave the go routine with items still enqueued on the w.compute channel
			// so this shutdown leaves unfinished work
			w.invocations = invocations
			w.millis = millis
			w.stop<- true // acknowledge stop
			return
		case c := <-w.compute:
			start := time.Now()
			c.ForceComputer(sbc)
			invocations++
			wg.Done()
			millis += time.Now().Sub(start).Milliseconds()
		default:
		}
	}
}

//
// Creates a work pool
//
// args:
//   goroutines  The number of go routines to create in the pool
//   sbc         The collection wrapper over the bodies in the sim
//
// returns:
//   pointer to created pool
//
func NewWorkPool(goroutines int, sbc interfaces.SimBodyCollection) *WorkPool {
	wp := WorkPool{
		wrkIdx:      0,
		workers:     []*Worker{},
		wg:          sync.WaitGroup{},
		submissions: 0,
		millis:      0,
		sbc:         sbc,
	}
	for i := 0; i < goroutines; i++ {
		w := Worker{
			id:          uint(i),
			stop:        make(chan bool),
			compute:     make(chan interfaces.SimBody, 5),
			invocations: 0,
			millis:      0,
		}
		wp.workers = append(wp.workers, &w)
		go worker(&w, &wp.wg, sbc)
	}
	return &wp
}

//
// Shuts down the workers in the pool, waiting for them all to acknowledge shutdown
// before returning to the caller
//
func (wp *WorkPool) ShutDownWorkPool() {
	for _, w := range wp.workers {
		w.stop <- true
		<-w.stop
	}
}

//
// Prints stats to the console
//
func (wp *WorkPool) PrintStats() {
	fmt.Printf("Worker Pool\n submissions: %v\n millis: %v\n millis/submission: %v\n", wp.submissions,
		wp.millis, float32(wp.millis)/float32(wp.submissions))
	for _, w := range wp.workers {
		fmt.Printf("> Worker id: %v invocations: %v millis: %v millis/invocation: %v\n", w.id,
			w.invocations, w.millis, float32(w.millis)/float32(w.invocations))
	}
}

//
// Submits a body to the pool for computation
//
func (wp *WorkPool) submit(c interfaces.SimBody) {
	start := time.Now()
	wp.wg.Add(1)
	wp.workers[wp.wrkIdx%uint(len(wp.workers))].compute <- c
	wp.submissions++
	wp.wrkIdx++
	wp.millis += time.Now().Sub(start).Milliseconds()
}

//
// Waits for all submitted work to complete
//
func (wp *WorkPool) wait() {
	wp.wg.Wait()
}
