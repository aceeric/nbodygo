package runner

import (
	"container/list"
	"fmt"
	"nbodygo/cmd/bodyrender"
	"sync"
)

// TODO refactor to use channels to regulate queue size

type ResultQueueHolder struct {
	maxQueues int
	queueNum  uint
	queues *list.List // queue of ResultQueue
	m sync.Mutex
}

type ResultQueue struct {
	computed bool
	queNum uint
	queue []bodyrender.Renderable // TODO RENAME - ITS NOT A QUEUE IS IT? JUST USED AS AN ARRAY
}

func (rq *ResultQueue) SetComputed() {
	rq.computed = true
}

func (rq *ResultQueue) Queue() []bodyrender.Renderable {
	return rq.queue
}

func NewResultQueue(queNum uint, capacity int) *ResultQueue  {
	return &ResultQueue{
		computed: false,
		queNum:   queNum,
		queue:    make([]bodyrender.Renderable, 0),
	}
}

func (rq *ResultQueue) AddRenderable(info bodyrender.Renderable) {
	rq.queue = append(rq.queue, info) // TODO HOW TO DO THIS MORE GO-LIKE?
}

func NewResultQueueHolder(maxQueues int) ResultQueueHolder {
	return ResultQueueHolder {
		maxQueues: maxQueues,
		queues: list.New(),
		m: sync.Mutex{},
	}
}

// return second arg false if full and can't add
func (rqh *ResultQueueHolder) NewQueue(capacity int) (*ResultQueue, bool) {
	if rqh.queues == nil { // TODO DELETEME
		//println("NewQueue: rqh.queues == nil")
	}
	rqh.m.Lock()
	defer rqh.m.Unlock()
	if rqh.IsFull() {
		return &ResultQueue{}, false
	}
	rq := NewResultQueue(rqh.nextQueueNum(), capacity)
	rqh.queues.PushFront(rq)
	return rq, true
}

// caller must synchronize
func (rqh *ResultQueueHolder) IsFull() bool {
	if rqh.queues == nil { // TODO DELETEME
		//println("IsFull: rqh.queues == nil")
	}
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("r: %+v -- rqh: %+v\n", r, rqh)
		}
	}()
	return rqh.queues.Len() >= rqh.maxQueues
}

// return second arg false if no computed queues
func (rqh *ResultQueueHolder) NextComputedQueue() (*ResultQueue, bool) {
	if rqh.queues == nil { // TODO DELETEME
		//println("NextComputedQueue: rqh.queues == nil")
	}
	rqh.m.Lock()
	defer rqh.m.Unlock()
	rq := rqh.queues.Back()
	if rq != nil && rq.Value.(*ResultQueue).computed {
		rqh.queues.Remove(rq)
		return rq.Value.(*ResultQueue), true
	}
	return &ResultQueue{}, false
}

// generate 1-up number (supports test/debug)
func (rqh *ResultQueueHolder) nextQueueNum() (queueNum uint) {
	queueNum = rqh.queueNum
	rqh.queueNum++
	return
}