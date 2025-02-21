package sixel

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"io"
	"os"
	"testing"

	"github.com/charmbracelet/x/ansi"
	gosixel "github.com/mattn/go-sixel"
)

func BenchmarkEncodingGoSixel(b *testing.B) {
	for i := 0; i < b.N; i++ {
		raw, err := loadImage("./../fixtures/graphics/JigokudaniMonkeyPark.png")
		if err != nil {
			os.Exit(1)
		}

		b := bytes.NewBuffer(nil)
		enc := gosixel.NewEncoder(b)
		if err := enc.Encode(raw); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		// fmt.Println(b)
	}
}

func writeSixelGraphics(w io.Writer, m image.Image) error {
	e := &Encoder{}

	data := bytes.NewBuffer(nil)
	if err := e.Encode(data, m); err != nil {
		return fmt.Errorf("failed to encode sixel image: %w", err)
	}

	_, err := io.WriteString(w, ansi.SixelGraphics(0, 1, 0, data.Bytes()))
	return err
}

func BenchmarkEncodingXSixel(b *testing.B) {
	for i := 0; i < b.N; i++ {
		raw, err := loadImage("./../fixtures/graphics/JigokudaniMonkeyPark.png")
		if err != nil {
			os.Exit(1)
		}

		b := bytes.NewBuffer(nil)
		if err := writeSixelGraphics(b, raw); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		// fmt.Println(b)
	}
}

func loadImage(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return png.Decode(f)
}
