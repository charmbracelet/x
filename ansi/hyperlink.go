package ansi

import (
	"io"
	"strings"
)

// WriteSetHyperlink writes a sequence for starting a hyperlink to w.
//
//	OSC 8 ; Params ; Uri ST
//	OSC 8 ; Params ; Uri BEL
//
// To reset the hyperlink, omit the URI.
//
// See: https://gist.github.com/egmontkob/eb114294efbcd5adb1944c9f3cb5feda
func WriteSetHyperlink(w io.Writer, uri string, params ...string) (int, error) {
	if len(uri) == 0 {
		return io.WriteString(w, ResetHyperlink)
	}
	return io.WriteString(w, "\x1b]8;"+strings.Join(params, ":")+";"+uri+"\x07")
}

// SetHyperlink returns a sequence for starting a hyperlink.
//
//	OSC 8 ; Params ; Uri ST
//	OSC 8 ; Params ; Uri BEL
//
// To reset the hyperlink, omit the URI.
//
// See: https://gist.github.com/egmontkob/eb114294efbcd5adb1944c9f3cb5feda
func SetHyperlink(uri string, params ...string) string {
	if len(uri) == 0 {
		return ResetHyperlink
	}

	var p string
	if len(params) > 0 {
		p = strings.Join(params, ":")
	}
	return "\x1b]8;" + p + ";" + uri + "\x07"
}

// ResetHyperlink is a sequence for resetting the hyperlink.
//
// See: https://gist.github.com/egmontkob/eb114294efbcd5adb1944c9f3cb5feda
const ResetHyperlink = "\x1b]8;;\x07"
