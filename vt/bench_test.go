package vt

import (
	"bytes"
	"fmt"
	uv "github.com/charmbracelet/ultraviolet"
	"strings"
	"testing"
)

// benchInputs are typical terminal outputs: command output that scrolls
// (short and long lines), colored output, full-screen TUI redraws, and
// non-ASCII text.
var benchInputs = func() []struct {
	name string
	data []byte
} {
	var seq, text, color, tui, unicode bytes.Buffer
	for i := 1; i <= 20000; i++ {
		fmt.Fprintf(&seq, "%d\r\n", i)
	}
	line := strings.Repeat("abcdefghij", 8)
	for range 2000 {
		text.WriteString(line + "\r\n")
	}
	for i := range 2000 {
		fmt.Fprintf(&color, "\x1b[1;3%dmdrwxr-xr-x\x1b[0m  1 user  staff  \x1b[38;5;%dm%6d\x1b[0m Nov  2 12:00 \x1b[34mfile-%d.txt\x1b[0m\r\n", i%8, i%256, i*37, i)
	}
	for frame := range 100 {
		tui.WriteString("\x1b[H\x1b[2J")
		for row := range 24 {
			fmt.Fprintf(&tui, "\x1b[%d;1H\x1b[48;2;%d;%d;%dm\x1b[3%dm│ row %2d frame %3d %s │\x1b[0m", row+1, row*10, frame, 100, row%8, row, frame, strings.Repeat("█", 40))
		}
	}
	for range 1000 {
		unicode.WriteString("héllo wörld ✓ → 日本語 中文 한국어 e\u0301 👍🏽\r\n")
	}
	return []struct {
		name string
		data []byte
	}{
		{"seq", seq.Bytes()},
		{"text", text.Bytes()},
		{"color", color.Bytes()},
		{"tui", tui.Bytes()},
		{"unicode", unicode.Bytes()},
	}
}()

func BenchmarkWrite(b *testing.B) {
	for _, in := range benchInputs {
		b.Run(in.name, func(b *testing.B) {
			// One long-lived emulator, as a terminal session has: its scrollback
			// fills up after the first run.
			e := NewEmulator(80, 24)
			b.SetBytes(int64(len(in.data)))
			b.ReportAllocs()
			for b.Loop() {
				// Write in the chunks a pty read returns.
				for off := 0; off < len(in.data); off += 32 << 10 {
					_, _ = e.Write(in.data[off:min(off+32<<10, len(in.data))])
				}
			}
		})
	}
}

func BenchmarkDraw(b *testing.B) {
	e := NewSafeEmulator(80, 24)
	_, _ = e.Write(benchInputs[3].data) // a TUI screen
	buf := uv.NewScreenBuffer(80, 24)
	b.ReportAllocs()
	for b.Loop() {
		e.Draw(buf, buf.Bounds())
	}
}
