package body

import "math"

//
// This go file localizes functionality associated with elastic collision calculation. If uses code from:
// https://www.plasmaphysics.org.uk/programs/coll3d_cpp.htm
//

//
// The 'collisionCalcResult' struct holds the output of an elastic collision calculation
//
type collisionCalcResult struct {
	collided bool
	vx1, vy1, vz1 float64
	vx2, vy2, vz2 float64
	vx_cm, vy_cm, vz_cm float64
}

//
// Uses the passed 'collisionCalcResult' struct to assign new velocities to colliding bodies
//
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
// The function is preserved as closely to the original as possible, hence the snake case names
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
		return collisionCalcResult{collided:false}
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
		return collisionCalcResult{collided:false}
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

	return collisionCalcResult{
		collided: true,
		vx1: ct*cp*vx1r-sp*vy1r+st*cp*vz1r+vx2,
		vy1: ct*sp*vx1r+cp*vy1r+st*sp*vz1r+vy2,
		vz1: ct*vz1r-st*vx1r+vz2,
		vx2: ct*cp*vx2r-sp*vy2r+st*cp*vz2r+vx2,
		vy2: ct*sp*vx2r+cp*vy2r+st*sp*vz2r+vy2,
		vz2: ct*vz2r-st*vx2r+vz2,
		vx_cm: vx_cm,
		vy_cm: vy_cm,
		vz_cm: vz_cm,
	}
}

