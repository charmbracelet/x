package vttest

import (
	"image"
	"image/color"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/vt"
)

// Modes represents terminal modes.
type Modes struct {
	ANSIModes map[ansi.ANSIMode]ansi.ModeSetting `json:"ansi_modes" yaml:"ansi_modes"`
	DECModes  map[ansi.DECMode]ansi.ModeSetting  `json:"dec_modes" yaml:"dec_modes"`
}

// Cursor represents the cursor state.
type Cursor struct {
	Position image.Point    `json:"position," yaml:"position"`
	Visible  bool           `json:"visible" yaml:"visible"`
	Color    color.Color    `json:"color,omitempty" yaml:"color,omitempty"`
	Style    vt.CursorStyle `json:"style" yaml:"style"`
	Blink    bool           `json:"blink" yaml:"blink"`
}

// Style represents the Style of a cell.
type Style struct {
	Fg             color.Color       `json:"fg,omitempty" yaml:"fg,omitempty"`
	Bg             color.Color       `json:"bg,omitempty" yaml:"bg,omitempty"`
	UnderlineColor color.Color       `json:"underline_color,omitempty" yaml:"underline_color,omitempty"`
	Underline      uv.UnderlineStyle `json:"underline,omitempty" yaml:"underline,omitempty"`
	Attrs          uv.StyleAttr      `json:"attrs,omitempty" yaml:"attrs,omitempty"`
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
	Content string `json:"content" yaml:"content"`

	// The style of the cell. Nil style means no style. Zero value prints a
	// reset sequence.
	Style Style `json:"style" yaml:"style"`

	// Link is the hyperlink of the cell.
	Link Link `json:"link" yaml:"link"`

	// Width is the mono-spaced width of the grapheme cluster.
	Width int `json:"width" yaml:"width"`
}

// Snapshot represents a snapshot of the terminal state at a given moment.
type Snapshot struct {
	Modes     Modes       `json:"modes" yaml:"modes"`
	Title     string      `json:"title" yaml:"title"`
	Rows      int         `json:"rows" yaml:"rows"`
	Cols      int         `json:"cols" yaml:"cols"`
	AltScreen bool        `json:"alt_screen" yaml:"alt_screen"`
	Cursor    Cursor      `json:"cursor" yaml:"cursor"`
	BgColor   color.Color `json:"bg_color" yaml:"bg_color"`
	FgColor   color.Color `json:"fg_color" yaml:"fg_color"`
	Cells     [][]Cell    `json:"cells" yaml:"cells"`
}
