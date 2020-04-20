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

// TODO attribution
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

// returns a 3-value vector with X, Y, and Z all > 0 and < max
func RandomVector(max int32) math32.Vector3 {
	x := float32(rand.Int31n(max))
	y := float32(rand.Int31n(max))
	z := float32(rand.Int31n(max))
	return math32.Vector3{x, y, z}
}

func NewVector3(x, y, z float64) *Vector3 {
	return &Vector3{X: x, Y: y, Z: z}
}
