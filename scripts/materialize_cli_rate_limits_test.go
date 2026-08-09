package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors/connsdk"
)

func TestBuildRateLimitSourceLedgerUsesOfficialResearchAndExplicitGaps(t *testing.T) {
	tempDir := t.TempDir()
	targetPath := filepath.Join(tempDir, "targets.json")
	researchPath := filepath.Join(tempDir, "research.json")

	targets := rateLimitTargetLedger{Targets: []rateLimitTarget{
		{Connector: "harvest", SourceKind: "openapi", SourceURL: "https://api.example.test/harvest", RetrievedAt: "2026-08-06"},
		{Connector: "widget", SourceKind: "reference", SourceURL: "https://api.example.test/widget", RetrievedAt: "2026-08-07"},
		{Connector: "faker", SourceKind: "none", SourceURL: "https://faker.example.test", RetrievedAt: "2026-08-08", SourceReason: "The target has no external provider HTTP/API surface."},
	}}
	research := rateLimitResearchLedger{Records: []rateLimitResearchRecord{{
		Connector:   "harvest",
		Verdict:     "declared",
		SourceURL:   "https://docs.example.test/harvest/rate-limits",
		RetrievedAt: "2026-08-06",
		Reason:      "The official documentation publishes 100 requests per 15 seconds per account.",
	}}}
	writeRateLimitTestJSON(t, targetPath, targets)
	writeRateLimitTestJSON(t, researchPath, research)

	ledger, err := buildRateLimitSourceLedger(targetPath, researchPath, "2026-08-09")
	if err != nil {
		t.Fatalf("buildRateLimitSourceLedger: %v", err)
	}
	if got, want := len(ledger.Entries), 3; got != want {
		t.Fatalf("entry count = %d, want %d", got, want)
	}
	byConnector := make(map[string]rateLimitSourceEntry, len(ledger.Entries))
	for _, entry := range ledger.Entries {
		byConnector[entry.Connector] = entry
	}

	harvest := byConnector["harvest"].Declaration
	if harvest.State != connsdk.RateLimitStateDeclared || len(harvest.Policies) != 1 {
		t.Fatalf("harvest declaration = %+v, want one declared policy", harvest)
	}
	if got := harvest.Policies[0].Source.URL; got != "https://docs.example.test/harvest/rate-limits" {
		t.Fatalf("harvest policy source = %q", got)
	}

	widget := byConnector["widget"].Declaration
	if widget.State != connsdk.RateLimitStateUnknown {
		t.Fatalf("widget state = %q, want unknown", widget.State)
	}
	if !strings.Contains(widget.Reason, "https://api.example.test/widget") || !strings.Contains(widget.Reason, "does not publish a complete enforceable") {
		t.Fatalf("widget reason does not preserve the official publication gap: %q", widget.Reason)
	}

	faker := byConnector["faker"].Declaration
	if faker.State != connsdk.RateLimitStateNotApplicable || !strings.Contains(faker.Reason, "no external provider HTTP/API") {
		t.Fatalf("faker declaration = %+v, want no-provider not_applicable", faker)
	}
}

func TestWriteRateLimitDeclarationRefusesToOverwriteDifferentEvidence(t *testing.T) {
	tempDir := t.TempDir()
	destination := filepath.Join(tempDir, "rate_limits.json")
	original := []byte("{\"schema_version\":1,\"state\":\"unknown\",\"reason\":\"retained source evidence\"}\n")
	if err := os.WriteFile(destination, original, 0o644); err != nil {
		t.Fatalf("write original declaration: %v", err)
	}

	_, err := writeRateLimitDeclaration(destination, connsdk.RateLimits{
		SchemaVersion: 1,
		State:         connsdk.RateLimitStateUnknown,
		Reason:        "different retained source evidence",
	})
	if err == nil || !strings.Contains(err.Error(), "refusing to overwrite") {
		t.Fatalf("writeRateLimitDeclaration error = %v, want overwrite refusal", err)
	}
	after, err := os.ReadFile(destination)
	if err != nil {
		t.Fatalf("read retained declaration: %v", err)
	}
	if string(after) != string(original) {
		t.Fatalf("existing declaration changed after refusal:\n got %q\nwant %q", after, original)
	}
}

func TestWriteRateLimitDeclarationPreservesDeclaredOverUnknownFallback(t *testing.T) {
	tempDir := t.TempDir()
	destination := filepath.Join(tempDir, "rate_limits.json")
	declared, err := rateLimitDeclarationFromResearch(rateLimitResearchRecord{
		Connector:   "harvest",
		Verdict:     "declared",
		SourceURL:   "https://docs.example.test/harvest/rate-limits",
		RetrievedAt: "2026-08-06",
		Reason:      "Official provider policy.",
	})
	if err != nil {
		t.Fatalf("build declared rate limit: %v", err)
	}
	if err := writeJSONAtomically(destination, declared); err != nil {
		t.Fatalf("write declared rate limit: %v", err)
	}

	status, err := writeRateLimitDeclaration(destination, connsdk.RateLimits{
		SchemaVersion: 1,
		State:         connsdk.RateLimitStateUnknown,
		Reason:        "generated fallback must not replace provider-cited declared evidence",
	})
	if err != nil {
		t.Fatalf("preserve declared rate limit: %v", err)
	}
	if status != "preserved_declared" {
		t.Fatalf("status = %q, want preserved_declared", status)
	}
	var after connsdk.RateLimits
	readRateLimitTestJSON(t, destination, &after)
	if after.State != connsdk.RateLimitStateDeclared || len(after.Policies) != 1 {
		t.Fatalf("declared record was downgraded: %+v", after)
	}
}

func TestPreserveDeclaredSourceEntriesOverUnknownFallback(t *testing.T) {
	declared, err := rateLimitDeclarationFromResearch(rateLimitResearchRecord{
		Connector:   "harvest",
		Verdict:     "declared",
		SourceURL:   "https://docs.example.test/harvest/rate-limits",
		RetrievedAt: "2026-08-06",
		Reason:      "Official provider policy.",
	})
	if err != nil {
		t.Fatalf("build declared rate limit: %v", err)
	}
	generated := []rateLimitSourceEntry{{
		Connector: "harvest",
		Evidence:  rateLimitEvidence{Kind: "official_operation_reference", URL: "https://api.example.test/harvest", RetrievedAt: "2026-08-09"},
		Declaration: connsdk.RateLimits{
			SchemaVersion: 1,
			State:         connsdk.RateLimitStateUnknown,
			Reason:        "fallback",
		},
	}}
	existing := []rateLimitSourceEntry{{
		Connector:   "harvest",
		Evidence:    rateLimitEvidence{Kind: "official_rate_limit_reference", URL: "https://docs.example.test/harvest/rate-limits", RetrievedAt: "2026-08-06"},
		Declaration: declared,
	}}
	merged, changed, err := preserveDeclaredSourceEntries(generated, existing)
	if err != nil {
		t.Fatalf("preserveDeclaredSourceEntries: %v", err)
	}
	if !changed || merged[0].Declaration.State != connsdk.RateLimitStateDeclared || merged[0].Evidence.URL != existing[0].Evidence.URL {
		t.Fatalf("merged source entry = %+v, want retained declared evidence", merged[0])
	}
}

func TestPreserveDeclaredDefinitionEntriesPromotesSourceLedgerFallback(t *testing.T) {
	defsRoot := t.TempDir()
	connectorRoot := filepath.Join(defsRoot, "harvest")
	if err := os.Mkdir(connectorRoot, 0o755); err != nil {
		t.Fatalf("create connector root: %v", err)
	}
	declared, err := rateLimitDeclarationFromResearch(rateLimitResearchRecord{
		Connector:   "harvest",
		Verdict:     "declared",
		SourceURL:   "https://docs.example.test/harvest/rate-limits",
		RetrievedAt: "2026-08-06",
		Reason:      "Official provider policy.",
	})
	if err != nil {
		t.Fatalf("build declared rate limit: %v", err)
	}
	if err := writeJSONAtomically(filepath.Join(connectorRoot, "rate_limits.json"), declared); err != nil {
		t.Fatalf("write declared rate limit: %v", err)
	}

	entries := []rateLimitSourceEntry{{
		Connector: "harvest",
		Evidence:  rateLimitEvidence{Kind: "official_operation_reference", URL: "https://api.example.test/harvest", RetrievedAt: "2026-08-09"},
		Declaration: connsdk.RateLimits{
			SchemaVersion: 1,
			State:         connsdk.RateLimitStateUnknown,
			Reason:        "fallback",
		},
	}}
	merged, changed, err := preserveDeclaredDefinitionEntries(entries, defsRoot)
	if err != nil {
		t.Fatalf("preserveDeclaredDefinitionEntries: %v", err)
	}
	if !changed || merged[0].Declaration.State != connsdk.RateLimitStateDeclared || merged[0].Evidence.Kind != "preserved_declared_bundle_policy" {
		t.Fatalf("definition declaration did not promote source ledger: %+v", merged[0])
	}
}

func TestSelectRateLimitEntriesIsBoundedAndPreservesSourceOrder(t *testing.T) {
	entries := make([]rateLimitSourceEntry, 42)
	for i := range entries {
		entries[i].Connector = strings.Repeat("a", 1) + string(rune('a'+i%26))
	}
	selected, err := selectRateLimitEntries(entries, 40, 40)
	if err != nil {
		t.Fatalf("selectRateLimitEntries: %v", err)
	}
	if got, want := len(selected), 2; got != want {
		t.Fatalf("selected length = %d, want %d", got, want)
	}
	if selected[0].Connector != entries[40].Connector || selected[1].Connector != entries[41].Connector {
		t.Fatalf("selection did not preserve deterministic source order: %+v", selected)
	}
	if _, err := selectRateLimitEntries(entries, 0, maxRateLimitMaterializeBatchSize+1); err == nil {
		t.Fatal("oversized batch unexpectedly accepted")
	}
}

func writeRateLimitTestJSON(t *testing.T, destination string, value any) {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal test JSON: %v", err)
	}
	if err := os.WriteFile(destination, raw, 0o644); err != nil {
		t.Fatalf("write test JSON: %v", err)
	}
}

func readRateLimitTestJSON(t *testing.T, source string, destination any) {
	t.Helper()
	raw, err := os.ReadFile(source)
	if err != nil {
		t.Fatalf("read test JSON: %v", err)
	}
	if err := json.Unmarshal(raw, destination); err != nil {
		t.Fatalf("decode test JSON: %v", err)
	}
}
