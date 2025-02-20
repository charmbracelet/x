package sixel

import (
	"bytes"
	"image/color"
	"testing"
)

func TestWriteColor(t *testing.T) {
	tests := []struct {
		name     string
		pc       int
		pu       int
		px       int
		py       int
		pz       int
		expected string
	}{
		{
			name:     "simple color number",
			pc:       1,
			pu:       0,
			expected: "#1",
		},
		{
			name:     "RGB color",
			pc:       1,
			pu:       2,
			px:       50,
			py:       60,
			pz:       70,
			expected: "#1;2;50;60;70",
		},
		{
			name:     "HLS color",
			pc:       2,
			pu:       1,
			px:       180,
			py:       50,
			pz:       100,
			expected: "#2;1;180;50;100",
		},
		{
			name:     "invalid pu > 2",
			pc:       1,
			pu:       3,
			expected: "#1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			n, err := WriteColor(buf, tt.pc, tt.pu, tt.px, tt.py, tt.pz)
			if err != nil {
				t.Errorf("WriteColor() unexpected error = %v", err)
				return
			}
			if got := buf.String(); got != tt.expected {
				t.Errorf("WriteColor() = %v, want %v", got, tt.expected)
			}
			if n != len(tt.expected) {
				t.Errorf("WriteColor() returned length = %v, want %v", n, len(tt.expected))
			}
		})
	}
}

func TestDecodeColor(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		wantC Color
		wantN int
	}{
		{
			name:  "simple color number",
			input: []byte("#1"),
			wantC: Color{Pc: 1},
			wantN: 2,
		},
		{
			name:  "RGB color",
			input: []byte("#1;2;50;60;70"),
			wantC: Color{Pc: 1, Pu: 2, Px: 50, Py: 60, Pz: 70},
			wantN: 13,
		},
		{
			name:  "HLS color",
			input: []byte("#2;1;180;50;100"),
			wantC: Color{Pc: 2, Pu: 1, Px: 180, Py: 50, Pz: 100},
			wantN: 15,
		},
		{
			name:  "empty input",
			input: []byte{},
			wantC: Color{},
			wantN: 0,
		},
		{
			name:  "invalid introducer",
			input: []byte("X1"),
			wantC: Color{},
			wantN: 0,
		},
		{
			name:  "incomplete sequence",
			input: []byte("#"),
			wantC: Color{},
			wantN: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotC, gotN := DecodeColor(tt.input)
			if gotC != tt.wantC {
				t.Errorf("DecodeColor() gotColor = %v, want %v", gotC, tt.wantC)
			}
			if gotN != tt.wantN {
				t.Errorf("DecodeColor() gotN = %v, want %v", gotN, tt.wantN)
			}
		})
	}
}

func TestColor_RGBA(t *testing.T) {
	tests := []struct {
		name  string
		color Color
		wantR uint32
		wantG uint32
		wantB uint32
		wantA uint32
	}{
		{
			name:  "default color map 0 (black)",
			color: Color{Pc: 0},
			wantR: 0x0000,
			wantG: 0x0000,
			wantB: 0x0000,
			wantA: 0xFFFF,
		},
		{
			name:  "RGB mode (50%, 60%, 70%)",
			color: Color{Pc: 1, Pu: 2, Px: 50, Py: 60, Pz: 70},
			wantR: 0x8080,
			wantG: 0x9999,
			wantB: 0xB3B3,
			wantA: 0xFFFF,
		},
		{
			name:  "HLS mode (180°, 50%, 100%)",
			color: Color{Pc: 1, Pu: 1, Px: 180, Py: 50, Pz: 100},
			wantR: 0x0000,
			wantG: 0xFFFF,
			wantB: 0xFFFF,
			wantA: 0xFFFF,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotR, gotG, gotB, gotA := tt.color.RGBA()
			if gotR != tt.wantR {
				t.Errorf("Color.RGBA() gotR = %v, want %v", gotR, tt.wantR)
			}
			if gotG != tt.wantG {
				t.Errorf("Color.RGBA() gotG = %v, want %v", gotG, tt.wantG)
			}
			if gotB != tt.wantB {
				t.Errorf("Color.RGBA() gotB = %v, want %v", gotB, tt.wantB)
			}
			if gotA != tt.wantA {
				t.Errorf("Color.RGBA() gotA = %v, want %v", gotA, tt.wantA)
			}
		})
	}
}

func TestSixelRGB(t *testing.T) {
	tests := []struct {
		name string
		r    int
		g    int
		b    int
		want color.Color
	}{
		{
			name: "black",
			r:    0,
			g:    0,
			b:    0,
			want: color.NRGBA{R: 0, G: 0, B: 0, A: 255},
		},
		{
			name: "white",
			r:    100,
			g:    100,
			b:    100,
			want: color.NRGBA{R: 255, G: 255, B: 255, A: 255},
		},
		{
			name: "red",
			r:    100,
			g:    0,
			b:    0,
			want: color.NRGBA{R: 255, G: 0, B: 0, A: 255},
		},
		{
			name: "half intensity",
			r:    50,
			g:    50,
			b:    50,
			want: color.NRGBA{R: 128, G: 128, B: 128, A: 255},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sixelRGB(tt.r, tt.g, tt.b)
			gotR, gotG, gotB, gotA := got.RGBA()
			wantR, wantG, wantB, wantA := tt.want.RGBA()
			if gotR != wantR || gotG != wantG || gotB != wantB || gotA != wantA {
				t.Errorf("sixelRGB() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSixelHLS(t *testing.T) {
	tests := []struct {
		name string
		h    int
		l    int
		s    int
		want color.Color
	}{
		{
			name: "black",
			h:    0,
			l:    0,
			s:    0,
			want: color.NRGBA{R: 0, G: 0, B: 0, A: 255},
		},
		{
			name: "white",
			h:    0,
			l:    100,
			s:    0,
			want: color.NRGBA{R: 255, G: 255, B: 255, A: 255},
		},
		{
			name: "pure red",
			h:    0,
			l:    50,
			s:    100,
			want: color.NRGBA{R: 255, G: 0, B: 0, A: 255},
		},
		{
			name: "pure green",
			h:    120,
			l:    50,
			s:    100,
			want: color.NRGBA{R: 0, G: 255, B: 0, A: 255},
		},
		{
			name: "pure blue",
			h:    240,
			l:    50,
			s:    100,
			want: color.NRGBA{R: 0, G: 0, B: 255, A: 255},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sixelHLS(tt.h, tt.l, tt.s)
			gotR, gotG, gotB, gotA := got.RGBA()
			wantR, wantG, wantB, wantA := tt.want.RGBA()
			if gotR != wantR || gotG != wantG || gotB != wantB || gotA != wantA {
				t.Errorf("sixelHLS() = %v, want %v", got, tt.want)
			}
		})
	}
}
