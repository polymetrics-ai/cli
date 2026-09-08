package main

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestSourceProjection206InheritedParameters(t *testing.T) {
	for _, archive := range []bool{false, true} {
		name := "raw_schema3"
		if archive {
			name = "archive_schema2"
		}
		t.Run(name, func(t *testing.T) {
			for _, duplicate := range []bool{false, true} {
				label := "inherited_and_override"
				if duplicate {
					label = "duplicate_inherited_scope"
				}
				t.Run(label, func(t *testing.T) {
					inherited := `[{"in":"path","name":"namespace","required":true,"schema":{"type":"string"}},{"in":"query","name":"limit","schema":{"type":"integer"}}]`
					if duplicate {
						inherited = strings.TrimSuffix(inherited, "]") + `,{"in":"path","name":"namespace","required":true,"schema":{"type":"string"}}]`
					}
					op := `{"parameters":[{"in":"query","name":"limit","required":true,"schema":{"type":"integer","maximum":25}}],"responses":{"200":{"description":"ok"}}}`
					node := `{"id":"fixture.list","protocol":"rest","method":"GET","path":"/namespaces/{namespace}"}`
					doc := retainedSourceDocument{ID: "archive", Payload: json.RawMessage(`{"schema_version":3,"rest":{}}`)}
					raw := &retainedSourceDocument{ID: "raw", Payload: json.RawMessage(`{"paths":{"/namespaces/{namespace}":{"parameters":` + inherited + `,"get":` + op + `}}}`)}
					pathPointer := "/paths/~1namespaces~1{namespace}/parameters"
					opPointer := "/paths/~1namespaces~1{namespace}/get/parameters/0"
					documentID := "raw"
					if archive {
						op = strings.TrimSuffix(op, "}") + `,"path_parameters":` + inherited + `}`
						node = strings.TrimSuffix(node, "}") + `,"source_operation":` + op + `}`
						doc.Payload = json.RawMessage(`{"schema_version":2,"rest":{"operations":[` + node + `]}}`)
						raw = nil
						pathPointer = "/rest/operations/0/source_operation/path_parameters"
						opPointer = "/rest/operations/0/source_operation/parameters/0"
						documentID = "archive"
					}
					facts := normalizeSourceFacts(retainedSourceOperation{Observed: true, Node: json.RawMessage(node), Pointer: "/rest/operations/0"}, doc, raw)
					if facts.Status != "available" || len(facts.Parameters) != 2 {
						t.Fatalf("actual normalizer lost inherited parameters: status=%s parameters=%+v diagnostics=%v", facts.Status, facts.Parameters, facts.Diagnostics)
					}
					for i, want := range []struct{ in, name, pointer string }{{"path", "namespace", pathPointer + "/0"}, {"query", "limit", opPointer}} {
						got := facts.Parameters[i]
						if got.In != want.in || got.Name != want.name || !got.Required || got.Ref.DocumentID != documentID || got.Ref.Pointer != want.pointer || got.Ref.ValueSHA256 == "" {
							t.Errorf("parameter or source precedence changed: %+v; want %+v", got, want)
						}
					}
					if !strings.Contains(string(facts.Parameters[1].Node), `"maximum":25`) {
						t.Error("operation-level schema override lost")
					}
					wantDiagnostic := "source_parameter_duplicate:" + pathPointer + "/0:" + pathPointer + "/2"
					if duplicate && (len(facts.Diagnostics) != 1 || facts.Diagnostics[0] != wantDiagnostic) {
						t.Errorf("duplicate inherited scope not diagnosed: %v", facts.Diagnostics)
					} else if !duplicate && len(facts.Diagnostics) != 0 {
						t.Errorf("valid inherited scope rejected: %v", facts.Diagnostics)
					}
				})
			}
		})
	}
}

func TestSourceProjection206ClosedArm(t *testing.T) {
	legacy := minimalVNextLockForTest()
	legacyBytes, err := json.Marshal(legacy)
	if err != nil {
		t.Fatal(err)
	}
	var root map[string]json.RawMessage
	if err := json.Unmarshal(legacyBytes, &root); err != nil {
		t.Fatal(err)
	}
	delete(root, "operations")
	delete(root, "schemas")
	root["source_projection"] = json.RawMessage(`{"version":1,"inventories":[{"id":"primary","path":"sources/operations.json","sha256":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","bytes":1024}],"semantics":[]}`)
	encode := func(values map[string]json.RawMessage) []byte {
		t.Helper()
		data, err := json.Marshal(values)
		if err != nil {
			t.Fatal(err)
		}
		return data
	}
	t.Run("supported_legacy_control", func(t *testing.T) {
		lock, err := decodeVNextSourceLock(legacyBytes)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := canonicalizeVNextSourceLock(lock); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("source_only_grammar", func(t *testing.T) {
		if _, err := decodeVNextSourceLock(encode(root)); err != nil {
			t.Fatalf("strict source-only grammar rejected: %v", err)
		}
	})
	for _, field := range []string{"operations", "schemas", "execution", "cli"} {
		t.Run("mixed_arm_"+field, func(t *testing.T) {
			mixed := make(map[string]json.RawMessage, len(root)+1)
			for k, v := range root {
				mixed[k] = v
			}
			mixed[field] = json.RawMessage(`null`)
			if _, err := decodeVNextSourceLock(encode(mixed)); err == nil {
				t.Fatal("present legacy execution arm accepted alongside projection")
			}
		})
	}
	for _, projection := range []string{
		`null`, `{}`, `{"version":2,"inventories":[]}`, `{"version":1,"inventories":[]}`,
		`{"version":1,"inventories":[{"id":"primary","path":"../outside","sha256":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","bytes":1}]}`,
		`{"version":1,"inventories":[{"id":"primary","path":"sources/a","sha256":"bad","bytes":1}]}`,
		`{"version":1,"inventories":[{"id":"primary","path":"sources/a","sha256":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","bytes":1,"maps_to":"query.x"}]}`,
	} {
		t.Run("invalid_projection", func(t *testing.T) {
			invalid := make(map[string]json.RawMessage, len(root))
			for k, v := range root {
				invalid[k] = v
			}
			invalid["source_projection"] = json.RawMessage(projection)
			if _, err := decodeVNextSourceLock(encode(invalid)); err == nil {
				t.Fatal("invalid source projection accepted")
			}
		})
	}
}

func TestSourceProjection206SemanticGrammar(t *testing.T) {
	valid := `{"source":{"inventory":"primary","id":"fixture.read"},"collection":{"records":{"response":{"status":"200","media":"application/json","pointer":"/items"}},"primary_key":["/id"]}}`
	for _, tc := range []struct {
		name, semantic string
		valid          bool
	}{
		{"response_collection", valid, true},
		{"empty_coordinate", strings.Replace(valid, `{"response":{"status":"200","media":"application/json","pointer":"/items"}}`, `{}`, 1), false},
		{"mixed_coordinate", strings.Replace(valid, `"response":`, `"parameter":{"in":"query","name":"limit"},"response":`, 1), false},
		{"missing_status", strings.Replace(valid, `"status":"200",`, ``, 1), false},
		{"missing_media", strings.Replace(valid, `"media":"application/json",`, ``, 1), false},
		{"invalid_pointer", strings.Replace(valid, `/items`, `/bad~2escape`, 1), false},
		{"missing_primary_key", strings.Replace(valid, `["/id"]`, `[]`, 1), false},
		{"duplicate_primary_key", strings.Replace(valid, `["/id"]`, `["/id","/id"]`, 1), false},
		{"invalid_primary_key", strings.Replace(valid, `["/id"]`, `["id"]`, 1), false},
		{"unbound_inventory", strings.Replace(valid, `"inventory":"primary"`, `"inventory":"other"`, 1), false},
		{"unknown_field", strings.Replace(valid, `"collection":`, `"maps_to":"body.raw","collection":`, 1), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			raw := []byte(`{"schema_version":4,"connector":"acme","source_projection":{"version":1,"inventories":[{"id":"primary","path":"sources/operations.json","sha256":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","bytes":1024}],"semantics":[` + tc.semantic + `]}}`)
			_, err := decodeVNextSourceLock(raw)
			if tc.valid && err != nil {
				t.Fatalf("valid semantic grammar refused: %v", err)
			}
			if !tc.valid && err == nil {
				t.Fatal("invalid semantic grammar admitted")
			}
		})
	}
}

func TestSourceProjection206WriteGrammar(t *testing.T) {
	valid := `{"source":{"inventory":"primary","id":"fixture.create"},"effect":"mutation","write":{"row_delivery":"one_request","batchable":true,"risk":"medium","retry":"single_attempt"}}`
	for _, tc := range []struct {
		name, semantic string
		valid          bool
	}{
		{"explicit_batchable", valid, true},
		{"explicit_individual", strings.Replace(valid, `"batchable":true`, `"batchable":false`, 1), true},
		{"missing_batchability", strings.Replace(valid, `"batchable":true,`, ``, 1), false},
		{"null_batchability", strings.Replace(valid, `"batchable":true`, `"batchable":null`, 1), false},
		{"read_effect", strings.Replace(valid, `"mutation"`, `"read"`, 1), false},
		{"unknown_delivery", strings.Replace(valid, `"one_request"`, `"arbitrary_requests"`, 1), false},
		{"implicit_retry", strings.Replace(valid, `"single_attempt"`, `"automatic"`, 1), false},
		{"unknown_risk", strings.Replace(valid, `"medium"`, `"unlimited"`, 1), false},
		{"authored_mapping", strings.Replace(valid, `"row_delivery"`, `"maps_to":"body.raw","row_delivery"`, 1), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			raw := []byte(`{"schema_version":4,"connector":"acme","source_projection":{"version":1,"inventories":[{"id":"primary","path":"sources/operations.json","sha256":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","bytes":1024}],"semantics":[` + tc.semantic + `]}}`)
			_, err := decodeVNextSourceLock(raw)
			if tc.valid && err != nil {
				t.Fatalf("selected typed write grammar refused: %v", err)
			}
			if !tc.valid && err == nil {
				t.Fatal("invalid typed write semantics admitted")
			}
		})
	}
}
