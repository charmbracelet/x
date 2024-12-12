package cellbuf

import (
	"bytes"
	"errors"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/term"
)

// ErrInvalidDimensions is returned when the dimensions of a window are invalid
// for the operation.
var ErrInvalidDimensions = errors.New("invalid dimensions")

// notLocal returns whether the coordinates are not considered local movement
// using the defined thresholds.
// This takes the number of columns, and the coordinates of the current and
// target positions.
func notLocal(cols, fx, fy, tx, ty int) bool {
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
		var yseq string
		if s.xtermLike && !s.opts.RelativeCursor {
			yseq = ansi.VerticalPositionAbsolute(ty + 1)
		}

		// OPTIM: Use [ansi.LF] and [ansi.ReverseIndex] as optimizations.

		if ty > fy {
			n := ty - fy
			if cud := ansi.CursorDown(n); yseq == "" || len(cud) < len(yseq) {
				yseq = cud
			}
			shouldScroll := !s.opts.AltScreen && fy+n >= s.newbuf.Height()
			if lf := strings.Repeat("\n", n); yseq == "" || shouldScroll || fy+n < s.newbuf.Height() && len(lf) < len(yseq) {
				// TODO: Ensure we're not unintentionally scrolling the screen down.
				yseq = lf
			}
		} else if ty < fy {
			n := fy - ty
			if cuu := ansi.CursorUp(n); yseq == "" || len(cuu) < len(yseq) {
				yseq = cuu
			}
			if yseq == "" || n == 1 && fy-1 > 0 {
				// TODO: Ensure we're not unintentionally scrolling the screen up.
				yseq = ansi.ReverseIndex
			}
		}

		seq += yseq
	}

	if tx != fx {
		var xseq string
		if s.xtermLike && !s.opts.RelativeCursor {
			xseq = ansi.HorizontalPositionAbsolute(tx + 1)
		}

		if tx > fx {
			n := tx - fx
			if cuf := ansi.CursorForward(n); xseq == "" || len(cuf) < len(xseq) {
				xseq = cuf
			}

			// OPTIM: Use [ansi.HT] and hard tabs as an optimization.

			// If we have no attribute and style changes, overwrite is cheaper.
			var ovw string
			if overwrite && ty >= 0 {
				for i := 0; i < n; i++ {
					cell := s.newbuf.Cell(fx+i, ty)
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
					cell := s.newbuf.Cell(fx+i, ty)
					if cell != nil {
						ovw += cell.String()
						i += cell.Width - 1
					} else {
						ovw += " "
					}
				}
			}

			if overwrite && (xseq == "" || len(ovw) < len(xseq)) {
				xseq = ovw
			}
		} else if tx < fx {
			n := fx - tx
			if cub := ansi.CursorBackward(n); xseq == "" || len(cub) < len(xseq) {
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

	if !s.opts.RelativeCursor {
		// Method #0: Use [ansi.CUP] if the distance is long.
		seq = ansi.CursorPosition(x+1, y+1)
		if fx == -1 || fy == -1 || notLocal(s.newbuf.Width(), fx, fy, x, y) {
			return
		}
	}

	// Method #1: Use local movement sequences.
	nseq := relativeCursorMove(s, fx, fy, x, y, overwrite)
	if seq == "" || len(nseq) < len(seq) {
		seq = nseq
	}

	// Method #2: Use [ansi.CR] and local movement sequences.
	nseq = "\r" + relativeCursorMove(s, 0, fy, x, y, overwrite)
	if seq == "" || len(nseq) < len(seq) {
		seq = nseq
	}

	if !s.opts.RelativeCursor {
		// Method #3: Use [ansi.CursorHomePosition] and local movement sequences.
		nseq = ansi.CursorHomePosition + relativeCursorMove(s, 0, 0, x, y, overwrite)
		if seq == "" || len(nseq) < len(seq) {
			seq = nseq
		}
	}

	return
}

// moveCursor moves the cursor to the specified position.
func (s *Screen) moveCursor(w io.Writer, x, y int, overwrite bool) {
	io.WriteString(w, moveCursor(s, x, y, overwrite)) //nolint:errcheck
	s.cur.X, s.cur.Y = x, y
}

func (s *Screen) move(w io.Writer, x, y int) {
	width, height := s.newbuf.Width(), s.newbuf.Height()
	if s.cur.X == x && s.cur.Y == y || width <= 0 || height <= 0 {
		return
	}

	if x >= width {
		// Handle autowrap
		y += (x / width)
		x %= width
	}

	// Disable styles if there's any
	var pen Style
	if !s.cur.Style.Empty() {
		pen = s.cur.Style
		io.WriteString(w, ansi.ResetStyle) //nolint:errcheck
	}

	if s.cur.X >= width {
		l := (s.cur.X + 1) / width

		s.cur.Y += l
		if s.cur.Y >= height {
			l -= s.cur.Y - height - 1
		}

		if l > 0 {
			s.cur.X = 0
			io.WriteString(w, "\r"+strings.Repeat("\n", l)) //nolint:errcheck
		}
	}

	if s.cur.Y > height-1 {
		s.cur.Y = height - 1
	}
	if y > height-1 {
		y = height - 1
	}

	// We set the new cursor in [Screen.moveCursor].
	s.moveCursor(w, x, y, true) // Overwrite cells if possible

	if !pen.Empty() {
		io.WriteString(w, pen.Sequence()) //nolint:errcheck
	}
}

// Cursor represents a terminal Cursor.
type Cursor struct {
	Style Style
	Link  Link
	Position
}

// ScreenOptions are options for the screen.
type ScreenOptions struct {
	// Term is the terminal type to use when writing to the screen. When empty,
	// `$TERM` is used from [os.Getenv].
	Term string
	// Width is the desired width of the screen. When 0, the width is
	// automatically determined using the terminal size.
	Width int
	// Height is the desired height of the screen. When 0, the height is
	// automatically determined using the terminal size.
	Height int
	// Profile is the color profile to use when writing to the screen.
	Profile colorprofile.Profile
	// RelativeCursor is whether to use relative cursor movements. This is
	// useful when alt-screen is not used or when using inline mode.
	RelativeCursor bool
	// AltScreen is whether to use the alternate screen buffer.
	AltScreen bool
	// ShowCursor is whether to show the cursor.
	ShowCursor bool
}

// Screen represents the terminal screen.
type Screen struct {
	w             io.Writer
	curbuf        *Buffer // the current buffer
	newbuf        *Buffer // the new buffer
	touch         map[int][2]int
	queueAbove    []string // the queue of strings to write above the screen
	cur, saved    Cursor   // the current and saved cursors
	opts          ScreenOptions
	pos           Position // the position of the cursor after the last render
	mu            sync.Mutex
	altScreenMode bool // whether alternate screen mode is enabled
	cursorHidden  bool // whether text cursor mode is enabled
	clear         bool // whether to force clear the screen
	xtermLike     bool // whether to use xterm-like optimizations, otherwise, it uses vt100 only
}

var _ Window = &Screen{}

// SetColorProfile sets the color profile to use when writing to the screen.
func (s *Screen) SetColorProfile(p colorprofile.Profile) {
	s.opts.Profile = p
}

// SetRelativeCursor sets whether to use relative cursor movements.
func (s *Screen) SetRelativeCursor(v bool) {
	s.opts.RelativeCursor = v
}

// EnterAltScreen enters the alternate screen buffer.
func (s *Screen) EnterAltScreen() {
	s.opts.AltScreen = true
	s.clear = true
	s.saved = s.cur
}

// ExitAltScreen exits the alternate screen buffer.
func (s *Screen) ExitAltScreen() {
	s.opts.AltScreen = false
	s.clear = true
	s.cur = s.saved
}

// ShowCursor shows the cursor.
func (s *Screen) ShowCursor() {
	s.opts.ShowCursor = true
}

// HideCursor hides the cursor.
func (s *Screen) HideCursor() {
	s.opts.ShowCursor = false
}

// Bounds implements Window.
func (s *Screen) Bounds() Rectangle {
	// Always return the new buffer bounds.
	return s.newbuf.Bounds()
}

// Cell implements Window.
func (s *Screen) Cell(x int, y int) *Cell {
	return s.newbuf.Cell(x, y)
}

// Clear implements Window.
func (s *Screen) Clear() bool {
	s.clear = true
	return s.ClearRect(s.newbuf.Bounds())
}

// ClearRect implements Window.
func (s *Screen) ClearRect(r Rectangle) bool {
	s.newbuf.ClearRect(r)
	s.mu.Lock()
	for i := r.Min.Y; i < r.Max.Y; i++ {
		s.touch[i] = [2]int{r.Min.X, r.Width() - 1}
	}
	s.mu.Unlock()
	return true
}

// Draw implements Window.
func (s *Screen) Draw(x int, y int, cell *Cell) (v bool) {
	cellWidth := 1
	if cell != nil {
		cellWidth = cell.Width
	}

	s.mu.Lock()
	chg := s.touch[y]
	chg[0] = min(chg[0], x)
	chg[1] = max(chg[1], x+cellWidth)
	s.touch[y] = chg
	s.mu.Unlock()

	return s.newbuf.Draw(x, y, cell)
}

// Fill implements Window.
func (s *Screen) Fill(cell *Cell) bool {
	return s.FillRect(cell, s.newbuf.Bounds())
}

// FillRect implements Window.
func (s *Screen) FillRect(cell *Cell, r Rectangle) bool {
	s.newbuf.FillRect(cell, r)
	s.mu.Lock()
	for i := r.Min.Y; i < r.Max.Y; i++ {
		s.touch[i] = [2]int{r.Min.X, r.Width() - 1}
	}
	s.mu.Unlock()
	return true
}

// isXtermLike returns whether the terminal is xterm-like. This means that the
// terminal supports ECMA-48 and ANSI X3.64 escape sequences.
// TODO: Should this be a lookup table into each $TERM terminfo database? Like
// we could keep a map of ANSI escape sequence to terminfo capability name and
// check if the database supports the escape sequence. Instead of keeping a
// list of terminal names here.
func isXtermLike(termtype string) (v bool) {
	parts := strings.Split(termtype, "-")
	if len(parts) == 0 {
		return
	}

	switch parts[0] {
	case
		"alacritty",
		"contour",
		"foot",
		"ghostty",
		"kitty",
		"linux",
		"rio",
		"screen",
		"st",
		"tmux",
		"wezterm",
		"xterm":
		v = true
	}

	return
}

// NewScreen creates a new Screen.
func NewScreen(w io.Writer, opts *ScreenOptions) (s *Screen) {
	s = new(Screen)
	s.w = w
	if opts != nil {
		s.opts = *opts
	}

	if s.opts.Term == "" {
		s.opts.Term = os.Getenv("TERM")
	}

	width, height := s.opts.Width, s.opts.Height
	if width <= 0 || height <= 0 {
		if f, ok := w.(term.File); ok {
			width, height, _ = term.GetSize(f.Fd())
		}
	}

	s.xtermLike = isXtermLike(s.opts.Term)
	s.curbuf = NewBuffer(width, height)
	s.newbuf = NewBuffer(width, height)
	s.reset()

	return
}

// cellEqual returns whether the two cells are equal. A nil cell is considered
// a [BlankCell].
func cellEqual(a, b *Cell) bool {
	if a == nil {
		a = &BlankCell
	}
	if b == nil {
		b = &BlankCell
	}
	return a.Equal(b)
}

// putCell draws a cell at the current cursor position.
func (s *Screen) putCell(w io.Writer, cell *Cell) {
	if cell != nil && cell.Empty() {
		return
	}

	if cell == nil {
		cell = s.clearBlank()
	}

	s.updatePen(w, cell)
	io.WriteString(w, cell.String()) //nolint:errcheck
	s.cur.X += cell.Width

	if s.cur.X >= s.newbuf.Width() {
		s.cur.X = s.newbuf.Width() - 1
	}
}

// updatePen updates the cursor pen styles.
func (s *Screen) updatePen(w io.Writer, cell *Cell) {
	if cell == nil {
		cell = &BlankCell
	}

	style := cell.Style
	link := cell.Link
	if s.opts.Profile != 0 {
		// Downsample colors to the given color profile.
		style = ConvertStyle(style, s.opts.Profile)
		link = ConvertLink(link, s.opts.Profile)
	}

	if !style.Equal(s.cur.Style) {
		seq := style.DiffSequence(s.cur.Style)
		if style.Empty() && len(seq) > len(ansi.ResetStyle) {
			seq = ansi.ResetStyle
		}
		io.WriteString(w, seq) //nolint:errcheck
		s.cur.Style = style
	}
	if !link.Equal(s.cur.Link) {
		io.WriteString(w, ansi.SetHyperlink(link.URL, link.URLID)) //nolint:errcheck
		s.cur.Link = link
	}
}

// emitRange emits a range of cells to the buffer. It it equivalent to calling
// [Screen.putCell] for each cell in the range. This is optimized to use
// [ansi.ECH] and [ansi.REP].
// Returns whether the cursor is at the end of interval or somewhere in the
// middle.
func (s *Screen) emitRange(w io.Writer, line Line, n int) (eoi bool) {
	for n > 0 {
		var count int
		for n > 1 && !cellEqual(line.At(0), line.At(1)) {
			s.putCell(w, line.At(0))
			line = line[1:]
			n--
		}

		cell0 := line[0]
		if n == 1 {
			s.putCell(w, cell0)
			return false
		}

		count = 2
		for count < n && cellEqual(line.At(count), cell0) {
			count++
		}

		ech := ansi.EraseCharacter(count)
		cup := ansi.CursorPosition(s.cur.X+count, s.cur.Y)
		rep := ansi.RepeatPreviousCharacter(count)
		if s.xtermLike && count > len(ech)+len(cup) && cell0 != nil && cell0.Clear() {
			s.updatePen(w, cell0)
			io.WriteString(w, ech) //nolint:errcheck

			// If this is the last cell, we don't need to move the cursor.
			if count < n {
				s.move(w, s.cur.X+count, s.cur.Y)
			} else {
				return true // cursor in the middle
			}
		} else if s.xtermLike && count > len(rep) &&
			(cell0 == nil || (len(cell0.Comb) == 0 && cell0.Rune < 256)) {
			// We only support ASCII characters. Most terminals will handle
			// non-ASCII characters correctly, but some might not, ahem xterm.
			//
			// NOTE: [ansi.REP] only repeats the last rune and won't work
			// if the last cell contains multiple runes.

			wrapPossible := s.cur.X+count >= s.newbuf.Width()
			repCount := count
			if wrapPossible {
				repCount--
			}

			s.updatePen(w, cell0)
			s.putCell(w, cell0)
			repCount-- // cell0 is a single width cell ASCII character

			io.WriteString(w, ansi.RepeatPreviousCharacter(repCount)) //nolint:errcheck
			s.cur.X += repCount
			if wrapPossible {
				s.putCell(w, cell0)
			}
		} else {
			for i := 0; i < count; i++ {
				s.putCell(w, line.At(i))
			}
		}

		line = line[clamp(count, 0, len(line)):]
		n -= count
	}

	return
}

// putRange puts a range of cells from the old line to the new line.
// Returns whether the cursor is at the end of interval or somewhere in the
// middle.
func (s *Screen) putRange(w io.Writer, oldLine, newLine Line, y, start, end int) (eoi bool) {
	inline := min(len(ansi.CursorPosition(start+1, y+1)),
		min(len(ansi.HorizontalPositionAbsolute(start+1)),
			len(ansi.CursorForward(start+1))))
	if (end - start + 1) > inline {
		var j, same int
		for j, same = start, 0; j <= end; j++ {
			oldCell, newCell := oldLine.At(j), newLine.At(j)
			if same == 0 && oldCell != nil && oldCell.Empty() {
				continue
			}
			if cellEqual(oldCell, newCell) {
				same++
			} else {
				if same > end-start {
					s.emitRange(w, newLine[start:], j-same-start)
					s.move(w, y, j)
					start = j
				}
				same = 0
			}
		}

		i := s.emitRange(w, newLine[start:], j-same-start)

		// Always return 1 for the next [Screen.move] after a [Screen.putRange] if
		// we found identical characters at end of interval.
		if same == 0 {
			return i
		}
		return true
	}

	return s.emitRange(w, newLine[start:], end-start+1)
}

// clearToEnd clears the screen from the current cursor position to the end of
// line.
func (s *Screen) clearToEnd(w io.Writer, blank *Cell, force bool) {
	if s.cur.Y >= 0 {
		curline := s.curbuf.Line(s.cur.Y)
		for j := s.cur.X; j < s.curbuf.Width(); j++ {
			if j >= 0 {
				c := curline.At(j)
				if !cellEqual(c, blank) {
					curline.Set(j, blank)
					force = true
				}
			}
		}
	}

	if force {
		s.updatePen(w, blank)
		count := s.newbuf.Width() - s.cur.X
		eraseRight := ansi.EraseLineRight
		if len(eraseRight) <= count {
			io.WriteString(w, eraseRight) //nolint:errcheck
		} else {
			for i := 0; i < count; i++ {
				s.putCell(w, blank)
			}
		}
	}
}

// clearBlank returns a blank cell based on the current cursor background color.
func (s *Screen) clearBlank() *Cell {
	c := BlankCell
	if !s.cur.Style.Empty() || !s.cur.Link.Empty() {
		c.Style = s.cur.Style
		c.Link = s.cur.Link
	}
	return &c
}

// insertCells inserts the count cells pointed by the given line at the current
// cursor position.
func (s *Screen) insertCells(w io.Writer, line Line, count int) {
	if s.xtermLike {
		// Use [ansi.ICH] as an optimization.
		io.WriteString(w, ansi.InsertCharacter(count)) //nolint:errcheck
	} else {
		// Otherwise, use [ansi.IRM] mode.
		io.WriteString(w, ansi.SetInsertReplaceMode) //nolint:errcheck
	}

	for i := 0; count > 0; i++ {
		s.putCell(w, line[i])
		count--
	}

	if !s.xtermLike {
		io.WriteString(w, ansi.ResetInsertReplaceMode) //nolint:errcheck
	}
}

// transformLine transforms the given line in the current window to the
// corresponding line in the new window. It uses [ansi.ICH] and [ansi.DCH] to
// insert or delete characters.
func (s *Screen) transformLine(w io.Writer, y int) {
	var firstCell, oLastCell, nLastCell int // first, old last, new last index
	oldLine := s.curbuf.Line(y)
	newLine := s.newbuf.Line(y)

	// Find the first changed cell in the line
	var lineChanged bool
	for i := 0; i < s.newbuf.Width(); i++ {
		if !cellEqual(newLine.At(i), oldLine.At(i)) {
			lineChanged = true
			break
		}
	}

	const ceolStandoutGlitch = false
	if ceolStandoutGlitch && lineChanged {
		s.move(w, 0, y)
		s.clearToEnd(w, nil, false)
		s.putRange(w, oldLine, newLine, y, 0, s.newbuf.Width()-1)
	} else {
		blank := newLine.At(0)

		// It might be cheaper to clear leading spaces with [ansi.EL] 1 i.e.
		// [ansi.EraseLineLeft].
		if blank == nil || blank.Clear() {
			var oFirstCell, nFirstCell int
			for oFirstCell = 0; oFirstCell < s.curbuf.Width(); oFirstCell++ {
				if !cellEqual(oldLine.At(oFirstCell), blank) {
					break
				}
			}
			for nFirstCell = 0; nFirstCell < s.newbuf.Width(); nFirstCell++ {
				if !cellEqual(newLine.At(nFirstCell), blank) {
					break
				}
			}

			if nFirstCell == oFirstCell {
				firstCell = nFirstCell

				// Find the first differing cell
				for firstCell < s.newbuf.Width() &&
					cellEqual(oldLine.At(firstCell), newLine.At(firstCell)) {
					firstCell++
				}
			} else if oFirstCell > nFirstCell {
				firstCell = nFirstCell
			} else /* if oFirstCell < nFirstCell */ {
				firstCell = oFirstCell
				el1Cost := len(ansi.EraseLineLeft)
				if el1Cost < nFirstCell-oFirstCell {
					if nFirstCell >= s.newbuf.Width() {
						s.move(w, 0, y)
						s.updatePen(w, blank)
						io.WriteString(w, ansi.EraseLineRight) //nolint:errcheck
					} else {
						s.move(w, nFirstCell-1, y)
						s.updatePen(w, blank)
						io.WriteString(w, ansi.EraseLineLeft) //nolint:errcheck
					}

					for firstCell < nFirstCell {
						oldLine.Set(firstCell, blank)
						firstCell++
					}
				}
			}
		} else {
			// Find the first differing cell
			for firstCell < s.newbuf.Width() && cellEqual(newLine.At(firstCell), oldLine.At(firstCell)) {
				firstCell++
			}
		}

		// If we didn't find one, we're done
		if firstCell >= s.newbuf.Width() {
			return
		}

		blank = newLine.At(s.newbuf.Width() - 1)
		if blank != nil && !blank.Clear() {
			// Find the last differing cell
			nLastCell = s.newbuf.Width() - 1
			for nLastCell > firstCell && cellEqual(newLine.At(nLastCell), oldLine.At(nLastCell)) {
				nLastCell--
			}

			if nLastCell >= firstCell {
				s.move(w, firstCell, y)
				s.putRange(w, oldLine, newLine, y, firstCell, nLastCell)
				copy(oldLine[firstCell:], newLine[firstCell:])
			}

			return
		}

		// Find last non-blank cell in the old line.
		oLastCell = s.curbuf.Width() - 1
		for oLastCell > firstCell && cellEqual(oldLine.At(oLastCell), blank) {
			oLastCell--
		}

		// Find last non-blank cell in the new line.
		nLastCell = s.newbuf.Width() - 1
		for nLastCell > firstCell && cellEqual(newLine.At(nLastCell), blank) {
			nLastCell--
		}

		el0Cost := len(ansi.EraseLineRight)
		if nLastCell == firstCell && el0Cost < oLastCell-nLastCell {
			s.move(w, firstCell, y)
			if !cellEqual(newLine.At(firstCell), blank) {
				s.putCell(w, newLine.At(firstCell))
			}
			s.clearToEnd(w, blank, false)
		} else if nLastCell != oLastCell &&
			!cellEqual(newLine.At(nLastCell), oldLine.At(oLastCell)) {
			s.move(w, firstCell, y)
			if oLastCell-nLastCell > el0Cost {
				if s.putRange(w, oldLine, newLine, y, firstCell, nLastCell) {
					s.move(w, nLastCell, y)
				}
				s.clearToEnd(w, blank, false)
			} else {
				n := max(nLastCell, oLastCell)
				s.putRange(w, oldLine, newLine, y, firstCell, n)
			}
		} else {
			nLastNonBlank := nLastCell
			oLastNonBlank := oLastCell

			// Find the last cells that really differ.
			// Can be -1 if no cells differ.
			for cellEqual(newLine.At(nLastCell), oldLine.At(oLastCell)) {
				if !cellEqual(newLine.At(nLastCell-1), oldLine.At(oLastCell-1)) {
					break
				}
				nLastCell--
				oLastCell--
				if nLastCell == -1 || oLastCell == -1 {
					break
				}
			}

			n := min(oLastCell, nLastCell)
			if n >= firstCell {
				s.move(w, firstCell, y)
				s.putRange(w, oldLine, newLine, y, firstCell, n)
			}

			if oLastCell < nLastCell {
				m := max(nLastNonBlank, oLastNonBlank)
				if n != 0 {
					for n > 0 {
						wide := newLine.At(n + 1)
						if wide == nil || !wide.Empty() {
							break
						}
						n--
						oLastCell--
					}
				} else if n >= firstCell && newLine.At(n) != nil && newLine.At(n).Width > 1 {
					next := newLine.At(n + 1)
					for next != nil && next.Empty() {
						n++
						oLastCell++
					}
				}

				s.move(w, n+1, y)
				ichCost := 3 + nLastCell - oLastCell
				if s.xtermLike && (nLastCell < nLastNonBlank || ichCost > (m-n)) {
					s.putRange(w, oldLine, newLine, y, n+1, m)
				} else {
					s.insertCells(w, newLine[n+1:], nLastCell-oLastCell)
				}
			} else if oLastCell > nLastCell {
				s.move(w, n+1, y)
				dchCost := 3 + oLastCell - nLastCell
				if dchCost > len(ansi.EraseLineRight)+nLastNonBlank-(n+1) {
					if s.putRange(w, oldLine, newLine, y, n+1, nLastNonBlank) {
						s.move(w, nLastNonBlank+1, y)
					}
					s.clearToEnd(w, blank, false)
				} else {
					s.updatePen(w, blank)
					s.deleteCells(w, oLastCell-nLastCell)
				}
			}
		}
	}

	// Update the old line with the new line
	if s.newbuf.Width() > firstCell && len(oldLine) != 0 {
		copy(oldLine[firstCell:], newLine[firstCell:])
	}
}

// deleteCells deletes the count cells at the current cursor position and moves
// the rest of the line to the left. This is equivalent to [ansi.DCH].
func (s *Screen) deleteCells(w io.Writer, count int) {
	// [ansi.DCH] will shift in cells from the right margin so we need to
	// ensure that they are the right style.
	io.WriteString(w, ansi.DeleteCharacter(count)) //nolint:errcheck
}

// clearToBottom clears the screen from the current cursor position to the end
// of the screen.
func (s *Screen) clearToBottom(w io.Writer, blank *Cell) {
	row, col := s.cur.Y, s.cur.X
	if row < 0 {
		row = 0
	}
	if col < 0 {
		col = 0
	}

	s.updatePen(w, blank)
	io.WriteString(w, ansi.EraseScreenBelow) //nolint:errcheck
	s.curbuf.ClearRect(Rect(col, row, s.curbuf.Width(), row+1))
	s.curbuf.ClearRect(Rect(0, row+1, s.curbuf.Width(), s.curbuf.Height()))
}

// clearBottom tests if clearing the end of the screen would satisfy part of
// the screen update. Scan backwards through lines in the screen checking if
// each is blank and one or more are changed.
// It returns the top line.
func (s *Screen) clearBottom(w io.Writer, total int, force bool) (top int) {
	top = total
	if total <= 0 {
		return
	}

	last := min(s.curbuf.Width(), s.newbuf.Width())
	blank := s.clearBlank()
	canClearWithBlank := blank == nil || blank.Clear()

	if canClearWithBlank || force {
		var row int
		for row = total - 1; row >= 0; row-- {
			var col int
			var ok bool
			for col, ok = 0, true; ok && col < last; col++ {
				ok = cellEqual(s.newbuf.Cell(col, row), blank)
			}
			if !ok {
				break
			}

			for col = 0; ok && col < last; col++ {
				ok = cellEqual(s.curbuf.Cell(col, row), blank)
			}
			if !ok {
				top = row
			}
		}

		if force || top < total {
			s.moveCursor(w, 0, top, false)
			s.clearToBottom(w, blank)
			if !s.opts.AltScreen {
				// Move to the last line of the screen
				s.moveCursor(w, 0, s.newbuf.Height()-1, false)
			}
			// TODO: Line hashing
		}
	}

	return
}

// clearScreen clears the screen and put cursor at home.
func (s *Screen) clearScreen(w io.Writer, blank *Cell) {
	s.updatePen(w, blank)
	io.WriteString(w, ansi.CursorHomePosition) //nolint:errcheck
	io.WriteString(w, ansi.EraseEntireScreen)  //nolint:errcheck
	s.cur.X, s.cur.Y = 0, 0
	s.curbuf.Fill(blank)
}

// clearBelow clears everything below the screen.
func (s *Screen) clearBelow(w io.Writer, blank *Cell, row int) {
	s.updatePen(w, blank)
	s.moveCursor(w, 0, row, false)
	s.clearToBottom(w, blank)
	s.cur.X, s.cur.Y = 0, row
	s.curbuf.FillRect(blank, Rect(0, row, s.curbuf.Width(), s.curbuf.Height()))
}

// clearUpdate forces a screen redraw.
func (s *Screen) clearUpdate(w io.Writer, partial bool) {
	blank := s.clearBlank()
	var nonEmpty int
	if s.opts.AltScreen {
		nonEmpty = min(s.curbuf.Height(), s.newbuf.Height())
		s.clearScreen(w, blank)
	} else {
		nonEmpty = s.newbuf.Height()
		s.clearBelow(w, blank, 0)
	}
	nonEmpty = s.clearBottom(w, nonEmpty, partial)
	for i := 0; i < nonEmpty; i++ {
		s.transformLine(w, i)
	}
}

// Render implements Window.
func (s *Screen) Render() {
	s.mu.Lock()
	b := new(bytes.Buffer)
	s.render(b)
	// Write the buffer
	if b.Len() > 0 {
		s.w.Write(b.Bytes()) //nolint:errcheck
	}
	s.mu.Unlock()
}

func (s *Screen) render(b *bytes.Buffer) {
	// Do we need alt-screen mode?
	if s.opts.AltScreen != s.altScreenMode {
		if s.opts.AltScreen {
			b.WriteString(ansi.SetAltScreenSaveCursorMode)
		} else {
			b.WriteString(ansi.ResetAltScreenSaveCursorMode)
		}
		s.altScreenMode = s.opts.AltScreen
	}

	// Do we need text cursor mode?
	if !s.opts.ShowCursor != s.cursorHidden {
		s.cursorHidden = !s.opts.ShowCursor
		if s.cursorHidden {
			b.WriteString(ansi.HideCursor)
		}
	}

	// Do we have queued strings to write above the screen?
	if len(s.queueAbove) > 0 {
		// TODO: Use scrolling region if available.
		// TODO: Use [Screen.Write] [io.Writer] interface.

		// We need to scroll the screen up by the number of lines in the queue.
		// We can't use [ansi.SU] because we want the cursor to move down until
		// it reaches the bottom of the screen.
		s.moveCursor(b, 0, s.newbuf.Height()-1, false)
		b.WriteString(strings.Repeat("\n", len(s.queueAbove)))
		s.cur.Y += len(s.queueAbove)
		// Now go to the top of the screen, insert new lines, and write the
		// queued strings.
		s.moveCursor(b, 0, 0, false)
		b.WriteString(ansi.InsertLine(len(s.queueAbove)))
		for _, line := range s.queueAbove {
			b.WriteString(line + "\r\n")
		}

		// Clear the queue
		s.queueAbove = s.queueAbove[:0]
	}

	var nonEmpty int

	// Force clear?
	// We only do partial clear if the screen is not in alternate screen mode
	partialClear := s.curbuf.Width() == s.newbuf.Width() &&
		s.curbuf.Height() > s.newbuf.Height()

	if s.clear {
		s.clearUpdate(b, partialClear)
		s.clear = false
	} else if len(s.touch) > 0 {
		var changedLines int
		var i int

		if s.opts.AltScreen {
			nonEmpty = min(s.curbuf.Height(), s.newbuf.Height())
		} else {
			nonEmpty = s.newbuf.Height()
		}

		nonEmpty = s.clearBottom(b, nonEmpty, partialClear)
		for i = 0; i < nonEmpty; i++ {
			_, wasTouched := s.touch[i]
			if wasTouched {
				s.transformLine(b, i)
				changedLines++
			}
		}

		// Mark changed lines
		if i <= s.newbuf.Height() {
			delete(s.touch, i)
		}
	}

	// Sync windows and screen
	for i := 0; i <= s.newbuf.Height(); i++ {
		delete(s.touch, i)
	}

	if s.curbuf.Width() != s.newbuf.Width() || s.curbuf.Height() != s.newbuf.Height() {
		// Resize the old buffer to match the new buffer.
		_, oldh := s.curbuf.Width(), s.curbuf.Height()
		s.curbuf.Resize(s.newbuf.Width(), s.newbuf.Height())
		// Sync new lines to old lines
		for i := oldh - 1; i < s.newbuf.Height(); i++ {
			copy(s.curbuf.Line(i), s.newbuf.Line(i))
		}
	}

	s.updatePen(b, nil) // nil indicates a blank cell with no styles

	// Move the cursor to the specified position.
	if s.pos != undefinedPos {
		s.move(b, s.pos.X, s.pos.Y)
		s.pos = undefinedPos
	}

	if b.Len() > 0 {
		// Is the cursor visible? If so, disable it while rendering.
		if s.opts.ShowCursor && !s.cursorHidden {
			nb := new(bytes.Buffer)
			nb.WriteString(ansi.HideCursor)
			nb.Write(b.Bytes())
			nb.WriteString(ansi.ShowCursor)
			*b = *nb
		}
	}
}

// undefinedPos is the position used when the cursor position is undefined and
// in its initial state.
var undefinedPos = Pos(-1, -1)

// Close writes the final screen update and resets the screen.
func (s *Screen) Close() (err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	b := new(bytes.Buffer)
	s.render(b)
	s.updatePen(b, nil)
	s.move(b, 0, s.newbuf.Height()-1)
	s.clearToEnd(b, nil, true)

	if s.altScreenMode {
		b.WriteString(ansi.ResetAltScreenSaveCursorMode)
		s.altScreenMode = false
	}

	if s.cursorHidden {
		b.WriteString(ansi.ShowCursor)
		s.cursorHidden = false
	}

	// Write the buffer
	_, err = s.w.Write(b.Bytes())
	if err != nil {
		return
	}

	s.reset()
	return
}

// reset resets the screen to its initial state.
func (s *Screen) reset() {
	s.cursorHidden = false
	s.altScreenMode = false
	if s.opts.RelativeCursor {
		s.cur = Cursor{}
	} else {
		s.cur = Cursor{Position: undefinedPos}
	}
	s.saved = s.cur
	s.touch = make(map[int][2]int)
	if s.curbuf != nil {
		s.curbuf.Clear()
	}
	if s.newbuf != nil {
		s.newbuf.Clear()
	}
}

// Resize resizes the screen.
func (s *Screen) Resize(width, height int) bool {
	oldw := s.newbuf.Width()
	oldh := s.newbuf.Height()

	if s.opts.AltScreen || width != oldw {
		// We only clear the whole screen if the width changes. Adding/removing
		// rows is handled by the [Screen.render] and [Screen.transformLine]
		// methods.
		s.clear = true
	}

	// Clear new columns and lines
	if width > oldh {
		s.ClearRect(Rect(oldw-2, 0, width-oldw, height))
	} else if width < oldw {
		s.ClearRect(Rect(width-1, 0, oldw-width, height))
	}

	if height > oldh {
		s.ClearRect(Rect(0, oldh-1, width, height-oldh))
	} else if height < oldh {
		s.ClearRect(Rect(0, height-1, width, oldh-height))
	}

	s.mu.Lock()
	s.newbuf.Resize(width, height)
	s.opts.Width, s.opts.Height = width, height
	s.mu.Unlock()

	return true
}

// MoveTo moves the cursor to the specified position.
func (s *Screen) MoveTo(x, y int) bool {
	pos := Pos(x, y)
	if !s.Bounds().Contains(pos) {
		return false
	}
	s.mu.Lock()
	s.pos = pos
	s.mu.Unlock()
	return true
}

// InsertAbove inserts string above the screen. The inserted string is not
// managed by the screen. This does nothing when alternate screen mode is
// enabled.
func (s *Screen) InsertAbove(str string) {
	if s.opts.AltScreen {
		return
	}
	s.mu.Lock()
	s.queueAbove = append(s.queueAbove, strings.Split(str, "\n")...)
	s.mu.Unlock()
}

// newWindow creates a new window.
func (s *Screen) newWindow(x, y, width, height int) (w *SubWindow, err error) {
	w = new(SubWindow)
	w.scr = s
	w.bounds = Rect(x, y, width, height)
	if x < 0 || y < 0 || width <= 0 || height <= 0 {
		return nil, ErrInvalidDimensions
	}

	scrw, scrh := s.Bounds().Width(), s.Bounds().Height()
	if x+width > scrw || y+height > scrh {
		return nil, ErrInvalidDimensions
	}

	return
}

// Window represents parts of the terminal screen.
type Window interface {
	Cell(x int, y int) *Cell
	Fill(cell *Cell) bool
	FillRect(cell *Cell, r Rectangle) bool
	Clear() bool
	ClearRect(r Rectangle) bool
	Draw(x int, y int, cell *Cell) (v bool)
	Bounds() Rectangle
	Resize(width, height int) bool
	MoveTo(x, y int) bool
}

// SubWindow represents a terminal SubWindow.
type SubWindow struct {
	scr    *Screen    // the screen where the window is attached
	par    *SubWindow // the parent screen (nil if the window is the primary window)
	bounds Rectangle  // the window's bounds
}

var _ Window = &SubWindow{}

// NewWindow creates a new sub-window.
func (s *Screen) NewWindow(x, y, width, height int) (*SubWindow, error) {
	return s.newWindow(x, y, width, height)
}

// NewWindow creates a new sub-window.
func (w *SubWindow) NewWindow(x, y, width, height int) (s *SubWindow, err error) {
	s, err = w.scr.newWindow(x, y, width, height)
	w.par = w
	return
}

// MoveTo moves the cursor to the specified position.
func (w *SubWindow) MoveTo(x, y int) bool {
	pos := Pos(x, y)
	if !w.Bounds().Contains(pos) {
		return false
	}

	x, y = w.bounds.Min.X+x, w.bounds.Min.Y+y
	return w.scr.MoveTo(x, y)
}

// Cell implements Window.
func (w *SubWindow) Cell(x int, y int) *Cell {
	if !Pos(x, y).In(w.Bounds().Rectangle) {
		return nil
	}
	bx, by := w.Bounds().Min.X, w.Bounds().Min.Y
	return w.scr.Cell(bx+x, by+y)
}

// Fill implements Window.
func (w *SubWindow) Fill(cell *Cell) bool {
	return w.FillRect(cell, w.Bounds())
}

// FillRect fills the cells in the specified rectangle with the specified
// cell.
func (w *SubWindow) FillRect(cell *Cell, r Rectangle) bool {
	if !r.In(w.Bounds().Rectangle) {
		return false
	}

	w.scr.FillRect(cell, r)
	return true
}

// Clear implements Window.
func (w *SubWindow) Clear() bool {
	return w.ClearRect(w.Bounds())
}

// ClearRect clears the cells in the specified rectangle based on the current
// cursor background color. Use [SetPen] to set the background color.
func (w *SubWindow) ClearRect(r Rectangle) bool {
	if !r.In(w.Bounds().Rectangle) {
		return false
	}

	w.scr.ClearRect(r)
	return true
}

// Draw implements Window.
func (w *SubWindow) Draw(x int, y int, cell *Cell) (v bool) {
	if !Pos(x, y).In(w.Bounds().Rectangle) {
		return
	}

	bx, by := w.Bounds().Min.X, w.Bounds().Min.Y
	return w.scr.newbuf.Draw(bx+x, by+y, cell)
}

// Bounds returns the window's bounds.
func (w *SubWindow) Bounds() Rectangle {
	return w.bounds
}

// Resize implements Window.
func (w *SubWindow) Resize(width int, height int) bool {
	if width <= 0 || height <= 0 {
		return false
	}

	if w.Bounds().Width() == width && w.Bounds().Height() == height {
		return true
	}

	x, y := w.bounds.Min.X, w.bounds.Min.Y
	scrw, scrh := w.scr.Bounds().Width(), w.scr.Bounds().Height()
	if x+width > scrw || y+height > scrh {
		return false
	}

	w.bounds = Rect(x, y, width, height)
	return true
}
