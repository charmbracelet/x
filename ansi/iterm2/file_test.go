package iterm2

import (
	"encoding/base64"
	"testing"
)

func TestCells(t *testing.T) {
	tests := []struct {
		input int
		want  string
	}{
		{0, "0"},
		{10, "10"},
		{-5, "-5"},
		{100, "100"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := Cells(tt.input); got != tt.want {
				t.Errorf("Cells(%d) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestPixels(t *testing.T) {
	tests := []struct {
		input int
		want  string
	}{
		{0, "0px"},
		{10, "10px"},
		{-5, "-5px"},
		{100, "100px"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := Pixels(tt.input); got != tt.want {
				t.Errorf("Pixels(%d) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestPercent(t *testing.T) {
	tests := []struct {
		input int
		want  string
	}{
		{0, "0%"},
		{10, "10%"},
		{-5, "-5%"},
		{100, "100%"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := Percent(tt.input); got != tt.want {
				t.Errorf("Percent(%d) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestFile_String(t *testing.T) {
	sampleContent := []byte("test-content")
	tests := []struct {
		name string
		file file
		want string
	}{
		{
			name: "empty file",
			file: file{},
			want: "",
		},
		{
			name: "basic file",
			file: file{
				Name: "test.png",
				Size: 1024,
			},
			want: "name=test.png;size=1024",
		},
		{
			name: "file with dimensions",
			file: file{
				Name:   "test.png",
				Width:  "100px",
				Height: "auto",
			},
			want: "name=test.png;width=100px;height=auto",
		},
		{
			name: "file with all options",
			file: file{
				Name:              "test.png",
				Size:              1024,
				Width:             "100px",
				Height:            "50%",
				IgnoreAspectRatio: true,
				Inline:            true,
				DoNotMoveCursor:   true,
				Content:           sampleContent,
			},
			want: "name=test.png;size=1024;width=100px;height=50%;preserveAspectRatio=0;inline=1;doNotMoveCursor=1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.file.String(); got != tt.want {
				t.Errorf("file.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFile_String_WithContent(t *testing.T) {
	sampleContent := []byte("test-content")
	encodedContent := base64.StdEncoding.EncodeToString(sampleContent)

	f := File{
		Name:    "test.png",
		Content: []byte(encodedContent),
	}

	want := "File=name=test.png:" + encodedContent
	if got := f.String(); got != want {
		t.Errorf("File.String() = %v, want %v", got, want)
	}
}

func TestMultipartFile_String(t *testing.T) {
	f := MultipartFile{
		Name:   "test.png",
		Size:   1024,
		Width:  "100px",
		Height: "50%",
	}

	want := "MultipartFile=name=test.png;size=1024;width=100px;height=50%"
	if got := f.String(); got != want {
		t.Errorf("MultipartFile.String() = %v, want %v", got, want)
	}
}

func TestFilePart_String(t *testing.T) {
	sampleContent := []byte("test-content")
	f := FilePart{
		Content: sampleContent,
	}

	want := "FilePart=" + string(sampleContent)
	if got := f.String(); got != want {
		t.Errorf("FilePart.String() = %v, want %v", got, want)
	}
}

func TestFileEnd_String(t *testing.T) {
	f := FileEnd{}
	want := "FileEnd"
	if got := f.String(); got != want {
		t.Errorf("FileEnd.String() = %v, want %v", got, want)
	}
}

func TestAuto_Constant(t *testing.T) {
	if Auto != "auto" {
		t.Errorf("Auto constant = %v, want 'auto'", Auto)
	}
}
