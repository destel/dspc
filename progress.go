package dspc

import (
	"bytes"
	"context"
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

func (p *Progress) prettyPrintToBuffer(buf *bytes.Buffer, title string) {
	maxKeySize := 0
	maxValueSize := 0

	for key, value := range p.All() {
		maxKeySize = max(maxKeySize, len(key))
		maxValueSize = max(maxValueSize, digitCount(value))
	}

	// Start with a blank line
	buf.WriteString("\033[K\n")

	// Print the title
	buf.WriteString("\033[K")
	buf.WriteString(title)
	buf.WriteString("\n")

	// Print the progress
	for key, value := range p.All() {
		buf.WriteString("\033[K")
		fmt.Fprintf(buf, "  %-*s  %*d", maxKeySize, key, maxValueSize, value)
		buf.WriteString("\n")
	}

	// End with a blank line
	buf.WriteString("\033[K\n")
}

func (p *Progress) PrettyPrint(w io.Writer, title string) error {
	var buf bytes.Buffer
	p.prettyPrintToBuffer(&buf, title)

	// Try to write all at once
	if _, err := w.Write(buf.Bytes()); err != nil {
		return err // this should never happen for stdout or stderr
	}
	return nil
}

func (p *Progress) PrettyPrintEvery(ctx context.Context, w io.Writer, t time.Duration, title string) {
	go func() {
		var buf bytes.Buffer

		for {
			buf.Reset()
			buf.WriteString("\033[s") // save cursor position
			p.prettyPrintToBuffer(&buf, title)
			buf.WriteString("\033[u") // restore cursor position

			// Try to write all at once
			if _, err := w.Write(buf.Bytes()); err != nil {
				// this should never happen for stdout or stderr
				fmt.Fprintln(os.Stderr, "Error writing progress:", err)
				return
			}

			if ctx.Err() != nil {
				return
			}
			time.Sleep(t)
		}
	}()
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
