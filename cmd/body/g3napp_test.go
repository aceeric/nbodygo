package body

import (
	"nbodygo/cmd/globals"
	"testing"
	"time"
)

// just a very brute-force test on running the G3N
func TestG3nApp(t *testing.T) {
	simDone := make(chan bool)
	rqh := NewResultQueueHolder(1)
	StartG3nApp(2560, 1440, rqh, simDone)
	time.Sleep(time.Second * 2)
	// now add a body and see (visually) if it got injected into the sim
	rq, _ := rqh.newQueue(1)
	rq.addRenderInfo(BodyRenderInfo{
		id: 1,
		exists:true,
		x: 0, y: 0, z: 0, radius: 5,
		isSun: false,
		bodyColor: globals.Green,
	})
	rq.computed = true
	<-simDone
}