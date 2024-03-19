package input

import (
	"io"
	"unicode/utf8"

	"github.com/erikgeiser/coninput"
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

	// When this flag is set, the driver will use Terminfo databases to
	// overwrite the default key sequences.
	FlagTerminfo

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
	rd     cancelreader.CancelReader
	parser *EventParser

	// paste is the bracketed paste mode buffer.
	// When nil, bracketed paste mode is disabled.
	paste []byte

	buf [256]byte // do we need a larger buffer?

	// prevMouseState keeps track of the previous mouse state to determine mouse
	// up button events.
	prevMouseState coninput.ButtonState
}

// NewDriver returns a new ANSI input driver.
// This driver uses ANSI control codes compatible with VT100/VT200 terminals,
// and XTerm. It supports reading Terminfo databases to overwrite the default
// key sequences.
//
// The parser argument is used to parse ANSI sequences into input events. If
// nil is passed, a default parser will be used.
func NewDriver(r io.Reader, parser *EventParser) (*Driver, error) {
	d := new(Driver)
	cr, err := newCancelreader(r)
	if err != nil {
		return nil, err
	}

	d.rd = cr
	if parser == nil {
		parser = &EventParser{}
	}
	d.parser = parser
	return d, nil
}

// Cancel cancels the underlying reader.
func (d *Driver) Cancel() bool {
	return d.rd.Cancel()
}

// Close closes the underlying reader.
func (d *Driver) Close() error {
	return d.rd.Close()
}

func (d *Driver) readEvents() (e []Event, err error) {
	nb, err := d.rd.Read(d.buf[:])
	if err != nil {
		return nil, err
	}

	buf := d.buf[:nb]

	// Lookup table first
	if k, ok := d.parser.LookupSequence(string(buf)); ok {
		e = append(e, KeyDownEvent(k))
		return
	}

	var i int
	for i < len(buf) {
		nb, ev := d.parser.ParseSequence(buf[i:])

		// Handle bracketed-paste
		if d.paste != nil {
			if _, ok := ev.(PasteEndEvent); !ok {
				d.paste = append(d.paste, buf[i])
				i++
				continue
			}
		}

		switch ev.(type) {
		case UnknownCsiEvent, UnknownSs3Event, UnknownEvent:
			// If the sequence is not recognized by the parser, try looking it up.
			if k, ok := d.parser.LookupSequence(string(buf[i : i+nb])); ok {
				ev = KeyDownEvent(k)
			}
		case PasteStartEvent:
			d.paste = []byte{}
		case PasteEndEvent:
			// Decode the captured data into runes.
			var paste []rune
			for len(d.paste) > 0 {
				r, w := utf8.DecodeRune(d.paste)
				if r != utf8.RuneError {
					paste = append(paste, r)
				}
				d.paste = d.paste[w:]
			}
			d.paste = nil // reset the buffer
			e = append(e, PasteEvent(paste))
		case nil:
			i++
			continue
		}

		if mevs, ok := ev.(MultiEvent); ok {
			e = append(e, []Event(mevs)...)
		} else {
			e = append(e, ev)
		}
		i += nb
	}

	return
}
