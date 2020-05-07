package body

import (
	"nbodygo/cmd/globals"
	"sync"
	"testing"
)

func TestMod(t *testing.T) {
	b := createTestBody(1)
	mods := []string{
		"x=41",
		"y=42",
		"z=43",
		"vx=44",
		"vy=45",
		"vz=46",
		"mass=47",
		"radius=48",
		"frag-factor=49",
		"frag-step=50",
		"collision=subsume", // elastic by default
		"color=green",       // red by default
		"telemetry=true",    // false by default
		"exists=false",      // true by default
	}
	b.ApplyMods(mods)
	failed := b.X != 41 ||
		b.Y != 42 ||
		b.Z != 43 ||
		b.Vx != 44 ||
		b.Vy != 45 ||
		b.Vz != 46 ||
		b.Mass != 47 ||
		b.Radius != 48 ||
		b.FragFactor != 49 ||
		b.FragStep != 50 ||
		b.CollisionBehavior != globals.Subsume ||
		b.BodyColor != globals.Green ||
		!b.WithTelemetry ||
		b.Exists
	if failed {
		t.Error("Mod body failed")
	}
}

//
// Tests the ID generator concurrently
//
func TestIdGen(t *testing.T) {
	wg := sync.WaitGroup{}
	const idCnt = 1000
	const funcCnt = 10
	wg.Add(funcCnt)
	for i := 0; i < funcCnt; i++ {
		go func() {
			for i := 0; i < idCnt; i++ {
				NextId()
			}
			wg.Done()
		}()
	}
	wg.Wait()
	nxtId := NextId()
	if nxtId != funcCnt * idCnt {
		t.Error("Incorrect ID gen")
	}
}