package runner

import (
	"container/list"
	"nbodygo/cmd/interfaces"
	"sync"
)

// TODO refactor to use channels to regulate queue size

//
// A result queue holder holds a fixed size queue of result queues allowing the computation
// runner to slighly outrun the render engine
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
// Sets a result queue to computed
//
func (rq *ResultQueue) SetComputed() {
	rq.computed = true
}

//
// Returns a ref to a queue in a result queue
//
func (rq *ResultQueue) Queue() []interfaces.Renderable {
	return rq.queue
}

//
// Creates and returns a new result queue
//
func newResultQueue(queNum uint, capacity int) *ResultQueue  {
	return &ResultQueue{
		computed: false,
		queNum:   queNum,
		queue:    make([]interfaces.Renderable, 0),
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
func (rqh *ResultQueueHolder) NewResultQueue(capacity int) (*ResultQueue, bool) {
	rqh.m.Lock()
	defer rqh.m.Unlock()
	if rqh.IsFull() {
		return nil, false
	}
	rq := newResultQueue(rqh.nextQueueNum(), capacity)
	rqh.queues.PushFront(rq)
	return rq, true
}

//
// Checks to see if the holder is full. Caller must synchronize.
//
// return: true of the holder is full, else false
//
func (rqh *ResultQueueHolder) IsFull() bool {
	return rqh.queues.Len() >= rqh.maxQueues
}

//
// Returns a computed queue in FIFO order, or second arg = false if no computed queues, else true
//
func (rqh *ResultQueueHolder) NextComputedQueue() (*ResultQueue, bool) {
	rqh.m.Lock()
	defer rqh.m.Unlock()
	rq := rqh.queues.Back()
	if rq != nil && rq.Value.(*ResultQueue).computed {
		rqh.queues.Remove(rq)
		return rq.Value.(*ResultQueue), true
	}
	return &ResultQueue{}, false
}

//
// Generates a 1-up number (supports test/debug). No concurrency guard because only
// called by the computation runner
//
func (rqh *ResultQueueHolder) nextQueueNum() (queueNum uint) {
	queueNum = rqh.queueNum
	rqh.queueNum++
	return
}