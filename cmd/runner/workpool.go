package runner

import (
	"fmt"
	"nbodygo/cmd/interfaces"
	"sync"
	"time"
)

type WorkPool struct {
	wrkIdx      uint
	workers     []*Worker
	wg          sync.WaitGroup
	submissions int64
	millis      int64
	cc          interfaces.SimBodyCollection
}

type Worker struct {
	id          uint
	stop        chan bool
	compute     chan interfaces.SimBody
	invocations int
	millis      int64
}

func worker(w *Worker, wg *sync.WaitGroup, cc interfaces.SimBodyCollection) {
	millis := int64(0)
	invocations := 0
	for {
		select {
		case <-w.stop:
			// note that this may leave the go routine with items still enqueued on the w.compute channel
			// so this shutdown leaves unfinished work
			w.invocations = invocations
			w.millis = millis
			w.stop<- true
			return
		case c := <-w.compute:
			start := time.Now()
			c.ForceComputer(cc)
			invocations++
			wg.Done()
			millis += time.Now().Sub(start).Milliseconds()
		default:
		}
	}
}

func NewWorkPool(goroutines int, cc interfaces.SimBodyCollection) *WorkPool {
	wp := WorkPool{
		wrkIdx:      0,
		workers:     []*Worker{},
		wg:          sync.WaitGroup{},
		submissions: 0,
		millis:      0,
		cc:          cc,
	}
	for i := 0; i < goroutines; i++ {
		w := Worker{
			id:          uint(i),
			stop:        make(chan bool),
			compute:     make(chan interfaces.SimBody, 5), // TODO buffer irrelevant?
			invocations: 0,
			millis:      0,
		}
		wp.workers = append(wp.workers, &w)
		go worker(&w, &wp.wg, cc)
	}
	return &wp
}

func (wp *WorkPool) ShutDownWorkPool() {
	for _, w := range wp.workers {
		w.stop <- true
		<-w.stop
	}
}

func (wp *WorkPool) PrintStats() {
	fmt.Printf("Worker Pool\n submissions: %v\n millis: %v\n millis/submission: %v\n", wp.submissions,
		wp.millis, float32(wp.millis)/float32(wp.submissions))
	for _, w := range wp.workers {
		fmt.Printf("> Worker id: %v invocations: %v millis: %v millis/invocation: %v\n", w.id,
			w.invocations, w.millis, float32(w.millis)/float32(w.invocations))
	}
}

func (wp *WorkPool) submit(c interfaces.SimBody) {
	start := time.Now()
	wp.wg.Add(1)
	wp.workers[wp.wrkIdx%uint(len(wp.workers))].compute <- c
	wp.submissions++
	wp.wrkIdx++
	wp.millis += time.Now().Sub(start).Milliseconds()
}

func (wp *WorkPool) wait() {
	wp.wg.Wait()
}
