package runner

import (
	"nbodygo/cmd/renderable"
)

//
// A result queue holder holds a fixed size queue of result queues allowing the computation
// runner to slightly outrun the render engine
//
type ResultQueueHolder struct {
	maxQueues int
	queueNum  uint
	ch chan *ResultQueue
}

//
// A result queue holds an array of renderable objects - each being the result of a computation
// cycle with updated position, etc.
//
type ResultQueue struct {
	computed bool
	queNum uint
	queue []renderable.Renderable
}

//
// Returns a ref to a queue in a result queue
//
func (rq *ResultQueue) Queue() []renderable.Renderable {
	return rq.queue
}

//
// Creates and returns a new result queue
//
func newResultQueue(queNum uint) *ResultQueue  {
	return &ResultQueue{
		computed: false,
		queNum:   queNum,
		queue:    make([]renderable.Renderable, 0),
	}
}

//
// Adds the passed renderable to the queue
//
func (rq *ResultQueue) AddRenderable(info renderable.Renderable) {
	rq.queue = append(rq.queue, info)
}

//
// Sets the passed queue as computed and adds it to the result queue holder. As a result,
// the next all to 'NextComputedQueue' can return it (eventually, if other computed queues
// are ahead of it)
//
func (rqh *ResultQueueHolder) SetComputed(queue *ResultQueue) {
	queue.computed = true
	rqh.ch <- queue
}

//
// Initializes a new result queue holder with capacity = 'maxQueues'
//
func NewResultQueueHolder(maxQueues int) ResultQueueHolder {
	return ResultQueueHolder {
		maxQueues: maxQueues,
		ch: make(chan *ResultQueue, maxQueues),
	}
}

//
// Creates and returns a new queue if there is capacity
//
// return: if there is capacity, returns a pointer to queue and true. If no
// capacity, returns nil and false
//
func (rqh *ResultQueueHolder) NewResultQueue() (*ResultQueue, bool) {
	if len(rqh.ch) == cap(rqh.ch) {
		return nil, false
	}
	queueNum := rqh.queueNum
	rqh.queueNum++
	return newResultQueue(queueNum), true
}

//
// Returns a computed queue in FIFO order
//
// return: if a computed queue is available, returns a pointer to queue and true. If no
// queues are available, returns nil and false
//
func (rqh *ResultQueueHolder) NextComputedQueue() (*ResultQueue, bool) {
	select {
	case queue := <-rqh.ch:
		return queue, true
	default:
		return nil, false
	}
}

//
// returns the max number of queues supported by the result queue holder
//
func (rqh *ResultQueueHolder) MaxQueues() int {
	return rqh.maxQueues
}
