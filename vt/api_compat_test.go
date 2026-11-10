package vt_test

import (
	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/vt"
)

// These compile-time fixtures pin the exported Screen and Scrollback method
// signatures that existed before semantic reflow was introduced.
type preReflowScrollback interface {
	Push(uv.Line)
	PushN(*uv.RenderBuffer, int, int)
	Len() int
	MaxLines() int
	SetMaxLines(int)
	Line(int) uv.Line
	Lines() []uv.Line
	Clear()
	CellAt(int, int) *uv.Cell
}

type preReflowScreen interface {
	Reset()
	Bounds() uv.Rectangle
	Touched() []*uv.LineData
	ClearTouched()
	CellAt(int, int) *uv.Cell
	SetCell(int, int, *uv.Cell)
	Height() int
	Resize(int, int)
	Width() int
	Clear()
	ClearWithScrollback()
	ClearArea(uv.Rectangle)
	Fill(*uv.Cell)
	FillArea(*uv.Cell, uv.Rectangle)
	Cursor() vt.Cursor
	CursorPosition() (int, int)
	ScrollRegion() uv.Rectangle
	SaveCursor()
	RestoreCursor()
	ShowCursor()
	HideCursor()
	InsertCell(int)
	DeleteCell(int)
	ScrollUp(int)
	ScrollDown(int)
	InsertLine(int) bool
	DeleteLine(int) bool
	Scrollback() *vt.Scrollback
	SetScrollback(*vt.Scrollback)
	SetScrollbackSize(int)
}

var (
	_ preReflowScrollback = (*vt.Scrollback)(nil)
	_ preReflowScreen     = (*vt.Screen)(nil)
)
