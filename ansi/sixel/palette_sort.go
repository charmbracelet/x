// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sixel

import (
	"math/bits"
)

const (
	unknownHint sortedHint = iota
	increasingHint
	decreasingHint
)

type ordered interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
		~float32 | ~float64 |
		~string
}

// Compare returns
//
//	-1 if x is less than y,
//	 0 if x equals y,
//	+1 if x is greater than y.
//
// For floating-point types, a NaN is considered less than any non-NaN,
// a NaN is considered equal to a NaN, and -0.0 is equal to 0.0.
func compare[T ordered](x, y T) int {
	xNaN := isNaN(x)
	yNaN := isNaN(y)
	if xNaN {
		if yNaN {
			return 0
		}
		return -1
	}
	if yNaN {
		return +1
	}
	if x < y {
		return -1
	}
	if x > y {
		return +1
	}
	return 0
}

// isNaN reports whether x is a NaN without requiring the math package.
// This will always return false if T is not floating-point.
func isNaN[T ordered](x T) bool {
	return x != x
}

func sortFunc[S ~[]E, E any](x S, cmp func(a, b E) int) {
	n := len(x)
	pdqsortCmpFunc(x, 0, n, bits.Len(uint(n)), cmp)
}

type sortedHint int // hint for pdqsort when choosing the pivot

// xorshift paper: https://www.jstatsoft.org/article/view/v008i14/xorshift.pdf
type xorshift uint64

func (r *xorshift) Next() uint64 {
	*r ^= *r << 13
	*r ^= *r >> 7
	*r ^= *r << 17
	return uint64(*r)
}

func nextPowerOfTwo(length int) uint {
	return 1 << bits.Len(uint(length)) //nolint:gosec
}

// insertionSortCmpFunc sorts data[a:b] using insertion sort.
func insertionSortCmpFunc[E any](data []E, a, b int, cmp func(a, b E) int) {
	for i := a + 1; i < b; i++ {
		for j := i; j > a && (cmp(data[j], data[j-1]) < 0); j-- {
			data[j], data[j-1] = data[j-1], data[j]
		}
	}
}

// siftDownCmpFunc implements the heap property on data[lo:hi].
// first is an offset into the array where the root of the heap lies.
func siftDownCmpFunc[E any](data []E, lo, hi, first int, cmp func(a, b E) int) {
	root := lo
	for {
		child := 2*root + 1
		if child >= hi {
			break
		}
		if child+1 < hi && (cmp(data[first+child], data[first+child+1]) < 0) {
			child++
		}
		if !(cmp(data[first+root], data[first+child]) < 0) { //nolint:staticcheck
			return
		}
		data[first+root], data[first+child] = data[first+child], data[first+root]
		root = child
	}
}

func heapSortCmpFunc[E any](data []E, a, b int, cmp func(a, b E) int) {
	first := a
	lo := 0
	hi := b - a

	// Build heap with greatest element at top.
	for i := (hi - 1) / 2; i >= 0; i-- {
		siftDownCmpFunc(data, i, hi, first, cmp)
	}

	// Pop elements, largest first, into end of data.
	for i := hi - 1; i >= 0; i-- {
		data[first], data[first+i] = data[first+i], data[first]
		siftDownCmpFunc(data, lo, i, first, cmp)
	}
}

// pdqsortCmpFunc sorts data[a:b].
// The algorithm based on pattern-defeating quicksort(pdqsort), but without the optimizations from BlockQuicksort.
// pdqsort paper: https://arxiv.org/pdf/2106.05123.pdf
// C++ implementation: https://github.com/orlp/pdqsort
// Rust implementation: https://docs.rs/pdqsort/latest/pdqsort/
// limit is the number of allowed bad (very unbalanced) pivots before falling back to heapsort.
func pdqsortCmpFunc[E any](data []E, a, b, limit int, cmp func(a, b E) int) {
	const maxInsertion = 12

	var (
		wasBalanced    = true // whether the last partitioning was reasonably balanced
		wasPartitioned = true // whether the slice was already partitioned
	)

	for {
		length := b - a

		if length <= maxInsertion {
			insertionSortCmpFunc(data, a, b, cmp)
			return
		}

		// Fall back to heapsort if too many bad choices were made.
		if limit == 0 {
			heapSortCmpFunc(data, a, b, cmp)
			return
		}

		// If the last partitioning was imbalanced, we need to breaking patterns.
		if !wasBalanced {
			breakPatternsCmpFunc(data, a, b)
			limit--
		}

		pivot, hint := choosePivotCmpFunc(data, a, b, cmp)
		if hint == decreasingHint {
			reverseRangeCmpFunc(data, a, b)
			// The chosen pivot was pivot-a elements after the start of the array.
			// After reversing it is pivot-a elements before the end of the array.
			// The idea came from Rust's implementation.
			pivot = (b - 1) - (pivot - a)
			hint = increasingHint
		}

		// The slice is likely already sorted.
		if wasBalanced && wasPartitioned && hint == increasingHint {
			if partialInsertionSortCmpFunc(data, a, b, cmp) {
				return
			}
		}

		// Probably the slice contains many duplicate elements, partition the slice into
		// elements equal to and elements greater than the pivot.
		if a > 0 && !(cmp(data[a-1], data[pivot]) < 0) { //nolint:staticcheck
			mid := partitionEqualCmpFunc(data, a, b, pivot, cmp)
			a = mid
			continue
		}

		mid, alreadyPartitioned := partitionCmpFunc(data, a, b, pivot, cmp)
		wasPartitioned = alreadyPartitioned

		leftLen, rightLen := mid-a, b-mid
		balanceThreshold := length / 8
		if leftLen < rightLen {
			wasBalanced = leftLen >= balanceThreshold
			pdqsortCmpFunc(data, a, mid, limit, cmp)
			a = mid + 1
		} else {
			wasBalanced = rightLen >= balanceThreshold
			pdqsortCmpFunc(data, mid+1, b, limit, cmp)
			b = mid
		}
	}
}

// partitionCmpFunc does one quicksort partition.
// Let p = data[pivot]
// Moves elements in data[a:b] around, so that data[i]<p and data[j]>=p for i<newpivot and j>newpivot.
// On return, data[newpivot] = p.
func partitionCmpFunc[E any](data []E, a, b, pivot int, cmp func(a, b E) int) (newpivot int, alreadyPartitioned bool) {
	data[a], data[pivot] = data[pivot], data[a]
	i, j := a+1, b-1 // i and j are inclusive of the elements remaining to be partitioned

	for i <= j && (cmp(data[i], data[a]) < 0) {
		i++
	}
	for i <= j && !(cmp(data[j], data[a]) < 0) { //nolint:staticcheck
		j--
	}
	if i > j {
		data[j], data[a] = data[a], data[j]
		return j, true
	}
	data[i], data[j] = data[j], data[i]
	i++
	j--

	for {
		for i <= j && (cmp(data[i], data[a]) < 0) {
			i++
		}
		for i <= j && !(cmp(data[j], data[a]) < 0) { //nolint:staticcheck
			j--
		}
		if i > j {
			break
		}
		data[i], data[j] = data[j], data[i]
		i++
		j--
	}
	data[j], data[a] = data[a], data[j]
	return j, false
}

// partitionEqualCmpFunc partitions data[a:b] into elements equal to data[pivot] followed by elements greater than data[pivot].
// It assumed that data[a:b] does not contain elements smaller than the data[pivot].
func partitionEqualCmpFunc[E any](data []E, a, b, pivot int, cmp func(a, b E) int) (newpivot int) {
	data[a], data[pivot] = data[pivot], data[a]
	i, j := a+1, b-1 // i and j are inclusive of the elements remaining to be partitioned

	for {
		for i <= j && !(cmp(data[a], data[i]) < 0) { //nolint:staticcheck
			i++
		}
		for i <= j && (cmp(data[a], data[j]) < 0) {
			j--
		}
		if i > j {
			break
		}
		data[i], data[j] = data[j], data[i]
		i++
		j--
	}
	return i
}

// partialInsertionSortCmpFunc partially sorts a slice, returns true if the slice is sorted at the end.
func partialInsertionSortCmpFunc[E any](data []E, a, b int, cmp func(a, b E) int) bool {
	const (
		maxSteps         = 5  // maximum number of adjacent out-of-order pairs that will get shifted
		shortestShifting = 50 // don't shift any elements on short arrays
	)
	i := a + 1
	for range maxSteps {
		for i < b && !(cmp(data[i], data[i-1]) < 0) { //nolint:staticcheck
			i++
		}

		if i == b {
			return true
		}

		if b-a < shortestShifting {
			return false
		}

		data[i], data[i-1] = data[i-1], data[i]

		// Shift the smaller one to the left.
		if i-a >= 2 {
			for j := i - 1; j >= 1; j-- {
				if !(cmp(data[j], data[j-1]) < 0) { //nolint:staticcheck
					break
				}
				data[j], data[j-1] = data[j-1], data[j]
			}
		}
		// Shift the greater one to the right.
		if b-i >= 2 {
			for j := i + 1; j < b; j++ {
				if !(cmp(data[j], data[j-1]) < 0) { //nolint:staticcheck
					break
				}
				data[j], data[j-1] = data[j-1], data[j]
			}
		}
	}
	return false
}

// breakPatternsCmpFunc scatters some elements around in an attempt to break some patterns
// that might cause imbalanced partitions in quicksort.
func breakPatternsCmpFunc[E any](data []E, a, b int) {
	length := b - a
	if length >= 8 {
		random := xorshift(length)
		modulus := nextPowerOfTwo(length)

		for idx := a + (length/4)*2 - 1; idx <= a+(length/4)*2+1; idx++ {
			other := int(uint(random.Next()) & (modulus - 1)) //nolint:gosec
			if other >= length {
				other -= length
			}
			data[idx], data[a+other] = data[a+other], data[idx]
		}
	}
}

// choosePivotCmpFunc chooses a pivot in data[a:b].
//
// [0,8): chooses a static pivot.
// [8,shortestNinther): uses the simple median-of-three method.
// [shortestNinther,∞): uses the Tukey ninther method.
func choosePivotCmpFunc[E any](data []E, a, b int, cmp func(a, b E) int) (pivot int, hint sortedHint) {
	const (
		shortestNinther = 50
		maxSwaps        = 4 * 3
	)

	l := b - a

	var (
		swaps int
		i     = a + l/4*1
		j     = a + l/4*2
		k     = a + l/4*3
	)

	if l >= 8 {
		if l >= shortestNinther {
			// Tukey ninther method, the idea came from Rust's implementation.
			i = medianAdjacentCmpFunc(data, i, &swaps, cmp)
			j = medianAdjacentCmpFunc(data, j, &swaps, cmp)
			k = medianAdjacentCmpFunc(data, k, &swaps, cmp)
		}
		// Find the median among i, j, k and stores it into j.
		j = medianCmpFunc(data, i, j, k, &swaps, cmp)
	}

	switch swaps {
	case 0:
		return j, increasingHint
	case maxSwaps:
		return j, decreasingHint
	default:
		return j, unknownHint
	}
}

// order2CmpFunc returns x,y where data[x] <= data[y], where x,y=a,b or x,y=b,a.
func order2CmpFunc[E any](data []E, a, b int, swaps *int, cmp func(a, b E) int) (int, int) {
	if cmp(data[b], data[a]) < 0 {
		*swaps++
		return b, a
	}
	return a, b
}

// medianCmpFunc returns x where data[x] is the median of data[a],data[b],data[c], where x is a, b, or c.
func medianCmpFunc[E any](data []E, a, b, c int, swaps *int, cmp func(a, b E) int) int {
	a, b = order2CmpFunc(data, a, b, swaps, cmp)
	b, _ = order2CmpFunc(data, b, c, swaps, cmp)
	_, b = order2CmpFunc(data, a, b, swaps, cmp)
	return b
}

// medianAdjacentCmpFunc finds the median of data[a - 1], data[a], data[a + 1] and stores the index into a.
func medianAdjacentCmpFunc[E any](data []E, a int, swaps *int, cmp func(a, b E) int) int {
	return medianCmpFunc(data, a-1, a, a+1, swaps, cmp)
}

func reverseRangeCmpFunc[E any](data []E, a, b int) {
	i := a
	j := b - 1
	for i < j {
		data[i], data[j] = data[j], data[i]
		i++
		j--
	}
}
