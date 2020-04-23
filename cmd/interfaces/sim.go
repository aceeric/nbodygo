package interfaces

import (
	"container/list"
	"nbodygo/cmd/body"
	"nbodygo/cmd/globals"
)

//
// Defines an interface for extenal entities (i.e. the gRPC server) to interact with a running
// simulation
//
type SimInteraction interface {
	//
	// Sets the result queue size, which enables the computation thread to outrun the render thread
	//
	// args:
	//   queueSize the size to set
	//
	SetResultQueueSize(queueSize int )

	//
	// returns the current result queue size
	//
	ResultQueueSize() int

	//
	// Sets a smoothing factor. All force and velocity calculations are scaled by this value
	//
	// args:
	//  smoothing  Larger is faster, smaller is slower
	//
	SetSmoothing(smoothing float64)

	//
	// returns the current smoothing factoer
	//
	Smoothing() float64

	//
	// Sets the number of workers dedicated to computing force/position of all bodies in the simulation
	// TODO needs to be implemented in the worker pool
	// args:
	//   workerCnt  The number of workers
	//
	SetComputationWorkers(workerCnt int)

	//
	// returns the current number of workers
	//
	ComputationWorkers() int

	//
	// Sets the coefficient of restitution for elastic collisions
	//
	// args:
	//   R the value to set
	//
	SetRestitutionCoefficient(R float64)

	//
	// returns the current coefficient of restitution
	//
	RestitutionCoefficient() float64

	//
	// Removes bodies from the simulation. The interface does not attempt to specify how bodies are selected
	// for removal, only the quantity
	//
	// args:
	//   countToRemove  The number of bodies to remove
	//
	RemoveBodies(countToRemove int)

	//
	// returns the current number of bodies in the simulation
	//
	BodyCount() int

	//
	// Adds a body to the simulation. Params are not documented, as they are consistent with the body package
	//
	// returns the ID of the body added
	//
	AddBody(mass, x, y, z, vx, vy, vz, radius float64,
		isSun bool, behavior globals.CollisionBehavior, color globals.BodyColor,
		fragFactor, fragStep float64,
		withTelemetry bool,
		name, class string,
		pinned bool) int

	//
	// Modifies body properties
	//
	// args:
	//   id    to modify a body by ID
	//   name  " name
	//   class " multiple bodies by class designator (e.g. "asteroid")
	//
	ModBody(id int, name string, class string, bodyMods *list.List) ModBodyResult

	//
	// Gets a body to display its properties
	//
	// args:
	//   id    to get a body by ID
	//   name  " name
	//
	GetBody(id int, name string) body.SimBody
}

// ModBodyResult enum defines the result of a call to ModBody
type ModBodyResult int
const (
	NoMatch ModBodyResult = 0
	ModAll  ModBodyResult = 1
	ModSome ModBodyResult = 2
	ModNone ModBodyResult = 3
)
func (mbr ModBodyResult) String() string {
	return [...]string{"No matching bodies", "All matching bodies were modified",
		"Some matching bodies were modified", "No matching bodies were modified"}[mbr]
}

