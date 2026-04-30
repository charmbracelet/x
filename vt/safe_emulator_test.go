package vt

import (
	"io"
	"sync"
	"testing"
)

func TestSafeEmulator_ReadAfterClose(t *testing.T) {
	t.Parallel()

	se := NewSafeEmulator(80, 24)
	if err := se.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
	n, err := se.Read(make([]byte, 1))
	if n != 0 || err != io.EOF {
		t.Fatalf("Read after Close: got (%d, %v), want (0, EOF)", n, err)
	}
}

func TestSafeEmulator_CloseTwiceThenReadEOF(t *testing.T) {
	t.Parallel()

	se := NewSafeEmulator(10, 10)
	if err := se.Close(); err != nil {
		t.Fatalf("first close: %v", err)
	}
	if err := se.Close(); err != nil {
		t.Fatalf("second close: %v", err)
	}
	n, err := se.Read(make([]byte, 8))
	if n != 0 || err != io.EOF {
		t.Fatalf("Read: got (%d, %v), want (0, EOF)", n, err)
	}
}

func TestSafeEmulator_ConcurrentReadWriteClose(t *testing.T) {
	t.Parallel()

	const rounds = 50
	const parallelism = 16

	for range rounds {
		se := NewSafeEmulator(80, 24)
		var wg sync.WaitGroup
		wg.Add(parallelism * 3)
		for range parallelism {
			go func() {
				defer wg.Done()
				var buf [64]byte
				_, _ = se.Read(buf[:])
			}()
			go func() {
				defer wg.Done()
				_, _ = se.Write([]byte{'x'})
			}()
			go func() {
				defer wg.Done()
				_ = se.Close()
			}()
		}
		wg.Wait()
	}
}
