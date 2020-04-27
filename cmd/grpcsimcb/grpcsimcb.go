package grpcsimcb

import (
	"nbodygo/cmd/globals"
)

// ModBodyResult enum defines the result of a call to the ModBody callback
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

//
// Defines a struct of callback functions that the gRPC server can use to call back into the simulation
// to modify the simulation while it is running
//
type GrpcSimCallbacks struct {
	SetResultQueueSize func(int) bool
	ResultQueueSize func() int
	SetSmoothing func(float64)
	Smoothing func() float64
	SetComputationWorkers func(int)
	ComputationWorkers func() int
	SetRestitutionCoefficient func(float64)
	RestitutionCoefficient func() float64
	RemoveBodies func(int)
	BodyCount func() int
	AddBody func(float64, float64, float64, float64, float64, float64, float64, float64,
		bool, globals.CollisionBehavior,globals.BodyColor, float64, float64, bool, string, string, bool) int
	ModBody func(int, string, string, []string) ModBodyResult
	GetBody func(int, string) interface{}
}
