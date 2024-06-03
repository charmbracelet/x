package input

// KeyMod represents modifier keys.
type KeyMod uint16

// Modifier keys.
const (
	Shift KeyMod = 1 << iota
	Alt
	Ctrl
	Meta

	// These modifiers are used with the Kitty protocol.
	// XXX: Meta and Super are swapped in the Kitty protocol,
	// this is to preserve compatibility with XTerm modifiers.

	Hyper
	Super // Windows/Command keys

	// These are key lock states.

	CapsLock
	NumLock
	ScrollLock // Defined in Windows API only
)

// HasShift reports whether the Shift modifier is set.
func (m KeyMod) HasShift() bool {
	return m&Shift != 0
}

// HasAlt reports whether the Alt modifier is set.
func (m KeyMod) HasAlt() bool {
	return m&Alt != 0
}

// HasCtrl reports whether the Ctrl modifier is set.
func (m KeyMod) HasCtrl() bool {
	return m&Ctrl != 0
}

// HasMeta reports whether the Meta modifier is set.
func (m KeyMod) HasMeta() bool {
	return m&Meta != 0
}

// HasHyper reports whether the Hyper modifier is set.
func (m KeyMod) HasHyper() bool {
	return m&Hyper != 0
}

// HasSuper reports whether the Super modifier is set.
func (m KeyMod) HasSuper() bool {
	return m&Super != 0
}

// HasCapsLock reports whether the CapsLock key is enabled.
func (m KeyMod) HasCapsLock() bool {
	return m&CapsLock != 0
}

// HasNumLock reports whether the NumLock key is enabled.
func (m KeyMod) HasNumLock() bool {
	return m&NumLock != 0
}

// HasScrollLock reports whether the ScrollLock key is enabled.
func (m KeyMod) HasScrollLock() bool {
	return m&ScrollLock != 0
}
