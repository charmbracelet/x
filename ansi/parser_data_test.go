package ansi

import (
	"bytes"
	"strings"
	"testing"
)

// TestDataBufferGrowsAsDataArrives checks that the data buffer is allocated
// as a string's data arrives, never up front, and never past the size set.
func TestDataBufferGrowsAsDataArrives(t *testing.T) {
	const size = 4 * 1024 * 1024
	var got []byte
	p := NewParser()
	p.SetHandler(Handler{HandleOsc: func(_ int, data []byte) { got = bytes.Clone(data) }})
	p.SetDataSize(size)
	if c := cap(p.data); c != 0 {
		t.Fatalf("want no buffer before any data, got %d bytes", c)
	}

	payload := "52;c;" + strings.Repeat("a", 100*1024)
	p.Parse([]byte("\x1b]" + payload + "\x07"))
	if string(got) != payload {
		t.Fatalf("want the %d bytes of data, got %d", len(payload), len(got))
	}
	if c := cap(p.data); c < len(payload) || c > 2*len(payload) {
		t.Fatalf("want a buffer about the data's size, got %d bytes for %d", c, len(payload))
	}

	p.Parse([]byte("\x1b]" + strings.Repeat("b", size+10) + "\x07"))
	if len(got) != size || cap(p.data) != size {
		t.Fatalf("want the data cut at %d bytes in a buffer that size, got %d in %d", size, len(got), cap(p.data))
	}
}

// TestDataBufferUnlimited checks that a size of 0 still takes any length.
func TestDataBufferUnlimited(t *testing.T) {
	const n = 200 * 1024
	var got int
	p := NewParser()
	p.SetHandler(Handler{HandleOsc: func(_ int, data []byte) { got = len(data) }})
	p.SetDataSize(0)
	p.Parse([]byte("\x1b]" + strings.Repeat("a", n) + "\x07"))
	if got != n {
		t.Fatalf("want all %d bytes, got %d", n, got)
	}
}
