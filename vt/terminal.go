package vt

import (
	"bytes"
	"image/color"
	"io"
	"sync"

	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/ansi/parser"
)

// Terminal represents a virtual terminal.
type Terminal struct {
	mu sync.Mutex

	// The input buffer of the terminal.
	buf    bytes.Buffer
	closed bool

	// The current focused screen.
	scr *Screen

	// Both main and alt screens.
	scrs [2]Screen

	// tabstop is the list of tab stops.
	tabstops TabStops

	// Terminal modes.
	modes  map[ansi.ANSIMode]ModeSetting
	pmodes map[ansi.DECMode]ModeSetting

	// The ANSI parser to use.
	parser *ansi.Parser

	// log is the logger to use.
	logger Logger

	// The terminal's icon name and title.
	iconName, title string

	// terminal default colors.
	fg, bg, cur color.Color
	colors      [256]color.Color

	// Bell handler. When set, this function is called when a bell character is
	// received.
	Bell func()

	// Damage handler. When set, this function is called when a cell is damaged
	// or changed.
	Damage func(Damage)
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
	t.scr = &t.scrs[0]
	t.parser = ansi.NewParser(parser.MaxParamsSize, 1024*1024*4) // 4MB data buffer
	t.modes = map[ansi.ANSIMode]ModeSetting{}
	t.pmodes = map[ansi.DECMode]ModeSetting{
		// These modes are set by default.
		ansi.AutowrapMode:     ModeSet,
		ansi.CursorEnableMode: ModeSet,
	}
	t.tabstops = DefaultTabStops(w)
	t.fg = defaultFg
	t.bg = defaultBg
	t.cur = defaultCur

	for _, opt := range opts {
		opt(t)
	}

	return t
}

// Screen returns the main terminal screen.
func (t *Terminal) Screen() *Screen {
	return &t.scrs[0]
}

// AltScreen returns the alternate terminal screen.
func (t *Terminal) AltScreen() *Screen {
	return &t.scrs[1]
}

// Cell returns the current focused screen cell at the given x, y position. It
// returns nil if the cell is out of bounds.
func (t *Terminal) Cell(x, y int) *Cell {
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
	t.tabstops = DefaultTabStops(width)
}

// Read reads data from the terminal input buffer.
func (t *Terminal) Read(p []byte) (n int, err error) {
	if t.closed {
		return 0, io.EOF
	}

	if t.buf.Len() == 0 {
		return 0, nil
	}

	t.mu.Lock()
	defer t.mu.Unlock()
	return t.buf.Read(p)
}

// Close closes the terminal.
func (t *Terminal) Close() error {
	if t.closed {
		return nil
	}

	t.closed = true
	return nil
}

// dispatcher parses and dispatches escape sequences and operates on the terminal.
func (t *Terminal) dispatcher(seq ansi.Sequence) {
	switch seq := seq.(type) {
	case ansi.ApcSequence:
	case ansi.PmSequence:
	case ansi.SosSequence:
	case ansi.OscSequence:
		t.handleOsc(seq)
	case ansi.CsiSequence:
		t.handleCsi(seq)
	case ansi.EscSequence:
		t.handleEsc(seq)
	case ansi.ControlCode:
		t.handleControl(rune(seq))
	case ansi.Rune:
		t.handleUtf8(seq)
	case ansi.Grapheme:
		t.handleUtf8(seq)
	}
}

// Write writes data to the terminal output buffer.
func (t *Terminal) Write(p []byte) (n int, err error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	var i int
	for i < len(p) {
		t.parser.Advance(t.dispatcher, p[i], i < len(p)-1)
		// TODO: Support grapheme clusters (mode 2027).
		i++
	}

	return i, nil
}

// Cursor returns the cursor.
func (t *Terminal) Cursor() Cursor {
	return t.scr.Cursor()
}

// Title returns the terminal's title.
func (t *Terminal) Title() string {
	return t.title
}

// IconName returns the terminal's icon name.
func (t *Terminal) IconName() string {
	return t.iconName
}

// InputPipe returns the terminal's input pipe.
// This can be used to send input to the terminal.
func (t *Terminal) InputPipe() io.Writer {
	return &t.buf
}

// Paste pastes text into the terminal.
// If bracketed paste mode is enabled, the text is bracketed with the
// appropriate escape sequences.
func (t *Terminal) Paste(text string) {
	if mode, ok := t.pmodes[ansi.BracketedPasteMode]; ok && mode.IsSet() {
		t.buf.WriteString(ansi.BracketedPasteStart)
		defer t.buf.WriteString(ansi.BracketedPasteEnd)
	}

	t.buf.WriteString(text)
}

// SendText sends text to the terminal.
func (t *Terminal) SendText(text string) {
	t.buf.WriteString(text)
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
		return ansi.ExtendedColor(i) //nolint:gosec
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
