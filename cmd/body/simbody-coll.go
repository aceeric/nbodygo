package body

//
// Function signature of a consumer provided to the 'IterateOnce' function
//
type IterationConsumer func (SimBody)

//
// SimBodyCollection interface
//
type SimBodyCollection interface {

	// allows other goroutines to add bodies to the queue and modify bodies in the queue without
	// synchronization on the bodies themselves
	Enqueue(Event)

	// Makes one traversal through the body array and invokes the passed consumer for each body
	IterateOnce(IterationConsumer)

	// Makes a copy of the body array for concurrent read access by another thread
	GetArrayCopy() *[]SimBody

	// Processes enqueued events that modify body state in a single goroutine to avoid race conditions
	ProcessMods()

	// Handles adds and deletes, preparing the array for the next iteration
	Cycle(float64)

	// Gets count of bodies
	Count() int

	// provides a copy of a body to the caller in a thread-safe manner
	GetBody(id int, name string) func() SimBody

	// TODO do I want this?
	HandleGetBody()
}
