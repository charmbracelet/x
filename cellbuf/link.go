package cellbuf

import "github.com/charmbracelet/colorprofile"

// Link represents a hyperlink in the terminal screen.
type Link struct {
	URL   string
	URLID string
}

// String returns a string representation of the hyperlink.
func (h Link) String() string {
	return h.URL
}

// Reset resets the hyperlink to the default state zero value.
func (h *Link) Reset() {
	h.URL = ""
	h.URLID = ""
}

// Equal returns whether the hyperlink is equal to the other hyperlink.
func (h Link) Equal(o Link) bool {
	return h == o
}

// Empty returns whether the hyperlink is empty.
func (h Link) Empty() bool {
	return h.URL == "" && h.URLID == ""
}

// Convert converts a hyperlink to respect the given color profile.
func (h Link) Convert(p colorprofile.Profile) Link {
	if p == colorprofile.NoTTY {
		return Link{}
	}

	return h
}
