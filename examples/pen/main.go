package main

import (
	"io"
	"log"
	"os"

	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/cellbuf"
)

func main() {
	pw := cellbuf.NewPenWriter(os.Stdout)
	defer pw.Close()

	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}

	io.WriteString(pw, ansi.Wrap(string(data), 10, ""))
}
