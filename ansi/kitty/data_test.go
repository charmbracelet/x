package kitty

import (
	"bytes"
	"compress/zlib"
	"errors"
	"fmt"
	"image/png"
	"io"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestEncodeGraphicsDataBoundaries(t *testing.T) {
	for _, size := range []int{0, 1, 3070, 3071, 3072, 3073, 6144, 6145} {
		for _, chunk := range []bool{false, true} {
			for _, action := range []byte{TransmitAndPut, Frame} {
				t.Run(fmt.Sprintf("size=%d/chunk=%t/action=%c", size, chunk, action), func(t *testing.T) {
					data := bytes.Repeat([]byte{0x42}, size)
					opts := &Options{Transmission: Direct, Format: PNG, Chunk: chunk, Action: action, Quiet: 1, Quite: 2, ID: 45}
					var plain, wrapped bytes.Buffer
					if err := EncodeGraphicsData(&plain, data, opts); err != nil {
						t.Fatal(err)
					}
					if got := checkGraphicsChunks(t, plain.String(), opts); !bytes.Equal(got, data) {
						t.Fatal("payload changed")
					}
					var inputs, formatted strings.Builder
					calls := 0
					opts.ChunkFormatter = func(s string) string {
						calls++
						inputs.WriteString(s)
						s = ansi.TmuxPassthrough(s)
						formatted.WriteString(s)
						return s
					}
					if err := EncodeGraphicsData(&wrapped, data, opts); err != nil {
						t.Fatal(err)
					}
					if chunk {
						if calls != strings.Count(plain.String(), "\x1b_G") || inputs.String() != plain.String() || formatted.String() != wrapped.String() {
							t.Fatal("formatter did not receive every original chunk")
						}
					} else if calls != 0 || plain.String() != wrapped.String() {
						t.Fatal("formatter must be ignored without chunking")
					}
				})
			}
		}
	}
}

func TestEncodeGraphicsDataMatchesImageEncoding(t *testing.T) {
	img := pngTestImage()
	for _, format := range []int{RGB, RGBA, PNG} {
		for _, compression := range []byte{0, Zlib} {
			t.Run(fmt.Sprintf("format=%d/compression=%d", format, compression), func(t *testing.T) {
				var data, imageOutput, dataOutput bytes.Buffer
				enc := Encoder{Format: format, Compress: compression == Zlib, PNGCompressionLevel: png.BestSpeed}
				if err := enc.Encode(&data, img); err != nil {
					t.Fatal(err)
				}
				opts := Options{Transmission: Direct, Format: format, Compression: compression, PNGCompressionLevel: png.BestSpeed, ImageWidth: img.Rect.Dx(), ImageHeight: img.Rect.Dy(), Quiet: 2, Chunk: true}
				copy := opts
				if err := EncodeGraphics(&imageOutput, img, &opts); err != nil {
					t.Fatal(err)
				}
				if err := EncodeGraphicsData(&dataOutput, data.Bytes(), &copy); err != nil {
					t.Fatal(err)
				}
				if !bytes.Equal(imageOutput.Bytes(), dataOutput.Bytes()) {
					t.Fatal("image and data entry points differ")
				}
				if got := checkGraphicsChunks(t, dataOutput.String(), &copy); !bytes.Equal(got, data.Bytes()) {
					t.Fatal("prepared payload was transformed")
				}
			})
		}
	}
}

func TestEncodeGraphicsDataResourceNames(t *testing.T) {
	// Intentionally nonexistent names: framing must not access the resources or
	// substitute Options.File for the supplied payload.
	for _, transmission := range []byte{File, TempFile, SharedMemory, 0} {
		t.Run(fmt.Sprintf("transmission=%d", transmission), func(t *testing.T) {
			data := []byte("/nonexistent/tty-graphics-protocol-caller-owned")
			opts := &Options{Transmission: transmission, File: "/ignored", Format: PNG, Size: 123, Offset: 7, Quiet: 2, Chunk: true}
			var out bytes.Buffer
			if err := EncodeGraphicsData(&out, data, opts); err != nil {
				t.Fatal(err)
			}
			if got := checkGraphicsChunks(t, out.String(), opts); !bytes.Equal(got, data) {
				t.Fatal("resource name changed")
			}
			wantTransmission := transmission
			if wantTransmission == 0 {
				wantTransmission = File
			}
			if !strings.Contains(out.String(), fmt.Sprintf("t=%c", wantTransmission)) {
				t.Fatal("missing transmission control")
			}
		})
	}
}

func TestEncodeGraphicsDataNilOptions(t *testing.T) {
	for _, data := range [][]byte{nil, {0, 1, 2, 255}} {
		var out bytes.Buffer
		if err := EncodeGraphicsData(&out, data, nil); err != nil {
			t.Fatal(err)
		}
		if got := checkGraphicsChunks(t, out.String(), &Options{}); !bytes.Equal(got, data) {
			t.Fatal("payload changed with default options")
		}
	}
}

type graphicsErrorWriter struct {
	calls, failAt int
	err           error
}

func (w *graphicsErrorWriter) Write(p []byte) (int, error) {
	w.calls++
	if w.calls == w.failAt {
		return 0, w.err
	}
	return len(p), nil
}

func TestEncodeGraphicsDataWriterErrors(t *testing.T) {
	sentinel := errors.New("writer failed")
	for _, chunk := range []bool{false, true} {
		failures := []int{1}
		if chunk {
			failures = []int{1, 2, 3}
		}
		for _, failAt := range failures {
			t.Run(fmt.Sprintf("chunk=%t/failAt=%d", chunk, failAt), func(t *testing.T) {
				w := &graphicsErrorWriter{failAt: failAt, err: sentinel}
				// Two full chunks and an empty terminating chunk.
				err := EncodeGraphicsData(w, make([]byte, 6144), &Options{Chunk: chunk})
				if !errors.Is(err, sentinel) || w.calls != failAt {
					t.Fatalf("got error %v after %d writes", err, w.calls)
				}
			})
		}
	}
}

func ExampleEncodeGraphicsData() {
	// A caller can prepare and retain compressed raw pixel bytes.
	var data bytes.Buffer
	zw := zlib.NewWriter(&data)
	_, _ = zw.Write([]byte{255, 0, 0, 255}) // One red RGBA pixel.
	_ = zw.Close()
	err := EncodeGraphicsData(io.Discard, data.Bytes(), &Options{
		Transmission: Direct,
		Format:       RGBA,
		ImageWidth:   1,
		ImageHeight:  1,
		Compression:  Zlib,
		Chunk:        true,
	})
	fmt.Println(err)
	// Output: <nil>
}
