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

// DefaultImageOptions are the default options used for creating terminal
// screen images.
var DefaultImageOptions = func() *ImageOptions {
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

	return &ImageOptions{
		CellWidth:      cellW,
		CellHeight:     cellH,
		RegularFace:    regularFace,
		BoldFace:       boldFace,
		ItalicFace:     italicFace,
		BoldItalicFace: boldItalicFace,
	}
}()

// ImageOptions contains options for drawing a terminal emulator screen to an
// image.
type ImageOptions struct {
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

// Image return s an image of the terminal emulator screen.
// If d is nil, [DefaultImageOptions] is used.
func (t *Terminal) Image(opts ...*ImageOptions) image.Image {
	t.mu.Lock()
	defer t.mu.Unlock()

	var opt *ImageOptions
	if len(opts) > 0 {
		opt = opts[0]
	}
	if opt == nil {
		opt = DefaultImageOptions
	}
	if opt.CellWidth <= 0 {
		opt.CellWidth = DefaultImageOptions.CellWidth
	}
	if opt.CellHeight <= 0 {
		opt.CellHeight = DefaultImageOptions.CellHeight
	}
	if opt.RegularFace == nil {
		opt.RegularFace = DefaultImageOptions.RegularFace
	}
	if opt.BoldFace == nil {
		opt.BoldFace = DefaultImageOptions.BoldFace
	}
	if opt.ItalicFace == nil {
		opt.ItalicFace = DefaultImageOptions.ItalicFace
	}
	if opt.BoldItalicFace == nil {
		opt.BoldItalicFace = DefaultImageOptions.BoldItalicFace
	}

	area := t.Bounds()
	width, height := area.Dx(), area.Dy()
	r := image.Rect(0, 0, width*opt.CellWidth, height*opt.CellHeight)
	img := image.NewRGBA(r)

	// Fill background
	bg := t.BackgroundColor()
	if bg == nil {
		bg = color.Black
	}
	draw.Draw(img, img.Bounds(), &image.Uniform{C: bg}, image.Point{}, draw.Src)

	// Draw cells
	drawCell := func(x, y int, cell *uv.Cell) {
		px := x * opt.CellWidth
		py := y * opt.CellHeight
		dot := fixed.P(px, py+opt.CellHeight-4) // 4 pixels from bottom for baseline
		style := cell.Style
		attrs := style.Attrs
		fg := style.Fg
		if fg == nil {
			fg = color.White
		}
		face := opt.RegularFace
		if attrs&uv.AttrBold != 0 && attrs&uv.AttrItalic != 0 {
			face = opt.BoldItalicFace
		} else if attrs&uv.AttrBold != 0 {
			face = opt.BoldFace
		} else if attrs&uv.AttrItalic != 0 {
			face = opt.ItalicFace
		}

		drawer := &font.Drawer{
			Dst:  img,
			Src:  image.NewUniform(fg),
			Face: face,
			Dot:  dot,
		}
		drawer.DrawString(cell.Content)

		// Handle underline
		//nolint:godox
		// TODO: Implement more underline styles
		// For now, we only support single underline
		if cell.Style.Underline > uv.UnderlineNone {
			col := cell.Style.UnderlineColor
			if col == nil {
				col = fg
			}
			for i := range opt.CellWidth {
				img.Set(px+i, py+opt.CellHeight-2, col)
			}
		}
	}

	// Iterate over screen cells
	for y := 0; y < height; y++ {
		for x := 0; x < width; {
			cell := t.CellAt(x, y)
			if cell == nil {
				cell = &uv.EmptyCell
			}
			drawCell(x, y, cell)
			x += cell.Width
		}
	}

	return img
}
