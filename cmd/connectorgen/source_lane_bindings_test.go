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
