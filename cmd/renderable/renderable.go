package renderable

import (
	"nbodygo/cmd/globals"
)

//
// Interface contract for a body to be able to be rendered into the scene graph
//
type Renderable interface {
	Id() int
	Exists() bool
	X() float32
	Y() float32
	Z() float32
	Radius() float64
	IsSun() bool
	Intensity() float32
	BodyColor() globals.BodyColor
}

//
// Holds only what is needed to interface with the rendering engine. As such, values are typed
// in accordance with how G3N will use them
//
type renderable struct {
	id int
	exists bool
	x, y, z float32
	radius float64
	isSun bool
	intensity float32 // of the light source
	bodyColor globals.BodyColor
}

//
// implementation of 'Renderable' interface
//
func (r renderable) Id() int                      {return r.id}
func (r renderable) Exists() bool                 {return r.exists}
func (r renderable) X() float32                   {return float32(r.x)}
func (r renderable) Y() float32                   {return float32(r.y)}
func (r renderable) Z() float32                   {return float32(r.z)}
func (r renderable) Radius() float64              {return r.radius}
func (r renderable) IsSun() bool                  {return r.isSun}
func (r renderable) Intensity() float32           {return float32(r.intensity)}
func (r renderable) BodyColor() globals.BodyColor {return r.bodyColor}

//
// Creates a 'renderable' struct from a 'Renderable'
//
func NewFromRenderable(r Renderable) Renderable {
	if !r.Exists() {
		return &renderable{
			id: r.Id(),
			exists: false,
		}
	}
	return &renderable{
		id: r.Id(),
		exists: r.Exists(),
		x: r.X(), y: r.Y(), z: r.Z(),
		radius: r.Radius(),
		isSun: r.IsSun(),
		intensity: r.Intensity(),
		bodyColor:r.BodyColor(),
	}
}

//
// Creates a new 'renderable' struct from the passed values
//
func New(id int, exists bool, x, y, z float32, radius float64, isSun bool, intensity float32,
	bodyColor globals.BodyColor) Renderable {
	return renderable{
		id:        id,
		exists:    exists,
		x:         x,
		y:         y,
		z:         z,
		radius:    radius,
		isSun:     isSun,
		intensity: intensity,
		bodyColor: bodyColor,
	}
}
