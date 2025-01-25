package vt

import "github.com/charmbracelet/x/cellbuf"

// eraseCharacter erases n characters starting from the cursor position. It
// does not move the cursor. This is equivalent to [ansi.ECH].
func (t *Terminal) eraseCharacter(n int) {
	x, y := t.scr.CursorPosition()
	rect := cellbuf.Rect(x, y, n, 1)
	t.scr.Fill(t.scr.blankCell(), rect)
	t.atPhantom = false
	// ECH does not move the cursor.
}
