package ansi

import (
	"strings"
	"testing"
)

func TestASCIIBaseGraphemeWidths(t *testing.T) {
	const keycap = "1\ufe0f\u20e3"
	tests := []struct{ name, got, want string }{
		{"truncate", Truncate(strings.Repeat(keycap, 3)+"abc", 3, ""), keycap},
		{"truncate with tail", Truncate(strings.Repeat(keycap, 3)+"abc", 5, "…"), keycap + keycap + "…"},
		{"truncate left", TruncateLeft(keycap+keycap+"x", 2, ""), keycap + "x"},
		{"cut", Cut(keycap+keycap+"x", 2, 4), keycap},
		{"hardwrap", Hardwrap(strings.Repeat(keycap, 3), 4, true), keycap + keycap + "\n" + keycap},
		{"wrap", Wrap(strings.Repeat(keycap, 3), 4, ""), keycap + keycap + "\n" + keycap},
		{"wordwrap", Wordwrap(keycap+" "+keycap, 4, ""), keycap + "\n" + keycap},
		{"styled truncate", Truncate("\x1b[31m"+keycap+keycap+"\x1b[0m", 2, ""), "\x1b[31m" + keycap + "\x1b[0m"},
		{"combining truncate", Truncate("a\u0301b", 1, ""), "a\u0301"},
		{"adjacent non-ASCII", Truncate("a界b", 3, ""), "a界"},
		{"wcwidth truncate", TruncateWc(strings.Repeat(keycap, 3), 2, ""), keycap + keycap},
		{"wcwidth hardwrap", HardwrapWc(strings.Repeat(keycap, 3), 2, true), keycap + keycap + "\n" + keycap},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("got %q, want %q", tt.got, tt.want)
			}
		})
	}
}

func TestDecodeASCIIBaseGrapheme(t *testing.T) {
	for _, cluster := range []string{"1\ufe0f\u20e3", "#\ufe0f\u20e3", "*\ufe0f\u20e3", "a\u0301"} {
		t.Run(cluster, func(t *testing.T) {
			seq, width, n, state := DecodeSequence(cluster+"x", NormalState, nil)
			if seq != cluster || width != StringWidth(cluster) || n != len(cluster) || state != NormalState {
				t.Fatalf("got (%q,%d,%d,%d), want (%q,%d,%d,%d)", seq, width, n, state, cluster, StringWidth(cluster), len(cluster), NormalState)
			}
			bseq, bwidth, bn, bstate := DecodeSequence([]byte(cluster+"x"), NormalState, nil)
			if string(bseq) != seq || bwidth != width || bn != n || bstate != state {
				t.Fatalf("byte decoding differs: (%q,%d,%d,%d)", bseq, bwidth, bn, bstate)
			}
		})
	}
	seq, width, n, _ := DecodeSequenceWc("1\ufe0f\u20e3x", NormalState, nil)
	if seq != "1" || width != 1 || n != 1 {
		t.Fatalf("wcwidth ASCII behavior changed: (%q,%d,%d)", seq, width, n)
	}
}
