package ansi

import (
	"bytes"
	"io"
	"unicode"
	"unicode/utf8"

	"github.com/charmbracelet/x/ansi/parser"
	"github.com/mattn/go-runewidth"
	"github.com/rivo/uniseg"
)

// nbsp is a non-breaking space
const nbsp = 0xA0

// Hardwrap wraps a string or a block of text to a given line length, breaking
// word boundaries. This will preserve ANSI escape codes and will account for
// wide-characters in the string.
// When preserveSpace is true, spaces at the beginning of a line will be
// preserved.
// This treats the text as a sequence of graphemes.
func Hardwrap(s string, limit int, preserveSpace bool) string {
	return hardwrap(GraphemeWidth, s, limit, preserveSpace)
}

// HardwrapWc wraps a string or a block of text to a given line length, breaking
// word boundaries. This will preserve ANSI escape codes and will account for
// wide-characters in the string.
// When preserveSpace is true, spaces at the beginning of a line will be
// preserved.
// This treats the text as a sequence of wide characters and runes.
func HardwrapWc(s string, limit int, preserveSpace bool) string {
	return hardwrap(WcWidth, s, limit, preserveSpace)
}

func hardwrap(m Method, s string, limit int, preserveSpace bool) string {
	if limit < 1 {
		return s
	}

	var (
		cluster      []byte
		buf          bytes.Buffer
		curWidth     int
		forceNewline bool
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
		if state == parser.Utf8State {
			var width int
			cluster, _, width, _ = uniseg.FirstGraphemeCluster(b[i:], -1)
			if m == WcWidth {
				width = runewidth.StringWidth(string(cluster))
			}
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
			pstate = parser.GroundState
			continue
		}

		switch action {
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
			if action == parser.PrintAction {
				curWidth++
			}
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
//
// This treats the text as a sequence of graphemes.
func Wordwrap(s string, limit int, breakpoints string) string {
	return wordwrap(GraphemeWidth, s, limit, breakpoints)
}

// WordwrapWc wraps a string or a block of text to a given line length, not
// breaking word boundaries. This will preserve ANSI escape codes and will
// account for wide-characters in the string.
// The breakpoints string is a list of characters that are considered
// breakpoints for word wrapping. A hyphen (-) is always considered a
// breakpoint.
//
// Note: breakpoints must be a string of 1-cell wide rune characters.
//
// This treats the text as a sequence of wide characters and runes.
func WordwrapWc(s string, limit int, breakpoints string) string {
	return wordwrap(WcWidth, s, limit, breakpoints)
}

// WordwrapWriter wraps a string or a block of text to a given line length, not
// breaking word boundaries. This will preserve ANSI escape codes and will
// account for wide-characters in the string.
// The breakpoints string is a list of characters that are considered
// breakpoints for word wrapping. A hyphen (-) is always considered a
// breakpoint.
type WordwrapWriter struct {
	Limit       int
	Breakpoints []rune
	Method      Method

	w        io.Writer
	cluster  []byte
	word     bytes.Buffer
	space    bytes.Buffer
	curWidth int
	wordLen  int
	pstate   byte // initial state
}

func (w *WordwrapWriter) addSpace() {
	w.curWidth += w.space.Len()
	_, _ = w.w.Write(w.space.Bytes())
	w.space.Reset()
}

func (w *WordwrapWriter) addWord() {
	if w.word.Len() == 0 {
		return
	}

	w.addSpace()
	w.curWidth += w.wordLen
	_, _ = w.w.Write(w.word.Bytes())
	w.word.Reset()
	w.wordLen = 0
}

func (w *WordwrapWriter) addNewline() {
	_, _ = w.w.Write([]byte("\n"))
	w.curWidth = 0
	w.space.Reset()
}

func wordwrap(m Method, s string, limit int, breakpoints string) string {
	var buf bytes.Buffer
	ww := NewWordwrapWriter(&buf, limit)
	ww.Method = m
	if len(breakpoints) > 0 {
		ww.Breakpoints = []rune(breakpoints)
	}
	_, _ = io.WriteString(ww, s)
	return buf.String()
}

// NewWordwrapWriter returns a new WordwrapWriter that writes to w.
func NewWordwrapWriter(w io.Writer, limit int) *WordwrapWriter {
	ww := &WordwrapWriter{Limit: limit}
	ww.w = w
	return ww
}

// Write writes the content of p into the internal buffer.
func (w *WordwrapWriter) Write(b []byte) (n int, err error) {
	if w.Limit < 1 {
		return w.w.Write(b)
	}

	i := 0
	for i < len(b) {
		state, action := parser.Table.Transition(w.pstate, b[i])
		if state == parser.Utf8State {
			var width int
			w.cluster, _, width, _ = uniseg.FirstGraphemeCluster(b[i:], -1)
			if w.Method == WcWidth {
				width = runewidth.StringWidth(string(w.cluster))
			}
			i += len(w.cluster)

			r, _ := utf8.DecodeRune(w.cluster)
			if r != utf8.RuneError && unicode.IsSpace(r) && r != nbsp {
				w.addWord()
				w.space.WriteRune(r)
			} else if bytes.ContainsAny(w.cluster, string(w.Breakpoints)) {
				w.addSpace()
				w.addWord()
				_, _ = w.w.Write(w.cluster)
				w.curWidth++
			} else {
				w.word.Write(w.cluster)
				w.wordLen += width
				if w.curWidth+w.space.Len()+w.wordLen > w.Limit &&
					w.wordLen < w.Limit {
					w.addNewline()
				}
			}

			w.pstate = parser.GroundState
			continue
		}

		switch action {
		case parser.PrintAction, parser.ExecuteAction:
			r := rune(b[i])
			switch {
			case r == '\n':
				if w.wordLen == 0 {
					if w.curWidth+w.space.Len() > w.Limit {
						w.curWidth = 0
					} else {
						_, _ = w.w.Write(w.space.Bytes())
					}
					w.space.Reset()
				}

				w.addWord()
				w.addNewline()
			case unicode.IsSpace(r):
				w.addWord()
				w.space.WriteByte(b[i])
			case r == '-':
				fallthrough
			case runeContainsAny(r, w.Breakpoints):
				w.addSpace()
				w.addWord()
				_, _ = w.w.Write([]byte{b[i]})
				w.curWidth++
			default:
				w.word.WriteByte(b[i])
				w.wordLen++
				if w.curWidth+w.space.Len()+w.wordLen > w.Limit &&
					w.wordLen < w.Limit {
					w.addNewline()
				}
			}

		default:
			w.word.WriteByte(b[i])
		}

		// We manage the UTF8 state separately manually above.
		if w.pstate != parser.Utf8State {
			w.pstate = state
		}
		i++
	}

	w.addWord()

	return len(b), nil
}

// Wrap wraps a string or a block of text to a given line length, breaking word
// boundaries if necessary. This will preserve ANSI escape codes and will
// account for wide-characters in the string. The breakpoints string is a list
// of characters that are considered breakpoints for word wrapping. A hyphen
// (-) is always considered a breakpoint.
//
// Note: breakpoints must be a string of 1-cell wide rune characters.
//
// This treats the text as a sequence of graphemes.
func Wrap(s string, limit int, breakpoints string) string {
	return wrap(GraphemeWidth, s, limit, breakpoints)
}

// WrapWc wraps a string or a block of text to a given line length, breaking word
// boundaries if necessary. This will preserve ANSI escape codes and will
// account for wide-characters in the string. The breakpoints string is a list
// of characters that are considered breakpoints for word wrapping. A hyphen
// (-) is always considered a breakpoint.
//
// Note: breakpoints must be a string of 1-cell wide rune characters.
//
// This treats the text as a sequence of wide characters and runes.
func WrapWc(s string, limit int, breakpoints string) string {
	return wrap(WcWidth, s, limit, breakpoints)
}

func wrap(m Method, s string, limit int, breakpoints string) string {
	var buf bytes.Buffer
	ww := NewWrapWriter(&buf, limit)
	ww.Method = m
	if len(breakpoints) > 0 {
		ww.Breakpoints = []rune(breakpoints)
	}
	_, _ = io.WriteString(ww, s)
	return buf.String()
}

// WrapWriter is a writer that wraps text to a given line length, breaking word
// boundaries if necessary. This will preserve ANSI escape codes and will
// account for wide-characters in the string. The breakpoints string is a list
// of characters that are considered breakpoints for word wrapping. A hyphen
// (-) is always considered a breakpoint.
type WrapWriter struct {
	Limit       int
	Breakpoints []rune
	Method      Method

	w        io.Writer
	cluster  []byte
	word     bytes.Buffer
	space    bytes.Buffer
	curWidth int  // written width of the line
	wordLen  int  // word buffer len without ANSI escape codes
	pstate   byte // initial state
}

// NewWrapWriter returns a new WrapWriter that writes to w.
func NewWrapWriter(w io.Writer, limit int) *WrapWriter {
	ww := &WrapWriter{Limit: limit}
	ww.w = w
	return ww
}

func (w *WrapWriter) addSpace() {
	w.curWidth += w.space.Len()
	_, _ = w.w.Write(w.space.Bytes())
	w.space.Reset()
}

func (w *WrapWriter) addWord() {
	if w.word.Len() == 0 {
		return
	}

	w.addSpace()
	w.curWidth += w.wordLen
	_, _ = w.w.Write(w.word.Bytes())
	w.word.Reset()
	w.wordLen = 0
}

func (w *WrapWriter) addNewline() {
	_, _ = w.w.Write([]byte("\n"))
	w.curWidth = 0
	w.space.Reset()
}

// Write writes the content of p into the internal buffer.
func (w *WrapWriter) Write(p []byte) (n int, err error) {
	if w.Limit < 1 {
		return w.w.Write(p)
	}

	i := 0
	for i < len(p) {
		state, action := parser.Table.Transition(w.pstate, p[i])
		if state == parser.Utf8State {
			var width int
			w.cluster, _, width, _ = uniseg.FirstGraphemeCluster(p[i:], -1)
			if w.Method == WcWidth {
				width = runewidth.StringWidth(string(w.cluster))
			}
			i += len(w.cluster)

			r, _ := utf8.DecodeRune(w.cluster)
			switch {
			case r != utf8.RuneError && unicode.IsSpace(r) && r != nbsp: // nbsp is a non-breaking space
				w.addWord()
				w.space.WriteRune(r)
			case bytes.ContainsAny(w.cluster, string(w.Breakpoints)):
				w.addSpace()
				if w.curWidth+w.wordLen+width > w.Limit {
					w.word.Write(w.cluster)
					w.wordLen += width
				} else {
					w.addWord()
					_, _ = w.w.Write(w.cluster)
					w.curWidth += width
				}
			default:
				if w.wordLen+width > w.Limit {
					// Hardwrap the word if it's too long
					w.addWord()
				}

				w.word.Write(w.cluster)
				w.wordLen += width

				if w.curWidth+w.wordLen+w.space.Len() > w.Limit {
					w.addNewline()
				}
			}

			w.pstate = parser.GroundState
			continue
		}

		switch action {
		case parser.PrintAction, parser.ExecuteAction:
			switch r := rune(p[i]); {
			case r == '\n':
				if w.wordLen == 0 {
					if w.curWidth+w.space.Len() > w.Limit {
						w.curWidth = 0
					} else {
						// preserve whitespaces
						_, _ = w.w.Write(w.space.Bytes())
					}
					w.space.Reset()
				}

				w.addWord()
				w.addNewline()
			case unicode.IsSpace(r):
				w.addWord()
				w.space.WriteRune(r)
			case r == '-':
				fallthrough
			case runeContainsAny(r, w.Breakpoints):
				w.addSpace()
				if w.curWidth+w.wordLen >= w.Limit {
					// We can't fit the breakpoint in the current line, treat
					// it as part of the word.
					w.word.WriteRune(r)
					w.wordLen++
				} else {
					w.addWord()
					_, _ = w.w.Write([]byte(string(r)))
					w.curWidth++
				}
			default:
				if w.curWidth == w.Limit {
					w.addNewline()
				}
				w.word.WriteRune(r)
				w.wordLen++

				if w.wordLen == w.Limit {
					// Hardwrap the word if it's too long
					w.addWord()
				}

				if w.curWidth+w.wordLen+w.space.Len() > w.Limit {
					w.addNewline()
				}
			}

		default:
			w.word.WriteByte(p[i])
		}

		// We manage the UTF8 state separately manually above.
		if w.pstate != parser.Utf8State {
			w.pstate = state
		}
		i++
	}

	if w.wordLen == 0 {
		if w.curWidth+w.space.Len() > w.Limit {
			w.curWidth = 0
		} else {
			// preserve whitespaces
			_, _ = w.w.Write(w.space.Bytes())
		}
		w.space.Reset()
	}

	w.addWord()

	return len(p), nil
}

func runeContainsAny(r rune, s []rune) bool {
	for _, c := range s {
		if c == r {
			return true
		}
	}
	return false
}
