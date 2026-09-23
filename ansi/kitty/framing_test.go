package kitty

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/png"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

// Verify controls independently from buildChunkOptions, including exact-boundary
// empty terminating chunks, then return the reassembled binary payload.
func checkGraphicsChunks(t *testing.T, output string, opts *Options) []byte {
	t.Helper()
	if !strings.HasSuffix(output, "\x1b\\") {
		t.Fatal("missing final ST")
	}
	chunks := strings.Split(strings.TrimSuffix(output, "\x1b\\"), "\x1b\\")
	var payload strings.Builder
	for i, chunk := range chunks {
		if !strings.HasPrefix(chunk, "\x1b_G") {
			t.Fatalf("chunk %d: missing APC", i)
		}
		controls, data, _ := strings.Cut(chunk[3:], ";")
		if opts.Chunk && (len(data) > MaxChunkSize || len(data)%4 != 0) {
			t.Fatalf("invalid chunk length %d", len(data))
		}
		if i < len(chunks)-1 && len(data) != MaxChunkSize {
			t.Fatal("short non-final chunk")
		}
		var want []string
		if i == 0 {
			copy := *opts
			want = copy.Options()
		} else {
			quiet := opts.Quiet
			if opts.Quite != 0 {
				quiet = opts.Quite
			}
			if quiet != 0 {
				want = append(want, fmt.Sprintf("q=%d", quiet))
			}
			if opts.Action == Frame {
				want = append(want, "a=f")
			}
		}
		if len(chunks) > 1 {
			more := "m=1"
			if i == len(chunks)-1 {
				more = "m=0"
			}
			want = append(want, more)
		}
		if controls != strings.Join(want, ",") {
			t.Fatalf("chunk %d controls: got %q, want %q", i, controls, strings.Join(want, ","))
		}
		payload.WriteString(data)
	}
	encodedLen := payload.Len()
	wantChunks := 1
	if opts.Chunk {
		wantChunks = encodedLen/MaxChunkSize + 1
	}
	if len(chunks) != wantChunks {
		t.Fatalf("got %d chunks, want %d", len(chunks), wantChunks)
	}
	decoded, err := base64.StdEncoding.DecodeString(payload.String())
	if err != nil {
		t.Fatal(err)
	}
	return decoded
}

func TestEncodeGraphicsPNGFraming(t *testing.T) {
	img := pngTestImage()
	for _, level := range []png.CompressionLevel{png.DefaultCompression, png.BestSpeed} {
		for _, chunk := range []bool{false, true} {
			for _, quiet := range []struct{ current, legacy byte }{{0, 0}, {1, 0}, {2, 0}, {1, 2}} {
				for _, action := range []byte{TransmitAndPut, Frame} {
					t.Run(fmt.Sprintf("level=%d/chunk=%t/quiet=%d,%d/action=%c", level, chunk, quiet.current, quiet.legacy, action), func(t *testing.T) {
						opts := &Options{Transmission: Direct, Format: PNG, PNGCompressionLevel: level, Chunk: chunk, Quiet: quiet.current, Quite: quiet.legacy, Action: action, ID: 45}
						var plain, wrapped bytes.Buffer
						if err := EncodeGraphics(&plain, img, opts); err != nil {
							t.Fatal(err)
						}
						assertPNG(t, checkGraphicsChunks(t, plain.String(), opts), img, level)
						var formatted, inputs strings.Builder
						calls := 0
						opts.ChunkFormatter = func(s string) string {
							calls++
							inputs.WriteString(s)
							s = ansi.TmuxPassthrough(s)
							formatted.WriteString(s)
							return s
						}
						if err := EncodeGraphics(&wrapped, img, opts); err != nil {
							t.Fatal(err)
						}
						if chunk {
							wantCalls := strings.Count(plain.String(), "\x1b_G")
							if calls != wantCalls || inputs.String() != plain.String() || wrapped.String() != formatted.String() {
								t.Fatal("formatter must apply once to every chunk")
							}
						} else if calls != 0 || wrapped.String() != plain.String() {
							t.Fatal("unchunked formatter behavior changed")
						}
					})
				}
			}
		}
	}
}

func TestEncodeGraphicsChunkBoundaries(t *testing.T) {
	// RGB produces three bytes per pixel, hence four base64 bytes per pixel.
	for _, pixels := range []int{0, 1, 1023, 1024, 1025, 2048, 2049} {
		t.Run(fmt.Sprint(pixels), func(t *testing.T) {
			img := image.NewRGBA(image.Rect(0, 0, pixels, 1))
			opts := &Options{Transmission: Direct, Format: RGB, Chunk: true, Quiet: 2}
			var out bytes.Buffer
			if err := EncodeGraphics(&out, img, opts); err != nil {
				t.Fatal(err)
			}
			if got := checkGraphicsChunks(t, out.String(), opts); !bytes.Equal(got, make([]byte, pixels*3)) {
				t.Fatal("chunking changed payload")
			}
		})
	}
}
