package conformance

import (
	"strings"
	"testing"
	"testing/fstest"

	"polymetrics.ai/internal/connectors/engine"
)

// --- dynamic checks: good bundle -----------------------------------------

func TestDynamicChecks_GoodBundleAllPass(t *testing.T) {
	b := loadTestBundle(t, "testdata/good", "acme")
	checks := runDynamicChecks(b)
	if len(checks) == 0 {
		t.Fatalf("runDynamicChecks returned zero checks")
	}
	for _, c := range checks {
		if !c.Passed && !c.Skipped {
			t.Errorf("dynamic check %s failed on good bundle: %s", c.Name, c.Error)
		}
	}
}

func TestPaginationTerminates_EachFixturePageServedExactlyOnce(t *testing.T) {
	b := loadTestBundle(t, "testdata/good", "acme")
	tracker := newHitTracker()
	result := checkPaginationTerminates(b, tracker)
	if !result.Passed {
		t.Fatalf("pagination_terminates failed: %s", result.Error)
	}
	// widgets has exactly 2 fixture pages (page_1 with 1 record, page_2
	// short/empty to stop); each must be served exactly once — no page
	// fetched twice, no infinite loop.
	hits := tracker.hitsFor("widgets")
	if hits != 2 {
		t.Fatalf("widgets fixture pages served %d times, want exactly 2 (one per page_N.json)", hits)
	}
}

func TestRecordsMatchSchema_FailsOnSeededTypeDrift(t *testing.T) {
	b := loadTestBundleFromDrift(t)
	result := checkRecordsMatchSchema(b)
	if result.Passed {
		t.Fatalf("records_match_schema passed despite seeded type drift (id as string, schema wants integer)")
	}
	if result.Error == "" {
		t.Fatalf("failing records_match_schema has no Error message")
	}
}

func TestRecordsMatchSchema_PassesOnGoodBundle(t *testing.T) {
	b := loadTestBundle(t, "testdata/good", "acme")
	result := checkRecordsMatchSchema(b)
	if !result.Passed {
		t.Fatalf("records_match_schema failed on good bundle: %s", result.Error)
	}
}

func TestCursorAdvances_PostReadCursorIsMaxFixtureCursorAndReReadSendsParam(t *testing.T) {
	b := loadTestBundle(t, "testdata/good", "acme")
	result := checkCursorAdvances(b)
	if !result.Passed {
		t.Fatalf("cursor_advances failed: %s", result.Error)
	}
}

// TestCursorAdvances_NumericCursorFieldSupported is the B2 regression test
// (REVIEW.md): a fixture whose incremental cursor field is a JSON NUMBER on
// the wire (Stripe's real `created` shape, decoded as json.Number since the
// engine's JSON decoders use UseNumber throughout) must be recognized by
// checkCursorAdvances exactly like a string cursor is — before the fix,
// dynamic.go's checkCursorAdvances only recognized a cursor value via a Go
// `string` type assertion, so a numeric cursor field silently produced "no
// cursor value observed across fixture records" even though two fixture
// records plainly carry increasing `created` values.
func TestCursorAdvances_NumericCursorFieldSupported(t *testing.T) {
	b := loadTestBundle(t, "testdata/good", "acme-numeric-cursor")
	result := checkCursorAdvances(b)
	if !result.Passed {
		t.Fatalf("cursor_advances failed on a numeric (json.Number) cursor field: %s", result.Error)
	}
}

// TestCursorAdvances_StringCursorFieldStillSupported locks in that the
// pre-existing string-cursor shape (param_format: rfc3339, x-cursor-field a
// JSON string) remains legal and unaffected by numeric-cursor support — both
// shapes are legal per B2's fix instructions.
func TestCursorAdvances_StringCursorFieldStillSupported(t *testing.T) {
	b := loadTestBundle(t, "testdata/good", "acme")
	result := checkCursorAdvances(b)
	if !result.Passed {
		t.Fatalf("cursor_advances failed on a string cursor field (regression): %s", result.Error)
	}
}

// TestCursorAdvances_GitHubDateRangeNumericCursorNormalized is the N1
// regression test (wave0 REVIEW.md re-review "New findings"): a
// param_format:github_date_range stream whose max-observed cursor is a bare
// digit string (the Unix-seconds app-persisted cursor shape,
// internal/app/sync_modes.go recordCursor -> toComparableString) must have
// its re-read request_param assertion normalized (digits -> RFC3339) exactly
// like the real engine's formatParam does — before the fix,
// formatCursorForAssertion's github_date_range branch returned
// ">=" + value VERBATIM (i.e. ">=1700000100", the raw digit string), which
// never matches what engine.Read actually sends on the wire
// (">=2023-11-14T22:15:00Z"), so this check falsely FAILS.
func TestCursorAdvances_GitHubDateRangeNumericCursorNormalized(t *testing.T) {
	b := loadTestBundle(t, "testdata/good", "acme-github-range-cursor")
	result := checkCursorAdvances(b)
	if !result.Passed {
		t.Fatalf("cursor_advances failed on github_date_range with a numeric (Unix-seconds) cursor: %s", result.Error)
	}
}

// TestCursorAdvances_GitHubDateRangeNonUTCOffsetCursorNormalized is N1's
// second regression shape: a param_format:github_date_range stream whose
// max-observed cursor is an RFC3339 STRING with a non-UTC offset (+05:30).
// The engine's formatParam parses it and re-emits UTC second precision
// (">=" + t.UTC().Format(time.RFC3339)); formatCursorForAssertion's
// pre-fix verbatim passthrough would instead assert against the
// un-normalized offset form (">=2023-11-14T22:15:00+05:30"), which never
// matches the engine's actual UTC-normalized request.
func TestCursorAdvances_GitHubDateRangeNonUTCOffsetCursorNormalized(t *testing.T) {
	b := loadTestBundle(t, "testdata/good", "acme-github-range-cursor-offset")
	result := checkCursorAdvances(b)
	if !result.Passed {
		t.Fatalf("cursor_advances failed on github_date_range with a non-UTC-offset RFC3339 cursor: %s", result.Error)
	}
}

func TestWriteRequestShape_MatchesExpectBlock(t *testing.T) {
	b := loadTestBundle(t, "testdata/good", "acme")
	results := checkWriteRequestShape(b)
	if len(results) == 0 {
		t.Fatalf("checkWriteRequestShape returned no results for a bundle with fixtures/writes/*")
	}
	for _, r := range results {
		if !r.Passed {
			t.Errorf("write_request_shape %s failed: %s", r.Name, r.Error)
		}
	}
}

func TestWriteRequestShape_MismatchFails(t *testing.T) {
	b := loadTestBundleFromWriteMismatch(t)
	results := checkWriteRequestShape(b)
	found := false
	for _, r := range results {
		if r.Name == "write_request_shape:update_widget" {
			found = true
			if r.Passed {
				t.Fatalf("expected write_request_shape mismatch to fail")
			}
		}
	}
	if !found {
		t.Fatalf("expected a write_request_shape:update_widget result, got %+v", results)
	}
}

func TestDeleteSemantics_MissingOkStatusHandledAsWritten(t *testing.T) {
	b := loadTestBundle(t, "testdata/good", "acme")
	result := checkDeleteSemantics(b)
	if !result.Passed {
		t.Fatalf("delete_semantics failed: %s", result.Error)
	}
}

func TestCheckFixture_AndReadFixtureNonempty(t *testing.T) {
	b := loadTestBundle(t, "testdata/good", "acme")
	checkFixtureResult := checkCheckFixture(b)
	if !checkFixtureResult.Passed {
		t.Fatalf("check_fixture failed: %s", checkFixtureResult.Error)
	}
	nonEmpty := checkReadFixtureNonempty(b, "widgets", true)
	if !nonEmpty.Passed {
		t.Fatalf("read_fixture_nonempty(widgets) failed: %s", nonEmpty.Error)
	}
}

func TestReusableStreamReplayServerResetsBetweenStreams(t *testing.T) {
	b := loadTestBundle(t, "testdata/good", "acme")
	replay := newReusableStreamReplayServer()
	defer replay.Close()

	var widgets int
	if err := readRawRecordsWithReplay(b, "widgets", nil, replay, func(map[string]any) error {
		widgets++
		return nil
	}); err != nil {
		t.Fatalf("read widgets: %v", err)
	}
	if widgets == 0 {
		t.Fatalf("widgets emitted zero records")
	}
	replayURL := replay.URL

	var notes int
	if err := readRawRecordsWithReplay(b, "notes", nil, replay, func(map[string]any) error {
		notes++
		return nil
	}); err != nil {
		t.Fatalf("read notes: %v", err)
	}
	if notes == 0 {
		t.Fatalf("notes emitted zero records")
	}
	if replay.URL != replayURL {
		t.Fatalf("replay server URL changed between streams: %q -> %q", replayURL, replay.URL)
	}
}

func TestReadRawRecordsWithReplayUsesFixtureReadQuery(t *testing.T) {
	b := engine.Bundle{
		Name: "acme",
		HTTP: engine.HTTPBase{URL: "http://placeholder.invalid"},
		Streams: []engine.StreamSpec{
			{
				Name:   "widgets",
				Method: "POST",
				Path:   "/graphql",
				GraphQL: &engine.GraphQLRequestSpec{
					Document:      "query Widget($number: Int!) { widget(number: $number) { id } }",
					OperationName: "Widget",
					Variables: map[string]any{
						"number": map[string]any{"template": "{{ query.number }}", "type": "integer"},
					},
				},
				Records:   engine.RecordsSpec{Path: "data.widget", SingleObject: true},
				SchemaRef: "schemas/widgets.json",
			},
		},
		Fixtures: fstest.MapFS{
			"streams/widgets/page_1.json": &fstest.MapFile{Data: []byte(`{
				"request": { "method": "POST", "path": "/graphql", "query": {} },
				"read_query": { "number": "7" },
				"response": { "status": 200, "body": { "data": { "widget": { "id": "w7" } } } }
			}`)},
		},
	}

	var records []map[string]any
	replay := newReusableStreamReplayServer()
	defer replay.Close()
	err := readRawRecordsWithReplay(b, "widgets", nil, replay, func(raw map[string]any) error {
		records = append(records, raw)
		return nil
	})
	if err != nil {
		t.Fatalf("readRawRecordsWithReplay: %v", err)
	}
	if len(records) != 1 || records[0]["id"] != "w7" {
		t.Fatalf("records = %+v, want one fixture-backed GraphQL record", records)
	}
}

// --- ENGINE DIALECT ADDITION (checkquery-ledger.md item 5): check_fixture
// must compare the fixture's recorded request.query, not just replay a
// canned response to any request whatsoever. --------------------------------

// TestCheckFixture_FailsWhenBundleSendsQueryFixtureDidNotRecord is the exact
// scenario the ledger's item 5 names: streams.json now declares
// base.check.query (RequestSpec.Query), but fixtures/check.json was recorded
// before that field existed (or was never updated) and carries no "request"
// field at all — so it implicitly expects no query string. Once Check()
// actually sends "?limit=1", the replay server must NOT match the fixture,
// and check_fixture must fail (not silently pass by ignoring the query the
// way it did pre-hardening).
func TestCheckFixture_FailsWhenBundleSendsQueryFixtureDidNotRecord(t *testing.T) {
	b := loadTestBundle(t, "testdata/dynamic-invalid", "check-query-mismatch")
	if b.HTTP.Check == nil || len(b.HTTP.Check.Query) == 0 {
		t.Fatalf("test bundle must declare base.check.query; got %+v", b.HTTP.Check)
	}
	result := checkCheckFixture(b)
	if result.Passed {
		t.Fatalf("check_fixture passed despite fixtures/check.json recording no query while the bundle now sends one")
	}
	if result.Error == "" {
		t.Fatalf("failing check_fixture has no Error message")
	}
}

// TestCheckFixture_PassesWhenFixtureRecordsMatchingQuery proves the positive
// case: a fixtures/check.json that DOES record the expected "request.query"
// matching what Check() actually sends passes cleanly — check_fixture is not
// merely stricter, it is CORRECTLY comparing the query, not just always
// failing once any check.query is declared.
func TestCheckFixture_PassesWhenFixtureRecordsMatchingQuery(t *testing.T) {
	b := loadTestBundle(t, "testdata/good", "acme-check-query")
	result := checkCheckFixture(b)
	if !result.Passed {
		t.Fatalf("check_fixture failed on a fixture whose recorded request.query matches base.check.query: %s", result.Error)
	}
}

// --- conformance skip markers (R3: hook-aware dynamic conformance) --------
//
// A bundle may declare an OPTIONAL, explicit "conformance": {"skip_dynamic":
// true, "reason": "..."} marker at stream level (streams.json) or bundle
// level (metadata.json), for a connector whose real behavior lives entirely
// behind a Tier-2 hook that a declarative fixture replay cannot exercise
// (e.g. monday's GraphQL StreamHook, gmail's custom-auth-only AuthHook).
// dynamic.go must Skip (never Pass/Fail) the affected checks and surface the
// marker's reason as the CheckResult's Error text, so a report reader always
// sees WHY a check didn't run and what proves the behavior instead.

// TestDynamicChecks_StreamLevelSkipMarkerSkipsReadCheckWithReason: a stream
// whose fixture is deliberately unreplayable via the declarative path (see
// testdata/good/acme-stream-marker) but carries a skip_dynamic marker must
// report read_fixture_nonempty:widgets as Skipped with the marker's reason,
// never as a failure.
func TestDynamicChecks_StreamLevelSkipMarkerSkipsReadCheckWithReason(t *testing.T) {
	b := loadTestBundle(t, "testdata/good", "acme-stream-marker")
	checks := runDynamicChecks(b)

	result := mustFindCheck(t, checks, "read_fixture_nonempty:widgets")
	if !result.Skipped {
		t.Fatalf("read_fixture_nonempty:widgets Skipped = false, want true (marker should skip, not fail): %+v", result)
	}
	if result.Passed {
		t.Fatalf("read_fixture_nonempty:widgets Passed = true, want false (a Skipped check is never also Passed)")
	}
	if result.Error == "" {
		t.Fatalf("read_fixture_nonempty:widgets Skipped but carries no reason text")
	}
	if !strings.Contains(result.Error, "archived parity evidence for acme-stream-marker") {
		t.Fatalf("read_fixture_nonempty:widgets Error %q does not name the authoritative substitute", result.Error)
	}

	// The unmarked sibling stream (notes) must still run normally (Passed),
	// proving the marker is scoped to the ONE marked stream, not a
	// connector-wide bypass.
	notes := mustFindCheck(t, checks, "read_fixture_nonempty:notes")
	if !notes.Passed {
		t.Fatalf("read_fixture_nonempty:notes failed on the unmarked sibling stream: %+v", notes)
	}
}

// TestDynamicChecks_StreamLevelMarkerExcludesStreamFromPaginationAndCursorChecks
// asserts that a marked FIRST stream (widgets — otherwise
// pagination_terminates' natural first-stream candidate, since it's first
// with fixtures) is excluded from candidate selection: pagination_terminates
// must not attempt (and therefore not fail against) its deliberately
// unreplayable fixture — it falls through to the next ELIGIBLE stream
// (notes) and Passes against that instead, proving exclusion means "treat as
// if this stream did not exist," not "abort the whole check." cursor_advances
// has no OTHER incremental stream in this bundle (only widgets declares
// `incremental`), so once widgets is excluded there is no candidate left at
// all — it Skips with the marker's reason rather than degrading to the
// pre-existing generic "no incremental stream" Skip.
func TestDynamicChecks_StreamLevelMarkerExcludesStreamFromPaginationAndCursorChecks(t *testing.T) {
	b := loadTestBundle(t, "testdata/good", "acme-stream-marker")
	checks := runDynamicChecks(b)

	pagination := mustFindCheck(t, checks, "pagination_terminates")
	if !pagination.Passed {
		t.Fatalf("pagination_terminates Passed = false, want true (should fall through to the next eligible stream, notes, not attempt the marker-excluded widgets): %+v", pagination)
	}

	cursor := mustFindCheck(t, checks, "cursor_advances")
	if !cursor.Skipped {
		t.Fatalf("cursor_advances Skipped = false, want true (widgets is the only incremental stream and is marker-excluded): %+v", cursor)
	}
	if cursor.Error == "" {
		t.Fatalf("cursor_advances Skipped but carries no reason text")
	}
}

// TestDynamicChecks_AllStreamsMarkedSkipsPaginationTerminatesWithReason is the
// companion case: when EVERY declared stream carries its OWN per-stream
// skip_dynamic marker (as opposed to acme-bundle-marker's single
// bundle-level marker, a different code path entirely — see
// runDynamicChecks), pagination_terminates has no eligible candidate stream
// left at all and must Skip with the (first) marker's reason, not silently
// fall back to the pre-existing no-reason "no streams" Skip.
func TestDynamicChecks_AllStreamsMarkedSkipsPaginationTerminatesWithReason(t *testing.T) {
	b := loadTestBundle(t, "testdata/good", "acme-all-streams-marked")
	result := checkPaginationTerminates(b, newHitTracker())
	if !result.Skipped {
		t.Fatalf("pagination_terminates Skipped = false, want true (every stream is marker-excluded): %+v", result)
	}
	if result.Error == "" {
		t.Fatalf("pagination_terminates Skipped but carries no reason text")
	}
}

// TestDynamicChecks_BundleLevelSkipMarkerSkipsAuthDependentChecks: a
// bundle-level marker (testdata/good/acme-bundle-marker — every stream
// fixture AND check.json deliberately unreplayable) must Skip check_fixture,
// every read_fixture_nonempty:*, pagination_terminates, and
// records_match_schema, all carrying the marker's reason — none may Fail.
func TestDynamicChecks_BundleLevelSkipMarkerSkipsAuthDependentChecks(t *testing.T) {
	b := loadTestBundle(t, "testdata/good", "acme-bundle-marker")
	checks := runDynamicChecks(b)

	for _, name := range []string{
		"check_fixture",
		"read_fixture_nonempty:widgets",
		"read_fixture_nonempty:notes",
		"pagination_terminates",
		"records_match_schema",
		"cursor_advances",
	} {
		result := mustFindCheck(t, checks, name)
		if !result.Skipped {
			t.Errorf("%s Skipped = false, want true (bundle-level marker): %+v", name, result)
		}
		if result.Passed {
			t.Errorf("%s Passed = true, want false (a Skipped check is never also Passed)", name)
		}
		if result.Error == "" {
			t.Errorf("%s Skipped but carries no reason text", name)
		}
	}
}

// TestDynamicChecks_UnmarkedHookFailureStillFails proves the marker
// mechanism does not accidentally widen into a blanket bypass: the
// pre-existing negative fixtures (no marker present at all) must still
// report a hard failure, exactly as before this feature existed.
func TestDynamicChecks_UnmarkedHookFailureStillFails(t *testing.T) {
	b := loadTestBundleFromDrift(t)
	result := checkRecordsMatchSchema(b)
	if result.Passed || result.Skipped {
		t.Fatalf("records_match_schema on an unmarked schema-drift bundle: Passed=%v Skipped=%v, want a hard failure", result.Passed, result.Skipped)
	}
	if result.Error == "" {
		t.Fatalf("failing records_match_schema has no Error message")
	}
}

func mustFindCheck(t *testing.T, checks []CheckResult, name string) CheckResult {
	t.Helper()
	for _, c := range checks {
		if c.Name == name {
			return c
		}
	}
	t.Fatalf("no check named %q in %+v", name, checks)
	return CheckResult{}
}

// --- helpers: alternate bundles for negative dynamic cases ---------------

// loadTestBundleFromDrift loads a bundle whose fixture record has a
// schema-violating type (id as a string where the schema declares integer),
// used to exercise records_match_schema's negative case in isolation from
// the static-check invalid corpus.
func loadTestBundleFromDrift(t *testing.T) engine.Bundle {
	t.Helper()
	return loadTestBundle(t, "testdata/dynamic-invalid", "schema-drift")
}

// loadTestBundleFromWriteMismatch loads a bundle whose fixtures/writes/*.json
// "expect" block does not match what the engine actually produces.
func loadTestBundleFromWriteMismatch(t *testing.T) engine.Bundle {
	t.Helper()
	return loadTestBundle(t, "testdata/dynamic-invalid", "write-mismatch")
}
