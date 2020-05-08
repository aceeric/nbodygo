package runner

import (
	"runtime"
	"testing"
	"time"
)

var adds, addFails, gets, getFails, maxReads, resizes int
var killChan = make(chan bool, 4)
var ackChan = make(chan bool, 4)
var outOfOrder = false

//
// ResultQueueHolder concurrency test. Runs goroutines to interact with a rqh: add, get, resize, get metrics from.
// Verifies that the holder returns queues in proper order. If it does, then it handled the concurrency
// correctly.
//
func TestResultQueue(t *testing.T) {
	adds, gets, addFails, getFails, maxReads, resizes = 0, 0, 0, 0, 0, 0
	rqh := NewResultQueueHolder(5)
	go add(rqh)
	go get(rqh)
	go resize(rqh)
	go metric(rqh)
	time.Sleep(time.Second * 10)
	killChan <- true
	killChan <- true
	killChan <- true
	killChan <- true
	<-ackChan
	<-ackChan
	<-ackChan
	<-ackChan
	if outOfOrder {
		t.Errorf("RQ out of order. adds:%v | addFails:%v | gets:%v | getFails:%v | maxReads:%v | resizes:%v\n",
			adds, addFails, gets, getFails, maxReads, resizes)
	}
}

//
// Calls the holder 'MaxQueues' function to get the queue size
//
func metric(rqh *ResultQueueHolder) {
	for {
		select {
		case <-killChan:
			ackChan <- true
			return
		default:
			rqh.MaxQueues()
			time.Sleep(time.Millisecond * 15)
			maxReads++
		}
		runtime.Gosched()
	}
}

//
// Adds a queue to the holder
//
func add(rqh *ResultQueueHolder) {
	for {
		select {
		case <-killChan:
			ackChan <- true
			return
		default:
			rq, ok := rqh.NewResultQueue()
			if ok {
				adds++
				rqh.Add(rq)
			} else {
				addFails++
			}
		}
		runtime.Gosched()
	}
}

//
// Gets a queue from the holder - if one comes back out of order, sets the 'outOfOrder' var to true,
// which is checked by the caller
//
func get(rqh *ResultQueueHolder) {
	first := true
	lastQueNum := uint(0)
	for {
		select {
		case <-killChan:
			ackChan <- true
			return
		default:
			rq, ok := rqh.Next()
			if ok {
				gets++
				if first {
					first = false
				} else if rq.QueueNum != lastQueNum + 1 {
					outOfOrder = true
				}
				lastQueNum = rq.QueueNum
			} else {
				getFails ++
			}
		}
		runtime.Gosched()
	}
}

//
// Resizes the queue up and down
//
func resize(rqh *ResultQueueHolder) {
	newSize := 3
	for {
		select {
		case <-killChan:
			ackChan <- true
			return
		default:
			time.Sleep(time.Second * 2)
			rqh.Resize(newSize)
			if newSize == 3 {
				newSize = 8
			} else {
				newSize = 3
			}
			resizes++
		}
		runtime.Gosched()
	}
}