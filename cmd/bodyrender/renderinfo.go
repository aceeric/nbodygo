package bodyrender

import (
	"nbodygo/cmd/globals"
	"nbodygo/cmd/interfaces"
)

type bodyRenderInfo struct {
	id int
	exists bool
	x, y, z, radius float64
	isSun bool
	bodyColor globals.BodyColor
}

// implementation of Renderable interface

func (r bodyRenderInfo) Id() int                      {return r.id}
func (r bodyRenderInfo) Exists() bool                 {return r.exists}
func (r bodyRenderInfo) X() float64                   {return r.x}
func (r bodyRenderInfo) Y() float64                   {return r.y}
func (r bodyRenderInfo) Z() float64                   {return r.z}
func (r bodyRenderInfo) X32() float32                 {return float32(r.x)}
func (r bodyRenderInfo) Y32() float32                 {return float32(r.y)}
func (r bodyRenderInfo) Z32() float32                 {return float32(r.z)}
func (r bodyRenderInfo) Radius() float64              {return r.radius}
func (r bodyRenderInfo) Radius32() float32            {return float32(r.radius)}
func (r bodyRenderInfo) IsSun() bool                  {return r.isSun}
func (r bodyRenderInfo) BodyColor() globals.BodyColor {return r.bodyColor}

func NewFromRenderable(r interfaces.Renderable) interfaces.Renderable {
	if !r.Exists() {
		return bodyRenderInfo{
			id: r.Id(),
			exists: false,
		}
	}
	return bodyRenderInfo{
		id: r.Id(),
		exists: r.Exists(),
		x: r.X(), y: r.Y(), z: r.Z(), radius: r.Radius(), isSun: r.IsSun(), bodyColor:r.BodyColor(),
	}
}

func New(id int, exists bool, x, y, z, radius float64, isSun bool, bodyColor globals.BodyColor) interfaces.Renderable {
	return bodyRenderInfo{
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

func NewEmpty() interfaces.Renderable {
	return bodyRenderInfo{
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