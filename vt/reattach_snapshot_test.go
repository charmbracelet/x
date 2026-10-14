package vt

import (
	"strings"
	"testing"
)

func snapshotString(t *testing.T, e *Emulator) string {
	t.Helper()
	snapshot, err := e.ReattachSnapshot()
	if err != nil {
		t.Fatalf("ReattachSnapshot() error = %v", err)
	}
	return string(snapshot)
}

func TestReattachSnapshotPreservesSoftAndHardBoundaries(t *testing.T) {
	t.Run("terminal autowrap stays soft", func(t *testing.T) {
		e := NewEmulator(5, 3)
		e.WriteString("abcdefghij")

		got := snapshotString(t, e)
		if !strings.Contains(got, "abcdefghij") {
			t.Fatalf("snapshot = %q, want contiguous soft-wrapped text", got)
		}
		if strings.Contains(got, "abcde\nfghij") || strings.Contains(got, "abcde\r\nfghij") {
			t.Fatalf("snapshot = %q, soft wrap was serialized as a hard break", got)
		}
	})

	t.Run("application newline stays hard", func(t *testing.T) {
		e := NewEmulator(5, 3)
		e.WriteString("abcde\r\nfghij")

		got := snapshotString(t, e)
		if !strings.Contains(got, "abcde\r\nfghij") {
			t.Fatalf("snapshot = %q, want application newline to remain hard", got)
		}
	})
}

func TestResizeReflowsPrimaryHistoryAndCursor(t *testing.T) {
	e := NewEmulator(5, 2)
	e.WriteString("abcdefghijk")

	if e.ScrollbackLen() == 0 {
		t.Fatal("test setup did not create scrollback")
	}

	e.Resize(10, 2)

	if got := e.ScrollbackLen(); got != 0 {
		t.Fatalf("scrollback len after widening = %d, want 0", got)
	}
	if got := e.Render(); !strings.Contains(got, "abcdefghij\nk") {
		t.Fatalf("render after widening = %q, want reflowed primary history", got)
	}
	if got := e.CursorPosition(); got.X != 1 || got.Y != 1 {
		t.Fatalf("cursor after widening = %v, want (1,1)", got)
	}

	e.Resize(5, 3)
	got := snapshotString(t, e)
	if !strings.Contains(got, "abcdefghijk") {
		t.Fatalf("snapshot after shrink = %q, want logical text preserved", got)
	}
}

func TestResizePreservesHardBoundariesAndUnicodeCells(t *testing.T) {
	e := NewEmulator(6, 3)
	e.WriteString("ab界e\u0301z\r\nsecond")

	e.Resize(12, 3)
	got := snapshotString(t, e)
	if !strings.Contains(got, "ab界e\u0301z\r\nsecond") {
		t.Fatalf("snapshot after widening = %q, want Unicode cells and hard newline preserved", got)
	}

	e.Resize(4, 4)
	got = snapshotString(t, e)
	if !strings.Contains(got, "ab界e\u0301z") {
		t.Fatalf("snapshot after narrowing = %q, want Unicode logical row preserved", got)
	}
	if !strings.Contains(got, "\r\nsecond") {
		t.Fatalf("snapshot after narrowing = %q, want hard newline preserved", got)
	}
}

func TestScrollbackCapMarksTruncatedSoftWrapHead(t *testing.T) {
	e := NewEmulator(5, 1)
	e.SetScrollbackSize(1)
	e.WriteString("abcdefghijk")

	rows := e.scr.scrollback.semanticRows()
	if len(rows) != 1 {
		t.Fatalf("scrollback rows = %d, want 1", len(rows))
	}
	if rows[0].boundary != boundaryTruncatedHead {
		t.Fatalf("retained head boundary = %v, want truncated head", rows[0].boundary)
	}
	got := snapshotString(t, e)
	if strings.Contains(got, "abcde") || !strings.Contains(got, "fghijk") {
		t.Fatalf("snapshot = %q, want only retained logical suffix", got)
	}
}

func TestResizePreservesExactColumnPendingWrap(t *testing.T) {
	e := NewEmulator(5, 2)
	e.WriteString("abcdefghij")
	e.Resize(10, 2)
	e.WriteString("k")

	if got := snapshotString(t, e); !strings.Contains(got, "abcdefghijk") {
		t.Fatalf("snapshot = %q, want pending-wrap continuation after widening", got)
	}
	if got := e.CursorPosition(); got.X != 1 || got.Y != 1 {
		t.Fatalf("cursor = %v, want (1,1)", got)
	}
}
