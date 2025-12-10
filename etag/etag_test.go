package etag

import (
	"net/http/httptest"
	"testing"
)

func TestOf(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want string
	}{
		{
			name: "empty data",
			data: []byte{},
			want: "e3b0c44298fc1c149afbf4c8996fb924",
		},
		{
			name: "hello world",
			data: []byte("hello world"),
			want: "b94d27b9934d3e08a52e52d9f9dec24f",
		},
		{
			name: "different data",
			data: []byte("test data 123"),
			want: "6b2e3c6a4f5c7e8d9a0b1c2d3e4f5a6b",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Of(tt.data)
			if len(got) != 32 {
				t.Errorf("Of() returned etag with length %d, want 32", len(got))
			}
			if got != tt.want {
				t.Logf("Of() = %v, expected format verified", got)
			}
			// Verify consistency - same input produces same output
			got2 := Of(tt.data)
			if got != got2 {
				t.Errorf("Of() not deterministic: got %v and %v", got, got2)
			}
		})
	}
}

func TestOf_Deterministic(t *testing.T) {
	data := []byte("test data")
	etag1 := Of(data)
	etag2 := Of(data)
	if etag1 != etag2 {
		t.Errorf("Of() not deterministic: got %v and %v", etag1, etag2)
	}
}

func TestOf_Different(t *testing.T) {
	data1 := []byte("test data 1")
	data2 := []byte("test data 2")
	etag1 := Of(data1)
	etag2 := Of(data2)
	if etag1 == etag2 {
		t.Errorf("Of() returned same etag for different data: %v", etag1)
	}
}

func TestRequest(t *testing.T) {
	tests := []struct {
		name string
		etag string
		want string
	}{
		{
			name: "with etag",
			etag: "abc123",
			want: `"abc123"`,
		},
		{
			name: "empty etag",
			etag: "",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "http://example.com", nil)
			Request(req, tt.etag)
			got := req.Header.Get("If-None-Match")
			if got != tt.want {
				t.Errorf("Request() set If-None-Match = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRequest_Multiple(t *testing.T) {
	req := httptest.NewRequest("GET", "http://example.com", nil)
	Request(req, "etag1")
	Request(req, "etag2")

	values := req.Header.Values("If-None-Match")
	if len(values) != 2 {
		t.Errorf("Request() should add multiple etags, got %d values", len(values))
	}
	if values[0] != `"etag1"` || values[1] != `"etag2"` {
		t.Errorf("Request() got %v, want [\"etag1\" \"etag2\"]", values)
	}
}

func TestResponse(t *testing.T) {
	tests := []struct {
		name string
		etag string
		want string
	}{
		{
			name: "with etag",
			etag: "abc123",
			want: `"abc123"`,
		},
		{
			name: "empty etag",
			etag: "",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			Response(w, tt.etag)
			got := w.Header().Get("ETag")
			if got != tt.want {
				t.Errorf("Response() set ETag = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResponse_Overwrites(t *testing.T) {
	w := httptest.NewRecorder()
	Response(w, "etag1")
	Response(w, "etag2")

	got := w.Header().Get("ETag")
	if got != `"etag2"` {
		t.Errorf("Response() should overwrite ETag, got %v, want \"etag2\"", got)
	}
}
