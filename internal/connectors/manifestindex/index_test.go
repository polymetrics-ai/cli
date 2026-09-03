package manifestindex

import "testing"

func TestIndexIsSortedUniqueAndBounded(t *testing.T) {
	index, err := New([]Entry{{Connector: "z", Generation: "g", Digest: "d", Executor: "e"}, {Connector: "a", Generation: "g", Digest: "d", Executor: "e"}}, 2)
	if err != nil {
		t.Fatal(err)
	}
	if got := index.List(); got[0].Connector != "a" {
		t.Fatalf("list = %#v", got)
	}
	if got, ok := index.Lookup("z"); !ok || got.Executor != "e" {
		t.Fatalf("lookup = %#v/%t", got, ok)
	}
	if _, err := New([]Entry{{Connector: "a", Generation: "g", Digest: "d", Executor: "e"}, {Connector: "a", Generation: "g", Digest: "d", Executor: "e"}}, 2); err == nil {
		t.Fatal("duplicate accepted")
	}
	if _, err := New([]Entry{{Connector: "a", Generation: "g", Digest: "d", Executor: "e"}}, 0); err == nil {
		t.Fatal("limit accepted")
	}
}
