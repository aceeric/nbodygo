package interfaces

//
// Function signature of a consumer provided to the 'IterateOnce' function
//
type IterationConsumer func (SimBody)

//
// SimBodyCollection interface
//
type SimBodyCollection interface {
	// Concurrent add
	Add(SimBody)

	// Concurrent update. Updates are enqueued and processed by a single thread at at the end of each compute
	// cycle
	// TODOUpdate(id int, op UpdateOp, params ... float64)

	// Makes one traversal through the body array and invokes the passed consumer for each body
	IterateOnce(IterationConsumer)

	// Makes a copy of the body array for concurrent read access by another thread
	GetArrayCopy() *[]SimBody

	// Handles adds and deletes, preparing the array for the next iteration
	Cycle()

	// Gets count of bodies
	Count() int
}

// Enum that defines the supported update operations
// TODO THIS DOESNT HANDLE GRPC WHIUCH CAN CHANGE ANYTHING
type UpdateOp int
const (
	// a body was subsumed by another body
	Subsumed   UpdateOp = 0
	// a body subsumed another body
	Subsumes   UpdateOp = 1
	// a body collided with another body and the result was an elastic collision
	Elastic    UpdateOp = 2
	// a body collided with another body and the result was fragmentation of the body
	StartFrag  UpdateOp = 3
	// fragmentation is performed in increments to avoid an interruption in the sim. This
	// event indicates that the fragmentation of a body is complete
	FinishFrag UpdateOp = 4
)