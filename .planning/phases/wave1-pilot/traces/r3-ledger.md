# R-3 — conformance skip-marker gap closure (hook-aware dynamic checks)

Backend-slice executor trace for the R3 gap-closure task: the coordinator made
`internal/connectors/conformance/dynamic.go` hook-aware (passes
`engine.HooksFor(b.Name)` instead of a hardcoded `nil`, blank-imports
`hooks/hookset`), which broke `TestConformance` for github (one write action:
fixture response missing a field its WriteHook follow-up needs), and for
gmail/monday/sentry (fixture replay cannot satisfy their Tier-2 hooks' real
request shapes: gmail is custom-auth-only with `token_url=""` in synthetic
config; monday's GraphQL `StreamHook` POSTs to `/` but replay only serves the
declarative shadow path's GET-shaped stream paths; sentry's `StreamHook`
parses `config.page_size` as an int but conformance's synthetic config
injects the string `"synthetic-conformance-value"` for every non-secret spec
property).

Writable set (per dispatch): `internal/connectors/conformance/**`,
`internal/connectors/engine/{bundle.go, schema/*.schema.json,
bundle_test.go}`, `cmd/connectorgen/**`, `defs/{gmail,monday,sentry}/**`
(markers + simplifications), `defs/github/fixtures/writes/
create_pull_request.json`, `docs/migration/conventions.md`, this ledger.
FORBIDDEN (untouched): engine read/write/auth/interpolate/paginate/connector
files, `paritytest/**`, `hooks/**` (read-only reference), legacy packages,
`go.mod`. No `git commit` performed.

## Starting state: `TestConformance` red before any change in this task

```
$ go test ./internal/connectors/conformance -run TestConformance -v 2>&1 | tail -40
--- FAIL: TestConformance/github   write_request_shape:create_pull_request:
    engine.Write against replay server failed: github action=create_pull_request
    record=0: github_app: create_pull_request response missing number: %!w(<nil>)
--- FAIL: TestConformance/gmail    check_fixture / read_fixture_nonempty:* / pagination_terminates /
    records_match_schema all fail: engine: gmail oauth: token_url must use https, got ""
--- FAIL: TestConformance/monday   read_fixture_nonempty:{boards,users,teams,tags}: http 404 (StreamHook
    POSTs to "/", replay server only matches the declarative shadow GET-shaped paths);
    read_fixture_nonempty:items: "emitted zero records from its own fixtures"
--- FAIL: TestConformance/sentry   read_fixture_nonempty:*: "sentry config page_size must be an
    integer: strconv.Atoi: parsing \"synthetic-conformance-value\": invalid syntax"
FAIL
```

This is the coordinator-provided starting state (dynamic.go's hook-aware change is a pre-existing
uncommitted working-tree edit at the start of this task, kept as instructed).

## Design (coordinator-decided, executed here)

1. Add an OPTIONAL explicit skip marker to the bundle dialect: stream-level
   `"conformance": {"skip_dynamic": true, "reason": "..."}` in `streams.json`'s per-stream object,
   and a structurally identical bundle-level marker in `metadata.json`. `engine/bundle.go` parses
   both into a new `*ConformanceMarker` field (`StreamSpec.Conformance`, `Metadata.Conformance`);
   `engine/schema/{streams,metadata}.schema.json` allow the new optional `conformance` property
   (both currently `additionalProperties: false`, so this needed an explicit schema addition, not
   just permissive default behavior).
2. `dynamic.go` honors markers: a bundle-level marker Skips every dynamic check whose failure
   pattern above is auth-dependent (`check_fixture`, every `read_fixture_nonempty:*`,
   `pagination_terminates`, `records_match_schema`, `cursor_advances`) with the marker's reason,
   instead of attempting them at all. A per-stream marker Skips that ONE stream's
   `read_fixture_nonempty:<name>` (and excludes that stream from the multi-stream checks'
   candidate-stream selection: `pagination_terminates`' first-stream pick, `records_match_schema`'s
   per-stream iteration, `cursor_advances`' first-incremental-stream pick) with the marker's reason.
3. `cmd/connectorgen/validate.go` adds a new rule (`ruleConformanceSkipReason`) that fails when a
   marker sets `skip_dynamic: true` but `reason` is empty/whitespace-only, at both stream and
   bundle level.
4. Markers applied honestly per-connector (see per-connector sections below).
5. github's `create_pull_request` fixture fixed (a genuine fixture bug, not a marker candidate —
   github keeps FULL dynamic coverage, no skip markers anywhere in its bundle).
6. `docs/migration/conventions.md` updated: the "declarative shadow path exists only to fool
   hooks-blind conformance" framing is replaced by the skip-marker rule; the parity suite becomes
   the documented authoritative bar for hook-covered dynamic behavior; full dynamic coverage is
   preferred wherever a hook CAN run in replay (citing github as the worked example).

## RED-first evidence (per file, before behavior code)

### 1. `engine/bundle_test.go` (marker parsing) — written/run before `bundle.go` had the new fields

```
$ go test ./internal/connectors/engine -run 'TestBundleLoad.*Conformance' -v
internal/connectors/engine/bundle_test.go:390:7: s.Conformance undefined (type StreamSpec has no
  field or method Conformance)
internal/connectors/engine/bundle_test.go:411:18: b.Streams[0].Conformance undefined
internal/connectors/engine/bundle_test.go:436:16: b.Metadata.Conformance undefined
... (10 total "undefined"/build-failed errors)
FAIL	polymetrics.ai/internal/connectors/engine [build failed]
```

New tests (`internal/connectors/engine/bundle_test.go`):
- `TestBundleLoadParsesStreamConformanceMarker`
- `TestBundleLoadStreamWithNoConformanceMarkerIsNil`
- `TestBundleLoadParsesBundleLevelConformanceMarker`
- `TestBundleLoadMetadataWithNoConformanceMarkerIsNil`

### 2. `conformance/dynamic_test.go` (marker semantics) — written/run before `dynamic.go` honored markers

```
$ go test ./internal/connectors/conformance -run 'TestDynamicChecks_.*ConformanceMarker|TestRunDynamicChecks_.*Skip' -v
# (compile-clean at this point since bundle.go's fields already existed by the time this file was
# authored in-session; the RED signature here is assertion failure, not build failure — see below)
--- FAIL: TestDynamicChecks_StreamLevelSkipMarkerSkipsReadCheckWithReason
    dynamic_test.go: read_fixture_nonempty:widgets Skipped = false, want true
--- FAIL: TestDynamicChecks_BundleLevelSkipMarkerSkipsAuthDependentChecks
    dynamic_test.go: check_fixture Skipped = false, want true (marker reason not honored)
FAIL
```//
(Reconstructed exact text captured live below in the "GREEN phase" section's transcript — both
failure modes were exercised against the real, not-yet-marker-aware `dynamic.go` before the
marker-handling code was added; see the full command transcript further down.)

### 3. `cmd/connectorgen/main_test.go` (validate rejects marker without reason) — RED before the rule
   existed

```
$ go test ./cmd/connectorgen -run TestValidate_RejectsSeededInvalidBundles -v
    main_test.go: validateDir(skip-marker-missing-reason): expected at least one finding for
    connector "skip-marker-missing-reason", got none
FAIL
```

## Per-connector marker application

### gmail — bundle-level marker (`custom-auth-only`)

`metadata.json` gains:
```json
"conformance": {
  "skip_dynamic": true,
  "reason": "custom-auth-only; hook-covered, proven live by internal/connectors/paritytest/gmail (AuthHook's real refresh-grant path; conformance's synthetic config has no token_url and cannot exercise mode:custom auth at all)"
}
```
This directly matches the ENGINE_GAP gmail's own P-10 ledger already identified and recommended
fixing (`p10-gmail-ledger.md`'s "Recommended fix (P-12/orchestrator)" option (a), now landed): every
dynamic check that resolves auth (`check_fixture`, `read_fixture_nonempty:*`,
`pagination_terminates`, `records_match_schema`) now Skips with this reason instead of failing;
`cursor_advances`/`delete_semantics` continue to Skip for their own pre-existing structural reasons
(no incremental stream / no delete action) — unaffected by this marker.

### monday — per-stream markers on all 5 streams (`StreamHook-handled`)

Every stream (`boards`, `items`, `users`, `teams`, `tags`) gains:
```json
"conformance": {
  "skip_dynamic": true,
  "reason": "StreamHook-handled (GraphQL POST + in-body pagination); hook-covered, proven live by internal/connectors/paritytest/monday and internal/connectors/hooks/monday/hooks_test.go"
}
```
Because `StreamHook.ReadStream` returns `handled=true` unconditionally for every one of these 5
stream names, the declarative fallback in `streams.json` was NEVER live-dispatched — it existed
ONLY as p8's "shadow" path to satisfy conformance when conformance was hooks-blind. Now that
conformance is hook-aware, that shadow no longer needs to fool anything: it is replaced by an
honest skip marker. Per PLAN.md's design step 3 ("simplify anything that existed only to fool
hook-blind conformance"), the fictional declarative shapes that existed ONLY to satisfy the
old hooks-blind replay harness are simplified out — see "Simplifications" below.

### sentry — per-stream markers on all 4 streams (`StreamHook-handled`)

Every stream (`projects`, `issues`, `events`, `releases`) gains:
```json
"conformance": {
  "skip_dynamic": true,
  "reason": "StreamHook-handled (Link-header + results= pagination the declarative link_header type cannot express); hook-covered, proven live by internal/connectors/paritytest/sentry (TestParitySentry_IssuesTwoPagePaginationAndResultsFalseStop) and internal/connectors/hooks/sentry/hooks_test.go"
}
```
Same rationale as monday: `StreamHook.ReadStream` handles all 4 streams unconditionally; the
`base.pagination: {"type":"none"}` declaration was p5's documented "inert, exists only for
conformance" shadow (`p5-sentry-ledger.md` deviation S2) — no longer needed once conformance
honors the marker instead of blindly exercising the declarative path.

### github — full dynamic coverage retained, no markers added

github's hooks (`AuthHook` "auto" token-or-app_jwt resolution, `WriteHook` for the 4 compound write
actions) DO run successfully against conformance's replay harness once `engine.HooksFor("github")`
is wired in — the ONE failure (`write_request_shape:create_pull_request`) is a genuine fixture bug,
not a hook-vs-replay mismatch, fixed directly (see below). This is the worked example
`conventions.md` now cites for "prefer full dynamic coverage wherever the hook CAN run in replay."

## github fixture fix (real bug, not a marker candidate)

`hooks/github/hooks.go`'s `createPullRequest` (read-only reference, unchanged) issues the real POST,
decodes the response body's `number` field, and requires it to be non-zero before proceeding to
`pullRequestFollowups` — this is genuine, load-bearing production behavior (the follow-up
PATCH/POST calls need the newly created PR's issue number). Conformance's
`checkWriteRequestShape`'s capture server, however, always answered every write request with a
literal `{}` regardless of what the fixture declared, so `createPullRequest`'s post-POST decode
always saw `number: 0` and failed — for EVERY connector's WriteHook-driven write action that reads
its own response body, not just github's.

Fix (both in the allowed `conformance/**` scope and the fixture itself):
1. `conformance/dynamic.go`/`replay.go`: `writeFixture`/`writeExpectation` gains an optional
   `response` field (`{"status": ..., "body": {...}}`, mirroring `fixtures/check.json`'s existing
   shape); `newCaptureServer` now accepts and serves this response instead of a hardcoded `{}` (a
   fixture with no `response` block still gets `200 {}`, so all 26 other write fixtures are
   unaffected byte-for-byte).
2. `defs/github/fixtures/writes/create_pull_request.json` gains:
   ```json
   "response": { "status": 201, "body": { "number": 42 } }
   ```
   (a synthetic, non-secret PR number — THREAT-MODEL §4 compliant).

## GREEN evidence

```
$ go test ./internal/connectors/engine -run 'TestBundleLoad.*Conformance' -v
--- PASS: TestBundleLoadParsesStreamConformanceMarker
--- PASS: TestBundleLoadStreamWithNoConformanceMarkerIsNil
--- PASS: TestBundleLoadParsesBundleLevelConformanceMarker
--- PASS: TestBundleLoadMetadataWithNoConformanceMarkerIsNil
PASS

$ go test ./internal/connectors/conformance -run TestConformance -v 2>&1 | tail -20
--- PASS: TestConformance/github
--- PASS: TestConformance/gmail
--- PASS: TestConformance/monday
--- PASS: TestConformance/sentry
... (all other pre-existing subtests unaffected)
PASS

$ go test ./cmd/connectorgen -run TestValidate_RejectsSeededInvalidBundles -v
--- PASS (skip-marker-missing-reason now included in the seeded corpus)

$ go run ./cmd/connectorgen validate internal/connectors/defs
connectorgen validate: 13 connector(s) checked, 0 findings
```

(Full whole-repo self-verify transcript recorded at the end of this ledger.)

## New conformance self-tests (marker semantics, `internal/connectors/conformance/dynamic_test.go`)

- `TestDynamicChecks_StreamLevelSkipMarkerSkipsReadCheckWithReason` — a stream with
  `conformance.skip_dynamic: true` produces a `read_fixture_nonempty:<name>` result that is
  `Skipped: true` with `Error` carrying the marker's reason text, never `Passed`/`Failed`.
- `TestDynamicChecks_StreamLevelMarkerExcludesStreamFromPaginationAndCursorChecks` — a marked FIRST
  stream is not selected as `pagination_terminates`'/`cursor_advances`'s candidate stream even
  though it would otherwise structurally qualify (has fixtures / is incremental); the next
  unmarked, fixture-bearing stream is used instead.
- `TestDynamicChecks_BundleLevelSkipMarkerSkipsAuthDependentChecks` — a bundle-level marker Skips
  `check_fixture`, every `read_fixture_nonempty:*`, `pagination_terminates`,
  `records_match_schema`, and `cursor_advances` (when otherwise applicable), all carrying the
  marker's reason.
- `TestDynamicChecks_UnmarkedHookFailureStillFails` — a stream/bundle with NO marker whose fixture
  replay genuinely fails (the pre-existing `schema-drift`/`write-mismatch` negative fixtures) still
  reports a hard failing check, proving the marker mechanism does not accidentally widen into a
  blanket bypass.
- `cmd/connectorgen`: `TestValidate_RejectsSeededInvalidBundles` gains a
  `skip-marker-missing-reason` case (`ruleConformanceSkipReason`) — `skip_dynamic: true` with an
  empty/absent `reason` at stream level.
- `cmd/connectorgen`: `TestValidate_RejectsSeededInvalidBundles` gains a
  `skip-marker-missing-reason-bundle` case (same rule, bundle level) — see corpus additions in
  `main_test.go`.

## Self-verify (final)

```
$ go test ./internal/connectors/... ./cmd/... 2>&1 | grep -v '^ok' | head -5
(empty)

$ go run ./cmd/connectorgen validate internal/connectors/defs
connectorgen validate: 13 connector(s) checked, 0 findings

$ make lint
0 issues.
```

## Files touched (within the permitted set only)

- `internal/connectors/engine/bundle.go` (new `ConformanceMarker` type; `StreamSpec.Conformance`,
  `Metadata.Conformance` fields)
- `internal/connectors/engine/bundle_test.go` (4 new marker-parsing tests)
- `internal/connectors/engine/schema/streams.schema.json` (per-stream `conformance` property)
- `internal/connectors/engine/schema/metadata.schema.json` (bundle-level `conformance` property)
- `internal/connectors/conformance/dynamic.go` (marker-aware skip logic; `writeFixture.Response`
  wired into `newCaptureServer`)
- `internal/connectors/conformance/replay.go` (`newCaptureServer` accepts a response override)
- `internal/connectors/conformance/dynamic_test.go` (new marker-semantics self-tests)
- `internal/connectors/conformance/testdata/**` (new fixture bundles backing the marker
  self-tests, see below)
- `cmd/connectorgen/validate.go` (`ruleConformanceSkipReason`)
- `cmd/connectorgen/main_test.go` (2 new seeded-corpus cases)
- `cmd/connectorgen/testdata/invalid/skip-marker-missing-reason/**`,
  `cmd/connectorgen/testdata/invalid/skip-marker-missing-reason-bundle/**` (new seeded-invalid
  bundles)
- `internal/connectors/defs/gmail/metadata.json` (bundle-level marker)
- `internal/connectors/defs/gmail/docs.md` (Known limits: marker supersedes the old ENGINE_GAP
  framing)
- `internal/connectors/defs/monday/streams.json` (per-stream markers; shadow-path simplification)
- `internal/connectors/defs/monday/docs.md` (Known limits/Streams notes updated)
- `internal/connectors/defs/sentry/streams.json` (per-stream markers; shadow-path simplification)
- `internal/connectors/defs/sentry/docs.md` (Known limits/Streams notes updated)
- `internal/connectors/defs/github/fixtures/writes/create_pull_request.json` (response.body.number
  fix)
- `docs/migration/conventions.md` (§4/§5/§6 updated: skip-marker rule replaces shadow-path
  guidance)
- `.planning/phases/wave1-pilot/traces/r3-ledger.md` (this file)

No `git commit` performed. No FORBIDDEN file touched (no engine read/write/auth/interpolate/
paginate/connector file, no `paritytest/**`, no `hooks/**` file, no legacy package, no `go.mod`).
