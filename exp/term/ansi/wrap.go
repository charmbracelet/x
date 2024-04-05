package ansi

import (
	"bytes"
	"unicode"
	"unicode/utf8"

	"github.com/charmbracelet/x/exp/term/ansi/parser"
	"github.com/rivo/uniseg"
)

// nbsp is a non-breaking space
const nbsp = 0xA0

// Hardwrap wraps a string or a block of text to a given line length, breaking
// word boundaries. This will preserve ANSI escape codes and will account for
// wide-characters in the string.
// When preserveSpace is true, spaces at the beginning of a line will be
// preserved.
func Hardwrap(s string, limit int, preserveSpace bool) string {
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
		case parser.PrintAction:
			if utf8ByteLen(b[i]) > 1 {
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
			}
			fallthrough
		case parser.ExecuteAction:
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
//
// Note: breakpoints must be a string of 1-cell wide rune characters.
func Wordwrap(s string, limit int, breakpoints string) string {
	if limit < 1 {
		return s
	}

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
		case parser.PrintAction:
			if utf8ByteLen(b[i]) > 1 {
				var width int
				cluster, _, width, gstate = uniseg.FirstGraphemeCluster(b[i:], gstate)
				i += len(cluster)

				r, _ := utf8.DecodeRune(cluster)
				if r != utf8.RuneError && unicode.IsSpace(r) && r != nbsp {
					addWord()
					space.WriteRune(r)
				} else if bytes.ContainsAny(cluster, breakpoints) {
					addSpace()
					addWord()
					buf.Write(cluster)
					curWidth++
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
			}
			fallthrough
		case parser.ExecuteAction:
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
			case r == '-':
				fallthrough
			case runeContainsAny(r, breakpoints):
				addSpace()
				addWord()
				buf.WriteByte(b[i])
				curWidth++
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

// Wrap wraps a string or a block of text to a given line length, breaking word
// boundaries if necessary. This will preserve ANSI escape codes and will
// account for wide-characters in the string. The breakpoints string is a list
// of characters that are considered breakpoints for word wrapping. A hyphen
// (-) is always considered a breakpoint.
//
// Note: breakpoints must be a string of 1-cell wide rune characters.
func Wrap(s string, limit int, breakpoints string) string {
	if limit < 1 {
		return s
	}

	var (
		cluster  []byte
		buf      bytes.Buffer
		word     bytes.Buffer
		space    bytes.Buffer
		curWidth int // written width of the line
		wordLen  int // word buffer len without ANSI escape codes
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
		case parser.PrintAction:
			if utf8ByteLen(b[i]) > 1 {
				var width int
				cluster, _, width, gstate = uniseg.FirstGraphemeCluster(b[i:], gstate)
				i += len(cluster)

				r, _ := utf8.DecodeRune(cluster)
				switch {
				case r != utf8.RuneError && unicode.IsSpace(r) && r != nbsp: // nbsp is a non-breaking space
					addWord()
					space.WriteRune(r)
				case bytes.ContainsAny(cluster, breakpoints):
					addSpace()
					if curWidth+wordLen+width > limit {
						word.Write(cluster)
						wordLen += width
					} else {
						addWord()
						buf.Write(cluster)
						curWidth += width
					}
				default:
					if wordLen+width > limit {
						// Hardwrap the word if it's too long
						addWord()
					}

					word.Write(cluster)
					wordLen += width

					if curWidth+wordLen+space.Len() > limit {
						addNewline()
					}
				}

				pstate = parser.GroundState
				continue
			}

			fallthrough
		case parser.ExecuteAction:
			switch r := rune(b[i]); {
			case r == '\n':
				if wordLen == 0 {
					if curWidth+space.Len() > limit {
						curWidth = 0
					} else {
						// preserve whitespaces
						buf.Write(space.Bytes())
					}
					space.Reset()
				}

				addWord()
				addNewline()
			case unicode.IsSpace(r):
				addWord()
				space.WriteRune(r)
			case r == '-':
				fallthrough
			case runeContainsAny(r, breakpoints):
				addSpace()
				if curWidth+wordLen+1 > limit {
					// We can't fit the breakpoint in the current line, treat
					// it as part of the word.
					word.WriteRune(r)
					wordLen++
				} else {
					addWord()
					buf.WriteRune(r)
					curWidth++
				}
			default:
				word.WriteRune(r)
				wordLen++

				if wordLen == limit {
					// Hardwrap the word if it's too long
					addWord()
				}

				if curWidth+wordLen+space.Len() > limit {
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

	if word.Len() != 0 {
		// Preserve ANSI wrapped spaces at the end of string
		if curWidth+space.Len() > limit {
			buf.WriteByte('\n')
		}
		addSpace()
	}
	buf.Write(word.Bytes())

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
