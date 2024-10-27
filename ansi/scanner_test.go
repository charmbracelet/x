package ansi

import (
	"fmt"
	"strings"
	"testing"
)

type scanResult struct {
	escape bool
	width  int
	text   string
}

func (sr scanResult) String() string {
	return fmt.Sprintf("%t %d %q", sr.escape, sr.width, strings.ReplaceAll(sr.text, "\x1b", "\\x1b"))
}

func TestScannerLinesWords(t *testing.T) {

	var testCases = []struct {
		name     string
		input    string
		expected []scanResult
	}{
		{
			name:  "simple",
			input: "I really \x1B[38;2;249;38;114mlove\x1B[0m Go!",
			expected: []scanResult{
				{false, 1, "I"},
				{false, 1, " "},
				{false, 6, "really"},
				{false, 1, " "},
				{true, 0, "\x1B[38;2;249;38;114m"},
				{false, 4, "love"},
				{true, 0, "\x1B[0m"},
				{false, 1, " "},
				{false, 3, "Go!"},
			},
		},
		{
			name:  "passthrough",
			input: "hello world",
			expected: []scanResult{
				{false, 5, "hello"},
				{false, 1, " "},
				{false, 5, "world"},
			},
		},
		{
			name:  "asian",
			input: "こんにち",
			expected: []scanResult{
				{false, 8, "こんにち"}},
		},
		{
			name:  "emoji",
			input: "😃👰🏻‍♀️🫧",
			expected: []scanResult{
				{false, 6, "😃👰🏻‍♀️🫧"},
			},
		},
		{
			name:  "long style",
			input: "\x1B[38;2;249;38;114ma really long string\x1B[0m",
			expected: []scanResult{
				{true, 0, "\x1B[38;2;249;38;114m"},
				{false, 1, "a"},
				{false, 1, " "},
				{false, 6, "really"},
				{false, 1, " "},
				{false, 4, "long"},
				{false, 1, " "},
				{false, 6, "string"},
				{true, 0, "\x1B[0m"},
			},
		},
		{
			name:  "long style nbsp",
			input: "\x1B[38;2;249;38;114ma really\u00a0long string\x1B[0m",
			expected: []scanResult{
				{true, 0, "\x1b[38;2;249;38;114m"},
				{false, 1, "a"},
				{false, 1, " "},
				{false, 6, "really"},
				{false, 1, "\u00a0"},
				{false, 4, "long"},
				{false, 1, " "},
				{false, 6, "string"},
				{true, 0, "\x1b[0m"},
			},
		},
		{
			name:  "exact",
			input: "\x1b[91mfoo\x1b[0",
			expected: []scanResult{
				{true, 0, "\x1b[91m"},
				{false, 3, "foo"},
				{true, 0, "\x1b[0"},
			},
		},
		{
			name:  "newline",
			input: "\x1b[91mfoo\nbar\x1b[0",
			expected: []scanResult{
				{true, 0, "\x1b[91m"},
				{false, 3, "foo"},
				{false, 1, "\n"},
				{false, 3, "bar"},
				{true, 0, "\x1b[0"},
			},
		},
		{
			name:  "carriage return",
			input: "\x1b[91mfoo\rbar\x1b[0",
			expected: []scanResult{
				{true, 0, "\x1b[91m"},
				{false, 3, "foo"},
				{false, 1, "\r"},
				{false, 3, "bar"},
				{true, 0, "\x1b[0"},
			},
		},
		{
			name:  "return & newline",
			input: "\x1b[91mfoo\r\nbar\x1b[0",
			expected: []scanResult{
				{true, 0, "\x1b[91m"},
				{false, 3, "foo"},
				{false, 2, "\r\n"},
				{false, 3, "bar"},
				{true, 0, "\x1b[0"},
			},
		},
		{
			name:  "extra extra return & newline",
			input: "\x1b[91mfoo\r\r\r\nbar\x1b[0",
			expected: []scanResult{
				{true, 0, "\x1b[91m"},
				{false, 3, "foo"},
				{false, 4, "\r\r\r\n"},
				{false, 3, "bar"},
				{true, 0, "\x1b[0"},
			},
		},
		{
			name:  "spaces return & newline",
			input: "\x1b[91mfoo \r\n bar \x1b[0",
			expected: []scanResult{
				{true, 0, "\x1b[91m"},
				{false, 3, "foo"},
				{false, 1, " "},
				{false, 2, "\r\n"},
				{false, 1, " "},
				{false, 3, "bar"},
				{false, 1, " "},
				{true, 0, "\x1b[0"},
			},
		},
		{
			name:  "multiple newlines emitted separately",
			input: "\x1b[91mfoo\n\n\r\nbar\x1b[0",
			expected: []scanResult{
				{true, 0, "\x1b[91m"},
				{false, 3, "foo"},
				{false, 1, "\n"},
				{false, 1, "\n"},
				{false, 2, "\r\n"},
				{false, 3, "bar"},
				{true, 0, "\x1b[0"},
			},
		},
		{
			name:  "whitespace with newlines",
			input: "   \n\n   \r\n   ",
			expected: []scanResult{
				{false, 3, "   "},
				{false, 1, "\n"},
				{false, 1, "\n"},
				{false, 3, "   "},
				{false, 2, "\r\n"},
				{false, 3, "   "},
			},
		},
	}

	for i, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := NewScanner(tc.input, ScanLines, ScanWords)
			for j, expected := range tc.expected {
				if !s.Scan() {
					t.Errorf("case %d, failed to scan", i+1)
				}
				isEscape := s.IsEscape()
				width := s.Width()
				text := s.Text()
				if isEscape != expected.escape || width != expected.width || text != expected.text {
					t.Errorf("case %d, input %q, expected %d %s, got %s", i+1, tc.input, j, expected, scanResult{isEscape, width, text})
				}
			}
		})
	}

}

func TestScannerLines(t *testing.T) {

	var testCases = []struct {
		name     string
		input    string
		expected []scanResult
	}{
		{
			name:  "simple",
			input: "I really \x1B[38;2;249;38;114mlove\x1B[0m Go!",
			expected: []scanResult{
				{false, 9, "I really "},
				{true, 0, "\x1B[38;2;249;38;114m"},
				{false, 4, "love"},
				{true, 0, "\x1B[0m"},
				{false, 4, " Go!"},
			},
		},
		{
			name:  "passthrough",
			input: "hello world",
			expected: []scanResult{
				{false, 11, "hello world"},
			},
		},
		{
			name:  "long style",
			input: "\x1B[38;2;249;38;114ma really long string\x1B[0m",
			expected: []scanResult{
				{true, 0, "\x1B[38;2;249;38;114m"},
				{false, 20, "a really long string"},
				{true, 0, "\x1B[0m"},
			},
		},
		{
			name:  "long style nbsp",
			input: "\x1B[38;2;249;38;114ma really\u00a0long string\x1B[0m",
			expected: []scanResult{
				{true, 0, "\x1b[38;2;249;38;114m"},
				{false, 20, "a really\u00a0long string"},
				{true, 0, "\x1b[0m"},
			},
		},
		{
			name:  "newline",
			input: "\x1b[91mfoo\nbar\x1b[0",
			expected: []scanResult{
				{true, 0, "\x1b[91m"},
				{false, 3, "foo"},
				{false, 1, "\n"},
				{false, 3, "bar"},
				{true, 0, "\x1b[0"},
			},
		},
		{
			name:  "spaces return & newline",
			input: "\x1b[91mfoo \r\n bar \x1b[0",
			expected: []scanResult{
				{true, 0, "\x1b[91m"},
				{false, 4, "foo "},
				{false, 2, "\r\n"},
				{false, 5, " bar "},
				{true, 0, "\x1b[0"},
			},
		},
		{
			name:  "spaces return",
			input: "\x1b[91mfoo \r bar \x1b[0",
			expected: []scanResult{
				{true, 0, "\x1b[91m"},
				{false, 4, "foo "},
				{false, 1, "\r"},
				{false, 5, " bar "},
				{true, 0, "\x1b[0"},
			},
		},
	}

	for i, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := NewScanner(tc.input, ScanLines)
			for j, expected := range tc.expected {
				if !s.Scan() {
					t.Errorf("case %d, failed to scan", i+1)
				}
				isEscape := s.IsEscape()
				width := s.Width()
				text := s.Text()
				if isEscape != expected.escape || width != expected.width || text != expected.text {
					t.Errorf("case %d, input %q, expected %d %s, got %s", i+1, tc.input, j, expected, scanResult{isEscape, width, text})
				}
			}
		})
	}
}

func TestScanner(t *testing.T) {

	var testCases = []struct {
		name     string
		input    string
		expected []scanResult
	}{
		{
			name:  "simple",
			input: "I really \x1B[38;2;249;38;114mlove\x1B[0m Go!",
			expected: []scanResult{
				{false, 9, "I really "},
				{true, 0, "\x1B[38;2;249;38;114m"},
				{false, 4, "love"},
				{true, 0, "\x1B[0m"},
				{false, 4, " Go!"},
			},
		},
		{
			name:  "passthrough",
			input: "hello world",
			expected: []scanResult{
				{false, 11, "hello world"},
			},
		},
		{
			name:  "long style",
			input: "\x1B[38;2;249;38;114ma really long string\x1B[0m",
			expected: []scanResult{
				{true, 0, "\x1B[38;2;249;38;114m"},
				{false, 20, "a really long string"},
				{true, 0, "\x1B[0m"},
			},
		},
		{
			name:  "long style nbsp",
			input: "\x1B[38;2;249;38;114ma really\u00a0long string\x1B[0m",
			expected: []scanResult{
				{true, 0, "\x1b[38;2;249;38;114m"},
				{false, 20, "a really\u00a0long string"},
				{true, 0, "\x1b[0m"},
			},
		},
		{
			name:  "spaces return & newline",
			input: "\x1b[91mfoo \r\n bar \x1b[0",
			expected: []scanResult{
				{true, 0, "\x1b[91m"},
				{false, 11, "foo \r\n bar "},
				{true, 0, "\x1b[0"},
			},
		},
	}

	for i, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := NewScanner(tc.input)
			for j, expected := range tc.expected {
				if !s.Scan() {
					t.Errorf("case %d, failed to scan", i+1)
				}
				isEscape := s.IsEscape()
				width := s.Width()
				text := s.Text()
				if isEscape != expected.escape || width != expected.width || text != expected.text {
					t.Errorf("case %d, input %q, expected %d %s, got %s", i+1, tc.input, j, expected, scanResult{isEscape, width, text})
				}
			}
		})
	}
}
