package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

func worker(in chan float32, out chan float32) {
	for secs := range in {
		time.Sleep(time.Duration(secs) * time.Second)
		out <- secs
	}
}

func initTimeouts() {
	inCh := make(chan float32)
	outCh := make(chan float32)
	defer close(inCh)
	defer close(outCh)

	go worker(inCh, outCh) // spawn worker on a different goroutine, main runs on current goroutine

	// the main runs a loop of sending jobs and waiting with a timer
	for {
		secs := rand.Float32()*2 + 1
		inCh <- secs

		// created everytime because we'd like to have a timer for each job instance
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second) // context

		select {
		case res := <-outCh:
			fmt.Printf("got a result back after sleeping for %f seconds \n", res)
			cancel()
		case <-ctx.Done():
			fmt.Printf("main detected timeout when job scheduler for %f \n", secs)
			cancel()
			return
		}
	}
}
