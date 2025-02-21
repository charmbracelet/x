package sixel

import (
	"bytes"
	"testing"
)

func TestWriteRaster(t *testing.T) {
	tests := []struct {
		name    string
		pan     int
		pad     int
		ph      int
		pv      int
		want    string
		wantErr bool
	}{
		{
			name: "basic case",
			pan:  1,
			pad:  2,
			want: "\"1;2",
		},
		{
			name: "with ph and pv",
			pan:  2,
			pad:  3,
			ph:   4,
			pv:   5,
			want: "\"2;3;4;5",
		},
		{
			name: "zero pad converts to 1,1",
			pan:  2,
			pad:  0,
			want: "\"1;1",
		},
		{
			name: "with ph only",
			pan:  1,
			pad:  2,
			ph:   3,
			pv:   0,
			want: "\"1;2;3;0",
		},
		{
			name: "with pv only",
			pan:  1,
			pad:  2,
			ph:   0,
			pv:   3,
			want: "\"1;2;0;3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			n, err := WriteRaster(&buf, tt.pan, tt.pad, tt.ph, tt.pv)
			if (err != nil) != tt.wantErr {
				t.Errorf("WriteRaster() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got := buf.String(); got != tt.want {
				t.Errorf("WriteRaster() = %q, want %q", got, tt.want)
			}
			if n != len(tt.want) {
				t.Errorf("WriteRaster() returned length %d, want %d", n, len(tt.want))
			}
		})
	}
}

func TestRaster_WriteTo(t *testing.T) {
	tests := []struct {
		name    string
		raster  Raster
		want    string
		wantErr bool
	}{
		{
			name:   "basic case",
			raster: Raster{Pan: 1, Pad: 2},
			want:   "\"1;2",
		},
		{
			name:   "full attributes",
			raster: Raster{Pan: 2, Pad: 3, Ph: 4, Pv: 5},
			want:   "\"2;3;4;5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			n, err := tt.raster.WriteTo(&buf)
			if (err != nil) != tt.wantErr {
				t.Errorf("Raster.WriteTo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got := buf.String(); got != tt.want {
				t.Errorf("Raster.WriteTo() = %q, want %q", got, tt.want)
			}
			if n != int64(len(tt.want)) {
				t.Errorf("Raster.WriteTo() returned length %d, want %d", n, len(tt.want))
			}
		})
	}
}

func TestDecodeRaster(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		want     Raster
		wantRead int
	}{
		{
			name:     "basic case",
			input:    "\"1;2",
			want:     Raster{Pan: 1, Pad: 2},
			wantRead: 4,
		},
		{
			name:     "full attributes",
			input:    "\"2;3;4;5",
			want:     Raster{Pan: 2, Pad: 3, Ph: 4, Pv: 5},
			wantRead: 8,
		},
		{
			name:     "empty input",
			input:    "",
			want:     Raster{},
			wantRead: 0,
		},
		{
			name:     "invalid start character",
			input:    "x1;2",
			want:     Raster{},
			wantRead: 0,
		},
		{
			name:     "too short",
			input:    "\"1",
			want:     Raster{Pan: 1},
			wantRead: 2,
		},
		{
			name:     "invalid character",
			input:    "\"1;a",
			want:     Raster{Pan: 1},
			wantRead: 3,
		},
		{
			name:     "partial attributes",
			input:    "\"1;2;3",
			want:     Raster{Pan: 1, Pad: 2, Ph: 3},
			wantRead: 6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, n := DecodeRaster([]byte(tt.input))
			if got != tt.want {
				t.Errorf("DecodeRaster() = %+v, want %+v", got, tt.want)
			}
			if n != tt.wantRead {
				t.Errorf("DecodeRaster() read = %d, want %d", n, tt.wantRead)
			}
		})
	}
}

func TestRaster_String(t *testing.T) {
	tests := []struct {
		name   string
		raster Raster
		want   string
	}{
		{
			name:   "basic case",
			raster: Raster{Pan: 1, Pad: 2},
			want:   "\"1;2",
		},
		{
			name:   "full attributes",
			raster: Raster{Pan: 2, Pad: 3, Ph: 4, Pv: 5},
			want:   "\"2;3;4;5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.raster.String(); got != tt.want {
				t.Errorf("Raster.String() = %q, want %q", got, tt.want)
			}
		})
	}
}
