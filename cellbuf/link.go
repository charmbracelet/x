package cellbuf

import (
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/vt"
)

// Link represents a hyperlink in the terminal screen.
type Link = vt.Link

// Convert converts a hyperlink to respect the given color profile.
func ConvertLink(h Link, p colorprofile.Profile) Link {
	if p == colorprofile.NoTTY {
		return Link{}
	}

	return h
}
