package main

import (
	"encoding/json"
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
