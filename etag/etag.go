// Package etag provides utilities for generating and handling ETag headers in
// HTTP requests and responses.
package etag

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"strings"
)

// Of returns the etag for the given data.
func Of(data []byte) string {
	hash := sha256.Sum256(data)
	return fmt.Sprintf(`%x`, hash[:16])
}

// Request sets the `If-None-Match` header in the given request, appropriately
// quoting the etag value.
func Request(req *http.Request, etag string) {
	if etag == "" {
		return
	}
	req.Header.Add("If-None-Match", fmt.Sprintf(`"%s"`, etag))
}

// Response sets the `ETag` header in the given response writer, appropriately
// quoting the etag value.
func Response(w http.ResponseWriter, etag string) {
	if etag == "" {
		return
	}
	w.Header().Set("ETag", fmt.Sprintf(`"%s"`, etag))
}

// Matches checks if the given request has `If-None-Match` header matching the
// given etag.
func Matches(r *http.Request, etag string) bool {
	header := r.Header.Get("If-None-Match")
	if header == "" || etag == "" {
		return false
	}
	return unquote(header) == unquote(etag)
}

func unquote(s string) string {
	return strings.TrimSuffix(strings.TrimPrefix(s, `"`), `"`)
}
