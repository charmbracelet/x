package vttest

import (
	"fmt"
	"image/color"
	"strconv"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/vt"
)

// Modes represents terminal modes.
type Modes struct {
	ANSI map[ansi.ANSIMode]ansi.ModeSetting `json:"ansi" yaml:"ansi"`
	DEC  map[ansi.DECMode]ansi.ModeSetting  `json:"dec" yaml:"dec"`
}

// Position represents a position in the terminal.
type Position struct {
	X int `json:"x" yaml:"x"`
	Y int `json:"y" yaml:"y"`
}

// Color represents a terminal color, which can be one of the following:
// - An ANSI 16 color (0-15) of type [ansi.BasicColor].
// - An ANSI 256 color (0-255) of type [ansi.IndexedColor].
// - Or any other 24-bit color that implements [color.Color].
type Color struct {
	Color color.Color `json:"color,omitempty" yaml:"color,omitempty"`
}

// MarshalText implements the [encoding.TextMarshaler] interface for Color.
func (c Color) MarshalText() ([]byte, error) {
	switch col := c.Color.(type) {
	case nil:
		return []byte{}, nil
	case ansi.BasicColor:
		return []byte(strconv.Itoa(int(col))), nil
	case ansi.IndexedColor:
		return []byte(strconv.Itoa(int(col))), nil
	default:
		r, g, b, _ := c.Color.RGBA()
		return fmt.Appendf(nil, "#%02x%02x%02x", r>>8, g>>8, b>>8), nil
	}
}

// UnmarshalText implements the [encoding.TextUnmarshaler] interface for Color.
func (c *Color) UnmarshalText(text []byte) error {
	s := string(text)
	if s == "" {
		return nil
	}
	if i, err := strconv.Atoi(s); err == nil {
		if i >= 0 && i <= 15 {
			c.Color = ansi.BasicColor(i)
			return nil
		} else if i >= 16 && i <= 255 {
			c.Color = ansi.IndexedColor(i)
			return nil
		}
	}

	col := ansi.XParseColor(s)
	if col == nil {
		return fmt.Errorf("invalid color: %s", s)
	}
	c.Color = col
	return nil
}

// Cursor represents the cursor state.
type Cursor struct {
	Position Position       `json:"position," yaml:"position"`
	Visible  bool           `json:"visible" yaml:"visible"`
	Color    Color          `json:"color,omitzero" yaml:"color,omitzero"`
	Style    vt.CursorStyle `json:"style" yaml:"style"`
	Blink    bool           `json:"blink" yaml:"blink"`
}

// Style represents the Style of a cell.
type Style struct {
	Fg             Color        `json:"fg,omitzero" yaml:"fg,omitzero"`
	Bg             Color        `json:"bg,omitzero" yaml:"bg,omitzero"`
	UnderlineColor Color        `json:"underline_color,omitzero" yaml:"underline_color,omitzero"`
	Underline      uv.Underline `json:"underline,omitempty" yaml:"underline,omitempty"`
	Attrs          byte         `json:"attrs,omitempty" yaml:"attrs,omitempty"`
}

// Link represents a hyperlink in the terminal screen.
type Link struct {
	URL    string `json:"url,omitempty" yaml:"url,omitempty"`
	Params string `json:"params,omitempty" yaml:"params,omitempty"`
}

// Cell represents a single cell in the terminal screen.
type Cell struct {
	// Content is the [Cell]'s content, which consists of a single grapheme
	// cluster. Most of the time, this will be a single rune as well, but it
	// can also be a combination of runes that form a grapheme cluster.
	Content string `json:"content,omitempty" yaml:"content,omitempty"`

	// The style of the cell. Nil style means no style. Zero value prints a
	// reset sequence.
	Style Style `json:"style,omitzero" yaml:"style,omitzero"`

	// Link is the hyperlink of the cell.
	Link Link `json:"link,omitzero" yaml:"link,omitzero"`

	// Width is the mono-spaced width of the grapheme cluster.
	Width int `json:"width,omitzero" yaml:"width,omitzero"`
}

// Snapshot represents a snapshot of the terminal state at a given moment.
type Snapshot struct {
	Modes     Modes    `json:"modes" yaml:"modes"`
	Title     string   `json:"title" yaml:"title"`
	Rows      int      `json:"rows" yaml:"rows"`
	Cols      int      `json:"cols" yaml:"cols"`
	AltScreen bool     `json:"alt_screen" yaml:"alt_screen"`
	Cursor    Cursor   `json:"cursor" yaml:"cursor"`
	BgColor   Color    `json:"bg_color,omitzero" yaml:"bg_color,omitzero"`
	FgColor   Color    `json:"fg_color,omitzero" yaml:"fg_color,omitzero"`
	Cells     [][]Cell `json:"cells" yaml:"cells"`
}
