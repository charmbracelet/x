package ansi

import (
	"github.com/charmbracelet/x/exp/term/ansi"
	"github.com/charmbracelet/x/exp/term/input"
)

func (d *driver) registerKeys(flags int) {
	nul := input.Key{Rune: '@', Mod: input.Ctrl} // ctrl+@ or ctrl+space
	if flags&Fctrlsp != 0 {
		if flags&Fspacesym != 0 {
			nul.Rune = 0
			nul.Sym = input.Space
		} else {
			nul.Rune = ' '
		}
	}

	tab := input.Key{Rune: 'i', Mod: input.Ctrl} // ctrl+i or tab
	if flags&Ftabsym != 0 {
		tab.Rune = 0
		tab.Mod = 0
		tab.Sym = input.Tab
	}

	enter := input.Key{Rune: 'm', Mod: input.Ctrl} // ctrl+m or enter
	if flags&Fentersym != 0 {
		enter.Rune = 0
		enter.Mod = 0
		enter.Sym = input.Enter
	}

	esc := input.Key{Rune: '[', Mod: input.Ctrl} // ctrl+[ or escape
	if flags&Fescsym != 0 {
		esc.Rune = 0
		esc.Mod = 0
		esc.Sym = input.Escape
	}

	sp := input.Key{Rune: ' '}
	if flags&Fspacesym != 0 {
		sp.Rune = 0
		sp.Sym = input.Space
	}

	del := input.Key{Sym: input.Delete}
	if flags&Fdelbackspace != 0 {
		del.Sym = input.Backspace
	}

	find := input.Key{Sym: input.Find}
	if flags&Ffindhome != 0 {
		find.Sym = input.Home
	}

	select_ := input.Key{Sym: input.Select}
	if flags&Fselectend != 0 {
		select_.Sym = input.End
	}

	// See: https://vt100.net/docs/vt100-ug/chapter3.html#S3.2
	// See: https://vt100.net/docs/vt220-rm/chapter3.html
	// See: https://vt100.net/docs/vt510-rm/chapter8.html
	d.table = map[string]input.Key{
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

		"\x1b[Z": {Sym: input.Tab, Mod: input.Shift},

		"\x1b[1~": find,
		"\x1b[2~": {Sym: input.Insert},
		"\x1b[3~": {Sym: input.Delete},
		"\x1b[4~": select_,
		"\x1b[5~": {Sym: input.PgUp},
		"\x1b[6~": {Sym: input.PgDown},
		"\x1b[7~": {Sym: input.Home},
		"\x1b[8~": {Sym: input.End},

		// Normal mode
		"\x1b[A": {Sym: input.Up},
		"\x1b[B": {Sym: input.Down},
		"\x1b[C": {Sym: input.Right},
		"\x1b[D": {Sym: input.Left},
		"\x1b[E": {Sym: input.Begin},
		"\x1b[F": {Sym: input.End},
		"\x1b[H": {Sym: input.Home},
		"\x1b[P": {Sym: input.F1},
		"\x1b[Q": {Sym: input.F2},
		"\x1b[R": {Sym: input.F3},
		"\x1b[S": {Sym: input.F4},

		// Application Cursor Key Mode (DECCKM)
		"\x1bOA": {Sym: input.Up},
		"\x1bOB": {Sym: input.Down},
		"\x1bOC": {Sym: input.Right},
		"\x1bOD": {Sym: input.Left},
		"\x1bOE": {Sym: input.Begin},
		"\x1bOF": {Sym: input.End},
		"\x1bOH": {Sym: input.Home},
		"\x1bOP": {Sym: input.F1},
		"\x1bOQ": {Sym: input.F2},
		"\x1bOR": {Sym: input.F3},
		"\x1bOS": {Sym: input.F4},

		// Keypad Application Mode (DECKPAM)

		"\x1bOM": {Sym: input.KpEnter},
		"\x1bOX": {Sym: input.KpEqual},
		"\x1bOj": {Sym: input.KpMul},
		"\x1bOk": {Sym: input.KpPlus},
		"\x1bOl": {Sym: input.KpComma},
		"\x1bOm": {Sym: input.KpMinus},
		"\x1bOn": {Sym: input.KpPeriod},
		"\x1bOo": {Sym: input.KpDiv},
		"\x1bOp": {Sym: input.Kp0},
		"\x1bOq": {Sym: input.Kp1},
		"\x1bOr": {Sym: input.Kp2},
		"\x1bOs": {Sym: input.Kp3},
		"\x1bOt": {Sym: input.Kp4},
		"\x1bOu": {Sym: input.Kp5},
		"\x1bOv": {Sym: input.Kp6},
		"\x1bOw": {Sym: input.Kp7},
		"\x1bOx": {Sym: input.Kp8},
		"\x1bOy": {Sym: input.Kp9},

		// Function keys

		"\x1b[11~": {Sym: input.F1},
		"\x1b[12~": {Sym: input.F2},
		"\x1b[13~": {Sym: input.F3},
		"\x1b[14~": {Sym: input.F4},
		"\x1b[15~": {Sym: input.F5},
		"\x1b[17~": {Sym: input.F6},
		"\x1b[18~": {Sym: input.F7},
		"\x1b[19~": {Sym: input.F8},
		"\x1b[20~": {Sym: input.F9},
		"\x1b[21~": {Sym: input.F10},
		"\x1b[23~": {Sym: input.F11},
		"\x1b[24~": {Sym: input.F12},
		"\x1b[25~": {Sym: input.F13},
		"\x1b[26~": {Sym: input.F14},
		"\x1b[28~": {Sym: input.F15},
		"\x1b[29~": {Sym: input.F16},
		"\x1b[31~": {Sym: input.F17},
		"\x1b[32~": {Sym: input.F18},
		"\x1b[33~": {Sym: input.F19},
		"\x1b[34~": {Sym: input.F20},
	}
}
