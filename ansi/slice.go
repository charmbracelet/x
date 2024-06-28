package ansi

import (
	"bytes"

	"github.com/charmbracelet/x/ansi/parser"
	"github.com/rivo/uniseg"
)

// Slice slices a string to a given length, starting from X cell position.
// This function is aware of ANSI escape codes and will not break them, and
// accounts for wide-characters (such as East Asians and emojis).
//
// If a string is cut in the middle of a wide character, padding (in the
// form of spaces) is inserted. This is done in order to maintain the width
// of the input string.
func Slice(s string, start int, end int) string {
	if end < start || start == end || s == "" {
		return ""
	}

	var cluster []byte
	var buf bytes.Buffer
	curPos := 0
	pstate := parser.GroundState // initial state
	b := []byte(s)

	// Here we iterate over the bytes of the string and collect printable
	// characters and runes. We also keep track of the scan position in cells.
	// Once we reach the given length, we start ignoring characters and only
	// collect ANSI escape codes until we reach the end of string.
	for i := 0; i < len(b); i++ {
		state, action := parser.Table.Transition(pstate, b[i])

		switch action {
		case parser.PrintAction:
			// Single/zero width character, fast path
			if utf8ByteLen(b[i]) <= 1 {
				if curPos >= start && curPos < end {
					buf.WriteByte(b[i])
				}
				curPos++
				continue
			}

			// This action happens when we transition to the Utf8State.
			var width int
			cluster, _, width, _ = uniseg.FirstGraphemeCluster(b[i:], -1)
			pstate = parser.GroundState
			oldPos := curPos
			curPos += width

			// When reading multiple characters, we need to advance i further.
			// We subtract one, because the loop adds that one by default.
			i += len(cluster) - 1

			// Before scope, skip
			if curPos <= start {
				continue
			}

			// Cut off at beginning, write begin padding
			if oldPos < start {
				diff := curPos - start
				for diff > 0 {
					buf.WriteByte(' ')
					diff--
				}
				continue
			}

			// Fits inside perfectly, write
			if curPos <= end {
				buf.Write(cluster)
				continue
			}

			// Cut off at end, write end padding
			if oldPos < end {
				diff := width - (curPos - end)
				for diff > 0 {
					buf.WriteByte(' ')
					diff--
				}
				continue
			}

			// Beyond scope, skip

		// Always collect ansi codes
		default:
			buf.WriteByte(b[i])
		}

		// Transition to the next state.
		pstate = state
	}

	// Ensure width matches requested
	if curPos < end-start {
		diff := (end - start) - curPos

		for diff > 0 {
			buf.WriteByte(' ')
			diff--
		}
	}

	// Return sliced string
	return buf.String()
}
