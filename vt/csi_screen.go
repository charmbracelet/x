package vt

import (
	"github.com/charmbracelet/x/ansi"
)

func (t *Terminal) handleScreen() {
	width, height := t.Width(), t.Height()
	x, y := t.scr.CursorPosition()

	switch t.parser.Cmd() {
	case 'J': // Erase in Display [ansi.ED]
		count, _ := t.parser.Param(0, 0)
		switch count {
		case 0: // Erase screen below (from after cursor position)
			rect1 := Rect(x, y, width, 1)            // cursor to end of line
			rect2 := Rect(0, y+1, width, height-y-1) // next line onwards
			for _, rect := range []Rectangle{rect1, rect2} {
				t.scr.Fill(t.scr.blankCell(), rect)
			}
		case 1: // Erase screen above (including cursor)
			rect := Rect(0, 0, width, y+1)
			t.scr.Fill(t.scr.blankCell(), rect)
		case 2: // erase screen
			fallthrough
		case 3: // erase display
			// TODO: Scrollback buffer support?
			t.scr.Clear()
		}
	case 'L': // Insert Line [ansi.IL]
		n, _ := t.parser.Param(0, 1)
		if n == 0 {
			n = 1
		}
		if t.scr.InsertLine(n) {
			// Move the cursor to the left margin.
			t.scr.setCursorX(0, true)
		}

	case 'M': // Delete Line [ansi.DL]
		n, _ := t.parser.Param(0, 1)
		if n == 0 {
			n = 1
		}
		if t.scr.DeleteLine(n) {
			// If the line was deleted successfully, move the cursor to the
			// left.
			// Move the cursor to the left margin.
			t.scr.setCursorX(0, true)
		}

	case 'X': // Erase Character [ansi.ECH]
		// It clears character attributes as well but not colors.
		n, _ := t.parser.Param(0, 1)
		if n == 0 {
			n = 1
		}
		t.eraseCharacter(n)

	case 'r': // Set Top and Bottom Margins [ansi.DECSTBM]
		top, _ := t.parser.Param(0, 1)
		if top < 1 {
			top = 1
		}

		bottom, _ := t.parser.Param(1, height)
		if bottom < 1 {
			bottom = height
		}

		if top >= bottom {
			break
		}

		// Rect is [x, y) which means y is exclusive. So the top margin
		// is the top of the screen minus one.
		t.scr.setVerticalMargins(top-1, bottom)

		// Move the cursor to the top-left of the screen or scroll region
		// depending on [ansi.DECOM].
		t.setCursorPosition(0, 0)

	case 's': // Set Left and Right Margins [ansi.DECSLRM]
		// These conflict with each other. When [ansi.DECSLRM] is set, the we
		// set the left and right margins. Otherwise, we save the cursor
		// position.

		if t.isModeSet(ansi.LeftRightMarginMode) {
			// Set Left Right Margins [ansi.DECSLRM]
			left, _ := t.parser.Param(0, 1)
			if left < 1 {
				left = 1
			}

			right, _ := t.parser.Param(1, width)
			if right < 1 {
				right = width
			}

			if left >= right {
				break
			}

			t.scr.setHorizontalMargins(left-1, right)

			// Move the cursor to the top-left of the screen or scroll region
			// depending on [ansi.DECOM].
			t.setCursorPosition(0, 0)
		} else {
			// Save Current Cursor Position [ansi.SCOSC]
			t.scr.SaveCursor()
		}
	}
}

func (t *Terminal) handleLine() {
	switch t.parser.Cmd() {
	case 'K': // Erase in Line [ansi.EL]
		// NOTE: Erase Line (EL) erases all character attributes but not cell
		// bg color.
		count, _ := t.parser.Param(0, 0)
		x, y := t.scr.CursorPosition()
		w := t.scr.Width()

		switch count {
		case 0: // Erase from cursor to end of line
			t.eraseCharacter(w - x)
		case 1: // Erase from start of line to cursor
			rect := Rect(0, y, x+1, 1)
			t.scr.Fill(t.scr.blankCell(), rect)
		case 2: // Erase entire line
			rect := Rect(0, y, w, 1)
			t.scr.Fill(t.scr.blankCell(), rect)
		}
	case 'S': // Scroll Up [ansi.SU]
		n, _ := t.parser.Param(0, 1)
		if n == 0 {
			n = 1
		}
		t.scr.ScrollUp(n)
	case 'T': // Scroll Down [ansi.SD]
		n, _ := t.parser.Param(0, 1)
		if n == 0 {
			n = 1
		}
		t.scr.ScrollDown(n)
	}
}

// eraseCharacter erases n characters starting from the cursor position. It
// does not move the cursor. This is equivalent to [ansi.ECH].
func (t *Terminal) eraseCharacter(n int) {
	x, y := t.scr.CursorPosition()
	rect := Rect(x, y, n, 1)
	t.scr.Fill(t.scr.blankCell(), rect)
	t.atPhantom = false
	// ECH does not move the cursor.
}
