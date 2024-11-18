package vt

// DefaultTabInterval is the default tab interval.
const DefaultTabInterval = 8

// TabStops represents horizontal line tab stops.
type TabStops struct {
	stops    []int
	interval int
	width    int
}

// NewTabStops creates a new set of tab stops from a number of columns and an
// interval.
func NewTabStops(width, interval int) *TabStops {
	ts := new(TabStops)
	ts.interval = interval
	ts.width = width
	ts.stops = make([]int, (width+(interval-1))/interval)
	for x := 0; x < width; x++ {
		if x%interval == 0 {
			ts.Set(x)
		} else {
			ts.Reset(x)
		}
	}
	return ts
}

// DefaultTabStops creates a new set of tab stops with the default interval.
func DefaultTabStops(cols int) *TabStops {
	return NewTabStops(cols, DefaultTabInterval)
}

// IsStop returns true if the given column is a tab stop.
func (ts TabStops) IsStop(col int) bool {
	mask := ts.mask(col)
	i := col >> 3
	if i < 0 || i >= len(ts.stops) {
		return false
	}
	return ts.stops[i]&mask != 0
}

// Next returns the next tab stop after the given column.
func (ts TabStops) Next(col int) int {
	return ts.Find(col, 1)
}

// Prev returns the previous tab stop before the given column.
func (ts TabStops) Prev(col int) int {
	return ts.Find(col, -1)
}

// Find returns the prev/next tab stop before/after the given column and delta.
// If delta is positive, it returns the next tab stop after the given column.
// If delta is negative, it returns the previous tab stop before the given column.
// If delta is zero, it returns the given column.
func (ts TabStops) Find(col, delta int) int {
	if delta == 0 {
		return col
	}

	var prev bool
	count := delta
	if count < 0 {
		count = -count
		prev = true
	}

	for count > 0 {
		if !prev {
			if col >= ts.width-1 {
				return col
			}

			col++
		} else {
			if col < 1 {
				return col
			}

			col--
		}

		if ts.IsStop(col) {
			count--
		}
	}

	return col
}

// Set adds a tab stop at the given column.
func (ts *TabStops) Set(col int) {
	mask := ts.mask(col)
	ts.stops[col>>3] |= mask
}

// Reset removes the tab stop at the given column.
func (ts *TabStops) Reset(col int) {
	mask := ts.mask(col)
	ts.stops[col>>3] &= ^mask
}

// Clear removes all tab stops.
func (ts *TabStops) Clear() {
	ts.stops = make([]int, 0, len(ts.stops))
}

// mask returns the mask for the given column.
func (ts *TabStops) mask(col int) int {
	return 1 << (col & (ts.interval - 1))
}

// resetTabStops resets the terminal tab stops to the default set.
func (t *Terminal) resetTabStops() {
	t.tabstops = DefaultTabStops(t.Width())
}
