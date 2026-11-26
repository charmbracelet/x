package vt

import (
	"errors"
	"io"
	"runtime"
	"sync"
	"testing"
	"time"
)

func TestSafeEmulatorClose(t *testing.T) {
	t.Run("ends a Read that waits for input", func(t *testing.T) {
		se := NewSafeEmulator(80, 24)

		read := make(chan error, 1)
		go func() {
			_, err := se.Read(make([]byte, 1))
			read <- err
		}()

		if err := se.Close(); err != nil {
			t.Fatalf("close: %v", err)
		}
		select {
		case err := <-read:
			if !errors.Is(err, io.EOF) {
				t.Fatalf("read: got %v, want EOF", err)
			}
		case <-time.After(5 * time.Second):
			t.Fatal("Read did not return after Close")
		}
	})

	t.Run("ends a send that nothing reads", func(t *testing.T) {
		se := NewSafeEmulator(80, 24)

		sent := make(chan struct{})
		go func() {
			defer close(sent)
			se.SendText("never read")
		}()
		// SendText holds the lock for as long as it waits for a reader.
		for se.mu.TryLock() {
			se.mu.Unlock()
			runtime.Gosched()
		}

		if err := se.Close(); err != nil {
			t.Fatalf("close: %v", err)
		}
		select {
		case <-sent:
		case <-time.After(5 * time.Second):
			t.Fatal("SendText did not return after Close")
		}
	})

	t.Run("is safe to call more than once, and from many goroutines", func(t *testing.T) {
		se := NewSafeEmulator(80, 24)

		var wg sync.WaitGroup
		for range 8 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				if err := se.Close(); err != nil {
					t.Errorf("close: %v", err)
				}
			}()
		}
		wg.Wait()

		if _, err := se.Read(make([]byte, 1)); !errors.Is(err, io.EOF) {
			t.Fatalf("read after close: got %v, want EOF", err)
		}
		if _, err := se.Write([]byte("x")); !errors.Is(err, io.ErrClosedPipe) {
			t.Fatalf("write after close: got %v, want ErrClosedPipe", err)
		}
	})
}
