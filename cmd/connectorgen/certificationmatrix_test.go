package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

var certificationMatrixTestCache struct {
	once   sync.Once
	matrix capabilityMatrix
	err    error
}

func TestCertificationDiscoverFunctionKindsFromRuntimeSource(t *testing.T) {
	kinds, err := discoverFunctionKinds(repoRootForCertificationTest(t))
	if err != nil {
		t.Fatalf("discoverFunctionKinds() error = %v", err)
	}

	want := map[string]bool{
		"capability:read":            false,
		"capability:write":           false,
		"operation:rest_read":        false,
		"operation:rest_write":       false,
		"operation:graphql_mutation": false,
		"operation:binary_download":  false,
		"operation:file_upload":      false,
	}
	for _, kind := range kinds {
		if _, ok := want[kind.ID]; ok {
			want[kind.ID] = true
		}
	}
	for id, found := range want {
		if !found {
			t.Errorf("discovered kinds missing %q: %#v", id, kinds)
		}
	}
}

func TestCertificationOperationExecutorAnnotationsAreRealPaths(t *testing.T) {
	kinds, err := discoverFunctionKinds(repoRootForCertificationTest(t))
	if err != nil {
		t.Fatalf("discoverFunctionKinds() error = %v", err)
	}

	wantImplemented := map[string]bool{
		"operation:rest_read":        false,
		"operation:rest_write":       false,
		"operation:graphql_mutation": false,
		"operation:binary_download":  false,
	}
	for _, kind := range kinds {
		if _, ok := wantImplemented[kind.ID]; ok && kind.ExecutorSource != "" {
			wantImplemented[kind.ID] = true
		}
	}
	for id, found := range wantImplemented {
		if !found {
			t.Errorf("%s has no real executor source annotation", id)
		}
	}
}

func TestCertificationMatrixRejectsDatabaseWriteStubs(t *testing.T) {
	matrix := certificationMatrixForTest(t)

	for _, connectorName := range []string{"postgres", "mysql"} {
		cell, ok := capabilityCellFor(matrix, connectorName, "capability:write")
		if !ok {
			t.Fatalf("%s write cell missing", connectorName)
		}
		if !cell.Applicable {
			t.Errorf("%s Write stub was hidden as not applicable", connectorName)
		}
		if cell.Declared {
			t.Errorf("%s Write stub reported declared", connectorName)
		}
		if cell.Implemented {
			t.Errorf("%s Write stub reported implemented", connectorName)
		}
	}
}

func TestCertificationMatrixRecognizesEngineCapabilityMethods(t *testing.T) {
	matrix := certificationMatrixForTest(t)

	cell, ok := capabilityCellFor(matrix, "github", "capability:read")
	if !ok {
		t.Fatal("github read cell missing")
	}
	if !cell.Implemented {
		t.Fatal("generic engine read implementation reported false")
	}
}

func TestCertificationApplicableCellWithoutLiveEvidencePreventsCompletion(t *testing.T) {
	cell := certificationCell{
		FunctionKind:  "capability:read",
		Applicable:    true,
		Declared:      true,
		Implemented:   true,
		FixtureTested: true,
		LiveTested:    false,
	}
	if certificationComplete([]certificationCell{cell}) {
		t.Fatal("applicable cell without live evidence completed certification")
	}
}

func TestCertificationRejectsNotApplicableWithoutNamedReason(t *testing.T) {
	cell := certificationCell{
		FunctionKind: "operation:graphql_query",
		NotApplicable: &notApplicableReason{
			Code: "n/a",
		},
	}
	err := validateCertificationCell(cell)
	if err == nil {
		t.Fatal("validateCertificationCell() error = nil, want named non-applicable reason rejection")
	}
	if !strings.Contains(err.Error(), "not_applicable") {
		t.Fatalf("validateCertificationCell() error = %q, want not_applicable context", err)
	}
}

func TestCertificationRejectsMalformedAcceptedLiveEvidence(t *testing.T) {
	err := validateAcceptedEvidence(acceptedEvidence{
		SchemaVersion:   1,
		Scope:           evidenceScopeCapability,
		Status:          evidenceStatusPassed,
		CredentialScope: credentialScopeFullParity,
		CredentialNote:  fullParityCredentialNote,
		Connector:       "github",
		FunctionKind:    "capability:read",
	})
	if err == nil {
		t.Fatal("validateAcceptedEvidence() error = nil, want missing real-provider evidence rejection")
	}
	if !strings.Contains(err.Error(), "provider") {
		t.Fatalf("validateAcceptedEvidence() error = %q, want provider context", err)
	}
}

func TestCertificationRejectsProoflessAcceptedLiveEvidence(t *testing.T) {
	err := validateAcceptedEvidence(acceptedEvidence{
		SchemaVersion:   1,
		Scope:           evidenceScopeCapability,
		Status:          evidenceStatusPassed,
		Connector:       "github",
		FunctionKind:    "capability:read",
		Provider:        "github",
		ExecutedAt:      "2026-08-10T00:00:00Z",
		RunID:           "live-run-123",
		CredentialScope: credentialScopeFullParity,
		CredentialNote:  fullParityCredentialNote,
	})
	if err == nil {
		t.Fatal("validateAcceptedEvidence() error = nil, want proof-bearing live evidence rejection")
	}
	if !strings.Contains(err.Error(), "proof") {
		t.Fatalf("validateAcceptedEvidence() error = %q, want proof context", err)
	}
}

func TestCertificationSanitizesPreparedValuesBeforeProofPersistence(t *testing.T) {
	secret := "token-value-used-by-the-prepared-request"
	evidence, err := newProofBearingEvidence(completedLiveEvidence{
		SchemaVersion:        certificationSchemaVersion,
		Scope:                evidenceScopeCapability,
		Connector:            "github",
		FunctionKind:         "operation:rest_read",
		Provider:             "github",
		ExecutedAt:           "2026-08-10T00:00:00Z",
		RunID:                "live-run-123",
		PMBinarySHA256:       strings.Repeat("a", 64),
		PMCommand:            "pm etl read --connector github --json",
		Passed:               true,
		CredentialFullParity: true,
		RepositorySalt:       []byte("0123456789abcdef0123456789abcdef"),
		PreparedValues:       []string{secret},
		HTTPExchanges: []completedHTTPExchange{{
			Operation: "github.repos.get",
			Request: completedHTTPRequest{
				Method:  "GET",
				Target:  "https://api.github.example/repos/acme/widget?token=" + secret,
				Headers: map[string][]string{"Authorization": {"Bearer " + secret}},
				Body:    []byte(`{"credential":"` + secret + `"}`),
			},
			Response: completedHTTPResponse{
				Status:  200,
				Headers: map[string][]string{"X-Provider-Token": {secret}},
				Body:    []byte(`{"access_token":"` + secret + `"}`),
			},
		}},
	})
	if err != nil {
		t.Fatalf("newProofBearingEvidence() error = %v", err)
	}

	raw, err := json.Marshal(evidence)
	if err != nil {
		t.Fatalf("marshal proof-bearing evidence: %v", err)
	}
	if bytes.Contains(raw, []byte(secret)) {
		t.Fatalf("persisted proof leaked a prepared credential: %s", raw)
	}
	proof := evidence.Proof.HTTPExchanges[0]
	if !isFingerprintSequence(proof.Request.Target) {
		t.Fatalf("proof did not fingerprint target: %#v", proof.Request.Target)
	}
	if len(evidence.Proof.CredentialFingerprints) != 1 || !isFingerprintSequence(evidence.Proof.CredentialFingerprints[0]) {
		t.Fatalf("proof did not retain a safe credential fingerprint: %#v", evidence.Proof.CredentialFingerprints)
	}
	var responseBody map[string]any
	if err := json.Unmarshal(proof.Response.Body.Value, &responseBody); err != nil {
		t.Fatalf("decode sanitized response body: %v", err)
	}
	if value, ok := responseBody["access_token"].(string); !ok || !isFingerprintSequence(value) {
		t.Fatalf("proof did not fingerprint response credential: %#v", proof.Response.Body)
	}
	if err := validateAcceptedEvidence(evidence); err != nil {
		t.Fatalf("validateAcceptedEvidence(sanitized) error = %v", err)
	}
}

func TestCertificationEvidenceWriterUsesRepositoryLocalSaltBeforePersistence(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "internal", "connectors", "certifications", "evidence", "github.json")
	secret := "credential-only-in-memory"
	completed := completedLiveEvidence{
		SchemaVersion:        certificationSchemaVersion,
		Scope:                evidenceScopeCapability,
		Connector:            "github",
		FunctionKind:         "operation:rest_read",
		Provider:             "github",
		ExecutedAt:           "2026-08-10T00:00:00Z",
		RunID:                "github-local-salt",
		PMBinarySHA256:       strings.Repeat("d", 64),
		PMCommand:            "pm etl read --connector github --json",
		Passed:               true,
		CredentialFullParity: true,
		PreparedValues:       []string{secret},
		HTTPExchanges: []completedHTTPExchange{{
			Operation: "github.repos.get",
			Request: completedHTTPRequest{
				Method:  "GET",
				Target:  "https://api.github.example/repos/acme/widget",
				Headers: map[string][]string{"Authorization": {"Bearer " + secret}},
			},
			Response: completedHTTPResponse{Status: 200, Body: []byte(`{"token":"` + secret + `"}`)},
		}},
	}
	evidence, err := writeProofBearingEvidence(root, path, completed)
	if err != nil {
		t.Fatalf("writeProofBearingEvidence() error = %v", err)
	}
	if len(evidence.Proof.CredentialFingerprints) != 1 {
		t.Fatalf("credential fingerprints = %#v, want one", evidence.Proof.CredentialFingerprints)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read persisted evidence: %v", err)
	}
	if bytes.Contains(raw, []byte(secret)) {
		t.Fatalf("persisted evidence leaked credential: %s", raw)
	}
	saltPath := filepath.Join(root, repositoryFingerprintSaltPath)
	salt, err := os.ReadFile(saltPath)
	if err != nil {
		t.Fatalf("read local fingerprint salt: %v", err)
	}
	if bytes.Contains(salt, []byte(secret)) {
		t.Fatal("repository fingerprint salt contains a credential")
	}
	info, err := os.Stat(saltPath)
	if err != nil {
		t.Fatalf("stat local fingerprint salt: %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("fingerprint salt mode = %o, want 600", info.Mode().Perm())
	}
	second, err := repositoryFingerprintSalt(root)
	if err != nil {
		t.Fatalf("reload repository fingerprint salt: %v", err)
	}
	if !bytes.Equal(salt, second) {
		t.Fatal("repository fingerprint salt was not deterministic within one checkout")
	}
	other, err := repositoryFingerprintSalt(t.TempDir())
	if err != nil {
		t.Fatalf("create second repository fingerprint salt: %v", err)
	}
	if bytes.Equal(salt, other) {
		t.Fatal("repository fingerprint salts unexpectedly match across checkouts")
	}
}

func TestCertificationRejectsUnsafeEmbeddedProof(t *testing.T) {
	secret := "unsafe-proof-token"
	evidence, err := newProofBearingEvidence(completedLiveEvidence{
		SchemaVersion:        certificationSchemaVersion,
		Scope:                evidenceScopeCapability,
		Connector:            "github",
		FunctionKind:         "operation:rest_read",
		Provider:             "github",
		ExecutedAt:           "2026-08-10T00:00:00Z",
		RunID:                "live-run-unsafe-proof",
		PMBinarySHA256:       strings.Repeat("b", 64),
		PMCommand:            "pm etl read --connector github --json",
		Passed:               true,
		CredentialFullParity: true,
		RepositorySalt:       []byte("0123456789abcdef0123456789abcdef"),
		PreparedValues:       []string{secret},
		HTTPExchanges: []completedHTTPExchange{{
			Operation: "github.repos.get",
			Request:   completedHTTPRequest{Method: "GET", Target: "https://api.github.example/repos/acme/widget"},
			Response:  completedHTTPResponse{Status: 200},
		}},
	})
	if err != nil {
		t.Fatalf("newProofBearingEvidence() error = %v", err)
	}
	evidence.Proof.HTTPExchanges[0].Response.Body.Value = json.RawMessage(`"unredacted provider response"`)
	if err := validateAcceptedEvidence(evidence); err == nil {
		t.Fatal("validateAcceptedEvidence() error = nil, want unsafe embedded response rejection")
	}
}

func TestCertificationArtifactProofValidationPrecedesCodeDrift(t *testing.T) {
	path := filepath.Join(t.TempDir(), "capability-matrix.json")
	artifact := capabilityMatrix{
		SchemaVersion: certificationSchemaVersion,
		FunctionKinds: []functionKind{{ID: "capability:read"}},
		Connectors: []capabilityConnector{{
			Name: "github",
			Cells: []certificationCell{{
				FunctionKind:    "capability:read",
				Applicable:      true,
				Declared:        true,
				Implemented:     true,
				FixtureTested:   true,
				LiveTested:      true,
				FixtureEvidence: []string{"fixture.json"},
				LiveEvidence:    []evidencePointer{},
			}},
		}},
	}
	raw, err := marshalGeneratedJSON(artifact)
	if err != nil {
		t.Fatalf("marshal artifact: %v", err)
	}
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		t.Fatalf("write artifact: %v", err)
	}
	err = checkGeneratedArtifact(path, []byte("{}\n"))
	if err == nil || !strings.Contains(err.Error(), "live_evidence") {
		t.Fatalf("checkGeneratedArtifact() error = %v, want live evidence validation before drift", err)
	}
}

func TestCertificationRejectsNarrowCredentialEvidence(t *testing.T) {
	_, err := newProofBearingEvidence(completedLiveEvidence{
		SchemaVersion:  certificationSchemaVersion,
		Scope:          evidenceScopeCapability,
		Connector:      "github",
		FunctionKind:   "operation:rest_read",
		Provider:       "github",
		ExecutedAt:     "2026-08-10T00:00:00Z",
		RunID:          "github-narrow-credential",
		PMBinarySHA256: strings.Repeat("c", 64),
		PMCommand:      "pm etl read --connector github --json",
		Passed:         true,
		RepositorySalt: []byte("0123456789abcdef0123456789abcdef"),
		PreparedValues: []string{"narrow-credential-token"},
		HTTPExchanges: []completedHTTPExchange{{
			Operation: "github.repos.get",
			Request:   completedHTTPRequest{Method: "GET", Target: "https://api.github.example/repos/acme/widget"},
			Response:  completedHTTPResponse{Status: 200},
		}},
	})
	if err == nil || !strings.Contains(err.Error(), "full-parity") {
		t.Fatalf("newProofBearingEvidence() error = %v, want full-parity rejection", err)
	}
}

func TestCertificationGeneratedArtifactDriftFails(t *testing.T) {
	path := filepath.Join(t.TempDir(), "capability-matrix.json")
	if err := os.WriteFile(path, []byte("{\"stale\":true}\n"), 0o600); err != nil {
		t.Fatalf("write stale artifact: %v", err)
	}
	if err := checkGeneratedArtifact(path, []byte("{\"current\":true}\n")); err == nil {
		t.Fatal("checkGeneratedArtifact() error = nil, want drift failure")
	}
}

func TestCertificationDiscoversStableWarehouseFacingSyncPrimitives(t *testing.T) {
	primitives, err := discoverSyncPrimitives([]matrixConnectorSource{{integrationType: "api"}, {integrationType: "database"}})
	if err != nil {
		t.Fatalf("discoverSyncPrimitives() error = %v", err)
	}
	want := map[string]bool{
		"api_read_into_warehouse":       false,
		"api_write_from_warehouse":      false,
		"database_read_into_warehouse":  false,
		"database_write_from_warehouse": false,
	}
	for _, primitive := range primitives {
		if _, ok := want[primitive.ID]; ok {
			want[primitive.ID] = true
		}
	}
	for primitive, found := range want {
		if !found {
			t.Errorf("discoverSyncPrimitives() omitted %q: %#v", primitive, primitives)
		}
	}
}

func TestCertificationSyncModeDatabaseWriteStubIsNotImplemented(t *testing.T) {
	matrix := certificationMatrixForTest(t)
	read, _ := capabilityCellFor(matrix, "postgres", "capability:read")
	write, _ := capabilityCellFor(matrix, "postgres", "capability:write")
	cdc, _ := capabilityCellFor(matrix, "postgres", "capability:cdc")
	cell := syncModeCellFor("postgres", "database", "incremental_dedupe", syncPrimitive{
		ID:                 "database_write_from_warehouse",
		IntegrationType:    "database",
		Capability:         "write",
		WarehouseDirection: "from_warehouse",
	}, read, write, cdc, false, nil)
	if !cell.Applicable {
		t.Fatal("PostgreSQL database write was hidden as non-applicable")
	}
	if cell.Implemented {
		t.Fatal("PostgreSQL ErrUnsupportedOperation write stub reported implemented for warehouse output")
	}
}

func TestCertificationChangeCaptureRequiresDatabaseReadIntoWarehouse(t *testing.T) {
	cell := syncModeCellFor("github", "api", "change_capture", syncPrimitive{
		ID:                 "api_read_into_warehouse",
		IntegrationType:    "api",
		Capability:         "read",
		WarehouseDirection: "into_warehouse",
	}, certificationCell{}, certificationCell{}, certificationCell{}, false, nil)
	if cell.Applicable || cell.NotApplicable == nil {
		t.Fatalf("API change-capture cell = %#v, want named non-applicable cell", cell)
	}
	if cell.NotApplicable.Code != "change_capture_requires_database_read" {
		t.Fatalf("change-capture reason = %#v, want database-read reason", cell.NotApplicable)
	}
	if err := validateSyncModeCertificationCell(cell); err != nil {
		t.Fatalf("validateSyncModeCertificationCell() error = %v", err)
	}
}

func TestCertificationWorkflowWithoutEvidencePreventsCompletion(t *testing.T) {
	cell := workflowCertificationCell{
		WorkflowKind:  "schedule",
		Applicable:    true,
		Declared:      true,
		Implemented:   true,
		FixtureTested: true,
		LiveTested:    false,
	}
	if workflowCellsComplete([]workflowCertificationCell{cell}) {
		t.Fatal("workflow without accepted live proof completed certification")
	}
}

func TestCertificationFlowPairAllowsGitHubToItselfThroughWarehouse(t *testing.T) {
	matrix := flowMatrix{PairSets: []flowPairSet{{
		FlowKind:              "api_to_api",
		Mediator:              localWarehouseMediator,
		SourceConnectors:      []string{"github"},
		DestinationConnectors: []string{"github"},
		Cell: flowCertificationCell{
			Applicable:      true,
			Declared:        true,
			Implemented:     false,
			FixtureEvidence: []string{},
			LiveEvidence:    []evidencePointer{},
		},
	}}}
	resolved, ok := resolveFlowPair(matrix, "api_to_api", "github", "github")
	if !ok {
		t.Fatal("GitHub -> warehouse -> GitHub did not resolve as an exact flow pair")
	}
	if resolved.Mediator != localWarehouseMediator || !resolved.Cell.Applicable {
		t.Fatalf("GitHub self pair = %#v, want applicable local-Parquet-mediated cell", resolved)
	}
}

func TestCertificationFlowEvidenceRequiresRoundTripProof(t *testing.T) {
	secret := "flow-evidence-token"
	evidence, err := newProofBearingEvidence(completedLiveEvidence{
		SchemaVersion:        certificationSchemaVersion,
		Scope:                evidenceScopeFlow,
		Source:               "github",
		Destination:          "github",
		FlowKind:             "api_to_api",
		Provider:             "github",
		ExecutedAt:           "2026-08-10T00:00:00Z",
		RunID:                "github-round-trip",
		PMBinarySHA256:       strings.Repeat("e", 64),
		PMCommand:            "pm flow run --file github-round-trip.json --json",
		Passed:               true,
		CredentialFullParity: true,
		RepositorySalt:       []byte("0123456789abcdef0123456789abcdef"),
		PreparedValues:       []string{secret},
		HTTPExchanges: []completedHTTPExchange{
			{Operation: "warehouse.readback", Request: completedHTTPRequest{Method: "GET", Target: "https://proof.example/warehouse"}, Response: completedHTTPResponse{Status: 200}},
			{Operation: "github.destination.readback", Request: completedHTTPRequest{Method: "GET", Target: "https://proof.example/github"}, Response: completedHTTPResponse{Status: 200}},
		},
		Flow: &completedFlowRoundTrip{
			PMCommand:                    "pm flow run --file github-round-trip.json --json",
			WarehouseReadbackOperation:   "warehouse.readback",
			DestinationReadbackOperation: "github.destination.readback",
			Delivery: deliveryGuarantees{Limitations: []deliveryLimitation{
				{Guarantee: "resumable", Code: "github_one_shot_no_workset", Reason: "one-shot mutation has no immutable workset"},
				{Guarantee: "receipt_backed", Code: "github_no_destination_receipt", Reason: "provider mutation supplies no destination receipt"},
				{Guarantee: "checkpointed", Code: "github_no_delivery_checkpoint", Reason: "provider mutation supplies no delivery checkpoint"},
				{Guarantee: "replay_identity", Code: "github_no_replay_identity", Reason: "provider mutation supplies no replay identity"},
				{Guarantee: "provider_idempotency_key", Code: "github_no_idempotency_key", Reason: "action declares no provider idempotency key"},
			}},
		},
	})
	if err != nil {
		t.Fatalf("newProofBearingEvidence(flow) error = %v", err)
	}
	if evidence.Proof.Flow == nil {
		t.Fatal("flow evidence did not embed a round-trip proof")
	}
	if err := validateAcceptedEvidence(evidence); err != nil {
		t.Fatalf("validateAcceptedEvidence(flow) error = %v", err)
	}
	evidence.Proof.Flow = nil
	if err := validateAcceptedEvidence(evidence); err == nil {
		t.Fatal("flow evidence without round-trip proof was accepted")
	}
}

func repoRootForCertificationTest(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolve repository root: %v", err)
	}
	return root
}

func certificationMatrixForTest(t *testing.T) capabilityMatrix {
	t.Helper()
	certificationMatrixTestCache.once.Do(func() {
		certificationMatrixTestCache.matrix, certificationMatrixTestCache.err = buildCapabilityMatrix(repoRootForCertificationTest(t))
	})
	if certificationMatrixTestCache.err != nil {
		t.Fatalf("buildCapabilityMatrix() error = %v", certificationMatrixTestCache.err)
	}
	return certificationMatrixTestCache.matrix
}
