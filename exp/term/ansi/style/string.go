package style

// String returns a styled string with the given SGR attributes applied.
func String(s string, attrs ...string) string {
	if len(attrs) == 0 {
		return s
	}
	return Sequence(attrs...) + s + ResetSequence
}
