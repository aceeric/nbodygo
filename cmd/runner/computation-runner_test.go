package runner

import (
	"nbodygo/cmd/body"
	"runtime"
	"testing"
	"time"
)

func TestRunnerSetters(t *testing.T) {
	rqh := NewResultQueueHolder(1)
	bc := body.NewSimBodyCollection([]*body.Body{})
	cr := NewComputationRunner(1, 1, rqh, bc)
	go func() {
		for {
			rqh.Next()
			runtime.Gosched()
		}
	}()
	cr.Start()
	cr.SetCoefficientOfRestitution(5)
	cr.SetTimeScaling(5)
	cr.SetWorkers(5)
	time.Sleep(time.Millisecond)
	coef := cr.CoefficientOfRestitution()
	scal := cr.TimeScaling()
	wcnt := cr.WorkerCount()
	if coef != 5 || scal != 5 || wcnt != 5 {
		t.Error("Setters failed")
	}
	cr.Stop()
}
