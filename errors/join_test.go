package errors

import (
	"errors"
	"fmt"
	"testing"
)

func TestJoin(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		err := Join(nil, nil, nil)
		if err != nil {
			t.Errorf("expected nil, got %v", err)
		}
	})
	t.Run("one err", func(t *testing.T) {
		expected := fmt.Errorf("fake")
		err := Join(nil, expected, nil)
		if !errors.Is(err, expected) {
			t.Errorf("expected %v, got %v", expected, err)
		}
		if s := err.Error(); s != expected.Error() {
			t.Errorf("expected %s, got %s", expected, err)
		}
	})
	t.Run("many errs", func(t *testing.T) {
		expected1 := fmt.Errorf("fake 1")
		expected2 := fmt.Errorf("fake 2")
		err := Join(nil, expected1, nil, nil, expected2, nil)
		if !errors.Is(err, expected1) {
			t.Errorf("expected %v, got %v", expected1, err)
		}
		if !errors.Is(err, expected2) {
			t.Errorf("expected %v, got %v", expected2, err)
		}
		expectedS := expected1.Error() + "\n" + expected2.Error()
		if s := err.Error(); s != expectedS {
			t.Errorf("expected %s, got %s", expectedS, err)
		}
	})
}
