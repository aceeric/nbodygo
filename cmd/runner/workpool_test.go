package runner

import (
	"sync"
	"testing"
	"time"
)

// basic body computer test
func TestBodyComputer(t *testing.T) {
	w := Worker{
		killChan:    make(chan bool),
		computation: make(chan Computation),
		cpu:         0,
	}
	wg := sync.WaitGroup{}
	wg.Add(1)
	go bodyComputer(w, &wg, nil)
	w.computation<- Computation{
		body:      nil,
		bodyQueue: nil,
	}
	time.Sleep(time.Millisecond * 1000)
	w.killChan<- true

}

// basic work pool test
func TestWorkPool(t *testing.T) {
	wp := NewWorkPool(5, nil)
	for i := 0; i < 20; i++ {
		wp.submit(nil, nil)
	}
	for i := 0; i < 20; i++ {
		wp.take()
	}
}

// tests the work pool with a custom function that increments a value by reading from a channel written to
// by a custom work function. As a result, the only blocking or throttling would be from that associated with the pool
// itself
func TestWorkPoolCustomFun(t *testing.T) {
	testVal := 0
	const workers = 20000
	workChan := make(chan int, workers * 1.1) // no blocking on the channel
	wp := NewWorkPool(20, func(*Computation) {
		workChan<- 1
	})
	go func() {
		for {
			i := <-workChan
			if i == 1 {
				testVal++
			} else {
				workChan <- 0
				// kind  of hokey but use the value 2 (sent below) to know we're done and send something to ack
				return
			}
		}
	}()
	for i := 0; i < workers; i++ {
		//t.Logf("submitting %v\n", i)
		wp.submit(nil, nil)
	}
	for i := 0; i < workers; i++ {
		//t.Logf("taking %v\n", i)
		wp.take()
	}
	workChan<- 2 // all works taken
	<-workChan // wait for all workers to complete
	if testVal != workers {
		t.Errorf("Worker pool calc failed. Expected: %v. Got actual: %v\n", workers, testVal)
	}
}
