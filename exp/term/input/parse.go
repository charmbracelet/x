package input

import (
	"unicode/utf8"

	"github.com/charmbracelet/x/exp/term/ansi"
	"github.com/erikgeiser/coninput"
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
			return 1, KeyDownEvent{Sym: KeyEscape}
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
				return 2, KeyDownEvent{Sym: KeyEscape, Mod: Alt}
			}
			fallthrough
		default:
			n, e := ParseSequence(buf[1:])
			if k, ok := e.(KeyDownEvent); ok {
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
	start, end := -1, -1 // start and end of the parameter bytes
	for j := 0; i < len(p) && p[i] >= 0x30 && p[i] <= 0x3F; i, j = i+1, j+1 {
		if j == 0 {
			initial = p[i]
			start = i
		}
		seq = append(seq, p[i])
	}

	end = i

	var params []byte
	if start > 0 && end > start {
		params = p[start:end]
	}

	// Scan intermediate bytes in the range 0x20-0x2F
	for ; i < len(p) && p[i] >= 0x20 && p[i] <= 0x2F; i++ {
		seq = append(seq, p[i])
	}

	// Final byte
	var final byte

	// Scan final byte in the range 0x40-0x7E
	if i >= len(p) || p[i] < 0x40 || p[i] > 0x7E {
		// Special case for URxvt keys
		// CSI <number> $ is an invalid sequence, but URxvt uses it for
		// shift modified keys.
		if p[i-1] == '$' {
			num := string(params)
			if k, ok := csiTildeKeys[num]; ok {
				k.Mod |= Shift
				return len(seq), KeyDownEvent(k)
			} else if k, ok := csiDollarKeys[num]; ok {
				return len(seq), KeyDownEvent(k)
			}
		}
		return len(seq), UnknownEvent(seq)
	}
	// Add the final byte
	final = p[i]
	seq = append(seq, p[i])
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
	case 'a', 'b', 'c', 'd':
		fallthrough
	case 'A', 'B', 'C', 'D', 'E', 'F', 'H', 'P', 'Q', 'R', 'S':
		fallthrough
	case 'Z':
		params := ansi.Params(params)
		return len(seq), parseCsiFunc(params, seq)
	case 'M':
		// Handle X10 mouse
		if i+3 > len(p) {
			return len(seq), UnknownCsiEvent(seq)
		}
		return len(seq) + 3, parseX10MouseEvent(append(seq, p[i:i+3]...))
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
	case '~':
		params := ansi.Params(params)
		if len(params) == 0 {
			return len(seq), UnknownCsiEvent(seq)
		}
		switch params[0][0] {
		case 1, 2, 3, 4, 5, 6, 7, 8:
			fallthrough
		case 11, 12, 13, 14, 15:
			fallthrough
		case 17, 18, 19, 20, 21, 23, 24, 25, 26:
			fallthrough
		case 28, 29, 31, 32, 33, 34:
			return len(seq), parseCsiTilde(params, seq)
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
		default:
			return len(seq), UnknownCsiEvent(seq)
		}
	case '^':
		params := ansi.Params(params)
		return len(seq), parseCsiCarat(params, seq)
	case '@':
		params := ansi.Params(params)
		return len(seq), parseCsiAt(params, seq)
	default:
		return len(seq), UnknownCsiEvent(seq)
	}
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

	// Scan numbers from 0-9
	start, end := -1, -1
	for ; i < len(p) && p[i] >= 0x30 && p[i] <= 0x39; i++ {
		if start == -1 {
			start = i
		}
		seq = append(seq, p[i])
	}
	end = i

	var mod []byte
	if start > 0 && end > start {
		mod = p[start:end]
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
	case 'a', 'b', 'c', 'd':
		fallthrough
	case 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z':
		fallthrough
	case 'A', 'B', 'C', 'D', 'E', 'F', 'H', 'P', 'Q', 'R', 'S', 'M', 'X':
		return len(seq), parseSs3Func(mod, seq)
	default:
		return len(seq), UnknownSs3Event(seq)
	}
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
	default:
		return len(seq), UnknownOscEvent(seq)
	}
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
	var seq []byte
	var i int
	if p[i] == ansi.DCS || p[i] == ansi.ESC {
		seq = append(seq, p[i])
		i++
	}
	if i < len(p) && p[i-1] == ansi.ESC && p[i] == 'P' {
		seq = append(seq, p[i])
		i++
	}

	// Scan parameter bytes in the range 0x30-0x3F
	var start, end int // start and end of the parameter bytes
	for j := 0; i < len(p) && p[i] >= 0x30 && p[i] <= 0x3F; i, j = i+1, j+1 {
		if j == 0 {
			start = i
		}
		seq = append(seq, p[i])
	}

	end = i

	// Scan intermediate bytes in the range 0x20-0x2F
	var istart, iend int
	for j := 0; i < len(p) && p[i] >= 0x20 && p[i] <= 0x2F; i, j = i+1, j+1 {
		if j == 0 {
			istart = i
		}
		seq = append(seq, p[i])
	}

	iend = i

	// Final byte
	var final byte

	// Scan final byte in the range 0x40-0x7E
	if i >= len(p) || p[i] < 0x40 || p[i] > 0x7E {
		return len(seq), UnknownEvent(seq)
	}
	// Add the final byte
	final = p[i]
	seq = append(seq, p[i])

	if i+1 >= len(p) {
		return len(seq), UnknownEvent(seq)
	}

	// Collect data bytes until a ST character is found
	// data bytes are in the range of 0x08-0x0D and 0x20-0x7F
	// but we don't care about the actual values for now
	var data []byte
	for i++; i < len(p) && p[i] != ansi.ST && p[i] != ansi.ESC; i++ {
		data = append(data, p[i])
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

	switch final {
	case 'r':
		inters := p[istart:iend] // intermediates
		if len(inters) == 0 {
			return len(seq), UnknownDcsEvent(seq)
		}
		switch inters[0] {
		case '+':
			// XTGETTCAP responses
			params := ansi.Params(p[start:end])
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

func parseApc(p []byte) (int, Event) {
	// APC sequences are introduced by APC (0x9f) or ESC _ (0x1b 0x5f)
	return parseCtrl(ansi.APC, '_')(p)
}

func parseUtf8(p []byte) (int, Event) {
	r, rw := utf8.DecodeRune(p)
	if r == utf8.RuneError || r <= ansi.US || r == ansi.DEL || r == ansi.SP {
		return 0, nil
	}
	return rw, KeyDownEvent{Rune: r}
}

func parseCtrl0(b byte) Event {
	switch b {
	case ansi.NUL:
		return KeyDownEvent{Rune: ' ', Sym: KeySpace, Mod: Ctrl}
	case ansi.SOH:
		return KeyDownEvent{Rune: 'a', Mod: Ctrl}
	case ansi.STX:
		return KeyDownEvent{Rune: 'b', Mod: Ctrl}
	case ansi.ETX:
		return KeyDownEvent{Rune: 'c', Mod: Ctrl}
	case ansi.EOT:
		return KeyDownEvent{Rune: 'd', Mod: Ctrl}
	case ansi.ENQ:
		return KeyDownEvent{Rune: 'e', Mod: Ctrl}
	case ansi.ACK:
		return KeyDownEvent{Rune: 'f', Mod: Ctrl}
	case ansi.BEL:
		return KeyDownEvent{Rune: 'g', Mod: Ctrl}
	case ansi.BS:
		return KeyDownEvent{Rune: 'h', Mod: Ctrl}
	case ansi.HT:
		return KeyDownEvent{Sym: KeyTab}
	case ansi.LF:
		return KeyDownEvent{Rune: 'j', Mod: Ctrl}
	case ansi.VT:
		return KeyDownEvent{Rune: 'k', Mod: Ctrl}
	case ansi.FF:
		return KeyDownEvent{Rune: 'l', Mod: Ctrl}
	case ansi.CR:
		return KeyDownEvent{Sym: KeyEnter}
	case ansi.SO:
		return KeyDownEvent{Rune: 'n', Mod: Ctrl}
	case ansi.SI:
		return KeyDownEvent{Rune: 'o', Mod: Ctrl}
	case ansi.DLE:
		return KeyDownEvent{Rune: 'p', Mod: Ctrl}
	case ansi.DC1:
		return KeyDownEvent{Rune: 'q', Mod: Ctrl}
	case ansi.DC2:
		return KeyDownEvent{Rune: 'r', Mod: Ctrl}
	case ansi.DC3:
		return KeyDownEvent{Rune: 's', Mod: Ctrl}
	case ansi.DC4:
		return KeyDownEvent{Rune: 't', Mod: Ctrl}
	case ansi.NAK:
		return KeyDownEvent{Rune: 'u', Mod: Ctrl}
	case ansi.SYN:
		return KeyDownEvent{Rune: 'v', Mod: Ctrl}
	case ansi.ETB:
		return KeyDownEvent{Rune: 'w', Mod: Ctrl}
	case ansi.CAN:
		return KeyDownEvent{Rune: 'x', Mod: Ctrl}
	case ansi.EM:
		return KeyDownEvent{Rune: 'y', Mod: Ctrl}
	case ansi.SUB:
		return KeyDownEvent{Rune: 'z', Mod: Ctrl}
	case ansi.ESC:
		return KeyDownEvent{Sym: KeyEscape}
	case ansi.FS:
		return KeyDownEvent{Rune: '\\', Mod: Ctrl}
	case ansi.GS:
		return KeyDownEvent{Rune: ']', Mod: Ctrl}
	case ansi.RS:
		return KeyDownEvent{Rune: '^', Mod: Ctrl}
	case ansi.US:
		return KeyDownEvent{Rune: '_', Mod: Ctrl}
	case ansi.SP:
		return KeyDownEvent{Sym: KeySpace, Rune: ' '}
	case ansi.DEL:
		return KeyDownEvent{Sym: KeyBackspace}
	default:
		return UnknownEvent(b)
	}
}
