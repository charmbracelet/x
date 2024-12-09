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

	defer term.Restore(os.Stdin.Fd(), state) //nolint:errcheck

	const altScreen = true
	if !altScreen {
		h = 10
	}

	termType := os.Getenv("TERM")
	scr := cellbuf.NewScreen(os.Stdout, &cellbuf.ScreenOptions{
		Width:          w,
		Height:         h,
		Term:           termType,
		RelativeCursor: !altScreen,
		AltScreen:      altScreen,
	})

	defer scr.Close() //nolint:errcheck

	drv, err := input.NewDriver(os.Stdin, termType, 0)
	if err != nil {
		log.Fatalf("creating input driver: %v", err)
	}

	modes := []ansi.Mode{
		ansi.ButtonEventMouseMode,
		ansi.SgrExtMouseMode,
	}

	os.Stdout.WriteString(ansi.SetMode(modes...))         //nolint:errcheck
	defer os.Stdout.WriteString(ansi.ResetMode(modes...)) //nolint:errcheck

	x, y := (w/2)-8, h/2

	render := func() {
		scr.Fill(cellbuf.NewCell('你'))
		text := " !Hello, world! "
		rect := cellbuf.Rect(x, y, ansi.StringWidth(text), 1)

		// This will produce the following escape sequence:
		// "\x1b[7m\x1b]8;;https://charm.sh\x07 ! Hello, world! \x1b]8;;\x07\x1b[m"
		content := ansi.Style{}.Reverse().String() +
			ansi.SetHyperlink("https://charm.sh") +
			text +
			ansi.ResetHyperlink() +
			ansi.ResetStyle

		cellbuf.PaintRect(scr, content, rect)
		scr.Render()
	}

	resize := func(nw, nh int) {
		if !altScreen {
			nh = h
			w = nw
		}
		scr.Resize(nw, nh)
	}

	if runtime.GOOS != "windows" {
		// Listen for resize events
		go listenForResize(func() {
			nw, nh, _ := term.GetSize(os.Stdout.Fd())
			resize(nw, nh)
		})
	}

	// First render
	render()

	for {
		evs, err := drv.ReadEvents()
		if err != nil {
			log.Fatalf("reading events: %v", err)
		}

		for _, ev := range evs {
			switch ev := ev.(type) {
			case input.WindowSizeEvent:
				resize(ev.Width, ev.Height)
			case input.MouseClickEvent:
				x, y = ev.X, ev.Y
			case input.KeyPressEvent:
				switch ev.String() {
				case "ctrl+c", "q":
					return
				case "left", "h":
					x--
				case "down", "j":
					y++
				case "up", "k":
					y--
				case "right", "l":
					x++
				}
			}
		}

		render()
	}
}

func init() {
	f, err := os.OpenFile("cellbuf.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0o666)
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(f)
}
