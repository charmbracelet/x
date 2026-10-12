package kitty

import "testing"

func TestDiacritic(t *testing.T) {
	tests := []struct {
		name  string
		index int
		want  rune
	}{
		{"first", 0, '\u0305'},
		{"second", 1, '\u030D'},
		{"last", len(diacritics) - 1, '\U0001D244'},
		{"negative", -1, '\u0305'},
		{"past end", len(diacritics), '\u0305'},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Diacritic(tt.index); got != tt.want {
				t.Fatalf("Diacritic(%d) = %U, want %U", tt.index, got, tt.want)
			}
		})
	}
}
