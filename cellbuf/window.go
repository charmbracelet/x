package cellbuf

import (
	"bytes"
	"io"
	"log"
	"strings"

	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/term"
)

type cmd interface {
	sequence(*Screen) string
}

type clearCmd int

const (
	clearBelow clearCmd = iota
	clearAbove
	clearScreen
	clearRight
	clearLeft
	clearLine
)

func (c clearCmd) sequence(*Screen) (seq string) {
	switch c {
	case clearBelow, clearAbove, clearScreen:
		seq = ansi.EraseDisplay(int(c))
	case clearRight, clearLeft, clearLine:
		seq = ansi.EraseLine(int(c - clearRight))
	}
	return
}

type scrollUpCmd struct {
	region Rectangle
	n      int
}

func (s scrollUpCmd) sequence(sc *Screen) (seq string) {
	if !s.region.Empty() {
		seq += ansi.SetTopBottomMargins(s.region.Min.Y+1, s.region.Max.Y)
		seq += ansi.SetCursorPosition(s.region.Min.X+1, s.region.Min.Y+1)
	}

	seq += ansi.ScrollUp(s.n)

	if !s.region.Empty() {
		seq += ansi.SetTopBottomMargins(1, sc.height)
	}

	seq += ansi.SetCursorPosition(sc.cur.X+1, sc.cur.Y+1)
	return
}

type setPenCmd struct {
	Style
}

func (p setPenCmd) sequence(s *Screen) (seq string) {
	if !p.Equal(s.cur.Style) {
		seq = p.Sequence()
		s.cur.Style = p.Style
	}
	return
}

// isLocal returns whether the coordinates are considered local movement using
// the defined thresholds.
// This takes the number of columns, and the coordinates of the current and
// target positions.
func isLocal(cols, fx, fy, tx, ty int) bool {
	// The typical distance for a [ansi.CUP] sequence. Anything less than this
	// is considered local movement.
	const longDist = 8 - 1
	return (tx > longDist) &&
		(tx < cols-1-longDist) &&
		(abs(ty-fy)+abs(tx-fx) > longDist)
}

// relativeCursorMove returns the relative cursor movement sequence using one or two
// of the following sequences [ansi.CUU], [ansi.CUD], [ansi.CUF], [ansi.CUB],
// [ansi.VPA], [ansi.HPA].
// When overwrite is true, this will try to optimize the sequence by using the
// screen cells values to move the cursor instead of using escape sequences.
func relativeCursorMove(s *Screen, fx, fy, tx, ty int, overwrite bool) (seq string) {
	if ty != fy {
		yseq := ansi.VerticalPositionAbsolute(ty + 1)

		// OPTIM: Use [ansi.LF] and [ansi.ReverseIndex] as optimizations.

		if ty > fy {
			n := ty - fy
			if cud := ansi.CursorDown(n); len(cud) < len(yseq) {
				yseq = cud
			}
		} else if ty < fy {
			n := fy - ty
			if cuu := ansi.CursorUp(n); len(cuu) < len(yseq) {
				yseq = cuu
			}
		}

		seq += yseq
	}

	if tx != fx {
		xseq := ansi.HorizontalPositionAbsolute(tx + 1)

		if tx > fx {
			n := tx - fx
			if cuf := ansi.CursorForward(n); len(cuf) < len(xseq) {
				xseq = cuf
			}

			// OPTIM: Use [ansi.HT] and hard tabs as an optimization.

			// If we have no attribute and style changes, overwrite is cheaper.
			if overwrite && ty >= 0 {
				for i := 0; i < n; i++ {
					cell := s.buf.Cell(fx+i, ty)
					if cell != nil {
						i += cell.Width - 1
						if !cell.Style.Equal(s.cur.Style) || !cell.Link.Equal(s.cur.Link) {
							overwrite = false
							break
						}
					}
				}
			}

			if overwrite && ty >= 0 {
				for i := 0; i < n; i++ {
					cell := s.buf.Cell(fx+i, ty)
					if cell != nil {
						xseq += cell.Content
						i += cell.Width - 1
					} else {
						xseq += " "
					}
				}
			}
		} else if tx < fx {
			n := fx - tx
			if cub := ansi.CursorBackward(n); len(cub) < len(xseq) {
				xseq = cub
			}

			// OPTIM: Use back tabs as an optimization.
		}

		seq += xseq
	}

	return
}

// moveCursor moves and returns the cursor movement sequence to move the cursor
// to the specified position.
// When overwrite is true, this will try to optimize the sequence by using the
// screen cells values to move the cursor instead of using escape sequences.
func moveCursor(s *Screen, x, y int, overwrite bool) (seq string) {
	fx, fy := s.cur.X, s.cur.Y

	if isLocal(s.width, fx, fy, x, y) {
		// Method #0: Use [ansi.CUP] if the distance is long.
		seq = ansi.CursorPosition(x+1, y+1)
	} else {
		// Method #1: Use local movement sequences.
		nseq := relativeCursorMove(s, fx, fy, x, y, overwrite)
		if len(nseq) < len(seq) {
			seq = nseq
		}

		// Method #2: Use [ansi.CR] and local movement sequences.
		nseq = "\r" + relativeCursorMove(s, 0, fy, x, y, overwrite)
		if len(nseq) < len(seq) {
			seq = nseq
		}

		// Method #3: Use [ansi.HomeCursorPosition] and local movement sequences.
		nseq = ansi.HomeCursorPosition + relativeCursorMove(s, 0, 0, x, y, overwrite)
		if len(nseq) < len(seq) {
			seq = nseq
		}
	}

	s.cur.X, s.cur.Y = x, y
	return
}

func move(s *Screen, x, y int) (seq string) {
	if s.cur.X == x && s.cur.Y == y {
		return
	}

	if x >= s.width {
		// Handle autowrap
		y += (x / s.width)
		x %= s.width
	}

	// Disable styles if there's any
	var pen Style
	if !s.cur.Style.Empty() {
		pen = s.cur.Style
		seq += ansi.ResetStyle
	}

	if s.cur.X >= s.width {
		l := (s.cur.X + 1) / s.width

		s.cur.Y += l
		if s.cur.Y >= s.height {
			l -= s.cur.Y - s.height - 1
		}

		if l > 0 {
			seq += "\r"
			s.cur.X = 0
			seq += strings.Repeat("\n", l)
		}
	}

	if s.cur.Y > s.height-1 {
		s.cur.Y = s.height - 1
	}
	if y > s.height-1 {
		y = s.height - 1
	}

	seq += moveCursor(s, x, y, true) // Overwrite cells if possible

	if !pen.Empty() {
		seq += pen.Sequence()
	}

	return
}

type moveCmd struct {
	x, y int
}

func (m moveCmd) sequence(s *Screen) string {
	return move(s, m.x, m.y)
}

func resetPen(s *Screen) (seq string) {
	if !s.cur.Link.Empty() {
		seq += ansi.ResetHyperlink()
		s.cur.Link.Reset()
	}
	if !s.cur.Style.Empty() {
		seq += ansi.ResetStyle
		s.cur.Style.Reset()
	}
	return
}

type posCmd struct {
	x, y int
}

func (p posCmd) sequence(s *Screen) (seq string) {
	// Did we already render this cell?
	pos := Pos(p.x, p.y)
	if _, ok := s.dirty[pos]; !ok {
		return
	}

	delete(s.dirty, pos)

	x, y := p.x, p.y
	cell := s.buf.Cell(x, y)
	if cell == nil {
		return ""
	}

	// Do we need to render the cell?
	prev := s.bufs[1].Cell(x, y)
	if prev != nil && prev.Equal(cell) {
		return
	}

	if s.cur.X != x || s.cur.Y != y {
		// Do we need to reset the style and hyperlink?
		if s.cur.X+cell.Width != x || s.cur.Y != y {
			seq += resetPen(s)
		}

		seq += move(s, x, y)
	}

	if !cell.Style.Empty() && !cell.Style.Equal(s.cur.Style) {
		seq += cell.Style.DiffSequence(s.cur.Style)
		s.cur.Style = cell.Style
	}
	if !cell.Link.Empty() && !cell.Link.Equal(s.cur.Link) {
		seq += ansi.SetHyperlink(cell.Link.URL, cell.Link.URLID)
		s.cur.Link = cell.Link
	}

	seq += cell.Content
	s.cur.X += cell.Width

	return
}

// cursor represents a terminal cursor.
type cursor struct {
	Style Style
	Link  Link
	Position
}

// ScreenOptions are options for the screen.
type ScreenOptions struct {
	// Term is the terminal type to use when writing to the screen.
	Term string
	// Profile is the color profile to use when writing to the screen.
	Profile colorprofile.Profile
	// NoAutoWrap is whether not to automatically wrap text when it reaches the
	// end of the line.
	NoAutoWrap bool
	// Origin is whether to use origin mode.
	Origin bool
}

// Screen represents the terminal screen.
type Screen struct {
	w             io.Writer
	q             []cmd // queue of commands to be written to the terminal
	curs          [2]cursor
	cur           *cursor               // the current active cursor
	dirty         map[Position]struct{} // dirty cells that need to be redrawn
	opts          ScreenOptions
	bufs          [2]*Buffer // the current active buffer
	buf           *Buffer    // the current active buffer
	width, height int
}

var _ Window = &Screen{}

// NewScreen creates a new Screen.
func NewScreen(w io.Writer, opts *ScreenOptions) (s *Screen) {
	s = new(Screen)
	s.w = w
	s.bufs[0] = new(Buffer)
	s.bufs[1] = new(Buffer)
	s.buf = s.bufs[0]
	s.cur = &s.curs[0]
	s.dirty = make(map[Position]struct{})
	if opts != nil {
		s.opts = *opts
	}

	if f, ok := w.(term.File); ok {
		width, height, err := term.GetSize(f.Fd())
		if err == nil {
			s.Resize(width, height)
		}
	}

	return
}

// queue queues a command to be written to the terminal on the next call to
// [Screen.Render].
func (s *Screen) queue(cmd cmd) {
	s.q = append(s.q, cmd)
}

// execute writes a command to the terminal immediately.
func (s *Screen) execute(cmd string) {
	log.Printf("execute: %q", cmd)
	s.w.Write([]byte(cmd)) //nolint:errcheck
}

// Bounds implements Window.
func (s *Screen) Bounds() Rectangle {
	return Rect(0, 0, s.width, s.height)
}

// Cell implements Window.
func (s *Screen) Cell(x int, y int) *Cell {
	return s.buf.Cell(x, y)
}

// SetPen sets the current cursor pen styles. Note that the cursor pen might
// get reset when moving the cursor and drawing cells.
func (s *Screen) SetPen(style Style) {
	s.q = append(s.q, setPenCmd{style})
	s.cur.Style = style
}

// Fill implements Window.
func (s *Screen) Fill(cell *Cell) {
	s.FillInRect(cell, s.Bounds())
}

// FillInRect fills the cells in the specified rectangle with the specified
// cell.
func (s *Screen) FillInRect(cell *Cell, r Rectangle) {
	s.bufs[0].FillInRect(cell, r)

	switch {
	case cell != nil && cell.Width == 1 && cell.Content == " " && !cell.Style.Empty():
		s.SetPen(cell.Style)
		fallthrough
	case cell == nil || cell.Equal(&BlankCell):
		s.ClearInRect(r)
	}

	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x += cell.Width {
			s.Draw(x, y, cell)
		}
	}
}

// Clear implements Window.
func (s *Screen) Clear() {
	s.ClearInRect(s.Bounds())
}

// ClearInRect clears the cells in the specified rectangle based on the current
// cursor background color. Use [SetPen] to set the background color.
func (s *Screen) ClearInRect(r Rectangle) {
	s.bufs[0].ClearInRect(r)
	s.bufs[1].ClearInRect(r)

	if s.width == r.Width() {
		// Above, below, or the entire screen
		switch {
		case r.Min.Y == 0 && r.Max.Y == s.height:
			s.q = append(s.q, clearScreen)
		case r.Min.Y == 0:
			s.q = append(s.q, moveCmd{x: r.Min.X, y: r.Min.Y})
			s.q = append(s.q, clearAbove)
		case r.Max.Y == s.height:
			s.q = append(s.q, moveCmd{x: r.Min.X, y: r.Min.Y})
			s.q = append(s.q, clearBelow)
		case r.Height() == 1:
			// Left, right, or the entire line
			switch {
			case r.Width() == s.width:
				s.q = append(s.q, moveCmd{x: r.Min.X, y: r.Min.Y})
				s.q = append(s.q, clearLine)
			case r.Min.X == 0:
				s.q = append(s.q, moveCmd{x: r.Max.X - 1, y: r.Min.Y})
				s.q = append(s.q, clearLeft)
			case r.Max.X == s.width:
				s.q = append(s.q, moveCmd{x: r.Min.X, y: r.Min.Y})
				s.q = append(s.q, clearRight)
			}
		default:
			// TODO: Support Origin Mode [ansi.OriginMode]
			s.q = append(s.q, scrollUpCmd{region: r, n: r.Height()})
		}
	} else {
		// TODO: Support Set Left Right Margins [ansi.DECSLRM].
		for y := r.Min.Y; y < r.Max.Y; y++ {
			for x := r.Min.X; x < r.Max.X; x++ {
				var cell *Cell
				if !s.cur.Style.Empty() {
					bc := BlankCell
					bc.Style = s.cur.Style
					cell = &bc
				}
				s.Draw(x, y, cell)
			}
		}
	}
}

// Draw implements Window.
func (s *Screen) Draw(x int, y int, cell *Cell) (v bool) {
	if x >= s.width-1 && cell != nil && cell.Width > 1 {
		// Too wide for the screen, we write a blank cell.
		bc := BlankCell
		bc.Style = cell.Style
		bc.Link = cell.Link
		cell = &bc
	}

	v = s.buf.SetCell(x, y, cell)
	if v {
		pos := Pos(x, y)
		if _, ok := s.dirty[pos]; !ok {
			s.dirty[pos] = struct{}{}
			s.q = append(s.q, posCmd{x: x, y: y})
		}
	}
	return
}

// Render implements Window.
func (s *Screen) Render() {
	var b bytes.Buffer
	for _, c := range s.q {
		b.WriteString(c.sequence(s))
	}

	log.Printf("Render: %q", b.String())
	s.w.Write(b.Bytes()) //nolint:errcheck

	// Copy the current buffer to the previous buffer
	for y := 0; y < len(s.buf.Lines); y++ {
		copy(s.bufs[1].Lines[y], s.buf.Lines[y])
	}

	// Update the cursor
	s.curs[1] = *s.cur
}

// MoveCursor implements Window.
func (s *Screen) MoveCursor(x int, y int) {
	s.q = append(s.q, moveCmd{x: x, y: y})
}

// Resize resizes the screen.
func (s *Screen) Resize(width, height int) {
	s.width, s.height = width, height
	s.bufs[0].Resize(width, height)
	s.bufs[1].Resize(width, height)
}

// NewWindow creates a new sub-window.
func (s *Screen) NewWindow(x, y, width, height int) Window {
	return &window{parent: s, x: x, y: y, width: width, height: height}
}

// window represents a terminal window.
type window struct {
	parent        *Screen // the parent screen (nil if the window is a screen)
	x, y          int     // the window's position relative to the parent
	width, height int
}

// Cell implements Window.
func (w *window) Cell(x int, y int) *Cell {
	if x >= w.width || y >= w.height || x < 0 || y < 0 {
		return nil
	}
	return w.parent.Cell(w.x+x, w.y+y)
}

// Fill implements Window.
func (w *window) Fill(cell *Cell) {
	w.parent.FillInRect(cell, w.Bounds())
}

// Clear implements Window.
func (w *window) Clear() {
	w.parent.ClearInRect(w.Bounds())
}

// Draw implements Window.
func (w *window) Draw(x int, y int, cell *Cell) bool {
	if x >= w.width || y >= w.height || x < 0 || y < 0 {
		return false
	}

	// Cell is out of bounds, we write a blank cell.
	if cell != nil && x+cell.Width >= w.width {
		if x+1 >= w.width {
			return false
		}

		cell = nil
	}

	return w.parent.Draw(w.x+x, w.y+y, cell)
}

// MoveCursor implements Window.
func (w *window) MoveCursor(x int, y int) {
	x = clamp(x, 0, w.width-1)
	y = clamp(y, 0, w.height-1)
	w.parent.MoveCursor(w.x+x, w.y+y)
}

// Bounds returns the window's bounds.
func (w *window) Bounds() Rectangle {
	return Rect(w.x, w.y, w.width, w.height)
}

// Resize resizes the window.
func (w *window) Resize(width, height int) {
	w.width, w.height = width, height
}

// Window represents a terminal window that can be used to draw to the screen.
type Window interface {
	// Clear clears the window.
	Clear()
	// Fill fills the window with the specified cell.
	Fill(cell *Cell)
	// Draw moves the cursor to the specified position, and draws the cell
	// moving the cursor by the cell's width. Use nil to clear a cell.
	Draw(x, y int, cell *Cell) bool
	// Cell returns the cell at the specified position. If the cell is out of
	// bounds, it returns nil.
	Cell(x, y int) *Cell
	// MoveCursor moves the cursor to the specified position.
	MoveCursor(x, y int)
	// Bounds returns the window's bounds.
	Bounds() Rectangle
	// Resize resizes the window.
	Resize(width, height int)
}
