package ansi

import (
	"fmt"
	"io"
)

// WriteITerm2 writes a sequence that uses the iTerm2 proprietary protocol to
// w. Use the iterm2 package for a more convenient API.
//
//	OSC 1337 ; key = value ST
//
// Example:
//
//	ITerm2(iterm2.File{...})
//
// See https://iterm2.com/documentation-escape-codes.html
// See https://iterm2.com/documentation-images.html
func WriteITerm2(w io.Writer, data any) (int, error) {
	return io.WriteString(w, "\x1b]1337;"+fmt.Sprint(data)+"\x07")
}

// ITerm2 returns a sequence that uses the iTerm2 proprietary protocol. Use the
// iterm2 package for a more convenient API.
//
//	OSC 1337 ; key = value ST
//
// Example:
//
//	ITerm2(iterm2.File{...})
//
// See https://iterm2.com/documentation-escape-codes.html
// See https://iterm2.com/documentation-images.html
func ITerm2(data any) string {
	return "\x1b]1337;" + fmt.Sprint(data) + "\x07"
}
