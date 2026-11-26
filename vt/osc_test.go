package vt

import (
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"
)

// assertIndexedColor checks that the terminal's indexed color at i has the
// given 16-bit RGBA values.
func assertIndexedColor(t *testing.T, term *Emulator, i int, r, g, b, a uint32) {
	t.Helper()

	c := term.IndexedColor(i)
	if c == nil {
		t.Fatalf("IndexedColor(%d) = nil", i)
	}
	gr, gg, gb, ga := c.RGBA()
	if gr != r || gg != g || gb != b || ga != a {
		t.Errorf("IndexedColor(%d) = RGBA(%04x, %04x, %04x, %04x), want RGBA(%04x, %04x, %04x, %04x)",
			i, gr, gg, gb, ga, r, g, b, a)
	}
}

// assertDefaultIndexedColor checks that the terminal's indexed color at i is
// still the default palette color.
func assertDefaultIndexedColor(t *testing.T, term *Emulator, i int) {
	t.Helper()

	got := term.IndexedColor(i)
	want := ansi.IndexedColor(i)
	gr, gg, gb, ga := got.RGBA()
	wr, wg, wb, wa := want.RGBA()
	if gr != wr || gg != wg || gb != wb || ga != wa {
		t.Errorf("IndexedColor(%d) = RGBA(%04x, %04x, %04x, %04x), want default RGBA(%04x, %04x, %04x, %04x)",
			i, gr, gg, gb, ga, wr, wg, wb, wa)
	}
}

// TestHandleIndexedColor covers OSC 4 (set/query palette entry) and OSC 104
// (reset palette entries) per xterm ctlseqs.
func TestHandleIndexedColor(t *testing.T) {
	t.Parallel()

	t.Run("issue repro: OSC 4 sets palette entry", func(t *testing.T) {
		t.Parallel()

		term := newTestTerminal(t, 80, 24)
		assertDefaultIndexedColor(t, term, 1)
		_, _ = term.Write([]byte("\x1b]4;1;#ff0000\x1b\\"))
		assertIndexedColor(t, term, 1, 0xffff, 0x0000, 0x0000, 0xffff)
	})

	t.Run("OSC 4 sets multiple entries in one sequence", func(t *testing.T) {
		t.Parallel()

		term := newTestTerminal(t, 80, 24)
		_, _ = term.Write([]byte("\x1b]4;1;#ff0000;2;#00ff00\x1b\\"))
		assertIndexedColor(t, term, 1, 0xffff, 0x0000, 0x0000, 0xffff)
		assertIndexedColor(t, term, 2, 0x0000, 0xffff, 0x0000, 0xffff)
	})

	t.Run("OSC 4 query replies with current color", func(t *testing.T) {
		t.Parallel()

		term := newTestTerminal(t, 80, 24)
		_, _ = term.Write([]byte("\x1b]4;1;#ff0000\x1b\\"))
		want := "\x1b]4;1;rgb:ffff/0000/0000\x1b\\"
		// Start reading before writing the query: the reply is written to
		// the terminal's output pipe while handling the sequence, and the
		// pipe writer blocks until a reader consumes it.
		replyCh := make(chan string, 1)
		go func() {
			buf := make([]byte, 64)
			var out []byte
			for len(out) < len(want) {
				n, err := term.Read(buf)
				if err != nil {
					return
				}
				out = append(out, buf[:n]...)
			}
			replyCh <- string(out)
		}()
		_, _ = term.Write([]byte("\x1b]4;1;?\x1b\\"))
		select {
		case got := <-replyCh:
			if got != want {
				t.Errorf("query reply = %q, want %q", got, want)
			}
		case <-time.After(2 * time.Second):
			t.Fatal("timed out waiting for terminal input")
		}
	})

	t.Run("OSC 104 resets a single entry", func(t *testing.T) {
		t.Parallel()

		term := newTestTerminal(t, 80, 24)
		_, _ = term.Write([]byte("\x1b]4;1;#ff0000;2;#00ff00\x1b\\"))
		_, _ = term.Write([]byte("\x1b]104;1\x1b\\"))
		assertDefaultIndexedColor(t, term, 1)
		assertIndexedColor(t, term, 2, 0x0000, 0xffff, 0x0000, 0xffff)
	})

	t.Run("OSC 104 with no index resets all entries", func(t *testing.T) {
		t.Parallel()

		term := newTestTerminal(t, 80, 24)
		_, _ = term.Write([]byte("\x1b]4;1;#ff0000;2;#00ff00\x1b\\"))
		_, _ = term.Write([]byte("\x1b]104\x1b\\"))
		assertDefaultIndexedColor(t, term, 1)
		assertDefaultIndexedColor(t, term, 2)
	})

	t.Run("OSC 4 ignores out-of-range and invalid indexes", func(t *testing.T) {
		t.Parallel()

		term := newTestTerminal(t, 80, 24)
		_, _ = term.Write([]byte("\x1b]4;300;#ff0000\x1b\\"))
		_, _ = term.Write([]byte("\x1b]4;abc;#ff0000\x1b\\"))
		_, _ = term.Write([]byte("\x1b]104;300\x1b\\"))
		assertDefaultIndexedColor(t, term, 1)
		assertDefaultIndexedColor(t, term, 2)
	})
}

// OSC 8 wire format per https://gist.github.com/egmontkov/eb6100b9 is
// ESC ] 8 ; <params> ; <uri> ST. handleHyperlink must store the
// <params> segment in Link.Params and the <uri> segment in Link.URL.
// Prior to the fix in this commit the two were swapped: Link.URL
// received parts[1] (params) and Link.Params received parts[2] (uri),
// which produced visible-but-unclickable links once Buffer.Render()
// emitted them back out via ansi.SetHyperlink(URL, Params).
func TestHandleHyperlink_StoresURLAndParamsInCorrectFields(t *testing.T) {
	t.Parallel()

	t.Run("empty params, populated uri (common case)", func(t *testing.T) {
		t.Parallel()
		term := newTestTerminal(t, 80, 24)
		// Feed an OSC 8 open + a single character + OSC 8 close so the
		// open's Link metadata sticks on the cell.
		_, _ = term.Write([]byte("\x1b]8;;https://example.com/\x1b\\X\x1b]8;;\x1b\\"))
		cell := term.CellAt(0, 0)
		if cell == nil {
			t.Fatal("cell at (0,0) is nil")
		}
		if got, want := cell.Link.URL, "https://example.com/"; got != want {
			t.Errorf("Link.URL = %q, want %q", got, want)
		}
		if got, want := cell.Link.Params, ""; got != want {
			t.Errorf("Link.Params = %q, want %q", got, want)
		}
	})

	t.Run("populated params and uri", func(t *testing.T) {
		t.Parallel()
		term := newTestTerminal(t, 80, 24)
		_, _ = term.Write([]byte("\x1b]8;id=42:tag=demo;https://example.com/path?x=1\x1b\\X\x1b]8;;\x1b\\"))
		cell := term.CellAt(0, 0)
		if cell == nil {
			t.Fatal("cell at (0,0) is nil")
		}
		if got, want := cell.Link.URL, "https://example.com/path?x=1"; got != want {
			t.Errorf("Link.URL = %q, want %q", got, want)
		}
		if got, want := cell.Link.Params, "id=42:tag=demo"; got != want {
			t.Errorf("Link.Params = %q, want %q", got, want)
		}
	})
}
