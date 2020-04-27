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
// carries related fragmentation values. Seems cleaner than returning and passing the same three
// args
//
type fragmentationCalcResult struct {
	shouldFragment bool
	thisFactor     float64
	otherFactor    float64
}

//
// returns a 'fragmentationCalcResult' struct indicating no fragmentatino
func noFragmentation() fragmentationCalcResult {
	return fragmentationCalcResult{
		false, 0, 0,
	}
}

func fragmentation(thisFactor float64, otherFactor float64) fragmentationCalcResult {
	return fragmentationCalcResult{
		true, thisFactor, otherFactor,
	}
}

func (b *Body) shouldFragment(otherBody *Body, r collisionCalcResult) fragmentationCalcResult {
	if !(b.CollisionBehavior == globals.Fragment ||
		otherBody.CollisionBehavior == globals.Fragment) {
		return noFragmentation()
	}
	vThis := b.Vx + b.Vy + b.Vz
	dvThis :=
		math.Abs(b.Vx-((r.vx1-r.vx_cm)*b.r+r.vx_cm)) +
			math.Abs(b.Vy-((r.vy1-r.vy_cm)*b.r+r.vy_cm)) +
			math.Abs(b.Vz-((r.vz1-r.vz_cm)*b.r+r.vz_cm))

	vThisFactor := dvThis / math.Abs(vThis)
	vOther := otherBody.Vx + otherBody.Vy + otherBody.Vz
	dvOther :=
		math.Abs(otherBody.Vx-((r.vx2-r.vx_cm)*b.r+r.vx_cm)) +
			math.Abs(otherBody.Vy-((r.vy2-r.vy_cm)*b.r+r.vy_cm)) +
			math.Abs(otherBody.Vz-((r.vz2-r.vz_cm)*b.r+r.vz_cm))

	vOtherFactor := dvOther / math.Abs(vOther)

	if b.CollisionBehavior == globals.Fragment && vThisFactor > b.FragFactor ||
		otherBody.CollisionBehavior == globals.Fragment && vOtherFactor > otherBody.FragFactor {
		return fragmentation(vThisFactor, vOtherFactor)
	}
	return noFragmentation()
}

func (b *Body) doFragment(otherBody *Body, fr fragmentationCalcResult) {
	if b.CollisionBehavior == globals.Fragment && fr.thisFactor > b.FragFactor {
		b.initiateFragmentation(fr.thisFactor)
	}
	if otherBody.CollisionBehavior == globals.Fragment && fr.otherFactor > otherBody.FragFactor {
		otherBody.initiateFragmentation(fr.otherFactor)
	}
}

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
// number of fragments all at once. Once fully fragmented, then sets this body to not exist.
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
