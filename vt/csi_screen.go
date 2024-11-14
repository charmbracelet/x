package vt

func (t *Terminal) handleScreen() {
	width, height := t.Width(), t.Height()
	_, y := t.scr.CursorPosition()

	switch t.parser.Cmd() {
	case 'J':
		count, _ := t.parser.Param(0, 0)
		switch count {
		case 0: // Erase screen below (including cursor)
			rect := Rect(0, y, width, height-y)
			t.scr.Fill(t.scr.blankCell(), rect)
			if t.Damage != nil {
				t.Damage(RectDamage(rect))
			}
		case 1: // Erase screen above (including cursor)
			rect := Rect(0, 0, width, y+1)
			t.scr.Fill(t.scr.blankCell(), rect)
			if t.Damage != nil {
				t.Damage(RectDamage(rect))
			}
		case 2: // erase screen
			fallthrough
		case 3: // erase display
			// TODO: Scrollback buffer support?
			t.scr.Clear()
			if t.Damage != nil {
				t.Damage(ScreenDamage{width, height})
			}
		}
	case 'L': // IL - Insert Line
		n, _ := t.parser.Param(0, 1)
		t.scr.InsertLine(n)
		// Move the cursor to the left margin.
		t.scr.setCursorX(0, true)

	case 'M': // DL - Delete Line
		n, _ := t.parser.Param(0, 1)
		t.scr.DeleteLine(n)
		// Move the cursor to the left margin.
		t.scr.setCursorX(0, true)

	case 'X':
		// ECH - Erase Character
		// It clears character attributes as well but not colors.
		n, _ := t.parser.Param(0, 1)
		t.eraseCharacter(n)

	case 'r': // DECSTBM - Set Top and Bottom Margins
		top, _ := t.parser.Param(0, 1)
		bottom, _ := t.parser.Param(1, height)
		if top >= bottom {
			break
		}

		// Rect is [x, y) which means y is exclusive. So the top margin
		// is the top of the screen minus one.
		t.scr.scroll.Min.Y = top - 1
		t.scr.scroll.Max.Y = bottom

		// Move the cursor to the top-left of the screen or scroll region
		// depending on [ansi.DECOM].
		t.setCursorPosition(0, 0)
	}
}

func (t *Terminal) handleLine() {
	switch t.parser.Cmd() {
	case 'K': // EL - Erase in Line
		// NOTE: Erase Line (EL) erases all character attributes but not cell
		// bg color.
		count, _ := t.parser.Param(0, 0)
		x, y := t.scr.CursorPosition()
		w := t.scr.Width()

		switch count {
		case 0: // Erase from cursor to end of line
			t.eraseCharacter(w - x)
			if t.Damage != nil {
				t.Damage(RectDamage(Rect(x, y, w-x, 1)))
			}
		case 1: // Erase from start of line to cursor
			rect := Rect(0, y, x+1, 1)
			t.scr.Fill(t.scr.blankCell(), rect)
			if t.Damage != nil {
				t.Damage(RectDamage(rect))
			}
		case 2: // Erase entire line
			rect := Rect(0, y, w, 1)
			t.scr.Fill(t.scr.blankCell(), rect)
			if t.Damage != nil {
				t.Damage(RectDamage(rect))
			}
		}
	case 'S': // SU - Scroll Up
		n, _ := t.parser.Param(0, 1)
		t.scr.ScrollUp(n)
	case 'T': // SD - Scroll Down
		n, _ := t.parser.Param(0, 1)
		t.scr.ScrollDown(n)
	}
}

// eraseCharacter erases n characters starting from the cursor position. It
// does not move the cursor. This is equivalent to [ansi.ECH].
func (t *Terminal) eraseCharacter(n int) {
	x, y := t.scr.CursorPosition()
	rect := Rect(x, y, n, 1)
	t.scr.Fill(t.scr.blankCell(), rect)
	// ECH does not move the cursor.
}
