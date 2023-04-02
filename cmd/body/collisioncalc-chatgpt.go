package body

import (
	"math"
)

type Vector struct {
	X, Y, Z float64
}

type Sphere struct {
	Position Vector
	Velocity Vector
	Radius   float64
}

type gpt_collisionCalcResult struct {
	collided      bool
	vx1, vy1, vz1 float64
	vx2, vy2, vz2 float64
}

func (b *Body) gpt_doElastic(otherBody *Body, r gpt_collisionCalcResult) {
	b.Vx = r.vx1
	b.Vy = r.vy1
	b.Vz = r.vz1
	otherBody.Vx = r.vx2
	otherBody.Vy = r.vy2
	otherBody.Vz = r.vz2
	b.collided = true
	otherBody.collided = true
}

// Uses the passed 'collisionCalcResult' struct to assign new velocities to colliding bodies
func (b *Body) gpt_calcElasticCollision(otherBody *Body) gpt_collisionCalcResult {
	// Calculate the vector between the spheres
	dx := otherBody.X - b.X
	dy := otherBody.Y - b.Y
	dz := otherBody.Z - b.Y

	// Calculate the distance between the spheres
	dist := math.Sqrt(dx*dx + dy*dy + dz*dz)

	// Calculate the unit normal and tangent vectors
	unitNormal := []float64{dx / dist, dy / dist, dz / dist}
	unitTangent := []float64{-unitNormal[1], unitNormal[0], 0}

	// Calculate the velocity components in the normal and tangent directions
	var v1n, v1t, v2n, v2t float64
	v1n = unitNormal[0]*b.Vx + unitNormal[1]*b.Vy + unitNormal[2]*b.Vz
	v1t = unitTangent[0]*b.Vx + unitTangent[1]*b.Vy + unitTangent[2]*b.Vz
	v2n = unitNormal[0]*otherBody.Vx + unitNormal[1]*otherBody.Vy + unitNormal[2]*otherBody.Vz
	v2t = unitTangent[0]*otherBody.Vx + unitTangent[1]*otherBody.Vy + unitTangent[2]*otherBody.Vz

	// Calculate the new normal velocities using the elastic collision equation
	newV1n := (b.Mass*v1n + otherBody.Mass*(2*v2n-v1n)) / (b.Mass + otherBody.Mass)
	newV2n := (otherBody.Mass*v2n + b.Mass*(2*v1n-v2n)) / (b.Mass + otherBody.Mass)

	// Calculate the new tangential velocities
	newV1t := v1t
	newV2t := v2t

	//// Convert the normal and tangent velocities back to the x, y, z coordinate system
	//b.Vx = newV1n*unitNormal[0] + newV1t*unitTangent[0]
	//b.Vy = newV1n*unitNormal[1] + newV1t*unitTangent[1]
	//b.Vz = newV1n*unitNormal[2] + newV1t*unitTangent[2]
	//otherBody.Vx = newV2n*unitNormal[0] + newV2t*unitTangent[0]
	//otherBody.Vy = newV2n*unitNormal[1] + newV2t*unitTangent[1]
	//otherBody.Vz = newV2n*unitNormal[2] + newV2t*unitTangent[2]

	return gpt_collisionCalcResult{
		collided: true,
		vx1:      newV1n*unitNormal[0] + newV1t*unitTangent[0],
		vy1:      newV1n*unitNormal[1] + newV1t*unitTangent[1],
		vz1:      newV1n*unitNormal[2] + newV1t*unitTangent[2],
		vx2:      newV2n*unitNormal[0] + newV2t*unitTangent[0],
		vy2:      newV2n*unitNormal[1] + newV2t*unitTangent[1],
		vz2:      newV2n*unitNormal[2] + newV2t*unitTangent[2],
	}
}
