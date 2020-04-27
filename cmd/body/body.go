package body

import (
	"math"
	"nbodygo/cmd/globals"
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
// creates a body that exists with hard-coded values and properties typically of
// interest specified as function args
//
func NewBody(id int, x, y, z, vx, vy, vz, mass, radius float64, collisionBehavior globals.CollisionBehavior,
	bodyColor globals.BodyColor, fragFactor, fragmentationStep float64, withTelemetry bool, name, class string,
	pinned bool) *Body {
	b := Body{
		Id:          id,
		Name:        name,
		Class:       class,
		collided:    false,
		fragmenting: false,
		X:           x,
		Y:           y,
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

func (b *Body) SetNotExists() {
	b.Mass = 0
	b.Exists = false
}

func (b *Body) SetSun(intensity float64) {
	b.IsSun = true
	b.intensity = intensity
}

func (b *Body) Update(timeScaling, R float64) *Renderable {
	if !b.Exists {
		return NewFromRenderable(b)
	}
	if !b.collided { // KEEP!
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
		// TODO print b to console
	}
	if math.IsNaN(b.X) || math.IsNaN(b.Y) || math.IsNaN(b.Z) {
		// TODO logger
		b.Exists = false
	}
	return NewFromRenderable(b)
}

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
			// todo metrics
			result := b.calcForceFrom(otherBody)
			if result.collided {
				if (b.CollisionBehavior == globals.Elastic || b.CollisionBehavior == globals.Fragment) &&
					(otherBody.CollisionBehavior == globals.Elastic || otherBody.CollisionBehavior == globals.Fragment) {
					bc.Enqueue(NewCollision(b, otherBody))
				} else if b.CollisionBehavior == globals.Subsume || otherBody.CollisionBehavior == globals.Subsume {
					if b.Radius > otherBody.Radius && result.dist <= b.Radius {
						bc.Enqueue(NewSubsume(b, otherBody))
					} else if otherBody.Radius > b.Radius && result.dist <= otherBody.Radius {
						bc.Enqueue(NewSubsume(otherBody, b))
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
	dx := otherBody.X - b.X
	dy := otherBody.Y - b.Y
	dz := otherBody.Z - b.Z
	dist := math.Sqrt(dx*dx + dy*dy + dz*dz)
	if dist > b.Radius+otherBody.Radius {
		force := G * b.Mass * otherBody.Mass / (dist * dist)
		b.fx += force * dx / dist
		b.fy += force * dy / dist
		b.fz += force * dz / dist
		return noCollision()
	} else {
		return collision(dist)
	}
}

func (b *Body) ResolveSubsume(otherBody *Body) {
	var thisMass, otherMass float64
	thisMass = b.Mass
	otherMass = otherBody.Mass
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
	b.Mass = thisMass + otherMass
	otherBody.SetNotExists()
	// TODO LOGGER
}

func (b *Body) ResolveCollision(otherBody *Body) {
	if ! b.Exists || !otherBody.Exists {
		return
	}
	if b.CollisionBehavior == globals.Elastic &&
		(otherBody.CollisionBehavior == globals.Elastic || otherBody.CollisionBehavior == globals.Fragment) {
		r := b.calcElasticCollision(otherBody)
		if r.collided {
			fcr := b.shouldFragment(otherBody, r)
			if fcr.shouldFragment {
				b.doFragment(otherBody, fcr)
			} else {
				b.doElastic(otherBody, r)
			}
		}
	}
}

func (b *Body) doElastic(otherBody *Body, r collisionCalcResult) {
	b.Vx = (r.vx1-r.vx_cm)*b.r + r.vx_cm
	b.Vy = (r.vy1-r.vy_cm)*b.r + r.vy_cm
	b.Vz = (r.vz1-r.vz_cm)*b.r + r.vz_cm
	otherBody.Vx = (r.vx2-r.vx_cm)*b.r + r.vx_cm
	otherBody.Vy = (r.vy2-r.vy_cm)*b.r + r.vy_cm
	otherBody.Vz = (r.vz2-r.vz_cm)*b.r + r.vz_cm
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

	m1 := b.Mass
	m2 := otherBody.Mass
	r1 := b.Radius
	r2 := otherBody.Radius
	x1 := b.X
	y1 := b.Y
	z1 := b.Z
	x2 := otherBody.X
	y2 := otherBody.Y
	z2 := otherBody.Z
	vx1 := b.Vx
	vy1 := b.Vy
	vz1 := b.Vz
	vx2 := otherBody.Vx
	vy2 := otherBody.Vy
	vz2 := otherBody.Vz
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
			case "SUN":
				b.IsSun = globals.ParseBoolean(nvp[1]) // todo implement in g3napp
			case "COLLISION":
				b.CollisionBehavior = globals.ParseCollisionBehavior(nvp[1])
			case "COLOR":
				b.BodyColor = globals.ParseBodyColor(nvp[1])
			case "TELEMETRY":
				b.WithTelemetry = globals.ParseBoolean(nvp[1])
			case "EXISTS":
				b.Exists = globals.ParseBoolean(nvp[1])
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
