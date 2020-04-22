package collection

import (
	"container/list"
	"nbodygo/cmd/interfaces"
	"runtime"
	"sync"
)

//
// The collection state
//
type simBodyCollection struct {
	// this is the body array that all goroutines will iterate
	arr    []interfaces.SimBody
	// holds adds until they can be copied into the array above
	list   *list.List
	// concurrency
	lock   sync.Mutex
	// supports concurrent adds
	addCh  chan interfaces.SimBody
}

//
// Creates a new collection struct, initializing it with the passed array of bodies, which it
// makes a copy of.
//
// returns: the struct
//
func NewSimBodyCollection(bodies []interfaces.SimBody) interfaces.SimBodyCollection {
	c := simBodyCollection{
		arr:    make([]interfaces.SimBody, len(bodies)),
		list:   list.New(),
		lock:   sync.Mutex{},
		addCh:  make(chan interfaces.SimBody, 500),
	}
	for i, size := 0, len(bodies); i < size; i++ {
		c.arr[i] = bodies[i]
	}
	go c.handleAdds()
	return &c
}

//
// Supports concurrent (deferred) adds. Adds are sent to the 'handleAdds' function
// through the 'addCh' channel.
//
func (sbc *simBodyCollection) Add(b interfaces.SimBody) {
	sbc.addCh<- b
}

//
// Go routine that supports concurrent (deferred) adds. Receives adds from the 'Add' function
// through the 'addCh' channel. Adds added bodies to an internal list, which is lazily added into the body
// array by the 'Cycle' function.
//
func (sbc *simBodyCollection) handleAdds() {
	for {
		select {
		default:
			runtime.Gosched()
		case sb:= <-sbc.addCh:
			sbc.lock.Lock()
			sbc.list.PushFront(sb)
			sbc.lock.Unlock()
		}
	}
}

//
// Iterator with a callback pattern. Since there is a lot of iteration, it seemed to make sense to
// encapsulate the iterator with the consumer as a callback. That way if there ever needs to be something
// unique about the iteration it can be hidden here and iterators don't need to be concerned with it
//
func (sbc *simBodyCollection) IterateOnce(c interfaces.IterationConsumer) {
	for i, size := 0, len(sbc.arr); i < size; i++ {
		c(sbc.arr[i])
	}
}

//
// Gets the number of bodies in the array
//
func (sbc *simBodyCollection) Count() int {
	return len(sbc.arr)
}

//
// For gRPC server since the collection is not thread safe: provide a copy of the array. This is
// not efficient but traversals by the gRPC interface are extremely infrequent. The computation runner
// doesn't need this because it orchestrates the state change of the collection - it's the only entity that
// calls 'Cycle' and it only does this when it knows that there are no goroutines computing force. But the
// gRPC server is event-driven and could request to iterate the body array at any time. Since this locks, it
// will slow the computation runner but - the gRPC interface is not intended to be frequently used for
// traversing the body array
//
func (sbc *simBodyCollection) GetArrayCopy() *[]interfaces.SimBody {
	sbc.lock.Lock()
	defer sbc.lock.Unlock()
	toReturn := make([]interfaces.SimBody, len(sbc.arr))
	for i, size := 0, len(sbc.arr); i < size; i++ {
		toReturn[i] = sbc.arr[i]
	}
	return &toReturn
}

//
// Called by computation runner. Handles adds and deletes as follows: Deletes are cases where a body is destroyed
// and so its 'exists' flag is set to false. If bodies are destroyed then this function copies the body array and
// excludes deletes from the copy. Adds are appended to the body array if there are any in the list populated
// by the 'handleAdds' function.
//
// 99.999% of the time - there are no adds or deletes and so this function tries to do as little as possible
// to support concurrent read-only iteration over an infrequently changing array
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
		// bodies were set to not exist so implement delete by copying/compacting the array
		cnt += sbc.list.Len()
		arr := make([]interfaces.SimBody, cnt)
		j := 0
		for i, size := 0, len(sbc.arr); i < size; i++ {
			if sbc.arr[i].Exists() {
				arr[j] = sbc.arr[i]
				j++
			}
		}
		for e := sbc.list.Front(); e != nil; e = e.Next() {
			arr[j] = e.Value.(interfaces.SimBody)
			j++
		}
		sbc.arr = arr
		sbc.list.Init()
	} else if sbc.list.Len() > 0 {
		// only adds - append may be inefficient but adds don't happen that often
		for e := sbc.list.Front(); e != nil; e = e.Next() {
			sbc.arr = append(sbc.arr, e.Value.(interfaces.SimBody))
		}
		sbc.list.Init()
	}
}


