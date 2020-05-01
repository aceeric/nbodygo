package body

//
// The 'body' go file has most of the functionality associated with representing a body in the simulation
//

import (
	"fmt"
	"log"
	"math"
	"nbodygo/cmd/globals"
	"nbodygo/cmd/instrumentation"
	"nbodygo/cmd/util"
	"strings"
	"sync"
)

const (
	G                float64 = 6.673e-11         // gravitational constant
	fourThirdsPi     float64 = math.Pi * (4 / 3) // used in volume/radius calcs
	fourPi           float64 = math.Pi * 4       // "
	maxFragsPerCycle int     = 100               // max number of bodies that a body can frag into per cycle
	maxFrags         float64 = 2000              // max fragments a body can fragment into
)

//
// Maintains fragmentation state as a body is fragmenting - potentially across compute cycles
//
type fragInfo struct {
	radius, newRadius, mass float64
	fragments               int
	curPos                  util.Vector3
}

//
// The simulation body
//
type Body struct {
	Id                                int
	Name                              string
	Class                             string
	X, Y, Z, Vx, Vy, Vz, Radius, Mass float64
	FragFactor, FragStep              float64
	CollisionBehavior                 globals.CollisionBehavior
	BodyColor                         globals.BodyColor
	IsSun                             bool
	Exists                            bool
	WithTelemetry                     bool
	Pinned                            bool
	r                                 float64
	fragmenting                       bool
	intensity                         float64
	fragInfo                          fragInfo
	fx, fy, fz                        float64
	collided                          bool
}

//
// Creates a body that exists with hard-coded values and properties typically of
// interest specified as function args
//
func NewBody(id int, x, y, z, vx, vy, vz, mass, radius float64, collisionBehavior globals.CollisionBehavior,
	bodyColor globals.BodyColor, fragFactor, fragmentationStep float64, withTelemetry bool, name, class string,
	pinned bool) *Body {
	b := Body{
		Id:                id,
		Name:              name,
		Class:             class,
		collided:          false,
		fragmenting:       false,
		X:                 x,
		Y:                 y,
		Z:                 z,
		Vx:                vx,
		Vy:                vy,
		Vz:                vz,
		Radius:            radius,
		Mass:              mass,
		fx:                0,
		fy:                0,
		fz:                0,
		FragFactor:        fragFactor,
		FragStep:          fragmentationStep,
		CollisionBehavior: collisionBehavior,
		BodyColor:         bodyColor,
		r:                 1,
		IsSun:             false,
		intensity:         0,
		Exists:            true,
		WithTelemetry:     withTelemetry,
		Pinned:            pinned,
		fragInfo:          fragInfo{},
	}
	return &b
}

//
// Sets the body not to exist, which will result in it being removed from the body collection
// and also from the rendering engine scene graph
//
func (b *Body) SetNotExists() {
	b.Mass = 0
	b.Exists = false
}

//
// Sets the body to be a sun, with the passed intensity. This results in a light source being
// associated with the body in the rendering engine. The intensity is needed to offset the G3N light
// source decay which causes the light to dim over distance
//
func (b *Body) SetSun(intensity float64) {
	b.IsSun = true
	b.intensity = intensity
}

//
// Applies the accumulated force to the velocity and position of the body. Intent is to call
// this once all bodies have calculated force on themselves from other bodies
//
// args:
//   timeScaling  time scale (see 'main' package for origin)
//   R            coefficient of restitution - allows that to be updated via the gRPC interface and
//                propagated out to all the bodies
//
func (b *Body) Update(timeScaling, R float64) *Renderable {
	if !b.Exists {
		return NewRenderable(b)
	}
	// todo consider removing the guard?
	if !b.collided {
		b.Vx += timeScaling * b.fx / b.Mass
		b.Vy += timeScaling * b.fy / b.Mass
		b.Vz += timeScaling * b.fz / b.Mass
	}
	b.X += timeScaling * b.Vx
	b.Y += timeScaling * b.Vy
	b.Z += timeScaling * b.Vz
	// clear collided flag for next cycle
	b.collided = false
	b.r = R
	if b.WithTelemetry {
		fmt.Printf("id:%v x:%v y:%v z:%v vx:%v vy:%v vz:%v m:%v r:%v", b.Id, b.X, b.Y, b.Z,
			b.Vx, b.Vy, b.Vz, b.Mass, b.Radius)
	}
	if math.IsNaN(b.X) || math.IsNaN(b.Y) || math.IsNaN(b.Z) {
		log.Printf("[ERROR] NaN values. id=%d (removing from sim)", b.Id)
		b.Exists = false
	}
	return NewRenderable(b)
}

//
// executes a for loop:
//
// for each body in the collection
//   update my force from other body
//   check for collision - if collision
//     enqueue resolution for subsequent (thread-safe) collision resolution
//
func (b *Body) Compute(bc *BodyCollection) {
	if !b.Exists {
		return
	}
	if b.fragmenting {
		b.fragment(bc)
		return
	}
	b.fx, b.fy, b.fz = 0, 0, 0
	bc.IterateOnce(func(otherBody *Body) {
		if !otherBody.Exists || b.fragmenting {
			return
		}
		if b != otherBody && otherBody.Exists && !otherBody.fragmenting {
			instrumentation.BodyComputations.Inc()
			collided, dist := b.calcForceFrom(otherBody)
			if collided {
				log.Printf("[INFO] Body id %v collided with id %v\n", b.Id, otherBody.Id)
				if (b.CollisionBehavior == globals.Elastic || b.CollisionBehavior == globals.Fragment) &&
					(otherBody.CollisionBehavior == globals.Elastic || otherBody.CollisionBehavior == globals.Fragment) {
					bc.Enqueue(newCollision(b, otherBody))
				} else if b.CollisionBehavior == globals.Subsume || otherBody.CollisionBehavior == globals.Subsume {
					if b.Radius > otherBody.Radius && dist <= b.Radius {
						bc.Enqueue(newSubsume(b, otherBody))
					} else if otherBody.Radius > b.Radius && dist <= otherBody.Radius {
						bc.Enqueue(newSubsume(otherBody, b))
					}
				}
			}
		}
	})
}

//
// Accumulates gravitational force on this body from other body. Also checks for collisions
//
// returns true if this body collided with otherBody, else false. If true, second return value
// is the distance between the centers of the two spheres.
//
func (b *Body) calcForceFrom(otherBody *Body) (bool, float64) {
	dx := otherBody.X - b.X
	dy := otherBody.Y - b.Y
	dz := otherBody.Z - b.Z
	dist := math.Sqrt(dx*dx + dy*dy + dz*dz)
	// Only allow one collision per body per cycle. Once a collision happens, continue to apply gravitational
	// force to the collided body. Allowing a body to collide multiple times caused odd things to happen
	// when many bodies were tightly compacted (not sure why) and it also impacts performance - the collision
	// calculation is expensive. This is a compromise
	if b.collided || dist > b.Radius+otherBody.Radius {
		force := G * b.Mass * otherBody.Mass / (dist * dist)
		b.fx += force * dx / dist
		b.fy += force * dy / dist
		b.fz += force * dz / dist
		return false, 0
	} else if dist <= b.Radius+otherBody.Radius {
		log.Printf("[INFO] collision: distance:%v this-radius:%v other-radius:%v this-id:%v other-id%v\n",
			dist, b.Radius, otherBody.Radius, b.Id, otherBody.Id)
		return true, dist
	}
	return false, 0
}

//
// Absorbs 'otherBody' into this body, and sets 'otherBody' not exists
//
func (b *Body) ResolveSubsume(otherBody *Body) {
	var thisMass, otherMass float64
	thisMass = b.Mass
	otherMass = otherBody.Mass
	/*
		// todo:
		// If I allow the radius to grow it occasionally causes a runaway condition in which a body
		// swallows the entire simulation. Need to figure this out
		volume := (fourThirdsPi * b.radius * b.radius * b.radius) +
			(fourThirdsPi * otherBody.radius * otherBody.radius * otherBody.radius)
		newRadius := math.Pow(((volume * 3) / fourPi, 1/3);
		b.radius = newRadius;
	*/
	b.Mass = thisMass + otherMass
	otherBody.SetNotExists()
	log.Printf("[INFO] Body ID %v (mass %v) subsumed ID %v (mass %v)\n", b.Id, thisMass, otherBody.Id, otherMass)
}

//
// Determines whether an elastic collision - or a fragmentation - should be the result of a collision
// and invokes the appropriate function
//
func (b *Body) ResolveCollision(otherBody *Body) {
	if ! b.Exists || !otherBody.Exists {
		return
	}
	if b.CollisionBehavior == globals.Elastic &&
		(otherBody.CollisionBehavior == globals.Elastic || otherBody.CollisionBehavior == globals.Fragment) {
		r := b.calcElasticCollision(otherBody)
		if r.collided {
			shouldFragment, thisFactor, otherFactor := b.shouldFragment(otherBody, r)
			if shouldFragment {
				b.doFragment(otherBody, thisFactor, otherFactor)
			} else {
				b.doElastic(otherBody, r)
			}
		}
	}
}

//
// Applies the passed modifications to the body. Supports the ability to change characteristics of
// a body in the simulation while the sim is running (i.e. via gRPC)
//
// args
//   mods An array of property=value strings. E.g.: "color=blue". Or "x=123". Unknown properties
//        are ignored. Parse errors are ignored. If an array element is not in property=value
//        form, it is ignored
//
func (b *Body) ApplyMods(mods []string) bool {
	for _, mod := range mods {
		nvp := strings.Split(mod, "=")
		if len(nvp) == 2 {
			switch strings.ToUpper(nvp[0]) {
			case "X":
				b.X = globals.SafeParseFloat(nvp[1], b.X)
			case "Y":
				b.Y = globals.SafeParseFloat(nvp[1], b.Y)
			case "Z":
				b.Z = globals.SafeParseFloat(nvp[1], b.Z)
			case "VX":
				b.Vx = globals.SafeParseFloat(nvp[1], b.Vx)
			case "VY":
				b.Vy = globals.SafeParseFloat(nvp[1], b.Vy)
			case "VZ":
				b.Vz = globals.SafeParseFloat(nvp[1], b.Vz)
			case "MASS":
				b.Mass = globals.SafeParseFloat(nvp[1], b.Mass)
			case "RADIUS":
				b.Radius = globals.SafeParseFloat(nvp[1], b.Radius)
			case "FRAG_FACTOR":
				b.FragFactor = globals.SafeParseFloat(nvp[1], b.FragFactor)
			case "FRAG_STEP":
				b.FragStep = globals.SafeParseFloat(nvp[1], b.FragStep)
			case "COLLISION":
				b.CollisionBehavior = globals.ParseCollisionBehavior(nvp[1])
			case "COLOR":
				if !b.IsSun { // suns are always white in the current version
					b.BodyColor = globals.ParseBodyColor(nvp[1])
				}
			case "TELEMETRY":
				b.WithTelemetry = globals.ParseBoolean(nvp[1])
			case "EXISTS":
				b.Exists = globals.ParseBoolean(nvp[1])
			}
		}
	}
	return true
}

//
// Generates a 1-up ID on each call to assign to bodies as they are created
//
var idGenerator = struct {
	lock sync.Mutex
	id   int
}{
	sync.Mutex{}, 0,
}

func NextId() (id int) {
	idGenerator.lock.Lock()
	id = idGenerator.id
	idGenerator.id++
	idGenerator.lock.Unlock()
	return
}
