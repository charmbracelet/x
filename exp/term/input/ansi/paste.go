package ansi

import (
	"fmt"

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
