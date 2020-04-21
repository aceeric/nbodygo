package interfaces

import "nbodygo/cmd/globals"

//
// Interface contract for a body to be able to be rendered into the scene graph
//
type Renderable interface {
	Id() int
	Exists() bool
	X() float64
	Y() float64
	Z() float64
	Radius() float64
	X32() float32
	Y32() float32
	Z32() float32
	Radius32() float32
	IsSun() bool
	Intensity() float32
	BodyColor() globals.BodyColor
}

