package sim

import (
	"nbodygo/cmd/body"
	"nbodygo/cmd/grpcsimcb"
	"nbodygo/cmd/runner"
)

//
// See grpcsim.GrpcSimCallbacks
//
func newGrpcSimCb(sbc body.SimBodyCollection, crunner *runner.ComputationRunner, rqh runner.ResultQueueHolder) grpcsimcb.GrpcSimCallbacks {
	return grpcsimcb.GrpcSimCallbacks{
		ComputationWorkers: func() int {
			return crunner.WorkerCount()
		},
		ResultQueueSize: func() int {
			return rqh.MaxQueues()
		},
		Smoothing: func() float64 {
			return crunner.TimeScaling()
		},
		RestitutionCoefficient: func() float64 {
			return body.RestitutionCoefficient()
		},
		BodyCount: func() int {
			return sbc.Count()
		},
	}
}
