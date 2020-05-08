package runner

import (
	"nbodygo/cmd/body"
	"nbodygo/cmd/globals"
	"runtime"
	"testing"
	"time"
)

//
// Tests resizing the work pool
//
func TestWpResize(t *testing.T) {
	bc := body.NewSimBodyCollection([]*body.Body{})
	grtn := runtime.NumGoroutine()
	wp := NewWorkPool(5, bc)
	if runtime.NumGoroutine() != grtn+5 {
		t.Error("Worker pool incorrect number of goroutines")
	}
	b := body.Body{}
	wp.submit(&b)
	wp.SetPoolSize(10)
	wp.submit(&b)
	if runtime.NumGoroutine() != grtn+10 {
		t.Error("Worker pool incorrect number of goroutines")
	}
	wp.SetPoolSize(3) // removes goroutines
	wp.submit(&b)
	time.Sleep(time.Millisecond) // give the runtime a moment
	if runtime.NumGoroutine() != grtn+3 {
		t.Error("Worker pool incorrect number of goroutines")
	}
}

//
// Tests that the worker pool actually computes bodies. Creates two non-colliding bodies with
// zero velocity, submits them to the work pool, then verifies that the body velocities were updated
// to a non-zero value (that means the Compute function on each body was called by the pool)
//
func TestWpCompute(t *testing.T) {
	bodies := [2]*body.Body{}
	bodies[0] = body.NewBody(1, 1, 1, 1, 0, 0, 0, 1, 1, globals.Elastic, globals.Blue,
		0, 0, false, "", "", false)
	bodies[1] = body.NewBody(2, 22, 22, 22, 0, 0, 0, 1, 1, globals.Elastic, globals.Blue,
		0, 0, false, "", "", false)
	bc := body.NewSimBodyCollection(bodies[:])
	wp := NewWorkPool(1, bc)
	wp.submitSlice(bc.GetArray())
	wp.wait()
	bc.GetArray()[0].Update(1, 1)
	bc.GetArray()[1].Update(1, 1)
	if bc.GetArray()[0].Vx == 0 || bc.GetArray()[1].Vx == 0 {
		t.Error("Worker pool did not compute body")
	}
}
