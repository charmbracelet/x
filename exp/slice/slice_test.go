package slice

import (
	"reflect"
	"testing"
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
	output := GroupBy(input, func(s string) string { return string(s[0]) })

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
		actual := Take(tc.input, tc.take)
		if len(actual) != len(tc.expected) {
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
		actual := Uniq(tc.input)
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
		actual := Intersperse(tc.input, tc.insert)
		if !reflect.DeepEqual(actual, tc.expected) {
			t.Errorf("Test %d: Expected %v, got %v", i, tc.expected, actual)
		}
	}
}
