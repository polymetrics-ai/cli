package main

import (
	"encoding/json"
	"testing"
)

func TestVNextSourceLockProjectsExecutionBundleDeterministically(t *testing.T) {
	lock := minimalVNextLockForTest()
	lock.Lanes["direct_read"] = "implemented"
	lock.Operations[0].Operation = json.RawMessage(`{"id":"widgets.list","kind":"rest_read","summary":"List widgets","risk":"low","approval":"none","output_policy":"json_redacted","rest":{"method":"GET","path":"/widgets","max_bytes":1024,"response":{"success_statuses":["200"]},"parameters":[]}}`)
	lock.Operations[0].Commands = []vNextCommandDescriptor{{Order: 0, Command: json.RawMessage(`{"path":"widgets list","summary":"List widgets","intent":"direct_read","availability":"implemented","operation":"widgets.list","flags":[]}`)}}
	lock.CLI = json.RawMessage(`{"usage":"pm acme <command> [flags]","tagline":"Acme commands"}`)
	canonical, err := canonicalizeVNextSourceLock(lock)
	if err != nil {
		t.Fatalf("canonicalizeVNextSourceLock() error = %v", err)
	}
	first, err := renderVNextExecutionBundle(canonical)
	if err != nil {
		t.Fatalf("renderVNextExecutionBundle() error = %v", err)
	}
	second, err := renderVNextExecutionBundle(canonical)
	if err != nil {
		t.Fatalf("second renderVNextExecutionBundle() error = %v", err)
	}
	if !executionBundlesEqual(first, second) {
		t.Fatal("vNext renderer is not byte-deterministic")
	}
	for _, file := range []string{"metadata.json", "spec.json", "streams.json", "schemas/widgets.json", "operations.json", "cli_surface.json"} {
		if len(first[file]) == 0 {
			t.Fatalf("rendered execution bundle is missing %s", file)
		}
	}
	if !sameJSON(first["operations.json"], []byte(`{"operations":[{"id":"widgets.list","kind":"rest_read","summary":"List widgets","risk":"low","approval":"none","output_policy":"json_redacted","rest":{"method":"GET","path":"/widgets","max_bytes":1024,"response":{"success_statuses":["200"]},"parameters":[]}}]}`)) {
		t.Fatalf("rendered operations.json = %s", first["operations.json"])
	}
}

func TestVNextSourceLockEvidenceCannotChangeExecution(t *testing.T) {
	lock := minimalVNextLockForTest()
	lock.CLI = json.RawMessage(`{"usage":"pm acme <command>","tagline":"Acme","source_cli":{"name":"acmectl","reference":"https://example.test/v1"}}`)
	first, err := canonicalizeVNextSourceLock(lock)
	if err != nil {
		t.Fatal(err)
	}
	lock.ProviderEvidence = json.RawMessage(`{"documents":{"rest":{"url":"https://changed.example/openapi.json"}}}`)
	lock.CLI = json.RawMessage(`{"usage":"pm acme <command>","tagline":"Acme","source_cli":{"name":"changedctl","reference":"https://example.test/v2"}}`)
	second, err := canonicalizeVNextSourceLock(lock)
	if err != nil {
		t.Fatal(err)
	}
	firstBundle, err := renderVNextExecutionBundle(first)
	if err != nil {
		t.Fatal(err)
	}
	secondBundle, err := renderVNextExecutionBundle(second)
	if err != nil {
		t.Fatal(err)
	}
	if !executionBundlesEqual(firstBundle, secondBundle) {
		t.Fatal("provider evidence changed projected execution JSON")
	}
}

func TestVNextSourceLockRejectsMissingSharedSchemaReference(t *testing.T) {
	lock := minimalVNextLockForTest()
	lock.Operations[0].SchemaRefs.Record = "schemas/missing.json"
	if _, err := canonicalizeVNextSourceLock(lock); err == nil {
		t.Fatal("canonicalizeVNextSourceLock() accepted a missing shared schema reference")
	}
}

func TestVNextSourceLockRejectsLegacyExecutionEvidence(t *testing.T) {
	tests := []struct {
		name string
		raw  json.RawMessage
	}{
		{name: "source operation", raw: json.RawMessage(`{"name":"widgets","source_operation":"widgets.list"}`)},
		{name: "source CLI path", raw: json.RawMessage(`{"name":"widgets","source_cli_path":"widgets list"}`)},
		{name: "conformance", raw: json.RawMessage(`{"name":"widgets","conformance":{"status":"certified"}}`)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			lock := minimalVNextLockForTest()
			lock.Operations[0].Stream = test.raw
			if _, err := canonicalizeVNextSourceLock(lock); err == nil {
				t.Fatal("canonicalizeVNextSourceLock() accepted legacy execution evidence")
			}
		})
	}
}

func minimalVNextLockForTest() vNextSourceLock {
	return vNextSourceLock{
		SchemaVersion: vNextSourceLockSchemaVersion,
		Connector:     "acme",
		Lanes: map[string]string{
			"direct_read": "unsupported", "direct_write": "unsupported", "binary_download": "unsupported", "binary_upload": "unsupported",
			"etl": "implemented", "reverse_etl": "unsupported", "sync_transport": "unsupported",
		},
		Metadata:     json.RawMessage(`{"name":"acme"}`),
		ConfigSchema: json.RawMessage(`{"type":"object"}`),
		HTTP:         json.RawMessage(`{"url":"https://api.acme.example"}`),
		Schemas:      map[string]json.RawMessage{"schemas/widgets.json": json.RawMessage(`{"type":"object"}`)},
		Operations: []vNextOperationDescriptor{{
			ID: "stream:widgets", Stream: json.RawMessage(`{"name":"widgets","path":"/widgets","schema":"schemas/widgets.json"}`),
			SchemaRefs: vNextSchemaReferences{Record: "schemas/widgets.json"},
		}},
	}
}

func sameJSON(left, right []byte) bool {
	var leftValue, rightValue any
	if json.Unmarshal(left, &leftValue) != nil || json.Unmarshal(right, &rightValue) != nil {
		return false
	}
	leftCanonical, _ := json.Marshal(leftValue)
	rightCanonical, _ := json.Marshal(rightValue)
	return string(leftCanonical) == string(rightCanonical)
}
