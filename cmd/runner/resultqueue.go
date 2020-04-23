package runner

import (
	"nbodygo/cmd/renderable"
)

/*
//
// A result queue holder holds a fixed size queue of result queues allowing the computation
// runner to slightly outrun the render engine
//
type ResultQueueHolder struct {
	maxQueues int
	queueNum  uint
	queues *list.List // queue of ResultQueue
	m sync.Mutex
}

//
// A result queue holds an array of renderable objects
//
type ResultQueue struct {
	computed bool
	queNum uint
	queue []interfaces.Renderable
}

//
// Returns a ref to a queue in a result queue
//
func (rq *ResultQueue) Queue() []interfaces.Renderable {
	return rq.queue
}

//
// Creates and returns a new result queue with passed capacity. Note that
// the underlying array might not actually be filled and so some of the pointers in the
// array could be null. Caller has to handle that
//
func newResultQueue(queNum uint, capacity int) *ResultQueue  {
	return &ResultQueue{
		computed: false,
		queNum:   queNum,
		queue:    make([]interfaces.Renderable, capacity),
	}
}

//
// Adds the passed renderable to the queue
//
func (rq *ResultQueue) AddRenderable(info interfaces.Renderable) {
	rq.queue = append(rq.queue, info)
}

//
// Initializes a new result queue holder with capacity = 'maxQueues'
//
func NewResultQueueHolder(maxQueues int) ResultQueueHolder {
	return ResultQueueHolder {
		maxQueues: maxQueues,
		queues: list.New(),
		m: sync.Mutex{},
	}
}

//
// Creates and returns a new queue
//
// return: second arg false if full and can't add, else true
//
func (rqh *ResultQueueHolder) NewResultQueue() (*ResultQueue, bool) {
	rqh.m.Lock()
	defer rqh.m.Unlock()
	if rqh.queues.Len() >= rqh.maxQueues {
		return nil, false
	}
	queueNum := rqh.queueNum
	rqh.queueNum++
	rq := newResultQueue(queueNum, 0)
	rqh.queues.PushFront(rq)
	return rq, true
}

//
// Returns a computed queue in FIFO order. Second arg = false if no computed queues, else true
//
func (rqh *ResultQueueHolder) NextComputedQueue() (*ResultQueue, bool) {
	rqh.m.Lock()
	defer rqh.m.Unlock()
	rq := rqh.queues.Back()
	if rq != nil && rq.Value.(*ResultQueue).computed {
		rqh.queues.Remove(rq)
		return rq.Value.(*ResultQueue), true
	}
	return nil, false
}

//
// Sets the last out queue to computed
//
func (rqh *ResultQueueHolder) SetComputed() {
	rqh.m.Lock()
	defer rqh.m.Unlock()
	rq := rqh.queues.Back()
	if rq != nil {
		rq.Value.(*ResultQueue).computed = true
	}
}
*/


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
// A result queue holds an array of renderable objects
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
// Adds the passed renderable to the queue
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
// Creates and returns a new queue
//
// return: second arg false if full and can't add, else true
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
// Returns a computed queue in FIFO order, or second arg = false if no computed queues, else true
//
func (rqh *ResultQueueHolder) NextComputedQueue() (*ResultQueue, bool) {
	select {
	case queue := <-rqh.ch:
		return queue, true
	default:
		return nil, false
	}
}

