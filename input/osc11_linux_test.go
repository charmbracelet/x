package input

import (
	"io"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestOSC11SplitAcrossReads(t *testing.T) {
	// first part *without* the final ST bytes
	p1 := []byte("\x1b]11;rgb:1a1a/1b1b/2c2c")
	// ST = ESC \
	p2 := []byte("\x1b\\")

	r := &chunkedReader{chunks: [][]byte{p1, p2}}
	ir, err := NewReader(r, "xterm-256color", 0)
	if err != nil {
		t.Fatal(err)
	}

	// 1st call consumes p1 – should not emit anything yet
	ev, _ := ir.ReadEvents()
	if len(ev) != 0 {
		t.Fatalf("unexpected event(s) after chunk1: %#v", ev)
	}

	// 2nd call must produce the colour
	ev, _ = ir.ReadEvents()
	if len(ev) != 1 {
		t.Fatalf("want 1 event, got %#v", ev)
	}
	bc, ok := ev[0].(BackgroundColorEvent)
	if !ok {
		t.Fatalf("got %T, want BackgroundColorEvent", ev[0])
	}
	want := ansi.XParseColor("rgb:1a1a/1b1b/2c2c")
	if bc.Color != want {
		t.Fatalf("wrong colour: %#v want %#v", bc.Color, want)
	}
}

// chunkedReader simulates a reader that returns data in separate chunks
type chunkedReader struct {
	chunks [][]byte
	index  int
}

func (r *chunkedReader) Read(p []byte) (n int, err error) {
	if r.index >= len(r.chunks) {
		return 0, io.EOF
	}

	chunk := r.chunks[r.index]
	r.index++

	n = copy(p, chunk)
	return n, nil
}

func TestOSC11SplitAcrossReadsBEL(t *testing.T) {
	// first part *without* the final BEL byte
	p1 := []byte("\x1b]11;rgb:1a1a/1b1b/2c2c")
	// BEL terminator
	p2 := []byte("\x07")

	r := &chunkedReader{chunks: [][]byte{p1, p2}}
	ir, err := NewReader(r, "xterm-256color", 0)
	if err != nil {
		t.Fatal(err)
	}

	// 1st call consumes p1 – should not emit anything yet
	ev, _ := ir.ReadEvents()
	if len(ev) != 0 {
		t.Fatalf("unexpected event(s) after chunk1: %#v", ev)
	}

	// 2nd call must produce the colour
	ev, _ = ir.ReadEvents()
	if len(ev) != 1 {
		t.Fatalf("want 1 event, got %#v", ev)
	}
	bc, ok := ev[0].(BackgroundColorEvent)
	if !ok {
		t.Fatalf("got %T, want BackgroundColorEvent", ev[0])
	}
	want := ansi.XParseColor("rgb:1a1a/1b1b/2c2c")
	if bc.Color != want {
		t.Fatalf("wrong colour: %#v want %#v", bc.Color, want)
	}
}
