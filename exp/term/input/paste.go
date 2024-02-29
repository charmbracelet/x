package input

import "strings"

// PasteEvent is an event that is emitted when a terminal receives pasted text
// using bracketed-paste.
type PasteEvent string

// String implements fmt.Stringer.
func (p PasteEvent) String() string {
	s := string(p)
	s = strings.ReplaceAll(s, "\n", "\r\n")
	return s
}

// PasteStartEvent is an event that is emitted when a terminal enters
// bracketed-paste mode.
type PasteStartEvent struct{}

// PasteEvent is an event that is emitted when a terminal receives pasted text.
type PasteEndEvent struct{}
