package kitty

import (
	"bytes"
	"encoding/base64"
	"image/png"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func graphicsPayload(t *testing.T, output string) []byte {
	t.Helper()
	var payload strings.Builder
	for _, chunk := range strings.SplitAfter(output, "\x1b\\") {
		if chunk == "" {
			continue
		}
		if !strings.HasPrefix(chunk, "\x1b_G") || !strings.HasSuffix(chunk, "\x1b\\") {
			t.Fatal("invalid graphics sequence")
		}
		_, data, _ := strings.Cut(strings.TrimSuffix(chunk, "\x1b\\"), ";")
		payload.WriteString(data)
	}
	data, err := base64.StdEncoding.DecodeString(payload.String())
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestEncodeGraphicsPNGChunkFormatter(t *testing.T) {
	img := pngTestImage()
	opts := &Options{
		Transmission:        Direct,
		Format:              PNG,
		PNGCompressionLevel: png.BestSpeed,
		Chunk:               true,
		Quiet:               2,
	}
	var plain, wrapped bytes.Buffer
	if err := EncodeGraphics(&plain, img, opts); err != nil {
		t.Fatal(err)
	}
	if got := graphicsPayload(t, plain.String()); !bytes.Equal(got, expectedPNG(t, img, png.BestSpeed)) {
		t.Fatal("chunking changed the PNG payload")
	}
	chunks := strings.SplitAfter(plain.String(), "\x1b\\")
	chunks = chunks[:len(chunks)-1]
	if len(chunks) < 2 {
		t.Fatal("test image must produce multiple chunks")
	}
	var want strings.Builder
	for _, chunk := range chunks {
		want.WriteString(ansi.TmuxPassthrough(chunk))
	}
	opts.ChunkFormatter = ansi.TmuxPassthrough
	if err := EncodeGraphics(&wrapped, img, opts); err != nil {
		t.Fatal(err)
	}
	if wrapped.String() != want.String() {
		t.Fatal("formatter must wrap each chunk separately")
	}
}
