package main

import (
	"fmt"
	"image/color"
	"log"

	"github.com/charmbracelet/x/exp/term"
)

func main() {
	t := term.Current()
	width, height, err := t.GetSize()
	if err != nil {
		log.Fatalf("error getting terminal size: %v", err)
	}

	log.Printf("terminal size: %d x %d", width, height)

	log.Printf("supports kitty keyboard: %v", t.SupportsKittyKeyboard())

	log.Printf("background color: %v", colorToHex(t.BackgroundColor()))

	log.Printf("foreground color: %v", colorToHex(t.ForegroundColor()))

	log.Printf("cursor color: %v", colorToHex(t.CursorColor()))
}

func colorToHex(c color.Color) string {
	r, g, b, _ := c.RGBA()
	r >>= 8
	g >>= 8
	b >>= 8
	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}
