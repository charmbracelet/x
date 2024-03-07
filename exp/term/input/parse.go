package input

import (
	"unicode/utf8"

	"github.com/charmbracelet/x/exp/term/ansi"
)

// ParseSequence finds the first recognized event sequence and returns it along
// with its length.
//
// It will return zero and nil no sequence is recognized or when the buffer is
// empty. If a sequence is not supported, an UnknownEvent is returned.
func ParseSequence(buf []byte) (n int, e Event) {
	if len(buf) == 0 {
		return 0, nil
	}

	switch b := buf[0]; b {
	case ansi.ESC:
		if len(buf) == 1 {
			// Escape key
			return 1, KeyEvent{Sym: KeyEscape}
		}

		switch b := buf[1]; b {
		case 'O': // Esc-prefixed SS3
			return parseSs3(buf)
		case 'P': // Esc-prefixed DCS
			return parseDcs(buf)
		case '[': // Esc-prefixed CSI
			return parseCsi(buf)
		case ']': // Esc-prefixed OSC
			return parseOsc(buf)
		case '_': // Esc-prefixed APC
			return parseApc(buf)
		case ansi.ESC:
			if len(buf) == 2 {
				// Double escape key
				return 2, KeyEvent{Sym: KeyEscape, Mod: Alt}
			}
			fallthrough
		default:
			n, e := ParseSequence(buf[1:])
			if k, ok := e.(KeyEvent); ok {
				k.Mod |= Alt
				return n + 1, k
			}

			return n + 1, e
		}
	case ansi.SS3:
		return parseSs3(buf)
	case ansi.DCS:
		return parseDcs(buf)
	case ansi.CSI:
		return parseCsi(buf)
	case ansi.OSC:
		return parseOsc(buf)
	case ansi.APC:
		return parseApc(buf)
	default:
		if b <= ansi.US || b == ansi.DEL || b == ansi.SP {
			return 1, parseCtrl0(b)
		}
		return parseUtf8(buf)
	}
}

func parseCsi(p []byte) (int, Event) {
	var seq []byte
	var i int
	if p[i] == ansi.CSI || p[i] == ansi.ESC {
		seq = append(seq, p[i])
		i++
	}
	if i < len(p) && p[i-1] == ansi.ESC && p[i] == '[' {
		seq = append(seq, p[i])
		i++
	}

	// Initial CSI byte
	var initial byte

	// Scan parameter bytes in the range 0x30-0x3F
	var start, end int // start and end of the parameter bytes
	for j := 0; i < len(p) && p[i] >= 0x30 && p[i] <= 0x3F; i, j = i+1, j+1 {
		if j == 0 {
			initial = p[i]
			start = i
		}
		seq = append(seq, p[i])
	}

	end = i

	// Scan intermediate bytes in the range 0x20-0x2F
	for ; i < len(p) && p[i] >= 0x20 && p[i] <= 0x2F; i++ {
		seq = append(seq, p[i])
	}

	// Final byte
	var final byte

	// Scan final byte in the range 0x40-0x7E
	if i >= len(p) || p[i] < 0x40 || p[i] > 0x7E {
		return len(seq), UnknownEvent(seq)
	}
	// Add the final byte
	final = p[i]
	seq = append(seq, p[i])

	switch initial {
	case '?':
		switch final {
		case 'c':
			// Primary Device Attributes
			params := ansi.Params(p[start:end])
			return len(seq), parsePrimaryDevAttrs(params)
		case 'u':
			// Kitty keyboard flags
			params := ansi.Params(p[start:end])
			if len(params) == 0 {
				return len(seq), UnknownCsiEvent(seq)
			}
			return len(seq), KittyKeyboardEvent(params[0][0])
		}
	case '<':
		switch final {
		case 'm', 'M':
			// Handle SGR mouse
			return len(seq), parseSGRMouseEvent(seq)
		}
	case '=', '>':
		// We don't support any of these yet
		return len(seq), UnknownCsiEvent(seq)
	}

	switch final {
	case 'a':
		return len(seq), KeyEvent{Sym: KeyUp, Mod: Shift}
	case 'b':
		return len(seq), KeyEvent{Sym: KeyDown, Mod: Shift}
	case 'c':
		return len(seq), KeyEvent{Sym: KeyRight, Mod: Shift}
	case 'd':
		return len(seq), KeyEvent{Sym: KeyLeft, Mod: Shift}
	case 'A':
		return len(seq), KeyEvent{Sym: KeyUp}
	case 'B':
		return len(seq), KeyEvent{Sym: KeyDown}
	case 'C':
		return len(seq), KeyEvent{Sym: KeyRight}
	case 'D':
		return len(seq), KeyEvent{Sym: KeyLeft}
	case 'E':
		return len(seq), KeyEvent{Sym: KeyBegin}
	case 'F':
		return len(seq), KeyEvent{Sym: KeyEnd}
	case 'H':
		return len(seq), KeyEvent{Sym: KeyHome}
	case 'P':
		return len(seq), KeyEvent{Sym: KeyF1}
	case 'Q':
		return len(seq), KeyEvent{Sym: KeyF2}
	case 'R':
		return len(seq), KeyEvent{Sym: KeyF3}
	case 'S':
		return len(seq), KeyEvent{Sym: KeyF4}
	case 'Z':
		return len(seq), KeyEvent{Sym: KeyTab, Mod: Shift}
	case 'M':
		// Handle X10 mouse
		if i+3 >= len(p) {
			return len(seq), UnknownCsiEvent(seq)
		}
		return len(seq) + 3, parseX10MouseEvent(append(seq, p[i+1:i+3]...))
	case 'u':
		// Kitty keyboard protocol
		params := ansi.Params(p[start:end])
		if len(params) == 0 {
			return len(seq), UnknownCsiEvent(seq)
		}
		return len(seq), parseKittyKeyboard(params)
	case '~':
		params := ansi.Params(p[start:end])
		if len(params) == 0 {
			return len(seq), UnknownCsiEvent(seq)
		}
		switch params[0][0] {
		case 1:
			return len(seq), KeyEvent{Sym: KeyHome}
		case 2:
			return len(seq), KeyEvent{Sym: KeyInsert}
		case 3:
			return len(seq), KeyEvent{Sym: KeyDelete}
		case 4:
			return len(seq), KeyEvent{Sym: KeyEnd}
		case 5:
			return len(seq), KeyEvent{Sym: KeyPgUp}
		case 6:
			return len(seq), KeyEvent{Sym: KeyPgDown}
		case 7:
			return len(seq), KeyEvent{Sym: KeyHome}
		case 8:
			return len(seq), KeyEvent{Sym: KeyEnd}
		case 11:
			return len(seq), KeyEvent{Sym: KeyF1}
		case 12:
			return len(seq), KeyEvent{Sym: KeyF2}
		case 13:
			return len(seq), KeyEvent{Sym: KeyF3}
		case 14:
			return len(seq), KeyEvent{Sym: KeyF4}
		case 15:
			return len(seq), KeyEvent{Sym: KeyF5}
		case 17:
			return len(seq), KeyEvent{Sym: KeyF6}
		case 18:
			return len(seq), KeyEvent{Sym: KeyF7}
		case 19:
			return len(seq), KeyEvent{Sym: KeyF8}
		case 20:
			return len(seq), KeyEvent{Sym: KeyF9}
		case 21:
			return len(seq), KeyEvent{Sym: KeyF10}
		case 23:
			return len(seq), KeyEvent{Sym: KeyF11}
		case 24:
			return len(seq), KeyEvent{Sym: KeyF12}
		case 25:
			return len(seq), KeyEvent{Sym: KeyF13}
		case 26:
			return len(seq), KeyEvent{Sym: KeyF14}
		case 28:
			return len(seq), KeyEvent{Sym: KeyF15}
		case 29:
			return len(seq), KeyEvent{Sym: KeyF16}
		case 31:
			return len(seq), KeyEvent{Sym: KeyF17}
		case 32:
			return len(seq), KeyEvent{Sym: KeyF18}
		case 33:
			return len(seq), KeyEvent{Sym: KeyF19}
		case 34:
			return len(seq), KeyEvent{Sym: KeyF20}
		case 27:
			// XTerm modifyOtherKeys 2
			if len(params) != 3 {
				return len(seq), UnknownCsiEvent(seq)
			}
			return len(seq), parseXTermModifyOtherKeys(params)
		case 200:
			// bracketed-paste start
			return len(seq), PasteStartEvent{}
		case 201:
			// bracketed-paste end
			return len(seq), PasteEndEvent{}
		}
	}

	return len(seq), UnknownCsiEvent(seq)
}

// parseSs3 parses a SS3 sequence.
// See https://vt100.net/docs/vt220-rm/chapter4.html#S4.4.4.2
func parseSs3(p []byte) (int, Event) {
	var seq []byte
	var i int
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
		return len(seq), UnknownEvent(seq)
	}

	// Add the GL character
	seq = append(seq, p[i])

	switch p[i] {
	case 'A':
		return len(seq), KeyEvent{Sym: KeyUp}
	case 'B':
		return len(seq), KeyEvent{Sym: KeyDown}
	case 'C':
		return len(seq), KeyEvent{Sym: KeyRight}
	case 'D':
		return len(seq), KeyEvent{Sym: KeyLeft}
	case 'F':
		return len(seq), KeyEvent{Sym: KeyEnd}
	case 'H':
		return len(seq), KeyEvent{Sym: KeyHome}
	case 'P':
		return len(seq), KeyEvent{Sym: KeyF1}
	case 'Q':
		return len(seq), KeyEvent{Sym: KeyF2}
	case 'R':
		return len(seq), KeyEvent{Sym: KeyF3}
	case 'S':
		return len(seq), KeyEvent{Sym: KeyF4}
	case 'a':
		return len(seq), KeyEvent{Sym: KeyUp, Mod: Shift}
	case 'b':
		return len(seq), KeyEvent{Sym: KeyDown, Mod: Shift}
	case 'c':
		return len(seq), KeyEvent{Sym: KeyRight, Mod: Shift}
	case 'd':
		return len(seq), KeyEvent{Sym: KeyLeft, Mod: Shift}
	case 'M':
		return len(seq), KeyEvent{Sym: KeyKpEnter}
	case 'X':
		return len(seq), KeyEvent{Sym: KeyKpEqual}
	case 'j':
		return len(seq), KeyEvent{Sym: KeyKpMul}
	case 'k':
		return len(seq), KeyEvent{Sym: KeyKpPlus}
	case 'l':
		return len(seq), KeyEvent{Sym: KeyKpComma}
	case 'm':
		return len(seq), KeyEvent{Sym: KeyKpMinus}
	case 'n':
		return len(seq), KeyEvent{Sym: KeyKpPeriod}
	case 'o':
		return len(seq), KeyEvent{Sym: KeyKpDiv}
	case 'p':
		return len(seq), KeyEvent{Sym: KeyKp0}
	case 'q':
		return len(seq), KeyEvent{Sym: KeyKp1}
	case 'r':
		return len(seq), KeyEvent{Sym: KeyKp2}
	case 's':
		return len(seq), KeyEvent{Sym: KeyKp3}
	case 't':
		return len(seq), KeyEvent{Sym: KeyKp4}
	case 'u':
		return len(seq), KeyEvent{Sym: KeyKp5}
	case 'v':
		return len(seq), KeyEvent{Sym: KeyKp6}
	case 'w':
		return len(seq), KeyEvent{Sym: KeyKp7}
	case 'x':
		return len(seq), KeyEvent{Sym: KeyKp8}
	case 'y':
		return len(seq), KeyEvent{Sym: KeyKp9}
	}

	return len(seq), UnknownSs3Event(seq)
}

func parseOsc(p []byte) (int, Event) {
	var seq []byte
	var i int
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
	var start, end int
	var dstart, dend int
	for j := 0; i < len(p) && p[i] != ansi.BEL && p[i] != ansi.ESC && p[i] != ansi.ST; i, j = i+1, j+1 {
		if end != 0 && dstart == 0 {
			dstart = i
		}
		if j == 0 {
			start = i
		}
		if p[i] == ';' {
			end = i
		}
		seq = append(seq, p[i])
	}

	dend = i

	if i >= len(p) {
		return len(seq), UnknownEvent(seq)
	}
	seq = append(seq, p[i])

	// Check 7-bit ST (string terminator) character
	if len(p) > i+1 && p[i] == ansi.ESC && p[i+1] == '\\' {
		i++
		seq = append(seq, p[i])
	}

	if end <= start || dend <= dstart {
		return len(seq), UnknownOscEvent(seq)
	}

	data := string(p[dstart:dend])
	switch string(seq[start:end]) {
	case "10":
		return len(seq), ForegroundColorEvent{xParseColor(data)}
	case "11":
		return len(seq), BackgroundColorEvent{xParseColor(data)}
	case "12":
		return len(seq), CursorColorEvent{xParseColor(data)}
	}

	return len(seq), UnknownOscEvent(seq)
}

// parseCtrl parses a control sequence that gets terminated by a ST character.
func parseCtrl(intro8, intro7 byte) func([]byte) (int, Event) {
	return func(p []byte) (int, Event) {
		var seq []byte
		var i int
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

func parseDcs(p []byte) (int, Event) {
	// DCS sequences are introduced by DCS (0x90) or ESC P (0x1b 0x50)
	return parseCtrl(ansi.DCS, 'P')(p)
}

func parseApc(p []byte) (int, Event) {
	// APC sequences are introduced by APC (0x9f) or ESC _ (0x1b 0x5f)
	return parseCtrl(ansi.APC, '_')(p)
}

func parseUtf8(p []byte) (int, Event) {
	r, rw := utf8.DecodeRune(p)
	if r == utf8.RuneError || r <= ansi.US || r == ansi.DEL || r == ansi.SP {
		return 0, nil
	}
	return rw, KeyEvent{Rune: r}
}

func parseCtrl0(b byte) Event {
	switch b {
	case ansi.NUL:
		return KeyEvent{Sym: KeySpace, Mod: Ctrl}
	case ansi.SOH:
		return KeyEvent{Rune: 'a', Mod: Ctrl}
	case ansi.STX:
		return KeyEvent{Rune: 'b', Mod: Ctrl}
	case ansi.ETX:
		return KeyEvent{Rune: 'c', Mod: Ctrl}
	case ansi.EOT:
		return KeyEvent{Rune: 'd', Mod: Ctrl}
	case ansi.ENQ:
		return KeyEvent{Rune: 'e', Mod: Ctrl}
	case ansi.ACK:
		return KeyEvent{Rune: 'f', Mod: Ctrl}
	case ansi.BEL:
		return KeyEvent{Rune: 'g', Mod: Ctrl}
	case ansi.BS:
		return KeyEvent{Rune: 'h', Mod: Ctrl}
	case ansi.HT:
		return KeyEvent{Sym: KeyTab}
	case ansi.LF:
		return KeyEvent{Rune: 'j', Mod: Ctrl}
	case ansi.VT:
		return KeyEvent{Rune: 'k', Mod: Ctrl}
	case ansi.FF:
		return KeyEvent{Rune: 'l', Mod: Ctrl}
	case ansi.CR:
		return KeyEvent{Sym: KeyEnter}
	case ansi.SO:
		return KeyEvent{Rune: 'n', Mod: Ctrl}
	case ansi.SI:
		return KeyEvent{Rune: 'o', Mod: Ctrl}
	case ansi.DLE:
		return KeyEvent{Rune: 'p', Mod: Ctrl}
	case ansi.DC1:
		return KeyEvent{Rune: 'q', Mod: Ctrl}
	case ansi.DC2:
		return KeyEvent{Rune: 'r', Mod: Ctrl}
	case ansi.DC3:
		return KeyEvent{Rune: 's', Mod: Ctrl}
	case ansi.DC4:
		return KeyEvent{Rune: 't', Mod: Ctrl}
	case ansi.NAK:
		return KeyEvent{Rune: 'u', Mod: Ctrl}
	case ansi.SYN:
		return KeyEvent{Rune: 'v', Mod: Ctrl}
	case ansi.ETB:
		return KeyEvent{Rune: 'w', Mod: Ctrl}
	case ansi.CAN:
		return KeyEvent{Rune: 'x', Mod: Ctrl}
	case ansi.EM:
		return KeyEvent{Rune: 'y', Mod: Ctrl}
	case ansi.SUB:
		return KeyEvent{Rune: 'z', Mod: Ctrl}
	case ansi.ESC:
		return KeyEvent{Sym: KeyEscape}
	case ansi.FS:
		return KeyEvent{Rune: '\\', Mod: Ctrl}
	case ansi.GS:
		return KeyEvent{Rune: ']', Mod: Ctrl}
	case ansi.RS:
		return KeyEvent{Rune: '^', Mod: Ctrl}
	case ansi.US:
		return KeyEvent{Rune: '_', Mod: Ctrl}
	case ansi.SP:
		return KeyEvent{Sym: KeySpace, Rune: ' '}
	case ansi.DEL:
		return KeyEvent{Sym: KeyBackspace}
	default:
		return UnknownEvent(string(b))
	}
}
