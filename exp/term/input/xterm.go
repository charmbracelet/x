package input

import (
	"github.com/charmbracelet/x/exp/term/ansi"
)

func parseXTermModifyOtherKeys(params [][]uint) Event {
	// XTerm modify other keys starts with ESC [ 27 ; <modifier> ; <code> ~
	mod := Mod(params[1][0] - 1)
	r := rune(params[2][0])
	k, ok := modifyOtherKeys[int(r)]
	if ok {
		k.Mod = mod
		return k
	}

	return KeyDownEvent{
		Mod:  mod,
		Rune: r,
	}
}

// CSI 27 ; <modifier> ; <code> ~ keys defined in XTerm modifyOtherKeys
var modifyOtherKeys = map[int]KeyDownEvent{
	ansi.BS:  {Sym: KeyBackspace},
	ansi.HT:  {Sym: KeyTab},
	ansi.CR:  {Sym: KeyEnter},
	ansi.ESC: {Sym: KeyEscape},
	ansi.DEL: {Sym: KeyBackspace},
}
