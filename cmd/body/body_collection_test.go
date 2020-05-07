package body

import (
	"nbodygo/cmd/globals"
	"nbodygo/cmd/grpcsimcb"
	"runtime"
	"testing"
	"time"
)

//
// Creates and returns a slice of pointers to bodies with length equal to the passed count. Each
// body's ID is its index into the slice
//
func createTestBodies(cnt int) []*Body {
	bodies := make([]*Body, cnt)
	for i := 0; i < cnt; i++ {
		bodies[i] = createTestBody(i)
	}
	return bodies
}

//
// Creates one body with the passed ID and returns a pointer to it
//
func createTestBody(id int) *Body {
	return &Body{
		Id:                id,
		Name:              string(id),
		Class:             string(id),
		X:                 float64(id),
		Y:                 float64(id),
		Z:                 float64(id),
		Vx:                float64(id),
		Vy:                float64(id),
		Vz:                float64(id),
		Radius:            float64(id),
		Mass:              float64(id),
		FragFactor:        float64(id),
		FragStep:          float64(id),
		CollisionBehavior: globals.Elastic,
		BodyColor:         globals.Red,
		IsSun:             false,
		Exists:            true,
		WithTelemetry:     false,
		Pinned:            false,
		r:                 1,
		fragmenting:       false,
		intensity:         0,
		fragInfo:          fragInfo{},
		fx:                0,
		fy:                0,
		fz:                0,
		collided:          false,
	}
}

//
// Verifies that when a collection is initialized, it doesn't lose any bodies
//
func TestInitSize(t *testing.T) {
	const cnt = 1_000_000
	bc := NewSimBodyCollection(createTestBodies(cnt))
	if bc.Count() != cnt {
		t.Error("Incorrect body count")
	}
}

//
// Sets a body to not exist, then invokes the Cycle method which should remove the body from
// the collection. Verifies that the body was removed
//
func TestRemove(t *testing.T) {
	const cnt = 1_000
	const idToDelete = 10
	bc := NewSimBodyCollection(createTestBodies(cnt))
	bodyArray := bc.GetArray()
	bodyArray[idToDelete].Exists = false
	bc.Cycle(1)
	if bc.Count() != cnt-1 {
		t.Errorf("Body was not removed. Expected count: %v, actual count: %v\n", cnt-1, bc.Count())
	}
	for i, sz := 0, bc.Count(); i < sz; i++ {
		if bc.GetArray()[i].Id == idToDelete {
			t.Error("Body should have been removed")
		}
	}
}

//
// Tests the GetBody function. This is intended for the gRPC interface to request a body via the collection
// API and uses channels internally to enqueue the request and return it to the caller. Tests by ID
//
func TestGetByID(t *testing.T) {
	const cnt = 1_000
	const idToGet = 10
	bc := NewSimBodyCollection(createTestBodies(cnt))
	var b *Body
	ch := make(chan *Body)
	go func() {
		ch <- nil
		b = bc.GetBody(idToGet, "") // blocks until HandleGetBody called
		ch <- b
	}()
	<-ch               // wait for the goroutine to start running
	bc.HandleGetBody() // provide the body to GetBody
	var gotBody *Body
	select {
	case gotBody = <-ch:
	case <-time.After(time.Millisecond):
		t.Error("Body was not gotten")
	}
	if gotBody == nil {
		t.Error("Got nil body")
	} else if gotBody.Id != idToGet {
		t.Error("Got wrong body")
	}
}

//
// Same as 'TestGetByID' except tests by name
//
func TestGetByName(t *testing.T) {
	const cnt = 2_000
	const idToGet = 1_999
	bc := NewSimBodyCollection(createTestBodies(cnt))
	var b *Body
	ch := make(chan *Body)
	go func() {
		ch <- nil
		b = bc.GetBody(-1, string(idToGet)) // blocks until HandleGetBody called
		ch <- b
	}()
	<-ch               // wait for the goroutine to start running
	bc.HandleGetBody() // provide the body to GetBody
	var gotBody *Body
	select {
	case gotBody = <-ch:
	case <-time.After(time.Millisecond):
		t.Error("Body was not gotten")
	}
	if gotBody == nil {
		t.Error("Got nil body")
	} else if gotBody.Id != idToGet {
		t.Error("Got wrong body")
	}
}

//
// Tests the mod body function which - like get body is designed to allow the caller to access
// the collection via the functions on the collection and internally enqueues/synchronizes access to
// the collection. Tests by ID
//
func TestModByID(t *testing.T) {
	const cnt = 1_000
	const idToGet = 10
	bc := NewSimBodyCollection(createTestBodies(cnt))
	ch := make(chan grpcsimcb.ModBodyResult)
	go func() {
		ch <- grpcsimcb.ModAll
		ch <- bc.ModBody(idToGet, "", "", []string{"color=blue"}) // blocks until HandleModBody called
	}()
	<-ch               // wait for the goroutine to start running
	bc.HandleModBody() // provide the body to ModBody
	var modResult grpcsimcb.ModBodyResult
	select {
	case modResult = <-ch:
	case <-time.After(time.Millisecond):
		t.Error("Body was not gotten")
	}
	if modResult != grpcsimcb.ModAll {
		t.Error("Mod body failed")
	}
	b := bc.GetArray()[idToGet]
	if b.BodyColor != globals.Blue {
		t.Error("Body modification failed")
	}
}

//
// Same as 'TestModByID' except tries to modify a body that doesn't exist and verifies that
// nothing happened
//
func TestModByIDNoMatch(t *testing.T) {
	const cnt = 1_000
	const idToGet = 2_000
	bc := NewSimBodyCollection(createTestBodies(cnt))
	ch := make(chan grpcsimcb.ModBodyResult)
	go func() {
		ch <- grpcsimcb.ModAll
		ch <- bc.ModBody(idToGet, "", "", []string{"color=blue"}) // blocks until HandleModBody called
	}()
	<-ch               // wait for the goroutine to start running
	bc.HandleModBody() // provide the body to ModBody
	var modResult grpcsimcb.ModBodyResult
	select {
	case modResult = <-ch:
	case <-time.After(time.Millisecond):
		t.Error("Body was not gotten")
	}
	if modResult != grpcsimcb.NoMatch {
		t.Error("Mod body failed")
	}
}

//
// Same as 'TestModByID' except mods by name
//
func TestModByName(t *testing.T) {
	const cnt = 1_000
	const idToGet = 10
	bc := NewSimBodyCollection(createTestBodies(cnt))
	ch := make(chan grpcsimcb.ModBodyResult)
	go func() {
		ch <- grpcsimcb.ModAll
		ch <- bc.ModBody(-1, string(idToGet), "", []string{"color=blue"}) // blocks until HandleModBody called
	}()
	<-ch               // wait for the goroutine to start running
	bc.HandleModBody() // provide the body to ModBody
	var modResult grpcsimcb.ModBodyResult
	select {
	case modResult = <-ch:
	case <-time.After(time.Millisecond):
		t.Error("Body was not gotten")
	}
	if modResult != grpcsimcb.ModAll {
		t.Error("Mod body failed")
	}
	b := bc.GetArray()[idToGet]
	if b.BodyColor != globals.Blue {
		t.Error("Body modification failed")
	}
}

//
// Same as 'TestModByID' except mods by class
//
func TestModByClass(t *testing.T) {
	const cnt = 1_000
	const idToGet = 10
	bc := NewSimBodyCollection(createTestBodies(cnt))
	ch := make(chan grpcsimcb.ModBodyResult)
	go func() {
		ch <- grpcsimcb.ModAll
		ch <- bc.ModBody(-1, "", string(idToGet), []string{"color=blue"}) // blocks until HandleModBody called
	}()
	<-ch               // wait for the goroutine to start running
	bc.HandleModBody() // provide the body to ModBody
	var modResult grpcsimcb.ModBodyResult
	select {
	case modResult = <-ch:
	case <-time.After(time.Millisecond):
		t.Error("Body was not gotten")
	}
	if modResult != grpcsimcb.ModAll {
		t.Error("Mod body failed")
	}
	b := bc.GetArray()[idToGet]
	if b.BodyColor != globals.Blue {
		t.Error("Body modification failed")
	}
}

//
// Enqueues the addition of a body and verifies that the Cycle method of the collection
// picks up the add
//
func TestAdds(t *testing.T) {
	const cnt = 500
	const idToAdd = 600
	bc := NewSimBodyCollection(createTestBodies(cnt))
	bc.Enqueue(NewAdd(createTestBody(idToAdd)))
	runtime.Gosched()
	bc.Cycle(1)
	for i, sz := 0, bc.Count(); i < sz; i++ {
		if bc.GetArray()[i].Id == idToAdd {
			return
		}
	}
	t.Error("Body was not added")
}

//
// Tests enqueueing and resolving a subsume event. This is how collisions are handled - bodies don't
// mod each other during the force calc tight nested loop - they instead enqueue modifications so all mods
// can be processed in one pass by the computation runner without locking
//
func TestSubsume(t *testing.T) {
	const cnt = 500
	const idSubsumes = 10
	const idSubsumed = 444
	bc := NewSimBodyCollection(createTestBodies(cnt))
	bArr := bc.GetArray()
	bc.Enqueue(newSubsume(bArr[idSubsumes], bArr[idSubsumed]))
	runtime.Gosched()
	bc.ProcessMods()
	bArr = bc.GetArray()
	var bSubsumes, bSubsumed *Body = nil, nil
	for i, sz := 0, bc.Count(); i < sz; i++ {
		if bArr[i].Id == idSubsumes {
			bSubsumes = bArr[i]
		} else if bArr[i].Id == idSubsumed {
			bSubsumed = bArr[i]
		}
	}
	if bSubsumes == nil || bSubsumed == nil {
		t.Error("Failed to find body in array")
	} else if !bSubsumes.Exists {
		t.Error("incorrect event resolution")
	} else if bSubsumed.Exists {
		t.Error("incorrect event resolution")
	}
}

//
// Same as 'TestSubsume' except tests collision resolution by placing two bodies in the same
// location
//
func TestCollide(t *testing.T) {
	const cnt = 5000
	const idOne = 2999
	const idTwo = 3999
	bc := NewSimBodyCollection(createTestBodies(cnt))
	bArr := bc.GetArray()
	bArr[idOne].X, bArr[idOne].Y, bArr[idOne].Z = 500, 500, 500
	bArr[idTwo].X, bArr[idTwo].Y, bArr[idTwo].Z = 500, 500, 500
	bc.Enqueue(newCollision(bArr[idOne], bArr[idTwo]))
	runtime.Gosched()
	bc.ProcessMods()
	bArr = bc.GetArray()
	var bOne, bTwo *Body = nil, nil
	for i, sz := 0, bc.Count(); i < sz; i++ {
		if bArr[i].Id == idOne {
			bOne = bArr[i]
		} else if bArr[i].Id == idTwo {
			bTwo = bArr[i]
		}
	}
	if bOne == nil || bTwo == nil {
		t.Error("Failed to find body in array")
	} else if !bOne.collided || !bTwo.collided {
		t.Error("incorrect event resolution")
	}
}
