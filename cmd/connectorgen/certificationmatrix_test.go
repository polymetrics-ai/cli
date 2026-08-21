package main

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"

	"polymetrics.ai/internal/connectors/certify"
	"polymetrics.ai/internal/connectors/engine"
	"polymetrics.ai/internal/synccontract"
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

	for _, connectorName := range []string{"mysql"} {
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

func TestCertificationMatrixRetainsPostgresManagedDestinationEvidenceAtExactModeScope(t *testing.T) {
	matrix := certificationMatrixForTest(t)
	write, ok := capabilityCellFor(matrix, "postgres", "capability:write")
	if !ok {
		t.Fatal("postgres write cell missing")
	}
	if !write.Applicable || write.Declared || write.Implemented || write.LiveTested || len(write.LiveEvidence) != 0 {
		t.Fatalf("PostgreSQL generic write cell = %#v, want applicable but unpublished and unimplemented", write)
	}
	repoRoot := repoRootForCertificationTest(t)
	bundles, err := loadSourceBundlesForConnectors(repoRoot, []string{"postgres"})
	if err != nil {
		t.Fatalf("loadSourceBundlesForConnectors() error = %v", err)
	}
	sources, err := matrixConnectorSourcesForNames(bundles, []string{"postgres"})
	if err != nil || len(sources) != 1 {
		t.Fatalf("matrixConnectorSourcesForNames() = %#v, %v; want PostgreSQL source", sources, err)
	}
	evidence, err := loadAcceptedEvidence(repoRoot, []string{"postgres"})
	if err != nil {
		t.Fatalf("loadAcceptedEvidence() error = %v", err)
	}
	cells, err := buildConnectorSyncModeCells(sources, matrix, []syncModeKind{
		{ID: string(synccontract.ModeFullOverwrite)},
		{ID: string(synccontract.ModeFullAppend)},
		{ID: string(synccontract.ModeIncrementalAppend)},
		{ID: string(synccontract.ModeIncrementalUpsert)},
		{ID: string(synccontract.ModeIncrementalDedupe)},
		{ID: string(synccontract.ModeIncrementalDedupeHistory)},
	}, []syncPrimitive{{
		ID:                 "database_write_from_warehouse",
		IntegrationType:    "database",
		Capability:         "write",
		WarehouseDirection: "from_warehouse",
	}}, evidence)
	if err != nil || len(cells) != 1 || len(cells[0].Cells) != 6 {
		t.Fatalf("buildConnectorSyncModeCells() = %#v, %v; want six PostgreSQL managed destination cells", cells, err)
	}
	for _, cell := range cells[0].Cells {
		if !cell.Applicable || !cell.Declared || !cell.Implemented || !cell.LiveTested || len(cell.LiveEvidence) != 1 {
			t.Fatalf("PostgreSQL managed destination cell = %#v, want exact declared live mode proof", cell)
		}
	}
}

// The PostgreSQL managed target is a closed destination transport. Its
// complete six-mode profile proves that exact transport, not the generic
// Connector.Write capability exposed by metadata.capabilities.write.
func TestCertificationMatrixDoesNotTreatPostgresManagedTransportAsGenericWrite(t *testing.T) {
	repoRoot := repoRootForCertificationTest(t)
	bundles, err := loadSourceBundlesForConnectors(repoRoot, []string{"postgres"})
	if err != nil {
		t.Fatalf("loadSourceBundlesForConnectors() error = %v", err)
	}
	sources, err := matrixConnectorSourcesForNames(bundles, []string{"postgres"})
	if err != nil || len(sources) != 1 {
		t.Fatalf("matrixConnectorSourcesForNames() = %#v, %v; want PostgreSQL source", sources, err)
	}
	// This is the publication shape required for PostgreSQL: its exact
	// destination transport remains declared, but generic Connector.Write is
	// not published. The direct Write method is intentionally unsupported.
	bundle := *sources[0].bundle
	bundle.Metadata.Capabilities.Write = false
	source := sources[0]
	source.bundle = &bundle
	implemented, err := capabilityImplemented(repoRoot, source, "write")
	if err != nil {
		t.Fatalf("capabilityImplemented() error = %v", err)
	}
	if implemented {
		t.Fatal("PostgreSQL managed destination evidence promoted generic write capability")
	}
}

func TestCertificationMatrixPromotesPostgresChangeCaptureOnlyWithReceiptBackedLiveProof(t *testing.T) {
	repoRoot := repoRootForCertificationTest(t)
	matrix := certificationMatrixForTest(t)
	cdc, ok := capabilityCellFor(matrix, "postgres", "capability:cdc")
	if !ok {
		t.Fatal("postgres CDC cell missing")
	}
	if !cdc.Applicable || !cdc.Declared || !cdc.Implemented || !cdc.LiveTested || len(cdc.LiveEvidence) != 1 {
		t.Fatalf("PostgreSQL CDC cell = %#v, want one declared, implemented, live receipt-backed proof", cdc)
	}
	bundles, err := loadSourceBundlesForConnectors(repoRoot, []string{"postgres"})
	if err != nil {
		t.Fatalf("loadSourceBundlesForConnectors() error = %v", err)
	}
	sources, err := matrixConnectorSourcesForNames(bundles, []string{"postgres"})
	if err != nil || len(sources) != 1 {
		t.Fatalf("matrixConnectorSourcesForNames() = %#v, %v; want PostgreSQL source", sources, err)
	}
	evidence, err := loadAcceptedEvidence(repoRoot, []string{"postgres"})
	if err != nil {
		t.Fatalf("loadAcceptedEvidence() error = %v", err)
	}
	cells, err := buildConnectorSyncModeCells(sources, matrix, []syncModeKind{{ID: string(synccontract.ModeChangeCapture)}}, []syncPrimitive{{
		ID:                 "database_read_into_warehouse",
		IntegrationType:    "database",
		Capability:         "read",
		WarehouseDirection: "into_warehouse",
	}}, evidence)
	if err != nil || len(cells) != 1 || len(cells[0].Cells) != 1 {
		t.Fatalf("buildConnectorSyncModeCells() = %#v, %v; want one PostgreSQL change-capture cell", cells, err)
	}
	cell := cells[0].Cells[0]
	if !cell.Applicable || !cell.Declared || !cell.Implemented || !cell.LiveTested || len(cell.LiveEvidence) != 1 {
		t.Fatalf("PostgreSQL change-capture cell = %#v, want one declared, implemented, live receipt-backed proof", cell)
	}
}

func TestPostgresPublishesOnlyGenericCapabilitiesWithMatchingLiveCertification(t *testing.T) {
	matrix := certificationMatrixForTest(t)
	bundles, err := loadSourceBundlesForConnectors(repoRootForCertificationTest(t), []string{"postgres"})
	if err != nil || len(bundles) != 1 {
		t.Fatalf("loadSourceBundlesForConnectors() = %#v, %v; want PostgreSQL bundle", bundles, err)
	}
	capabilities := bundles[0].Metadata.Capabilities
	if capabilities.Write {
		t.Fatalf("PostgreSQL write publication = true, want false because the managed target is a closed transport rather than generic Connector.Write")
	}
	write, ok := capabilityCellFor(matrix, "postgres", "capability:write")
	if !ok || !write.Applicable || write.Declared || write.Implemented || write.LiveTested || len(write.LiveEvidence) != 0 {
		t.Fatalf("PostgreSQL generic write cell = %#v, want applicable but unpublished and unimplemented", write)
	}
	published, found := boolFieldForKind(capabilities, "cdc")
	if !found || !published {
		t.Fatalf("PostgreSQL CDC publication = %t, want true with current evidence", published)
	}
	cdc, ok := capabilityCellFor(matrix, "postgres", "capability:cdc")
	if !ok || !cdc.Declared || !cdc.Implemented || !cdc.LiveTested || len(cdc.LiveEvidence) == 0 {
		t.Fatalf("PostgreSQL published CDC cell = %#v, want declaration, implementation, and accepted live evidence", cdc)
	}
	if capabilities.Query {
		t.Fatalf("PostgreSQL query publication = true, want concrete false while no query route is certified")
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

func TestCertificationMatrixDistinguishesGitHubGraphQLAndLocalGitExecutors(t *testing.T) {
	matrix := certificationMatrixForTest(t)

	for _, tc := range []struct {
		kind            string
		wantImplemented bool
	}{
		{kind: "operation:graphql_query", wantImplemented: true},
		{kind: "operation:local_git", wantImplemented: false},
	} {
		cell, ok := capabilityCellFor(matrix, "github", tc.kind)
		if !ok {
			t.Fatalf("github %s cell missing", tc.kind)
		}
		if !cell.Applicable || !cell.Declared {
			t.Fatalf("github %s cell = %+v, want applicable declared operation", tc.kind, cell)
		}
		if cell.Implemented != tc.wantImplemented {
			t.Errorf("github %s implemented = %t, want %t", tc.kind, cell.Implemented, tc.wantImplemented)
		}
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

func TestCertificationGitHubRestReadWithoutAcceptedEvidenceStaysNotLiveTested(t *testing.T) {
	repoRoot := repoRootForCertificationTest(t)
	bundles, err := loadSourceBundlesForConnectors(repoRoot, []string{"github"})
	if err != nil || len(bundles) != 1 {
		t.Fatalf("loadSourceBundlesForConnectors() = %#v, %v; want GitHub bundle", bundles, err)
	}
	sources, err := matrixConnectorSourcesForNames(bundles, []string{"github"})
	if err != nil || len(sources) != 1 {
		t.Fatalf("matrixConnectorSourcesForNames() = %#v, %v; want GitHub source", sources, err)
	}
	kinds, err := discoverFunctionKinds(repoRoot)
	if err != nil {
		t.Fatalf("discoverFunctionKinds() = %v", err)
	}
	for _, kind := range kinds {
		if kind.ID != "operation:rest_read" {
			continue
		}
		cell, err := buildCertificationCell(repoRoot, sources[0], kind, nil)
		if err != nil {
			t.Fatalf("buildCertificationCell() = %v", err)
		}
		if cell.LiveTested || len(cell.LiveEvidence) != 0 {
			t.Fatalf("GitHub rest-read cell without accepted evidence = %#v, want not live-tested", cell)
		}
		return
	}
	t.Fatal("operation:rest_read function kind missing")
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
		SchemaVersion:        acceptedEvidenceSchemaVersion,
		Scope:                evidenceScopeCapability,
		Status:               evidenceStatusPassed,
		CredentialScope:      credentialScopeObservedOperations,
		CredentialNote:       observedOperationsCredentialNote,
		CredentialScopeProof: credentialScopeProofProtocolExchanges,
		Connector:            "github",
		FunctionKind:         "capability:read",
	})
	if err == nil {
		t.Fatal("validateAcceptedEvidence() error = nil, want missing real-provider evidence rejection")
	}
	if !strings.Contains(err.Error(), "provider") {
		t.Fatalf("validateAcceptedEvidence() error = %q, want provider context", err)
	}
}

func TestCertificationRejectsUnsafeEvidenceIdentifiers(t *testing.T) {
	err := validateAcceptedEvidence(acceptedEvidence{
		SchemaVersion:        acceptedEvidenceSchemaVersion,
		Scope:                evidenceScopeCapability,
		Status:               evidenceStatusPassed,
		CredentialScope:      credentialScopeObservedOperations,
		CredentialNote:       observedOperationsCredentialNote,
		CredentialScopeProof: credentialScopeProofProtocolExchanges,
		Connector:            "github",
		FunctionKind:         "capability:read",
		Provider:             "github credential",
	})
	if err == nil {
		t.Fatal("validateAcceptedEvidence() error = nil, want unsafe provider rejection")
	}
	if !strings.Contains(err.Error(), "provider") {
		t.Fatalf("validateAcceptedEvidence() error = %q, want provider context", err)
	}

	root := t.TempDir()
	_, err = acceptedEvidenceOutputPath(root, filepath.Join(root, "internal", "connectors", "certifications", "evidence", "credential value.json"))
	if err == nil {
		t.Fatal("acceptedEvidenceOutputPath() error = nil, want unsafe record-name rejection")
	}
}

func TestCertificationRejectsProoflessAcceptedLiveEvidence(t *testing.T) {
	err := validateAcceptedEvidence(acceptedEvidence{
		SchemaVersion:        acceptedEvidenceSchemaVersion,
		Scope:                evidenceScopeCapability,
		Status:               evidenceStatusPassed,
		Connector:            "github",
		FunctionKind:         "capability:read",
		Provider:             "github",
		ExecutedAt:           "2026-08-10T00:00:00Z",
		RunID:                "live-run-123",
		CredentialScope:      credentialScopeObservedOperations,
		CredentialNote:       observedOperationsCredentialNote,
		CredentialScopeProof: credentialScopeProofProtocolExchanges,
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
		SchemaVersion:  certificationSchemaVersion,
		Scope:          evidenceScopeCapability,
		Connector:      "github",
		FunctionKind:   "operation:rest_read",
		Provider:       "github",
		ExecutedAt:     "2026-08-10T00:00:00Z",
		RunID:          "live-run-123",
		PMBinarySHA256: strings.Repeat("a", 64),
		PMCommand:      "pm etl read --connector github --json",
		Passed:         true,
		RepositorySalt: []byte("0123456789abcdef0123456789abcdef"),
		PreparedValues: []string{secret},
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
	subject := writeTestCurrentCertificationSubject(t, root, strings.Repeat("d", 64))
	path := filepath.Join(root, "internal", "connectors", "certifications", "evidence", "github.json")
	secret := "credential-only-in-memory"
	completed := completedLiveEvidence{
		SchemaVersion:  certificationSchemaVersion,
		Scope:          evidenceScopeCapability,
		Connector:      "github",
		FunctionKind:   "operation:rest_read",
		Provider:       "github",
		ExecutedAt:     "2026-08-10T00:00:00Z",
		RunID:          "github-local-salt",
		PMBinarySHA256: strings.Repeat("d", 64),
		PMCommand:      "pm etl read --connector github --json",
		Passed:         true,
		PreparedValues: []string{secret},
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
	if !certificationSubjectsEqual(evidence.Proof.CertificationSubject, subject) {
		t.Fatalf("writer subject = %#v, want exact current subject", evidence.Proof.CertificationSubject)
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
		SchemaVersion:  certificationSchemaVersion,
		Scope:          evidenceScopeCapability,
		Connector:      "github",
		FunctionKind:   "operation:rest_read",
		Provider:       "github",
		ExecutedAt:     "2026-08-10T00:00:00Z",
		RunID:          "live-run-unsafe-proof",
		PMBinarySHA256: strings.Repeat("b", 64),
		PMCommand:      "pm etl read --connector github --json",
		Passed:         true,
		RepositorySalt: []byte("0123456789abcdef0123456789abcdef"),
		PreparedValues: []string{secret},
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

func TestCertificationPublishesNarrowCredentialEvidence(t *testing.T) {
	evidence, err := newProofBearingEvidence(completedLiveEvidence{
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
	if err != nil {
		t.Fatalf("newProofBearingEvidence() = %v, want bounded evidence publication", err)
	}
	if evidence.CredentialScope != credentialScopeObservedOperations || evidence.CredentialScopeProof != credentialScopeProofProtocolExchanges {
		t.Fatalf("bounded credential claim = (%q, %q), want (%q, %q)", evidence.CredentialScope, evidence.CredentialScopeProof, credentialScopeObservedOperations, credentialScopeProofProtocolExchanges)
	}
}

func TestCertificationBoundedScopePublishesObservedOperations(t *testing.T) {
	evidence, err := newProofBearingEvidence(completedLiveEvidence{
		SchemaVersion:  certificationSchemaVersion,
		Scope:          evidenceScopeCapability,
		Connector:      "test_connector",
		FunctionKind:   "operation:rest_read",
		Provider:       "test_provider",
		ExecutedAt:     "2026-08-10T00:00:00Z",
		RunID:          "bounded-credential-run",
		PMBinarySHA256: strings.Repeat("f", 64),
		PMCommand:      "pm test_connector records list --json",
		Passed:         true,
		RepositorySalt: []byte("0123456789abcdef0123456789abcdef"),
		PreparedValues: []string{"bounded-credential-token"},
		HTTPExchanges: []completedHTTPExchange{{
			Operation: "records.list",
			Request:   completedHTTPRequest{Method: "GET", Target: "https://api.example.test/records"},
			Response:  completedHTTPResponse{Status: 200},
		}},
	})
	if err != nil {
		t.Fatalf("newProofBearingEvidence() = %v, want bounded evidence publication", err)
	}
	if evidence.CredentialScope != credentialScopeObservedOperations {
		t.Fatalf("credential_scope = %q, want %q", evidence.CredentialScope, credentialScopeObservedOperations)
	}
	if evidence.CredentialScopeProof != credentialScopeProofProtocolExchanges {
		t.Fatalf("credential_scope_proof = %q, want %q", evidence.CredentialScopeProof, credentialScopeProofProtocolExchanges)
	}
	if err := validateAcceptedEvidence(evidence); err != nil {
		t.Fatalf("validateAcceptedEvidence(bounded) = %v", err)
	}
}

func TestCertificationFullParityScopeRequiresPassedReportStage(t *testing.T) {
	completed := completedLiveEvidence{
		SchemaVersion:  certificationSchemaVersion,
		Scope:          evidenceScopeCapability,
		Connector:      "test_connector",
		FunctionKind:   "operation:rest_read",
		Provider:       "test_provider",
		ExecutedAt:     "2026-08-10T00:00:00Z",
		RunID:          "unverified-full-parity-run",
		PMBinarySHA256: strings.Repeat("e", 64),
		PMCommand:      "pm test_connector records list --json",
		Passed:         true,
		RepositorySalt: []byte("0123456789abcdef0123456789abcdef"),
		PreparedValues: []string{"full-parity-credential-token"},
		HTTPExchanges: []completedHTTPExchange{{
			Operation: "records.list",
			Request:   completedHTTPRequest{Method: "GET", Target: "https://api.example.test/records"},
			Response:  completedHTTPResponse{Status: 200},
		}},
	}

	// This is a passed direct-read-like run. It has no full-parity stage, so
	// its evidence must fail before a full-parity record can be constructed.
	_, err := newFullParityProofBearingEvidence(completed, certify.Report{
		Stages: []certify.StageResult{{Name: "direct_read_sweep", Passed: true}},
	})
	if err == nil || !strings.Contains(err.Error(), "full-parity stage") {
		t.Fatalf("newFullParityProofBearingEvidence(unverified report) error = %v, want full-parity stage refusal", err)
	}

	evidence, err := newFullParityProofBearingEvidence(completed, certify.Report{
		Stages: []certify.StageResult{{Name: "full_parity", Passed: true}},
	})
	if err != nil {
		t.Fatalf("newFullParityProofBearingEvidence(verified report) = %v", err)
	}
	if evidence.CredentialScope != credentialScopeFullParity || evidence.CredentialScopeProof != credentialScopeProofFullParityStage {
		t.Fatalf("verified full-parity credential claim = (%q, %q), want (%q, %q)", evidence.CredentialScope, evidence.CredentialScopeProof, credentialScopeFullParity, credentialScopeProofFullParityStage)
	}

	evidence.CredentialScopeProof = credentialScopeProofProtocolExchanges
	if err := validateAcceptedEvidence(evidence); err == nil || !strings.Contains(err.Error(), "scope proof") {
		t.Fatalf("validateAcceptedEvidence(mismatched full-parity proof) error = %v, want scope proof rejection", err)
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

func TestCertificationShardsRoundTripGeneratedMatrices(t *testing.T) {
	root := repoRootForCertificationTest(t)
	shards, err := buildCertificationShards(root, certificationConnectorAllowlist)
	if err != nil {
		t.Fatalf("buildCertificationShards() error = %v", err)
	}
	if len(shards) != len(certificationConnectorAllowlist) {
		t.Fatalf("buildCertificationShards() produced %d shards, want %d", len(shards), len(certificationConnectorAllowlist))
	}
	capabilities, flows, err := reconstructCertificationMatrices(shards)
	if err != nil {
		t.Fatalf("reconstructCertificationMatrices() error = %v", err)
	}
	if len(capabilities.Connectors) != len(certificationConnectorAllowlist) || len(flows.ConnectorRoles) != len(certificationConnectorAllowlist) {
		t.Fatalf("reconstructed scope = capabilities=%d roles=%d, want %d", len(capabilities.Connectors), len(flows.ConnectorRoles), len(certificationConnectorAllowlist))
	}
	for _, connector := range certificationConnectorAllowlist {
		if _, found := capabilityConnectorByName(capabilities, connector); !found {
			t.Errorf("reconstructed capability matrix omitted %q", connector)
		}
	}
	wantCapabilities, err := buildCapabilityMatrixForConnectors(root, certificationConnectorAllowlist)
	if err != nil {
		t.Fatalf("buildCapabilityMatrixForConnectors() error = %v", err)
	}
	wantFlows, err := buildFlowMatrixForConnectors(root, wantCapabilities, certificationConnectorAllowlist)
	if err != nil {
		t.Fatalf("buildFlowMatrixForConnectors() error = %v", err)
	}
	// Legacy harness metadata is deliberately not certification data and never
	// belongs inside a connector shard. The rest must reconstruct precisely.
	wantCapabilities.LegacyCertificationInputs = capabilities.LegacyCertificationInputs
	wantCapabilities.GeneratedCommand = capabilities.GeneratedCommand
	wantFlows.GeneratedCommand = flows.GeneratedCommand
	if !reflect.DeepEqual(capabilities, wantCapabilities) {
		t.Fatal("shard union did not reconstruct the GitHub/PostgreSQL capability aggregate")
	}
	if !reflect.DeepEqual(flows, wantFlows) {
		t.Fatal("shard union did not reconstruct the GitHub/PostgreSQL flow aggregate")
	}
}

func TestCertificationScopedGenerationLeavesOtherShardByteIdentical(t *testing.T) {
	sourceRoot := repoRootForCertificationTest(t)
	shards, err := buildCertificationShards(sourceRoot, certificationConnectorAllowlist)
	if err != nil {
		t.Fatalf("buildCertificationShards() error = %v", err)
	}
	payloads := make(map[string][]byte, len(shards))
	for connector, shard := range shards {
		payloads[connector], err = marshalGeneratedJSON(shard)
		if err != nil {
			t.Fatalf("marshal %q shard: %v", connector, err)
		}
	}

	outputRoot := t.TempDir()
	for connector, payload := range payloads {
		path := certificationShardPath(outputRoot, connector)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir shard parent: %v", err)
		}
		if err := os.WriteFile(path, payload, 0o600); err != nil {
			t.Fatalf("write initial %q shard: %v", connector, err)
		}
	}
	other := "postgres"
	before, err := os.ReadFile(certificationShardPath(outputRoot, other))
	if err != nil {
		t.Fatalf("read initial %q shard: %v", other, err)
	}
	statusPath := filepath.Join(outputRoot, certificationStatusPath)
	statusBefore := []byte("unchanged shared status projection\n")
	if err := os.MkdirAll(filepath.Dir(statusPath), 0o755); err != nil {
		t.Fatalf("mkdir status parent: %v", err)
	}
	if err := os.WriteFile(statusPath, statusBefore, 0o600); err != nil {
		t.Fatalf("write initial status: %v", err)
	}
	if _, err := generateCertificationMatrix(sourceRoot, outputRoot, false, false, "github"); err != nil {
		t.Fatalf("generateCertificationMatrix() error = %v", err)
	}
	after, err := os.ReadFile(certificationShardPath(outputRoot, other))
	if err != nil {
		t.Fatalf("read final %q shard: %v", other, err)
	}
	if !bytes.Equal(before, after) {
		t.Fatalf("scoped github generation rewrote %q shard", other)
	}
	statusAfter, err := os.ReadFile(statusPath)
	if err != nil {
		t.Fatalf("read final status: %v", err)
	}
	if !bytes.Equal(statusBefore, statusAfter) {
		t.Fatal("scoped github generation rewrote the shared status projection")
	}
}

func TestCertificationCheckValidatesCommittedStatusBeforeSourceReconstruction(t *testing.T) {
	repoRoot := repoRootForCertificationTest(t)
	shards, err := buildCertificationShards(repoRoot, certificationConnectorAllowlist)
	if err != nil {
		t.Fatalf("buildCertificationShards() error = %v", err)
	}
	payloads, err := marshalCertificationShards(shards)
	if err != nil {
		t.Fatalf("marshalCertificationShards() error = %v", err)
	}
	outputRoot := t.TempDir()
	if err := writeCertificationShardScope(outputRoot, payloads, certificationConnectorAllowlist); err != nil {
		t.Fatalf("writeCertificationShardScope() error = %v", err)
	}
	statusPath := filepath.Join(outputRoot, certificationStatusPath)
	if err := os.MkdirAll(filepath.Dir(statusPath), 0o755); err != nil {
		t.Fatalf("mkdir status parent: %v", err)
	}
	if err := os.WriteFile(statusPath, []byte("{}\n"), 0o600); err != nil {
		t.Fatalf("write malformed status: %v", err)
	}
	_, err = generateCertificationMatrix(t.TempDir(), outputRoot, true, false, "")
	if err == nil || !strings.Contains(err.Error(), "generated certification artifact") {
		t.Fatalf("generateCertificationMatrix() error = %v, want committed status validation", err)
	}
}

func TestCertificationScopedCheckDoesNotReadGlobalStatusOrOtherConnectorShards(t *testing.T) {
	sourceRoot := repoRootForCertificationTest(t)
	shards, err := buildCertificationShards(sourceRoot, []string{"github"})
	if err != nil {
		t.Fatalf("buildCertificationShards() error = %v", err)
	}
	payloads, err := marshalCertificationShards(shards)
	if err != nil {
		t.Fatalf("marshalCertificationShards() error = %v", err)
	}
	outputRoot := t.TempDir()
	if err := writeCertificationShardScope(outputRoot, payloads, []string{"github"}); err != nil {
		t.Fatalf("writeCertificationShardScope() error = %v", err)
	}
	if _, err := generateCertificationMatrix(sourceRoot, outputRoot, true, false, "github"); err != nil {
		t.Fatalf("scoped matrix check error = %v, want it to validate only GitHub's generated shard", err)
	}
}

func TestCertificationShardDriftFails(t *testing.T) {
	root := repoRootForCertificationTest(t)
	shards, err := buildCertificationShards(root, certificationConnectorAllowlist)
	if err != nil {
		t.Fatalf("buildCertificationShards() error = %v", err)
	}
	payloads, err := marshalCertificationShards(shards)
	if err != nil {
		t.Fatalf("marshalCertificationShards() error = %v", err)
	}
	outputRoot := t.TempDir()
	if err := writeCertificationShardScope(outputRoot, payloads, certificationConnectorAllowlist); err != nil {
		t.Fatalf("writeCertificationShardScope() error = %v", err)
	}
	path := certificationShardPath(outputRoot, "github")
	stale := append([]byte(nil), payloads["github"]...)
	stale = append([]byte(" "), stale...)
	if err := os.WriteFile(path, stale, 0o600); err != nil {
		t.Fatalf("write stale shard: %v", err)
	}
	if err := checkCertificationShards(outputRoot, payloads); err == nil {
		t.Fatal("checkCertificationShards() error = nil, want shard drift failure")
	}
}

func TestCertificationScopeRetainsSourceOwnedCrossPairClaims(t *testing.T) {
	scope := []string{"github"}
	for _, evidence := range []acceptedEvidence{
		{Scope: evidenceScopeCapability, Connector: "mysql"},
		{Scope: evidenceScopeWorkflow, Connector: "postgres"},
		{Scope: evidenceScopeFlow, Source: "github", Destination: "mysql"},
	} {
		if acceptedEvidenceWithinScope(evidence, scope) {
			t.Fatalf("evidence %#v unexpectedly entered github certification scope", evidence)
		}
	}
	if !acceptedEvidenceWithinScope(acceptedEvidence{Scope: evidenceScopeCapability, Connector: "github"}, scope) {
		t.Fatal("github capability evidence did not enter github certification scope")
	}
	if !acceptedEvidenceWithinScope(acceptedEvidence{Scope: evidenceScopeFlow, Source: "github", Destination: "postgres"}, scope) {
		t.Fatal("github source-owned cross-pair evidence did not enter github certification scope")
	}
}

func TestCertificationEvidenceScopeFiltersBeforeStrictValidation(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, acceptedEvidenceDirectory)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir evidence directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "mysql.json"), []byte(`{"scope":"capability","connector":"mysql","unexpected":true}`), 0o600); err != nil {
		t.Fatalf("write mysql evidence: %v", err)
	}
	items, err := loadAcceptedEvidence(root, certificationConnectorAllowlist)
	if err != nil {
		t.Fatalf("loadAcceptedEvidence() with malformed nonallowlisted evidence: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("loadAcceptedEvidence() returned %#v, want no evidence", items)
	}
	if err := os.WriteFile(filepath.Join(dir, "github.json"), []byte(`{"scope":"capability","connector":"github","unexpected":true}`), 0o600); err != nil {
		t.Fatalf("write github evidence: %v", err)
	}
	if _, err := loadAcceptedEvidence(root, certificationConnectorAllowlist); err == nil || !strings.Contains(err.Error(), "unknown field") {
		t.Fatalf("loadAcceptedEvidence() with malformed allowlisted evidence = %v, want strict decoder rejection", err)
	}
}

func TestCertificationEvidencePrefilterSkipsOnlyConclusiveNonallowlisted(t *testing.T) {
	for _, test := range []struct {
		name     string
		identity acceptedEvidenceScopeIdentity
		wantSkip bool
	}{
		{
			name:     "nonallowlisted capability",
			identity: acceptedEvidenceScopeIdentity{Scope: evidenceScopeCapability, Connector: "mysql"},
			wantSkip: true,
		},
		{
			name:     "nonallowlisted flow",
			identity: acceptedEvidenceScopeIdentity{Scope: evidenceScopeFlow, Source: "mysql", Destination: "mariadb"},
			wantSkip: true,
		},
		{
			name:     "unsupported scope",
			identity: acceptedEvidenceScopeIdentity{Scope: "unsupported", Connector: "mysql"},
		},
		{
			name:     "missing capability connector",
			identity: acceptedEvidenceScopeIdentity{Scope: evidenceScopeCapability},
		},
		{
			name:     "allowlisted connector",
			identity: acceptedEvidenceScopeIdentity{Scope: evidenceScopeCapability, Connector: "github"},
		},
		{
			name:     "allowlisted flow destination",
			identity: acceptedEvidenceScopeIdentity{Scope: evidenceScopeFlow, Source: "mysql", Destination: "postgres"},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := acceptedEvidenceScopeIdentityIsConclusiveNonallowlisted(test.identity); got != test.wantSkip {
				t.Fatalf("acceptedEvidenceScopeIdentityIsConclusiveNonallowlisted(%#v) = %t, want %t", test.identity, got, test.wantSkip)
			}
		})
	}
}

func TestCertificationEvidencePrefilterRejectsDuplicateIdentityKeys(t *testing.T) {
	for _, test := range []struct {
		name     string
		field    string
		evidence string
	}{
		{
			name:     "scope",
			field:    "scope",
			evidence: `{"scope":"unsupported","scope":"capability","connector":"mysql","unexpected":true}`,
		},
		{
			name:     "connector",
			field:    "connector",
			evidence: `{"scope":"capability","connector":"github","connector":"mysql","unexpected":true}`,
		},
		{
			name:     "source",
			field:    "source",
			evidence: `{"scope":"flow","source":"github","source":"mysql","destination":"mariadb","unexpected":true}`,
		},
		{
			name:     "destination",
			field:    "destination",
			evidence: `{"scope":"flow","source":"mysql","destination":"github","destination":"mariadb","unexpected":true}`,
		},
		{
			name:     "case variant scope",
			field:    "scope",
			evidence: `{"Scope":"unsupported","scope":"capability","connector":"mysql","unexpected":true}`,
		},
		{
			name:     "case variant connector",
			field:    "connector",
			evidence: `{"scope":"capability","Connector":"github","connector":"mysql","unexpected":true}`,
		},
		{
			name:     "case variant source",
			field:    "source",
			evidence: `{"scope":"flow","Source":"github","source":"mysql","destination":"mariadb","unexpected":true}`,
		},
		{
			name:     "case variant destination",
			field:    "destination",
			evidence: `{"scope":"flow","source":"mysql","Destination":"github","destination":"mariadb","unexpected":true}`,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			dir := filepath.Join(root, acceptedEvidenceDirectory)
			if err := os.MkdirAll(dir, 0o755); err != nil {
				t.Fatalf("mkdir evidence directory: %v", err)
			}
			if err := os.WriteFile(filepath.Join(dir, "evidence.json"), []byte(test.evidence), 0o600); err != nil {
				t.Fatalf("write evidence: %v", err)
			}
			_, err := loadAcceptedEvidence(root, certificationConnectorAllowlist)
			if err == nil || !strings.Contains(err.Error(), `duplicate accepted evidence identity field "`+test.field+`"`) {
				t.Fatalf("loadAcceptedEvidence() with duplicate %s identity = %v, want duplicate identity rejection", test.field, err)
			}
		})
	}
}

func TestCertificationEvidenceUnsupportedAllowlistedScopeReachesStrictValidation(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, acceptedEvidenceDirectory)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir evidence directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "github.json"), []byte(`{"scope":"unsupported","connector":"github","unexpected":true}`), 0o600); err != nil {
		t.Fatalf("write github evidence: %v", err)
	}
	if _, err := loadAcceptedEvidence(root, certificationConnectorAllowlist); err == nil || !strings.Contains(err.Error(), "unknown field") {
		t.Fatalf("loadAcceptedEvidence() with malformed unsupported allowlisted evidence = %v, want strict decoder rejection", err)
	}
}

func TestCertificationSourceAnchorsUseSymbols(t *testing.T) {
	shards, err := buildCertificationShards(repoRootForCertificationTest(t), certificationConnectorAllowlist)
	if err != nil {
		t.Fatalf("buildCertificationShards() error = %v", err)
	}
	for connector, shard := range shards {
		if err := validateCertificationShard(shard); err != nil {
			t.Fatalf("validate %q shard: %v", connector, err)
		}
		for _, kind := range shard.FunctionKinds {
			for _, anchor := range []string{kind.DiscoverySource, kind.ExecutorSource} {
				if anchor == "" {
					continue
				}
				if err := validateSymbolSourceAnchor(anchor); err != nil {
					t.Errorf("%q has invalid source anchor %q: %v", connector, anchor, err)
				}
			}
		}
	}
	if err := validateSymbolSourceAnchor("internal/cli/cli.go:603"); err == nil {
		t.Fatal("numeric source anchor was accepted")
	}
	if err := validateSymbolSourceAnchor("internal/cli/cli.go:603notasymbol"); err == nil {
		t.Fatal("non-symbol source anchor was accepted")
	}
}

func TestCertificationScopedSourceResolutionUsesTargetConnector(t *testing.T) {
	bundles, err := loadSourceBundlesForConnectors(repoRootForCertificationTest(t), []string{"github"})
	if err != nil {
		t.Fatalf("loadSourceBundlesForConnectors() error = %v", err)
	}
	connector, found := scopedMatrixConnector("github", &bundles[0])
	if !found || !isEngineConnector(connector) {
		t.Fatalf("scopedMatrixConnector(github) = %T, %t; want engine connector", connector, found)
	}
}

func TestCertificationScopedSourceResolutionUsesScopedPostgresBundle(t *testing.T) {
	bundles, err := loadSourceBundlesForConnectors(repoRootForCertificationTest(t), []string{"postgres"})
	if err != nil {
		t.Fatalf("loadSourceBundlesForConnectors() error = %v", err)
	}
	connector, found := scopedMatrixConnector("postgres", &bundles[0])
	if !found {
		t.Fatal("scopedMatrixConnector(postgres) did not find the connector")
	}
	postgresConnector, ok := connector.(scopedPostgresMatrixConnector)
	if !ok {
		t.Fatalf("scopedMatrixConnector(postgres) = %T, want scopedPostgresMatrixConnector", connector)
	}
	method, found := capabilityMethod(postgresConnector, "write")
	if !found || method != "Write" {
		t.Fatalf("scopedMatrixConnector(postgres) write method = %q, %t; want Write, true", method, found)
	}
	unsupported, err := methodDirectlyReturnsUnsupported(repoRootForCertificationTest(t), postgresConnector, method)
	if err != nil {
		t.Fatalf("inspect scoped PostgreSQL Write method: %v", err)
	}
	if !unsupported {
		t.Fatal("scoped PostgreSQL Write does not retain the native unsupported stub")
	}
	matrix, err := buildCapabilityMatrixForConnectors(repoRootForCertificationTest(t), certificationConnectorAllowlist)
	if err != nil {
		t.Fatalf("buildCapabilityMatrixForConnectors() error = %v", err)
	}
	write, found := capabilityCellFor(matrix, "postgres", "capability:write")
	if !found || !write.Applicable || write.Declared || write.Implemented || write.LiveTested || len(write.LiveEvidence) != 0 {
		t.Fatalf("scoped PostgreSQL generic write cell = %#v, %t; want applicable but unpublished and unimplemented", write, found)
	}
}

func TestCertificationScopedSourceResolutionIgnoresUnrelatedRuntimeLedgerEntry(t *testing.T) {
	defsRoot := filepath.Join(repoRootForCertificationTest(t), "internal", "connectors", "defs")
	raw, err := os.ReadFile(filepath.Join(defsRoot, engine.RuntimeOperationEndpointLedgerFile))
	if err != nil {
		t.Fatalf("read runtime operation endpoint ledger: %v", err)
	}
	var entries map[string]json.RawMessage
	if err := json.Unmarshal(raw, &entries); err != nil {
		t.Fatalf("parse runtime operation endpoint ledger: %v", err)
	}
	entries["other"] = json.RawMessage(`[{"unexpected":true}]`)
	malformed, err := json.Marshal(entries)
	if err != nil {
		t.Fatalf("encode malformed runtime operation endpoint ledger: %v", err)
	}
	source := certificationRuntimeOperationEndpointLedgerFS{FS: os.DirFS(defsRoot), raw: malformed}
	if _, err := engine.Load(source, "github"); err == nil || !strings.Contains(err.Error(), "unknown field") {
		t.Fatalf("runtime Load with malformed unrelated ledger entry = %v, want strict decoder rejection", err)
	}
	scoped, err := scopedRuntimeOperationEndpointLedgerForCertification(source, "github")
	if err != nil {
		t.Fatalf("scope runtime operation endpoint ledger: %v", err)
	}
	if _, err := engine.Load(scoped, "github"); err != nil {
		t.Fatalf("generator-scoped Load with malformed unrelated ledger entry: %v", err)
	}
}

func TestCertificationCheckIgnoresMalformedNonAllowlistedRuntimeLedgerEntry(t *testing.T) {
	sourceRoot := repoRootForCertificationTest(t)
	sourcePostgresMatrix := filepath.Join(sourceRoot, "internal", "connectors", "defs", "postgres", "certification-matrix.json")
	sourceBefore, err := os.ReadFile(sourcePostgresMatrix)
	if err != nil {
		t.Fatalf("read source PostgreSQL matrix before isolated generator: %v", err)
	}
	root := certificationCommandWorkspace(t)
	writeTestCurrentCertificationSubject(t, root, strings.Repeat("a", 64))
	bootstrap := exec.Command("go", "run", "./cmd/connectorgen", "certification-matrix", "--all")
	bootstrap.Dir = root
	bootstrap.Env = append(os.Environ(), "GOCACHE="+t.TempDir())
	if output, err := bootstrap.CombinedOutput(); err != nil {
		t.Fatalf("bootstrap certification-matrix --all: %v\n%s", err, output)
	}
	sourceAfter, err := os.ReadFile(sourcePostgresMatrix)
	if err != nil {
		t.Fatalf("read source PostgreSQL matrix after isolated generator: %v", err)
	}
	if !bytes.Equal(sourceBefore, sourceAfter) {
		t.Fatal("isolated certification generator rewrote the source PostgreSQL matrix")
	}
	ledgerPath := filepath.Join(root, "internal", "connectors", "defs", engine.RuntimeOperationEndpointLedgerFile)
	raw, err := os.ReadFile(ledgerPath)
	if err != nil {
		t.Fatalf("read runtime operation endpoint ledger: %v", err)
	}
	var entries map[string]json.RawMessage
	if err := json.Unmarshal(raw, &entries); err != nil {
		t.Fatalf("parse runtime operation endpoint ledger: %v", err)
	}
	if _, found := entries["mysql"]; !found {
		t.Fatal("runtime operation endpoint ledger does not contain mysql")
	}
	entries["mysql"] = json.RawMessage(`[{"unexpected":true}]`)
	malformed, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		t.Fatalf("encode malformed runtime operation endpoint ledger: %v", err)
	}
	malformed = append(malformed, '\n')
	if err := os.Remove(ledgerPath); err != nil {
		t.Fatalf("unlink copied runtime operation endpoint ledger: %v", err)
	}
	if err := os.WriteFile(ledgerPath, malformed, 0o644); err != nil {
		t.Fatalf("write malformed runtime operation endpoint ledger: %v", err)
	}

	command := exec.Command("go", "run", "./cmd/connectorgen", "certification-matrix", "--check")
	command.Dir = root
	command.Env = append(os.Environ(), "GOCACHE="+t.TempDir())
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	if err := command.Run(); err != nil {
		t.Fatalf("certification-matrix --check with malformed mysql ledger: %v\nstdout:\n%s\nstderr:\n%s", err, stdout.String(), stderr.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("certification-matrix --check stderr = %q, want empty", stderr.String())
	}
	if !strings.Contains(stdout.String(), "certification shards are current:") {
		t.Fatalf("certification-matrix --check stdout = %q, want current-shards confirmation", stdout.String())
	}
}

func certificationCommandWorkspace(t *testing.T) string {
	t.Helper()
	root := repoRootForCertificationTest(t)
	workspace := t.TempDir()
	for _, relative := range []string{"go.mod", "go.sum", "cmd", "internal"} {
		copyCertificationCommandPath(t, filepath.Join(root, relative), filepath.Join(workspace, relative))
	}
	return workspace
}

func copyCertificationCommandPath(t *testing.T, source, destination string) {
	t.Helper()
	info, err := os.Lstat(source)
	if err != nil {
		t.Fatalf("inspect command workspace source %q: %v", source, err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		target, err := os.Readlink(source)
		if err != nil {
			t.Fatalf("read command workspace symlink %q: %v", source, err)
		}
		if err := os.Symlink(target, destination); err != nil {
			t.Fatalf("copy command workspace symlink %q: %v", source, err)
		}
		return
	}
	if !info.IsDir() {
		copyCertificationCommandFile(t, source, destination, info.Mode())
		return
	}
	if err := filepath.WalkDir(source, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relative, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		target := filepath.Join(destination, relative)
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return os.MkdirAll(target, info.Mode().Perm())
		}
		if entry.Type()&os.ModeSymlink != 0 {
			link, err := os.Readlink(path)
			if err != nil {
				return err
			}
			return os.Symlink(link, target)
		}
		copyCertificationCommandFile(t, path, target, info.Mode())
		return nil
	}); err != nil {
		t.Fatalf("copy command workspace tree %q: %v", source, err)
	}
}

func copyCertificationCommandFile(t *testing.T, source, destination string, mode os.FileMode) {
	t.Helper()
	in, err := os.Open(source)
	if err != nil {
		t.Fatalf("open command workspace source %q: %v", source, err)
	}
	defer func() { _ = in.Close() }()
	out, err := os.OpenFile(destination, os.O_WRONLY|os.O_CREATE|os.O_EXCL, mode.Perm())
	if err != nil {
		t.Fatalf("create command workspace destination %q: %v", destination, err)
	}
	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		t.Fatalf("copy command workspace file %q: %v", source, err)
	}
	if err := out.Close(); err != nil {
		t.Fatalf("close command workspace destination %q: %v", destination, err)
	}
}

func TestCertificationDiscoversStableWarehouseFacingSyncPrimitives(t *testing.T) {
	primitives, err := discoverSyncPrimitives(repoRootForCertificationTest(t))
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

func TestCertificationSyncModeDiscoveryUsesAllModesSource(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "internal", "synccontract", "all_modes.go")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir synccontract source: %v", err)
	}
	if err := os.WriteFile(path, []byte("package synccontract\nfunc AllModes() {}\n"), 0o600); err != nil {
		t.Fatalf("write synccontract source: %v", err)
	}
	modes, err := discoverSyncModes(root)
	if err != nil {
		t.Fatalf("discoverSyncModes() error = %v", err)
	}
	want := "internal/synccontract/all_modes.go:AllModes"
	for _, mode := range modes {
		if mode.DiscoverySource != want {
			t.Fatalf("sync mode %q discovery source = %q, want %q", mode.ID, mode.DiscoverySource, want)
		}
	}
}

func TestCertificationSyncModeDiscoveryRequiresAllModesSymbol(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "internal", "synccontract", "mode.go")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir synccontract source: %v", err)
	}
	if err := os.WriteFile(path, []byte("package synccontract\n"), 0o600); err != nil {
		t.Fatalf("write synccontract source: %v", err)
	}
	if _, err := discoverSyncModes(root); err == nil || !strings.Contains(err.Error(), "AllModes") {
		t.Fatalf("discoverSyncModes() error = %v, want missing AllModes failure", err)
	}
}

func TestCertificationSyncPrimitiveDiscoveryRequiresMetadataSymbol(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "internal", "connectors", "connectors.go")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir connector source: %v", err)
	}
	if err := os.WriteFile(path, []byte("package connectors\ntype Connector interface { Name() string }\n"), 0o600); err != nil {
		t.Fatalf("write connector source: %v", err)
	}
	if _, err := discoverSyncPrimitives(root); err == nil || !strings.Contains(err.Error(), "Connector.Metadata") {
		t.Fatalf("discoverSyncPrimitives() error = %v, want missing Connector.Metadata failure", err)
	}
}

func TestCertificationShardRequiresDestinationRoleContextForSourceOwnedPair(t *testing.T) {
	shards, err := buildCertificationShards(repoRootForCertificationTest(t), certificationConnectorAllowlist)
	if err != nil {
		t.Fatalf("buildCertificationShards() error = %v", err)
	}
	github := shards["github"]
	postgres := shards["postgres"]
	sourceRole, found := flowRoleForConnector([]connectorFlowRoles{github.ConnectorRoles}, "github", "api_source")
	if !found {
		t.Fatal("github api_source role is missing")
	}
	destinationRole, found := flowRoleForConnector([]connectorFlowRoles{postgres.ConnectorRoles}, "postgres", "database_destination")
	if !found {
		t.Fatal("postgres database_destination role is missing")
	}
	override := flowPairOverride{
		FlowKind:        "api_to_database",
		Source:          "github",
		Destination:     "postgres",
		Mediator:        localWarehouseMediator,
		DestinationRole: destinationRole,
		Cell:            baseFlowCell(sourceRole, destinationRole),
	}
	github.PairOverrides = []flowPairOverride{override}
	if err := validateCertificationShard(github); err != nil {
		t.Fatalf("validateCertificationShard() error = %v", err)
	}
	github.PairOverrides[0].DestinationRole = connectorFlowRole{}
	if err := validateCertificationShard(github); err == nil {
		t.Fatal("validateCertificationShard() accepted source-owned pair without destination role context")
	}
}

func TestCertificationStatusArtifactRemediationUsesAll(t *testing.T) {
	path := filepath.Join(t.TempDir(), "status.json")
	if err := validateCertificationStatusArtifactFile(path); err == nil || !strings.Contains(err.Error(), "certification-matrix --all") {
		t.Fatalf("validateCertificationStatusArtifactFile() error = %v, want --all remediation", err)
	}
	if err := checkCertificationStatusGeneratedArtifact(path, []byte("{}\n")); err == nil || !strings.Contains(err.Error(), "certification-matrix --all") {
		t.Fatalf("checkCertificationStatusGeneratedArtifact() error = %v, want --all remediation", err)
	}
}

func TestCertificationStatusArtifactUsesAllGenerator(t *testing.T) {
	statuses := make([]connectorCertificationStatus, 0, len(certificationConnectorAllowlist))
	for _, connector := range certificationConnectorAllowlist {
		statuses = append(statuses, connectorCertificationStatus{
			Connector: connector,
			Label:     "COMMUNITY BUILD, UNCERTIFIED",
			Warning:   "This connector is reachable but is a COMMUNITY BUILD, UNCERTIFIED.",
		})
	}
	artifact := buildCertificationStatusArtifact(flowMatrix{ConnectorStatuses: statuses})
	if artifact.GeneratedCommand != "go run ./cmd/connectorgen certification-matrix --all" {
		t.Fatalf("status generated command = %q, want --all", artifact.GeneratedCommand)
	}
	if !reflect.DeepEqual(artifact.CertificationScope, certificationConnectorAllowlist) {
		t.Fatalf("status certification scope = %#v, want %#v", artifact.CertificationScope, certificationConnectorAllowlist)
	}
	raw, err := marshalGeneratedJSON(artifact)
	if err != nil {
		t.Fatalf("marshal status artifact: %v", err)
	}
	if err := validateCertificationStatusArtifactJSON(raw); err != nil {
		t.Fatalf("validate status artifact: %v", err)
	}
}

func TestCertificationSyncModePostgresDeclaredTransportPairRequiresLiveProof(t *testing.T) {
	matrix := certificationMatrixForTest(t)
	read, _ := capabilityCellFor(matrix, "postgres", "capability:read")
	write, _ := capabilityCellFor(matrix, "postgres", "capability:write")
	cdc, _ := capabilityCellFor(matrix, "postgres", "capability:cdc")
	bundles, err := loadSourceBundlesForConnectors(repoRootForCertificationTest(t), []string{"postgres"})
	if err != nil {
		t.Fatalf("loadSourceBundlesForConnectors() error = %v", err)
	}
	sources, err := matrixConnectorSourcesForNames(bundles, []string{"postgres"})
	if err != nil || len(sources) != 1 {
		t.Fatalf("matrixConnectorSourcesForNames() = %#v, %v; want PostgreSQL source", sources, err)
	}
	cell := syncModeCellFor(sources[0], "incremental_dedupe", syncPrimitive{
		ID:                 "database_write_from_warehouse",
		IntegrationType:    "database",
		Capability:         "write",
		WarehouseDirection: "from_warehouse",
	}, read, write, cdc, false, nil)
	if !cell.Applicable {
		t.Fatal("PostgreSQL database write was hidden as non-applicable")
	}
	if !cell.Declared || cell.Implemented {
		t.Fatalf("PostgreSQL declared managed transport cell = %#v, want declared but unimplemented until live proof", cell)
	}
}

func TestCertificationSyncModePostgresTransportPromotionRequiresExactPair(t *testing.T) {
	matrix := certificationMatrixForTest(t)
	read, _ := capabilityCellFor(matrix, "postgres", "capability:read")
	write, _ := capabilityCellFor(matrix, "postgres", "capability:write")
	cdc, _ := capabilityCellFor(matrix, "postgres", "capability:cdc")
	bundles, err := loadSourceBundlesForConnectors(repoRootForCertificationTest(t), []string{"postgres"})
	if err != nil {
		t.Fatalf("loadSourceBundlesForConnectors() error = %v", err)
	}
	sources, err := matrixConnectorSourcesForNames(bundles, []string{"postgres"})
	if err != nil || len(sources) != 1 {
		t.Fatalf("matrixConnectorSourcesForNames() = %#v, %v; want PostgreSQL source", sources, err)
	}
	bundle := *sources[0].bundle
	transport := bundle.SyncTransport.Clone()
	transport.Destination.Executor.ID = "not_postgres_managed_target"
	bundle.SyncTransport = transport
	source := sources[0]
	source.bundle = &bundle
	cell := syncModeCellFor(source, "incremental_dedupe", syncPrimitive{
		ID:                 "database_write_from_warehouse",
		IntegrationType:    "database",
		Capability:         "write",
		WarehouseDirection: "from_warehouse",
	}, read, write, cdc, false, nil)
	if cell.Implemented {
		t.Fatalf("non-PostgreSQL destination executor promoted database write: %#v", cell)
	}
}

func TestCertificationSyncModeReadRequiresDeclaredSourceTransportMode(t *testing.T) {
	matrix := certificationMatrixForTest(t)
	bundles, err := loadSourceBundlesForConnectors(repoRootForCertificationTest(t), []string{"github"})
	if err != nil {
		t.Fatalf("loadSourceBundlesForConnectors() error = %v", err)
	}
	sources, err := matrixConnectorSourcesForNames(bundles, []string{"github"})
	if err != nil {
		t.Fatalf("matrixConnectorSourcesForNames() error = %v", err)
	}
	if len(sources) != 1 || sources[0].bundle == nil || sources[0].bundle.SyncTransport == nil || sources[0].bundle.SyncTransport.Source == nil {
		t.Fatalf("GitHub matrix source = %#v, want one source transport declaration", sources)
	}

	// Keep this source transport intentionally narrower than GitHub's authored
	// declaration. The cell must follow this concrete declaration, not the
	// connector-level read capability that remains true for GitHub.
	bundle := *sources[0].bundle
	transport := bundle.SyncTransport.Clone()
	transport.Source.Modes = []synccontract.Mode{synccontract.ModeFullAppend}
	bundle.SyncTransport = transport
	source := sources[0]
	source.bundle = &bundle
	source.connector = engine.New(bundle, nil)

	cells, err := buildConnectorSyncModeCells([]matrixConnectorSource{source}, matrix, []syncModeKind{{ID: string(synccontract.ModeIncrementalDedupe)}}, []syncPrimitive{{
		ID:                 "api_read_into_warehouse",
		IntegrationType:    "api",
		Capability:         "read",
		WarehouseDirection: "into_warehouse",
	}}, nil)
	if err != nil {
		t.Fatalf("buildConnectorSyncModeCells() error = %v", err)
	}
	if len(cells) != 1 || len(cells[0].Cells) != 1 {
		t.Fatalf("sync cells = %#v, want one GitHub read cell", cells)
	}
	cell := cells[0].Cells[0]
	if !cell.Applicable {
		t.Fatalf("source mode cell = %#v, want applicable mode cell", cell)
	}
	if cell.Declared || cell.Implemented {
		t.Fatalf("source mode cell = %#v, want unadmitted transport mode to be neither declared nor implemented", cell)
	}
}

func TestCertificationChangeCaptureRequiresDatabaseReadIntoWarehouse(t *testing.T) {
	cell := syncModeCellFor(matrixConnectorSource{name: "github", integrationType: "api"}, "change_capture", syncPrimitive{
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

func TestCertificationChangeCaptureRequiresImplementedChangefeedContract(t *testing.T) {
	matrix := certificationMatrixForTest(t)
	read, _ := capabilityCellFor(matrix, "postgres", "capability:read")
	write, _ := capabilityCellFor(matrix, "postgres", "capability:write")
	cdc, _ := capabilityCellFor(matrix, "postgres", "capability:cdc")
	bundles, err := loadSourceBundlesForConnectors(repoRootForCertificationTest(t), []string{"postgres"})
	if err != nil {
		t.Fatalf("loadSourceBundlesForConnectors() error = %v", err)
	}
	sources, err := matrixConnectorSourcesForNames(bundles, []string{"postgres"})
	if err != nil || len(sources) != 1 {
		t.Fatalf("matrixConnectorSourcesForNames() = %#v, %v; want PostgreSQL source", sources, err)
	}
	bundle := *sources[0].bundle
	bundle.Changefeed = nil
	source := sources[0]
	source.bundle = &bundle
	cell := syncModeCellFor(source, string(synccontract.ModeChangeCapture), syncPrimitive{
		ID:                 "database_read_into_warehouse",
		IntegrationType:    "database",
		Capability:         "read",
		WarehouseDirection: "into_warehouse",
	}, read, write, cdc, false, nil)
	if cell.Declared || cell.Implemented {
		t.Fatalf("change-capture cell without implemented declaration = %#v, want neither declared nor implemented", cell)
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
		SchemaVersion:  certificationSchemaVersion,
		Scope:          evidenceScopeFlow,
		Source:         "github",
		Destination:    "github",
		FlowKind:       "api_to_api",
		Provider:       "github",
		ExecutedAt:     "2026-08-10T00:00:00Z",
		RunID:          "github-round-trip",
		PMBinarySHA256: strings.Repeat("e", 64),
		PMCommand:      "pm flow run --file github-round-trip.json --json",
		Passed:         true,
		RepositorySalt: []byte("0123456789abcdef0123456789abcdef"),
		PreparedValues: []string{secret},
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
