package vt

import (
	"slices"

	uv "github.com/charmbracelet/ultraviolet"
)

// DefaultScrollbackSize is the default maximum number of lines in the scrollback buffer.
const DefaultScrollbackSize = 10000

// Scrollback represents a scrollback buffer that stores lines scrolled off the screen.
//
// It is a ring buffer: once full, each new line replaces the oldest one in
// place, so adding a line costs the same however many lines are kept.
type Scrollback struct {
	// lines holds the lines, oldest at head once the buffer is full. Until
	// then head is 0 and new lines are appended.
	lines    []uv.Line
	head     int
	maxLines int
}

// NewScrollback creates a new scrollback buffer with the given maximum number of lines.
func NewScrollback(maxLines int) *Scrollback {
	if maxLines <= 0 {
		maxLines = DefaultScrollbackSize
	}
	return &Scrollback{
		lines:    make([]uv.Line, 0, min(maxLines, 1000)), // Pre-allocate reasonable capacity
		maxLines: maxLines,
	}
}

// Push adds a line to the scrollback buffer.
// If the buffer is full, the oldest line is removed.
//
// The line is copied, so the caller may reuse it.
func (s *Scrollback) Push(line uv.Line) {
	s.pushRow(line, len(line))
}

// pushRow is Push for a line that is known to be blank from column used on,
// which saves looking at the blank cells to trim them.
func (s *Scrollback) pushRow(line uv.Line, used int) {
	if s == nil || s.maxLines <= 0 {
		return
	}

	// Trim trailing empty cells. This helps with wrapping and window
	// resizing.
	n := min(used, len(line))
	for n > 0 && isBlank(&line[n-1]) {
		n--
	}

	if len(s.lines) < s.maxLines {
		s.lines = append(s.lines, slices.Clone(line[:n]))
		return
	}

	// Replace the oldest line, reusing its memory when it is large enough.
	old := s.lines[s.head]
	if cap(old) >= n {
		old = old[:n]
		copy(old, line[:n])
		// Drop what the rest of the old line referenced.
		clear(old[n:cap(old)])
	} else {
		old = slices.Clone(line[:n])
	}
	s.lines[s.head] = old
	s.head++
	if s.head == len(s.lines) {
		s.head = 0
	}
}

// isBlank reports whether c is a zero cell or equal to [uv.EmptyCell]. It is
// the same as checking c.IsZero() || c.Equal(&uv.EmptyCell), without
// comparing colors.
func isBlank(c *uv.Cell) bool {
	switch c.Width {
	case 1:
		if len(c.Content) != 1 || c.Content[0] != ' ' {
			return false
		}
	case 0:
		if len(c.Content) != 0 {
			return false
		}
	default:
		return false
	}
	return c.Style.Fg == nil && c.Style.Bg == nil && c.Style.UnderlineColor == nil &&
		c.Style.Attrs == 0 && c.Style.Underline == uv.UnderlineNone &&
		len(c.Link.URL) == 0 && len(c.Link.Params) == 0
}

// PushN adds n lines from the buffer starting at line y to the scrollback.
func (s *Scrollback) PushN(buf *uv.RenderBuffer, y, n int) {
	if s == nil || buf == nil || n <= 0 {
		return
	}

	for i := range min(n, buf.Height()-y) {
		if line := buf.Line(y + i); line != nil {
			s.Push(line)
		}
	}
}

// Len returns the number of lines in the scrollback buffer.
func (s *Scrollback) Len() int {
	if s == nil {
		return 0
	}
	return len(s.lines)
}

// MaxLines returns the maximum number of lines the scrollback buffer can hold.
func (s *Scrollback) MaxLines() int {
	if s == nil {
		return 0
	}
	return s.maxLines
}

// SetMaxLines sets the maximum number of lines in the scrollback buffer.
// If the current number of lines exceeds the new maximum, oldest lines are removed.
func (s *Scrollback) SetMaxLines(maxLines int) {
	if s == nil || maxLines <= 0 {
		return
	}

	lines := s.Lines()
	if len(lines) > maxLines {
		// Remove oldest lines
		lines = lines[len(lines)-maxLines:]
	}
	// Copy the lines in order into a new buffer, which also lets go of the
	// lines that were removed.
	s.lines = append(make([]uv.Line, 0, max(len(lines), min(maxLines, 1000))), lines...)
	s.head = 0
	s.maxLines = maxLines
}

// Line returns the line at the given index.
// Index 0 is the oldest line, Len()-1 is the most recent.
// Returns nil if index is out of bounds.
//
// The line belongs to the scrollback buffer and may be overwritten once
// another line is added.
func (s *Scrollback) Line(index int) uv.Line {
	if s == nil || index < 0 || index >= len(s.lines) {
		return nil
	}
	index += s.head
	if index >= len(s.lines) {
		index -= len(s.lines)
	}
	return s.lines[index]
}

// Lines returns all lines in the scrollback buffer.
// Index 0 is the oldest line.
//
// The lines belong to the scrollback buffer and may be overwritten once
// another line is added.
func (s *Scrollback) Lines() []uv.Line {
	if s == nil {
		return nil
	}
	if s.head == 0 {
		return s.lines
	}
	lines := make([]uv.Line, 0, len(s.lines))
	lines = append(lines, s.lines[s.head:]...)
	return append(lines, s.lines[:s.head]...)
}

// Clear removes all lines from the scrollback buffer.
func (s *Scrollback) Clear() {
	if s == nil {
		return
	}
	clear(s.lines)
	s.lines = s.lines[:0]
	s.head = 0
}

// CellAt returns the cell at the given position in the scrollback buffer.
// x is the column, y is the line index (0 = oldest).
// Returns nil if position is out of bounds.
func (s *Scrollback) CellAt(x, y int) *uv.Cell {
	line := s.Line(y)
	if line == nil || x < 0 || x >= len(line) {
		return nil
	}
	return &line[x]
}
