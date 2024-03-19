package input

import (
	"github.com/charmbracelet/x/exp/term/ansi"
)

func parseXTermModifyOtherKeys(params [][]uint) Event {
	// XTerm modify other keys starts with ESC [ 27 ; <modifier> ; <code> ~
	mod := KeyMod(params[1][0] - 1)
	r := rune(params[2][0])
	k, ok := modifyOtherKeys[int(r)]
	if ok {
		k.Mod = mod
		return KeyDownEvent(k)
	}

	return KeyDownEvent{
		Mod:  mod,
		Rune: r,
	}
}

// CSI 27 ; <modifier> ; <code> ~ keys defined in XTerm modifyOtherKeys
var modifyOtherKeys = map[int]Key{
	ansi.BS:  {Sym: KeyBackspace},
	ansi.HT:  {Sym: KeyTab},
	ansi.CR:  {Sym: KeyEnter},
	ansi.ESC: {Sym: KeyEscape},
	ansi.DEL: {Sym: KeyBackspace},
}

// ModifyOtherKeysEvent represents a modifyOtherKeys event.
//
//	0: disable
//	1: enable mode 1
//	2: enable mode 2
//
// See: https://invisible-island.net/xterm/ctlseqs/ctlseqs.html#h3-Functions-using-CSI-_-ordered-by-the-final-character_s_
// See: https://invisible-island.net/xterm/manpage/xterm.html#VT100-Widget-Resources:modifyOtherKeys
type ModifyOtherKeysEvent uint8
