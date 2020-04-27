package runner

import (
	"fmt"
	"nbodygo/cmd/body"
	"runtime"
	"sync"
	"time"
)

/*
#define _GNU_SOURCE
#include <sched.h>
#include <pthread.h>
void lock_thread(int cpuid) {
        pthread_t tid;
        cpu_set_t cpuset;
        tid = pthread_self();
        CPU_ZERO(&cpuset);
        CPU_SET(cpuid, &cpuset);
    pthread_setaffinity_np(tid, sizeof(cpu_set_t), &cpuset);
}
*/
import "C"

//
// Defines the worker pool state
//
type WorkPool struct {
	// dynamically updated by the 'submit' function to round-robin work into the pool
	wrkIdx uint
	// workers
	workers []*Worker
	// wait group - waits for all submitted work to complete
	wg sync.WaitGroup
	// metrics
	submissions int64
	millis      int64
	// the simulation body collection
	bc *body.BodyCollection
	// how to resize the pool
	resizeCh chan int
	// tracks the ID of the most recently created worker
	maxId    int
}

//
// Defines the worker state
//
type Worker struct {
	// unique ID per worker
	id int
	// how the worker pool shuts down workers
	stop chan bool
	// how the worker receives work
	compute chan *body.Body
	// how the worker receives work
	computeSlice chan []*body.Body
	// metrics
	invocations int
	millis      int64
}

//
// A worker goroutine that is started by the work pool. Waits for a body on its 'compute' channel and when
// it gets a body, calls the 'Compute' method on the the body. Is stopped using the 'stop'
// channel
//
// args:
//   w   This worker
//   wg  Wait group to signal completion of force calculation for the body
//   bc  The collection wrapper over the bodies in the sim
//
func worker(w *Worker, wg *sync.WaitGroup, bc *body.BodyCollection) {
	millis := int64(0)
	//runtime.LockOSThread()
	//C.lock_thread(C.int(w.id))
	invocations := 0
	for {
		select {
		case <-w.stop:
			// note that this may leave the go routine with items still enqueued on the compute channel
			// so this shutdown can leave unfinished work
			w.invocations = invocations
			w.millis = millis
			w.stop <- true // acknowledge stop
			return
		case c := <-w.compute:
			start := time.Now()
			c.Compute(bc)
			invocations++
			wg.Done()
			millis += time.Now().Sub(start).Milliseconds()
		case slice := <-w.computeSlice:
			start := time.Now()
			for _, b := range slice {
				b.Compute(bc)
			}
			invocations++
			wg.Done()
			millis += time.Now().Sub(start).Milliseconds()
		default:
		}
		runtime.Gosched()
	}
}

//
// Creates a work pool
//
// args:
//   goroutines  The number of go routines to create in the pool
//   bc          The collection wrapper over the bodies in the sim
//
// returns:
//   pointer to created pool
//
func NewWorkPool(goroutines int, bc *body.BodyCollection) *WorkPool {
	wp := &WorkPool{
		wrkIdx:      0,
		workers:     []*Worker{},
		wg:          sync.WaitGroup{},
		submissions: 0,
		millis:      0,
		bc:          bc,
		resizeCh:    make(chan int, 1),
		maxId:       -1,
	}
	wp.createWorkers(goroutines)
	return wp
}

//
// Creates workers until the number of workers in the pool equals the passed value
//
// args:
//   goroutines - the number of go routines desired in the pool
//
func (wp *WorkPool) createWorkers(goroutines int) {
	for ; len(wp.workers) < goroutines; {
		wp.maxId++
		w := Worker{
			id:           wp.maxId,
			stop:         make(chan bool),
			compute:      make(chan *body.Body, 5), // TODO REVISIT COUNT
			computeSlice: make(chan []*body.Body, 5),
			invocations:  0,
			millis:       0,
		}
		wp.workers = append(wp.workers, &w)
		go worker(&w, &wp.wg, wp.bc)
	}
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
// signals the internal channel with a new pool size. The 'submit' functions monitor
// the channel and implement the resize
//
func (wp *WorkPool) SetPoolSize(workers int) {
	wp.resizeCh <- workers
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
func (wp *WorkPool) submit(c *body.Body) {
	wp.chkResize()
	start := time.Now()
	wp.wg.Add(1)
	wp.workers[wp.wrkIdx%uint(len(wp.workers))].compute <- c
	wp.submissions++
	wp.wrkIdx++
	wp.millis += time.Now().Sub(start).Milliseconds()
}

//
// Submits a slice of the body array to the pool for computation - slightly faster
// than the other way
//
func (wp *WorkPool) submitSlice(bodySlice []*body.Body) {
	wp.chkResize()
	start := time.Now()
	wp.wg.Add(1)
	wp.workers[wp.wrkIdx%uint(len(wp.workers))].computeSlice <- bodySlice
	wp.submissions++
	wp.wrkIdx++
	wp.millis += time.Now().Sub(start).Milliseconds()
}

//
// Checks the internal work pool channel that is used to signal a resize request
// and handles the request
//
func (wp *WorkPool) chkResize() {
	select {
	case goroutines := <-wp.resizeCh:
		switch {
		case goroutines < len(wp.workers):
			for i := goroutines; i < len(wp.workers); i++ {
				wp.workers[i].stop <- true
				<-wp.workers[i].stop
			}
			wp.workers = wp.workers[0:goroutines]
			wp.wrkIdx = 0
		case goroutines > len(wp.workers):
			wp.createWorkers(goroutines)
			wp.wrkIdx = 0
		}
	default:
		return
	}
}

//
// Waits for all submitted work to complete
//
func (wp *WorkPool) wait() {
	wp.wg.Wait()
}
