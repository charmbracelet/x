package errors

import "strings"

// Join returns an error that wraps the given errors.
// Any nil error values are discarded.
// Join returns nil if every value in errs is nil.
// The error formats as the concatenation of the strings obtained
// by calling the Error method of each element of errs, with a newline
// between each string.
//
// A non-nil error returned by Join implements the Unwrap() []error method.
//
// This is copied from Go 1.20 errors.Unwrap, with some tuning to avoid using unsafe.
// The main goal is to have this available in older Go versions.
func Join(errs ...error) error {
	var nonNil []error
	for _, err := range errs {
		if err == nil {
			continue
		}
		nonNil = append(nonNil, err)
	}
	if len(nonNil) == 0 {
		return nil
	}
	return &joinError{
		errs: nonNil,
	}
}

type joinError struct {
	errs []error
}

func (e *joinError) Error() string {
	strs := make([]string, 0, len(e.errs))
	for _, err := range e.errs {
		strs = append(strs, err.Error())
	}
	return strings.Join(strs, "\n")
}

func (e *joinError) Unwrap() []error {
	return e.errs
}
