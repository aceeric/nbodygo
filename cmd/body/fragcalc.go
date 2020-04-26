package body

import (
	"math"
	"nbodygo/cmd/globals"
	"nbodygo/cmd/util"
)

type fragmentationCalcResult struct { // todo consider multi-return value rather than struct - same for fragcalc
	shouldFragment bool
	thisFactor     float64
	otherFactor    float64
}

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
	if !(b.collisionBehavior == globals.Fragment ||
		otherBody.collisionBehavior == globals.Fragment) {
		return noFragmentation()
	}
	vThis := b.vx + b.vy + b.vz
	dvThis :=
		math.Abs(b.vx-((r.vx1-r.vx_cm)*b.R+r.vx_cm)) +
			math.Abs(b.vy-((r.vy1-r.vy_cm)*b.R+r.vy_cm)) +
			math.Abs(b.vz-((r.vz1-r.vz_cm)*b.R+r.vz_cm))

	vThisFactor := dvThis / math.Abs(vThis)
	vOther := otherBody.vx + otherBody.vy + otherBody.vz
	dvOther :=
		math.Abs(otherBody.vx-((r.vx2-r.vx_cm)*b.R+r.vx_cm)) +
			math.Abs(otherBody.vy-((r.vy2-r.vy_cm)*b.R+r.vy_cm)) +
			math.Abs(otherBody.vz-((r.vz2-r.vz_cm)*b.R+r.vz_cm))

	vOtherFactor := dvOther / math.Abs(vOther)

	if b.collisionBehavior == globals.Fragment && vThisFactor > b.fragFactor ||
		otherBody.collisionBehavior == globals.Fragment && vOtherFactor > otherBody.fragFactor {
		return fragmentation(vThisFactor, vOtherFactor)
	}
	return noFragmentation()
}

func (b *Body) doFragment(otherBody *Body, fr fragmentationCalcResult) {
	if b.collisionBehavior == globals.Fragment && fr.thisFactor > b.fragFactor {
		b.initiateFragmentation(fr.thisFactor)
	}
	if otherBody.collisionBehavior == globals.Fragment && fr.otherFactor > otherBody.fragFactor {
		otherBody.initiateFragmentation(fr.otherFactor)
	}
}

func (b *Body) initiateFragmentation(fragFactor float64) {
	var fragDelta float64
	if b.fragFactor > 10 {
		fragDelta = 10
	} else {
		fragDelta = fragFactor - b.fragFactor
	}
	fragments := math.Min(fragDelta*b.fragmentationStep, maxFrags)
	if fragments <= 1 {
		b.collisionBehavior = globals.Fragment
		return
	}
	b.fragmenting = true
	curPos := util.Vector3{X: b.x, Y: b.y, Z: b.z}
	volume := fourThirdsPi * b.radius * b.radius * b.radius
	newRadius := math.Max(math.Pow(((volume/fragments)*3)/fourPi, 1/3), .1)
	newMass := b.mass / fragments
	b.fragInfo = fragInfo{b.radius, newRadius, newMass, int(fragments), curPos}
}

//
// Called by the 'Compute' function if a body has been marked as fragmenting. Splits the body into
// fragments, potentially over multiple compute cycles so the sim pace isn't held up creating a large
// number of fragments all at once. Once fully fragmented, then sets this body to not exist.
//
func (b *Body) fragment(sbc SimBodyCollection) {
	cnt := 0
	for ; b.fragInfo.fragments > 0; {
		b.fragInfo.fragments--
		v := util.GetVectorEven(b.fragInfo.curPos, b.fragInfo.radius*.9)
		toAdd := &Body{
			id:   NextId(),
			name: b.name, class: b.class,
			x: v.X, y: v.Y, z: v.Z, vx: b.vx, vy: b.vy, vz: b.vz, mass: b.fragInfo.mass, radius: b.fragInfo.newRadius,
			fragFactor: 0, fragmentationStep: 0,
			collisionBehavior: globals.Elastic, bodyColor: b.bodyColor,
			isSun: false, exists: true,
			withTelemetry: false, pinned: false,
		}
		sbc.Enqueue(NewAdd(toAdd))
		cnt++
		if cnt > maxFragsPerCycle {
			break
		}
	}
	if b.fragInfo.fragments <= 0 {
		b.exists = false
	} else {
		// TODO REMOVE UNLESS CAN IMPLEMENT IN G3N
		// shrink the b a little each time
		b.radius = b.radius * .9
	}
}
