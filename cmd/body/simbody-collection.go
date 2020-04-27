package body

import (
	"container/list"
	"nbodygo/cmd/grpcsimcb"
	"runtime"
	"sync"
)

//
// The collection state
//
type simBodyCollection struct {
	// this is the body array that all goroutines will iterate
	arr []SimBody
	// a list of concurrent adds, as well as collisions accumulated during each compute cycle
	events *list.List
	// concurrency
	lock sync.Mutex
	// handle add/modify body events
	evCh chan Event
	// allows the gRPC server to get a body synchronized
	getBodyCh chan struct {
		id   int
		name string
	}
	sendBodyCh chan SimBody
	// diagnostic/debugging aid
	cycle     int
	modBodyCh chan struct {
		id          int
		name, class string
		mods        []string
	}
	modBodyResultCh chan grpcsimcb.ModBodyResult
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
		getBodyCh: make(chan
		struct {
			id   int
			name string
		}, 10),
		sendBodyCh: make(chan SimBody, 1),
		modBodyCh: make(chan
		struct {
			id          int
			name, class string
			mods        []string
		}, 10),
		modBodyResultCh: make(chan grpcsimcb.ModBodyResult, 1),
	}
	for i, size := 0, len(bodies); i < size; i++ {
		c.arr[i] = bodies[i]
	}
	go c.handleEvents()
	return &c
}

func (sbc *simBodyCollection) GetArray() []SimBody {
	return sbc.arr
}

//
// supports post-processing events that would cause race conditions or - would require synchronization - if
// done concurrently. Synchronization in the tight nested body computation loop has a prohibitive impact
// on performance
//
func (sbc *simBodyCollection) Enqueue(ev Event) {
	sbc.evCh <- ev
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
		case ev := <-sbc.evCh:
			sbc.lock.Lock()
			sbc.events.PushFront(ev)
			sbc.lock.Unlock()
		}
	}
}

//
// because the computation cycle is always running, this function provides a way for callers
// to register a request to get a body. The function writes to a channel which is checked by the
// 'HandleGetBody' function which is called by the collection's 'Cycle' method. That method finds
// the body, clones it, and writes it to the channel that is checked by the function returned by
// this function. This gives the caller a copy of the body created in a thread-safe way. E.g.:
//
// Assume 'sbc' is pointer to the collection:
//
// var b SimBody = sbc.GetBody()()
//
// The second parens invoke the return function, and doesn't require exposing the channel to the
// caller
//
func (sbc *simBodyCollection) GetBody(id int, name string) func() SimBody {
	sbc.getBodyCh <- struct {
		id   int
		name string
	}{id: id, name: name}
	return func() SimBody {
		return <-sbc.sendBodyCh
	}
}

//
// if there is a 'get body' message on the 'getBodyCh' channel, then searches the body array for the
// body and if found, calls doSendBody to send the body on the 'sendBodyCh' channel
//
func (sbc *simBodyCollection) HandleGetBody() {
	select {
	default:
	case ev := <-sbc.getBodyCh:
		id := ev.id
		name := ev.name
		for i, size := 0, len(sbc.arr); i < size; i++ {
			if (name != "" && name == sbc.arr[i].Name()) || (id == sbc.arr[i].Id()) {
				sbc.doSendBody(sbc.arr[i])
				return
			}
		}
		sbc.doSendBody(nil)
	}
}

//
// Clones the passed body and sends it to the 'sendBodyCh' channel. That channel is monitored
// by the function returned by the 'GetBody' function
//
func (sbc *simBodyCollection) doSendBody(b SimBody) {
	if len(sbc.sendBodyCh) < cap(sbc.sendBodyCh) { // if too many requests just discard them
		if b != nil {
			bb := b.(*Body)
			clone := NewBody(bb.id, bb.x, bb.y, bb.z, bb.vx, bb.vy, bb.vz, bb.mass, bb.radius, bb.collisionBehavior, bb.bodyColor,
				bb.fragFactor, bb.fragmentationStep, bb.withTelemetry, bb.name, bb.class, bb.pinned)
			sbc.sendBodyCh <- &clone
		} else {
			sbc.sendBodyCh <- nil
		}
	}
}

//
// Enqueues a request to modify the properties of a body - or bodies - in the collection. Uses the same
// pattern as GetBody / HandleGetBody / doSendBody except since all it has to return is a result code, it
// doesn't need a doModBody
//
func (sbc *simBodyCollection) ModBody(id int, name, class string, mods []string) func() grpcsimcb.ModBodyResult {
	sbc.modBodyCh <- struct {
		id   int
		name, class string
		mods []string
	}{id: id, name: name, class: class, mods: mods}
	return func() grpcsimcb.ModBodyResult {
		return <-sbc.modBodyResultCh
	}
}

//
// if there is a 'mod body' message on the 'modBodyCh' channel, then searches the body array for all matching
// bodies and if found, calls ApplyMods on the body, then sends the result on the 'modBodyResultCh' channel
//
func (sbc *simBodyCollection) HandleModBody() {
	select {
	default:
	case mod := <-sbc.modBodyCh:
		var found, modified = 0, 0
		for i, size := 0, len(sbc.arr); i < size; i++ {
			b := sbc.arr[i].(*Body)
			if mod.class != "" && mod.class == b.class ||
				mod.name != "" && mod.name == b.name ||
				mod.id == b.id {
				found++
				if b.ApplyMods(mod.mods) {
					modified++
				}
			}
		}
		switch {
		case found == 0: sbc.modBodyResultCh <- grpcsimcb.NoMatch
		case modified == 0: sbc.modBodyResultCh <- grpcsimcb.ModNone
		case found == modified: sbc.modBodyResultCh <- grpcsimcb.ModAll
		default: sbc.modBodyResultCh <- grpcsimcb.ModSome
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
// changing body state in such a way that would require synchronization to avoid race conditions. Adds
// are excluded from this processing. (Handled in the 'Cycle' function.)
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
// done in a single thread to avoid race conditions.
//
// args:
//   R  coefficient of restitution for elastic collision gets plugged into each added body
//
func (sbc *simBodyCollection) Cycle(R float64) {
	cnt := 0
	for i, size := 0, len(sbc.arr); i < size; i++ {
		if sbc.arr[i].Exists() {
			cnt++
		}
	}
	sbc.lock.Lock() // prevents new events
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
				arr[j].SetR(R)
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
					b := e.Value.(Event).GetAdd()
					b.SetR(R)
					sbc.arr = append(sbc.arr, b)
				}
			}
		}
		sbc.events.Init()
	}
	sbc.cycle++
}
