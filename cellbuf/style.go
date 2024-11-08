package cellbuf

import (
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"
)

// AttrMask is a bitmask for text attributes that can change the look of text.
// These attributes can be combined to create different styles.
type AttrMask uint8

// These are the available text attributes that can be combined to create
// different styles.
const (
	BoldAttr AttrMask = 1 << iota
	FaintAttr
	ItalicAttr
	SlowBlinkAttr
	RapidBlinkAttr
	ReverseAttr
	ConcealAttr
	StrikethroughAttr

	ResetAttr AttrMask = 0
)

var attrMaskNames = map[AttrMask]string{
	BoldAttr:          "BoldAttr",
	FaintAttr:         "FaintAttr",
	ItalicAttr:        "ItalicAttr",
	SlowBlinkAttr:     "SlowBlinkAttr",
	RapidBlinkAttr:    "RapidBlinkAttr",
	ReverseAttr:       "ReverseAttr",
	ConcealAttr:       "ConcealAttr",
	StrikethroughAttr: "StrikethroughAttr",
	ResetAttr:         "ResetAttr",
}

// UnderlineStyle is the style of underline to use for text.
type UnderlineStyle uint8

// These are the available underline styles.
const (
	NoUnderline UnderlineStyle = iota
	SingleUnderline
	DoubleUnderline
	CurlyUnderline
	DottedUnderline
	DashedUnderline
)

var underlineStyleNames = map[UnderlineStyle]string{
	NoUnderline:     "NoUnderline",
	SingleUnderline: "SingleUnderline",
	DoubleUnderline: "DoubleUnderline",
	CurlyUnderline:  "CurlyUnderline",
	DottedUnderline: "DottedUnderline",
	DashedUnderline: "DashedUnderline",
}

// String returns a string representation of the underline style.
func (u UnderlineStyle) String() string {
	return underlineStyleNames[u]
}

// Style represents the Style of a cell.
type Style struct {
	Fg      ansi.Color
	Bg      ansi.Color
	Ul      ansi.Color
	Attrs   AttrMask
	UlStyle UnderlineStyle
}

// Sequence returns the ANSI sequence that sets the style.
func (s Style) Sequence() string {
	if s.Empty() {
		return ansi.ResetStyle
	}

	var b ansi.Style

	if s.Attrs != 0 {
		if s.Attrs&BoldAttr != 0 {
			b = b.Bold()
		}
		if s.Attrs&FaintAttr != 0 {
			b = b.Faint()
		}
		if s.Attrs&ItalicAttr != 0 {
			b = b.Italic()
		}
		if s.Attrs&SlowBlinkAttr != 0 {
			b = b.SlowBlink()
		}
		if s.Attrs&RapidBlinkAttr != 0 {
			b = b.RapidBlink()
		}
		if s.Attrs&ReverseAttr != 0 {
			b = b.Reverse()
		}
		if s.Attrs&ConcealAttr != 0 {
			b = b.Conceal()
		}
		if s.Attrs&StrikethroughAttr != 0 {
			b = b.Strikethrough()
		}
	}
	if s.UlStyle != NoUnderline {
		switch s.UlStyle {
		case SingleUnderline:
			b = b.Underline()
		case DoubleUnderline:
			b = b.DoubleUnderline()
		case CurlyUnderline:
			b = b.CurlyUnderline()
		case DottedUnderline:
			b = b.DottedUnderline()
		case DashedUnderline:
			b = b.DashedUnderline()
		}
	}
	if s.Fg != nil {
		b = b.ForegroundColor(s.Fg)
	}
	if s.Bg != nil {
		b = b.BackgroundColor(s.Bg)
	}
	if s.Ul != nil {
		b = b.UnderlineColor(s.Ul)
	}

	return b.String()
}

// DiffSequence returns the ANSI sequence that sets the style as a diff from
// another style.
func (s Style) DiffSequence(o Style) string {
	if o.Empty() {
		return s.Sequence()
	}

	var b ansi.Style

	if !colorEqual(s.Fg, o.Fg) {
		b = b.ForegroundColor(s.Fg)
	}

	if !colorEqual(s.Bg, o.Bg) {
		b = b.BackgroundColor(s.Bg)
	}

	if !colorEqual(s.Ul, o.Ul) {
		b = b.UnderlineColor(s.Ul)
	}

	var (
		noBlink  bool
		isNormal bool
	)

	if s.Attrs != o.Attrs {
		if s.Attrs&BoldAttr != o.Attrs&BoldAttr {
			if s.Attrs&BoldAttr != 0 {
				b = b.Bold()
			} else if !isNormal {
				isNormal = true
				b = b.NormalIntensity()
			}
		}
		if s.Attrs&FaintAttr != o.Attrs&FaintAttr {
			if s.Attrs&FaintAttr != 0 {
				b = b.Faint()
			} else if !isNormal {
				isNormal = true
				b = b.NormalIntensity()
			}
		}
		if s.Attrs&ItalicAttr != o.Attrs&ItalicAttr {
			if s.Attrs&ItalicAttr != 0 {
				b = b.Italic()
			} else {
				b = b.NoItalic()
			}
		}
		if s.Attrs&SlowBlinkAttr != o.Attrs&SlowBlinkAttr {
			if s.Attrs&SlowBlinkAttr != 0 {
				b = b.SlowBlink()
			} else if !noBlink {
				b = b.NoBlink()
			}
		}
		if s.Attrs&RapidBlinkAttr != o.Attrs&RapidBlinkAttr {
			if s.Attrs&RapidBlinkAttr != 0 {
				b = b.RapidBlink()
			} else if !noBlink {
				b = b.NoBlink()
			}
		}
		if s.Attrs&ReverseAttr != o.Attrs&ReverseAttr {
			if s.Attrs&ReverseAttr != 0 {
				b = b.Reverse()
			} else {
				b = b.NoReverse()
			}
		}
		if s.Attrs&ConcealAttr != o.Attrs&ConcealAttr {
			if s.Attrs&ConcealAttr != 0 {
				b = b.Conceal()
			} else {
				b = b.NoConceal()
			}
		}
		if s.Attrs&StrikethroughAttr != o.Attrs&StrikethroughAttr {
			if s.Attrs&StrikethroughAttr != 0 {
				b = b.Strikethrough()
			} else {
				b = b.NoStrikethrough()
			}
		}
	}

	return b.String()
}

// Equal returns true if the style is equal to the other style.
func (s Style) Equal(o Style) bool {
	return colorEqual(s.Fg, o.Fg) &&
		colorEqual(s.Bg, o.Bg) &&
		colorEqual(s.Ul, o.Ul) &&
		s.Attrs == o.Attrs &&
		s.UlStyle == o.UlStyle
}

func colorEqual(c, o ansi.Color) bool {
	if c == nil && o == nil {
		return true
	}
	if c == nil || o == nil {
		return false
	}
	cr, cg, cb, ca := c.RGBA()
	or, og, ob, oa := o.RGBA()
	return cr == or && cg == og && cb == ob && ca == oa
}

// Bold sets the bold attribute.
func (s *Style) Bold(v bool) *Style {
	if v {
		s.Attrs |= BoldAttr
	} else {
		s.Attrs &^= BoldAttr
	}
	return s
}

// Faint sets the faint attribute.
func (s *Style) Faint(v bool) *Style {
	if v {
		s.Attrs |= FaintAttr
	} else {
		s.Attrs &^= FaintAttr
	}
	return s
}

// Italic sets the italic attribute.
func (s *Style) Italic(v bool) *Style {
	if v {
		s.Attrs |= ItalicAttr
	} else {
		s.Attrs &^= ItalicAttr
	}
	return s
}

// SlowBlink sets the slow blink attribute.
func (s *Style) SlowBlink(v bool) *Style {
	if v {
		s.Attrs |= SlowBlinkAttr
	} else {
		s.Attrs &^= SlowBlinkAttr
	}
	return s
}

// RapidBlink sets the rapid blink attribute.
func (s *Style) RapidBlink(v bool) *Style {
	if v {
		s.Attrs |= RapidBlinkAttr
	} else {
		s.Attrs &^= RapidBlinkAttr
	}
	return s
}

// Reverse sets the reverse attribute.
func (s *Style) Reverse(v bool) *Style {
	if v {
		s.Attrs |= ReverseAttr
	} else {
		s.Attrs &^= ReverseAttr
	}
	return s
}

// Conceal sets the conceal attribute.
func (s *Style) Conceal(v bool) *Style {
	if v {
		s.Attrs |= ConcealAttr
	} else {
		s.Attrs &^= ConcealAttr
	}
	return s
}

// Strikethrough sets the strikethrough attribute.
func (s *Style) Strikethrough(v bool) *Style {
	if v {
		s.Attrs |= StrikethroughAttr
	} else {
		s.Attrs &^= StrikethroughAttr
	}
	return s
}

// UnderlineStyle sets the underline style.
func (s *Style) UnderlineStyle(style UnderlineStyle) *Style {
	s.UlStyle = style
	return s
}

// Underline sets the underline attribute.
// This is a syntactic sugar for [UnderlineStyle].
func (s *Style) Underline(v bool) *Style {
	if v {
		return s.UnderlineStyle(SingleUnderline)
	}
	return s.UnderlineStyle(NoUnderline)
}

// Foreground sets the foreground color.
func (s *Style) Foreground(c ansi.Color) *Style {
	s.Fg = c
	return s
}

// Background sets the background color.
func (s *Style) Background(c ansi.Color) *Style {
	s.Bg = c
	return s
}

// UnderlineColor sets the underline color.
func (s *Style) UnderlineColor(c ansi.Color) *Style {
	s.Ul = c
	return s
}

// Reset resets the style to default.
func (s *Style) Reset() *Style {
	s.Fg = nil
	s.Bg = nil
	s.Ul = nil
	s.Attrs = ResetAttr
	s.UlStyle = NoUnderline
	return s
}

// Empty returns true if the style is empty.
func (s *Style) Empty() bool {
	return s.Fg == nil && s.Bg == nil && s.Ul == nil && s.Attrs == ResetAttr && s.UlStyle == NoUnderline
}

// Convert converts a style to respect the given color profile.
func (s Style) Convert(p colorprofile.Profile) Style {
	switch p {
	case colorprofile.TrueColor:
		return s
	case colorprofile.Ascii:
		s.Fg = nil
		s.Bg = nil
		s.Ul = nil
	case colorprofile.NoTTY:
		return Style{}
	}

	if s.Fg != nil {
		s.Fg = p.Convert(s.Fg)
	}
	if s.Bg != nil {
		s.Bg = p.Convert(s.Bg)
	}
	if s.Ul != nil {
		s.Ul = p.Convert(s.Ul)
	}

	return s
}
