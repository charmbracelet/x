package cellbuf

import (
	"bytes"
	"strings"

	"github.com/charmbracelet/x/ansi"
)

const newline = "\r\n"

// Pos represents a position in a 2D grid.
type Pos struct {
	X, Y int
}

// screen represents a 2D grid of cells.
type screen struct {
	curr      *Buffer // current buffer to be committed
	prev      *Buffer // last committed buffer
	pos       Pos     // the current cursor position
	altScreen bool    // whether this is an alternate screen
}

// newScreen creates a new screen with the given options.
func newScreen(altScreen bool, width int, method WidthMethod) screen {
	return screen{
		altScreen: altScreen,
		curr:      NewBuffer(width, method),
		prev:      NewBuffer(0, method),
	}
}

// Render returns a string representation of the screen with ANSI escape sequences.
func (s *screen) Render() string {
	return s.curr.Render()
}

// RenderLine returns a string representation of the screen n line with ANSI
// escape sequences.
func (s *screen) RenderLine(n int) string {
	return s.curr.RenderLine(n)
}

// SetCell sets the cell at the given position.
func (s *screen) SetCell(x, y int, c Cell) {
	s.curr.Set(x, y, c)
}

// SetContent writes the given data to the buffer starting from the first cell.
func (s *screen) SetContent(data string) {
	s.curr.SetContent(data)
}

// Commit returns the necessary changes and commits the buffer.
func (s *screen) Commit() string {
	if s.curr == nil || s.prev == nil {
		return ""
	}
	// Return empty string if the buffers are the same.
	if s.curr.Equal(s.prev) {
		return ""
	}

	var buf bytes.Buffer
	var pen CellStyle
	var link CellLink
	changes := Changes(s.prev, s.curr)
	for _, ch := range changes {
		switch v := ch.Change.(type) {
		case ClearScreen:
			// log.Printf("ClearScreen\r\n")
			if s.altScreen {
				buf.WriteString(ansi.EraseEntireDisplay)
				buf.WriteString(ansi.MoveCursorOrigin)
			} else {
				buf.WriteByte(ansi.CR)
				buf.WriteString(ansi.CursorUp(s.pos.Y))
				buf.WriteString(ansi.EraseDisplayBelow)
			}
			s.pos = Pos{}
		case EraseRight:
			s.moveCursor(&buf, ch.X, ch.Y)
			buf.WriteString(ansi.EraseLineRight)
		case Line:
			// log.Printf("Line %q\r\n", v.Content)
			if len(v) == 0 {
				// Erase line
				buf.WriteString(ansi.EraseEntireLine)
			} else {
				if s.pos.X > 0 {
					buf.WriteByte(ansi.CR)
				}
				buf.WriteString(string(v))
			}
			buf.WriteString(newline)
			s.pos.X = 0
			s.pos.Y++
		case Segment:
			if v.Style.IsEmpty() && !pen.IsEmpty() {
				buf.WriteString(ansi.ResetStyle) //nolint:errcheck
				pen.Reset()
			}
			if v.Link != link && link.URL != "" {
				buf.WriteString(ansi.ResetHyperlink()) //nolint:errcheck
				link.Reset()
			}

			s.moveCursor(&buf, ch.X, ch.Y)

			// log.Printf("Segment x: %d, y: %d, %q\r\n", ch.X, ch.Y, v.Content)
			if !v.Style.Equal(pen) {
				buf.WriteString(v.Style.DiffSequence(pen)) // nolint:errcheck
				pen = v.Style
			}
			if v.Link != link {
				buf.WriteString(ansi.SetHyperlink(v.Link.URL, v.Link.URLID)) // nolint:errcheck
				link = v.Link
			}

			buf.WriteString(v.Content)
			s.pos.X += v.Width
		}
	}

	// Reset the style and hyperlink if necessary.
	if link.URL != "" {
		buf.WriteString(ansi.ResetHyperlink()) //nolint:errcheck
	}
	if !pen.IsEmpty() {
		buf.WriteString(ansi.ResetStyle) //nolint:errcheck
	}

	s.prev.width = s.curr.width
	if len(s.curr.cells) > len(s.prev.cells) {
		// Expand the buffer if necessary.
		s.prev.cells = append(s.prev.cells, make([]Cell, len(s.curr.cells)-len(s.prev.cells))...)
	}
	copy(s.prev.cells, s.curr.cells)

	return buf.String()
}

// moveCursor moves the cursor to the given position.
func (s *screen) moveCursor(b *bytes.Buffer, x, y int) {
	if s.altScreen {
		b.WriteString(ansi.MoveCursor(x, y))
	} else {
		if s.pos.Y < y {
			if diff := y - s.pos.Y; diff >= 3 {
				// [ansi.CursorDown] is at least 3 bytes long, so we use "\n" when
				// we can to avoid writing more bytes than necessary.
				b.WriteString(ansi.CursorDown(diff))
			} else {
				b.WriteString(strings.Repeat("\n", diff))
			}
		} else if s.pos.Y > y {
			b.WriteString(ansi.CursorUp(s.pos.Y - y))
		}
		if s.pos.X < x {
			b.WriteString(ansi.CursorRight(x - s.pos.X))
		} else if s.pos.X > x {
			if x == 0 {
				// We use [ansi.CR] instead of [ansi.CursorLeft] to avoid
				// writing multiple bytes.
				b.WriteByte(ansi.CR)
			} else if diff := s.pos.X - x; diff >= 3 {
				// [ansi.CursorLeft] is at least 3 bytes long, so we use [ansi.BS]
				// when we can to avoid writing more bytes than necessary.
				b.WriteString(ansi.CursorLeft(s.pos.X - x))
			} else {
				b.Write(bytes.Repeat([]byte{ansi.BS}, diff))
			}
		}
	}

	s.pos.X, s.pos.Y = x, y
}

// Screen represents a 2D inline screen.
type Screen struct {
	screen
}

// NewScreen creates a new Screen with the given width and method.
func NewScreen(width int, method WidthMethod) *Screen {
	return &Screen{
		screen: newScreen(false, width, method),
	}
}

// AltScreen represents a 2D alternate screen.
type AltScreen struct {
	screen
}

// NewAltScreen creates a new AltScreen with the given width and method.
func NewAltScreen(width int, method WidthMethod) *AltScreen {
	return &AltScreen{
		screen: newScreen(true, width, method),
	}
}
