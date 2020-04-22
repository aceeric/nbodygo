package body

import (
	"fmt"
	"math"
	"nbodygo/cmd/bodyrender"
	"nbodygo/cmd/globals"
	"nbodygo/cmd/interfaces"
	"nbodygo/cmd/util"
	"sync"
	"sync/atomic"
	"time"
)

const (
	G                float64 = 6.673e-11         // gravitational constant
	fourThirdsPi     float64 = math.Pi * (4 / 3) // used in volume/radius calcs
	fourPi           float64 = math.Pi * 4       // "
	R                float64 = 1                 // coefficient of  restitution
	maxFragsPerCycle int     = 100
	maxFrags         float64 = 2000
)

type Body struct {
	id                                int
	name                              string
	class                             string
	collided                          bool
	fragmenting                       bool
	x, y, z, vx, vy, vz, radius, mass float64
	fx, fy, fz                        float64
	fragFactor, fragmentationStep     float64
	collisionBehavior                 globals.CollisionBehavior
	bodyColor                         globals.BodyColor
	// TODO R should be "class" scope
	R             float64
	isSun         bool
	intensity     float32
	exists        bool
	lock          int32
	withTelemetry bool
	pinned        bool
	fragInfo      FragInfo
}

// creates a body that exists with hard-coded values and properties typically of interest specified as function args
func NewBody(id int, x, y, z, vx, vy, vz, mass, radius float64, collisionBehavior globals.CollisionBehavior,
	bodyColor globals.BodyColor, fragFactor, fragmentationStep float64, withTelemetry bool, name, class string,
	pinned bool) Body {
	b := Body{
		id:                id,
		name:              name,
		class:             class,
		collided:          false,
		fragmenting:       false,
		x:                 x,
		y:                 y,
		z:                 z,
		vx:                vx,
		vy:                vy,
		vz:                vz,
		radius:            radius,
		mass:              mass,
		fx:                0,
		fy:                0,
		fz:                0,
		fragFactor:        fragFactor,
		fragmentationStep: fragmentationStep,
		collisionBehavior: collisionBehavior,
		bodyColor:         bodyColor,
		R:                 R,
		isSun:             false,
		intensity:         0,
		exists:            true,
		lock:              0,
		withTelemetry:     withTelemetry,
		pinned:            pinned,
		fragInfo:          FragInfo{},
	}
	return b
}

// TODO getters?

func (b *Body) mod() bool {
	return false // TODO
}

func (b *Body) tryLock() bool {
	return atomic.CompareAndSwapInt32(&b.lock, 0, 1)
}

func (b *Body) unLock() {
	atomic.CompareAndSwapInt32(&b.lock, 1, 0)
}

// begin SimBody and Renderable interface implementation(s)

func (b *Body) Id() int                      { return b.id }
func (b *Body) Exists() bool                 { return b.exists }
func (b *Body) X() float64                   { return b.x }
func (b *Body) Y() float64                   { return b.y }
func (b *Body) Z() float64                   { return b.z }
func (b *Body) X32() float32                 { return float32(b.x) }
func (b *Body) Y32() float32                 { return float32(b.y) }
func (b *Body) Z32() float32                 { return float32(b.z) }
func (b *Body) Radius() float64              { return b.radius }
func (b *Body) Radius32() float32            { return float32(b.radius) }
func (b *Body) IsSun() bool                  { return b.isSun }
func (b *Body) Intensity() float32           { return b.intensity }
func (b *Body) SetSun(intensity float32)     { b.isSun = true; b.intensity = intensity }
func (b *Body) BodyColor() globals.BodyColor { return b.bodyColor }
func (b *Body) SetCollisionBehavior(behavior globals.CollisionBehavior) {
	b.collisionBehavior = behavior
}

func (b *Body) Update(timeScaling float64) interfaces.Renderable {
	if !b.exists {
		return bodyrender.NewFromRenderable(b)
	}
	if !b.collided {
		b.vx += timeScaling * b.fx / b.mass
		b.vy += timeScaling * b.fy / b.mass
		b.vz += timeScaling * b.fz / b.mass
	}
	b.x += timeScaling * b.vx
	b.y += timeScaling * b.vy
	b.z += timeScaling * b.vz
	// clear collided flag for next cycle
	b.collided = false
	if b.withTelemetry {
		// TODO print b to console
	}
	if math.IsNaN(b.x) || math.IsNaN(b.y) || math.IsNaN(b.z) {
		// TODO logger
		b.exists = false
	}
	return bodyrender.NewFromRenderable(b)
}

// TODO equivalent of Java synchronized primitive which is a guaranteed atomic read/write
func (b *Body) ForceComputer(sbc interfaces.SimBodyCollection) {
	// TODO panic/recover
	if b.fragmenting {
		b.fragment(sbc)
	} else {
		b.fx, b.fy, b.fz = 0, 0, 0
		sbc.IterateOnce(func(c interfaces.SimBody) {
			otherBody := c.(*Body)
			if !otherBody.exists || b.fragmenting {
				return
			}
			if b != otherBody && otherBody.exists && !otherBody.fragmenting {
				// todo metrics
				result := b.calcForceFrom(otherBody)
				if result.collided {
					b.resolveCollision(result.dist, otherBody)
					//fmt.Printf("%v -- id %v collided with id %v\n", time.Now(), b.id, otherBody.id)
				}
			}
		})
	}
}

// end SimBody interface implementation

func (b *Body) subsume(dist float64, otherBody *Body) {
	if dist + otherBody.radius >= b.radius*1.2 {
		// leave unless most of the other b is inside this b
		return
	}
	subsumed := false
	var thisMass, otherMass float64
	// todo defer handle unlock
	if b.tryLock() {
		otherLock := otherBody.tryLock()
		if otherLock {
			thisMass = b.mass
			otherMass = otherBody.mass
			/*
				TODO:
				If I allow the radius to grow it occasionally causes a runaway condition in which a b
				swallows the entire simulation. Need to figure this out
				volume := (fourThirdsPi * b.radius  * b.radius  * b.radius) +
					(fourThirdsPi * otherBody.radius  * otherBody.radius  * otherBody.radius)
				newRadius := math.Pow(((volume * 3) / fourPi, 1/3);
				TODOLOG("old radius: {} -- new radius: {}", radius, newRadius);
				b.radius = newRadius;
			*/
			b.mass = thisMass + otherMass
			otherBody.setNotExists()
			subsumed = true
		}
		b.unLock()
		if otherLock {
			otherBody.unLock()
		}
	}
	if subsumed {
		// TODO LOGGER
	}

}

func (b *Body) CopyOf() *Body {
	return &Body{
		b.id,
		b.name,
		b.class,
		b.collided,
		b.fragmenting,
		b.x,
		b.y,
		b.z,
		b.vx,
		b.vy,
		b.vz,
		b.radius,
		b.mass,
		b.fx,
		b.fy,
		b.fz,
		b.fragFactor,
		b.fragmentationStep,
		b.collisionBehavior,
		b.bodyColor,
		b.R,
		b.isSun,
		b.intensity,
		b.exists,
		b.lock,
		b.withTelemetry,
		b.pinned,
		b.fragInfo,
	}
}

func (b *Body) calcForceFromOrig(otherBody *Body) ForceCalcResult {
	dx := otherBody.x - b.x
	dy := otherBody.y - b.y
	dz := otherBody.z - b.z
	dist := math.Sqrt(dx*dx + dy*dy + dz*dz)
	if b.collided || dist > b.radius + otherBody.radius {
		force := G * b.mass * otherBody.mass / (dist * dist)
		b.fx += force * dx / dist
		b.fy += force * dy / dist
		b.fz += force * dz / dist
	} else if dist <= b.radius + otherBody.radius {
		return Collision(dist)
	}
	return NoCollision()
}

func (b *Body) calcForceFrom(otherBody *Body) ForceCalcResult {
	dx := otherBody.x - b.x
	dy := otherBody.y - b.y
	dz := otherBody.z - b.z
	dist := math.Sqrt(dx*dx + dy*dy + dz*dz)
	if dist > b.radius + otherBody.radius {
		force := G * b.mass * otherBody.mass / (dist * dist)
		b.fx += force * dx / dist
		b.fy += force * dy / dist
		b.fz += force * dz / dist
	} else {
		if b.tryLock() {
			defer b.unLock()
			if b.collided {
				force := G * b.mass * otherBody.mass / (dist * dist)
				b.fx += force * dx / dist
				b.fy += force * dy / dist
				b.fz += force * dz / dist
			} else {
				return Collision(dist)
			}
		}
	}
	return NoCollision()
}

func (b *Body) resolveCollision(dist float64, otherBody *Body) {
	if b.collisionBehavior == globals.Subsume || otherBody.collisionBehavior == globals.Subsume {
		if b.radius > otherBody.radius {
			b.subsume(dist, otherBody)
			fmt.Printf("%v, this subsume other\n", time.Now())
		} else {
			otherBody.subsume(dist, b)
			fmt.Printf("%v, other subsume this\n", time.Now())
		}
	} else if (b.collisionBehavior == globals.Elastic || b.collisionBehavior == globals.Fragment) &&
		(otherBody.collisionBehavior == globals.Elastic || otherBody.collisionBehavior == globals.Fragment) {
		r := b.calcElasticCollision(otherBody)
		if r.collided {
			if b.tryLock() {
				otherLock := otherBody.tryLock()
				if otherLock && b.exists && otherBody.exists {
					fr := b.shouldFragment(otherBody, r)
					if fr.shouldFragment {
						b.doFragment(otherBody, fr)
						fmt.Printf("%v, initiate fragment\n", time.Now())
					} else {
						b.doElastic(otherBody, r)
					}
				}
				b.unLock()
				if otherLock {
					otherBody.unLock()
				}
			}
			// RACE COND MOVE INSIDE
			//if b.collided {
			//	// TODO LOGGING
			//}
		}
	}
}

func (b *Body) doElastic(otherBody *Body, r CollisionCalcResult) {
	b.vx = (r.vx1-r.vx_cm) * R + r.vx_cm
	b.vy = (r.vy1-r.vy_cm) * R + r.vy_cm
	b.vz = (r.vz1-r.vz_cm) * R + r.vz_cm
	otherBody.vx = (r.vx2-r.vx_cm) * R + r.vx_cm
	otherBody.vy = (r.vy2-r.vy_cm) * R + r.vy_cm
	otherBody.vz = (r.vz2-r.vz_cm) * R + r.vz_cm
	b.collided = true
	otherBody.collided = true
}

func (b *Body) calcElasticCollision(otherBody *Body) CollisionCalcResult {
	var r12, m21, d, v, theta2, phi2, st, ct, sp, cp, vx1r, vy1r, vz1r, fvz1r,
	thetav, phiv, dr, alpha, beta, sbeta, cbeta, t, a, dvz2,
	vx2r, vy2r, vz2r, x21, y21, z21, vx21, vy21, vz21, vx_cm, vy_cm, vz_cm float64

	m1 := b.mass
	m2 := otherBody.mass
	r1 := b.radius
	r2 := otherBody.radius
	x1 := b.x
	y1 := b.y
	z1 := b.z
	x2 := otherBody.x
	y2 := otherBody.y
	z2 := otherBody.z
	vx1 := b.vx
	vy1 := b.vy
	vz1 := b.vz
	vx2 := otherBody.vx
	vy2 := otherBody.vy
	vz2 := otherBody.vz
	_ = t // unused in source as well but keep as close to the source as possible

	r12 = r1 + r2
	m21 = m2 / m1
	x21 = x2 - x1
	y21 = y2 - y1
	z21 = z2 - z1
	vx21 = vx2 - vx1
	vy21 = vy2 - vy1
	vz21 = vz2 - vz1

	vx_cm = (m1*vx1 + m2*vx2) / (m1 + m2)
	vy_cm = (m1*vy1 + m2*vy2) / (m1 + m2)
	vz_cm = (m1*vz1 + m2*vz2) / (m1 + m2)

	// calculate relative distance and relative speed
	d = math.Sqrt(x21*x21 + y21*y21 + z21*z21)
	v = math.Sqrt(vx21*vx21 + vy21*vy21 + vz21*vz21)

	// commented this out from the original - if the radii overlap run the calc anyway because the sim doesn't
	// prevent bodies overlapping - that would take way too much compute power

	// return if distance between balls smaller than sum of radii
	// if (d < r12) {return NoElasticCollision()}

	// return if relative speed = 0
	if v == 0 {
		// TODO LOGGING
		return NoElasticCollision()
	}
	// shift coordinate system so that ball 1 is at the origin
	x2 = x21
	y2 = y21
	z2 = z21

	// boost coordinate system so that ball 2 is resting
	vx1 = -vx21
	vy1 = -vy21
	vz1 = -vz21

	// find the polar coordinates of the location of ball 2
	theta2 = math.Acos(z2 / d)
	if x2 == 0 && y2 == 0 {
		phi2 = 0
	} else {
		phi2 = math.Atan2(y2, x2)
	}
	st = math.Sin(theta2)
	ct = math.Cos(theta2)
	sp = math.Sin(phi2)
	cp = math.Cos(phi2)

	// express the velocity vector of ball 1 in a rotated coordinate system where ball 2 lies on the z-axis
	vx1r = ct*cp*vx1 + ct*sp*vy1 - st*vz1
	vy1r = cp*vy1 - sp*vx1
	vz1r = st*cp*vx1 + st*sp*vy1 + ct*vz1
	fvz1r = vz1r / v
	if fvz1r > 1 {
		// fix for possible rounding errors
		fvz1r = 1
	} else if fvz1r < -1 {
		fvz1r = -1
	}
	thetav = math.Acos(fvz1r)
	if vx1r == 0 && vy1r == 0 {
		phiv = 0
	} else {
		phiv = math.Atan2(vy1r, vx1r)
	}

	// calculate the normalized impact parameter
	dr = d * math.Sin(thetav) / r12

	// if balls do not collide, do nothing
	if thetav > math.Pi/2 || math.Abs(dr) > 1 {
		// TODO LOGGING
		return NoElasticCollision()
	}
	// calculate impact angles if balls do collide
	alpha = math.Asin(-dr)
	beta = phiv
	sbeta = math.Sin(beta)
	cbeta = math.Cos(beta)

	// commented out from original - position is assigned in the update method

	// calculate time to collision
	//t = (d * Math.cos(thetav) - r12 * Math.sqrt(1 - dr * dr)) / v;
	// update positions and reverse the coordinate shift
	// x2 = x2 + vx2 * t + x1;
	// y2 = y2 + vy2 * t + y1;
	// z2 = z2 + vz2 * t + z1;
	// x1 = (vx1 + vx2) * t + x1;
	// y1 = (vy1 + vy2) * t + y1;
	// z1 = (vz1 + vz2) * t + z1;

	// update velocities

	a = math.Tan(thetav + alpha)

	dvz2 = 2 * (vz1r + a*(cbeta*vx1r+sbeta*vy1r)) / ((1 + a*a) * (1 + m21))

	vz2r = dvz2
	vx2r = a * cbeta * dvz2
	vy2r = a * sbeta * dvz2
	vz1r = vz1r - m21*vz2r
	vx1r = vx1r - m21*vx2r
	vy1r = vy1r - m21*vy2r

	// rotate the velocity vectors back and add the initial velocity
	// vector of ball 2 to retrieve the original coordinate system

	return ElasticCollision(
		ct*cp*vx1r-sp*vy1r+st*cp*vz1r+vx2,
		ct*sp*vx1r+cp*vy1r+st*sp*vz1r+vy2,
		ct*vz1r-st*vx1r+vz2,
		ct*cp*vx2r-sp*vy2r+st*cp*vz2r+vx2,
		ct*sp*vx2r+cp*vy2r+st*sp*vz2r+vy2,
		ct*vz2r-st*vx2r+vz2,
		vx_cm, vy_cm, vz_cm)
}

func (b *Body) shouldFragment(otherBody *Body, r CollisionCalcResult) FragmentationCalcResult {
	if !(b.collisionBehavior == globals.Fragment ||
		otherBody.collisionBehavior == globals.Fragment) {
		return NoFragmentation()
	}
	vThis := b.vx + b.vy + b.vz
	dvThis :=
		math.Abs(b.vx - ((r.vx1 - r.vx_cm) * R + r.vx_cm)) +
		math.Abs(b.vy - ((r.vy1 - r.vy_cm) * R + r.vy_cm)) +
		math.Abs(b.vz - ((r.vz1 - r.vz_cm) * R + r.vz_cm))

	vThisFactor := dvThis / math.Abs(vThis)
	vOther := otherBody.vx + otherBody.vy + otherBody.vz
	dvOther :=
		math.Abs(otherBody.vx - ((r.vx2-r.vx_cm) * R + r.vx_cm)) +
		math.Abs(otherBody.vy - ((r.vy2-r.vy_cm) * R + r.vy_cm)) +
		math.Abs(otherBody.vz - ((r.vz2-r.vz_cm) * R + r.vz_cm))

	vOtherFactor := dvOther / math.Abs(vOther)

	if b.collisionBehavior == globals.Fragment && vThisFactor > b.fragFactor ||
		otherBody.collisionBehavior == globals.Fragment && vOtherFactor > otherBody.fragFactor {
		return Fragmentation(vThisFactor, vOtherFactor)
	}
	return NoFragmentation()
}

func (b *Body) doFragment(otherBody *Body, fr FragmentationCalcResult) {
	if b.collisionBehavior == globals.Fragment && fr.thisFactor > b.fragFactor {
		b.initiateFragmentation(fr.thisFactor)
	}
	if otherBody.collisionBehavior == globals.Fragment && fr.otherFactor > otherBody.fragFactor {
		otherBody.initiateFragmentation(fr.otherFactor)
	}
}

// TODO clean up locking calls
func (b *Body) initiateFragmentation(fragFactor float64) {
	var fragDelta float64
	if b.fragFactor > 10 {
		fragDelta = 10
	} else {
		fragDelta = fragFactor - b.fragFactor
	}
	fragments := math.Min(fragDelta * b.fragmentationStep, maxFrags)
	if fragments <= 1 {
		b.collisionBehavior = globals.Fragment
		return
	}
	b.fragmenting = true
	curPos := util.Vector3{b.x, b.y, b.z}
	volume := fourThirdsPi * b.radius * b.radius * b.radius
	newRadius := math.Max(math.Pow(((volume/fragments)*3) / fourPi, 1/3), .1)
	newMass := b.mass / fragments
	b.fragInfo = FragInfo{b.radius, newRadius, newMass, int(fragments), curPos}
}

// TODO clean up locking calls
func (b *Body) fragment(cc interfaces.SimBodyCollection) {
	cnt := 0
	for ; b.fragInfo.fragments > 0; {
		b.fragInfo.fragments--
		v := util.GetVectorEven(b.fragInfo.curPos, b.fragInfo.radius * .9)
		toAdd := &Body{
			id:   NextId(),
			name: b.name, class: b.class,
			x:  v.X, y: v.Y, z: v.Z, vx: b.vx, vy: b.vy, vz: b.vz, mass: b.fragInfo.mass, radius: b.fragInfo.newRadius,
			fragFactor: 0, fragmentationStep: 0,
			collisionBehavior: globals.Elastic, bodyColor: b.bodyColor,
			R: R, // todo fix this
			isSun: false, exists: true,
			withTelemetry: false, pinned: false,
		}
		cc.Add(toAdd)
		cnt++
		if cnt > maxFragsPerCycle {
			break
		}
	}
	if b.fragInfo.fragments <= 0 {
		// turn this instance into a fragment
		b.mass = b.fragInfo.mass
		b.radius = b.fragInfo.newRadius
		b.collisionBehavior = globals.Elastic
		b.fragmenting = false
		// TODO DON'T KNOW HOW TO CHANGE RADIUS IN G3N SO JUST SET IT NOT EXISTS FOR NOW
		b.exists = false
	} else {
		// shrink the b a little each time
		b.radius = b.radius * .9
	}
}

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

func (b *Body) setNotExists() {
	b.mass = 0
	b.exists = false
}
