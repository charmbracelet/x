package cellbuf

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"
)

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
		yseq := ansi.VerticalPositionAbsolute(ty + 1)

		// OPTIM: Use [ansi.LF] and [ansi.ReverseIndex] as optimizations.

		if ty > fy {
			n := ty - fy
			if cud := ansi.CursorDown(n); len(cud) < len(yseq) {
				yseq = cud
			}
			if lf := strings.Repeat("\n", n); fy+n < s.height && len(lf) < len(yseq) {
				// TODO: Ensure we're not unintentionally scrolling the screen down.
				yseq = lf
			}
		} else if ty < fy {
			n := fy - ty
			if cuu := ansi.CursorUp(n); len(cuu) < len(yseq) {
				yseq = cuu
			}
			if n == 1 && fy+1 < s.height {
				// TODO: Ensure we're not unintentionally scrolling the screen up.
				yseq = ansi.ReverseIndex
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
					cell := s.newwin.Cell(fx+i, ty)
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
					cell := s.newwin.Cell(fx+i, ty)
					if cell != nil {
						xseq += cell.Content()
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

	// Method #0: Use [ansi.CUP] if the distance is long.
	seq = ansi.CursorPosition(x+1, y+1)
	if fx == -1 || fy == -1 || notLocal(s.width, fx, fy, x, y) {
		return
	}

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

	return
}

func (s *Screen) move(w *bytes.Buffer, x, y int) {
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
		w.WriteString(ansi.ResetStyle)
	}

	if s.cur.X >= s.width {
		l := (s.cur.X + 1) / s.width

		s.cur.Y += l
		if s.cur.Y >= s.height {
			l -= s.cur.Y - s.height - 1
		}

		if l > 0 {
			w.WriteByte(ansi.CR) // '\r'
			s.cur.X = 0
			w.WriteString(strings.Repeat("\n", l))
		}
	}

	if s.cur.Y > s.height-1 {
		s.cur.Y = s.height - 1
	}
	if y > s.height-1 {
		y = s.height - 1
	}

	w.WriteString(moveCursor(s, x, y, true)) // Overwrite cells if possible

	if !pen.Empty() {
		w.WriteString(pen.Sequence())
	}

	s.cur.X, s.cur.Y = x, y
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
	// LeaveCursor is whether to leave the cursor at the location after rendering.
	LeaveCursor bool
}

// Screen represents the terminal screen.
type Screen struct {
	w             io.Writer
	curwin        *Window // the previous window
	newwin        *Window // the current window
	opts          ScreenOptions
	cur           cursor // the current cursor
	mu            sync.Mutex
	lastChar      rune // the last character written to the screen
	width, height int  // the screen's width and height
}

// NewScreen creates a new Screen.
func NewScreen(w io.Writer, width, height int, opts *ScreenOptions) (s *Screen) {
	s = new(Screen)
	s.w = w
	if opts != nil {
		s.opts = *opts
	}

	s.width, s.height = width, height
	s.reset()

	return
}

// Bounds implements Window.
func (s *Screen) Bounds() Rectangle {
	return Rect(0, 0, s.width, s.height)
}

// Window returns the screen's window.
func (s *Screen) Window() *Window {
	return s.newwin
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

// cellRunes returns the runes of the cell content. A nil cell is considered a
// [BlankCell].
func cellRunes(c *Cell) []rune {
	if c == nil {
		return []rune{BlankCell.Rune}
	}
	return append([]rune{c.Rune}, c.Comb...)
}

// putCell draws a cell at the current cursor position.
func (s *Screen) putCell(w *bytes.Buffer, cell *Cell) {
	if cell != nil && cell.Width == 0 && cell.Rune == 0 {
		return
	}

	if cell == nil {
		cell = s.clearBlank()
	}

	s.updatePen(w, cell)
	w.WriteString(cell.Content())
	s.cur.X += cell.Width
	s.lastChar = cell.Rune

	if s.cur.X >= s.width {
		s.cur.X = s.width - 1
	}
}

// updatePen updates the cursor pen styles.
func (s *Screen) updatePen(w *bytes.Buffer, cell *Cell) {
	if cell == nil {
		cell = &BlankCell
	}

	if !cell.Style.Equal(s.cur.Style) {
		seq := cell.Style.DiffSequence(s.cur.Style)
		if cell.Style.Empty() && len(seq) > len(ansi.ResetStyle) {
			seq = ansi.ResetStyle
		}
		w.WriteString(seq)
		s.cur.Style = cell.Style
	}
	if !cell.Link.Equal(s.cur.Link) {
		w.WriteString(ansi.SetHyperlink(cell.Link.URL, cell.Link.URLID))
		s.cur.Link = cell.Link
	}
}

// emitRange emits a range of cells to the buffer. It it equivalent to calling
// [Screen.putCell] for each cell in the range. This is optimized to use
// [ansi.ECH] and [ansi.REP].
// Returns whether the cursor is at the end of interval or somewhere in the
// middle.
func (s *Screen) emitRange(w *bytes.Buffer, line Line, n int) (eoi bool) {
	for n > 0 {
		var count int
		for n > 1 && !cellEqual(line[0], line[1]) {
			s.putCell(w, line[0])
			line = line[1:]
			n--
		}

		cell0 := line[0]
		if n == 1 {
			s.putCell(w, cell0)
			return
		}

		count = 2
		for count < n && cellEqual(line[count], cell0) {
			count++
		}

		ech := ansi.EraseCharacter(count)
		cup := ansi.CursorPosition(s.cur.X+count, s.cur.Y)
		rep := ansi.RepeatPreviousCharacter(count)
		if count > len(ech)+len(cup) && cell0 != nil && cell0.Clear() {
			s.updatePen(w, cell0)
			w.WriteString(ech)

			// If this is the last cell, we don't need to move the cursor.
			if count < n {
				s.move(w, s.cur.X+count, s.cur.Y)
			} else {
				return true // cursor in the middle
			}
		} else if runes := cellRunes(cell0); count > len(rep) &&
			len(runes) == 1 {
			// NOTE: [ansi.REP] only repeats the last rune and won't work
			// if the last cell contains multiple runes.

			wrapPossible := s.cur.X+count >= s.width

			repCount := count
			if runes[0] != s.lastChar {
				s.putCell(w, cell0)
				repCount--
			}

			if wrapPossible {
				repCount--
				rep = ansi.RepeatPreviousCharacter(repCount)
			}

			s.updatePen(w, cell0)
			w.WriteString(rep)
			s.cur.X += repCount
			if wrapPossible {
				s.putCell(w, cell0)
			}
		} else {
			for i := 0; i < count; i++ {
				s.putCell(w, line[i])
			}
		}

		line = line[count:]
		n -= count
	}

	return
}

// putRange puts a range of cells from the old line to the new line.
// Returns whether the cursor is at the end of interval or somewhere in the
// middle.
func (s *Screen) putRange(w *bytes.Buffer, oldLine, newLine Line, y, start, end int) (eoi bool) {
	inline := min(len(ansi.CursorPosition(start+1, y+1)),
		min(len(ansi.HorizontalPositionAbsolute(start+1)),
			len(ansi.CursorForward(start+1))))
	if (end - start + 1) > inline {
		var j, same int
		for j, same = start, 0; j <= end; j++ {
			oldCell, newCell := oldLine[j], newLine[j]
			if same > 0 && oldCell != nil && oldCell.Rune == 0 && oldCell.Width == 0 {
				continue
			}
			if cellEqual(oldCell, newCell) {
				same++
			} else {
				if same > end-start {
					s.emitRange(w, newLine[start:], j-same-start)
					s.move(w, y, start)
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
func (s *Screen) clearToEnd(w *bytes.Buffer, blank *Cell, force bool) {
	if s.cur.Y >= 0 {
		for j := s.cur.X; j < s.width; j++ {
			c := s.curwin.Cell(j, s.cur.Y)
			if c != nil && !c.Equal(blank) {
				c = blank
				force = true
				break
			}
		}
	}

	if blank == nil {
		blank = &BlankCell
	}

	if force {
		s.updatePen(w, blank)

		count := s.width - s.cur.X
		eraseRight := ansi.EraseLineRight
		if len(eraseRight) <= count {
			w.WriteString(eraseRight)
		} else {
			for i := 0; i < count; i++ {
				s.putCell(w, blank)
			}
		}
	}
}

// clearBlank returns a blank cell based on the current cursor background color.
func (s *Screen) clearBlank() (c *Cell) {
	c = &BlankCell
	if !s.cur.Style.Empty() || !s.cur.Link.Empty() {
		c = new(Cell)
		*c = BlankCell
		c.Style = s.cur.Style
		c.Link = s.cur.Link
	}
	return
}

// insertCells inserts the count cells pointed by the given line at the current
// cursor position.
func (s *Screen) insertCells(w *bytes.Buffer, line Line, count int) {
	w.WriteString(ansi.InsertCharacter(count))
	for i := 0; count > 0; i++ {
		s.putCell(w, line[i])
		count--
	}
}

// transformLine transforms the given line in the current window to the
// corresponding line in the new window. It uses [ansi.ICH] and [ansi.DCH] to
// insert or delete characters.
func (s *Screen) transformLine(w *bytes.Buffer, y int) {
	var firstCell, oLastCell, nLastCell int // first, old last, new last index
	oldLine := s.curwin.buf.Lines[y]
	newLine := s.newwin.buf.Lines[y]

	var oline string
	for _, c := range oldLine {
		if c == nil {
			oline += " "
		} else {
			oline += c.Content()
		}
	}

	var nline string
	for _, c := range newLine {
		if c == nil {
			nline += " "
		} else {
			nline += c.Content()
		}
	}

	// Find the first changed cell in the line
	var lineChanged bool
	for i := 0; i < s.width; i++ {
		if !cellEqual(newLine[i], oldLine[i]) {
			lineChanged = true
			break
		}
	}

	const ceolStandoutGlitch = false
	if ceolStandoutGlitch && lineChanged {
		s.move(w, 0, y)
		s.clearToEnd(w, s.clearBlank(), false)
		s.putRange(w, oldLine, newLine, y, 0, s.width-1)
	} else {
		blank := newLine[0]

		// It might be cheaper to clear leading spaces with [ansi.EL] 1 i.e.
		// [ansi.EraseLineLeft].
		if blank == nil || blank.Clear() {
			var oFirstCell, nFirstCell int
			for oFirstCell = 0; oFirstCell < s.width; oFirstCell++ {
				if !cellEqual(oldLine[oFirstCell], blank) {
					break
				}
			}
			for nFirstCell = 0; nFirstCell < s.width; nFirstCell++ {
				if !cellEqual(newLine[nFirstCell], blank) {
					break
				}
			}

			if nFirstCell == oFirstCell {
				firstCell = nFirstCell

				// Find the first differing cell
				for firstCell < s.width &&
					cellEqual(oldLine[firstCell], newLine[firstCell]) {
					firstCell++
				}
			} else if oFirstCell > nFirstCell {
				firstCell = nFirstCell
			} else {
				firstCell = oFirstCell
				if el1 := ansi.EraseLineLeft; len(el1) < nFirstCell-oFirstCell {
					if nFirstCell >= s.width {
						s.move(w, 0, y)
						s.updatePen(w, blank)
						w.WriteString(ansi.EraseLineRight)
					} else {
						s.move(w, nFirstCell-1, y)
						s.updatePen(w, blank)
						w.WriteString(el1)
					}

					for firstCell < nFirstCell {
						var c *Cell
						if blank != nil {
							c = new(Cell)
							*c = *blank
						}

						oldLine[firstCell] = c
						firstCell++
					}
				}
			}
		} else {
			// Find the first differing cell
			for firstCell < s.width && cellEqual(newLine[firstCell], oldLine[firstCell]) {
				firstCell++
			}
		}

		// If we didn't find one, we're done
		if firstCell >= s.width {
			return
		}

		blank = newLine[s.width-1]
		if blank != nil && !blank.Clear() {
			// Find the last differing cell
			nLastCell = s.width - 1
			for nLastCell > firstCell && cellEqual(newLine[nLastCell], oldLine[nLastCell]) {
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
		oLastCell = s.width - 1
		for oLastCell > firstCell && cellEqual(oldLine[oLastCell], blank) {
			oLastCell--
		}

		// Find last non-blank cell in the new line.
		nLastCell = s.width - 1
		for nLastCell > firstCell && cellEqual(newLine[nLastCell], blank) {
			nLastCell--
		}

		el0 := ansi.EraseLineRight
		if nLastCell == firstCell && len(el0) < oLastCell-nLastCell {
			s.move(w, firstCell, y)
			if !cellEqual(newLine[firstCell], blank) {
				s.putCell(w, newLine[firstCell])
			}
			s.clearToEnd(w, blank, false)
		} else if nLastCell != oLastCell &&
			!cellEqual(newLine[nLastCell], oldLine[oLastCell]) {
			s.move(w, firstCell, y)
			if oLastCell-nLastCell > len(el0) {
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
			for cellEqual(newLine[nLastCell], oldLine[oLastCell]) {
				if !cellEqual(newLine[nLastCell-1], oldLine[oLastCell-1]) {
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
						wide := newLine[n+1]
						if wide == nil || wide.Rune != 0 || wide.Width != 0 {
							break
						}
						n--
						oLastCell--
					}
				} else if n >= firstCell && newLine[n] != nil && newLine[n].Width > 1 {
					for newLine[n+1] != nil &&
						newLine[n+1].Rune == 0 &&
						newLine[n+1].Width == 0 {
						n++
						oLastCell++
					}
				}

				s.move(w, n+1, y)
				ich := ansi.InsertCharacter(nLastCell - oLastCell)
				if nLastCell < nLastNonBlank || len(ich) > (m-n) {
					s.putRange(w, oldLine, newLine, y, n+1, m)
				} else {
					s.insertCells(w, newLine[n+1:], nLastCell-oLastCell)
				}
			} else if oLastCell > nLastCell {
				s.move(w, n+1, y)
				s.clearToEnd(w, blank, false)
				dch := ansi.DeleteCharacter(oLastCell - nLastCell)
				if len(dch) > len(ansi.EraseLineRight)+nLastNonBlank-(n+1) {
					if s.putRange(w, oldLine, newLine, y, n+1, nLastNonBlank) {
						s.move(w, nLastNonBlank+1, y)
					}
					s.clearToEnd(w, blank, false)
				} else {
					// [ansi.DCH] will shift in cells from the right margin so we need to
					// ensure that they are the right style.
					s.updatePen(w, blank)
					w.WriteString(dch)
				}
			}
		}
	}

	// Update the old line with the new line
	if s.width > firstCell {
		copy(oldLine[firstCell:], newLine[firstCell:])
	}
}

// clearToBottom clears the screen from the current cursor position to the end
// of the screen.
func (s *Screen) clearToBottom(w *bytes.Buffer, blank *Cell) {
	row, col := s.cur.Y, s.cur.X
	if row < 0 {
		row = 0
	}
	if col < 0 {
		col = 0
	}

	s.updatePen(w, blank)
	w.WriteString(ansi.EraseScreenBelow)

	for col < s.width {
		s.curwin.buf.Lines[row][col] = blank
		col++
	}

	for row = row + 1; row < s.height; row++ {
		for col = 0; col < s.width; col++ {
			s.curwin.buf.Lines[row][col] = blank
		}
	}
}

// clearBottom tests if clearing the end of the screen would satisfy part of
// the screen update. Scan backwards through lines in the screen checking if
// each is blank and on or more are changed.
// It returns the top line.
func (s *Screen) clearBottom(w *bytes.Buffer, total int) (top int) {
	top = total
	if total <= 0 {
		return
	}

	last := min(s.width, s.newwin.width+1)

	var blank *Cell
	nLines := s.newwin.buf.Lines
	oLines := s.curwin.buf.Lines
	if total-1 >= 0 && total-1 < len(nLines) && last-1 >= 0 && last-1 < len(nLines[total-1]) {
		blank = nLines[total-1][last-1]
	}

	if blank == nil || blank.Clear() {
		var row int
		for row = total - 1; row >= 0; row-- {
			var col int
			var ok bool
			for col, ok = 0, true; ok && col < last; col++ {
				ok = cellEqual(nLines[row][col], blank)
			}
			if !ok {
				break
			}

			for col = 0; ok && col < last; col++ {
				ok = cellEqual(oLines[row][col], blank)
			}
			if !ok {
				top = row
			}
		}

		if top < total {
			s.move(w, 0, top)
			s.clearToBottom(w, blank)
			// TODO: Line hashing
		}
	}

	return
}

// clearScreen clears the screen and put cursor at home.
func (s *Screen) clearScreen(w *bytes.Buffer, blank *Cell) {
	s.updatePen(w, blank)
	w.WriteString(ansi.CursorHomePosition)
	w.WriteString(ansi.EraseEntireScreen)
	s.cur.X, s.cur.Y = 0, 0

	for i := 0; i < s.height; i++ {
		for j := 0; j < s.width; j++ {
			s.curwin.buf.Lines[i][j] = blank
		}
	}
}

// clearUpdate forces a screen redraw.
func (s *Screen) clearUpdate(w *bytes.Buffer) {
	blank := s.clearBlank()
	nonEmpty := min(s.height, s.newwin.height+1)
	s.clearScreen(w, blank)
	nonEmpty = s.clearBottom(w, nonEmpty)
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
	var nonEmpty int

	// Force clear?
	if s.curwin.clear || s.newwin.clear {
		s.clearUpdate(b)
		s.curwin.clear = false
		s.newwin.clear = false
	} else {
		var changedLines int
		var i int
		nonEmpty = min(s.height, s.newwin.height+1)
		nonEmpty = s.clearBottom(b, nonEmpty)
		for i = 0; i < nonEmpty; i++ {
			_, nok := s.newwin.dirty[i]
			_, ook := s.curwin.dirty[i]
			if nok || ook {
				s.transformLine(b, i)
				changedLines++
			}
		}

		// Mark changed lines
		if i <= s.newwin.height {
			delete(s.newwin.dirty, i)
		}
		if i <= s.curwin.height {
			delete(s.curwin.dirty, i)
		}
	}

	// Sync windows and screen
	for i := nonEmpty; i <= s.newwin.height; i++ {
		delete(s.newwin.dirty, i)
	}
	for i := nonEmpty; i <= s.curwin.height; i++ {
		delete(s.curwin.dirty, i)
	}

	s.updatePen(b, nil)
}

// Close writes the final screen update and resets the screen.
func (s *Screen) Close() (err error) {
	b := new(bytes.Buffer)
	s.render(b)
	s.updatePen(b, &BlankCell)
	s.move(b, 0, s.height-1)
	s.clearToEnd(b, &BlankCell, true)

	// TODO: Set cursor to visible if needed.
	// TODO: Exit alternate screen buffer if needed.

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
	s.lastChar = -1
	s.cur = cursor{Position: Position{X: -1, Y: -1}}
	s.curwin = s.newWindow(0, 0, s.width, s.height)
	s.newwin = s.newWindow(0, 0, s.width, s.height)
}

// Resize resizes the screen.
func (s *Screen) Resize(width, height int) {
	s.width, s.height = width, height
	s.curwin.buf.Resize(width, height)
	s.newwin.buf.Resize(width, height)
	s.curwin.Resize(width, height)
	s.newwin.Resize(width, height)
}

// newWindow creates a new window.
func (s *Screen) newWindow(x, y, width, height int) (w *Window) {
	w = new(Window)
	w.x, w.y, w.width, w.height = x, y, width, height
	w.buf = NewBuffer(width, height)
	w.dirty = make(map[int][2]int)
	return
}

// Window represents a terminal Window.
type Window struct {
	buf           *Buffer
	parent        *Window        // the parent screen (nil if the window is a screen)
	dirty         map[int][2]int // map of the first and last changed cells in a row
	x, y          int            // the window's position relative to the parent
	width, height int
	clear         bool // whether to force refresh the screen
}

// NewWindow creates a new sub-window.
func (s *Screen) NewWindow(x, y, width, height int) (*Window, error) {
	r := Rect(x, y, width, height)
	if !r.In(s.Bounds().Rectangle) {
		return nil, fmt.Errorf("window out of bounds: %v not in %v", r, s.Bounds())
	}

	w := s.newWindow(x, y, width, height)
	w.parent = s.newwin
	return w, nil
}

// Cell implements Window.
func (w *Window) Cell(x int, y int) *Cell {
	if !Pos(x, y).In(w.Bounds().Rectangle) {
		return nil
	}
	return w.buf.Cell(w.x+x, w.y+y)
}

// Fill implements Window.
func (w *Window) Fill(cell *Cell) {
	w.FillInRect(cell, w.Bounds())
}

// FillInRect fills the cells in the specified rectangle with the specified
// cell.
func (w *Window) FillInRect(cell *Cell, r Rectangle) {
	if !r.In(w.Bounds().Rectangle) {
		return
	}

	w.buf.FillInRect(cell, r)
	for i := r.Min.Y; i < r.Max.Y; i++ {
		w.dirty[i] = [2]int{r.Min.X, r.Max.X - 1}
	}
}

// Clear implements Window.
func (w *Window) Clear() {
	w.ClearInRect(w.Bounds())
}

// ClearInRect clears the cells in the specified rectangle based on the current
// cursor background color. Use [SetPen] to set the background color.
func (w *Window) ClearInRect(r Rectangle) {
	if !r.In(w.Bounds().Rectangle) {
		return
	}

	w.buf.ClearInRect(r)
	w.clear = true
	for i := r.Min.Y; i < r.Max.Y; i++ {
		w.dirty[i] = [2]int{r.Min.X, r.Max.X - 1}
	}
}

// Draw implements Window.
func (w *Window) Draw(x int, y int, cell *Cell) (v bool) {
	if !Pos(x, y).In(w.Bounds().Rectangle) {
		return
	}

	cellWidth := 1
	if cell != nil {
		cellWidth = cell.Width
	}

	chg := w.dirty[y]
	chg[0] = min(chg[0], x)
	chg[1] = max(chg[1], x+cellWidth)
	w.dirty[y] = chg

	return w.buf.Draw(w.x+x, w.y+y, cell)
}

// Bounds returns the window's bounds.
func (w *Window) Bounds() Rectangle {
	return Rect(w.x, w.y, w.width, w.height)
}

// Resize resizes the window.
func (w *Window) Resize(width, height int) {
	w.width, w.height = width, height
}
