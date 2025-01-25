package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/charmbracelet/x/ansi"
)

func main() {
	input, err := io.ReadAll(os.Stdin)
	if err != nil && err != io.EOF {
		log.Fatal(err)
	}

	var state byte
	p := ansi.NewParser()
	for len(input) > 0 {
		seq, width, n, newState := ansi.DecodeSequence(input, state, p)
		switch {
		case ansi.HasOscPrefix(seq):
			fmt.Printf("OSC sequence: %q, cmd: %d, data: %q", seq, p.Command(), p.Data())
			fmt.Println()
		case ansi.HasDcsPrefix(seq):
			c := ansi.Cmd(p.Command())
			intermed, prefix, cmd := c.Intermediate(), c.Prefix(), c.Final()
			fmt.Printf("DCS sequence: %q,", seq)
			if intermed != 0 {
				fmt.Printf(" intermed: %q,", intermed)
			}
			if prefix != 0 {
				fmt.Printf(" prefix: %q,", prefix)
			}
			if cmd != 0 {
				fmt.Printf(" cmd: %q,", cmd)
			}
			fmt.Print(" params: [")
			var more bool
			params := p.Params()
			for i, r := range params {
				param, hasMore := r.Param(-1), r.HasMore()
				if more != hasMore {
					fmt.Print("[")
				}
				if param == -1 {
					fmt.Print("MISSING")
				} else {
					fmt.Printf("%d", param)
				}
				if i != len(params)-1 {
					fmt.Print(", ")
				}
				if more != hasMore {
					fmt.Print("]")
				}
				more = hasMore
			}
			fmt.Printf("], data: %q", p.Data())
			fmt.Println()

		case ansi.HasSosPrefix(seq):
			fmt.Printf("SOS sequence: %q, data: %q", seq, p.Data())
			fmt.Println()
		case ansi.HasPmPrefix(seq):
			fmt.Printf("PM sequence: %q, data: %q", seq, p.Data())
			fmt.Println()
		case ansi.HasApcPrefix(seq):
			fmt.Printf("APC sequence: %q, data: %q", seq, p.Data())
			fmt.Println()
		case ansi.HasCsiPrefix(seq):
			c := ansi.Cmd(p.Command())
			intermed, prefix, cmd := c.Intermediate(), c.Prefix(), c.Final()
			fmt.Printf("CSI sequence: %q,", seq)
			if intermed != 0 {
				fmt.Printf(" intermed: %q,", intermed)
			}
			if prefix != 0 {
				fmt.Printf(" prefix: %q,", prefix)
			}
			if cmd != 0 {
				fmt.Printf(" cmd: %q,", cmd)
			}
			fmt.Print(" params: [")
			var more bool
			params := p.Params()
			for i, r := range params {
				param, hasMore := r.Param(-1), r.HasMore()
				if hasMore && more != hasMore {
					fmt.Print("[")
				}
				if param == -1 {
					fmt.Print("MISSING")
				} else {
					fmt.Printf("%d", param)
				}
				if !hasMore && more != hasMore {
					fmt.Print("]")
				}
				if i != len(params)-1 {
					fmt.Print(", ")
				}
				more = hasMore
			}
			fmt.Print("]")
			fmt.Println()

		case ansi.HasEscPrefix(seq):
			if !bytes.Equal(seq, []byte{ansi.ESC}) {
				c := ansi.Cmd(p.Command())
				intermed, cmd := c.Intermediate(), c.Final()
				fmt.Printf("ESC sequence: %q", seq)
				if intermed != 0 {
					fmt.Printf(", intermed: %q", intermed)
				}
				if cmd != 0 {
					fmt.Printf(", cmd: %q", cmd)
				}
				fmt.Println()
				break
			}
			fallthrough
		default:
			if width > 0 {
				fmt.Printf("Print: %s, width: %d", seq, width)
			} else {
				fmt.Printf("Execute: %q", seq)
			}
			fmt.Println()
		}
		state = newState
		input = input[n:]
	}
}
