package ansi

import "testing"

func TestCsiSequenceIsValid(t *testing.T) {
	cases := []struct {
		seq   CsiSequence
		valid bool
	}{
		{CsiSequence(""), false},
		{CsiSequence("\x1b["), false},
		{CsiSequence("\x1b]"), false},
		{CsiSequence("\x9b"), false},
		{CsiSequence("\x1b[?1;2:1230"), false},
		{CsiSequence("\x1b[0A"), true},
		{CsiSequence("\x1b[A"), true},
		{CsiSequence("\x1b[ A"), true},
		{CsiSequence("\x1b[ #A"), true},
		{CsiSequence("\x1b[1 #A"), true},
		{CsiSequence("\x1b[1; #A"), true},
		{CsiSequence("\x1b[1;2 #A"), true},
		{CsiSequence("\x1b[1;2:3:4 #A"), true},
		{CsiSequence("\x1b[1;2:3:4: #["), true},
		{CsiSequence("\x1b[1;2;3;4;5;6;7;8;9A"), true},
		{CsiSequence("\x1b[?1;2A"), true},
		{CsiSequence("\x1b[?1;2:123A"), true},
	}
	for _, c := range cases {
		if got, want := c.seq.IsValid(), c.valid; got != want {
			t.Errorf("got %t, want %t", got, want)
		}
	}
}

func TestCsiSequenceParams(t *testing.T) {
	cases := []struct {
		seq    CsiSequence
		params string
	}{
		{CsiSequence("\x1b[012;3"), ""},
		{CsiSequence("\x1b[A"), ""},
		{CsiSequence("\x1b[0A"), "0"},
		{CsiSequence("\x1b[1;2;3;4;5;6;7;8;9A"), "1;2;3;4;5;6;7;8;9"},
		{CsiSequence("\x1b[?1;2A"), "?1;2"},
		{CsiSequence("\x1b[?1;2:123A"), "?1;2:123"},
	}
	for _, c := range cases {
		if got, want := string(c.seq.Params()), c.params; got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}
}

func TestCsiSequenceIntermediates(t *testing.T) {
	cases := []struct {
		seq          CsiSequence
		intermediate string
	}{
		{CsiSequence("\x1b[0A"), ""},
		{CsiSequence("\x1b[1;2;3;4;5;6;7;8;9A"), ""},
		{CsiSequence("\x1b[?1;2A"), ""},
		{CsiSequence("\x1b[?1;2:123A"), ""},
		{CsiSequence("\x1b[?1;2:123 A"), " "},
		{CsiSequence("\x1b[123 #!A"), " #!"},
	}
	for _, c := range cases {
		if got, want := string(c.seq.Intermediates()), c.intermediate; got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}
}

func TestCsiSequenceCommand(t *testing.T) {
	cases := []struct {
		seq     CsiSequence
		command byte
	}{
		{CsiSequence(""), 0},
		{CsiSequence("\x1b[0A"), 'A'},
		{CsiSequence("\x1b[1;2;3;4;5;6;7;8;9A"), 'A'},
		{CsiSequence("\x1b[?1;2A"), 'A'},
		{CsiSequence("\x1b[?1;2:123A"), 'A'},
	}
	for _, c := range cases {
		if got, want := c.seq.Command(), c.command; got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}
}
