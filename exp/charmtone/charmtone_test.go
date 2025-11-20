package charmtone

import (
	"regexp"
	"testing"
)

var hexRegexp = regexp.MustCompile(`^#(?:[0-9a-fA-F]{3}){1,2}$`)

func TestValidateHexes(t *testing.T) {
	for _, key := range Keys() {
		if !hexRegexp.MatchString(key.Hex()) {
			t.Errorf("Key %s: invalid hex format %s", key, key.Hex())
		}
	}
}
