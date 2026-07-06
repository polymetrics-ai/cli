# gaploop-r2-ledger — wave0-engine-harness CONFORMANCE/GOLDENS/CERTIFY/DOCS batch

Backend gap-closure pass over REVIEW.md (B2 block, F6 flag, F7 flag, F10 doc
slips) and SECURITY-REVIEW.md (M2 major, m4 minor), following R1's engine-core
landing (formatParam digit-cursor passthrough, path interpolation, chained
filters + `join:<sep>` + static-literal computed_fields, SSRF guards on
link_header+next_url, typed unresolved-key errors, Bundle.RawSpec,
ResolveCheckAuthSpec). Strict TDD per item: RED test recorded here BEFORE the
fix, then GREEN.

Scope: `internal/connectors/conformance/**`, `internal/connectors/defs/stripe/**`
(fixtures/schemas only), `internal/connectors/defs/searxng/**`,
`internal/connectors/engine/parity_searxng_test.go` +
`parity_stripe_test.go` (strengthen only), `internal/connectors/certify/**`,
`cmd/connectorgen/**`, `docs/migration/conventions.md`.

---

## B2 — numeric cursors in conformance + real-wire stripe fixtures

### RED (1a — cursor_advances numeric-cursor support)

New self-test bundle `conformance/testdata/good/acme-numeric-cursor/` (a
single `events` stream, `x-cursor-field: created`, `param_format:
unix_seconds`, fixture `created` values as JSON NUMBERS `1700000000` /
`1700000100`, decoded as `json.Number` by the engine's UseNumber decoders)
plus new test `TestCursorAdvances_NumericCursorFieldSupported` in
`dynamic_test.go`. Before the fix:

```
$ go test ./internal/connectors/conformance/... -run TestCursorAdvances_NumericCursorFieldSupported -v
--- FAIL: TestCursorAdvances_NumericCursorFieldSupported (0.00s)
    dynamic_test.go:80: cursor_advances failed on a numeric (json.Number) cursor field: stream "events": no cursor value observed across fixture records
FAIL
```

Exactly the B2 failure: `checkCursorAdvances`'s `raw[sch.CursorField].(string)`
type assertion silently fails (no error) for a `json.Number` cursor, so
`maxCursor` stays `""` and the check reports "no cursor value observed" even
though two fixture records plainly carry increasing numeric `created` values.

Also added `TestCursorAdvances_StringCursorFieldStillSupported` (existing
`testdata/good/acme` bundle, string `updated_at` cursor + `rfc3339`
param_format) as a locked-in companion proving both cursor value shapes must
remain legal after the fix — this one already passes pre-fix (no regression
risk) and stays green post-fix.

### Fix (1a)

`checkCursorAdvances` (dynamic.go): replace the `v.(string)` type assertion
with a small `cursorValueString(v any) (string, bool)` helper that accepts:
- `string` (existing behavior, unchanged — compared/maxed lexicographically as
  before, which is correct for RFC3339 strings since lexicographic order
  matches chronological order for a fixed-width fixed-offset representation),
- `json.Number` (the real-world numeric-cursor decode shape throughout this
  codebase — connsdk/engine JSON decoding always uses `UseNumber()`) — for
  numeric comparison, a `json.Number` cursor is compared/maxed via
  numeric (`big.Float`) comparison, not string comparison (since `"9" >
  "10"` lexicographically but not numerically), and the winning value is
  canonicalized to its digit-string form before being handed to
  `formatCursorForAssertion`, matching the exact digit-string shape the real
  engine's `formatParam`/`parseLowerBoundTime` (R1) now accepts as a valid
  unix_seconds passthrough,
- `float64` (defensive: any caller/fixture path that doesn't go through
  UseNumber decoding, mirroring the same defensive handling R1 already added
  to `lastRecordFieldValue` in paginate.go for F3).

`formatCursorForAssertion`'s own `unix_seconds`/`date` branches already
delegate to `time.Parse(time.RFC3339, ...)`, which cannot parse a bare digit
string — so `formatCursorForAssertion` gains the same digit-passthrough
treatment R1 gave the real `formatParam`/`parseLowerBoundTime` (a small
duplicated `parseLowerBoundTimeForAssertion` helper, consistent with the
existing documented duplication rationale in `formatCursorForAssertion`'s doc
comment: this package intentionally does not reach into engine internals, so
it re-derives the same semantics independently for the assertion side).

### GREEN (1a)

```
$ go test ./internal/connectors/conformance/... -run 'TestCursorAdvances' -v
=== RUN   TestCursorAdvances_PostReadCursorIsMaxFixtureCursorAndReReadSendsParam
--- PASS
=== RUN   TestCursorAdvances_NumericCursorFieldSupported
--- PASS
=== RUN   TestCursorAdvances_StringCursorFieldStillSupported
--- PASS
PASS
```

### Fix (1b/1c — real-wire stripe fixtures + tightened schemas)

Rewrote every `defs/stripe/fixtures/streams/*/page_*.json` so `created` (and,
for `products`, the pre-existing `updated`) is emitted as a bare JSON NUMBER
(Stripe's real wire shape: Unix seconds), not an RFC3339 string — customers
page_1/page_2, charges, invoices, products, subscriptions. `fixtures/check.json`
has no `created` field (a bare `{"object":"list","data":[],"has_more":false}`
check probe) so it needed no change.

Tightened `defs/stripe/schemas/{customers,charges,invoices,products,
subscriptions}.json`: `created` reverted from `["integer","string"]` back to
`"integer"`-only (removing the B2-flagged widening); `customers.json`'s
description note about "represented as an RFC3339 string in this bundle's own
fixtures" deleted (no longer true, and was itself the B2-flagged
false-institutionalization).

### GREEN (1b/1c/1d — parity + conformance + validate all green with real-wire shapes)

```
$ go test ./internal/connectors/engine/... -run TestParityStripe -v
PASS (all stripe parity tests; parity_stripe_test.go's own inline httptest
fixtures already used numeric created — no test change needed there, only
defs/stripe's on-disk fixtures/schemas)
$ go run ./cmd/connectorgen validate internal/connectors/defs
connectorgen validate: 3 connector(s) checked, 0 findings
$ go test ./internal/connectors/conformance/... -v 2>&1 | tail -20
PASS
```

---

## F7 — searxng full data parity via R1 primitives (engines join + stream static-literal)

### RED

Removed the `normalizeSearxngRecord` workarounds in `parity_searxng_test.go`
(the `delete(r, "stream")` + `canonicalEngines` array/string normalization)
WITHOUT yet touching `defs/searxng/streams.json`'s `computed_fields`, to prove
the strengthened test genuinely fails against the old (unresolved-deviation)
bundle shape:

```
$ go test ./internal/connectors/engine/... -run 'TestParitySearxng_SearchStreamRecords|TestParitySearxng_RedditStreamScopesQuery' -v
--- FAIL: TestParitySearxng_SearchStreamRecords
    parity_searxng_test.go:190: record 0 mismatch:
        engine:  map[... engines:[reddit] ... (no "stream" key) ...]
        legacy:  map[... engines:reddit ... stream:search ...]
--- FAIL: TestParitySearxng_RedditStreamScopesQuery
    (same mismatch shape)
FAIL
```

Exactly the F7 failure: `engines` is an unjoined array on the engine side vs.
legacy's comma-joined string, and `stream` is entirely absent on the engine
side — both change the emitted record DATA for identical inputs, violating
conventions.md §5's own meta-rule.

### Fix

`defs/searxng/streams.json`: both streams' `computed_fields` gain
- `"engines": "{{ record.engines | join:, }}"` — R1's `join:<sep>` filter,
  applied to the raw (pre-projection) `record.engines` array, producing the
  identical comma-joined string legacy's `joinAny` emits.
- `"stream": "search"` / `"stream": "reddit"` — a static-literal
  `computed_fields` value (no `{{ }}` markers at all; `Interpolate`/
  `applyComputedFields` already returns a template with no markers verbatim,
  R1's "static-literal computed_fields" primitive), matching legacy's
  `searxngResultRecord`'s per-stream `"stream"` marker exactly.

`defs/searxng/schemas/{search,reddit}.json`: `engines` type narrowed from
`["array","string","null"]` to `["string","null"]` (it is now always a
string, never the raw array); added `stream` (required-shaped as `"string"`,
always emitted).

`parity_searxng_test.go`'s `normalizeSearxngRecord` now does ONLY canonical
JSON re-encoding (no field deletion, no field-value normalization) — a
genuine RAW-record-equality assertion.

### GREEN

```
$ go test ./internal/connectors/engine/... -run TestParitySearxng -v
PASS (all 8 searxng parity tests, including the two strengthened ones)
```

Subreddit-narrowing (`site:reddit.com/r/<sub>`) remains genuinely
inexpressible: query-param templating (`stream.Query`) has no
absent-key-falsy tolerance (that tolerance is scoped to `EvalWhen`/`auth`
specs only — confirmed by reading `interpolate.go`'s `resolveExpr`/
`resolveRefValue`, which query resolution uses directly, vs.
`resolveRefForWhen`), so an unconditional `{{ config.subreddit }}` reference
would hard-error whenever `subreddit` is unset (the common case). Verified
`TestParitySearxng_RedditStreamScopesQuery` already asserts legacy's own
default (no-subreddit) query-scoping behavior matches the bundle's — the
documented base-case-only scope is accurate. Deviation ledger entries 4
(engines array vs. string) and 6 (stream field dropped) move to RESOLVED in
`docs/migration/conventions.md`.

---

## F6 — searxng golden hygiene (dead spec keys, optional bearer-proxy auth)

### RED

`TestParitySearxng_ApiKeySecretSendsBearerAuth` (new, `parity_searxng_test.go`):
both connectors driven with an `api_key` secret configured, asserting an
identical `Authorization: Bearer <token>` header on the outgoing request.
Before wiring `streams.json`'s `auth` block:

```
$ go test ./internal/connectors/engine/... -run TestParitySearxng_ApiKey -v
--- FAIL: TestParitySearxng_ApiKeySecretSendsBearerAuth
    parity_searxng_test.go:292: engine Authorization = "", want "Bearer proxy-token-12345"
--- PASS: TestParitySearxng_ApiKeyAbsentSendsNoAuth (already passing: no auth
    block at all trivially sends no Authorization header, matching legacy's
    default)
FAIL
```

Confirmed via reading `auth.go`'s `authSpecMatches`/`EvalWhen`'s
absent-key-falsy tolerance (R1, already engine-side) that a `when`-gated
bearer spec IS now safe for the common no-credential case — the F6 finding's
premise ("api_key is never applied" / "silently 401s behind an auth proxy")
is fixable, not a permanent limitation.

### Fix

`defs/searxng/streams.json`: `base.auth` gains
`[{"mode":"bearer","token":"{{ secrets.api_key }}","when":"{{ secrets.api_key }}"},
{"mode":"none"}]` — when `api_key` is set, the bearer spec's `when` evaluates
truthy and it's selected; when absent, `EvalWhen`'s absent-key-falsy handling
makes it evaluate false (not error), so `selectAuth` falls through to the
unconditional `none` spec. This mirrors legacy's own conditional behavior
(`searxng.go:184-189`: `if token := ...; strings.TrimSpace(token) != "" { auth
= connsdk.Bearer(token) }`) exactly.

`defs/searxng/spec.json`: dropped the 8 genuinely-inert declared keys
(`subreddit`, `categories`, `engines`, `language`, `time_range`, `safesearch`,
`page_size`, `max_pages`) — confirmed via grep that `streams.json` never
templates any of them (only `base_url` and `query` are referenced), and that
query-param templating has no absent-key-falsy tolerance (unlike `auth`/
`when`), so wiring them unconditionally would hard-error whenever any one of
them is unset (the common case for every one of these optional filters).
`api_key` is KEPT (now genuinely wired, see above). `page_size`/`max_pages`
were ALSO inert for a second, independent reason: `read.go`'s pagination
resolution reads `PaginationSpec.PageSize`/`MaxPages` only from the loaded
`streams.json` base spec, never from `req.Config` — there is no runtime
config-driven override mechanism at all for these, so declaring them as
config properties never had any effect regardless of the `when`-tolerance
question.

### GREEN

```
$ go test ./internal/connectors/engine/... -run TestParitySearxng -v
PASS (all 8, including both new api_key auth tests)
$ go run ./cmd/connectorgen validate internal/connectors/defs
connectorgen validate: 3 connector(s) checked, 0 findings
```

`docs/migration/conventions.md`'s "declared config must be consumed" rule
(F6) — see the F10/docs section below for the conventions.md rewrite that
documents this rule going forward.

---

## F6 (stripe half) — rate_limit placement + dead cursor-paginator fields

### RED

`TestParityStripe_NoDeadPaginationFields` (new, `parity_stripe_test.go`):
asserts `bundle.HTTP.Pagination.LimitParam == "" && PageSize == 0`. Confirmed
via reading `paginate.go`'s `newCursorPaginator` that the
`cursor`+`last_record_field` paginator (stripe's shape) never reads
`PaginationSpec.LimitParam`/`PageSize` (only `page_number`/`offset_limit`
do), so declaring `limit_param: "limit"`/`page_size: 100` on stripe's base
pagination block was dead config — `limit=100` is actually sent via each
stream's static `query: {"limit": "100"}`. Reverted `streams.json`'s
pagination block to include the dead fields to prove RED:

```
$ go test ./internal/connectors/engine/... -run TestParityStripe_NoDeadPaginationFields -v
--- FAIL: TestParityStripe_NoDeadPaginationFields
    parity_stripe_test.go:578: pagination.limit_param = "limit", want empty (dead for cursor+last_record_field paginator, F6)
FAIL
```

Also added `TestParityStripe_MetadataRateLimitIsInformationalOnly`: asserts
`bundle.HTTP.RateLimit == nil` (streams.json declares no base `rate_limit` —
legacy stripe enforces none, so this bundle must not add new
behavior-changing throttling) and `bundle.Metadata.RateLimit.RequestsPerMinute
== 100` (informational only, never consumed by the read path — confirmed via
grep that `read.go` reads only `b.HTTP.RateLimit`, never
`Metadata.RateLimit`).

### Fix

`defs/stripe/streams.json`: removed `limit_param`/`page_size` from
`base.pagination` (dead fields for this paginator type; `limit=100` continues
to be sent via each stream's own static `query.limit`, no behavior change).
`defs/stripe/metadata.json`: removed the `rate_limit.strategy: "token_bucket"`
key (not even a field on the Go `RateLimitSpec{RequestsPerMinute int}` type —
already silently dropped by `encoding/json`, so this was purely dead JSON);
kept `requests_per_minute: 100` as the informational-only published-limit
statement. `docs.md` updated to document both: `limit=100`'s actual wire
mechanism (static query, not pagination fields) and `rate_limit`'s
informational-only status (no client-side throttling added, matching
legacy's real lack of one).

### GREEN

```
$ go test ./internal/connectors/engine/... -run TestParityStripe -v
PASS (all, including both new hygiene regression tests)
$ go run ./cmd/connectorgen validate internal/connectors/defs
connectorgen validate: 3 connector(s) checked, 0 findings
```

conventions.md's deviation ledger entry 3 (stripe `limit_param`/`page_size`
"remain declared... as an honest statement of intended page size") is
superseded by this fix — the fields are now simply absent rather than
declared-but-dead; see the F10 docs rewrite below.

---

## M2 (security major) — certify redaction scan completeness

### RED

`TestSourceStagesSecretLeakInStdoutFailsSecretRedactionNamingStage` (new,
`stages_source_test.go`): a new self-test seam `SabotageStdoutLeak(r, stage,
secret)` plants a KNOWN secret value (the same value the harness already
watches for via `Options.SecretEnv` — planting an arbitrary unrelated string
would not exercise `ScanForSecrets`, which only ever matches against a run's
own known secret values, exactly as it would need to for a real leaked
credential) into the `etl_full_refresh_append` stage's captured stdout
(chosen as the highest-risk stage per the security review: "`etl run` is the
highest-risk call... most likely place for a secret to leak into an error
message or verbose output"), without touching that stage's own pass/fail
outcome. Reverted `finalizeSecretRedaction` to the prior
argv-only-scan implementation to confirm RED:

```
$ go test ./internal/connectors/certify/... -run TestSourceStagesSecretLeakInStdoutFailsSecretRedactionNamingStage -v
--- FAIL: TestSourceStagesSecretLeakInStdoutFailsSecretRedactionNamingStage
    stages_source_test.go:284: Capabilities.SecretRedaction.Result = "pass", want fail (a secret was planted in etl_full_refresh_append's stdout)
FAIL
```

Exactly the M2 failure: a secret leaking into a stage's stdout is silently
missed while `secret_redaction` still reports "pass".

### Fix

- `CLIResult` (cliharness.go) already captured `Stdout`/`Stderr`
  (SECURITY-REVIEW.md confirms this — the gap was never in capture, only in
  scanning) — no change needed there.
- New `runContext.run(args...) CLIResult` wrapper method (stages_source.go)
  replaces all ~21 production `rc.harness.Run(...)` call sites (verified via
  grep: every call site outside `run` itself now goes through it): calls the
  harness, then appends a `capturedOutput{stage, stdout, stderr}` (RAW,
  unredacted) to `rc.capturedOutputs`, tagged with `rc.currentStage`.
- `recordStage` gained an `rc *runContext` parameter, setting
  `rc.currentStage = name` for the duration of the stage's body closure
  (restored via `defer` afterward) so every `rc.run(...)` call made from
  within any stage's closure — including ones made from shared helpers like
  `queryRowCount`/`setupCaptureConnection` that don't themselves know the
  calling stage's name — is tagged with the correct owning stage
  automatically, without threading a stage-name parameter through every
  helper function individually.
- `finalizeSecretRedaction` (renamed signature: `(rc *runContext, rep
  *Report, secretValues []string)`) now scans, in order: (1) every stage's
  redacted argv (unchanged — a distinct, independent control), (2) every
  captured stdout AND stderr in `rc.capturedOutputs`, (3) the report itself
  marshaled to JSON exactly as `Report.Save` will persist it (pre-save: this
  runs before `Run` returns, and nothing in this package calls `Save` before
  that). A hit anywhere sets `Result: "fail"` and a `Reason` naming the
  offending stage.
- `Runner.Run`'s final `rep.Passed` computation now also requires
  `Capabilities.SecretRedaction.Result != "fail"` (previously computed from
  `allStagesPassed(rep.Stages)` alone, which `SecretRedaction` — a
  Capabilities field, not a Stage — could never influence): per the
  certification design doc's own enablement gate (`cmd/certstatus` keys off
  `Report.Passed`), a "fail" secret-redaction result must actually fail the
  report, not just annotate it.

### GREEN

```
$ go test ./internal/connectors/certify/... -v
PASS (all, including the new M2 regression test; TestSourceStagesAgainstSample
still asserts SecretRedaction.Result == "pass" against REAL stdout/stderr
now, not just argv, and every pre-existing test is unaffected — the run()
wrapper refactor is behavior-preserving for every non-sabotaged path)
```

---

## m4 — ScanForSecrets JSON-escaped form detection

### RED

`TestScanForSecretsDetectsJSONEscapedForm` (new, `cliharness_test.go`): a
secret containing a double-quote and a backslash, rendered as it would
appear inside a `--json` envelope's string value (RFC 8259 escaping:
`\"`/`\\`), must still be detected.

```
$ go test ./internal/connectors/certify/... -run TestScanForSecretsDetectsJSONEscapedForm -v
--- FAIL: TestScanForSecretsDetectsJSONEscapedForm
    cliharness_test.go:188: ScanForSecrets(JSON-escaped form) = empty, want a hit...
FAIL
```

### Fix

`containsSecretForm` (cliharness.go) gains a `jsonEscapedForm(secret)` check:
`json.Marshal(secret)` (which always yields a quoted JSON string) with the
wrapping quote bytes stripped, added alongside the existing
exact/base64/URL-encoded checks.

### GREEN

```
$ go test ./internal/connectors/certify/... -run TestScanForSecrets -v
PASS (all, including the exact/base64/URL-encoded cases unchanged)
```

---

## connectorgen wiring — ResolveCheckAuthSpec + filter-name validation

### RED

Two new invalid corpus cases added to `cmd/connectorgen/testdata/invalid/`:
- `auth-field-unknown-spec-key`: a `basic`-mode auth spec whose `username`
  field templates `{{ config.nope_username }}`, an undeclared spec key.
- `unknown-filter-in-template`: a `computed_fields` template
  `{{ record.tags | reverse }}` referencing an unknown filter name.

Both seeded into `TestValidate_RejectsSeededInvalidBundles` expecting
`ruleInterpolationUnresolved`.

```
$ go test ./cmd/connectorgen/... -run 'TestValidate_RejectsSeededInvalidBundles/auth-field-unknown-spec-key|TestValidate_RejectsSeededInvalidBundles/unknown-filter-in-template' -v
--- FAIL: .../auth-field-unknown-spec-key
    main_test.go:99: expected at least one finding for connector
    "auth-field-unknown-spec-key", got none
--- PASS: .../unknown-filter-in-template
FAIL
```

The filter-name case was ALREADY caught pre-fix: `checkInterpolations`
already runs every `computed_fields` template through
`interpolationResolveCheck` -> `engine.ResolveCheck`, and `ResolveCheck`
already validates every filter-chain stage against `isKnownFilter` (R1's
join:-prefix support included) — so filter-name validation was already fully
wired via the existing per-template call sites; only the AuthSpec-field gap
was real RED. Kept both cases in the corpus regardless (per the task's
explicit instruction to seed 2 new cases, and the filter case is a genuine,
valuable locked-in regression test even though it required no new
production code).

### Fix

`checkInterpolations` (cmd/connectorgen/validate.go): replaced the
3-field-only manual check (`a.Token`, `a.Value`, `a.When`) for each
`b.HTTP.Auth` entry with a single call to `engine.ResolveCheckAuthSpec(a,
specKeys)` (R1's engine-exported helper, previously unreachable from
connectorgen per the R1 ledger's own note that `internal/cli`/
`cmd/connectorgen` were outside that task's editable scope) — this validates
every templated AuthSpec field (token/username/password/token_url/client_id/
client_secret/scopes/when) against declared spec keys, named with the
existing `ruleInterpolationUnresolved` rule.

### GREEN

```
$ go test ./cmd/connectorgen/... -run TestValidate_RejectsSeededInvalidBundles -v
PASS (all 16 cases, including both new ones)
$ go test ./cmd/connectorgen/... -v
PASS (full package)
$ go run ./cmd/connectorgen validate internal/connectors/defs
connectorgen validate: 3 connector(s) checked, 0 findings
```

---

## F10 + B2/F7 documentation — conventions.md + REVIEW follow-through

Doc-only changes (no test attached — these are prose accuracy fixes, verified
by direct cross-reference against the source they describe rather than a
RED/GREEN pair):

- Fixed the misnamed rule attribution (§2): `connectorgen validate`'s actual
  rule names are `primary_key_missing`/`cursor_field_missing`
  (`cmd/connectorgen/validate.go`'s `rulePrimaryKeyMissing`/
  `ruleCursorFieldMissing` constants); `pk_fields_exist`/`cursor_fields_exist`
  are `conformance/static.go`'s names — both now correctly attributed.
- Fixed the "load-time error" mislabel (§3): `cursor`'s `token_path`+
  `last_record_field` mutual-exclusivity check runs from `paginate.go`'s
  `newPaginator`, called once per `Read`/`Check` (via `read.go`'s
  `newRuntime`), not once at bundle load — confirmed `connectorgen validate`
  has zero `PaginationSpec` references, so this genuinely only surfaces at
  read time.
- Deleted §4's "RFC3339 string cursors in conformance fixtures" convention
  entirely (superseded by B2's fix — see above); replaced with "fixtures use
  the API's real wire shape, `cursor_advances` handles both string and
  numeric cursor values."
- Rewrote the §5 parity-deviation ledger: entries 2 (stripe RFC3339
  fixtures), 4 (searxng engines array), 6 (searxng stream field), and 8
  (searxng api_key auth) all marked RESOLVED with a one-line pointer to the
  fix; entry 3 (stripe limit_param/page_size) marked RESOLVED (fields
  removed, no longer declared-but-dead); entry 7 (subreddit-narrowing)
  confirmed still genuinely open and left ACCEPTABLE (documented scope
  narrowing) — verified query-templating has no absent-key-falsy tolerance
  by reading `interpolate.go`'s `resolveExpr` (used by query resolution)
  vs. `resolveRefForWhen` (used only by `EvalWhen`).
- Added the R1-established engine rules throughout §3: chained filters
  (`{{ ref | f1 | f2 }}`), `join:<sep>`, static-literal `computed_fields`,
  path interpolation (`InterpolatePath` now wired into reads, not just
  writes, plus the `..`-segment traversal guard), the full
  namespace-scoped header-omission decision table (secrets always hard-error;
  config declared-optional omits; config required/undeclared hard-errors),
  and `metadata.json.rate_limit` informational-only vs. `streams.json`'s
  `base.rate_limit` (the only field actually enforced).
- Stale test comment removed: `parity_searxng_test.go`'s
  `searxngShortPageServer` doc comment referenced
  `TestParityStripe_MaxPagesStopEngineGap` (wrong test name entirely, AND a
  test that no longer exists — it was renamed/strengthened to
  `TestParitySearxng_MaxPagesStop` by the earlier waveF repair). Rewritten to
  correctly describe the current (closed) state of the max_pages gap.
- `docs/migration/conventions.md` final line count: 457 (≤ ~460 target).

### Verification

```
$ go build ./... && go vet ./internal/connectors/...
exit 0
$ go test ./internal/connectors/... ./cmd/... 2>&1 | grep -v '^ok\|no test files'
(empty — zero failures)
$ go run ./cmd/connectorgen validate internal/connectors/defs
connectorgen validate: 3 connector(s) checked, 0 findings
$ make lint
0 issues.
$ gofmt -l internal/connectors/conformance internal/connectors/certify internal/connectors/defs cmd/connectorgen internal/connectors/engine
(empty — clean)
$ go test ./internal/connectors/engine/... -cover
coverage: 85.7% of statements (gate ≥85%, unchanged from R1's landing)
$ wc -l docs/migration/conventions.md
457
```
