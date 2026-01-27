package ansi

import (
	"io"
	"testing"
)

func BenchmarkWritePrimaryDeviceAttributes(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = WritePrimaryDeviceAttributes(io.Discard, 1, 4, 18)
	}
}

func BenchmarkPrimaryDeviceAttributes(b *testing.B) {
	for i := 0; i < b.N; i++ {
		io.WriteString(io.Discard, PrimaryDeviceAttributes(1, 4, 18))
	}
}

func BenchmarkWriteSecondaryDeviceAttributes(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = WriteSecondaryDeviceAttributes(io.Discard, 1, 2, 3, 4)
	}
}

func BenchmarkSecondaryDeviceAttributes(b *testing.B) {
	for i := 0; i < b.N; i++ {
		io.WriteString(io.Discard, SecondaryDeviceAttributes(1, 2, 3, 4))
	}
}

func BenchmarkWriteTertiaryDeviceAttributes(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = WriteTertiaryDeviceAttributes(io.Discard, "TERM-1234")
	}
}

func BenchmarkTertiaryDeviceAttributes(b *testing.B) {
	for i := 0; i < b.N; i++ {
		io.WriteString(io.Discard, TertiaryDeviceAttributes("TERM-1234"))
	}
}

func TestPrimaryDeviceAttributes(t *testing.T) {
	tt := []struct {
		attrs  []int
		expect string
	}{
		{[]int{}, "\x1b[c"},
		{[]int{0}, "\x1b[c"},
		{[]int{1}, "\x1b[?1c"},
		{[]int{1, 4}, "\x1b[?1;4c"},
		{[]int{1, 4, 18}, "\x1b[?1;4;18c"},
	}
	for _, tp := range tt {
		pda := PrimaryDeviceAttributes(tp.attrs...)
		if pda != tp.expect {
			t.Errorf("PrimaryDeviceAttributes(%v) = %q, want %q", tp.attrs, pda, tp.expect)
		}
	}
}

func TestSecondaryDeviceAttributes(t *testing.T) {
	tt := []struct {
		attrs  []int
		expect string
	}{
		{[]int{}, "\x1b[>c"},
		{[]int{0}, "\x1b[>c"},
		{[]int{1}, "\x1b[>1c"},
		{[]int{1, 2}, "\x1b[>1;2c"},
		{[]int{1, 2, 3}, "\x1b[>1;2;3c"},
		{[]int{1, 2, 3, 4}, "\x1b[>1;2;3;4c"},
	}
	for _, tp := range tt {
		sda := SecondaryDeviceAttributes(tp.attrs...)
		if sda != tp.expect {
			t.Errorf("SecondaryDeviceAttributes(%v) = %q, want %q", tp.attrs, sda, tp.expect)
		}
	}
}

func TestTertiaryDeviceAttributes(t *testing.T) {
	tt := []struct {
		unitID string
		expect string
	}{
		{"", "\x1b[=c"},
		{"0", "\x1b[=c"},
		{"TERM-1234", "\x1bP!|TERM-1234\x1b\\"},
	}
	for _, tp := range tt {
		tda := TertiaryDeviceAttributes(tp.unitID)
		if tda != tp.expect {
			t.Errorf("TertiaryDeviceAttributes(%q) = %q, want %q", tp.unitID, tda, tp.expect)
		}
	}
}
