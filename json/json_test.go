package json

import (
	"fmt"
	"io"
	"reflect"
	"testing"
)

func TestReader(t *testing.T) {
	r := Reader(map[string]int{
		"foo": 2,
	})
	bts, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	if string(bts) != `{"foo":2}` {
		t.Fatalf("wrong json: %s", string(bts))
	}
}

func TestFrom(t *testing.T) {
	in := map[string]int{"foo": 10, "bar": 20}
	m, err := From(Reader(in), map[string]int{})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if !reflect.DeepEqual(m, in) {
		t.Fatalf("maps should be equal: %v vs %v", in, m)
	}
}

func TestErrReader(t *testing.T) {
	err := fmt.Errorf("foo")
	_, err2 := io.ReadAll(&ErrorReader{err})
	if err != err2 {
		t.Fatalf("expected same error")
	}
}

func TestIsValid(t *testing.T) {
	tests := []struct {
		name string
		data any
		want bool
	}{
		{
			name: "empty string",
			data: "",
			want: false,
		},
		{
			name: "empty bytes",
			data: []byte(""),
			want: false,
		},
		{
			name: "valid json string",
			data: `{"foo": 2}`,
			want: true,
		},
		{
			name: "valid json bytes",
			data: []byte(`{"foo": 2}`),
			want: true,
		},
		{
			name: "invalid json string",
			data: `{"foo": 2`,
			want: false,
		},
		{
			name: "invalid json bytes",
			data: []byte(`{"foo": 2`),
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got bool
			switch v := tt.data.(type) {
			case string:
				got = IsValid(v)
			case []byte:
				got = IsValid(v)
			default:
				t.Fatalf("unsupported type: %T", tt.data)
			}
			if got != tt.want {
				t.Errorf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}
