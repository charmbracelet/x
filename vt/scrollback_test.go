package vt

import (
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
)

func TestScrollback(t *testing.T) {
	t.Run("basic push and len", func(t *testing.T) {
		sb := NewScrollback(100)
		if sb.Len() != 0 {
			t.Errorf("expected len 0, got %d", sb.Len())
		}
		if sb.MaxLines() != 100 {
			t.Errorf("expected max 100, got %d", sb.MaxLines())
		}
	})

	t.Run("scrollback in emulator", func(t *testing.T) {
		// Create a small terminal
		e := NewEmulator(10, 5)

		// Fill the screen with numbered lines and force scrolling
		for i := 0; i < 10; i++ {
			e.WriteString("\r\n") // Scroll up
		}

		// Check scrollback has captured some lines
		sbLen := e.ScrollbackLen()
		t.Logf("scrollback length after 10 newlines: %d", sbLen)

		if sbLen == 0 {
			t.Error("expected scrollback to have captured lines, got 0")
		}
	})

	t.Run("scrollback with content", func(t *testing.T) {
		e := NewEmulator(20, 5)

		// Write content that will scroll
		for i := 0; i < 10; i++ {
			e.WriteString("line\r\n")
		}

		// Verify scrollback captured the scrolled content
		sb := e.Scrollback()
		if sb == nil {
			t.Fatal("scrollback is nil")
		}

		// Should have captured lines (at least 5, since screen is 5 tall and we wrote 10 lines)
		if sb.Len() < 5 {
			t.Errorf("expected at least 5 lines in scrollback, got %d", sb.Len())
		}
	})

	t.Run("scrollback max lines", func(t *testing.T) {
		sb := NewScrollback(5)

		// Push more lines than max
		for i := 0; i < 10; i++ {
			sb.Push(nil, false)
		}

		if sb.Len() != 5 {
			t.Errorf("expected len 5 after overflow, got %d", sb.Len())
		}
	})

	t.Run("clear scrollback", func(t *testing.T) {
		e := NewEmulator(20, 5)

		// Write content that will scroll
		for i := 0; i < 10; i++ {
			e.WriteString("line\r\n")
		}

		// Verify we have scrollback
		if e.ScrollbackLen() == 0 {
			t.Error("expected scrollback before clear")
		}

		// Clear it
		e.ClearScrollback()

		if e.ScrollbackLen() != 0 {
			t.Errorf("expected empty scrollback after clear, got %d", e.ScrollbackLen())
		}
	})

	t.Run("alt screen does not have scrollback", func(t *testing.T) {
		e := NewEmulator(20, 5)

		// Write some content to main screen
		for i := 0; i < 10; i++ {
			e.WriteString("line\r\n")
		}

		mainScrollbackLen := e.ScrollbackLen()
		if mainScrollbackLen == 0 {
			t.Error("expected scrollback on main screen")
		}

		// Enter alt screen
		e.WriteString("\x1b[?1049h") // DECSET alt screen

		// Scrollback should still be from main screen
		if e.ScrollbackLen() != mainScrollbackLen {
			t.Errorf("expected scrollback len %d in alt screen, got %d",
				mainScrollbackLen, e.ScrollbackLen())
		}

		// Write to alt screen - should not affect main scrollback
		for i := 0; i < 10; i++ {
			e.WriteString("alt\r\n")
		}

		// Main screen scrollback should be unchanged
		if e.ScrollbackLen() != mainScrollbackLen {
			t.Errorf("expected scrollback len %d after alt screen writes, got %d",
				mainScrollbackLen, e.ScrollbackLen())
		}
	})

	t.Run("ED 2 saves to scrollback", func(t *testing.T) {
		e := NewEmulator(20, 5)

		// Write some content (not enough to scroll)
		e.WriteString("line 1\r\n")
		e.WriteString("line 2\r\n")
		e.WriteString("line 3\r\n")

		// Should have no scrollback yet (didn't scroll)
		initialLen := e.ScrollbackLen()

		// Clear screen with ED 2 (ESC[2J)
		e.WriteString("\x1b[2J")

		// Should have saved lines to scrollback
		newLen := e.ScrollbackLen()
		if newLen <= initialLen {
			t.Errorf("expected scrollback to grow after ED 2, was %d now %d", initialLen, newLen)
		}
		t.Logf("scrollback after ED 2: %d lines", newLen)
	})

	t.Run("ED 3 clears scrollback", func(t *testing.T) {
		e := NewEmulator(20, 5)

		// Write content that will scroll
		for i := 0; i < 10; i++ {
			e.WriteString("line\r\n")
		}

		// Verify we have scrollback
		if e.ScrollbackLen() == 0 {
			t.Error("expected scrollback before ED 3")
		}

		// ED 3 (ESC[3J) should clear scrollback
		e.WriteString("\x1b[3J")

		if e.ScrollbackLen() != 0 {
			t.Errorf("expected empty scrollback after ED 3, got %d", e.ScrollbackLen())
		}
	})
}

func TestReflow(t *testing.T) {
	t.Run("basic reflow shrink", func(t *testing.T) {
		// Create terminal with 20 cols
		e := NewEmulator(20, 5)

		// Write a line that will wrap when shrunk to 10 cols
		e.WriteString("12345678901234567890")

		// Resize to 10 cols - should reflow
		e.Resize(10, 5)

		// First row should have "1234567890"
		var line0 string
		for x := 0; x < 10; x++ {
			cell := e.CellAt(x, 0)
			if cell != nil && cell.Content != "" && cell.Content != " " {
				line0 += cell.Content
			}
		}

		// Second row should have "1234567890"
		var line1 string
		for x := 0; x < 10; x++ {
			cell := e.CellAt(x, 1)
			if cell != nil && cell.Content != "" && cell.Content != " " {
				line1 += cell.Content
			}
		}

		if line0 != "1234567890" {
			t.Errorf("line 0: expected '1234567890', got %q", line0)
		}
		if line1 != "1234567890" {
			t.Errorf("line 1: expected '1234567890', got %q", line1)
		}
	})

	t.Run("reflow grow unwraps", func(t *testing.T) {
		// Create terminal with 10 cols
		e := NewEmulator(10, 5)

		// Write text that wraps
		e.WriteString("1234567890ABCDEFGHIJ")

		// Resize to 20 cols - should unwrap
		e.Resize(20, 5)

		// First row should now have all 20 chars
		var line0 string
		for x := 0; x < 20; x++ {
			cell := e.CellAt(x, 0)
			if cell != nil && cell.Content != "" && cell.Content != " " {
				line0 += cell.Content
			}
		}

		if line0 != "1234567890ABCDEFGHIJ" {
			t.Errorf("expected '1234567890ABCDEFGHIJ', got %q", line0)
		}
	})

	t.Run("reflow cursor tracking", func(t *testing.T) {
		// Create terminal with 20 cols
		e := NewEmulator(20, 5)

		// Write some text and move cursor
		e.WriteString("Hello, World!")
		// Cursor should be at position 13 (after the !)

		pos := e.CursorPosition()
		if pos.X != 13 || pos.Y != 0 {
			t.Errorf("initial cursor: expected (13,0), got (%d,%d)", pos.X, pos.Y)
		}

		// Resize - cursor should stay at same logical position
		e.Resize(10, 5)

		pos = e.CursorPosition()
		// "Hello, Wor" on line 0 (10 chars)
		// "ld!" on line 1 (3 chars)
		// Cursor was at offset 13, which is now (3, 1)
		if pos.X != 3 || pos.Y != 1 {
			t.Errorf("after resize: expected (3,1), got (%d,%d)", pos.X, pos.Y)
		}
	})

	t.Run("scrollback reflow", func(t *testing.T) {
		// Create a screen with scrollback to test reflow
		// Start with width 10 (same as line1 length) so we can reflow to 20
		scr := NewScreen(10, 10)
		sb := NewScrollback(100)
		scr.SetScrollback(sb)

		// Push lines that represent soft-wrapped content
		line1 := make([]byte, 10)
		for i := range line1 {
			line1[i] = byte('A' + i)
		}
		// Convert to uv.Line manually
		uvLine1 := make([]uv.Cell, 10)
		for i, b := range line1 {
			uvLine1[i] = uv.Cell{Content: string(b), Width: 1}
		}
		sb.Push(uvLine1, true) // soft-wrapped

		line2 := make([]byte, 5)
		for i := range line2 {
			line2[i] = byte('K' + i)
		}
		uvLine2 := make([]uv.Cell, 5)
		for i, b := range line2 {
			uvLine2[i] = uv.Cell{Content: string(b), Width: 1}
		}
		sb.Push(uvLine2, false) // not soft-wrapped

		// Should have 2 lines
		if sb.Len() != 2 {
			t.Errorf("expected 2 lines, got %d", sb.Len())
		}

		// Reflow via screen resize (width change triggers reflow)
		scr.Reflow(20, 10, 0, 0)

		if sb.Len() != 1 {
			t.Errorf("after reflow to 20: expected 1 line, got %d", sb.Len())
		}

		// Check content
		resultLine := sb.Line(0)
		if len(resultLine) != 15 {
			t.Errorf("expected 15 cells, got %d", len(resultLine))
		}
	})
}
