package input

import (
	"strconv"
)

// parseCsiFunc parses a CSI function key sequence.
func parseCsiFunc(params [][]uint, seq []byte) Event {
	final := string(seq[len(seq)-1])
	if len(params) == 0 || params[0][0] == 0 {
		if k, ok := csiFuncKeys[final]; ok {
			return KeyDownEvent(k)
		} else if k, ok := csiFuncKeysURxvt[final]; ok {
			return KeyDownEvent(k)
		} else if final == "Z" {
			return KeyDownEvent(Key{Sym: KeyTab, Mod: Shift})
		}
		return UnknownCsiEvent(seq)
	}

	if len(params) != 2 || params[0][0] != 1 {
		return UnknownCsiEvent(seq)
	}

	var k Key
	switch final {
	case "a", "b", "c", "d":
		k = csiFuncKeysURxvt[final]
	case "A", "B", "C", "D", "E", "F", "H", "P", "Q", "R", "S":
		k = csiFuncKeys[final]
	case "Z":
		k.Sym = KeyTab
		k.Mod |= Shift
	default:
		return UnknownCsiEvent(seq)
	}

	k.Mod |= KeyMod(params[1][0] - 1)

	return KeyDownEvent(k)
}

// parseCsiTilde parses a CSI ~ sequence.
func parseCsiTilde(params [][]uint, seq []byte) Event {
	if len(params) == 0 || seq[len(seq)-1] != '~' {
		return UnknownCsiEvent(seq)
	}

	k, ok := csiTildeKeys[strconv.FormatUint(uint64(params[0][0]), 10)]
	if !ok {
		return UnknownCsiEvent(seq)
	}

	switch len(params) {
	case 1:
		return KeyDownEvent(k)
	case 2:
		k.Mod |= KeyMod(params[1][0] - 1)
		return KeyDownEvent(k)
	}

	return UnknownCsiEvent(seq)
}

func parseCsiCarat(params [][]uint, seq []byte) Event {
	if len(params) == 0 || seq[len(seq)-1] != '^' {
		return UnknownCsiEvent(seq)
	}

	num := strconv.FormatUint(uint64(params[0][0]), 10)
	if k, ok := csiTildeKeys[num]; ok {
		k.Mod |= Ctrl
		return KeyDownEvent(k)
	} else if k, ok := csiCaratKeys[num]; ok {
		return KeyDownEvent(k)
	}

	return UnknownCsiEvent(seq)
}

func parseCsiAt(params [][]uint, seq []byte) Event {
	if len(params) == 0 || seq[len(seq)-1] != '@' {
		return UnknownCsiEvent(seq)
	}

	num := strconv.FormatUint(uint64(params[0][0]), 10)
	if k, ok := csiTildeKeys[num]; ok {
		k.Mod |= Shift | Ctrl
		return KeyDownEvent(k)
	} else if k, ok := csiAtKeys[num]; ok {
		return KeyDownEvent(k)
	}

	return UnknownCsiEvent(seq)
}

// CSI function keys
var (
	// XTerm keys
	csiFuncKeys = map[string]Key{
		"A": {Sym: KeyUp}, "B": {Sym: KeyDown},
		"C": {Sym: KeyRight}, "D": {Sym: KeyLeft},
		"E": {Sym: KeyBegin}, "F": {Sym: KeyEnd},
		"H": {Sym: KeyHome}, "P": {Sym: KeyF1},
		"Q": {Sym: KeyF2}, "R": {Sym: KeyF3},
		"S": {Sym: KeyF4},
	}
	// URXvt weird keys
	csiFuncKeysURxvt = map[string]Key{
		"a": {Sym: KeyUp, Mod: Shift}, "b": {Sym: KeyDown, Mod: Shift},
		"c": {Sym: KeyRight, Mod: Shift}, "d": {Sym: KeyLeft, Mod: Shift},
	}

	// CSI ~ sequence keys
	csiTildeKeys = map[string]Key{
		"1": {Sym: KeyHome}, "2": {Sym: KeyInsert},
		"3": {Sym: KeyDelete}, "4": {Sym: KeyEnd},
		"5": {Sym: KeyPgUp}, "6": {Sym: KeyPgDown},
		"7": {Sym: KeyHome}, "8": {Sym: KeyEnd},
		// There are no 9 and 10 keys
		"11": {Sym: KeyF1}, "12": {Sym: KeyF2},
		"13": {Sym: KeyF3}, "14": {Sym: KeyF4},
		"15": {Sym: KeyF5}, "17": {Sym: KeyF6},
		"18": {Sym: KeyF7}, "19": {Sym: KeyF8},
		"20": {Sym: KeyF9}, "21": {Sym: KeyF10},
		"23": {Sym: KeyF11}, "24": {Sym: KeyF12},
		"25": {Sym: KeyF13}, "26": {Sym: KeyF14},
		"28": {Sym: KeyF15}, "29": {Sym: KeyF16},
		"31": {Sym: KeyF17}, "32": {Sym: KeyF18},
		"33": {Sym: KeyF19}, "34": {Sym: KeyF20},
	}

	// CSI ^ sequence keys
	// Mostly used in URxvt
	csiCaratKeys = map[string]Key{
		"11": {Sym: KeyF1, Mod: Ctrl},
		"12": {Sym: KeyF2, Mod: Ctrl},
		"13": {Sym: KeyF3, Mod: Ctrl},
		"14": {Sym: KeyF4, Mod: Ctrl},
		"15": {Sym: KeyF5, Mod: Ctrl},
		"17": {Sym: KeyF6, Mod: Ctrl},
		"18": {Sym: KeyF7, Mod: Ctrl},
		"19": {Sym: KeyF8, Mod: Ctrl},
		"20": {Sym: KeyF9, Mod: Ctrl},
		"21": {Sym: KeyF10, Mod: Ctrl},
		"23": {Sym: KeyF11, Mod: Ctrl},
		"24": {Sym: KeyF12, Mod: Ctrl},
		"25": {Sym: KeyF13, Mod: Ctrl},
		"26": {Sym: KeyF14, Mod: Ctrl},
		"28": {Sym: KeyF15, Mod: Ctrl},
		"29": {Sym: KeyF16, Mod: Ctrl},
		"31": {Sym: KeyF17, Mod: Ctrl},
		"32": {Sym: KeyF18, Mod: Ctrl},
		"33": {Sym: KeyF19, Mod: Ctrl},
		"34": {Sym: KeyF20, Mod: Ctrl},
	}

	// CSI @ sequence keys
	// Mostly used in URxvt
	csiAtKeys = map[string]Key{
		"23": {Sym: KeyF11, Mod: Shift | Ctrl},
		"24": {Sym: KeyF12, Mod: Shift | Ctrl},
		"25": {Sym: KeyF13, Mod: Shift | Ctrl},
		"26": {Sym: KeyF14, Mod: Shift | Ctrl},
		"28": {Sym: KeyF15, Mod: Shift | Ctrl},
		"29": {Sym: KeyF16, Mod: Shift | Ctrl},
		"31": {Sym: KeyF17, Mod: Shift | Ctrl},
		"32": {Sym: KeyF18, Mod: Shift | Ctrl},
		"33": {Sym: KeyF19, Mod: Shift | Ctrl},
		"34": {Sym: KeyF20, Mod: Shift | Ctrl},
	}

	// CSI $ sequence keys
	// These are invalid CSI sequences that are used in URxvt
	csiDollarKeys = map[string]Key{
		"23": {Sym: KeyF11, Mod: Shift},
		"24": {Sym: KeyF12, Mod: Shift},
		"25": {Sym: KeyF13, Mod: Shift},
		"26": {Sym: KeyF14, Mod: Shift},
		"28": {Sym: KeyF15, Mod: Shift},
		"29": {Sym: KeyF16, Mod: Shift},
		"31": {Sym: KeyF17, Mod: Shift},
		"32": {Sym: KeyF18, Mod: Shift},
		"33": {Sym: KeyF19, Mod: Shift},
		"34": {Sym: KeyF20, Mod: Shift},
	}
)
