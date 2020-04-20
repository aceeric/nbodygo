package body

import (
	"nbodygo/cmd/util"
)

type FragInfo struct {
	radius, newRadius, mass float64
	fragments int
	curPos util.Vector3
}
