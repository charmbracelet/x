package input

import (
	"fmt"
	"image/color"
	"strconv"
	"strings"
)

// FgColorEvent represents a foreground color change event.
type FgColorEvent struct{ color.Color }

var _ Event = FgColorEvent{}

// String implements Event.
func (e FgColorEvent) String() string {
	r, g, b, a := e.RGBA()
	r >>= 8
	g >>= 8
	b >>= 8
	a >>= 8
	return fmt.Sprintf("FgColor: [%02x]#%02x%02x%02x", a, r, g, b)
}

// Type implements Event.
func (FgColorEvent) Type() string {
	return "FgColor"
}

// BgColorEvent represents a background color change event.
type BgColorEvent struct{ color.Color }

var _ Event = BgColorEvent{}

// String implements Event.
func (e BgColorEvent) String() string {
	r, g, b, a := e.RGBA()
	r >>= 8
	g >>= 8
	b >>= 8
	a >>= 8
	return fmt.Sprintf("BgColor: [%02x]#%02x%02x%02x", a, r, g, b)
}

// Type implements Event.
func (BgColorEvent) Type() string {
	return "BgColor"
}

// CursorColorEvent represents a cursor color change event.
type CursorColorEvent struct{ color.Color }

var _ Event = CursorColorEvent{}

// String implements Event.
func (e CursorColorEvent) String() string {
	r, g, b, a := e.RGBA()
	r >>= 8
	g >>= 8
	b >>= 8
	a >>= 8
	return fmt.Sprintf("CursorColor: [%02x]#%02x%02x%02x", a, r, g, b)
}

// Type implements Event.
func (CursorColorEvent) Type() string {
	return "CursorColor"
}

func xParseColor(s string) color.Color {
	switch {
	case strings.HasPrefix(s, "rgb:"):
		parts := strings.Split(s[4:], "/")
		if len(parts) != 3 {
			return color.Black
		}

		r, _ := strconv.ParseUint(parts[0], 16, 32)
		g, _ := strconv.ParseUint(parts[1], 16, 32)
		b, _ := strconv.ParseUint(parts[2], 16, 32)

		return color.RGBA{uint8(r), uint8(g), uint8(b), 255}
	case strings.HasPrefix(s, "rgba:"):
		parts := strings.Split(s[5:], "/")
		if len(parts) != 4 {
			return color.Black
		}

		r, _ := strconv.ParseUint(parts[0], 16, 32)
		g, _ := strconv.ParseUint(parts[1], 16, 32)
		b, _ := strconv.ParseUint(parts[2], 16, 32)
		a, _ := strconv.ParseUint(parts[3], 16, 32)

		return color.RGBA{uint8(r), uint8(g), uint8(b), uint8(a)}
	}
	return color.Black
}
