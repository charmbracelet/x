package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/charmbracelet/x/exp/term/ansi"
)

type dispatcher struct{}

func (p *dispatcher) Print(r rune) {
	fmt.Printf("[Print] %c\n", r)
}

func (p *dispatcher) Execute(code byte) {
	fmt.Printf("[Execute] 0x%02x\n", code)
}

func (p *dispatcher) DcsDispatch(marker byte, params [][]uint, inter byte, r byte, data []byte, ignore bool) {
	fmt.Printf("[DcsDispatch] marker=%c params=%v, inter=%c, final=%q, data=%q, ignore=%v\n", marker, params, inter, r, data, ignore)
}

func (p *dispatcher) OscDispatch(params [][]byte, bellTerminated bool) {
	fmt.Printf("[OscDispatch]")
	for _, param := range params {
		fmt.Printf(" param=%q", param)
	}
	fmt.Printf(" bellTerminated=%v\n", bellTerminated)
}

func (p *dispatcher) CsiDispatch(marker byte, params [][]uint, inter byte, r byte, ignore bool) {
	fmt.Print("[CsiDispatch]")
	fmt.Printf(" marker=%c params=%v, inter=%c, final=%c, ignore=%v\n", marker, params, inter, r, ignore)
}

func (p *dispatcher) EscDispatch(inter byte, r byte, ignore bool) {
	fmt.Printf("[EscDispatch] inter=%c, final=%c, ignore=%v\n", inter, r, ignore)
}

func main() {
	bts, err := io.ReadAll(os.Stdin)
	if err != nil && err != io.EOF {
		log.Fatal(err)
	}
	dispatcher := &dispatcher{}
	ansi.NewParser(ansi.Handler{
		Rune:       dispatcher.Print,
		Execute:    dispatcher.Execute,
		DcsHandler: dispatcher.DcsDispatch,
		OscHandler: dispatcher.OscDispatch,
		CsiHandler: dispatcher.CsiDispatch,
		EscHandler: dispatcher.EscDispatch,
	}).Parse(bts)
}
