package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestSourceLaneBindingClaims(t *testing.T) {
	const artifact = "internal/connectors/defs/fixture/operations.json"
	raw := []byte(`{"operations":[{"id":"read_widget","kind":"rest_read","rest":{"method":"GET","path":"/widgets/{id}"}},{"id":"other_widget","kind":"rest_read","rest":{"method":"GET","path":"/other/{id}"}}]}`)
	node := json.RawMessage(`{"id":"provider.read","protocol":"rest","method":"GET","path":"/widgets/{id}","source_operation":{"summary":"Get widget","responses":{"204":{"description":"No content"}}}}`)
	key := sourceOperationKey{Connector: "fixture", Inventory: "primary", ID: "provider.read"}
	row := retainedSourceOperation{Key: key, Observed: true, Node: node, Pointer: "/rest/operations/0"}
	doc := retainedSourceDocument{ID: "fixture:primary", Payload: json.RawMessage(`{"rest":{"operations":[` + string(node) + `]}}`)}
	for _, tc := range []struct {
		name, id, pointer, hash string
		materialized            bool
		want                    string
	}{
		{"matching observed endpoint", "read_widget", "/operations/0", "", false, "target_contract_unverified"},
		{"existing target without authored pointer", "read_widget", "", "", false, "target_contract_unverified"},
		{"existing target with missing pointer", "other_widget", "/operations/9", "", false, "target_pointer_mismatch"},
		{"absent intended", "future_read", "/operations/9", "", false, "target_absent"},
		{"dangling present", "future_read", "/operations/9", sourceBytesHash(raw), true, "materialized_target_absent"},
		{"existing wrong endpoint", "other_widget", "/operations/1", sourceBytesHash(raw), true, "target_semantics_mismatch"},
		{"wrong identity at valid pointer", "other_widget", "/operations/0", sourceBytesHash(raw), true, "target_identity_mismatch"},
		{"changed target bytes", "read_widget", "/operations/0", "stale", true, "target_digest_mismatch"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			facts := normalizeSourceFacts(row, doc, nil)
			facts.bindings = &sourceLaneBindingInputs{Artifacts: map[string][]byte{artifact: raw}, Canonical: map[string]vNextCanonicalDescriptor{}}
			ref := sourceLaneTargetRef{Kind: "operation", Connector: "fixture", ID: tc.id, Lane: "direct_read", Artifact: artifact, Pointer: tc.pointer, ArtifactSHA256: tc.hash}
			a := sourceSemanticAnnotation{Key: key, Citation: facts.Refs["summary"], Clause: "Get widget"}
			if tc.materialized {
				a.MaterializedBindings = []sourceLaneTargetRef{ref}
			} else {
				a.IntendedBindings = []sourceLaneTargetRef{ref}
			}
			cells := classifySourceLanes(key, facts, &a)
			cell := requireSourceLane(t, cells, "direct_read", "applicable")
			found := false
			for _, d := range cell.Diagnostics {
				if d.Code == tc.want {
					found = true
				}
			}
			if !found {
				t.Fatalf("reference %s: want diagnostic %s, got %+v", tc.name, tc.want, cell.Diagnostics)
			}
			if len(cell.References) != 0 || cell.State != "mapped_unproven" {
				t.Fatalf("invalid/absent target promoted: %+v", cell)
			}
		})
	}
}

func TestSourceLaneBindingCanonicalPositive(t *testing.T) {
	lock := operationDirectReadLockForSemanticAdmissionTest()
	descriptor, err := canonicalizeVNextSourceLock(lock)
	if err != nil {
		t.Fatal(err)
	}
	artifact := "internal/connectors/defs/acme/operations.json"
	raw := descriptor.Staged.Outputs["operations.json"]
	node := json.RawMessage(`{"id":"provider.widgets.get","protocol":"rest","method":"GET","path":"/widgets","source_operation":{"summary":"Get widgets","responses":{"200":{"description":"Response semantics retained separately"}}}}`)
	key := sourceOperationKey{Connector: "acme", Inventory: "primary", ID: "provider.widgets.get"}
	row := retainedSourceOperation{Key: key, Observed: true, Node: node, Pointer: "/rest/operations/0"}
	doc := retainedSourceDocument{ID: "acme:primary", Payload: json.RawMessage(`{"rest":{"operations":[` + string(node) + `]}}`)}
	facts := normalizeSourceFacts(row, doc, nil)
	facts.bindings = &sourceLaneBindingInputs{Artifacts: map[string][]byte{artifact: raw}, Canonical: map[string]vNextCanonicalDescriptor{"acme": descriptor}}
	ref := sourceLaneTargetRef{Kind: "operation", Connector: "acme", ID: "widgets.get", Lane: "direct_read", Artifact: artifact, Pointer: "/operations/0", ArtifactSHA256: sourceBytesHash(raw), CanonicalID: "operation:widgets.get", CanonicalPointer: "/operations/0/operation", Generation: descriptor.Staged.Identity.Digest}
	annotation := sourceSemanticAnnotation{Key: key, Citation: facts.Refs["summary"], Clause: "Get widgets", IntendedBindings: []sourceLaneTargetRef{ref}}
	cells := classifySourceLanes(key, facts, &annotation)
	cell := requireSourceLane(t, cells, "direct_read", "applicable")
	// Identity/provenance can be observed without promoting unavailable response
	// semantics or behavior. A correct observation must survive that distinction.
	found := false
	for _, d := range cell.Diagnostics {
		if d.Code == "target_response_contract_unverified" {
			found = true
		}
	}
	if !found {
		t.Fatalf("exact admitted positive target not distinguished from missing provenance: %+v", cell.Diagnostics)
	}
	if cell.State != "mapped_unproven" || len(cell.ProofRefs) != 0 {
		t.Fatalf("identity promoted behavior: %+v", cell)
	}
}

func TestSourceLaneGitLabBridge(t *testing.T) {
	for _, tc := range []struct{ name, sourcePath, targetPath, bridge, want string }{
		{"declared exact boundary", "/api/v4/projects/{id}", "/projects/{id}", `{"source_prefix":"/api/v4","connector_prefix":""}`, "target_contract_unverified"},
		{"prefix lookalike", "/api/v40/projects/{id}", "/0/projects/{id}", `{"source_prefix":"/api/v4","connector_prefix":""}`, "target_semantics_mismatch"},
		{"undeclared prefix", "/api/v4/projects/{id}", "/projects/{id}", `null`, "target_semantics_mismatch"},
		{"unrelated prefix declaration", "/private/projects/{id}", "/projects/{id}", `{"source_prefix":"/private","connector_prefix":""}`, "target_semantics_mismatch"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			node, err := json.Marshal(map[string]any{"id": "provider.project", "protocol": "rest", "method": "GET", "path": tc.sourcePath, "source_operation": map[string]any{"summary": "Get project", "responses": map[string]any{"204": map[string]any{"description": "No content"}}}})
			if err != nil {
				t.Fatal(err)
			}
			key := sourceOperationKey{Connector: "fixture", Inventory: "primary", ID: "provider.project"}
			row := retainedSourceOperation{Key: key, Observed: true, Node: node, Pointer: "/rest/operations/0"}
			doc := retainedSourceDocument{ID: "fixture", Payload: json.RawMessage(`{"rest":{"path_bridge":` + tc.bridge + `,"operations":[` + string(node) + `]}}`)}
			facts := normalizeSourceFacts(row, doc, nil)
			raw, err := json.Marshal(map[string]any{"operations": []any{map[string]any{"id": "project", "kind": "rest_read", "rest": map[string]any{"method": "GET", "path": tc.targetPath}}}})
			if err != nil {
				t.Fatal(err)
			}
			artifact := "internal/connectors/defs/fixture/operations.json"
			facts.bindings = &sourceLaneBindingInputs{Artifacts: map[string][]byte{artifact: raw}}
			ref := sourceLaneTargetRef{Kind: "operation", Connector: "fixture", ID: "project", Lane: "direct_read", Artifact: artifact, Pointer: "/operations/0"}
			a := sourceSemanticAnnotation{Key: key, Citation: facts.Refs["summary"], Clause: "Get project", IntendedBindings: []sourceLaneTargetRef{ref}}
			cells := classifySourceLanes(key, facts, &a)
			cell := requireSourceLane(t, cells, "direct_read", "applicable")
			found := false
			for _, d := range cell.Diagnostics {
				if d.Code == tc.want {
					found = true
				}
			}
			if !found {
				t.Fatalf("bridge %s: want %s got %+v", tc.name, tc.want, cell.Diagnostics)
			}
			if facts.Path != tc.sourcePath {
				t.Fatalf("source route was rewritten: %q", facts.Path)
			}
		})
	}
}

func TestSourceLaneBindingParameterContract(t *testing.T) {
	for _, tc := range []struct {
		name, sourceType string
		sourceRequired   bool
		sourceMaximum    any
		want             string
	}{
		{"matching bounded parameter", "integer", true, 100, "target_response_contract_unverified"},
		{"equivalent exponent bound", "integer", true, json.Number("1e2"), "target_response_contract_unverified"},
		{"equivalent decimal bound", "integer", true, json.Number("100.0"), "target_response_contract_unverified"},
		{"wrong existing type", "string", true, 100, "target_parameter_mismatch"},
		{"wrong requiredness", "integer", false, 100, "target_parameter_mismatch"},
		{"wrong bound", "integer", true, 10, "target_parameter_mismatch"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			lock := operationDirectReadLockForSemanticAdmissionTest()
			lock.Operations[0].Operation = json.RawMessage(`{"id":"widgets.get","kind":"rest_read","summary":"Get widgets","risk":"low","approval":"none","output_policy":"json_redacted","rest":{"method":"GET","path":"/widgets","max_bytes":1024,"parameters":[{"name":"limit","in":"query","type":"integer","required":true,"minimum":1,"maximum":100}]}}`)
			lock.Operations[0].Commands[0].Command = json.RawMessage(`{"path":"widgets get","summary":"Get widgets","intent":"direct_read","availability":"implemented","operation":"widgets.get","api_surface":[{"method":"GET","path":"/widgets"}],"output_policy":"json_redacted","flags":[{"name":"limit","maps_to":"query.limit","type":"integer","required":true,"minimum":1,"maximum":100}]}`)
			descriptor, err := canonicalizeVNextSourceLock(lock)
			if err != nil {
				t.Fatal(err)
			}
			node, err := json.Marshal(map[string]any{"id": "provider.widgets.get", "protocol": "rest", "method": "GET", "path": "/widgets", "source_operation": map[string]any{"summary": "Get widgets", "parameters": []any{map[string]any{"name": "limit", "in": "query", "required": tc.sourceRequired, "schema": map[string]any{"type": tc.sourceType, "minimum": 1, "maximum": tc.sourceMaximum}}}, "responses": map[string]any{"200": map[string]any{"description": "Unresolved response"}}}})
			if err != nil {
				t.Fatal(err)
			}
			key := sourceOperationKey{Connector: "acme", Inventory: "primary", ID: "provider.widgets.get"}
			row := retainedSourceOperation{Key: key, Observed: true, Node: node, Pointer: "/rest/operations/0"}
			doc := retainedSourceDocument{ID: "acme:primary", Payload: json.RawMessage(`{"rest":{"operations":[` + string(node) + `]}}`)}
			facts := normalizeSourceFacts(row, doc, nil)
			artifact := "internal/connectors/defs/acme/operations.json"
			raw := descriptor.Staged.Outputs["operations.json"]
			facts.bindings = &sourceLaneBindingInputs{Artifacts: map[string][]byte{artifact: raw}, Canonical: map[string]vNextCanonicalDescriptor{"acme": descriptor}}
			ref := sourceLaneTargetRef{Kind: "operation", Connector: "acme", ID: "widgets.get", Lane: "direct_read", Artifact: artifact, Pointer: "/operations/0", ArtifactSHA256: sourceBytesHash(raw), CanonicalID: "operation:widgets.get", CanonicalPointer: "/operations/0/operation", Generation: descriptor.Staged.Identity.Digest}
			a := sourceSemanticAnnotation{Key: key, Citation: facts.Refs["summary"], Clause: "Get widgets", IntendedBindings: []sourceLaneTargetRef{ref}}
			cell := requireSourceLane(t, classifySourceLanes(key, facts, &a), "direct_read", "applicable")
			found := false
			for _, d := range cell.Diagnostics {
				if d.Code == tc.want {
					found = true
				}
			}
			if !found {
				t.Fatalf("%s: want %s got %+v", tc.name, tc.want, cell.Diagnostics)
			}
		})
	}
}

func TestSourceLaneBindingNumericOracle(t *testing.T) {
	for _, tc := range []struct {
		a, b         string
		equal, known bool
	}{
		{"9007199254740993", "9007199254740992", false, true},
		{"0.0001", "1e-4", true, true},
		{"-0.0", "0", true, true},
		{"-10", "10", false, true},
		{"1e999999999", "10e999999998", true, true},
		{"null", "0", false, true},
		{`"100"`, "100", false, false},
	} {
		equal, known := sourceLaneNumericBoundEqual([]byte(tc.a), []byte(tc.b))
		if equal != tc.equal || known != tc.known {
			t.Errorf("%s vs %s: got (%v,%v), want (%v,%v)", tc.a, tc.b, equal, known, tc.equal, tc.known)
		}
	}
}

func sourceBindingRepositoryFixture(t *testing.T) (string, sourceLaneCohort, vNextCanonicalDescriptor) {
	t.Helper()
	lock := operationDirectReadLockForSemanticAdmissionTest()
	descriptor, err := canonicalizeVNextSourceLock(lock)
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	outputs := map[string][]byte{}
	for name, raw := range descriptor.Staged.Outputs {
		outputs[name] = raw
	}
	raw, err := json.Marshal(lock)
	if err != nil {
		t.Fatal(err)
	}
	outputs["source.lock.json"] = raw
	for name, raw := range outputs {
		file := filepath.Join(root, "internal/connectors/defs/acme", name)
		if err := os.MkdirAll(filepath.Dir(file), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(file, raw, 0600); err != nil {
			t.Fatal(err)
		}
	}
	return root, sourceLaneCohort{Inventories: []sourceInventoryAnchor{{Connector: "acme", Inventory: "primary"}, {Connector: "acme", Inventory: "supplement"}}}, descriptor
}

func sourceBindingFixtureSnapshot(t *testing.T, root string) map[string]string {
	t.Helper()
	result := map[string]string{}
	err := filepath.WalkDir(root, func(name string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(root, name)
		if err != nil {
			return err
		}
		if entry.Type()&os.ModeSymlink != 0 {
			target, err := os.Readlink(name)
			if err != nil {
				return err
			}
			result[rel] = "symlink:" + target
			return nil
		}
		raw, err := os.ReadFile(name)
		if err != nil {
			return err
		}
		result[rel] = string(raw)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func TestSourceLaneBindingCollector(t *testing.T) {
	root, cohort, descriptor := sourceBindingRepositoryFixture(t)
	before := sourceBindingFixtureSnapshot(t, root)
	inputs := collectSourceLaneBindings(context.Background(), root, cohort)
	if len(inputs.Observations) != 0 {
		t.Fatalf("valid canonical/execution inputs rejected: %+v", inputs.Observations)
	}
	if len(inputs.Canonical) != 1 || inputs.Canonical["acme"].Staged.Identity.Digest != descriptor.Staged.Identity.Digest || len(inputs.Bundles) != 1 {
		t.Fatalf("missing exact canonical/execution observation: %+v", inputs.Observations)
	}
	if len(inputs.Artifacts) != len(descriptor.Staged.Outputs) {
		t.Fatalf("collected%d files, want%d closed runtime artifacts", len(inputs.Artifacts), len(descriptor.Staged.Outputs))
	}
	for name := range inputs.Artifacts {
		if strings.Contains(name, "source.lock") {
			t.Fatalf("authoring lock entered runtime inputs: %s", name)
		}
	}
	if after := sourceBindingFixtureSnapshot(t, root); !reflect.DeepEqual(before, after) {
		t.Fatal("read-only collection changed fixture files")
	}
}

func TestSourceLaneBindingCollectorFaults(t *testing.T) {
	for _, which := range []string{"cancel", "malformed authoring", "unsafe artifact"} {
		t.Run(which, func(t *testing.T) {
			root, cohort, _ := sourceBindingRepositoryFixture(t)
			ctx := context.Background()
			switch which {
			case "cancel":
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			case "malformed authoring":
				if err := os.WriteFile(filepath.Join(root, "internal/connectors/defs/acme/source.lock.json"), []byte(`{"schema_version":4,"operations":`), 0600); err != nil {
					t.Fatal(err)
				}
			case "unsafe artifact":
				name := filepath.Join(root, "internal/connectors/defs/acme/metadata.json")
				if err := os.Remove(name); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink("spec.json", name); err != nil {
					t.Fatal(err)
				}
			}
			before := sourceBindingFixtureSnapshot(t, root)
			inputs := collectSourceLaneBindings(ctx, root, cohort)
			want := "cancelled"
			if which == "malformed authoring" {
				want = "canonical_input_invalid"
			}
			if which == "unsafe artifact" {
				want = "artifact_input_invalid"
			}
			found := false
			for _, o := range inputs.Observations {
				if o.Code == want {
					found = true
				}
			}
			if !found {
				t.Fatalf("fault%s lacked %s: %+v", which, want, inputs.Observations)
			}
			if which == "malformed authoring" && (len(inputs.Canonical) != 0 || len(inputs.Bundles) != 1) {
				t.Fatalf("authoring fault erased runtime observation or created canonical admission")
			}
			if which == "unsafe artifact" && len(inputs.Bundles) != 0 {
				t.Fatal("unsafe artifact became a loaded bundle")
			}
			if which == "cancel" && (len(inputs.Bundles) != 0 || len(inputs.Canonical) != 0 || len(inputs.Artifacts) != 0) {
				t.Fatal("cancelled collection continued to artifact/admission work")
			}
			if after := sourceBindingFixtureSnapshot(t, root); !reflect.DeepEqual(before, after) {
				t.Fatal("failure changed fixture state")
			}
		})
	}
}

func TestSourceLaneBindingCollectorCurrentClosedInputs(t *testing.T) {
	root, err := repoRoot()
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(root, "data/connector-canon/batch1-source-lane-cohort.json"))
	if err != nil {
		t.Fatal(err)
	}
	var cohort sourceLaneCohort
	if err := json.Unmarshal(raw, &cohort); err != nil {
		t.Fatal(err)
	}
	inputs := collectSourceLaneBindings(context.Background(), root, cohort)
	expected := map[string]bool{"asana": true, "gitlab": true, "bitbucket": true, "circleci": true, "dockerhub": true, "jira": true, "notion": true, "sentry": true, "stripe": true, "vercel": true}
	if len(inputs.Bundles) != len(expected) {
		t.Fatalf("closed input bundles%d, want10; observations=%+v", len(inputs.Bundles), inputs.Observations)
	}
	for name := range expected {
		if _, exists := inputs.Bundles[name]; !exists {
			t.Errorf("closed execution input %s missing", name)
		}
	}
	if len(inputs.Canonical) != 2 {
		t.Fatalf("canonical observations%d, want existing Asana/GitLab pair; observations=%+v", len(inputs.Canonical), inputs.Observations)
	}
	for _, name := range []string{"asana", "gitlab"} {
		if _, exists := inputs.Canonical[name]; !exists {
			t.Errorf("canonical observation missing%s", name)
		}
	}
	for _, observation := range inputs.Observations {
		if observation.Code != "canonical_input_unavailable" || observation.Connector == "asana" || observation.Connector == "gitlab" {
			t.Errorf("unexpected current input observation: %+v", observation)
		}
	}
	// These are static input observations only, never executable lane proof.
}

func TestSourceLaneReferenceCounterexample(t *testing.T) {
	for _, tc := range []struct{ name, raw, id, pointer, code string }{
		{"exact operation", `{"operations":[{"id":"a"},{"id":"b"}]}`, "b", "/operations/1", ""},
		{"duplicate operation", `{"operations":[{"id":"b"},{"id":"b"}]}`, "b", "", "target_identity_ambiguous"},
		{"absent operation", `{"operations":[{"id":"a"}]}`, "b", "", "target_absent"},
		{"wrong collection", `{"actions":[{"name":"b"}]}`, "b", "", "target_shape_invalid"},
	} {
		pointer, code := sourceLaneTargetPointer([]byte(tc.raw), sourceLaneTargetRef{Kind: "operation", ID: tc.id})
		if pointer != tc.pointer || code != tc.code {
			t.Errorf("%s: pointer/code=(%s,%s), want(%s,%s)", tc.name, pointer, code, tc.pointer, tc.code)
		}
	}
}
