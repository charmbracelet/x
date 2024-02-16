package ansi

import (
	"github.com/charmbracelet/x/exp/term/input"
)

// Kitty Clipboard Control Sequences
var kittyKeyMap = map[int]input.KeySym{
	9:   input.KeyTab,
	13:  input.KeyEnter,
	27:  input.KeyEscape,
	127: input.KeyBackspace,

	57344: input.KeyEscape,
	57345: input.KeyEnter,
	57346: input.KeyTab,
	57347: input.KeyBackspace,
	57348: input.KeyInsert,
	57349: input.KeyDelete,
	57350: input.KeyLeft,
	57351: input.KeyRight,
	57352: input.KeyUp,
	57353: input.KeyDown,
	57354: input.KeyPgUp,
	57355: input.KeyPgDown,
	57356: input.KeyHome,
	57357: input.KeyEnd,
	57358: input.KeyCapsLock,
	57359: input.KeyScrollLock,
	57360: input.KeyNumLock,
	57361: input.KeyPrintScreen,
	57362: input.KeyPause,
	57363: input.KeyMenu,
	57364: input.KeyF1,
	57365: input.KeyF2,
	57366: input.KeyF3,
	57367: input.KeyF4,
	57368: input.KeyF5,
	57369: input.KeyF6,
	57370: input.KeyF7,
	57371: input.KeyF8,
	57372: input.KeyF9,
	57373: input.KeyF10,
	57374: input.KeyF11,
	57375: input.KeyF12,
	57376: input.KeyF13,
	57377: input.KeyF14,
	57378: input.KeyF15,
	57379: input.KeyF16,
	57380: input.KeyF17,
	57381: input.KeyF18,
	57382: input.KeyF19,
	57383: input.KeyF20,
	57384: input.KeyF21,
	57385: input.KeyF22,
	57386: input.KeyF23,
	57387: input.KeyF24,
	57388: input.KeyF25,
	57389: input.KeyF26,
	57390: input.KeyF27,
	57391: input.KeyF28,
	57392: input.KeyF29,
	57393: input.KeyF30,
	57394: input.KeyF31,
	57395: input.KeyF32,
	57396: input.KeyF33,
	57397: input.KeyF34,
	57398: input.KeyF35,
	57399: input.KeyKp0,
	57400: input.KeyKp1,
	57401: input.KeyKp2,
	57402: input.KeyKp3,
	57403: input.KeyKp4,
	57404: input.KeyKp5,
	57405: input.KeyKp6,
	57406: input.KeyKp7,
	57407: input.KeyKp8,
	57408: input.KeyKp9,
	57409: input.KeyKpPeriod,
	57410: input.KeyKpDiv,
	57411: input.KeyKpMul,
	57412: input.KeyKpMinus,
	57413: input.KeyKpPlus,
	57414: input.KeyKpEnter,
	57415: input.KeyKpEqual,
	57416: input.KeyKpSep,
	57417: input.KeyKpLeft,
	57418: input.KeyKpRight,
	57419: input.KeyKpUp,
	57420: input.KeyKpDown,
	57421: input.KeyKpPgUp,
	57422: input.KeyKpPgDown,
	57423: input.KeyKpHome,
	57424: input.KeyKpEnd,
	57425: input.KeyKpInsert,
	57426: input.KeyKpDelete,
	57427: input.KeyKpBegin,
	57428: input.KeyMediaPlay,
	57429: input.KeyMediaPause,
	57430: input.KeyMediaPlayPause,
	57431: input.KeyMediaReverse,
	57432: input.KeyMediaStop,
	57433: input.KeyMediaFastForward,
	57434: input.KeyMediaRewind,
	57435: input.KeyMediaNext,
	57436: input.KeyMediaPrev,
	57437: input.KeyMediaRecord,
	57438: input.KeyLowerVol,
	57439: input.KeyRaiseVol,
	57440: input.KeyMute,
	57441: input.KeyLeftShift,
	57442: input.KeyLeftCtrl,
	57443: input.KeyLeftAlt,
	57444: input.KeyLeftSuper,
	57445: input.KeyLeftHyper,
	57446: input.KeyLeftMeta,
	57447: input.KeyRightShift,
	57448: input.KeyRightCtrl,
	57449: input.KeyRightAlt,
	57450: input.KeyRightSuper,
	57451: input.KeyRightHyper,
	57452: input.KeyRightMeta,
	57453: input.KeyIsoLevel3Shift,
	57454: input.KeyIsoLevel5Shift,
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

func fromKittyMod(mod int) input.Mod {
	var m input.Mod
	if mod&kittyShift != 0 {
		m |= input.Shift
	}
	if mod&kittyAlt != 0 {
		m |= input.Alt
	}
	if mod&kittyCtrl != 0 {
		m |= input.Ctrl
	}
	if mod&kittySuper != 0 {
		m |= input.Super
	}
	if mod&kittyHyper != 0 {
		m |= input.Hyper
	}
	if mod&kittyMeta != 0 {
		m |= input.Meta
	}
	if mod&kittyCapsLock != 0 {
		m |= input.CapsLock
	}
	if mod&kittyNumLock != 0 {
		m |= input.NumLock
	}
	return m
}
