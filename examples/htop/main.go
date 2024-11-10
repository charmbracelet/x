package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"log"
	"os"
	"os/exec"
	"sync/atomic"
	"time"

	"github.com/charmbracelet/x/termios"
	"github.com/charmbracelet/x/vt"
	"github.com/creack/pty"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
	"golang.org/x/sys/unix"
)

const (
	fontWidth  = 7
	fontHeight = 13
)

func drawChar(img *image.RGBA, x, y int, fg color.Color, text string) {
	point := fixed.Point26_6{fixed.Int26_6(x * 64), fixed.Int26_6(y * 64)}
	if fg == nil {
		fg = color.White
	}
	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(fg),
		Face: basicfont.Face7x13,
		Dot:  point,
	}
	d.DrawString(text)
}

var counter int32 = 1

func putImage(vt *vt.Terminal) {
	rows, cols := vt.Height(), vt.Width()
	img := image.NewRGBA(image.Rect(0, 0, cols*fontWidth, rows*fontHeight))
	draw.Draw(img, img.Bounds(), &image.Uniform{color.Black}, image.ZP, draw.Src)

	for row := 0; row < rows; row++ {
		for col := 0; col < cols; col++ {
			cell, ok := vt.At(col, row)
			if !ok {
				log.Printf("failed to get at %d %d", row, col)
			}
			txt := cell.Content
			if len(txt) > 0 && txt[0] != 0 {
				fg, bg := cell.Style.Fg, cell.Style.Bg
				if bg == nil {
					bg = color.Black
				}

				// Draw background
				x0, y0 := (col+1)*fontWidth, ((row+1)*fontHeight)-11
				x1, y1 := x0+fontWidth, y0+fontHeight
				draw.Draw(img, image.Rect(x0, y0, x1, y1), &image.Uniform{bg}, image.ZP, draw.Over)

				drawChar(img, (col+1)*fontWidth, (row+1)*fontHeight, fg, txt)
			}
		}
	}

	f, err := os.Create(fmt.Sprintf("output%d.png", counter))
	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()
	err = png.Encode(f, img)
	if err != nil {
		log.Fatal(err)
	}

	atomic.AddInt32(&counter, 1)
}

const (
	width  = 165
	height = 25
)

func main() {
	vt := vt.NewTerminal(width, height)
	cmd := exec.Command("htop")

	go func() {
		for {
			time.Sleep(1 * time.Second)
			putImage(vt)
		}
	}()

	go func() {
		time.Sleep(5 * time.Second)
		cmd.Process.Kill()
	}()

	ptm, err := pty.Start(cmd)
	if err != nil {
		log.Fatal(err)
	}

	if err := termios.SetWinsize(int(ptm.Fd()), &unix.Winsize{Row: uint16(height), Col: uint16(width)}); err != nil {
		log.Fatal(err)
	}

	go io.Copy(ptm, vt)
	go io.Copy(vt, ptm)

	if err := cmd.Wait(); err != nil {
		log.Fatal(err)
	}
}
