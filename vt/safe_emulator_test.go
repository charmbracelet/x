package vt

import (
	"testing"
	"time"

	uv "github.com/charmbracelet/ultraviolet"
)

// TestSafeEmulatorPendingInputDoesNotBlockOutput verifies that input waiting
// for the program to read it doesn't stop the program's output from being
// processed, which would deadlock a program that only reads its input after
// writing output.
func TestSafeEmulatorPendingInputDoesNotBlockOutput(t *testing.T) {
	se := NewSafeEmulator(10, 2)
	defer se.Close() //nolint:errcheck

	// Nothing reads the input pipe, so these block until Close.
	go se.SendKey(uv.KeyPressEvent{Code: 'a', Text: "a"})
	go se.SendText("b")
	go se.Paste("c")
	go se.SendMouse(uv.MouseClickEvent{Button: MouseLeft})
	time.Sleep(50 * time.Millisecond)

	done := make(chan struct{})
	go func() {
		_, _ = se.Write([]byte("hi"))
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("output blocked by pending input")
	}
	if got := se.CellAt(0, 0).Content; got != "h" {
		t.Fatalf("expected %q, got %q", "h", got)
	}
}

func TestSafeEmulatorCellAtReturnsCopy(t *testing.T) {
	se := NewSafeEmulator(10, 2)
	defer se.Close() //nolint:errcheck
	_, _ = se.Write([]byte("x"))

	c := se.CellAt(0, 0)
	c.Content = "y"
	if got := se.CellAt(0, 0).Content; got != "x" {
		t.Fatalf("modifying the returned cell changed the screen to %q", got)
	}
	if se.CellAt(100, 100) != nil || se.ScrollbackCellAt(0, 0) != nil {
		t.Fatal("expected nil for out of bounds cells")
	}
}

func TestSafeEmulatorCloseWhileReading(t *testing.T) {
	se := NewSafeEmulator(10, 2)
	done := make(chan struct{})
	go func() {
		buf := make([]byte, 8)
		for {
			if _, err := se.Read(buf); err != nil {
				close(done)
				return
			}
		}
	}()
	time.Sleep(10 * time.Millisecond)
	if err := se.Close(); err != nil {
		t.Fatal(err)
	}
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("read did not return after close")
	}
}
