package charmtone

import (
	"strconv"
	"strings"
	"testing"
)

func TestValidateHexes(t *testing.T) {
	for _, key := range Keys() {
		hex := strings.TrimPrefix(key.Hex(), "#")
		if len(hex) != 6 && len(hex) != 3 {
			t.Errorf("Key %s: invalid hex length %d for %s", key, len(hex), key.Hex())
		}
		if _, err := strconv.ParseUint(hex, 16, 32); err != nil {
			t.Errorf("Key %s: invalid hex value %s", key, key.Hex())
		}
	}
}

func TestNameFromHex(t *testing.T) {
	for _, key := range Keys() {
		name := NameFromHex(key.Hex())
		if name != key.String() {
			t.Errorf("NameFromHex(%q) = %q, want %q", key.Hex(), name, key.String())
		}
	}
}

func TestNameFromHex_CaseInsensitive(t *testing.T) {
	tests := []struct {
		hex  string
		want string
	}{
		{"#6b50ff", "Charple"},
		{"#6B50FF", "Charple"},
		{"#6B50Ff", "Charple"},
		{"#201f26", "Pepper"},
		{"#ecebf0", "Sash"},
	}
	for _, tt := range tests {
		got := NameFromHex(tt.hex)
		if got != tt.want {
			t.Errorf("NameFromHex(%q) = %q, want %q", tt.hex, got, tt.want)
		}
	}
}

func TestNameFromHex_NoMatch(t *testing.T) {
	tests := []string{
		"#000000",
		"#FFFFFF",
		"",
		"not-a-hex",
		"#FFF",
	}
	for _, hex := range tests {
		if got := NameFromHex(hex); got != "" {
			t.Errorf("NameFromHex(%q) = %q, want empty", hex, got)
		}
	}
}
