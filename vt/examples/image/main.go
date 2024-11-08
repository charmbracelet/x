package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"os"

	"github.com/charmbracelet/x/vt"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

func drawChar(img *image.RGBA, x, y int, c color.Color, text string) {
	point := fixed.Point26_6{fixed.Int26_6(x * 64), fixed.Int26_6(y * 64)}
	if c == nil {
		c = color.White
	}
	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(c),
		Face: basicfont.Face7x13,
		Dot:  point,
	}
	d.DrawString(text)
}

func main() {
	vt := vt.NewTerminal(100, 25)

	for i := 0; i < 5; i++ {
		_, err := fmt.Fprintf(vt, "\033[32mHello \033[%dmGolang\033[0m\r\n", 32+i)
		if err != nil {
			log.Fatal(err)
		}
	}

	rows, cols := vt.Height(), vt.Width()
	img := image.NewRGBA(image.Rect(0, 0, cols*7, rows*13))
	draw.Draw(img, img.Bounds(), &image.Uniform{color.Black}, image.ZP, draw.Src)

	for row := 0; row < rows; row++ {
		for col := 0; col < cols; col++ {
			cell, ok := vt.At(col, row)
			if !ok {
				log.Printf("failed to get at %d %d", row, col)
			}
			txt := cell.Content
			// fmt.Println(cell.Style.Fg())
			if len(txt) > 0 && txt[0] != 0 {
				drawChar(img, (col+1)*7, (row+1)*13, cell.Style.Fg, txt)
			}
		}
	}
	f, err := os.Create("output.png")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	err = png.Encode(f, img)
	if err != nil {
		log.Fatal(err)
	}
}
