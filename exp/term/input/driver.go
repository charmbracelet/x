package input

import (
	"bufio"
	"io"
	"unicode/utf8"

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
		nb, e := ParseSequence(p[i:])

		// Handle bracketed-paste
		if d.paste != nil {
			if _, ok := e.(PasteEndEvent); !ok {
				d.paste = append(d.paste, p[i])
				i++
				continue
			}
		}

		switch e.(type) {
		case UnknownCsiEvent, UnknownSs3Event, UnknownEvent:
			// If the sequence is not recognized by the parser, try looking it up.
			if k, ok := d.table[string(p[i:i+nb])]; ok {
				e = k
			}
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
		case nil:
			i++
			continue
		}

		ev = append(ev, e)
		i += nb
	}

	return i, ev, nil
}

func parsePrimaryDevAttrs(params [][]uint) Event {
	// Primary Device Attributes
	da1 := make([]uint, len(params))
	for i, p := range params {
		da1[i] = p[0]
	}
	return PrimaryDeviceAttributesEvent(da1)
}
