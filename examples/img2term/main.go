package main

import (
	"bytes"
	"flag"
	"fmt"
	"image"
	"io"
	"log"
	"os"

	_ "image/jpeg"
	_ "image/png"

	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/ansi/sixel"
	"github.com/charmbracelet/x/graphics"
)

// $ go run . ./../../ansi/fixtures/graphics/JigokudaniMonkeyPark.png
func main() {
	imageProtocols := graphics.DetectImageProtocols()
	fmt.Println("sixel  supported:", imageProtocols.Sixel)
	fmt.Println("iTerm2 supported:", imageProtocols.ITerm2)
	fmt.Println("kitty  supported", imageProtocols.Kitty)

	flag.Parse()
	args := flag.Args()
	if len(args) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	f, err := os.Open(args[0])
	if err != nil {
		log.Fatal(err)
	}

	defer f.Close() //nolint:errcheck
	img, _, err := image.Decode(f)
	if err != nil {
		log.Fatal(err)
	}

	if _, err := writeSixel(os.Stdout, img); err != nil {
		log.Fatal(err)
	}
}

func writeSixel(w io.Writer, img image.Image) (int, error) {
	var buf bytes.Buffer
	var e sixel.Encoder
	if err := e.Encode(&buf, img); err != nil {
		return 0, err
	}

	return io.WriteString(w, ansi.SixelGraphics(0, 1, 0, buf.Bytes()))
}
