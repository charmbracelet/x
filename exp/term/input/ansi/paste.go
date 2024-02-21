package ansi

import (
	"fmt"
	"unicode/utf8"

	"github.com/charmbracelet/x/exp/term/input"
)

// PasteEvent represents a bracketed paste event.
type PasteEvent string

var _ input.Event = PasteEvent("")

// String implements Event.
func (e PasteEvent) String() string {
	return fmt.Sprintf("paste: %q", string(e))
}

// Type implements Event.
func (PasteEvent) Type() string {
	return "Paste"
}

func parseBracketedPaste(p []byte, buf *[]byte) input.Event {
	switch string(p) {
	case "\x1b[200~":
		*buf = []byte{}
	case "\x1b[201~":
		var paste []rune
		for len(*buf) > 0 {
			r, w := utf8.DecodeRune(*buf)
			if r != utf8.RuneError {
				*buf = (*buf)[w:]
			}
			paste = append(paste, r)
		}
		*buf = nil
		return PasteEvent(paste)
	}
	return nil
}
