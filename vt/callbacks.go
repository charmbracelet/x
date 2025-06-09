package vt

import (
	"image/color"

	"github.com/charmbracelet/uv"
)

// Callbacks represents a set of callbacks for a terminal.
type Callbacks struct {
	// Bell callback. When set, this function is called when a bell character is
	// received.
	Bell func()

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
	CursorPosition func(old, new uv.Position) //nolint:predeclared,revive

	// CursorVisibility callback. When set, this function is called when the
	// cursor visibility changes.
	CursorVisibility func(visible bool)

	// CursorStyle callback. When set, this function is called when the cursor
	// style changes.
	CursorStyle func(style CursorStyle, blink bool)

	// CursorColor callback. When set, this function is called when the cursor
	// color changes. Nil indicates the default terminal color.
	CursorColor func(color color.Color)

	// BackgroundColor callback. When set, this function is called when the
	// background color changes. Nil indicates the default terminal color.
	BackgroundColor func(color color.Color)

	// ForegroundColor callback. When set, this function is called when the
	// foreground color changes. Nil indicates the default terminal color.
	ForegroundColor func(color color.Color)
}
