package cmap

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type TestBod struct {
	id int
}

type ITestBod interface {
	iterateOverMap(*testing.T, *ConcurrentMap, chan<- bool)
}

const bodyCount = 1000

func (tb *TestBod) iterateOverMap(t *testing.T, bodyQueue * ConcurrentMap, result chan<- bool) {
	count := 0
	for item := range bodyQueue.IterBuffered() {
		count++
		body := item.Val.(*TestBod)
		if body.id == -1 {
			//println("this will never happen")
		}
	}
	if count != bodyCount {
		t.Error("Did not iterate over all items")
	}
	result <- true
}

//
// Tests nested concurrent iteration. The computation runner in the runner package iterates over this map
// implementation and - as it is iterating - it feeds concurrent goroutines which in turn iterate the map all
// concurrently. This test exercises that flow to make sure that all these concurrent iterations are handled
// correctly. Note that this test does not worry about CPU/thread affinity for the concurrent goroutines and
// as a result I see very high utilization across all CPUs which, according to my reading, is caused by
// the Go runtime creating and destroying threads across all CPUs which is very inefficient. However for this
// test, we don't care about that we are only concerned with the map implementation's handling of
// concurrent iteration.
//
func TestNestedConcurrentIteration(t *testing.T) {
	bodyQueue := New()
	for bodyId := 0; bodyId < bodyCount; bodyId++ {
		bodyQueue.Set(bodyId, &TestBod{bodyId})
	}
	const concurrentGoroutines = 5
	runners := make(chan bool, concurrentGoroutines)
	for i := 0; i < concurrentGoroutines; i++ {
		runners <- true
	}
	myGoroutines := int32(0)
	go func() {
		for {
			//fmt.Printf("my goroutines: %v\n", myGoroutines)
			time.Sleep(time.Millisecond * 15)
		}
	}()
	stopped := make(chan bool)
	start := time.Now()
	// iterate over the map, repeatedly - top level
	var wg sync.WaitGroup
	go func() {
		for {
			for item := range bodyQueue.IterBuffered() {
				<-runners // throttles concurrent goroutines by thread count
				wg.Add(1)
				body := item.Val.(ITestBod)
				// each element also iterates over the map
				atomic.AddInt32(&myGoroutines, 1)
				go func() {
					// iterate over the map concurrently
					body.iterateOverMap(t, &bodyQueue, runners)
					wg.Done()
					atomic.AddInt32(&myGoroutines, -1)
				}()
			}
			wg.Wait()
			duration := time.Now().Sub(start)
			if duration.Seconds() >= 30 {
				stopped <- true
				return
			}
		}
	}()
	<-stopped
}

