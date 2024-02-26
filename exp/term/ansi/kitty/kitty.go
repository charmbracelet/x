package kitty

import "strconv"

// Kitty keyboard protocol progressive enhancement flags.
// See: https://sw.kovidgoyal.net/kitty/keyboard-protocol/#progressive-enhancement
const (
	DisambiguateEscapeCodes = 1 << iota
	ReportEventTypes
	ReportAlternateKeys
	ReportAllKeys
	ReportAssociatedKeys

	AllFlags = DisambiguateEscapeCodes | ReportEventTypes | ReportAlternateKeys | ReportAllKeys | ReportAssociatedKeys
)

// Request is a sequence to request the terminal Kitty keyboard protocol
// enabled flags.
//
// See: https://sw.kovidgoyal.net/kitty/keyboard-protocol/
const Request = "\x1b[?u"

// Push returns a sequence to push the given flags to the terminal Kitty
// Keyboard stack.
//
//	CSI > flags u
//
// See https://sw.kovidgoyal.net/kitty/keyboard-protocol/#progressive-enhancement
func Push(flags int) string {
	var f string
	if flags > 0 {
		f = strconv.Itoa(flags)
	}

	return "\x1b" + "[" + ">" + f + "u"
}

// Pop returns a sequence to pop n number of flags from the terminal Kitty
// Keyboard stack.
//
//	CSI < flags u
//
// See https://sw.kovidgoyal.net/kitty/keyboard-protocol/#progressive-enhancement
func Pop(n int) string {
	var num string
	if n > 0 {
		num = strconv.Itoa(n)
	}

	return "\x1b" + "[" + "<" + num + "u"
}
