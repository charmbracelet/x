package ansi

import (
	"encoding/base64"
	"testing"

	"github.com/charmbracelet/x/ansi/iterm2"
)

func TestITerm2(t *testing.T) {
	tests := []struct {
		name string
		data any
		want string
	}{
		{
			name: "empty file",
			data: iterm2.File{},
			want: "\x1b]1337;File=\x07",
		},
		{
			name: "basic file",
			data: iterm2.File{
				Name: "test.png",
				Size: 1024,
			},
			want: "\x1b]1337;File=name=test.png;size=1024\x07",
		},
		{
			name: "file with dimensions",
			data: iterm2.File{
				Name:   "test.png",
				Width:  iterm2.Pixels(100),
				Height: iterm2.Auto,
			},
			want: "\x1b]1337;File=name=test.png;width=100px;height=auto\x07",
		},
		{
			name: "file with all options",
			data: iterm2.File{
				Name:              "test.png",
				Size:              1024,
				Width:             iterm2.Cells(100),
				Height:            iterm2.Percent(50),
				IgnoreAspectRatio: true,
				Inline:            true,
				DoNotMoveCursor:   true,
			},
			want: "\x1b]1337;File=name=test.png;size=1024;width=100;height=50%;preserveAspectRatio=0;inline=1;doNotMoveCursor=1\x07",
		},
		{
			name: "file with content",
			data: iterm2.File{
				Name:    "test.png",
				Content: []byte(base64.StdEncoding.EncodeToString([]byte("test-content"))),
			},
			want: "\x1b]1337;File=name=test.png:dGVzdC1jb250ZW50\x07",
		},
		{
			name: "multipart file",
			data: iterm2.MultipartFile{
				Name:   "test.png",
				Size:   1024,
				Width:  iterm2.Pixels(100),
				Height: iterm2.Percent(50),
			},
			want: "\x1b]1337;MultipartFile=name=test.png;size=1024;width=100px;height=50%\x07",
		},
		{
			name: "file part",
			data: iterm2.FilePart{
				Content: []byte("part-content"),
			},
			want: "\x1b]1337;FilePart=part-content\x07",
		},
		{
			name: "file end",
			data: iterm2.FileEnd{},
			want: "\x1b]1337;FileEnd\x07",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ITerm2(tt.data); got != tt.want {
				t.Errorf("ITerm2() = %v, want %v", got, tt.want)
			}
		})
	}
}
