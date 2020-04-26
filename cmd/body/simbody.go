package body

import (
	"nbodygo/cmd/globals"
	"nbodygo/cmd/renderable"
)

//
// The SimBody interface defines the functionality required to integrate a body into
// the computation runner, and hence into the n-body simulation. This interface expresses only the
// requirements for computation, not rendering
//
type SimBody interface {
	// Computes force on this body from all other bodies in the sim. Detects and resolves collisions with
	// other bodies. Accumulates force in the body, and may change other body properties based on collision
	Compute(SimBodyCollection)

	// Updates velocity and position from the results of gravitational force calculation and collision
	// resolution. The function returns a 'Renderable' which contains only the information needed to render
	// the body in the rendering engine. Also updates coefficient of restitution, allowing this to be
	// changed globally in the simulation
	Update(timeScaling, R float64) renderable.Renderable

	// Returns true if the body exists, false if the body has been destroyed and should be
	// removed from the simulation and the scene graph. A body can be destroyed if - for example - if gets
	// subsumed into a larger body on impact
	Exists() bool

	// sets the body not to exist
	SetNotExists()

	// returns whether the body is pinned or not

	IsPinned() bool

	// Every body must have a unique ID, which is used in various maps throughout the app
	Id() int

	// Returns the body name, which could be the empty string
	Name() string

	// sets the body to be a Sun, and sets the intensity to the passed value
	SetSun(float64)

	// sets the collision behavior
	SetCollisionBehavior(behavior globals.CollisionBehavior)

	// sets the coefficient of restitution
	SetR(R float64)

	// resolves collision between bodies
	ResolveCollision(SimBody)

	// resolves one body subsuming another
	ResolveSubsume(SimBody)
}
