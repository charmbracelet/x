package vt

// ModeSetting represents a mode setting.
type ModeSetting int

// ModeSetting constants.
const (
	ModeNotRecognized ModeSetting = iota
	ModeSet
	ModeReset
	ModePermanentlySet
	ModePermanentlyReset
)

// IsSet returns true if the mode is set or permanently set.
func (m ModeSetting) IsSet() bool {
	return m == ModeSet || m == ModePermanentlySet
}

// IsReset returns true if the mode is reset or permanently reset.
func (m ModeSetting) IsReset() bool {
	return m == ModeReset || m == ModePermanentlyReset
}

// IsPermanentlySet returns true if the mode is permanently set.
func (m ModeSetting) IsPermanentlySet() bool {
	return m == ModePermanentlySet
}

// IsPermanentlyReset returns true if the mode is permanently reset.
func (m ModeSetting) IsPermanentlyReset() bool {
	return m == ModePermanentlyReset
}
