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
