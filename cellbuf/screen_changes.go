package cellbuf

import (
	"bytes"

	"github.com/charmbracelet/x/ansi"
)

// changes returns a string to perform the changes of the dirty cell buffers
// based on the given from and to cursor positions.
func (s *Screen) changes(buf *bytes.Buffer) {
	width := s.buf.Width()
	if width <= 0 {
		return
	}

	height := s.buf.Height()
	var x int
	if s.lastRender == "" {
		// We render the changes line by line to be able to get the cursor
		// position using the width of each line.
		for y := 0; y < height; y++ {
			var line string
			x, line = RenderLine(s.buf, y)
			buf.WriteString(line)
			if y < height-1 {
				x = 0
				buf.WriteString("\r\n")
			}
		}

		s.pos.X, s.pos.Y = x, height-1
		buf.WriteString(Render(s.buf))
		return
	}

	// We use this to optimize moving the cursor around.
	ccur, _ := s.buf.At(s.pos.X, s.pos.Y)

	var pen CellStyle
	var link CellLink
	for y := 0; y < height; y++ {
		var seg *Segment
		var segX int    // The start position of the current segment.
		var eraser bool // Whether we're erasing using spaces and no styles or links.
		for x := 0; x < width; x++ {
			cell, _ := s.buf.At(x, y)
			if !s.buf.IsDirty(x, y) {
				if seg != nil {
					s.flushSegment(buf, seg, &pen, &link, Pos{segX, y}, &ccur, eraser)
					seg = nil
				}
				continue
			}

			if seg == nil {
				segX = x
				eraser = cell.Equal(spaceCell)
				seg = &Segment{
					Style:   cell.Style,
					Link:    cell.Link,
					Content: cell.Content,
					Width:   cell.Width,
				}
				continue
			}

			if !seg.Style.Equal(cell.Style) || seg.Link != cell.Link {
				s.flushSegment(buf, seg, &pen, &link, Pos{segX, y}, &ccur, eraser)
				segX = x
				eraser = cell.Equal(spaceCell)
				seg = &Segment{
					Style:   cell.Style,
					Link:    cell.Link,
					Content: cell.Content,
					Width:   cell.Width,
				}
				continue
			}

			eraser = eraser && cell.Equal(spaceCell)
			seg.Content += cell.Content
			seg.Width += cell.Width
		}

		if seg != nil {
			s.flushSegment(buf, seg, &pen, &link, Pos{segX, y}, &ccur, eraser)
			seg = nil
		}
	}

	// Delete extra lines from previous render.
	if s.lastHeight > height {
		// Move the cursor to the last line of this render.
		s.moveCursor(buf, &ccur, 0, height-1)
		// Save the cursor position to be restored later.
		buf.WriteString(ansi.SaveCursor) //nolint:errcheck
		for y := height; y < s.lastHeight; y++ {
			buf.WriteString(ansi.EraseEntireLine) //nolint:errcheck
			if y < s.lastHeight-1 {
				buf.WriteByte(ansi.LF) //nolint:errcheck
			}
		}
		buf.WriteString(ansi.RestoreCursor) //nolint:errcheck
	}

	// Reset the style and hyperlink if necessary.
	if link.URL != "" {
		buf.WriteString(ansi.ResetHyperlink()) //nolint:errcheck
	}
	if !pen.IsEmpty() {
		buf.WriteString(ansi.ResetStyle) //nolint:errcheck
	}
}

func (s *Screen) flushSegment(
	buf *bytes.Buffer, seg *Segment, pen *CellStyle,
	link *CellLink, to Pos, curCell *Cell, eraser bool,
) {
	if s.pos != to {
		renderReset(buf, seg, pen, link)
		s.moveCursor(buf, curCell, to.X, to.Y)
	}

	// We use [ansi.EraseLineRight] to erase the rest of the line if the segment
	// is an "eraser" i.e. it's just a bunch of spaces with no styles or links. We erase the
	// rest of the line when:
	// 1. The segment is an eraser.
	// 2. The segment reaches the end of the line to erase i.e. the new line is shorter.
	// 3. The segment takes more bytes than [ansi.EraseLineRight] to erase which is 4 bytes.
	if eraser && to.Y < len(s.linew) && seg.Width > 4 && s.linew[to.Y] < seg.Width+to.X {
		buf.WriteString(ansi.EraseLineRight) //nolint:errcheck
	} else {
		renderSegment(buf, seg, pen, link)
		s.pos.X += seg.Width
	}
}

func renderReset(buf *bytes.Buffer, seg *Segment, pen *CellStyle, link *CellLink) {
	if seg.Link != *link && link.URL != "" {
		buf.WriteString(ansi.ResetHyperlink()) //nolint:errcheck
		link.Reset()
	}
	if seg.Style.IsEmpty() && !pen.IsEmpty() {
		buf.WriteString(ansi.ResetStyle) //nolint:errcheck
		pen.Reset()
	}
}

func renderSegment(buf *bytes.Buffer, seg *Segment, pen *CellStyle, link *CellLink) {
	if !seg.Style.Equal(*pen) {
		buf.WriteString(seg.Style.DiffSequence(*pen)) // nolint:errcheck
		*pen = seg.Style
	}
	if seg.Link != *link {
		buf.WriteString(ansi.SetHyperlink(seg.Link.URL, seg.Link.URLID)) // nolint:errcheck
		*link = seg.Link
	}

	buf.WriteString(seg.Content)
}
