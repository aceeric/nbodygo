package body

import (
	"math/rand"
	"nbodygo/cmd/cmap"
	"nbodygo/cmd/globals"
	"nbodygo/cmd/util"
	"nbodygo/internal/pkg/math32"
	"testing"
	"time"
)

var testBody = Body{
	id:                0,
	name:              "b1",
	class:             "c1",
	collided:          false,
	fragmenting:       false,
	x:                 0,
	y:                 0,
	z:                 0,
	vx:                0,
	vy:                0,
	vz:                0,
	radius:            10,
	mass:              1E10,
	fx:                0,
	fy:                0,
	fz:                0,
	fragFactor:        0,
	fragmentationStep: 0,
	collisionBehavior: globals.Elastic,
	bodyColor:         globals.Red,
	R:                 1,
	isSun:             false,
	exists:            true,
	lock:              0,
	withTelemetry:     false,
	pinned:            false,
	fragInfo:          FragInfo{0, 0, 0, 0, math32.Vector3{0, 0, 0}},
}

// just testing basic mechanics
func TestBodyCreate1(t *testing.T) {
	tb1 := Body{
		id:                1,
		name:              "b1",
		class:             "c1",
		collided:          false,
		fragmenting:       false,
		x:                 0,
		y:                 0,
		z:                 0,
		vx:                0,
		vy:                0,
		vz:                0,
		radius:            10,
		mass:              1E10,
		fx:                0,
		fy:                0,
		fz:                0,
		fragFactor:        0,
		fragmentationStep: 0,
		collisionBehavior: globals.Elastic,
		bodyColor:         globals.Red,
		R:                 1,
		isSun:             false,
		exists:            true,
		lock:              0,
		withTelemetry:     false,
		pinned:            false,
		fragInfo:          FragInfo{0, 0, 0, 0, math32.Vector3{0, 0, 0}},
	}
	t.Logf("%+v\n", tb1)
}

// just testing basic mechanics
func TestBodyCreate2(t *testing.T) {
	b := NewBody(1, 0, 0, 0, 0, 0, 0, 1E20, 100, globals.Elastic, globals.Blue,
		0, 0, false, "foo", "bar", false)
	_ = b
	//t.Logf("%+v\n", b)
}

// just testing basic mechanics
func TestBodyCreate3(t *testing.T) {
	tb := testBody
	//t.Logf("%+v\n", tb)
	//t.Logf("Same values? %v\n", tb == testBody)   // true, := copies values
	//t.Logf("Same object? %v\n", &tb == &testBody) // false := creates new instance
	_ = tb
}

// initializes some bodies and directly runs the force computation the way the computation runner
// would. Verifies that force was set to non-zero on all bodies
func TestBodyForceCalcDirect(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	bodies := [10000]*Body{}
	for i := range bodies {
		b := testBody.CopyOf()
		b.id = i
		v := randomVector(100)
		b.x = v.X
		b.y = v.Y
		b.z = v.Z
		b.mass *= 1 + float32(rand.Int31n(3))
		bodies[i] = b
	}
	for _, bodyOuter := range bodies {
		for _, bodyInner := range bodies {
			if &bodyOuter != &bodyInner {
				bodyOuter.calcForceFrom(bodyInner)
			}
		}
	}
	for _, body := range bodies {
		body.Update(.000000001)
		if body.fx == 0 {
			t.Error("Force computation failed")
			break
		}
	}
}

// Just a brute force test of the body force computer to make sure it doesn't crash
func TestBodyForceComputer(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	bodyQueue := cmap.New()
	for i := 0; i < 2000; i++ {
		b := testBody.CopyOf()
		b.id = i
		v := util.RandomVector(100)
		b.x = v.X
		b.y = v.Y
		b.z = v.Z
		v = util.RandomVector(100000000)
		b.vx = v.X
		b.vy = v.Y
		b.vz = v.Z
		b.mass *= 1 + float32(rand.Int31n(3))
		b.radius *= 1 + float32(rand.Int31n(3))
		bodyQueue.Set(i, b) // b is already a pointer
	}
	start := time.Now()
	var stop time.Time
	ch := make(chan bool)
	computations := int64(0)
	go func() {
		for {
			for item := range bodyQueue.IterBuffered() {
				b := item.Val.(SimBody) // no pointer
				b.ForceComputer(&bodyQueue)
			}
			for item := range bodyQueue.IterBuffered() {
				b := item.Val.(SimBody) // no pointer
				b.Update(.000000001)
			}
			computations++
			stop = time.Now()
			if stop.Sub(start).Seconds() >= 10 {
				ch <- true
				return
			}
		}
	}()
	<-ch
	millis := stop.Sub(start).Milliseconds()
	millisPerComputation := millis / computations
	_ = millisPerComputation
	//t.Logf("milliseconds per compute cycle: %v\n", millisPerComputation)
}
