package higherorder

import "testing"

func Test_Foldl(t *testing.T) {
	x := Foldl(func(a, b int) int {
		return a + b
	}, 0, []int{1, 2, 3})

	const expect = 6
	if x != expect {
		t.Errorf("Expected %d, got %d", expect, x)
	}
}

func Test_Foldr(t *testing.T) {
	x := Foldl(func(a, b int) int {
		return a - b
	}, 6, []int{1, 2, 3})

	const expect = 0
	if x != expect {
		t.Errorf("Expected %d, got %d", expect, x)
	}
}

func Test_Map(t *testing.T) {
	{
		// Map over ints, returning the square of each int.
		// (Take ints, return ints.)
		x := Map(func(a int) int {
			return a * a
		}, []int{2, 3, 4})

		expected := []int{4, 9, 16}
		for i, v := range x {
			if v != expected[i] {
				t.Errorf("Index %d: expected %d, got %d", i, expected[i], v)
			}
		}
	}
	{
		// Map over strings, returning the length of each string.
		// (Take ints, return strings.)
		x := Map(func(a string) int {
			return len([]rune(a))
		}, []string{"one", "two", "three"})

		expected := []int{3, 3, 5}
		for i, v := range x {
			if v != expected[i] {
				t.Errorf("Index %d: expected %d, got %d", i, expected[i], v)
			}
		}
	}
}
