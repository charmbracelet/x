package ansi

import (
	"testing"
)

func TestKittyGraphics(t *testing.T) {
	tests := []struct {
		name    string
		payload []byte
		opts    []string
		want    string
	}{
		{
			name:    "empty payload no options",
			payload: []byte{},
			opts:    nil,
			want:    "\x1b_G\x1b\\",
		},
		{
			name:    "with payload no options",
			payload: []byte("test"),
			opts:    nil,
			want:    "\x1b_G;test\x1b\\",
		},
		{
			name:    "with payload and options",
			payload: []byte("test"),
			opts:    []string{"a=t", "f=100"},
			want:    "\x1b_Ga=t,f=100;test\x1b\\",
		},
		{
			name:    "multiple options no payload",
			payload: []byte{},
			opts:    []string{"q=2", "C=1", "f=24"},
			want:    "\x1b_Gq=2,C=1,f=24\x1b\\",
		},
		{
			name:    "with special characters in payload",
			payload: []byte("\x1b_G"),
			opts:    []string{"a=t"},
			want:    "\x1b_Ga=t;\x1b_G\x1b\\",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := KittyGraphics(tt.payload, tt.opts...)
			if got != tt.want {
				t.Errorf("KittyGraphics() = %q, want %q", got, tt.want)
			}
		})
	}
}
