package vt

func max(a, b int) int { //nolint:predeclared
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int { //nolint:predeclared
	if a > b {
		return b
	}
	return a
}

func clamp(v, low, high int) int {
	if high < low {
		low, high = high, low
	}
	return min(high, max(low, v))
}
