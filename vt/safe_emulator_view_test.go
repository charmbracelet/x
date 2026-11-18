package vt

import (
	"sync"
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
)

// TestSafeEmulatorViewUpdate verifies View/Update give locked access
// to the underlying emulator and exclude concurrent writers.
func TestSafeEmulatorViewUpdate(t *testing.T) {
	se := NewSafeEmulator(80, 24)
	se.Update(func(e *Emulator) {
		c := &uv.Cell{Content: "x", Width: 1}
		e.SetCell(0, 0, c)
	})
	var got string
	se.View(func(e *Emulator) {
		if c := e.CellAt(0, 0); c != nil {
			got = c.Content
		}
	})
	if got != "x" {
		t.Fatalf("View read %q, want x", got)
	}
}

// TestSafeEmulatorViewConcurrent runs grid-walking Views against
// concurrent Writes under -race: batched reads must be as safe as
// per-cell CellAt.
func TestSafeEmulatorViewConcurrent(t *testing.T) {
	se := NewSafeEmulator(80, 24)
	var wg sync.WaitGroup
	stop := make(chan struct{})
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-stop:
				return
			default:
			}
			_, _ = se.Write([]byte("hello world\r\n"))
		}
	}()
	for range 100 {
		se.View(func(e *Emulator) {
			for y := 0; y < e.Height(); y++ {
				for x := 0; x < e.Width(); x++ {
					_ = e.CellAt(x, y)
				}
			}
		})
	}
	close(stop)
	wg.Wait()
}

// BenchmarkCellAtGridWalk is the per-cell locking cost View exists
// to avoid: an 80x24 grid read through CellAt.
func BenchmarkCellAtGridWalk(b *testing.B) {
	se := NewSafeEmulator(80, 24)
	for b.Loop() {
		for y := 0; y < 24; y++ {
			for x := 0; x < 80; x++ {
				_ = se.CellAt(x, y)
			}
		}
	}
}

// BenchmarkViewGridWalk is the same read batched under one lock.
func BenchmarkViewGridWalk(b *testing.B) {
	se := NewSafeEmulator(80, 24)
	for b.Loop() {
		se.View(func(e *Emulator) {
			for y := 0; y < 24; y++ {
				for x := 0; x < 80; x++ {
					_ = e.CellAt(x, y)
				}
			}
		})
	}
}
