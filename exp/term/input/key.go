package input

// KeySym is a keyboard symbol.
type KeySym int

// Symbol constants.
const (
	KeyNone KeySym = iota

	// Special names in C0

	KeyBackspace
	KeyTab
	KeyEnter
	KeyEscape

	// Special names in G0

	KeySpace
	KeyDel

	// Special keys

	KeyUp
	KeyDown
	KeyLeft
	KeyRight
	KeyBegin
	KeyFind
	KeyInsert
	KeyDelete
	KeySelect
	KeyPgUp
	KeyPgDown
	KeyHome
	KeyEnd

	// Keypad keys

	KeyKpEnter
	KeyKpEqual
	KeyKpMul
	KeyKpPlus
	KeyKpComma
	KeyKpMinus
	KeyKpPeriod
	KeyKpDiv
	KeyKp0
	KeyKp1
	KeyKp2
	KeyKp3
	KeyKp4
	KeyKp5
	KeyKp6
	KeyKp7
	KeyKp8
	KeyKp9

	// Function keys

	KeyF1
	KeyF2
	KeyF3
	KeyF4
	KeyF5
	KeyF6
	KeyF7
	KeyF8
	KeyF9
	KeyF10
	KeyF11
	KeyF12
	KeyF13
	KeyF14
	KeyF15
	KeyF16
	KeyF17
	KeyF18
	KeyF19
	KeyF20
)

// KeyEvent is a keyboard key event.
type KeyEvent struct {
	Sym  KeySym
	Rune rune
	Mod  Mod
}

var _ Event = KeyEvent{}

// String implements Event.
func (k KeyEvent) String() string {
	var s string
	if k.Mod.IsCtrl() {
		s += "ctrl+"
	}
	if k.Mod.IsAlt() {
		s += "alt+"
	}
	if k.Mod.IsShift() {
		s += "shift+"
	}
	if k.Rune != 0 {
		// Space is the only invisible printable character.
		if k.Rune == ' ' {
			s += "space"
		} else {
			s += string(k.Rune)
		}
	} else {
		sym := keySymString[k.Sym]
		if sym == "" {
			s += "unknown"
		} else {
			s += sym
		}
	}
	return s
}

// Type implements Event.
func (KeyEvent) Type() string {
	return "Key"
}

var keySymString = map[KeySym]string{
	KeyEnter:     "enter",
	KeyTab:       "tab",
	KeyBackspace: "backspace",
	KeyEscape:    "escape",
	KeySpace:     "space",
	KeyDel:       "del",
	KeyUp:        "up",
	KeyDown:      "down",
	KeyLeft:      "left",
	KeyRight:     "right",
	KeyBegin:     "begin",
	KeyFind:      "find",
	KeyInsert:    "insert",
	KeyDelete:    "delete",
	KeySelect:    "select",
	KeyPgUp:      "pgup",
	KeyPgDown:    "pgdown",
	KeyHome:      "home",
	KeyEnd:       "end",
	KeyKpEnter:   "kpenter",
	KeyKpEqual:   "kpequal",
	KeyKpMul:     "kpmul",
	KeyKpPlus:    "kpplus",
	KeyKpComma:   "kpcomma",
	KeyKpMinus:   "kpminus",
	KeyKpPeriod:  "kpperiod",
	KeyKpDiv:     "kpdiv",
	KeyKp0:       "kp0",
	KeyKp1:       "kp1",
	KeyKp2:       "kp2",
	KeyKp3:       "kp3",
	KeyKp4:       "kp4",
	KeyKp5:       "kp5",
	KeyKp6:       "kp6",
	KeyKp7:       "kp7",
	KeyKp8:       "kp8",
	KeyKp9:       "kp9",
	KeyF1:        "f1",
	KeyF2:        "f2",
	KeyF3:        "f3",
	KeyF4:        "f4",
	KeyF5:        "f5",
	KeyF6:        "f6",
	KeyF7:        "f7",
	KeyF8:        "f8",
	KeyF9:        "f9",
	KeyF10:       "f10",
	KeyF11:       "f11",
	KeyF12:       "f12",
	KeyF13:       "f13",
	KeyF14:       "f14",
	KeyF15:       "f15",
	KeyF16:       "f16",
	KeyF17:       "f17",
	KeyF18:       "f18",
	KeyF19:       "f19",
	KeyF20:       "f20",
}
