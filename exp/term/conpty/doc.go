// Package conpty implements Windows Console Pseudo-terminal support.
//
// https://learn.microsoft.com/en-us/windows/console/creating-a-pseudoconsole-session

package conpty

import "errors"

// ErrUnsupported is returned when the current platform is not supported.
var ErrUnsupported = errors.New("conpty: unsupported platform")
