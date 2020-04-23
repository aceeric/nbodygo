package body

type EventType int

const (
	CollisionEvent EventType = 0
	SubsumeEvent   EventType = 1
	FragmentEvent  EventType = 2
	AddEvent       EventType = 3
)

type TwoBodies struct {
	b1 SimBody
	b2 SimBody
}

type OneBody struct {
	b SimBody
}

type Event struct {
	evType EventType
	TwoBodies
	OneBody
}

func (ev Event) Handle() {
	switch ev.evType {
	case CollisionEvent:
		ev.TwoBodies.b1.ResolveCollision(ev.TwoBodies.b2)
	case SubsumeEvent:
		ev.TwoBodies.b1.ResolveSubsume(ev.TwoBodies.b2)
	case FragmentEvent:
	case AddEvent:
		// not handled here - handled by body collection
	}
}

func (ev Event) GetAdd() SimBody {
	if ev.evType == AddEvent {
		return ev.OneBody.b
	}  else {
		return nil
	}
}

func NewCollision(b1 SimBody, b2 SimBody) Event {
	return Event{
		evType:    CollisionEvent,
		TwoBodies: TwoBodies{
			b1:b1,
			b2:b2,
		},
	}
}

func NewSubsume(b1 SimBody, b2 SimBody) Event {
	return Event{
		evType:    SubsumeEvent,
		TwoBodies: TwoBodies{
			b1: b1, // subsumes b2
			b2: b2, // subsumed by b1
		},
	}
}

func NewAdd(b SimBody) Event {
	return Event{
		evType:    AddEvent,
		OneBody: OneBody{
			b: b,
		},
	}
}