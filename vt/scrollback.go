package vt

import (
	"slices"

	uv "github.com/charmbracelet/ultraviolet"
)

// DefaultScrollbackSize is the default maximum number of lines in the scrollback buffer.
const DefaultScrollbackSize = 10000

// Scrollback represents a scrollback buffer that stores lines scrolled off the screen.
//
// Storage is a ring once the buffer reaches maxLines: head indexes
// the oldest line and pushes overwrite in place. The previous
// implementation removed the oldest line with slices.Delete — an
// O(maxLines) memmove on EVERY push once full, which profiled as the
// dominant cost of bulk output in a real emulator (every newline at
// the bottom of the screen pushes a line; at the default 10k cap
// that was ~240KB moved per line of output).
type Scrollback struct {
	lines    []uv.Line
	head     int // index of the oldest line; 0 until the ring is full
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
func (s *Scrollback) Push(line uv.Line) {
	if s == nil || s.maxLines <= 0 {
		return
	}

	// Find last non-empty cell to trim trailing empty cells.
	// This helps with wrapping and window resizing.
	lastNonEmpty := -1
	for i := len(line) - 1; i >= 0; i-- {
		c := &line[i]
		// Fast path: a structurally-blank cell (zero value or plain
		// space, no style/link) via single whole-struct compares —
		// one type-generated equality call per cell instead of
		// Cell.Equal's interface-unwrapping color comparisons. This
		// scan runs for every line pushed, which is every line of
		// output once the screen is full.
		if *c == uv.EmptyCell || *c == (uv.Cell{}) {
			continue
		}
		if !c.IsZero() && !c.Equal(&uv.EmptyCell) {
			lastNonEmpty = i
			break
		}
	}

	// Clone the line content up to and including the last non-empty cell
	cloned := slices.Clone(line[:lastNonEmpty+1])

	if len(s.lines) >= s.maxLines {
		// Ring is full: overwrite the oldest line in place.
		s.lines[s.head] = cloned
		s.head = (s.head + 1) % len(s.lines)
		return
	}
	s.lines = append(s.lines, cloned)
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

	s.maxLines = maxLines
	if len(s.lines) > maxLines {
		// Keep the newest maxLines lines, linearized.
		all := s.Lines()
		s.lines = all[len(all)-maxLines:]
		s.head = 0
	}
}

// Line returns the line at the given index.
// Index 0 is the oldest line, Len()-1 is the most recent.
// Returns nil if index is out of bounds.
func (s *Scrollback) Line(index int) uv.Line {
	if s == nil || index < 0 || index >= len(s.lines) {
		return nil
	}
	return s.lines[(s.head+index)%len(s.lines)]
}

// Lines returns all lines in the scrollback buffer.
// Index 0 is the oldest line. When the ring has wrapped, this
// allocates to return the lines in order; Line(i) is the
// allocation-free accessor.
func (s *Scrollback) Lines() []uv.Line {
	if s == nil {
		return nil
	}
	if s.head == 0 {
		return s.lines
	}
	out := make([]uv.Line, len(s.lines))
	n := copy(out, s.lines[s.head:])
	copy(out[n:], s.lines[:s.head])
	return out
}

// Clear removes all lines from the scrollback buffer.
func (s *Scrollback) Clear() {
	if s == nil {
		return
	}
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
