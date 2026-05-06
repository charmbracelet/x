package protocol

import (
	"path/filepath"
	"runtime"
	"testing"
)

// TestDocumentURIPath exercises DocumentURI.Path across the URI shapes
// that real LSP clients send. The Windows-drive cases are the main
// thing being locked in: previously DocumentURI("file:///C:/dev/foo").Path()
// returned "\C:\dev\foo" (with a stray leading separator) because the
// drive-letter check in filename() never matched the "/C:/..." form
// produced by url.URL.Path.
func TestDocumentURIPath(t *testing.T) {
	type testCase struct {
		name string
		uri  DocumentURI
		// want is the expected path in forward-slash form. The test
		// converts it via filepath.FromSlash before comparing, so a
		// case can express "C:/dev/foo" once and have it match
		// "C:\dev\foo" on Windows and "C:/dev/foo" on POSIX.
		want string
	}
	cases := []testCase{
		{name: "empty", uri: "", want: ""},
		{name: "POSIX absolute", uri: "file:///home/sven/foo.go", want: "/home/sven/foo.go"},
		{name: "POSIX with space", uri: "file:///home/sven/My%20Code/foo.go", want: "/home/sven/My Code/foo.go"},
	}
	if runtime.GOOS == "windows" {
		// Windows-only: the drive-path normalization runs through
		// filepath.VolumeName, which only recognizes drive letters on
		// Windows. On POSIX, "C:/dev/foo" comes back as "/C:/dev/foo"
		// because there's no notion of a drive there - which is fine
		// for callers, because a POSIX host will never be asked to
		// open such a path anyway.
		cases = append(
			cases,
			testCase{name: "Windows drive", uri: "file:///C:/dev/foo", want: "C:/dev/foo"},
			testCase{name: "Windows drive lowercase", uri: "file:///c:/dev/foo", want: "C:/dev/foo"},
			testCase{name: "Windows drive percent-encoded colon", uri: "file:///C%3A/dev/foo", want: "C:/dev/foo"},
			testCase{name: "Windows drive with space", uri: "file:///C:/Program%20Files/foo.exe", want: "C:/Program Files/foo.exe"},
		)
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := tc.uri.Path()
			if err != nil {
				t.Fatalf("Path() error: %v", err)
			}
			want := filepath.FromSlash(tc.want)
			if got != want {
				t.Errorf("Path() = %q, want %q", got, want)
			}
		})
	}
}

// TestDocumentURIRoundTrip checks that URIFromPath and DocumentURI.Path
// are inverses for the path shapes the public API supports.
func TestDocumentURIRoundTrip(t *testing.T) {
	var paths []string
	if runtime.GOOS == "windows" {
		paths = []string{
			`C:\dev\foo`,
			`C:\Program Files\foo.exe`,
		}
	} else {
		paths = []string{
			"/home/sven/foo.go",
			"/tmp/My Code/foo.go",
		}
	}
	for _, p := range paths {
		t.Run(p, func(t *testing.T) {
			got, err := URIFromPath(p).Path()
			if err != nil {
				t.Fatalf("URIFromPath(%q).Path() error: %v", p, err)
			}
			if got != p {
				t.Errorf("round-trip: got %q, want %q", got, p)
			}
		})
	}
}
