package ansi

import (
	"bytes"
	"unicode"
	"unicode/utf8"

	. "github.com/charmbracelet/x/exp/term/ansi/parser"
	"github.com/rivo/uniseg"
)

// Wrap wraps a string or a block of text to a given line length, breaking word
// boundaries. This will preserve ANSI escape codes and will account for
// wide-characters in the string.
// When preserveSpace is true, spaces at the beginning of a line will be
// preserved.
func Wrap(s string, limit int, preserveSpace bool) string {
	if limit < 1 {
		return s
	}

	var (
		cluster      []byte
		buf          bytes.Buffer
		curWidth     int
		forceNewline bool
		gstate       = -1
		pstate       = GroundState // initial state
		b            = []byte(s)
		i            = 0
	)

	addNewline := func() {
		buf.WriteByte('\n')
		curWidth = 0
	}

	for i < len(b) {
		state, action := Table.Transition(pstate, b[i])
		// log.Printf("pstate: %s, state: %s, action: %s, code: %q", StateNames[pstate], StateNames[state], ActionNames[action], b[i])

		switch action {
		case CollectAction:
			if w := utf8ByteLen(b[i]); w > 1 {
				var width int
				cluster, _, width, gstate = uniseg.FirstGraphemeCluster(b[i:], gstate)
				// log.Printf("cluster: %q, width: %d, curWidth: %d, buf: %q", string(cluster), width, curWidth, b[i:])
				i += len(cluster)

				if curWidth+width > limit {
					addNewline()
				}
				if !preserveSpace && curWidth == 0 && len(cluster) <= 4 {
					// Skip spaces at the beginning of a line
					if r, _ := utf8.DecodeRune(cluster); r != utf8.RuneError && unicode.IsSpace(r) {
						pstate = GroundState
						continue
					}
				}
				buf.Write(cluster)
				curWidth += width
				gstate = -1 // reset grapheme state otherwise, width calculation might be off

				pstate = GroundState
				continue
			} else {
				// Collect sequence intermediate bytes
				buf.WriteByte(b[i])
			}
		case PrintAction, ExecuteAction:
			if b[i] == '\n' {
				addNewline()
				forceNewline = false
				break
			}

			if curWidth+1 > limit {
				addNewline()
				forceNewline = true
			}

			// Skip spaces at the beginning of a line
			if curWidth == 0 {
				if !preserveSpace && forceNewline && unicode.IsSpace(rune(b[i])) {
					break
				}
				forceNewline = false
			}

			buf.WriteByte(b[i])
			curWidth++
		default:
			buf.WriteByte(b[i])
		}
		// log.Printf("curWidth: %d, limit: %d", curWidth, limit)

		// We manage the UTF8 state separately manually above.
		if pstate != Utf8State {
			pstate = state
		}
		i++
	}

	return buf.String()
}

// Wordwrap wraps a string or a block of text to a given line length, not
// breaking word boundaries. This will preserve ANSI escape codes and will
// account for wide-characters in the string.
// The breakpoints string is a list of characters that are considered
// breakpoints for word wrapping. A hyphen (-) is always considered a
// breakpoint.
func Wordwrap(s string, limit int, breakpoints string) string {
	if limit < 1 {
		return s
	}

	// Add a hyphen to the breakpoints
	breakpoints += "-"

	var (
		cluster  []byte
		buf      bytes.Buffer
		word     bytes.Buffer
		space    bytes.Buffer
		curWidth int
		wordLen  int
		gstate   = -1
		pstate   = GroundState // initial state
		b        = []byte(s)
		i        = 0
	)

	addSpace := func() {
		curWidth += space.Len()
		buf.Write(space.Bytes())
		space.Reset()
	}

	addWord := func() {
		if word.Len() == 0 {
			return
		}
		addSpace()
		curWidth += wordLen
		buf.Write(word.Bytes())
		word.Reset()
		wordLen = 0
	}

	addNewline := func() {
		buf.WriteByte('\n')
		curWidth = 0
		space.Reset()
	}

	for i < len(b) {
		state, action := Table.Transition(pstate, b[i])
		// log.Printf("pstate: %s, state: %s, action: %s, code: %q", StateNames[pstate], StateNames[state], ActionNames[action], b[i])
		// log.Printf("curWidth: %d, limit: %d", curWidth, limit)
		// log.Printf("word: %q, wordLen: %d, space: %q", word.String(), wordLen, space.String())

		switch action {
		case CollectAction:
			if w := utf8ByteLen(b[i]); w > 1 {
				var width int
				cluster, _, width, gstate = uniseg.FirstGraphemeCluster(b[i:], gstate)
				// log.Printf("cluster: %q, width: %d, buf: %q", cluster, width, b[i:])
				i += len(cluster)

				r, _ := utf8.DecodeRune(cluster)
				if r != utf8.RuneError && unicode.IsSpace(r) {
					addWord()
					space.WriteRune(r)
				} else if bytes.ContainsAny(cluster, breakpoints) {
					addSpace()
					addWord()
					buf.Write(cluster)
				} else {
					word.Write(cluster)
					wordLen += width
					if curWidth+space.Len()+wordLen > limit &&
						wordLen < limit {
						addNewline()
					}
				}

				pstate = GroundState
				continue
			} else {
				// Collect sequence intermediate bytes
				word.WriteByte(b[i])
			}
		case PrintAction, ExecuteAction:
			r := rune(b[i])
			if r == '\n' {
				if wordLen == 0 {
					if curWidth+space.Len() > limit {
						curWidth = 0
					} else {
						buf.Write(space.Bytes())
					}
					space.Reset()
				}

				addWord()
				addNewline()
			} else if unicode.IsSpace(r) {
				addWord()
				space.WriteByte(b[i])
			} else if runeContainsAny(r, breakpoints) {
				addSpace()
				addWord()
				buf.WriteByte(b[i])
			} else {
				word.WriteByte(b[i])
				wordLen++
				if curWidth+space.Len()+wordLen > limit &&
					wordLen < limit {
					addNewline()
				}
			}

		default:
			word.WriteByte(b[i])
		}
		// We manage the UTF8 state separately manually above.
		if pstate != Utf8State {
			pstate = state
		}
		i++
	}

	addWord()

	return buf.String()
}

func runeContainsAny(r rune, s string) bool {
	for _, c := range s {
		if c == r {
			return true
		}
	}
	return false
}
