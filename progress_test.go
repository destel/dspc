package dspc

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestRead(t *testing.T) {
	var progress Progress

	// Read from zero value Progress
	expectValue(t, progress.Get("any"), 0)

	// Returns correct values after Inc
	progress.Inc("foo", 5)
	expectValue(t, progress.Get("foo"), 5)

	// Still returns 0 for non-existent
	expectValue(t, progress.Get("nonexistent"), 0)
}

func TestWrite(t *testing.T) {
	var progress Progress

	progress.Set("foo", 10)
	progress.Inc("foo", 5)
	progress.Inc("foo", -7)
	progress.Inc("bar", 2)

	expectValue(t, progress.Get("foo"), 8)
	expectValue(t, progress.Get("bar"), 2)
}

func TestIteration(t *testing.T) {
	var progress Progress

	// Read from zero value Progress
	for _, _ = range progress.All() {
		t.Fatalf("expected no values")
	}

	// Should iterate in alphabetical order
	progress.Inc("foo4", 1)
	progress.Set("foo2", 2)
	progress.Inc("foo3", -3)
	progress.Inc("foo1", 4)

	type Pair struct {
		Key   string
		Value int64
	}

	var actual []Pair
	for k, v := range progress.All() {
		actual = append(actual, Pair{k, v})
	}

	expectSlice(t, actual, []Pair{{"foo1", 4}, {"foo2", 2}, {"foo3", -3}, {"foo4", 1}})
}

func TestConcurrency(t *testing.T) {
	concurrency := 5
	totalKeys := 1000

	keysStream := make(chan string, concurrency)

	// Start a key producer
	// The idea is to create contention by forcing multiple goroutines to
	// write to the same key at the same time.
	// With the current settings (5 goroutines, 1000 keys), test runs have shown
	// about 3.5k failures of the CAS operation.
	go func() {
		defer close(keysStream)

		for k := range totalKeys {
			key := fmt.Sprintf("key-%d", k)
			for range concurrency {
				keysStream <- key
			}
			// Sleep for a bit to allow goroutines to finish with the current key
			time.Sleep(1 * time.Microsecond)
		}
	}()

	// Start the workers
	var progress Progress

	var wg sync.WaitGroup
	for range concurrency {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for key := range keysStream {
				progress.Inc(key, 1)
			}
		}()
	}
	wg.Wait()

	// Verify each key got incremented exactly 'concurrency' times
	for k := range totalKeys {
		key := fmt.Sprintf("key-%d", k)
		expectValue(t, progress.Get(key), int64(concurrency))
	}
}

func TestPrinting(t *testing.T) {
	var progress Progress
	progress.Inc("foo", -100)
	progress.Inc("bar", 20)
	progress.Set("grault", 0)

	var out customWriter
	outParts := make([]string, 0, 5)
	enough := make(chan struct{})

	stop := progress.PrettyPrintEvery(&out, 100*time.Millisecond, "Test progress:")

	out.WriteFunc = func(p []byte) (n int, err error) {
		outParts = append(outParts, string(p))
		if len(outParts) == 2 {
			enough <- struct{}{} // tell the main goroutine to call stop()
		}
		return len(p), nil
	}

	<-enough
	stop() // this should also print the final state w/o ansi

	const expectedOutput = `
Test progress:
  bar       20
  foo     -100
  grault     0

`
	const ansiOutput = "\033[J" + expectedOutput + "\033[6A"

	expectValue(t, len(outParts), 3)
	expectValue(t, outParts[0], ansiOutput)
	expectValue(t, outParts[1], ansiOutput)
	expectValue(t, outParts[2], expectedOutput)
}
