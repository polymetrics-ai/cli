# T/B-13 — conformance v2

Phase: wave0-engine-harness · Wave: E · Executor: gsd-loop-backend (sonnet, with tester duties)

## Scope

New package `internal/connectors/conformance/`:

- `conformance.go` — `Report`/`CheckResult` shape, `RunBundle`, `ReportFromLoadError`, `TestConformance`
- `static.go` — 10 static checks per design §E.2
- `replay.go` — fixture-backed `httptest.Server` replay builder (`fixtures/streams/<stream>/page_N.json`)
- `dynamic.go` — 8 dynamic checks running the REAL engine against the replay server
- `conformance_test.go`, `static_test.go`, `dynamic_test.go`, `replay_test.go`
- `testdata/good/acme/**` (own control bundle, NOT shared with `cmd/connectorgen/testdata`)
- `testdata/invalid/**` (10 seeded static-defect bundles, one per static check)
- `testdata/dynamic-invalid/**` (2 seeded dynamic-defect bundles: schema drift, write mismatch)

`internal/connectors/native_conformance.go` (legacy) read-only reference, untouched.

## RED evidence

Test files authored first (`conformance_test.go`, `static_test.go`, `dynamic_test.go`,
`replay_test.go`) plus the full `testdata/` corpus (good bundle + invalid corpus + dynamic-invalid
corpus), against a package containing ONLY `testdata/` — no `.go` production files yet:

```
$ go vet ./internal/connectors/conformance/...
# polymetrics.ai/internal/connectors/conformance
# [polymetrics.ai/internal/connectors/conformance]
vet: internal/connectors/conformance/conformance_test.go:132:39: undefined: Report
```

`Report`, `CheckResult`, `RunBundle`, `ReportFromLoadError`, `runStaticChecks`, `staticCheckNames`,
`runDynamicChecks`, `checkPaginationTerminates`, `checkRecordsMatchSchema`, `checkCursorAdvances`,
`checkWriteRequestShape`, `checkDeleteSemantics`, `checkCheckFixture`, `checkReadFixtureNonempty`,
`newHitTracker`, `newStreamReplayServer`, `loadFixturePages`, `jsonMarshal` did not exist anywhere
in the tree before this task — the compile failure is the correct RED signal for a brand-new
package built test-first.

Corpus validated pre-implementation with scratch `engine.Load` probes (removed before RED capture,
not committed): `bad-spec-schema`/`bad-stream-schema` fail `engine.Load` itself (meta-schema/compile
errors); the other 8 invalid-corpus bundles + both dynamic-invalid bundles load successfully via
`engine.Load` (defects are semantic, not structural) — confirming each seeded bundle exercises
exactly the intended check rather than accidentally tripping the loader.

## Test coverage authored RED-first

`conformance_test.go`:
- `TestReportJSONShape` / `TestReportPassedTrueWhenNoFailures` — `{connector, checks:
  [{name,passed,skipped,error}], passed}` shape + roll-up semantics (`computePassed`: any failing,
  non-skipped check ⇒ `Passed=false`).
- `TestConformance` — iterates `engine.LoadAll(defs.FS)`, one `t.Run(bundle.Name, ...)` subtest per
  bundle (exact pattern later waves depend on: `-run 'TestConformance/<name>'`). Zero bundles today
  (defs/ ships empty until Wave F) ⇒ trivially passes (no subtests registered, no failures).
- `TestConformance_EmptyDefsTreePassesTrivially` — explicit belt-and-suspenders lock on the
  "empty tree passes" contract.
- `TestRunBundle_GoodBundlePassesEveryCheck` — the own `testdata/good/acme` control bundle: every
  static AND dynamic check present and passing; spot-asserts representative check names from both
  categories are present in the report (`spec_schema_valid`, `docs_present`, `fixtures_present`,
  `pagination_terminates`, `records_match_schema`, `cursor_advances`, `write_request_shape`,
  `delete_semantics`).
- `TestStaticChecks_TargetedFailures` — table-driven, 10 cases (one per static check), each loads a
  single seeded-invalid bundle (own corpus, `testdata/invalid/**`) and asserts: `Report.Passed ==
  false`, the NAMED check is present and failing, and its `Error` is non-empty. Corpus size assertion
  `len(cases) >= 10` locks the "every static check has a failing negative case" EVAL-PLAN §4 metric
  (10/10).

`static_test.go`:
- `TestRunStaticChecks_GoodBundleAllPass` — `runStaticChecks` alone (no dynamic/replay) returns
  exactly `len(staticCheckNames)` results, all passing, on the good bundle.
- `TestStaticCheckNames_MatchDesignList` — locks the exact design §E.2 static check name list and
  order.
- `TestReportFromLoadError_ClassifiesMetaSchemaFailure` / `..._SkipsRemainingChecks` — a bundle that
  fails `engine.Load` itself (spec.json meta-schema/compile violation) still produces a full report:
  the specific failing check is named (`spec_schema_valid`), and every check that needs a loaded
  bundle to run is marked `Skipped` (never silently absent).
- `TestInvalidCorpus_DirsExist` — sanity guard the 10 seeded directories exist on disk.

`dynamic_test.go`:
- `TestDynamicChecks_GoodBundleAllPass` — `runDynamicChecks` on the good bundle: every result
  passing or explicitly skipped, none failing.
- `TestPaginationTerminates_EachFixturePageServedExactlyOnce` — replay-server hit-tracker asserts
  the `widgets` stream's 2 fixture pages (`page_1.json` 1-record, `page_2.json` empty/stop) are each
  served EXACTLY once by a full engine `Read` (bounded page count, no infinite loop, no duplicate
  fetch) — the design §E.2 "multi-page fixture consumed exactly once" requirement, verified by an
  actual counter rather than just "Read returned no error".
- `TestRecordsMatchSchema_FailsOnSeededTypeDrift` / `..._PassesOnGoodBundle` — a dedicated
  `testdata/dynamic-invalid/schema-drift` bundle (widgets `id` fixture value is a string; schema
  declares `integer`) fails with a non-empty `Error`; the good bundle passes.
- `TestCursorAdvances_PostReadCursorIsMaxFixtureCursorAndReReadSendsParam` — full read against
  fixtures, then a second read seeded with the resulting cursor state, asserting the re-read request
  actually carries `since=<cursor>` (the declared `incremental.request_param`).
- `TestWriteRequestShape_MatchesExpectBlock` / `..._MismatchFails` — `checkWriteRequestShape` runs
  the real engine's write-request resolution (method/path/body) against every
  `fixtures/writes/<action>.json`'s `expect` block; a dedicated
  `testdata/dynamic-invalid/write-mismatch` bundle (deliberately wrong `expect.path`) fails with a
  named, non-empty-error result keyed `write_request_shape:<action>`.
- `TestDeleteSemantics_MissingOkStatusHandledAsWritten` — `delete_widget`'s fixture-driven dry
  run/execute against a 404 response is treated as written-not-failed per `missing_ok_status`.
- `TestCheckFixture_AndReadFixtureNonempty` — `check_fixture` (declarative Check() against a replay
  server) and `read_fixture_nonempty` (first-stream-mandatory nonempty read) both pass on the good
  bundle.

`replay_test.go`:
- `TestNewStreamReplayServer_ServesPagesInOrderExactlyOnce` / `..._UnmatchedRequestIs404` — the
  fixture replay server itself: matches recorded `page_N.json` request shape (method+path+query),
  serves the recorded response, 404s on anything unmatched, and the shared `hitTracker` counts
  per-stream hits independently of which specific dynamic check is driving the read.
- `TestLoadFixturePages_ParsesRequestAndResponse` / `..._MissingStreamReturnsEmpty` — fixture-file
  parsing in isolation from the HTTP server.

## Implementation plan (before writing code)

- `Report{Connector string; Checks []CheckResult; Passed bool}`,
  `CheckResult{Name string; Passed bool; Skipped bool; Error string}` — JSON tags `connector`,
  `checks`, `passed`, `name`, `passed`, `skipped,omitempty`, `error,omitempty`. Mirrors
  `native_conformance.go`'s `{Name, Passed, Error}` shape plus a `Skipped` bit (needed because a
  Load failure must still report every OTHER check as skipped rather than omitting them) and a
  `Passed` rollup at both `CheckResult`-list and `Report` level.
- `RunBundle(b engine.Bundle) Report` = `runStaticChecks(b)` ++ `runDynamicChecks(b)`, rolled up via
  `computePassed`.
- `ReportFromLoadError(name string, err error) Report`: classifies the Load error into the specific
  static check name using the SAME string-based classification `cmd/connectorgen/validate.go`
  already uses for `loadErrorFinding` (kept independent — no cross-package import, this package's
  own copy — per the "no cross-package sharing" note in PLAN.md's corpus split), then marks every
  other static+dynamic check name `Skipped:true`.
- `staticCheckNames` — the fixed, ordered §E.2 list; `runStaticChecks` composes 10 check funcs, one
  per name, mirroring `cmd/connectorgen/validate.go`'s check functions closely (same rules, but
  in-package — the brief explicitly allows overlap: "conformance may overlap with connectorgen
  validate, that's fine, but implement in-package so conformance is self-contained").
- `replay.go`: fixture file shape `{"request":{"method","path","query"},"response":{"status","body"}}`
  — locked from the ALREADY-COMMITTED `internal/connectors/engine/testdata/bundles/widget-demo/
  fixtures/streams/widgets/page_1.json` (Wave A/B reference, read-only). `newStreamReplayServer`
  loads every `page_N.json` under `fixtures/streams/<stream>/`, sorted by filename (`page_1` before
  `page_2`, ...), and serves an `httptest.Server` that on each request finds the first
  not-yet-served fixture whose recorded method+path+query matches the incoming request, replays its
  response, and marks it served (never serves the same fixture file twice) — a `hitTracker` records
  per-stream hit counts for `pagination_terminates`' "each page served exactly once" assertion. An
  unmatched request → 404 (surfaces as a normal `*connsdk.HTTPError` to the engine, which is itself
  a useful conformance signal if the bundle's declared paging query doesn't match its own fixtures).
- `dynamic.go` checks run the REAL `engine.Read`/`engine.Check`/`engine.ValidateWrite`/
  `engine.DryRunWrite` against a `newStreamReplayServer`-backed `Bundle` copy whose `HTTP.URL` is
  swapped to the replay server's URL (a shallow bundle copy with `HTTP.URL` overridden — `Bundle` is
  a value type per `bundle.go`, so this is a plain struct-literal copy, no mutation of the loaded
  bundle).
- `write_request_shape`/`delete_semantics` do NOT perform a live network call against a real
  external API even in "dry" mode per THREAT-MODEL: they run against the SAME replay server
  (fixtures/writes/<action>.json's expected request is asserted by intercepting what the replay
  server actually received, not by trusting `DryRunWrite`'s redacted preview string alone) — decided
  during implementation; see Notes if this diverges.

## Implementation notes (as built)

- `conformance.go`: `Report{Connector, Checks []CheckResult, Passed}` / `CheckResult{Name, Passed,
  Skipped, Error}` built exactly as planned. `RunBundle` = `runStaticChecks` ++ `runDynamicChecks`,
  rolled up via `Report.computePassed()` (a `Skipped` check never affects `Passed`). `jsonMarshal` is
  a one-line indirection kept purely so `conformance_test.go` has a package-local name to call.
- `static.go`: `staticCheckNames` is the fixed §E.2-ordered list; `runStaticChecks` composes 10 check
  functions almost identical in spirit to `cmd/connectorgen/validate.go`'s (`checkAPISurface`,
  `checkDocsHeadings`, `checkFixtureSecrets`, etc.) but reimplemented in-package per the brief's
  explicit "may overlap, implement in-package so conformance is self-contained" instruction — no
  import of `cmd/connectorgen` (that would be an inverted, disallowed dependency direction anyway:
  `cmd/*` importing back from a library package is fine, the reverse is not attempted here regardless).
  `classifyLoadError` is this package's OWN copy of the same string-classification idea
  `loadErrorFinding` uses (independently written, not shared code) — needed because
  `spec_schema_valid`/`stream_schemas_valid` can only fail via `engine.Load` itself erroring (both are
  already-enforced-at-Load invariants once a bundle loads), so their FAILING case is only reachable
  through `ReportFromLoadError`, never through `runStaticChecks` on a loaded bundle.
- `checkFixturesPresent`: "first stream mandatory" per design §E.2 — a bundle with streams but zero
  fixture pages under `fixtures/streams/<first-stream>/` fails; other (non-first) streams may
  legitimately ship no fixtures. A bundle with zero streams (Tier-3 dynamic-schema natives, out of
  wave0 scope but structurally possible) trivially passes.
- `replay.go`: fixture page shape locked from the ALREADY-COMMITTED
  `internal/connectors/engine/testdata/bundles/widget-demo/fixtures/streams/widgets/page_1.json`
  (`{"request":{"method","path","query"},"response":{"status","body"}}`), reused verbatim rather than
  inventing a new shape. `newStreamReplayServer` serves pages in filename order, matching the first
  not-yet-served page whose recorded method+path+query is a subset-match of the incoming request
  (extra incoming query params not in the fixture are tolerated — a spec's own optional params
  shouldn't force every fixture author to spell out every param); an unmatched request is a 404,
  which becomes a normal `*connsdk.HTTPError` bubbling up through `engine.Read`/`Check` — itself a
  useful, honest failure signal rather than a silent skip. `hitTracker` (mutex-guarded map) records
  per-stream serve counts for `pagination_terminates`'s "exactly one request per fixture page"
  assertion. Added (beyond the original plan) `loadCheckFixture`/`newCheckReplayServer` for a
  DEDICATED `fixtures/check.json` file (`{"response":{"status","body"}}`, no "request" field needed
  since `Check()` always issues exactly one request to one declared path) — the original plan of
  reusing the first stream's replay server for `check_fixture` turned out to 404 (the check path
  `/ping` never matches a stream's recorded `/widgets` or `/notes` request), so a small dedicated
  fixture file was the correct fix rather than a workaround; documented here as the one concrete
  deviation from the pre-code plan.
- `dynamic.go`: every dynamic check builds a synthetic `connectors.RuntimeConfig` via
  `runtimeConfigForEngine` (every spec-declared property gets a synthetic value; `x-secret` properties
  get a synthetic value in `Secrets`, never `Config` — THREAT-MODEL §4, conformance never touches real
  credentials) and a per-check `withReplayURL` bundle copy (a plain struct-field override — `Bundle` is
  a value type per `bundle.go`, so no mutation of the caller's loaded bundle). `checkCursorAdvances`
  runs a full fixture-backed read to find the max observed cursor value, independently
  re-derives the EXPECTED formatted `request_param` value via a small local `formatCursorForAssertion`
  (a deliberate, documented, minimal duplication of read.go's unexported `formatParam` — needed so the
  assertion is derived independently of the code under test, not reaching into engine internals which
  are out of this task's file scope anyway), then re-reads against a dedicated
  `paramCaptureServer` (always answers one empty page so the read terminates after exactly the single
  request this check inspects) and compares the captured query param against that independently
  computed expectation. `checkWriteRequestShape` folds `write_validate` into the same per-action
  result (a fixture record that fails `engine.ValidateWrite` fails the whole
  `write_request_shape:<action>` check, since design §E.2 lists them as one combined item per action)
  and asserts against what a dedicated `captureServer` ACTUALLY received (method/path/decoded JSON
  body), not against `DryRunWrite`'s human-readable preview string — a stronger, more direct assertion
  than string-matching a preview. `checkDeleteSemantics` runs the real `engine.Write` against a
  `newAlwaysStatusServer` that always answers the action's first `missing_ok_status` value and asserts
  `RecordsWritten==1, RecordsFailed==0`.
- All per-stream/per-action dynamic checks that have no applicable target (e.g. `cursor_advances` on a
  bundle with no incremental+fixtured stream, `delete_semantics` on a bundle with no delete action,
  `write_request_shape:<action>` for an action with no `fixtures/writes/<action>.json`) report
  `Skipped: true` rather than being silently absent from the check list, so `Report.Checks` is always
  a stable, explainable shape for any bundle.

## GREEN evidence

`go test ./internal/connectors/conformance -v` (24 top-level tests, incl. the 10-case table-driven
`TestStaticChecks_TargetedFailures` subtest group — 33 total pass/subtest lines):

```
=== RUN   TestReportJSONShape
--- PASS: TestReportJSONShape (0.00s)
=== RUN   TestReportPassedTrueWhenNoFailures
--- PASS: TestReportPassedTrueWhenNoFailures (0.00s)
=== RUN   TestConformance
--- PASS: TestConformance (0.00s)
=== RUN   TestConformance_EmptyDefsTreePassesTrivially
--- PASS: TestConformance_EmptyDefsTreePassesTrivially (0.00s)
=== RUN   TestRunBundle_GoodBundlePassesEveryCheck
--- PASS: TestRunBundle_GoodBundlePassesEveryCheck (0.01s)
=== RUN   TestStaticChecks_TargetedFailures
=== RUN   TestStaticChecks_TargetedFailures/bad-spec-schema
=== RUN   TestStaticChecks_TargetedFailures/bad-stream-schema
=== RUN   TestStaticChecks_TargetedFailures/pk-missing
=== RUN   TestStaticChecks_TargetedFailures/cursor-missing
=== RUN   TestStaticChecks_TargetedFailures/unresolved-interp
=== RUN   TestStaticChecks_TargetedFailures/write-schema-invalid
=== RUN   TestStaticChecks_TargetedFailures/surface-incomplete
=== RUN   TestStaticChecks_TargetedFailures/docs-missing-heading
=== RUN   TestStaticChecks_TargetedFailures/secret-in-fixture
=== RUN   TestStaticChecks_TargetedFailures/no-fixtures
--- PASS: TestStaticChecks_TargetedFailures (0.04s)
    --- PASS: TestStaticChecks_TargetedFailures/bad-spec-schema (0.00s)
    --- PASS: TestStaticChecks_TargetedFailures/bad-stream-schema (0.00s)
    --- PASS: TestStaticChecks_TargetedFailures/pk-missing (0.01s)
    --- PASS: TestStaticChecks_TargetedFailures/cursor-missing (0.01s)
    --- PASS: TestStaticChecks_TargetedFailures/unresolved-interp (0.00s)
    --- PASS: TestStaticChecks_TargetedFailures/write-schema-invalid (0.00s)
    --- PASS: TestStaticChecks_TargetedFailures/surface-incomplete (0.00s)
    --- PASS: TestStaticChecks_TargetedFailures/docs-missing-heading (0.01s)
    --- PASS: TestStaticChecks_TargetedFailures/secret-in-fixture (0.01s)
    --- PASS: TestStaticChecks_TargetedFailures/no-fixtures (0.00s)
=== RUN   TestDynamicChecks_GoodBundleAllPass
--- PASS: TestDynamicChecks_GoodBundleAllPass (0.00s)
=== RUN   TestPaginationTerminates_EachFixturePageServedExactlyOnce
--- PASS: TestPaginationTerminates_EachFixturePageServedExactlyOnce (0.00s)
=== RUN   TestRecordsMatchSchema_FailsOnSeededTypeDrift
--- PASS: TestRecordsMatchSchema_FailsOnSeededTypeDrift (0.00s)
=== RUN   TestRecordsMatchSchema_PassesOnGoodBundle
--- PASS: TestRecordsMatchSchema_PassesOnGoodBundle (0.00s)
=== RUN   TestCursorAdvances_PostReadCursorIsMaxFixtureCursorAndReReadSendsParam
--- PASS: TestCursorAdvances_PostReadCursorIsMaxFixtureCursorAndReReadSendsParam (0.00s)
=== RUN   TestWriteRequestShape_MatchesExpectBlock
--- PASS: TestWriteRequestShape_MatchesExpectBlock (0.00s)
=== RUN   TestWriteRequestShape_MismatchFails
--- PASS: TestWriteRequestShape_MismatchFails (0.00s)
=== RUN   TestDeleteSemantics_MissingOkStatusHandledAsWritten
--- PASS: TestDeleteSemantics_MissingOkStatusHandledAsWritten (0.00s)
=== RUN   TestCheckFixture_AndReadFixtureNonempty
--- PASS: TestCheckFixture_AndReadFixtureNonempty (0.00s)
=== RUN   TestNewStreamReplayServer_ServesPagesInOrderExactlyOnce
--- PASS: TestNewStreamReplayServer_ServesPagesInOrderExactlyOnce (0.00s)
=== RUN   TestNewStreamReplayServer_UnmatchedRequestIs404
--- PASS: TestNewStreamReplayServer_UnmatchedRequestIs404 (0.00s)
=== RUN   TestLoadFixturePages_ParsesRequestAndResponse
--- PASS: TestLoadFixturePages_ParsesRequestAndResponse (0.00s)
=== RUN   TestLoadFixturePages_MissingStreamReturnsEmpty
--- PASS: TestLoadFixturePages_MissingStreamReturnsEmpty (0.00s)
=== RUN   TestRunStaticChecks_GoodBundleAllPass
--- PASS: TestRunStaticChecks_GoodBundleAllPass (0.00s)
=== RUN   TestStaticCheckNames_MatchDesignList
--- PASS: TestStaticCheckNames_MatchDesignList (0.00s)
=== RUN   TestReportFromLoadError_ClassifiesMetaSchemaFailure
--- PASS: TestReportFromLoadError_ClassifiesMetaSchemaFailure (0.00s)
=== RUN   TestReportFromLoadError_SkipsRemainingChecks
--- PASS: TestReportFromLoadError_SkipsRemainingChecks (0.00s)
=== RUN   TestInvalidCorpus_DirsExist
--- PASS: TestInvalidCorpus_DirsExist (0.00s)
PASS
ok  	polymetrics.ai/internal/connectors/conformance	0.506s
```

## Self-verify (all commands from the dispatch, run at the end of the session)

```
$ go build ./...                                                     # clean, no output
$ go vet ./internal/connectors/conformance                           # clean, no output
$ go vet ./...                                                       # clean, no output (whole tree)
$ go test ./... 2>&1 | grep -v "^ok"
?   	polymetrics.ai/cmd/pm	[no test files]
?   	polymetrics.ai/cmd/pm-cataloggen	[no test files]
?   	polymetrics.ai/cmd/registrygen	[no test files]
?   	polymetrics.ai/internal/connectors/defs	[no test files]
?   	polymetrics.ai/internal/connectors/hooks/hookset	[no test files]
?   	polymetrics.ai/internal/connectors/native/nativeset	[no test files]
?   	polymetrics.ai/internal/coordination	[no test files]
                                                                       # every other package: ok
$ gofmt -l internal/connectors/conformance                            # empty output = clean
$ golangci-lint run ./internal/connectors/conformance/...
0 issues.
$ make lint    # repo-wide wave0 lint target, includes this package
golangci-lint run ./internal/connectors/engine/... ./internal/connectors/defs/... \
  ./internal/connectors/hooks/... ./internal/connectors/native/... \
  ./internal/connectors/conformance/... ./internal/connectors/certify/... \
  ./cmd/connectorgen/... ./cmd/inventorygen/...
0 issues.
$ make connectorgen-validate
go run ./cmd/connectorgen validate internal/connectors/defs
connectorgen validate: 0 connector(s) checked, 0 findings
$ go test -cover ./internal/connectors/conformance
ok  	polymetrics.ai/internal/connectors/conformance	0.415s	coverage: 82.2% of statements
```

## Notes / deviations

- No new go.mod dependencies. No `ENGINE_GAP` blockers — every engine API needed
  (`engine.Load`/`LoadAll`, `Bundle`/`StreamSpec`/`WriteAction`/`StreamSchema`/`APISurface` fields,
  `Read`/`Check`/`ValidateWrite`/`DryRunWrite`/`Write`, `Schema.Validate/Properties/SecretKeys`,
  `CompileSchema`, `ResolveCheck`) already existed exactly as documented from Waves A–D and matched
  API-CONTRACT.md.
- One concrete implementation-time deviation from the pre-code plan (documented above and in
  `replay.go`'s doc comments): `check_fixture` uses a DEDICATED `fixtures/check.json` file instead of
  reusing a stream's replay server, because a bundle's declarative `Check()` request path (e.g.
  `/ping`) generally does not match any stream's recorded fixture path — reusing stream fixtures would
  have made `check_fixture` fail on every bundle whose check path differs from its stream paths
  (which is the common case). This is a small, additive, documented extension to the fixture
  vocabulary, not a workaround around a real defect.
- `formatCursorForAssertion` in `dynamic.go` intentionally duplicates read.go's unexported
  `formatParam` logic in miniature (rfc3339/unix_seconds/date/github_date_range) rather than reaching
  into `engine` internals (not exported, and `read.go` is outside this task's exclusive file list
  regardless) — this exists specifically so `cursor_advances`' assertion is derived independently of
  the code path it is testing.
- Files touched: exactly `internal/connectors/conformance/**` (conformance.go, static.go, dynamic.go,
  replay.go, conformance_test.go, static_test.go, dynamic_test.go, replay_test.go, testdata/good/**,
  testdata/invalid/** [10 seeded static-defect bundles], testdata/dynamic-invalid/** [2 seeded
  dynamic-defect bundles]) plus this ledger file. `internal/connectors/native_conformance.go` (legacy)
  was read as reference only, never modified. No files under `internal/connectors/certify/**` or
  `internal/connectors/engine/**` were modified by this session — those files' more recent mtimes
  (observed via `find -newer`) belong to a parallel Wave E agent (T/B-14, certify source stages)
  working concurrently, confirmed by diffing this session's own edit list against them.
- Static/dynamic check counts: 10 static checks (exact §E.2 list), 8 dynamic check "kinds"
  (`check_fixture`, `read_fixture_nonempty` [once per stream], `pagination_terminates`,
  `records_match_schema`, `cursor_advances`, `write_request_shape` [once per write action],
  `delete_semantics`) — `write_request_shape`/`read_fixture_nonempty` fan out per-action/per-stream,
  so the good control bundle's report carries 19 total check entries (10 static + 9 dynamic instances:
  check_fixture, 2×read_fixture_nonempty, pagination_terminates, records_match_schema,
  cursor_advances, 2×write_request_shape, delete_semantics).
