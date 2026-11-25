package ansi

import (
	"fmt"
	"image/color"
	"strconv"
	"strings"

	"github.com/lucasb-eyer/go-colorful"
)

// colorToHexString returns a hex string representation of a color.
func colorToHexString(c color.Color) string { //nolint:unused
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

// rgbToHex converts red, green, and blue values to a hexadecimal value.
//
//	hex := rgbToHex(0, 0, 255) // 0x0000FF
func rgbToHex(r, g, b uint32) uint32 { //nolint:unused
	return r<<16 + g<<8 + b
}

type shiftable interface {
	~uint | ~uint16 | ~uint32 | ~uint64
}

func shift[T shiftable](x T) T {
	if x > 0xff {
		x >>= 8
	}
	return x
}

// XParseColor is a helper function that parses a string into a color.Color. It
// provides a similar interface to the XParseColor function in Xlib. It
// supports the following formats:
//
//   - #RGB
//   - #RRGGBB
//   - rgb:RRRR/GGGG/BBBB
//   - rgba:RRRR/GGGG/BBBB/AAAA
//
// The rgb: and rgba: formats accept one to four hex digits per channel.
// If the string is not a valid color, nil is returned.
//
// See: https://linux.die.net/man/3/xparsecolor
func XParseColor(s string) color.Color {
	switch {
	case strings.HasPrefix(s, "#"):
		c, err := colorful.Hex(s)
		if err != nil {
			return nil
		}

		return c
	case strings.HasPrefix(s, "rgb:"):
		return parseXColor(s[4:], 3)
	case strings.HasPrefix(s, "rgba:"):
		return parseXColor(s[5:], 4)
	}
	return nil
}

func parseXColor(s string, numChannels int) color.Color {
	parts := strings.Split(s, "/")
	if len(parts) != numChannels {
		return nil
	}

	channels := [4]uint8{3: 0xff}
	for i, part := range parts {
		if len(part) == 0 || len(part) > 4 {
			return nil
		}
		value, err := strconv.ParseUint(part, 16, 16)
		if err != nil {
			return nil
		}

		// Scale to the 16-bit range as Xlib does, then take the high byte.
		maxValue := (uint64(1) << (4 * len(part))) - 1
		channels[i] = uint8((value * 0xffff / maxValue) >> 8) //nolint:gosec // value <= maxValue
	}
	return color.RGBA{channels[0], channels[1], channels[2], channels[3]}
}
