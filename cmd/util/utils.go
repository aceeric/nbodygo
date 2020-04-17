package util

import (
	"math/rand"
	"nbodygo/internal/pkg/math32"
)

// returns a 3-value vector with X, Y, and Z all > 0 and < max
func RandomVector(max int32) math32.Vector3 {
	x := float32(rand.Int31n(max))
	y := float32(rand.Int31n(max))
	z := float32(rand.Int31n(max))
	return math32.Vector3{x, y, z}
}
