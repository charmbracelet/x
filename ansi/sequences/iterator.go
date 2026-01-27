package sequences

import (
	"github.com/charmbracelet/x/ansi/internal/iterators"
	"github.com/clipperhouse/stringish"
)

// Iterator is an iterator over ANSI escape sequences and grapheme clusters.
//
// Iterate using the [Iterator.Next] method, and get the width of the current
// token using the [Iterator.Width] method.
type Iterator[T stringish.Interface] struct {
	*iterators.Iterator[T]
}

// FromString returns an iterator for escape sequences and grapheme clusters
// from the given string.
func FromString(s string) Iterator[string] {
	return Iterator[string]{
		iterators.New(SplitFunc[string], s),
	}
}

// FromBytes returns an iterator for escape sequences and grapheme clusters
// from the given byte slice.
func FromBytes(b []byte) Iterator[[]byte] {
	return Iterator[[]byte]{
		iterators.New(SplitFunc[[]byte], b),
	}
}
