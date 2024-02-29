package term

import (
	"errors"
	"image/color"
	"io"
	"os"
	"time"

	"github.com/charmbracelet/x/exp/term/ansi/ctrl"
	"github.com/charmbracelet/x/exp/term/ansi/kitty"
	"github.com/charmbracelet/x/exp/term/ansi/sys"
	"github.com/charmbracelet/x/exp/term/input"
	"github.com/muesli/cancelreader"
)

type terminal struct {
	input  io.Reader
	output io.Writer

	inFile  *os.File
	outFile *os.File

	inIstty  bool
	outIstty bool

	isRaw     bool
	prevState *State

	inputHandler input.Driver

	kittyFlags int
	bgColor    color.Color
	fgColor    color.Color
	curColor   color.Color
}

var _ Terminal = &terminal{}

func newTerminal(input io.Reader, output io.Writer) *terminal {
	c := new(terminal)
	c.input = input
	c.output = output
	c.kittyFlags = -1
	return c
}

// Read implements io.Reader.
func (c *terminal) Read(p []byte) (n int, err error) {
	return c.input.Read(p)
}

// Write implements io.Writer.
func (c *terminal) Write(p []byte) (n int, err error) {
	return c.output.Write(p)
}

func (c *terminal) initTerminal() {
	if c.inFile != nil && c.outFile != nil {
		return
	}
	if inFile, ok := c.input.(*os.File); ok {
		c.inFile = inFile
		c.inIstty = IsTerminal(inFile.Fd())
	}
	if outFile, ok := c.output.(*os.File); ok {
		c.outFile = outFile
		c.outIstty = IsTerminal(outFile.Fd())
	}
}

// IsTerminal implements Terminal.
func (c *terminal) IsTerminal() bool {
	c.initTerminal()
	return c.inIstty && c.outIstty
}

// MakeRaw implements Terminal.
func (c *terminal) MakeRaw() (rErr error) {
	c.initTerminal()
	if c.inFile == nil {
		return ErrNotTerminal
	}
	defer func() {
		c.isRaw = rErr == nil
	}()
	if err := control(c.inFile, func(fd uintptr) {
		c.prevState, rErr = MakeRaw(fd)
	}); err != nil {
		return err
	}
	return
}

// Restore implements Terminal.
func (c *terminal) Restore() (rErr error) {
	c.initTerminal()
	if c.inFile == nil {
		return ErrNotTerminal
	}
	defer func() {
		if rErr == nil {
			c.isRaw = false
		}
	}()
	if err := control(c.inFile, func(fd uintptr) {
		rErr = Restore(fd, c.prevState)
	}); err != nil {
		return err
	}
	return
}

// GetSize implements Terminal.
func (c *terminal) GetSize() (w int, h int, rErr error) {
	c.initTerminal()
	if c.inFile == nil {
		return 0, 0, ErrNotTerminal
	}
	if err := control(c.inFile, func(fd uintptr) {
		w, h, rErr = GetSize(fd)
	}); err != nil {
		return 0, 0, err
	}
	return
}

func control(f *os.File, cb func(fd uintptr)) error {
	conn, err := f.SyscallConn()
	if err != nil {
		return err
	}

	return conn.Control(cb)
}

func (c *terminal) initInputHandler() {
	if c.inputHandler != nil {
		return
	}
	c.inputHandler = input.NewDriver(c.input, os.Getenv("TERM"), 0)
}

// SupportsKittyKeyboard implements Terminal.
func (c *terminal) SupportsKittyKeyboard() bool {
	if c.kittyFlags != -1 {
		return true
	}

	if !c.isRaw {
		if err := c.MakeRaw(); err != nil {
			return false
		}
		defer c.Restore() // nolint: errcheck
	}

	c.queryTerminal()

	return c.kittyFlags != -1
}

func (c *terminal) queryTerminal() {
	c.initInputHandler()
	const query = sys.RequestBackgroundColor +
		sys.RequestForegroundColor +
		sys.RequestCursorColor +
		kitty.Request +
		ctrl.RequestPrimaryDeviceAttributes

	evc := make(chan input.Event)
	go func() {
		for {
			ev, err := c.inputHandler.ReadInput()
			if errors.Is(err, cancelreader.ErrCanceled) {
				return
			}
			if err != nil {
				return
			}
			for _, e := range ev {
				evc <- e
			}
		}
	}()

	io.WriteString(c.output, query) // nolint: errcheck

loop:
	for {
		select {
		case <-time.After(1 * time.Second):
			break loop
		case e := <-evc:
			switch e := e.(type) {
			case input.KittyKeyboardEvent:
				c.kittyFlags = int(e)
			case input.BgColorEvent:
				c.bgColor = e.Color
			case input.FgColorEvent:
				c.fgColor = e.Color
			case input.CursorColorEvent:
				c.curColor = e.Color
			case input.PrimaryDeviceAttributesEvent:
				if c.bgColor == nil {
					c.bgColor = color.Black
				}
				if c.fgColor == nil {
					c.fgColor = color.White
				}
				if c.curColor == nil {
					c.curColor = color.White
				}
				break loop
			}
		}
	}

	close(evc)
	c.inputHandler.Cancel()
	c.inputHandler.Close() // nolint: errcheck
}

func (c *terminal) colorAttr(ptr *color.Color, def color.Color) color.Color {
	if ptr != nil && *ptr != nil {
		return *ptr
	}

	if !c.isRaw {
		if err := c.MakeRaw(); err != nil {
			return def
		}
		defer c.Restore() // nolint: errcheck
	}

	c.queryTerminal()

	if ptr != nil {
		return *ptr
	}

	return color.Black
}

// BackgroundColor implements Terminal.
func (c *terminal) BackgroundColor() color.Color {
	return c.colorAttr(&c.bgColor, color.Black)
}

// CursorColor implements Terminal.
func (c *terminal) CursorColor() color.Color {
	return c.colorAttr(&c.curColor, color.White)
}

// ForegroundColor implements Terminal.
func (c *terminal) ForegroundColor() color.Color {
	return c.colorAttr(&c.fgColor, color.White)
}
