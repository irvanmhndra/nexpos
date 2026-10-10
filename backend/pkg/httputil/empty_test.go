package httputil

import (
	"encoding/json"
	"testing"
)

type inner struct{ Options []string }
type outer struct {
	Items  []inner
	Groups []*inner
	Tags   []string `json:"tags,omitempty"`
	Child  *inner
}

func TestEmptyLists(t *testing.T) {
	var nilSlice []outer
	b, _ := json.Marshal(emptyLists(nilSlice))
	if string(b) != "[]" {
		t.Fatalf("top-level nil slice = %s", b)
	}
	v := []*outer{{Items: []inner{{}}, Groups: []*inner{{}}, Child: &inner{}}, nil}
	b, _ = json.Marshal(emptyLists(v))
	want := `[{"Items":[{"Options":[]}],"Groups":[{"Options":[]}],"Child":{"Options":[]}},null]`
	if string(b) != want {
		t.Fatalf("got %s\nwant %s", b, want)
	}
	// A struct passed by value is copied and filled.
	b, _ = json.Marshal(emptyLists(outer{}))
	if string(b) != `{"Items":[],"Groups":[],"Child":null}` {
		t.Fatalf("value struct: %s", b)
	}
}
