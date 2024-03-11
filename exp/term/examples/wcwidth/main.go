package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/charmbracelet/x/exp/term/ansi"
	"github.com/rivo/uniseg"
)

func wc(data []byte) {
	s := string(data)
	var words int
	var chars int
	lines := strings.Split(s, "\n")
	for _, line := range lines {
		l := ansi.Strip(line)
		words += len(strings.Fields(l))
		chars += uniseg.StringWidth(l)
	}
	fmt.Println("\t", len(lines)-1, "\t", words, "\t", chars)
}

func main() {
	if len(os.Args) > 1 {
		for _, file := range os.Args[1:] {
			data, err := os.ReadFile(file)
			if err != nil {
				log.Fatalf("error reading file: %v", err)
			}
			wc(data)
		}
	} else {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			log.Fatalf("error reading from stdin: %v", err)
		}
		wc(data)
	}
}
