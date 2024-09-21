package color

import (
	"fmt"
	"image/color"
	"math"
)

// Color is convenience type that wraps an RGBA color with additional methods
// for converion.
type Color struct {
	v color.RGBA
}

// RGBA returns the RGBA values of the color.
func (c Color) RGBA() (r, g, b, a uint32) {
	return c.v.RGBA()
}

// Hex returns the hex value of the color as a string.
func (c Color) Hex() string {
	return ColorToHex(c)
}

// HSV returns the HSV values of the color.
func (c Color) HSV() (h, s, v float64) {
	return ColorToHSV(c)
}

// FromHSV sets the color from HSV values.
func (c *Color) FromHSV(h, s, v float64) {
	c.v = HSVToRGBA(h, s, v)
}

// FromRGB sets the color from RGB values.
func (c *Color) FromRGB(r, g, b uint8) {
	c.v.R = r
	c.v.G = g
	c.v.B = b
	c.v.A = 255
}

// ColorToHex converts a color to a hex string.
func ColorToHex(c color.Color) string {
	r, g, b, _ := c.RGBA()
	return fmt.Sprintf("#%02X%02X%02X", r>>8, g>>8, b>>8)
}

// HSVToRGBA converts HSV values to an RGBA color. HSV values should be in the
// ranges [0, 360], [0, 1], and [0, 1] respectively.
func HSVToRGBA(h, s, v float64) color.RGBA {
	h = math.Mod(h, 360)            // Ensure h is in the range [0, 360]
	s = math.Max(0, math.Min(1, s)) // Clamp s to [0, 1]
	v = math.Max(0, math.Min(1, v)) // Clamp v to [0, 1]

	c := v * s
	x := c * (1 - math.Abs(math.Mod(h/60, 2)-1))
	m := v - c

	var r, g, b float64

	switch {
	case h < 60:
		r, g, b = c, x, 0
	case h < 120:
		r, g, b = x, c, 0
	case h < 180:
		r, g, b = 0, c, x
	case h < 240:
		r, g, b = 0, x, c
	case h < 300:
		r, g, b = x, 0, c
	default:
		r, g, b = c, 0, x
	}

	r = math.Round((r + m) * 255)
	g = math.Round((g + m) * 255)
	b = math.Round((b + m) * 255)

	return color.RGBA{
		R: uint8(r),
		G: uint8(g),
		B: uint8(b),
		A: 255, // Full opacity
	}
}

// ColorToHSV converts a color to HSV.
//
// A few things to note:
//
//   - Due to rounding errors, we mayget a slightly different color when
//     converting back to RGB.
//   - The hue will be rounded to the nearest degree, while the saturation and
//     value will be rounded to two decimal places.
func ColorToHSV(c color.Color) (h, s, v float64) {
	r, g, b, _ := c.RGBA()

	// Convert from uint32 (0-MaxUint32) to float64 (0-1).
	rf := float64(r) / float64(0xFFFF)
	gf := float64(g) / float64(0xFFFF)
	bf := float64(b) / float64(0xFFFF)

	minimum := math.Min(rf, math.Min(gf, bf))
	maximum := math.Max(rf, math.Max(gf, bf))

	v = maximum
	delta := maximum - minimum

	if maximum == 0 {
		// Black.
		s = 0
		h = 0
		return
	}

	s = delta / maximum

	if delta == 0 {
		// Gray.
		h = 0
		return
	}

	switch maximum {
	case rf:
		h = (gf - bf) / delta
		if gf < bf {
			h += 6
		}
	case gf:
		h = 2 + (bf-rf)/delta
	case bf:
		h = 4 + (rf-gf)/delta
	}

	h *= 60

	h = math.Round(h)           // Round to nearest degree.
	s = math.Round(s*100) / 100 // Round to two decimal places.
	v = math.Round(v*100) / 100 // Round to two decimal places.

	return
}
