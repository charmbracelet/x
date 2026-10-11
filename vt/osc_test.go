package vt

import "testing"

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

func TestOscPayloadContainingSemicolon(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		input     string
		wantTitle string
		wantIcon  string
		wantCwd   string
		wantURL   string
		wantParam string
	}{
		{name: "OSC 0", input: "\x1b]0;a;b\x07", wantTitle: "a;b", wantIcon: "a;b"},
		{name: "OSC 1", input: "\x1b]1;a;b\x07", wantIcon: "a;b"},
		{name: "OSC 2", input: "\x1b]2;a;b;c\x07", wantTitle: "a;b;c"},
		{name: "OSC 7", input: "\x1b]7;file://host/tmp/a;b\x07", wantCwd: "file://host/tmp/a;b"},
		{name: "OSC 8", input: "\x1b]8;id=1;https://example.com/a;b\x07", wantURL: "https://example.com/a;b", wantParam: "id=1"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			term := newTestTerminal(t, 20, 2)
			_, _ = term.WriteString(tc.input)
			if term.title != tc.wantTitle || term.iconName != tc.wantIcon || term.cwd != tc.wantCwd {
				t.Errorf("got title %q icon %q cwd %q, want title %q icon %q cwd %q",
					term.title, term.iconName, term.cwd, tc.wantTitle, tc.wantIcon, tc.wantCwd)
			}
			if l := term.scr.cur.Link; l.URL != tc.wantURL || l.Params != tc.wantParam {
				t.Errorf("link: got URL %q params %q, want URL %q params %q", l.URL, l.Params, tc.wantURL, tc.wantParam)
			}
		})
	}
}
