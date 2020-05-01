package runner

import (
	"log"
	"nbodygo/cmd/body"
	"nbodygo/cmd/instrumentation"
	"sync"
)

//
// Implements a holder of queues. Each queue holds the information needed to render a body in the
// graphics engine. The holder allows computation to outrun rendering by a fixed amount. The holder
// provides queues in FIFO order, and can also be concurrently resized. On resize, a new queue is
// created, the existing queue contents are copied, and then the the new queue is activated. The
// design only supports one thread adding queues into the holder - with no limit on consumers.
// Note: this would be much simpler to implement if Go channels were re-sizable because then the
// holder could simply be a channel. But - Go channels are not resizable, at least as of 1.14
//

//
// A result queue holds an array of renderable objects
// engine
//
type ResultQueue struct {
	QueueNum uint
	queue    []*body.Renderable
}

//
// Returns a ref to a queue in a result queue
//
func (rq *ResultQueue) Queue() []*body.Renderable {
	return rq.queue
}

//
// Adds the passed item to the queue
//
func (rq *ResultQueue) Add(info *body.Renderable) {
	rq.queue = append(rq.queue, info)
}

//
// A result queue holder holds a fixed size queue of result queues allowing the computation
// runner to slightly outrun the render engine
//
type ResultQueueHolder struct {
	selector int
	lock sync.RWMutex
	queues   [2]struct {
		ch        chan *ResultQueue
		maxQueues int
	}
	queueNum  uint
}

//
// Adds the passed queue to the queue holder
//
func (rqh *ResultQueueHolder) Add(queue *ResultQueue) {
	rqh.lock.RLock()
	if len(rqh.queues[rqh.selector].ch) >= cap(rqh.queues[rqh.selector].ch) {
		log.Fatalf("No queue capacity selector=%v len=%v cap=%v max=%v\n", rqh.selector,
			len(rqh.queues[rqh.selector].ch), cap(rqh.queues[rqh.selector].ch), rqh.queues[rqh.selector].maxQueues)
	}
	rqh.queues[rqh.selector].ch <- queue
	rqh.lock.RUnlock()
}

//
// Initializes a new result queue holder with the passed capacity
//
func NewResultQueueHolder(maxQueues int) *ResultQueueHolder {
	instrumentation.MaxQueues.Set(float64(maxQueues))
	rqh := ResultQueueHolder{
		selector: 0,
		queues: [2]struct {
			ch        chan *ResultQueue
			maxQueues int
		}{
			{
				ch:        make(chan *ResultQueue, maxQueues),
				maxQueues: maxQueues,
			},
		},
		queueNum:  0,
	}
	return &rqh
}

//
// Creates and returns a new queue if there is capacity
//
// return: if there is capacity, returns a pointer to a new queue and true. If no
// capacity, returns nil and false
//
func (rqh *ResultQueueHolder) NewResultQueue() (*ResultQueue, bool) {
	rqh.lock.RLock()
	curLen := len(rqh.queues[rqh.selector].ch)
	max := rqh.queues[rqh.selector].maxQueues
	rqh.lock.RUnlock()
	if curLen >= max {
		return nil, false
	}
	queueNum := rqh.queueNum
	rqh.queueNum++
	instrumentation.CurQueues.Set(float64(curLen))
	return &ResultQueue{
		QueueNum: queueNum,
		queue:    make([]*body.Renderable, 0),
	}, true
}

//
// return: if a queue is available, returns a pointer to queue and true. If no
// queues are available, returns nil and false
//
func (rqh *ResultQueueHolder) Next() (*ResultQueue, bool) {
	rqh.lock.RLock()
	var queue *ResultQueue = nil
	var ok = false
	select {
	case queue = <-rqh.queues[rqh.selector].ch:
		ok = true
	default:
	}
	rqh.lock.RUnlock()
	return queue, ok
}

//
// returns: the max number of queues supported by the result queue holder
//
func (rqh *ResultQueueHolder) MaxQueues() int {
	rqh.lock.RLock()
	max := rqh.queues[rqh.selector].maxQueues
	rqh.lock.RUnlock()
	return max
}

//
// Resizes the queue holder. Note regarding concurrency: the computation runner checks to
// see if there is capacity in the holder before beginning a compute cycle. If there is, it gets
// a queue and runs the compute cycle. Upon completion of the compute cycle the runner then
// adds the result queue to the holder. This is important because the runner doesn't want to waste
// cpu if there is no queue capacity and - once a compute cycle finishes those results can't be
// lost without causing jumpiness in the rendering. So the resize functionality leaves extra space
// in the resized queue to accommodate this requirement. The use case is:
//
// 1 - computation runner asks for a new queue from the holder - and gets one - so there is capacity
// 2 - resize event concurrently resizes the queue down, but leaving extra space as described
// 3 - runner adds the queue to the holder - the add succeeds because of the extra space
//
// This only works because there is only one thread adding to the queue. It wouldn't work with more
// than one thread adding. But the design of this holder is not intended to be general purpose - it is
// specifically tailored to the requirements of the n-body simulation
//
func (rqh *ResultQueueHolder) Resize(maxQueues int) bool {
	// ok to use selector unguarded because only one thread performs resize
	if maxQueues == rqh.queues[rqh.selector].maxQueues {
		// don't resize to the same size
		return false
	}
	sizeToUse := maxQueues
	rqh.lock.Lock()
	curLen := len(rqh.queues[rqh.selector].ch)
	if maxQueues < curLen {
		// need to preserve current queue contents. Since a concurrent call to NewResultQueue may have
		// already reported that it is ok to add a queue, if we shrink the queue that could fail the caller's
		// subsequent (blocked) call to 'Add'. So preserve physical queue size plus space to add, but
		// set logical queue size to match the caller's stipulation
		sizeToUse = rqh.queues[rqh.selector].maxQueues + 1
	}
	rqh.queues[rqh.selector^1] = struct {
		ch        chan *ResultQueue
		maxQueues int
	}{
		ch:        make(chan *ResultQueue, sizeToUse), // could be bigger than maxQueues per above
		maxQueues: maxQueues,
	}
	for ; len(rqh.queues[rqh.selector].ch) != 0; {
		// transfer queue contents
		rq := <-rqh.queues[rqh.selector].ch
		rqh.queues[rqh.selector^1].ch <- rq
	}
	rqh.selector ^= 1 // enable new queue
	rqh.lock.Unlock()
	// enable old queue resoures to be garbage-collected
	rqh.queues[rqh.selector^1] = struct {
		ch        chan *ResultQueue
		maxQueues int
	}{
		ch:        nil,
		maxQueues: 0,
	}
	instrumentation.MaxQueues.Set(float64(maxQueues))
	return true
}
