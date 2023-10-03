package slice

// Take returns the first n elements of the given slice. If there are not
// enough elements in the slice, the whole slice is returned.
func Take[A any](slice []A, n int) []A {
	if n > len(slice) {
		return slice
	}
	return slice[:n]
}

// Reverse returns a new slice with the elements of the given slice in reverse.
func Reverse[A any](s []A) []A {
	r := make([]A, len(s))
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		r[i], r[j] = s[j], s[i]
	}
	return r
}

// Delete returns a new slice with the element at the given index removed.
// Slice order is preserved. Runs in linear time, aka O(n).
func Delete[A any](s []A, i int) []A {
	if i < 0 || i >= len(s) {
		return s
	}
	var d = make([]A, len(s))
	copy(d, s)
	copy(d[i:], d[i+1:])  // shift left
	d[len(d)-1] = *new(A) // zero out last element
	return s[:len(s)-1]   // truncate
}
