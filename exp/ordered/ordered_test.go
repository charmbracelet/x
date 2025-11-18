package ordered

import (
	"cmp"
	"fmt"
	"testing"
)

func assertEqual[T cmp.Ordered](tb testing.TB, result, expect T) {
	tb.Helper()
	if result != expect {
		tb.Errorf("expected %v, got %v", expect, result)
	}
}

func TestClamp(t *testing.T) {
	name := func(n, low, high, expect any) string {
		return fmt.Sprintf("clamp(%v, %v, %v) = %v", n, low, high, expect)
	}

	for expect, input := range map[int]struct {
		n, low, high int
	}{
		10: {10, 1, 10},
		4:  {2, 4, 30},
		32: {45, 20, 32},
		15: {15, 33, 11},
	} {
		t.Run(name(input.n, input.low, input.high, expect), func(t *testing.T) {
			assertEqual(t, Clamp(input.n, input.low, input.high), expect)
		})
	}

	for expect, input := range map[float64]struct {
		n, low, high float64
	}{
		1.0: {1.0, 1.0, 10.3},
		0.4: {0.2, 0.4, 30},
		3.2: {4.5, 2.0, 3.2},
	} {
		t.Run(name(input.n, input.low, input.high, expect), func(t *testing.T) {
			assertEqual(t, Clamp(input.n, input.low, input.high), expect)
		})
	}

	for expect, input := range map[string]struct {
		n, low, high string
	}{
		"a":    {"a", "a", "aaaa"},
		"aaa":  {"aaa", "aa", "aaaa"},
		"aaaa": {"aaaaaa", "aa", "aaaa"},
	} {
		t.Run(name(input.n, input.low, input.high, expect), func(t *testing.T) {
			assertEqual(t, Clamp(input.n, input.low, input.high), expect)
		})
	}
}

func TestFirst(t *testing.T) {
	name := func(args []any, expect any) string {
		return fmt.Sprintf("first(%v) = %v", args, expect)
	}
	for expect, args := range map[string]struct {
		x string
		y []string
	}{
		"a": {"", []string{"", "a", "b", ""}},
		"c": {"c", []string{"", "a", "b", ""}},
		"":  {"", nil},
	} {
		fnargs := []any{args.x}
		for _, y := range args.y {
			fnargs = append(fnargs, y)
		}

		t.Run(name(fnargs, expect), func(t *testing.T) {
			assertEqual(t, First(args.x, args.y...), expect)
		})
	}

	for expect, args := range map[int]struct {
		x int
		y []int
	}{
		1:   {0, []int{0, 1, 2}},
		0:   {0, []int{0, 0, 0, 0}},
		100: {100, []int{0}},
	} {
		fnargs := []any{args.x}
		for _, y := range args.y {
			fnargs = append(fnargs, y)
		}

		t.Run(name(fnargs, expect), func(t *testing.T) {
			assertEqual(t, First(args.x, args.y...), expect)
		})
	}
}
