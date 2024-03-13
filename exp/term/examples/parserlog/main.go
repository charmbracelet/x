package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strconv"

	"github.com/charmbracelet/x/exp/term/ansi"
)

type dispatcher struct {
	str string
}

func (p *dispatcher) Print(r rune) {
	p.str += string(r)
	// fmt.Printf("[Print] %c\n", r)
}

func (p *dispatcher) Execute(code byte) {
	if p.str != "" {
		fmt.Printf("[Print] %s\n", p.str)
		p.str = ""
	}
	fmt.Printf("[Execute] 0x%02x\n", code)
}

func (p *dispatcher) DcsDispatch(marker byte, params [][]uint, inter byte, r byte, data []byte, ignore bool) {
	if p.str != "" {
		fmt.Printf("[Print] %s\n", p.str)
		p.str = ""
	}
	fmt.Printf("[DcsDispatch] marker=%c params=%v, inter=%c, final=%q, data=%q, ignore=%v\n", marker, params, inter, r, data, ignore)
}

func (p *dispatcher) OscDispatch(params [][]byte, bellTerminated bool) {
	if p.str != "" {
		fmt.Printf("[Print] %s\n", p.str)
		p.str = ""
	}
	fmt.Printf("[OscDispatch]")
	for _, param := range params {
		fmt.Printf(" param=%q", param)
	}
	fmt.Printf(" bellTerminated=%v\n", bellTerminated)
}

func (p *dispatcher) CsiDispatch(marker byte, params [][]uint, inter byte, r byte, ignore bool) {
	if p.str != "" {
		fmt.Printf("[Print] %s\n", p.str)
		p.str = ""
	}
	fmt.Print("[CsiDispatch]")
	fmt.Printf(" marker=%c params=%v, inter=%c, final=%c, ignore=%v\n", marker, params, inter, r, ignore)
}

func (p *dispatcher) EscDispatch(inter byte, r byte, ignore bool) {
	if p.str != "" {
		fmt.Printf("[Print] %s\n", p.str)
		p.str = ""
	}
	fmt.Printf("[EscDispatch] inter=%c, final=%c, ignore=%v\n", inter, r, ignore)
}

func (p *dispatcher) SosPmApcDispatch(kind byte, data []byte) {
	if p.str != "" {
		fmt.Printf("[Print] %s\n", p.str)
		p.str = ""
	}
	var k string
	switch kind {
	case ansi.SOS, 'X':
		k = "SOS"
	case ansi.PM, '^':
		k = "PM"
	case ansi.APC, '_':
		k = "APC"
	default:
		k = strconv.Itoa(int(kind))
	}
	fmt.Printf("[SosPmApcDispatch] kind=%v, data=%q\n", k, data)
}

func main() {
	bts, err := io.ReadAll(os.Stdin)
	if err != nil && err != io.EOF {
		log.Fatal(err)
	}
	dispatcher := &dispatcher{}
	parser := ansi.Parser{
		Print:            dispatcher.Print,
		Execute:          dispatcher.Execute,
		DcsDispatch:      dispatcher.DcsDispatch,
		OscDispatch:      dispatcher.OscDispatch,
		CsiDispatch:      dispatcher.CsiDispatch,
		EscDispatch:      dispatcher.EscDispatch,
		SosPmApcDispatch: dispatcher.SosPmApcDispatch,
	}
	parser.Parse(bts)
}
