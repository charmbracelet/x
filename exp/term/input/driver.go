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
	rd    cancelreader.CancelReader
	table map[string]Key

	term string // the $TERM name to use

	// paste is the bracketed paste mode buffer.
	// When nil, bracketed paste mode is disabled.
	paste []byte

	internalEvents []Event   // holds peeked events
	buf            [256]byte // do we need a larger buffer?

	// prevMouseState keeps track of the previous mouse state to determine mouse
	// up button events.
	prevMouseState coninput.ButtonState

	// flags to control the behavior of the driver.
	flags int
}

// NewDriver returns a new ANSI input driver.
// This driver uses ANSI control codes compatible with VT100/VT200 terminals,
// and XTerm. It supports reading Terminfo databases to overwrite the default
// key sequences.
func NewDriver(r io.Reader, term string, flags int) (*Driver, error) {
	d := new(Driver)
	d.internalEvents = make([]Event, 0, 10) // initial size of 10

	cr, err := newCancelreader(r)
	if err != nil {
		return nil, err
	}

	d.rd = cr
	d.flags = flags
	d.term = term
	// Populate the key sequences table.
	d.table = registerKeys(flags, term)
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

func (d *Driver) readInput(e []Event) (n int, err error) {
	if len(e) == 0 {
		return 0, nil
	}

	// If there are any peeked events, return them first.
	if len(d.internalEvents) > 0 {
		n = copy(e, d.internalEvents)
		d.internalEvents = d.internalEvents[n:]
	}

	// Read new events
	if n < len(e) {
		ev, err := d.PeekInput(len(e) - n)
		if err != nil {
			return n, err
		}
		nl := copy(e[n:], ev)
		n += nl

		// Consume the events from the internalEvents buffer.
		d.internalEvents = d.internalEvents[nl:]
	}

	return
}

func (d *Driver) peekInput(n int) ([]Event, error) {
	if n <= 0 {
		return []Event{}, nil
	}

	// Peek events from the internalEvents buffer first.
	if len(d.internalEvents) > 0 {
		if len(d.internalEvents) >= n {
			return d.internalEvents[:n], nil
		}
		n -= len(d.internalEvents)
	}

	// Peek new events
	nb, err := d.rd.Read(d.buf[:])
	if err != nil {
		return nil, err
	}

	buf := d.buf[:nb]

	// Lookup table first
	if k, ok := d.table[string(buf)]; ok {
		d.internalEvents = append(d.internalEvents, KeyDownEvent(k))
		return d.internalEvents, nil
	}

	var i int
	for i < len(buf) {
		nb, ev := ParseSequence(buf[i:])

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
			if k, ok := d.table[string(buf[i:i+nb])]; ok {
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
			d.internalEvents = append(d.internalEvents, PasteEvent(paste))
		case nil:
			i++
			continue
		}

		if mevs, ok := ev.(MultiEvent); ok {
			d.internalEvents = append(d.internalEvents, []Event(mevs)...)
		} else {
			d.internalEvents = append(d.internalEvents, ev)
		}
		i += nb
	}

	if len(d.internalEvents) >= n {
		return d.internalEvents[:n], nil
	}

	return d.internalEvents, nil
}
