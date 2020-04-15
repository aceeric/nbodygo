package body

import "nbodygo/cmd/cmap"

// The SimBody interface defines the functionality required to integrate a body into the computation runner
type SimBody interface {
	ForceComputer(bodyQueue *cmap.ConcurrentMap, result chan<- bool)
	Update(timeScaling float32) BodyRenderInfo
	Exists() bool
	Id() int
}
