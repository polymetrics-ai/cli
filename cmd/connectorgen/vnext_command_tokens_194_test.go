package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestVNextFlagLikeCommandPublicationRejected194(t *testing.T) {
	cases := []struct{ id, path, alias, param, flag string }{
		{"source.literal", "/widgets/fixed", "widgets --help", "", ""},
	}
	lock := operationDirectReadLockForSemanticAdmissionTest()
	lock.Operations = nil
	for i, tc := range cases {
		parameters := []map[string]any{}
		flags := []map[string]any{}
		if tc.param != "" {
			parameters = append(parameters, map[string]any{"name": tc.param, "in": "path", "type": "string", "required": true})
			flags = append(flags, map[string]any{"name": tc.flag, "maps_to": "path." + tc.param, "type": "string", "required": true})
		}
		operation, err := json.Marshal(map[string]any{"id": tc.id, "kind": "rest_read", "summary": "Retained source fixture", "risk": "low", "approval": "none", "output_policy": "json_redacted", "rest": map[string]any{"method": "GET", "path": tc.path, "max_bytes": 1024, "response": map[string]any{"success_statuses": []string{"200"}}, "parameters": parameters}})
		if err != nil {
			t.Fatal(err)
		}
		command, err := json.Marshal(map[string]any{"path": tc.alias, "summary": "Retained command", "intent": "direct_read", "availability": "implemented", "operation": tc.id, "api_surface": []map[string]any{{"method": "GET", "path": tc.path}}, "output_policy": "json_redacted", "flags": flags})
		if err != nil {
			t.Fatal(err)
		}
		source, err := json.Marshal(map[string]any{"method": "GET", "path": tc.path})
		if err != nil {
			t.Fatal(err)
		}
		lock.Operations = append(lock.Operations, vNextOperationDescriptor{ID: tc.id, Source: source, Operation: operation, Commands: []vNextCommandDescriptor{{Order: i, Command: command}}})
	}
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "acme"), 0700); err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(lock)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "acme", "source.lock.json"), raw, 0600); err != nil {
		t.Fatal(err)
	}
	var out, diag bytes.Buffer
	if code := runLockRender([]string{"lock-render", "acme", "--defs", root}, &out, &diag); code == 0 {
		t.Fatalf("flag-like command published: %s", out.String())
	} else {
		t.Logf("rejected before publication: %s", diag.String())
	}
}
