package dspc

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"
)

func TestBasic(t *testing.T) {
	ctx := context.Background()

	var wg sync.WaitGroup
	var progress Progress
	progress.PrettyPrintEvery(ctx, os.Stdout, 100*time.Millisecond, "Progress:")
	defer progress.PrettyPrint(os.Stdout, "Final progress:")

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			progress.Inc("foo", 1)
			time.Sleep(10 * time.Millisecond)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			progress.Inc("bar", 1)
			time.Sleep(10 * time.Millisecond)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			progress.Inc("baz", 1)
			time.Sleep(10 * time.Millisecond)
		}
	}()

	wg.Wait()
}
