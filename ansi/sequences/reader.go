package sequences

import (
	"bufio"
	"io"
)

// Scanner is a [bufio.Scanner] wrapped to use ANSI escape sequence splitting.
type Scanner struct {
	*bufio.Scanner
	state splitState[[]byte]
}

// FromReader returns a new [Scanner] that splits ANSI escape sequences and
// grapheme clusters.
//
// It embeds a [bufio.Scanner], so you can use [Scanner.Scan] and
// [Scanner.Text] as usual.
func FromReader(r io.Reader) *Scanner {
	s := new(Scanner)
	s.Scanner = bufio.NewScanner(r)
	s.Scanner.Split(s.state.splitFunc)
	return s
}

// Width returns the display width of the most recently scanned token.
func (s *Scanner) Width() int {
	return s.state.width
}
