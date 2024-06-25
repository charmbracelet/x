package maps

import (
	"cmp"
	"slices"
)

// SortedKeys returns the keys of the map m.
// The keys will be sorted.
func SortedKeys[M ~map[K]V, K cmp.Ordered, V any](m M) []K {
	r := Keys(m)
	slices.Sort(r)
	return r
}

// Keys returns the keys of the map m.
func Keys[M ~map[K]V, K cmp.Ordered, V any](m M) []K {
	r := make([]K, 0, len(m))
	for k := range m {
		r = append(r, k)
	}
	return r
}
