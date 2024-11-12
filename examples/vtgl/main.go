package main

import (
	"image/color"
	"io"
	"log"
	"math"
	"os/exec"
	"syscall"
	"time"

	"github.com/charmbracelet/x/cellbuf"
	"github.com/charmbracelet/x/vt"
	"github.com/creack/pty"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/fogleman/gg"
	"golang.org/x/image/font"
)

// mapKey converts pixelgl key to vt key
func mapKey(key pixelgl.Button) (k vt.Key, ok bool) {
	ok = true
	switch key {
	case pixelgl.KeyLeft:
		k = vt.Key{Code: vt.KeyLeft}
	case pixelgl.KeyRight:
		k = vt.Key{Code: vt.KeyRight}
	case pixelgl.KeyUp:
		k = vt.Key{Code: vt.KeyUp}
	case pixelgl.KeyDown:
		k = vt.Key{Code: vt.KeyDown}
	case pixelgl.KeyBackspace:
		k = vt.Key{Code: vt.KeyBackspace}
	case pixelgl.KeyDelete:
		k = vt.Key{Code: vt.KeyDelete}
	case pixelgl.KeyEnter:
		k = vt.Key{Code: vt.KeyEnter}
	case pixelgl.KeyEscape:
		k = vt.Key{Code: vt.KeyEscape}
	case pixelgl.KeyTab:
		k = vt.Key{Code: vt.KeyTab}
	default:
		ok = false
	}
	return
}

const (
	// width      = 165
	// height     = 45
	width      = 80
	height     = 24
	cellWidth  = 8
	cellHeight = 16
)

type Terminal struct {
	vt         *vt.Terminal
	context    *gg.Context
	sprite     *pixel.Sprite
	picture    pixel.Picture
	font       font.Face
	lastCursor vt.Cursor
	dirty      bool
}

func NewTerminal(width, height int) *Terminal {
	vterm := vt.NewTerminal(width, height)
	term := &Terminal{
		vt:      vterm,
		context: gg.NewContext(width*cellWidth, height*cellHeight),
	}

	// Load a monospace font
	// face, err := gg.LoadFontFace("./JetBrainsMono-Regular.ttf", 12)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// term.font = face
	// term.context.SetFontFace(face)

	term.vt.Damage = term.Damage
	term.context.SetColor(color.Black)
	term.context.Clear()
	return term
}

func (t *Terminal) Damage(d vt.Damage) {
	switch d := d.(type) {
	case vt.CellDamage:
		t.DrawAt(d.X, d.Y, d.Cell)
	case vt.ScreenDamage:
		t.context.SetColor(color.Black)
		t.context.Clear()
	case vt.RectDamage:
		t.context.SetColor(color.Black)
		rect := d.Bounds()
		t.context.DrawRectangle(float64(rect.Min.X*cellWidth), float64(rect.Min.Y*cellHeight),
			float64(rect.Width()*cellWidth), float64(rect.Height()*cellHeight))
		t.context.Fill()
	}

	t.dirty = true
}

func (t *Terminal) DrawAt(x, y int, cell cellbuf.Cell) {
	// Convert terminal coordinates to pixel coordinates
	px := float64(x * cellWidth)
	py := float64(y * cellHeight)

	// Draw background
	bg := cell.Style.Bg
	if bg == nil {
		bg = t.vt.BackgroundColor()
	}
	t.context.SetColor(bg)
	t.context.DrawRectangle(px, py, cellWidth, cellHeight)
	t.context.Fill()

	// Draw text
	if len(cell.Content) != 0 {
		fg := cell.Style.Fg
		if fg == nil {
			fg = t.vt.ForegroundColor()
		}
		t.context.SetColor(fg)
		t.context.DrawString(cell.Content, px, py+cellHeight-4) // Adjust Y for baseline
	}

	// Handle underline
	if cell.Style.UlStyle > cellbuf.NoUnderline {
		ul := cell.Style.Ul
		if ul == nil {
			ul = t.vt.ForegroundColor()
		}
		t.context.SetColor(ul)
		t.context.DrawLine(px, py+cellHeight-2, px+cellWidth, py+cellHeight-2)
		t.context.Stroke()
	}
}

func (t *Terminal) Draw() {
	// Only update sprite if terminal is dirty
	if !t.dirty {
		return
	}

	// Get current cursor
	cursor := t.vt.Cursor()

	if false && cursor != t.lastCursor {
		// FIXME: This causes a crash

		// Restore the previous cursor cell if it was visible
		if !t.lastCursor.Hidden {
			if cell, ok := t.vt.At(t.lastCursor.X, t.lastCursor.Y); ok {
				t.DrawAt(t.lastCursor.X, t.lastCursor.Y, cell)
			}
		}

		// Draw new cursor
		if !cursor.Hidden {
			px := float64(cursor.X * cellWidth)
			py := float64(cursor.Y * cellHeight)

			// Get the cell at cursor position
			cell, _ := t.vt.At(cursor.X, cursor.Y)

			// Draw cursor background
			t.context.SetColor(t.vt.CursorColor())
			t.context.DrawRectangle(px, py, cellWidth, cellHeight)
			t.context.Fill()

			// Draw cursor text in inverted color
			if len(cell.Content) > 0 {
				// TODO: Invert [CursorColor].
				t.context.SetColor(color.Black)
				t.context.DrawString(cell.Content, px, py+cellHeight-4)
			}
		}

		// Store current cursor for next frame
		t.lastCursor = cursor
	}

	img := t.context.Image()
	pic := pixel.PictureDataFromImage(img)
	t.picture = pic
	t.sprite = pixel.NewSprite(t.picture, t.picture.Bounds())
	t.dirty = false
}

func run() {
	cfg := pixelgl.WindowConfig{
		Title:     "Terminal",
		Bounds:    pixel.R(0, 0, float64(width*cellWidth), float64(height*cellHeight)),
		VSync:     true,
		Resizable: true,
	}
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		log.Fatal(err)
	}

	term := NewTerminal(width, height)
	lastBounds := win.Bounds()
	// cmd := exec.Command("nvim")
	cmd := exec.Command("htop")
	// cmd := exec.Command("ssh", "git.charm.sh")
	// cmd := exec.Command("zsh", "-i", "-l")

	attrs := syscall.SysProcAttr{
		Setsid:  true,
		Setctty: true,
	}
	ptm, err := pty.StartWithAttrs(cmd, &pty.Winsize{Rows: height, Cols: width}, &attrs)
	if err != nil {
		log.Fatal(err)
	}

	// TODO: Handle window resize
	go io.Copy(ptm, term.vt)
	go io.Copy(term.vt, ptm)

	last := time.Now()
	for !win.Closed() {
		dt := time.Since(last).Seconds()
		last = time.Now()

		// Handle special keys
		for key := pixelgl.KeyWorld1; key < pixelgl.KeyLast; key++ {
			if win.JustPressed(key) {
				if vtKey, ok := mapKey(key); ok {
					term.vt.SendKey(vtKey)
				} else {
					log.Printf("unhandled key: %v", key)
				}
			}
		}

		// Handle typed characters
		if win.Typed() != "" {
			term.vt.SendText(win.Typed())
		}

		// Handle mouse input
		mousePos := win.MousePosition()
		// Convert window coordinates to cell coordinates
		cellX := int(math.Floor(mousePos.X / cellWidth))
		cellY := height - int(math.Floor(mousePos.Y/cellHeight)) - 1

		// Handle mouse buttons
		if win.JustPressed(pixelgl.MouseButtonLeft) {
			term.vt.SendMouse(vt.MouseClick{
				Button: vt.MouseLeft,
				X:      cellX,
				Y:      cellY,
			})
		}
		if win.JustReleased(pixelgl.MouseButtonLeft) {
			term.vt.SendMouse(vt.MouseRelease{
				Button: vt.MouseLeft,
				X:      cellX,
				Y:      cellY,
			})
		}

		// Handle scroll wheel
		scroll := win.MouseScroll()
		if scroll.Y != 0 {
			button := vt.MouseWheelUp
			if scroll.Y < 0 {
				button = vt.MouseWheelDown
			}
			term.vt.SendMouse(vt.MouseWheel{
				Button: button,
				X:      cellX,
				Y:      cellY,
			})
		}

		term.Draw()
		win.Clear(color.Black)
		if term.sprite != nil {
			term.sprite.Draw(win, pixel.IM.Moved(win.Bounds().Center()))
		}
		win.Update()

		// Handle window resize
		bounds := win.Bounds()
		if bounds != lastBounds {
			newWidth := int(bounds.W() / cellWidth)
			newHeight := int(bounds.H() / cellHeight)

			// Resize the terminal
			term.vt.Resize(newWidth, newHeight)

			// Create new context with new size
			term.context = gg.NewContext(newWidth*cellWidth, newHeight*cellHeight)
			if term.font != nil {
				term.context.SetFontFace(term.font)
			}

			// Resize the pty
			pty.Setsize(ptm, &pty.Winsize{
				Rows: uint16(newHeight),
				Cols: uint16(newWidth),
			})

			lastBounds = bounds
			term.dirty = true
		}

		time.Sleep(time.Second/time.Duration(60) - time.Duration(dt*float64(time.Second)))
	}

	cmd.Process.Kill()
}

func main() {
	pixelgl.Run(run)
}
