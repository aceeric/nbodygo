package runner

import (
	"fmt"
	"nbodygo/cmd/interfaces"
	"runtime"
	"time"
)

type ComputationRunner struct {
	stop                     chan bool
	workerCnt                int
	iterations, computations uint
	submits, waits           uint
	submitMillis             int64
	waitMillis               int64
	wp                       *WorkPool
	cc                       interfaces.SimBodyCollection
	maxIteration             int
	running                  bool
	startTime                time.Time
	stopTime                 time.Time
	goroutines               int
	timeScaling              float64
	resultQueueHolder        ResultQueueHolder
}

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

func NewComputationRunner(workerCnt int, timeScaling float64, resultQueueHolder ResultQueueHolder,
	cc interfaces.SimBodyCollection) *ComputationRunner {
	r := ComputationRunner{
		stop:              make(chan bool),
		workerCnt:         workerCnt,
		wp:                NewWorkPool(workerCnt, cc),
		cc:                cc,
		timeScaling:       timeScaling,
		resultQueueHolder: resultQueueHolder,
	}
	return &r
}

// supports debugging
func (r *ComputationRunner) SetMaxIterations(maxIteration int) *ComputationRunner {
	r.maxIteration = maxIteration
	return r
}

func (r *ComputationRunner) Start() *ComputationRunner {
	go r.run()
	return r
}

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

func (r *ComputationRunner) runOneComputation() {
	r.iterations++
	if r.resultQueueHolder.IsFull() {
		return
	}
	start := time.Now()
	bodies := 0
	r.cc.IterateOnce(func(c interfaces.SimBody) {
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

	rq, _ := r.resultQueueHolder.NewQueue(bodies)
	r.cc.IterateOnce(func(c interfaces.SimBody) {
		ri := c.Update(r.timeScaling)
		rq.AddRenderable(ri)
	})
	rq.computed = true
	r.cc.Cycle()
	r.computations++
	r.goroutines += runtime.NumGoroutine()
}
