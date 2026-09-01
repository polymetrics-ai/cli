package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var vNextReferenceLockConnectors = []string{"asana", "github", "gitlab"}

func TestVNextSourceLockProjectsMinimalExecutionBundleDeterministically(t *testing.T) {
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

// TestVNextSourceLockDeterministicallyRendersReferenceConnectors is the
// Foundation Atlas proof for schema-4 authoring. It reads each committed lock,
// renders the complete execution set in memory twice, and compares the closed
// set byte-for-byte with the committed execution artifacts without publishing
// anything.
func TestVNextSourceLockDeterministicallyRendersReferenceConnectors(t *testing.T) {
	root, err := repoRoot()
	if err != nil {
		t.Fatalf("repoRoot(): %v", err)
	}
	defsRoot := filepath.Join(root, "internal", "connectors", "defs")
	for _, connector := range vNextReferenceLockConnectors {
		t.Run(connector, func(t *testing.T) {
			lock := readVNextSourceLockForTest(t, filepath.Join(defsRoot, connector, "source.lock.json"))
			canonical, err := canonicalizeVNextSourceLock(lock)
			if err != nil {
				t.Fatalf("canonicalizeVNextSourceLock(%s): %v", connector, err)
			}
			first, err := renderVNextExecutionBundle(canonical)
			if err != nil {
				t.Fatalf("renderVNextExecutionBundle(%s): %v", connector, err)
			}
			second, err := renderVNextExecutionBundle(canonical)
			if err != nil {
				t.Fatalf("second renderVNextExecutionBundle(%s): %v", connector, err)
			}
			if !executionBundlesEqual(first, second) {
				t.Fatal("reference source lock renderer is not byte-deterministic")
			}
			committed := readVNextExecutionOutputsForTest(t, defsRoot, connector)
			if err := compareVNextExecutionOutputSets(first, committed); err != nil {
				t.Fatalf("%s committed execution set differs from its in-memory render: %v", connector, err)
			}
		})
	}
}

// This negative assertion keeps the closed-set comparison honest: a rendered
// bundle cannot silently admit a known execution artifact that the lock did not
// render. The production publication defect itself remains separately frozen
// for #4427; this test makes the #4423 in-memory proof reject that shape.
func TestVNextReferenceLockClosedSetRejectsUnrenderedArtifact(t *testing.T) {
	want := map[string][]byte{"metadata.json": []byte(`{"name":"acme"}`)}
	got := map[string][]byte{
		"metadata.json":    []byte(`{"name":"acme"}`),
		"rate_limits.json": []byte(`{"requests_per_minute":1}`),
	}
	if err := compareVNextExecutionOutputSets(want, got); err == nil {
		t.Fatal("closed-set comparison accepted an unrendered execution artifact")
	}
}

func TestVNextFoundationProofSelectorsResolve(t *testing.T) {
	root, err := repoRoot()
	if err != nil {
		t.Fatalf("repoRoot(): %v", err)
	}
	catalogPath := filepath.Join(root, "docs", "connector-canon", "foundations", "catalog.json")
	raw, err := os.ReadFile(catalogPath)
	if err != nil {
		t.Fatalf("ReadFile(%s): %v", catalogPath, err)
	}
	var catalog struct {
		Foundations []struct {
			ID         string `json:"id"`
			ProofTests []struct {
				File string `json:"file"`
				Name string `json:"name"`
			} `json:"proof_tests"`
		} `json:"foundations"`
	}
	if err := json.Unmarshal(raw, &catalog); err != nil {
		t.Fatalf("Unmarshal(%s): %v", catalogPath, err)
	}
	for _, foundation := range catalog.Foundations {
		if foundation.ID != "authoring.source-lock-vnext.v1" {
			continue
		}
		if len(foundation.ProofTests) == 0 {
			t.Fatal("source-lock-vnext foundation declares zero proof tests")
		}
		for _, proof := range foundation.ProofTests {
			filePath := filepath.Join(root, proof.File)
			fileSet := token.NewFileSet()
			parsed, err := parser.ParseFile(fileSet, filePath, nil, 0)
			if err != nil {
				t.Fatalf("parse proof file %s: %v", proof.File, err)
			}
			if !declaresTestFunction(parsed, proof.Name) {
				t.Errorf("Foundation Atlas proof %s in %s does not declare an executable test function", proof.Name, proof.File)
			}
		}
		return
	}
	t.Fatal("source-lock-vnext foundation is absent from the Foundation Atlas")
}

func readVNextSourceLockForTest(t *testing.T, lockPath string) vNextSourceLock {
	t.Helper()
	raw, err := os.ReadFile(lockPath)
	if err != nil {
		t.Fatalf("ReadFile(%s): %v", lockPath, err)
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	var lock vNextSourceLock
	if err := decoder.Decode(&lock); err != nil {
		t.Fatalf("decode %s: %v", lockPath, err)
	}
	return lock
}

func readVNextExecutionOutputsForTest(t *testing.T, defsRoot, connector string) map[string][]byte {
	t.Helper()
	source := os.DirFS(defsRoot)
	outputs := make(map[string][]byte)
	err := fs.WalkDir(source, connector, func(filePath string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		relative := strings.TrimPrefix(filePath, connector+"/")
		if !isVNextRenderedExecutionPath(relative) {
			return nil
		}
		data, err := fs.ReadFile(source, filePath)
		if err != nil {
			return err
		}
		outputs[relative] = data
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s execution artifacts: %v", connector, err)
	}
	return outputs
}

func isVNextRenderedExecutionPath(name string) bool {
	if strings.HasPrefix(name, "schemas/") && strings.HasSuffix(name, ".json") {
		return true
	}
	switch name {
	case "metadata.json", "spec.json", "streams.json", "writes.json", "operations.json", "cli_surface.json":
		return true
	default:
		return isVNextOptionalExecutionFile(name)
	}
}

func compareVNextExecutionOutputSets(want, got map[string][]byte) error {
	for _, name := range sortedOutputNames(want) {
		current, ok := got[name]
		if !ok {
			return fmt.Errorf("missing %s", name)
		}
		if !bytes.Equal(want[name], current) {
			return fmt.Errorf("%s differs", name)
		}
	}
	for _, name := range sortedOutputNames(got) {
		if _, ok := want[name]; !ok {
			return fmt.Errorf("unexpected %s", name)
		}
	}
	return nil
}

func declaresTestFunction(file *ast.File, name string) bool {
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if ok && function.Recv == nil && function.Name.Name == name && hasTestingTParameter(function) {
			return true
		}
	}
	return false
}

func hasTestingTParameter(function *ast.FuncDecl) bool {
	if function.Type.Params == nil || len(function.Type.Params.List) != 1 {
		return false
	}
	pointer, ok := function.Type.Params.List[0].Type.(*ast.StarExpr)
	if !ok {
		return false
	}
	selector, ok := pointer.X.(*ast.SelectorExpr)
	if !ok || selector.Sel.Name != "T" {
		return false
	}
	packageName, ok := selector.X.(*ast.Ident)
	return ok && packageName.Name == "testing"
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
