package body

import (
	"math/rand"
	"nbodygo/internal/pkg/math32"
	"time"
)

func getVectorEven(center math32.Vector3, radius float32) math32.Vector3 {
	var x, y, z float32
	rand.Seed(time.Now().UnixNano())
	for d := float32(2); d > 1; {
		x = rand.Float32() * 2 - 1
		y = rand.Float32() * 2 - 1
		z = rand.Float32() * 2 - 1
		d = x*x + y*y + z*z
	}
	return math32.Vector3{((x * radius) + center.X), ((y * radius) + center.Y), ((z * radius) + center.Z)};

}