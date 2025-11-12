package vcr

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"gopkg.in/dnaeon/go-vcr.v4/pkg/cassette"
	"gopkg.in/dnaeon/go-vcr.v4/pkg/recorder"
)

func customMatcher(t *testing.T) recorder.MatcherFunc {
	return func(r *http.Request, i cassette.Request) bool {
		if r.Body == nil || r.Body == http.NoBody {
			return cassette.DefaultMatcher(r, i)
		}
		if r.Method != i.Method || r.URL.String() != i.URL {
			return false
		}

		reqBody, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("vcr: failed to read request body")
		}
		_ = r.Body.Close()
		r.Body = io.NopCloser(bytes.NewBuffer(reqBody))

		// Some providers can sometimes generate JSON requests with keys in
		// a different order, which means a direct string comparison will fail.
		// Falling back to deserializing the content if we don't have a match.
		requestContent := normalizeLineEndings(reqBody)
		cassetteContent := normalizeLineEndings(i.Body)
		if requestContent == cassetteContent {
			return true
		}
		var content1, content2 any
		if err := json.Unmarshal([]byte(requestContent), &content1); err != nil {
			printDiff(t, requestContent, cassetteContent)
			return false
		}
		if err := json.Unmarshal([]byte(cassetteContent), &content2); err != nil {
			printDiff(t, requestContent, cassetteContent)
			return false
		}
		if isEqual := reflect.DeepEqual(content1, content2); !isEqual {
			printDiff(t, requestContent, cassetteContent)
			return false
		}
		return true
	}
}

// normalizeLineEndings does not only replace `\r\n` into `\n`,
// but also replaces `\\r\\n` into `\\n`. That's because we want the content
// inside JSON string to be replaces as well.
func normalizeLineEndings[T string | []byte](s T) string {
	str := string(s)
	str = strings.ReplaceAll(str, "\r\n", "\n")
	str = strings.ReplaceAll(str, `\r\n`, `\n`)
	return str
}

func printDiff(t *testing.T, requestContent, cassetteContent string) {
	t.Logf("Request interaction not found for %q.\nDiff:\n%s", t.Name(), cmp.Diff(requestContent, cassetteContent))
}
