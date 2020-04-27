package sim

import (
	"nbodygo/cmd/body"
	"nbodygo/cmd/globals"
	"nbodygo/cmd/grpcsimcb"
	"nbodygo/cmd/runner"
)

//
// See grpcsimcb package
//
func newGrpcSimCb(bc *body.BodyCollection, crunner *runner.ComputationRunner,
	rqh *runner.ResultQueueHolder) grpcsimcb.GrpcSimCallbacks {
	return grpcsimcb.GrpcSimCallbacks{
		SetResultQueueSize: func(maxQueues int) bool {
			return rqh.Resize(maxQueues)
		},
		ResultQueueSize: func() int {
			max, _ := rqh.MaxQueues()
			return max
		},
		SetSmoothing: func(timeScale float64) {
			crunner.SetTimeScaling(timeScale)
		},
		Smoothing: func() float64 {
			return crunner.TimeScaling()
		},
		SetComputationWorkers: func(count int) {
			crunner.SetWorkers(count)
		},
		ComputationWorkers: func() int {
			return crunner.WorkerCount()
		},
		SetRestitutionCoefficient: func(R float64) {
			crunner.SetCoefficientOfRestitution(R)
		},
		RestitutionCoefficient: func() float64 {
			return crunner.CoefficientOfRestitution()
		},
		RemoveBodies: func(count int) {
			crunner.RemoveBodies(count)
		},
		BodyCount: func() int {
			return bc.Count()
		},
		AddBody: func(mass, x, y, z, vx, vy, vz, radius float64,
			isSun bool, behavior globals.CollisionBehavior, bodyColor globals.BodyColor,
			fragFactor, fragStep float64,
			withTelemetry bool, name, class string,
			pinned bool) int {
			id := body.NextId()
			b := body.NewBody(id, x, y, z, vx, vy, vz, mass, radius, behavior, bodyColor, fragFactor, fragStep,
				withTelemetry, name, class, pinned)
			if isSun {
				b.SetSun(100) // todo support passing intensity
			}
			bc.Enqueue(body.NewAdd(b))
			return id
		},
		ModBody: func(id int, name, class string, mods []string) grpcsimcb.ModBodyResult {
			return bc.ModBody(id, name, class, mods)()
		},
		GetBody: func(id int, name string) interface{} {
			return bc.GetBody(id, name)
		},
	}
}
