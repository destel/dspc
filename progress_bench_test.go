package dspc

import (
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"testing"
)

func BenchmarkSingleThreaded(b *testing.B) {
	keys := []string{
		"key-1", "key-2", "key-3", "key-4", "key-5", "key-6",
		"key-7", "key-8", "key-9", "key-10", "key-11", "key-12",
	}

	b.Run("Map", func(b *testing.B) {
		m := make(map[string]int64)

		b.ReportAllocs()

		for i := range b.N {
			key := keys[i%len(keys)]
			m[key]++
		}
	})

	b.Run("Progress", func(b *testing.B) {
		var progress Progress

		b.ReportAllocs()

		for i := range b.N {
			key := keys[i%len(keys)]
			progress.Inc(key, 1)
		}
	})
}

func BenchmarkMultiThreaded(b *testing.B) {
	keys := []string{
		"key-1", "key-2", "key-3", "key-4", "key-5", "key-6",
		"key-7", "key-8", "key-9", "key-10", "key-11", "key-12",
	}

	b.Run("Map", func(b *testing.B) {
		m := make(map[string]int64)
		var mu sync.Mutex

		b.ReportAllocs()

		b.RunParallel(func(pb *testing.PB) {
			for i := 0; pb.Next(); i++ {
				key := keys[i%len(keys)]
				mu.Lock()
				m[key]++
				mu.Unlock()
			}
		})
	})

	b.Run("Progress", func(b *testing.B) {
		var progress Progress

		b.ReportAllocs()

		b.RunParallel(func(pb *testing.PB) {
			for i := 0; pb.Next(); i++ {
				key := keys[i%len(keys)]
				progress.Inc(key, 1)
			}
		})
	})

	// each goroutine uses its own key
	b.Run("Progress w disjoint keys", func(b *testing.B) {
		var progress Progress

		b.ReportAllocs()

		i := int64(-1)
		b.RunParallel(func(pb *testing.PB) {
			key := keys[int(atomic.AddInt64(&i, 1))%len(keys)]

			for pb.Next() {
				progress.Inc(key, 1)
			}
		})
	})
}

func BenchmarkPrinting(b *testing.B) {
	var progress Progress

	for i := range 10 {
		key := fmt.Sprintf("key-%d", i)
		progress.Inc(key, 1)
	}

	b.ReportAllocs()

	for range b.N {
		progress.prettyPrint(io.Discard, "Test progress:", false)
	}

}
