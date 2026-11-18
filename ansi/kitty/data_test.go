package kitty

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"errors"
	"fmt"
	"image/png"
	"io"
	"strings"
	"testing"
)

func TestEncodeGraphicsDataChunkBoundaries(t *testing.T) {
	full := "\x1b_Gq=2,m=1;" + strings.Repeat("A", 4096) + "\x1b\\"
	end := "\x1b_Gq=2,m=0\x1b\\"
	tests := []struct {
		name string
		size int
		want string
	}{
		{"empty", 0, "\x1b_Gq=2\x1b\\"},
		{"one byte", 1, "\x1b_Gq=2;AA==\x1b\\"},
		{"below boundary", 3069, "\x1b_Gq=2;" + strings.Repeat("A", 4092) + "\x1b\\"},
		{"two padding bytes", 3070, "\x1b_Gq=2,m=1;" + strings.Repeat("A", 4094) + "==\x1b\\" + end},
		{"one padding byte", 3071, "\x1b_Gq=2,m=1;" + strings.Repeat("A", 4095) + "=\x1b\\" + end},
		{"exact boundary", 3072, full + end},
		{"above boundary", 3073, full + "\x1b_Gq=2,m=0;AA==\x1b\\"},
		{"two full chunks", 6144, full + full + end},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			err := EncodeGraphicsData(&out, make([]byte, tt.size), &Options{Chunk: true, Quiet: 2})
			if err != nil {
				t.Fatal(err)
			}
			if out.String() != tt.want {
				t.Fatal("unexpected chunk boundaries or payload")
			}
		})
	}
}

func TestEncodeGraphicsDataContinuationControls(t *testing.T) {
	tests := []struct {
		name        string
		opts        Options
		first, last string
	}{
		{"default", Options{}, "m=1", "m=0"},
		{"quiet 1", Options{Quiet: 1}, "q=1,m=1", "q=1,m=0"},
		{"quiet 2", Options{Quiet: 2}, "q=2,m=1", "q=2,m=0"},
		{"legacy quiet", Options{Quiet: 1, Quite: 2}, "q=2,m=1", "q=2,m=0"},
		{"frame", Options{Action: Frame, Quiet: 2, ID: 45}, "q=2,i=45,a=f,m=1", "q=2,a=f,m=0"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.opts.Chunk = true
			var out bytes.Buffer
			if err := EncodeGraphicsData(&out, make([]byte, 3072), &tt.opts); err != nil {
				t.Fatal(err)
			}
			want := "\x1b_G" + tt.first + ";" + strings.Repeat("A", 4096) + "\x1b\\" +
				"\x1b_G" + tt.last + "\x1b\\"
			if out.String() != want {
				t.Fatal("unexpected continuation controls")
			}
		})
	}
}

func TestEncodeGraphicsDataChunkFormatter(t *testing.T) {
	payload := strings.Repeat("A", 4096)
	tests := []struct {
		name  string
		chunk bool
		want  string
	}{
		{"enabled", true, "[\x1b_Gm=1;" + payload + "\x1b\\][\x1b_Gm=0\x1b\\]"},
		{"disabled", false, "\x1b_G;" + payload + "\x1b\\"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			opts := &Options{Chunk: tt.chunk, ChunkFormatter: func(s string) string { return "[" + s + "]" }}
			if err := EncodeGraphicsData(&out, make([]byte, 3072), opts); err != nil {
				t.Fatal(err)
			}
			if out.String() != tt.want {
				t.Fatal("unexpected formatter output")
			}
		})
	}
}

func TestEncodeGraphicsDataPreservesPayload(t *testing.T) {
	tests := []struct {
		name     string
		opts     Options
		controls string
	}{
		{"default", Options{}, ""},
		{"PNG", Options{Format: PNG, PNGCompressionLevel: png.BestSpeed}, "f=100"},
		{"RGB", Options{Format: RGB, ImageWidth: 1, ImageHeight: 1}, "f=24,s=1,v=1"},
		{"RGBA", Options{Format: RGBA, ImageWidth: 1, ImageHeight: 1}, "s=1,v=1"},
		{"compressed", Options{Compression: Zlib}, "o=z"},
		{"file", Options{Transmission: File, File: "/ignored"}, "t=f"},
		{"implicit file", Options{File: "/ignored"}, "t=f"},
		{"temporary file", Options{Transmission: TempFile}, "t=t"},
		{"shared memory", Options{Transmission: SharedMemory}, "t=s"},
	}
	data := []byte("caller-prepared-payload")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			if err := EncodeGraphicsData(&out, data, &tt.opts); err != nil {
				t.Fatal(err)
			}
			want := "\x1b_G" + tt.controls + ";" + base64.StdEncoding.EncodeToString(data) + "\x1b\\"
			if out.String() != want {
				t.Fatalf("got %q, want %q", out.String(), want)
			}
		})
	}
}

func TestEncodeGraphicsDataNilOptions(t *testing.T) {
	var out bytes.Buffer
	if err := EncodeGraphicsData(&out, []byte{0, 1, 2, 255}, nil); err != nil {
		t.Fatal(err)
	}
	if want := "\x1b_G;AAEC/w==\x1b\\"; out.String() != want {
		t.Fatalf("got %q, want %q", out.String(), want)
	}
}

func TestEncodeGraphicsDataMatchesPNG(t *testing.T) {
	img := testImage()
	var data, fromImage, fromData bytes.Buffer
	if err := png.Encode(&data, img); err != nil {
		t.Fatal(err)
	}
	opts := &Options{Transmission: Direct, Format: PNG, Chunk: true, Quiet: 2}
	if err := EncodeGraphics(&fromImage, img, opts); err != nil {
		t.Fatal(err)
	}
	if err := EncodeGraphicsData(&fromData, data.Bytes(), opts); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(fromImage.Bytes(), fromData.Bytes()) {
		t.Fatal("image and data entry points differ")
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
	tests := []struct {
		name   string
		chunk  bool
		failAt int
	}{
		{"unchunked", false, 1},
		{"first chunk", true, 1},
		{"middle chunk", true, 2},
		{"last chunk", true, 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &graphicsErrorWriter{failAt: tt.failAt, err: sentinel}
			err := EncodeGraphicsData(w, make([]byte, 6144), &Options{Chunk: tt.chunk})
			if !errors.Is(err, sentinel) || w.calls != tt.failAt {
				t.Fatalf("got error %v after %d writes", err, w.calls)
			}
		})
	}
}

func ExampleEncodeGraphicsData() {
	var data bytes.Buffer
	zw := zlib.NewWriter(&data)
	_, _ = zw.Write([]byte{255, 0, 0, 255})
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
