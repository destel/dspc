package main

import (
	"errors"
	"log"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/destel/dspc"
)

var ErrNotFound = errors.New("not found")
var ErrTimeout = errors.New("timeout")
var ErrSkip = errors.New("skip")

func main() {
	var progress dspc.Progress
	var wg sync.WaitGroup

	// Print progress to stdout periodically.
	defer progress.PrettyPrintEvery(os.Stdout, 100*time.Millisecond, "Progress:")()

	// Produce some work.
	workChan := make(chan int)
	go func() {
		defer close(workChan)
		for i := range 1000 {
			workChan <- i
		}
	}()

	// This function spawns a worker.
	// dur argument controls how much time the worker needs to process one item.
	runWorker := func(dur time.Duration) {
		wg.Add(1)

		go func() {
			defer wg.Done()

			log.Printf("Starting %v worker\n", dur)
			defer log.Printf("Stopped %v worker\n", dur)

			for i := range workChan {
				err := doWork(dur, &progress)

				if errors.Is(err, ErrSkip) {
					progress.Inc("skipped", 1)
					continue
				}
				if err != nil {
					progress.Inc("errors", 1)
					progress.Inc("errors["+err.Error()+"]", 1)

					log.Printf("Failed to process item %d: %v", i, err)
					continue
				}

				progress.Inc("done", 1)
			}
		}()
	}

	// Spawn some workers with different processing times
	runWorker(100 * time.Millisecond)
	runWorker(50 * time.Millisecond)
	runWorker(10 * time.Millisecond)

	wg.Wait()
	log.Println("All work is done")
}

func doWork(dur time.Duration, progress *dspc.Progress) error {
	progress.Inc("in_progress", 1)
	defer progress.Inc("in_progress", -1)

	switch rand.Int31n(100) {
	case 0:
		return ErrNotFound // 1% chance
	case 1:
		return ErrTimeout // 1% chance
	case 10, 11, 12, 13:
		return ErrSkip // 4% chance
	default:
		time.Sleep(dur) // do some work
		return nil
	}
}
