package cellbuf

import (
	"bytes"
	"strings"

	"github.com/charmbracelet/x/ansi"
)

// Pos represents a position in a 2D grid.
type Pos struct {
	X, Y int
}

// Screen represents a 2D grid of cells.
type Screen struct {
	// buf   *Buffer     // current buffer to be committed
	// dirty map[int]int // represents the dirty cells, when nil, all cells are dirty

	buf         *DirtyBuffer
	lastContent string   // the last set content
	lastRender  string   // the last committed render of the screen
	queueAbove  []string // queue of lines to be written above the screen
	linew       []int    // the width of each line

	pos        Pos // the current cursor position
	lastHeight int // the last height of the screen
	// noAutoWrap  bool     // whether autowrap is disabled
	altScreen bool // whether this is an alternate screen
}

var _ Window = &Screen{}

// SetWidth sets the width of the screen.
func (s *Screen) SetWidth(width int) {
	s.buf.SetWidth(width)
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
	return s.buf.At(x, y)
}

// Child implements Window.
func (s *Screen) Child(x int, y int, width int, height int) (Window, error) {
	return newChildWindow(s.buf, s, x, y, width, height)
}

// Height implements Window.
func (s *Screen) Height() int {
	return s.buf.Height()
}

// Width implements Window.
func (s *Screen) Width() int {
	return s.buf.Width()
}

// newScreen creates a new screen with the given options.
func newScreen(altScreen bool, width int, method WidthMethod) *Screen {
	return &Screen{
		altScreen: altScreen,
		buf:       NewDirtyBuffer(width, method),
	}
}

// Pos returns the x and y position of the cursor.
func (s *Screen) Pos() (x int, y int) {
	return s.pos.X, s.pos.Y
}

// Set sets the cell at the given position.
func (s *Screen) Set(x, y int, c Cell) {
	s.buf.Set(x, y, c)
}

// SetContent writes the given data to the buffer starting from the first cell.
func (s *Screen) SetContent(data string) []int {
	s.lastContent = data
	linew := s.buf.SetContent(data)
	s.linew = linew
	return linew
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
	s.lastRender = ""
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
		s.moveCursor(&buf, nil, 0, 0)
		for _, line := range s.queueAbove {
			buf.WriteString(line + ansi.EraseLineRight + "\r\n")
		}
		s.queueAbove = s.queueAbove[:0]
		s.Repaint()
	}

	if s.pos.X > s.Width()-1 {
		buf.WriteByte(ansi.CR)
		s.pos.X = 0
	}

	if s.lastRender == "" {
		// First render clear the screen.
		s.pos.X, s.pos.Y = 0, 0
		s.moveCursor(&buf, nil, 0, 0)
		buf.WriteString(s.ClearScreen())
	}

	s.changes(&buf)

	if s.lastContent != "" {
		s.lastRender = s.lastContent
	}
	s.lastHeight = Height(s.lastRender)
	s.buf.Commit()

	return buf.String()
}

// moveCursor moves the cursor to the given position.
func (s *Screen) moveCursor(b *bytes.Buffer, curCell *Cell, x, y int) {
	if s.pos.X == x && s.pos.Y == y {
		return
	}

	if s.altScreen {
		b.WriteString(ansi.MoveCursor(y+1, x+1))
	} else {
		if s.pos.X < x {
			diff := x - s.pos.X
			switch diff {
			case 1:
				// We check if the cell at cursor position is a space cell.
				// A single space is more efficient than [ansi.CursorRight(1)]
				// which takes at least 3 bytes `ESC [ D`.
				if curCell != nil && curCell.Equal(spaceCell) {
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
	}

	s.pos.X, s.pos.Y = x, y
	if curCell != nil {
		*curCell, _ = s.At(s.pos.X, s.pos.Y)
	}
}

// NewScreen creates a new Screen with the given width and method.
func NewScreen(width int, method WidthMethod) *Screen {
	return newScreen(false, width, method)
}

// NewAltScreen creates a new AltScreen with the given width and method.
func NewAltScreen(width int, method WidthMethod) *Screen {
	return newScreen(true, width, method)
}
