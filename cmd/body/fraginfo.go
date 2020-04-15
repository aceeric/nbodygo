package body

import "nbodygo/internal/pkg/math32"

type FragInfo struct {
	radius float32
	newRadius float32
	mass float32
	fragments int
	curPos math32.Vector3
}
