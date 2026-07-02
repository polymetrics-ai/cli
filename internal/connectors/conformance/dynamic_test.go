package conformance

import (
	"testing"

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
