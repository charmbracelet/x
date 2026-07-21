package vt

import (
	"fmt"
	"slices"

	uv "github.com/charmbracelet/ultraviolet"
)

type logicalRows struct {
	cells        uv.Line
	headBoundary rowBoundary
	hasCursor    bool
	cursorOffset int
	hasSaved     bool
	savedOffset  int
}

type mappedPosition struct {
	x, y  int
	found bool
}

type packedLogicalRows struct {
	rows          []semanticRow
	cursor, saved mappedPosition
}

func (s *Screen) resizeReflow(width, height int) {
	if width <= 0 || height <= 0 {
		return
	}

	rows, screenStartBefore := s.semanticRows()
	groups := groupSemanticRows(
		rows,
		screenStartBefore+s.cur.Y,
		s.cur.X,
		screenStartBefore+s.saved.Y,
		s.saved.X,
	)

	var (
		reflowed       []semanticRow
		cursorGlobalY  int
		cursorX        int
		cursorWasFound bool
		savedGlobalY   int
		savedX         int
		savedWasFound  bool
	)
	for _, group := range groups {
		packed := packLogicalRows(group, width)
		if packed.cursor.found {
			cursorGlobalY = len(reflowed) + packed.cursor.y
			cursorX = packed.cursor.x
			cursorWasFound = true
		}
		if packed.saved.found {
			savedGlobalY = len(reflowed) + packed.saved.y
			savedX = packed.saved.x
			savedWasFound = true
		}
		reflowed = append(reflowed, packed.rows...)
	}

	for len(reflowed) < height {
		reflowed = append(reflowed, semanticRow{line: uv.NewLine(width), boundary: boundaryHard})
	}
	screenStart := max(0, len(reflowed)-height)
	if s.scrollback != nil {
		s.scrollback.replaceSemanticRows(reflowed[:screenStart])
	}
	s.buf.replace(width, height, reflowed[screenStart:])

	if cursorWasFound {
		s.cur.X = min(max(cursorX, 0), width-1)
		s.cur.Y = min(max(cursorGlobalY-screenStart, 0), height-1)
	} else {
		s.cur.X = min(max(s.cur.X, 0), width-1)
		s.cur.Y = min(max(s.cur.Y, 0), height-1)
	}
	if savedWasFound {
		s.saved.X = min(max(savedX, 0), width-1)
		s.saved.Y = min(max(savedGlobalY-screenStart, 0), height-1)
	} else {
		s.saved.X = min(max(s.saved.X, 0), width-1)
		s.saved.Y = min(max(s.saved.Y, 0), height-1)
	}
}

func (s *Screen) resizePhysical(width, height int) {
	if width <= 0 || height <= 0 {
		return
	}
	shift := s.buf.resizePhysical(width, height, s.cur.Y)
	s.cur.X = min(max(s.cur.X, 0), width-1)
	s.cur.Y = min(max(s.cur.Y+shift, 0), height-1)
	s.saved.X = min(max(s.saved.X, 0), width-1)
	s.saved.Y = min(max(s.saved.Y+shift, 0), height-1)
	s.scroll = s.buf.bounds()
}

func (s *Screen) semanticRows() ([]semanticRow, int) {
	var rows []semanticRow
	if s.scrollback != nil {
		rows = s.scrollback.semanticRows()
	}
	screenStart := len(rows)
	rows = append(rows, s.buf.allRows()...)
	return rows, screenStart
}

func groupSemanticRows(rows []semanticRow, cursorRow, cursorX, savedRow, savedX int) []logicalRows {
	groups := make([]logicalRows, 0, len(rows))
	for i, row := range rows {
		if len(groups) == 0 || row.boundary != boundarySoft {
			groups = append(groups, logicalRows{headBoundary: row.boundary})
		}
		group := &groups[len(groups)-1]
		take := semanticLineLength(row.line)
		if i+1 < len(rows) && rows[i+1].boundary == boundarySoft {
			take = len(row.line)
		}
		if i == cursorRow {
			take = max(take, min(cursorX, len(row.line)))
			group.hasCursor = true
			group.cursorOffset = len(group.cells) + min(cursorX, take)
		}
		if i == savedRow {
			take = max(take, min(savedX, len(row.line)))
			group.hasSaved = true
			group.savedOffset = len(group.cells) + min(savedX, take)
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

func packLogicalRows(group logicalRows, width int) packedLogicalRows {
	rows := []semanticRow{{line: uv.NewLine(width), boundary: group.headBoundary}}
	y, x := 0, 0
	var cursor, saved mappedPosition

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
			rows = append(rows, semanticRow{line: uv.NewLine(width), boundary: boundarySoft})
		}
		if group.hasCursor && !cursor.found && group.cursorOffset >= i && group.cursorOffset < i+cellWidth {
			cursor = mappedPosition{x: min(x+group.cursorOffset-i, width-1), y: y, found: true}
		}
		if group.hasSaved && !saved.found && group.savedOffset >= i && group.savedOffset < i+cellWidth {
			saved = mappedPosition{x: min(x+group.savedOffset-i, width-1), y: y, found: true}
		}
		rows[y].line.Set(x, &cell)
		x += cellWidth
		i += cellWidth
		if x >= width && i < len(group.cells) {
			y++
			x = 0
			rows = append(rows, semanticRow{line: uv.NewLine(width), boundary: boundarySoft})
		}
	}

	if group.hasCursor && !cursor.found {
		cursor = mappedPosition{x: min(x, width-1), y: y, found: true}
	}
	if group.hasSaved && !saved.found {
		saved = mappedPosition{x: min(x, width-1), y: y, found: true}
	}
	return packedLogicalRows{rows: rows, cursor: cursor, saved: saved}
}

func (s *Screen) snapshotRows(includeScrollback bool) ([]semanticRow, error) {
	if err := s.buf.validate(); err != nil {
		return nil, fmt.Errorf("visible semantic buffer: %w", err)
	}
	var rows []semanticRow
	if includeScrollback && s.scrollback != nil {
		rows = s.scrollback.semanticRows()
	}
	rows = append(rows, s.buf.allRows()...)
	for i, row := range rows {
		if !row.boundary.valid() {
			return nil, fmt.Errorf("snapshot row %d has invalid boundary %d", i, row.boundary)
		}
		rows[i].line = slices.Clone(row.line)
	}
	if len(rows) > 0 && rows[0].boundary == boundarySoft {
		return nil, fmt.Errorf("snapshot begins with soft boundary")
	}
	return rows, nil
}
