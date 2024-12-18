package dspc

import (
	"bytes"
	"cmp"
	"slices"
	"strconv"
)

// buffer extension with some zero allocation helpers
type betterBuffer struct {
	bytes.Buffer
}

func (b *betterBuffer) WriteInt64(n int64) {
	var data [20]byte
	res := strconv.AppendInt(data[:0], n, 10)
	b.Write(res)
}

func (b *betterBuffer) WriteInt(n int) {
	b.WriteInt64(int64(n))
}

func (b *betterBuffer) WriteByteRepeated(c byte, n int) {
	const bufSize = 32
	var data [bufSize]byte
	for i := range min(bufSize, n) {
		data[i] = c
	}

	for {
		if n < bufSize {
			b.Write(data[:n])
			return
		}

		b.Write(data[:])
		n -= bufSize
	}
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

// takes sorted slice and returns its copy with the new value inserted at the correct position
func cloneSortedSliceAndInsert[T cmp.Ordered](s []T, v T) []T {
	res := make([]T, len(s)+1)
	pos, _ := slices.BinarySearch(s, v)

	copy(res, s[:pos])
	res[pos] = v
	copy(res[pos+1:], s[pos:])

	return res
}
