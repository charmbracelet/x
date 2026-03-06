package vt

import (
	"slices"

	uv "github.com/charmbracelet/ultraviolet"
)

// DefaultScrollbackSize is the default maximum number of lines in the scrollback buffer.
const DefaultScrollbackSize = 10000

// ScrollbackLine represents a line in the scrollback buffer with metadata.
type ScrollbackLine struct {
	// Cells contains the cell data for this line.
	Cells uv.Line
	// SoftWrapped indicates this line was soft-wrapped (continued on next line
	// due to terminal width, not an explicit newline). This enables proper
	// reflow on terminal resize.
	SoftWrapped bool
}

// Scrollback represents a scrollback buffer that stores lines scrolled off the screen.
type Scrollback struct {
	lines    []ScrollbackLine
	maxLines int
}

// NewScrollback creates a new scrollback buffer with the given maximum number of lines.
func NewScrollback(maxLines int) *Scrollback {
	if maxLines <= 0 {
		maxLines = DefaultScrollbackSize
	}
	return &Scrollback{
		lines:    make([]ScrollbackLine, 0, min(maxLines, 1000)), // Pre-allocate reasonable capacity
		maxLines: maxLines,
	}
}

// Push adds a line to the scrollback buffer.
// If the buffer is full, the oldest line is removed.
// The wrapped parameter indicates if this line was soft-wrapped (no explicit newline).
func (s *Scrollback) Push(line uv.Line, wrapped bool) {
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

	entry := ScrollbackLine{
		Cells:       cloned,
		SoftWrapped: wrapped,
	}

	if len(s.lines) >= s.maxLines {
		// Remove oldest line and append new one
		s.lines = slices.Delete(s.lines, 0, 1)
	}
	s.lines = append(s.lines, entry)
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
		// Remove oldest lines
		s.lines = s.lines[len(s.lines)-maxLines:]
	}
}

// Line returns the line cells at the given index.
// Index 0 is the oldest line, Len()-1 is the most recent.
// Returns nil if index is out of bounds.
func (s *Scrollback) Line(index int) uv.Line {
	if s == nil || index < 0 || index >= len(s.lines) {
		return nil
	}
	return s.lines[index].Cells
}

// LineEntry returns the full ScrollbackLine at the given index.
// Index 0 is the oldest line, Len()-1 is the most recent.
// Returns nil if index is out of bounds.
func (s *Scrollback) LineEntry(index int) *ScrollbackLine {
	if s == nil || index < 0 || index >= len(s.lines) {
		return nil
	}
	return &s.lines[index]
}

// Lines returns all line cells in the scrollback buffer.
// Index 0 is the oldest line.
func (s *Scrollback) Lines() []uv.Line {
	if s == nil {
		return nil
	}
	result := make([]uv.Line, len(s.lines))
	for i, entry := range s.lines {
		result[i] = entry.Cells
	}
	return result
}

// Clear removes all lines from the scrollback buffer.
func (s *Scrollback) Clear() {
	if s == nil {
		return
	}
	s.lines = s.lines[:0]
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

// Reflow reflows the scrollback buffer for a new terminal width.
// Lines that were soft-wrapped are joined and re-wrapped to the new width.
// This should be called when the terminal is resized.
func (s *Scrollback) Reflow(newWidth int) {
	if s == nil || len(s.lines) == 0 || newWidth <= 0 {
		return
	}

	// Collect all logical lines (joining wrapped physical lines)
	var logicalLines []uv.Line
	var current uv.Line

	for _, entry := range s.lines {
		current = append(current, entry.Cells...)
		if !entry.SoftWrapped {
			// End of logical line
			logicalLines = append(logicalLines, current)
			current = nil
		}
	}
	// Handle trailing wrapped line
	if len(current) > 0 {
		logicalLines = append(logicalLines, current)
	}

	// Re-wrap logical lines to new width
	s.lines = s.lines[:0]
	for _, logical := range logicalLines {
		if len(logical) == 0 {
			s.lines = append(s.lines, ScrollbackLine{Cells: nil, SoftWrapped: false})
			continue
		}

		// Split into chunks of newWidth
		for len(logical) > newWidth {
			chunk := slices.Clone(logical[:newWidth])
			s.lines = append(s.lines, ScrollbackLine{Cells: chunk, SoftWrapped: true})
			logical = logical[newWidth:]

			// Enforce max lines
			if len(s.lines) >= s.maxLines {
				s.lines = slices.Delete(s.lines, 0, 1)
			}
		}

		// Final chunk (not wrapped)
		if len(logical) > 0 {
			s.lines = append(s.lines, ScrollbackLine{Cells: slices.Clone(logical), SoftWrapped: false})
			if len(s.lines) >= s.maxLines {
				s.lines = slices.Delete(s.lines, 0, 1)
			}
		}
	}
}
