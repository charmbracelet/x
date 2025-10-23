package vt

import (
	"image/color"
	"sync"

	uv "github.com/charmbracelet/ultraviolet"
)

// SafeEmulator is a wrapper around an Emulator that adds concurrency safety.
type SafeEmulator struct {
	*Emulator
	mu sync.RWMutex
}

// NewSafeEmulator creates a new SafeEmulator instance.
func NewSafeEmulator(w, h int) *SafeEmulator {
	return &SafeEmulator{
		Emulator: NewEmulator(w, h),
	}
}

// Write writes data to the emulator in a concurrency-safe manner.
func (se *SafeEmulator) Write(data []byte) (int, error) {
	se.mu.Lock()
	defer se.mu.Unlock()
	return se.Emulator.Write(data)
}

// Read reads data from the emulator in a concurrency-safe manner.
func (se *SafeEmulator) Read(p []byte) (int, error) {
	return se.Emulator.Read(p)
}

// Resize resizes the emulator in a concurrency-safe manner.
func (se *SafeEmulator) Resize(w, h int) {
	se.mu.Lock()
	defer se.mu.Unlock()
	se.Emulator.Resize(w, h)
}

// Render renders the emulator's current state in a concurrency-safe manner.
func (se *SafeEmulator) Render() string {
	se.mu.RLock()
	defer se.mu.RUnlock()
	return se.Emulator.Render()
}

// SetCell sets a cell in the emulator in a concurrency-safe manner.
func (se *SafeEmulator) SetCell(x, y int, cell *uv.Cell) {
	se.mu.Lock()
	defer se.mu.Unlock()
	se.Emulator.SetCell(x, y, cell)
}

// CellAt retrieves a cell from the emulator in a concurrency-safe manner.
func (se *SafeEmulator) CellAt(x, y int) *uv.Cell {
	se.mu.RLock()
	defer se.mu.RUnlock()
	return se.Emulator.CellAt(x, y)
}

// SendKey sends a key event to the emulator in a concurrency-safe manner.
func (se *SafeEmulator) SendKey(key uv.KeyEvent) {
	se.mu.Lock()
	defer se.mu.Unlock()
	se.Emulator.SendKey(key)
}

// SendMouse sends a mouse event to the emulator in a concurrency-safe manner.
func (se *SafeEmulator) SendMouse(mouse uv.MouseEvent) {
	se.mu.Lock()
	defer se.mu.Unlock()
	se.Emulator.SendMouse(mouse)
}

// SendText sends text input to the emulator in a concurrency-safe manner.
func (se *SafeEmulator) SendText(text string) {
	se.mu.Lock()
	defer se.mu.Unlock()
	se.Emulator.SendText(text)
}

// Paste pastes text into the emulator in a concurrency-safe manner.
func (se *SafeEmulator) Paste(text string) {
	se.mu.Lock()
	defer se.mu.Unlock()
	se.Emulator.Paste(text)
}

// SetForegroundColor sets the foreground color in a concurrency-safe manner.
func (se *SafeEmulator) SetForegroundColor(color color.Color) {
	se.mu.Lock()
	defer se.mu.Unlock()
	se.Emulator.SetForegroundColor(color)
}

// SetBackgroundColor sets the background color in a concurrency-safe manner.
func (se *SafeEmulator) SetBackgroundColor(color color.Color) {
	se.mu.Lock()
	defer se.mu.Unlock()
	se.Emulator.SetBackgroundColor(color)
}

// SetCursorColor sets the cursor color in a concurrency-safe manner.
func (se *SafeEmulator) SetCursorColor(color color.Color) {
	se.mu.Lock()
	defer se.mu.Unlock()
	se.Emulator.SetCursorColor(color)
}

// SetIndexedColor sets an indexed color in a concurrency-safe manner.
func (se *SafeEmulator) SetIndexedColor(index int, color color.Color) {
	se.mu.Lock()
	defer se.mu.Unlock()
	se.Emulator.SetIndexedColor(index, color)
}

// IndexedColor retrieves an indexed color in a concurrency-safe manner.
func (se *SafeEmulator) IndexedColor(index int) color.Color {
	se.mu.RLock()
	defer se.mu.RUnlock()
	return se.Emulator.IndexedColor(index)
}

// Touched returns the touched lines in a concurrency-safe manner.
func (se *SafeEmulator) Touched() []*uv.LineData {
	se.mu.RLock()
	defer se.mu.RUnlock()
	return se.Emulator.Touched()
}

// Height returns the height of the emulator in a concurrency-safe manner.
func (se *SafeEmulator) Height() int {
	se.mu.RLock()
	defer se.mu.RUnlock()
	return se.Emulator.Height()
}

// Width returns the width of the emulator in a concurrency-safe manner.
func (se *SafeEmulator) Width() int {
	se.mu.RLock()
	defer se.mu.RUnlock()
	return se.Emulator.Width()
}

// ForegroundColor returns the foreground color in a concurrency-safe manner.
func (se *SafeEmulator) ForegroundColor() color.Color {
	se.mu.RLock()
	defer se.mu.RUnlock()
	return se.Emulator.ForegroundColor()
}

// BackgroundColor returns the background color in a concurrency-safe manner.
func (se *SafeEmulator) BackgroundColor() color.Color {
	se.mu.RLock()
	defer se.mu.RUnlock()
	return se.Emulator.BackgroundColor()
}

// CursorColor returns the cursor color in a concurrency-safe manner.
func (se *SafeEmulator) CursorColor() color.Color {
	se.mu.RLock()
	defer se.mu.RUnlock()
	return se.Emulator.CursorColor()
}

// CursorPosition returns the cursor position in a concurrency-safe manner.
func (se *SafeEmulator) CursorPosition() uv.Position {
	se.mu.RLock()
	defer se.mu.RUnlock()
	return se.Emulator.CursorPosition()
}

// Draw draws the emulator's content onto a given surface in a concurrency-safe manner.
func (se *SafeEmulator) Draw(s uv.Screen, a uv.Rectangle) {
	se.mu.RLock()
	defer se.mu.RUnlock()
	se.Emulator.Draw(s, a)
}
