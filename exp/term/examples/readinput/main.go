package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/charmbracelet/x/exp/term"
	"github.com/charmbracelet/x/exp/term/ansi"
	"github.com/charmbracelet/x/exp/term/input"
)

func init() {
	// suppress logger time prefix
	log.SetFlags(0)
}

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

	defer io.WriteString(os.Stdout, ansi.PopKittyKeyboard(ansi.KittyAllFlags)) // Disable Kitty keyboard
	defer disableMouse()
	defer execute(ansi.DisableWin32Input)

	rd, err := input.NewDriver(in, os.Getenv("TERM"), 0)
	if err != nil {
		log.Printf("error creating driver: %v\r\n", err)
		return
	}

	defer rd.Cancel()
	defer rd.Close()

	printHelp()

	var (
		kittyFlags int

		paste bool

		mouse       bool
		mouseHilite bool
		mouseCell   bool
		mouseAll    bool
		mouseExt    bool

		win32Input bool
	)
	last := input.Event(nil)
	var buffer [16]input.Event
OUT:
	for {
		n, err := rd.ReadInput(buffer[:])
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

		buf := buffer[:n]
		if last != nil && len(buf) > 0 {
			currKey, ok1 := buf[len(buf)-1].(input.KeyEvent)
			prevKey, ok2 := last.(input.KeyEvent)
			if ok1 && ok2 && currKey.Sym == 0 && prevKey.Sym == 0 && currKey.Action == 0 && prevKey.Action == 0 {
				prev := string(prevKey.Rune)
				curr := string(currKey.Rune)
				switch {
				case prev == "q" && curr == "q":
					break OUT
				case prev == "h" && curr == "h":
					printHelp()
				case prev == "p" && curr == "p":
					if paste {
						execute(ansi.DisableBracketedPaste)
					} else {
						execute(ansi.EnableBracketedPaste)
					}
					paste = !paste
				case prev == "w" && curr == "m":
					if win32Input {
						execute(ansi.DisableWin32Input)
					} else {
						execute(ansi.EnableWin32Input)
					}
					win32Input = !win32Input
				case prev == "k":
					switch curr {
					case "0":
						kittyFlags = 0
						execute(ansi.PopKittyKeyboard(kittyFlags))
					case "1":
						if kittyFlags&ansi.KittyDisambiguateEscapeCodes == 0 {
							kittyFlags |= ansi.KittyDisambiguateEscapeCodes
							execute(ansi.PushKittyKeyboard(kittyFlags))
						} else {
							kittyFlags &^= ansi.KittyDisambiguateEscapeCodes
							execute(ansi.PopKittyKeyboard(kittyFlags))
						}
					case "2":
						if kittyFlags&ansi.KittyReportEventTypes == 0 {
							kittyFlags |= ansi.KittyReportEventTypes
							execute(ansi.PushKittyKeyboard(kittyFlags))
						} else {
							kittyFlags &^= ansi.KittyReportEventTypes
							execute(ansi.PopKittyKeyboard(kittyFlags))
						}
					case "3":
						if kittyFlags&ansi.KittyReportAlternateKeys == 0 {
							kittyFlags |= ansi.KittyReportAlternateKeys
							execute(ansi.PushKittyKeyboard(kittyFlags))
						} else {
							kittyFlags &^= ansi.KittyReportAlternateKeys
							execute(ansi.PopKittyKeyboard(kittyFlags))
						}
					case "4":
						if kittyFlags&ansi.KittyReportAllKeys == 0 {
							kittyFlags |= ansi.KittyReportAllKeys
							execute(ansi.PushKittyKeyboard(kittyFlags))
						} else {
							kittyFlags &^= ansi.KittyReportAllKeys
							execute(ansi.PopKittyKeyboard(kittyFlags))
						}
					case "5":
						if kittyFlags&ansi.KittyReportAssociatedKeys == 0 {
							kittyFlags |= ansi.KittyReportAssociatedKeys
							execute(ansi.PushKittyKeyboard(kittyFlags))
						} else {
							kittyFlags &^= ansi.KittyReportAssociatedKeys
							execute(ansi.PopKittyKeyboard(kittyFlags))
						}
					}
				case prev == "r":
					switch curr {
					case "k":
						execute(ansi.RequestKittyKeyboard)
					case "b":
						execute(ansi.RequestBackgroundColor)
					case "f":
						execute(ansi.RequestForegroundColor)
					case "c":
						execute(ansi.RequestCursorColor)
					case "d":
						execute(ansi.RequestPrimaryDeviceAttributes)
					case "x":
						execute(ansi.RequestXTVersion)
					}
				case prev == "m":
					switch string(currKey.Rune) {
					case "0":
						disableMouse()
					case "1":
						if mouse {
							execute(ansi.DisableMouse)
						} else {
							execute(ansi.EnableMouse)
						}
						mouse = !mouse
					case "2":
						if mouseHilite {
							execute(ansi.DisableMouseHilite)
						} else {
							execute(ansi.EnableMouseHilite)
						}
						mouseHilite = !mouseHilite
					case "3":
						if mouseCell {
							execute(ansi.DisableMouseCellMotion)
						} else {
							execute(ansi.EnableMouseCellMotion)
						}
						mouseCell = !mouseCell
					case "4":
						if mouseAll {
							execute(ansi.DisableMouseAllMotion)
						} else {
							execute(ansi.EnableMouseAllMotion)
						}
						mouseAll = !mouseAll
					case "5":
						if mouseExt {
							execute(ansi.DisableMouseSgrExt)
						} else {
							execute(ansi.EnableMouseSgrExt)
						}
						mouseExt = !mouseExt
					}
				}
			}
		}

		for _, e := range buf {
			if _, ok := e.(fmt.Stringer); ok {
				log.Printf("=== %T: %s\r\n\r\n", e, e)
			} else {
				log.Printf("=== %T\r\n\r\n", e)
			}
		}

		// Store last keypress
		if len(buf) > 0 {
			key, ok := buf[len(buf)-1].(input.KeyEvent)
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
	execute(ansi.DisableMouseSgrExt)
	execute(ansi.DisableMouseAllMotion)
	execute(ansi.DisableMouseCellMotion)
	execute(ansi.DisableMouseHilite)
	execute(ansi.DisableMouse)
}

func printHelp() {
	fmt.Fprintf(os.Stdout, "Welcome to input demo!\r\n\r\n")
	fmt.Fprintf(os.Stdout, "Press 'qq' to quit.\r\n")
	fmt.Fprintf(os.Stdout, "Press 'hh' to print this help again.\r\n")
	fmt.Fprintf(os.Stdout, "Press 'pp' to toggle bracketed paste mode.\r\n")
	fmt.Fprintf(os.Stdout, "Press 'wm' to toggle Win32 input mode.\r\n")
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
