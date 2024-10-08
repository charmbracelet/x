package cellbuf

// Line represents a single line of cells.
type Line string

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

// Change represents a single change between two cell buffers.
type Change struct {
	X, Y int // The starting position of the change.

	// A change can be a [Cell], [Line], [Segment], or [ClearScreen].
	Change any
}

// Changes returns a list of changes between two cell buffers.
func Changes(a, b *Buffer) (chs []Change) {
	if a == nil || b == nil {
		return nil
	}

	bHeight := len(b.cells) / b.width
	if a.width == 0 || a.width != b.width {
		// Clear the screen and redraw everything if the widths are different.
		chs = append(chs, Change{Change: ClearScreen{}})
		for y := 0; y < bHeight; y++ {
			chs = append(chs, Change{
				X: 0, Y: y,
				Change: Line(b.RenderLine(y)),
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

			if b.lastInLine(x, y) && !a.lastInLine(x, y) {
				// PERF: This is expensive. We should find a better way to handle this.
				chs = append(chs,
					Change{X: startX, Y: y, Change: *seg},
					Change{X: x, Y: y, Change: EraseRight{}},
				)
				seg = nil
				break
			}

			seg.Content += cellB.Content
			seg.Width += cellB.Width
		}

		if seg != nil {
			chs = append(chs, Change{X: startX, Y: y, Change: *seg})
		}
	}

	// Return any remaining lines.
	aHeight := len(a.cells) / a.width
	if aHeight > bHeight {
		for y := bHeight; y < aHeight; y++ {
			chs = append(chs, Change{
				X: 0, Y: y,
				Change: Line(""), // Empty line signals a clear line.
			})
		}
	}

	return
}
