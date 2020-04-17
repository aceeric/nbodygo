package cmap

import (
	"math/rand"
	"testing"
	"time"
)

type TestBody struct {
	id int
}

// tests concurrent add, remove, and iterate of the map
func TestConcurrent(t *testing.T) {
	bodyQueue := New()
	rand.Seed(time.Now().UnixNano())
	sleepMillis := 10
	var bodyId int
	for bodyId = 0; bodyId < 100; bodyId++ {
		bodyQueue.Set(bodyId, &TestBody{bodyId})
	}
	c := make(chan bool)
	// this goroutine iterates
	go func() {
		start := time.Now()
		for {
			itemCnt := 0
			for range bodyQueue.IterBuffered() {
				itemCnt++
			}
			//t.Logf("item cnt: %v\n", itemCnt)
			duration := time.Now().Sub(start)
			if duration.Seconds() >= 5 {
				c <- true
				return
			}
			time.Sleep(time.Millisecond * time.Duration(rand.Intn(sleepMillis)))
		}
	}()
	// this goroutine removes items randomly
	removed := 0
	go func() {
		start := time.Now()
		for {
			idx := rand.Intn(bodyQueue.Count() - 1)
			var keys []int
			for key, _ := range bodyQueue.Items() {
				keys = append(keys, key)
			}
			item := bodyQueue.Items()[keys[idx]]
			if item == nil {
				t.Error("Unable to access bodyQueue as expected")
			}
			bodyQueue.Remove(item.(*TestBody).id) // Items() we have to access as pointer but Iter not!?!
			removed++
			duration := time.Now().Sub(start)
			if duration.Seconds() >= 5 {
				c <- true
				return
			}
			time.Sleep(time.Millisecond * time.Duration(rand.Intn(sleepMillis)))
		}
	}()
	// this goroutine adds items sequentially (as would happen in the sim)
	added :=0
	go func() {
		start := time.Now()
		for {
			bodyQueue.Set(bodyId, &TestBody{bodyId})
			bodyId++
			added++
			duration := time.Now().Sub(start)
			if duration.Seconds() >= 5 {
				c <- true
				return
			}
			time.Sleep(time.Millisecond * time.Duration(rand.Intn(sleepMillis)))
		}
	}()
	<-c; <-c; <-c
	t.Logf("Queue size: %v. Added: %v. Removed: %v", bodyQueue.Count(), added, removed)
}
