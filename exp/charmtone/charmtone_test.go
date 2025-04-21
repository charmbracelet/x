package charmtone

import (
	"testing"

	"github.com/lucasb-eyer/go-colorful"
)

func TestValidateHexes(t *testing.T) {
	for _, key := range Keys() {
		if _, err := colorful.Hex(key.Hex()); err != nil {
			t.Errorf("Key %s: %v", key, err)
		}
	}
}
