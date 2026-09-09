package main

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors/engine"
)

// The normal source-lane producer must read an independently supplied review
// catalog.  Test fixtures are deliberately scoped to this test root; no
// production reviewer identity is authored here.
func TestBatch1CP18GitLabAdmission277ReviewCatalog(t *testing.T) {
	externalReviewRecord := func(record sourceLaneProofRecord) sourceLaneProofRecord {
		// The catalog protocol deliberately excludes fixture authority. Keep the
		// test hermetic while exercising the same primary/C2-shaped tuple that a
		// separately supplied reviewer must attest in production.
		record.Key.Inventory = "primary"
		record.Scope = "connector_hermetic"
		record.ExecutionClass = "C2"
		return record
	}
	writeCatalog := func(t *testing.T, root string, document sourceLaneProofReviewCatalogDocument) {
		t.Helper()
		raw, err := json.Marshal(document)
		if err != nil {
			t.Fatal(err)
		}
		proofWrite(t, root, sourceLaneProofReviewCatalogPath, raw)
	}

	t.Run("valid_external_review", func(t *testing.T) {
		root, record, _ := proofFixture(t)
		record = externalReviewRecord(record)
		proofDocument(t, root, []sourceLaneProofRecord{record})
		writeCatalog(t, root, sourceLaneProofReviewCatalogDocument{
			SchemaVersion: 1,
			Reviews:       []sourceLaneExternalProofReview{{Record: record}},
		})
		catalog, diagnostics := sourceLaneProofCatalogFromRepository(context.Background(), root, proofFixturePolicy())
		if len(diagnostics) != 0 {
			t.Fatalf("external review catalog diagnostics = %+v", diagnostics)
		}
		proofs := loadSourceLaneProofs(context.Background(), root, proofFixturePolicy(), catalog)
		if !proofs.accepted[record.ID] {
			t.Fatalf("valid external review did not admit record: %+v", proofs.Diagnostics)
		}
	})

	t.Run("missing_review_is_not_admitted", func(t *testing.T) {
		root, record, _ := proofFixture(t)
		record = externalReviewRecord(record)
		proofDocument(t, root, []sourceLaneProofRecord{record})
		catalog, diagnostics := sourceLaneProofCatalogFromRepository(context.Background(), root, proofFixturePolicy())
		if len(diagnostics) != 0 {
			t.Fatalf("missing catalog diagnostics = %+v", diagnostics)
		}
		proofs := loadSourceLaneProofs(context.Background(), root, proofFixturePolicy(), catalog)
		if proofs.accepted[record.ID] || !proofHasDiagnostic(proofs.Diagnostics, "proof_review_unavailable", "deficit") {
			t.Fatalf("missing review admitted record: %+v", proofs.Diagnostics)
		}
	})

	t.Run("stale_review_is_rejected", func(t *testing.T) {
		root, record, _ := proofFixture(t)
		record = externalReviewRecord(record)
		proofDocument(t, root, []sourceLaneProofRecord{record})
		stale := record
		stale.ReceiptSHA256 = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
		writeCatalog(t, root, sourceLaneProofReviewCatalogDocument{
			SchemaVersion: 1,
			Reviews:       []sourceLaneExternalProofReview{{Record: stale}},
		})
		catalog, diagnostics := sourceLaneProofCatalogFromRepository(context.Background(), root, proofFixturePolicy())
		if len(diagnostics) != 0 {
			t.Fatalf("stale catalog diagnostics = %+v", diagnostics)
		}
		proofs := loadSourceLaneProofs(context.Background(), root, proofFixturePolicy(), catalog)
		if proofs.accepted[record.ID] || !proofHasDiagnostic(proofs.Diagnostics, "proof_review_stale", "error") {
			t.Fatalf("stale review admitted record: %+v", proofs.Diagnostics)
		}
	})

	t.Run("mismatched_review_is_rejected", func(t *testing.T) {
		root, record, _ := proofFixture(t)
		record = externalReviewRecord(record)
		proofDocument(t, root, []sourceLaneProofRecord{record})
		mismatch := record
		mismatch.Key.ID = "fixture.other"
		writeCatalog(t, root, sourceLaneProofReviewCatalogDocument{
			SchemaVersion: 1,
			Reviews:       []sourceLaneExternalProofReview{{Record: mismatch}},
		})
		catalog, diagnostics := sourceLaneProofCatalogFromRepository(context.Background(), root, proofFixturePolicy())
		if len(diagnostics) != 0 {
			t.Fatalf("mismatched catalog diagnostics = %+v", diagnostics)
		}
		proofs := loadSourceLaneProofs(context.Background(), root, proofFixturePolicy(), catalog)
		if proofs.accepted[record.ID] || !proofHasDiagnostic(proofs.Diagnostics, "proof_review_mismatch", "error") {
			t.Fatalf("mismatched review admitted record: %+v", proofs.Diagnostics)
		}
	})
}

// A source-lock operation that owns a typed write response must be admissible
// through the existing schema-4 producer. This is separate from the runtime
// consumer assertion: it guards the source -> canonical -> writes.json join.
func TestBatch1CP18GitLabAdmission277WriteResponseSchemaRole(t *testing.T) {
	response := json.RawMessage(`{"type":"object","properties":{"key":{"type":"string"}}}`)
	operation := vNextCanonicalOperation{
		Index: 17,
		SchemaRefs: vNextSchemaReferences{
			Response: "schemas/responses/ci-variable.json",
		},
		Write: &vNextCanonicalWrite{Spec: engine.WriteAction{
			Name:           "delete_ci_variable",
			ResponseSchema: response,
		}},
	}
	if err := validateVNextSchemaRoles(operation, map[string]json.RawMessage{
		"schemas/responses/ci-variable.json": response,
	}); err != nil {
		t.Fatalf("typed write response schema rejected by source-lock producer: %v", err)
	}
}

// A typed write has one physical route and one request-side parameter contract,
// even when its source declaration also binds a distinct typed success response.
// JSON Schema's document-level $schema keyword must not make that route or its
// request projection unverifiable.
func TestBatch1CP18GitLabAdmission277CrossRoleWriteProjection(t *testing.T) {
	for _, tc := range []struct {
		name           string
		removePathMap  bool
		wantReferences int
		wantCodes      []string
	}{
		{name: "request_mapping_covers_typed_response_role", wantReferences: 2},
		{name: "missing_request_mapping_is_refused", removePathMap: true, wantCodes: []string{"target_parameter_mismatch", "target_path_projection_mismatch"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			requestSchema := json.RawMessage(`{"$schema":"http://json-schema.org/draft-07/schema#","type":"object","properties":{"id":{"type":"string"},"data":{"type":"object","properties":{"name":{"type":"string"}},"required":["name"],"additionalProperties":false}},"required":["id","data"],"additionalProperties":false}`)
			responseSchema := json.RawMessage(`{"type":"object","properties":{"id":{"type":"string"}},"required":["id"],"additionalProperties":false}`)
			key, facts, annotation := sourceBindingFixture099F(t, "body", func(lock *vNextSourceLock) {
				lock.Metadata = json.RawMessage(`{"name":"acme","display_name":"Acme","description":"Test connector","integration_type":"api","release_stage":"ga","capabilities":{"check":true,"read":false,"write":true,"query":false,"cdc":false,"dynamic_schema":false}}`)
				lock.Lanes["reverse_etl"] = "implemented"
				lock.Schemas["schemas/request.json"] = requestSchema
				lock.Schemas["schemas/response.json"] = responseSchema
				lock.Operations[0].SchemaRefs.Response = "schemas/response.json"
				var write map[string]json.RawMessage
				if err := json.Unmarshal(lock.Operations[0].Write, &write); err != nil {
					t.Fatal(err)
				}
				write["record_schema"] = requestSchema
				write["response_schema"] = responseSchema
				raw, err := json.Marshal(write)
				if err != nil {
					t.Fatal(err)
				}
				lock.Operations[0].Write = raw
			})
			facts = sourceBindingRepin099F(t, key, facts, func(document map[string]any) {
				operation := document["rest"].(map[string]any)["operations"].([]any)[0].(map[string]any)["source_operation"].(map[string]any)
				operation["responses"] = map[string]any{
					"200": map[string]any{
						"description": "The created widget.",
						"content": map[string]any{
							"application/json": map[string]any{
								"schema": map[string]any{
									"type": "object",
									"properties": map[string]any{
										"id": map[string]any{"type": "string"},
									},
									"required":             []any{"id"},
									"additionalProperties": false,
								},
							},
						},
					},
				}
			})

			response := canonicalSourceLaneTargetRef(annotation.IntendedBindings[0])
			response.SchemaRole = sourceLaneSchemaResponse
			responseCitation := sourceBindingCitation099F(t, facts, "/rest/operations/0/source_operation/responses/200/content/application~1json/schema")
			response.SourceSchema = &responseCitation
			root := ""
			response.FieldMappings = []sourceLaneFieldMapping{{
				Source: responseCitation,
				Target: sourceLaneFieldTarget{Kind: sourceLaneFieldSchema, Pointer: &root},
			}}
			annotation.IntendedBindings = append(annotation.IntendedBindings, response)
			if tc.removePathMap {
				annotation.IntendedBindings[0].FieldMappings = annotation.IntendedBindings[0].FieldMappings[1:]
			}

			wantReferences := tc.wantReferences
			if tc.removePathMap {
				wantReferences = 0
			}
			sourceBindingOutcome099F(t, key, facts, annotation, wantReferences, tc.wantCodes...)
		})
	}
}

// The retained GitLab source-lane projection identifies its admitted
// generation as sha256:<digest>. Proof records must accept that exact
// coordinate while retaining bare-digest compatibility for existing records.
func TestBatch1CP18GitLabAdmission277ProofGenerationIdentity(t *testing.T) {
	base := func(generation string) sourceLaneProofRecord {
		return sourceLaneProofRecord{
			ID:   "gitlab-proof-generation",
			Key:  sourceOperationKey{Connector: "gitlab", Inventory: "primary", ID: "deleteApiV4AdminCiVariablesKey"},
			Lane: "direct_write",
			Targets: []sourceLaneTargetRef{{
				Kind: "write", Connector: "gitlab", ID: "source_write_test", Lane: "direct_write",
				Artifact: "internal/connectors/defs/gitlab/writes.json", Pointer: "/actions/0",
				ArtifactSHA256: strings.Repeat("a", 64), CanonicalID: "write:source_write_test",
				CanonicalPointer: "/operations/0/write", Generation: generation,
			}},
			Inputs: []sourceLaneProofInput{
				{Path: "cmd/connectorgen/proof.go", SHA256: strings.Repeat("b", 64), Role: "code"},
				{Path: "cmd/connectorgen/proof_test.go", SHA256: strings.Repeat("c", 64), Role: "test"},
				{Path: "go.mod", SHA256: strings.Repeat("d", 64), Role: "dependency"},
				{Path: "go.sum", SHA256: strings.Repeat("e", 64), Role: "dependency"},
			},
			TestPath: "cmd/connectorgen/proof_test.go", TestSymbol: "TestProof", SelectedTest: "TestProof/selected",
			Package: "polymetrics.ai/cmd/connectorgen", ExecutionClass: "C2", Scope: "connector_hermetic",
			ReceiptPath: "data/connector-canon/proof-receipts/gitlab/proof.jsonl", ReceiptSHA256: strings.Repeat("f", 64),
			ObservableContract: "Exact retained generation identity is part of the proof target.",
			Limitations:        []string{"Hermetic C2 shape check only."},
		}
	}
	for _, tc := range []struct {
		name, generation string
		want             bool
	}{
		{name: "bare_digest", generation: strings.Repeat("a", 64), want: true},
		{name: "canonical_sha256_digest", generation: "sha256:" + strings.Repeat("a", 64), want: true},
		{name: "malformed_prefix", generation: "sha512:" + strings.Repeat("a", 64), want: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := sourceLaneProofShape(base(tc.generation)); got != tc.want {
				t.Fatalf("generation shape accepted=%t want=%t generation=%q", got, tc.want, tc.generation)
			}
		})
	}
}
