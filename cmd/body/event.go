package body

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
// Creates an even representing a collision between two bodies
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
// Creates an even representing one body subsuming another body
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
// Creates an even representing a body being added into a running simulation
//
func NewAdd(b *Body) Event {
	return Event{
		evType: AddEvent,
		OneBody: OneBody{
			b: b,
		},
	}
}
