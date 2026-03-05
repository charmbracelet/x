package vt

import (
	"testing"
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
			sb.Push(nil)
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
