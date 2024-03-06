package input

import (
	"unicode/utf8"

	"github.com/charmbracelet/x/exp/term/ansi"
)

// KittyKeyboardEvent represents Kitty keyboard progressive enhancement flags.
type KittyKeyboardEvent int

// IsDisambiguateEscapeCodes returns true if the DisambiguateEscapeCodes flag is set.
func (e KittyKeyboardEvent) IsDisambiguateEscapeCodes() bool {
	return e&ansi.KittyDisambiguateEscapeCodes != 0
}

// IsReportEventTypes returns true if the ReportEventTypes flag is set.
func (e KittyKeyboardEvent) IsReportEventTypes() bool {
	return e&ansi.KittyReportEventTypes != 0
}

// IsReportAlternateKeys returns true if the ReportAlternateKeys flag is set.
func (e KittyKeyboardEvent) IsReportAlternateKeys() bool {
	return e&ansi.KittyReportAlternateKeys != 0
}

// IsReportAllKeys returns true if the ReportAllKeys flag is set.
func (e KittyKeyboardEvent) IsReportAllKeys() bool {
	return e&ansi.KittyReportAllKeys != 0
}

// IsReportAssociatedKeys returns true if the ReportAssociatedKeys flag is set.
func (e KittyKeyboardEvent) IsReportAssociatedKeys() bool {
	return e&ansi.KittyReportAssociatedKeys != 0
}

// String implements fmt.Stringer.
func (e KittyKeyboardEvent) String() string {
	s := "Flags:"
	if e == 0 {
		return s + " none"
	}
	if e.IsDisambiguateEscapeCodes() {
		s += " DisambiguateEscapeCodes"
	}
	if e.IsReportEventTypes() {
		s += " ReportEventTypes"
	}
	if e.IsReportAlternateKeys() {
		s += " ReportAlternateKeys"
	}
	if e.IsReportAllKeys() {
		s += " ReportAllKeys"
	}
	if e.IsReportAssociatedKeys() {
		s += " ReportAssociatedKeys"
	}
	return s
}

// Kitty Clipboard Control Sequences
var kittyKeyMap = map[int]KeySym{
	ansi.BS:  KeyBackspace,
	ansi.HT:  KeyTab,
	ansi.CR:  KeyEnter,
	ansi.ESC: KeyEscape,
	ansi.DEL: KeyBackspace,

	57344: KeyEscape,
	57345: KeyEnter,
	57346: KeyTab,
	57347: KeyBackspace,
	57348: KeyInsert,
	57349: KeyDelete,
	57350: KeyLeft,
	57351: KeyRight,
	57352: KeyUp,
	57353: KeyDown,
	57354: KeyPgUp,
	57355: KeyPgDown,
	57356: KeyHome,
	57357: KeyEnd,
	57358: KeyCapsLock,
	57359: KeyScrollLock,
	57360: KeyNumLock,
	57361: KeyPrintScreen,
	57362: KeyPause,
	57363: KeyMenu,
	57364: KeyF1,
	57365: KeyF2,
	57366: KeyF3,
	57367: KeyF4,
	57368: KeyF5,
	57369: KeyF6,
	57370: KeyF7,
	57371: KeyF8,
	57372: KeyF9,
	57373: KeyF10,
	57374: KeyF11,
	57375: KeyF12,
	57376: KeyF13,
	57377: KeyF14,
	57378: KeyF15,
	57379: KeyF16,
	57380: KeyF17,
	57381: KeyF18,
	57382: KeyF19,
	57383: KeyF20,
	57384: KeyF21,
	57385: KeyF22,
	57386: KeyF23,
	57387: KeyF24,
	57388: KeyF25,
	57389: KeyF26,
	57390: KeyF27,
	57391: KeyF28,
	57392: KeyF29,
	57393: KeyF30,
	57394: KeyF31,
	57395: KeyF32,
	57396: KeyF33,
	57397: KeyF34,
	57398: KeyF35,
	57399: KeyKp0,
	57400: KeyKp1,
	57401: KeyKp2,
	57402: KeyKp3,
	57403: KeyKp4,
	57404: KeyKp5,
	57405: KeyKp6,
	57406: KeyKp7,
	57407: KeyKp8,
	57408: KeyKp9,
	57409: KeyKpPeriod,
	57410: KeyKpDiv,
	57411: KeyKpMul,
	57412: KeyKpMinus,
	57413: KeyKpPlus,
	57414: KeyKpEnter,
	57415: KeyKpEqual,
	57416: KeyKpSep,
	57417: KeyKpLeft,
	57418: KeyKpRight,
	57419: KeyKpUp,
	57420: KeyKpDown,
	57421: KeyKpPgUp,
	57422: KeyKpPgDown,
	57423: KeyKpHome,
	57424: KeyKpEnd,
	57425: KeyKpInsert,
	57426: KeyKpDelete,
	57427: KeyKpBegin,
	57428: KeyMediaPlay,
	57429: KeyMediaPause,
	57430: KeyMediaPlayPause,
	57431: KeyMediaReverse,
	57432: KeyMediaStop,
	57433: KeyMediaFastForward,
	57434: KeyMediaRewind,
	57435: KeyMediaNext,
	57436: KeyMediaPrev,
	57437: KeyMediaRecord,
	57438: KeyLowerVol,
	57439: KeyRaiseVol,
	57440: KeyMute,
	57441: KeyLeftShift,
	57442: KeyLeftCtrl,
	57443: KeyLeftAlt,
	57444: KeyLeftSuper,
	57445: KeyLeftHyper,
	57446: KeyLeftMeta,
	57447: KeyRightShift,
	57448: KeyRightCtrl,
	57449: KeyRightAlt,
	57450: KeyRightSuper,
	57451: KeyRightHyper,
	57452: KeyRightMeta,
	57453: KeyIsoLevel3Shift,
	57454: KeyIsoLevel5Shift,
}

const (
	kittyShift = 1 << iota
	kittyAlt
	kittyCtrl
	kittySuper
	kittyHyper
	kittyMeta
	kittyCapsLock
	kittyNumLock
)

func fromKittyMod(mod int) Mod {
	var m Mod
	if mod&kittyShift != 0 {
		m |= Shift
	}
	if mod&kittyAlt != 0 {
		m |= Alt
	}
	if mod&kittyCtrl != 0 {
		m |= Ctrl
	}
	if mod&kittySuper != 0 {
		m |= Super
	}
	if mod&kittyHyper != 0 {
		m |= Hyper
	}
	if mod&kittyMeta != 0 {
		m |= Meta
	}
	if mod&kittyCapsLock != 0 {
		m |= CapsLock
	}
	if mod&kittyNumLock != 0 {
		m |= NumLock
	}
	return m
}

func parseKittyKeyboard(params [][]uint) KeyEvent {
	key := KeyEvent{}
	if len(params) > 0 {
		code := int(params[0][0])
		if sym, ok := kittyKeyMap[code]; ok {
			key.Sym = sym
		} else {
			r := rune(code)
			if !utf8.ValidRune(r) {
				r = utf8.RuneError
			}
			key.Rune = r
			if len(params[0]) > 1 {
				al := rune(params[0][1])
				if utf8.ValidRune(al) {
					key.AltRune = al
				}
			}
		}
	}
	if len(params) > 1 {
		mod := int(params[1][0])
		if mod > 1 {
			key.Mod = fromKittyMod(int(params[1][0] - 1))
		}
		if len(params[1]) > 1 {
			switch int(params[1][1]) {
			case 0, 1:
				key.Action = KeyPress
			case 2:
				key.Action = KeyRepeat
			case 3:
				key.Action = KeyRelease
			}
		}
	}
	if len(params) > 2 {
		r := rune(params[2][0])
		if !utf8.ValidRune(r) {
			r = utf8.RuneError
		}
		key.AltRune = r
	}
	return key
}
