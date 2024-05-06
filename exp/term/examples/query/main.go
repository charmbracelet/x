package main

import (
	"fmt"
	"image/color"
	"log"
	"os"

	"github.com/charmbracelet/x/exp/term"
)

func main() {
	in, out := os.Stdin, os.Stdout
	hasKitty, _ := term.QueryKittyKeyboard(in, out)
	log.Printf("Kitty keyboard support: %v", hasKitty)
	bg, _ := term.QueryBackgroundColor(in, out)
	log.Printf("Background color: %s", colorToHexString(bg))
}

// colorToHexString returns a hex string representation of a color.
func colorToHexString(c color.Color) string {
	if c == nil {
		return ""
	}
	shift := func(v uint32) uint32 {
		if v > 0xff {
			return v >> 8
		}
		return v
	}
	r, g, b, _ := c.RGBA()
	r, g, b = shift(r), shift(g), shift(b)
	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}
