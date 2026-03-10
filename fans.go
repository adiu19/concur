package main

import (
	"fmt"
	"sync"
)

type fanjob struct {
	id    int
	value float32
}

type fanworker struct {
	id    int
	inCh  chan fanjob
	outCh chan fanjob
}

func printer(outCh chan fanjob, wg *sync.WaitGroup) {
	for job := range outCh {
		fmt.Printf("got %f via job %d \n", job.value, job.id)
		wg.Done() // decrease wg reference counter, since one job is done e2e
	}
}

func collector(done chan bool, outCh chan fanjob, inCh1 chan fanjob, inCh2 chan fanjob, inCh3 chan fanjob, inCh4 chan fanjob) {
	for {
		select {
		case _ = <-done:
			close(outCh) // closes printer too
			return
		case job, ok := <-inCh1:
			if !ok {
				inCh1 = nil
				continue
			}
			outCh <- fanjob{id: job.id, value: job.value}
		case job, ok := <-inCh2:
			if !ok {
				inCh2 = nil
				continue
			}
			outCh <- fanjob{id: job.id, value: job.value}
		case job, ok := <-inCh3:
			if !ok {
				inCh3 = nil
				continue
			}
			outCh <- fanjob{id: job.id, value: job.value}
		case job, ok := <-inCh4:
			if !ok {
				inCh4 = nil
				continue
			}
			outCh <- fanjob{id: job.id, value: job.value}
		}
	}
}

func worker(instance *fanworker) {
	defer close(instance.outCh) // close output channel
	for input := range instance.inCh {
		instance.outCh <- fanjob{id: input.id, value: input.value * float32(2)}
	}
}

func initFans() {
	inCh := make(chan fanjob)
	outCh := make(chan fanjob)
	done := make(chan bool)
	workerMap := make(map[int]fanworker) // ignore for now; kept to keep code easy to scale with num_workers

	var wg sync.WaitGroup

	// spawn worker goroutines
	for i := 1; i <= 4; i++ {
		workerInstance := fanworker{id: i, inCh: inCh, outCh: make(chan fanjob)}
		workerMap[i] = workerInstance
		go worker(&workerInstance)
	}

	// spawn collector
	go collector(done, outCh, workerMap[1].outCh, workerMap[2].outCh, workerMap[3].outCh, workerMap[4].outCh)
	go printer(outCh, &wg) // spawn printer

	//send work to shared channel
	for i := 1; i <= 20; i++ {
		job := fanjob{id: i, value: float32(i)}
		wg.Add(1)
		inCh <- job
	}

	wg.Wait()
	close(inCh) // this triggers closure of all worker output channels too, since that's part of each worker's defer
	done <- true

}
