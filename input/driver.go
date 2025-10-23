//nolint:unused,revive,nolintlint
package input

import (
	"bytes"
	"io"
	"unicode/utf8"

	"github.com/muesli/cancelreader"
)

// Logger is a simple logger interface.
type Logger interface {
	Printf(format string, v ...any)
}

// win32InputState is a state machine for parsing key events from the Windows
// Console API into escape sequences and utf8 runes, and keeps track of the last
// control key state to determine modifier key changes. It also keeps track of
// the last mouse button state and window size changes to determine which mouse
// buttons were released and to prevent multiple size events from firing.
type win32InputState struct {
	ansiBuf                    [256]byte
	ansiIdx                    int
	utf16Buf                   [2]rune
	utf16Half                  bool
	lastCks                    uint32 // the last control key state for the previous event
	lastMouseBtns              uint32 // the last mouse button state for the previous event
	lastWinsizeX, lastWinsizeY int16  // the last window size for the previous event to prevent multiple size events from firing
}

// Reader represents an input event reader. It reads input events and parses
// escape sequences from the terminal input buffer and translates them into
// human-readable events.
type Reader struct {
	rd    cancelreader.CancelReader
	table map[string]Key // table is a lookup table for key sequences.

	term string // term is the terminal name $TERM.

	// paste is the bracketed paste mode buffer.
	// When nil, bracketed paste mode is disabled.
	paste []byte

	buf [256]byte // do we need a larger buffer?

	// keyState keeps track of the current Windows Console API key events state.
	// It is used to decode ANSI escape sequences and utf16 sequences.
	keyState win32InputState

	// pending holds partial sequences that need more data to complete
	pending []byte
	// inStringTerminated tracks if we're inside an OSC/DCS/APC/SOS/PM sequence
	inStringTerminated bool

	parser Parser
	logger Logger
}

// NewReader returns a new input event reader. The reader reads input events
// from the terminal and parses escape sequences into human-readable events. It
// supports reading Terminfo databases. See [Parser] for more information.
//
// Example:
//
//	r, _ := input.NewReader(os.Stdin, os.Getenv("TERM"), 0)
//	defer r.Close()
//	events, _ := r.ReadEvents()
//	for _, ev := range events {
//	  log.Printf("%v", ev)
//	}
func NewReader(r io.Reader, termType string, flags int) (*Reader, error) {
	d := new(Reader)
	cr, err := newCancelreader(r, flags)
	if err != nil {
		return nil, err
	}

	d.rd = cr
	d.table = buildKeysTable(flags, termType)
	d.term = termType
	d.parser.flags = flags
	return d, nil
}

// SetLogger sets a logger for the reader.
func (d *Reader) SetLogger(l Logger) {
	d.logger = l
}

// Read implements [io.Reader].
func (d *Reader) Read(p []byte) (int, error) {
	return d.rd.Read(p) //nolint:wrapcheck
}

// Cancel cancels the underlying reader.
func (d *Reader) Cancel() bool {
	return d.rd.Cancel()
}

// Close closes the underlying reader.
func (d *Reader) Close() error {
	return d.rd.Close() //nolint:wrapcheck
}

func (d *Reader) readEvents() ([]Event, error) {
	nb, err := d.rd.Read(d.buf[:])
	if err != nil {
		return nil, err //nolint:wrapcheck
	}

	var events []Event
	buf := d.buf[:nb]

	// Check if we had pending data from previous incomplete sequences
	hadPending := len(d.pending) > 0

	// Prepend any pending data from previous incomplete sequences
	if hadPending {
		combined := make([]byte, len(d.pending)+len(buf))
		copy(combined, d.pending)
		copy(combined[len(d.pending):], buf)
		buf = combined
		d.pending = nil
		d.inStringTerminated = false
	}

	// Lookup table first (only if no pending data)
	if !hadPending && bytes.HasPrefix(buf, []byte{'\x1b'}) {
		if k, ok := d.table[string(buf)]; ok {
			if d.logger != nil {
				d.logger.Printf("input: %q", buf)
			}
			events = append(events, KeyPressEvent(k))
			return events, nil
		}
	}

	var i int
	for i < len(buf) {
		nb, ev := d.parser.parseSequence(buf[i:])
		if d.logger != nil {
			d.logger.Printf("input: %q", buf[i:i+nb])
		}

		// Handle bracketed-paste
		if d.paste != nil {
			if _, ok := ev.(PasteEndEvent); !ok {
				d.paste = append(d.paste, buf[i])
				i++
				continue
			}
		}

		switch ev.(type) {
		case UnknownEvent:
			// Check if this might be an incomplete string-terminated sequence
			if d.isIncompleteStringTerminated(buf[i : i+nb]) {
				// Buffer this data and wait for more
				d.pending = make([]byte, len(buf[i:]))
				copy(d.pending, buf[i:])
				d.inStringTerminated = true
				return events, nil
			}
			// If the sequence is not recognized by the parser, try looking it up.
			if k, ok := d.table[string(buf[i:i+nb])]; ok {
				ev = KeyPressEvent(k)
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
			events = append(events, PasteEvent(paste))
		case nil:
			i++
			continue
		}

		if mevs, ok := ev.(MultiEvent); ok {
			events = append(events, []Event(mevs)...)
		} else {
			events = append(events, ev)
		}
		i += nb
	}

	return events, nil
}

// isIncompleteStringTerminated checks if the given bytes represent an incomplete
// string-terminated sequence (OSC, DCS, APC, SOS, PM) that needs more data.
func (d *Reader) isIncompleteStringTerminated(b []byte) bool {
	if len(b) < 2 {
		return false
	}

	// Check for OSC sequences: ESC ] or 0x9D
	if (b[0] == '\x1b' && b[1] == ']') || b[0] == '\x9d' {
		return d.isIncompleteOscLike(b)
	}

	// Check for DCS sequences: ESC P or 0x90
	if (b[0] == '\x1b' && b[1] == 'P') || b[0] == '\x90' {
		return d.isIncompleteOscLike(b)
	}

	// Check for APC sequences: ESC _ or 0x9F
	if (b[0] == '\x1b' && b[1] == '_') || b[0] == '\x9f' {
		return d.isIncompleteOscLike(b)
	}

	// Check for SOS sequences: ESC X or 0x98
	if (b[0] == '\x1b' && b[1] == 'X') || b[0] == '\x98' {
		return d.isIncompleteOscLike(b)
	}

	// Check for PM sequences: ESC ^ or 0x9E
	if (b[0] == '\x1b' && b[1] == '^') || b[0] == '\x9e' {
		return d.isIncompleteOscLike(b)
	}

	return false
}

// isIncompleteOscLike checks if a string-terminated sequence is incomplete
func (d *Reader) isIncompleteOscLike(b []byte) bool {
	// Look for terminators: BEL (0x07), ESC \ (ST), CAN (0x18), SUB (0x1A)
	for i := 0; i < len(b); i++ {
		switch b[i] {
		case '\x07': // BEL
			return false
		case '\x18', '\x1a': // CAN, SUB
			return false
		case '\x1b': // ESC
			if i+1 < len(b) && b[i+1] == '\\' {
				return false // Found ST (ESC \)
			}
		}
	}
	// No terminator found, sequence is incomplete
	return true
}
