package g3napp

import (
	"nbodygo/cmd/bodyrender"
	"nbodygo/cmd/globals"
	"nbodygo/cmd/runner"
	"testing"
	"time"
)

//
// Just a very rudimentary test of running the G3N App, and demonstrating all the basic mechanics.
// Performs the following steps:
//
//  1) Creates a window
//  2) Starts the app (which is a goroutine)
//  3) Populates render queues in a loop for a fixed number of iterations. The result should be moving bodies
//     in the window.
//  4) Stops the engine and waits for the app goroutine to signal completed
//
// This test is skipped because it's intended as a manual test: it opens up a window and requires
// a person to observe it to verify that it works. So its not suitable as an automated test
//
func TestG3nApp(t *testing.T) {
	t.Skip()
	simDone := make(chan bool)
	rqh := runner.NewResultQueueHolder(1)
	const framesPerSecond = 16
	StartG3nApp(2560, 1440, rqh, simDone)
	for i := 0; i < 5000/framesPerSecond; i++ {
		rq, _ := rqh.NewResultQueue(1)
		rq.AddRenderable(bodyrender.New(1, true, float32(-i)*.1, 0, 0, 5, false, globals.Green))
		rq.AddRenderable(bodyrender.New(2, true, 20, 20+(float32(-i)*.1), 20, 5, true, 0))
		rq.SetComputed()
		time.Sleep(time.Millisecond * 16)
	}
	StopG3nApp()
	<-simDone
}
