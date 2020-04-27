package body

import "nbodygo/cmd/globals"

//
// Holds only what is needed to interface with the rendering engine. As such, values are typed
// in accordance with how G3N will use them
//
type Renderable struct {
	Id        int
	Exists    bool
	X, Y, Z   float32
	Radius    float64
	IsSun     bool
	Intensity float32 // of the light source
	BodyColor globals.BodyColor
}

//
// Creates a 'Renderable' from the passed 'Body'
//
func NewFromRenderable(b *Body) *Renderable {
	if !b.Exists {
		return &Renderable{
			Id:     b.Id,
			Exists: false,
		}
	}
	return &Renderable{
		Id:        b.Id,
		Exists:    b.Exists,
		X:         float32(b.X),
		Y:         float32(b.Y),
		Z:         float32(b.Z),
		Radius:    b.Radius,
		IsSun:     b.IsSun,
		Intensity: float32(b.intensity),
		BodyColor: b.BodyColor,
	}
}
