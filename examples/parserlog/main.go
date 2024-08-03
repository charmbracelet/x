package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/ansi/parser"
)

func main() {
	bts, err := io.ReadAll(os.Stdin)
	if err != nil && err != io.EOF {
		log.Fatal(err)
	}

	var str string
	parser := ansi.NewParser(parser.MaxParamsSize, 0)
	dispatcher := func(s ansi.Sequence) {
		if _, ok := s.(ansi.Rune); !ok && str != "" {
			fmt.Printf("[Print] %s\n", str)
			str = ""
		}
		switch s := s.(type) {
		case ansi.Rune:
			str += string(s)
		case ansi.ControlCode:
			fmt.Printf("[ControlCode] %q\n", s)
		case ansi.EscSequence:
			fmt.Print("[EscSequence] ")
			fmt.Printf("Cmd=%c ", s.Command())
			if intermed := s.Intermediate(); intermed != 0 {
				fmt.Printf("Inter=%c", intermed)
			}
			fmt.Println()
		case ansi.CsiSequence:
			fmt.Print("[CsiSequence] ")
			fmt.Printf("Cmd=%q ", s.Command())
			if marker := s.Marker(); marker != 0 {
				fmt.Printf("Marker=%q, ", marker)
			}
			if intermed := s.Intermediate(); intermed != 0 {
				fmt.Printf("Intermed=%q, ", intermed)
			}
			for i := 0; i < s.Len(); i++ {
				if i == 0 {
					fmt.Printf("Params=[")
				}
				fmt.Printf("%+v", s.Subparams(i))
				if i != s.Len()-1 {
					fmt.Print(", ")
				}
				if i == s.Len()-1 {
					fmt.Print("]")
				}
			}
			fmt.Println()
		case ansi.OscSequence:
			fmt.Print("[OscSequence] ")
			fmt.Printf("Cmd=%d ", s.Command())
			fmt.Printf("Params=%+v\n", s.Params())
		case ansi.DcsSequence:
			fmt.Print("[DcsSequence] ")
			fmt.Printf("Cmd=%q ", s.Command())
			if marker := s.Marker(); marker != 0 {
				fmt.Printf("Marker=%q, ", marker)
			}
			if intermed := s.Intermediate(); intermed != 0 {
				fmt.Printf("Intermed=%q, ", intermed)
			}
			for i := 0; i < s.Len(); i++ {
				if i == 0 {
					fmt.Printf("Params=[")
				}
				fmt.Printf("%+v", s.Subparams(i))
				if i != s.Len()-1 {
					fmt.Print(", ")
				}
				if i == s.Len()-1 {
					fmt.Print("] ")
				}
			}
			fmt.Printf("Data=%q\n", s.Data)
		case ansi.SosSequence:
			fmt.Printf("[SosSequence] Data=%q\n", s.Data)
		case ansi.PmSequence:
			fmt.Printf("[PmSequence] Data=%q\n", s.Data)
		case ansi.ApcSequence:
			fmt.Printf("[ApcSequence] Data=%q\n", s.Data)
		}
	}

	parser.Parse(dispatcher, bts)
	if str != "" {
		fmt.Printf("[Print] %s\n", str)
	}
}
