// Package snapshot provides helpers for working with terminal snapshots.
package snapshot

import (
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/charmbracelet/x/vttest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// update indicates whether to update testdata files.
var update = flag.Bool("update", false, "update testdata files")

// Snapshotter is an interface for types that can produce snapshots of their state.
type Snapshotter interface {
	Snapshot() vttest.Snapshot
}

// Imager represents types that can produce an image representation of their state.
type Imager interface {
	Image() image.Image
}

// TestdataEqualf compares the snapshot of the given [Snapshotter] with the
// expected snapshot stored in the "testdata" directory, using the provided
// format and arguments to construct the filename.
//
// If the snapshots do not match, it reports an error on the testing.TB.
func TestdataEqualf(tb testing.TB, expectedNameSuffix string, actual Snapshotter, format string, args ...any) {
	tb.Helper()
	testdataEq(tb, expectedNameSuffix, actual, func(expected, actual vttest.Snapshot) {
		assert.Equalf(tb, expected, actual, format, args...)
	})
}

// TestdataEqual compares the snapshot of the given [Snapshotter] with the
// expected snapshot stored in the "testdata" directory, using the test name
// and the provided expectedNameSuffix to construct the filename.
//
// If the snapshots do not match, it reports an error on the testing.TB.
func TestdataEqual(tb testing.TB, expectedNameSuffix string, actual Snapshotter, msgAndArgs ...any) {
	tb.Helper()
	testdataEq(tb, expectedNameSuffix, actual, func(expected, actual vttest.Snapshot) {
		assert.Equal(tb, expected, actual, msgAndArgs...)
	})
}

// TestdataRequireEqualf compares the snapshot of the given [Snapshotter] with the
// expected snapshot stored in the "testdata" directory, using the provided
// format and arguments to construct the filename.
//
// If the snapshots do not match, it fails the test immediately.
func TestdataRequireEqualf(tb testing.TB, expectedNameSuffix string, actual Snapshotter, format string, args ...any) {
	tb.Helper()
	testdataEq(tb, expectedNameSuffix, actual, func(expected, actual vttest.Snapshot) {
		require.Equalf(tb, expected, actual, format, args...)
	})
}

// TestdataRequireEqual compares the snapshot of the given [Snapshotter] with the
// expected snapshot stored in the "testdata" directory, using the test name
// and the provided expectedNameSuffix to construct the filename.
//
// If the snapshots do not match, it fails the test immediately.
func TestdataRequireEqual(tb testing.TB, expectedNameSuffix string, actual Snapshotter, msgAndArgs ...any) {
	tb.Helper()
	testdataEq(tb, expectedNameSuffix, actual, func(expected, actual vttest.Snapshot) {
		require.Equal(tb, expected, actual, msgAndArgs...)
	})
}

func testdataEq(tb testing.TB, expectedNameSuffix string, actual Snapshotter, cb func(expected, actual vttest.Snapshot)) {
	tb.Helper()

	actualSnap := actual.Snapshot()
	fp := filepath.Join("testdata", fmt.Sprintf("%s_%s.json", tb.Name(), expectedNameSuffix))
	if *update {
		if err := os.MkdirAll(filepath.Dir(fp), 0o750); err != nil { //nolint: mnd
			tb.Fatal(err)
		}

		f, err := os.Create(fp)
		if err != nil {
			tb.Fatalf("failed to create snapshot file: %v", err)
		}
		defer f.Close()

		if err := json.NewEncoder(f).Encode(actualSnap); err != nil {
			tb.Fatalf("failed to encode snapshot: %v", err)
		}

		if imgSnap, ok := actual.(Imager); ok {
			// Create image representation
			img := imgSnap.Image()
			fp := filepath.Join("testdata", fmt.Sprintf("%s_%s.png", tb.Name(), expectedNameSuffix))
			imgFile, err := os.Create(fp)
			if err != nil {
				tb.Fatalf("failed to create image file: %v", err)
			}
			defer imgFile.Close()

			if err := png.Encode(imgFile, img); err != nil {
				tb.Fatalf("failed to encode image: %v", err)
			}
		}
	}

	expectedSnapFile, err := os.Open(fp)
	if err != nil {
		tb.Fatalf("failed to read snapshot file: %v", err)
	}

	var expectedSnap vttest.Snapshot
	if err := json.NewDecoder(expectedSnapFile).Decode(&expectedSnap); err != nil {
		tb.Fatalf("failed to decode snapshot: %v", err)
	}

	cb(expectedSnap, actualSnap)
}
