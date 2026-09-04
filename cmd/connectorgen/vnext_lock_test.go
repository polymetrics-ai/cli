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

// TestEverySourceLockBuildsCanonicalGraph proves that every committed schema-4
// authoring input is accepted by the strict, no-I/O canonical graph builder.
func TestEverySourceLockBuildsCanonicalGraph(t *testing.T) {
	root, err := repoRoot()
	if err != nil {
		t.Fatalf("repoRoot(): %v", err)
	}
	defsRoot := filepath.Join(root, "internal", "connectors", "defs")
	entries, err := os.ReadDir(defsRoot)
	if err != nil {
		t.Fatalf("ReadDir(%s): %v", defsRoot, err)
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		connector := entry.Name()
		lockPath := filepath.Join(defsRoot, connector, "source.lock.json")
		if _, err := os.Stat(lockPath); os.IsNotExist(err) {
			continue
		} else if err != nil {
			t.Fatalf("Stat(%s): %v", lockPath, err)
		}
		t.Run(connector, func(t *testing.T) {
			lock := readVNextSourceLockForTest(t, lockPath)
			canonical, err := canonicalizeVNextSourceLock(lock)
			if err != nil {
				t.Fatalf("canonicalizeVNextSourceLock(%s): %v", connector, err)
			}
			if canonical.Graph.Identity.Digest == "" {
				t.Fatal("canonical graph did not retain its loaded execution identity")
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

func TestFoundationAtlasSelectorsResolve(t *testing.T) {
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
			ID    string `json:"id"`
			Owner struct {
				Files   []string `json:"files"`
				Symbols []struct {
					File string `json:"file"`
					Name string `json:"name"`
				} `json:"symbols"`
			} `json:"owner"`
			SupportedContracts struct {
				Guarantees []string `json:"guarantees"`
			} `json:"supported_contracts"`
			Selection struct {
				DefinitionFiles []string `json:"definition_files"`
				Selectors       []string `json:"selectors"`
			} `json:"selection"`
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
		t.Run(foundation.ID, func(t *testing.T) {
			if len(foundation.Owner.Files) == 0 || len(foundation.Owner.Symbols) == 0 {
				t.Fatal("Foundation Atlas owner must declare files and symbols")
			}
			if len(foundation.SupportedContracts.Guarantees) == 0 {
				t.Fatal("Foundation Atlas supported contract must declare guarantees")
			}
			if len(foundation.Selection.Selectors) == 0 {
				t.Fatal("Foundation Atlas selection must declare selectors")
			}
			if len(foundation.ProofTests) == 0 {
				t.Fatal("Foundation Atlas foundation declares zero proof tests")
			}
			ownerFiles := make(map[string]struct{}, len(foundation.Owner.Files))
			for _, ownerFile := range foundation.Owner.Files {
				if strings.TrimSpace(ownerFile) == "" {
					t.Error("Foundation Atlas owner has a blank file")
					continue
				}
				info, err := os.Stat(filepath.Join(root, ownerFile))
				if err != nil {
					t.Errorf("Foundation Atlas owner file %s does not resolve: %v", ownerFile, err)
					continue
				}
				if !info.Mode().IsRegular() {
					t.Errorf("Foundation Atlas owner file %s is not regular", ownerFile)
				}
				ownerFiles[ownerFile] = struct{}{}
			}
			for _, owner := range foundation.Owner.Symbols {
				if _, declared := ownerFiles[owner.File]; !declared {
					t.Errorf("Foundation Atlas owner symbol %s names undeclared owner file %s", owner.Name, owner.File)
				}
				filePath := filepath.Join(root, owner.File)
				fileSet := token.NewFileSet()
				parsed, err := parser.ParseFile(fileSet, filePath, nil, 0)
				if err != nil {
					t.Fatalf("parse owner file %s: %v", owner.File, err)
				}
				if !declaresSymbol(parsed, owner.Name) {
					t.Errorf("Foundation Atlas owner symbol %s in %s does not resolve", owner.Name, owner.File)
				}
			}
			for _, guarantee := range foundation.SupportedContracts.Guarantees {
				if strings.TrimSpace(guarantee) == "" {
					t.Error("Foundation Atlas guarantee is blank")
				}
			}
			for _, selector := range foundation.Selection.Selectors {
				if strings.TrimSpace(selector) == "" {
					t.Error("Foundation Atlas selector is blank")
				}
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
		})
	}
}

func TestDecodeVNextSourceLockRejectsTrailingAndDuplicateJSON(t *testing.T) {
	raw, err := json.Marshal(minimalVNextLockForTest())
	if err != nil {
		t.Fatalf("marshal minimal lock: %v", err)
	}
	tests := []struct {
		name string
		raw  []byte
	}{
		{name: "trailing document", raw: append(append([]byte(nil), raw...), []byte("\n{}")...)},
		{name: "duplicate root member", raw: bytes.Replace(raw, []byte(`"connector":"acme"`), []byte(`"connector":"acme","connector":"other"`), 1)},
		{name: "duplicate nested member", raw: bytes.Replace(raw, []byte(`"metadata":{"name":"acme","display_name":"Acme"`), []byte(`"metadata":{"name":"acme","name":"other","display_name":"Acme"`), 1)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := decodeVNextSourceLock(test.raw); err == nil {
				t.Fatal("decodeVNextSourceLock accepted non-canonical JSON")
			}
		})
	}
}

func TestRunLockRenderRejectsNonCanonicalSourceBeforeWriting(t *testing.T) {
	root := t.TempDir()
	connectorRoot := filepath.Join(root, "acme")
	if err := os.MkdirAll(connectorRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(minimalVNextLockForTest())
	if err != nil {
		t.Fatalf("marshal minimal lock: %v", err)
	}
	if err := os.WriteFile(filepath.Join(connectorRoot, "source.lock.json"), append(raw, []byte("\n{}")...), 0o600); err != nil {
		t.Fatal(err)
	}
	const sentinel = "must not be replaced"
	outputPath := filepath.Join(connectorRoot, "metadata.json")
	if err := os.WriteFile(outputPath, []byte(sentinel), 0o600); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	if code := runLockRender([]string{"lock-render", "acme", "--defs", root}, &stdout, &stderr); code != 1 {
		t.Fatalf("runLockRender() = %d, want parse failure; stderr=%s", code, stderr.String())
	}
	got, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != sentinel {
		t.Fatalf("lock render overwrote output before rejecting non-canonical source: %q", got)
	}
}

func TestRunLockRenderRejectsInvalidCanonicalGraphBeforeWriting(t *testing.T) {
	tests := []struct {
		name     string
		wantPath string
		mutate   func(*vNextSourceLock)
	}{
		{
			name:     "unknown stream execution field",
			wantPath: "/operations/0/stream",
			mutate: func(lock *vNextSourceLock) {
				lock.Operations[0].Stream = replaceVNextRaw(t, lock.Operations[0].Stream, `"schema":"schemas/widgets.json"`, `"schema":"schemas/widgets.json","unknown_execution":true`)
			},
		},
		{
			name:     "wrong request schema role",
			wantPath: "/operations/0/schema_refs/request",
			mutate: func(lock *vNextSourceLock) {
				lock.Operations[0].SchemaRefs.Request = "schemas/widgets.json"
			},
		},
		{
			name:     "unsupported stream body encoder",
			wantPath: "/operations/0/stream",
			mutate: func(lock *vNextSourceLock) {
				lock.Operations[0].Stream = replaceVNextRaw(t, lock.Operations[0].Stream, `"path":"/widgets"`, `"path":"/widgets","body_type":"opaque"`)
			},
		},
		{
			name:     "normalized command alias collision",
			wantPath: "/operations/0/commands/1/path",
			mutate: func(lock *vNextSourceLock) {
				lock.Lanes["direct_read"] = "implemented"
				lock.Operations[0].Operation = json.RawMessage(`{"id":"widgets.list","kind":"rest_read","summary":"List widgets","risk":"low","approval":"none","output_policy":"json_redacted","rest":{"method":"GET","path":"/widgets","max_bytes":1024,"response":{"success_statuses":["200"]},"parameters":[]}}`)
				lock.Operations[0].Commands = []vNextCommandDescriptor{
					{Order: 0, Command: json.RawMessage(`{"path":"widgets list","summary":"List widgets","intent":"direct_read","availability":"implemented","operation":"widgets.list","flags":[]}`)},
					{Order: 1, Command: json.RawMessage(`{"path":"widgets  list","summary":"Alias","intent":"direct_read","availability":"implemented","operation":"widgets.list","flags":[]}`)},
				}
				lock.CLI = json.RawMessage(`{"usage":"pm acme <command>","tagline":"Acme"}`)
			},
		},
		{
			name:     "missing record schema binding",
			wantPath: "/operations/0/schema_refs/record",
			mutate: func(lock *vNextSourceLock) {
				lock.Schemas["schemas/other.json"] = json.RawMessage(`{"type":"object","properties":{"id":{"type":"string"}},"x-primary-key":["id"]}`)
				lock.Operations[0].Stream = replaceVNextRaw(t, lock.Operations[0].Stream, `"schema":"schemas/widgets.json"`, `"schema":"schemas/other.json"`)
			},
		},
		{
			name:     "non-object source identity",
			wantPath: "/operations/0/source",
			mutate: func(lock *vNextSourceLock) {
				lock.Operations[0].Source = json.RawMessage(`["source-widgets"]`)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			connectorRoot := filepath.Join(root, "acme")
			if err := os.MkdirAll(connectorRoot, 0o755); err != nil {
				t.Fatal(err)
			}
			lock := minimalVNextLockForTest()
			test.mutate(&lock)
			raw, err := json.Marshal(lock)
			if err != nil {
				t.Fatalf("marshal source lock: %v", err)
			}
			if err := os.WriteFile(filepath.Join(connectorRoot, "source.lock.json"), raw, 0o600); err != nil {
				t.Fatal(err)
			}
			outputPath := filepath.Join(connectorRoot, "metadata.json")
			const sentinel = "must not be replaced"
			if err := os.WriteFile(outputPath, []byte(sentinel), 0o600); err != nil {
				t.Fatal(err)
			}
			var stdout, stderr bytes.Buffer
			if code := runLockRender([]string{"lock-render", "acme", "--defs", root}, &stdout, &stderr); code != 1 {
				t.Fatalf("runLockRender() = %d, want canonical-graph refusal; stdout=%s stderr=%s", code, stdout.String(), stderr.String())
			}
			if !strings.Contains(stderr.String(), test.wantPath) {
				t.Fatalf("refusal %q lacks source path %q: %s", test.name, test.wantPath, stderr.String())
			}
			got, err := os.ReadFile(outputPath)
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != sentinel {
				t.Fatalf("lock render replaced output after rejecting %s: %q", test.name, got)
			}
		})
	}
}

func TestVNextCanonicalGraphAllowsDirectOperationSchemaRoles(t *testing.T) {
	lock := minimalVNextLockForTest()
	lock.Lanes["etl"] = "unsupported"
	lock.Lanes["direct_read"] = "implemented"
	lock.Schemas["schemas/widgets-request.json"] = json.RawMessage(`{"type":"object","properties":{"limit":{"type":"integer"}}}`)
	lock.Schemas["schemas/widgets-response.json"] = json.RawMessage(`{"type":"object","properties":{"data":{"type":"array"}}}`)
	lock.Operations = []vNextOperationDescriptor{{
		ID:         "operation:widgets.get",
		SchemaRefs: vNextSchemaReferences{Request: "schemas/widgets-request.json", Response: "schemas/widgets-response.json"},
		Operation:  json.RawMessage(`{"id":"widgets.get","kind":"rest_read","summary":"Get widgets","risk":"low","approval":"none","output_policy":"json_redacted","rest":{"method":"GET","path":"/widgets","max_bytes":1024,"response":{"success_statuses":["200"]},"parameters":[]}}`),
		Commands: []vNextCommandDescriptor{{
			Order:   0,
			Command: json.RawMessage(`{"path":"widgets get","summary":"Get widgets","intent":"direct_read","availability":"implemented","operation":"widgets.get","flags":[]}`),
		}},
	}}
	lock.CLI = json.RawMessage(`{"usage":"pm acme <command>","tagline":"Acme"}`)

	if _, err := canonicalizeVNextSourceLock(lock); err != nil {
		t.Fatalf("canonicalizeVNextSourceLock() rejected direct operation schema roles: %v", err)
	}
}

func TestVNextCanonicalGraphIgnoresIrrelevantOperationOrdering(t *testing.T) {
	first := minimalVNextLockForTest()
	first.Operations[0].Source = json.RawMessage(`{"provider_operation":"widgets.list","reference":"https://provider.example/widgets"}`)
	first.Operations[0].StreamOrder = 0
	first.Schemas["schemas/gadgets.json"] = json.RawMessage(`{"type":"object","properties":{"id":{"type":"string"}},"x-primary-key":["id"]}`)
	first.Operations = append(first.Operations, vNextOperationDescriptor{
		ID:          "stream:gadgets",
		Source:      json.RawMessage(`{"provider_operation":"gadgets.list","reference":"https://provider.example/gadgets"}`),
		StreamOrder: 0,
		SchemaRefs:  vNextSchemaReferences{Record: "schemas/gadgets.json"},
		Stream:      json.RawMessage(`{"name":"gadgets","path":"/gadgets","records":{"path":"data"},"schema":"schemas/gadgets.json"}`),
	})
	second := first
	second.Operations = []vNextOperationDescriptor{first.Operations[1], first.Operations[0]}

	firstCanonical, err := canonicalizeVNextSourceLock(first)
	if err != nil {
		t.Fatalf("canonicalize first source lock: %v", err)
	}
	secondCanonical, err := canonicalizeVNextSourceLock(second)
	if err != nil {
		t.Fatalf("canonicalize reordered source lock: %v", err)
	}
	firstOutput, err := renderVNextExecutionBundle(firstCanonical)
	if err != nil {
		t.Fatalf("render first source lock: %v", err)
	}
	secondOutput, err := renderVNextExecutionBundle(secondCanonical)
	if err != nil {
		t.Fatalf("render reordered source lock: %v", err)
	}
	if !executionBundlesEqual(firstOutput, secondOutput) {
		t.Fatalf("irrelevant source operation order changed rendered bytes:\nfirst=%s\nsecond=%s", firstOutput["streams.json"], secondOutput["streams.json"])
	}
	if firstCanonical.Graph.Identity.Digest == "" || firstCanonical.Graph.Identity.Digest != secondCanonical.Graph.Identity.Digest {
		t.Fatalf("irrelevant source operation order changed graph digest: %q != %q", firstCanonical.Graph.Identity.Digest, secondCanonical.Graph.Identity.Digest)
	}
	seen := map[string]json.RawMessage{}
	for _, operation := range firstCanonical.Graph.Operations {
		seen[operation.ID] = operation.Source
	}
	for _, id := range []string{"stream:widgets", "stream:gadgets"} {
		if _, exists := seen[id]; !exists {
			t.Fatalf("canonical graph lost source identity %q", id)
		}
	}
	if !sameJSON(seen["stream:widgets"], first.Operations[0].Source) {
		t.Fatalf("canonical graph changed widgets source identity: %s", seen["stream:widgets"])
	}
	if !sameJSON(seen["stream:gadgets"], first.Operations[1].Source) {
		t.Fatalf("canonical graph changed gadgets source identity: %s", seen["stream:gadgets"])
	}
}

func replaceVNextRaw(t *testing.T, raw json.RawMessage, old, replacement string) json.RawMessage {
	t.Helper()
	updated := strings.Replace(string(raw), old, replacement, 1)
	if updated == string(raw) {
		t.Fatalf("source fixture did not contain %q", old)
	}
	return json.RawMessage(updated)
}

func readVNextSourceLockForTest(t *testing.T, lockPath string) vNextSourceLock {
	t.Helper()
	raw, err := os.ReadFile(lockPath)
	if err != nil {
		t.Fatalf("ReadFile(%s): %v", lockPath, err)
	}
	lock, err := decodeVNextSourceLock(raw)
	if err != nil {
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

func declaresSymbol(file *ast.File, name string) bool {
	receiverName, declaredName := splitAtlasSymbol(name)
	for _, declaration := range file.Decls {
		switch declaration := declaration.(type) {
		case *ast.FuncDecl:
			if declaration.Name.Name != declaredName {
				continue
			}
			if receiverName == "" || receiverTypeName(declaration.Recv) == receiverName {
				return true
			}
		case *ast.GenDecl:
			if receiverName != "" {
				continue
			}
			for _, spec := range declaration.Specs {
				switch spec := spec.(type) {
				case *ast.TypeSpec:
					if spec.Name.Name == declaredName {
						return true
					}
				case *ast.ValueSpec:
					for _, declared := range spec.Names {
						if declared.Name == declaredName {
							return true
						}
					}
				}
			}
		}
	}
	return false
}

func splitAtlasSymbol(name string) (receiverName, declaredName string) {
	receiverEnd := strings.LastIndex(name, ".")
	if receiverEnd < 0 {
		return "", name
	}
	receiverName = strings.Trim(name[:receiverEnd], "()*")
	return receiverName, name[receiverEnd+1:]
}

func receiverTypeName(receivers *ast.FieldList) string {
	if receivers == nil || len(receivers.List) != 1 {
		return ""
	}
	switch typ := receivers.List[0].Type.(type) {
	case *ast.Ident:
		return typ.Name
	case *ast.StarExpr:
		if identifier, ok := typ.X.(*ast.Ident); ok {
			return identifier.Name
		}
	}
	return ""
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
	firstCLI := cloneRawJSON(lock.CLI)
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
	if first.Graph.Identity.Digest == "" || first.Graph.Identity.Digest != second.Graph.Identity.Digest {
		t.Fatalf("provider evidence changed graph execution digest: %q != %q", first.Graph.Identity.Digest, second.Graph.Identity.Digest)
	}
	if len(first.Graph.ProviderEvidence) != 0 {
		t.Fatalf("canonical graph retained absent provider evidence: %s", first.Graph.ProviderEvidence)
	}
	if !sameJSON(second.Graph.ProviderEvidence, lock.ProviderEvidence) {
		t.Fatalf("canonical graph did not retain provider evidence: %s", second.Graph.ProviderEvidence)
	}
	if !sameJSON(first.Graph.AuthoredCLI, firstCLI) {
		t.Fatalf("canonical graph did not retain source CLI evidence: %s", first.Graph.AuthoredCLI)
	}
	if !sameJSON(second.Graph.AuthoredCLI, lock.CLI) {
		t.Fatalf("canonical graph did not retain changed source CLI evidence: %s", second.Graph.AuthoredCLI)
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
		Metadata:     json.RawMessage(`{"name":"acme","display_name":"Acme","description":"Test connector","integration_type":"api","release_stage":"ga","capabilities":{"check":true,"read":true,"write":false,"query":false,"cdc":false,"dynamic_schema":false}}`),
		ConfigSchema: json.RawMessage(`{"type":"object","properties":{}}`),
		HTTP:         json.RawMessage(`{"url":"https://api.acme.example","headers":{},"auth":[],"pagination":{"type":"none"},"check":{"method":"GET","path":"/check"},"error_map":[]}`),
		Schemas: map[string]json.RawMessage{
			"schemas/widgets.json": json.RawMessage(`{"type":"object","properties":{"id":{"type":"string"}},"x-primary-key":["id"]}`),
		},
		Operations: []vNextOperationDescriptor{{
			ID: "stream:widgets", Stream: json.RawMessage(`{"name":"widgets","path":"/widgets","records":{"path":"data"},"schema":"schemas/widgets.json"}`),
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
