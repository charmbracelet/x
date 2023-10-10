package ordered

import (
	"cmp"
	"fmt"
	"testing"
)

func minName[T cmp.Ordered](x, y, expect T) string {
	return fmt.Sprintf("min(%v, %v) = %v", x, y, expect)
}

func assertMin[T cmp.Ordered](tb testing.TB, x, y, expect T) {
	tb.Helper()
	if r := Min(x, y); r != expect {
		tb.Errorf("expected %v, got %v", expect, r)
	}
}

func TestMin(t *testing.T) {
	for expect, args := range map[int][2]int{
		1:   {1, 2},
		0:   {1, 0},
		-10: {1, -10},
	} {
		t.Run(minName(args[0], args[1], expect), func(t *testing.T) {
			assertMin(t, args[0], args[1], expect)
		})
	}
	for expect, args := range map[float64][2]float64{
		0.1:  {0.1, 2},
		0.0:  {1, 0},
		-1.0: {1, -1.0},
	} {
		t.Run(minName(args[0], args[1], expect), func(t *testing.T) {
			assertMin(t, args[0], args[1], expect)
		})
	}
	for expect, args := range map[string][2]string{
		"":   {"", "a"},
		"a":  {"aa", "a"},
		"aa": {"aa", "aaaa"},
	} {
		t.Run(minName(args[0], args[1], expect), func(t *testing.T) {
			assertMin(t, args[0], args[1], expect)
		})
	}
}

func maxName[T cmp.Ordered](x, y, expect T) string {
	return fmt.Sprintf("max(%v, %v) = %v", x, y, expect)
}

func assertMax[T cmp.Ordered](tb testing.TB, x, y, expect T) {
	tb.Helper()
	if r := Max(x, y); r != expect {
		tb.Errorf("expected %v, got %v", expect, r)
	}
}

func TestMax(t *testing.T) {
	for expect, args := range map[int][2]int{
		2: {1, 2},
		1: {1, 0},
		0: {0, -10},
	} {
		t.Run(maxName(args[0], args[1], expect), func(t *testing.T) {
			assertMax(t, args[0], args[1], expect)
		})
	}
	for expect, args := range map[float64][2]float64{
		0.1:  {0.1, 0.02},
		1.0:  {1, 0},
		-1.0: {-1.1, -1.0},
	} {
		t.Run(maxName(args[0], args[1], expect), func(t *testing.T) {
			assertMax(t, args[0], args[1], expect)
		})
	}
	for expect, args := range map[string][2]string{
		"a":    {"", "a"},
		"aa":   {"aa", "a"},
		"aaaa": {"aa", "aaaa"},
	} {
		t.Run(maxName(args[0], args[1], expect), func(t *testing.T) {
			assertMax(t, args[0], args[1], expect)
		})
	}
}

func clampName[T cmp.Ordered](n, low, high, expect T) string {
	return fmt.Sprintf("clamp(%v, %v, %v) = %v", n, low, high, expect)
}

func assertClamp[T cmp.Ordered](tb testing.TB, n, low, high, expect T) {
	tb.Helper()
	if r := Clamp(n, low, high); r != expect {
		tb.Errorf("expected %v, got %v", expect, r)
	}
}

func TestClamp(t *testing.T) {
	for expect, input := range map[int]struct {
		n, low, high int
	}{
		10: {10, 1, 10},
		4:  {2, 4, 30},
		32: {45, 20, 32},
		15: {15, 33, 11},
	} {
		t.Run(clampName(input.n, input.low, input.high, expect), func(t *testing.T) {
			assertClamp(t, input.n, input.low, input.high, expect)
		})
	}

	for expect, input := range map[float64]struct {
		n, low, high float64
	}{
		1.0: {1.0, 1.0, 10.3},
		0.4: {0.2, 0.4, 30},
		3.2: {4.5, 2.0, 3.2},
	} {
		t.Run(clampName(input.n, input.low, input.high, expect), func(t *testing.T) {
			assertClamp(t, input.n, input.low, input.high, expect)
		})
	}

	for expect, input := range map[string]struct {
		n, low, high string
	}{
		"a":    {"a", "a", "aaaa"},
		"aaa":  {"aaa", "aa", "aaaa"},
		"aaaa": {"aaaaaa", "aa", "aaaa"},
	} {
		t.Run(clampName(input.n, input.low, input.high, expect), func(t *testing.T) {
			assertClamp(t, input.n, input.low, input.high, expect)
		})
	}
}

func firstName[T cmp.Ordered](args []T, expect T) string {
	return fmt.Sprintf("first(%v) = %v", args, expect)
}

func assertFirst[T cmp.Ordered](tb testing.TB, args []T, expect T) {
	tb.Helper()
	if r := First(args[0], args[1:]...); r != expect {
		tb.Errorf("expected %v, got %v", expect, r)
	}
}

func TestFirst(t *testing.T) {
	for expect, args := range map[string][]string{
		"a": {"", "", "a", "b", ""},
		"c": {"c", "", "a", "b", ""},
	} {
		t.Run(firstName(args, expect), func(t *testing.T) {
			assertFirst(t, args, expect)
		})
	}

	for expect, args := range map[int][]int{
		1:   {0, 0, 1, 2},
		0:   {0, 0},
		100: {100, 0},
	} {
		t.Run(firstName(args, expect), func(t *testing.T) {
			assertFirst(t, args, expect)
		})
	}
}
