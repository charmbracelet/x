package vt

import "github.com/charmbracelet/x/cellbuf"

// Callbacks represents a set of callbacks for a terminal.
type Callbacks struct {
	// Bell callback. When set, this function is called when a bell character is
	// received.
	Bell func()

	// Damage callback. When set, this function is called when a cell is damaged
	// or changed.
	Damage func(Damage)

	// Title callback. When set, this function is called when the terminal title
	// changes.
	Title func(string)

	// IconName callback. When set, this function is called when the terminal
	// icon name changes.
	IconName func(string)

	// AltScreen callback. When set, this function is called when the alternate
	// screen is activated or deactivated.
	AltScreen func(bool)

	// CursorPosition callback. When set, this function is called when the cursor
	// position changes.
	CursorPosition func(old, new cellbuf.Position) //nolint:predeclared

	// CursorVisibility callback. When set, this function is called when the
	// cursor visibility changes.
	CursorVisibility func(visible bool)

	// CursorStyle callback. When set, this function is called when the cursor
	// style changes.
	CursorStyle func(style CursorStyle, blink bool)
}
