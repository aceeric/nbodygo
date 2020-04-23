package util

import (
	"math/rand"
	"nbodygo/internal/pkg/math32"
	"time"
)

type Vector3 struct {
	X float64
	Y float64
	Z float64
}

//
//  Generates a vector that is evenly distributed within a virtual sphere around the vector defined
//  in the passed param. Meaning - if called multiple times, the result will be a set of vectors
//  evenly distributed within a sphere. This function is based on:
//
//  https://karthikkaranth.me/blog/generating-random-points-in-a-sphere/
//
//  args:
//   center The Vector around which to center the generated vector
//   radius The radius within which to generate the vector
//
//  return: a vector as as described
//
func GetVectorEven(center Vector3, radius float64) Vector3 {
	var x, y, z float64
	rand.Seed(time.Now().UnixNano())
	for d := float64(2); d > 1; {
		x = rand.Float64() * 2 - 1
		y = rand.Float64() * 2 - 1
		z = rand.Float64() * 2 - 1
		d = x*x + y*y + z*z
	}
	return Vector3{(x * radius) + center.X, (y * radius) + center.Y, (z * radius) + center.Z};
}

//
// returns a 3-value vector with x, y, and z all randomly generated from  0 to < max
//
func RandomVector(max int32) math32.Vector3 {
	x := float32(rand.Int31n(max))
	y := float32(rand.Int31n(max))
	z := float32(rand.Int31n(max))
	return math32.Vector3{X: x, Y: y, Z:z}
}

//
// Returns a 3-value vector from the passed x, y, z args
//
func NewVector3(x, y, z float64) *Vector3 {
	return &Vector3{X: x, Y: y, Z: z}
}
