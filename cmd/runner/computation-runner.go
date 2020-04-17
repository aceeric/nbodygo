package runner

import (
	"fmt"
	"nbodygo/cmd/body"
	"nbodygo/cmd/cmap"
	"time"
)

type ComputationRunner struct {
	stopChannel              chan bool
	timeScaling              float32
	resultQueueHolder        ResultQueueHolder
	bodyQueue                *cmap.ConcurrentMap // TODO rename
	threads                  int
	iterations, computations uint
	wp                       *WorkPool
}

type RunnerStats struct {
	TimeScaling              float32
	Threads                  int
	Iterations, Computations uint
}

// singleton
var runner ComputationRunner

func newRunner(threadCount int, timeScaling float32, resultQueueHolder ResultQueueHolder, bodyQueue *cmap.ConcurrentMap) {
	if runner.stopChannel != nil {
		panic("Cannot call newRunner twice")
	}
	runner = ComputationRunner{
		stopChannel:       make(chan bool, 1),
		timeScaling:       timeScaling,
		resultQueueHolder: resultQueueHolder,
		bodyQueue:         bodyQueue,
		threads:           threadCount,
		wp:                NewWorkPool(threadCount, calcForce),
	}
	if runner.resultQueueHolder.queues == nil { // TODO DELETEME
		panic("newRunner runner.resultQueueHolder.queues == nil")
	}
}

func StartComputationRunner(threadCount int, bodyQueue *cmap.ConcurrentMap, timeScaling float32,
	resultQueueHolder ResultQueueHolder) *ComputationRunner {
	newRunner(threadCount, timeScaling, resultQueueHolder, bodyQueue)
	go runner.run()
	return &runner
}

func StopComputationRunner() RunnerStats {
	stats := RunnerStats{
		TimeScaling:  runner.timeScaling,
		Threads:      runner.threads,
		Iterations:   runner.iterations,
		Computations: runner.computations,
	}
	if runner.stopChannel != nil {
		runner.stopChannel <- true
		<-runner.stopChannel
		runner.wp.StopAll()
		runner = ComputationRunner{}
	}
	return stats
}

// TODO this doesn't do anything yet - has to be integrated into the work pool
func (runner *ComputationRunner) SetThreads(threads int) {
	runner.threads = threads
}

func (runner *ComputationRunner) run() {
	// TODO PANIC/RECOVER
	for {
		select {
		default:
			runner.runOneComputation()
		case <-runner.stopChannel:
			runner.stopChannel <- true
			return
		}
	}
}

func calcForce(computation *Computation) {
	start := time.Now()
	b := *computation.body
	b.ForceComputer(computation.bodyQueue)
	stop := time.Now()
	millis := stop.Sub(start).Milliseconds()
	_ = start; _ = stop; _ = millis
	//fmt.Printf("Force computation for id %v, elapsed millis=%v\n", b.Id(), millis)
}

func (runner *ComputationRunner) runOneComputation() {
	runner.iterations++
	//fmt.Printf("runOneComputation runner.iterations=%v\n", runner.iterations)
	if runner.resultQueueHolder.queues == nil { // TODO DELETEME
		panic(fmt.Sprintf("runOneComputation runner.resultQueueHolder.queues == nil. runner.iterations=%v\n", runner.iterations))
	}
	if runner.resultQueueHolder.IsFull() {
		//fmt.Printf("runOneComputation no queues -- runner.iterations=%v\n", runner.iterations)
		time.Sleep(time.Millisecond * 5)
		return
	}
	//fmt.Printf("begin submit\n")
	bodies := 0
	start := time.Now()
	for item := range runner.bodyQueue.IterBuffered() {
		b := item.Val.(body.SimBody) // no pointer
		//fmt.Printf("submit body: %+v\n", b)
		runner.wp.submit(&b, runner.bodyQueue)
		bodies++
	}
	stop := time.Now()
	millis := stop.Sub(start).Milliseconds()
	//fmt.Printf("Computation runner submit millis=%v\n",  millis)
	if bodies == 0 {
		//fmt.Printf("no bodies\n")
		// no bodies in the sim - don't peg the CPU
		time.Sleep(5 * time.Millisecond)
		return
	}
	//fmt.Printf("begin take\n")
	start = time.Now()
	for i := 0; i < bodies; i++ {
		runner.wp.take()
	}
	stop = time.Now()
	millis = stop.Sub(start).Milliseconds()
	//fmt.Printf("Computation runner take millis=%v\n",  millis)
	//fmt.Printf("begin update\n")
	start = time.Now()
	rq, _ := runner.resultQueueHolder.NewQueue(bodies)
	for item := range runner.bodyQueue.IterBuffered() {
		b := item.Val.(body.SimBody) // no pointer
		ri := b.Update(runner.timeScaling)
		rq.AddRenderable(ri)
		if !b.Exists() {
			runner.bodyQueue.Remove(b.Id()) // TODO TEST THIS
		}
	}
	stop = time.Now()
	millis = stop.Sub(start).Milliseconds()
	_ = start; _ = stop; _ = millis
	//fmt.Printf("Computation runner update millis=%v\n",  millis)
	//fmt.Printf("set computed\n")
	rq.computed = true
	runner.computations++
}
