package input

// KeySym is a keyboard symbol.
type KeySym int

// Symbol constants.
const (
	None KeySym = iota

	// Special names in C0

	Backspace
	Tab
	Enter
	Escape

	// Special names in G0

	Space
	Del

	// Special keys

	Up
	Down
	Left
	Right
	Begin
	Find
	Insert
	Delete
	Select
	PgUp
	PgDown
	Home
	End

	// Keypad keys

	KpEnter
	KpEqual
	KpMul
	KpPlus
	KpComma
	KpMinus
	KpPeriod
	KpDiv
	Kp0
	Kp1
	Kp2
	Kp3
	Kp4
	Kp5
	Kp6
	Kp7
	Kp8
	Kp9

	// Function keys

	F1
	F2
	F3
	F4
	F5
	F6
	F7
	F8
	F9
	F10
	F11
	F12
	F13
	F14
	F15
	F16
	F17
	F18
	F19
	F20
)

// Key is a keyboard key event.
type Key struct {
	Sym  KeySym
	Rune rune
	Mod  Mod
}

// KeyEvent is a keyboard key event.
type KeyEvent Key

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
	Enter:     "enter",
	Tab:       "tab",
	Backspace: "backspace",
	Escape:    "escape",
	Space:     "space",
	Del:       "del",
	Up:        "up",
	Down:      "down",
	Left:      "left",
	Right:     "right",
	Begin:     "begin",
	Find:      "find",
	Insert:    "insert",
	Delete:    "delete",
	Select:    "select",
	PgUp:      "pgup",
	PgDown:    "pgdown",
	Home:      "home",
	End:       "end",
	KpEnter:   "kpenter",
	KpEqual:   "kpequal",
	KpMul:     "kpmul",
	KpPlus:    "kpplus",
	KpComma:   "kpcomma",
	KpMinus:   "kpminus",
	KpPeriod:  "kpperiod",
	KpDiv:     "kpdiv",
	Kp0:       "kp0",
	Kp1:       "kp1",
	Kp2:       "kp2",
	Kp3:       "kp3",
	Kp4:       "kp4",
	Kp5:       "kp5",
	Kp6:       "kp6",
	Kp7:       "kp7",
	Kp8:       "kp8",
	Kp9:       "kp9",
	F1:        "f1",
	F2:        "f2",
	F3:        "f3",
	F4:        "f4",
	F5:        "f5",
	F6:        "f6",
	F7:        "f7",
	F8:        "f8",
	F9:        "f9",
	F10:       "f10",
	F11:       "f11",
	F12:       "f12",
	F13:       "f13",
	F14:       "f14",
	F15:       "f15",
	F16:       "f16",
	F17:       "f17",
	F18:       "f18",
	F19:       "f19",
	F20:       "f20",
}
