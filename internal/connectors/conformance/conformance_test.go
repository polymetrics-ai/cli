// Package conformance implements conformance v2 (design §E.2): static
// structural/policy checks plus dynamic fixture-backed replay checks that
// exercise the REAL engine (internal/connectors/engine) against recorded
// fixture pages.
package conformance

import (
	"errors"
	"io/fs"
	"os"
	"testing"
	"testing/fstest"

	"polymetrics.ai/internal/connectors/engine"
)

// --- report shape ------------------------------------------------------

// TestReportJSONShape locks the {Connector, Checks: [{Name, Passed, Skipped,
// Error}], Passed} shape.
func TestReportJSONShape(t *testing.T) {
	rep := Report{
		Connector: "acme",
		Checks: []CheckResult{
			{Name: "spec_schema_valid", Passed: true},
			{Name: "fixtures_present", Passed: false, Error: "boom"},
			{Name: "write_validate", Skipped: true},
		},
	}
	rep.Passed = rep.computePassed()
	if rep.Passed {
		t.Fatalf("Report.Passed = true, want false (one failing check)")
	}

	raw, err := jsonMarshal(rep)
	if err != nil {
		t.Fatalf("marshal report: %v", err)
	}
	for _, key := range []string{`"connector"`, `"checks"`, `"passed"`, `"name"`, `"skipped"`} {
		if !containsString(string(raw), key) {
			t.Fatalf("report JSON missing key %s: %s", key, raw)
		}
	}
}

func TestReportPassedTrueWhenNoFailures(t *testing.T) {
	rep := Report{Checks: []CheckResult{
		{Name: "a", Passed: true},
		{Name: "b", Skipped: true},
	}}
	if !rep.computePassed() {
		t.Fatalf("computePassed() = false, want true (no failing checks)")
	}
}

// --- TestConformance over the on-disk defs tree ------------------------

func realDefsFS() fs.FS {
	// Use the on-disk bundle tree for conformance so replay fixtures and
	// api_surface.json remain covered without embedding those test-only files
	// in the production cmd/pm binary.
	return os.DirFS("../defs")
}

// TestConformance iterates the real defs tree via engine.LoadAll with per-bundle
// subtests t.Run(name, ...). Zero bundles today (wave0 ships no goldens
// yet) must pass trivially; goldens auto-join in Wave F. Later waves run
// `go test ./internal/connectors/conformance -run 'TestConformance/<name>'`,
// so the subtest name MUST be exactly the bundle/connector name.
//
// ENGINE HARDENING (hardening-ledger.md): engine.LoadAll now returns a
// non-nil *engine.LoadAllError whenever one or more (but not necessarily
// all) bundles fail to load, alongside every bundle that DID load cleanly
// (internal/connectors/engine's bundle.go doc comment on LoadAll has the
// full rationale — a single malformed bundle must not hide the rest of a
// ~400-bundle fleet from discovery). TestConformance no longer treats that
// error as fatal for the whole run: every bundle engine.Load actually
// failed to parse gets its own FAILING subtest (named after the bundle, so
// `-run 'TestConformance/<name>'` still isolates it) whose failure message
// is the load error itself, and every bundle that loaded still runs the
// full RunBundle check set exactly as before. This is a strengthening, not
// a loosening: previously a single bad bundle anywhere in the fleet made
// TestConformance report ZERO information about any other bundle at all
// (a bare t.Fatalf on the aggregate error, no subtests created); now every
// bundle — loadable or not — gets a named, individually-inspectable result.
func TestConformance(t *testing.T) {
	bundles, err := engine.LoadAll(realDefsFS())
	var loadErr *engine.LoadAllError
	if err != nil && !errors.As(err, &loadErr) {
		t.Fatalf("engine.LoadAll(realDefsFS()): %v", err)
	}
	for _, f := range loadErr.GetFailures() {
		f := f
		t.Run(f.Name, func(t *testing.T) {
			t.Fatalf("bundle %s failed to load: %v", f.Name, f.Err)
		})
	}
	for _, b := range bundles {
		b := b
		t.Run(b.Name, func(t *testing.T) {
			rep := RunBundle(b)
			if !rep.Passed {
				t.Fatalf("conformance failed for %s: %+v", b.Name, rep.Checks)
			}
		})
	}
}

func TestConformance_EmptyDefsTreePassesTrivially(t *testing.T) {
	// defs.FS ships zero bundles in wave0 (goldens land in Wave F); this
	// locks in the "empty tree passes" contract independent of whatever
	// the real defs tree happens to contain when this runs (belt & suspenders
	// on top of TestConformance itself, which already iterates whatever's there).
	bundles, err := engine.LoadAll(fstest.MapFS{})
	var loadErr *engine.LoadAllError
	if err != nil && !errors.As(err, &loadErr) {
		t.Fatalf("engine.LoadAll(empty FS): %v", err)
	}
	if len(bundles) != 0 {
		t.Skip("defs.FS is no longer empty (Wave F goldens landed); covered by TestConformance subtests instead")
	}
}

// --- good bundle: every check passes ------------------------------------

func loadTestBundle(t *testing.T, root, name string) engine.Bundle {
	t.Helper()
	b, err := engine.Load(os.DirFS(root), name)
	if err != nil {
		t.Fatalf("engine.Load(%s, %s): %v", root, name, err)
	}
	return b
}

func TestRunBundle_GoodBundlePassesEveryCheck(t *testing.T) {
	b := loadTestBundle(t, "testdata/good", "acme")
	rep := RunBundle(b)
	if !rep.Passed {
		t.Fatalf("good bundle failed conformance: %+v", rep.Checks)
	}
	if len(rep.Checks) == 0 {
		t.Fatalf("expected a non-empty check list")
	}
	for _, c := range rep.Checks {
		if !c.Passed && !c.Skipped {
			t.Errorf("check %s failed: %s", c.Name, c.Error)
		}
	}
	// Spot-check a representative static AND dynamic check name is present,
	// so a future refactor can't silently drop a whole category.
	assertHasCheck(t, rep, "spec_schema_valid")
	assertHasCheck(t, rep, "docs_present")
	assertHasCheck(t, rep, "fixtures_present")
	assertHasCheck(t, rep, "pagination_terminates")
	assertHasCheck(t, rep, "records_match_schema")
	assertHasCheck(t, rep, "cursor_advances")
	assertHasCheck(t, rep, "write_request_shape:update_widget")
	assertHasCheck(t, rep, "delete_semantics")
}

func assertHasCheck(t *testing.T, rep Report, name string) *CheckResult {
	t.Helper()
	for i := range rep.Checks {
		if rep.Checks[i].Name == name {
			return &rep.Checks[i]
		}
	}
	t.Fatalf("report missing check %q; got %+v", name, rep.Checks)
	return nil
}

// --- static checks: each fails on its own targeted invalid bundle -------

func TestStaticChecks_TargetedFailures(t *testing.T) {
	cases := []struct {
		dir   string
		check string
	}{
		{"bad-spec-schema", "spec_schema_valid"},
		{"bad-stream-schema", "stream_schemas_valid"},
		{"pk-missing", "pk_fields_exist"},
		{"cursor-missing", "cursor_fields_exist"},
		{"unresolved-interp", "interpolations_resolve"},
		{"write-schema-invalid", "write_schemas_valid"},
		{"surface-incomplete", "surface_complete"},
		{"docs-missing-heading", "docs_present"},
		{"secret-in-fixture", "secret_redaction"},
		{"no-fixtures", "fixtures_present"},
	}

	for _, tc := range cases {
		t.Run(tc.dir, func(t *testing.T) {
			rep := RunBundleDir(t, "testdata/invalid", tc.dir)
			if rep.Passed {
				t.Fatalf("expected conformance to fail for %s, got Passed=true: %+v", tc.dir, rep.Checks)
			}
			c := assertHasCheck(t, rep, tc.check)
			if c.Passed {
				t.Fatalf("check %q passed for %s, want failing", tc.check, tc.dir)
			}
			if c.Error == "" {
				t.Fatalf("failing check %q for %s has no Error message", tc.check, tc.dir)
			}
		})
	}

	if len(cases) < 10 {
		t.Fatalf("static self-test corpus has %d cases, want >= 10", len(cases))
	}
}

// RunBundleDir loads a single named bundle out of a directory that may
// contain sibling seeded-invalid bundles (mirrors
// cmd/connectorgen/main_test.go's singleBundleFS pattern: a Load error
// itself — e.g. a meta-schema compile failure — must still surface as a
// named, failing check rather than aborting the whole test).
func RunBundleDir(t *testing.T, root, name string) Report {
	t.Helper()
	b, err := engine.Load(os.DirFS(root), name)
	if err != nil {
		return ReportFromLoadError(name, err)
	}
	return RunBundle(b)
}

func containsString(haystack, needle string) bool {
	return len(needle) == 0 || (len(haystack) >= len(needle) && indexOf(haystack, needle) >= 0)
}

func indexOf(haystack, needle string) int {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return i
		}
	}
	return -1
}
