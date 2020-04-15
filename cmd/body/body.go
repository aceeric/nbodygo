package body

import (
	"math"
	"nbodygo/cmd/cmap"
	"nbodygo/cmd/globals"
	"nbodygo/internal/pkg/math32"
	"sync"
)

const (
	G                float64 = 6.673e-11         // gravitational constant
	fourThirdsPi     float32 = math.Pi * (4 / 3) // used in volume/radius calcs
	fourPi           float32 = math.Pi * 4       // "
	R                float32 = 1                 // coefficient of  restitution
	maxFragsPerCycle int     = 100
	maxFrags         float32 = 2000
)

type Body struct {
	id                                int
	name                              string
	class                             string
	collided                          bool
	fragmenting                       bool
	x, y, z, vx, vy, vz, radius, mass float32
	fx, fy, fz                        float64
	fragFactor, fragmentationStep     float32
	collisionBehavior                 globals.CollisionBehavior
	bodyColor                         globals.BodyColor
	R                                 float32
	isSun                             bool
	exists                            bool
	// TODO LOCK
	withTelemetry bool
	pinned        bool
	fragInfo FragInfo
}



func NewBody(id int, x, y, z, vx, vy, vz, mass, radius float32, collisionBehavior globals.CollisionBehavior,
	bodyColor globals.BodyColor, fragFactor, fragmentationStep float32, withTelemetry bool, name, class string,
	pinned bool) Body {
	b := Body{
		// TODO LOCK
		id: id, x: x, y: y, z: z, vx: vx, vy: vy, vz: vz, mass: mass, radius: radius,
		collisionBehavior: collisionBehavior, bodyColor: bodyColor,
		fragFactor: fragFactor, fragmentationStep: fragmentationStep,
		withTelemetry: withTelemetry,
		name:          name, class: class, pinned: pinned,
	}
	return b
}

// TODO getters?

func (body *Body) mod() bool {
	return false  // TODO
}

func (body *Body) tryLock() bool {
	return false  // TODO
}

func (body *Body) unLock() {
	// TODO
}

// begin SimBody interface implementation

func (body *Body) Exists() bool {
	return body.exists
}

func (body *Body) Id() int {
	return body.id
}

func (body *Body) Update(timeScaling float32) BodyRenderInfo {
	if !body.exists {
		return NewBodyRenderInfo(body)
	}
	if !body.collided {
		body.vx += timeScaling * float32(body.fx) / body.mass
		body.vy += timeScaling * float32(body.fy) / body.mass
		body.vz += timeScaling * float32(body.fz) / body.mass
	}
	body.x += timeScaling * body.vx;
	body.y += timeScaling * body.vy;
	body.z += timeScaling * body.vz;
	// clear collided flag for next cycle
	body.collided = false;
	if body.withTelemetry {
		// TODO print body to console
	}
	if math32.IsNaN(body.x) || math32.IsNaN(body.y) || math32.IsNaN(body.z) {
		// TODO logger
		body.exists = false
	}
	return NewBodyRenderInfo(body)
}

// TODO rename bodyQueue to bodies or some such
func (body *Body) ForceComputer(bodyQueue *cmap.ConcurrentMap, result chan<- bool) {
	// TODO panic/recover
	if body.fragmenting {
		body.fragment(bodyQueue)
	} else {
		body.fx, body.fy, body.fz = 0, 0, 0
		for item := range bodyQueue.IterBuffered() {
			otherBody := item.Val.(*Body)
			if !otherBody.exists || body.fragmenting {
				break
			}
			if body != otherBody && otherBody.exists && !otherBody.fragmenting {
				// todo metrics
				result := body.calcForceFrom(otherBody)
				if result.collided {
					body.resolveCollision(result.dist, otherBody)
				}
			}
		}
	}
	result<- true
}
// end SimBody interface implementation


func (body *Body) subsume(dist float32, otherBody *Body) {
	// TODO
}

func (body *Body) calcForceFrom(otherBody *Body) ForceCalcResult {
	dx := otherBody.x - body.x
	dy := otherBody.y - body.y
	dz := otherBody.z - body.z
	dist := float32(math.Sqrt(float64(dx*dx + dy*dy + dz*dz)))
	if body.collided || dist > body.radius + otherBody.radius {
		force := G * float64(body.mass) * float64(otherBody.mass) / float64(dist * dist)
		body.fx += force * float64(dx) / float64(dist)
		body.fy += force * float64(dy) / float64(dist)
		body.fz += force * float64(dz) / float64(dist)
	} else if dist <= body.radius + otherBody.radius {
		return Collision(dist)  // TODO NOT idiomatic?
	}
	return NoCollision() // TODO NOT idiomatic?
}

func (body *Body) resolveCollision(dist float32, otherBody *Body) {
	if body.collisionBehavior == globals.Subsume || otherBody.collisionBehavior == globals.Subsume {
		if body.radius > otherBody.radius {
			body.subsume(dist, otherBody)
		} else {
			otherBody.subsume(dist, body)
		}
	} else if (body.collisionBehavior == globals.Elastic || body.collisionBehavior == globals.Fragment) &&
		(otherBody.collisionBehavior == globals.Elastic || otherBody.collisionBehavior == globals.Fragment) {
		 //>>> HERE
	}
}

func (body *Body) calcElasticCollision(otherBody *Body) CollisionCalcResult {
	var r12, m21, d, v, theta2, phi2, st, ct, sp, cp, vx1r, vy1r, vz1r, fvz1r,
	thetav, phiv, dr, alpha, beta, sbeta, cbeta, t, a, dvz2,
	vx2r, vy2r, vz2r, x21, y21, z21, vx21, vy21, vz21, vx_cm, vy_cm, vz_cm float64

	m1  := float64(body.mass)
	m2  := float64(otherBody.mass)
	r1  := float64(body.radius)
	r2  := float64(otherBody.radius)
	x1  := float64(body.x)
	y1  := float64(body.y)
	z1  := float64(body.z)
	x2  := float64(otherBody.x)
	y2  := float64(otherBody.y)
	z2  := float64(otherBody.z)
	vx1 := float64(body.vx)
	vy1 := float64(body.vy)
	vz1 := float64(body.vz)
	vx2 := float64(otherBody.vx)
	vy2 := float64(otherBody.vy)
	vz2 := float64(otherBody.vz)
	t = t // unused in source as well but keep as close to the source as possible

	r12 = r1 + r2
	m21 = m2 / m1
	x21 = x2 - x1
	y21 = y2 - y1
	z21 = z2 - z1
	vx21 = vx2 - vx1
	vy21 = vy2 - vy1
	vz21 = vz2 - vz1

	vx_cm = (m1 * vx1 + m2 * vx2) / (m1 + m2)
	vy_cm = (m1 * vy1 + m2 * vy2) / (m1 + m2)
	vz_cm = (m1 * vz1 + m2 * vz2) / (m1 + m2)

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
	theta2 = math.Acos(z2/d)
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
	vx1r = ct * cp * vx1 + ct * sp * vy1 - st * vz1
	vy1r = cp * vy1 - sp * vx1
	vz1r = st * cp * vx1 + st * sp * vy1 + ct * vz1
	fvz1r = vz1r / v
	if fvz1r > 1 {
		// fix for possible rounding errors
		fvz1r=1
	} else if fvz1r < -1 {
		fvz1r=-1
	}
	thetav = math.Acos(fvz1r)
	if vx1r == 0 && vy1r == 0 {
		phiv=0
	} else {
		phiv = math.Atan2(vy1r,vx1r)
	}

	// calculate the normalized impact parameter
	dr = d * math.Sin(thetav) / r12

	// if balls do not collide, do nothing
	if thetav > math.Pi/ 2 || math.Abs(dr) > 1 {
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

	dvz2 = 2 * (vz1r + a * (cbeta * vx1r + sbeta * vy1r)) / ((1 + a * a) * (1 + m21))

	vz2r = dvz2
	vx2r = a * cbeta * dvz2
	vy2r = a * sbeta * dvz2
	vz1r = vz1r - m21 * vz2r
	vx1r = vx1r - m21 * vx2r
	vy1r = vy1r - m21 * vy2r

	// rotate the velocity vectors back and add the initial velocity
	// vector of ball 2 to retrieve the original coordinate system

	return ElasticCollision(
		ct * cp * vx1r - sp * vy1r + st * cp * vz1r + vx2,
		ct * sp * vx1r + cp * vy1r + st * sp * vz1r + vy2,
		ct * vz1r - st * vx1r                       + vz2,
		ct * cp * vx2r - sp * vy2r + st * cp * vz2r + vx2,
		ct * sp * vx2r + cp * vy2r + st * sp * vz2r + vy2,
		ct * vz2r - st * vx2r                       + vz2,
		vx_cm, vy_cm, vz_cm)
}

func (body *Body) shouldFragment(otherBody *Body, r CollisionCalcResult) FragmentationCalcResult {
	if !(body.collisionBehavior == globals.Fragment ||
		otherBody.collisionBehavior == globals.Fragment) {
		return NoFragmentation();
	}
	vThis := body.vx + body.vy + body.vz
	dvThis :=
	 	math32.Abs(body.vx - ((float32(r.vx1 - r.vx_cm)) * R + float32(r.vx_cm))) +
		math32.Abs(body.vy - ((float32(r.vy1 - r.vy_cm)) * R + float32(r.vy_cm))) +
		math32.Abs(body.vz - ((float32(r.vz1 - r.vz_cm)) * R + float32(r.vz_cm)))

	vThisFactor := dvThis / math32.Abs(vThis);
	vOther := otherBody.vx + otherBody.vy + otherBody.vz;
	dvOther :=
		math32.Abs(otherBody.vx - ((float32(r.vx2 - r.vx_cm)) * R + float32(r.vx_cm))) +
		math32.Abs(otherBody.vy - ((float32(r.vy2 - r.vy_cm)) * R + float32(r.vy_cm))) +
		math32.Abs(otherBody.vz - ((float32(r.vz2 - r.vz_cm)) * R + float32(r.vz_cm)))

	vOtherFactor := dvOther / math32.Abs(vOther);

	if body.collisionBehavior == globals.Fragment && vThisFactor > body.fragFactor ||
		otherBody.collisionBehavior == globals.Fragment && vOtherFactor > otherBody.fragFactor {
		return Fragmentation(vThisFactor, vOtherFactor);
	}
	return NoFragmentation()
}

func (body *Body) doFragment(otherBody *Body, fr FragmentationCalcResult) {
	if body.collisionBehavior == globals.Fragment && fr.thisFactor > body.fragFactor {
		body.initiateFragmentation(fr.thisFactor)
	}
	if otherBody.collisionBehavior == globals.Fragment && fr.otherFactor > otherBody.fragFactor {
		otherBody.initiateFragmentation(fr.otherFactor);
	}
}

func (body *Body) initiateFragmentation(fragFactor float32) {
	var fragDelta float32
	if body.fragFactor > 10 {
		fragDelta = 10
	} else {
		fragDelta = fragFactor - body.fragFactor
	}
	fragments := math32.Min(fragDelta * body.fragmentationStep, maxFrags)
	if fragments <= 1 {
		body.collisionBehavior = globals.Fragment
		return
	}
	body.fragmenting = true
	curPos := math32.Vector3{body.x, body.y, body.z}
	volume := fourThirdsPi * body.radius * body.radius * body.radius
	newRadius := math32.Max(math32.Pow(((volume / fragments) * 3) / fourPi, 1/3), .1);
	newMass := body.mass / fragments;
	body.fragInfo = FragInfo{body.radius, newRadius, newMass, int(fragments), curPos}
}

func (body *Body) fragment(bodyQueue *cmap.ConcurrentMap) {
	cnt := 0
	for ; body.fragInfo.fragments > 0; {
		body.fragInfo.fragments--
		v := getVectorEven(body.fragInfo.curPos, body.fragInfo.radius *.9)
		bodyQueue.Set(nextId(), Body{
			id: nextId(),
			x: v.X, y: v.Y, z: v.Z, vx: body.vx, vy: body.vy, vz: body.vz, mass: body.fragInfo.mass, radius: body.fragInfo.newRadius,
			collisionBehavior: globals.Elastic, bodyColor: body.bodyColor, fragFactor: 0, fragmentationStep: 0,
			withTelemetry: false, name: body.name, class: body.class, pinned: false,
		})
		cnt++
		if cnt > maxFragsPerCycle {
			break
		}
	}
	if body.fragInfo.fragments <= 0  {
		// turn this instance into a fragment
		body.mass = body.fragInfo.mass;
		body.radius = body.fragInfo.newRadius;
		body.collisionBehavior = globals.Elastic;
		body.fragmenting = false;
	} else {
		// shrink the body a little each time
		body.radius = body.radius * .9;
	}
}

var idGenerator = struct {
	lock sync.Mutex
	id   int
}{sync.Mutex{}, 0,}

func nextId() (id int) {
	idGenerator.lock.Lock()
	id = idGenerator.id
	idGenerator.id++
	idGenerator.lock.Unlock()
	return
}

func (body *Body) setNotExists() {
	body.mass = 0
	body.exists = false
}
