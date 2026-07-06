// Command connectorgen is the wave0 migration-tooling CLI: it validates
// declarative connector definition bundles (defs/), regenerates the two
// deterministic wiring files hookset_gen.go/nativeset_gen.go, and scaffolds
// new bundles.
package main

import (
	"bytes"
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

// --- validate: accepts the golden control bundle -----------------------------

func TestValidate_AcceptsGoodBundle(t *testing.T) {
	report, err := validateDir(os.DirFS("testdata/valid"))
	if err != nil {
		t.Fatalf("validateDir: %v", err)
	}
	if len(report.Findings) != 0 {
		t.Fatalf("expected zero findings for the good bundle, got %+v", report.Findings)
	}
	if report.ConnectorsChecked != 1 {
		t.Fatalf("ConnectorsChecked = %d, want 1", report.ConnectorsChecked)
	}
}

// TestValidate_WhenClauseEqualityAndMembershipAgainstSpecKnownKeyPasses is the
// S3 engine mini-wave item 2 regression case (wave1-pilot SUMMARY.md carried
// queue / REVIEW-A.md re-review R1/R3): a `when` clause using the `==`/`in`
// grammar against a REAL, spec-declared key (`auth_type`) must pass
// connectorgen validate cleanly. Before ResolveCheckWhen existed,
// ResolveCheck's bare-namespace.key-only parsing treated the entire
// "auth_type == 'token'" expression as one dotted reference and always
// hard-failed with an "unknown spec key" finding — even though `auth_type`
// IS declared — because no `==`/`in`-shaped reference could ever look like a
// valid two-segment "namespace.key" split. This fixture lives in its own
// parent dir (not testdata/valid, which TestValidate_AcceptsGoodBundle
// asserts contains exactly one connector) so it doesn't disturb that count.
func TestValidate_WhenClauseEqualityAndMembershipAgainstSpecKnownKeyPasses(t *testing.T) {
	fsys := singleBundleFS(t, "testdata/valid-extra", "when-clause-equality-valid")
	report, err := validateDir(fsys)
	if err != nil {
		t.Fatalf("validateDir: %v", err)
	}
	if len(report.Findings) != 0 {
		t.Fatalf("expected zero findings for a spec-known ==/in when clause, got %+v", report.Findings)
	}
}

// TestValidate_FanOutBundlePassesCleanly proves a well-formed fan_out block
// (S4 engine mini-wave item 2) — including a "{{ fanout.id }}" reference in
// stream.Path — passes connectorgen validate with zero findings.
func TestValidate_FanOutBundlePassesCleanly(t *testing.T) {
	fsys := singleBundleFS(t, "testdata/valid-extra", "fanout-valid")
	report, err := validateDir(fsys)
	if err != nil {
		t.Fatalf("validateDir: %v", err)
	}
	if len(report.Findings) != 0 {
		t.Fatalf("expected zero findings for a well-formed fan_out block, got %+v", report.Findings)
	}
}

// TestValidate_KeyedObjectBundlePassesCleanly proves a well-formed
// records.keyed_object/key_field block (S4 engine mini-wave item 3) passes
// connectorgen validate with zero findings.
func TestValidate_KeyedObjectBundlePassesCleanly(t *testing.T) {
	fsys := singleBundleFS(t, "testdata/valid-extra", "keyed-object-valid")
	report, err := validateDir(fsys)
	if err != nil {
		t.Fatalf("validateDir: %v", err)
	}
	if len(report.Findings) != 0 {
		t.Fatalf("expected zero findings for a well-formed records.keyed_object block, got %+v", report.Findings)
	}
}

// TestValidate_OAuth2ExtraParamsBundlePassesCleanly proves a well-formed
// oauth2_client_credentials auth.extra_params block (S4 engine mini-wave item
// 4) passes connectorgen validate with zero findings.
func TestValidate_OAuth2ExtraParamsBundlePassesCleanly(t *testing.T) {
	fsys := singleBundleFS(t, "testdata/valid-extra", "oauth2-extra-params-valid")
	report, err := validateDir(fsys)
	if err != nil {
		t.Fatalf("validateDir: %v", err)
	}
	if len(report.Findings) != 0 {
		t.Fatalf("expected zero findings for a well-formed oauth2_client_credentials extra_params block, got %+v", report.Findings)
	}
}

func TestValidate_CLISurfaceValidReferencesPassCleanly(t *testing.T) {
	report, err := validateDir(cliSurfaceBundleFS(validCLISurfaceJSON()))
	if err != nil {
		t.Fatalf("validateDir: %v", err)
	}
	if len(report.Findings) != 0 {
		t.Fatalf("expected zero findings for valid cli_surface.json, got %+v", report.Findings)
	}
}

func TestValidate_CLISurfaceUnknownStreamIsHardFinding(t *testing.T) {
	report, err := validateDir(cliSurfaceBundleFS(strings.ReplaceAll(validCLISurfaceJSON(), `"stream": "widgets"`, `"stream": "missing_widgets"`)))
	if err != nil {
		t.Fatalf("validateDir: %v", err)
	}
	assertFindingRule(t, report, "cli-surface", ruleCLISurfaceUnknownTarget)
}

func TestValidate_CLISurfaceImplementedETLRequiresStream(t *testing.T) {
	report, err := validateDir(cliSurfaceBundleFS(strings.ReplaceAll(validCLISurfaceJSON(), `"stream": "widgets",`, "")))
	if err != nil {
		t.Fatalf("validateDir: %v", err)
	}
	assertFindingRule(t, report, "cli-surface", ruleCLISurfaceMissingMapping)
}

func TestValidate_CLISurfaceSecretLookingExampleIsHardFinding(t *testing.T) {
	token := "gh" + "p_" + "1234567890abcdef1234567890abcdef1234"
	report, err := validateDir(cliSurfaceBundleFS(strings.ReplaceAll(validCLISurfaceJSON(), `pm cli-surface widget list --json`, `pm cli-surface auth --token `+token)))
	if err != nil {
		t.Fatalf("validateDir: %v", err)
	}
	assertFindingRule(t, report, "cli-surface", ruleSecretLiteral)
}

func TestValidate_CLISurfaceAPIRefCannotUseExcludedEndpoint(t *testing.T) {
	cliSurface := strings.Replace(validCLISurfaceJSON(), `{ "method": "GET", "path": "/widgets" }`, `{ "method": "GET", "path": "/widgets/export" }`, 1)
	fsys := cliSurfaceBundleFS(cliSurface)
	fsys["cli-surface/api_surface.json"] = &fstest.MapFile{Data: []byte(`{
		"api": "test API v1",
		"endpoints": [
			{ "method": "GET", "path": "/widgets", "covered_by": { "stream": "widgets" } },
			{ "method": "POST", "path": "/widgets", "covered_by": { "write": "create_widget" } },
			{ "method": "GET", "path": "/widgets/export", "excluded": { "category": "out_of_scope", "reason": "not exposed" } }
		]
	}`)}

	report, err := validateDir(fsys)
	if err != nil {
		t.Fatalf("validateDir: %v", err)
	}
	assertFindingRule(t, report, "cli-surface", ruleCLISurfaceSafety)
}

func TestValidate_CLISurfaceAPIRefMustMatchStreamOrWrite(t *testing.T) {
	cliSurface := strings.Replace(validCLISurfaceJSON(), `{ "method": "GET", "path": "/widgets" }`, `{ "method": "GET", "path": "/widget-writes" }`, 1)
	fsys := cliSurfaceBundleFS(cliSurface)
	fsys["cli-surface/api_surface.json"] = &fstest.MapFile{Data: []byte(`{
		"api": "test API v1",
		"endpoints": [
			{ "method": "GET", "path": "/widgets", "covered_by": { "stream": "widgets" } },
			{ "method": "GET", "path": "/widget-writes", "covered_by": { "write": "create_widget" } },
			{ "method": "POST", "path": "/widgets", "covered_by": { "write": "create_widget" } }
		]
	}`)}

	report, err := validateDir(fsys)
	if err != nil {
		t.Fatalf("validateDir: %v", err)
	}
	assertFindingRule(t, report, "cli-surface", ruleCLISurfaceSafety)
}

func TestValidate_CLISurfaceAPIRefFailsWhenSurfaceHasZeroEndpoints(t *testing.T) {
	fsys := cliSurfaceBundleFS(validCLISurfaceJSON())
	fsys["cli-surface/api_surface.json"] = &fstest.MapFile{Data: []byte(`{
		"api": "test API v1",
		"endpoints": []
	}`)}

	report, err := validateDir(fsys)
	if err != nil {
		t.Fatalf("validateDir: %v", err)
	}
	assertFindingRule(t, report, "cli-surface", ruleCLISurfaceUnknownTarget)
}

func TestValidate_CLISurfaceReverseETLRequiresRiskAndApproval(t *testing.T) {
	cliSurface := strings.ReplaceAll(validCLISurfaceJSON(), `
				"risk": "creates a widget",
				"approval": "reverse ETL writes require plan, preview, approval, execute",
`, "")
	report, err := validateDir(cliSurfaceBundleFS(cliSurface))
	if err != nil {
		t.Fatalf("validateDir: %v", err)
	}
	assertFindingRule(t, report, "cli-surface", ruleCLISurfaceSafety)
}

func TestValidate_CLISurfaceImplementedRawAPIIsBlocked(t *testing.T) {
	cliSurface := strings.Replace(validCLISurfaceJSON(), `"intent": "etl"`, `"intent": "raw_api"`, 1)
	report, err := validateDir(cliSurfaceBundleFS(cliSurface))
	if err != nil {
		t.Fatalf("validateDir: %v", err)
	}
	assertFindingRule(t, report, "cli-surface", ruleCLISurfaceSafety)
}

func TestValidate_CLISurfaceImplementedDirectWriteIsBlocked(t *testing.T) {
	cliSurface := strings.Replace(validCLISurfaceJSON(), `"intent": "reverse_etl"`, `"intent": "direct_write"`, 1)
	report, err := validateDir(cliSurfaceBundleFS(cliSurface))
	if err != nil {
		t.Fatalf("validateDir: %v", err)
	}
	assertFindingRule(t, report, "cli-surface", ruleCLISurfaceSafety)
}

// TestValidate_EmptyTreeIsFine mirrors the loader contract: an empty defs/
// tree (no bundle directories) passes with a zero connector count, so wave0's
// bundle-less internal/connectors/defs/ tree does not fail CI.
func TestValidate_EmptyTreeIsFine(t *testing.T) {
	dir := t.TempDir()
	report, err := validateDir(os.DirFS(dir))
	if err != nil {
		t.Fatalf("validateDir on empty tree: %v", err)
	}
	if report.ConnectorsChecked != 0 {
		t.Fatalf("ConnectorsChecked = %d, want 0", report.ConnectorsChecked)
	}
	if len(report.Findings) != 0 {
		t.Fatalf("expected zero findings on an empty tree, got %+v", report.Findings)
	}
}

// --- validate: seeded-invalid corpus (>=10 seeded, >=8 distinct classes) ----

func TestValidate_RejectsSeededInvalidBundles(t *testing.T) {
	cases := []struct {
		dir      string // testdata/invalid/<dir>
		wantRule string
	}{
		{"missing-metadata-file", ruleMissingFile},
		{"bad-spec-schema", ruleMetaSchema},
		{"unresolvable-interpolation", ruleInterpolationUnresolved},
		{"missing-schema-ref", ruleSchemaRefMissing},
		{"pk-not-in-schema", rulePrimaryKeyMissing},
		{"cursor-not-in-schema", ruleCursorFieldMissing},
		{"write-path-fields-not-in-schema", ruleWritePathFields},
		{"surface-both-covered-and-excluded", ruleSurfaceCoverage},
		{"surface-missing-stream", ruleSurfaceIncomplete},
		{"Source-GitHub", ruleNameRegex},
		{"secret-literal-in-fixture", ruleSecretLiteral},
		{"docs-missing-heading", ruleDocsHeading},
		{"surface-unknown-category", ruleMetaSchema},
		{"write-false-with-mutation-endpoint", ruleSurfaceFailFirstRun},
		{"auth-field-unknown-spec-key", ruleInterpolationUnresolved},
		{"unknown-filter-in-template", ruleInterpolationUnresolved},
		{"when-clause-equality-unknown-spec-key", ruleInterpolationUnresolved},
		{"skip-marker-missing-reason", ruleConformanceSkipReason},
		{"skip-marker-missing-reason-bundle", ruleConformanceSkipReason},
		{"default-type-mismatch", ruleDefaultTypeMismatch},
		{"unknown-base-key", ruleMetaSchema},
		// checkquery-ledger.md: base.check.query is now a real, engine-level
		// field (RequestSpec.Query) rather than an unknown key, so its
		// templates must be statically validated exactly like stream.Query's
		// — a check.query entry templating an undeclared spec key is a
		// ruleInterpolationUnresolved finding, the same rule
		// auth-field-unknown-spec-key already exercises for base.auth.
		{"check-query-unknown-spec-key", ruleInterpolationUnresolved},
		// S4 engine mini-wave item 2: fan_out.ids_from.request.path gets the
		// same static ResolveCheck treatment as an ordinary stream.Path.
		{"fanout-request-path-unknown-spec-key", ruleInterpolationUnresolved},
		// S4 engine mini-wave item 4: oauth2_client_credentials auth.extra_params
		// values get the same static ResolveCheck treatment as token_url/
		// client_id/client_secret/scopes.
		{"oauth2-extra-params-unknown-spec-key", ruleInterpolationUnresolved},
	}

	seenRules := map[string]bool{}
	for _, tc := range cases {
		t.Run(tc.dir, func(t *testing.T) {
			// validateDir mirrors engine.LoadAll's contract: its fsys root is
			// the PARENT of bundle directories, not a bundle directory
			// itself (a bundle's own subdirectories like schemas/ and
			// fixtures/ must never be mistaken for sibling bundles). So each
			// seeded case is validated in isolation by rooting fsys one
			// level up and filtering findings down to that one connector.
			fsys := singleBundleFS(t, "testdata/invalid", tc.dir)
			report, err := validateDir(fsys)
			if err != nil {
				// A hard structural error (e.g. missing metadata.json) surfaces
				// as a returned error from the loader rather than a Finding;
				// validate must still translate it into a named finding via
				// the caller. Exercise that path explicitly here too.
				t.Fatalf("validateDir(%s) returned a bare error instead of findings: %v", tc.dir, err)
			}
			var relevant []Finding
			for _, f := range report.Findings {
				if f.Connector == tc.dir {
					relevant = append(relevant, f)
				}
			}
			if len(relevant) == 0 {
				t.Fatalf("validateDir(%s): expected at least one finding for connector %q, got none (all findings: %+v)", tc.dir, tc.dir, report.Findings)
			}
			var found *Finding
			for i := range relevant {
				if relevant[i].Rule == tc.wantRule {
					found = &relevant[i]
					break
				}
			}
			if found == nil {
				t.Fatalf("validateDir(%s): no finding with rule %q; got %+v", tc.dir, tc.wantRule, relevant)
			}
			if found.Connector == "" {
				t.Fatalf("finding %+v missing connector name", found)
			}
			if found.File == "" {
				t.Fatalf("finding %+v missing file name", found)
			}
			if found.Message == "" {
				t.Fatalf("finding %+v missing message", found)
			}
		})
		seenRules[tc.wantRule] = true
	}

	if len(cases) < 10 {
		t.Fatalf("seeded corpus has %d cases, want >= 10", len(cases))
	}
	if len(seenRules) < 8 {
		t.Fatalf("seeded corpus covers %d distinct rules, want >= 8: %v", len(seenRules), seenRules)
	}
}

// --- gap-loop cycle-1 item 6 (REVIEW-A.md C3): validate-time hard FINDING
// for a spec.json "default" that does not type-check against its own
// declared "type" -------------------------------------------------------
//
// C3's materialization increment (engine/read.go's materializeConfigDefaults)
// fills an absent config key from spec.json's "default" verbatim; a default
// whose JSON type mismatches the property's declared type would silently
// materialize a wrong-shaped config value (e.g. default: 100 landing in a
// string-typed RuntimeConfig.Config, or a non-boolean string landing where a
// boolean was declared) — a hard validate FINDING (not a warning: this is a
// structural defect a bundle author can and must fix, unlike N2's
// plausibility heuristic below).

func TestValidate_DefaultTypeMismatchIsHardFinding(t *testing.T) {
	fsys := singleBundleFS(t, "testdata/invalid", "default-type-mismatch")
	report, err := validateDir(fsys)
	if err != nil {
		t.Fatalf("validateDir: %v", err)
	}
	var found *Finding
	for i := range report.Findings {
		if report.Findings[i].Connector == "default-type-mismatch" && report.Findings[i].Rule == ruleDefaultTypeMismatch {
			found = &report.Findings[i]
			break
		}
	}
	if found == nil {
		t.Fatalf("validateDir(default-type-mismatch): expected a %q finding, got %+v", ruleDefaultTypeMismatch, report.Findings)
	}
	if !strings.Contains(found.Message, "max_pages") {
		t.Fatalf("finding message %q does not name the offending property max_pages", found.Message)
	}
}

// TestValidate_WellTypedDefaultDoesNotTriggerMismatchRule proves a
// well-typed default (base_url's string default in the same seeded bundle)
// never itself triggers the rule — only the genuinely mismatched property
// (max_pages) does.
func TestValidate_WellTypedDefaultDoesNotTriggerMismatchRule(t *testing.T) {
	fsys := singleBundleFS(t, "testdata/invalid", "default-type-mismatch")
	report, err := validateDir(fsys)
	if err != nil {
		t.Fatalf("validateDir: %v", err)
	}
	for _, f := range report.Findings {
		if f.Rule == ruleDefaultTypeMismatch && strings.Contains(f.Message, "base_url") {
			t.Fatalf("well-typed base_url default incorrectly flagged: %+v", f)
		}
	}
}

// --- N2 (wave0 REVIEW.md carried flag): validate-time WARNING for a
// digit-shaped non-unix start_config_key value ------------------------------
//
// N2's narrow, honest scope (SPEC.md §4 "noted, not blocking... promote to a
// validate-time guard"): formatParam's digits-passthrough (B1) is CORRECT
// for param_format unix_seconds (an all-digits config value there really
// does mean Unix seconds) but is a silent-misinterpretation risk for
// param_format date/github_date_range, where a free-form (no declared
// date-ish format) start_config_key spec property could hold a value like
// "20260101" (yyyymmdd) that would be silently treated as a 1970s-era
// Unix-seconds lower bound instead of erroring. A property that DOES
// declare format:date-time (or format:date) is not flagged: an operator
// filling in a date-time-typed config field is exceedingly unlikely to type
// a bare yyyymmdd digit string, and the risk is specifically about
// UNDECLARED free-form string config. This is a WARNING (Report.Warnings,
// not Report.Findings) — never blocks validate's exit code or the "0
// findings" self-verify contract — because it is a plausibility heuristic,
// not a structural defect: a legitimately-Unix-seconds start_config_key
// used with date/github_date_range (unusual but not inexpressible) would
// otherwise be a false positive if this were a hard error.

func TestValidate_StartDateFreeFormStringWarns(t *testing.T) {
	fsys := singleBundleFS(t, "testdata/invalid", "start-date-free-form-string")
	report, err := validateDir(fsys)
	if err != nil {
		t.Fatalf("validateDir: %v", err)
	}
	if len(report.Findings) != 0 {
		t.Fatalf("expected zero hard Findings (this is warning-only), got %+v", report.Findings)
	}
	var found *Finding
	for i := range report.Warnings {
		if report.Warnings[i].Connector == "start-date-free-form-string" && report.Warnings[i].Rule == ruleStartDateFreeFormString {
			found = &report.Warnings[i]
			break
		}
	}
	if found == nil {
		t.Fatalf("expected a %s warning for start-date-free-form-string, got %+v", ruleStartDateFreeFormString, report.Warnings)
	}
	if found.File == "" || found.Message == "" {
		t.Fatalf("warning %+v missing file/message", found)
	}
}

// TestValidate_StartDateWithDateTimeFormatNoWarning is the no-false-positive
// companion: an identical incremental shape whose start_config_key spec
// property DOES declare format:date-time must not warn.
func TestValidate_StartDateWithDateTimeFormatNoWarning(t *testing.T) {
	fsys := singleBundleFS(t, "testdata/invalid", "start-date-rfc3339-format-no-warning")
	report, err := validateDir(fsys)
	if err != nil {
		t.Fatalf("validateDir: %v", err)
	}
	for _, w := range report.Warnings {
		if w.Connector == "start-date-rfc3339-format-no-warning" && w.Rule == ruleStartDateFreeFormString {
			t.Fatalf("unexpected %s warning for a format:date-time start_config_key: %+v", ruleStartDateFreeFormString, w)
		}
	}
}

// TestValidate_UnixSecondsStartDateNeverWarns locks in that param_format
// unix_seconds is never flagged by this rule at all (digits ARE the correct
// shape there) — reusing the real stripe golden, whose start_date has no
// declared format annotation either, proves this directly against
// production defs, not just a synthetic fixture.
func TestValidate_UnixSecondsStartDateNeverWarns(t *testing.T) {
	report, err := validateDir(os.DirFS(filepath.Join("..", "..", "internal", "connectors", "defs")))
	if err != nil {
		t.Fatalf("validateDir(defs): %v", err)
	}
	for _, w := range report.Warnings {
		if w.Rule == ruleStartDateFreeFormString {
			t.Fatalf("unexpected %s warning against the real defs corpus: %+v", ruleStartDateFreeFormString, w)
		}
	}
}

// TestValidate_ExitCodeReflectsFindings exercises the run() entry point (the
// one main() calls) end to end: a directory with findings must exit 1; a
// clean directory must exit 0.
func TestValidate_ExitCodeReflectsFindings(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"validate", "testdata/invalid/bad-spec-schema"}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("run(validate invalid) exit = %d, want 1; stderr=%s", code, stderr.String())
	}

	stdout.Reset()
	stderr.Reset()
	code = run([]string{"validate", "testdata/valid"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("run(validate valid) exit = %d, want 0; stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
}

// --- validate --json shape ---------------------------------------------------

func TestValidate_JSONOutputShape(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"validate", "testdata/invalid/bad-spec-schema", "--json"}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("run --json exit = %d, want 1", code)
	}

	var generic map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &generic); err != nil {
		t.Fatalf("--json output is not valid JSON: %v\noutput: %s", err, stdout.String())
	}
	if _, ok := generic["connectors_checked"]; !ok {
		t.Fatalf("--json output missing connectors_checked: %s", stdout.String())
	}
	findingsRaw, ok := generic["findings"].([]any)
	if !ok || len(findingsRaw) == 0 {
		t.Fatalf("--json output missing non-empty findings array: %s", stdout.String())
	}
	entry, ok := findingsRaw[0].(map[string]any)
	if !ok {
		t.Fatalf("findings[0] is not an object: %v", findingsRaw[0])
	}
	for _, key := range []string{"connector", "file", "rule", "message"} {
		if _, ok := entry[key]; !ok {
			t.Fatalf("findings[0] missing key %q: %s", key, stdout.String())
		}
	}
}

// TestValidate_JSONOutputCleanRun asserts the --json shape on a passing run
// too (empty findings array, not a missing key / null).
func TestValidate_JSONOutputCleanRun(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"validate", "testdata/valid", "--json"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("run --json (clean) exit = %d, want 0; stderr=%s", code, stderr.String())
	}
	var generic map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &generic); err != nil {
		t.Fatalf("--json output is not valid JSON: %v", err)
	}
	findingsRaw, ok := generic["findings"].([]any)
	if !ok {
		t.Fatalf("--json output findings is not an array: %s", stdout.String())
	}
	if len(findingsRaw) != 0 {
		t.Fatalf("--json output findings should be empty for a clean run, got %v", findingsRaw)
	}
}

// --- gen: deterministic byte-stable regeneration -----------------------------

func TestGen_HooksetWritesEmptyImportList(t *testing.T) {
	hooksRoot := t.TempDir()
	if err := os.MkdirAll(filepath.Join(hooksRoot, "hookset"), 0o755); err != nil {
		t.Fatalf("mkdir hookset: %v", err)
	}

	if err := genHookset(hooksRoot); err != nil {
		t.Fatalf("genHookset: %v", err)
	}

	raw, err := os.ReadFile(filepath.Join(hooksRoot, "hookset", "hookset_gen.go"))
	if err != nil {
		t.Fatalf("read hookset_gen.go: %v", err)
	}
	if !strings.Contains(string(raw), "package hookset") {
		t.Fatalf("hookset_gen.go missing package clause: %s", raw)
	}
	if !strings.Contains(string(raw), "Code generated") {
		t.Fatalf("hookset_gen.go missing generated-by header: %s", raw)
	}
	if strings.Contains(string(raw), "_ \"") {
		t.Fatalf("hookset_gen.go should have an empty import list (no hooks/<name> packages exist yet): %s", raw)
	}
}

func TestGen_HooksetImportsEveryHookPackageExceptHookset(t *testing.T) {
	hooksRoot := t.TempDir()
	for _, name := range []string{"hookset", "acme"} {
		if err := os.MkdirAll(filepath.Join(hooksRoot, name), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", name, err)
		}
	}
	if err := os.WriteFile(filepath.Join(hooksRoot, "acme", "hooks.go"), []byte("package acme\n"), 0o644); err != nil {
		t.Fatalf("write acme/hooks.go: %v", err)
	}

	if err := genHookset(hooksRoot); err != nil {
		t.Fatalf("genHookset: %v", err)
	}

	raw, err := os.ReadFile(filepath.Join(hooksRoot, "hookset", "hookset_gen.go"))
	if err != nil {
		t.Fatalf("read hookset_gen.go: %v", err)
	}
	if !strings.Contains(string(raw), `_ "polymetrics.ai/internal/connectors/hooks/acme"`) {
		t.Fatalf("hookset_gen.go missing blank import for acme: %s", raw)
	}
	if strings.Contains(string(raw), "/hooks/hookset\"") {
		t.Fatalf("hookset_gen.go must not import itself: %s", raw)
	}
}

func TestGen_NativesetWritesEmptyImportListWhenNoNativePackages(t *testing.T) {
	nativeRoot := t.TempDir()
	if err := os.MkdirAll(filepath.Join(nativeRoot, "nativeset"), 0o755); err != nil {
		t.Fatalf("mkdir nativeset: %v", err)
	}

	if err := genNativeset(nativeRoot); err != nil {
		t.Fatalf("genNativeset: %v", err)
	}

	raw, err := os.ReadFile(filepath.Join(nativeRoot, "nativeset", "nativeset_gen.go"))
	if err != nil {
		t.Fatalf("read nativeset_gen.go: %v", err)
	}
	if !strings.Contains(string(raw), "package nativeset") {
		t.Fatalf("nativeset_gen.go missing package clause: %s", raw)
	}
	if strings.Contains(string(raw), "_ \"") {
		t.Fatalf("nativeset_gen.go should have an empty import list, got: %s", raw)
	}
}

func TestGen_NativesetImportsEveryNativePackageExceptNativeset(t *testing.T) {
	nativeRoot := t.TempDir()
	for _, name := range []string{"nativeset", "postgres"} {
		if err := os.MkdirAll(filepath.Join(nativeRoot, name), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", name, err)
		}
	}
	if err := os.WriteFile(filepath.Join(nativeRoot, "postgres", "connector.go"), []byte("package postgres\n"), 0o644); err != nil {
		t.Fatalf("write postgres/connector.go: %v", err)
	}

	if err := genNativeset(nativeRoot); err != nil {
		t.Fatalf("genNativeset: %v", err)
	}

	raw, err := os.ReadFile(filepath.Join(nativeRoot, "nativeset", "nativeset_gen.go"))
	if err != nil {
		t.Fatalf("read nativeset_gen.go: %v", err)
	}
	if !strings.Contains(string(raw), `_ "polymetrics.ai/internal/connectors/native/postgres"`) {
		t.Fatalf("nativeset_gen.go missing blank import for postgres: %s", raw)
	}
}

// TestGen_ByteStableOnRerun is the core determinism guarantee: running gen
// twice against the same input tree must produce byte-identical output.
func TestGen_ByteStableOnRerun(t *testing.T) {
	hooksRoot := t.TempDir()
	for _, name := range []string{"hookset", "acme", "beta"} {
		if err := os.MkdirAll(filepath.Join(hooksRoot, name), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", name, err)
		}
	}
	if err := os.WriteFile(filepath.Join(hooksRoot, "acme", "hooks.go"), []byte("package acme\n"), 0o644); err != nil {
		t.Fatalf("write acme/hooks.go: %v", err)
	}
	if err := os.WriteFile(filepath.Join(hooksRoot, "beta", "hooks.go"), []byte("package beta\n"), 0o644); err != nil {
		t.Fatalf("write beta/hooks.go: %v", err)
	}

	if err := genHookset(hooksRoot); err != nil {
		t.Fatalf("genHookset (1st): %v", err)
	}
	first, err := os.ReadFile(filepath.Join(hooksRoot, "hookset", "hookset_gen.go"))
	if err != nil {
		t.Fatalf("read 1st: %v", err)
	}

	if err := genHookset(hooksRoot); err != nil {
		t.Fatalf("genHookset (2nd): %v", err)
	}
	second, err := os.ReadFile(filepath.Join(hooksRoot, "hookset", "hookset_gen.go"))
	if err != nil {
		t.Fatalf("read 2nd: %v", err)
	}

	if !bytes.Equal(first, second) {
		t.Fatalf("genHookset not byte-stable across reruns:\n1st: %s\n2nd: %s", first, second)
	}
}

// TestGen_RunCommandRegeneratesBothFiles exercises `connectorgen gen` through
// run() against a scratch tree shaped like the real repo layout.
func TestGen_RunCommandRegeneratesBothFiles(t *testing.T) {
	root := t.TempDir()
	hooksDir := filepath.Join(root, "internal/connectors/hooks")
	nativeDir := filepath.Join(root, "internal/connectors/native")
	if err := os.MkdirAll(filepath.Join(hooksDir, "hookset"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(nativeDir, "nativeset"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	var stdout, stderr bytes.Buffer
	code := runGenAt([]string{"gen"}, &stdout, &stderr, hooksDir, nativeDir)
	if code != 0 {
		t.Fatalf("runGenAt(gen) exit = %d, stderr=%s", code, stderr.String())
	}

	if _, err := os.Stat(filepath.Join(hooksDir, "hookset", "hookset_gen.go")); err != nil {
		t.Fatalf("hookset_gen.go not written: %v", err)
	}
	if _, err := os.Stat(filepath.Join(nativeDir, "nativeset", "nativeset_gen.go")); err != nil {
		t.Fatalf("nativeset_gen.go not written: %v", err)
	}
}

// --- new: scaffolding ---------------------------------------------------------

func TestNew_ScaffoldsBundleThatPassesValidate(t *testing.T) {
	root := t.TempDir()

	if err := scaffoldNew(root, "acme-widgets"); err != nil {
		t.Fatalf("scaffoldNew: %v", err)
	}

	for _, f := range []string{"metadata.json", "spec.json", "streams.json", "api_surface.json", "docs.md"} {
		if _, err := os.Stat(filepath.Join(root, "acme-widgets", f)); err != nil {
			t.Fatalf("scaffold missing %s: %v", f, err)
		}
	}
	if _, err := os.Stat(filepath.Join(root, "acme-widgets", "schemas")); err != nil {
		t.Fatalf("scaffold missing schemas/ dir: %v", err)
	}

	report, err := validateDir(os.DirFS(root))
	if err != nil {
		t.Fatalf("validateDir(scaffold): %v", err)
	}
	if len(report.Findings) != 0 {
		t.Fatalf("scaffolded bundle failed validate: %+v", report.Findings)
	}
}

func TestNew_RejectsInvalidName(t *testing.T) {
	root := t.TempDir()
	for _, name := range []string{"Acme", "-acme", "acme_widgets", "", "acme widgets"} {
		if err := scaffoldNew(root, name); err == nil {
			t.Fatalf("scaffoldNew(%q) succeeded, want name-regex rejection", name)
		}
	}
}

func TestNew_RejectsExistingDir(t *testing.T) {
	root := t.TempDir()
	if err := scaffoldNew(root, "acme-widgets"); err != nil {
		t.Fatalf("scaffoldNew (first): %v", err)
	}
	if err := scaffoldNew(root, "acme-widgets"); err == nil {
		t.Fatalf("scaffoldNew (second, same name) succeeded, want existing-dir rejection")
	}
}

// TestNew_RunCommandScaffolds exercises `connectorgen new <name>` through
// run() end to end.
func TestNew_RunCommandScaffolds(t *testing.T) {
	root := t.TempDir()
	var stdout, stderr bytes.Buffer
	code := runNewAt([]string{"new", "acme-widgets"}, &stdout, &stderr, root)
	if code != 0 {
		t.Fatalf("runNewAt(new) exit = %d, stderr=%s", code, stderr.String())
	}
	if _, err := os.Stat(filepath.Join(root, "acme-widgets", "metadata.json")); err != nil {
		t.Fatalf("new did not scaffold metadata.json: %v", err)
	}
}

func TestNew_RunCommandMissingArgIsUsageError(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"new"}, &stdout, &stderr)
	if code == 0 {
		t.Fatalf("run(new) with no name should fail, exit = 0")
	}
}

// --- main() usage / unknown subcommand ----------------------------------------

func TestRun_UnknownSubcommand(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"bogus"}, &stdout, &stderr)
	if code == 0 {
		t.Fatalf("run(bogus) should fail, exit = 0")
	}
}

func TestRun_NoArgsIsUsageError(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run(nil, &stdout, &stderr)
	if code == 0 {
		t.Fatalf("run(nil) should fail, exit = 0")
	}
}

// --- test helpers --------------------------------------------------------------

// singleBundleFS returns an fs.FS rooted at parent that exposes exactly one
// top-level directory (name), so validateDir (which walks its root looking
// for candidate bundle dirs, exactly like engine.LoadAll) sees only that one
// bundle and not any of parent's other sibling seeded-invalid fixtures.
func singleBundleFS(t *testing.T, parent, name string) fs.FS {
	t.Helper()
	return onlyDirFS{FS: os.DirFS(parent), name: name}
}

func assertFindingRule(t *testing.T, report Report, connector, rule string) {
	t.Helper()
	for _, f := range report.Findings {
		if f.Connector == connector && f.Rule == rule {
			return
		}
	}
	t.Fatalf("no finding for connector %q with rule %q; findings=%+v", connector, rule, report.Findings)
}

func cliSurfaceBundleFS(cliSurface string) fstest.MapFS {
	return fstest.MapFS{
		"cli-surface/metadata.json": &fstest.MapFile{Data: []byte(`{
			"name": "cli-surface",
			"display_name": "CLI Surface",
			"description": "test connector",
			"integration_type": "api",
			"release_stage": "ga",
			"capabilities": { "check": true, "read": true, "write": true, "query": false, "cdc": false, "dynamic_schema": false }
		}`)},
		"cli-surface/spec.json": &fstest.MapFile{Data: []byte(`{
			"$schema": "http://json-schema.org/draft-07/schema#",
			"type": "object",
			"required": ["base_url"],
			"properties": {
				"base_url": { "type": "string" }
			}
		}`)},
		"cli-surface/streams.json": &fstest.MapFile{Data: []byte(`{
			"base": {
				"url": "{{ config.base_url }}",
				"check": { "method": "GET", "path": "/widgets" }
			},
			"streams": [
				{ "name": "widgets", "path": "/widgets", "records": { "path": "data" }, "schema": "schemas/widgets.json" }
			]
		}`)},
		"cli-surface/writes.json": &fstest.MapFile{Data: []byte(`{
			"actions": [
				{
					"name": "create_widget",
					"kind": "create",
					"method": "POST",
					"path": "/widgets",
					"record_schema": { "type": "object", "required": ["name"], "properties": { "name": { "type": "string" } } },
					"risk": "creates a widget"
				}
			]
		}`)},
		"cli-surface/api_surface.json": &fstest.MapFile{Data: []byte(`{
			"api": "test API v1",
			"endpoints": [
				{ "method": "GET", "path": "/widgets", "covered_by": { "stream": "widgets" } },
				{ "method": "POST", "path": "/widgets", "covered_by": { "write": "create_widget" } }
			]
		}`)},
		"cli-surface/schemas/widgets.json": &fstest.MapFile{Data: []byte(`{
			"$schema": "http://json-schema.org/draft-07/schema#",
			"type": "object",
			"x-primary-key": ["id"],
			"properties": {
				"id": { "type": "integer" },
				"name": { "type": "string" }
			}
		}`)},
		"cli-surface/docs.md": &fstest.MapFile{Data: []byte(`# Overview

test

## Auth setup

none

## Streams notes

none

## Write actions & risks

none

## Known limits

none
`)},
		"cli-surface/cli_surface.json": &fstest.MapFile{Data: []byte(cliSurface)},
	}
}

func validCLISurfaceJSON() string {
	return `{
		"tagline": "Work with CLI Surface from the command line.",
		"usage": "pm cli-surface <command> [flags]",
		"commands": [
			{
				"path": "widget list",
				"summary": "List widgets",
				"intent": "etl",
				"availability": "implemented",
				"stream": "widgets",
				"source_cli_path": "clis widget list",
				"api_surface": [
					{ "method": "GET", "path": "/widgets" }
				],
				"examples": ["pm cli-surface widget list --json"]
			},
			{
				"path": "widget create",
				"summary": "Create a widget",
				"intent": "reverse_etl",
				"availability": "implemented",
				"write": "create_widget",
				"source_cli_path": "clis widget create",
				"api_surface": [
					{ "method": "POST", "path": "/widgets" }
				],
				"risk": "creates a widget",
				"approval": "reverse ETL writes require plan, preview, approval, execute",
				"examples": ["pm cli-surface widget create --json"]
			}
		]
	}`
}

// onlyDirFS wraps an fs.FS and restricts ReadDir(".") to a single named
// entry, while passing every other operation straight through (so reads
// inside name/... still resolve normally).
type onlyDirFS struct {
	fs.FS
	name string
}

func (o onlyDirFS) ReadDir(dir string) ([]fs.DirEntry, error) {
	entries, err := fs.ReadDir(o.FS, dir)
	if err != nil {
		return nil, err
	}
	if dir != "." {
		return entries, nil
	}
	var out []fs.DirEntry
	for _, e := range entries {
		if e.Name() == o.name {
			out = append(out, e)
		}
	}
	return out, nil
}
