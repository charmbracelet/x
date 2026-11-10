package kitty

import (
	"bytes"
	"compress/zlib"
	"errors"
	"image"
	"image/color"
	"image/png"
	"io"
	"math/rand/v2"
	"os"
	"testing"
)

// Exercise nonzero bounds, a padded stride, and all alpha values.
func pngTestImage() *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, 140, 100)).SubImage(image.Rect(7, 9, 131, 93)).(*image.NRGBA)
	random := rand.New(rand.NewPCG(1, 2))
	for y := img.Rect.Min.Y; y < img.Rect.Max.Y; y++ {
		for x := img.Rect.Min.X; x < img.Rect.Max.X; x++ {
			img.SetNRGBA(x, y, color.NRGBA{uint8(random.Uint32()), uint8(random.Uint32()), uint8(random.Uint32()), uint8(x*3 + y)})
		}
	}
	return img
}

func expectedPNG(t *testing.T, img image.Image, level png.CompressionLevel) []byte {
	t.Helper()
	var out bytes.Buffer
	var err error
	if level == png.DefaultCompression {
		// Compare against the original png.Encode path, not just our new encoder.
		err = png.Encode(&out, img)
	} else {
		enc := png.Encoder{CompressionLevel: level}
		err = enc.Encode(&out, img)
	}
	if err != nil {
		t.Fatal(err)
	}
	return out.Bytes()
}

func assertPNG(t *testing.T, data []byte, img *image.NRGBA, level png.CompressionLevel) {
	t.Helper()
	if !bytes.Equal(data, expectedPNG(t, img, level)) {
		t.Fatal("PNG differs from requested standard-library compression level")
	}
	decoded, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	if decoded.Bounds().Size() != img.Bounds().Size() {
		t.Fatal("PNG dimensions changed")
	}
	for y := 0; y < img.Rect.Dy(); y++ {
		for x := 0; x < img.Rect.Dx(); x++ {
			got := color.NRGBAModel.Convert(decoded.At(x, y)).(color.NRGBA)
			want := img.NRGBAAt(x+img.Rect.Min.X, y+img.Rect.Min.Y)
			if got != want {
				t.Fatalf("pixel %d,%d: got %v, want %v", x, y, got, want)
			}
		}
	}
}

func TestEncoderPNGCompressionLevel(t *testing.T) {
	img := pngTestImage()
	tests := []struct {
		name  string
		level png.CompressionLevel
	}{
		{"default", png.DefaultCompression},
		{"best speed", png.BestSpeed},
		{"best compression", png.BestCompression},
		{"no compression", png.NoCompression},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			enc := Encoder{Format: PNG, PNGCompressionLevel: tt.level}
			if err := enc.Encode(&out, img); err != nil {
				t.Fatal(err)
			}
			assertPNG(t, out.Bytes(), img, tt.level)
		})
	}
}

func TestEncoderPNGWithZlib(t *testing.T) {
	img := pngTestImage()
	var out bytes.Buffer
	enc := Encoder{Format: PNG, PNGCompressionLevel: png.BestSpeed, Compress: true}
	if err := enc.Encode(&out, img); err != nil {
		t.Fatal(err)
	}
	zr, err := zlib.NewReader(&out)
	if err != nil {
		t.Fatal(err)
	}
	defer zr.Close()
	data, err := io.ReadAll(zr)
	if err != nil {
		t.Fatal(err)
	}
	assertPNG(t, data, img, png.BestSpeed)
}

type pngErrorWriter struct {
	err error
}

func (w pngErrorWriter) Write([]byte) (int, error) {
	return 0, w.err
}

func TestEncoderPNGWriterError(t *testing.T) {
	sentinel := errors.New("PNG write failed")
	enc := Encoder{Format: PNG, PNGCompressionLevel: png.BestSpeed}
	if err := enc.Encode(pngErrorWriter{sentinel}, testImage()); !errors.Is(err, sentinel) {
		t.Fatalf("got error %v, want wrapped writer error %v", err, sentinel)
	}
}

func TestEncodeGraphicsPNGInvalidDimensions(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 0, 1))
	var out bytes.Buffer
	opts := &Options{Transmission: Direct, Format: PNG, PNGCompressionLevel: png.BestSpeed}
	err := EncodeGraphics(&out, img, opts)
	var formatErr png.FormatError
	if !errors.As(err, &formatErr) {
		t.Fatalf("got error %v, want wrapped PNG format error", err)
	}
	if out.Len() != 0 {
		t.Fatal("image encoding failure must not write protocol output")
	}
}

func TestPNGCompressionLevelIgnoredForRawFormats(t *testing.T) {
	for _, format := range []int{RGB, RGBA} {
		var want, got bytes.Buffer
		enc := Encoder{Format: format}
		if err := enc.Encode(&want, testImage()); err != nil {
			t.Fatal(err)
		}
		enc.PNGCompressionLevel = png.BestSpeed
		if err := enc.Encode(&got, testImage()); err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(want.Bytes(), got.Bytes()) {
			t.Fatal("PNG level affected raw pixels")
		}
	}
}

func TestPNGCompressionLevelNotSerialized(t *testing.T) {
	opts := Options{Format: PNG, Compression: Zlib}
	want := opts.String()
	opts.PNGCompressionLevel = png.BestSpeed
	if got := opts.String(); got != want {
		t.Fatalf("local PNG level changed controls: %q != %q", got, want)
	}
}

func TestEncodeGraphicsPNGCompressionLevel(t *testing.T) {
	img := pngTestImage()
	tests := []struct {
		name         string
		level        png.CompressionLevel
		transmission byte
	}{
		{"default", png.DefaultCompression, Direct},
		{"best speed", png.BestSpeed, Direct},
		{"best speed temporary file", png.BestSpeed, TempFile},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			opts := &Options{Transmission: tt.transmission, Format: PNG, PNGCompressionLevel: tt.level}
			if err := EncodeGraphics(&out, img, opts); err != nil {
				t.Fatal(err)
			}
			data := graphicsPayload(t, out.String())
			if tt.transmission == TempFile {
				path := string(data)
				defer os.Remove(path)
				var err error
				data, err = os.ReadFile(path)
				if err != nil {
					t.Fatal(err)
				}
			}
			if !bytes.Equal(data, expectedPNG(t, img, tt.level)) {
				t.Fatal("EncodeGraphics did not use the selected PNG compression")
			}
		})
	}
}
