package vt

import "testing"

// TestHyperlinkURLWithSemicolon verifies that an OSC 8 hyperlink whose URI
// contains a semicolon is kept and attached to the cells, rather than being
// silently dropped. URLs may legally contain semicolons (for example in query
// strings or path parameters), and only the first two semicolons of the OSC 8
// payload are field separators.
//
// The assertion is agnostic to whether the URI lands in Link.URL or
// Link.Params: the exchange of those two fields is a separate bug tracked by
// PR #868, and this test only checks that the URI survives the parser at all.
func TestHyperlinkURLWithSemicolon(t *testing.T) {
	const uri = "https://example.com/f?a=1;b=2"

	term := newTestTerminal(t, 10, 1)
	// Set the hyperlink, write "AB", then reset the hyperlink, matching the
	// reproduction from the issue: OSC 8 ; ; <uri> ST  AB  OSC 8 ; ; ST.
	term.Write([]byte("\x1b]8;;" + uri + "\x1b\\AB\x1b]8;;\x1b\\"))

	for _, x := range []int{0, 1} {
		cell := term.CellAt(x, 0)
		if cell == nil {
			t.Fatalf("expected a cell at (%d, 0)", x)
		}
		got := cell.Link.URL
		if got != uri {
			got = cell.Link.Params
		}
		if got != uri {
			t.Errorf("OSC 8 hyperlink with a semicolon in the URI was dropped at (%d, 0):\nwant URI %q in Link.URL or Link.Params\ngot  URL=%q Params=%q",
				x, uri, cell.Link.URL, cell.Link.Params)
		}
	}
}
