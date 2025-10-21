package ansi

import "testing"

func TestUrxvtExt(t *testing.T) {
	tests := []struct {
		extension string
		params    []string
		expected  string
	}{
		{
			extension: "foo",
			params:    []string{"bar", "baz"},
			expected:  "\x1b]777;foo;bar;baz\x07",
		},
		{
			extension: "test",
			params:    []string{},
			expected:  "\x1b]777;test;\x07",
		},
		{
			extension: "example",
			params:    []string{"param1"},
			expected:  "\x1b]777;example;param1\x07",
		},
		{
			extension: "notify",
			params:    []string{"message", "info"},
			expected:  "\x1b]777;notify;message;info\x07",
		},
	}

	for _, tt := range tests {
		result := URxvtExt(tt.extension, tt.params...)
		if result != tt.expected {
			t.Errorf("URxvtExt(%q, %v) = %q; want %q", tt.extension, tt.params, result, tt.expected)
		}
	}
}
