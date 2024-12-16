package dspc

import "testing"

func expectValue[T comparable](t *testing.T, actual T, expected T, msgAndArgs ...any) {
	t.Helper()

	if actual != expected {
		t.Fatalf("expected %v, got %v", expected, actual)
	}
}

func expectSlice[T comparable](t *testing.T, actual []T, expected []T) {
	t.Helper()
	if len(expected) != len(actual) {
		t.Fatalf("expected %v, got %v", expected, actual)
		return
	}

	for i := range expected {
		if expected[i] != actual[i] {
			t.Errorf("expected %v, got %v", expected, actual)
			return
		}
	}
}

type customWriter struct {
	WriteFunc func(p []byte) (n int, err error)
}

func (w *customWriter) Write(p []byte) (n int, err error) {
	return w.WriteFunc(p)
}
