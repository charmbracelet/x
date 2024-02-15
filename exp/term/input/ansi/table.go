package ansi

import (
	"github.com/charmbracelet/x/exp/term/ansi"
	"github.com/charmbracelet/x/exp/term/input"
)

func (d *driver) registerKeys(flags int) {
	nul := input.KeyEvent{Rune: '@', Mod: input.Ctrl} // ctrl+@ or ctrl+space
	if flags&Fctrlsp != 0 {
		if flags&Fspacesym != 0 {
			nul.Rune = 0
			nul.Sym = input.KeySpace
		} else {
			nul.Rune = ' '
		}
	}

	tab := input.KeyEvent{Rune: 'i', Mod: input.Ctrl} // ctrl+i or tab
	if flags&Ftabsym != 0 {
		tab.Rune = 0
		tab.Mod = 0
		tab.Sym = input.KeyTab
	}

	enter := input.KeyEvent{Rune: 'm', Mod: input.Ctrl} // ctrl+m or enter
	if flags&Fentersym != 0 {
		enter.Rune = 0
		enter.Mod = 0
		enter.Sym = input.KeyEnter
	}

	esc := input.KeyEvent{Rune: '[', Mod: input.Ctrl} // ctrl+[ or escape
	if flags&Fescsym != 0 {
		esc.Rune = 0
		esc.Mod = 0
		esc.Sym = input.KeyEscape
	}

	sp := input.KeyEvent{Rune: ' '}
	if flags&Fspacesym != 0 {
		sp.Rune = 0
		sp.Sym = input.KeySpace
	}

	del := input.KeyEvent{Sym: input.KeyDelete}
	if flags&Fdelbackspace != 0 {
		del.Sym = input.KeyBackspace
	}

	find := input.KeyEvent{Sym: input.KeyFind}
	if flags&Ffindhome != 0 {
		find.Sym = input.KeyHome
	}

	sel := input.KeyEvent{Sym: input.KeySelect}
	if flags&Fselectend != 0 {
		sel.Sym = input.KeyEnd
	}

	// The following is a table of key sequences and their corresponding key
	// events based on the VT100/VT200 terminal specs.
	//
	// See: https://vt100.net/docs/vt100-ug/chapter3.html#S3.2
	// See: https://vt100.net/docs/vt220-rm/chapter3.html
	//
	// XXX: These keys may be overwritten by other options like XTerm or
	// Terminfo.
	d.table = map[string]input.KeyEvent{
		// C0 control characters
		string(ansi.NUL): nul,
		string(ansi.SOH): {Rune: 'a', Mod: input.Ctrl},
		string(ansi.STX): {Rune: 'b', Mod: input.Ctrl},
		string(ansi.ETX): {Rune: 'c', Mod: input.Ctrl},
		string(ansi.EOT): {Rune: 'd', Mod: input.Ctrl},
		string(ansi.ENQ): {Rune: 'e', Mod: input.Ctrl},
		string(ansi.ACK): {Rune: 'f', Mod: input.Ctrl},
		string(ansi.BEL): {Rune: 'g', Mod: input.Ctrl},
		string(ansi.BS):  {Rune: 'h', Mod: input.Ctrl},
		string(ansi.HT):  tab,
		string(ansi.LF):  {Rune: 'j', Mod: input.Ctrl},
		string(ansi.VT):  {Rune: 'k', Mod: input.Ctrl},
		string(ansi.FF):  {Rune: 'l', Mod: input.Ctrl},
		string(ansi.CR):  enter,
		string(ansi.SO):  {Rune: 'n', Mod: input.Ctrl},
		string(ansi.SI):  {Rune: 'o', Mod: input.Ctrl},
		string(ansi.DLE): {Rune: 'p', Mod: input.Ctrl},
		string(ansi.DC1): {Rune: 'q', Mod: input.Ctrl},
		string(ansi.DC2): {Rune: 'r', Mod: input.Ctrl},
		string(ansi.DC3): {Rune: 's', Mod: input.Ctrl},
		string(ansi.DC4): {Rune: 't', Mod: input.Ctrl},
		string(ansi.NAK): {Rune: 'u', Mod: input.Ctrl},
		string(ansi.SYN): {Rune: 'v', Mod: input.Ctrl},
		string(ansi.ETB): {Rune: 'w', Mod: input.Ctrl},
		string(ansi.CAN): {Rune: 'x', Mod: input.Ctrl},
		string(ansi.EM):  {Rune: 'y', Mod: input.Ctrl},
		string(ansi.SUB): {Rune: 'z', Mod: input.Ctrl},
		string(ansi.ESC): esc,
		string(ansi.FS):  {Rune: '\\', Mod: input.Ctrl},
		string(ansi.GS):  {Rune: ']', Mod: input.Ctrl},
		string(ansi.RS):  {Rune: '^', Mod: input.Ctrl},
		string(ansi.US):  {Rune: '_', Mod: input.Ctrl},

		// Special keys in G0
		string(ansi.SP):  sp,
		string(ansi.DEL): del,

		// Special keys

		"\x1b[Z": {Sym: input.KeyTab, Mod: input.Shift},

		"\x1b[1~": find,
		"\x1b[2~": {Sym: input.KeyInsert},
		"\x1b[3~": {Sym: input.KeyDelete},
		"\x1b[4~": sel,
		"\x1b[5~": {Sym: input.KeyPgUp},
		"\x1b[6~": {Sym: input.KeyPgDown},
		"\x1b[7~": {Sym: input.KeyHome},
		"\x1b[8~": {Sym: input.KeyEnd},

		// Normal mode
		"\x1b[A": {Sym: input.KeyUp},
		"\x1b[B": {Sym: input.KeyDown},
		"\x1b[C": {Sym: input.KeyRight},
		"\x1b[D": {Sym: input.KeyLeft},
		"\x1b[E": {Sym: input.KeyBegin},
		"\x1b[F": {Sym: input.KeyEnd},
		"\x1b[H": {Sym: input.KeyHome},
		"\x1b[P": {Sym: input.KeyF1},
		"\x1b[Q": {Sym: input.KeyF2},
		"\x1b[R": {Sym: input.KeyF3},
		"\x1b[S": {Sym: input.KeyF4},

		// Application Cursor Key Mode (DECCKM)
		"\x1bOA": {Sym: input.KeyUp},
		"\x1bOB": {Sym: input.KeyDown},
		"\x1bOC": {Sym: input.KeyRight},
		"\x1bOD": {Sym: input.KeyLeft},
		"\x1bOE": {Sym: input.KeyBegin},
		"\x1bOF": {Sym: input.KeyEnd},
		"\x1bOH": {Sym: input.KeyHome},
		"\x1bOP": {Sym: input.KeyF1},
		"\x1bOQ": {Sym: input.KeyF2},
		"\x1bOR": {Sym: input.KeyF3},
		"\x1bOS": {Sym: input.KeyF4},

		// Keypad Application Mode (DECKPAM)

		"\x1bOM": {Sym: input.KeyKpEnter},
		"\x1bOX": {Sym: input.KeyKpEqual},
		"\x1bOj": {Sym: input.KeyKpMul},
		"\x1bOk": {Sym: input.KeyKpPlus},
		"\x1bOl": {Sym: input.KeyKpComma},
		"\x1bOm": {Sym: input.KeyKpMinus},
		"\x1bOn": {Sym: input.KeyKpPeriod},
		"\x1bOo": {Sym: input.KeyKpDiv},
		"\x1bOp": {Sym: input.KeyKp0},
		"\x1bOq": {Sym: input.KeyKp1},
		"\x1bOr": {Sym: input.KeyKp2},
		"\x1bOs": {Sym: input.KeyKp3},
		"\x1bOt": {Sym: input.KeyKp4},
		"\x1bOu": {Sym: input.KeyKp5},
		"\x1bOv": {Sym: input.KeyKp6},
		"\x1bOw": {Sym: input.KeyKp7},
		"\x1bOx": {Sym: input.KeyKp8},
		"\x1bOy": {Sym: input.KeyKp9},

		// Function keys

		"\x1b[11~": {Sym: input.KeyF1},
		"\x1b[12~": {Sym: input.KeyF2},
		"\x1b[13~": {Sym: input.KeyF3},
		"\x1b[14~": {Sym: input.KeyF4},
		"\x1b[15~": {Sym: input.KeyF5},
		"\x1b[17~": {Sym: input.KeyF6},
		"\x1b[18~": {Sym: input.KeyF7},
		"\x1b[19~": {Sym: input.KeyF8},
		"\x1b[20~": {Sym: input.KeyF9},
		"\x1b[21~": {Sym: input.KeyF10},
		"\x1b[23~": {Sym: input.KeyF11},
		"\x1b[24~": {Sym: input.KeyF12},
		"\x1b[25~": {Sym: input.KeyF13},
		"\x1b[26~": {Sym: input.KeyF14},
		"\x1b[28~": {Sym: input.KeyF15},
		"\x1b[29~": {Sym: input.KeyF16},
		"\x1b[31~": {Sym: input.KeyF17},
		"\x1b[32~": {Sym: input.KeyF18},
		"\x1b[33~": {Sym: input.KeyF19},
		"\x1b[34~": {Sym: input.KeyF20},
	}

	// CSI function keys
	csiFuncKeys := map[string]input.KeyEvent{
		"A": {Sym: input.KeyUp}, "B": {Sym: input.KeyDown},
		"C": {Sym: input.KeyRight}, "D": {Sym: input.KeyLeft},
		"E": {Sym: input.KeyBegin}, "F": {Sym: input.KeyEnd},
		"H": {Sym: input.KeyHome}, "P": {Sym: input.KeyF1},
		"Q": {Sym: input.KeyF2}, "R": {Sym: input.KeyF3},
		"S": {Sym: input.KeyF4},
	}

	// CSI ~ sequence keys
	csiTildeKeys := map[string]input.KeyEvent{
		"1": find, "2": {Sym: input.KeyInsert},
		"3": {Sym: input.KeyDelete}, "4": sel,
		"5": {Sym: input.KeyPgUp}, "6": {Sym: input.KeyPgDown},
		"7": {Sym: input.KeyHome}, "8": {Sym: input.KeyEnd},
		// There are no 9 and 10 keys
		"11": {Sym: input.KeyF1}, "12": {Sym: input.KeyF2},
		"13": {Sym: input.KeyF3}, "14": {Sym: input.KeyF4},
		"15": {Sym: input.KeyF5}, "17": {Sym: input.KeyF6},
		"18": {Sym: input.KeyF7}, "19": {Sym: input.KeyF8},
		"20": {Sym: input.KeyF9}, "21": {Sym: input.KeyF10},
		"23": {Sym: input.KeyF11}, "24": {Sym: input.KeyF12},
		"25": {Sym: input.KeyF13}, "26": {Sym: input.KeyF14},
		"28": {Sym: input.KeyF15}, "29": {Sym: input.KeyF16},
		"31": {Sym: input.KeyF17}, "32": {Sym: input.KeyF18},
		"33": {Sym: input.KeyF19}, "34": {Sym: input.KeyF20},
	}

	if flags&Fxterm != 0 {
		// XTerm modifiers
		// These are offset by 1 to be compatible with our Mod type.
		// See https://invisible-island.net/xterm/ctlseqs/ctlseqs.html#h3-PC-Style-Function-Keys
		for _, m := range []input.Mod{
			input.Shift,                                       // 1
			input.Alt,                                         // 2
			input.Shift | input.Alt,                           // 3
			input.Ctrl,                                        // 4
			input.Shift | input.Ctrl,                          // 5
			input.Alt | input.Ctrl,                            // 6
			input.Shift | input.Alt | input.Ctrl,              // 7
			input.Meta,                                        // 8
			input.Meta | input.Shift,                          // 9
			input.Meta | input.Alt,                            // 10
			input.Meta | input.Shift | input.Alt,              // 11
			input.Meta | input.Ctrl,                           // 12
			input.Meta | input.Shift | input.Ctrl,             // 13
			input.Meta | input.Alt | input.Ctrl,               // 14
			input.Meta | input.Shift | input.Alt | input.Ctrl, // 15
		} {
			// XTerm modifier offset +1
			xtermMod := string('0' + byte(m+1))

			//  CSI 1 ; <modifier> <func>
			for k, v := range csiFuncKeys {
				// Functions always have a leading 1 param
				seq := "\x1b[1;" + xtermMod + k
				key := v
				key.Mod = m
				d.table[seq] = key
			}
			//  CSI <number> ; <modifier> ~
			for k, v := range csiTildeKeys {
				seq := "\x1b[" + k + ";" + xtermMod + "~"
				key := v
				key.Mod = m
				d.table[seq] = key
			}
		}
	}

	// URxvt keys
	// See https://manpages.ubuntu.com/manpages/trusty/man7/urxvt.7.html#key%20codes
	d.table["\x1b[a"] = input.KeyEvent{Sym: input.KeyUp, Mod: input.Shift}
	d.table["\x1b[b"] = input.KeyEvent{Sym: input.KeyDown, Mod: input.Shift}
	d.table["\x1b[c"] = input.KeyEvent{Sym: input.KeyRight, Mod: input.Shift}
	d.table["\x1b[d"] = input.KeyEvent{Sym: input.KeyLeft, Mod: input.Shift}
	d.table["\x1bOa"] = input.KeyEvent{Sym: input.KeyUp, Mod: input.Ctrl}
	d.table["\x1bOb"] = input.KeyEvent{Sym: input.KeyDown, Mod: input.Ctrl}
	d.table["\x1bOc"] = input.KeyEvent{Sym: input.KeyRight, Mod: input.Ctrl}
	d.table["\x1bOd"] = input.KeyEvent{Sym: input.KeyLeft, Mod: input.Ctrl}
	// TODO: invistigate if shift-ctrl arrow keys collide with DECCKM keys i.e.
	// "\x1bOA", "\x1bOB", "\x1bOC", "\x1bOD"

	// URxvt modifier CSI ~ keys
	for k, v := range csiTildeKeys {
		key := v
		// Normal (no modifier) already defined part of VT100/VT200
		// Shift modifier
		key.Mod = input.Shift
		d.table["\x1b["+k+"$"] = key
		// Ctrl modifier
		key.Mod = input.Ctrl
		d.table["\x1b["+k+"^"] = key
		// Shift-Ctrl modifier
		key.Mod = input.Shift | input.Ctrl
		d.table["\x1b["+k+"@"] = key
	}

	// URxvt F keys
	// Note: Shift + F1-F10 generates F11-F20.
	// This means Shift + F1 and Shift + F2 will generate F11 and F12, the same
	// applies to Ctrl + Shift F1 & F2.
	//
	// P.S. Don't like this? Blame URxvt, configure your terminal to use
	// different escapes like XTerm, or switch to a better terminal ¯\_(ツ)_/¯
	//
	// See https://manpages.ubuntu.com/manpages/trusty/man7/urxvt.7.html#key%20codes
	d.table["\x1b[23$"] = input.KeyEvent{Sym: input.KeyF11, Mod: input.Shift}
	d.table["\x1b[24$"] = input.KeyEvent{Sym: input.KeyF12, Mod: input.Shift}
	d.table["\x1b[25$"] = input.KeyEvent{Sym: input.KeyF13, Mod: input.Shift}
	d.table["\x1b[26$"] = input.KeyEvent{Sym: input.KeyF14, Mod: input.Shift}
	d.table["\x1b[28$"] = input.KeyEvent{Sym: input.KeyF15, Mod: input.Shift}
	d.table["\x1b[29$"] = input.KeyEvent{Sym: input.KeyF16, Mod: input.Shift}
	d.table["\x1b[31$"] = input.KeyEvent{Sym: input.KeyF17, Mod: input.Shift}
	d.table["\x1b[32$"] = input.KeyEvent{Sym: input.KeyF18, Mod: input.Shift}
	d.table["\x1b[33$"] = input.KeyEvent{Sym: input.KeyF19, Mod: input.Shift}
	d.table["\x1b[34$"] = input.KeyEvent{Sym: input.KeyF20, Mod: input.Shift}
	d.table["\x1b[11^"] = input.KeyEvent{Sym: input.KeyF1, Mod: input.Ctrl}
	d.table["\x1b[12^"] = input.KeyEvent{Sym: input.KeyF2, Mod: input.Ctrl}
	d.table["\x1b[13^"] = input.KeyEvent{Sym: input.KeyF3, Mod: input.Ctrl}
	d.table["\x1b[14^"] = input.KeyEvent{Sym: input.KeyF4, Mod: input.Ctrl}
	d.table["\x1b[15^"] = input.KeyEvent{Sym: input.KeyF5, Mod: input.Ctrl}
	d.table["\x1b[17^"] = input.KeyEvent{Sym: input.KeyF6, Mod: input.Ctrl}
	d.table["\x1b[18^"] = input.KeyEvent{Sym: input.KeyF7, Mod: input.Ctrl}
	d.table["\x1b[19^"] = input.KeyEvent{Sym: input.KeyF8, Mod: input.Ctrl}
	d.table["\x1b[20^"] = input.KeyEvent{Sym: input.KeyF9, Mod: input.Ctrl}
	d.table["\x1b[21^"] = input.KeyEvent{Sym: input.KeyF10, Mod: input.Ctrl}
	d.table["\x1b[23^"] = input.KeyEvent{Sym: input.KeyF11, Mod: input.Ctrl}
	d.table["\x1b[24^"] = input.KeyEvent{Sym: input.KeyF12, Mod: input.Ctrl}
	d.table["\x1b[25^"] = input.KeyEvent{Sym: input.KeyF13, Mod: input.Ctrl}
	d.table["\x1b[26^"] = input.KeyEvent{Sym: input.KeyF14, Mod: input.Ctrl}
	d.table["\x1b[28^"] = input.KeyEvent{Sym: input.KeyF15, Mod: input.Ctrl}
	d.table["\x1b[29^"] = input.KeyEvent{Sym: input.KeyF16, Mod: input.Ctrl}
	d.table["\x1b[31^"] = input.KeyEvent{Sym: input.KeyF17, Mod: input.Ctrl}
	d.table["\x1b[32^"] = input.KeyEvent{Sym: input.KeyF18, Mod: input.Ctrl}
	d.table["\x1b[33^"] = input.KeyEvent{Sym: input.KeyF19, Mod: input.Ctrl}
	d.table["\x1b[34^"] = input.KeyEvent{Sym: input.KeyF20, Mod: input.Ctrl}
	d.table["\x1b[23@"] = input.KeyEvent{Sym: input.KeyF11, Mod: input.Shift | input.Ctrl}
	d.table["\x1b[24@"] = input.KeyEvent{Sym: input.KeyF12, Mod: input.Shift | input.Ctrl}
	d.table["\x1b[25@"] = input.KeyEvent{Sym: input.KeyF13, Mod: input.Shift | input.Ctrl}
	d.table["\x1b[26@"] = input.KeyEvent{Sym: input.KeyF14, Mod: input.Shift | input.Ctrl}
	d.table["\x1b[28@"] = input.KeyEvent{Sym: input.KeyF15, Mod: input.Shift | input.Ctrl}
	d.table["\x1b[29@"] = input.KeyEvent{Sym: input.KeyF16, Mod: input.Shift | input.Ctrl}
	d.table["\x1b[31@"] = input.KeyEvent{Sym: input.KeyF17, Mod: input.Shift | input.Ctrl}
	d.table["\x1b[32@"] = input.KeyEvent{Sym: input.KeyF18, Mod: input.Shift | input.Ctrl}
	d.table["\x1b[33@"] = input.KeyEvent{Sym: input.KeyF19, Mod: input.Shift | input.Ctrl}
	d.table["\x1b[34@"] = input.KeyEvent{Sym: input.KeyF20, Mod: input.Shift | input.Ctrl}

	// Register Alt + <key> combinations
	for k, v := range d.table {
		v.Mod |= input.Alt
		d.table["\x1b"+k] = v
	}

	// Register terminfo keys
	d.registerTerminfoKeys()
}
