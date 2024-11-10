package vt

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
