package vt

import (
	"bytes"
	"image/color"
	"io"
	"time"

	"github.com/charmbracelet/uv"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/ansi/parser"
	"github.com/rivo/uniseg"
)

// Terminal represents a virtual terminal.
type Terminal struct {
	handlers

	// The terminal's indexed 256 colors.
	colors [256]color.Color

	// Both main and alt screens.
	scrs [2]Screen

	// Character sets
	charsets [4]CharSet

	// log is the logger to use.
	logger Logger

	// terminal default colors.
	fg, bg, cur color.Color

	// Terminal modes.
	modes map[ansi.Mode]ansi.ModeSetting

	// The current focused screen.
	scr *Screen

	// The last written character.
	lastChar rune // either ansi.Rune or ansi.Grapheme

	// The ANSI parser to use.
	parser *ansi.Parser

	cb Callbacks

	// The terminal's icon name and title.
	iconName, title string

	// tabstop is the list of tab stops.
	tabstops *uv.TabStops

	// I/O pipes.
	pr *io.PipeReader
	pw *io.PipeWriter

	// The GL and GR character set identifiers.
	gl, gr  int
	gsingle int // temporarily select GL or GR

	// Indicates if the terminal is closed.
	closed bool

	// atPhantom indicates if the cursor is out of bounds.
	// When true, and a character is written, the cursor is moved to the next line.
	atPhantom bool
}

var (
	defaultFg  = color.White
	defaultBg  = color.Black
	defaultCur = color.White
)

// NewTerminal creates a new terminal.
func NewTerminal(w, h int, opts ...Option) *Terminal {
	t := new(Terminal)
	t.scrs[0] = *NewScreen(w, h)
	t.scrs[1] = *NewScreen(w, h)
	t.scrs[0].cb = &t.Callbacks
	t.scrs[1].cb = &t.Callbacks
	t.scr = &t.scrs[0]
	t.parser = ansi.NewParser() // 4MB data buffer
	t.parser.SetHandler(ansi.Handler{
		Print:     t.handlePrint,
		Execute:   t.handleControl,
		HandleCsi: t.handleCsi,
		HandleEsc: t.handleEsc,
		HandleDcs: t.handleDcs,
		HandleOsc: t.handleOsc,
		HandleApc: t.handleApc,
		// Pm:      t.handlePm,
		// Sos:     t.handleSos,
	})
	t.pr, t.pw = io.Pipe()
	t.resetModes()
	t.tabstops = uv.DefaultTabStops(w)
	t.fgColor = defaultFg
	t.bgColor = defaultBg
	t.curColor = defaultCur
	t.registerDefaultHandlers()

	for _, opt := range opts {
		opt(t)
	}

	return t
}

// SetCallbacks sets the terminal's callbacks.
func (t *Terminal) SetCallbacks(cb Callbacks) {
	t.cb = cb
	t.scrs[0].cb = &t.cb
	t.scrs[1].cb = &t.cb
}

// CellAt returns the current focused screen cell at the given x, y position.
// It returns nil if the cell is out of bounds.
func (t *Terminal) CellAt(x, y int) *uv.Cell {
	return t.scr.CellAt(x, y)
}

// Height returns the height of the terminal.
func (t *Terminal) Height() int {
	return t.scr.Height()
}

// Width returns the width of the terminal.
func (t *Terminal) Width() int {
	return t.scr.Width()
}

// CursorPosition returns the terminal's cursor position.
func (t *Terminal) CursorPosition() uv.Position {
	x, y := t.scr.CursorPosition()
	return uv.Pos(x, y)
}

// Resize resizes the terminal.
func (t *Terminal) Resize(width int, height int) {
	x, y := t.scr.CursorPosition()
	if t.atPhantom {
		if x < width-1 {
			t.atPhantom = false
			x++
		}
	}

	if y < 0 {
		y = 0
	}
	if y >= height {
		y = height - 1
	}
	if x < 0 {
		x = 0
	}
	if x >= width {
		x = width - 1
	}

	t.scrs[0].Resize(width, height)
	t.scrs[1].Resize(width, height)
	t.tabstops = uv.DefaultTabStops(width)

	t.setCursor(x, y)
}

// Read reads data from the terminal input buffer.
func (t *Terminal) Read(p []byte) (n int, err error) {
	if t.closed {
		return 0, io.EOF
	}

	return t.pr.Read(p)
}

// Close closes the terminal.
func (t *Terminal) Close() error {
	if t.closed {
		return nil
	}

	t.closed = true
	return nil
}

// Write writes data to the terminal output buffer.
func (t *Terminal) Write(p []byte) (int, error) {
	t.parser.Parse(p)
	return len(p), nil
}

// InputPipe returns the terminal's input pipe.
// This can be used to send input to the terminal.
func (t *Terminal) InputPipe() io.Writer {
	return t.pw
}

// Paste pastes text into the terminal.
// If bracketed paste mode is enabled, the text is bracketed with the
// appropriate escape sequences.
func (t *Terminal) Paste(text string) {
	if t.isModeSet(ansi.BracketedPasteMode) {
		io.WriteString(t.pw, ansi.BracketedPasteStart)     //nolint:errcheck
		defer io.WriteString(t.pw, ansi.BracketedPasteEnd) //nolint:errcheck
	}

	io.WriteString(t.pw, text) //nolint:errcheck
}

// SendText sends text to the terminal.
func (t *Terminal) SendText(text string) {
	io.WriteString(t.pw, text) //nolint:errcheck
}

// SendKeys sends multiple keys to the terminal.
func (t *Terminal) SendKeys(keys ...Key) {
	for _, k := range keys {
		t.SendKey(k)
	}
}

// ForegroundColor returns the terminal's foreground color.
func (t *Terminal) ForegroundColor() color.Color {
	return t.fg
}

// SetForegroundColor sets the terminal's foreground color.
func (t *Terminal) SetForegroundColor(c color.Color) {
	t.fg = c
}

// BackgroundColor returns the terminal's background color.
func (t *Terminal) BackgroundColor() color.Color {
	return t.bg
}

// SetBackgroundColor sets the terminal's background color.
func (t *Terminal) SetBackgroundColor(c color.Color) {
	t.bg = c
}

// CursorColor returns the terminal's cursor color.
func (t *Terminal) CursorColor() color.Color {
	return t.cur
}

// SetCursorColor sets the terminal's cursor color.
func (t *Terminal) SetCursorColor(c color.Color) {
	t.cur = c
}

// IndexedColor returns a terminal's indexed color. An indexed color is a color
// between 0 and 255.
func (t *Terminal) IndexedColor(i int) color.Color {
	if i < 0 || i > 255 {
		return nil
	}

	c := t.colors[i]
	if c == nil {
		// Return the default color.
		return ansi.ExtendedColor(i) //nolint:gosec,staticcheck
	}

	return c
}

// SetIndexedColor sets a terminal's indexed color.
// The index must be between 0 and 255.
func (t *Terminal) SetIndexedColor(i int, c color.Color) {
	if i < 0 || i > 255 {
		return
	}

	t.colors[i] = c
}

// resetTabStops resets the terminal tab stops to the default set.
func (t *Terminal) resetTabStops() {
	t.tabstops = uv.DefaultTabStops(t.Width())
}
