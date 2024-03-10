package ansi

import (
	"github.com/rivo/uniseg"
)

// StringWidth returns the width of a string in cells. This is the number of
// cells that the string will occupy when printed in a terminal. ANSI escape
// codes are ignored and wide characters (such as East Asians and emojis) are
// accounted for.
func StringWidth(s string) int {
	var b []byte
	var ri int
	var rw int
	pstate := GroundState

	// This implements a subset of the Parser to only collect runes and
	// printable characters.
	for i := 0; i < len(s); i++ {
		state, action := table.Transition(pstate, s[i])
		switch {
		case pstate == Utf8State:
			// During this state, keep collecting the rw bytes till we have
			// enough to form a valid rune. Then transition to the GroundState
			// and reset the counters.
			b = append(b, s[i])
			ri++
			if ri < rw {
				continue
			}
			pstate = GroundState
			ri = 0
			rw = 0
		case action == CollectAction:
			// This action happens when we transition to the Utf8State.
			if w := utf8ByteLen(s[i]); w > 1 {
				rw = w
				b = append(b, s[i])
				ri++
			}
		case action == PrintAction:
			// PrintAction is just ASCII characters
			b = append(b, s[i])
		}
		if pstate != Utf8State {
			pstate = state
		}
	}

	return uniseg.StringWidth(string(b))
}
