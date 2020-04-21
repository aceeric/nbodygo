package interfaces

//
// Function signature of a consumer provided to the 'IterateOnce' function
//
type IterationConsumer func (SimBody)

//
// SimBodyCollection interface
//
type SimBodyCollection interface {
	// concurrent add
	Add(SimBody)

	// makes one traversal through the body array and invokes the passed consumer for each body
	IterateOnce(IterationConsumer)

	// makes a copy of the body array for concurrent read access by another thread
	GetArrayCopy() *[]SimBody

	// handles adds and deletes, preparing the array for the next iteration
	Cycle()

	// gets count of bodies
	Count() int
}
