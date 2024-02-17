package ansi

import (
	"bufio"
	"fmt"
	"io"
	"unicode/utf8"

	"github.com/charmbracelet/x/exp/term/ansi"
	"github.com/charmbracelet/x/exp/term/input"
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

// driver represents a terminal ANSI input driver.
type driver struct {
	table map[string]input.KeyEvent
	rd    *bufio.Reader
	term  string
	flags int
}

var _ input.Driver = &driver{}

// NewDriver returns a new ANSI input driver.
// This driver uses ANSI control codes compatible with VT100/VT200 terminals,
// and XTerm. It supports reading Terminfo databases to overwrite the default
// key sequences.
func NewDriver(r io.Reader, term string, flags int) input.Driver {
	d := &driver{
		rd:    bufio.NewReaderSize(r, 256),
		flags: flags,
		term:  term,
	}
	// Populate the key sequences table.
	d.registerKeys(flags)
	return d
}

// ReadInput implements input.Driver.
func (d *driver) ReadInput() ([]input.Event, error) {
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

const esc = string(ansi.ESC)

// PeekInput implements input.Driver.
func (d *driver) PeekInput() ([]input.Event, error) {
	_, ne, err := d.peekInput()
	if err != nil {
		return nil, err
	}

	return ne, err
}

func (d *driver) peekInput() (int, []input.Event, error) {
	ev := make([]input.Event, 0)
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
		return len(p), []input.Event{k}, nil
	}

	i := 0 // index of the current byte

	for i < len(p) {
		var alt bool
		b := p[i]

	begin:
		switch b {
		case ansi.ESC:
			if bufferedBytes == 1 {
				// Special case for Esc
				i++
				ev = append(ev, d.table[esc])
				continue
			}

			if i+1 >= len(p) {
				// Not enough bytes to peek
				break
			}

			switch p[i+1] {
			case ansi.ESC:
				alt = true
				b = p[i+1]
				// Start over with the next byte
				goto begin
			case 'O': // Esc-prefixed SS3
				d.handleSeq(d.parseSs3, i, p, alt, &i, &ev)
				continue
			case 'P': // Esc-prefixed DCS
			case '[': // Esc-prefixed CSI
				d.handleSeq(d.parseCsi, i, p, alt, &i, &ev)
				continue
			case ']': // Esc-prefixed OSC
				d.handleSeq(d.parseOsc, i, p, alt, &i, &ev)
				continue
			}
		case ansi.SS3:
			d.handleSeq(d.parseSs3, i, p, alt, &i, &ev)
			continue
		case ansi.DCS:
		case ansi.CSI:
			d.handleSeq(d.parseCsi, i, p, alt, &i, &ev)
			continue
		case ansi.OSC:
			d.handleSeq(d.parseOsc, i, p, alt, &i, &ev)
			continue
		default:
			// Unknown sequence
		}

		// Single byte control code or printable ASCII/UTF-8
		if b <= ansi.US || b == ansi.DEL || b == ansi.SP {
			k := d.table[string(b)]
			nb := 1
			if alt {
				k.Mod |= input.Alt
			}
			i += nb
			ev = append(ev, k)
			continue
		} else if utf8.RuneStart(b) { // Printable ASCII/UTF-8
			nb := utf8ByteLen(b)
			if nb == -1 || nb > bufferedBytes {
				return i, ev, fmt.Errorf("invalid UTF-8 sequence: %x", p)
			}

			r := rune(b)
			if nb > 1 {
				r, _ = utf8.DecodeRune(p[i : i+nb])
			}

			k := input.KeyEvent{Rune: r}
			if alt {
				k.Mod |= input.Alt
			}

			i += nb
			ev = append(ev, k)
			continue
		}
	}

	return i, ev, nil
}

// helper function to handle adding events and the number of bytes consumed.
func (d *driver) handleSeq(
	seqFn func(int, []byte, bool) (int, input.Event),
	i int, p []byte, alt bool,
	np *int, ne *[]input.Event,
) {
	n, e := seqFn(i, p, alt)
	*np += n
	*ne = append(*ne, e)
}

func (d *driver) parseCsi(i int, p []byte, alt bool) (int, input.Event) {
	var seq string
	if p[i] == ansi.CSI || p[i] == ansi.ESC {
		seq += string(p[i])
		i++
	}
	if i < len(p) && p[i-1] == ansi.ESC && p[i] == '[' {
		seq += string(p[i])
		i++
	}

	// Scan parameter bytes in the range 0x30-0x3F
	for ; i < len(p) && p[i] >= 0x30 && p[i] <= 0x3F; i++ {
		seq += string(p[i])
	}
	// Scan intermediate bytes in the range 0x20-0x2F
	for ; i < len(p) && p[i] >= 0x20 && p[i] <= 0x2F; i++ {
		seq += string(p[i])
	}
	// Scan final byte in the range 0x40-0x7E
	if i >= len(p) || p[i] < 0x40 || p[i] > 0x7E {
		if key, ok := d.table[seq]; ok {
			if alt {
				key.Mod |= input.Alt
			}
			return len(seq), key
		}
		return len(seq), input.UnknownEvent{string(seq)}
	}

	seq += string(p[i])
	csi := ansi.CsiSequence(seq)
	initial := csi.Initial()
	cmd := csi.Command()
	switch {
	case seq == "\x1b[M" && i+3 < len(p):
		// Handle X10 mouse
		return len(seq) + 3, parseX10MouseEvent(append([]byte(seq), p[i+1:i+3]...))
	case initial == '<' && (cmd == 'm' || cmd == 'M'):
		return len(seq), parseSGRMouseEvent([]byte(seq))
	case initial == 0 && cmd == 'u':
		// Kitty keyboard protocol
		params := ansi.Params(csi.Params())
		key := input.KeyEvent{}
		if len(params) > 0 {
			code := int(params[0][0])
			if sym, ok := kittyKeyMap[code]; ok {
				key.Sym = sym
			} else {
				key.Rune = rune(code)
				// TODO: support alternate keys
			}
		}
		if len(params) > 1 {
			mod := int(params[1][0])
			if mod > 1 {
				key.Mod = fromKittyMod(int(params[1][0] - 1))
			}
			if len(params[1]) > 1 {
				switch int(params[1][1]) {
				case 0, 1:
					key.Action = input.KeyPress
				case 2:
					key.Action = input.KeyRepeat
				case 3:
					key.Action = input.KeyRelease
				}
			}
		}
		if len(params) > 2 {
			key.Rune = rune(params[2][0])
		}
		return len(seq), key
	}

	k, ok := d.table[seq]
	if ok {
		return len(seq), k
	}

	return len(seq), UnknownCsiEvent(seq)
}

// parseSs3 parses a SS3 sequence.
// See https://vt100.net/docs/vt220-rm/chapter4.html#S4.4.4.2
func (d *driver) parseSs3(i int, p []byte, alt bool) (int, input.Event) {
	var seq string
	if p[i] == ansi.SS3 || p[i] == ansi.ESC {
		seq += string(p[i])
		i++
	}
	if i < len(p) && p[i-1] == ansi.ESC && p[i] == 'O' {
		seq += string(p[i])
		i++
	}

	// Scan a GL character
	// A GL character is a single byte in the range 0x21-0x7E
	// See https://vt100.net/docs/vt220-rm/chapter2.html#S2.3.2
	if i >= len(p) || p[i] < 0x21 || p[i] > 0x7E {
		if key, ok := d.table[seq]; ok {
			if alt {
				key.Mod |= input.Alt
			}
			return len(seq), key
		}
		return len(seq), input.UnknownEvent{string(seq)}
	}
	seq += string(p[i])

	k, ok := d.table[seq]
	if ok {
		if alt {
			k.Mod |= input.Alt
		}
		return len(seq), k
	}

	return len(seq), UnknownSs3Event(seq)
}

func (d *driver) parseOsc(i int, p []byte, _ bool) (int, input.Event) {
	var seq string
	if p[i] == ansi.OSC || p[i] == ansi.ESC {
		seq += string(p[i])
		i++
	}
	if i < len(p) && p[i-1] == ansi.ESC && p[i] == ']' {
		seq += string(p[i])
		i++
	}

	// Scan a OSC sequence
	// An OSC sequence is terminated by a BEL, ESC, or ST character
	for ; i < len(p) && p[i] != ansi.BEL && p[i] != ansi.ESC && p[i] != ansi.ST; i++ {
		seq += string(p[i])
	}

	if i >= len(p) {
		return len(seq), input.UnknownEvent{string(seq)}
	}
	seq += string(p[i])

	// Check 7-bit ST (string terminator) character
	if len(p) > i+1 && p[i] == ansi.ESC && p[i+1] == '\\' {
		i++
		seq += string(p[i])
	}

	osc := ansi.OscSequence(seq)
	switch osc.Identifier() {
	case "10":
		return len(seq), FgColorEvent{xParseColor(osc.Data())}
	case "11":
		return len(seq), BgColorEvent{xParseColor(osc.Data())}
	case "12":
		return len(seq), CursorColorEvent{xParseColor(osc.Data())}
	}

	return len(seq), UnknownOscEvent(seq)
}

func utf8ByteLen(b byte) int {
	if b <= 0b0111_1111 { // 0x00-0x7F
		return 1
	} else if b >= 0b1100_0000 && b <= 0b1101_1111 { // 0xC0-0xDF
		return 2
	} else if b >= 0b1110_0000 && b <= 0b1110_1111 { // 0xE0-0xEF
		return 3
	} else if b >= 0b1111_0000 && b <= 0b1111_0111 { // 0xF0-0xF7
		return 4
	}
	return -1
}
