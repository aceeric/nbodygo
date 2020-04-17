package bodyrender

import "nbodygo/cmd/globals"

type BodyRenderInfo struct {
	id int
	exists bool
	x, y, z, radius float32
	isSun bool
	bodyColor globals.BodyColor
}

type Renderable interface {
	Id() int
	Exists() bool
	X() float32
	Y() float32
	Z() float32
	Radius() float32
	IsSun() bool
	BodyColor() globals.BodyColor
}

// implementation of Renderable interface

func (r BodyRenderInfo) Id() int {return r.id}
func (r BodyRenderInfo) Exists() bool {return r.exists}
func (r BodyRenderInfo) X() float32 {return r.x}
func (r BodyRenderInfo) Y() float32 {return r.y}
func (r BodyRenderInfo) Z() float32 {return r.z}
func (r BodyRenderInfo) Radius() float32 {return r.radius}
func (r BodyRenderInfo) IsSun() bool {return r.isSun}
func (r BodyRenderInfo) BodyColor() globals.BodyColor {return r.bodyColor}

func NewFromRenderable(r Renderable) Renderable {
	if !r.Exists() {
		return BodyRenderInfo{
			id: r.Id(),
			exists: false,
		}
	}
	return BodyRenderInfo{
		id: r.Id(),
		exists: r.Exists(),
		x: r.X(), y: r.Y(), z: r.Z(), radius: r.Radius(), isSun: r.IsSun(), bodyColor:r.BodyColor(),
	}
}

func New(id int, exists bool, x, y, z, radius float32, isSun bool, bodyColor globals.BodyColor) Renderable {
	return BodyRenderInfo{
		id:        id,
		exists:    exists,
		x:         x,
		y:         y,
		z:         z,
		radius:    radius,
		isSun:     isSun,
		bodyColor: bodyColor,
	}
}

func NewEmpty() Renderable {
	return BodyRenderInfo{
		id:        0,
		exists:    true,
		x:         0,
		y:         0,
		z:         0,
		radius:    0,
		isSun:     false,
		bodyColor: globals.Random,
	}
}