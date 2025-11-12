package vcr

import (
	"path/filepath"
	"testing"

	"gopkg.in/dnaeon/go-vcr.v4/pkg/recorder"
)

// Recorder is an alias for the go-vcr Recorder.
type Recorder = recorder.Recorder

type options struct {
	mode           recorder.Mode
	keepAllHeaders bool
}

// Option defines a functional option for configuring the VCR recorder.
type Option func(*options) error

// WithMode sets the recorder mode.
func WithMode(mode recorder.Mode) Option {
	return func(o *options) error {
		o.mode = mode
		return nil
	}
}

// WithKeepAllHeaders configures the recorder to keep all HTTP headers.
func WithKeepAllHeaders() Option {
	return func(o *options) error {
		o.keepAllHeaders = true
		return nil
	}
}

// NewRecorder creates a new VCR recorder for the given test with the provided options.
func NewRecorder(t *testing.T, opts ...Option) *Recorder {
	o := options{
		mode:           recorder.ModeRecordOnce,
		keepAllHeaders: false,
	}
	for _, opt := range opts {
		if err := opt(&o); err != nil {
			t.Fatalf("vcr: failed to apply option: %v", err)
		}
	}

	cassetteName := filepath.Join("testdata", t.Name())

	r, err := recorder.New(
		cassetteName,
		recorder.WithMode(o.mode),
		recorder.WithMatcher(customMatcher(t)),
		recorder.WithMarshalFunc(customMarshaler),
		recorder.WithSkipRequestLatency(true), // disable sleep to simulate response time, makes tests faster
		recorder.WithHook(hookRemoveHeaders(o.keepAllHeaders), recorder.AfterCaptureHook),
	)
	if err != nil {
		t.Fatalf("vcr: failed to create recorder: %v", err)
	}

	t.Cleanup(func() {
		if err := r.Stop(); err != nil {
			t.Errorf("vcr: failed to stop recorder: %v", err)
		}
	})

	return r
}
