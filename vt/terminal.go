package vt

import (
	"bytes"
	"io"
	"log"
	"sync"
	"unicode"
	"unicode/utf8"

	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/ansi/parser"
	"github.com/charmbracelet/x/cellbuf"
	"github.com/charmbracelet/x/wcwidth"
)

type (
	// Style represents a style.
	Style = cellbuf.Style

	// Link represents a hyperlink.
	Link = cellbuf.Link
)

// Terminal represents a virtual terminal.
type Terminal struct {
	mu sync.Mutex

	tmp []byte
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
	logger *log.Logger

	// The terminal's icon name and title.
	iconName, title string

	// Bell handler. When set, this function is called when a bell character is
	// received.
	Bell func()

	// Damage handler. When set, this function is called when a cell is damaged
	// or changed.
	Damage func(Damage)
}

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

	for _, opt := range opts {
		opt(t)
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
	t.mu.Lock()
	t.scrs[0].Resize(width, height)
	t.scrs[1].Resize(width, height)
	t.tabstops = DefaultTabStops(width)
	t.mu.Unlock()
}

// Read reads data from the terminal input buffer.
func (t *Terminal) Read(p []byte) (n int, err error) {
	if t.closed {
		return 0, io.EOF
	}

	if t.buf.Len() == 0 {
		return 0, nil
	}

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
		t.handleOsc(seq.Bytes())
	case ansi.CsiSequence:
		t.handleCsi(seq.Bytes())
	case ansi.EscSequence:
		t.handleEsc(seq.Bytes())
	case ansi.ControlCode:
		t.handleControl(rune(seq))
	case ansi.Rune:
		t.handleUtf8([]byte{byte(seq)}, wcwidth.RuneWidth(rune(seq)))
	case ansi.Grapheme:
		t.handleUtf8([]byte(seq.Cluster), seq.Width)
	}
}

// Write writes data to the terminal output buffer.
func (t *Terminal) Write(p []byte) (n int, err error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	// TODO: Use just a parser and a dispatcher. We gotta make [ansi.Parser]
	// support graphemes first tho.
	t.parser.Parse(t.dispatcher, p)
	return len(p), nil

	t.tmp = append(t.tmp, p...)
	n += len(p)

	var state byte
	for len(t.tmp) > 0 {
		seq, width, m, newState, ok := ansi.DecodeSequence(t.tmp, state, t.parser)
		if !ok {
			// Incomplete sequence.
			return
		}

		switch {
		case ansi.HasSosPrefix(seq): /* Ignore */
		case ansi.HasApcPrefix(seq): /* Ignore */
		case ansi.HasPmPrefix(seq): /* Ignore */
		case ansi.HasCsiPrefix(seq):
			t.handleCsi(seq)
		case ansi.HasOscPrefix(seq):
			t.handleOsc(seq)
		case ansi.HasDcsPrefix(seq):
			t.handleDcs(seq)
		case ansi.HasEscPrefix(seq):
			t.handleEsc(seq)
		default:
			r, _ := utf8.DecodeRune(seq)
			if len(seq) == 1 && unicode.IsControl(r) {
				t.handleControl(r)
			} else {
				t.handleUtf8(seq, width)
			}
		}

		state = newState
		t.tmp = t.tmp[m:]
		// n += m
	}

	return
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

// String returns the terminal's content as a string.
func (t *Terminal) String() string {
	return cellbuf.Render(t.scr)
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
