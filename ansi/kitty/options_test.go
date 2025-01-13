package kitty

import (
	"reflect"
	"sort"
	"testing"
)

func TestOptions_Options(t *testing.T) {
	tests := []struct {
		name     string
		options  Options
		expected []string
	}{
		{
			name:     "default options",
			options:  Options{},
			expected: []string{}, // Default values don't generate options
		},
		{
			name: "basic transmission options",
			options: Options{
				Format: PNG,
				ID:     1,
				Action: TransmitAndPut,
			},
			expected: []string{
				"f=100",
				"i=1",
				"a=T",
			},
		},
		{
			name: "display options",
			options: Options{
				X:      100,
				Y:      200,
				Z:      3,
				Width:  400,
				Height: 300,
			},
			expected: []string{
				"x=100",
				"y=200",
				"z=3",
				"w=400",
				"h=300",
			},
		},
		{
			name: "compression and chunking",
			options: Options{
				Compression: Zlib,
				Chunk:       true,
				Size:        1024,
			},
			expected: []string{
				"S=1024",
				"o=z",
			},
		},
		{
			name: "delete options",
			options: Options{
				Delete:          DeleteID,
				DeleteResources: true,
			},
			expected: []string{
				"d=I", // Uppercase due to DeleteResources being true
			},
		},
		{
			name: "virtual placement",
			options: Options{
				VirtualPlacement:  true,
				ParentID:          5,
				ParentPlacementID: 2,
			},
			expected: []string{
				"U=1",
				"P=5",
				"Q=2",
			},
		},
		{
			name: "cell positioning",
			options: Options{
				OffsetX: 10,
				OffsetY: 20,
				Columns: 80,
				Rows:    24,
			},
			expected: []string{
				"X=10",
				"Y=20",
				"c=80",
				"r=24",
			},
		},
		{
			name: "transmission details",
			options: Options{
				Transmission: File,
				File:         "/tmp/image.png",
				Offset:       100,
				Number:       2,
				PlacementID:  3,
			},
			expected: []string{
				"p=3",
				"I=2",
				"t=f",
				"O=100",
			},
		},
		{
			name: "quiet mode and format",
			options: Options{
				Quite:  2,
				Format: RGB,
			},
			expected: []string{
				"f=24",
				"q=2",
			},
		},
		{
			name: "all zero values",
			options: Options{
				Format: 0,
				Action: 0,
				Delete: 0,
			},
			expected: []string{}, // Should use defaults and not generate options
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.options.Options()

			// Sort both slices to ensure consistent comparison
			sortStrings(got)
			sortStrings(tt.expected)

			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("Options.Options() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestOptions_Validation(t *testing.T) {
	tests := []struct {
		name    string
		options Options
		check   func([]string) bool
	}{
		{
			name: "format validation",
			options: Options{
				Format: 999, // Invalid format
			},
			check: func(opts []string) bool {
				// Should still output the format even if invalid
				return containsOption(opts, "f=999")
			},
		},
		{
			name: "delete with resources",
			options: Options{
				Delete:          DeleteID,
				DeleteResources: true,
			},
			check: func(opts []string) bool {
				// Should be uppercase when DeleteResources is true
				return containsOption(opts, "d=I")
			},
		},
		{
			name: "transmission with file",
			options: Options{
				File: "/tmp/test.png",
			},
			check: func(opts []string) bool {
				return containsOption(opts, "t=f")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.options.Options()
			if !tt.check(got) {
				t.Errorf("Options validation failed for %s: %v", tt.name, got)
			}
		})
	}
}

// Helper functions

func sortStrings(s []string) {
	sort.Strings(s)
}

func containsOption(opts []string, target string) bool {
	for _, opt := range opts {
		if opt == target {
			return true
		}
	}
	return false
}
