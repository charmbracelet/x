package input

// ClipboardSelection represents a clipboard selection. The most common
// clipboard selections are "system" and "primary" and selections.
type ClipboardSelection byte

// Clipboard selections.
const (
	SystemClipboard  = ClipboardSelection('c')
	PrimaryClipboard = ClipboardSelection('p')
)

// ClipboardEvent is a clipboard read message event. This message is emitted when
// a terminal receives an OSC52 clipboard read message event.
type ClipboardEvent struct {
	Content   string
	Selection ClipboardSelection
}

// String returns the string representation of the clipboard message.
func (e ClipboardEvent) String() string {
	return e.Content
}
