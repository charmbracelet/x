package charmtone

import (
	"strconv"
	"strings"
	"testing"
)

func TestValidateHexes(t *testing.T) {
	for _, key := range Keys() {
		hex := strings.TrimPrefix(key.Hex(), "#")
		if len(hex) != 6 && len(hex) != 3 {
			t.Errorf("Key %s: invalid hex length %d for %s", key, len(hex), key.Hex())
		}
		if _, err := strconv.ParseUint(hex, 16, 32); err != nil {
			t.Errorf("Key %s: invalid hex value %s", key, key.Hex())
		}
	}
}
