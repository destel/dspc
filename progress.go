package dspc

import (
	"bytes"
	"fmt"
	"io"
	"iter"
	"maps"
	"os"
	"slices"
	"sync/atomic"
	"time"
)

type Progress struct {
	state atomic.Pointer[progressState]
}

type progressState struct {
	counters   map[string]*int64
	sortedKeys []string
}

func (p *Progress) Inc(key string, delta int64) {
	c := p.getOrCreateCounter(key)
	atomic.AddInt64(c, delta)
}

func (p *Progress) Set(key string, value int64) {
	c := p.getOrCreateCounter(key)
	atomic.StoreInt64(c, value)
}

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
		} else {
			newState.counters = make(map[string]*int64, 1)
		}
		newState.counters[key] = newCounter
		newState.rebuildSortedKeys()

		if p.state.CompareAndSwap(state, newState) {
			return newCounter
		}
	}
}

func (s *progressState) rebuildSortedKeys() {
	s.sortedKeys = slices.Sorted(maps.Keys(s.counters))
}

type entry struct {
	key   string
	value int64
}

type printingState struct {
	buf     bytes.Buffer
	entries []entry
}

func (p *Progress) prettyPrint(w io.Writer, title string, inPlace bool, state *printingState) error {
	if state == nil {
		var localState printingState
		state = &localState
	} else {
		state.buf.Reset()
		state.entries = state.entries[:0]
	}

	maxKeySize := 0
	maxValueSize := 0

	for key, value := range p.All() {
		state.entries = append(state.entries, entry{key, value})

		maxKeySize = max(maxKeySize, len(key))
		maxValueSize = max(maxValueSize, digitCount(value))
	}

	if inPlace {
		state.buf.WriteString("\033[J") // clear the screen after the cursor todo: always clear?
		//state.buf.WriteString("\033[s") // save cursor position
	}

	// Start with a blank line
	state.buf.WriteString("\n")

	// Print the title
	state.buf.WriteString(title)
	state.buf.WriteString("\n")

	// Print the progress
	for _, entry := range state.entries {
		fmt.Fprintf(&state.buf, "  %-*s  %*d", maxKeySize, entry.key, maxValueSize, entry.value)
		state.buf.WriteString("\n")
	}

	// End with a blank line
	state.buf.WriteString("\n")

	if inPlace {
		//state.buf.WriteString("\033[u") // restore cursor position
		fmt.Fprintf(&state.buf, "\033[%dA", len(state.entries)+3)
	}

	// Flush the buffer in a single Write call
	_, err := w.Write(state.buf.Bytes())
	return err
}

// Usage:
//
//	stop := progress.PrettyPrintEvery(os.Stdout, time.Second, "Progress:")
//	defer stop()
//
// Or better:
//
//	defer progress.PrettyPrintEvery(os.Stdout, time.Second, "Progress:")()
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

		var state printingState

		if err := p.prettyPrint(w, title, true, &state); err != nil {
			printError(err)
			return
		}

		for {
			select {
			case <-ticker.C:
				if err := p.prettyPrint(w, title, true, &state); err != nil {
					printError(err)
					return
				}
			case <-stop:
				// w/o ansi
				if err := p.prettyPrint(w, title, false, &state); err != nil {
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

func digitCount(n int64) int {
	if n == 0 {
		return 1
	}

	count := 0
	if n < 0 {
		count = 1 // for the minus sign
	}

	for n != 0 {
		n /= 10
		count++
	}
	return count
}
