package sequences

import (
	"fmt"
	"iter"
)

// FromString returns an iterator for escape sequences and grapheme clusters
// from the given string.
func FromString(s string) iter.Seq[string] {
	var state state[string]
	return func(yield func(string) bool) {
		for len(s) > 0 {
			n, tok, err := state.splitFunc(s, len(s) == 0)
			if err != nil {
				panic(fmt.Sprintf("error in split function: %v", err))
			}
			if !yield(tok) {
				return
			}
			s = s[n:]
		}
	}
}

// FromBytes returns an iterator for escape sequences and grapheme clusters
// from the given byte slice.
func FromBytes(b []byte) iter.Seq[[]byte] {
	var state state[[]byte]
	return func(yield func([]byte) bool) {
		for len(b) > 0 {
			n, tok, err := state.splitFunc(b, len(b) == 0)
			if err != nil {
				panic(fmt.Sprintf("error in split function: %v", err))
			}
			if !yield(tok) {
				return
			}
			b = b[n:]
		}
	}
}
