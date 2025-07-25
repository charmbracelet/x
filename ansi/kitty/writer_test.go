package kitty

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteKittyGraphics(t *testing.T) {
	// Create a test image
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	img.Set(0, 0, color.RGBA{R: 255, G: 0, B: 0, A: 255})
	img.Set(1, 0, color.RGBA{R: 0, G: 255, B: 0, A: 255})
	img.Set(0, 1, color.RGBA{R: 0, G: 0, B: 255, A: 255})
	img.Set(1, 1, color.RGBA{R: 255, G: 255, B: 255, A: 255})

	// Create large test image (larger than [MaxChunkSize] 4096 bytes)
	imgLarge := image.NewRGBA(image.Rect(0, 0, 100, 100))
	for y := range 100 {
		for x := range 100 {
			imgLarge.Set(x, y, color.RGBA{R: 255, G: 0, B: 0, A: 255})
		}
	}

	// Create a temporary test file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test-image")
	if err := os.WriteFile(tmpFile, []byte("test image data"), 0o644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name      string
		img       image.Image
		opts      *Options
		wantError bool
		check     func(t *testing.T, output string)
	}{
		{
			name: "direct transmission",
			img:  img,
			opts: &Options{
				Transmission: Direct,
				Format:       RGB,
			},
			wantError: false,
			check: func(t *testing.T, output string) {
				if !strings.HasPrefix(output, "\x1b_G") {
					t.Error("output should start with ESC sequence")
				}
				if !strings.HasSuffix(output, "\x1b\\") {
					t.Error("output should end with ST sequence")
				}
				if !strings.Contains(output, "f=24") {
					t.Error("output should contain format specification")
				}
			},
		},
		{
			name: "chunked transmission",
			img:  imgLarge,
			opts: &Options{
				Transmission: Direct,
				Format:       RGB,
				Chunk:        true,
			},
			wantError: false,
			check: func(t *testing.T, output string) {
				chunks := strings.Split(output, "\x1b\\")
				if len(chunks) < 2 {
					t.Error("output should contain multiple chunks")
				}

				chunks = chunks[:len(chunks)-1] // Remove last empty chunk
				for i, chunk := range chunks {
					if i == len(chunks)-1 {
						if !strings.Contains(chunk, "m=0") {
							t.Errorf("output should contain chunk end-of-data indicator for chunk %d %q", i, chunk)
						}
					} else {
						if !strings.Contains(chunk, "m=1") {
							t.Errorf("output should contain chunk indicator for chunk %d %q", i, chunk)
						}
					}
				}
			},
		},
		{
			name: "file transmission",
			img:  img,
			opts: &Options{
				Transmission: File,
				File:         tmpFile,
			},
			wantError: false,
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, base64.StdEncoding.EncodeToString([]byte(tmpFile))) {
					t.Error("output should contain encoded file path")
				}
			},
		},
		{
			name: "temp file transmission",
			img:  img,
			opts: &Options{
				Transmission: TempFile,
			},
			wantError: false,
			check: func(t *testing.T, output string) {
				output = strings.TrimPrefix(output, "\x1b_G")
				output = strings.TrimSuffix(output, "\x1b\\")
				payload := strings.Split(output, ";")[1]
				fn, err := base64.StdEncoding.DecodeString(payload)
				if err != nil {
					t.Error("output should contain base64 encoded temp file path")
				}
				if !strings.Contains(string(fn), "tty-graphics-protocol") {
					t.Error("output should contain temp file path")
				}
				if !strings.Contains(output, "t=t") {
					t.Error("output should contain transmission specification")
				}
			},
		},
		{
			name: "compression enabled",
			img:  img,
			opts: &Options{
				Transmission: Direct,
				Compression:  Zlib,
			},
			wantError: false,
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "o=z") {
					t.Error("output should contain compression specification")
				}
			},
		},
		{
			name: "invalid file path",
			img:  img,
			opts: &Options{
				Transmission: File,
				File:         "/nonexistent/file",
			},
			wantError: true,
			check:     nil,
		},
		{
			name:      "nil options",
			img:       img,
			opts:      nil,
			wantError: false,
			check: func(t *testing.T, output string) {
				if !strings.HasPrefix(output, "\x1b_G") {
					t.Error("output should start with ESC sequence")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := EncodeGraphics(&buf, tt.img, tt.opts)

			if (err != nil) != tt.wantError {
				t.Errorf("WriteKittyGraphics() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if !tt.wantError && tt.check != nil {
				tt.check(t, buf.String())
			}
		})
	}
}

func TestWriteKittyGraphicsEdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		img       image.Image
		opts      *Options
		wantError bool
	}{
		{
			name: "zero size image",
			img:  image.NewRGBA(image.Rect(0, 0, 0, 0)),
			opts: &Options{
				Transmission: Direct,
			},
			wantError: false,
		},
		{
			name: "shared memory transmission",
			img:  image.NewRGBA(image.Rect(0, 0, 1, 1)),
			opts: &Options{
				Transmission: SharedMemory,
			},
			wantError: true, // Not implemented
		},
		{
			name: "file transmission without file path",
			img:  image.NewRGBA(image.Rect(0, 0, 1, 1)),
			opts: &Options{
				Transmission: File,
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := EncodeGraphics(&buf, tt.img, tt.opts)

			if (err != nil) != tt.wantError {
				t.Errorf("WriteKittyGraphics() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}
