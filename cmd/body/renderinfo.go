package body

import "nbodygo/cmd/globals"

type BodyRenderInfo struct {
	id int
	exists bool
	x, y, z, radius float32
	isSun bool
	bodyColor globals.BodyColor
}

func NewBodyRenderInfo(b *Body) BodyRenderInfo {
	if !b.exists {
		return BodyRenderInfo{
			id: b.id,
			exists: false,
		}
	}
	return BodyRenderInfo{
		id: b.id,
		exists: true,
		x: b.x, y: b.y, z: b.z, radius: b.radius, isSun: b.isSun, bodyColor:b.bodyColor,
	}
}
// TODO REMOVE THIS
func NewNonExistentBodyRenderInfo(b *Body) BodyRenderInfo {
	return BodyRenderInfo{
		id: b.id,
		exists: false,
	}
}
