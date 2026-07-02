# Wave A ledger — engine foundations (T/B-01..04)

## T-01

Status: red-confirmed
Timestamp: 2026-07-02T07:52:03Z

```
$ go test ./internal/connectors/engine -run TestSchema -v 2>&1 | tail -40
# polymetrics.ai/internal/connectors/engine [polymetrics.ai/internal/connectors/engine.test]
internal/connectors/engine/schema_test.go:64:14: undefined: CompileSchema
internal/connectors/engine/schema_test.go:185:16: undefined: CompileSchema
internal/connectors/engine/schema_test.go:216:14: undefined: CompileSchema
internal/connectors/engine/schema_test.go:240:14: undefined: CompileSchema
internal/connectors/engine/schema_test.go:266:14: undefined: CompileSchema
internal/connectors/engine/schema_test.go:270:9: undefined: StreamSchema
internal/connectors/engine/schema_test.go:280:12: undefined: CompileSchema
internal/connectors/engine/schema_test.go:285:11: undefined: CompileSchema
FAIL	polymetrics.ai/internal/connectors/engine [build failed]
FAIL
```

RED confirmed: schema_test.go authored per PLAN.md T-01 scope (keyword matrix, annotations,
x-secret/x-primary-key/x-cursor-field accessors, unknown-keyword compile error, JSON-pointer-ish
invalid-instance error paths). Compile errors are expected RED evidence since `schema.go` does not
exist yet.

Status: green
Timestamp: 2026-07-02T07:55:00Z

```
$ go test ./internal/connectors/engine -run TestSchema -v 2>&1 | tail -20
=== RUN   TestSchemaSecretKeys
--- PASS: TestSchemaSecretKeys (0.00s)
=== RUN   TestSchemaProperties
--- PASS: TestSchemaProperties (0.00s)
=== RUN   TestSchemaCompileErrorMessages
--- PASS: TestSchemaCompileErrorMessages (0.00s)
PASS
ok  	polymetrics.ai/internal/connectors/engine	0.367s
```

Implemented `internal/connectors/engine/schema.go`: minimal draft-07 subset compiler/validator,
zero imports outside stdlib (encoding/json, fmt, regexp, sort, strings). Compile-once (CompileSchema)
/ validate-many (Validate) split. Unknown keyword -> compile error. JSON-pointer-ish paths built
incrementally (`/field/nested/0`). `SecretKeys()`/`Properties()`/`PrimaryKeys()`/`CursorFieldName()`
accessors + `StreamSchema` wrapper per API-CONTRACT.md. `gofmt -l` clean, `go vet` clean.

## T-02

Status: red-confirmed
Timestamp: 2026-07-02T07:54:27Z

```
$ go test ./internal/connectors/engine -run 'TestInterpolate|TestEvalWhen|TestResolveCheck' -v 2>&1 | tail -30
# polymetrics.ai/internal/connectors/engine [polymetrics.ai/internal/connectors/engine.test]
internal/connectors/engine/interpolate_test.go:8:17: undefined: Vars
internal/connectors/engine/interpolate_test.go:9:9: undefined: Vars
internal/connectors/engine/interpolate_test.go:50:16: undefined: Interpolate
internal/connectors/engine/interpolate_test.go:63:12: undefined: Interpolate
internal/connectors/engine/interpolate_test.go:71:11: undefined: Interpolate
internal/connectors/engine/interpolate_test.go:112:16: undefined: InterpolatePath
internal/connectors/engine/interpolate_test.go:127:15: undefined: Interpolate
internal/connectors/engine/interpolate_test.go:139:13: undefined: Interpolate
internal/connectors/engine/interpolate_test.go:146:15: undefined: Interpolate
internal/connectors/engine/interpolate_test.go:156:15: undefined: Interpolate
internal/connectors/engine/interpolate_test.go:156:15: too many errors
FAIL	polymetrics.ai/internal/connectors/engine [build failed]
FAIL
```

RED confirmed: interpolate_test.go authored per PLAN.md T-02 scope + coordinator correction
(CRLF/header-injection rejection cases) — config/secrets/record/cursor resolution, dotted record
paths, urlencode default for path segments (traversal/metachar/space/unicode/double-encode-guard
cases), explicit filters (unix_seconds, base64, urlencode), CRLF injection rejection for both
InterpolateHeader and InterpolatePath, unresolved-key errors naming key+namespace, `when` grammar
(==, in [...], truthiness, unknown-operator compile error), and ResolveCheck static validation API.

Status: green
Timestamp: 2026-07-02T07:56:53Z

```
$ go test ./internal/connectors/engine -run TestInterpolate -v 2>&1 | tail -30
=== RUN   TestInterpolateFilters/explicit_urlencode_filter_in_non-path_context
--- PASS: TestInterpolateFilters (0.00s)
    --- PASS: TestInterpolateFilters/unix_seconds_on_rfc3339 (0.00s)
    --- PASS: TestInterpolateFilters/unix_seconds_on_bad_input_errors (0.00s)
    --- PASS: TestInterpolateFilters/base64 (0.00s)
    --- PASS: TestInterpolateFilters/explicit_urlencode_filter_in_non-path_context (0.00s)
=== RUN   TestInterpolateHeaderCRLFInjectionRejected
--- PASS: TestInterpolateHeaderCRLFInjectionRejected (0.00s)
PASS
ok  	polymetrics.ai/internal/connectors/engine	0.450s
```

Implemented `internal/connectors/engine/interpolate.go`: `Interpolate`/`InterpolatePath`/
`InterpolateHeader`/`EvalWhen`/`ResolveCheck` per API-CONTRACT.md (InterpolateHeader is an
additional helper, not in the contract table, added for THREAT-MODEL §2 header-injection guard
callers will need in read.go/auth.go). Key implementation notes:
- CR/LF is rejected on the RAW resolved reference value (pre-filter), for every namespace, in both
  `Interpolate` and `InterpolatePath` — not just as a post-hoc string scan of the final output
  (which would miss values escaped by the urlencode default). This closes the injection surface
  from THREAT-MODEL §2 uniformly.
- `urlencode` default for `InterpolatePath` uses a custom `urlencodeSegment` (QueryEscape with
  `+`->`%20`) so both metachars (`?`,`&`,`=`) and spaces encode as expected for path segments per
  TEST-PLAN §1.1; `%` is always re-encoded (double-encode guard).
- `when` grammar is a restricted hand-rolled parser: `==` (quoted literal), `in [...]` (quoted
  literal list), and bare truthiness; any other recognizable operator (`!=`,`>=`,`<=`,`>`,`<`,
  `&&`,`||`) is an explicit compile error rather than falling through to truthiness.
- `ResolveCheck` only checks `config.*` references against specKeys (record/secrets/cursor are not
  spec-declared); unknown namespace is also an error.
File size: 331 lines (design target ~150; the CRLF-guard additions, ResolveCheck, and the
hand-rolled when-parser exceed the estimate — flagging as a documented deviation, not a blocker,
since all T-02 acceptance cases pass and zero non-stdlib imports are used).
`gofmt -l` clean, `go vet` clean, full `go test ./internal/connectors/engine -v` green (schema +
interpolate suites).

## T-03

Status: red-confirmed
Timestamp: 2026-07-02T07:59:12Z

```
$ go test ./internal/connectors/engine -run TestBundle -v 2>&1 | tail -30
# polymetrics.ai/internal/connectors/engine [polymetrics.ai/internal/connectors/engine.test]
internal/connectors/engine/bundle_test.go:115:12: undefined: Load
internal/connectors/engine/bundle_test.go:162:12: undefined: Load
internal/connectors/engine/bundle_test.go:182:12: undefined: Load
internal/connectors/engine/bundle_test.go:195:12: undefined: Load
internal/connectors/engine/bundle_test.go:208:12: undefined: Load
internal/connectors/engine/bundle_test.go:221:12: undefined: Load
internal/connectors/engine/bundle_test.go:231:12: undefined: Load
internal/connectors/engine/bundle_test.go:251:12: undefined: Load
internal/connectors/engine/bundle_test.go:263:18: undefined: LoadAll
internal/connectors/engine/bundle_test.go:280:18: undefined: LoadAll
internal/connectors/engine/bundle_test.go:280:18: too many errors
FAIL	polymetrics.ai/internal/connectors/engine [build failed]
FAIL
```

RED confirmed: bundle_test.go authored per PLAN.md T-03 scope, using `fstest.MapFS` — happy path
full bundle, optional files absent (fixtures/ + no writes.json), streams.json optional iff
`capabilities.dynamic_schema`, dir-name/metadata.name mismatch, bad name regex
(`Source-GitHub`-shape), missing required file (api_surface.json), meta-schema violation
(metadata.json missing required `capabilities`), and `LoadAll` iterating multiple bundles +
tolerating an empty tree. Test function names prefixed `TestBundleLoad*` so the plan's verify
pattern (`-run TestBundle`) matches. Meta-schema JSON files authored first at
`internal/connectors/engine/schema/{metadata,spec,streams,writes,api_surface}.schema.json` in the
minimal draft-07 dialect (verified as valid JSON); they are consumed by the loader once `bundle.go`
exists.

Status: green
Timestamp: 2026-07-02T08:04:11Z

```
$ go test ./internal/connectors/engine -run TestBundle -v 2>&1 | tail -30
=== RUN   TestBundleLoadAllIteratesBundles
--- PASS: TestBundleLoadAllIteratesBundles (0.00s)
=== RUN   TestBundleLoadAllEmptyTreeIsFine
--- PASS: TestBundleLoadAllEmptyTreeIsFine (0.00s)
=== RUN   TestBundleLoadAllDefsFSEmpty
--- PASS: TestBundleLoadAllDefsFSEmpty (0.00s)
=== RUN   TestBundleLoadFromOnDiskTestdata
--- PASS: TestBundleLoadFromOnDiskTestdata (0.00s)
PASS
ok  	polymetrics.ai/internal/connectors/engine	0.457s
```

Implemented:
- `internal/connectors/engine/bundle.go` — `Bundle`, `Metadata`, `Capabilities`, `BatchSpec`,
  `RiskSpec`, `HTTPBase`, `RequestSpec`, `AuthSpec`, `PaginationSpec` (incl. coordinator-corrected
  `LastRecordField`/`StopPath`/`AllowCrossHost` fields), `ErrorRule`, `RateLimitSpec`, `StreamSpec`,
  `RecordsSpec`, `FilterSpec`, `IncrementalSpec`, `WriteAction`, `DeleteSpec`, `APISurface`,
  `SurfaceEndpoint`, `SurfaceCoverage`, `SurfaceExclusion` — field-for-field per API-CONTRACT.md/
  design §B.2/DATA-MODEL.md §2. `LoadAll(fsys fs.FS)`/`Load(fsys fs.FS, name string)` per the exact
  contract signatures.
- `internal/connectors/engine/metaschemas.go` — `//go:embed` of the 5 meta-schema JSON files as
  package-level strings, compiled once via `init()` into `metaSchemas` (a package var caching the
  5 compiled `*Schema`s + first compile error, since meta-schema compile failure is a programming
  error, not a per-bundle error).
- `internal/connectors/defs/defs.go` — `package defs`, single `//go:embed all:*`, `var FS embed.FS`.
  DESIGN DEVIATION NOTED (per task instructions): rather than adding a `.gitkeep`-style placeholder
  file, verified experimentally that `//go:embed all:*` on a directory containing only `defs.go`
  compiles successfully (embeds `defs.go` itself). `engine.LoadAll` already skips non-directory root
  entries via `fs.ReadDir` + `e.IsDir()`, so the stray embedded `defs.go` file is harmless and
  `LoadAll(defs.FS)` returns 0 bundles cleanly — proven by `TestBundleLoadAllDefsFSEmpty`, an
  engine-package test that imports `defs` directly (no import cycle: `defs` has zero deps).
- `internal/connectors/engine/testdata/bundles/widget-demo/**` — an on-disk (not just in-memory
  MapFS) loader test fixture bundle, exercised via `os.DirFS` in `TestBundleLoadFromOnDiskTestdata`,
  satisfying the PLAN.md instruction to place "loader testdata under
  internal/connectors/engine/testdata/bundles/".

Loader validation order per bundle: required-file existence check (metadata.json, spec.json,
api_surface.json, docs.md always required; streams.json required unless
`capabilities.dynamic_schema`) -> per-file meta-schema validation -> Go struct unmarshal -> semantic
checks (name regex, dir-name==metadata.name). Two test cases (`TestBundleLoadDirNameMismatch`,
`TestBundleLoadBadNameRegex`) were adjusted during GREEN to supply a complete bundle FS rather than
only `metadata.json`, so each test exercises its intended distinct error class instead of being
masked by the (also-true) missing-required-file error — this is a test-quality fix made during the
RED->GREEN step, not a scope change; the originally-authored RED state already failed to compile
for the correct reason (undefined `Load`).

`gofmt -l` clean, `go vet ./internal/connectors/engine/... ./internal/connectors/defs/...` clean.

## T-04

Status: red-confirmed
Timestamp: 2026-07-02T08:05:03Z

```
$ go test ./internal/connectors/engine -run TestError -v 2>&1 | tail -40
# polymetrics.ai/internal/connectors/engine [polymetrics.ai/internal/connectors/engine.test]
internal/connectors/engine/errors_test.go:13:8: undefined: Error
internal/connectors/engine/errors_test.go:32:8: undefined: Error
internal/connectors/engine/errors_test.go:39:8: undefined: Error
internal/connectors/engine/errors_test.go:54:9: undefined: Error
internal/connectors/engine/errors_test.go:77:17: undefined: applyErrorMap
internal/connectors/engine/errors_test.go:92:17: undefined: applyErrorMap
internal/connectors/engine/errors_test.go:107:17: undefined: applyErrorMap
internal/connectors/engine/errors_test.go:120:17: undefined: applyErrorMap
internal/connectors/engine/errors_test.go:128:17: undefined: applyErrorMap
internal/connectors/engine/errors_test.go:139:17: undefined: applyErrorMap
internal/connectors/engine/errors_test.go:139:17: too many errors
FAIL	polymetrics.ai/internal/connectors/engine [build failed]
FAIL
```

RED confirmed: errors_test.go authored per PLAN.md T-04 scope — `engine.Error` wraps
`*connsdk.HTTPError` reachable via `errors.As`, `Unwrap()` reachable via `errors.Is`, context fields
{connector, stream|action, page|record_index} surfaced in `Error()`, `applyErrorMap` matching by
status alone / status+match_body / match_body-mismatch-falls-through / no-match / non-HTTPError
input, hint surfacing verbatim, and a secret-redaction case (planted `sk_live_` token in both URL
query and body must not appear in `Error()`).

Status: green
Timestamp: 2026-07-02T08:06:39Z

```
$ go test ./internal/connectors/engine -run TestError -v 2>&1 | tail -30
=== RUN   TestErrorMapNonHTTPError
--- PASS: TestErrorMapNonHTTPError (0.00s)
=== RUN   TestErrorHintSurfacesVerbatim
--- PASS: TestErrorHintSurfacesVerbatim (0.00s)
=== RUN   TestErrorRedactsSecrets
--- PASS: TestErrorRedactsSecrets (0.00s)
PASS
ok  	polymetrics.ai/internal/connectors/engine	0.358s
```

Implemented `internal/connectors/engine/errors.go` per API-CONTRACT.md / design §F.4:
- `Error{Connector, Stream, Action, Page, RecordIndex, Class, Hint, Err}` with `Error()`/`Unwrap()`.
  `errors.As`/`errors.Is` reach the wrapped `*connsdk.HTTPError` (or any sentinel) through the
  standard `Unwrap() error` protocol — no custom errors.As reimplementation (uses stdlib
  `errors.As` inside `applyErrorMap` directly, per golang-error-handling skill guidance).
- `applyErrorMap(rules []ErrorRule, err error) (class, hint string)`: first rule whose `status`
  matches AND (if `match_body` set) whose body substring matches; non-`*connsdk.HTTPError` errors
  (checked via `errors.As`, so still reachable through arbitrary wrapping) never match, since
  error_map is an HTTP-status concept.
- `Error()` message construction: context fields rendered as `connector stream=x page=n record=m
  class=c: <redacted inner error>`, then error_map hint appended in parentheses VERBATIM (never
  redacted/truncated) per design §F.4 "hints surface to the CLI verbatim". The inner error text
  passes through `safety.RedactErrorText` (same helper `connsdk.HTTPError.Error()` already applies
  internally, so redaction is layered: the wrapped HTTPError redacts its own URL/body first, and the
  outer engine.Error message is redacted again as defense-in-depth for any additional secret-shaped
  substrings the context fields might carry).

Test function names in the T-04 error_map cases renamed from `TestApplyErrorMap*` to
`TestErrorMap*` during RED->GREEN so the plan's documented verify pattern (`-run TestError`) matches
every case (same rationale as the T-03 bundle test renames).

Full package status: `gofmt -l ./internal/connectors/engine` clean, `go vet` clean, all engine
package tests green, `go test -cover ./internal/connectors/engine` = 80.0% (T-01..T-04 files only;
the ≥85% EVAL-PLAN gate is a phase-exit metric evaluated once every engine file lands in later
waves, not a per-task gate for Wave A).
