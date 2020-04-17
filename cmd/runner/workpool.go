package runner

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

import (
	"nbodygo/cmd/body"
	"nbodygo/cmd/cmap"
	"runtime"
	"sync"
	"time"
)

// contains a body - to compute force for - and a map of all other bodies in the simulation
type Computation struct {
	body *body.SimBody
	bodyQueue *cmap.ConcurrentMap
}

// defines a function that computes force on a body from other bodies in the simulation
type ComputeFunc func(*Computation)

// Intended to be run as a goroutine. Pins itself to a CPU and runs indefinitely until
// it is signalled to stop. Receives work to perform on a channel in the Worker param
func bodyComputer(worker Worker, wg *sync.WaitGroup, computeFunc ComputeFunc) {
    // TODO PUT BACK
	runtime.LockOSThread()
	C.lock_thread(C.int(worker.cpu))
	if computeFunc == nil {
        computeFunc = DefaultComputeFunc
    }
	for {
		select {
		case <-worker.killChan:
			return
		case c := <-worker.computation:
            computeFunc(&c)
            worker.invocations++
            wg.Done()
		default:
			time.Sleep(time.Millisecond * 5)
		}
	}
}

// A dummy default compute function
func DefaultComputeFunc(*Computation) {
    time.Sleep(time.Millisecond * 50)
}

// defines that things needed by a body computer
type Worker struct {
    id          uint
	killChan    chan bool
	computation chan Computation
	cpu         uint
	invocations uint
}

// defines a worker pool
type WorkPool struct {
    wrkIdx uint
    cpus   uint
	workers []Worker
	wg      sync.WaitGroup
}

// creates a new worker pool with the passed number of threads, and the passed computation
// function. Spins up 'threads' number of goroutines running the  passed function
func NewWorkPool(threads int, computeFunc ComputeFunc) *WorkPool {
	wp := WorkPool{
        wrkIdx:  0,
        cpus:    uint(runtime.NumCPU()),
		workers: []Worker{},
		wg:      sync.WaitGroup{},
    }
	for i := 0; i < threads; i++ {
		w := Worker{
            id:          uint(i),
			killChan:    make(chan bool),
			computation: make(chan Computation, 4), // 4 is a guess, probably 2 would be ok...
			cpu:         uint(i) % wp.cpus,
            invocations: 0,
		}
		wp.workers = append(wp.workers, w)
		if computeFunc == nil {
            computeFunc = DefaultComputeFunc
        }
		go bodyComputer(w, &wp.wg, computeFunc)
	}
	return &wp
}

// signals all goroutines in the worker pool to stop. (They may not stop right away if they are performing
// a computation)
func (wp *WorkPool) StopAll() {
	for _, worker := range wp.workers {
		worker.killChan <- true
	}
}

// future
func (wp *WorkPool) setThreads(threads int) {
	// TODO handle increase / decrease in threads balancing across cpus
}

// creates a Computation from the passed args, and submits it round-robin to the pool. The design
// assumption is each computation will take approximately the same time to complete and so there doesn't
// need to be anything fancy with regard to finding the least utilized goroutine and assigning the
// work to that routine
func (wp *WorkPool) submit(b *body.SimBody, bodyQueue *cmap.ConcurrentMap) {
    worker := wp.workers[wp.wrkIdx % uint(len(wp.workers))]
    wp.wg.Add(1)
    // computation channel is buffered so we get concurrency as well as a limited number of threads
    // with thread-cpu affinity
	//start := time.Now()
    worker.computation<- Computation{
        body:      b,
        bodyQueue: bodyQueue,
    }
	//stop := time.Now()
	//millis := stop.Sub(start).Milliseconds()
	//fmt.Printf("Worker id %v, latency millis=%v\n", worker.id, millis)
    wp.wrkIdx++
}

// waits for one work to complete
func (wp *WorkPool) take() {
	wp.wg.Wait()
}
