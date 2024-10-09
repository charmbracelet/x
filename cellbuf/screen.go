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

// Screen represents a 2D grid of cells.
type Screen struct {
	curr       *Buffer  // current buffer to be committed
	prev       *Buffer  // last committed buffer
	queueAbove []string // queue of lines to be written above the screen
	pos        Pos      // the current cursor position
	noAutoWrap bool     // whether autowrap is disabled
	altScreen  bool     // whether this is an alternate screen
}

var _ Window = &Screen{}

// SetWidth sets the width of the screen.
func (s *Screen) SetWidth(width int) {
	s.curr.width = width
}

// AbsX implements Window.
func (s *Screen) AbsX() int {
	return 0
}

// AbsY implements Window.
func (s *Screen) AbsY() int {
	return 0
}

// At implements Window.
func (s *Screen) At(x int, y int) (Cell, error) {
	return s.curr.At(x, y)
}

// Child implements Window.
func (s *Screen) Child(x int, y int, width int, height int) (Window, error) {
	return newChildWindow(s.curr, s, x, y, width, height)
}

// Height implements Window.
func (s *Screen) Height() int {
	return len(s.curr.cells) / s.curr.width
}

// Width implements Window.
func (s *Screen) Width() int {
	return s.curr.width
}

// newScreen creates a new screen with the given options.
func newScreen(altScreen bool, width int, method WidthMethod) *Screen {
	return &Screen{
		altScreen: altScreen,
		curr:      NewBuffer(width, method),
		prev:      NewBuffer(0, method),
	}
}

// Pos returns the x and y position of the cursor.
func (s *Screen) Pos() (x int, y int) {
	return s.pos.X, s.pos.Y
}

// Render returns a string representation of the screen with ANSI escape sequences.
func (s *Screen) Render() string {
	return s.curr.Render()
}

// RenderLine returns a string representation of the screen n line with ANSI
// escape sequences.
func (s *Screen) RenderLine(n int) string {
	return s.curr.RenderLine(n)
}

// Set sets the cell at the given position.
func (s *Screen) Set(x, y int, c Cell) {
	s.curr.Set(x, y, c)
}

// SetContent writes the given data to the buffer starting from the first cell.
func (s *Screen) SetContent(data string) {
	s.curr.SetContent(data)
}

// ClearScreen returns a string to clear the screen and moves the cursor to the
// origin location i.e. top-left.
func (s *Screen) ClearScreen() string {
	if s.altScreen {
		return ansi.EraseEntireDisplay + ansi.MoveCursorOrigin
	}

	var seq string
	if s.pos.X > 0 {
		seq += "\r"
	}
	if s.pos.Y > 0 {
		seq += ansi.CursorUp(s.pos.Y)
	}
	return seq + ansi.EraseDisplayBelow
}

// Repaint forces a full repaint of the screen.
func (s *Screen) Repaint() {
	if s.prev != nil && s.prev.width > 0 {
		s.prev = NewBuffer(0, s.curr.method)
	}
}

// InsertAbove inserts a line above the screen.
func (s *Screen) InsertAbove(line string) {
	if !s.altScreen {
		s.queueAbove = append(s.queueAbove, strings.Split(line, "\n")...)
	}
}

// Commit returns the necessary changes and commits the buffer.
func (s *Screen) Commit() string {
	var buf bytes.Buffer
	if !s.altScreen && len(s.queueAbove) > 0 {
		s.moveCursor(&buf, 0, 0)
		for _, line := range s.queueAbove {
			buf.WriteString(line + ansi.EraseLineRight + newline)
		}
		s.queueAbove = s.queueAbove[:0]
		s.Repaint()
	}

	if s.pos.X > s.Width()-1 {
		buf.WriteByte(ansi.CR)
		s.pos.X = 0
	}

	var pen CellStyle
	var link CellLink
	var pos Pos // Use to store/restore cursor position
	changes := Changes(s.prev, s.curr)
	for _, ch := range changes {
		// log.Printf("Change posX: %d, posY: %d, x: %d, y: %d", s.pos.X, s.pos.Y, ch.X, ch.Y)
		switch v := ch.Change.(type) {
		case ClearScreen:
			// log.Printf("ClearScreen")
			s.pos.X, s.pos.Y = 0, 0
			buf.WriteString(s.ClearScreen())
		case EraseRight:
			// log.Printf("EraseRight")
			s.moveCursor(&buf, ch.X, ch.Y)
			buf.WriteString(ansi.EraseLineRight)
		case EraseLine:
			// log.Printf("EraseLine")
			s.moveCursor(&buf, ch.X, ch.Y)
			buf.WriteString(ansi.EraseEntireLine)
		case SaveCursor:
			// log.Printf("SaveCursor")
			pos.X, pos.Y = s.pos.X, s.pos.Y
			buf.WriteString(ansi.SaveCursor)
		case RestoreCursor:
			// log.Printf("RestoreCursor")
			buf.WriteString(ansi.RestoreCursor)
			s.pos.X, s.pos.Y = pos.X, pos.Y
		case Line:
			// log.Printf("Line %q", v.Content)
			s.moveCursor(&buf, ch.X, ch.Y)

			if v.Content != "" {
				buf.WriteString(v.Content)
				s.pos.X += v.Width
			}
			if v.Erase {
				buf.WriteString(ansi.EraseLineRight)
			}

			if !s.noAutoWrap && s.pos.X > s.Width() {
				// When autowrap is enabled (the default DECAWM) and the cursor
				// draws the last cell on the row, it goes into a "phantom"
				// cell i.e. `x == width`. When the cursor is at "phantom"
				// cell, and autowrap is enabled, the terminal moves the cursor
				// to the beginning of the next line. So we need to keep track
				// of that and move the cursor position accordingly.
				s.pos.X -= s.Width() - 1
				s.pos.Y++
			}

			// Move the cursor to the next line if necessary.
			if s.pos.Y < s.Height()-1 {
				buf.WriteString(newline)
				s.pos.X = 0
				s.pos.Y++
			}
		case Segment:
			// log.Printf("Segment %q", v.Content)
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

	s.prev = s.curr.Clone()

	return buf.String()
}

// moveCursor moves the cursor to the given position.
func (s *Screen) moveCursor(b *bytes.Buffer, x, y int) {
	if s.altScreen && (x != s.pos.X || y != s.pos.Y) {
		b.WriteString(ansi.MoveCursor(x+1, y+1))
	} else {
		if s.pos.Y < y {
			diff := y - s.pos.Y
			if diff >= 3 {
				// [ansi.CursorDown] is at least 3 bytes long, so we use "\n" when
				// we can to avoid writing more bytes than necessary.
				b.WriteString(ansi.CursorDown(diff))
			} else {
				b.WriteString(strings.Repeat("\n", diff))
			}
		} else if s.pos.Y > y {
			diff := s.pos.Y - y
			b.WriteString(ansi.CursorUp(diff))
		}
		if s.pos.X < x {
			diff := x - s.pos.X
			switch diff {
			case 1:
				// A single space is more efficient than [ansi.CursorRight(1)]
				// which takes at least 3 bytes `ESC [ D`.
				if cell, _ := s.curr.At(s.pos.X, s.pos.Y); cell.Equal(spaceCell) {
					b.WriteByte(' ')
					break
				}
				fallthrough
			default:
				b.WriteString(ansi.CursorRight(diff))
			}
		} else if s.pos.X > x {
			if x == 0 {
				// We use [ansi.CR] instead of [ansi.CursorLeft] to avoid
				// writing multiple bytes.
				b.WriteByte(ansi.CR)
			} else {
				diff := s.pos.X - x
				if diff >= 3 {
					// [ansi.CursorLeft] is at least 3 bytes long, so we use [ansi.BS]
					// when we can to avoid writing more bytes than necessary.
					b.WriteString(ansi.CursorLeft(s.pos.X - x))
				} else {
					b.Write(bytes.Repeat([]byte{ansi.BS}, diff))
				}
			}
		}
	}

	s.pos.X, s.pos.Y = x, y
}

// NewScreen creates a new Screen with the given width and method.
func NewScreen(width int, method WidthMethod) *Screen {
	return newScreen(false, width, method)
}

// NewAltScreen creates a new AltScreen with the given width and method.
func NewAltScreen(width int, method WidthMethod) *Screen {
	return newScreen(true, width, method)
}
