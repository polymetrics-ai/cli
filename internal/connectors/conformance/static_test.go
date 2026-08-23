package conformance

import (
	"os"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors/engine"
)

// TestRunStaticChecks_GoodBundleAllPass exercises runStaticChecks directly
// (bypassing dynamic/replay checks) against the good control bundle.
func TestRunStaticChecks_GoodBundleAllPass(t *testing.T) {
	b := loadTestBundle(t, "testdata/good", "acme")
	checks := runStaticChecks(b)
	if len(checks) != len(staticCheckNames) {
		t.Fatalf("runStaticChecks returned %d checks, want %d (%v)", len(checks), len(staticCheckNames), staticCheckNames)
	}
	for _, c := range checks {
		if !c.Passed {
			t.Errorf("static check %s failed on good bundle: %s", c.Name, c.Error)
		}
	}
}

// TestStaticCheckNames_MatchDesignList locks the exact static check-name
// list from design §E.2 so a future refactor can't silently rename/drop one
// without the test noticing.
func TestStaticCheckNames_MatchDesignList(t *testing.T) {
	want := []string{
		"spec_schema_valid",
		"stream_schemas_valid",
		"pk_fields_exist",
		"cursor_fields_exist",
		"interpolations_resolve",
		"write_schemas_valid",
		"surface_complete",
		"docs_present",
		"secret_redaction",
		"fixtures_present",
	}
	if len(staticCheckNames) != len(want) {
		t.Fatalf("staticCheckNames = %v, want %v", staticCheckNames, want)
	}
	for i, name := range want {
		if staticCheckNames[i] != name {
			t.Fatalf("staticCheckNames[%d] = %q, want %q (full: %v)", i, staticCheckNames[i], name, staticCheckNames)
		}
	}
}

// TestReportFromLoadError_ClassifiesMetaSchemaFailure exercises the
// Load-error path used when a bundle fails to load at all (e.g. spec.json's
// meta-schema compile failure) — the report must still name a specific
// failing static check, not a bare unclassified error.
func TestReportFromLoadError_ClassifiesMetaSchemaFailure(t *testing.T) {
	rep := RunBundleDir(t, "testdata/invalid", "bad-spec-schema")
	if rep.Passed {
		t.Fatalf("expected Passed=false for a bundle that fails to load")
	}
	c := assertHasCheck(t, rep, "spec_schema_valid")
	if c.Passed {
		t.Fatalf("spec_schema_valid passed despite a Load error")
	}
}

// TestReportFromLoadError_SkipsRemainingChecks: when Load itself fails,
// every other check (which needs a loaded Bundle to run) is reported
// Skipped rather than silently absent, so downstream tooling always sees
// the full check list.
func TestReportFromLoadError_SkipsRemainingChecks(t *testing.T) {
	rep := RunBundleDir(t, "testdata/invalid", "bad-spec-schema")
	skipped := 0
	for _, c := range rep.Checks {
		if c.Name == "spec_schema_valid" {
			continue
		}
		if c.Skipped {
			skipped++
		}
	}
	if skipped == 0 {
		t.Fatalf("expected at least one Skipped check when Load fails, got none: %+v", rep.Checks)
	}
}

// TestCheckInterpolationsResolve_AuthWhenClauseUsesFullGrammar is the S3
// engine mini-wave item 2 regression case (wave1-pilot SUMMARY.md carried
// queue / REVIEW-A.md re-review R1/R3): conformance's own
// interpolations_resolve check must accept an `==`/`in`-shaped `when` clause
// on a spec-known key, not just bare truthiness — this is the exact bug the
// github auth_type restoration surfaced (TestConformance/github regressed
// with "unknown spec key ... referenced as config.auth_type in [...]" until
// checkInterpolationsResolve's `when` field was routed through
// engine.ResolveCheckWhen instead of plain engine.ResolveCheck).
func TestCheckInterpolationsResolve_AuthWhenClauseUsesFullGrammar(t *testing.T) {
	specRaw := []byte(`{
		"type": "object",
		"properties": { "auth_type": {"type": "string"}, "base_url": {"type": "string"} }
	}`)
	spec, err := engine.CompileSchema(specRaw)
	if err != nil {
		t.Fatalf("CompileSchema: %v", err)
	}

	b := engine.Bundle{
		Name: "acme",
		Spec: spec,
		HTTP: engine.HTTPBase{
			URL: "{{ config.base_url }}",
			Auth: []engine.AuthSpec{
				{Mode: "none", When: "{{ config.auth_type in ['public', 'none'] }}"},
			},
		},
	}

	if err := checkInterpolationsResolve(b); err != nil {
		t.Fatalf("checkInterpolationsResolve: unexpected error for a spec-known ==/in when clause: %v", err)
	}

	// A spec-UNKNOWN key inside the same grammar must still be rejected.
	badB := b
	badB.HTTP.Auth = []engine.AuthSpec{
		{Mode: "none", When: "{{ config.auth_typo in ['public', 'none'] }}"},
	}
	if err := checkInterpolationsResolve(badB); err == nil {
		t.Fatalf("checkInterpolationsResolve: expected an error for a spec-unknown key in an `in` when clause")
	}
}

func TestCheckSurfaceComplete_ClosedReadOperationsSatisfyCoverage(t *testing.T) {
	for _, tc := range []struct {
		intent string
		method string
	}{
		{intent: "direct_read", method: "GET"},
		{intent: "binary_download", method: "GET"},
		{intent: "text_export", method: "GET"},
		{intent: "status_check", method: "HEAD"},
	} {
		t.Run(tc.intent, func(t *testing.T) {
			path := "file " + strings.ReplaceAll(tc.intent, "_", "-")
			b := engine.Bundle{
				Name: "acme",
				Metadata: engine.Metadata{
					Capabilities: engine.Capabilities{Read: true},
				},
				Surface: &engine.APISurface{
					API: "https://api.acme.test",
					Endpoints: []engine.SurfaceEndpoint{{
						Method:    tc.method,
						Path:      "/files/{file_id}/download",
						CoveredBy: &engine.SurfaceCoverage{DirectReads: []string{path}},
					}},
				},
				CLISurface: &engine.CLISurface{
					Commands: []engine.CLICommand{{
						Path:         path,
						Intent:       tc.intent,
						Availability: "implemented",
					}},
				},
			}

			if err := checkSurfaceComplete(b); err != nil {
				t.Fatalf("checkSurfaceComplete: implemented %s must cover its endpoint: %v", tc.intent, err)
			}

			plannedB := b
			plannedB.CLISurface = &engine.CLISurface{
				Commands: []engine.CLICommand{{
					Path:         path,
					Intent:       tc.intent,
					Availability: "planned",
				}},
			}
			if err := checkSurfaceComplete(plannedB); err == nil {
				t.Fatalf("checkSurfaceComplete: a planned %s command must not satisfy covered_by.direct_reads", tc.intent)
			}
		})
	}
}

func TestCheckSurfaceComplete_AllowsClosedReadOperationModels(t *testing.T) {
	for _, model := range []string{"text_export", "status_check"} {
		t.Run(model, func(t *testing.T) {
			b := engine.Bundle{
				Name: "acme",
				Surface: &engine.APISurface{
					OperationLedgerVersion: 1,
					Endpoints: []engine.SurfaceEndpoint{{
						Method: "GET",
						Path:   "/files",
						Operation: &engine.SurfaceOperation{
							Model:            model,
							Status:           "blocked",
							Risk:             "low",
							BlockedByDefault: true,
							Reason:           "declaration-only test operation",
						},
					}},
				},
			}
			if err := checkSurfaceComplete(b); err != nil {
				t.Fatalf("checkSurfaceComplete: %s operation model = %v, want pass", model, err)
			}
		})
	}
}

// TestCheckSurfaceComplete_FixedGraphQLOperationsSatisfyCoverage keeps a
// shared physical GraphQL transport honest.  The endpoint is covered only by
// the fixed operation IDs it names; it is neither an arbitrary POST read nor a
// generic raw GraphQL escape hatch.
func TestCheckSurfaceComplete_FixedGraphQLOperationsSatisfyCoverage(t *testing.T) {
	b := engine.Bundle{
		Name:     "acme",
		Metadata: engine.Metadata{Capabilities: engine.Capabilities{Read: true, Write: true}},
		Operations: []engine.OperationSpec{
			{ID: "acme.graphql.query.viewer", Kind: "graphql_query", GraphQL: &engine.GraphQLOperationSpec{Path: "/graphql"}},
			{ID: "acme.graphql.mutation.close_widget", Kind: "graphql_mutation", GraphQL: &engine.GraphQLOperationSpec{Path: "/graphql"}},
		},
		Surface: &engine.APISurface{
			API: "https://api.acme.test",
			Endpoints: []engine.SurfaceEndpoint{{
				Method:    "POST",
				Path:      "/graphql",
				CoveredBy: &engine.SurfaceCoverage{Operations: []string{"acme.graphql.query.viewer", "acme.graphql.mutation.close_widget"}},
			}},
		},
	}

	if err := checkSurfaceComplete(b); err != nil {
		t.Fatalf("checkSurfaceComplete: fixed GraphQL operation coverage = %v, want pass", err)
	}

	missing := b
	missing.Surface = &engine.APISurface{
		API: "https://api.acme.test",
		Endpoints: []engine.SurfaceEndpoint{{
			Method:    "POST",
			Path:      "/graphql",
			CoveredBy: &engine.SurfaceCoverage{Operations: []string{"acme.graphql.query.missing"}},
		}},
	}
	if err := checkSurfaceComplete(missing); err == nil || !strings.Contains(err.Error(), "not a declared fixed GraphQL operation") {
		t.Fatalf("checkSurfaceComplete missing operation error = %v, want exact fixed GraphQL binding rejection", err)
	}
}

func TestCheckSurfaceComplete_RequiresV2EndpointProvenance(t *testing.T) {
	b := engine.Bundle{
		Name: "acme",
		Metadata: engine.Metadata{
			Capabilities: engine.Capabilities{Read: true},
		},
		Streams: []engine.StreamSpec{{Name: "widgets"}},
		Surface: &engine.APISurface{
			OperationLedgerVersion: 2,
			Artifacts: []engine.SurfaceArtifact{{
				ID:          "acme-openapi-2026-08-06",
				URL:         "https://docs.acme.test/openapi.yaml",
				RetrievedAt: "2026-08-06",
			}},
			Endpoints: []engine.SurfaceEndpoint{{
				Method:    "GET",
				Path:      "/widgets",
				CoveredBy: &engine.SurfaceCoverage{Stream: "widgets"},
			}},
		},
	}

	err := checkSurfaceComplete(b)
	if err == nil {
		t.Fatal("checkSurfaceComplete: v2 endpoint without provenance passed")
	}
	if !strings.Contains(err.Error(), "provenance is required") {
		t.Fatalf("checkSurfaceComplete error = %q, want missing provenance diagnostic", err)
	}
}

func TestCheckSurfaceComplete_ProvenanceDoesNotSatisfyCoverage(t *testing.T) {
	b := engine.Bundle{
		Name: "acme",
		Metadata: engine.Metadata{
			Capabilities: engine.Capabilities{Read: true},
		},
		Surface: &engine.APISurface{
			OperationLedgerVersion: 2,
			Artifacts: []engine.SurfaceArtifact{{
				ID:          "acme-openapi-2026-08-06",
				URL:         "https://docs.acme.test/openapi.yaml",
				RetrievedAt: "2026-08-06",
			}},
			Endpoints: []engine.SurfaceEndpoint{{
				Method: "GET",
				Path:   "/widgets",
				Provenance: &engine.SurfaceProvenance{
					Artifact:  "acme-openapi-2026-08-06",
					SourceURL: "https://docs.acme.test/api/widgets",
				},
			}},
		},
	}

	err := checkSurfaceComplete(b)
	if err == nil {
		t.Fatal("checkSurfaceComplete: provenance-only endpoint passed without a coverage classifier")
	}
	if !strings.Contains(err.Error(), "has no classifier") {
		t.Fatalf("checkSurfaceComplete error = %q, want unchanged covered_by classifier diagnostic", err)
	}
}

func TestCheckSurfaceComplete_V2ProvenanceSuppliesBlockedOperationCitation(t *testing.T) {
	b := engine.Bundle{
		Name: "acme",
		Surface: &engine.APISurface{
			OperationLedgerVersion: 2,
			Artifacts: []engine.SurfaceArtifact{{
				ID:          "acme-openapi-2026-08-06",
				URL:         "https://docs.acme.test/openapi.yaml",
				RetrievedAt: "2026-08-06",
			}},
			Endpoints: []engine.SurfaceEndpoint{{
				Method: "POST",
				Path:   "/widgets",
				Provenance: &engine.SurfaceProvenance{
					Artifact:  "acme-openapi-2026-08-06",
					SourceURL: "https://docs.acme.test/api/widgets#create",
				},
				Operation: &engine.SurfaceOperation{
					Model:            "sensitive_reverse_etl",
					Status:           "blocked",
					Risk:             "high",
					BlockedByDefault: true,
					Reason:           "requires sensitive-data safeguards",
				},
			}},
		},
	}

	if err := checkSurfaceComplete(b); err != nil {
		t.Fatalf("checkSurfaceComplete: v2 endpoint provenance should supply the blocked-operation citation: %v", err)
	}
}

func TestCheckSurfaceComplete_POSTDirectReadDoesNotRequireWriteCapability(t *testing.T) {
	b := engine.Bundle{
		Name: "acme",
		Metadata: engine.Metadata{
			Capabilities: engine.Capabilities{Read: true, Write: false},
		},
		Surface: &engine.APISurface{
			API: "https://api.acme.test",
			Endpoints: []engine.SurfaceEndpoint{{
				Method:    "POST",
				Path:      "/freeBusy",
				CoveredBy: &engine.SurfaceCoverage{DirectRead: "freebusy query"},
			}},
		},
		CLISurface: &engine.CLISurface{
			Commands: []engine.CLICommand{{
				Path:         "freebusy query",
				Intent:       "direct_read",
				Availability: "implemented",
			}},
		},
	}

	if err := checkSurfaceComplete(b); err != nil {
		t.Fatalf("checkSurfaceComplete rejected POST direct read with write disabled: %v", err)
	}
}

// sanity: confirm the invalid corpus directories actually exist on disk
// under this package (own corpus, not shared with cmd/connectorgen).
func TestInvalidCorpus_DirsExist(t *testing.T) {
	for _, dir := range []string{
		"bad-spec-schema", "bad-stream-schema", "pk-missing", "cursor-missing",
		"unresolved-interp", "stream-path-literal-placeholder",
		"write-schema-invalid", "surface-incomplete",
		"docs-missing-heading", "secret-in-fixture", "no-fixtures",
	} {
		if _, err := os.Stat("testdata/invalid/" + dir); err != nil {
			t.Fatalf("testdata/invalid/%s missing: %v", dir, err)
		}
	}
}
