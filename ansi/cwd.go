package ansi

import (
	"net/url"
)

// NotifyWorkingDirectory returns a sequence for notifying the program's
// current working directory.
//
//	OSC 7 ; Pt BEL
//
// Where Pt is a URL in the format "file://[host]/[path]".
// Set host to "localhost" if this is a path on the local computer.
func NotifyWorkingDirectory(host string, path string) string {
	u := &url.URL{
		Scheme: "file",
		Host:   host,
		Path:   path,
	}
	return "\x1b]7;" + u.String() + "\x07"
}
