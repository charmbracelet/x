// Package ordered provides utility functions for ordered types.
package ordered

import "cmp"

// Clamp returns a value clamped between the given low and high values.
func Clamp[T cmp.Ordered](n, low, high T) T {
	if low > high {
		low, high = high, low
	}
	return min(high, max(low, n))
}

// First returns the first non-default value of a fixed number of
// arguments of [cmp.Ordered] types.
func First[T cmp.Ordered](x T, y ...T) T {
	var empty T
	if x != empty {
		return x
	}
	for _, s := range y {
		if s != empty {
			return s
		}
	}
	return empty
}
