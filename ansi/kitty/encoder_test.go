package kitty

import (
	"bytes"
	"compress/zlib"
	"image"
	"image/color"
	"io"
	"testing"
)

// taken from "image/png" package
const pngHeader = "\x89PNG\r\n\x1a\n"

// testImage creates a simple test image with a red and blue pattern
func testImage() *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	img.Set(0, 0, color.RGBA{R: 255, G: 0, B: 0, A: 255}) // Red
	img.Set(1, 0, color.RGBA{R: 0, G: 0, B: 255, A: 255}) // Blue
	img.Set(0, 1, color.RGBA{R: 0, G: 0, B: 255, A: 255}) // Blue
	img.Set(1, 1, color.RGBA{R: 255, G: 0, B: 0, A: 255}) // Red
	return img
}

func TestEncoder_Encode(t *testing.T) {
	tests := []struct {
		name    string
		encoder Encoder
		img     image.Image
		wantErr bool
		verify  func([]byte) error
	}{
		{
			name: "nil image",
			encoder: Encoder{
				Format: RGBA,
			},
			img:     nil,
			wantErr: false,
			verify: func(got []byte) error {
				if len(got) != 0 {
					t.Errorf("expected empty output for nil image, got %d bytes", len(got))
				}
				return nil
			},
		},
		{
			name: "RGBA format",
			encoder: Encoder{
				Format: RGBA,
			},
			img:     testImage(),
			wantErr: false,
			verify: func(got []byte) error {
				expected := []byte{
					255, 0, 0, 255, // Red pixel
					0, 0, 255, 255, // Blue pixel
					0, 0, 255, 255, // Blue pixel
					255, 0, 0, 255, // Red pixel
				}
				if !bytes.Equal(got, expected) {
					t.Errorf("unexpected RGBA output\ngot:  %v\nwant: %v", got, expected)
				}
				return nil
			},
		},
		{
			name: "RGB format",
			encoder: Encoder{
				Format: RGB,
			},
			img:     testImage(),
			wantErr: false,
			verify: func(got []byte) error {
				expected := []byte{
					255, 0, 0, // Red pixel
					0, 0, 255, // Blue pixel
					0, 0, 255, // Blue pixel
					255, 0, 0, // Red pixel
				}
				if !bytes.Equal(got, expected) {
					t.Errorf("unexpected RGB output\ngot:  %v\nwant: %v", got, expected)
				}
				return nil
			},
		},
		{
			name: "PNG format",
			encoder: Encoder{
				Format: PNG,
			},
			img:     testImage(),
			wantErr: false,
			verify: func(got []byte) error {
				// Verify PNG header
				// if len(got) < 8 || !bytes.Equal(got[:8], []byte{137, 80, 78, 71, 13, 10, 26, 10}) {
				if len(got) < 8 || !bytes.Equal(got[:8], []byte(pngHeader)) {
					t.Error("invalid PNG header")
				}
				return nil
			},
		},
		{
			name: "invalid format",
			encoder: Encoder{
				Format: 999, // Invalid format
			},
			img:     testImage(),
			wantErr: true,
			verify:  nil,
		},
		{
			name: "RGBA with compression",
			encoder: Encoder{
				Format:   RGBA,
				Compress: true,
			},
			img:     testImage(),
			wantErr: false,
			verify: func(got []byte) error {
				// Decompress the data
				r, err := zlib.NewReader(bytes.NewReader(got))
				if err != nil {
					return err //nolint:wrapcheck
				}
				defer r.Close()

				decompressed, err := io.ReadAll(r)
				if err != nil {
					return err //nolint:wrapcheck
				}

				expected := []byte{
					255, 0, 0, 255, // Red pixel
					0, 0, 255, 255, // Blue pixel
					0, 0, 255, 255, // Blue pixel
					255, 0, 0, 255, // Red pixel
				}
				if !bytes.Equal(decompressed, expected) {
					t.Errorf("unexpected decompressed output\ngot:  %v\nwant: %v", decompressed, expected)
				}
				return nil
			},
		},
		{
			name: "zero format defaults to RGBA",
			encoder: Encoder{
				Format: 0,
			},
			img:     testImage(),
			wantErr: false,
			verify: func(got []byte) error {
				expected := []byte{
					255, 0, 0, 255, // Red pixel
					0, 0, 255, 255, // Blue pixel
					0, 0, 255, 255, // Blue pixel
					255, 0, 0, 255, // Red pixel
				}
				if !bytes.Equal(got, expected) {
					t.Errorf("unexpected RGBA output\ngot:  %v\nwant: %v", got, expected)
				}
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := tt.encoder.Encode(&buf, tt.img)

			if (err != nil) != tt.wantErr {
				t.Errorf("Encode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.verify != nil {
				if err := tt.verify(buf.Bytes()); err != nil {
					t.Errorf("verification failed: %v", err)
				}
			}
		})
	}
}

func TestEncoder_EncodeWithDifferentImageTypes(t *testing.T) {
	// Create different image types for testing
	rgba := image.NewRGBA(image.Rect(0, 0, 1, 1))
	rgba.Set(0, 0, color.RGBA{R: 255, G: 0, B: 0, A: 255})

	gray := image.NewGray(image.Rect(0, 0, 1, 1))
	gray.Set(0, 0, color.Gray{Y: 128})

	tests := []struct {
		name    string
		img     image.Image
		format  int
		wantLen int
	}{
		{
			name:    "RGBA image to RGBA format",
			img:     rgba,
			format:  RGBA,
			wantLen: 4, // 4 bytes per pixel
		},
		{
			name:    "Gray image to RGBA format",
			img:     gray,
			format:  RGBA,
			wantLen: 4, // 4 bytes per pixel
		},
		{
			name:    "RGBA image to RGB format",
			img:     rgba,
			format:  RGB,
			wantLen: 3, // 3 bytes per pixel
		},
		{
			name:    "Gray image to RGB format",
			img:     gray,
			format:  RGB,
			wantLen: 3, // 3 bytes per pixel
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			enc := Encoder{Format: tt.format}

			err := enc.Encode(&buf, tt.img)
			if err != nil {
				t.Errorf("Encode() error = %v", err)
				return
			}

			if got := buf.Len(); got != tt.wantLen {
				t.Errorf("Encode() output length = %v, want %v", got, tt.wantLen)
			}
		})
	}
}
