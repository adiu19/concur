package main

import (
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type cbstate int

const (
	open     cbstate = iota
	closed           = 1
	halfopen         = 2
)

type cbinfo struct {
	state                    cbstate
	consecutiveErrs          int
	t                        *time.Timer
	mu                       sync.RWMutex
	done                     chan bool
	allowSingleReqOnHalfOpen bool
}

/*
1. goroutine that checks every 5 seconds, and once woken up, checks cb state. if cb state = open, moves it to half open
2.
*/

func (cb *cbinfo) monitor() {
	for {
		select {
		case _ = <-cb.done:
			return
		case _ = <-cb.t.C:
			cb.mu.Lock()
			if cb.state == open {
				fmt.Println("setting cb state to halfopen.....")
				cb.state = halfopen
				cb.allowSingleReqOnHalfOpen = true
			}
			cb.mu.Unlock()
			// cb.t.Reset(time.Duration(5) * time.Second) // reset timer to begin the next 5 second timer
			// problem is that this is not beginning after the cb was moved to open
			// so probably, this needs to move to a separate place, wherein we
			// check for 3 consecutive failure
			// or check for one failure after halfopen cb
		}
	}
}

func (cb *cbinfo) shouldOpenCB(currentRequestErrs bool) bool {
	return (cb.state == closed && cb.consecutiveErrs >= 3) || (cb.state == halfopen && currentRequestErrs)
}

func (cb *cbinfo) shouldAllowRequestToProceed() bool {

	if cb.state == closed {
		return true
	}

	if cb.state == halfopen {
		cb.mu.Lock()
		if cb.allowSingleReqOnHalfOpen {
			fmt.Println("allowing single request to pass through.....")

			cb.allowSingleReqOnHalfOpen = false
			cb.mu.Unlock()
			return true
		}
		cb.mu.Unlock()
		return false

	}

	return false

}

// TODO, allowSingular is being updated but not implemented; it should only allow one request to flow in
func (cb *cbinfo) sharedWorkToGateAccessOn(workerID int) error {
	if !cb.shouldAllowRequestToProceed() {
		return errors.New("denied")
	}

	currentRequestErrs := rand.Float32() <= 0.4
	fmt.Printf("request for worker %d had failure status %t \n", workerID, currentRequestErrs)
	if !currentRequestErrs {
		cb.mu.Lock()
		cb.consecutiveErrs = 0
		if cb.state == halfopen {
			fmt.Println("setting cb state to closed.....")
			cb.state = closed
		}
		cb.mu.Unlock()
		return nil
	} else {
		cb.mu.Lock()
		cb.consecutiveErrs++
		if cb.shouldOpenCB(currentRequestErrs) {
			fmt.Println("setting cb state to open.....")
			cb.state = open
			cb.t.Reset(time.Duration(5) * time.Second)
		}
		cb.mu.Unlock()
		return nil

	}
}

func (cb *cbinfo) circuitBreakerWorker(workerID int, wg *sync.WaitGroup) {
	defer wg.Done()
	cb.sharedWorkToGateAccessOn(workerID)
}

func initCircuitBreaker() {
	// spawn 15 goroutines
	var wg sync.WaitGroup
	timer := time.NewTimer(5 * time.Second)
	defer timer.Stop()
	cb := cbinfo{
		state:                    closed,
		consecutiveErrs:          0,
		t:                        timer,
		allowSingleReqOnHalfOpen: false,
	}

	go cb.monitor()
	for i := 1; i <= 1000; i++ {
		time.Sleep(time.Duration(200) * time.Millisecond)
		wg.Add(1)
		go cb.circuitBreakerWorker(i, &wg)
	}

	wg.Wait()
	cb.t.Stop()
	cb.done <- true

}
