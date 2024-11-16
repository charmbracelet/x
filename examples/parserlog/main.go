package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/charmbracelet/x/ansi"
)

func main() {
	bts, err := io.ReadAll(os.Stdin)
	if err != nil && err != io.EOF {
		log.Fatal(err)
	}

	var str string
	dispatcher := func(s ansi.Sequence) {
		if _, ok := s.(ansi.Rune); !ok && str != "" {
			fmt.Printf("[Print] %s\n", str)
			str = ""
		}
		switch s := s.(type) {
		case ansi.Rune:
			str += string(s)
		default:
			fmt.Printf("[%T] %q\n", s, s)
		}
	}

	parser := ansi.NewParser(dispatcher)
	for _, b := range bts {
		parser.Advance(b)
	}
	if str != "" {
		fmt.Printf("[Print] %s\n", str)
	}
}
