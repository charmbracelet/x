package ansi

import (
	"unicode"
	"unicode/utf8"

	"github.com/charmbracelet/x/ansi/parser"
	"github.com/rivo/uniseg"
)

// ScannerToken is the scanner type identifier
type ScannerToken int

const (
	EmptyToken ScannerToken = -iota
	EndToken
	ErrorToken
	ControlToken
	RuneToken
	TextToken
	SpaceToken
	LineToken
)

// Scanner implements the reading of strings with ANSI escape and control codes and
// accounts for wide-characters (such as East Asians and emojis). Used to split the
// codes from the string without getting into the details of the codes.
//
// The default splitter ScanAll]  will split the string into separate control codes
// and regular strings stripped of encoding.
type Scanner struct {
	b     []byte
	width int
	end   int
	start int
	state byte
	split ScannerSplit
	token ScannerToken
}

// ScannerSplit is the signature of the split function used to further tokenize
// the input. The arguments are the current substring of the remaining unprocessed
// data, the current width of the text and the current token. The return values
// are the number of bytes to advance the input and the next token to return to
// the user.
//
// The split function is called repeatedly as the current token is read. If the
// advance return value is 0 or less the next rune is read and the split is called
// again.
type ScannerSplit func(b []byte, width int, token ScannerToken) (advance int, newToken ScannerToken)

// NewScanner creates a new Scanner for reading the string.
func NewScanner(s string, splitters ...ScannerSplit) *Scanner {
	scanner := &Scanner{
		b:     []byte(s),
		state: parser.GroundState,
		split: composeSplitters(splitters),
	}
	return scanner
}

func composeSplitters(splitters []ScannerSplit) ScannerSplit {
	switch len(splitters) {
	case 0:
		return ScanAll
	case 1:
		return splitters[0]
	}
	return func(b []byte, width int, token ScannerToken) (advance int, newToken ScannerToken) {
		var w int
		for _, split := range splitters {
			w, token = split(b, width, token)
			if w > 0 {
				return w, token
			}
		}
		return 0, token
	}
}

// Split sets the split function for the [Scanner].
// The default split function is [ScanAll].
func (s *Scanner) Split(f ScannerSplit) {
	s.split = f
}

// Text returns the string for current token.
func (s *Scanner) Text() string {
	return string(s.data())
}

func (s *Scanner) data() []byte {
	return s.b[s.start:s.end]
}

// Len returns the length for current token.
func (s *Scanner) Len() int {
	return s.end - s.start
}

// Width returns the width for current token.
func (s *Scanner) Width() int {
	return s.width
}

// EOF returns true if at the end of the input string
func (s *Scanner) EOF() bool {
	return s.end >= len(s.b)
}

func (s *Scanner) advance(size, width int) bool {
	s.end += size
	s.width += width
	n, tk := s.split(s.b[s.start:s.end], s.width, s.token)
	s.token = tk
	if n > s.Len() {
		s.token = ErrorToken
		return false
	}
	switch n {
	case 0:
		return false
	case s.Len():
		return true
	case s.Len() - size:
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
func (s *Scanner) Scan() (ScannerToken, string) {
	if s.token == ErrorToken {
		return ErrorToken, ""
	}
	s.token = EmptyToken
	s.start = s.end
	s.width = 0
	if s.end >= len(s.b) {
		return EndToken, ""
	}

	// Here we iterate over the bytes of the string and collect characters
	// and runes.
	// On change of token we emit the current token.
	for s.end < len(s.b) {
		state, action := parser.Table.Transition(s.state, s.b[s.end])

		if state == parser.Utf8State {
			switch s.token {
			case EmptyToken:
				s.token = TextToken
			case ControlToken:
				// emit on a change from control type
				if s.Len() > 0 {
					return s.token, s.Text()
				}
				s.token = TextToken
			}
			// This action happens when we transition to the Utf8State.
			cluster, _, width, _ := uniseg.FirstGraphemeCluster(s.b[s.end:], -1)
			if s.advance(len(cluster), width) {
				return s.token, s.Text()
			}
			// Done collecting, now we're back in the ground state.
			s.state = parser.GroundState
			continue
		}

		switch action {
		case parser.PrintAction, parser.ExecuteAction:
			switch s.token {
			case EmptyToken:
				s.token = TextToken
			case ControlToken:
				// emit on a change from control type
				if s.Len() > 0 {
					return s.token, s.Text()
				}
				s.token = TextToken
			}
			if s.advance(1, 1) {
				return s.token, s.Text()
			}

		default:
			if s.token != ControlToken && s.Len() > 0 {
				return s.token, s.Text()
			}
			s.token = ControlToken
			s.end++
		}
		// Transition to the next state.
		s.state = state
	}

	return s.token, s.Text()
}

// Splitter Functions

// ScanAll is a split function for a [Scanner] that returns all data as Text.
func ScanAll(b []byte, width int, token ScannerToken) (int, ScannerToken) {
	return 0, TextToken
}

// ScanWords is a split function for a [Scanner] that returns each space
// separated word, and spaces as tokens.
func ScanWords(b []byte, width int, token ScannerToken) (int, ScannerToken) {
	r0, _ := utf8.DecodeRune(b)
	if len(b) == 1 {
		if unicode.IsSpace(r0) {
			return 0, SpaceToken
		}
		return 0, TextToken
	}
	r1, r1w := utf8.DecodeLastRune(b)
	if unicode.IsSpace(r0) != unicode.IsSpace(r1) {
		if unicode.IsSpace(r0) {
			return len(b) - r1w, SpaceToken
		}
		return len(b) - r1w, TextToken
	}
	return 0, TextToken
}

// ScanRunes is a split function for a [Scanner] that returns each rune.
func ScanRunes(b []byte, width int, token ScannerToken) (int, ScannerToken) {
	return len(b), RuneToken
}

// ScanWords is a split function for a [Scanner] that returns lines and
// and newlines as tokens.
func ScanLines(b []byte, width int, token ScannerToken) (int, ScannerToken) {
	r0, r0w := utf8.DecodeRune(b)
	if r0 == '\n' {
		return r0w, LineToken
	}
	if len(b) == 1 {
		if r0 == '\r' {
			return 0, LineToken
		}
		return 0, token
	}
	r1, r1w := utf8.DecodeLastRune(b)
	if r0 == '\r' {
		switch r1 {
		case '\r':
			return 0, LineToken
		case '\n':
			return len(b), LineToken
		default:
			return len(b) - r1w, LineToken
		}
	}
	switch r1 {
	case '\r', '\n':
		return len(b) - r1w, token
	default:
		return 0, token
	}
}

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
