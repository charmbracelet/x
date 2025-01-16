package input

import (
	"fmt"
	"strings"
)

// Event represents a terminal event.
type Event interface{}

// UnknownEvent represents an unknown event.
type UnknownEvent string

// String returns a string representation of the unknown event.
func (e UnknownEvent) String() string {
	return fmt.Sprintf("%q", string(e))
}

// MultiEvent represents multiple messages event.
type MultiEvent []Event

// String returns a string representation of the multiple messages event.
func (e MultiEvent) String() string {
	var sb strings.Builder
	for _, ev := range e {
		sb.WriteString(fmt.Sprintf("%v\n", ev))
	}
	return sb.String()
}

// WindowSizeEvent is used to report the terminal size. Note that Windows does
// not have support for reporting resizes via SIGWINCH signals and relies on
// the Windows Console API to report window size changes.
type WindowSizeEvent struct {
	Width  int
	Height int
}

// WindowAreaEvent is used to report the terminal area size in pixels. This is
// the response from a [ansi.WindowOp] [ansi.ReportWindowSizeWinOp] request.
// You can use this along with [WindowSizeEvent] to get the terminal cell size.
//
// Example:
//
//	    // Assuming we already have a WindowSizeEvent
//		var winsize WindowSizeEvent
//		switch ev := ev.(type) {
//		case WindowAreaEvent:
//		  cellWidth := ev.Width / winsize.Width
//		  cellHeight := ev.Height / winsize.Height
//		}
type WindowAreaEvent struct {
	Width  int
	Height int
}
