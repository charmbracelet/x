package vt

import "log"

// Option is a terminal option.
type Option func(*Terminal)

// WithLogger returns an [Option] that sets the terminal's logger.
// The logger is used for debugging and logging.
// By default, the terminal does not log anything.
func WithLogger(logger *log.Logger) Option {
	return func(t *Terminal) {
		t.logger = logger
	}
}

// log logs a message if the terminal has a logger.
func (t *Terminal) log(msg string) {
	if t.logger != nil {
		t.logger.Print(msg)
	}
}

// logln logs a message if the terminal has a logger.
func (t *Terminal) logln(msg string) {
	if t.logger != nil {
		t.logger.Println(msg)
	}
}

// logf logs a formatted message if the terminal has a logger.
func (t *Terminal) logf(format string, v ...interface{}) {
	if t.logger != nil {
		t.logger.Printf(format, v...)
	}
}
