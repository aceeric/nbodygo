package body

//
// Events are enqueued in the body collection as they occur during each compute cycle. This allows bodies
// to modify each other (e.g. collide and exchange velocity) in a thread safe way. The enqueued events are
// processed by the computation once each cycle. This avoids synchronization between bodies which would not
// be feasible given the number of concurrent reads. In the Java version, the threads could  modify each
// other's body objects because the JVM guaranteed atomic reads/writes. Granted this resulted in a "dirty
// read" for the Java app but the individual values were immune from race conditions and so - in the
// interests of concurrency it was an acceptable compromise. In Go, however, any concurrent access is a
// race condition. But the high concurrency (2K bodies reading 2K bodies) makes synchronization using
// a mutex infeasible from a performance perspective.
//

//
// Enum that defines the events that are handled
//
type EventType int

const (
	CollisionEvent EventType = 0
	SubsumeEvent   EventType = 1
	AddEvent       EventType = 2
)

//
// The event definition
//
type Event struct {
	evType EventType
	TwoBodies
	OneBody
}

//
// If the event needs two bodies
//
type TwoBodies struct {
	b1 *Body
	b2 *Body
}

//
// If the event needs one body
//
type OneBody struct {
	b *Body
}

//
// Dispatches the event to the appropriate Body handler function
//
func (ev Event) Handle() {
	switch ev.evType {
	case CollisionEvent:
		ev.TwoBodies.b1.ResolveCollision(ev.TwoBodies.b2)
	case SubsumeEvent:
		ev.TwoBodies.b1.ResolveSubsume(ev.TwoBodies.b2)
	case AddEvent:
		panic("Event not handled here -- handled by body collection")
	}
}

//
// Gets the body being added from the passed event
//
func (ev Event) GetAdd() *Body {
	if ev.evType == AddEvent {
		return ev.OneBody.b
	} else {
		return nil
	}
}

//
// Creates an event representing a collision between two bodies
//
func NewCollision(b1 *Body, b2 *Body) Event {
	return Event{
		evType: CollisionEvent,
		TwoBodies: TwoBodies{
			b1: b1,
			b2: b2,
		},
	}
}

//
// Creates an event representing one body subsuming another body
//
func NewSubsume(b1 *Body, b2 *Body) Event {
	return Event{
		evType: SubsumeEvent,
		TwoBodies: TwoBodies{
			b1: b1, // subsumes b2
			b2: b2, // subsumed by b1
		},
	}
}

//
// Creates an event representing a body being added into a running simulation
//
func NewAdd(b *Body) Event {
	return Event{
		evType: AddEvent,
		OneBody: OneBody{
			b: b,
		},
	}
}