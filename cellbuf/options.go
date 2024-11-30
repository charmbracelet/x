package cellbuf

import (
	"bytes"

	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/wcwidth"
)

// Options are options for manipulating the buffer.
type Options struct {
	// Parser is the parser to use when writing to the buffer.
	Parser *ansi.Parser
	// Area is the area to write to.
	Area Rectangle
	// Profile is the profile to use when writing to the buffer.
	Profile colorprofile.Profile
	// Method is the width calculation method to use when writing to the buffer.
	Method Method
	// AutoWrap is whether to automatically wrap text when it reaches the end
	// of the line.
	AutoWrap bool
	// NewLine whether to automatically insert a carriage returns [ansi.CR]
	// when a linefeed [ansi.LF] is encountered.
	NewLine bool
}

// SetString sets the string at the given x, y position. It returns the new x
// and y position after writing the string. If the string is wider than the
// buffer and auto-wrap is enabled, it will wrap to the next line. Otherwise,
// it will be truncated.
func (opt Options) SetString(b *Buffer, x, y int, s string) (int, int) {
	if opt.Area.Empty() {
		opt.Area = b.Bounds()
	}

	p := opt.Parser
	if p == nil {
		p = ansi.GetParser()
		defer ansi.PutParser(p)
	}

	var pen Style
	var link Link
	var state byte

	handleCell := func(content string, width int) {
		var r rune
		var comb []rune
		for i, c := range content {
			if i == 0 {
				r = c
			} else {
				comb = append(comb, c)
			}
		}

		c := &Cell{
			Rune:  r,
			Comb:  comb,
			Width: width,
			Style: pen,
			Link:  link,
		}

		b.SetCell(x, y, c)
		if x+width >= opt.Area.Max.X && opt.AutoWrap {
			x = opt.Area.Min.X
			y++
		} else {
			x += width
		}
	}

	blankCell := func() *Cell {
		if pen.Bg != nil {
			return &Cell{
				Rune:  ' ',
				Width: 1,
				Style: pen,
			}
		}
		return nil
	}

	for len(s) > 0 {
		seq, width, n, newState := ansi.DecodeSequence(s, state, p)
		switch width {
		case 1, 2, 3, 4: // wide cells can go up to 4 cells wide
			switch opt.Method {
			case WcWidth:
				for _, r := range seq {
					width = wcwidth.RuneWidth(r)
					handleCell(string(r), width)
				}
			case GraphemeWidth:
				handleCell(seq, width)
			}
		case 0:
			switch {
			case ansi.HasCsiPrefix(seq):
				switch p.Cmd() {
				case 'm': // Select Graphic Rendition [ansi.SGR]
					handleSgr(p, &pen)
				case 'L': // Insert Line [ansi.IL]
					count := 1
					if n, ok := p.Param(0, 1); ok && n > 0 {
						count = n
					}

					b.InsertLine(y, count, blankCell())
				case 'M': // Delete Line [ansi.DL]
					count := 1
					if n, ok := p.Param(0, 1); ok && n > 0 {
						count = n
					}

					b.DeleteLine(y, count, blankCell())
				}
			case ansi.HasOscPrefix(seq):
				switch p.Cmd() {
				case 8: // Hyperlinks
					handleHyperlinks(p, &link)
				}
			case ansi.HasEscPrefix(seq):
				switch p.Cmd() {
				case 'M': // Reverse Index [ansi.RI]
					// Move the cursor up one line in the same column. If the
					// cursor is at the top margin, the screen performs a scroll-up.
					if y > opt.Area.Min.Y {
						y--
					}
				}
			case seq == "\n", seq == "\v", seq == "\f":
				if opt.NewLine {
					x = opt.Area.Min.X
				}
				y++
			case seq == "\r":
				x = opt.Area.Min.X
			}
		default:
			// Should never happen
			panic("invalid cell width")
		}

		s = s[n:]
		state = newState
	}

	return x, y
}

// Render renders the buffer to a string.
func (opt Options) Render(b *Buffer) string {
	var buf bytes.Buffer
	height := b.Height()
	for y := 0; y < height; y++ {
		_, line := renderLine(b, y, opt)
		buf.WriteString(line)
		if y < height-1 {
			buf.WriteString("\r\n")
		}
	}
	return buf.String()
}

// RenderLine renders a single line of the buffer to a string.
// It returns the width of the line and the rendered string.
func (opt Options) RenderLine(b *Buffer, y int) (int, string) {
	return renderLine(b, y, opt)
}

// Span represents a span of cells with the same style and link.
type Span struct {
	// Segment is the content of the span.
	Segment
	// Position is the starting position of the span.
	Position
}

// Diff computes the diff between two buffers as a slice of affected cells. It
// only returns affected cells withing the given rectangle.
func (opt Options) Diff(b, prev *Buffer) (diff []Span) {
	if prev == nil {
		return nil
	}

	area := opt.Area
	if area.Empty() {
		area = b.Bounds()
	}

	for y := area.Min.Y; y < area.Max.Y; y++ {
		var span *Span
		for x := area.Min.X; x < area.Max.X; x++ {
			cellA := b.Cell(x, y)
			cellB := prev.Cell(x, y)

			if cellA.Equal(cellB) {
				continue
			}

			if cellB == nil {
				cellB = &BlankCell
			}

			if span == nil {
				span = &Span{
					Position: Pos(x, y),
					Segment:  cellB.Segment(),
				}
				continue
			}

			if span.X+span.Width == x &&
				span.Style.Equal(cellB.Style) &&
				span.Link == cellB.Link {
				span.Content += cellB.Content()
				span.Width += cellB.Width
				continue
			}

			diff = append(diff, *span)
			span = &Span{
				Position: Pos(x, y),
				Segment:  cellB.Segment(),
			}
		}

		if span != nil {
			diff = append(diff, *span)
		}
	}

	return
}
