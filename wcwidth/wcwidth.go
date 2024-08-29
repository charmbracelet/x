// wcwidth is a Go implementation of wcwidth(3) that uses
// golang.org/x/text/width. It is a drop-in replacement for
// github.com/mattn/go-runewidth.
//
// Unlike go-runewidth, wcwidth treats East Asian ambiguous characters as
// single-width characters. This is consistent with the behavior of wcwidth(3).

package wcwidth

import (
	"unicode"

	"golang.org/x/text/width"
)

// IsComb returns true if r is a Unicode combining character. Alias of:
//
//	unicode.Is(unicode.Mn, r)
func IsComb(r rune) bool { return unicode.Is(unicode.Mn, r) }

// RuneWidth returns fixed-width width of rune.
// https://en.wikipedia.org/wiki/Halfwidth_and_fullwidth_forms#In_Unicode
func RuneWidth(r rune) int {
	if r == 0 || !unicode.IsPrint(r) || IsComb(r) {
		return 0
	}
	k := width.LookupRune(r)
	switch k.Kind() {
	case width.EastAsianWide, width.EastAsianFullwidth:
		return 2
	case width.EastAsianNarrow, width.EastAsianHalfwidth, width.EastAsianAmbiguous, width.Neutral:
		return 1
	default:
		return 0
	}
}

// StringWidth returns fixed-width width of string.
// https://en.wikipedia.org/wiki/Halfwidth_and_fullwidth_forms#In_Unicode
func StringWidth(s string) (n int) {
	for _, r := range s {
		n += RuneWidth(r)
	}
	return n
}
