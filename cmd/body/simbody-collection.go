package body

import (
	"container/list"
	"runtime"
	"sync"
)

//
// The collection state
//
type simBodyCollection struct {
	// this is the body array that all goroutines will iterate
	arr    []SimBody
	// a list of concurrent adds, as well as collisions accumulated during each compute cycle
	events *list.List
	// concurrency
	lock   sync.Mutex
	// handle add/modify body events
	evCh  chan Event
	// diagnostic/debugging aid
	cycle  int
}

//
// Creates a new collection struct, initializing it with the passed array of bodies, which it
// makes a copy of.
//
// returns: the struct
//
func NewSimBodyCollection(bodies []SimBody) SimBodyCollection {
	c := simBodyCollection{
		arr:    make([]SimBody, len(bodies)),
		events: list.New(),
		lock:   sync.Mutex{},
		evCh:   make(chan Event, 5000), // todo factor of body count
	}
	for i, size := 0, len(bodies); i < size; i++ {
		c.arr[i] = bodies[i]
	}
	go c.handleEvents()
	return &c
}

//
// supports post-processing events that would cause race conditions or - would require synchronization - if
// done concurrently. Synchronization in the tight nested body computation loop has a prohibitive impact
// on performance
//
func (sbc *simBodyCollection) Enqueue(ev Event) {
	sbc.evCh<- ev
}

//
// Go routine that supports concurrent (deferred) adds and modifications to body state. Receives events
// from the 'Enqueue' function through the 'evCh' channel. Adds events to an internal list, which is handled
// by a call to the 'ProcessMods' function. (See computation runner.)
//
func (sbc *simBodyCollection) handleEvents() {
	for {
		select {
		default:
			runtime.Gosched()
		case ev:= <-sbc.evCh:
			sbc.lock.Lock()
			sbc.events.PushFront(ev)
			sbc.lock.Unlock()
		}
	}
}

//
// Iterator with a callback pattern. Since there is a lot of iteration, it seemed to make sense to
// encapsulate the iterator with the consumer as a callback. That way if there ever needs to be something
// unique about the iteration it can be hidden here and iterators don't need to be concerned with it
//
func (sbc *simBodyCollection) IterateOnce(c IterationConsumer) {
	for i, size := 0, len(sbc.arr); i < size; i++ {
		c(sbc.arr[i])
	}
}

//
// Gets the number of bodies in the array
//
func (sbc *simBodyCollection) Count() int {
	sbc.lock.Lock()
	defer sbc.lock.Unlock()
	return len(sbc.arr)
}

//
// For gRPC server since the collection is not thread safe: provide a copy of the array. This is
// not efficient but traversals by the gRPC interface are extremely infrequent. The computation runner
// doesn't need this because it orchestrates the state change of the collection - it is the only entity that
// calls 'Cycle' and it only does this when it knows that there are no goroutines computing force. But the
// gRPC server is event-driven and could request to iterate the body array at any time. Since this locks, it
// will slow the computation runner but - the gRPC interface is not intended to be frequently used for
// traversing the body array
//
func (sbc *simBodyCollection) GetArrayCopy() *[]SimBody {
	sbc.lock.Lock()
	defer sbc.lock.Unlock()
	arrCopy := make([]SimBody, len(sbc.arr))
	for i, size := 0, len(sbc.arr); i < size; i++ {
		arrCopy[i] = sbc.arr[i]
	}
	return &arrCopy
}

//
// Walks the internal 'events' list and processes all enqueued events. These are events that require
// changing body state in such a way that would require synchronization to avoid race conditions.
//
func (sbc *simBodyCollection) ProcessMods() {
	sbc.lock.Lock()
	if sbc.events.Len() == 0 {
		sbc.lock.Unlock()
		return
	}
	evs := []Event{}
	var prev *list.Element = nil
	for e := sbc.events.Front(); e != nil; e = e.Next() {
		if prev != nil {
			sbc.events.Remove(prev)
		}
		if e.Value.(Event).evType != AddEvent {
			evs = append(evs, e.Value.(Event))
			prev = e
		}
	}
	sbc.lock.Unlock()
	for i, len := 0, len(evs); i < len; i++ {
		evs[i].Handle()
	}
}

//
// Returns the count of events in the internal 'events' list that are Adds
//
func (sbc *simBodyCollection) countAdds() int {
	cnt := 0
	for e := sbc.events.Front(); e != nil; e = e.Next() {
		if e.Value.(Event).evType == AddEvent {
			cnt++
		}
	}
	return cnt
}

//
// Called by computation runner to prepare the body collection for another compute cycle. Removes refs
// to bodies that have been set not to exist, and resolves collisions and fragmentations which have to be
// done in a single thread to avoid race conditions
//
func (sbc *simBodyCollection) Cycle() {
	cnt := 0
	for i, size := 0, len(sbc.arr); i < size; i++ {
		if sbc.arr[i].Exists() {
			cnt++
		}
	}
	sbc.lock.Lock() // prevents adds
	defer sbc.lock.Unlock()
	if cnt < len(sbc.arr) {
		cnt += sbc.countAdds()
		// bodies were set to not exist so implement delete by copying/compacting the array
		arr := make([]SimBody, cnt)
		j := 0
		for i, size := 0, len(sbc.arr); i < size; i++ {
			if sbc.arr[i].Exists() {
				arr[j] = sbc.arr[i]
				j++
			}
		}
		for e := sbc.events.Front(); e != nil; e = e.Next() {
			if e.Value.(Event).evType == AddEvent {
				arr[j] = e.Value.(Event).GetAdd()
				j++
			}
		}
		sbc.arr = arr
		sbc.events.Init()
	} else {
		cnt = sbc.countAdds()
		if cnt > 0 {
			for e := sbc.events.Front(); e != nil; e = e.Next() {
				if e.Value.(Event).evType == AddEvent {
					sbc.arr = append(sbc.arr, e.Value.(Event).GetAdd())
				}
			}
		}
		sbc.events.Init()
	}
	sbc.cycle++
}


