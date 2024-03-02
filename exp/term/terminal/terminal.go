package terminal

import (
	"errors"
	"image/color"
	"io"
	"os"
)

var (
	// ErrNotTerminal is returned when the console is not a terminal.
	ErrNotTerminal = errors.New("not a terminal")

	// ErrUnsupported is returned when the platform does not support the operation.
	ErrUnsupported = errors.New("unsupported platform")
)

// Terminal represents a terminal console interface.
type Terminal interface {
	io.ReadWriter

	IsTerminal() bool
	MakeRaw() error
	Restore() error
	GetSize() (width int, height int, err error)

	SupportsKittyKeyboard() bool
	BackgroundColor() color.Color
	ForegroundColor() color.Color
	CursorColor() color.Color
}

// NewTerminal returns a new Console interface.
func NewTerminal(in io.Reader, out io.Writer) Terminal {
	return newTerminal(in, out)
}

// Current returns the current console.
func Current() Terminal {
	return NewTerminal(os.Stdin, os.Stdout)
}
