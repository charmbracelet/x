package input

import "strconv"

func parseSs3Func(p []byte, seq []byte) Event {
	final := string(seq[len(seq)-1])

	var key KeyDownEvent
	if k, ok := ss3UrvtKeys[final]; ok {
		key = KeyDownEvent(k)
	} else if k, ok := ss3FuncKeys[final]; ok {
		key = KeyDownEvent(k)
	} else {
		return UnknownEvent(seq)
	}

	if len(p) > 0 {
		m, err := strconv.Atoi(string(p))
		if err == nil {
			key.Mod |= KeyMod(m - 1)
		}
	}

	return key
}

var (
	// SS3 URxvt keys
	ss3UrvtKeys = map[string]Key{
		"a": {Sym: KeyUp, Mod: Ctrl}, "b": {Sym: KeyDown, Mod: Ctrl},
		"c": {Sym: KeyRight, Mod: Ctrl}, "d": {Sym: KeyLeft, Mod: Ctrl},
	}

	// SS3 keypad function keys
	ss3FuncKeys = map[string]Key{
		// These are defined in XTerm
		// Taken from Foot keymap.h and XTerm modifyOtherKeys
		// https://codeberg.org/dnkl/foot/src/branch/master/keymap.h
		"A": {Sym: KeyUp}, "B": {Sym: KeyDown},
		"C": {Sym: KeyRight}, "D": {Sym: KeyLeft},
		"E": {Sym: KeyBegin}, "F": {Sym: KeyEnd},
		"H": {Sym: KeyHome},
		"P": {Sym: KeyF1}, "Q": {Sym: KeyF2},
		"R": {Sym: KeyF3}, "S": {Sym: KeyF4},
		"M": {Sym: KeyKpEnter}, "X": {Sym: KeyKpEqual},
		"j": {Sym: KeyKpMul}, "k": {Sym: KeyKpPlus},
		"l": {Sym: KeyKpComma}, "m": {Sym: KeyKpMinus},
		"n": {Sym: KeyKpPeriod}, "o": {Sym: KeyKpDiv},
		"p": {Sym: KeyKp0}, "q": {Sym: KeyKp1},
		"r": {Sym: KeyKp2}, "s": {Sym: KeyKp3},
		"t": {Sym: KeyKp4}, "u": {Sym: KeyKp5},
		"v": {Sym: KeyKp6}, "w": {Sym: KeyKp7},
		"x": {Sym: KeyKp8}, "y": {Sym: KeyKp9},
	}
)
