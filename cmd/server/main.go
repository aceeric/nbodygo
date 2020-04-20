package main

import (
	"nbodygo/cmd/globals"
	"nbodygo/cmd/sim"
	"nbodygo/internal/pkg/math32"
)

func main() {

	bodies, workerFunc := sim.Generate("Sim1", 3000, globals.Elastic, globals.Random, "50")
	sim.NewNBodySimBuilder().
	Bodies(bodies).
	Threads(9).
	Scaling(.000000002).
	InitialCam(*math32.NewVector3(10, 100, 800)).
	SimThread(workerFunc).
	Render(true).
	Resolution([2]int{2560, 1405}).
	VSync(true).
	FrameRate(-1).
	Build().
	Run()
}
