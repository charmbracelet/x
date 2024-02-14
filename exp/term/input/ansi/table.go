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

	select_ := input.KeyEvent{Sym: input.KeySelect}
	if flags&Fselectend != 0 {
		select_.Sym = input.KeyEnd
	}

	// See: https://vt100.net/docs/vt100-ug/chapter3.html#S3.2
	// See: https://vt100.net/docs/vt220-rm/chapter3.html
	// See: https://vt100.net/docs/vt510-rm/chapter8.html
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
		"\x1b[4~": select_,
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
}
