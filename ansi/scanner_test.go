package ansi

import (
	"fmt"
	"strings"
	"testing"
)

type scanResult struct {
	kind ScannerToken
	text string
}

func (sr scanResult) String() string {
	return fmt.Sprintf("{%d %q}", sr.kind, strings.ReplaceAll(sr.text, "\x1b", "\\x1b"))
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
				{TextToken, "I"},
				{SpaceToken, " "},
				{TextToken, "really"},
				{SpaceToken, " "},
				{ControlToken, "\x1B[38;2;249;38;114m"},
				{TextToken, "love"},
				{ControlToken, "\x1B[0m"},
				{SpaceToken, " "},
				{TextToken, "Go!"},
			},
		},
		{
			name:  "passthrough",
			input: "hello world",
			expected: []scanResult{
				{TextToken, "hello"},
				{SpaceToken, " "},
				{TextToken, "world"},
			},
		},
		{
			name:  "asian",
			input: "こんにち",
			expected: []scanResult{
				{TextToken, "こんにち"}},
		},
		{
			name:  "emoji",
			input: "😃👰🏻‍♀️🫧",
			expected: []scanResult{
				{TextToken, "😃👰🏻‍♀️🫧"},
			},
		},
		{
			name:  "long style",
			input: "\x1B[38;2;249;38;114ma really long string\x1B[0m",
			expected: []scanResult{
				{ControlToken, "\x1B[38;2;249;38;114m"},
				{TextToken, "a"},
				{SpaceToken, " "},
				{TextToken, "really"},
				{SpaceToken, " "},
				{TextToken, "long"},
				{SpaceToken, " "},
				{TextToken, "string"},
				{ControlToken, "\x1B[0m"},
			},
		},
		{
			name:  "long style nbsp",
			input: "\x1B[38;2;249;38;114ma really\u00a0long string\x1B[0m",
			expected: []scanResult{
				{ControlToken, "\x1b[38;2;249;38;114m"},
				{TextToken, "a"},
				{SpaceToken, " "},
				{TextToken, "really"},
				{SpaceToken, "\u00a0"},
				{TextToken, "long"},
				{SpaceToken, " "},
				{TextToken, "string"},
				{ControlToken, "\x1b[0m"},
			},
		},
		{
			name:  "exact",
			input: "\x1b[91mfoo\x1b[0",
			expected: []scanResult{
				{ControlToken, "\x1b[91m"},
				{TextToken, "foo"},
				{ControlToken, "\x1b[0"},
			},
		},
		{
			name:  "newline",
			input: "\x1b[91mfoo\nbar\x1b[0",
			expected: []scanResult{
				{ControlToken, "\x1b[91m"},
				{TextToken, "foo"},
				{LineToken, "\n"},
				{TextToken, "bar"},
				{ControlToken, "\x1b[0"},
			},
		},
		{
			name:  "carriage return",
			input: "\x1b[91mfoo\rbar\x1b[0",
			expected: []scanResult{
				{ControlToken, "\x1b[91m"},
				{TextToken, "foo"},
				{LineToken, "\r"},
				{TextToken, "bar"},
				{ControlToken, "\x1b[0"},
			},
		},
		{
			name:  "return & newline",
			input: "\x1b[91mfoo\r\nbar\x1b[0",
			expected: []scanResult{
				{ControlToken, "\x1b[91m"},
				{TextToken, "foo"},
				{LineToken, "\r\n"},
				{TextToken, "bar"},
				{ControlToken, "\x1b[0"},
			},
		},
		{
			name:  "extra extra return & newline",
			input: "\x1b[91mfoo\r\r\r\nbar\x1b[0",
			expected: []scanResult{
				{ControlToken, "\x1b[91m"},
				{TextToken, "foo"},
				{LineToken, "\r\r\r\n"},
				{TextToken, "bar"},
				{ControlToken, "\x1b[0"},
			},
		},
		{
			name:  "spaces return & newline",
			input: "\x1b[91mfoo \r\n bar \x1b[0",
			expected: []scanResult{
				{ControlToken, "\x1b[91m"},
				{TextToken, "foo"},
				{SpaceToken, " "},
				{LineToken, "\r\n"},
				{SpaceToken, " "},
				{TextToken, "bar"},
				{SpaceToken, " "},
				{ControlToken, "\x1b[0"},
			},
		},
		{
			name:  "multiple newlines emitted separately",
			input: "\x1b[91mfoo\n\n\r\nbar\x1b[0",
			expected: []scanResult{
				{ControlToken, "\x1b[91m"},
				{TextToken, "foo"},
				{LineToken, "\n"},
				{LineToken, "\n"},
				{LineToken, "\r\n"},
				{TextToken, "bar"},
				{ControlToken, "\x1b[0"},
			},
		},
	}

	for i, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := NewScanner(tc.input, ScanLines, ScanWords)
			for j, expected := range tc.expected {
				k, s := s.Scan()
				if k != expected.kind || s != expected.text {
					t.Errorf("case %d, input %q, expected %d %s, got %s", i+1, tc.input, j, expected, scanResult{k, s})
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
				{TextToken, "I really "},
				{ControlToken, "\x1B[38;2;249;38;114m"},
				{TextToken, "love"},
				{ControlToken, "\x1B[0m"},
				{TextToken, " Go!"},
			},
		},
		{
			name:  "passthrough",
			input: "hello world",
			expected: []scanResult{
				{TextToken, "hello world"},
			},
		},
		{
			name:  "long style",
			input: "\x1B[38;2;249;38;114ma really long string\x1B[0m",
			expected: []scanResult{
				{ControlToken, "\x1B[38;2;249;38;114m"},
				{TextToken, "a really long string"},
				{ControlToken, "\x1B[0m"},
			},
		},
		{
			name:  "long style nbsp",
			input: "\x1B[38;2;249;38;114ma really\u00a0long string\x1B[0m",
			expected: []scanResult{
				{ControlToken, "\x1b[38;2;249;38;114m"},
				{TextToken, "a really\u00a0long string"},
				{ControlToken, "\x1b[0m"},
			},
		},
		{
			name:  "newline",
			input: "\x1b[91mfoo\nbar\x1b[0",
			expected: []scanResult{
				{ControlToken, "\x1b[91m"},
				{TextToken, "foo"},
				{LineToken, "\n"},
				{TextToken, "bar"},
				{ControlToken, "\x1b[0"},
			},
		},
		{
			name:  "spaces return & newline",
			input: "\x1b[91mfoo \r\n bar \x1b[0",
			expected: []scanResult{
				{ControlToken, "\x1b[91m"},
				{TextToken, "foo "},
				{LineToken, "\r\n"},
				{TextToken, " bar "},
				{ControlToken, "\x1b[0"},
			},
		},
		{
			name:  "spaces return",
			input: "\x1b[91mfoo \r bar \x1b[0",
			expected: []scanResult{
				{ControlToken, "\x1b[91m"},
				{TextToken, "foo "},
				{LineToken, "\r"},
				{TextToken, " bar "},
				{ControlToken, "\x1b[0"},
			},
		},
	}

	for i, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := NewScanner(tc.input, ScanLines)
			for j, expected := range tc.expected {
				k, s := s.Scan()
				if k != expected.kind || s != expected.text {
					t.Errorf("case %d, input %q, expected %d %s, got %s", i+1, tc.input, j, expected, scanResult{k, s})
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
				{TextToken, "I really "},
				{ControlToken, "\x1B[38;2;249;38;114m"},
				{TextToken, "love"},
				{ControlToken, "\x1B[0m"},
				{TextToken, " Go!"},
			},
		},
		{
			name:  "passthrough",
			input: "hello world",
			expected: []scanResult{
				{TextToken, "hello world"},
			},
		},
		{
			name:  "long style",
			input: "\x1B[38;2;249;38;114ma really long string\x1B[0m",
			expected: []scanResult{
				{ControlToken, "\x1B[38;2;249;38;114m"},
				{TextToken, "a really long string"},
				{ControlToken, "\x1B[0m"},
			},
		},
		{
			name:  "long style nbsp",
			input: "\x1B[38;2;249;38;114ma really\u00a0long string\x1B[0m",
			expected: []scanResult{
				{ControlToken, "\x1b[38;2;249;38;114m"},
				{TextToken, "a really\u00a0long string"},
				{ControlToken, "\x1b[0m"},
			},
		},
		{
			name:  "spaces return & newline",
			input: "\x1b[91mfoo \r\n bar \x1b[0",
			expected: []scanResult{
				{ControlToken, "\x1b[91m"},
				{TextToken, "foo \r\n bar "},
				{ControlToken, "\x1b[0"},
			},
		},
	}

	for i, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := NewScanner(tc.input)
			for j, expected := range tc.expected {
				k, s := s.Scan()
				if k != expected.kind || s != expected.text {
					t.Errorf("case %d, input %q, expected %d %s, got %s", i+1, tc.input, j, expected, scanResult{k, s})
				}
			}
		})
	}
}
