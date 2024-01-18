package internal

import (
	"fmt"
	"image/color"
)

// ColorToHexString returns a hex string representation of a color.
func ColorToHexString(c color.Color) string {
	if c == nil {
		return ""
	}
	r, g, b, _ := c.RGBA()
	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}

// RgbToHex converts red, green, and blue values to a hexadecimal value.
//
//	hex := RgbToHex(0, 0, 255) // 0x0000FF
func RgbToHex(r, g, b uint32) uint32 {
	return r<<16 + g<<8 + b
}
