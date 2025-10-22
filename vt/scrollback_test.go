package vt

import (
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
)

func TestScrollback_Basic(t *testing.T) {
	sb := NewScrollback(100)

	if sb.Len() != 0 {
		t.Errorf("new scrollback should be empty, got %d lines", sb.Len())
	}

	if sb.MaxLines() != 100 {
		t.Errorf("expected max lines 100, got %d", sb.MaxLines())
	}

	// Add a line
	line := []uv.Cell{
		{Content: "H", Width: 1},
		{Content: "i", Width: 1},
	}
	sb.PushLine(line)

	if sb.Len() != 1 {
		t.Errorf("expected 1 line after push, got %d", sb.Len())
	}

	// Retrieve the line
	retrieved := sb.Line(0)
	if len(retrieved) != 2 {
		t.Errorf("expected line length 2, got %d", len(retrieved))
	}
	if retrieved[0].Content != "H" || retrieved[1].Content != "i" {
		t.Errorf("line content mismatch")
	}
}

func TestScrollback_Overflow(t *testing.T) {
	sb := NewScrollback(3) // Small buffer

	// Add 5 lines
	for i := 0; i < 5; i++ {
		line := []uv.Cell{{Content: string(rune('A' + i)), Width: 1}}
		sb.PushLine(line)
	}

	// Should only have 3 lines (newest)
	if sb.Len() != 3 {
		t.Errorf("expected 3 lines after overflow, got %d", sb.Len())
	}

	// Oldest should be 'C' (lines A and B were dropped)
	if sb.Line(0)[0].Content != "C" {
		t.Errorf("expected oldest line 'C', got %s", sb.Line(0)[0].Content)
	}

	// Newest should be 'E'
	if sb.Line(2)[0].Content != "E" {
		t.Errorf("expected newest line 'E', got %s", sb.Line(2)[0].Content)
	}
}

func TestScrollback_Clear(t *testing.T) {
	sb := NewScrollback(100)

	for i := 0; i < 10; i++ {
		line := []uv.Cell{{Content: string(rune('A' + i)), Width: 1}}
		sb.PushLine(line)
	}

	if sb.Len() != 10 {
		t.Errorf("expected 10 lines, got %d", sb.Len())
	}

	sb.Clear()

	if sb.Len() != 0 {
		t.Errorf("expected 0 lines after clear, got %d", sb.Len())
	}
}

func TestScrollback_SetMaxLines(t *testing.T) {
	sb := NewScrollback(100)

	// Add 10 lines
	for i := 0; i < 10; i++ {
		line := []uv.Cell{{Content: string(rune('A' + i)), Width: 1}}
		sb.PushLine(line)
	}

	// Reduce max to 5
	sb.SetMaxLines(5)

	if sb.Len() != 5 {
		t.Errorf("expected 5 lines after reducing max, got %d", sb.Len())
	}

	// Should keep newest 5 lines (F-J)
	if sb.Line(0)[0].Content != "F" {
		t.Errorf("expected oldest remaining line 'F', got %s", sb.Line(0)[0].Content)
	}
}

func TestScreen_ScrollUpWithScrollback(t *testing.T) {
	term := newTestTerminal(t, 5, 3)

	// Fill screen with content
	term.Write([]byte("Line1\r\n"))
	term.Write([]byte("Line2\r\n"))
	term.Write([]byte("Line3"))

	// Check initial scrollback is empty
	if term.ScrollbackLen() != 0 {
		t.Errorf("expected empty scrollback initially, got %d lines", term.ScrollbackLen())
	}

	// Write more lines to trigger scrolling
	term.Write([]byte("\r\nLine4"))

	// One line should have scrolled into scrollback
	if term.ScrollbackLen() != 1 {
		t.Errorf("expected 1 line in scrollback, got %d", term.ScrollbackLen())
	}

	// Check the scrollback contains "Line1"
	line := term.ScrollbackLine(0)
	if line == nil {
		t.Fatal("expected scrollback line, got nil")
	}

	content := ""
	for _, cell := range line {
		if cell.Content != "" {
			content += string(cell.Content)
		}
	}
	if content != "Line1" {
		t.Errorf("expected scrollback to contain 'Line1', got %q", content)
	}
}

func TestScreen_ClearScrollback(t *testing.T) {
	term := newTestTerminal(t, 5, 2)

	// Fill and scroll
	for i := 0; i < 5; i++ {
		term.Write([]byte("Line\r\n"))
	}

	if term.ScrollbackLen() == 0 {
		t.Error("expected scrollback to have lines")
	}

	// Clear scrollback
	term.ClearScrollback()

	if term.ScrollbackLen() != 0 {
		t.Errorf("expected empty scrollback after clear, got %d lines", term.ScrollbackLen())
	}
}

func TestScreen_EraseDisplayWithScrollback(t *testing.T) {
	term := newTestTerminal(t, 10, 3)

	// Fill and scroll
	for i := 0; i < 10; i++ {
		term.Write([]byte("TestLine\r\n"))
	}

	if term.ScrollbackLen() == 0 {
		t.Error("expected scrollback to have lines")
	}

	// ED 3 should clear scrollback
	term.Write([]byte("\x1b[3J"))

	if term.ScrollbackLen() != 0 {
		t.Errorf("expected scrollback cleared after ED 3, got %d lines", term.ScrollbackLen())
	}
}

func TestScreen_ScrollRegionNoScrollback(t *testing.T) {
	term := newTestTerminal(t, 10, 5)

	// Set a scroll region that doesn't start at top
	term.Write([]byte("\x1b[2;4r")) // Lines 2-4

	// Move to scroll region and fill
	term.Write([]byte("\x1b[2;1H"))
	for i := 0; i < 5; i++ {
		term.Write([]byte("Line\r\n"))
	}

	// Scrolling within a limited region should NOT add to scrollback
	if term.ScrollbackLen() != 0 {
		t.Errorf("expected no scrollback for limited scroll region, got %d lines", term.ScrollbackLen())
	}
}

func TestScrollback_EmptyLine(t *testing.T) {
	sb := NewScrollback(100)

	// Empty lines should not be added
	sb.PushLine([]uv.Cell{})

	if sb.Len() != 0 {
		t.Errorf("expected empty line not to be added, got %d lines", sb.Len())
	}
}

func TestScrollback_OutOfBounds(t *testing.T) {
	sb := NewScrollback(10)
	sb.PushLine([]uv.Cell{{Content: "A", Width: 1}})

	// Test negative index
	if line := sb.Line(-1); line != nil {
		t.Error("expected nil for negative index")
	}

	// Test index beyond length
	if line := sb.Line(10); line != nil {
		t.Error("expected nil for out of bounds index")
	}
}

func TestScrollback_DefaultSize(t *testing.T) {
	sb := NewScrollback(0) // Should use default

	if sb.MaxLines() != 10000 {
		t.Errorf("expected default max lines 10000, got %d", sb.MaxLines())
	}

	sb2 := NewScrollback(-5) // Should also use default
	if sb2.MaxLines() != 10000 {
		t.Errorf("expected default max lines 10000 for negative input, got %d", sb2.MaxLines())
	}
}

func TestScreen_AlternateScreenNoScrollback(t *testing.T) {
	term := newTestTerminal(t, 10, 3)

	// Switch to alternate screen
	term.Write([]byte("\x1b[?1049h"))

	// Fill alternate screen
	for i := 0; i < 10; i++ {
		term.Write([]byte("AltLine\r\n"))
	}

	// Alternate screen scrolling should not affect main scrollback
	// (alternate screen has its own scrollback, but we check main)
	mainScrollback := term.scrs[0].ScrollbackLen()
	if mainScrollback != 0 {
		t.Errorf("expected no scrollback in main screen, got %d lines", mainScrollback)
	}

	// Switch back
	term.Write([]byte("\x1b[?1049l"))

	// Main scrollback should still be empty
	if term.ScrollbackLen() != 0 {
		t.Errorf("expected main scrollback to remain empty, got %d lines", term.ScrollbackLen())
	}
}

func TestScrollback_LineCopy(t *testing.T) {
	sb := NewScrollback(10)

	original := []uv.Cell{{Content: "A", Width: 1}}
	sb.PushLine(original)

	// Modify original
	original[0].Content = "B"

	// Retrieved line should still be 'A' (deep copy)
	retrieved := sb.Line(0)
	if retrieved[0].Content != "A" {
		t.Errorf("expected line to be deep copied, got %s instead of 'A'", retrieved[0].Content)
	}
}

func TestExtractLine(t *testing.T) {
	buf := uv.NewBuffer(5, 3)

	// Set some cells
	buf.SetCell(0, 0, &uv.Cell{Content: "H", Width: 1})
	buf.SetCell(1, 0, &uv.Cell{Content: "i", Width: 1})

	line := extractLine(buf, 0, 5)

	if len(line) != 5 {
		t.Errorf("expected line length 5, got %d", len(line))
	}

	if line[0].Content != "H" || line[1].Content != "i" {
		t.Errorf("extracted line content mismatch")
	}

	// Remaining cells should be blank (width 1, but rune may be space)
	if line[2].Width != 1 {
		t.Errorf("expected blank cell width 1 at index 2, got %d", line[2].Width)
	}
}
