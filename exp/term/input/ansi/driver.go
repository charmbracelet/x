package ansi

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"unicode/utf8"

	"github.com/charmbracelet/x/exp/term/ansi"
	"github.com/charmbracelet/x/exp/term/input"
)

// ErrUnsupportedReader is returned when the reader is not a *bufio.Reader.
var ErrUnsupportedReader = fmt.Errorf("unsupported reader")

// Flags to control the behavior of the driver.
const (
	Fctrlsp       = 1 << iota // treat NUL as ctrl+space, otherwise ctrl+@
	Ftabsym                   // treat tab as a symbol
	Fentersym                 // treat enter as a symbol
	Fescsym                   // treat escape as a symbol
	Fspacesym                 // treat space as a symbol
	Fdelbackspace             // treat DEL as a symbol
	Ffindhome                 // treat find symbol as home
	Fselectend                // treat select symbol as end

	Stdflags = Ftabsym | Fentersym | Fescsym | Fspacesym | Fdelbackspace | Ffindhome | Fselectend
)

// driver represents a terminal ANSI input driver.
type driver struct {
	table map[string]input.Key
	rd    *bufio.Reader
	term  string
	flags int
}

var _ input.Driver = &driver{}

// NewDriver returns a new ANSI input driver.
// This driver uses ANSI control codes compatible with VT100/VT200 terminals.
func NewDriver(r io.Reader, term string, flags int) input.Driver {
	if r == nil {
		r = os.Stdin
	}
	if term == "" {
		term = os.Getenv("TERM")
	}
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

	peekedBytes := 0
	i := 0 // index of the current byte

	addEvent := func(n int, e input.Event) {
		peekedBytes += n
		i += n
		ev = append(ev, e)
	}

	for i < len(p) {
		var alt bool
		b := p[i]

	begin:
		switch b {
		case ansi.ESC:
			if bufferedBytes == 1 {
				// Special case for Esc
				addEvent(1, input.KeyEvent(d.table[esc]))
				continue
			}

			if i+1 >= len(p) {
				// Not enough bytes to peek
				break
			}

			i++ // we know there's at least one more byte
			peekedBytes++
			switch p[i] {
			case 'O': // Esc-prefixed SS3
				nb, e, err := d.parseSs3(i, p, alt)
				if err != nil {
					return peekedBytes, ev, err
				}

				addEvent(nb, e)
				continue
			case 'P': // Esc-prefixed DCS
			case '[': // Esc-prefixed CSI
				nb, e, err := d.parseCsi(i, p, alt)
				if err != nil {
					return peekedBytes, ev, err
				}

				addEvent(nb, e)
				continue
			case ']': // Esc-prefixed OSC
				nb, e, err := d.parseOsc(i, p, alt)
				if err != nil {
					return peekedBytes, ev, err
				}

				addEvent(nb, e)
				continue
			}

			alt = true
			b = p[i]

			goto begin
		case ansi.SS3:
			nb, e, err := d.parseSs3(i, p, alt)
			if err != nil {
				return peekedBytes, ev, err
			}

			addEvent(nb, e)
			continue
		case ansi.DCS:
		case ansi.CSI:
			nb, e, err := d.parseCsi(i, p, alt)
			if err != nil {
				return peekedBytes, ev, err
			}

			addEvent(nb, e)
			continue
		case ansi.OSC:
			nb, e, err := d.parseOsc(i, p, alt)
			if err != nil {
				return peekedBytes, ev, err
			}

			addEvent(nb, e)
			continue
		}

		// Single byte control code or printable ASCII/UTF-8
		if b <= ansi.US || b == ansi.DEL || b == ansi.SP {
			k := input.KeyEvent(d.table[string(b)])
			nb := 1
			if alt {
				k.Mod |= input.Alt
			}
			addEvent(nb, k)
			continue
		} else if utf8.RuneStart(b) { // Printable ASCII/UTF-8
			nb := utf8ByteLen(b)
			if nb == -1 || nb > bufferedBytes {
				return peekedBytes, ev, fmt.Errorf("invalid UTF-8 sequence: %x", p)
			}

			r := rune(b)
			if nb > 1 {
				r, _ = utf8.DecodeRune(p[i : i+nb])
			}

			k := input.KeyEvent{Rune: r}
			if alt {
				k.Mod |= input.Alt
			}

			addEvent(nb, k)
			continue
		}
	}

	return peekedBytes, ev, nil
}

func (d *driver) parseCsi(i int, p []byte, alt bool) (n int, e input.Event, err error) {
	if p[i] == '[' {
		n++
	}

	i++
	seq := "\x1b["

	// Scan parameter bytes in the range 0x30-0x3F
	for ; p[i] >= 0x30 && p[i] <= 0x3F; i++ {
		n++
		seq += string(p[i])
	}
	// Scan intermediate bytes in the range 0x20-0x2F
	for ; p[i] >= 0x20 && p[i] <= 0x2F; i++ {
		n++
		seq += string(p[i])
	}
	// Scan final byte in the range 0x40-0x7E
	if p[i] < 0x40 || p[i] > 0x7E {
		return n, nil, fmt.Errorf("%w: invalid CSI sequence: %q", input.ErrUnknownEvent, seq[2:])
	}
	n++
	seq += string(p[i])

	// Handle X10 mouse
	if seq == "\x1b[M" && i+3 < len(p) {
		btn := int(p[i+1] - 32)
		x := int(p[i+2] - 32)
		y := int(p[i+3] - 32)
		return n + 3, input.MouseEvent{X: x, Y: y, Btn: input.Button(btn)}, nil
	}

	k, ok := d.table[seq]
	if ok {
		if alt {
			k.Mod |= input.Alt
		}
		return n, input.KeyEvent(k), nil
	}

	return n, csiSequence(seq), nil
}

// parseSs3 parses a SS3 sequence.
// See https://vt100.net/docs/vt220-rm/chapter4.html#S4.4.4.2
func (d *driver) parseSs3(i int, p []byte, alt bool) (n int, e input.Event, err error) {
	if p[i] == 'O' {
		n++
	}

	i++
	seq := "\x1bO"

	// Scan a GL character
	// A GL character is a single byte in the range 0x21-0x7E
	// See https://vt100.net/docs/vt220-rm/chapter2.html#S2.3.2
	if p[i] < 0x21 || p[i] > 0x7E {
		return n, nil, fmt.Errorf("%w: invalid SS3 sequence: %q", input.ErrUnknownEvent, p[i])
	}
	n++
	seq += string(p[i])

	k, ok := d.table[seq]
	if ok {
		if alt {
			k.Mod |= input.Alt
		}
		return n, input.KeyEvent(k), nil
	}

	return n, ss3Sequence(seq), nil
}

func (d *driver) parseOsc(i int, p []byte, _ bool) (n int, e input.Event, err error) {
	if p[i] == ']' {
		n++
	}

	i++
	seq := "\x1b]"

	// Scan a OSC sequence
	// An OSC sequence is terminated by a BEL, ESC, or ST character
	for ; p[i] != ansi.BEL && p[i] != ansi.ESC && p[i] != ansi.ST; i++ {
		n++
		seq += string(p[i])
	}
	n++
	seq += string(p[i])

	// Check 7-bit ST (string terminator) character
	if len(p) > i+1 && p[i] == ansi.ESC && p[i+1] == '\\' {
		seq += string(p[i+1])
		n++
	}

	return n, oscSequence(seq), nil
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
