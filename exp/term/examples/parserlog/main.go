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

func (p *dispatcher) DcsDispatch(params [][]uint, intermediates [2]byte, r byte, data []byte, ignore bool) {
	fmt.Printf("[DcsDispatch] params=%v, intermediates=%q, final=%q, data=%q, ignore=%v\n", params, intermediates, r, data, ignore)
}

func (p *dispatcher) OscDispatch(params [][]byte, bellTerminated bool) {
	fmt.Printf("[OscDispatch]")
	for _, param := range params {
		fmt.Printf(" param=%q", param)
	}
	fmt.Printf(" bellTerminated=%v\n", bellTerminated)
}

func (p *dispatcher) CsiDispatch(params [][]uint, intermediates [2]byte, r byte, ignore bool) {
	fmt.Print("[CsiDispatch]")
	fmt.Printf(" params=%v, intermediates=%v, final=%c, ignore=%v\n", params, intermediates, r, ignore)
}

func (p *dispatcher) EscDispatch(intermediates [2]byte, r byte, ignore bool) {
	fmt.Printf("[EscDispatch] intermediates=%v, final=%c, ignore=%v\n", intermediates, r, ignore)
}

func main() {
	bts, err := io.ReadAll(os.Stdin)
	if err != nil && err != io.EOF {
		log.Fatal(err)
	}
	dispatcher := &dispatcher{}
	ansi.New(&ansi.Handler{
		Rune:       dispatcher.Print,
		Execute:    dispatcher.Execute,
		DcsHandler: dispatcher.DcsDispatch,
		OscHandler: dispatcher.OscDispatch,
		CsiHandler: dispatcher.CsiDispatch,
		EscHandler: dispatcher.EscDispatch,
	}).Parse(bts)
}
