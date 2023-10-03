package slice

import "testing"

func Test_Take(t *testing.T) {
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
		actual := Take(tc.input, tc.take)
		if len(actual) != len(tc.expected) {
			t.Errorf("Test %d: Expected %v, got %v", i, tc.expected, actual)
		}
	}
}

func Test_Reverse(t *testing.T) {
	for i, tc := range []struct {
		input    []int
		expected []int
	}{
		{
			input:    []int{1, 2, 3, 4, 5},
			expected: []int{5, 4, 3, 2, 1},
		},
		{
			input:    []int{1, 2, 3},
			expected: []int{3, 2, 1},
		},
		{
			input:    []int{},
			expected: []int{},
		},
		{
			input:    nil,
			expected: []int{},
		},
	} {
		actual := Reverse(tc.input)
		if len(actual) != len(tc.expected) {
			t.Errorf("Test %d: Expected %v, got %v", i, tc.expected, actual)
		}
	}
}

func Test_Delete(t *testing.T) {
	for i, tc := range []struct {
		input    []int
		index    int
		expected []int
	}{
		{
			input:    []int{1, 2, 3},
			index:    1,
			expected: []int{1, 3},
		},
		{
			input:    []int{1, 2, 3},
			index:    0,
			expected: []int{2, 3},
		},
		{
			input:    []int{1, 2, 3},
			index:    2,
			expected: []int{1, 2},
		},
		{
			input:    []int{1, 2, 3},
			index:    4,
			expected: []int{1, 2, 3},
		},
		{
			input:    []int{1, 2, 3},
			index:    -1,
			expected: []int{1, 2, 3},
		},
		{
			input:    []int{},
			index:    0,
			expected: []int{},
		},
		{
			input:    nil,
			index:    0,
			expected: []int{},
		},
	} {
		actual := Delete(tc.input, tc.index)
		if len(actual) != len(tc.expected) {
			t.Errorf("Test %d: Expected %v, got %v", i, tc.expected, actual)
		}
	}
}
