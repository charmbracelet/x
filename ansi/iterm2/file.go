// Package iterm2 provides iTerm2-specific functionality.
package iterm2

import (
	"strconv"
	"strings"
)

// Auto is a constant that represents the "auto" value.
const Auto = "auto"

// Cells returns a string that represents the number of cells. This is simply a
// wrapper around strconv.Itoa.
func Cells(n int) string {
	return strconv.Itoa(n)
}

// Pixels returns a string that represents the number of pixels.
func Pixels(n int) string {
	return strconv.Itoa(n) + "px"
}

// Percent returns a string that represents the percentage.
func Percent(n int) string {
	return strconv.Itoa(n) + "%"
}

// file represents the optional arguments for the iTerm2 Inline Image Protocol.
//
// See https://iterm2.com/documentation-images.html
type file struct {
	// Name is the name of the file. Defaults to "Unnamed file" if empty.
	Name string
	// Size is the file size in bytes. Used for progress indication. This is
	// optional.
	Size int64
	// Width is the width of the image. This can be specified by a number
	// followed by by a unit or "auto". The unit can be none, "px" or "%". None
	// means the number is in cells. Defaults to "auto" if empty.
	// For convenience, the [Pixels], [Cells] and [Percent] functions and
	// [Auto] can be used.
	Width string
	// Height is the height of the image. This can be specified by a number
	// followed by by a unit or "auto". The unit can be none, "px" or "%". None
	// means the number is in cells. Defaults to "auto" if empty.
	// For convenience, the [Pixels], [Cells] and [Percent] functions and
	// [Auto] can be used.
	Height string
	// IgnoreAspectRatio is a flag that indicates that the image should be
	// stretched to fit the specified width and height. Defaults to false
	// meaning the aspect ratio is preserved.
	IgnoreAspectRatio bool
	// Inline is a flag that indicates that the image should be displayed
	// inline. Otherwise, it is downloaded to the Downloads folder with no
	// visual representation in the terminal. Defaults to false.
	Inline bool
	// DoNotMoveCursor is a flag that indicates that the cursor should not be
	// moved after displaying the image. This is an extension introduced by
	// WezTerm and might not work on all terminals supporting the iTerm2
	// protocol. Defaults to false.
	DoNotMoveCursor bool
	// Content is the base64 encoded data of the file.
	Content []byte
}

// String implements fmt.Stringer.
func (f file) String() string {
	var opts []string
	if f.Name != "" {
		opts = append(opts, "name="+f.Name)
	}
	if f.Size != 0 {
		opts = append(opts, "size="+strconv.FormatInt(f.Size, 10))
	}
	if f.Width != "" {
		opts = append(opts, "width="+f.Width)
	}
	if f.Height != "" {
		opts = append(opts, "height="+f.Height)
	}
	if f.IgnoreAspectRatio {
		opts = append(opts, "preserveAspectRatio=0")
	}
	if f.Inline {
		opts = append(opts, "inline=1")
	}
	if f.DoNotMoveCursor {
		opts = append(opts, "doNotMoveCursor=1")
	}
	return strings.Join(opts, ";")
}

// File represents the optional arguments for the iTerm2 Inline Image Protocol.
type File file

// String implements fmt.Stringer.
func (f File) String() string {
	var s strings.Builder
	s.WriteString("File=")
	s.WriteString(file(f).String())
	if len(f.Content) > 0 {
		s.WriteString(":")
		s.Write(f.Content)
	}

	return s.String()
}

// MultipartFile represents the optional arguments for the iTerm2 Inline Image Protocol.
type MultipartFile file

// String implements fmt.Stringer.
func (f MultipartFile) String() string {
	return "MultipartFile=" + file(f).String()
}

// FilePart represents the optional arguments for the iTerm2 Inline Image Protocol.
type FilePart file

// String implements fmt.Stringer.
func (f FilePart) String() string {
	return "FilePart=" + string(f.Content)
}

// FileEnd represents the optional arguments for the iTerm2 Inline Image Protocol.
type FileEnd struct{}

// String implements fmt.Stringer.
func (f FileEnd) String() string {
	return "FileEnd"
}
