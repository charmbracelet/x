package ansi

// Method is a type that represents the how to calculate the cell widths in the
// terminal. The default is to use [WcWidth]. Some terminals use grapheme
// clustering by default. Some support mode 2027 to tell the terminal to use
// mode 2027 instead of wcwidth.
type Method uint8

// Display width modes.
const (
	WcWidth Method = iota
	GraphemeWidth
)

// String returns the string representation of the Method.
func (m Method) String() string {
	switch m {
	case WcWidth:
		return "WcWidth"
	case GraphemeWidth:
		return "GraphemeWidth"
	default:
		return "Unknown"
	}
}
