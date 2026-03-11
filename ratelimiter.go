package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

func sharedWorker(id int, tokens chan int, wg *sync.WaitGroup) {
	for token := range tokens {
		fmt.Printf("worker %d got token %d \n", id, token)
		ms := rand.Float32()*100 + 100
		time.Sleep(time.Duration(ms) * time.Millisecond)
		now := time.Now().UnixMilli()
		fmt.Printf("work %d done at %d \n", id, now)
		wg.Done()
		return // worker returns once its work is done
	}
}

func initRateLimiter() {
	tokens := make(chan int, 3)
	defer close(tokens)

	var wg sync.WaitGroup

	// start workers
	for i := 1; i <= 20; i++ {
		wg.Add(1)
		go sharedWorker(i, tokens, &wg)
	}

	ticker := time.NewTicker(time.Duration(1) * time.Second)
	defer ticker.Stop() // this also cleans up the goroutine that replenishes tokens

	// bootstrap channel with initial token
	for i := 1; i <= 3; i++ {
		tokens <- i
	}

	currTokenID := 4
	// goroutine to replish token
	go func() {
		for range ticker.C {
			// snapshot length so that a worker concurrently consuming a token doesn't cause us to replenish more than
			// what is allowed
			snapshottedLen := len(tokens)
			for snapshottedLen < 3 {
				tokens <- currTokenID
				currTokenID++
				snapshottedLen++
			}
		}
	}()

	wg.Wait()
}
