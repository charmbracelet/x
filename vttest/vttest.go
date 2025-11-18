// Package vttest provides a virtual terminal implementation for testing
// terminal applications. It allows you to create a terminal instance with a
// pseudo-terminal (PTY) and capture its state at any moment, enabling you to
// write tests that verify the behavior of terminal applications.
package vttest

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"io"
	"maps"
	"os"
	"os/exec"
	"sync"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/vt"
	"github.com/charmbracelet/x/xpty"
)

// Terminal represents a virtual terminal with it's PTY and state.
type Terminal struct {
	*vt.Emulator

	cols, rows  int
	title       string
	altScreen   bool
	ansiModes   map[ansi.ANSIMode]ansi.ModeSetting
	decModes    map[ansi.DECMode]ansi.ModeSetting
	cursorPos   image.Point
	cursorVis   bool
	cursorColor color.Color
	cursorStyle vt.CursorStyle
	cursorBlink bool
	bgColor     color.Color
	fgColor     color.Color

	pty    xpty.Pty
	ptyIn  io.Reader
	ptyOut io.Writer

	mu sync.Mutex
}

// NewTerminal creates a new virtual terminal with the given size for testing
// purposes. At any moment, you can take a snapshot of the terminal state by
// calling the [Terminal.Snapshot] method on the returned Terminal instance.
func NewTerminal(cols, rows int) (*Terminal, error) {
	pty, err := xpty.NewPty(cols, rows)
	if err != nil {
		return nil, fmt.Errorf("failed to create pty: %w", err)
	}

	term := new(Terminal)
	term.cols = cols
	term.rows = rows
	term.ansiModes = make(map[ansi.ANSIMode]ansi.ModeSetting)
	term.decModes = make(map[ansi.DECMode]ansi.ModeSetting)

	switch p := pty.(type) {
	case *xpty.UnixPty:
		term.ptyIn = p.Slave()
		term.ptyOut = p.Slave()
	case *xpty.ConPty:
		inFile := os.NewFile(p.InPipeReadFd(), "|0")
		outFile := os.NewFile(p.OutPipeWriteFd(), "|1")
		term.ptyIn = inFile
		term.ptyOut = outFile
	}

	vterm := vt.NewEmulator(cols, rows)
	vterm.SetCallbacks(vt.Callbacks{
		Title: func(title string) {
			term.mu.Lock()
			defer term.mu.Unlock()
			term.title = title
		},
		AltScreen: func(alt bool) {
			term.mu.Lock()
			defer term.mu.Unlock()
			term.altScreen = alt
		},
		EnableMode: func(mode ansi.Mode) {
			term.mu.Lock()
			defer term.mu.Unlock()
			switch m := mode.(type) {
			case ansi.ANSIMode:
				term.ansiModes[m] = ansi.ModeSet
			case ansi.DECMode:
				term.decModes[m] = ansi.ModeSet
			}
		},
		DisableMode: func(mode ansi.Mode) {
			term.mu.Lock()
			defer term.mu.Unlock()
			switch m := mode.(type) {
			case ansi.ANSIMode:
				term.ansiModes[m] = ansi.ModeReset
			case ansi.DECMode:
				term.decModes[m] = ansi.ModeReset
			}
		},
		CursorPosition: func(_, newpos uv.Position) {
			term.mu.Lock()
			defer term.mu.Unlock()
			term.cursorPos = newpos
		},
		CursorVisibility: func(visible bool) {
			term.mu.Lock()
			defer term.mu.Unlock()
			term.cursorVis = visible
		},
		CursorStyle: func(style vt.CursorStyle, blink bool) {
			term.mu.Lock()
			defer term.mu.Unlock()
			term.cursorStyle = style
			term.cursorBlink = blink
		},
		CursorColor: func(color color.Color) {
			term.mu.Lock()
			defer term.mu.Unlock()
			term.cursorColor = color
		},
		BackgroundColor: func(color color.Color) {
			term.mu.Lock()
			defer term.mu.Unlock()
			term.bgColor = color
		},
		ForegroundColor: func(color color.Color) {
			term.mu.Lock()
			defer term.mu.Unlock()
			term.fgColor = color
		},
	})

	term.Emulator = vterm
	term.pty = pty

	// Copy PTY input to terminal
	go io.Copy(vterm, pty) //nolint:errcheck
	// Copy terminal output to PTY
	go io.Copy(pty, vterm) //nolint:errcheck

	return term, nil
}

// Start starts a process attached to the terminal's PTY.
func (t *Terminal) Start(cmd *exec.Cmd) error {
	if err := t.pty.Start(cmd); err != nil {
		return fmt.Errorf("failed to start process: %w", err)
	}
	return nil
}

// Close closes the terminal and its PTY.
func (t *Terminal) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if err := t.Emulator.Close(); err != nil && !errors.Is(err, io.EOF) {
		_ = t.pty.Close()
		return fmt.Errorf("failed to close emulator: %w", err)
	}

	if err := t.pty.Close(); err != nil {
		return fmt.Errorf("failed to close pty: %w", err)
	}
	return nil
}

// Resize resizes the terminal and its PTY.
func (t *Terminal) Resize(cols, rows int) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.cols = cols
	t.rows = rows
	t.Emulator.Resize(cols, rows)
	if err := t.pty.Resize(cols, rows); err != nil {
		return fmt.Errorf("failed to resize pty: %w", err)
	}

	return nil
}

// Input returns the input side of the terminal's PTY.
func (t *Terminal) Input() io.Reader {
	return t.ptyIn
}

// Output returns the output side of the terminal's PTY.
func (t *Terminal) Output() io.Writer {
	return t.ptyOut
}

// Snapshot takes a snapshot of the current terminal state.
// The returned [Snapshot] can be used to inspect the terminal state at the
// moment the snapshot was taken. It can also be serialized to JSON or YAML for
// further analysis or testing purposes.
func (t *Terminal) Snapshot() Snapshot {
	t.mu.Lock()
	defer t.mu.Unlock()

	snap := Snapshot{
		Modes: Modes{
			ANSI: maps.Clone(t.ansiModes),
			DEC:  maps.Clone(t.decModes),
		},
		Title:     t.title,
		Rows:      t.rows,
		Cols:      t.cols,
		AltScreen: t.altScreen,
		Cursor: Cursor{
			Position: Position(t.cursorPos),
			Visible:  t.cursorVis,
			Color:    Color{t.cursorColor},
			Style:    t.cursorStyle,
			Blink:    t.cursorBlink,
		},
		BgColor: Color{t.bgColor},
		FgColor: Color{t.fgColor},
		Cells:   make([][]Cell, t.rows),
	}

	for r := 0; r < t.rows; r++ {
		snap.Cells[r] = make([]Cell, t.cols)
		for c := 0; c < t.cols; c++ {
			cell := t.CellAt(c, r)
			snap.Cells[r][c] = Cell{
				Content: cell.Content,
				Style: Style{
					Fg:             Color{cell.Style.Fg},
					Bg:             Color{cell.Style.Bg},
					UnderlineColor: Color{cell.Style.UnderlineColor},
					Underline:      cell.Style.Underline,
					Attrs:          cell.Style.Attrs,
				},
				Link:  Link(cell.Link),
				Width: cell.Width,
			}
		}
	}

	return snap
}
