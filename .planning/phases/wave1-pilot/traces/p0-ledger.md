# p0-ledger — wave1-pilot P-0 (pre-pilot engine follow-ups: N1, N2, N4)

Scope: `internal/connectors/conformance/**` (dynamic.go + dynamic_test.go + new self-test
bundles), `cmd/connectorgen/**` (validate.go + main.go + main_test.go + new invalid-corpus
bundles), `internal/connectors/engine/read.go` (COMMENT-ONLY, N4). Strict TDD per item: RED
recorded here BEFORE the fix, then GREEN. No `git commit` performed by this task.

Note on scope vs. dispatch instructions: PLAN.md's P-0 section and SPEC.md §4 scope P-0 to
**N1 + N4 only** and explicitly say N2 is "noted, not blocking... promote to a validate-time
guard only if a pilot actually hits it" (SPEC.md §4) — no pilot has landed yet. The dispatch
instructions for this task explicitly asked for N2 as well, with an allowance to "pick the
narrow, honest check per SPEC/PLAN guidance and document." Given the direct conflict, N2 is
implemented here as a narrow, additive, WARNING-only (never blocking) connectorgen rule that
satisfies the dispatch instruction without contradicting SPEC.md's "not blocking" instruction —
it adds no new hard `Finding`, never changes validate's exit code, and the "0 findings"
self-verify contract for the 3 real goldens is unaffected (see N2 GREEN evidence below).

---

## N1 — `formatCursorForAssertion` `github_date_range` alignment

### Diagnosis

`conformance/dynamic.go`'s `formatCursorForAssertion`'s `github_date_range` branch returned
`">=" + value` VERBATIM. The engine's real `formatParam` (`engine/read.go:429-439`, post-R1/B1)
instead: parses `value` via `parseLowerBoundTime` (digits-only -> Unix seconds, else RFC3339),
then re-emits `">=" + t.UTC().Format(time.RFC3339)` — i.e. it ALWAYS normalizes to UTC second
precision regardless of the input cursor's shape. The verbatim assertion-side version only
happened to agree with the engine when `value` was ALREADY exactly UTC-normalized RFC3339 with
no fractional seconds; a numeric (Unix-seconds digit string) cursor or a non-UTC-offset RFC3339
cursor would falsely FAIL `cursor_advances`.

### RED

New self-test bundles (modeled on `testdata/good/acme-numeric-cursor`, the B2 pattern), plus two
new tests in `dynamic_test.go`:

- `testdata/good/acme-github-range-cursor/` — `events` stream, `x-cursor-field: created` (JSON
  NUMBER, Unix seconds), `param_format: github_date_range`. New test
  `TestCursorAdvances_GitHubDateRangeNumericCursorNormalized`.
- `testdata/good/acme-github-range-cursor-offset/` — `events` stream, `x-cursor-field:
  updated_at` (JSON STRING, RFC3339 with a non-UTC `+05:30` offset), `param_format:
  github_date_range`. New test `TestCursorAdvances_GitHubDateRangeNonUTCOffsetCursorNormalized`.

```
$ go test ./internal/connectors/conformance/... -run 'TestCursorAdvances_GitHubDateRange' -v
=== RUN   TestCursorAdvances_GitHubDateRangeNumericCursorNormalized
    dynamic_test.go:111: cursor_advances failed on github_date_range with a numeric (Unix-seconds) cursor: re-read request_param "created_range" = ">=2023-11-14T22:15:00Z", want ">=1700000100" (cursor "1700000100", param_format "github_date_range")
--- FAIL: TestCursorAdvances_GitHubDateRangeNumericCursorNormalized (0.00s)
=== RUN   TestCursorAdvances_GitHubDateRangeNonUTCOffsetCursorNormalized
    dynamic_test.go:127: cursor_advances failed on github_date_range with a non-UTC-offset RFC3339 cursor: re-read request_param "updated_range" = ">=2023-11-14T16:45:00Z", want ">=2023-11-14T22:15:00+05:30" (cursor "2023-11-14T22:15:00+05:30", param_format "github_date_range")
--- FAIL: TestCursorAdvances_GitHubDateRangeNonUTCOffsetCursorNormalized (0.00s)
FAIL
FAIL	polymetrics.ai/internal/connectors/conformance	0.454s
FAIL
```

Note the failure direction: `got` (`>=2023-11-14T22:15:00Z` / `>=2023-11-14T16:45:00Z`) is what
the REAL engine actually sent (correctly UTC-normalized); `want` (the pre-fix
`formatCursorForAssertion` output) is the un-normalized verbatim value — exactly the N1 defect
(the assertion side was wrong, not the engine).

### Fix

`formatCursorForAssertion`'s `github_date_range` case now calls
`parseLowerBoundTimeForAssertion(value)` (the same helper already used by the `unix_seconds`/
`date` branches, itself already mirroring `engine/read.go`'s `parseLowerBoundTime`) and emits
`">=" + t.UTC().Format(time.RFC3339)`, byte-for-byte mirroring `formatParam`'s
`github_date_range` case. No change to `unix_seconds`/`date`/`rfc3339` branches.

### GREEN

```
$ go test ./internal/connectors/conformance/... -run 'TestCursorAdvances' -v
=== RUN   TestCursorAdvances_PostReadCursorIsMaxFixtureCursorAndReReadSendsParam
--- PASS
=== RUN   TestCursorAdvances_NumericCursorFieldSupported
--- PASS
=== RUN   TestCursorAdvances_StringCursorFieldStillSupported
--- PASS
=== RUN   TestCursorAdvances_GitHubDateRangeNumericCursorNormalized
--- PASS
=== RUN   TestCursorAdvances_GitHubDateRangeNonUTCOffsetCursorNormalized
--- PASS
PASS
ok  	polymetrics.ai/internal/connectors/conformance	0.374s
```

Full package regression + both new bundles' full `RunBundle` (static+dynamic) suite also verified
green (`TestDynamicChecks_GoodBundleAllPass`-style scratch check run and discarded; full package
`go test ./internal/connectors/conformance/... -v` below).

---

## N2 — validate-time WARNING for digit-shaped non-unix `start_config_key` values

### Diagnosis / scope decision

Per SPEC.md §4 and N2's own carried-flag text (wave0 REVIEW.md re-review): `formatParam`'s
digits-passthrough (B1) is CORRECT and intentional for `param_format: unix_seconds` — a
digit-shaped config value there really does mean Unix seconds, no misinterpretation risk.
The genuine risk is narrower: `param_format: date` and `param_format: github_date_range` also
route through `parseLowerBoundTime` (an all-digits value is STILL silently treated as Unix
seconds there), but for these two formats a digit-shaped config value is much more likely to be
an operator typo (e.g. a yyyymmdd `"20260101"` for a `start_date` field) than a genuine
intentional Unix-seconds value. `rfc3339` never attempts digit parsing at all (verbatim
passthrough), so it is out of scope entirely.

Chose a **narrow, honest, WARNING-level** rule, not an error: `checkIncrementalStartDateFormat`
flags a stream whose `incremental.param_format` is `date`/`github_date_range` AND whose
`start_config_key` names a spec.json property that declares NO date-ish `format` annotation
(`date-time`/`date`). A property that DOES declare `format: date-time` is not flagged — an
operator filling in a spec-declared timestamp field typing a bare yyyymmdd digit string is the
much rarer case this heuristic accepts as a false-negative trade-off, in exchange for zero false
positives against any bundle that already documents its start_date's shape. This is
`Report.Warnings`, a field kept SEPARATE from `Report.Findings`, specifically so it can never
regress the "0 findings" self-verify contract or validate's exit code — a plausibility heuristic
must never gate the pipeline the way a structural defect does.

### RED

New invalid-corpus bundle `cmd/connectorgen/testdata/invalid/start-date-free-form-string/`
(`events` stream, `param_format: github_date_range`, `start_config_key: start_date`,
`spec.json`'s `start_date` property has NO `format` annotation) plus a no-false-positive
companion `start-date-rfc3339-format-no-warning/` (identical shape, but `start_date` DOES
declare `format: date-time`). Three new tests in `main_test.go`:
`TestValidate_StartDateFreeFormStringWarns`, `TestValidate_StartDateWithDateTimeFormatNoWarning`,
`TestValidate_UnixSecondsStartDateNeverWarns` (the last reuses the REAL stripe golden, whose
`start_date` has no `format` annotation either but uses `param_format: unix_seconds`, proving the
rule's `unix_seconds`-exclusion directly against production defs, not just a synthetic fixture).

Before `Report.Warnings`/`ruleStartDateFreeFormString` existed, this was a compile failure (no
behavior to even exercise yet — the honest RED for a task that adds a new field/rule from
scratch):

```
$ go test ./cmd/connectorgen/... -run 'TestValidate_StartDateFreeFormStringWarns|TestValidate_StartDateWithDateTimeFormatNoWarning|TestValidate_UnixSecondsStartDateNeverWarns' -v
# polymetrics.ai/cmd/connectorgen [polymetrics.ai/cmd/connectorgen.test]
cmd/connectorgen/main_test.go:163:24: report.Warnings undefined (type Report has no field or method Warnings)
cmd/connectorgen/main_test.go:164:13: report.Warnings undefined (type Report has no field or method Warnings)
cmd/connectorgen/main_test.go:164:98: undefined: ruleStartDateFreeFormString
...
FAIL	polymetrics.ai/cmd/connectorgen [build failed]
FAIL
```

### Fix

- `Report` (validate.go) gains `Warnings []Finding`, deliberately never merged into `Findings`.
- `validateDir`/`validateBundleDir` now return/aggregate/sort both lists independently.
- New rule `ruleStartDateFreeFormString = "start_date_free_form_string"`.
- New `checkIncrementalStartDateFormat(b engine.Bundle) []Finding`: for every stream with
  `Incremental != nil && StartConfigKey != "" && dateShapedParamFormats[ParamFormat]` (date /
  github_date_range only), parses `b.RawSpec` (F5's verbatim spec.json bytes — the compiled
  `*engine.Schema` does not expose the `format` annotation keyword through any accessor;
  schema.go documents `format` as accepted-but-only-preserved) to check
  `properties.<start_config_key>.format`; warns iff that format is not in
  `{"date-time","date"}`. De-duplicates per `start_config_key` (a shared config key referenced
  by multiple streams warns once).
- `main.go`'s `renderText` now also prints warnings after the (unchanged-wording) findings
  summary line, under a distinct `connectorgen validate: N warning(s)` line — the exit code
  (`runValidate`) still keys off `len(report.Findings)` only, unchanged.

### GREEN

```
$ go test ./cmd/connectorgen/... -run 'TestValidate_StartDateFreeFormStringWarns|TestValidate_StartDateWithDateTimeFormatNoWarning|TestValidate_UnixSecondsStartDateNeverWarns' -v
=== RUN   TestValidate_StartDateFreeFormStringWarns
--- PASS: TestValidate_StartDateFreeFormStringWarns (0.00s)
=== RUN   TestValidate_StartDateWithDateTimeFormatNoWarning
--- PASS: TestValidate_StartDateWithDateTimeFormatNoWarning (0.00s)
=== RUN   TestValidate_UnixSecondsStartDateNeverWarns
--- PASS: TestValidate_UnixSecondsStartDateNeverWarns (0.00s)
PASS
ok  	polymetrics.ai/cmd/connectorgen	0.426s

$ go run ./cmd/connectorgen validate internal/connectors/defs
connectorgen validate: 3 connector(s) checked, 0 findings
```

No warning line printed for the real defs corpus (stripe's `start_date` is `param_format:
unix_seconds`, out of this rule's scope by design — confirmed by
`TestValidate_UnixSecondsStartDateNeverWarns`), so the "0 findings" self-verify contract and the
3-goldens-unaffected requirement both hold.

---

## N4 — stale doc comment on `incrementalLowerBoundValue`

### Fix (docs-only, no test pair per PLAN.md)

`engine/read.go`'s `incrementalLowerBoundValue` doc comment said the lower bound is "always
RFC3339 when present" — false since B1's digits-passthrough fix (a state cursor may be an
all-digits Unix-seconds string, the app-persisted shape for a numeric cursor field). Rewrote the
comment to describe the digits-or-RFC3339 reality and point at B1/N4. Proven comment-only via
`git diff -w` (whitespace-insensitive diff) showing zero non-blank-line changes outside `//`
comment lines:

```
$ git diff -w -- internal/connectors/engine/read.go
--- a/internal/connectors/engine/read.go
+++ b/internal/connectors/engine/read.go
@@ -377,9 +377,13 @@ func buildInitialQuery(stream StreamSpec, req connectors.ReadRequest) (url.Value
 	return q, nil
 }

-// incrementalLowerBoundValue returns the raw (unformatted, always RFC3339
-// when present) incremental lower bound: the state cursor if set, else the
-// start_config_key config value, else "" (full sync / no lower bound).
+// incrementalLowerBoundValue returns the raw (unformatted) incremental lower
+// bound: the state cursor if set, else the start_config_key config value,
+// else "" (full sync / no lower bound). This value is NOT always RFC3339
+// (N4, wave0 REVIEW.md re-review): the state cursor may be an all-digits
+// Unix-seconds string (the app-persisted shape for a numeric cursor field,
+// B1) or an RFC3339 timestamp (a string cursor field, or a config
+// start_date); formatParam/parseLowerBoundTime accept both shapes.
 // client_filtered streams (no server-side request_param) still need this
 // value to drop old records client-side.
 func incrementalLowerBoundValue(stream StreamSpec, req connectors.ReadRequest) (string, error) {

$ git diff -w -- internal/connectors/engine/read.go | grep -E '^\+|^-' | grep -v -e '^+++' -e '^---' -e '^+//' -e '^-//'
(empty — every changed line is a // comment line)
```

Verified behavior-unchanged: `go build ./... && go test ./internal/connectors/engine/...` green.

---

## Full self-verify (all three items combined)

```
$ go build ./... && go test ./internal/connectors/conformance ./cmd/connectorgen ./internal/connectors/engine 2>&1 | tail -5
ok  	polymetrics.ai/internal/connectors/conformance	0.423s
ok  	polymetrics.ai/cmd/connectorgen	0.516s
ok  	polymetrics.ai/internal/connectors/engine	(cached)

$ make lint
golangci-lint run ./internal/connectors/engine/... ./internal/connectors/defs/... ./internal/connectors/hooks/... ./internal/connectors/native/... ./internal/connectors/conformance/... ./internal/connectors/certify/... ./cmd/connectorgen/... ./cmd/inventorygen/...
0 issues.

$ go run ./cmd/connectorgen validate internal/connectors/defs
connectorgen validate: 3 connector(s) checked, 0 findings

$ go vet ./...
(exit 0)

$ gofmt -l internal/connectors/conformance internal/connectors/engine cmd/connectorgen
(empty — clean)

$ go build ./... && go test ./... 2>&1 | grep -v '^ok'
?   	polymetrics.ai/cmd/pm	[no test files]
?   	polymetrics.ai/cmd/pm-cataloggen	[no test files]
?   	polymetrics.ai/cmd/registrygen	[no test files]
?   	polymetrics.ai/internal/connectors/defs	[no test files]
?   	polymetrics.ai/internal/connectors/hooks/hookset	[no test files]
?   	polymetrics.ai/internal/connectors/native/nativeset	[no test files]
?   	polymetrics.ai/internal/coordination	[no test files]
(no failures anywhere in the full repo)
```

## Files touched (exclusively within the permitted set)

- `internal/connectors/conformance/dynamic.go` (N1 fix)
- `internal/connectors/conformance/dynamic_test.go` (N1 RED tests)
- `internal/connectors/conformance/testdata/good/acme-github-range-cursor/**` (N1 fixture, new)
- `internal/connectors/conformance/testdata/good/acme-github-range-cursor-offset/**` (N1 fixture, new)
- `cmd/connectorgen/validate.go` (N2 rule + `Report.Warnings`)
- `cmd/connectorgen/main.go` (N2 warning rendering, exit code unchanged)
- `cmd/connectorgen/main_test.go` (N2 RED tests)
- `cmd/connectorgen/testdata/invalid/start-date-free-form-string/**` (N2 fixture, new)
- `cmd/connectorgen/testdata/invalid/start-date-rfc3339-format-no-warning/**` (N2 fixture, new)
- `internal/connectors/engine/read.go` (N4, comment-only — proven via `git diff -w`)

No `git commit` performed. No files touched outside the permitted set.
