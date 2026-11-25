package ansi

import "testing"

func TestXParseColor(t *testing.T) {
	cases := []struct {
		input string
		want  uint32
		valid bool
	}{
		{"rgb:1/A/f", 0x11aaffff, true},
		{"rgb:01/a5/ff", 0x01a5ffff, true},
		{"rgb:00f/800/AbC", 0x0080abff, true},
		{"rgb:00ee/8000/ffff", 0x0080ffff, true},
		{"rgb:00ee/00ee/00ee", 0x000000ff, true},
		{"rgb:f/80/00ee", 0xff8000ff, true},
		{"rgba:1/02/030/8000", 0x11020380, true},
		{"rgba:11/22/33/44", 0x11223344, true},
		{"#aBc", 0xaabbccff, true},
		{"#0123aB", 0x0123abff, true},
		{"", 0, false},
		{"#xyz", 0, false},
		{"rgb:1/2", 0, false},
		{"rgb:1/2/3/4", 0, false},
		{"rgb:1//3", 0, false},
		{"rgb:00001/2/3", 0, false},
		{"rgb:1/2/3g", 0, false},
		{"rgba:1/2/3/", 0, false},
		{"rgba:1/2/3/10000", 0, false},
		{"rgba:1/2/3/gg", 0, false},
	}

	for _, c := range cases {
		got := XParseColor(c.input)
		if !c.valid {
			if got != nil {
				t.Errorf("XParseColor(%q): got %v, want nil", c.input, got)
			}
			continue
		}
		if got == nil {
			t.Errorf("XParseColor(%q): got nil, want #%08x", c.input, c.want)
			continue
		}
		r, g, b, a := got.RGBA()
		rgba := (r>>8)<<24 | (g>>8)<<16 | (b>>8)<<8 | (a >> 8)
		if rgba != c.want {
			t.Errorf("XParseColor(%q): got #%08x, want #%08x", c.input, rgba, c.want)
		}
	}
}
