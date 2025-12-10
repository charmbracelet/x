package etag

import (
	"crypto/sha256"
	"fmt"
	"net/http"
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
