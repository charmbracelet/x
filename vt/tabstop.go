package vt

import "slices"

// DefaultTabInterval is the default tab interval.
const DefaultTabInterval = 8

// TabStops represents horizontal line tab stops.
type TabStops []int

// NewTabStops creates a new set of tab stops from a number of columns and an
// interval.
func NewTabStops(cols, interval int) TabStops {
	ts := make(TabStops, 0, cols/interval)
	for i := interval; i < cols; i += interval {
		ts = append(ts, i)
	}
	return ts
}

// DefaultTabStops creates a new set of tab stops with the default interval.
func DefaultTabStops(cols int) TabStops {
	return NewTabStops(cols, DefaultTabInterval)
}

// Next returns the next tab stop after the given column.
func (ts TabStops) Next(col int) int {
	for _, t := range ts {
		if t > col {
			return t
		}
	}
	return col
}

// Prev returns the previous tab stop before the given column.
func (ts TabStops) Prev(col int) int {
	for i := len(ts) - 1; i >= 0; i-- {
		if ts[i] < col {
			return ts[i]
		}
	}
	return col
}

// Set adds a tab stop at the given column.
func (ts *TabStops) Set(col int) {
	i, ok := binarySearch(*ts, col)
	if !ok {
		return
	}

	*ts = slices.Insert(*ts, i, col)
}

// Reset removes the tab stop at the given column.
func (ts *TabStops) Reset(col int) {
	i, ok := binarySearch(*ts, col)
	if !ok {
		return
	}

	*ts = slices.Delete(*ts, i, i+1)
}

// Clear removes all tab stops.
func (ts *TabStops) Clear() {
	*ts = (*ts)[:0]
}

// resetTabStops resets the terminal tab stops to the default set.
func (t *Terminal) resetTabStops() {
	t.tabstops = DefaultTabStops(t.Width())
}
