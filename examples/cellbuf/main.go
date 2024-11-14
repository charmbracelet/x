package main

import (
	"log"
	"os"
	"runtime"

	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/cellbuf"
	"github.com/charmbracelet/x/input"
	"github.com/charmbracelet/x/term"
	"github.com/charmbracelet/x/vt"
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

	var style cellbuf.Style
	buf := vt.NewBuffer(w, h)
	style.Reverse(true)
	x, y := (w/2)-8, h/2

	reset(buf, x, y)

	if runtime.GOOS != "windows" {
		// Listen for resize events
		go listenForResize(func() {
			updateWinsize(buf)
			reset(buf, x, y)
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
				updateWinsize(buf)
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

		reset(buf, x, y)
	}
}

func reset(buf cellbuf.Buffer, x, y int) {
	cellbuf.Fill(buf, vt.NewCell('你'))
	rect := cellbuf.Rect(x, y, 16, 1)
	cellbuf.Paint(buf, cellbuf.WcWidth, "\x1b[7m !Hello, world! \x1b[m", &rect)
	os.Stdout.WriteString(ansi.SetCursorPosition(1, 1) + cellbuf.Render(buf))
}

func updateWinsize(buf cellbuf.Resizable) (w, h int) {
	w, h, _ = term.GetSize(os.Stdout.Fd())
	buf.Resize(w, h)
	return
}
