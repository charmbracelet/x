package vt

import (
	"slices"

	uv "github.com/charmbracelet/ultraviolet"
)

type semanticRow struct {
	line          uv.Line
	wrapped       bool
	headTruncated bool
}

type logicalRows struct {
	cells         uv.Line
	headTruncated bool
	hasCursor     bool
	cursorOffset  int
}

func (s *Screen) resizeReflow(width, height int) {
	if width <= 0 || height <= 0 {
		return
	}

	rows, cursorRow := s.semanticRows()
	groups := groupSemanticRows(rows, cursorRow, s.cur.X)

	var (
		reflowed       []ScrollbackRow
		cursorGlobalY  int
		cursorX        int
		cursorWasFound bool
	)
	for _, group := range groups {
		packed, cy, cx, found := packLogicalRows(group, width)
		if found {
			cursorGlobalY = len(reflowed) + cy
			cursorX = cx
			cursorWasFound = true
		}
		reflowed = append(reflowed, packed...)
	}

	for len(reflowed) < height {
		reflowed = append(reflowed, ScrollbackRow{Line: uv.NewLine(width)})
	}
	screenStart := max(0, len(reflowed)-height)
	if s.scrollback != nil {
		s.scrollback.ReplaceRows(reflowed[:screenStart])
	}

	newbuf := uv.NewRenderBuffer(width, height)
	newWrapped := make([]bool, height)
	for y, row := range reflowed[screenStart:] {
		copy(newbuf.Line(y), row.Line)
		newWrapped[y] = row.Wrapped
	}
	newbuf.Touched = nil
	s.buf = newbuf
	s.wrapped = newWrapped

	if cursorWasFound {
		s.cur.X = min(max(cursorX, 0), width-1)
		s.cur.Y = min(max(cursorGlobalY-screenStart, 0), height-1)
	} else {
		s.cur.X = min(max(s.cur.X, 0), width-1)
		s.cur.Y = min(max(s.cur.Y, 0), height-1)
	}
	s.saved.X = min(max(s.saved.X, 0), width-1)
	s.saved.Y = min(max(s.saved.Y, 0), height-1)
}

func (s *Screen) semanticRows() ([]semanticRow, int) {
	scrollbackLen := 0
	if s.scrollback != nil {
		scrollbackLen = s.scrollback.Len()
	}
	rows := make([]semanticRow, 0, s.buf.Height()+scrollbackLen)
	if s.scrollback != nil {
		for _, row := range s.scrollback.Rows() {
			rows = append(rows, semanticRow{
				line:          slices.Clone(row.Line),
				wrapped:       row.Wrapped,
				headTruncated: row.HeadTruncated,
			})
		}
	}
	cursorRow := len(rows) + s.cur.Y
	for y := 0; y < s.buf.Height(); y++ {
		rows = append(rows, semanticRow{
			line:    slices.Clone(s.buf.Line(y)),
			wrapped: s.rowWrapped(y),
		})
	}
	return rows, cursorRow
}

func groupSemanticRows(rows []semanticRow, cursorRow, cursorX int) []logicalRows {
	groups := make([]logicalRows, 0, len(rows))
	for i, row := range rows {
		if len(groups) == 0 || !row.wrapped {
			groups = append(groups, logicalRows{headTruncated: row.headTruncated})
		}
		group := &groups[len(groups)-1]
		take := semanticLineLength(row.line)
		if i+1 < len(rows) && rows[i+1].wrapped {
			take = len(row.line)
		}
		if i == cursorRow {
			take = max(take, min(cursorX, len(row.line)))
			group.hasCursor = true
			group.cursorOffset = len(group.cells) + min(cursorX, take)
		}
		group.cells = append(group.cells, row.line[:take]...)
	}
	return groups
}

func semanticLineLength(line uv.Line) int {
	for i := len(line) - 1; i >= 0; i-- {
		cell := line[i]
		if !cell.IsZero() && !cell.Equal(&uv.EmptyCell) {
			return min(len(line), i+max(cell.Width, 1))
		}
	}
	return 0
}

func packLogicalRows(group logicalRows, width int) ([]ScrollbackRow, int, int, bool) {
	rows := []ScrollbackRow{{
		Line:          uv.NewLine(width),
		HeadTruncated: group.headTruncated,
	}}
	y, x := 0, 0
	cursorY, cursorX := 0, 0
	cursorMapped := false

	for i := 0; i < len(group.cells); {
		cell := group.cells[i]
		cellWidth := cell.Width
		if cellWidth <= 0 {
			if cell.Content == "" {
				i++
				continue
			}
			cellWidth = 1
		}
		if x > 0 && x+cellWidth > width {
			y++
			x = 0
			rows = append(rows, ScrollbackRow{Line: uv.NewLine(width), Wrapped: true})
		}
		if group.hasCursor && !cursorMapped && group.cursorOffset >= i && group.cursorOffset < i+cellWidth {
			cursorY = y
			cursorX = min(x+group.cursorOffset-i, width-1)
			cursorMapped = true
		}
		rows[y].Line.Set(x, &cell)
		x += cellWidth
		i += cellWidth
		if x >= width && i < len(group.cells) {
			y++
			x = 0
			rows = append(rows, ScrollbackRow{Line: uv.NewLine(width), Wrapped: true})
		}
	}

	if group.hasCursor && !cursorMapped {
		cursorY = y
		cursorX = min(x, width-1)
		cursorMapped = true
	}
	return rows, cursorY, cursorX, cursorMapped
}

func (s *Screen) snapshotRows(includeScrollback bool) []ScrollbackRow {
	scrollbackLen := 0
	if s.scrollback != nil {
		scrollbackLen = s.scrollback.Len()
	}
	rows := make([]ScrollbackRow, 0, s.buf.Height()+scrollbackLen)
	if includeScrollback && s.scrollback != nil {
		rows = append(rows, s.scrollback.Rows()...)
	}
	for y := 0; y < s.buf.Height(); y++ {
		rows = append(rows, ScrollbackRow{
			Line:    s.buf.Line(y),
			Wrapped: s.rowWrapped(y),
		})
	}
	return rows
}
