package main

import (
	"bytes"
	"io"
	"log"
	"os"

	"github.com/charmbracelet/x/exp/term"
	"github.com/charmbracelet/x/exp/term/input"
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

	rd := input.NewDriver(in, os.Getenv("TERM"), 0)

	evs, err := rd.PeekInput(10)
	if err != nil {
		log.Fatalf("error peeking input: %v\r\n", err)
	}
	for _, e := range evs {
		log.Printf("=== %T: %s\r\n\r\n", e, e)
	}
}
