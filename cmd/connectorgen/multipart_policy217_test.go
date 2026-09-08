package main

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"testing"
)

func TestMultipartIdentityCanonical217(t *testing.T) {
	root, err := repoRoot()
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, "internal/connectors/defs/asana/source.lock.json")
	var rendered [2]map[string][]byte
	for i := range rendered {
		lock := readVNextSourceLockForTest(t, path)
		found := false
		for j := range lock.Operations {
			op := &lock.Operations[j]
			if op.ID != "write:upload_attachment_file" {
				continue
			}
			found = true
			var action map[string]any
			if err = json.Unmarshal(op.Write, &action); err != nil {
				t.Fatal(err)
			}
			parts := action["multipart"].(map[string]any)["parts"].([]any)
			file := parts[1].(map[string]any)
			delete(file, "filename_encoding")
			if i == 1 {
				file["filename_encoding"] = "identity"
			}
			op.Write, err = json.Marshal(action)
			if err != nil {
				t.Fatal(err)
			}
		}
		if !found {
			t.Fatal("missing source action")
		}
		canonical, err := canonicalizeVNextSourceLock(lock)
		if err != nil {
			t.Fatal(err)
		}
		rendered[i], err = renderVNextExecutionBundle(canonical)
		if err != nil {
			t.Fatal(err)
		}
	}
	if !bytes.Equal(rendered[0]["writes.json"], rendered[1]["writes.json"]) {
		t.Fatal("explicit default changed generated execution bytes")
	}
}

func TestMultipartIdentityOperationScope217(t *testing.T) {
	raw := json.RawMessage(`{"id":"attachment","kind":"rest_write","rest":{"method":"POST","multipart":{"max_bytes":9007199254740993,"parts":[{"name":"file","type":"file","filename_encoding":"identity"},{"name":"field","type":"field","filename_encoding":"identity"}]},"body_schema":{"const":"identity"}}}`)
	got := canonicalMultipartFilenamePolicy(raw, true)
	want := []byte(`{"id":"attachment","kind":"rest_write","rest":{"method":"POST","multipart":{"max_bytes":9007199254740993,"parts":[{"name":"file","type":"file"},{"name":"field","type":"field","filename_encoding":"identity"}]},"body_schema":{"const":"identity"}}}`)
	var compactGot, compactWant bytes.Buffer
	if err := json.Compact(&compactGot, got); err != nil {
		t.Fatal(err)
	}
	if err := json.Compact(&compactWant, want); err != nil {
		t.Fatal(err)
	}
	var root map[string]json.RawMessage
	if err := json.Unmarshal(got, &root); err != nil {
		t.Fatal(err)
	}
	if len(root) != 3 {
		t.Fatalf("operation root changed shape: %s", got)
	}
	if !sameJSON(compactGot.Bytes(), compactWant.Bytes()) {
		t.Fatalf("unrelated declaration changed: %s", got)
	}
	if !bytes.Contains(got, []byte("9007199254740993")) {
		t.Fatal("exact bound rounded")
	}
}
