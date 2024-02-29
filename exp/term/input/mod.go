package input

// Mod represents modifier keys.
type Mod uint16

// Modifier keys.
const (
	Shift Mod = 1 << iota
	Alt
	Ctrl
	Meta

	// These modifiers are used with the Kitty protocol.
	// XXX: Meta and Super are swapped in the Kitty protocol,
	// this is to preserve compatibility with XTerm modifiers.

	Hyper
	Super // Windows/Command keys
	CapsLock
	NumLock
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

// IsHyper reports whether the Hyper modifier is set.
func (m Mod) IsHyper() bool {
	return m&Hyper != 0
}

// IsSuper reports whether the Super modifier is set.
func (m Mod) IsSuper() bool {
	return m&Super != 0
}

// IsCapsLock reports whether the CapsLock modifier is set.
func (m Mod) IsCapsLock() bool {
	return m&CapsLock != 0
}

// IsNumLock reports whether the NumLock modifier is set.
func (m Mod) IsNumLock() bool {
	return m&NumLock != 0
}
