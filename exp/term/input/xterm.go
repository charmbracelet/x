package input

import (
	"github.com/charmbracelet/x/exp/term/ansi"
)

func parseXTermModifyOtherKeys(seq []byte) Event {
	csi := ansi.CsiSequence(seq)
	params := ansi.Params(csi.Params())

	// XTerm modify other keys starts with ESC [ 27 ; <modifier> ; <code> ~
	if len(params) != 3 || params[0][0] != 27 {
		return UnknownCsiEvent{csi}
	}

	mod := Mod(params[1][0] - 1)
	r := rune(params[2][0])
	k, ok := modifyOtherKeys[int(r)]
	if ok {
		k.Mod = mod
		return k
	}

	return KeyEvent{
		Mod:   mod,
		Runes: []rune{r},
	}
}

// CSI 27 ; <modifier> ; <code> ~ keys defined in XTerm modifyOtherKeys
var modifyOtherKeys = map[int]KeyEvent{
	ansi.BS:  {Sym: KeyBackspace},
	ansi.HT:  {Sym: KeyTab},
	ansi.CR:  {Sym: KeyEnter},
	ansi.ESC: {Sym: KeyEscape},
	ansi.DEL: {Sym: KeyBackspace},
}
