package dspc

import (
	"bytes"
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
