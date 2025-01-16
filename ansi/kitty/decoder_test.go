package kitty

import (
	"bytes"
	"compress/zlib"
	"image"
	"image/color"
	"image/png"
	"reflect"
	"testing"
)

func TestDecoder_Decode(t *testing.T) {
	// Helper function to create compressed data
	compress := func(data []byte) []byte {
		var buf bytes.Buffer
		w := zlib.NewWriter(&buf)
		w.Write(data)
		w.Close()
		return buf.Bytes()
	}

	tests := []struct {
		name    string
		decoder Decoder
		input   []byte
		want    image.Image
		wantErr bool
	}{
		{
			name: "RGBA format 2x2",
			decoder: Decoder{
				Format: RGBA,
				Width:  2,
				Height: 2,
			},
			input: []byte{
				255, 0, 0, 255, // Red pixel
				0, 0, 255, 255, // Blue pixel
				0, 0, 255, 255, // Blue pixel
				255, 0, 0, 255, // Red pixel
			},
			want: func() image.Image {
				img := image.NewRGBA(image.Rect(0, 0, 2, 2))
				img.Set(0, 0, color.RGBA{R: 255, G: 0, B: 0, A: 255})
				img.Set(1, 0, color.RGBA{R: 0, G: 0, B: 255, A: 255})
				img.Set(0, 1, color.RGBA{R: 0, G: 0, B: 255, A: 255})
				img.Set(1, 1, color.RGBA{R: 255, G: 0, B: 0, A: 255})
				return img
			}(),
			wantErr: false,
		},
		{
			name: "RGB format 2x2",
			decoder: Decoder{
				Format: RGB,
				Width:  2,
				Height: 2,
			},
			input: []byte{
				255, 0, 0, // Red pixel
				0, 0, 255, // Blue pixel
				0, 0, 255, // Blue pixel
				255, 0, 0, // Red pixel
			},
			want: func() image.Image {
				img := image.NewRGBA(image.Rect(0, 0, 2, 2))
				img.Set(0, 0, color.RGBA{R: 255, G: 0, B: 0, A: 255})
				img.Set(1, 0, color.RGBA{R: 0, G: 0, B: 255, A: 255})
				img.Set(0, 1, color.RGBA{R: 0, G: 0, B: 255, A: 255})
				img.Set(1, 1, color.RGBA{R: 255, G: 0, B: 0, A: 255})
				return img
			}(),
			wantErr: false,
		},
		{
			name: "RGBA with compression",
			decoder: Decoder{
				Format:     RGBA,
				Width:      2,
				Height:     2,
				Decompress: true,
			},
			input: compress([]byte{
				255, 0, 0, 255,
				0, 0, 255, 255,
				0, 0, 255, 255,
				255, 0, 0, 255,
			}),
			want: func() image.Image {
				img := image.NewRGBA(image.Rect(0, 0, 2, 2))
				img.Set(0, 0, color.RGBA{R: 255, G: 0, B: 0, A: 255})
				img.Set(1, 0, color.RGBA{R: 0, G: 0, B: 255, A: 255})
				img.Set(0, 1, color.RGBA{R: 0, G: 0, B: 255, A: 255})
				img.Set(1, 1, color.RGBA{R: 255, G: 0, B: 0, A: 255})
				return img
			}(),
			wantErr: false,
		},
		{
			name: "PNG format",
			decoder: Decoder{
				Format: PNG,
				// Width and height are embedded and inferred from the PNG data
			},
			input: func() []byte {
				img := image.NewRGBA(image.Rect(0, 0, 1, 1))
				img.Set(0, 0, color.RGBA{R: 255, G: 0, B: 0, A: 255})
				var buf bytes.Buffer
				png.Encode(&buf, img)
				return buf.Bytes()
			}(),
			want: func() image.Image {
				img := image.NewRGBA(image.Rect(0, 0, 1, 1))
				img.Set(0, 0, color.RGBA{R: 255, G: 0, B: 0, A: 255})
				return img
			}(),
			wantErr: false,
		},
		{
			name: "invalid format",
			decoder: Decoder{
				Format: 999,
				Width:  2,
				Height: 2,
			},
			input:   []byte{0, 0, 0},
			want:    nil,
			wantErr: true,
		},
		{
			name: "incomplete RGBA data",
			decoder: Decoder{
				Format: RGBA,
				Width:  2,
				Height: 2,
			},
			input:   []byte{255, 0, 0}, // Incomplete pixel data
			want:    nil,
			wantErr: true,
		},
		{
			name: "invalid compressed data",
			decoder: Decoder{
				Format:     RGBA,
				Width:      2,
				Height:     2,
				Decompress: true,
			},
			input:   []byte{1, 2, 3}, // Invalid zlib data
			want:    nil,
			wantErr: true,
		},
		{
			name: "default format (RGBA)",
			decoder: Decoder{
				Width:  1,
				Height: 1,
			},
			input: []byte{255, 0, 0, 255},
			want: func() image.Image {
				img := image.NewRGBA(image.Rect(0, 0, 1, 1))
				img.Set(0, 0, color.RGBA{R: 255, G: 0, B: 0, A: 255})
				return img
			}(),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.decoder.Decode(bytes.NewReader(tt.input))

			if (err != nil) != tt.wantErr {
				t.Errorf("Decode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Decode() output mismatch")
				if bounds := got.Bounds(); bounds != tt.want.Bounds() {
					t.Errorf("bounds got %v, want %v", bounds, tt.want.Bounds())
				}

				// Compare pixels
				bounds := got.Bounds()
				for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
					for x := bounds.Min.X; x < bounds.Max.X; x++ {
						gotColor := got.At(x, y)
						wantColor := tt.want.At(x, y)
						if !reflect.DeepEqual(gotColor, wantColor) {
							t.Errorf("pixel at (%d,%d) = %v, want %v", x, y, gotColor, wantColor)
						}
					}
				}
			}
		})
	}
}

func TestDecoder_DecodeEdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		decoder Decoder
		input   []byte
		wantErr bool
	}{
		{
			name: "zero dimensions",
			decoder: Decoder{
				Format: RGBA,
				Width:  0,
				Height: 0,
			},
			input:   []byte{},
			wantErr: false,
		},
		{
			name: "negative width",
			decoder: Decoder{
				Format: RGBA,
				Width:  -1,
				Height: 1,
			},
			input:   []byte{255, 0, 0, 255},
			wantErr: false, // The image package handles this gracefully
		},
		{
			name: "very large dimensions",
			decoder: Decoder{
				Format: RGBA,
				Width:  1,
				Height: 1000000, // Very large height
			},
			input:   []byte{255, 0, 0, 255}, // Not enough data
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.decoder.Decode(bytes.NewReader(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("Decode() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
