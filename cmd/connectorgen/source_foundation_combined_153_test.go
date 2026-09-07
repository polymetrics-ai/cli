package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors/commandrunner"
	"polymetrics.ai/internal/connectors/engine"
)

// Independently retained provider body; it is not taken from generated output.
const sourceFoundationProviderBody153 = `{"type":"object","additionalProperties":false,"required":["label","attributes","targets"],"properties":{"label":{"type":"string"},"attributes":{"type":"object","additionalProperties":false,"required":["owner","active"],"properties":{"owner":{"type":"string"},"active":{"type":"boolean"}}},"targets":{"type":"array","minItems":1,"maxItems":2,"items":{"type":"object","additionalProperties":false,"required":["id"],"properties":{"id":{"type":"string"}}}}}}`

func sourceFoundationCombinedRetained155(t *testing.T, missingCLI bool) (string, sourceArtifactPin, sourceFoundationAssessmentDocument) {
	return sourceFoundationCombinedRetainedProfile155(t, missingCLI, "single")
}

func sourceFoundationCombinedRetainedProfile155(t *testing.T, missingCLI bool, profile string) (string, sourceArtifactPin, sourceFoundationAssessmentDocument) {
	t.Helper()
	repo, proofDocument := sourceFoundationProofFixture153(t, true)
	writeSourceFoundationObservationFixture(t, repo, proofDocument)
	if len(proofDocument.Records) != 4 {
		t.Fatal("four original proof records required")
	}
	lock := minimalVNextLockForTest()
	lock.Lanes["etl"], lock.Lanes["direct_write"] = "unsupported", "implemented"
	lock.Metadata = json.RawMessage(`{"name":"acme","display_name":"Acme","description":"Structured fixture","integration_type":"api","release_stage":"ga","capabilities":{"check":true,"read":false,"write":true,"query":false,"cdc":false,"dynamic_schema":false}}`)
	lock.ConfigSchema = json.RawMessage(`{"type":"object","properties":{"api_key":{"type":"string","x-secret":true}},"required":["api_key"]}`)
	lock.HTTP = json.RawMessage(`{"url":"https://api.acme.example","headers":{},"auth":[{"mode":"api_key_header","header":"X-API-Key","value":"{{ secrets.api_key }}"}],"pagination":{"type":"none"},"check":{"method":"GET","path":"/check"},"error_map":[]}`)
	// Target declaration is literal and independent from the retained source.
	lock.Schemas = map[string]json.RawMessage{"schemas/request.json": json.RawMessage(`{"type":"object","additionalProperties":false,"required":["label","attributes","targets"],"properties":{"label":{"type":"string"},"attributes":{"type":"object","additionalProperties":false,"required":["owner","active"],"properties":{"owner":{"type":"string"},"active":{"type":"boolean"}}},"targets":{"type":"array","minItems":1,"maxItems":2,"items":{"type":"object","additionalProperties":false,"required":["id"],"properties":{"id":{"type":"string"}}}}}}`)}
	lock.Operations = []vNextOperationDescriptor{{ID: "operation:widgets.create", Source: json.RawMessage(`{"provider_operation":"createWidget","method":"POST","path":"/workspaces/{workspace_id}/widgets"}`), SchemaRefs: vNextSchemaReferences{Request: "schemas/request.json"},
		Operation: json.RawMessage(`{"id":"widgets.create","kind":"rest_write","summary":"Create a widget","risk":"high","approval":"plan-preview-confirm-execute","output_policy":"json","mutation_class":"create","batchable":false,"rest":{"method":"POST","path":"/workspaces/{workspace_id}/widgets","content_type":"application/json","max_bytes":1024,"response":{"success_statuses":["204"]},"parameters":[{"name":"workspace_id","in":"path","type":"string","required":true},{"name":"dry_run","in":"query","type":"boolean"}],"body_schema":` + string(lock.Schemas["schemas/request.json"]) + `}}`),
		Commands:  []vNextCommandDescriptor{{Order: 0, Command: json.RawMessage(`{"path":"widgets create","summary":"Create a widget","intent":"direct_write","availability":"implemented","operation":"widgets.create","api_surface":[{"method":"POST","path":"/workspaces/{workspace_id}/widgets"}],"output_policy":"json","flags":[{"name":"workspace-id","type":"string","maps_to":"path.workspace_id","required":true},{"name":"dry-run","type":"boolean","maps_to":"query.dry_run"},{"name":"label","type":"string","maps_to":"body.label","required":true},{"name":"attributes","type":"json","maps_to":"body.attributes","required":true},{"name":"targets","type":"json","maps_to":"body.targets","required":true}]}`)}},
	}}
	if profile == "joined" || profile == "unjoined" {
		var second map[string]any
		if json.Unmarshal(lock.Operations[0].Commands[0].Command, &second) != nil {
			t.Fatal("command literal")
		}
		second["path"] = "widgets create-secondary"
		encoded, err := json.Marshal(second)
		if err != nil {
			t.Fatal(err)
		}
		lock.Operations[0].Commands = append(lock.Operations[0].Commands, vNextCommandDescriptor{Order: 1, Command: encoded})
	}
	if strings.HasPrefix(profile, "auth_") {
		var http map[string]any
		if json.Unmarshal(lock.HTTP, &http) != nil {
			t.Fatal("HTTP literal")
		}
		auth := http["auth"].([]any)[0].(map[string]any)
		switch profile {
		case "auth_header":
			auth["header"] = "Other-Key"
		case "auth_query":
			auth["mode"] = "api_key_query"
			auth["param"] = "api_key"
			delete(auth, "header")
		case "auth_prefix":
			auth["prefix"] = "Bearer "
		case "auth_first_none":
			http["auth"] = []any{map[string]any{"mode": "none"}, auth}
		case "auth_conditional":
			auth["when"] = "false"
		case "auth_template":
			auth["value"] = "prefix {{ secrets.api_key }}"
		case "auth_nonsecret":
			lock.ConfigSchema = json.RawMessage(`{"type":"object","properties":{"api_key":{"type":"string"}},"required":["api_key"]}`)
		}
		encoded, err := json.Marshal(http)
		if err != nil {
			t.Fatal(err)
		}
		lock.HTTP = encoded
	}
	if profile == "body_form" || profile == "body_flat" {
		lock.Schemas["schemas/request.json"] = json.RawMessage(`{"type":"object","additionalProperties":false,"required":["label"],"properties":{"label":{"type":"string"}}}`)
		var operation, command map[string]any
		if json.Unmarshal(lock.Operations[0].Operation, &operation) != nil || json.Unmarshal(lock.Operations[0].Commands[0].Command, &command) != nil {
			t.Fatal("flat body literals")
		}
		operation["rest"].(map[string]any)["body_schema"] = lock.Schemas["schemas/request.json"]
		command["flags"] = command["flags"].([]any)[:3]
		encodedOperation, err := json.Marshal(operation)
		if err != nil {
			t.Fatal(err)
		}
		lock.Operations[0].Operation = encodedOperation
		lock.Operations[0].Commands[0].Command, err = json.Marshal(command)
		if err != nil {
			t.Fatal(err)
		}
	}
	if profile == "body_form" {
		lock.Operations[0].Operation = bytes.ReplaceAll(lock.Operations[0].Operation, []byte(`"content_type":"application/json"`), []byte(`"content_type":"application/x-www-form-urlencoded"`))
	}
	lock.CLI = json.RawMessage(`{"usage":"pm acme <command>","tagline":"Acme typed commands"}`)
	descriptor, err := canonicalizeVNextSourceLock(lock)
	if err != nil {
		t.Fatalf("real canonical body/auth admission: %v", err)
	}
	raw, err := json.Marshal(lock)
	if err != nil {
		t.Fatal(err)
	}
	prefix := "internal/connectors/defs/acme/"
	proofWrite(t, repo, prefix+"source.lock.json", raw)
	for name, raw := range descriptor.Staged.Outputs {
		proofWrite(t, repo, prefix+name, raw)
	}
	bundle, err := engine.Load(newVNextExecutionFS("acme", descriptor.Staged.Outputs), "acme")
	if err != nil {
		t.Fatal(err)
	}
	connector := engine.New(bundle, nil)
	paths := [][]string{{"widgets", "create"}}
	if profile == "joined" || profile == "unjoined" {
		paths = append(paths, []string{"widgets", "create-secondary"})
	}
	for _, commandPath := range paths {
		if err := commandrunner.Preflight(connector, commandPath); err != nil {
			t.Fatalf("actual generated command preflight: %v", err)
		}
		flags := map[string][]string{"workspace-id": {"workspace-1"}, "dry-run": {"true"}, "label": {"fixture widget"}, "attributes": {`{"owner":"owner-1","active":true}`}, "targets": {`[{"id":"target-1"}]`}}
		if profile == "body_form" || profile == "body_flat" {
			delete(flags, "attributes")
			delete(flags, "targets")
		}
		write, err := commandrunner.BuildWriteCommand(t.Context(), connector, commandrunner.Request{Path: commandPath, Flags: flags})
		if err != nil || write.PathParams["workspace_id"] != "workspace-1" || write.Query["dry_run"] != "true" || write.Record["label"] != "fixture widget" || (profile != "body_form" && profile != "body_flat" && !reflect.DeepEqual(write.Record["attributes"], map[string]any{"owner": "owner-1", "active": true})) {
			t.Fatalf("actual bounded command fields: %+v %v", write, err)
		}
	}
	// Full retained source: two independently named keys, all fourteen cells.
	sourceRaw := []byte(`{"schema_version":2,"connector":"acme","source_contract":{"components":{"securitySchemes":{"WidgetKey":{"type":"apiKey","in":"header","name":"X-API-Key"}}}},"rest":{"operations":[{"id":"source.a","operation_id":"createWidget","protocol":"rest","method":"post","path":"/workspaces/{workspace_id}/widgets","source_operation":{"summary":"Create a widget","parameters":[{"name":"workspace_id","in":"path","required":true,"schema":{"type":"string"}},{"name":"dry_run","in":"query","schema":{"type":"boolean"}}],"security":[{"WidgetKey":[]}],"requestBody":{"required":true,"content":{"application/json":{"schema":` + sourceFoundationProviderBody153 + `}}},"responses":{"204":{"description":"No content"}}}},{"id":"source.b","operation_id":"getOther","protocol":"rest","method":"get","path":"/other","source_operation":{"summary":"Get another item","responses":{"204":{"description":"No content"}}}}]},"counts":{"total":2}}`)
	if profile == "body_form" || profile == "body_flat" {
		sourceRaw = bytes.ReplaceAll(sourceRaw, []byte(sourceFoundationProviderBody153), []byte(`{"type":"object","additionalProperties":false,"required":["label"],"properties":{"label":{"type":"string"}}}`))
	}
	if profile == "body_form" {
		sourceRaw = bytes.ReplaceAll(sourceRaw, []byte(`"application/json"`), []byte(`"application/x-www-form-urlencoded"`))
	}
	// Source auth negatives retain current canonical execution and exact citations.
	// They reach the requirement mechanism consumer, not a stale binding check.
	switch profile {
	case "auth_scoped":
		sourceRaw = bytes.ReplaceAll(sourceRaw, []byte(`"security":[{"WidgetKey":[]}]`), []byte(`"security":[{"WidgetKey":["required-scope"]}]`))
	case "auth_conjunctive":
		sourceRaw = bytes.ReplaceAll(sourceRaw, []byte(`"security":[{"WidgetKey":[]}]`), []byte(`"security":[{"WidgetKey":[],"OtherKey":[]}]`))
	case "auth_oauth":
		sourceRaw = bytes.ReplaceAll(sourceRaw, []byte(`"type":"apiKey","in":"header","name":"X-API-Key"`), []byte(`"type":"oauth2"`))
	}
	proofWrite(t, repo, "source.json", sourceRaw)
	_, cohort := sourceInventoryFixture(t, []string{"source.a", "source.b"}, 2)
	cohort.Inventories[0].Connector = "acme"
	cohort.Inventories[0].SHA256 = sourceBytesHash(sourceRaw)
	universe, err := buildSourceFoundationUniverse(t.Context(), repo, cohort, nil)
	if err != nil {
		t.Fatal(err)
	}
	facts := universe.manifest.SourceOperations[0].Facts
	if universe.manifest.SourceOperations[0].Source.Key.ID != "source.a" || facts.OperationID != "createWidget" {
		t.Fatal("independent retained source identity changed")
	}
	cite := func(pointer string) sourceFactRef {
		t.Helper()
		value, err := sourceJSONPointer(facts.Document, pointer)
		if err != nil {
			t.Fatal(err)
		}
		canonical, err := canonicalSourceJSON(value)
		if err != nil {
			t.Fatal(err)
		}
		return sourceFactRef{DocumentID: facts.Refs["source_operation"].DocumentID, Pointer: pointer, ValueSHA256: sourceBytesHash(canonical)}
	}
	mediaPointer := "application~1json"
	if profile == "body_form" {
		mediaPointer = "application~1x-www-form-urlencoded"
	}
	body := cite(facts.Refs["request_body"].Pointer + "/content/" + mediaPointer + "/schema")
	op := sourceLaneTargetRef{Kind: "operation", Connector: "acme", ID: "widgets.create", Lane: "direct_write", Artifact: prefix + "operations.json", Pointer: "/operations/0", ArtifactSHA256: sourceBytesHash(descriptor.Staged.Outputs["operations.json"]), CanonicalID: "operation:widgets.create", CanonicalPointer: "/operations/0/operation", Generation: descriptor.Staged.Identity.Digest, SchemaRole: sourceLaneSchemaRequest, SourceSchema: &body}
	wholeBody := ""
	op.FieldMappings = []sourceLaneFieldMapping{{Source: body, Target: sourceLaneFieldTarget{Kind: sourceLaneFieldSchema, Pointer: &wholeBody}}}
	command := op
	command.Kind, command.ID, command.Artifact, command.Pointer, command.CanonicalPointer = "command", "widgets create", prefix+"cli_surface.json", "/commands/0", "/operations/0/commands/0"
	command.ArtifactSHA256 = sourceBytesHash(descriptor.Staged.Outputs["cli_surface.json"])
	annotation := sourceSemanticAnnotation{Key: universe.manifest.SourceOperations[0].Source.Key, Citation: facts.Refs["summary"], Clause: "Create a widget", IntendedBindings: []sourceLaneTargetRef{op, command}}
	secondCommand := command
	secondCommand.ID = "widgets create-secondary"
	secondCommand.Pointer = "/commands/1"
	secondCommand.CanonicalPointer = "/operations/0/commands/1"
	if profile == "joined" {
		annotation.IntendedBindings = append(annotation.IntendedBindings, secondCommand)
	}
	annotations := []sourceSemanticAnnotation{annotation}
	universe, err = buildSourceFoundationUniverse(t.Context(), repo, cohort, annotations)
	if err != nil {
		t.Fatal(err)
	}
	complete := requireSourceLane(t, universe.manifest.SourceOperations[0].Lanes, "direct_write", "applicable")
	expectedReferences := 2
	if profile == "joined" {
		expectedReferences = 3
	}
	if profile == "body_form" {
		expectedReferences = 0
		mediaDeficits := 0
		for _, diagnostic := range complete.Diagnostics {
			if diagnostic.Code == "target_media_contract_unverified" && diagnostic.Severity == "deficit" {
				mediaDeficits++
			}
		}
		if mediaDeficits != 2 {
			t.Fatal("form source projection lost exact operation/command media deficits")
		}
	}
	if len(complete.References) != expectedReferences {
		t.Fatalf("actual complete canonical bindings: %+v", complete.Diagnostics)
	}
	proofs := map[string]sourceFoundationProofRecord{}
	for _, record := range proofDocument.Records {
		proofs[record.ID] = record
	}
	makeRequirement := func(id, proofID, status string, ref sourceLaneTargetRef, refs []sourceFactRef) sourceFoundationRequirement {
		proof := proofs[proofID]
		return sourceFoundationRequirement{ID: id, SourceRefs: refs, Statement: proof.Assertion.Statement, AtlasLookup: sourceFoundationLookup{Atlas: universe.atlasPin, Candidates: []sourceFoundationLookupCandidate{{AtlasID: proof.AtlasID, Contract: proof.Contract, Disposition: status, Rationale: "Exact retained source and reviewed mechanism"}}}, Assessment: status, ProofIDs: []string{proofID}, AffectedArtifacts: []string{ref.Artifact}, EvidenceRequirements: []string{"Exact retained source fit and current reviewed mechanism witness"}, DecisionRefs: []sourceFoundationDecision{}, FitBindings: []sourceLaneTargetRef{ref}}
	}
	bodyReq := makeRequirement("shared-structured-body", "runtime.direct-execution.v1.structured-rest-body", "existing_shared_capability", op, []sourceFactRef{facts.Refs["request_body"], body})
	authReq := makeRequirement("shared-static-header-auth", "runtime.direct-execution.v1.static-api-key-header", "existing_shared_capability", op, []sourceFactRef{facts.Refs["security"], facts.Refs["security_schemes"], cite(facts.Refs["security_schemes"].Pointer + "/WidgetKey")})
	assessment := sourceFoundationAssessmentDocument{SchemaVersion: 1, Kind: "foundation_cell_assessments", Atlas: universe.atlasPin, Assessments: []sourceFoundationCellAssessment{{Key: annotation.Key, Lane: "direct_write", NextOwner: "polymetrics.ai/internal/connectors/engine", Requirements: []sourceFoundationRequirement{bodyReq, authReq}}}}
	if missingCLI {
		local := makeRequirement("local-command-artifact", "runtime.direct-execution.v1.structured-rest-body", "connector_local_configuration", command, bodyReq.SourceRefs)
		local.EvidenceRequirements = []string{"materialize the exact canonical CLI artifact"}
		assessment.Assessments[0].Requirements = append(assessment.Assessments[0].Requirements, local)
		if profile == "joined" {
			secondLocal := makeRequirement("local-secondary-command-artifact", "runtime.direct-execution.v1.structured-rest-body", "connector_local_configuration", secondCommand, bodyReq.SourceRefs)
			secondLocal.EvidenceRequirements = []string{"materialize the exact secondary canonical CLI artifact"}
			assessment.Assessments[0].Requirements = append(assessment.Assessments[0].Requirements, secondLocal)
		}
	}
	writeSourceFoundationAssessmentFixture(t, repo, assessment)
	raw, err = json.Marshal(cohort)
	if err != nil {
		t.Fatal(err)
	}
	proofWrite(t, repo, sourceLaneCohortPath, raw)
	raw, err = json.Marshal(struct {
		SchemaVersion int                        `json:"schema_version"`
		Annotations   []sourceSemanticAnnotation `json:"annotations"`
	}{1, annotations})
	if err != nil {
		t.Fatal(err)
	}
	proofWrite(t, repo, sourceLaneAnnotationsPath, raw)
	var manifest, diag bytes.Buffer
	if code := runSourceLanesContext(t.Context(), []string{"source-lanes", "--repo", repo}, &manifest, &diag); code != 0 {
		t.Fatal(diag.String())
	}
	proofWrite(t, repo, sourceLaneManifestPath, manifest.Bytes())
	baseline := sourceFoundationObligationDocument{SchemaVersion: 1, Kind: "foundation_known_obligations", BaselineManifest: sourceArtifactPin{Path: sourceLaneManifestPath, SHA256: sourceBytesHash(manifest.Bytes()), Bytes: int64(manifest.Len())}, SourceEvidenceSHA256: sourceBytesHash(sourceRaw), HistoricalReceiverCandidates: []sourceFoundationCell{}, SentryRegistration: []sourceFoundationCell{}, Obligations: []sourceFoundationKnownObligation{{Identity: sourceFoundationCell{Key: annotation.Key, Lane: "direct_write"}, Source: universe.manifest.SourceOperations[0].Source, SourceRefs: []sourceFactRef{facts.Refs["request_body"]}, SourceDocumentPins: []sourceArtifactPin{{Path: "source.json", SHA256: sourceBytesHash(sourceRaw), Bytes: int64(len(sourceRaw))}}}}}
	expectedCells := []sourceFoundationCell{}
	for _, id := range []string{"source.a", "source.b"} {
		for _, lane := range []string{"direct_read", "direct_write", "binary_download", "binary_upload", "etl", "reverse_etl", "sync_transport"} {
			expectedCells = append(expectedCells, sourceFoundationCell{Key: sourceOperationKey{Connector: "acme", Inventory: "primary", ID: id}, Lane: lane})
		}
	}
	baseline.BaselineUniverseSHA256, err = sourceFoundationCellsHash(expectedCells)
	if err != nil {
		t.Fatal(err)
	}
	baselinePin := writeSourceFoundationObligationsFixture(t, repo, baseline)
	if missingCLI {
		if err := os.Remove(filepath.Join(repo, prefix+"cli_surface.json")); err != nil {
			t.Fatal(err)
		}
		manifest.Reset()
		diag.Reset()
		// Missing CLI changes the physical generation. CP12 must continue to
		// emit its complete invalid report and exit1, including saved checks.
		if code := runSourceLanesContext(t.Context(), []string{"source-lanes", "--repo", repo}, &manifest, &diag); code != 1 {
			t.Fatalf("missing CLI source-lanes exit=%d: %s", code, diag.String())
		}
		var actual sourceLaneManifest
		if json.Unmarshal(manifest.Bytes(), &actual) != nil || actual.Validation.Status != "invalid" || actual.Validation.Errors != 1 || actual.Validation.Deficits != map[bool]int{true: 5, false: 4}[profile == "joined"] {
			t.Fatal("actual missing-file producer diagnostics changed")
		}
		cell := requireSourceLane(t, actual.SourceOperations[0].Lanes, "direct_write", "applicable")
		if len(cell.References) != 0 {
			t.Fatal("missing file retained fabricated complete-generation references")
		}
		proofWrite(t, repo, sourceLaneManifestPath, manifest.Bytes())
	}
	return repo, baselinePin, assessment
}

func sourceFoundationCombinedFixture153(t *testing.T, missingCLI bool) (string, sourceArtifactPin, sourceFoundationDemandInputs, sourceFoundationAssessmentDocument) {
	t.Helper()
	repo, baseline, assessment := sourceFoundationCombinedRetained155(t, missingCLI)
	var cohort sourceLaneCohort
	raw, err := os.ReadFile(filepath.Join(repo, sourceLaneCohortPath))
	if err != nil || json.Unmarshal(raw, &cohort) != nil {
		t.Fatal("retained fixture cohort unreadable")
	}
	var annotations struct {
		Annotations []sourceSemanticAnnotation `json:"annotations"`
	}
	raw, err = os.ReadFile(filepath.Join(repo, sourceLaneAnnotationsPath))
	if err != nil || json.Unmarshal(raw, &annotations) != nil {
		t.Fatal("retained fixture annotations unreadable")
	}
	universe, err := buildSourceFoundationUniverse(t.Context(), repo, cohort, annotations.Annotations)
	if err != nil {
		t.Fatal(err)
	}
	inputs, err := observeSourceFoundationDemandInputs(t.Context(), repo, sourceFoundationAssessmentsPath, universe, baseline)
	if err != nil {
		t.Fatal(err)
	}
	return repo, baseline, inputs, assessment
}

func TestSourceFoundationCombinedBodyAuthLocal153(t *testing.T) {
	for _, missing := range []bool{false, true} {
		t.Run(map[bool]string{false: "materialized", true: "canonical_intended_with_local"}[missing], func(t *testing.T) {
			repo, baseline, inputs, _ := sourceFoundationCombinedFixture153(t, missing)
			before, err := json.Marshal(inputs.assessments.universe.manifest)
			if err != nil {
				t.Fatal(err)
			}
			got, err := buildSourceFoundationRegister(t.Context(), repo, inputs)
			if err != nil {
				t.Fatal(err)
			}
			want := map[string]string{"shared-structured-body": "existing_shared_capability", "shared-static-header-auth": "existing_shared_capability"}
			state := "materialized"
			if missing {
				want["local-command-artifact"] = "connector_local_configuration"
				state = "canonical_intended"
			}
			if got.Coverage.UniverseCount != 14 || len(got.Coverage.Assessed) != 1 || len(got.Coverage.Unassessed) != 13 || len(got.Known) != 1 || len(got.Requirements) != len(want) || len(got.Adopters) != len(want) {
				t.Fatal("independent source/aspect coverage collapsed")
			}
			for _, row := range got.Requirements {
				if row.Status != want[row.ID] || len(row.MechanismFits) != 1 || row.MechanismFits[0].DeclarationState != state || row.Identity.Key.ID != "source.a" || row.Identity.Lane != "direct_write" {
					t.Fatalf("independent aspect mismatch: %+v", row)
				}
				mechanism, proof := "rest_write_structured_json_body", "runtime.direct-execution.v1.structured-rest-body"
				if row.ID == "shared-static-header-auth" {
					mechanism, proof = "rest_write_static_api_key_header", "runtime.direct-execution.v1.static-api-key-header"
				}
				if row.MechanismFits[0].Mechanism != mechanism || row.MechanismFits[0].ProofID != proof || len(row.MechanismFits[0].Selectors) < 2 {
					t.Fatal("body/auth proof or exact selectors borrowed")
				}
				sourceFoundationExpectedSelectors156(t, repo, row.ID, row.MechanismFits[0])
				delete(want, row.ID)
			}
			if len(want) != 0 {
				t.Fatal("independent body/auth/local row erased")
			}
			after, err := json.Marshal(inputs.assessments.universe.manifest)
			if err != nil || !bytes.Equal(before, after) {
				t.Fatal("fit changed actual lane/source authority")
			}
			var out, diag bytes.Buffer
			if code := runSourceDemandsPolicy(t.Context(), []string{"source-demands", "--repo", repo}, &out, &diag, baseline); code != 0 {
				t.Fatalf("real full generated command: %s", diag.String())
			}
			var command sourceFoundationRegister
			if err := json.Unmarshal(out.Bytes(), &command); err != nil {
				t.Fatal(err)
			}
			// Compare the complete emitted tuples. The private matched-review
			// mechanism in Proofs is deliberately not a JSON trust field.
			for _, pair := range [][2]any{{command.Requirements, got.Requirements}, {command.Adopters, got.Adopters}} {
				actual, err := json.Marshal(pair[0])
				if err != nil {
					t.Fatal(err)
				}
				expected, err := json.Marshal(pair[1])
				if err != nil || !bytes.Equal(actual, expected) {
					t.Fatal("actual command disagrees with exact requirements/adopter wire tuples")
				}
			}
		})
	}
}

// Fixed selector coordinates are independent from the emitted selector list.
// Current source.lock supplies bytes only; it cannot choose expected coordinates.
func sourceFoundationExpectedSelectors156(t *testing.T, repo, id string, fit sourceFoundationMechanismFit) {
	t.Helper()
	prefix := "internal/connectors/defs/acme/"
	raw, err := os.ReadFile(filepath.Join(repo, prefix+"source.lock.json"))
	if err != nil {
		t.Fatal(err)
	}
	var lock vNextSourceLock
	if json.Unmarshal(raw, &lock) != nil {
		t.Fatal("expected canonical source")
	}
	descriptor, err := canonicalizeVNextSourceLock(lock)
	if err != nil {
		t.Fatal(err)
	}
	coordinates := [][2]string{{"operations.json", "/operations/0"}, {"schemas/request.json", ""}}
	switch id {
	case "shared-static-header-auth":
		coordinates = [][2]string{{"operations.json", "/operations/0"}, {"spec.json", "/properties/api_key"}, {"streams.json", "/base/auth/0"}}
	case "local-command-artifact":
		coordinates = append([][2]string{{"cli_surface.json", "/commands/0"}}, coordinates...)
	}
	if len(fit.Selectors) != len(coordinates) {
		t.Fatal("missing or extra exact selector")
	}
	for i, coordinate := range coordinates {
		raw := descriptor.Staged.Outputs[coordinate[0]]
		value, err := sourceJSONPointer(raw, coordinate[1])
		if err != nil {
			t.Fatal(err)
		}
		canonical, err := canonicalSourceJSON(value)
		if err != nil {
			t.Fatal(err)
		}
		want := sourceFoundationSelector{Path: prefix + coordinate[0], Pointer: coordinate[1], ArtifactSHA256: sourceBytesHash(raw), ValueSHA256: sourceBytesHash(canonical)}
		if fit.Selectors[i] != want {
			t.Fatalf("exact selector %d differs: got%+v want%+v", i, fit.Selectors[i], want)
		}
	}
}
