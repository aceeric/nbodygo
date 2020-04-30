package body

import (
	"container/list"
	"nbodygo/cmd/grpcsimcb"
	"runtime"
	"sync"
)

type IterationConsumer func(*Body)

//
// The collection state
//
type BodyCollection struct {
	// this is the body array that all goroutines will iterate
	arr []*Body
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
	sendBodyCh chan *Body
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
func NewSimBodyCollection(bodies []*Body) *BodyCollection {
	c := BodyCollection{
		arr:    make([]*Body, len(bodies)),
		events: list.New(),
		lock:   sync.Mutex{},
		evCh:   make(chan Event, 5000), // todo factor of body count?
		getBodyCh: make(chan
		struct {
			id   int
			name string
		}, 10),
		sendBodyCh: make(chan *Body, 1),
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

func (bc *BodyCollection) GetArray() []*Body {
	return bc.arr
}

//
// supports post-processing events that would cause race conditions or - would require synchronization - if
// done concurrently. Synchronization in the tight nested body computation loop has a prohibitive impact
// on performance
//
func (bc *BodyCollection) Enqueue(ev Event) {
	bc.evCh <- ev
}

//
// Go routine that supports concurrent (deferred) adds and modifications to body state. Receives events
// from the 'Enqueue' function through the 'evCh' channel. Adds events to an internal list, which is handled
// by a call to the 'ProcessMods' function. (See computation runner.)
//
func (bc *BodyCollection) handleEvents() {
	for {
		select {
		default:
			runtime.Gosched()
		case ev := <-bc.evCh:
			bc.lock.Lock()
			bc.events.PushFront(ev)
			bc.lock.Unlock()
		}
	}
}

//
// Because the computation cycle is always running, this function provides a way for callers
// to register a request to get a body. The function writes to a channel which is checked by the
// 'HandleGetBody' function which is called by the collection's 'Cycle' method. That method finds
// the body, clones it, and writes it to the channel that is checked by this function. This gives
// the caller a copy of the body created in a thread-safe way.
//
func (bc *BodyCollection) GetBody(id int, name string) *Body {
	bc.getBodyCh <- struct {
		id   int
		name string
	}{id: id, name: name}
	return <-bc.sendBodyCh
}

//
// If there is a 'get body' message on the 'getBodyCh' channel, then searches the body array for the
// body and if found, calls doSendBody to send the body on the 'sendBodyCh' channel
//
func (bc *BodyCollection) HandleGetBody() {
	select {
	default:
	case ev := <-bc.getBodyCh:
		id := ev.id
		name := ev.name
		for i, size := 0, len(bc.arr); i < size; i++ {
			if (name != "" && name == bc.arr[i].Name) || (id == bc.arr[i].Id) {
				bc.doSendBody(bc.arr[i])
				return
			}
		}
		bc.doSendBody(nil)
	}
}

//
// Clones the passed body and sends it to the 'sendBodyCh' channel. That channel is monitored
// by the function returned by the 'GetBody' function
//
func (bc *BodyCollection) doSendBody(b *Body) {
	if len(bc.sendBodyCh) < cap(bc.sendBodyCh) { // if too many requests just discard them
		if b != nil {
			clone := NewBody(b.Id, b.X, b.Y, b.Z, b.Vx, b.Vy, b.Vz, b.Mass, b.Radius, b.CollisionBehavior, b.BodyColor,
				b.FragFactor, b.FragStep, b.WithTelemetry, b.Name, b.Class, b.Pinned)
			bc.sendBodyCh <- clone
		} else {
			bc.sendBodyCh <- nil
		}
	}
}

//
// Enqueues a request to modify the properties of a body - or bodies - in the collection. Uses the same
// pattern as GetBody / HandleGetBody / doSendBody except since all it has to return is a result code, it
// doesn't need a doModBody
//
func (bc *BodyCollection) ModBody(id int, name, class string, mods []string) func() grpcsimcb.ModBodyResult {
	bc.modBodyCh <- struct {
		id          int
		name, class string
		mods        []string
	}{id: id, name: name, class: class, mods: mods}
	return func() grpcsimcb.ModBodyResult {
		return <-bc.modBodyResultCh
	}
}

//
// If there is a 'mod body' message on the 'modBodyCh' channel, then searches the body array for all matching
// bodies and if found, calls ApplyMods on the body, then sends the result on the 'modBodyResultCh' channel
//
func (bc *BodyCollection) HandleModBody() {
	select {
	default:
	case mod := <-bc.modBodyCh:
		var found, modified = 0, 0
		for i, size := 0, len(bc.arr); i < size; i++ {
			if mod.class != "" && mod.class == bc.arr[i].Class ||
				mod.name != "" && mod.name == bc.arr[i].Name ||
				mod.id == bc.arr[i].Id {
				found++
				if bc.arr[i].ApplyMods(mod.mods) {
					modified++
				}
			}
		}
		switch {
		case found == 0:
			bc.modBodyResultCh <- grpcsimcb.NoMatch
		case modified == 0:
			bc.modBodyResultCh <- grpcsimcb.ModNone
		case found == modified:
			bc.modBodyResultCh <- grpcsimcb.ModAll
		default:
			bc.modBodyResultCh <- grpcsimcb.ModSome
		}
	}
}

//
// Iterator with a callback pattern. Since there is a lot of iteration, it seemed to make sense to
// encapsulate the iterator with the consumer as a callback. That way if there ever needs to be something
// unique about the iteration it can be hidden here and iterators don't need to be concerned with it
//
func (bc *BodyCollection) IterateOnce(c IterationConsumer) {
	for i, size := 0, len(bc.arr); i < size; i++ {
		c(bc.arr[i])
	}
}

//
// Gets the number of bodies in the collection
//
func (bc *BodyCollection) Count() int {
	bc.lock.Lock()
	defer bc.lock.Unlock()
	return len(bc.arr)
}

//
// Walks the internal 'events' list and processes all enqueued events. These are events that require
// changing body state in such a way that would require synchronization to avoid race conditions. Adds
// are excluded from this processing. (Handled in the 'Cycle' function.)
//
func (bc *BodyCollection) ProcessMods() {
	bc.lock.Lock()
	if bc.events.Len() == 0 {
		bc.lock.Unlock()
		return
	}
	evs := []Event{}
	var prev *list.Element = nil
	for e := bc.events.Front(); e != nil; e = e.Next() {
		if prev != nil {
			bc.events.Remove(prev)
		}
		if e.Value.(Event).evType != AddEvent {
			evs = append(evs, e.Value.(Event))
			prev = e
		}
	}
	bc.lock.Unlock()
	for i, len := 0, len(evs); i < len; i++ {
		evs[i].Handle()
	}
}

//
// Returns the count of events in the internal 'events' list that are Adds
//
func (bc *BodyCollection) countAdds() int {
	cnt := 0
	for e := bc.events.Front(); e != nil; e = e.Next() {
		if e.Value.(Event).evType == AddEvent {
			cnt++
		}
	}
	return cnt
}

//
// Called by computation runner to prepare the body collection for another compute cycle. Removes refs
// to bodies that have been set not to exist, and adds bodies that have been enqueued for addition by the
// gRPC interface
//
// args:
//   R  coefficient of restitution for elastic collision gets plugged into each added body
//
func (bc *BodyCollection) Cycle(R float64) {
	cnt := 0
	for i, size := 0, len(bc.arr); i < size; i++ {
		if bc.arr[i].Exists {
			cnt++
		}
	}
	bc.lock.Lock() // prevents new events
	defer bc.lock.Unlock()
	if cnt < len(bc.arr) {
		cnt += bc.countAdds()
		// bodies were set to not exist so implement delete by copying/compacting the array
		arr := make([]*Body, cnt)
		j := 0
		for i, size := 0, len(bc.arr); i < size; i++ {
			if bc.arr[i].Exists {
				arr[j] = bc.arr[i]
				j++
			}
		}
		for e := bc.events.Front(); e != nil; e = e.Next() {
			if e.Value.(Event).evType == AddEvent {
				arr[j] = e.Value.(Event).GetAdd()
				arr[j].r = R
				j++
			}
		}
		bc.arr = arr
		bc.events.Init()
	} else {
		cnt = bc.countAdds()
		if cnt > 0 {
			for e := bc.events.Front(); e != nil; e = e.Next() {
				if e.Value.(Event).evType == AddEvent {
					b := e.Value.(Event).GetAdd()
					b.r = R
					bc.arr = append(bc.arr, b)
				}
			}
		}
		bc.events.Init()
	}
	bc.cycle++
}
