package errors

import (
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
		je := err.(*joinError)
		un := je.Unwrap()
		if len(un) != 1 {
			t.Fatalf("expected 1 err, got %d", len(un))
		}
		if s := un[0].Error(); s != expected.Error() {
			t.Errorf("expected %v, got %v", expected, un[0])
		}
		if s := err.Error(); s != expected.Error() {
			t.Errorf("expected %s, got %s", expected, err)
		}
	})
	t.Run("many errs", func(t *testing.T) {
		expected1 := fmt.Errorf("fake 1")
		expected2 := fmt.Errorf("fake 2")
		err := Join(nil, expected1, nil, nil, expected2, nil)
		je := err.(*joinError)
		un := je.Unwrap()
		if len(un) != 2 {
			t.Fatalf("expected 2 err, got %d", len(un))
		}
		if s := un[0].Error(); s != expected1.Error() {
			t.Errorf("expected %v, got %v", expected1, un[0])
		}
		if s := un[1].Error(); s != expected2.Error() {
			t.Errorf("expected %v, got %v", expected2, un[1])
		}
		expectedS := expected1.Error() + "\n" + expected2.Error()
		if s := err.Error(); s != expectedS {
			t.Errorf("expected %s, got %s", expectedS, err)
		}
	})
}
