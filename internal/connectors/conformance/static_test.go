package conformance

import (
	"os"
	"testing"
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

// sanity: confirm the invalid corpus directories actually exist on disk
// under this package (own corpus, not shared with cmd/connectorgen).
func TestInvalidCorpus_DirsExist(t *testing.T) {
	for _, dir := range []string{
		"bad-spec-schema", "bad-stream-schema", "pk-missing", "cursor-missing",
		"unresolved-interp", "write-schema-invalid", "surface-incomplete",
		"docs-missing-heading", "secret-in-fixture", "no-fixtures",
	} {
		if _, err := os.Stat("testdata/invalid/" + dir); err != nil {
			t.Fatalf("testdata/invalid/%s missing: %v", dir, err)
		}
	}
}
