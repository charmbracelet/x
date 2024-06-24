package main

import (
	"flag"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/charmbracelet/x/exp/term/ansi"
)

var (
	length = flag.Int("l", 80, "length of the output")
	tail   = flag.String("tail", "", "tail of the output")
)

func main() {
	flag.Parse()

	var err error
	input := strings.Join(flag.Args(), " ")
	input, err = strconv.Unquote(`"` + input + `"`)
	if err != nil {
		log.Fatalf("could not unquote input: %v", err)
	}

	output := ansi.Truncate(input, *length, *tail)
	output = strconv.Quote(output)
	output = output[1 : len(output)-1] // remove quotes
	fmt.Print(output)
}
