package charmtone

import (
	"testing"

	"github.com/lucasb-eyer/go-colorful"
)

func TestValidateHexes(t *testing.T) {
	for k, v := range colors {
		if _, err := colorful.Hex(v); err != nil {
			t.Errorf("Key %s: %v", k, err)
		}
	}
}
