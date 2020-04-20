package collection

import (
	"container/list"
	"nbodygo/cmd/interfaces"
	"runtime"
	"sync"
)

type simBodyCollection struct {
	arr    []interfaces.SimBody
	list   *list.List
	lock   sync.Mutex
	addCh  chan interfaces.SimBody
}

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

// Supports concurrent (deferred). Adds are sent to the 'handleAdds' function
// through the 'addCh' channel.
func (cm *simBodyCollection) Add(b interfaces.SimBody) {
	cm.addCh<- b
}

// Supports concurrent (deferred). Receives Adds from the 'Add' function through the 'addCh' channel.
// Adds added bodies to a list, which is then added into the body array by the 'Cycle' function.
func (cm *simBodyCollection) handleAdds() {
	for {
		select {
		default:
			runtime.Gosched()
		case sb:= <-cm.addCh:
			cm.lock.Lock()
			cm.list.PushFront(sb)
			cm.lock.Unlock()
		}
	}
}

// Iterator with a callback pattern
func (cm *simBodyCollection) IterateOnce(callback func (item interfaces.SimBody)) {
	for i, size := 0, len(cm.arr); i < size; i++ {
		callback(cm.arr[i])
	}
}

// for gRPC server since the collection is not thread safe: provide a copy of the array. This is
// not efficient but traversals by the gRPC interface are extremely infrequent
func (cm *simBodyCollection) GetArrayCopy() *[]interfaces.SimBody {
	cm.lock.Lock()
	defer cm.lock.Unlock()
	toReturn := make([]interfaces.SimBody, len(cm.arr))
	for i, size := 0, len(cm.arr); i < size; i++ {
		toReturn[i] = cm.arr[i]
	}
	return &toReturn
}

// called by computation runner. Handles adds and deletes as follows: Deletes are cases where a body is destroyed
// and so its exists flag is set to false. If bodies are destroyed then the function copies the body array and
// excludes deletes from the copy. Adds are appended to the body array if there are any in the list populated
// by the 'handleAdds' function.
//
// 99.999% of the time - there are no adds or deletes and so this function tries to do as little
// as possible to support read-only iteration over an infrequently changing array
func (cm *simBodyCollection) Cycle() {
	cnt := 0
	for i, size := 0, len(cm.arr); i < size; i++ {
		if cm.arr[i].Exists() {
			cnt++
		}
	}
	cm.lock.Lock() // prevents adds
	defer cm.lock.Unlock()
	if cnt < len(cm.arr) {
		// bodies were set to not exist so implement delete by copying/compacting the array
		cnt += cm.list.Len()
		arr := make([]interfaces.SimBody, cnt)
		j := 0
		for i, size := 0, len(cm.arr); i < size; i++ {
			if cm.arr[i].Exists() {
				arr[j] = cm.arr[i]
				j++
			}
		}
		for e := cm.list.Front(); e != nil; e = e.Next() {
			arr[j] = e.Value.(interfaces.SimBody)
			j++
		}
		cm.arr = arr
		cm.list.Init()
	} else if cm.list.Len() > 0 {
		// only adds - append may be inefficient but adds don't happen that often
		for e := cm.list.Front(); e != nil; e = e.Next() {
			cm.arr = append(cm.arr, e.Value.(interfaces.SimBody))
		}
		cm.list.Init()
	}
}


