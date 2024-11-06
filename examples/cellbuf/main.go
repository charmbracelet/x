package main

import (
	"log"
	"os"
	"runtime"

	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/cellbuf"
	"github.com/charmbracelet/x/input"
	"github.com/charmbracelet/x/term"
)

func main() {
	w, h, err := term.GetSize(os.Stdout.Fd())
	if err != nil {
		log.Fatalf("getting terminal size: %v", err)
	}

	state, err := term.MakeRaw(os.Stdin.Fd())
	if err != nil {
		log.Fatalf("making raw: %v", err)
	}

	defer term.Restore(os.Stdin.Fd(), state)

	drv, err := input.NewDriver(os.Stdin, os.Getenv("TERM"), 0)
	if err != nil {
		log.Fatalf("creating input driver: %v", err)
	}

	os.Stdout.WriteString(ansi.EnableAltScreenBuffer + ansi.EnableMouseCellMotion + ansi.EnableMouseSgrExt)
	defer os.Stdout.WriteString(ansi.DisableMouseSgrExt + ansi.DisableMouseCellMotion + ansi.DisableAltScreenBuffer)

	var buf cellbuf.Buffer
	var style cellbuf.Style
	style.Reverse(true)
	x, y := (w/2)-8, h/2
	buf.Resize(w, h)

	reset(&buf, x, y)

	if runtime.GOOS != "windows" {
		// Listen for resize events
		go listenForResize(func() {
			updateWinsize(&buf)
			reset(&buf, x, y)
		})
	}

	for {
		evs, err := drv.ReadEvents()
		if err != nil {
			log.Fatalf("reading events: %v", err)
		}

		for _, ev := range evs {
			switch ev := ev.(type) {
			case input.WindowSizeEvent:
				updateWinsize(&buf)
			case input.MouseClickEvent:
				x, y = ev.X, ev.Y
			case input.KeyPressEvent:
				switch ev.String() {
				case "ctrl+c", "q":
					return
				case "left":
					x--
				case "down":
					y++
				case "up":
					y--
				case "right":
					x++
				}
			}
		}

		reset(&buf, x, y)
	}
}

func reset(buf *cellbuf.Buffer, x, y int) {
	buf.Fill(cellbuf.Cell{Content: "你", Width: 2}, nil)
	buf.Paint(0, "\x1b[7m !Hello, world! \x1b[m", &cellbuf.Rectangle{X: x, Y: y, Width: 16, Height: 1})
	os.Stdout.WriteString(ansi.SetCursorPosition(1, 1) + buf.Render())
}

func updateWinsize(buf *cellbuf.Buffer) (w, h int) {
	w, h, _ = term.GetSize(os.Stdout.Fd())
	buf.Resize(w, h)
	return
}
