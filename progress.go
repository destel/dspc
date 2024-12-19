// Package dspc provides a progress counter for tracking and displaying
// real-time progress of concurrent operations in terminal apps.
// It's minimalistic, lock-free, zero-allocation and provides in-place pretty-printing out of the box.
package dspc

import (
	"fmt"
	"io"
	"iter"
	"maps"
	"os"
	"sync/atomic"
	"time"
)

// Progress tracks multiple named counters. It's like a concurrent map[string]int64
// but optimized for progress tracking with small, stable sets of keys (typically
// fitting on a single screen).
//
// All operations are atomic, lock-free and safe for concurrent use.
// After the set of keys becomes stable, Progress does zero allocations for all operations.
//
// The zero Progress is empty and ready for use
type Progress struct {
	state atomic.Pointer[progressState]

	// for printing w/o allocations
	buf     betterBuffer
	entries []entry
}

type progressState struct {
	counters   map[string]*int64
	sortedKeys []string
}

type entry struct {
	key   string
	value int64
}

// Inc atomically adds delta to the counter associated with the given key.
// If the key doesn't exist, it's created with an initial value of 0 before adding delta.
func (p *Progress) Inc(key string, delta int64) {
	c := p.getOrCreateCounter(key)
	atomic.AddInt64(c, delta)
}

// Set atomically sets the counter associated with the given key to value.
// If the key doesn't exist, it's created with the specified value.
func (p *Progress) Set(key string, value int64) {
	c := p.getOrCreateCounter(key)
	atomic.StoreInt64(c, value)
}

// Get returns the current value of the counter associated with key.
// Returns 0 if the key doesn't exist.
func (p *Progress) Get(key string) int64 {
	state := p.state.Load()
	if state == nil {
		return 0
	}

	counter := state.counters[key]
	if counter == nil {
		return 0
	}

	return atomic.LoadInt64(counter)
}

// All returns an iterator over all counters in lexicographical key order.
// The iterator yields (key, value) pairs. The values represent atomic snapshots
// of the counters at the time they are read.
func (p *Progress) All() iter.Seq2[string, int64] {
	return func(yield func(string, int64) bool) {
		state := p.state.Load()
		if state == nil {
			return
		}

		for _, key := range state.sortedKeys {
			if !yield(key, atomic.LoadInt64(state.counters[key])) {
				return
			}
		}
	}
}

func (p *Progress) size() int {
	state := p.state.Load()
	if state == nil {
		return 0
	}

	return len(state.counters)
}

func (p *Progress) getOrCreateCounter(key string) *int64 {
	for {
		state := p.state.Load()

		// happy path: map contains the key
		if state != nil {
			if counter := state.counters[key]; counter != nil {
				return counter
			}
		}

		// Unhappy path: need to clone the state and add new key to it with CAS
		newCounter := new(int64)
		newState := &progressState{}

		if state != nil {
			newState.counters = make(map[string]*int64, len(state.counters)+1)
			maps.Copy(newState.counters, state.counters)
			newState.counters[key] = newCounter

			newState.sortedKeys = cloneSortedSliceAndInsert(state.sortedKeys, key)
		} else {
			newState.counters = map[string]*int64{key: newCounter}
			newState.sortedKeys = []string{key}
		}

		if p.state.CompareAndSwap(state, newState) {
			return newCounter
		}
	}
}

func (p *Progress) prettyPrint(w io.Writer, title string, inPlace bool) error {
	p.buf.Reset()
	p.entries = p.entries[:0]

	maxKeySize := 0
	maxValueSize := 0

	for key, value := range p.All() {
		p.entries = append(p.entries, entry{key, value})

		maxKeySize = max(maxKeySize, len(key))
		maxValueSize = max(maxValueSize, digitCount(value))
	}

	// clear the screen after the cursor
	p.buf.WriteString("\033[J")

	// Start with a blank line
	p.buf.WriteString("\n")

	// Print the title
	p.buf.WriteString(title)
	p.buf.WriteString("\n")

	// Print the progress
	for _, ent := range p.entries {
		p.buf.WriteString("  ")
		p.buf.WriteString(ent.key)
		p.buf.WriteByteRepeated(' ', maxKeySize-len(ent.key))
		p.buf.WriteString("  ")
		p.buf.WriteByteRepeated(' ', maxValueSize-digitCount(ent.value))
		p.buf.WriteInt64(ent.value)
		p.buf.WriteString("\n")
	}

	// End with a blank line
	p.buf.WriteString("\n")

	if inPlace {
		// Move the cursor up to the start of the progress.
		// Works more reliably that doing save/restore of the cursor position.
		p.buf.WriteString("\033[")
		p.buf.WriteInt(len(p.entries) + 3)
		p.buf.WriteString("A")
	}

	// Flush the buffer in a single Write call
	_, err := w.Write(p.buf.Bytes())
	return err
}

// PrettyPrintEvery periodically prints the current state of Progress to w (typically stdout ot stderr).
// It updates the output in-place and won't damage the log output of the application
// (assuming logs are printed line by line).
// PrettyPrintEvery returns the function that stops the printing when called.
//
// Usage:
//
//	stop := progress.PrettyPrintEvery(os.Stdout, time.Second, "Progress:")
//	defer stop()
//
// Or better:
//
//	defer progress.PrettyPrintEvery(os.Stdout, time.Second, "Progress:")()
//
// Example output:
//
//	Progress:
//	  completed  15
//	  failed      3
//	  skipped     7
func (p *Progress) PrettyPrintEvery(w io.Writer, t time.Duration, title string) func() {
	stop := make(chan struct{})
	done := make(chan struct{})

	printError := func(err error) {
		// Should never happen, especially when writing to stdout/stderr
		fmt.Fprintln(os.Stderr, "Error writing progress:", err)
	}

	go func() {
		defer close(done)

		ticker := time.NewTicker(t)
		defer ticker.Stop()

		if err := p.prettyPrint(w, title, true); err != nil {
			printError(err)
			return
		}

		for {
			select {
			case <-ticker.C:
				if err := p.prettyPrint(w, title, true); err != nil {
					printError(err)
					return
				}
			case <-stop:
				// w/o ansi
				if err := p.prettyPrint(w, title, false); err != nil {
					printError(err)
				}
				return
			}
		}

	}()

	stopPrinting := func() {
		close(stop)
		<-done
	}

	return stopPrinting
}
