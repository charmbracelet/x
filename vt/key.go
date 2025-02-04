package vt

import (
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/input"
)

// KeyMod represents a key modifier.
type KeyMod = input.KeyMod

// Modifier keys.
const (
	ModShift = input.ModShift
	ModAlt   = input.ModAlt
	ModCtrl  = input.ModCtrl
	ModMeta  = input.ModMeta
)

// Key represents a key press event.
type Key = input.Key

// SendKey returns the default key map.
func (t *Terminal) SendKey(k Key) {
	var seq string

	ack := t.isModeSet(ansi.CursorKeysMode)    // Application cursor keys mode
	akk := t.isModeSet(ansi.NumericKeypadMode) // Application keypad keys mode

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

// Key codes.
const (
	KeyExtended         = input.KeyExtended
	KeyUp               = input.KeyUp
	KeyDown             = input.KeyDown
	KeyRight            = input.KeyRight
	KeyLeft             = input.KeyLeft
	KeyBegin            = input.KeyBegin
	KeyFind             = input.KeyFind
	KeyInsert           = input.KeyInsert
	KeyDelete           = input.KeyDelete
	KeySelect           = input.KeySelect
	KeyPgUp             = input.KeyPgUp
	KeyPgDown           = input.KeyPgDown
	KeyHome             = input.KeyHome
	KeyEnd              = input.KeyEnd
	KeyKpEnter          = input.KeyKpEnter
	KeyKpEqual          = input.KeyKpEqual
	KeyKpMultiply       = input.KeyKpMultiply
	KeyKpPlus           = input.KeyKpPlus
	KeyKpComma          = input.KeyKpComma
	KeyKpMinus          = input.KeyKpMinus
	KeyKpDecimal        = input.KeyKpDecimal
	KeyKpDivide         = input.KeyKpDivide
	KeyKp0              = input.KeyKp0
	KeyKp1              = input.KeyKp1
	KeyKp2              = input.KeyKp2
	KeyKp3              = input.KeyKp3
	KeyKp4              = input.KeyKp4
	KeyKp5              = input.KeyKp5
	KeyKp6              = input.KeyKp6
	KeyKp7              = input.KeyKp7
	KeyKp8              = input.KeyKp8
	KeyKp9              = input.KeyKp9
	KeyKpSep            = input.KeyKpSep
	KeyKpUp             = input.KeyKpUp
	KeyKpDown           = input.KeyKpDown
	KeyKpLeft           = input.KeyKpLeft
	KeyKpRight          = input.KeyKpRight
	KeyKpPgUp           = input.KeyKpPgUp
	KeyKpPgDown         = input.KeyKpPgDown
	KeyKpHome           = input.KeyKpHome
	KeyKpEnd            = input.KeyKpEnd
	KeyKpInsert         = input.KeyKpInsert
	KeyKpDelete         = input.KeyKpDelete
	KeyKpBegin          = input.KeyKpBegin
	KeyF1               = input.KeyF1
	KeyF2               = input.KeyF2
	KeyF3               = input.KeyF3
	KeyF4               = input.KeyF4
	KeyF5               = input.KeyF5
	KeyF6               = input.KeyF6
	KeyF7               = input.KeyF7
	KeyF8               = input.KeyF8
	KeyF9               = input.KeyF9
	KeyF10              = input.KeyF10
	KeyF11              = input.KeyF11
	KeyF12              = input.KeyF12
	KeyF13              = input.KeyF13
	KeyF14              = input.KeyF14
	KeyF15              = input.KeyF15
	KeyF16              = input.KeyF16
	KeyF17              = input.KeyF17
	KeyF18              = input.KeyF18
	KeyF19              = input.KeyF19
	KeyF20              = input.KeyF20
	KeyF21              = input.KeyF21
	KeyF22              = input.KeyF22
	KeyF23              = input.KeyF23
	KeyF24              = input.KeyF24
	KeyF25              = input.KeyF25
	KeyF26              = input.KeyF26
	KeyF27              = input.KeyF27
	KeyF28              = input.KeyF28
	KeyF29              = input.KeyF29
	KeyF30              = input.KeyF30
	KeyF31              = input.KeyF31
	KeyF32              = input.KeyF32
	KeyF33              = input.KeyF33
	KeyF34              = input.KeyF34
	KeyF35              = input.KeyF35
	KeyF36              = input.KeyF36
	KeyF37              = input.KeyF37
	KeyF38              = input.KeyF38
	KeyF39              = input.KeyF39
	KeyF40              = input.KeyF40
	KeyF41              = input.KeyF41
	KeyF42              = input.KeyF42
	KeyF43              = input.KeyF43
	KeyF44              = input.KeyF44
	KeyF45              = input.KeyF45
	KeyF46              = input.KeyF46
	KeyF47              = input.KeyF47
	KeyF48              = input.KeyF48
	KeyF49              = input.KeyF49
	KeyF50              = input.KeyF50
	KeyF51              = input.KeyF51
	KeyF52              = input.KeyF52
	KeyF53              = input.KeyF53
	KeyF54              = input.KeyF54
	KeyF55              = input.KeyF55
	KeyF56              = input.KeyF56
	KeyF57              = input.KeyF57
	KeyF58              = input.KeyF58
	KeyF59              = input.KeyF59
	KeyF60              = input.KeyF60
	KeyF61              = input.KeyF61
	KeyF62              = input.KeyF62
	KeyF63              = input.KeyF63
	KeyCapsLock         = input.KeyCapsLock
	KeyScrollLock       = input.KeyScrollLock
	KeyNumLock          = input.KeyNumLock
	KeyPrintScreen      = input.KeyPrintScreen
	KeyPause            = input.KeyPause
	KeyMenu             = input.KeyMenu
	KeyMediaPlay        = input.KeyMediaPlay
	KeyMediaPause       = input.KeyMediaPause
	KeyMediaPlayPause   = input.KeyMediaPlayPause
	KeyMediaReverse     = input.KeyMediaReverse
	KeyMediaStop        = input.KeyMediaStop
	KeyMediaFastForward = input.KeyMediaFastForward
	KeyMediaRewind      = input.KeyMediaRewind
	KeyMediaNext        = input.KeyMediaNext
	KeyMediaPrev        = input.KeyMediaPrev
	KeyMediaRecord      = input.KeyMediaRecord
	KeyLowerVol         = input.KeyLowerVol
	KeyRaiseVol         = input.KeyRaiseVol
	KeyMute             = input.KeyMute
	KeyLeftShift        = input.KeyLeftShift
	KeyLeftAlt          = input.KeyLeftAlt
	KeyLeftCtrl         = input.KeyLeftCtrl
	KeyLeftSuper        = input.KeyLeftSuper
	KeyLeftHyper        = input.KeyLeftHyper
	KeyLeftMeta         = input.KeyLeftMeta
	KeyRightShift       = input.KeyRightShift
	KeyRightAlt         = input.KeyRightAlt
	KeyRightCtrl        = input.KeyRightCtrl
	KeyRightSuper       = input.KeyRightSuper
	KeyRightHyper       = input.KeyRightHyper
	KeyRightMeta        = input.KeyRightMeta
	KeyIsoLevel3Shift   = input.KeyIsoLevel3Shift
	KeyIsoLevel5Shift   = input.KeyIsoLevel5Shift
	KeyBackspace        = input.KeyBackspace
	KeyTab              = input.KeyTab
	KeyEnter            = input.KeyEnter
	KeyReturn           = input.KeyReturn
	KeyEscape           = input.KeyEscape
	KeyEsc              = input.KeyEsc
	KeySpace            = input.KeySpace
)
