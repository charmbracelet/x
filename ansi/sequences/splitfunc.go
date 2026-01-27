package sequences

import (
	"github.com/charmbracelet/x/ansi"
	"github.com/clipperhouse/stringish"
)

type splitState[T stringish.Interface] struct {
	state byte
	width int // width of the last token
}

// splitFunc is a [bufio.SplitFunc]-compatible function that splits ANSI escape
// sequences and grapheme clusters.
func (s *splitState[T]) splitFunc(data T, atEOF bool) (n int, token T, err error) {
	var empty T
	if atEOF {
		return 0, empty, nil
	}
	token, s.width, n, s.state = ansi.DecodeSequence(data, s.state, nil)
	return n, token, nil
}

// SplitFunc is a [bufio.SplitFunc]-compatible function that splits ANSI escape
// sequences and grapheme clusters for the given stringish type.
func SplitFunc[T stringish.Interface](data T, atEOF bool) (n int, token T, err error) {
	var state splitState[T]
	return state.splitFunc(data, atEOF)
}
