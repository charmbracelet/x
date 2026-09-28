package vt

import (
	"io"
	"sync"
	"testing"
)

func TestEmulatorCloseDataRace(t *testing.T) {
	emu := NewEmulator(80, 24)

	var wg sync.WaitGroup
	wg.Add(2)

	// Goroutine 1: Read (blocks on pipe)
	go func() {
		defer wg.Done()
		buf := make([]byte, 1024)
		n, err := emu.Read(buf)
		if err != io.EOF {
			t.Errorf("expected io.EOF, got n=%d err=%v", n, err)
		}
	}()

	// Goroutine 2: Close (should unblock Read)
	go func() {
		defer wg.Done()
		if err := emu.Close(); err != nil {
			t.Errorf("unexpected close error: %v", err)
		}
	}()

	wg.Wait()
}

func TestSafeEmulatorCloseDataRace(t *testing.T) {
	emu := NewSafeEmulator(80, 24)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		buf := make([]byte, 1024)
		n, err := emu.Read(buf)
		if err != io.EOF {
			t.Errorf("expected io.EOF, got n=%d err=%v", n, err)
		}
	}()

	go func() {
		defer wg.Done()
		if err := emu.Close(); err != nil {
			t.Errorf("unexpected close error: %v", err)
		}
	}()

	wg.Wait()
}

func TestEmulatorCloseIdempotent(t *testing.T) {
	emu := NewEmulator(80, 24)

	if err := emu.Close(); err != nil {
		t.Fatalf("first close: %v", err)
	}

	if err := emu.Close(); err != nil {
		t.Fatalf("second close should be no-op: %v", err)
	}
}

func TestEmulatorWriteAfterClose(t *testing.T) {
	emu := NewEmulator(80, 24)

	if err := emu.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}

	n, err := emu.Write([]byte("hello"))
	if err != io.ErrClosedPipe {
		t.Errorf("expected ErrClosedPipe, got n=%d err=%v", n, err)
	}
}
