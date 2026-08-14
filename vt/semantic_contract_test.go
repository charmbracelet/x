package vt

import (
	"errors"
	"strings"
	"sync"
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
)

func TestEraseLineBreaksStaleSoftBoundary(t *testing.T) {
	e := NewEmulator(5, 3)
	e.WriteString("abcdef")
	e.WriteString("\x1b[2KZ")

	got := snapshotString(t, e)
	if !strings.Contains(got, "abcde\r\n Z") {
		t.Fatalf("snapshot = %q, want hard boundary before rewritten row", got)
	}
}

func TestScrollbackReadsAreDefensive(t *testing.T) {
	sb := NewScrollback(2)
	line := uv.NewLine(2)
	line.Set(0, &uv.Cell{Content: "a", Width: 1})
	sb.Push(line)

	fromLine := sb.Line(0)
	fromLine[0].Content = "x"
	fromLines := sb.Lines()
	fromLines[0][0].Content = "y"
	fromCell := sb.CellAt(0, 0)
	fromCell.Content = "z"

	if got := sb.Line(0)[0].Content; got != "a" {
		t.Fatalf("stored cell = %q, want defensive read preserving a", got)
	}
}

func TestScreenCellReadIsDefensive(t *testing.T) {
	s := NewScreen(2, 1)
	s.SetCell(0, 0, &uv.Cell{Content: "a", Width: 1})
	cell := s.CellAt(0, 0)
	cell.Content = "x"
	if got := s.CellAt(0, 0).Content; got != "a" {
		t.Fatalf("stored screen cell = %q, want defensive read preserving a", got)
	}
}

func TestAlternateResizePreservesPhysicalRowsAndClippedCells(t *testing.T) {
	e := NewEmulator(6, 2)
	e.WriteString("\x1b[?1049habcdef\r\nXYZ")

	e.Resize(3, 2)
	if got := e.Render(); !strings.Contains(got, "abc\nXYZ") {
		t.Fatalf("alternate render after shrink = %q, want clipped physical rows", got)
	}
	e.Resize(6, 2)

	got := e.Render()
	if !strings.Contains(got, "abcdef\nXYZ") {
		t.Fatalf("alternate render after shrink-grow = %q, want physical rows and clipped cells preserved", got)
	}
	if e.scrs[0].scrollback.Len() != 0 {
		t.Fatalf("primary scrollback changed during alternate resize: %d", e.scrs[0].scrollback.Len())
	}
}

func TestAlternateHeightResizeIsCursorAnchored(t *testing.T) {
	e := NewEmulator(4, 4)
	e.WriteString("\x1b[?1049hA\r\nB\r\nC\r\nD\x1b[3;1H")

	e.Resize(4, 2)

	if got := e.Render(); !strings.Contains(got, "B\nC") {
		t.Fatalf("alternate render after height shrink = %q, want rows around cursor", got)
	}
	if got := e.CursorPosition(); got.X != 0 || got.Y != 1 {
		t.Fatalf("cursor after height shrink = %v, want (0,1)", got)
	}
	e.Resize(4, 4)
	if got := e.Render(); !strings.HasPrefix(got, "B\nC\n\n") {
		t.Fatalf("alternate render after height growth = %q, want retained rows plus blanks", got)
	}
}

func TestSnapshotFailureReturnsNoPartialBytes(t *testing.T) {
	e := NewEmulator(5, 2)
	e.WriteString("abc")
	e.scr.buf.boundaries[0] = rowBoundary(255)

	got, err := e.ReattachSnapshot()
	if got != nil {
		t.Fatalf("failed snapshot bytes = %q, want nil", got)
	}
	var snapshotErr *SnapshotError
	if !errors.As(err, &snapshotErr) {
		t.Fatalf("snapshot error = %v, want *SnapshotError", err)
	}
}

func TestHorizontalEditCancelsPendingWrap(t *testing.T) {
	for _, sequence := range []string{"\x1b[@", "\x1b[P"} {
		t.Run(sequence, func(t *testing.T) {
			e := NewEmulator(5, 2)
			e.WriteString("abcde")
			e.WriteString(sequence)
			e.WriteString("Z")

			if got := e.CursorPosition(); got.Y != 0 {
				t.Fatalf("cursor after edit and write = %v, want current row", got)
			}
		})
	}
}

func TestPrimaryResizePreservesSavedCursorLogicalPosition(t *testing.T) {
	e := NewEmulator(5, 3)
	e.WriteString("abcdefgh")
	e.WriteString("\x1b7")

	e.Resize(10, 3)
	e.WriteString("\x1b8")

	if got := e.CursorPosition(); got.X != 8 || got.Y != 0 {
		t.Fatalf("restored cursor after widening = %v, want logical position (8,0)", got)
	}
}

func TestResizePreservesCustomTabStops(t *testing.T) {
	e := NewEmulator(12, 2)
	e.WriteString("\x1b[3g\x1b[1;4H\x1bH")

	e.Resize(10, 2)

	if !e.tabstops.IsStop(3) {
		t.Fatal("custom tab stop at column 3 was lost during resize")
	}
	if e.tabstops.IsStop(8) {
		t.Fatal("cleared default tab stop at column 8 was recreated during resize")
	}
}

func TestSafeEmulatorSnapshotConcurrentObservation(t *testing.T) {
	e := NewSafeEmulator(10, 3)
	var wg sync.WaitGroup
	for range 20 {
		wg.Add(3)
		go func() {
			defer wg.Done()
			_, _ = e.Write([]byte("abcdefghij"))
		}()
		go func() {
			defer wg.Done()
			e.Resize(5, 3)
			e.Resize(10, 3)
		}()
		go func() {
			defer wg.Done()
			_, _ = e.ReattachSnapshot()
		}()
	}
	wg.Wait()
}
