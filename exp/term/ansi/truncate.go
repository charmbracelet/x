package ansi

import (
	"bytes"

	. "github.com/charmbracelet/x/exp/term/ansi/parser"
	"github.com/rivo/uniseg"
)

// Truncate truncates a string to a given length, adding a tail to the
// end if the string is longer than the given length.
// This function is aware of ANSI escape codes and will not break them, and
// accounts for wide-characters (such as East Asians and emojis).
func Truncate(s string, length int, tail string) string {
	tw := StringWidth(tail)
	length -= tw
	if length < 0 {
		return ""
	}

	var cluster []byte
	var buf bytes.Buffer
	curWidth := 0
	ignoring := false
	gstate := -1
	pstate := GroundState // initial state
	b := []byte(s)
	i := 0

	// Here we iterate over the bytes of the string and collect printable
	// characters and runes. We also keep track of the width of the string
	// in cells.
	// Once we reach the given length, we start ignoring characters and only
	// collect ANSI escape codes until we reach the end of string.
	for i < len(b) {
		state, action := Table.Transition(pstate, b[i])
		// log.Printf("pstate: %s, state: %s, action: %s, code: %q", StateNames[pstate], StateNames[state], ActionNames[action], s[i])

		switch action {
		case CollectAction:
			// This action happens when we transition to the Utf8State.
			if w := utf8ByteLen(b[i]); w > 1 {
				var width int
				cluster, _, width, gstate = uniseg.FirstGraphemeCluster(b[i:], gstate)

				// log.Printf("cluster: %q, width: %d, curWidth: %d", string(cluster), width, curWidth)

				// increment the index by the length of the cluster
				i += len(cluster)

				// Are we ignoring? Skip to the next byte
				if ignoring {
					continue
				}

				// Is this gonna be too wide?
				// If so write the tail and stop collecting.
				if curWidth+width >= length && !ignoring {
					ignoring = true
					buf.WriteString(tail)
				}

				if curWidth+width > length {
					continue
				}

				curWidth += width
				for _, r := range cluster {
					buf.WriteByte(r)
				}

				// Done collecting, now we're back in the ground state.
				pstate = GroundState
				continue
			} else {
				// Collecting sequence intermediate bytes
				buf.WriteByte(b[i])
			}
		case PrintAction:
			// Is this gonna be too wide?
			// If so write the tail and stop collecting.
			if curWidth >= length && !ignoring {
				ignoring = true
				buf.WriteString(tail)
			}

			// Skip to the next byte if we're ignoring
			if ignoring {
				i++
				continue
			}

			// collects printable ASCII
			curWidth++
			fallthrough
		default:
			buf.WriteByte(b[i])
			i++
		}

		// Transition to the next state.
		pstate = state

		// log.Printf("buf: %q, curWidth: %d, ignoring: %v", buf.String(), curWidth, ignoring)

		// Once we reach the given length, we start ignoring runes and write
		// the tail to the buffer.
		if curWidth > length && !ignoring {
			ignoring = true
			buf.WriteString(tail)
		}
	}

	return buf.String()
}
