package body

import (
	"nbodygo/cmd/cmap"
	"time"
)

type ComputationRunner struct {
	stopChannel chan bool
	timeScaling float32
	resultQueueHolder ResultQueueHolder
	bodyQueue *cmap.ConcurrentMap // TODO rename
	threads int
	iterations, computations uint
}

// singleton
var runner ComputationRunner

func newRunner(timeScaling float32, resultQueueHolder ResultQueueHolder, bodyQueue *cmap.ConcurrentMap) {
	if runner.stopChannel != nil {
		panic("Cannot call newRunner twice")
	}
	runner = ComputationRunner {
		stopChannel: make(chan bool, 1),
		timeScaling: timeScaling,
		resultQueueHolder: resultQueueHolder,
		bodyQueue: bodyQueue,
		threads: 5, // TODO
	}
}

func StartComputationRunner(timeScaling float32, resultQueueHolder ResultQueueHolder,
	bodyQueue *cmap.ConcurrentMap) *ComputationRunner {
	newRunner(timeScaling, resultQueueHolder, bodyQueue)
	go runner.run()
	return &runner
}

func StopComputationRunner() {
	if runner.stopChannel != nil {
		runner.stopChannel <- true
		<-runner.stopChannel
		runner = ComputationRunner{}
	}
}
// deprecated
func (runner *ComputationRunner) Stop() {
	if runner.stopChannel != nil {
		runner.stopChannel <- true
		<-runner.stopChannel
		runner = &ComputationRunner{}
	}
}

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
			runner.stopChannel<- true
			return
		}
	}
}

func (runner *ComputationRunner) runOneComputation() {
	runner.iterations++
	if runner.resultQueueHolder.isFull() {
		time.Sleep(time.Millisecond * 5)
		return
	}
	// the 'runners' channel is used to ensure that only 'runner.threads' number of goroutines are
	// concurrently running - thus providing direct control over thread utilization. Based on:
	// https://gist.github.com/AntoineAugusti/80e99edfe205baf7a094
	runners := make(chan bool, runner.threads)
	for i := 0; i < runner.threads; i++ {
		runners <- true
	}
	bodies := 0
	for item := range runner.bodyQueue.IterBuffered(){
		<-runners
		body := item.Val
		body.(SimBody).ForceComputer(runner.bodyQueue, runners)
		bodies++
	}
	if bodies == 0 {
		// no bodies in the sim - don't peg the CPU
		time.Sleep(5 * time.Millisecond)
		return
	}
	rq, _ := runner.resultQueueHolder.newQueue(bodies)
	for item := range runner.bodyQueue.IterBuffered() {
		body := item.Val.(SimBody)
		ri := body.Update(runner.timeScaling)
		rq.addRenderInfo(ri)
		if !body.Exists() {
			runner.bodyQueue.Remove(body.Id())
		}
	}
	rq.computed = true
	runner.computations++
}
