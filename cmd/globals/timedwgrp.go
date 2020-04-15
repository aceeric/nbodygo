package globals

import "sync"

type TimedWaitGroup struct {
	wg  *sync.WaitGroup
	chn chan bool
}

func NewTimedWaitGroup(waiters int) *TimedWaitGroup {
	twg := TimedWaitGroup {
		wg:  new(sync.WaitGroup),
		chn: make(chan bool),
	}
	twg.wg.Add(waiters)
	return &twg
}

func (twg *TimedWaitGroup) Channel() chan bool {
	return twg.chn
}

func (twg *TimedWaitGroup) Done() {
	twg.wg.Done()
}

func (twg *TimedWaitGroup) GoWait() {
	go func() {
		twg.wg.Wait()
		close(twg.chn)
	}()
}