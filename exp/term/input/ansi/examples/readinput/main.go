package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/charmbracelet/x/exp/term"
	"github.com/charmbracelet/x/exp/term/ansi/ctrl"
	"github.com/charmbracelet/x/exp/term/ansi/kitty"
	"github.com/charmbracelet/x/exp/term/ansi/mode"
	"github.com/charmbracelet/x/exp/term/ansi/sys"
	"github.com/charmbracelet/x/exp/term/input"
	"github.com/charmbracelet/x/exp/term/input/ansi"
)

func main() {
	var in io.Reader = os.Stdin
	if !term.IsTerminal(os.Stdin.Fd()) {
		bts, err := io.ReadAll(os.Stdin)
		if err != nil {
			log.Fatalf("error reading from stdin: %v\r\n", err)
		}

		in = bytes.NewReader(bts)
	} else {
		state, err := term.MakeRaw(os.Stdin.Fd())
		if err != nil {
			log.Fatalf("error making raw: %v", err)
		}

		defer term.Restore(os.Stdin.Fd(), state)
	}

	defer io.WriteString(os.Stdout, kitty.Disable(kitty.AllFlags)) // Disable Kitty keyboard
	defer disableMouse()

	rd := ansi.NewDriver(in, os.Getenv("TERM"), 0)

	printHelp()

	var (
		kittyFlags int

		paste bool

		mouse       bool
		mouseHilite bool
		mouseCell   bool
		mouseAll    bool
		mouseExt    bool
	)
	last := input.Event(nil)
OUT:
	for {
		n, err := rd.ReadInput()
		if err == io.EOF {
			break
		}
		if err != nil {
			if errors.Is(err, input.ErrUnknownEvent) {
				log.Printf("%v\r\n", err)
				continue
			}
			log.Fatalf("error reading input: %v\r\n", err)
		}

		if last != nil && len(n) > 0 {
			currKey, ok1 := n[len(n)-1].(input.KeyEvent)
			prevKey, ok2 := last.(input.KeyEvent)
			if ok1 && ok2 && currKey.Sym == 0 && prevKey.Sym == 0 && currKey.Action == 0 && prevKey.Action == 0 {
				prev := string(prevKey.Runes)
				curr := string(currKey.Runes)
				switch {
				case prev == "q" && curr == "q":
					break OUT
				case prev == "h" && curr == "h":
					printHelp()
				case prev == "p" && curr == "p":
					if paste {
						execute(mode.DisableBracketedPaste)
					} else {
						execute(mode.EnableBracketedPaste)
					}
					paste = !paste
				case prev == "k":
					switch curr {
					case "0":
						kittyFlags = 0
						execute(kitty.Disable(kittyFlags))
					case "1":
						if kittyFlags&kitty.DisambiguateEscapeCodes == 0 {
							kittyFlags |= kitty.DisambiguateEscapeCodes
							execute(kitty.Enable(kittyFlags))
						} else {
							kittyFlags &^= kitty.DisambiguateEscapeCodes
							execute(kitty.Disable(kittyFlags))
						}
					case "2":
						if kittyFlags&kitty.ReportEventTypes == 0 {
							kittyFlags |= kitty.ReportEventTypes
							execute(kitty.Enable(kittyFlags))
						} else {
							kittyFlags &^= kitty.ReportEventTypes
							execute(kitty.Disable(kittyFlags))
						}
					case "3":
						if kittyFlags&kitty.ReportAlternateKeys == 0 {
							kittyFlags |= kitty.ReportAlternateKeys
							execute(kitty.Enable(kittyFlags))
						} else {
							kittyFlags &^= kitty.ReportAlternateKeys
							execute(kitty.Disable(kittyFlags))
						}
					case "4":
						if kittyFlags&kitty.ReportAllKeys == 0 {
							kittyFlags |= kitty.ReportAllKeys
							execute(kitty.Enable(kittyFlags))
						} else {
							kittyFlags &^= kitty.ReportAllKeys
							execute(kitty.Disable(kittyFlags))
						}
					case "5":
						if kittyFlags&kitty.ReportAssociatedKeys == 0 {
							kittyFlags |= kitty.ReportAssociatedKeys
							execute(kitty.Enable(kittyFlags))
						} else {
							kittyFlags &^= kitty.ReportAssociatedKeys
							execute(kitty.Disable(kittyFlags))
						}
					}
				case prev == "r":
					switch curr {
					case "k":
						execute(kitty.Request)
					case "b":
						execute(sys.RequestBackgroundColor)
					case "f":
						execute(sys.RequestForegroundColor)
					case "c":
						execute(sys.RequestCursorColor)
					case "d":
						execute(ctrl.RequestPrimaryDeviceAttributes)
					case "x":
						execute(ctrl.RequestXTVersion)
					}
				case prev == "m":
					switch string(currKey.Runes) {
					case "0":
						disableMouse()
					case "1":
						if mouse {
							execute(mode.DisableMouseTracking)
						} else {
							execute(mode.EnableMouseTracking)
						}
						mouse = !mouse
					case "2":
						if mouseHilite {
							execute(mode.DisableHiliteMouseTracking)
						} else {
							execute(mode.EnableHiliteMouseTracking)
						}
						mouseHilite = !mouseHilite
					case "3":
						if mouseCell {
							execute(mode.DisableCellMotionMouseTracking)
						} else {
							execute(mode.EnableCellMotionMouseTracking)
						}
						mouseCell = !mouseCell
					case "4":
						if mouseAll {
							execute(mode.DisableAllMouseTracking)
						} else {
							execute(mode.EnableAllMouseTracking)
						}
						mouseAll = !mouseAll
					case "5":
						if mouseExt {
							execute(mode.DisableSgrMouseExt)
						} else {
							execute(mode.EnableSgrMouseExt)
						}
						mouseExt = !mouseExt
					}
				}
			}
		}

		for _, e := range n {
			log.Printf("event: %s\r\n\r\n", e)
		}

		// Store last keypress
		if len(n) > 0 {
			key, ok := n[len(n)-1].(input.KeyEvent)
			if ok && key.Action == 0 {
				last = key
			}
		}
	}
}

func execute(s string) {
	io.WriteString(os.Stdout, s) // nolint: errcheck
}

func disableMouse() {
	execute(mode.DisableSgrMouseExt)
	execute(mode.DisableAllMouseTracking)
	execute(mode.DisableCellMotionMouseTracking)
	execute(mode.DisableHiliteMouseTracking)
	execute(mode.DisableMouseTracking)
}

func printHelp() {
	fmt.Fprintf(os.Stdout, "Welcome to input demo!\r\n\r\n")
	fmt.Fprintf(os.Stdout, "Press 'qq' to quit.\r\n")
	fmt.Fprintf(os.Stdout, "Press 'hh' to print this help again.\r\n")
	fmt.Fprintf(os.Stdout, "Press 'pp' to toggle bracketed paste mode.\r\n")
	fmt.Fprintf(os.Stdout, "Press 'k' followed by a number to toggle Kitty keyboard protocol flags.\r\n")
	fmt.Fprintf(os.Stdout, "  1: DisambiguateEscapeCodes\r\n")
	fmt.Fprintf(os.Stdout, "  2: ReportEventTypes\r\n")
	fmt.Fprintf(os.Stdout, "  3: ReportAlternateKeys\r\n")
	fmt.Fprintf(os.Stdout, "  4: ReportAllKeys\r\n")
	fmt.Fprintf(os.Stdout, "  5: ReportAssociatedKeys\r\n")
	fmt.Fprintf(os.Stdout, "  0: Disable all flags\r\n")
	fmt.Fprintf(os.Stdout, "\r\n")
	fmt.Fprintf(os.Stdout, "Press 'm' followed by a number to toggle mouse events.\r\n")
	fmt.Fprintf(os.Stdout, "  0: Disable all mouse events\r\n")
	fmt.Fprintf(os.Stdout, "  1: Enable mouse events\r\n")
	fmt.Fprintf(os.Stdout, "  2: Enable mouse events with highlighting\r\n")
	fmt.Fprintf(os.Stdout, "  3: Enable mouse events with cell motion\r\n")
	fmt.Fprintf(os.Stdout, "  4: Enable all mouse events\r\n")
	fmt.Fprintf(os.Stdout, "  5: Enable extended mouse events (SGR)\r\n")
	fmt.Fprintf(os.Stdout, "\r\n")
	fmt.Fprintf(os.Stdout, "Press 'r' followed by a letter to request a terminal capability.\r\n")
	fmt.Fprintf(os.Stdout, "  k: Kitty keyboard protocol flags\r\n")
	fmt.Fprintf(os.Stdout, "  b: Background color\r\n")
	fmt.Fprintf(os.Stdout, "  f: Foreground color\r\n")
	fmt.Fprintf(os.Stdout, "  c: Cursor color\r\n")
	fmt.Fprintf(os.Stdout, "  d: Primary Device Attributes\r\n")
	fmt.Fprintf(os.Stdout, "  x: XTVERSION\r\n")
	fmt.Fprintf(os.Stdout, "\r\n")
}
