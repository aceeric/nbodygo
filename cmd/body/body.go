package body

import (
	"math"
	"nbodygo/cmd/globals"
	"nbodygo/cmd/renderable"
	"nbodygo/cmd/util"
	"strings"
	"sync"
)

const (
	G                float64 = 6.673e-11         // gravitational constant
	fourThirdsPi     float64 = math.Pi * (4 / 3) // used in volume/radius calcs
	fourPi           float64 = math.Pi * 4       // "
	maxFragsPerCycle int     = 100
	maxFrags         float64 = 2000
)

//
// Maintains fragmentation state as a body is fragmenting potentially across compute cycles
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
	id                                int
	name                              string
	class                             string
	fragmenting                       bool
	x, y, z, vx, vy, vz, radius, mass float64
	fragFactor, fragmentationStep     float64
	collisionBehavior                 globals.CollisionBehavior
	bodyColor                         globals.BodyColor
	R                                 float64
	isSun                             bool
	intensity                         float64
	exists                            bool
	withTelemetry                     bool
	pinned                            bool
	fragInfo                          fragInfo
	fx, fy, fz                        float64
	collided                          bool
}

type RawBody struct {
	Id                                int
	Name                              string
	Class                             string
	Fragmenting                       bool
	X, Y, Z, Vx, Vy, Vz, Radius, Mass float64
	FragFactor, FragmentationStep     float64
	CollisionBehavior                 globals.CollisionBehavior
	BodyColor                         globals.BodyColor
	R                                 float64
	IsSun                             bool
	Intensity                         float64
	Exists                            bool
	WithTelemetry                     bool
	Pinned                            bool
	FragInfo                          fragInfo
	Fx, Fy, Fz                        float64
	Collided                          bool
}

// SO BAD !!!!! TODO DELETEME FAST!
func (b *Body) RawBodyFromSimBody() RawBody {
	return RawBody{
		Id:                b.id,
		Name:              b.name,
		Class:             b.class,
		Fragmenting:       b.fragmenting,
		X:                 b.x,
		Y:                 b.y,
		Z:                 b.z,
		Vx:                b.vx,
		Vy:                b.vy,
		Vz:                b.vz,
		Radius:            b.radius,
		Mass:              b.mass,
		FragFactor:        b.fragFactor,
		FragmentationStep: b.fragmentationStep,
		CollisionBehavior: b.collisionBehavior,
		BodyColor:         b.bodyColor,
		R:                 b.R,
		IsSun:             b.isSun,
		Intensity:         b.intensity,
		Exists:            b.exists,
		WithTelemetry:     b.withTelemetry,
		Pinned:            b.pinned,
		FragInfo:          b.fragInfo,
		Fx:                b.fx,
		Fy:                b.fy,
		Fz:                b.fz,
		Collided:          b.collided,
	}
}

//
// creates a body that exists with hard-coded values and properties typically of
// interest specified as function args
//
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
		R:                 1,
		isSun:             false,
		intensity:         0,
		exists:            true,
		withTelemetry:     withTelemetry,
		pinned:            pinned,
		fragInfo:          fragInfo{},
	}
	return b
}

// begin SimBody and Renderable interface implementation(s)

func (b *Body) Id() int                      { return b.id }
func (b *Body) Name() string                 { return b.name }
func (b *Body) Exists() bool                 { return b.exists }
func (b *Body) X() float32                   { return float32(b.x) }
func (b *Body) Y() float32                   { return float32(b.y) }
func (b *Body) Z() float32                   { return float32(b.z) }
func (b *Body) Radius() float64              { return b.radius }
func (b *Body) IsSun() bool                  { return b.isSun }
func (b *Body) Intensity() float32           { return float32(b.intensity) }
func (b *Body) SetSun(intensity float64)     { b.isSun = true; b.intensity = intensity }
func (b *Body) SetR(R float64)               { b.R = R }
func (b *Body) IsPinned() bool               { return b.pinned }
func (b *Body) BodyColor() globals.BodyColor { return b.bodyColor }
func (b *Body) SetCollisionBehavior(behavior globals.CollisionBehavior) {
	b.collisionBehavior = behavior
}
func (b *Body) SetNotExists() {
	b.mass = 0
	b.exists = false
}

func (b *Body) Update(timeScaling, R float64) renderable.Renderable {
	if !b.exists {
		return renderable.NewFromRenderable(b)
	}
	if !b.collided { // KEEP!
		b.vx += timeScaling * b.fx / b.mass
		b.vy += timeScaling * b.fy / b.mass
		b.vz += timeScaling * b.fz / b.mass
	}
	b.x += timeScaling * b.vx
	b.y += timeScaling * b.vy
	b.z += timeScaling * b.vz
	// clear collided flag for next cycle
	b.collided = false
	b.R = R
	if b.withTelemetry {
		// TODO print b to console
	}
	if math.IsNaN(b.x) || math.IsNaN(b.y) || math.IsNaN(b.z) {
		// TODO logger
		b.exists = false
	}
	return renderable.NewFromRenderable(b)
}

func (b *Body) Compute(sbc SimBodyCollection) {
	if !b.exists {
		return
	}
	if b.fragmenting {
		b.fragment(sbc)
		return
	}
	b.fx, b.fy, b.fz = 0, 0, 0
	sbc.IterateOnce(func(c SimBody) {
		otherBody := c.(*Body)
		if !otherBody.exists || b.fragmenting {
			return
		}
		if b != otherBody && otherBody.exists && !otherBody.fragmenting {
			// todo metrics
			result := b.calcForceFrom(otherBody)
			if result.collided {
				if (b.collisionBehavior == globals.Elastic || b.collisionBehavior == globals.Fragment) &&
					(otherBody.collisionBehavior == globals.Elastic || otherBody.collisionBehavior == globals.Fragment) {
					sbc.Enqueue(NewCollision(b, otherBody))
				} else if b.collisionBehavior == globals.Subsume || otherBody.collisionBehavior == globals.Subsume {
					if b.radius > otherBody.radius && result.dist <= b.radius {
						sbc.Enqueue(NewSubsume(b, otherBody))
					} else if otherBody.radius > b.radius && result.dist <= otherBody.radius {
						sbc.Enqueue(NewSubsume(otherBody, b))
					}
				}
			}
		}
	})
}

//
// Accumulates gravitational force on this body from other body. Also checks proximity for collision
//
// returns a 'forceCalcResult' with 'collided' true if this body collided with otherBody,
// else false.
//
func (b *Body) calcForceFrom(otherBody *Body) forceCalcResult {
	dx := otherBody.x - b.x
	dy := otherBody.y - b.y
	dz := otherBody.z - b.z
	dist := math.Sqrt(dx*dx + dy*dy + dz*dz)
	if dist > b.radius+otherBody.radius {
		force := G * b.mass * otherBody.mass / (dist * dist)
		b.fx += force * dx / dist
		b.fy += force * dy / dist
		b.fz += force * dz / dist
		return noCollision()
	} else {
		return collision(dist)
	}
}

func (b *Body) ResolveSubsume(otherBody SimBody) {
	ob := otherBody.(*Body)
	var thisMass, otherMass float64
	thisMass = b.mass
	otherMass = ob.mass
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
	ob.SetNotExists()
	// TODO LOGGER
}

func (b *Body) ResolveCollision(otherBody SimBody) {
	ob := otherBody.(*Body)
	if ! b.exists || !ob.exists {
		return
	}
	if b.collisionBehavior == globals.Elastic &&
		(ob.collisionBehavior == globals.Elastic || ob.collisionBehavior == globals.Fragment) {
		r := b.calcElasticCollision(ob)
		if r.collided {
			fcr := b.shouldFragment(ob, r)
			if fcr.shouldFragment {
				b.doFragment(ob, fcr)
			} else {
				b.doElastic(ob, r)
			}
		}
	}
}

func (b *Body) doElastic(otherBody *Body, r collisionCalcResult) {
	b.vx = (r.vx1-r.vx_cm)*b.R + r.vx_cm
	b.vy = (r.vy1-r.vy_cm)*b.R + r.vy_cm
	b.vz = (r.vz1-r.vz_cm)*b.R + r.vz_cm
	otherBody.vx = (r.vx2-r.vx_cm)*b.R + r.vx_cm
	otherBody.vy = (r.vy2-r.vy_cm)*b.R + r.vy_cm
	otherBody.vz = (r.vz2-r.vz_cm)*b.R + r.vz_cm
	b.collided = true
	otherBody.collided = true
}

//
// Elastic collision algorithm from https://www.plasmaphysics.org.uk/programs/coll3d_cpp.htm
// The function is preserved as close to the original as possible, hence the snake case names
// rather than camel case.
//
func (b *Body) calcElasticCollision(otherBody *Body) collisionCalcResult {
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
	// if (d < r12) {return noElasticCollision()}

	// return if relative speed = 0
	if v == 0 {
		// TODO LOGGING
		return noElasticCollision()
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
		return noElasticCollision()
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

	return elasticCollision(
		ct*cp*vx1r-sp*vy1r+st*cp*vz1r+vx2,
		ct*sp*vx1r+cp*vy1r+st*sp*vz1r+vy2,
		ct*vz1r-st*vx1r+vz2,
		ct*cp*vx2r-sp*vy2r+st*cp*vz2r+vx2,
		ct*sp*vx2r+cp*vy2r+st*sp*vz2r+vy2,
		ct*vz2r-st*vx2r+vz2,
		vx_cm, vy_cm, vz_cm)
}

//
// Applies the passed modifications to the body. Supports the ability to change characteristics of
// a body in the simulation while the sim is running. The passed mods are an array of property=value
// strings. E.g.: "sun=true". Or "x=123". Unknown properties are ignored. Parse errors are ignored.
// If an array element is not in property=value form, it is ignored
//
func (b *Body) ApplyMods(mods []string) bool {
	for _, mod := range mods {
		nvp := strings.Split(mod, "=")
		if len(nvp) == 2 {
			switch strings.ToUpper(nvp[0]) {
			case "X":
				b.x = globals.SafeParseFloat(nvp[1], b.x)
			case "Y":
				b.y = globals.SafeParseFloat(nvp[1], b.y)
			case "Z":
				b.z = globals.SafeParseFloat(nvp[1], b.z)
			case "VX":
				b.vx = globals.SafeParseFloat(nvp[1], b.vx)
			case "VY":
				b.vy = globals.SafeParseFloat(nvp[1], b.vy)
			case "VZ":
				b.vz = globals.SafeParseFloat(nvp[1], b.vz)
			case "MASS":
				b.mass = globals.SafeParseFloat(nvp[1], b.mass)
			case "RADIUS":
				b.radius = globals.SafeParseFloat(nvp[1], b.radius)
			case "FRAG_FACTOR":
				b.fragFactor = globals.SafeParseFloat(nvp[1], b.fragFactor)
			case "FRAG_STEP":
				b.fragmentationStep = globals.SafeParseFloat(nvp[1], b.fragmentationStep)
			case "SUN":
				b.isSun = globals.ParseBoolean(nvp[1]) // todo implement in g3napp
			case "COLLISION":
				b.collisionBehavior = globals.ParseCollisionBehavior(nvp[1])
			case "COLOR":
				b.bodyColor = globals.ParseBodyColor(nvp[1])
			case "TELEMETRY":
				b.withTelemetry = globals.ParseBoolean(nvp[1])
			case "EXISTS":
				b.exists = globals.ParseBoolean(nvp[1])
			}
		}
	}
	return true
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
