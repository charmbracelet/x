package style

import (
	"image/color"
	"strconv"
	"strings"
)

// Attribute is a SGR (Select Graphic Rendition) style attribute.
type Attribute = string

// ResetSequence is a SGR (Select Graphic Rendition) style sequence that resets
// all attributes.
const ResetSequence = "\x1b[m"

// SGR (Select Graphic Rendition) style attributes.
const (
	Reset                  Attribute = "0"
	Bold                   Attribute = "1"
	Faint                  Attribute = "2"
	Italic                 Attribute = "3"
	Underline              Attribute = "4"
	DoubleUnderline        Attribute = "4:2"
	CurlyUnderline         Attribute = "4:3"
	DottedUnderline        Attribute = "4:4"
	DashedUnderline        Attribute = "4:5"
	SlowBlink              Attribute = "5"
	RapidBlink             Attribute = "6"
	Reverse                Attribute = "7"
	Conceal                Attribute = "8"
	Strikethrough          Attribute = "9"
	NoBold                 Attribute = "21" // Some terminals treat this as double underline.
	NormalIntensity        Attribute = "22"
	NoItalic               Attribute = "23"
	NoUnderline            Attribute = "24"
	NoBlink                Attribute = "25"
	NoReverse              Attribute = "27"
	NoStrikethrough        Attribute = "29"
	DefaultForegroundColor Attribute = "39"
	DefaultBackgroundColor Attribute = "49"
	DefaultUnderlineColor  Attribute = "59"
)

// Sequence creates a SGR (Select Graphic Rendition) style sequence with the
// given attributes.
// If no attributes are given, it will return a reset sequence.
func Sequence(attrs ...Attribute) string {
	if len(attrs) == 0 {
		return ResetSequence
	}
	return "\x1b" + "[" + strings.Join(attrs, ";") + "m"
}

// Foreground returns the SGR attribute for the given foreground color.
func ForegroundColor(c Color) Attribute {
	switch c := c.(type) {
	case BasicColor:
		// 3-bit or 4-bit ANSI foreground
		// "3<n>" or "9<n>" where n is the color number from 0 to 7
		if c < 8 {
			return "3" + string('0'+c)
		} else if c < 16 {
			return "9" + string('0'+c-8)
		}
	case ExtendedColor:
		// 256-color ANSI foreground
		// "38;5;<n>"
		return "38;5;" + strconv.FormatUint(uint64(c), 10)
	case TrueColor, color.Color:
		// 24-bit "true color" foreground
		// "38;2;<r>;<g>;<b>"
		r, g, b, _ := c.RGBA()
		return "38;2;" +
			strconv.FormatUint(uint64(r), 10) + ";" +
			strconv.FormatUint(uint64(g), 10) + ";" +
			strconv.FormatUint(uint64(b), 10)
	}
	return DefaultForegroundColor
}

// BackgroundColor returns the SGR attribute for the given background color.
func BackgroundColor(c Color) Attribute {
	switch c := c.(type) {
	case BasicColor:
		// 3-bit or 4-bit ANSI foreground
		// "4<n>" or "10<n>" where n is the color number from 0 to 7
		if c < 8 {
			return "4" + string('0'+c)
		} else {
			return "10" + string('0'+c-8)
		}
	case ExtendedColor:
		// 256-color ANSI foreground
		// "48;5;<n>"
		return "48;5;" + strconv.FormatUint(uint64(c), 10)
	case TrueColor, color.Color:
		// 24-bit "true color" foreground
		// "38;2;<r>;<g>;<b>"
		r, g, b, _ := c.RGBA()
		return "48;2;" +
			strconv.FormatUint(uint64(r), 10) + ";" +
			strconv.FormatUint(uint64(g), 10) + ";" +
			strconv.FormatUint(uint64(b), 10)
	}
	return DefaultBackgroundColor
}

// UnderlineColor returns the SGR attribute for the given underline color.
func UnderlineColor(c Color) Attribute {
	switch c := c.(type) {
	// NOTE: we can't use 3-bit and 4-bit ANSI color codes with underline
	// color, use 256-color instead.
	//
	// 256-color ANSI underline color
	// "58;5;<n>"
	case BasicColor:
		return "58;5;" + strconv.FormatUint(uint64(c), 10)
	case ExtendedColor:
		return "58;5;" + strconv.FormatUint(uint64(c), 10)
	case TrueColor, color.Color:
		// 24-bit "true color" foreground
		// "38;2;<r>;<g>;<b>"
		r, g, b, _ := c.RGBA()
		return "58;2;" +
			strconv.FormatUint(uint64(r), 10) + ";" +
			strconv.FormatUint(uint64(g), 10) + ";" +
			strconv.FormatUint(uint64(b), 10)
	}
	return DefaultUnderlineColor
}
