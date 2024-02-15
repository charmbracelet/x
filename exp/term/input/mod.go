package input

// Mod represents modifier keys.
type Mod uint16

// Modifier keys.
const (
	Shift Mod = 1 << iota
	Alt
	Ctrl
	Meta
)

// IsShift reports whether the Shift modifier is set.
func (m Mod) IsShift() bool {
	return m&Shift != 0
}

// IsAlt reports whether the Alt modifier is set.
func (m Mod) IsAlt() bool {
	return m&Alt != 0
}

// IsCtrl reports whether the Ctrl modifier is set.
func (m Mod) IsCtrl() bool {
	return m&Ctrl != 0
}

// IsMeta reports whether the Meta modifier is set.
func (m Mod) IsMeta() bool {
	return m&Meta != 0
}
