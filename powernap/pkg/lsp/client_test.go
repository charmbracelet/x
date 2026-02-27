package lsp

import (
	"context"
	"sync"
	"testing"

	"github.com/charmbracelet/x/powernap/pkg/lsp/protocol"
)

func TestPositionToByteOffset(t *testing.T) {
	tests := []struct {
		name      string
		lineText  string
		utf16Char uint32
		expected  int
	}{
		{
			name:      "ASCII only",
			lineText:  "hello world",
			utf16Char: 6,
			expected:  6,
		},
		{
			name:      "CJK characters (3 bytes each in UTF-8, 1 UTF-16 unit)",
			lineText:  "你好world",
			utf16Char: 2,
			expected:  6,
		},
		{
			name:      "CJK - position after CJK",
			lineText:  "var x = \"你好world\"",
			utf16Char: 11,
			expected:  15,
		},
		{
			name:      "Emoji (4 bytes in UTF-8, 2 UTF-16 units)",
			lineText:  "👋hello",
			utf16Char: 2,
			expected:  4,
		},
		{
			name:      "Multiple emoji",
			lineText:  "👋👋world",
			utf16Char: 4,
			expected:  8,
		},
		{
			name:      "Mixed content",
			lineText:  "Hello👋你好",
			utf16Char: 8,
			expected:  12,
		},
		{
			name:      "Position 0",
			lineText:  "hello",
			utf16Char: 0,
			expected:  0,
		},
		{
			name:      "Position beyond end",
			lineText:  "hi",
			utf16Char: 100,
			expected:  2,
		},
		{
			name:      "Empty string",
			lineText:  "",
			utf16Char: 0,
			expected:  0,
		},
		{
			name:      "Surrogate pair at start",
			lineText:  "𐐷hello",
			utf16Char: 2,
			expected:  4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PositionToByteOffset(tt.lineText, tt.utf16Char)
			if result != tt.expected {
				t.Errorf("PositionToByteOffset(%q, %d) = %d, want %d",
					tt.lineText, tt.utf16Char, result, tt.expected)
			}
		})
	}
}

func TestParseOffsetEncoding(t *testing.T) {
	tests := []struct {
		name             string
		positionEncoding *protocol.PositionEncodingKind
		offsetEncoding   string
		expected         OffsetEncoding
	}{
		{
			name:           "defaults to UTF16 when no encoding provided",
			offsetEncoding: "",
			expected:       UTF16,
		},
		{
			name:           "uses legacy offsetEncoding",
			offsetEncoding: "utf-8",
			expected:       UTF8,
		},
		{
			name:           "uses positionEncoding when provided",
			positionEncoding: func() *protocol.PositionEncodingKind {
				v := protocol.UTF8
				return &v
			}(),
			offsetEncoding: "utf-16",
			expected:       UTF8,
		},
		{
			name:           "falls back to legacy offsetEncoding when positionEncoding unknown",
			positionEncoding: func() *protocol.PositionEncodingKind {
				v := protocol.PositionEncodingKind("unknown")
				return &v
			}(),
			offsetEncoding: "utf-32",
			expected:       UTF32,
		},
		{
			name:           "defaults to UTF16 when both encodings unknown",
			positionEncoding: func() *protocol.PositionEncodingKind {
				v := protocol.PositionEncodingKind("unknown")
				return &v
			}(),
			offsetEncoding: "also-unknown",
			expected:       UTF16,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseOffsetEncoding(tt.positionEncoding, tt.offsetEncoding)
			if result != tt.expected {
				t.Errorf("parseOffsetEncoding(%v, %q) = %v, want %v", tt.positionEncoding, tt.offsetEncoding, result, tt.expected)
			}
		})
	}
}

func TestProcessCloser_ConcurrentClose(t *testing.T) {
	config := ClientConfig{
		Command: "cat",
		RootURI: "file:///tmp",
	}

	stream, err := startServerProcess(t.Context(), config)
	if err != nil {
		t.Fatalf("failed to start process: %v", err)
	}

	var wg sync.WaitGroup
	for range 10 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = stream.Close()
		}()
	}
	wg.Wait()
}

func TestProcessCloser_CloseAfterContextCancel(t *testing.T) {
	config := ClientConfig{
		Command: "cat",
		RootURI: "file:///tmp",
	}

	ctx, cancel := context.WithCancel(t.Context())
	stream, err := startServerProcess(ctx, config)
	if err != nil {
		t.Fatalf("failed to start process: %v", err)
	}

	cancel()
	_ = stream.Close()
}

func TestProcessCloser_ConcurrentCancelAndClose(t *testing.T) {
	config := ClientConfig{
		Command: "cat",
		RootURI: "file:///tmp",
	}

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	stream, err := startServerProcess(ctx, config)
	if err != nil {
		t.Fatalf("failed to start process: %v", err)
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		cancel()
	}()
	go func() {
		defer wg.Done()
		_ = stream.Close()
	}()
	wg.Wait()
}
