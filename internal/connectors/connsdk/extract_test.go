package connsdk

import "testing"

func TestRecordsAtRootArray(t *testing.T) {
	recs, err := RecordsAt([]byte(`[{"id":1},{"id":2}]`), "")
	if err != nil {
		t.Fatalf("RecordsAt: %v", err)
	}
	if len(recs) != 2 {
		t.Fatalf("len = %d, want 2", len(recs))
	}
}

func TestRecordsAtNestedPath(t *testing.T) {
	body := []byte(`{"meta":{"x":1},"result":{"items":[{"a":1},{"a":2},{"a":3}]}}`)
	recs, err := RecordsAt(body, "result.items")
	if err != nil {
		t.Fatalf("RecordsAt: %v", err)
	}
	if len(recs) != 3 {
		t.Fatalf("len = %d, want 3", len(recs))
	}
}

func TestRecordsAtSingleObject(t *testing.T) {
	recs, err := RecordsAt([]byte(`{"id":"abc","name":"x"}`), "")
	if err != nil {
		t.Fatalf("RecordsAt: %v", err)
	}
	if len(recs) != 1 || recs[0]["id"] != "abc" {
		t.Fatalf("recs = %+v", recs)
	}
}

func TestRecordsAtMissingPath(t *testing.T) {
	recs, err := RecordsAt([]byte(`{"data":[]}`), "missing.here")
	if err != nil {
		t.Fatalf("RecordsAt: %v", err)
	}
	if len(recs) != 0 {
		t.Fatalf("len = %d, want 0", len(recs))
	}
}

func TestStringAt(t *testing.T) {
	body := []byte(`{"meta":{"next_cursor":"c2","count":42,"done":false}}`)
	cases := map[string]string{
		"meta.next_cursor": "c2",
		"meta.count":       "42",
		"meta.done":        "false",
		"meta.absent":      "",
	}
	for path, want := range cases {
		got, err := StringAt(body, path)
		if err != nil {
			t.Fatalf("StringAt(%q): %v", path, err)
		}
		if got != want {
			t.Errorf("StringAt(%q) = %q, want %q", path, got, want)
		}
	}
}
