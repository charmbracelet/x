package ansi

import (
	"bytes"
	"unicode"
	"unicode/utf8"

	"github.com/charmbracelet/x/exp/term/ansi/parser"
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
		pstate       = parser.GroundState // initial state
		b            = []byte(s)
	)

	addNewline := func() {
		buf.WriteByte('\n')
		curWidth = 0
	}

	i := 0
	for i < len(b) {
		state, action := parser.Table.Transition(pstate, b[i])

		switch action {
		case parser.CollectAction:
			if w := utf8ByteLen(b[i]); w <= 1 {
				// Collect sequence intermediate bytes
				buf.WriteByte(b[i])
				break
			}

			var width int
			cluster, _, width, gstate = uniseg.FirstGraphemeCluster(b[i:], gstate)
			i += len(cluster)

			if curWidth+width > limit {
				addNewline()
			}
			if !preserveSpace && curWidth == 0 && len(cluster) <= 4 {
				// Skip spaces at the beginning of a line
				if r, _ := utf8.DecodeRune(cluster); r != utf8.RuneError && unicode.IsSpace(r) {
					pstate = parser.GroundState
					continue
				}
			}

			buf.Write(cluster)
			curWidth += width
			gstate = -1 // reset grapheme state otherwise, width calculation might be off
			pstate = parser.GroundState
			continue
		case parser.PrintAction, parser.ExecuteAction:
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

		// We manage the UTF8 state separately manually above.
		if pstate != parser.Utf8State {
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
		pstate   = parser.GroundState // initial state
		b        = []byte(s)
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

	i := 0
	for i < len(b) {
		state, action := parser.Table.Transition(pstate, b[i])

		switch action {
		case parser.CollectAction:
			if w := utf8ByteLen(b[i]); w <= 1 {
				// Collect sequence intermediate bytes
				word.WriteByte(b[i])
				break
			}

			var width int
			cluster, _, width, gstate = uniseg.FirstGraphemeCluster(b[i:], gstate)
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

			pstate = parser.GroundState
			continue
		case parser.PrintAction, parser.ExecuteAction:
			r := rune(b[i])
			switch {
			case r == '\n':
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
			case unicode.IsSpace(r):
				addWord()
				space.WriteByte(b[i])
			case runeContainsAny(r, breakpoints):
				addSpace()
				addWord()
				buf.WriteByte(b[i])
			default:
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
		if pstate != parser.Utf8State {
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
