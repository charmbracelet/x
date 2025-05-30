package vt

// Logger represents a logger interface.
type Logger interface {
	Printf(format string, v ...any)
}

// Option is a terminal option.
type Option func(*Terminal)

// WithLogger returns an [Option] that sets the terminal's logger.
// The logger is used for debugging and logging.
// By default, the terminal does not log anything.
//
// Example:
//
//	vterm := vt.NewTerminal(80, 24, vt.WithLogger(log.Default()))
func WithLogger(logger Logger) Option {
	return func(t *Terminal) {
		t.logger = logger
	}
}

// logf logs a formatted message if the terminal has a logger.
func (t *Terminal) logf(format string, v ...any) {
	if t.logger != nil {
		t.logger.Printf(format, v...)
	}
}
