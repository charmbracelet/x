package vt

import (
	"bytes"
	"unicode"
	"unicode/utf8"

	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/ansi/parser"
	"github.com/charmbracelet/x/cellbuf"
)

type (
	// Style represents a style.
	Style = cellbuf.Style

	// Link represents a hyperlink.
	Link = cellbuf.Link
)

// Terminal represents a virtual terminal.
type Terminal struct {
	// The input buffer of the terminal.
	buf bytes.Buffer

	// The current focused screen.
	scr *Screen

	// Both main and alt screens.
	scrs [2]Screen

	// Terminal modes.
	modes  map[ansi.Mode]ModeSetting
	pmodes map[ansi.PrivateMode]ModeSetting

	// The ANSI parser to use.
	parser *ansi.Parser

	// The terminal's icon name and title.
	iconName, title string

	// Bell handler. When set, this function is called when a bell character is
	// received.
	Bell func()
}

// NewTerminal creates a new terminal.
func NewTerminal(w, h int) *Terminal {
	t := new(Terminal)
	t.scrs[0] = *NewScreen(w, h)
	t.scrs[1] = *NewScreen(w, h)
	t.scr = &t.scrs[0]
	t.parser = ansi.NewParser(parser.MaxParamsSize, 1024*4) // 4MB data buffer
	t.modes = map[ansi.Mode]ModeSetting{}
	t.pmodes = map[ansi.PrivateMode]ModeSetting{
		// These modes are set by default.
		ansi.AutoWrapMode:     ModeSet,
		ansi.CursorEnableMode: ModeSet,
	}
	return t
}

// At returns the cell at the given position.
func (t *Terminal) At(x int, y int) (cellbuf.Cell, bool) {
	return t.scr.Cell(x, y)
}

// Height returns the height of the terminal.
func (t *Terminal) Height() int {
	return t.scr.Height()
}

// Width returns the width of the terminal.
func (t *Terminal) Width() int {
	return t.scr.Width()
}

// Resize resizes the terminal.
func (t *Terminal) Resize(width int, height int) {
	t.scrs[0].Resize(width, height)
	t.scrs[1].Resize(width, height)
}

// Read reads data from the terminal input buffer.
func (t *Terminal) Read(p []byte) (n int, err error) {
	return t.buf.Read(p)
}

// Write writes data to the terminal output buffer.
func (t *Terminal) Write(p []byte) (n int, err error) {
	var state byte
	for len(p) > 0 {
		seq, width, m, newState := ansi.DecodeSequence(p, state, t.parser)
		r, rw := utf8.DecodeRune(seq)

		switch {
		case ansi.HasCsiPrefix(seq):
			t.handleCsi(seq)
		case ansi.HasOscPrefix(seq):
			t.handleOsc(seq)
		case ansi.HasDcsPrefix(seq):
			t.handleDcs(seq)
		case ansi.HasEscPrefix(seq):
			t.handleEsc(seq)
		case len(seq) == 1 && unicode.IsControl(r):
			t.handleControl(r)
		default:
			t.handleUtf8(seq, width, r, rw)
		}

		state = newState
		p = p[m:]
		n += m

		// x, y := t.Cursor().Pos.X, t.Cursor().Pos.Y
		// fmt.Printf("%q: %d %d\n", seq, x, y)
	}

	return
}

// Cursor returns the cursor.
func (t *Terminal) Cursor() Cursor {
	return t.scr.Cursor()
}

// Pos returns the cursor position.
func (t *Terminal) Pos() (int, int) {
	return t.scr.Pos()
}

// Title returns the terminal's title.
func (t *Terminal) Title() string {
	return t.title
}

// IconName returns the terminal's icon name.
func (t *Terminal) IconName() string {
	return t.iconName
}

// String returns the terminal's content as a string.
func (t *Terminal) String() string {
	return cellbuf.Render(t.scr)
}
