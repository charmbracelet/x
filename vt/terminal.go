package vt

import (
	"image/color"
	"io"

	uv "github.com/charmbracelet/ultraviolet"
)

// Terminal represents a virtual terminal interface.
type Terminal interface {
	BackgroundColor() color.Color
	Blur()
	Bounds() uv.Rectangle
	CellAt(x int, y int) *uv.Cell
	Close() error
	CursorColor() color.Color
	CursorPosition() uv.Position
	Draw(scr uv.Screen, area uv.Rectangle)
	Focus()
	ForegroundColor() color.Color
	Height() int
	IndexedColor(i int) color.Color
	InputPipe() io.Writer
	Paste(text string)
	Read(p []byte) (n int, err error)
	RegisterApcHandler(handler ApcHandler)
	RegisterCsiHandler(cmd int, handler CsiHandler)
	RegisterDcsHandler(cmd int, handler DcsHandler)
	RegisterEscHandler(cmd int, handler EscHandler)
	RegisterOscHandler(cmd int, handler OscHandler)
	RegisterPmHandler(handler PmHandler)
	RegisterSosHandler(handler SosHandler)
	Render() string
	Resize(width int, height int)
	SendKey(k uv.KeyEvent)
	SendKeys(keys ...uv.KeyEvent)
	SendMouse(m Mouse)
	SendText(text string)
	SetBackgroundColor(c color.Color)
	SetCallbacks(cb Callbacks)
	SetCell(x int, y int, c *uv.Cell)
	SetCursorColor(c color.Color)
	SetDefaultBackgroundColor(c color.Color)
	SetDefaultCursorColor(c color.Color)
	SetDefaultForegroundColor(c color.Color)
	SetForegroundColor(c color.Color)
	SetIndexedColor(i int, c color.Color)
	SetLogger(l Logger)
	String() string
	Touched() []*uv.LineData
	Width() int
	WidthMethod() uv.WidthMethod
	Write(p []byte) (n int, err error)
	WriteString(s string) (n int, err error)
}
