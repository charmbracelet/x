package ansi

import (
	"image/color"
	"strconv"
	"strings"
)

// ResetStyle is a SGR (Select Graphic Rendition) style sequence that resets
// all attributes.
// See: https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_(Select_Graphic_Rendition)_parameters
const ResetStyle = "\x1b[m"

// Attr is a SGR (Select Graphic Rendition) style attribute.
type Attr = string

// Style represents an ANSI SGR (Select Graphic Rendition) style.
type Style []Attr

// String returns the ANSI SGR (Select Graphic Rendition) style sequence for
// the given style.
func (s Style) String() string {
	if len(s) == 0 {
		return ResetStyle
	}
	return "\x1b[" + strings.Join(s, ";") + "m"
}

// Styled returns a styled string with the given style applied.
func (s Style) Styled(str string) string {
	if len(s) == 0 {
		return str
	}
	return s.String() + str + ResetStyle
}

// Reset appends the reset style attribute to the style.
func (s Style) Reset() Style {
	return append(s, resetAttr)
}

// Bold appends the bold style attribute to the style.
func (s Style) Bold() Style {
	return append(s, boldAttr)
}

// Faint appends the faint style attribute to the style.
func (s Style) Faint() Style {
	return append(s, faintAttr)
}

// Italic appends the italic style attribute to the style.
func (s Style) Italic() Style {
	return append(s, italicAttr)
}

// Underline appends the underline style attribute to the style.
func (s Style) Underline() Style {
	return append(s, underlineAttr)
}

// DoubleUnderline appends the double underline style attribute to the style.
func (s Style) DoubleUnderline() Style {
	return append(s, doubleUnderlineAttr)
}

// CurlyUnderline appends the curly underline style attribute to the style.
func (s Style) CurlyUnderline() Style {
	return append(s, curlyUnderlineAttr)
}

// DottedUnderline appends the dotted underline style attribute to the style.
func (s Style) DottedUnderline() Style {
	return append(s, dottedUnderlineAttr)
}

// DashedUnderline appends the dashed underline style attribute to the style.
func (s Style) DashedUnderline() Style {
	return append(s, dashedUnderlineAttr)
}

// SlowBlink appends the slow blink style attribute to the style.
func (s Style) SlowBlink() Style {
	return append(s, slowBlinkAttr)
}

// RapidBlink appends the rapid blink style attribute to the style.
func (s Style) RapidBlink() Style {
	return append(s, rapidBlinkAttr)
}

// Reverse appends the reverse style attribute to the style.
func (s Style) Reverse() Style {
	return append(s, reverseAttr)
}

// Conceal appends the conceal style attribute to the style.
func (s Style) Conceal() Style {
	return append(s, concealAttr)
}

// Strikethrough appends the strikethrough style attribute to the style.
func (s Style) Strikethrough() Style {
	return append(s, strikethroughAttr)
}

// NoBold appends the no bold style attribute to the style.
func (s Style) NoBold() Style {
	return append(s, noBoldAttr)
}

// NormalIntensity appends the normal intensity style attribute to the style.
func (s Style) NormalIntensity() Style {
	return append(s, normalIntensityAttr)
}

// NoItalic appends the no italic style attribute to the style.
func (s Style) NoItalic() Style {
	return append(s, noItalicAttr)
}

// NoUnderline appends the no underline style attribute to the style.
func (s Style) NoUnderline() Style {
	return append(s, noUnderlineAttr)
}

// NoBlink appends the no blink style attribute to the style.
func (s Style) NoBlink() Style {
	return append(s, noBlinkAttr)
}

// NoReverse appends the no reverse style attribute to the style.
func (s Style) NoReverse() Style {
	return append(s, noReverseAttr)
}

// NoStrikethrough appends the no strikethrough style attribute to the style.
func (s Style) NoStrikethrough() Style {
	return append(s, noStrikethroughAttr)
}

// DefaultForegroundColor appends the default foreground color style attribute to the style.
func (s Style) DefaultForegroundColor() Style {
	return append(s, defaultForegroundColorAttr)
}

// DefaultBackgroundColor appends the default background color style attribute to the style.
func (s Style) DefaultBackgroundColor() Style {
	return append(s, defaultBackgroundColorAttr)
}

// DefaultUnderlineColor appends the default underline color style attribute to the style.
func (s Style) DefaultUnderlineColor() Style {
	return append(s, defaultUnderlineColorAttr)
}

// ForegroundColor appends the foreground color style attribute to the style.
func (s Style) ForegroundColor(c Color) Style {
	return append(s, foregroundColorAttr(c))
}

// BackgroundColor appends the background color style attribute to the style.
func (s Style) BackgroundColor(c Color) Style {
	return append(s, backgroundColorAttr(c))
}

// UnderlineColor appends the underline color style attribute to the style.
func (s Style) UnderlineColor(c Color) Style {
	return append(s, underlineColorAttr(c))
}

// SGR (Select Graphic Rendition) style attributes.
// See: https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_(Select_Graphic_Rendition)_parameters
const (
	resetAttr                  Attr = "0"
	boldAttr                   Attr = "1"
	faintAttr                  Attr = "2"
	italicAttr                 Attr = "3"
	underlineAttr              Attr = "4"
	doubleUnderlineAttr        Attr = "4:2"
	curlyUnderlineAttr         Attr = "4:3"
	dottedUnderlineAttr        Attr = "4:4"
	dashedUnderlineAttr        Attr = "4:5"
	slowBlinkAttr              Attr = "5"
	rapidBlinkAttr             Attr = "6"
	reverseAttr                Attr = "7"
	concealAttr                Attr = "8"
	strikethroughAttr          Attr = "9"
	noBoldAttr                 Attr = "21" // Some terminals treat this as double underline.
	normalIntensityAttr        Attr = "22"
	noItalicAttr               Attr = "23"
	noUnderlineAttr            Attr = "24"
	noBlinkAttr                Attr = "25"
	noReverseAttr              Attr = "27"
	noStrikethroughAttr        Attr = "29"
	defaultForegroundColorAttr Attr = "39"
	defaultBackgroundColorAttr Attr = "49"
	defaultUnderlineColorAttr  Attr = "59"
)

// Foreground returns the SGR attribute for the given foreground color.
// See: https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_(Select_Graphic_Rendition)_parameters
func foregroundColorAttr(c Color) Attr {
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
	return defaultForegroundColorAttr
}

// backgroundColorAttr returns the SGR attribute for the given background color.
// See: https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_(Select_Graphic_Rendition)_parameters
func backgroundColorAttr(c Color) Attr {
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
	return defaultBackgroundColorAttr
}

// underlineColorAttr returns the SGR attribute for the given underline color.
// See: https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_(Select_Graphic_Rendition)_parameters
func underlineColorAttr(c Color) Attr {
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
	return defaultUnderlineColorAttr
}
