package ansi

import (
	"reflect"
	"testing"
)

func TestParamsParameters(t *testing.T) {
	cases := []struct {
		params string
		want   [][]uint
	}{
		{"", [][]uint{}},
		{"0", [][]uint{{0}}},
		{"1;2", [][]uint{{1}, {2}}},
		{"1:2;3", [][]uint{{1, 2}, {3}}},
		{"1;2;q", [][]uint{{1}, {2}, {0}}},
		{"1;;2:255:255:0", [][]uint{{1}, {0}, {2, 255, 255, 0}}},
		{"1;2:::0", [][]uint{{1}, {2, 0, 0, 0}}},
	}
	for i, c := range cases {
		got := Params([]byte(c.params))
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("case %d, got %v, want %v", i+1, got, c.want)
		}
	}
}
