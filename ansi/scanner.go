package ansi

import (
	"errors"
	"unicode"
	"unicode/utf8"

	"github.com/charmbracelet/x/ansi/parser"
	"github.com/rivo/uniseg"
)

// Scanner implements the reading of strings with ANSI escape and control codes and
// accounts for wide-characters (such as East Asians and emojis). Used to split the
// codes from the string without getting into the details of the codes.
//
// The default splitter ScanAll]  will split the string into separate control codes
// and regular strings stripped of encoding.
type Scanner struct {
	b      []byte
	width  int
	end    int
	err    error
	start  int
	pState byte
	escape bool
	split  SplitFunc
	token  []byte
}

// Scanner Errors
var (
	ErrAdvanceTooFar   = errors.New("ansi.Scanner: split function returned advance larger than buffer")
	ErrAdvanceNegative = errors.New("ansi.Scanner: split function returned negative advance")
)

// SplitFunc is the signature of the split function used to further tokenize
// the input. The arguments are the current substring of the remaining unprocessed
// data, the current width of the text and the current token. The return values
// are the number of bytes to advance the input and the next token to return to
// the user.
//
// The split function is called repeatedly as the current token is read. If the
// advance return value is 0 or less the next rune is read and the split is called
// again.
type SplitFunc func(b []byte, width int, eof bool) (advance int, token []byte, err error)

// NewScanner creates a new Scanner for reading the string.
func NewScanner(s string, splitters ...SplitFunc) *Scanner {
	scanner := &Scanner{
		b:      []byte(s),
		pState: parser.GroundState,
		split:  composeSplitters(splitters),
	}
	return scanner
}

func composeSplitters(splitters []SplitFunc) SplitFunc {
	switch len(splitters) {
	case 0:
		return ScanAll
	case 1:
		return splitters[0]
	}
	return func(b []byte, width int, eof bool) (int, []byte, error) {
		for _, split := range splitters {
			w, token, err := split(b, width, eof)
			if w > 0 {
				return w, token, err
			}
		}
		return 0, []byte(nil), nil
	}
}

// Split sets the split function for the [Scanner].
// The default split function is [ScanAll].
func (s *Scanner) Split(f SplitFunc) {
	s.split = f
}

// Text returns the string for current token.
func (s *Scanner) Text() string {
	return string(s.Bytes())
}

// Error returns the current error.
func (s *Scanner) Error() error {
	return s.err
}

// Bytes returns the current token.
func (s *Scanner) Bytes() []byte {
	return s.b[s.start:s.end]
}

// Token returns the current token, width and escape flag.
func (s *Scanner) Token() ([]byte, int, bool) {
	return s.b[s.start:s.end], s.width, s.escape
}

// Len returns the length for current token.
func (s *Scanner) Len() int {
	return s.end - s.start
}

// Width returns the width for current token.
func (s *Scanner) Width() int {
	return s.width
}

// IsEscape returns if token is an escape sequence.
func (s *Scanner) IsEscape() bool {
	return s.escape
}

// EOF returns true if at the end of the input string
func (s *Scanner) EOF() bool {
	return s.end >= len(s.b)
}

func (s *Scanner) advance(size, width int) bool {
	s.end += size
	s.width += width
	n, tk, err := s.split(s.b[s.start:s.end], s.width, s.EOF())
	s.token = tk
	switch {
	case err != nil:
		s.token = nil
		s.err = err
		return false
	case n < 0:
		s.token = nil
		s.err = ErrAdvanceNegative
		return false
	case n > s.Len():
		s.token = nil
		s.err = ErrAdvanceTooFar
		return false
	case n == 0:
		return false
	case n == s.Len():
		return true
	case n == s.Len()-size:
		// can backup if completed without accepting the last rune
		s.end -= size
		s.width -= width
		return true
	default:
		// not using the whole buffer, update the end
		// and re-scan the string for the new width
		s.end = s.start + n
		s.width = stringWidth(s.Text())
		return true
	}
}

// Scan reads the next token from source and returns it.
func (s *Scanner) Scan() bool {
	if s.err != nil {
		return false
	}
	if s.EOF() {
		return false
	}
	s.start = s.end
	s.width = 0
	if s.end >= len(s.b) {
		return false
	}

	// Here we iterate over the bytes of the string and collect characters
	// and runes.
	// On change of token we emit the current token.
	for s.end < len(s.b) {
		state, action := parser.Table.Transition(s.pState, s.b[s.end])
		if state == parser.Utf8State {
			if s.escape {
				// emit on a change from escape sequence
				// if there is a buffer
				if s.Len() > 0 {
					return true
				}
				s.escape = false
			}
			// This action happens when we transition to the Utf8State.
			cluster, _, width, _ := uniseg.FirstGraphemeCluster(s.b[s.end:], -1)
			if s.advance(len(cluster), width) {
				return true
			}
			// Done collecting, now we're back in the ground state.
			s.pState = parser.GroundState
			continue
		}

		switch action {
		case parser.PrintAction, parser.ExecuteAction:
			if s.escape {
				// emit on a change from escape sequence
				// if there is a buffer
				if s.Len() > 0 {
					return true
				}
				s.escape = false
			}
			if s.advance(1, 1) {
				return true
			}
		default:
			if !s.escape {
				// emit on a change to escape sequence
				// if there is a buffer
				if s.Len() > 0 {
					return true
				}
				s.escape = true
			}
			s.end++
		}
		// Transition to the next state.
		s.pState = state
	}

	return true
}

// Splitter Functions

// ScanAll is a split function for a [Scanner] that returns all data as Text.
func ScanAll(b []byte, width int, end bool) (int, []byte, error) {
	return 0, []byte(nil), nil
}

var _ SplitFunc = ScanAll

// ScanRunes is a split function for a [Scanner] that returns each rune.
func ScanRunes(b []byte, width int, end bool) (int, []byte, error) {
	return len(b), b, nil
}

var _ SplitFunc = ScanRunes

// ScanWords is a split function for a [Scanner] that returns each space
// separated word, and spaces as tokens.
func ScanWords(b []byte, width int, end bool) (int, []byte, error) {
	if len(b) == 1 {
		return 0, []byte(nil), nil
	}
	first, _ := utf8.DecodeRune(b)
	last, lastWidth := utf8.DecodeLastRune(b)
	if unicode.IsSpace(first) != unicode.IsSpace(last) {
		return len(b) - lastWidth, b[:len(b)-lastWidth], nil
	}
	return 0, []byte(nil), nil
}

var _ SplitFunc = ScanWords

// ScanLines is a split function for a [Scanner] that returns lines and
// and newlines as tokens.
func ScanLines(b []byte, width int, end bool) (int, []byte, error) {
	first, _ := utf8.DecodeRune(b)
	if len(b) == 1 {
		if first == '\n' {
			return len(b), b, nil
		}
		return 0, []byte(nil), nil
	}
	last, lastWidth := utf8.DecodeLastRune(b)
	if first == '\r' {
		switch last {
		case '\r':
			return 0, []byte(nil), nil
		case '\n':
			return len(b), b, nil
		}
		n := len(b) - lastWidth
		return n, b[:n], nil
	}
	switch last {
	case '\r', '\n':
		n := len(b) - lastWidth
		return n, b[:n], nil
	}
	return 0, []byte(nil), nil
}

var _ SplitFunc = ScanLines

// utility functions

// stringWidth returns the width of a string in cells. The argument is a string
// without ANSI escape sequences. The return value is the number of cells that
// the string will occupy when printed in a terminal. Wide characters (such as
// East Asians and emojis) are accounted for.
//
// ANSI escape not accounted for and not expected to be present in the input.
func stringWidth(s string) int {
	width := 0
	g := uniseg.NewGraphemes(s)
	for g.Next() {
		width += g.Width()
	}
	return width
}
