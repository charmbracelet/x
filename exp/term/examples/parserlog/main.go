package main

import (
	"fmt"
	"log"
	"os"

	parser "github.com/charmbracelet/x/exp/term/vtparser"
)

type dispatcher struct{}

func (p *dispatcher) Print(r rune) {
	fmt.Printf("[Print] %c\n", r)
}

func (p *dispatcher) Execute(code byte) {
	fmt.Printf("[Execute] 0x%02x\n", code)
}

func (p *dispatcher) DcsPut(code byte) {
	fmt.Printf("[DcsPut] %02x\n", code)
}

func (p *dispatcher) DcsUnhook() {
	fmt.Printf("[DcsUnhook]\n")
}

func (p *dispatcher) DcsHook(prefix string, params [][]uint16, intermediates []byte, r rune, ignore bool) {
	fmt.Print("[DcsHook]")
	if prefix != "" {
		fmt.Printf(" prefix=%s", prefix)
	}
	fmt.Printf(" params=%v, intermediates=%v, final=%c, ignore=%v\n", params, intermediates, r, ignore)
}

func (p *dispatcher) OscDispatch(params [][]byte, bellTerminated bool) {
	fmt.Printf("[OscDispatch]")
	for _, param := range params {
		fmt.Printf(" param=%q", param)
	}
	fmt.Printf(" bellTerminated=%v\n", bellTerminated)
}

func (p *dispatcher) CsiDispatch(prefix string, params [][]uint16, intermediates []byte, r rune, ignore bool) {
	fmt.Print("[CsiDispatch]")
	if prefix != "" {
		fmt.Printf(" prefix=%s", prefix)
	}
	fmt.Printf(" params=%v, intermediates=%v, final=%c, ignore=%v\n", params, intermediates, r, ignore)
}

func (p *dispatcher) EscDispatch(intermediates []byte, r rune, ignore bool) {
	fmt.Printf("[EscDispatch] intermediates=%v, final=%c, ignore=%v\n", intermediates, r, ignore)
}

func main() {
	dispatcher := &dispatcher{}
	if err := parser.New(dispatcher).Parse(os.Stdin); err != nil {
		log.Fatal(err)
	}
}
