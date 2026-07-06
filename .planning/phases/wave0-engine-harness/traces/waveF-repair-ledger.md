# TDD ledger — Wave0 engine-harness repair: MaxPages wiring + EvalWhen absent-key

Repairs two confirmed ENGINE_GAPs surfaced by the searxng golden migration (commit e58ff6e;
`.planning/phases/wave0-engine-harness/traces/waveF-b16-ledger.md`'s ENGINE_GAP section). Scope:
`internal/connectors/engine/{read.go,read_test.go,interpolate.go,interpolate_test.go,
parity_searxng_test.go}` + `internal/connectors/defs/searxng/streams.json` (max_pages value only).
No other file touched.

## GAP 1 — PaginationSpec.MaxPages unwired in read.go's readDeclarative

### Discovery notes (ground truth before writing tests)

- `internal/connectors/engine/bundle.go:132`: `PaginationSpec.MaxPages int` (json `max_pages`)
  already exists as a struct field. Confirmed via grep that, before this repair, `read.go` never
  reads it (zero references to `MaxPages` outside `bundle.go`'s declaration and
  `paginate_test.go`'s `TestNewPaginatorPageNumberMaxPagesStop`, which drives `connsdk.Harvest`
  directly with a hard-coded literal `1` for Harvest's own `maxPages` parameter — not through any
  bundle-driven production path).
- `read.go`'s `readDeclarative` resolves the effective `PaginationSpec` via: `pag := stream.Pagination;
  if pag == nil { pag = b.HTTP.Pagination }` (read.go:66-69, pre-fix). This is whole-spec
  replacement, NOT a field-by-field merge — the stream-level spec, when non-nil, is used wholesale
  and the base-level spec is ignored entirely, matching the plan's required
  "stream-level overrides base-level" semantics with no partial-merge surprises.
- `connsdk.Harvest` (`internal/connectors/connsdk/paginate.go:195-201`) is the reference
  implementation of the exact stop semantics required: `if maxPages > 0 && pageNum >= maxPages {
  return nil }`, checked BEFORE issuing the request for `pageNum` (0-indexed), i.e. a hard cap of
  exactly `maxPages` requests regardless of page fullness, with `maxPages <= 0` meaning unbounded.
  This repair mirrors that exact check inline in `readDeclarative`'s own loop (read.go is the
  sanctioned file for this fix; `connsdk` and `paginate.go` are NOT touched — the cap belongs in
  read.go's loop, not paginate.go, because `newPaginator`/the `connsdk.Paginator` interface has no
  page-count-aware method, and adding one would touch the `connsdk.Paginator` contract shared by
  every paginator type, which is out of this repair's sanctioned file set).

### RED evidence (before the fix)

Three new tests added to `read_test.go`:

- `TestReadMaxPagesHardStopsRequestCount`: source always returns a FULL page (2 records, PageSize=2)
  so the short-page stop signal never fires; MaxPages=2 must still stop at exactly 2 requests.
- `TestReadMaxPagesZeroIsUnbounded`: MaxPages=0 must NOT introduce any cap (existing short-page-stop
  behavior only) — guards against a regression once the fix lands.
- `TestReadMaxPagesAbsentPaginationSpecIsUnbounded`: no Pagination spec at all (nil) — the
  `nonePaginator` path — must stay a single request, unaffected by the MaxPages wiring.
- `TestReadMaxPagesStreamLevelOverridesBase`: base declares MaxPages=1, stream declares MaxPages=3;
  the stream-level override must win (3 requests observed), proving the existing whole-spec-replace
  merge semantics are preserved by the fix.

```
$ go test ./internal/connectors/engine -run TestReadMaxPagesHardStopsRequestCount -v -timeout 10s
=== RUN   TestReadMaxPagesHardStopsRequestCount
panic: test timed out after 10s
        running tests:
                TestReadMaxPagesHardStopsRequestCount (10s)
...
polymetrics.ai/internal/connectors/engine.readDeclarative(...)
        .../internal/connectors/engine/read.go:117 +0x8dc
polymetrics.ai/internal/connectors/engine.ReadWithSleeper(...)
        .../internal/connectors/engine/read.go:60 +0x354
...
FAIL    polymetrics.ai/internal/connectors/engine      10.464s
FAIL
```

This is the exact shape of the gap: with an always-full-page source and MaxPages=2 declared, the
engine's real Read() loops forever (would loop until the test's `httptest.Server` is torn down /
the process is killed) because `readDeclarative` never consults `MaxPages` at all. This confirms
the gap is a genuine unbounded-request risk, not merely "reports a slightly wrong count."

```
$ go test ./internal/connectors/engine -run TestReadMaxPagesStreamLevelOverridesBase -v -timeout 10s
(times out identically — same gap, stream-level MaxPages=3 also never consulted)
FAIL    polymetrics.ai/internal/connectors/engine      ~10s
FAIL
```

```
$ go test ./internal/connectors/engine -run TestReadMaxPagesZeroIsUnbounded -v -timeout 10s
--- PASS (0.00s) — already green pre-fix, confirming this guard test does not exercise the gap
                    itself (MaxPages=0 was always unbounded; this test protects the "preserve
                    current behavior" requirement once the cap is wired in).

$ go test ./internal/connectors/engine -run TestReadMaxPagesAbsentPaginationSpecIsUnbounded -v -timeout 10s
--- PASS (0.00s) — same: guards nil-Pagination behavior stays a single request post-fix.
```

### Fix

`internal/connectors/engine/read.go`'s `readDeclarative` page loop: added a `maxPages` local
(resolved from the SAME effective `specForPaginator` already used to build the paginator — no new
merge logic, reusing the existing stream-overrides-base resolution) and a hard stop check
`if maxPages > 0 && pageNum >= maxPages { break }`, placed at the TOP of the loop body (mirroring
`connsdk.Harvest`'s placement: checked before issuing the request for that page number, so
`pageNum` counts REQUESTS ISSUED, not pages fully processed). `MaxPages <= 0` (the zero value,
i.e. absent/unset) is unbounded — no behavior change from today for every existing test.

### GREEN evidence (after the fix)

See "Full suite GREEN evidence" section below (all four new tests pass, no timeouts).

## GAP 2 — EvalWhen hard-errors on an absent config/secrets key

### Discovery notes (ground truth before writing tests)

- `internal/connectors/engine/interpolate.go`'s `EvalWhen` (pre-fix) delegates every branch
  (equality, `in`, truthiness) to `resolveRef`, the SAME helper `Interpolate`/`InterpolatePath`/
  `InterpolateHeader` use for general template resolution. `resolveRef` hard-errors
  (`fmt.Errorf("interpolate: unresolved key %q in %s", ...)`) whenever a `config.*` or `secrets.*`
  key is absent from the `Vars` maps — by design, for general interpolation (a missing base_url or
  bearer token SHOULD be a hard error, not silently emptied). But `EvalWhen` reused this same
  strict helper, so a `when` condition referencing an optional, caller-may-not-have-populated
  secret (e.g. `when: "{{ secrets.api_key }}"` gating an optional Bearer-proxy auth spec) also hard
  errors instead of evaluating falsy — breaking the "optional credential" pattern documented as a
  scope simplification in `waveF-b16-ledger.md` item 5 (searxng's own `api_key` secret had to be
  omitted from an `auth` block entirely because of this exact gap).
- Confirmed via reading `auth.go` in full: `EvalWhen` is called from exactly ONE production call
  site, `authSpecMatches` (auth.go:50-55), itself called only from `selectAuth`. No other call site
  exists in `engine/*.go` (grep confirms). This repair changes `EvalWhen`'s absent-key behavior
  only; `authSpecMatches`'s own signature and `selectAuth`'s signature are UNCHANGED (no call-site
  update needed beyond what already exists — `EvalWhen`'s exported signature
  `(cond string, vars Vars) (bool, error)` does not change).
- Confirmed `ResolveCheck` (interpolate.go:306-331, static validator used by connectorgen validate)
  is a SEPARATE code path from `EvalWhen` — it only checks that a referenced `config.*` key exists
  in `specKeys` (the spec.json-declared property set); it has no runtime `Vars` at all and was
  never affected by `resolveRef`'s runtime absent-key error. This repair does not touch
  `ResolveCheck`; a dedicated regression test proves it. `secrets`/`record` references are (and
  remain) not statically checkable there (no specKeys equivalent for secrets/record namespaces
  exists in the dialect).
- Scope boundary (explicitly verified, not assumed): general `Interpolate`/`InterpolatePath`/
  `InterpolateHeader` calls (bearer tokens, base URL, headers, query params, computed_fields) must
  NOT be affected — an absent key there is still a hard error. The fix is therefore implemented as
  new, `EvalWhen`-local resolution logic (a small unexported helper), not as a behavior change to
  `resolveRef` itself (which remains exactly as strict as before for every other caller).

### RED evidence (before the fix)

Four new/updated test functions added to `interpolate_test.go`:

- `TestEvalWhenAbsentKeyEvaluatesFalsy` (table-driven, 7 cases): truthiness on an absent secret/
  config key -> false; `==` against an absent key compares as empty string (both the "mismatch"
  and "matches the empty-string literal" cases); `in [...]` against an absent key is "not
  contained" UNLESS the list itself contains the empty-string literal (proving the absent value
  really is treated as `""`, not specially short-circuited).
- `TestEvalWhenAbsentKeyDoesNotLeakIntoGeneralInterpolation`: proves `Interpolate` itself (not
  `EvalWhen`) still hard-errors on the exact same absent keys — the scope boundary.
- `TestResolveCheckStillRejectsSpecUnknownKeyForWhenTemplates`: a when-template referencing a
  spec-KNOWN key passes `ResolveCheck` even though runtime-absent; a spec-UNKNOWN key (typo) still
  fails `ResolveCheck`.

```
$ go test ./internal/connectors/engine -run 'TestEvalWhenAbsentKey|TestResolveCheckStillRejectsSpecUnknownKeyForWhenTemplates' -v -timeout 20s
=== RUN   TestEvalWhenAbsentKeyEvaluatesFalsy
=== RUN   TestEvalWhenAbsentKeyEvaluatesFalsy/truthiness:_absent_secret_key
    interpolate_test.go:249: EvalWhen("{{ secrets.api_key }}") unexpected error: interpolate: unresolved key "api_key" in secrets (absent key must evaluate falsy in a when-condition, not error)
=== RUN   TestEvalWhenAbsentKeyEvaluatesFalsy/truthiness:_absent_config_key
    interpolate_test.go:249: EvalWhen("{{ config.missing_cfg }}") unexpected error: interpolate: unresolved key "missing_cfg" in config (absent key must evaluate falsy in a when-condition, not error)
=== RUN   TestEvalWhenAbsentKeyEvaluatesFalsy/equality:_absent_secret_key_compares_as_empty_string,_mismatch
    interpolate_test.go:249: EvalWhen("{{ secrets.api_key == 'sekret-token' }}") unexpected error: interpolate: unresolved key "api_key" in secrets ...
=== RUN   TestEvalWhenAbsentKeyEvaluatesFalsy/equality:_absent_secret_key_compares_as_empty_string,_match_empty_literal
    interpolate_test.go:249: EvalWhen("{{ secrets.api_key == '' }}") unexpected error: interpolate: unresolved key "api_key" in secrets ...
=== RUN   TestEvalWhenAbsentKeyEvaluatesFalsy/equality:_absent_config_key_compares_as_empty_string
    interpolate_test.go:249: EvalWhen("{{ config.missing_cfg == 'anything' }}") unexpected error: interpolate: unresolved key "missing_cfg" in config ...
=== RUN   TestEvalWhenAbsentKeyEvaluatesFalsy/in:_absent_secret_key_is_not_contained_in_any_non-empty_list
    interpolate_test.go:249: EvalWhen("{{ secrets.api_key in ['a', 'b'] }}") unexpected error: interpolate: unresolved key "api_key" in secrets ...
=== RUN   TestEvalWhenAbsentKeyEvaluatesFalsy/in:_absent_config_key_is_not_contained_even_in_a_list_containing_empty_string
    interpolate_test.go:249: EvalWhen("{{ config.missing_cfg in ['', 'x'] }}") unexpected error: interpolate: unresolved key "missing_cfg" in config ...
--- FAIL: TestEvalWhenAbsentKeyEvaluatesFalsy (0.00s)
    (all 7 subtests FAIL)
=== RUN   TestEvalWhenAbsentKeyDoesNotLeakIntoGeneralInterpolation
--- PASS: TestEvalWhenAbsentKeyDoesNotLeakIntoGeneralInterpolation (0.00s)
=== RUN   TestResolveCheckStillRejectsSpecUnknownKeyForWhenTemplates
--- PASS: TestResolveCheckStillRejectsSpecUnknownKeyForWhenTemplates (0.00s)
FAIL
FAIL    polymetrics.ai/internal/connectors/engine      0.220s
FAIL
```

The two PASS results (pre-fix) are expected: they assert behavior that must NOT change (general
interpolation still errors; ResolveCheck already rejects unknown keys) — genuine regression guards,
not red/green pairs for new production code, matching the wave0-b16 ledger's own precedent for its
`TestSearxngRegistrygenSkipMapRegression`.

### Fix

`internal/connectors/engine/interpolate.go`: added an unexported `resolveRefForWhen(ref string,
vars Vars) (string, error)` used ONLY inside `EvalWhen`'s three branches (equality, `in`,
truthiness), in place of the shared `resolveRef`. It delegates to `resolveRef` for every case
EXCEPT a `config.*`/`secrets.*` reference whose key is absent from the respective map — in that one
case it returns `("", nil)` (empty string, no error) instead of propagating the unresolved-key
error. `record.*` and bare `cursor` references are unaffected (delegated straight through — a
`when` condition referencing an absent record path was already possible only in record-scoped
contexts that don't apply to auth/`when` evaluation today, so no new tolerance is introduced there
beyond what `resolveRef` itself already does for `record.*` paths via `resolveRecordPath`, which
already returns "" instead of erroring... actually `resolveRecordPath` DOES error on an absent
record path per interpolate.go's own doc comment claim vs actual code — verified: it does return an
error `unresolved key %q in record`, same as config/secrets; `read.go`'s `applyComputedFields`
catches that specific error string itself via `isUnresolvedRecordPath`. `EvalWhen` has no `Record`
in its `Vars` today (only `authVars` builds `Vars` for it, which never sets `Record`), so this edge
case is currently unreachable via the one production call site; `resolveRefForWhen` nonetheless
treats an absent `record.*` path the same tolerant way for consistency/future-proofing, since a
`when` condition gating on record-shape is conceptually the same "presence check" pattern as
config/secrets).

`resolveRef` itself is UNCHANGED — every other caller (`Interpolate`, `InterpolatePath`,
`InterpolateHeader`, `buildAuthenticator`'s per-mode template resolution) keeps hard-erroring on an
absent key, exactly as before.

`auth.go` required NO call-site change: `EvalWhen`'s signature is unchanged
(`func EvalWhen(cond string, vars Vars) (bool, error)`), and `authSpecMatches`/`selectAuth` already
just propagate whatever `EvalWhen` returns.

### GREEN evidence (after the fix)

See "Full suite GREEN evidence" below.

## searxng bundle: max_pages value (streams.json)

`internal/connectors/defs/searxng/streams.json`'s base pagination block already declared
`"max_pages": 1` (authored in the T/B-16 golden migration, before this repair — see
`waveF-b16-ledger.md`), which is legacy's real default (`searxngDefaultMaxPages = 1`,
`searxng.go:39`). No value change was needed in this repair: GAP 1's fix means this static `1` now
actually takes effect on the engine read path for the first time. Per the repair brief,
config-driven `max_pages` (legacy reads `max_pages` from CONFIG at runtime, allowing a caller to
raise/disable the cap per-read) is NOT modeled — `PaginationSpec.MaxPages` is a static int with no
template support, and adding template support to that field would be a dialect change outside this
repair's sanctioned scope. This remains a documented, deliberate deviation (already noted in
`docs.md`'s "Known limits" from the original T/B-16 task); parity is asserted for the DEFAULT case
only (max_pages=1, the common/default path), not the config-override path.

## parity_searxng_test.go: TestParitySearxng_MaxPagesStopEngineGap -> TestParitySearxng_MaxPagesStop

Flipped from asserting the gapped behavior (`engHits > 1`, i.e. "the engine does NOT stop") to
asserting REAL byte-for-byte-equivalent parity with legacy's max_pages=1 default hard stop: both
legacy and engine, driven against an always-full-page source (so the short-page stop signal never
fires on its own), must issue EXACTLY 1 request and return EXACTLY `searxngPageSize` records. Test
renamed `TestParitySearxng_MaxPagesStop`; its stale doc-comment block (the "KNOWN, DOCUMENTED
ENGINE_GAP" explanation) was rewritten to describe the CLOSED gap and point back at this ledger
instead of the b16 one.

### Interim regression caught and fixed while making the RED tests GREEN

After wiring MaxPages into read.go, the FULL engine suite (not just the new tests) was re-run and
caught a real regression in an EXISTING test, `TestParitySearxng_PagenoSequenceAndShortPageStop`:
that test's legacy side explicitly passes `max_pages: "all"` (unbounded) to isolate its own concern
(pageno sequence + short-page stop across 2 pages) from the max_pages cap, but the engine side had
no equivalent override — so once the cap started being honored, the bundle's declared
`max_pages: 1` stopped the engine side after page 1, before the short-page signal on page 2 was
ever reached (`engine records = 10, want 11`). Fixed by adding a
`withSearxngUnboundedMaxPages` test helper (shallow-copies `b.HTTP.Pagination` with `MaxPages: 0`,
mirroring legacy's own `max_pages:"all"` config override) and using it ONLY in that one test, so
this test's isolation-from-the-cap intent (already present on the legacy side, per the original
b16 authoring) is now symmetric on the engine side too. No other existing test needed this
treatment (every other parity/read test either uses a naturally short final page, so the ordinary
stop signal fires before any cap would matter, or doesn't use page_number pagination at all).

### RED evidence for the renamed parity test (before GAP 1's fix)

```
$ go test ./internal/connectors/engine -run TestParitySearxng_MaxPagesStopEngineGap -v
=== RUN   TestParitySearxng_MaxPagesStopEngineGap
--- PASS (0.00s)   // pre-fix: asserted engHits > 1 (the gapped behavior) — this was the OLD
                    // test's passing assertion before it was rewritten/renamed; it is not a
                    // red/green pair itself (it deliberately asserted the bug). The rewritten
                    // TestParitySearxng_MaxPagesStop (asserting real parity: exactly 1 request)
                    // was authored directly in its place and confirmed GREEN only after GAP 1's
                    // read.go fix landed — verified by reverting the read.go MaxPages check
                    // locally and re-running just this test, which reproduces the same timeout/
                    // request-count failure documented for TestReadMaxPagesHardStopsRequestCount
                    // above (both drive the exact same readDeclarative code path).
```

## Full suite GREEN evidence (after both fixes + parity test flip + regression fix)

```
$ go build ./...
(clean, no output)

$ go vet ./internal/connectors/...
(clean, no output)

$ gofmt -l internal/connectors
(empty)

$ go test ./internal/connectors/engine/... -v -timeout 60s
... 162 "--- PASS" lines, 0 "--- FAIL", exit 0, including:
--- PASS: TestReadMaxPagesHardStopsRequestCount
--- PASS: TestReadMaxPagesZeroIsUnbounded
--- PASS: TestReadMaxPagesAbsentPaginationSpecIsUnbounded
--- PASS: TestReadMaxPagesStreamLevelOverridesBase
--- PASS: TestEvalWhenAbsentKeyEvaluatesFalsy (all 7 subtests)
--- PASS: TestEvalWhenAbsentKeyDoesNotLeakIntoGeneralInterpolation
--- PASS: TestResolveCheckStillRejectsSpecUnknownKeyForWhenTemplates
--- PASS: TestParitySearxng_SearchStreamRecords
--- PASS: TestParitySearxng_RedditStreamScopesQuery
--- PASS: TestParitySearxng_PagenoSequenceAndShortPageStop
--- PASS: TestParitySearxng_MaxPagesStop
--- PASS: TestParitySearxng_ManifestSurface
--- PASS: TestParitySearxng_BundleLoadsAndValidates
--- PASS: TestSearxngRegistrygenSkipMapRegression
--- PASS: TestParityStripe_* (all, unaffected)
PASS

$ go test ./internal/connectors/... (full module connectors tree)
565 "ok" package results, 0 FAIL, exit 0

$ go run ./cmd/connectorgen validate internal/connectors/defs
connectorgen validate: 3 connector(s) checked, 0 findings

$ make lint
golangci-lint run ./internal/connectors/engine/... ./internal/connectors/defs/... ./internal/connectors/hooks/... ./internal/connectors/native/... ./internal/connectors/conformance/... ./internal/connectors/certify/... ./cmd/connectorgen/... ./cmd/inventorygen/...
0 issues.

$ go test ./internal/connectors/engine -cover -timeout 60s
ok  	polymetrics.ai/internal/connectors/engine	0.423s	coverage: 85.0% of statements
```

## Files touched (exhaustive)

- `internal/connectors/engine/read.go` (MaxPages wiring in `readDeclarative`'s loop)
- `internal/connectors/engine/read_test.go` (4 new tests)
- `internal/connectors/engine/interpolate.go` (`resolveRefForWhen` helper; `EvalWhen` branches
  updated to use it)
- `internal/connectors/engine/interpolate_test.go` (3 new/updated test functions)
- `internal/connectors/engine/parity_searxng_test.go` (`TestParitySearxng_MaxPagesStopEngineGap`
  renamed/flipped to `TestParitySearxng_MaxPagesStop`, doc comment rewritten)
- `internal/connectors/defs/searxng/streams.json` (no value change; `max_pages: 1` already correct
  from the original T/B-16 task — confirmed, not re-authored)
- `.planning/phases/wave0-engine-harness/traces/waveF-repair-ledger.md` (this file)

No other file was touched. `internal/connectors/engine/auth.go` required no edit (call site
signature unchanged); `internal/connectors/engine/paginate.go` was read but not modified (the
MaxPages cap belongs in `read.go`'s loop, not the `connsdk.Paginator` contract — justified above).
