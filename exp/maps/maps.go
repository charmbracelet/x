package maps

import (
	"golang.org/x/exp/constraints"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

// SortedKeys returns the keys of the map m.
// The keys will be sorted.
func SortedKeys[M ~map[K]V, K constraints.Ordered, V any](m M) []K {
	r := maps.Keys(m)
	slices.Sort(r)
	return r
}
