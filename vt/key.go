package vt

import (
	"unicode"

	"github.com/charmbracelet/x/ansi"
)

// KeyMod represents a key modifier.
type KeyMod int

// Modifier keys.
const (
	ModShift KeyMod = 1 << iota
	ModAlt
	ModCtrl
	ModMeta
)

// Key represents a key press event.
type Key struct {
	Code rune
	Mod  KeyMod
}

// SendKey returns the default key map.
func (t *Terminal) SendKey(k Key) {
	var seq string

	var (
		ack bool // Application cursor keys mode
		akk bool // Application keypad keys mode
	)

	if mode, ok := t.pmodes[ansi.CursorKeysMode]; ok && mode.IsSet() {
		ack = true
	}
	if mode, ok := t.pmodes[ansi.NumericKeypadMode]; ok && mode.IsSet() {
		akk = true
	}

	switch k {
	// Control keys
	case Key{Code: KeySpace, Mod: ModCtrl}:
		seq = "\x00"
	case Key{Code: 'a', Mod: ModCtrl}:
		seq = "\x01"
	case Key{Code: 'b', Mod: ModCtrl}:
		seq = "\x02"
	case Key{Code: 'c', Mod: ModCtrl}:
		seq = "\x03"
	case Key{Code: 'd', Mod: ModCtrl}:
		seq = "\x04"
	case Key{Code: 'e', Mod: ModCtrl}:
		seq = "\x05"
	case Key{Code: 'f', Mod: ModCtrl}:
		seq = "\x06"
	case Key{Code: 'g', Mod: ModCtrl}:
		seq = "\x07"
	case Key{Code: 'h', Mod: ModCtrl}:
		seq = "\x08"
	case Key{Code: 'j', Mod: ModCtrl}:
		seq = "\x0a"
	case Key{Code: 'k', Mod: ModCtrl}:
		seq = "\x0b"
	case Key{Code: 'l', Mod: ModCtrl}:
		seq = "\x0c"
	case Key{Code: 'n', Mod: ModCtrl}:
		seq = "\x0e"
	case Key{Code: 'o', Mod: ModCtrl}:
		seq = "\x0f"
	case Key{Code: 'p', Mod: ModCtrl}:
		seq = "\x10"
	case Key{Code: 'q', Mod: ModCtrl}:
		seq = "\x11"
	case Key{Code: 'r', Mod: ModCtrl}:
		seq = "\x12"
	case Key{Code: 's', Mod: ModCtrl}:
		seq = "\x13"
	case Key{Code: 't', Mod: ModCtrl}:
		seq = "\x14"
	case Key{Code: 'u', Mod: ModCtrl}:
		seq = "\x15"
	case Key{Code: 'v', Mod: ModCtrl}:
		seq = "\x16"
	case Key{Code: 'w', Mod: ModCtrl}:
		seq = "\x17"
	case Key{Code: 'x', Mod: ModCtrl}:
		seq = "\x18"
	case Key{Code: 'y', Mod: ModCtrl}:
		seq = "\x19"
	case Key{Code: 'z', Mod: ModCtrl}:
		seq = "\x1a"
	case Key{Code: '\\', Mod: ModCtrl}:
		seq = "\x1c"
	case Key{Code: ']', Mod: ModCtrl}:
		seq = "\x1d"
	case Key{Code: '^', Mod: ModCtrl}:
		seq = "\x1e"
	case Key{Code: '_', Mod: ModCtrl}:
		seq = "\x1f"

	case Key{Code: KeyEnter}:
		seq = "\r"
	case Key{Code: KeyTab}:
		seq = "\t"
	case Key{Code: KeyBackspace}:
		seq = "\x7f"
	case Key{Code: KeyEscape}:
		seq = "\x1b"

	case Key{Code: KeyUp}:
		if ack {
			seq = "\x1bOA"
		} else {
			seq = "\x1b[A"
		}
	case Key{Code: KeyDown}:
		if ack {
			seq = "\x1bOB"
		} else {
			seq = "\x1b[B"
		}
	case Key{Code: KeyRight}:
		if ack {
			seq = "\x1bOC"
		} else {
			seq = "\x1b[C"
		}
	case Key{Code: KeyLeft}:
		if ack {
			seq = "\x1bOD"
		} else {
			seq = "\x1b[D"
		}

	case Key{Code: KeyInsert}:
		seq = "\x1b[2~"
	case Key{Code: KeyDelete}:
		seq = "\x1b[3~"
	case Key{Code: KeyHome}:
		seq = "\x1b[H"
	case Key{Code: KeyEnd}:
		seq = "\x1b[F"
	case Key{Code: KeyPgUp}:
		seq = "\x1b[5~"
	case Key{Code: KeyPgDown}:
		seq = "\x1b[6~"

	case Key{Code: KeyF1}:
		seq = "\x1bOP"
	case Key{Code: KeyF2}:
		seq = "\x1bOQ"
	case Key{Code: KeyF3}:
		seq = "\x1bOR"
	case Key{Code: KeyF4}:
		seq = "\x1bOS"
	case Key{Code: KeyF5}:
		seq = "\x1b[15~"
	case Key{Code: KeyF6}:
		seq = "\x1b[17~"
	case Key{Code: KeyF7}:
		seq = "\x1b[18~"
	case Key{Code: KeyF8}:
		seq = "\x1b[19~"
	case Key{Code: KeyF9}:
		seq = "\x1b[20~"
	case Key{Code: KeyF10}:
		seq = "\x1b[21~"
	case Key{Code: KeyF11}:
		seq = "\x1b[23~"
	case Key{Code: KeyF12}:
		seq = "\x1b[24~"

	case Key{Code: KeyKp0}:
		if akk {
			seq = "\x1bOp"
		} else {
			seq = "0"
		}
	case Key{Code: KeyKp1}:
		if akk {
			seq = "\x1bOq"
		} else {
			seq = "1"
		}
	case Key{Code: KeyKp2}:
		if akk {
			seq = "\x1bOr"
		} else {
			seq = "2"
		}
	case Key{Code: KeyKp3}:
		if akk {
			seq = "\x1bOs"
		} else {
			seq = "3"
		}
	case Key{Code: KeyKp4}:
		if akk {
			seq = "\x1bOt"
		} else {
			seq = "4"
		}
	case Key{Code: KeyKp5}:
		if akk {
			seq = "\x1bOu"
		} else {
			seq = "5"
		}
	case Key{Code: KeyKp6}:
		if akk {
			seq = "\x1bOv"
		} else {
			seq = "6"
		}
	case Key{Code: KeyKp7}:
		if akk {
			seq = "\x1bOw"
		} else {
			seq = "7"
		}
	case Key{Code: KeyKp8}:
		if akk {
			seq = "\x1bOx"
		} else {
			seq = "8"
		}
	case Key{Code: KeyKp9}:
		if akk {
			seq = "\x1bOy"
		} else {
			seq = "9"
		}
	case Key{Code: KeyKpEnter}:
		if akk {
			seq = "\x1bOM"
		} else {
			seq = "\r"
		}
	case Key{Code: KeyKpEqual}:
		if akk {
			seq = "\x1bOX"
		} else {
			seq = "="
		}
	case Key{Code: KeyKpMultiply}:
		if akk {
			seq = "\x1bOj"
		} else {
			seq = "*"
		}
	case Key{Code: KeyKpPlus}:
		if akk {
			seq = "\x1bOk"
		} else {
			seq = "+"
		}
	case Key{Code: KeyKpComma}:
		if akk {
			seq = "\x1bOl"
		} else {
			seq = ","
		}
	case Key{Code: KeyKpMinus}:
		if akk {
			seq = "\x1bOm"
		} else {
			seq = "-"
		}
	case Key{Code: KeyKpDecimal}:
		if akk {
			seq = "\x1bOn"
		} else {
			seq = "."
		}

	case Key{Code: KeyTab, Mod: ModShift}:
		seq = "\x1b[Z"
	}

	if k.Mod&ModAlt != 0 {
		// Handle alt-modified keys
		seq = "\x1b" + seq
	}

	t.buf.WriteString(seq) //nolint:errcheck
}

const (
	// KeyExtended is a special key code used to signify that a key event
	// contains multiple runes.
	KeyExtended = unicode.MaxRune + 1
)

// Special key symbols.
const (

	// Special keys

	KeyUp rune = KeyExtended + iota + 1
	KeyDown
	KeyRight
	KeyLeft
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
	KeyKpMultiply
	KeyKpPlus
	KeyKpComma
	KeyKpMinus
	KeyKpDecimal
	KeyKpDivide
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

	// The following are keys defined in the Kitty keyboard protocol.
	// TODO: Investigate the names of these keys
	KeyKpSep
	KeyKpUp
	KeyKpDown
	KeyKpLeft
	KeyKpRight
	KeyKpPgUp
	KeyKpPgDown
	KeyKpHome
	KeyKpEnd
	KeyKpInsert
	KeyKpDelete
	KeyKpBegin

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
	KeyF21
	KeyF22
	KeyF23
	KeyF24
	KeyF25
	KeyF26
	KeyF27
	KeyF28
	KeyF29
	KeyF30
	KeyF31
	KeyF32
	KeyF33
	KeyF34
	KeyF35
	KeyF36
	KeyF37
	KeyF38
	KeyF39
	KeyF40
	KeyF41
	KeyF42
	KeyF43
	KeyF44
	KeyF45
	KeyF46
	KeyF47
	KeyF48
	KeyF49
	KeyF50
	KeyF51
	KeyF52
	KeyF53
	KeyF54
	KeyF55
	KeyF56
	KeyF57
	KeyF58
	KeyF59
	KeyF60
	KeyF61
	KeyF62
	KeyF63

	// The following are keys defined in the Kitty keyboard protocol.
	// TODO: Investigate the names of these keys

	KeyCapsLock
	KeyScrollLock
	KeyNumLock
	KeyPrintScreen
	KeyPause
	KeyMenu

	KeyMediaPlay
	KeyMediaPause
	KeyMediaPlayPause
	KeyMediaReverse
	KeyMediaStop
	KeyMediaFastForward
	KeyMediaRewind
	KeyMediaNext
	KeyMediaPrev
	KeyMediaRecord

	KeyLowerVol
	KeyRaiseVol
	KeyMute

	KeyLeftShift
	KeyLeftAlt
	KeyLeftCtrl
	KeyLeftSuper
	KeyLeftHyper
	KeyLeftMeta
	KeyRightShift
	KeyRightAlt
	KeyRightCtrl
	KeyRightSuper
	KeyRightHyper
	KeyRightMeta
	KeyIsoLevel3Shift
	KeyIsoLevel5Shift

	// Special names in C0

	KeyBackspace = rune(ansi.DEL)
	KeyTab       = rune(ansi.HT)
	KeyEnter     = rune(ansi.CR)
	KeyReturn    = KeyEnter
	KeyEscape    = rune(ansi.ESC)
	KeyEsc       = KeyEscape

	// Special names in G0

	KeySpace = rune(ansi.SP)
)
