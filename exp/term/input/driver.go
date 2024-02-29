package input

import (
	"bufio"
	"io"
	"unicode/utf8"

	"github.com/charmbracelet/x/exp/term/ansi"
	"github.com/muesli/cancelreader"
)

// Flags to control the behavior of the driver.
const (
	// When this flag is set, the driver will treat both Ctrl+Space and Ctrl+@
	// as the same key sequence.
	//
	// Historically, the ANSI specs generate NUL (0x00) on both the Ctrl+Space
	// and Ctrl+@ key sequences. This flag allows the driver to treat both as
	// the same key sequence.
	FlagCtrlAt = 1 << iota

	// When this flag is set, the driver will treat the Tab key and Ctrl+I as
	// the same key sequence.
	//
	// Historically, the ANSI specs generate HT (0x09) on both the Tab key and
	// Ctrl+I. This flag allows the driver to treat both as the same key
	// sequence.
	FlagCtrlI

	// When this flag is set, the driver will treat the Enter key and Ctrl+M as
	// the same key sequence.
	//
	// Historically, the ANSI specs generate CR (0x0D) on both the Enter key
	// and Ctrl+M. This flag allows the driver to treat both as the same key
	FlagCtrlM

	// When this flag is set, the driver will treat Escape and Ctrl+[ as
	// the same key sequence.
	//
	// Historically, the ANSI specs generate ESC (0x1B) on both the Escape key
	// and Ctrl+[. This flag allows the driver to treat both as the same key
	// sequence.
	FlagCtrlOpenBracket

	// When this flag is set, the driver will treat space as a key rune instead
	// of a key symbol.
	FlagSpace

	// When this flag is set, the driver will send a BS (0x08 byte) character
	// instead of a DEL (0x7F byte) character when the Backspace key is
	// pressed.
	//
	// The VT100 terminal has both a Backspace and a Delete key. The VT220
	// terminal dropped the Backspace key and replaced it with the Delete key.
	// Both terminals send a DEL character when the Delete key is pressed.
	// Modern terminals and PCs later readded the Delete key but used a
	// different key sequence, and the Backspace key was standardized to send a
	// DEL character.
	FlagBackspace

	// When this flag is set, the driver will recognize the Find key instead of
	// treating it as a Home key.
	//
	// The Find key was part of the VT220 keyboard, and is no longer used in
	// modern day PCs.
	FlagFind

	// When this flag is set, the driver will recognize the Select key instead
	// of treating it as a End key.
	//
	// The Symbol key was part of the VT220 keyboard, and is no longer used in
	// modern day PCs.
	FlagSelect

	// When this flag is set, the driver won't register XTerm key sequences.
	//
	// Most modern terminals are compatible with XTerm, so this flag is
	// generally not needed.
	FlagNoXTerm

	// When this flag is set, the driver won't use Terminfo databases to
	// overwrite the default key sequences.
	FlagNoTerminfo

	// When this flag is set, the driver will preserve function keys (F13-F63)
	// as symbols.
	//
	// Since these keys are not part of today's standard 20th century keyboard,
	// we treat them as F1-F12 modifier keys i.e. ctrl/shift/alt + Fn combos.
	// Key definitions come from Terminfo, this flag is only useful when
	// FlagTerminfo is not set.
	FlagFKeys
)

// Driver represents an ANSI terminal input Driver.
// It reads input events and parses ANSI sequences from the terminal input
// buffer.
type Driver struct {
	table map[string]KeyEvent
	rd    *bufio.Reader
	cr    cancelreader.CancelReader
	term  string

	// paste is the bracketed paste mode buffer.
	// When nil, bracketed paste mode is disabled.
	paste []byte

	// flags to control the behavior of the driver.
	flags int
}

// NewDriver returns a new ANSI input driver.
// This driver uses ANSI control codes compatible with VT100/VT200 terminals,
// and XTerm. It supports reading Terminfo databases to overwrite the default
// key sequences.
func NewDriver(r io.Reader, term string, flags int) *Driver {
	d := new(Driver)
	// TODO: implement cancelable reader
	cr, err := cancelreader.NewReader(r)
	if err == nil {
		d.cr = cr
		r = cr
	}
	d.rd = bufio.NewReaderSize(r, 256)
	d.flags = flags
	d.term = term
	// Populate the key sequences table.
	d.registerKeys(flags)
	return d
}

// Cancel cancels the underlying reader.
func (d *Driver) Cancel() bool {
	if d.cr != nil {
		return d.cr.Cancel()
	}
	return false
}

// Close closes the underlying reader.
func (d *Driver) Close() error {
	if d.cr != nil {
		return d.cr.Close()
	}
	return nil
}

// ReadInput reads input events from the terminal.
func (d *Driver) ReadInput() ([]Event, error) {
	nb, ne, err := d.peekInput()
	if err != nil {
		return nil, err
	}

	// Consume the event
	if _, err := d.rd.Discard(nb); err != nil {
		return nil, err
	}

	return ne, nil
}

// PeekInput peeks at input events from the terminal without consuming
// them.
func (d *Driver) PeekInput() ([]Event, error) {
	_, ne, err := d.peekInput()
	if err != nil {
		return nil, err
	}

	return ne, err
}

// TODO:: split this into a static Parse function that can be used by other
// drivers.
func (d *Driver) peekInput() (int, []Event, error) {
	ev := make([]Event, 0)
	p, err := d.rd.Peek(1)
	if err != nil {
		return 0, nil, err
	}

	// The number of bytes buffered.
	bufferedBytes := d.rd.Buffered()
	// Peek more bytes if needed.
	if bufferedBytes > len(p) {
		p, err = d.rd.Peek(bufferedBytes)
		if err != nil {
			return 0, nil, err
		}
	}

	// Lookup table first
	if k, ok := d.table[string(p)]; ok {
		return len(p), []Event{k}, nil
	}

	i := 0 // index of the current byte
	for i < len(p) {
		nb, e := d.peekOne(p[i:])
		if e != nil {
			switch e.(type) {
			case PasteStartEvent:
				d.paste = []byte{}
			case PasteEndEvent:
				// Decode the captured data into runes.
				var paste []rune
				for len(d.paste) > 0 {
					r, w := utf8.DecodeRune(d.paste)
					if r != utf8.RuneError {
						d.paste = d.paste[w:]
					}
					paste = append(paste, r)
				}
				d.paste = nil // reset the buffer
				ev = append(ev, PasteEvent(paste))
			}
			ev = append(ev, e)
		}
		i += nb
	}

	return i, ev, nil
}

// peekOne peeks a single event from the input buffer.
// This may return a nil event.
func (d *Driver) peekOne(p []byte) (int, Event) {
	if len(p) == 0 {
		return 0, nil
	}

	var (
		alt = false
		i   = 0
	)

	var parser func(int, []byte, bool) (int, Event)

begin:
	b := p[i]
	switch b {
	case ansi.ESC:
		if i+1 >= len(p) {
			break
		}

		switch p[i+1] {
		case 'O': // Esc-prefixed SS3
			parser = d.parseSs3
		case 'P': // Esc-prefixed DCS
			parser = d.parseDcs
		case '[': // Esc-prefixed CSI
			parser = d.parseCsi
		case ']': // Esc-prefixed OSC
			parser = d.parseOsc
		case '_': // Esc-prefixed APC
			parser = d.parseApc
		}

		if parser != nil || alt {
			break
		}

		alt = true
		i++
		goto begin
	case ansi.SS3:
		parser = d.parseSs3
	case ansi.DCS:
		parser = d.parseDcs
	case ansi.CSI:
		parser = d.parseCsi
	case ansi.OSC:
		parser = d.parseOsc
	case ansi.APC:
		parser = d.parseApc
	}

	if parser != nil {
		n, e := parser(i, p, alt)
		if d.paste != nil {
			if _, ok := e.(PasteEndEvent); !ok {
				// Not a valid sequence. We collect bytes until we reach a
				// bracketed-paste end sequence (ESC [ 201 ~).
				d.paste = append(d.paste, b)
				return 1, nil
			}
		}
		i += n
		return i, e
	}

	if d.paste != nil {
		// Collect bytes until we reach a bracketed-paste end sequence.
		d.paste = append(d.paste, b)
		return 1, nil
	}

	if b <= ansi.US || b == ansi.DEL || b == ansi.SP {
		// Single byte control code or printable ASCII/UTF-8
		k := d.table[string(b)]
		if alt {
			k.Mod |= Alt
		}
		i++
		return i, k
	} else if utf8.RuneStart(b) {
		// Collect UTF-8 sequences into a slice of runes.
		// We need to do this for multi-rune emojis to work.
		var k KeyEvent
		for rw := 0; i < len(p); i += rw {
			var r rune
			r, rw = utf8.DecodeRune(p[i:])
			if r == utf8.RuneError || r <= ansi.US || r == ansi.DEL || r == ansi.SP {
				break
			}
			k.Runes = append(k.Runes, r)
		}

		if alt {
			k.Mod |= Alt
		}

		// Zero runes means we didn't find a valid UTF-8 sequence.
		if len(k.Runes) > 0 {
			return i, k
		}
	}

	return 1, UnknownEvent(string(p[0]))
}

func (d *Driver) parseCsi(i int, p []byte, alt bool) (int, Event) {
	var seq []byte
	if p[i] == ansi.CSI || p[i] == ansi.ESC {
		seq = append(seq, p[i])
		i++
	}
	if i < len(p) && p[i-1] == ansi.ESC && p[i] == '[' {
		seq = append(seq, p[i])
		i++
	}

	// Scan parameter bytes in the range 0x30-0x3F
	for ; i < len(p) && p[i] >= 0x30 && p[i] <= 0x3F; i++ {
		seq = append(seq, p[i])
	}
	// Scan intermediate bytes in the range 0x20-0x2F
	for ; i < len(p) && p[i] >= 0x20 && p[i] <= 0x2F; i++ {
		seq = append(seq, p[i])
	}
	// Scan final byte in the range 0x40-0x7E
	if i >= len(p) || p[i] < 0x40 || p[i] > 0x7E {
		// XXX: Some terminals like URxvt send invalid CSI sequences on key
		// events such as shift modified keys (\x1b [ <func> $). We try to
		// lookup the sequence in the table and return it as a key event if it
		// exists. Otherwise, we report an unknown event.
		var e Event = UnknownEvent(seq)
		if key, ok := d.table[string(seq)]; ok {
			if alt {
				key.Mod |= Alt
			}
			e = key
		}
		return len(seq), e
	}

	// Add the final byte
	seq = append(seq, p[i])
	k, ok := d.table[string(seq)]
	if ok {
		if alt {
			k.Mod |= Alt
		}
		return len(seq), k
	}

	// TODO: improve and cleanup this block
	csi := ansi.CsiSequence(seq)
	initial := csi.Initial()
	cmd := csi.Command()
	params := ansi.Params(csi.Params())
	switch {
	// Bracketed-paste must come before XTerm modifyOtherKeys
	case string(seq) == "\x1b[200~":
		// bracketed-paste start
		return len(seq), PasteStartEvent{}
	case string(seq) == "\x1b[201~":
		// bracketed-paste end
		return len(seq), PasteEndEvent{}
	case string(seq) == "\x1b[M" && i+3 < len(p):
		// Handle X10 mouse
		return len(seq) + 3, parseX10MouseEvent(append(seq, p[i+1:i+3]...))
	case initial == '<' && (cmd == 'm' || cmd == 'M'):
		// Handle SGR mouse
		return len(seq), parseSGRMouseEvent(seq)
	case initial == 0 && cmd == 'u':
		// Kitty keyboard protocol
		return len(seq), parseKittyKeyboard(seq)
	case initial == '?' && cmd == 'u' && len(params) > 0:
		// Kitty keyboard flags
		return len(seq), KittyKeyboardEvent(params[0][0])
	case initial == 0 && cmd == '~':
		// XTerm modifyOtherKeys 2
		return len(seq), parseXTermModifyOtherKeys(seq)
	case initial == '?' && cmd == 'c':
		// Primary Device Attributes
		da1 := make([]uint, len(params))
		for i, p := range params {
			da1[i] = p[0]
		}
		return len(seq), PrimaryDeviceAttributesEvent(da1)
	}

	return len(seq), UnknownCsiEvent{csi}
}

// parseSs3 parses a SS3 sequence.
// See https://vt100.net/docs/vt220-rm/chapter4.html#S4.4.4.2
func (d *Driver) parseSs3(i int, p []byte, alt bool) (int, Event) {
	var seq []byte
	if p[i] == ansi.SS3 || p[i] == ansi.ESC {
		seq = append(seq, p[i])
		i++
	}
	if i < len(p) && p[i-1] == ansi.ESC && p[i] == 'O' {
		seq = append(seq, p[i])
		i++
	}

	// Scan a GL character
	// A GL character is a single byte in the range 0x21-0x7E
	// See https://vt100.net/docs/vt220-rm/chapter2.html#S2.3.2
	if i >= len(p) || p[i] < 0x21 || p[i] > 0x7E {
		var e Event = UnknownEvent(seq)
		if key, ok := d.table[string(seq)]; ok {
			if alt {
				key.Mod |= Alt
			}
			e = key
		}
		return len(seq), e
	}

	// Add the GL character
	seq = append(seq, p[i])
	k, ok := d.table[string(seq)]
	if ok {
		if alt {
			k.Mod |= Alt
		}
		return len(seq), k
	}

	return len(seq), UnknownEvent(seq)
}

func (d *Driver) parseOsc(i int, p []byte, _ bool) (int, Event) {
	var seq []byte
	if p[i] == ansi.OSC || p[i] == ansi.ESC {
		seq = append(seq, p[i])
		i++
	}
	if i < len(p) && p[i-1] == ansi.ESC && p[i] == ']' {
		seq = append(seq, p[i])
		i++
	}

	// Scan a OSC sequence
	// An OSC sequence is terminated by a BEL, ESC, or ST character
	for ; i < len(p) && p[i] != ansi.BEL && p[i] != ansi.ESC && p[i] != ansi.ST; i++ {
		seq = append(seq, p[i])
	}

	if i >= len(p) {
		return len(seq), UnknownEvent(seq)
	}
	seq = append(seq, p[i])

	// Check 7-bit ST (string terminator) character
	if len(p) > i+1 && p[i] == ansi.ESC && p[i+1] == '\\' {
		i++
		seq = append(seq, p[i])
	}

	osc := ansi.OscSequence(seq)
	switch osc.Identifier() {
	case "10":
		return len(seq), ForegroundColorEvent{xParseColor(osc.Data())}
	case "11":
		return len(seq), BackgroundColorEvent{xParseColor(osc.Data())}
	case "12":
		return len(seq), CursorColorEvent{xParseColor(osc.Data())}
	}

	return len(seq), UnknownOscEvent{osc}
}

// parseCtrl parses a control sequence that gets terminated by a ST character.
func (d *Driver) parseCtrl(intro8, intro7 byte) func(int, []byte, bool) (int, Event) {
	return func(i int, p []byte, _ bool) (int, Event) {
		var seq []byte
		if p[i] == intro8 || p[i] == ansi.ESC {
			seq = append(seq, p[i])
			i++
		}
		if i < len(p) && p[i-1] == ansi.ESC && p[i] == intro7 {
			seq = append(seq, p[i])
			i++
		}

		// Scan control sequence
		// Most common control sequence is terminated by a ST character
		// ST is a 7-bit string terminator character is (ESC \)
		for ; i < len(p) && p[i] != ansi.ST && p[i] != ansi.ESC; i++ {
			seq = append(seq, p[i])
		}

		if i >= len(p) {
			switch intro8 {
			case ansi.DCS:
				return len(seq), UnknownDcsEvent(seq)
			case ansi.APC:
				return len(seq), UnknownApcEvent(seq)
			default:
				return len(seq), UnknownEvent(seq)
			}
		}
		seq = append(seq, p[i])

		// Check 7-bit ST (string terminator) character
		if len(p) > i+1 && p[i] == ansi.ESC && p[i+1] == '\\' {
			i++
			seq = append(seq, p[i])
		}

		switch intro8 {
		case ansi.DCS:
			return len(seq), UnknownDcsEvent(seq)
		case ansi.APC:
			return len(seq), UnknownApcEvent(seq)
		default:
			return len(seq), UnknownEvent(seq)
		}
	}
}

func (d *Driver) parseDcs(i int, p []byte, alt bool) (int, Event) {
	// DCS sequences are introduced by DCS (0x90) or ESC P (0x1b 0x50)
	return d.parseCtrl(ansi.DCS, 'P')(i, p, alt)
}

func (d *Driver) parseApc(i int, p []byte, alt bool) (int, Event) {
	// APC sequences are introduced by APC (0x9f) or ESC _ (0x1b 0x5f)
	return d.parseCtrl(ansi.APC, '_')(i, p, alt)
}
