package maps

import (
	"slices"
	"testing"
)

func TestSortedKeys(t *testing.T) {
	m := map[string]int{
		"foo":   1,
		"bar":   10,
		"aaaaa": 11,
	}

	keys := SortedKeys(m)
	if slices.Compare(keys, []string{"aaaaa", "bar", "foo"}) != 0 {
		t.Fatalf("unexpected keys order: %v", keys)
	}
}
