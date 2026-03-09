package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

func cancellationWorker(ctx context.Context, wg *sync.WaitGroup, id int) {
	defer wg.Done()

	timer := time.NewTimer(500 * time.Millisecond)
	defer timer.Stop()

	counter := 1
	for {
		select {
		case <-timer.C:
			fmt.Printf("worker %d did job %d \n", id, counter)
			counter++
			timer.Reset(500 * time.Millisecond)
		case <-ctx.Done():
			fmt.Printf("goodbye from worker %d \n", id)
			return
		}
	}
}

func initContextCancellation() {

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go cancellationWorker(ctx, &wg, i)
	}

	wg.Wait()
	fmt.Println("main signing off")
}
