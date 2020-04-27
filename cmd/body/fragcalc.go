package body

import (
	"math"
	"nbodygo/cmd/globals"
	"nbodygo/cmd/util"
)

//
// This go file just localizes the calculations that are related to implementing fragmentation
// of a body on impact
//

//
// Uses a previous elastic collision calc result and determines - based on approaching velocity
// and fragmentation characteristics whether to initiate fragmentation of a body. Fragmentation
// is initiated and then completed over the course of some number of compute cycles because it is
// expensive to create bodies and add them to the body collection and the goal is to do this
// without introducing noticeable lag
//
// return true and fragmentation factors if should fragment. If false then the frag factors
// are zero and should not be used
//
func (b *Body) shouldFragment(otherBody *Body, r collisionCalcResult) (bool, float64, float64) {
	if !(b.CollisionBehavior == globals.Fragment ||
		otherBody.CollisionBehavior == globals.Fragment) {
		return false, 0, 0
	}
	vThis := b.Vx + b.Vy + b.Vz
	dvThis :=
		math.Abs(b.Vx-((r.vx1-r.vx_cm)*b.r+r.vx_cm)) +
			math.Abs(b.Vy-((r.vy1-r.vy_cm)*b.r+r.vy_cm)) +
			math.Abs(b.Vz-((r.vz1-r.vz_cm)*b.r+r.vz_cm))

	thisFactor := dvThis / math.Abs(vThis)
	vOther := otherBody.Vx + otherBody.Vy + otherBody.Vz
	dvOther :=
		math.Abs(otherBody.Vx-((r.vx2-r.vx_cm)*b.r+r.vx_cm)) +
			math.Abs(otherBody.Vy-((r.vy2-r.vy_cm)*b.r+r.vy_cm)) +
			math.Abs(otherBody.Vz-((r.vz2-r.vz_cm)*b.r+r.vz_cm))

	otherFactor := dvOther / math.Abs(vOther)

	if b.CollisionBehavior == globals.Fragment && thisFactor > b.FragFactor ||
		otherBody.CollisionBehavior == globals.Fragment && otherFactor > otherBody.FragFactor {
		return true, thisFactor, otherFactor
	}
	return false, 0, 0
}

//
// Initiates fragmentation of this body and/or the other body
//
func (b *Body) doFragment(otherBody *Body, thisFactor, otherFactor float64) {
	if b.CollisionBehavior == globals.Fragment && thisFactor > b.FragFactor {
		b.initiateFragmentation(thisFactor)
	}
	if otherBody.CollisionBehavior == globals.Fragment && otherFactor > otherBody.FragFactor {
		otherBody.initiateFragmentation(otherFactor)
	}
}

//
// Initiates fragmentation of a body. The passed fragfactor determines how many fragments will
// be created. The actual fragmentation will be handled the 'fragment' function below
//
func (b *Body) initiateFragmentation(fragFactor float64) {
	var fragDelta float64
	if b.FragFactor > 10 {
		fragDelta = 10
	} else {
		fragDelta = fragFactor - b.FragFactor
	}
	fragments := math.Min(fragDelta*b.FragStep, maxFrags)
	if fragments <= 1 {
		b.CollisionBehavior = globals.Fragment
		return
	}
	b.fragmenting = true
	curPos := util.Vector3{X: b.X, Y: b.Y, Z: b.Z}
	volume := fourThirdsPi * b.Radius * b.Radius * b.Radius
	newRadius := math.Max(math.Pow(((volume/fragments)*3)/fourPi, 1/3), .1)
	newMass := b.Mass / fragments
	b.fragInfo = fragInfo{b.Radius, newRadius, newMass, int(fragments), curPos}
}

//
// Called by the 'Compute' function if a body has been marked as fragmenting. Splits the body into
// fragments, potentially over multiple compute cycles so the sim pace isn't held up creating a large
// number of fragments all at once. Once fully fragmented, then sets this body to not exist. As bodies
// are created, they are enqueued to the passed body collection. The body collection will insert them
// in a thread-safe manner.
//
func (b *Body) fragment(bc *BodyCollection) {
	cnt := 0
	for ; b.fragInfo.fragments > 0; {
		b.fragInfo.fragments--
		v := util.GetVectorEven(b.fragInfo.curPos, b.fragInfo.radius*.9)
		toAdd := &Body{
			Id:   NextId(),
			Name: b.Name, Class: b.Class,
			X: v.X, Y: v.Y, Z: v.Z, Vx: b.Vx, Vy: b.Vy, Vz: b.Vz, Mass: b.fragInfo.mass, Radius: b.fragInfo.newRadius,
			FragFactor: 0, FragStep: 0,
			CollisionBehavior: globals.Elastic, BodyColor: b.BodyColor,
			IsSun: false, Exists: true,
			WithTelemetry: false, Pinned: false,
		}
		bc.Enqueue(NewAdd(toAdd))
		cnt++
		if cnt > maxFragsPerCycle {
			break
		}
	}
	if b.fragInfo.fragments <= 0 {
		b.Exists = false
	} else {
		b.Radius = b.Radius * .9 // todo consider removing given the G3N cost
	}
}
