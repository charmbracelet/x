// Package main is a parser example.
package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/charmbracelet/x/ansi"
)

func main() {
	bts, err := io.ReadAll(os.Stdin)
	if err != nil && err != io.EOF {
		log.Fatal(err)
	}
	var str string
	printStr := func() {
		if str != "" {
			fmt.Printf("[Print] %s\n", str)
			str = ""
		}
	}

	parser := ansi.NewParser()
	parser.SetHandler(ansi.Handler{
		Print: func(r rune) { str += string(r) },
		Execute: func(b byte) {
			printStr()
			fmt.Printf("[Execute] %q\n", b)
		},
		HandleCsi: func(cmd ansi.Cmd, params ansi.Params) {
			printStr()
			fmt.Printf("[CSI] %q\n", csiString(cmd, params))
		},
		HandleEsc: func(cmd ansi.Cmd) {
			printStr()
			fmt.Printf("[ESC] ")
			if inter := cmd.Intermediate(); inter != 0 {
				fmt.Printf("%c ", inter)
			}
			if final := cmd.Final(); final != 0 {
				fmt.Printf("%c", final)
			}
			fmt.Println()
		},
		HandleDcs: func(cmd ansi.Cmd, params ansi.Params, data []byte) {
			printStr()
			fmt.Printf("[DCS] %q %q\n", csiString(cmd, params), string(data))
		},
		HandleOsc: func(cmd int, data []byte) {
			printStr()
			fmt.Printf("[OSC] %d %q\n", cmd, data)
		},
		HandleApc: func(data []byte) {
			printStr()
			fmt.Printf("[APC] %q\n", string(data))
		},
		HandleSos: func(data []byte) {
			printStr()
			fmt.Printf("[SOS] %q\n", string(data))
		},
		HandlePm: func(data []byte) {
			printStr()
			fmt.Printf("[PM] %q\n", string(data))
		},
	})
	for _, b := range bts {
		parser.Advance(b)
	}

	printStr()
}

func csiString(cmd ansi.Cmd, params ansi.Params) string {
	var s strings.Builder
	s.WriteString("CSI ")
	if mark := cmd.Prefix(); mark != 0 {
		s.WriteByte(mark)
	}
	params.ForEach(-1, func(i, p int, more bool) {
		s.WriteString(fmt.Sprintf("%d", p))
		if i < len(params)-1 {
			if more {
				s.WriteByte(':')
			} else {
				s.WriteByte(';')
			}
		}
	})
	if inter := cmd.Intermediate(); inter != 0 {
		s.WriteByte(inter)
	}
	if final := cmd.Final(); final != 0 {
		s.WriteByte(final)
	}
	return s.String()
}
