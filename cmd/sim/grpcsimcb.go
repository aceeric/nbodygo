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
func newGrpcSimCb(sbc body.SimBodyCollection, crunner *runner.ComputationRunner,
	rqh runner.ResultQueueHolder) grpcsimcb.GrpcSimCallbacks {
	return grpcsimcb.GrpcSimCallbacks{
		SetResultQueueSize: nil, // TODO
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
		SetComputationWorkers: nil, // TODO
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
			return sbc.Count()
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
			sbc.Enqueue(body.NewAdd(&b))
			return id
		},
		ModBody: nil, // TODO
		GetBody: func(id int, name string) grpcsimcb.BodyRaw {
			b := sbc.GetBody(id, name)()
			if b == nil {
				return grpcsimcb.BodyRaw{Id:-1}
			}
			bb := b.(*body.Body).RawBodyFromSimBody() // TODO SO BAD FIX FIX FIX!!!
			return grpcsimcb.BodyRaw{
				Id:                int64(bb.Id),
				X:                 float32(bb.X),
				Y:                 float32(bb.Y),
				Z:                 float32(bb.Z),
				Vx:                float32(bb.Vx),
				Vy:                float32(bb.Vy),
				Vz:                float32(bb.Vz),
				Mass:              float32(bb.Mass),
				Radius:            float32(bb.Radius),
				IsSun:             bb.IsSun,
				CollisionBehavior: bb.CollisionBehavior,
				BodyColor:         bb.BodyColor,
				FragFactor:        float32(bb.FragFactor),
				FragStep:          float32(bb.FragmentationStep),
				WithTelemetry:     bb.WithTelemetry,
				Name:              bb.Name,
				Class:             bb.Class,
				Pinned:            bb.Pinned,
			}
		},
	}
}
