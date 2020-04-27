package runner

import (
	"fmt"
	"nbodygo/cmd/body"
	"runtime"
	"testing"
	"time"
)

func TestWpResize(t *testing.T) {
	fmt.Printf("start goroutines: %v\n", runtime.NumGoroutine())
	sbc := body.NewSimBodyCollection([]body.SimBody{})
	fmt.Printf("after sbc goroutines: %v\n", runtime.NumGoroutine())
	wp := NewWorkPool(5, sbc)
	fmt.Printf("wp=5 goroutines: %v\n", runtime.NumGoroutine())
	b := body.Body{}
	wp.submit(&b)
	wp.SetPoolSize(10)
	wp.submit(&b)
	fmt.Printf("wp=10 goroutines: %v\n", runtime.NumGoroutine())
	wp.SetPoolSize(3)
	wp.submit(&b)
	time.Sleep(time.Second)
	fmt.Printf("wp=3 goroutines: %v\n", runtime.NumGoroutine())
	sbc = nil
	time.Sleep(time.Second)
	fmt.Printf("sbc=nil goroutines: %v\n", runtime.NumGoroutine())
}
