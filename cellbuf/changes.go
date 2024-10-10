package cellbuf

// Line represents a single line of cells.
type Line struct {
	Content string
	Width   int
	Erase   bool // Whether to erase the rest of the line.
}

// Segment represents multiple cells with the same style and link.
type Segment struct {
	Style   CellStyle
	Link    CellLink
	Content string
	Width   int
}

// ClearScreen represents a clear screen operation.
type ClearScreen struct{}

// EraseRight represents an erase right operation.
type EraseRight struct{}

// EraseLine represents an erase line operation.
type EraseLine struct{}

// SaveCursor represents a save cursor position operation.
type SaveCursor struct{}

// RestoreCursor represents a restore cursor position operation.
type RestoreCursor struct{}

// Change represents a single change between two cell buffers.
type Change struct {
	X, Y int // The starting position of the change.

	// A change can be a [Cell], [Line], [Segment], or [ClearScreen].
	Change any
}

// Changes returns a list of changes between two cell buffers.
func Changes(a, b *Buffer) (chs []Change) {
	if a == nil || b == nil || b.width == 0 {
		return nil
	}

	bHeight := len(b.cells) / b.width
	if a.width == 0 || a.width != b.width {
		// Clear the screen and redraw everything if the widths are different.
		chs = append(chs, Change{Change: ClearScreen{}})
		for y := 0; y < bHeight; y++ {
			width, line := RenderLine(b, y)
			chs = append(chs, Change{
				X: 0, Y: y,
				Change: Line{line, width, false},
			})
		}
		return chs
	}

	// Find the different cells and create cells, segments, and lines.
	for y := 0; y < bHeight; y++ {
		var seg *Segment
		var startX int
		var x int
		for x = 0; x < b.width; x++ {
			cellA, _ := a.At(x, y)
			cellB, _ := b.At(x, y)

			if cellB.Equal(cellA) {
				if seg != nil {
					chs = append(chs, Change{
						X: startX, Y: y,
						Change: *seg,
					})
					seg = nil
				}
				continue
			}

			if seg == nil {
				startX = x
				seg = &Segment{
					Style:   cellB.Style,
					Link:    cellB.Link,
					Content: cellB.Content,
					Width:   cellB.Width,
				}
				continue
			}

			if !seg.Style.Equal(cellB.Style) || !seg.Link.Equal(cellB.Link) {
				chs = append(chs, Change{
					X: startX, Y: y,
					Change: *seg,
				})
				startX = x
				seg = &Segment{
					Style:   cellB.Style,
					Link:    cellB.Link,
					Content: cellB.Content,
					Width:   cellB.Width,
				}
				continue
			}

			seg.Content += cellB.Content
			seg.Width += cellB.Width

			if b.lastInLine(x, y) && !a.lastInLine(x, y) {
				// PERF: This is expensive. We should find a better way to handle this.
				chs = append(chs,
					Change{X: startX, Y: y, Change: *seg},
					Change{X: x + 1, Y: y, Change: EraseRight{}},
				)
				seg = nil

				// Skip to the next line. We already know that the rest of the line is spaces.
				break
			}
		}

		if seg != nil {
			chs = append(chs, Change{X: startX, Y: y, Change: *seg})
		}
	}

	// Delete any remaining lines in a.
	if a.width > 0 {
		aHeight := len(a.cells) / a.width
		if aHeight > bHeight {
			// Ensure the cursor is at the last line of the current buffer.
			chs = append(chs, Change{X: 0, Y: bHeight - 1, Change: Line{"", 0, false}})

			chs = append(chs, Change{Change: SaveCursor{}})
			for y := bHeight; y < aHeight; y++ {
				chs = append(chs, Change{X: 0, Y: y, Change: EraseLine{}})
			}
			chs = append(chs, Change{Change: RestoreCursor{}})
		}
	}

	return
}
