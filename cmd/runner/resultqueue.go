package runner

import (
	"nbodygo/cmd/body"
)

//
// Implements a holder of queues. Each queue holds the information needed to render a body in the
// graphics engine. The holder allows computation to outrun rendering by a fixed amount. The holder
// provides queues in FIFO order, and can also be concurrently resized. The resize functionality depends
// on a channel which is continually read/written containing the index into the active queue. On resize, a
// new queue is created, the existing queue contents are copied, and the the new queue is activated
// by writing its index into the selector channel. So the active queue is either at index zero or
// one. This approach also synchronizes calls into the queue so the resize can be done in a
// thread-safe manner. The design only supports one thread adding queues into the holder - with
// no limit on consumers.
//

//
// A result queue holds an array of objects with information needed to render them in the graphics
// engine
//
type ResultQueue struct {
	queueNum uint
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
	selector chan int
	queues   [2]struct {
		ch        chan *ResultQueue
		maxQueues int
	}
	queueNum  uint
	resizable bool
}

//
// Adds the passed queue to the queue holder
//
func (rqh *ResultQueueHolder) Add(queue *ResultQueue) {
	if !rqh.resizable {
		rqh.queues[0].ch <- queue
		return
	}
	q := <-rqh.selector
	if len(rqh.queues[q].ch) >= rqh.queues[q].maxQueues {
		panic("No queue capacity")
	}
	rqh.queues[q].ch <- queue
	rqh.selector <- q
}

//
// Initializes a new result queue holder with capacity = 'maxQueues', and resizable based on the
// resizable arg
//
func NewResultQueueHolder(maxQueues int, resizable bool) *ResultQueueHolder {
	rqh := ResultQueueHolder{
		selector: make(chan int, 1),
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
		resizable: resizable,
	}
	rqh.selector <- 0
	return &rqh
}

//
// Creates and returns a new queue if there is capacity
//
// return: if there is capacity, returns a pointer to a new queue and true. If no
// capacity, returns nil and false
//
func (rqh *ResultQueueHolder) NewResultQueue() (*ResultQueue, bool) {
	if !rqh.resizable {
		if len(rqh.queues[0].ch) >= rqh.queues[0].maxQueues {
			return nil, false
		}
	} else {
		q := <-rqh.selector
		l := len(rqh.queues[q].ch)
		max := rqh.queues[q].maxQueues
		rqh.selector <- q
		if l >= max {
			return nil, false
		}
	}
	queueNum := rqh.queueNum
	rqh.queueNum++
	return &ResultQueue{
		queueNum: queueNum,
		queue:    make([]*body.Renderable, 0),
	}, true

}

//
// return: if a queue is available, returns a pointer to queue and true. If no
// queues are available, returns nil and false
//
func (rqh *ResultQueueHolder) Next() (*ResultQueue, bool) {
	if !rqh.resizable {
		select {
		case queue := <-rqh.queues[0].ch:
			return queue, true
		default:
			return nil, false
		}
	}
	q := <-rqh.selector
	select {
	case queue := <-rqh.queues[q].ch:
		rqh.selector <- q
		return queue, true
	default:
		rqh.selector <- q
		return nil, false
	}
}

//
// return: the max number of queues in return value one supported by the result queue holder,
// and the current queue length in return value two
//
func (rqh *ResultQueueHolder) MaxQueues() (int, int) {
	if !rqh.resizable {
		return rqh.queues[0].maxQueues, len(rqh.queues[0].ch)
	}
	q := <-rqh.selector
	max := rqh.queues[q].maxQueues
	ln := len(rqh.queues[q].ch)
	rqh.selector <- q
	return max, ln
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
// computation runner asks for a new queue from the holder - and gets one - so there is capacity
// resize event concurrently resizes the queue down, while leaving extra space as described
// runner adds the queue to the holder - the add succeeds because of the extra space
//
// This only works because there is only one thread adding to the queue. It wouldn't work with more
// than one thread adding. But the design of this holder is not intended to be general purpose - it is
// specifically tailored to the requirements of the n-body simulation
//
func (rqh *ResultQueueHolder) Resize(maxQueues int) bool {
	if !rqh.resizable {
		return false
	}
	q := <-rqh.selector
	if maxQueues == rqh.queues[q].maxQueues {
		// don't resize to the same size
		rqh.selector <- q
		return false
	}
	sizeToUse := maxQueues
	curLen := len(rqh.queues[q].ch)
	if maxQueues < curLen {
		// need to preserve current queue contents. Since a concurrent call to NewResultQueue may have
		// already reported that it is ok to add a queue, if we shrink the queue that could fail the caller's
		// subsequent concurrent call to 'Add'. So preserve physical queue size plus space to add, but
		// set logical queue size to match the caller's stipulation
		sizeToUse = rqh.queues[q].maxQueues + 1
	}
	rqh.queues[q^1] = struct {
		ch        chan *ResultQueue
		maxQueues int
	}{
		ch:        make(chan *ResultQueue, sizeToUse), // could be bigger than maxQueues per above
		maxQueues: maxQueues,
	}
	for ; len(rqh.queues[q].ch) != 0; {
		// transfer queue contents
		rq := <-rqh.queues[q].ch
		rqh.queues[q^1].ch <- rq
	}
	rqh.selector <- q ^ 1 // enable new queue
	rqh.queues[q] = struct {
		ch        chan *ResultQueue
		maxQueues int
	}{
		ch:        nil,
		maxQueues: 0,
	}
	return true
}
