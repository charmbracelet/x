package wcwidth

import (
	"github.com/mattn/go-runewidth"
)

// RuneWidth returns fixed-width width of rune.
//
// Deprecated: this is now a wrapper around go-runewidth. Use go-runewidth
// directly.
func RuneWidth(r rune) int {
	return runewidth.RuneWidth(r)
}

// StringWidth returns fixed-width width of string.
//
// Deprecated: this is now a wrapper around go-runewidth. Use go-runewidth
// directly.
func StringWidth(s string) (n int) {
	return runewidth.StringWidth(s)
}
