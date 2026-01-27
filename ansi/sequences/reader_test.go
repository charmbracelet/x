package sequences_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi/sequences"
)

func TestScanner(t *testing.T) {
	cases := []struct {
		name          string
		input         string
		expected      []string
		expectedWidth []int
	}{
		{
			name:  "simple text",
			input: "Hello, World!",
			expected: []string{
				"H",
				"e",
				"l",
				"l",
				"o",
				",",
				" ",
				"W",
				"o",
				"r",
				"l",
				"d",
				"!",
			},
			expectedWidth: []int{
				1,
				1,
				1,
				1,
				1,
				1,
				1,
				1,
				1,
				1,
				1,
				1,
				1,
			},
		},
		{
			name:  "text with escape sequences",
			input: "\x1b[31mRed Text\x1b[0m Normal Text",
			expected: []string{
				"\x1b[31m",
				"R",
				"e",
				"d",
				" ",
				"T",
				"e",
				"x",
				"t",
				"\x1b[0m",
				" ",
				"N",
				"o",
				"r",
				"m",
				"a",
				"l",
				" ",
				"T",
				"e",
				"x",
				"t",
			},
			expectedWidth: []int{
				0,
				1,
				1,
				1,
				1,
				1,
				1,
				1,
				1,
				0,
				1,
				1,
				1,
				1,
				1,
				1,
				1,
				1,
				1,
				1,
				1,
				1,
			},
		},
		{
			name:  "emoji and complex characters",
			input: "Hello, 🌍! 👩‍💻",
			expected: []string{
				"H",
				"e",
				"l",
				"l",
				"o",
				",",
				" ",
				"🌍",
				"!",
				" ",
				"👩‍💻",
			},
			expectedWidth: []int{
				1,
				1,
				1,
				1,
				1,
				1,
				1,
				2,
				1,
				1,
				2,
			},
		},
		{
			name:  "mixed content",
			input: "\x1b[32mGreen 🌿 Text\x1b[0m and more text 👨‍🚀 \x1bP123qtext\x1b\\\x7fabc",
			expected: []string{
				"\x1b[32m",
				"G",
				"r",
				"e",
				"e",
				"n",
				" ",
				"🌿",
				" ",
				"T",
				"e",
				"x",
				"t",
				"\x1b[0m",
				" ",
				"a",
				"n",
				"d",
				" ",
				"m",
				"o",
				"r",
				"e",
				" ",
				"t",
				"e",
				"x",
				"t",
				" ",
				"👨‍🚀",
				" ",
				"\x1bP123qtext\x1b\\",
				"\x7f",
				"a",
				"b",
				"c",
			},
			expectedWidth: []int{
				0,
				1,
				1,
				1,
				1,
				1,
				1,
				2,
				1,
				1,
				1,
				1,
				1,
				0,
				1,
				1,
				1,
				1,
				1,
				1,
				1,
				1,
				1,
				1,
				1,
				1,
				1,
				1,
				1,
				2,
				1,
				0,
				0,
				1,
				1,
				1,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			scanner := sequences.FromReader(strings.NewReader(tc.input))
			var results []string
			var widths []int
			for scanner.Scan() {
				results = append(results, scanner.Text())
				widths = append(widths, scanner.Width())
			}
			if err := scanner.Err(); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(results, tc.expected) {
				t.Errorf("expected %q, got %q", tc.expected, results)
			}
			if !reflect.DeepEqual(widths, tc.expectedWidth) {
				t.Errorf("expected widths %v, got %v", tc.expectedWidth, widths)
			}
		})
	}
}
