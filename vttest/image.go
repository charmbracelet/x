package vttest

import (
	"image"
	"image/color"
	"image/draw"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gomono"
	"golang.org/x/image/font/gofont/gomonobold"
	"golang.org/x/image/font/gofont/gomonobolditalic"
	"golang.org/x/image/font/gofont/gomonoitalic"
	"golang.org/x/image/math/fixed"
)

// DefaultDrawer are the default options used for creating terminal
// screen images.
var DefaultDrawer = func() *Drawer {
	cellW, cellH := 10, 16
	regular, _ := freetype.ParseFont(gomono.TTF)
	bold, _ := freetype.ParseFont(gomonobold.TTF)
	italic, _ := freetype.ParseFont(gomonoitalic.TTF)
	boldItalic, _ := freetype.ParseFont(gomonobolditalic.TTF)
	faceOpts := &truetype.Options{
		Size:    14, // Font size -2 to account for padding
		DPI:     72,
		Hinting: font.HintingFull,
	}
	regularFace := truetype.NewFace(regular, faceOpts)
	boldFace := truetype.NewFace(bold, faceOpts)
	italicFace := truetype.NewFace(italic, faceOpts)
	boldItalicFace := truetype.NewFace(boldItalic, faceOpts)

	return &Drawer{
		CellWidth:      cellW,
		CellHeight:     cellH,
		RegularFace:    regularFace,
		BoldFace:       boldFace,
		ItalicFace:     italicFace,
		BoldItalicFace: boldItalicFace,
	}
}()

// Draw draws a terminal emulator screen to an image.
func Draw(e Emulator) image.Image {
	return DefaultDrawer.Draw(e)
}

// Emulator represents a terminal emulator interface.
type Emulator interface {
	uv.Screen
	BackgroundColor() color.Color
}

// Drawer contains options for drawing a terminal emulator screen to an image.
type Drawer struct {
	// CellWidth is the width of each cell in pixels. Default is 10.
	CellWidth int
	// CellHeight is the height of each cell in pixels. Default is 18.
	CellHeight int
	// RegularFace is the font face to use for regular text. Default is Go
	// mono.
	RegularFace font.Face
	// BoldFace is the font face to use for bold text. If nil, Go mono bold is
	// used.
	BoldFace font.Face
	// ItalicFace is the font face to use for italic text. If nil, Go mono italic
	// is used.
	ItalicFace font.Face
	// BoldItalicFace is the font face to use for bold italic text. If nil, Go
	// mono bold italic is used.
	BoldItalicFace font.Face
}

// Draw draws a terminal emulator screen to an image using the drawer's
// options.
func (d *Drawer) Draw(e Emulator) image.Image {
	if d == nil {
		d = DefaultDrawer
	}
	if d.CellWidth <= 0 {
		d.CellWidth = DefaultDrawer.CellWidth
	}
	if d.CellHeight <= 0 {
		d.CellHeight = DefaultDrawer.CellHeight
	}
	if d.RegularFace == nil {
		d.RegularFace = DefaultDrawer.RegularFace
	}
	if d.BoldFace == nil {
		d.BoldFace = DefaultDrawer.BoldFace
	}
	if d.ItalicFace == nil {
		d.ItalicFace = DefaultDrawer.ItalicFace
	}
	if d.BoldItalicFace == nil {
		d.BoldItalicFace = DefaultDrawer.BoldItalicFace
	}

	area := e.Bounds()
	width, height := area.Dx(), area.Dy()
	r := image.Rect(0, 0, width*d.CellWidth, height*d.CellHeight)
	img := image.NewRGBA(r)

	// Fill background
	bg := e.BackgroundColor()
	if bg == nil {
		bg = color.Black
	}
	draw.Draw(img, img.Bounds(), &image.Uniform{C: bg}, image.Point{}, draw.Src)

	// Draw cells
	drawCell := func(x, y int, cell *uv.Cell) {
		px := x * d.CellWidth
		py := y * d.CellHeight
		dot := fixed.P(px, py+d.CellHeight-4) // 4 pixels from bottom for baseline
		style := cell.Style
		attrs := style.Attrs
		fg := style.Fg
		if fg == nil {
			fg = color.White
		}
		face := d.RegularFace
		if attrs&uv.AttrBold != 0 && attrs&uv.AttrItalic != 0 {
			face = d.BoldItalicFace
		} else if attrs&uv.AttrBold != 0 {
			face = d.BoldFace
		} else if attrs&uv.AttrItalic != 0 {
			face = d.ItalicFace
		}

		drawer := &font.Drawer{
			Dst:  img,
			Src:  image.NewUniform(fg),
			Face: face,
			Dot:  dot,
		}
		drawer.DrawString(cell.Content)

		// Handle underline
		// TODO: Implement more underline styles
		// For now, we only support single underline
		if cell.Style.Underline > uv.UnderlineStyleNone {
			col := cell.Style.UnderlineColor
			if col == nil {
				col = fg
			}
			for i := range d.CellWidth {
				img.Set(px+i, py+d.CellHeight-2, col)
			}
		}
	}

	// Iterate over screen cells
	for y := 0; y < height; y++ {
		for x := 0; x < width; {
			cell := e.CellAt(x, y)
			if cell == nil {
				cell = &uv.EmptyCell
			}
			drawCell(x, y, cell)
			x += cell.Width
		}
	}

	return img
}
