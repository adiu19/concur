package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type job struct {
	id       int
	duration float32
}

// bundled worker execution
func bundledWorker(workerID int, inCh chan job, outCh chan job) {
	for job := range inCh {
		fmt.Printf("worker %d polled job %d and now sleeping for %f \n", workerID, job.id, job.duration)
		time.Sleep(time.Duration(job.duration) * time.Millisecond)
		outCh <- job
	}
}

func outputWorker(workerID int, outCh chan job, wg *sync.WaitGroup) {
	for job := range outCh {
		fmt.Printf("output worker %d printing result for job id %d with duration %f \n", workerID, job.id, job.duration)
		wg.Done()
	}
}
func initWorkerPool() {
	inCh := make(chan job)
	outCh := make(chan job)
	var wg sync.WaitGroup // wait on all print tasks being done

	// spawn workers
	for i := 1; i <= 3; i++ {
		go bundledWorker(i, inCh, outCh)
	}

	// a separate goroutine responsible for printing
	go outputWorker(99, outCh, &wg)

	// each iteration is a job that needs to be sent
	for i := 1; i <= 15; i++ {
		ms := rand.Float32()*400 + 100
		wg.Add(1)
		inCh <- job{id: i, duration: ms}
	}
	close(inCh)
	wg.Wait()
	close(outCh)
}
