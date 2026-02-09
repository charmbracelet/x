package slice_test

import (
	"reflect"
	"slices"
	"testing"

	"github.com/charmbracelet/x/exp/slice"
)

func TestGroupBy(t *testing.T) {
	expected := map[string][]string{
		"a": {"andrey", "ayman"},
		"b": {"bash"},
		"c": {"carlos", "christian"},
		"r": {"raphael"},
	}
	input := []string{
		"andrey",
		"ayman",
		"bash",
		"carlos",
		"christian",
		"raphael",
	}
	output := slice.GroupBy(input, func(s string) string { return string(s[0]) })

	if !reflect.DeepEqual(expected, output) {
		t.Errorf("Expected %v, got %v", expected, output)
	}
}

func TestTake(t *testing.T) {
	for i, tc := range []struct {
		input    []int
		take     int
		expected []int
	}{
		{
			input:    []int{1, 2, 3, 4, 5},
			take:     3,
			expected: []int{1, 2, 3},
		},
		{
			input:    []int{1, 2, 3},
			take:     5,
			expected: []int{1, 2, 3},
		},
		{
			input:    []int{},
			take:     2,
			expected: []int{},
		},
		{
			input:    []int{1, 2, 3},
			take:     0,
			expected: []int{},
		},
		{
			input:    nil,
			take:     2,
			expected: []int{},
		},
	} {
		actual := slice.Take(tc.input, tc.take)
		if len(actual) != len(tc.expected) {
			t.Errorf("Test %d: Expected %v, got %v", i, tc.expected, actual)
		}
	}
}

func TestLast(t *testing.T) {
	for i, tc := range []struct {
		input    []int
		ok       bool
		expected int
	}{
		{
			input:    []int{1, 2, 3, 4, 5},
			ok:       true,
			expected: 5,
		},
		{
			input:    []int{1, 2, 3},
			ok:       true,
			expected: 3,
		},
		{
			input:    []int{1},
			ok:       true,
			expected: 1,
		},
		{
			input:    []int{},
			ok:       false,
			expected: 0,
		},
	} {
		actual, ok := slice.Last(tc.input)
		if ok != tc.ok {
			t.Errorf("Test %d: Expected ok %v, got %v", i, tc.ok, ok)
		}
		if actual != tc.expected {
			t.Errorf("Test %d: Expected %v, got %v", i, tc.expected, actual)
		}
	}
}

func TestUniq(t *testing.T) {
	for i, tc := range []struct {
		input    []int
		expected []int
	}{
		{
			input:    []int{1, 2, 3, 4, 5},
			expected: []int{1, 2, 3, 4, 5},
		},
		{
			input:    []int{1, 2, 2, 3, 4, 4, 5},
			expected: []int{1, 2, 3, 4, 5},
		},
		{
			input:    []int{1, 2, 3, 4, 5, 1, 2, 3, 4, 5, 1, 2, 3, 4, 5},
			expected: []int{1, 2, 3, 4, 5},
		},
		{
			input:    []int{},
			expected: []int{},
		},
	} {
		actual := slice.Uniq(tc.input)
		if !reflect.DeepEqual(actual, tc.expected) {
			t.Errorf("Test %d: Expected %v, got %v", i, tc.expected, actual)
		}
	}
}

func TestIntersperse(t *testing.T) {
	for i, tc := range []struct {
		input    []string
		insert   string
		expected []string
	}{
		{
			input:    []string{},
			insert:   "-",
			expected: []string{},
		},
		{
			input:    []string{"a"},
			insert:   "-",
			expected: []string{"a"},
		},
		{
			input:    []string{"a", "b"},
			insert:   "-",
			expected: []string{"a", "-", "b"},
		},
		{
			input:    []string{"a", "b", "c"},
			insert:   "-",
			expected: []string{"a", "-", "b", "-", "c"},
		},
	} {
		actual := slice.Intersperse(tc.input, tc.insert)
		if !reflect.DeepEqual(actual, tc.expected) {
			t.Errorf("Test %d: Expected %v, got %v", i, tc.expected, actual)
		}
	}
}

func TestContainsAny(t *testing.T) {
	for i, tc := range []struct {
		input    []string
		values   []string
		expected bool
	}{
		{
			input:    []string{"a", "b", "c"},
			values:   []string{"a", "b"},
			expected: true,
		},
		{
			input:    []string{"a", "b", "c"},
			values:   []string{"d", "e"},
			expected: false,
		},
		{
			input:    []string{"a", "b", "c"},
			values:   []string{"c", "d"},
			expected: true,
		},
		{
			input:    []string{},
			values:   []string{"d", "e"},
			expected: false,
		},
	} {
		actual := slice.ContainsAny(tc.input, tc.values...)
		if actual != tc.expected {
			t.Errorf("Test %d: Expected %v, got %v", i, tc.expected, actual)
		}
	}
}

func TestShift(t *testing.T) {
	for i, tc := range []struct {
		input         []int
		ok            bool
		expectedVal   int
		expectedSlice []int
	}{
		{
			input:         []int{1, 2, 3, 4, 5},
			ok:            true,
			expectedVal:   1,
			expectedSlice: []int{2, 3, 4, 5},
		},
		{
			input:         []int{1, 2, 3},
			ok:            true,
			expectedVal:   1,
			expectedSlice: []int{2, 3},
		},
		{
			input:         []int{1},
			ok:            true,
			expectedVal:   1,
			expectedSlice: []int{},
		},
		{
			input:         []int{},
			ok:            false,
			expectedVal:   0,
			expectedSlice: []int{},
		},
	} {
		actual, newSlice, ok := slice.Shift(tc.input)
		if ok != tc.ok {
			t.Errorf("test %d: expected ok %v, got %v", i, tc.ok, ok)
		}
		if actual != tc.expectedVal {
			t.Errorf("test %d: expected val %v, got %v", i, tc.expectedVal, actual)
		}
		if !reflect.DeepEqual(newSlice, tc.expectedSlice) {
			t.Errorf("test %d: expected slice %v, got %v", i, tc.expectedSlice, newSlice)
		}
	}
}

func TestPop(t *testing.T) {
	for i, tc := range []struct {
		input         []int
		ok            bool
		expectedVal   int
		expectedSlice []int
	}{
		{
			input:         []int{1, 2, 3, 4, 5},
			ok:            true,
			expectedVal:   5,
			expectedSlice: []int{1, 2, 3, 4},
		},
		{
			input:         []int{1, 2, 3},
			ok:            true,
			expectedVal:   3,
			expectedSlice: []int{1, 2},
		},
		{
			input:         []int{1},
			ok:            true,
			expectedVal:   1,
			expectedSlice: []int{},
		},
		{
			input:         []int{},
			ok:            false,
			expectedVal:   0,
			expectedSlice: []int{},
		},
	} {
		actual, newSlice, ok := slice.Pop(tc.input)
		if ok != tc.ok {
			t.Errorf("test %d: expected ok %v, got %v", i, tc.ok, ok)
		}
		if actual != tc.expectedVal {
			t.Errorf("test %d: expected val %v, got %v", i, tc.expectedVal, actual)
		}
		if !reflect.DeepEqual(newSlice, tc.expectedSlice) {
			t.Errorf("test %d: expected slice %v, got %v", i, tc.expectedSlice, newSlice)
		}
	}
}

func TestDeleteAt(t *testing.T) {
	for i, tc := range []struct {
		input         []int
		index         int
		ok            bool
		expectedVal   int
		expectedSlice []int
	}{
		{
			input:         []int{1, 2, 3, 4, 5},
			index:         2,
			ok:            true,
			expectedVal:   3,
			expectedSlice: []int{1, 2, 4, 5},
		},
		{
			input:         []int{1, 2, 3},
			index:         0,
			ok:            true,
			expectedVal:   1,
			expectedSlice: []int{2, 3},
		},
		{
			input:         []int{1, 2, 3},
			index:         2,
			ok:            true,
			expectedVal:   3,
			expectedSlice: []int{1, 2},
		},
		{
			input:         []int{1},
			index:         0,
			ok:            true,
			expectedVal:   1,
			expectedSlice: []int{},
		},
		{
			input:         []int{},
			index:         0,
			ok:            false,
			expectedVal:   0,
			expectedSlice: []int{},
		},
	} {
		actual, newSlice, ok := slice.DeleteAt(tc.input, tc.index)
		if ok != tc.ok {
			t.Errorf("test %d: expected ok %v, got %v", i, tc.ok, ok)
		}
		if actual != tc.expectedVal {
			t.Errorf("test %d: expected val %v, got %v", i, tc.expectedVal, actual)
		}
		if !reflect.DeepEqual(newSlice, tc.expectedSlice) {
			t.Errorf("test %d: expected slice %v, got %v", i, tc.expectedSlice, newSlice)
		}
	}
}

func TestIsSubset(t *testing.T) {
	tests := []struct {
		name     string
		a        []string
		b        []string
		expected bool
	}{
		// Basic subset cases
		{
			name:     "empty subset of empty",
			a:        []string{},
			b:        []string{},
			expected: true,
		},
		{
			name:     "empty subset of non-empty",
			a:        []string{},
			b:        []string{"a", "b", "c"},
			expected: true,
		},
		{
			name:     "non-empty not subset of empty",
			a:        []string{"a"},
			b:        []string{},
			expected: false,
		},
		{
			name:     "single element subset",
			a:        []string{"b"},
			b:        []string{"a", "b", "c"},
			expected: true,
		},
		{
			name:     "single element not subset",
			a:        []string{"d"},
			b:        []string{"a", "b", "c"},
			expected: false,
		},
		{
			name:     "multiple elements subset",
			a:        []string{"a", "c"},
			b:        []string{"a", "b", "c", "d"},
			expected: true,
		},
		{
			name:     "multiple elements not subset",
			a:        []string{"a", "e"},
			b:        []string{"a", "b", "c", "d"},
			expected: false,
		},
		{
			name:     "equal sets are subsets",
			a:        []string{"a", "b", "c"},
			b:        []string{"a", "b", "c"},
			expected: true,
		},
		{
			name:     "larger set not subset of smaller",
			a:        []string{"a", "b", "c", "d"},
			b:        []string{"a", "b"},
			expected: false,
		},

		// Order independence
		{
			name:     "subset with different order",
			a:        []string{"c", "a"},
			b:        []string{"b", "a", "d", "c"},
			expected: true,
		},

		// Duplicate handling
		{
			name:     "duplicates in subset",
			a:        []string{"a", "a", "b"},
			b:        []string{"a", "b", "c"},
			expected: true,
		},
		{
			name:     "duplicates in superset",
			a:        []string{"a", "b"},
			b:        []string{"a", "a", "b", "b", "c"},
			expected: true,
		},
		{
			name:     "duplicates in both",
			a:        []string{"a", "a", "b"},
			b:        []string{"a", "a", "b", "b", "c"},
			expected: true,
		},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := slice.IsSubset(tt.a, tt.b)
			if actual != tt.expected {
				t.Errorf("Test %d: Expected %v, got %v", i, tt.expected, actual)
			}
		})
	}
}

func TestIsSubsetWithInts(t *testing.T) {
	tests := []struct {
		name     string
		a        []int
		b        []int
		expected bool
	}{
		{
			name:     "int subset",
			a:        []int{1, 3},
			b:        []int{1, 2, 3, 4},
			expected: true,
		},
		{
			name:     "int not subset",
			a:        []int{1, 5},
			b:        []int{1, 2, 3, 4},
			expected: false,
		},
		{
			name:     "empty int subset",
			a:        []int{},
			b:        []int{1, 2, 3},
			expected: true,
		},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := slice.IsSubset(tt.a, tt.b)
			if actual != tt.expected {
				t.Errorf("Test %d: Expected %v, got %v", i, tt.expected, actual)
			}
		})
	}
}

func TestMapStringToInt(t *testing.T) {
	seq := slices.Values([]string{"a", "ab", "abc", "abcd"})
	mapped := slice.Map(seq, func(s string) int { return len(s) })
	expected := []int{1, 2, 3, 4}

	result := slices.Collect(mapped)

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}
