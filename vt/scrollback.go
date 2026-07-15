package vt

import (
	"slices"

	uv "github.com/charmbracelet/ultraviolet"
)

// DefaultScrollbackSize is the default maximum number of lines in the scrollback buffer.
const DefaultScrollbackSize = 10000

// Scrollback represents a scrollback buffer that stores lines scrolled off the screen.
type Scrollback struct {
	rows     []ScrollbackRow
	maxLines int
}

// ScrollbackRow is a physical terminal row plus the semantic boundary that
// precedes it. Wrapped is true only when the row continues a terminal
// autowrap from the previous retained row. HeadTruncated marks a retained
// fragment whose preceding soft-wrapped rows were evicted by the cap.
type ScrollbackRow struct {
	Line          uv.Line
	Wrapped       bool
	HeadTruncated bool
}

// NewScrollback creates a new scrollback buffer with the given maximum number of lines.
func NewScrollback(maxLines int) *Scrollback {
	if maxLines <= 0 {
		maxLines = DefaultScrollbackSize
	}
	return &Scrollback{
		rows:     make([]ScrollbackRow, 0, min(maxLines, 1000)), // Pre-allocate reasonable capacity
		maxLines: maxLines,
	}
}

// Push adds a line to the scrollback buffer.
// If the buffer is full, the oldest line is removed.
func (s *Scrollback) Push(line uv.Line) {
	s.PushWrapped(line, false)
}

// PushWrapped adds a row while preserving whether it is a soft-wrap
// continuation of the previous physical row.
func (s *Scrollback) PushWrapped(line uv.Line, wrapped bool) {
	if s == nil || s.maxLines <= 0 {
		return
	}

	// Find last non-empty cell to trim trailing empty cells.
	// This helps with wrapping and window resizing.
	lastNonEmpty := -1
	for i := len(line) - 1; i >= 0; i-- {
		c := &line[i]
		if !c.IsZero() && !c.Equal(&uv.EmptyCell) {
			lastNonEmpty = i
			break
		}
	}

	// Clone the line content up to and including the last non-empty cell
	cloned := slices.Clone(line[:lastNonEmpty+1])

	evicted := len(s.rows) >= s.maxLines
	if evicted {
		// Remove oldest line and append new one
		s.rows = slices.Delete(s.rows, 0, 1)
		if len(s.rows) > 0 && s.rows[0].Wrapped {
			s.rows[0].Wrapped = false
			s.rows[0].HeadTruncated = true
		}
	}
	s.rows = append(s.rows, ScrollbackRow{Line: cloned, Wrapped: wrapped})
	if evicted && len(s.rows) > 0 && s.rows[0].Wrapped {
		s.rows[0].Wrapped = false
		s.rows[0].HeadTruncated = true
	}
}

// PushN adds n lines from the buffer starting at line y to the scrollback.
func (s *Scrollback) PushN(buf *uv.RenderBuffer, wrapped []bool, y, n int) {
	if s == nil || buf == nil || n <= 0 {
		return
	}

	for i := range min(n, buf.Height()-y) {
		if line := buf.Line(y + i); line != nil {
			isWrapped := y+i < len(wrapped) && wrapped[y+i]
			s.PushWrapped(line, isWrapped)
		}
	}
}

// Len returns the number of lines in the scrollback buffer.
func (s *Scrollback) Len() int {
	if s == nil {
		return 0
	}
	return len(s.rows)
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
	if len(s.rows) > maxLines {
		// Remove oldest lines
		s.rows = s.rows[len(s.rows)-maxLines:]
		if len(s.rows) > 0 && s.rows[0].Wrapped {
			s.rows[0].Wrapped = false
			s.rows[0].HeadTruncated = true
		}
	}
}

// Line returns the line at the given index.
// Index 0 is the oldest line, Len()-1 is the most recent.
// Returns nil if index is out of bounds.
func (s *Scrollback) Line(index int) uv.Line {
	if s == nil || index < 0 || index >= len(s.rows) {
		return nil
	}
	return s.rows[index].Line
}

// Lines returns all lines in the scrollback buffer.
// Index 0 is the oldest line.
func (s *Scrollback) Lines() []uv.Line {
	if s == nil {
		return nil
	}
	lines := make([]uv.Line, len(s.rows))
	for i := range s.rows {
		lines[i] = s.rows[i].Line
	}
	return lines
}

// Rows returns the retained rows and their boundary provenance.
func (s *Scrollback) Rows() []ScrollbackRow {
	if s == nil {
		return nil
	}
	return s.rows
}

// ReplaceRows atomically replaces retained history after a semantic reflow.
func (s *Scrollback) ReplaceRows(rows []ScrollbackRow) {
	if s == nil {
		return
	}
	if len(rows) > s.maxLines {
		rows = rows[len(rows)-s.maxLines:]
		if len(rows) > 0 && rows[0].Wrapped {
			rows[0].Wrapped = false
			rows[0].HeadTruncated = true
		}
	}
	s.rows = slices.Clone(rows)
}

// Clear removes all lines from the scrollback buffer.
func (s *Scrollback) Clear() {
	if s == nil {
		return
	}
	s.rows = s.rows[:0]
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
