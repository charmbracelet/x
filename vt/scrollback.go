package vt

import (
	"slices"

	uv "github.com/charmbracelet/ultraviolet"
)

// DefaultScrollbackSize is the default maximum number of lines in the scrollback buffer.
const DefaultScrollbackSize = 10000

// Scrollback stores retained semantic rows. Its public read API returns value
// projections so callers cannot mutate retained terminal state through aliases.
type Scrollback struct {
	rows     []semanticRow
	maxLines int
}

// NewScrollback creates a new scrollback buffer with the given maximum number of lines.
func NewScrollback(maxLines int) *Scrollback {
	if maxLines <= 0 {
		maxLines = DefaultScrollbackSize
	}
	return &Scrollback{
		rows:     make([]semanticRow, 0, min(maxLines, 1000)),
		maxLines: maxLines,
	}
}

// Push adds a hard-boundary line to the scrollback buffer.
func (s *Scrollback) Push(line uv.Line) {
	s.pushSemanticRow(semanticRow{line: line, boundary: boundaryHard})
}

// PushN adds n hard-boundary lines from the buffer starting at line y.
// Semantic screen transfers use pushSemanticRows so boundary provenance stays
// inside the owning package rather than widening this established API.
func (s *Scrollback) PushN(buf *uv.RenderBuffer, y, n int) {
	if s == nil || buf == nil || n <= 0 || y < 0 || y >= buf.Height() {
		return
	}
	for i := range min(n, buf.Height()-y) {
		if line := buf.Line(y + i); line != nil {
			s.Push(line)
		}
	}
}

func (s *Scrollback) pushSemanticRows(rows []semanticRow) {
	for _, row := range rows {
		s.pushSemanticRow(row)
	}
}

func (s *Scrollback) pushSemanticRow(row semanticRow) {
	if s == nil || s.maxLines <= 0 {
		return
	}
	row.line = trimAndCloneLine(row.line)
	if !row.boundary.valid() {
		row.boundary = boundaryHard
	}
	if len(s.rows) == 0 && row.boundary == boundarySoft {
		row.boundary = boundaryTruncatedHead
	}
	s.rows = append(s.rows, row)
	if len(s.rows) > s.maxLines {
		s.rows = slices.Delete(s.rows, 0, len(s.rows)-s.maxLines)
		s.normalizeHead()
	}
}

func trimAndCloneLine(line uv.Line) uv.Line {
	lastNonEmpty := -1
	for i := len(line) - 1; i >= 0; i-- {
		cell := &line[i]
		if !cell.IsZero() && !cell.Equal(&uv.EmptyCell) {
			lastNonEmpty = i
			break
		}
	}
	return slices.Clone(line[:lastNonEmpty+1])
}

func (s *Scrollback) normalizeHead() {
	if len(s.rows) > 0 && s.rows[0].boundary == boundarySoft {
		s.rows[0].boundary = boundaryTruncatedHead
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
func (s *Scrollback) SetMaxLines(maxLines int) {
	if s == nil || maxLines <= 0 {
		return
	}
	s.maxLines = maxLines
	if len(s.rows) > maxLines {
		s.rows = slices.Clone(s.rows[len(s.rows)-maxLines:])
		s.normalizeHead()
	}
}

// Line returns a defensive copy of the line at the given index.
func (s *Scrollback) Line(index int) uv.Line {
	if s == nil || index < 0 || index >= len(s.rows) {
		return nil
	}
	return slices.Clone(s.rows[index].line)
}

// Lines returns defensive copies of all retained lines, oldest first.
func (s *Scrollback) Lines() []uv.Line {
	if s == nil {
		return nil
	}
	lines := make([]uv.Line, len(s.rows))
	for i := range s.rows {
		lines[i] = slices.Clone(s.rows[i].line)
	}
	return lines
}

func (s *Scrollback) semanticRows() []semanticRow {
	if s == nil {
		return nil
	}
	rows := make([]semanticRow, len(s.rows))
	for i, row := range s.rows {
		rows[i] = cloneSemanticRow(row)
	}
	return rows
}

func (s *Scrollback) replaceSemanticRows(rows []semanticRow) {
	if s == nil {
		return
	}
	if len(rows) > s.maxLines {
		rows = rows[len(rows)-s.maxLines:]
	}
	s.rows = make([]semanticRow, len(rows))
	for i, row := range rows {
		s.rows[i] = cloneSemanticRow(row)
	}
	s.normalizeHead()
}

// Clear removes all lines from the scrollback buffer.
func (s *Scrollback) Clear() {
	if s != nil {
		s.rows = s.rows[:0]
	}
}

// CellAt returns a defensive copy of a cell in the scrollback buffer.
func (s *Scrollback) CellAt(x, y int) *uv.Cell {
	if s == nil || y < 0 || y >= len(s.rows) || x < 0 || x >= len(s.rows[y].line) {
		return nil
	}
	return s.rows[y].line[x].Clone()
}
