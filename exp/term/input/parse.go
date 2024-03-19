package input

import (
	"strconv"
	"unicode/utf8"

	"github.com/charmbracelet/x/exp/term/ansi"
	"github.com/erikgeiser/coninput"
)

// EventParser represents a parser for input events.
type EventParser struct {
	table map[string]Key
	flags int
}

// NewEventParser returns a new EventParser.
func NewEventParser(term string, flags int) EventParser {
	t := registerKeys(flags, term)
	return EventParser{
		table: t,
		flags: flags,
	}
}

// LookupSequence looks up a key sequence in the parser's table and returns the
// corresponding key.
func (p EventParser) LookupSequence(seq string) (Key, bool) {
	k, ok := p.table[seq]
	return k, ok
}

// ParseSequence finds the first recognized event sequence and returns it along
// with its length.
//
// It will return zero and nil no sequence is recognized or when the buffer is
// empty. If a sequence is not supported, an UnknownEvent is returned.
func (p EventParser) ParseSequence(buf []byte) (n int, e Event) {
	if len(buf) == 0 {
		return 0, nil
	}

	switch b := buf[0]; b {
	case ansi.ESC:
		if len(buf) == 1 {
			// Escape key
			return 1, KeyDownEvent{Sym: KeyEscape}
		}

		switch b := buf[1]; b {
		case 'O': // Esc-prefixed SS3
			return p.parseSs3(buf)
		case 'P': // Esc-prefixed DCS
			return p.parseDcs(buf)
		case '[': // Esc-prefixed CSI
			return p.parseCsi(buf)
		case ']': // Esc-prefixed OSC
			return p.parseOsc(buf)
		case '_': // Esc-prefixed APC
			return p.parseApc(buf)
		case ansi.ESC:
			if len(buf) == 2 {
				// Double escape key
				return 2, KeyDownEvent{Sym: KeyEscape, Mod: Alt}
			}
			fallthrough
		default:
			n, e := p.ParseSequence(buf[1:])
			if k, ok := e.(KeyDownEvent); ok {
				if !k.Mod.IsAlt() {
					k.Mod |= Alt
					return n + 1, k
				}
			}

			return 1, KeyDownEvent{Sym: KeyEscape}
		}
	case ansi.SS3:
		return p.parseSs3(buf)
	case ansi.DCS:
		return p.parseDcs(buf)
	case ansi.CSI:
		return p.parseCsi(buf)
	case ansi.OSC:
		return p.parseOsc(buf)
	case ansi.APC:
		return p.parseApc(buf)
	default:
		if b <= ansi.US || b == ansi.DEL || b == ansi.SP {
			return 1, p.parseCtrl0(b)
		}
		return p.parseUtf8(buf)
	}
}

func (p *EventParser) parseCsi(b []byte) (int, Event) {
	var seq []byte
	var i int
	if b[i] == ansi.CSI || b[i] == ansi.ESC {
		seq = append(seq, b[i])
		i++
	}
	if i < len(b) && b[i-1] == ansi.ESC && b[i] == '[' {
		seq = append(seq, b[i])
		i++
	}

	// Initial CSI byte
	var initial byte

	// Scan parameter bytes in the range 0x30-0x3F
	start := -11 // start of the parameter bytes
	for j := 0; i < len(b) && b[i] >= 0x30 && b[i] <= 0x3F; i, j = i+1, j+1 {
		if j == 0 {
			initial = b[i]
			start = i
		}
		seq = append(seq, b[i])
	}

	end := i

	var params []byte
	if start > 0 && end > start {
		params = b[start:end]
	}

	// Scan intermediate bytes in the range 0x20-0x2F
	for ; i < len(b) && b[i] >= 0x20 && b[i] <= 0x2F; i++ {
		seq = append(seq, b[i])
	}

	// Final byte
	var final byte

	// Scan final byte in the range 0x40-0x7E
	if i >= len(b) || b[i] < 0x40 || b[i] > 0x7E {
		// Special case for URxvt keys
		// CSI <number> $ is an invalid sequence, but URxvt uses it for
		// shift modified keys.
		if seq[i-1] == '$' {
			n, ev := p.parseCsi(append(seq[:i-1], '~'))
			if k, ok := ev.(KeyDownEvent); ok {
				k.Mod |= Shift
				return n, k
			}
		}
		return len(seq), UnknownEvent(seq)
	}
	// Add the final byte
	final = b[i]
	seq = append(seq, b[i])
	i++

	switch initial {
	case '?':
		switch final {
		case 'c':
			// Primary Device Attributes
			params := ansi.Params(params)
			return len(seq), parsePrimaryDevAttrs(params)
		case 'u':
			// Kitty keyboard flags
			params := ansi.Params(params)
			if len(params) == 0 {
				return len(seq), UnknownCsiEvent(seq)
			}
			return len(seq), KittyKeyboardEvent(params[0][0])
		default:
			return len(seq), UnknownCsiEvent(seq)
		}
	case '<':
		switch final {
		case 'm', 'M':
			// Handle SGR mouse
			return len(seq), parseSGRMouseEvent(seq)
		default:
			return len(seq), UnknownCsiEvent(seq)
		}
	case '>':
		switch final {
		case 'm':
			// XTerm modifyOtherKeys
			params := ansi.Params(params)
			if len(params) != 2 || params[0][0] != 4 {
				return len(seq), UnknownCsiEvent(seq)
			}

			return len(seq), ModifyOtherKeysEvent(params[1][0])
		default:
			return len(seq), UnknownCsiEvent(seq)
		}
	case '=':
		// We don't support any of these yet
		return len(seq), UnknownCsiEvent(seq)
	}

	switch final {
	case 'a', 'b', 'c', 'd', 'A', 'B', 'C', 'D', 'E', 'F', 'H', 'P', 'Q', 'R', 'S', 'Z':
		var k KeyDownEvent
		switch final {
		case 'a', 'b', 'c', 'd':
			k = KeyDownEvent{Sym: KeyUp + KeySym(final-'a'), Mod: Shift}
		case 'A', 'B', 'C', 'D':
			k = KeyDownEvent{Sym: KeyUp + KeySym(final-'A')}
		case 'E':
			k = KeyDownEvent{Sym: KeyBegin}
		case 'F':
			k = KeyDownEvent{Sym: KeyEnd}
		case 'H':
			k = KeyDownEvent{Sym: KeyHome}
		case 'P', 'Q', 'R', 'S':
			k = KeyDownEvent{Sym: KeyF1 + KeySym(final-'P')}
		case 'Z':
			k = KeyDownEvent{Sym: KeyTab, Mod: Shift}
		}
		if len(params) > 0 {
			params := ansi.Params(params)
			// CSI 1 ; <modifiers> A
			if len(params) > 1 {
				k.Mod |= KeyMod(params[1][0] - 1)
			}
		}
		return len(seq), k
	case 'M':
		// Handle X10 mouse
		if i+3 > len(b) {
			return len(seq), UnknownCsiEvent(seq)
		}
		return len(seq) + 3, parseX10MouseEvent(append(seq, b[i:i+3]...))
	case 'u':
		// Kitty keyboard protocol
		params := ansi.Params(params)
		if len(params) == 0 {
			return len(seq), UnknownCsiEvent(seq)
		}
		return len(seq), parseKittyKeyboard(params)
	case '_':
		// Win32 Input Mode
		params := ansi.Params(params)
		if len(params) != 6 {
			return len(seq), UnknownCsiEvent(seq)
		}

		rc := uint16(params[5][0])
		if rc == 0 {
			rc = 1
		}

		event := parseWin32InputKeyEvent(
			coninput.VirtualKeyCode(params[0][0]),  // Vk wVirtualKeyCode
			coninput.VirtualKeyCode(params[1][0]),  // Sc wVirtualScanCode
			rune(params[2][0]),                     // Uc UnicodeChar
			params[3][0] == 1,                      // Kd bKeyDown
			coninput.ControlKeyState(params[4][0]), // Cs dwControlKeyState
			rc,                                     // Rc wRepeatCount
		)

		if event == nil {
			return len(seq), UnknownCsiEvent(seq)
		}

		return len(seq), event
	case '@', '^', '~':
		params := ansi.Params(params)
		if len(params) == 0 {
			return len(seq), UnknownCsiEvent(seq)
		}

		switch final {
		case '~':
			switch params[0][0] {
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

		switch params[0][0] {
		case 1, 2, 3, 4, 5, 6, 7, 8:
			fallthrough
		case 11, 12, 13, 14, 15:
			fallthrough
		case 17, 18, 19, 20, 21, 23, 24, 25, 26:
			fallthrough
		case 28, 29, 31, 32, 33, 34:
			var k KeyDownEvent
			v := params[0][0]
			switch v {
			case 1:
				if p.flags&FlagFind != 0 {
					k = KeyDownEvent{Sym: KeyFind}
				} else {
					k = KeyDownEvent{Sym: KeyHome}
				}
			case 2:
				k = KeyDownEvent{Sym: KeyInsert}
			case 3:
				k = KeyDownEvent{Sym: KeyDelete}
			case 4:
				if p.flags&FlagSelect != 0 {
					k = KeyDownEvent{Sym: KeySelect}
				} else {
					k = KeyDownEvent{Sym: KeyEnd}
				}
			case 5:
				k = KeyDownEvent{Sym: KeyPgUp}
			case 6:
				k = KeyDownEvent{Sym: KeyPgDown}
			case 7:
				k = KeyDownEvent{Sym: KeyHome}
			case 8:
				k = KeyDownEvent{Sym: KeyEnd}
			case 11, 12, 13, 14, 15:
				k = KeyDownEvent{Sym: KeyF1 + KeySym(v-11)}
			case 17, 18, 19, 20, 21:
				k = KeyDownEvent{Sym: KeyF6 + KeySym(v-17)}
			case 23, 24, 25, 26:
				k = KeyDownEvent{Sym: KeyF11 + KeySym(v-23)}
			case 28, 29:
				k = KeyDownEvent{Sym: KeyF15 + KeySym(v-28)}
			case 31, 32, 33, 34:
				k = KeyDownEvent{Sym: KeyF17 + KeySym(v-31)}
			}

			// modifiers
			if len(params) > 1 {
				k.Mod |= KeyMod(params[1][0] - 1)
			}

			// Handle URxvt weird keys
			switch final {
			case '^':
				k.Mod |= Ctrl
			case '@':
				k.Mod |= Ctrl | Shift
			}

			return len(seq), k
		}
	}
	return len(seq), UnknownCsiEvent(seq)
}

// parseSs3 parses a SS3 sequence.
// See https://vt100.net/docs/vt220-rm/chapter4.html#S4.4.4.2
func (p *EventParser) parseSs3(b []byte) (int, Event) {
	var seq []byte
	var i int
	if b[i] == ansi.SS3 || b[i] == ansi.ESC {
		seq = append(seq, b[i])
		i++
	}
	if i < len(b) && b[i-1] == ansi.ESC && b[i] == 'O' {
		seq = append(seq, b[i])
		i++
	}

	// Scan numbers from 0-9
	start := -1
	for ; i < len(b) && b[i] >= 0x30 && b[i] <= 0x39; i++ {
		if start == -1 {
			start = i
		}
		seq = append(seq, b[i])
	}
	end := i

	var mod []byte
	if start > 0 && end > start {
		mod = b[start:end]
	}

	// Scan a GL character
	// A GL character is a single byte in the range 0x21-0x7E
	// See https://vt100.net/docs/vt220-rm/chapter2.html#S2.3.2
	if i >= len(b) || b[i] < 0x21 || b[i] > 0x7E {
		return len(seq), UnknownEvent(seq)
	}

	// Add the GL character
	seq = append(seq, b[i])

	var k KeyDownEvent
	switch b[i] {
	case 'a', 'b', 'c', 'd':
		k = KeyDownEvent{Sym: KeyUp + KeySym(b[i]-'a'), Mod: Ctrl}
	case 'A', 'B', 'C', 'D':
		k = KeyDownEvent{Sym: KeyUp + KeySym(b[i]-'A')}
	case 'E':
		k = KeyDownEvent{Sym: KeyBegin}
	case 'F':
		k = KeyDownEvent{Sym: KeyEnd}
	case 'H':
		k = KeyDownEvent{Sym: KeyHome}
	case 'P', 'Q', 'R', 'S':
		k = KeyDownEvent{Sym: KeyF1 + KeySym(b[i]-'P')}
	case 'M':
		k = KeyDownEvent{Sym: KeyKpEnter}
	case 'X':
		k = KeyDownEvent{Sym: KeyKpEqual}
	case 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y':
		k = KeyDownEvent{Sym: KeyKpMultiply + KeySym(b[i]-'j')}
	default:
		return len(seq), UnknownSs3Event(seq)
	}

	// Handle weird SS3 <modifier> Func
	if len(mod) > 0 {
		m, err := strconv.Atoi(string(mod))
		if err == nil {
			k.Mod |= KeyMod(m - 1)
		}
	}

	return len(seq), k
}

func (p *EventParser) parseOsc(b []byte) (int, Event) {
	var seq []byte
	var i int
	if b[i] == ansi.OSC || b[i] == ansi.ESC {
		seq = append(seq, b[i])
		i++
	}
	if i < len(b) && b[i-1] == ansi.ESC && b[i] == ']' {
		seq = append(seq, b[i])
		i++
	}

	// Scan a OSC sequence
	// An OSC sequence is terminated by a BEL, ESC, or ST character
	var start, end int
	var dstart, dend int
	for j := 0; i < len(b) && b[i] != ansi.BEL && b[i] != ansi.ESC && b[i] != ansi.ST; i, j = i+1, j+1 {
		if end != 0 && dstart == 0 {
			dstart = i
		}
		if j == 0 {
			start = i
		}
		if b[i] == ';' {
			end = i
		}
		seq = append(seq, b[i])
	}

	dend = i

	if i >= len(b) {
		return len(seq), UnknownEvent(seq)
	}
	seq = append(seq, b[i])

	// Check 7-bit ST (string terminator) character
	if len(b) > i+1 && b[i] == ansi.ESC && b[i+1] == '\\' {
		i++
		seq = append(seq, b[i])
	}

	if end <= start || dend <= dstart {
		return len(seq), UnknownOscEvent(seq)
	}

	data := string(b[dstart:dend])
	switch string(seq[start:end]) {
	case "10":
		return len(seq), ForegroundColorEvent{xParseColor(data)}
	case "11":
		return len(seq), BackgroundColorEvent{xParseColor(data)}
	case "12":
		return len(seq), CursorColorEvent{xParseColor(data)}
	default:
		return len(seq), UnknownOscEvent(seq)
	}
}

// parseCtrl parses a control sequence that gets terminated by a ST character.
func (p *EventParser) parseCtrl(intro8, intro7 byte) func([]byte) (int, Event) {
	return func(b []byte) (int, Event) {
		var seq []byte
		var i int
		if b[i] == intro8 || b[i] == ansi.ESC {
			seq = append(seq, b[i])
			i++
		}
		if i < len(b) && b[i-1] == ansi.ESC && b[i] == intro7 {
			seq = append(seq, b[i])
			i++
		}

		// Scan control sequence
		// Most common control sequence is terminated by a ST character
		// ST is a 7-bit string terminator character is (ESC \)
		for ; i < len(b) && b[i] != ansi.ST && b[i] != ansi.ESC; i++ {
			seq = append(seq, b[i])
		}

		if i >= len(b) {
			switch intro8 {
			case ansi.DCS:
				return len(seq), UnknownDcsEvent(seq)
			case ansi.APC:
				return len(seq), UnknownApcEvent(seq)
			default:
				return len(seq), UnknownEvent(seq)
			}
		}
		seq = append(seq, b[i])

		// Check 7-bit ST (string terminator) character
		if len(b) > i+1 && b[i] == ansi.ESC && b[i+1] == '\\' {
			i++
			seq = append(seq, b[i])
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

func (p *EventParser) parseDcs(b []byte) (int, Event) {
	// DCS sequences are introduced by DCS (0x90) or ESC P (0x1b 0x50)
	var seq []byte
	var i int
	if b[i] == ansi.DCS || b[i] == ansi.ESC {
		seq = append(seq, b[i])
		i++
	}
	if i < len(b) && b[i-1] == ansi.ESC && b[i] == 'P' {
		seq = append(seq, b[i])
		i++
	}

	// Scan parameter bytes in the range 0x30-0x3F
	var start, end int // start and end of the parameter bytes
	for j := 0; i < len(b) && b[i] >= 0x30 && b[i] <= 0x3F; i, j = i+1, j+1 {
		if j == 0 {
			start = i
		}
		seq = append(seq, b[i])
	}

	end = i

	// Scan intermediate bytes in the range 0x20-0x2F
	var istart, iend int
	for j := 0; i < len(b) && b[i] >= 0x20 && b[i] <= 0x2F; i, j = i+1, j+1 {
		if j == 0 {
			istart = i
		}
		seq = append(seq, b[i])
	}

	iend = i

	// Final byte
	var final byte

	// Scan final byte in the range 0x40-0x7E
	if i >= len(b) || b[i] < 0x40 || b[i] > 0x7E {
		return len(seq), UnknownEvent(seq)
	}
	// Add the final byte
	final = b[i]
	seq = append(seq, b[i])

	if i+1 >= len(b) {
		return len(seq), UnknownEvent(seq)
	}

	// Collect data bytes until a ST character is found
	// data bytes are in the range of 0x08-0x0D and 0x20-0x7F
	// but we don't care about the actual values for now
	var data []byte
	for i++; i < len(b) && b[i] != ansi.ST && b[i] != ansi.ESC; i++ {
		data = append(data, b[i])
		seq = append(seq, b[i])
	}

	if i >= len(b) {
		return len(seq), UnknownEvent(seq)
	}

	seq = append(seq, b[i])

	// Check 7-bit ST (string terminator) character
	if len(b) > i+1 && b[i] == ansi.ESC && b[i+1] == '\\' {
		i++
		seq = append(seq, b[i])
	}

	switch final {
	case 'r':
		inters := b[istart:iend] // intermediates
		if len(inters) == 0 {
			return len(seq), UnknownDcsEvent(seq)
		}
		switch inters[0] {
		case '+':
			// XTGETTCAP responses
			params := ansi.Params(b[start:end])
			if len(params) == 0 {
				return len(seq), UnknownDcsEvent(seq)
			}

			switch params[0][0] {
			case 0, 1:
				tc := parseTermcap(data)
				// XXX: some terminals like KiTTY report invalid responses with
				// their queries i.e. sending a query for "Tc" using "\x1bP+q5463\x1b\\"
				// returns "\x1bP0+r5463\x1b\\".
				// The specs says that invalid responses should be in the form of
				// DCS 0 + r ST "\x1bP0+r\x1b\\"
				//
				// See: https://invisible-island.net/xterm/ctlseqs/ctlseqs.html#h3-Operating-System-Commands
				tc.IsValid = params[0][0] == 1
				return len(seq), tc
			}
		}
	}

	return len(seq), UnknownDcsEvent(seq)
}

func (p *EventParser) parseApc(b []byte) (int, Event) {
	// APC sequences are introduced by APC (0x9f) or ESC _ (0x1b 0x5f)
	return p.parseCtrl(ansi.APC, '_')(b)
}

func (p *EventParser) parseUtf8(b []byte) (int, Event) {
	r, rw := utf8.DecodeRune(b)
	if r == utf8.RuneError || r <= ansi.US || r == ansi.DEL || r == ansi.SP {
		return 0, nil
	}
	return rw, KeyDownEvent{Rune: r}
}

func (p *EventParser) parseCtrl0(b byte) Event {
	switch b {
	case ansi.NUL:
		if p.flags&FlagCtrlAt != 0 {
			return KeyDownEvent{Rune: '@', Mod: Ctrl}
		}
		return KeyDownEvent{Rune: ' ', Sym: KeySpace, Mod: Ctrl}
	case ansi.HT:
		if p.flags&FlagCtrlI != 0 {
			return KeyDownEvent{Rune: 'i', Mod: Ctrl}
		}
		return KeyDownEvent{Sym: KeyTab}
	case ansi.CR:
		if p.flags&FlagCtrlM != 0 {
			return KeyDownEvent{Rune: 'm', Mod: Ctrl}
		}
		return KeyDownEvent{Sym: KeyEnter}
	case ansi.ESC:
		if p.flags&FlagCtrlOpenBracket != 0 {
			return KeyDownEvent{Rune: '[', Mod: Ctrl}
		}
		return KeyDownEvent{Sym: KeyEscape}
	case ansi.DEL:
		if p.flags&FlagBackspace != 0 {
			return KeyDownEvent{Sym: KeyDelete}
		}
		return KeyDownEvent{Sym: KeyBackspace}
	case ansi.SP:
		return KeyDownEvent{Sym: KeySpace, Rune: ' '}
	default:
		if b >= ansi.SOH && b <= ansi.SUB {
			return KeyDownEvent{Rune: rune(b - ansi.SOH + 'a'), Mod: Ctrl}
		} else if b >= ansi.FS && b <= ansi.US {
			return KeyDownEvent{Rune: rune(b - ansi.FS + '\\'), Mod: Ctrl}
		}
		return UnknownEvent(b)
	}
}
