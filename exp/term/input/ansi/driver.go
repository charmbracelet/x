package ansi

import (
	"bufio"
	"fmt"
	"unicode/utf8"

	"github.com/charmbracelet/x/exp/term/ansi"
	"github.com/charmbracelet/x/exp/term/input"
)

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
	r     *bufio.Reader
	table map[string]input.Key
	flags int
}

var _ input.Driver = &driver{}

// NewDriver returns a new ANSI input driver.
// This driver uses ANSI control codes compatible with VT100/VT200 terminals.
func NewDriver(r *bufio.Reader, flags int) input.Driver {
	d := &driver{
		r:     bufio.NewReaderSize(r, 256),
		flags: flags,
	}
	// Populate the key sequences table.
	d.registerKeys(flags)
	return d
}

// ReadInput implements input.Driver.
func (d *driver) ReadInput() (int, input.Event, error) {
	n, e, err := d.PeekInput()
	if err != nil {
		return n, e, err
	}

	// Consume the event
	p := make([]byte, n)
	if _, err := d.r.Read(p); err != nil {
		return 0, nil, err
	}

	return n, e, nil
}

const esc = string(ansi.ESC)

// PeekInput implements input.Driver.
func (d *driver) PeekInput() (int, input.Event, error) {
	p, err := d.r.Peek(1)
	if err != nil {
		return 0, nil, err
	}

	// The number of bytes buffered.
	nb := d.r.Buffered()
	// Peek more bytes if needed.
	if nb > len(p) {
		p, err = d.r.Peek(nb)
		if err != nil {
			return 0, nil, err
		}
	}

	for {
		var alt bool
		i := 0
		b := p[i]

	begin:
		switch b {
		case ansi.ESC:
			if nb == 1 {
				return 1, input.KeyEvent(d.table[esc]), nil
			}

			i++ // i = 1
			switch p[i] {
			case 'O': // Esc-prefixed SS3
				return d.parseSs3(i+1, p)
			case 'P': // Esc-prefixed DCS
			case '[': // Esc-prefixed CSI
				return d.parseCsi(i+1, p)
			case ']': // Esc-prefixed OSC
			}

			alt = true
			b = p[i]

			goto begin
		case ansi.SS3:
			return d.parseSs3(i+1, p)
		case ansi.DCS:
		case ansi.CSI:
			return d.parseCsi(i+1, p)
		case ansi.OSC:
		}

		// Single byte control code or printable ASCII/UTF-8
		if b <= ansi.US || b == ansi.DEL || b == ansi.SP {
			k := input.KeyEvent(d.table[string(b)])
			l := 1
			if alt {
				k.Mod |= input.Alt
				l++
			}
			return l, k, nil
		} else if utf8.RuneStart(b) { // Printable ASCII/UTF-8
			ul := utf8ByteLen(b)
			if ul == -1 || ul > nb {
				return 0, nil, fmt.Errorf("invalid UTF-8 sequence: %x", p)
			}

			r := rune(b)
			if ul > 1 {
				r, _ = utf8.DecodeRune(p[i : i+ul])
			}

			k := input.KeyEvent{Rune: r}
			if alt {
				k.Mod |= input.Alt
				ul++
			}

			return ul, k, nil
		}

		return nb, nil, input.ErrUnknownEvent
	}
}

func (d *driver) parseCsi(i int, p []byte) (int, input.Event, error) {
	start := i
	seq := "\x1b["

	// Scan parameter bytes in the range 0x30-0x3F
	for ; p[i] >= 0x30 && p[i] <= 0x3F; i++ {
		seq += string(p[i])
	}
	// Scan intermediate bytes in the range 0x20-0x2F
	for ; p[i] >= 0x20 && p[i] <= 0x2F; i++ {
		seq += string(p[i])
	}
	// Scan final byte in the range 0x40-0x7E
	if p[i] < 0x40 || p[i] > 0x7E {
		return i, nil, fmt.Errorf("%w: invalid CSI sequence: %q", input.ErrUnknownEvent, p[start:i+1])
	}
	seq += string(p[i])

	k, ok := d.table[seq]
	if ok {
		return i + 1, input.KeyEvent(k), nil
	}

	return i + 1, nil, fmt.Errorf("%w: unknown CSI sequence: %q (%q)", input.ErrUnknownEvent, seq, p[start:i+1])
}

// parseSs3 parses a SS3 sequence.
// See https://vt100.net/docs/vt220-rm/chapter4.html#S4.4.4.2
func (d *driver) parseSs3(i int, p []byte) (int, input.Event, error) {
	seq := "\x1bO"

	// Scan a GL character
	// A GL character is a single byte in the range 0x21-0x7E
	// See https://vt100.net/docs/vt220-rm/chapter2.html#S2.3.2
	if p[i] < 0x21 || p[i] > 0x7E {
		return i, nil, fmt.Errorf("%w: invalid SS3 sequence: %q", input.ErrUnknownEvent, p[i])
	}
	seq += string(p[i])

	k, ok := d.table[seq]
	if ok {
		return i + 1, input.KeyEvent(k), nil
	}

	return i + 1, nil, fmt.Errorf("%w: unknown SS3 sequence: %q", input.ErrUnknownEvent, seq)
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
