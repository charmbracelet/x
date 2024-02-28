package ansi

import (
	"github.com/charmbracelet/x/exp/term/ansi"
	"github.com/charmbracelet/x/exp/console/input"
)

func parseXTermModifyOtherKeys(seq []byte) input.Event {
	csi := ansi.CsiSequence(seq)
	params := ansi.Params(csi.Params())

	// XTerm modify other keys starts with ESC [ 27 ; <modifier> ; <code> ~
	if len(params) != 3 || params[0][0] != 27 {
		return UnknownCsiEvent{csi}
	}

	mod := input.Mod(params[1][0] - 1)
	r := rune(params[2][0])
	k, ok := modifyOtherKeys[int(r)]
	if ok {
		k.Mod = mod
		return k
	}

	return input.KeyEvent{
		Mod:   mod,
		Runes: []rune{r},
	}
}

// CSI 27 ; <modifier> ; <code> ~ keys defined in XTerm modifyOtherKeys
var modifyOtherKeys = map[int]input.KeyEvent{
	ansi.BS:  {Sym: input.KeyBackspace},
	ansi.HT:  {Sym: input.KeyTab},
	ansi.CR:  {Sym: input.KeyEnter},
	ansi.ESC: {Sym: input.KeyEscape},
	ansi.DEL: {Sym: input.KeyBackspace},
}
