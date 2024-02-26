package ansi

import "testing"

func TestOscSequenceIsValid(t *testing.T) {
	cases := []struct {
		in     string
		expect bool
	}{
		{"", false},
		{"\x1b]0", false},
		{"\x1b]0;", false},
		{"\x1b]0;hello", false},
		{"\x1b]1:hello\x07", false},
		{"\x1b]0;hello\x07", true},
		{"\x1b]0;hello\x1b\\", true},
		{"\x1b]1234;hello\x1b\\", true},
		{"\x1b]1234\x1b\\", true},
		{"\x1b]1234;abc;hello\x1b\\", true},
		{"\x9b1234;hello\x9c", false},
		{"\x9d]1234;hello\x9c", false},
		{"\x9d1234;hello\x9c", true},
	}

	for i, c := range cases {
		seq := OscSequence(c.in)
		if seq.IsValid() != c.expect {
			t.Errorf("case %d: expected %v, got %v", i+1, c.expect, seq.IsValid())
		}
	}
}

func TestOscSequenceIdentifier(t *testing.T) {
	cases := []struct {
		in     string
		expect string
	}{
		{"", ""},
		{"\x1b]0", ""},
		{"\x1b]0;", "0"},
		{"\x1b]0;hello", "0"},
		{"\x1b]1:hello\x07", ""},
		{"\x1b]0;hello\x07", "0"},
		{"\x1b]0\x07", "0"},
		{"\x1b]0;hello\x1b\\", "0"},
		{"\x1b]1234;hello\x1b\\", "1234"},
		{"\x1b]1234;abc;hello\x1b\\", "1234"},
		{"\x9b1234;hello\x9c", "1234"},
		{"\x9d]1234;hello\x9c", "1234"},
		{"\x9d1234\x9c", "1234"},
		{"\x9d1234;hello\x9c", "1234"},
	}

	for i, c := range cases {
		seq := OscSequence(c.in)
		if seq.Identifier() != c.expect {
			t.Errorf("case %d: expected %q, got %q", i+1, c.expect, seq.Identifier())
		}
	}
}

func TestOscSequenceData(t *testing.T) {
	cases := []struct {
		in     string
		expect string
	}{
		{"", ""},
		{"\x1b]0", ""},
		{"\x1b]0;", ""},
		{"\x1b]0;hello", ""},
		{"\x1b]1:hello\x07", ""},
		{"\x1b]0;hello\x07", "hello"},
		{"\x1b]0;hello\x1b\\", "hello"},
		{"\x1b]1234;hello\x1b\\", "hello"},
		{"\x1b]1234;abc;hello\x1b\\", "abc;hello"},
		{"\x9b1234;hello\x9c", "hello"},
		{"\x9d]1234;hello\x9c", "hello"},
		{"\x9d1234;hello\x9c", "hello"},
	}

	for i, c := range cases {
		seq := OscSequence(c.in)
		if seq.Data() != c.expect {
			t.Errorf("case %d: expected %q, got %q", i+1, c.expect, seq.Data())
		}
	}
}

func TestOscSequenceTerminator(t *testing.T) {
	cases := []struct {
		in     string
		expect string
	}{
		{"", ""},
		{"\x1b]1:hello\x07", "\x07"},
		{"\x1b]0;hello\x07", "\x07"},
		{"\x1b]0;hello\x1b\\", "\x1b\\"},
		{"\x1b]1234;hello\x1b\\", "\x1b\\"},
		{"\x1b]1234;abc;hello\x1b\\", "\x1b\\"},
		{"\x9b1234;hello\x9c", "\x9c"},
		{"\x9d]1234;hello\x9c", "\x9c"},
		{"\x9d1234;hello\x9c", "\x9c"},
	}

	for i, c := range cases {
		seq := OscSequence(c.in)
		if seq.Terminator() != c.expect {
			t.Errorf("case %d: expected %q, got %q", i+1, c.expect, seq.Terminator())
		}
	}
}
