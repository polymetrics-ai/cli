# S3 engine mini-wave (pre-fan-out) — TDD ledger

Executor: gsd-loop-backend. Branch `connector-architecture-v2`, HEAD at dispatch start
`20913329a311f0a209866e4941cac9959bb7d085`. Scope: the two carried `ENGINE_GAP` items from
wave1-pilot's SUMMARY.md ("Carried queue") + REVIEW-A.md re-review adjudications R1/R2/R3, plus the
carried minors list, plus a `docs/migration/conventions.md` update. No commits made (per dispatch
instructions).

Read first (per dispatch mandate): `.planning/phases/wave1-pilot/SUMMARY.md` (carried queue),
`REVIEW-A.md` "Re-review (gap loop cycle 1)" section (adjudications R1/R2/R3 + carried minors list),
`traces/s2-chargebee-sentry-ledger.md` (chargebee item 2's STOP evidence — the exact `read.go`
mechanics this dispatch had to change), `traces/s2-github-gmail-ledger.md` (github's `public_access`
adjudication + the `ResolveCheck` `==`/`in` gap discovery).

---

## Item 1 — incremental lower-bound query vars (`{{ incremental.lower_bound }}`)

### Design

`engine/read.go`'s `buildInitialQuery` previously resolved `stream.Query` (with `Vars.Cursor=""`,
no lower-bound reference at all) BEFORE computing the incremental lower bound — the S-2 STOP
(`s2-chargebee-sentry-ledger.md`) established this ordering is exactly why chargebee's
`sort_by[asc]=updated_at` (sent in the SAME branch as `updated_at[after]`, `chargebee.go:151-155`)
could not be expressed: the optional-query dialect's `omit_when_absent` is keyed to a
config/secrets reference resolving or not, and there is no config/secrets key that tracks "the
incremental lower bound resolved" on the state-cursor-driven repeat-sync path (an app-persisted
`State["cursor"]` is not a config key at all).

Fix shape: reorder `buildInitialQuery` to compute+format the lower bound FIRST, then resolve
`stream.Query` against `Vars` that carry the formatted value as `IncrementalLowerBound`. A new
`incremental` reference namespace (`interpolate.go`'s `resolveIncrementalRef`) resolves
`incremental.lower_bound` to that value, or to the SAME `unresolvedKeyError` shape (`Namespace:
"incremental"`) config/secrets absence already uses when the value is empty — this is what lets
`omit_when_absent`/`default` compose with it via the exact same classification helper
(`isUnresolvedConfigSecretOrIncremental`, extended from `isUnresolvedConfigOrSecret`), no new
tolerance mechanism required.

Since legacy sends a FIXED literal (`"updated_at"`, not the lower bound's own timestamp value)
gated on whether the lower bound resolves, a second small primitive was needed: a `const:<value>`
filter (discards the resolved value entirely, always returns the literal after the first `:`) so
`omit_when_absent` gates on `incremental.lower_bound` while the emitted value is the fixed literal
`chargebee.go` actually sends.

### RED (engine/read_test.go, engine/interpolate_test.go)

```
$ go test ./internal/connectors/engine/... -run 'TestReadIncrementalLowerBoundQueryVar' -v
--- FAIL: TestReadIncrementalLowerBoundQueryVarOmittedOnFreshFullSync
    read_test.go:431: Read: acme stream=widgets: engine: resolve query "sort_by[asc]":
    interpolate: unknown namespace "incremental" in reference "incremental.lower_bound"
--- FAIL: TestReadIncrementalLowerBoundQueryVarSentFromStateCursor (same error)
--- FAIL: TestReadIncrementalLowerBoundQueryVarSentFromStartConfigKey (same error)
--- FAIL: TestReadIncrementalLowerBoundQueryVarAbsentIncrementalSpecOmits (same error)
FAIL

$ go test ./internal/connectors/engine/... -run 'TestApplyFilterConst' -v
--- FAIL: TestApplyFilterConstReplacesResolvedValueWithFixedLiteral
    interpolate: unknown filter "const:updated_at"
--- PASS: TestApplyFilterConstStillFailsWhenReferenceUnresolved (trivially true pre-fix too)
--- FAIL: TestApplyFilterConstComposesWithOmitWhenAbsentQueryDialect
--- FAIL: TestApplyFilterConstKnownToResolveCheck
FAIL
```

### GREEN

- `engine/interpolate.go`: `Vars.IncrementalLowerBound string`; `resolveIncrementalRef` (new
  `incremental` namespace, key `lower_bound`, `knownIncrementalKeys` for static validation);
  `const:<value>` filter added to `applyFilterValue`/`isKnownFilter`.
- `engine/read.go`: `buildInitialQuery` reordered (lower bound computed+formatted BEFORE the
  `stream.Query` loop; `vars.IncrementalLowerBound` set before that loop runs;
  `stream.Incremental.RequestParam` set from the SAME `formattedLower` value afterward, unchanged
  behavior for existing bundles); `isUnresolvedConfigOrSecret` renamed
  `isUnresolvedConfigSecretOrIncremental`, extended to accept the `"incremental"` namespace.
- `defs/chargebee/streams.json`: all 5 streams gained
  `"sort_by[asc]": {"template": "{{ incremental.lower_bound | const:updated_at }}",
  "omit_when_absent": true}`.

```
$ go test ./internal/connectors/engine/... -run 'TestReadIncrementalLowerBoundQueryVar|TestApplyFilterConst' -v
--- PASS (all 8 subtests)

$ go test ./internal/connectors/engine/... -v   # full package, no regressions
ok

$ go run ./cmd/connectorgen validate internal/connectors/defs
connectorgen validate: 13 connector(s) checked, 0 findings
```

### Chargebee parity RED/GREEN (paritytest/chargebee/parity_test.go)

```
$ go test ./internal/connectors/paritytest/chargebee/... -run 'TestParityChargebee_SortByAsc' -v
--- FAIL: TestParityChargebee_SortByAscSentOnIncrementalFromState
    engine sort_by[asc] = "", want "updated_at" (legacy)
--- FAIL: TestParityChargebee_SortByAscSentOnIncrementalFromStartDate (same)
--- PASS: TestParityChargebee_SortByAscOmittedOnFullSync (trivially true pre-fix: absent both sides)
FAIL
```

After the streams.json fix:

```
$ go test ./internal/connectors/paritytest/chargebee/... -run 'TestParityChargebee_SortByAsc' -v
--- PASS (all 3)

$ go test ./internal/connectors/paritytest/chargebee/... -v
--- PASS (all subtests, full package)
```

`docs.md`'s "OPEN — sort_by[asc]" bullet flipped to strikethrough+RESOLVED (conventions.md §5
pattern); Streams-notes paragraph updated from "Not yet reproduced" to "RESOLVED".

---

## Item 2a — `ResolveCheck`/`ResolveCheckWhen`: full when-grammar static parsing

### Design

`ResolveCheck` (used for every `{{ }}` template EXCEPT the runtime-only `EvalWhen` path) parsed a
template's inner `{{ }}` content as a single dotted `namespace.key` reference — an `==`/`in`-shaped
`when` clause (`{{ config.auth_type == 'public' }}`) therefore had its ENTIRE inner text
(`config.auth_type == 'public'`) treated as one reference, split on `.` into
`namespace="config"`, `key="auth_type == 'public'"` — which then failed the `specKeys[key]` lookup
even when `auth_type` IS a declared spec property, since the whole comparison string was checked as
if it were the key. Verified via a throwaway probe before writing any test:

```
$ (probe) ResolveCheck("{{ config.auth_type == 'public' }}", {"auth_type": true})
  -> resolve check: unknown spec key "auth_type == 'public'" referenced as "config.auth_type == 'public'"
```

Fix: `ResolveCheckWhen(template, specKeys)` parses the IDENTICAL grammar `EvalWhen` evaluates at
runtime (`==`, ` in `, truthiness, plus the same rejected-operator set) and validates only the
left-hand-side reference via a new shared `checkNamespaceRef` helper (extracted from `ResolveCheck`'s
own reference-checking body so both entry points enforce identical namespace/key rules); the RHS
literal/list syntax is checked for well-formedness only (`parseLiteral`/`parseList`, reused
verbatim from `EvalWhen`'s own helpers), never checked against an enum (there is none to check
against). `ResolveCheckAuthSpec`'s `when` field now routes through `ResolveCheckWhen` instead of
`ResolveCheck`; every other AuthSpec field is unchanged.

### RED (engine/interpolate_test.go)

```
$ go test ./internal/connectors/engine/... -run 'TestResolveCheckWhen|TestResolveCheckAuthSpecWhenUsesFullGrammar' -v
# build failed: undefined: ResolveCheckWhen (8 call sites)
FAIL
```

### GREEN

`engine/interpolate.go`: `checkNamespaceRef` (shared reference-check helper, also used by the
existing `ResolveCheck`), `knownIncrementalKeys` (used by both `ResolveCheck` and
`ResolveCheckWhen` for the `incremental.*` namespace from item 1), `ResolveCheckWhen` (new,
mirrors `EvalWhen`'s grammar dispatch). `ResolveCheckAuthSpec`'s `when` field routed through it.

```
$ go test ./internal/connectors/engine/... -run 'TestResolveCheckWhen|TestResolveCheckAuthSpecWhenUsesFullGrammar|TestResolveCheck' -v
--- PASS (all subtests, including the pre-existing TestResolveCheckStillRejectsSpecUnknownKeyForWhenTemplates
    and the "when condition still checked" AuthSpec subtest — unaffected)

$ go test ./internal/connectors/engine/... -v
ok   # full package, zero regressions
```

### `cmd/connectorgen` corpus additions (RED verified against a simulated pre-fix build)

New fixtures: `testdata/invalid/when-clause-equality-unknown-spec-key/` (added to
`TestValidate_RejectsSeededInvalidBundles`'s table, rule `ruleInterpolationUnresolved`) and
`testdata/valid-extra/when-clause-equality-valid/` (new dedicated positive-regression test
`TestValidate_WhenClauseEqualityAndMembershipAgainstSpecKnownKeyPasses`, kept out of
`testdata/valid/` so it doesn't disturb `TestValidate_AcceptsGoodBundle`'s `ConnectorsChecked==1`
assertion).

```
# Simulated pre-fix (ResolveCheckAuthSpec's when field routed through plain ResolveCheck):
$ go test ./cmd/connectorgen/... -run TestValidate_WhenClauseEqualityAndMembershipAgainstSpecKnownKeyPasses -v
--- FAIL: expected zero findings for a spec-known ==/in when clause, got
  [... unknown spec key "auth_type == 'token'" ... unknown spec key "auth_type in [...]" ...]
FAIL

# Restored (real fix):
$ go test ./cmd/connectorgen/... -run 'TestValidate_WhenClauseEqualityAndMembershipAgainstSpecKnownKeyPasses|TestValidate_RejectsSeededInvalidBundles|TestValidate_AcceptsGoodBundle' -v
--- PASS (all)
```

### `internal/connectors/conformance/static.go` — the same gap existed here too (found via github's TestConformance regression)

Wiring github's `auth_type in [...]` candidate (item 2b below) immediately regressed
`TestConformance/github`'s `interpolations_resolve` check:

```
$ go test ./internal/connectors/conformance/... -run 'TestConformance/github' -v
--- FAIL: interpolations_resolve: resolve check: unknown spec key
  "auth_type in ['public', 'none', 'anonymous', 'unauthenticated']" referenced as
  "config.auth_type in [...]"
FAIL
```

`checkInterpolationsResolve` (conformance's OWN static check, separate from `cmd/connectorgen`'s)
called plain `engine.ResolveCheck` for `AuthSpec.When` too — same bug, independent call site. Fixed
identically (`checkWhen` helper using `ResolveCheckWhen`, wired for `a.When` only). Added a direct
unit test (`TestCheckInterpolationsResolve_AuthWhenClauseUsesFullGrammar`) since this call site has
no corpus-fixture equivalent to the cmd/connectorgen table.

```
$ go test ./internal/connectors/conformance/... -run 'TestConformance/github|TestCheckInterpolationsResolve_AuthWhenClauseUsesFullGrammar' -v
--- PASS (all)

$ go test ./internal/connectors/conformance/... -v
ok   # full package, zero regressions
```

---

## Item 2b — github `auth_type` decision: ADDITIVE restoration (not a replacement)

### Decision

Restored `auth_type` as an ADDITIONAL, purely-additive opt-in for the 4 legacy PUBLIC synonyms
(`public`/`none`/`anonymous`/`unauthenticated`) alongside the existing `public_access` boolean —
`public_access` remains the PRIMARY documented surface; `auth_type` is a second `mode:none`
candidate in `streams.json`'s `auth` list, not a replacement. Two separate candidates (not one
OR'd `when`) because `EvalWhen`'s grammar has no `||` operator (confirmed: `!=`/`>=`/`<=`/`>`/`<`/
`&&`/`||` are all explicitly rejected as unsupported) — `public_access` truthy OR `auth_type` in
the public enum cannot be expressed as a single `when` clause.

Scope is deliberately NARROW: only the 4 public-synonym STRING VALUES are reproduced, never
legacy's full `auth_type` mode-selection semantics (`auth_type=github_app` forcing app auth ahead
of a configured token; the token/oauth/actions/installation synonym distinctions, which this
bundle already collapses into one `bearer` candidate). Reproducing those would require rewriting
the auth candidate list's entire precedence model — a materially larger, auth-behavior-changing
scope than this dispatch's `when`-grammar-gap-closure mandate, and per the role brief's human-gate
list ("auth/security changes") that broader rework was NOT attempted here. This restoration is
purely additive (widens accepted input to a SECOND explicit opt-in reaching the SAME already
fail-loud-gated `mode:none` outcome) — no new implicit/silent path to unauthenticated reads exists;
verified by `TestParityGithub_AuthTypeUnrelatedValueDoesNotGrantPublicAccess` (an unrelated
`auth_type` value still hard-errors, same as no credentials at all).

### RED (paritytest/github/parity_test.go)

```
$ go test ./internal/connectors/paritytest/github/... -run 'TestParityGithub_AuthType' -v
--- FAIL: TestParityGithub_AuthTypePublicEnumOptIn/public
    engine Read with auth_type="public": engine: select auth: no auth spec matched for auth_type "public"
--- FAIL: .../none, .../anonymous, .../unauthenticated (same)
--- PASS: TestParityGithub_AuthTypeUnrelatedValueDoesNotGrantPublicAccess (trivially true pre-fix)
FAIL
```

### GREEN

`defs/github/streams.json`: added
`{"mode":"none","when":"{{ config.auth_type in ['public', 'none', 'anonymous', 'unauthenticated'] }}"}`
as a 4th `auth` candidate (after the existing `public_access`-gated one). `defs/github/spec.json`:
added `auth_type` property (documented scope: additive, narrow, non-public values inert).

```
$ go test ./internal/connectors/paritytest/github/... -run 'TestParityGithub_AuthType' -v
--- PASS (all 5 subtests)

$ go test ./internal/connectors/paritytest/github/... ./internal/connectors/hooks/github/... -v
--- PASS (full packages, zero regressions)

$ go test ./internal/connectors/conformance/... -run 'TestConformance/github' -v
--- PASS

$ go run ./cmd/connectorgen validate internal/connectors/defs
connectorgen validate: 13 connector(s) checked, 0 findings
```

`docs.md`'s "Auth setup" candidate-list paragraph, the auth-config-surface paragraph, and the
matching "Known limits" bullet all rewritten to describe both opt-ins, the additive/narrow scope,
and the `||`-grammar-gap reason for two candidates instead of one.

---

## Item 3 — carried minors

All 7 carried minors from SUMMARY.md addressed:

1. **chargebee blanket-stringify main-compare** — `TestParityChargebee_StreamRecords` flipped from
   `normalizeRecordsStringify` to raw `reflect.DeepEqual` on emitted records directly. Verified via
   probe (`TestProbeRawDeepEqualChargebee`, not committed) that raw DeepEqual already passes on
   every stream/record before flipping the real test — confirms the strengthening is safe, not
   accidentally red. `normalizeRecordStringify`/`normalizeRecordsStringify` deleted entirely
   (genuinely unused after the flip); `stringifyAny` kept (still used by `recordIDs`, an unrelated,
   still-legitimate use for pagination-order assertions).
   ```
   $ go test ./internal/connectors/paritytest/chargebee/... -v
   --- PASS (all subtests)
   ```
2. **github compound-write body assertions** — `TestParityGithub_WriteCreatePullRequestCompound`
   flipped from method/path-only compare to full `reflect.DeepEqual` on each captured request
   (method+path+body), matching every non-compound write test's bar.
   ```
   $ go test ./internal/connectors/paritytest/github/... -run TestParityGithub_WriteCreatePullRequestCompound -v
   --- PASS
   ```
3. **github close_issue/close_pull_request fixture bodies omit `state`** —
   `defs/github/fixtures/writes/close_issue.json`/`close_pull_request.json`'s `expect.body` now
   assert `"state": "closed"` (the WriteHook's `closeResource` always sends this; the fixture was
   silently under-asserting via `compareWriteExpectation`'s subset-check semantics). Verified as a
   genuine strengthening by sabotaging `hooks/github/hooks.go`'s `closeResource` payload (temporary,
   reverted) and re-running conformance:
   ```
   $ (sabotaged: payload := map[string]any{}) go test ./internal/connectors/conformance/... -run 'TestConformance/github' -v
   close_issue Passed:false Error:body missing key "state" (want closed)
   close_pull_request Passed:false Error:body missing key "state" (want closed)
   $ (restored) go test ./internal/connectors/conformance/... -run 'TestConformance/github' -v
   --- PASS
   ```
4. **monday max_pages declared but undeclared in spec.json** — `defs/monday/spec.json` gained the
   `max_pages` property (genuinely consumed by `hooks/monday/hooks.go`'s `mondayMaxPages`, never
   declared before); new lock-in test `TestParityMonday_MaxPagesDeclaredInSpec`, RED-verified
   against a version of spec.json with the key removed:
   ```
   $ (max_pages removed) go test ./internal/connectors/paritytest/monday/... -run TestParityMonday_MaxPagesDeclaredInSpec -v
   --- FAIL: spec.json properties = [...], want "max_pages" declared
   $ (restored) --- PASS
   ```
5. **sentry inert static per_page entries + check per_page=1 note** — removed the 4 streams' dead
   `query: {"per_page": "100"}` entries (every stream is `StreamHook`-handled, so
   `readDeclarative`'s `stream.Query` resolution never runs for any of them — the hook builds its
   own `per_page` from `config.page_size` independently); new lock-in test
   `TestParitySentry_NoInertStaticQueryEntries`, RED-verified against a version with the entries
   restored. `Check`'s `per_page=1` (legacy sends it, this bundle's `base.check` has no query field
   to express it — `engine.RequestSpec` declares only method/path) documented as an OPEN, verified
   BENIGN residual gap (not pursued as an engine increment — single-connector, non-functional
   payload-size optimization, does not meet the ≥3-occurrence recurrence bar).
   ```
   $ (entries restored) go test ./internal/connectors/paritytest/sentry/... -run TestParitySentry_NoInertStaticQueryEntries -v
   --- FAIL: stream "projects" declares query map[per_page:...], want none
   $ (removed, real fix) --- PASS
   ```
6. **gmail interpolateOptional comment drift** — the doc comment claimed CRLF-injection/
   unknown-filter errors "still propagate", but the code (`if err != nil { return "" }`) swallows
   EVERY interpolation error identically. Fixed the COMMENT to match the code (not the code itself —
   verified benign: `interpolateOptional`'s only 2 call sites, `client_secret`/`scope`, are optional
   OAuth POST-form values per legacy's own "omit when unset" semantics, not headers/paths). New
   direct unit test `TestInterpolateOptional_AnyErrorResolvesToEmptyString` pins the actual behavior
   (absent key, CRLF value, and unknown filter all resolve to `""`; a clean value passes through) so
   the comment can't silently drift from the code again.
   ```
   $ go test ./internal/connectors/hooks/gmail/... -run TestInterpolateOptional_AnyErrorResolvesToEmptyString -v
   --- PASS (all 4 subtests)
   ```
7. **github hooks.go at exactly the 400-line hard ceiling** — re-evaluated for safe reduction;
   concluded STANDING EXCEPTION (documented in `defs/github/docs.md` Known limits, citing
   REVIEW-A.md's re-review disposition table + SUMMARY.md's carried-minors list): the prior repair
   round already trimmed 424->400 by removing redundant comment prose and collapsing 3 near-
   identical `updateLabel` ifs into a loop; remaining comments are load-bearing documentation
   (non-obvious legacy-parity rationale, not restated code), remaining blank lines are single
   standard `gofmt`-conventional separators (`gofmt -l` reports the file clean), and the one
   candidate logic consolidation (merging `createLabel`/`updateLabel`) was rejected as a net-negative
   trade (different validation rules; merging adds coupling/risk for marginal line savings — exactly
   the "gaming" the dispatch warned against). No functional change made; `wc -l` confirms 400
   unchanged.

Full re-verify after all 7 minors:
```
$ go build ./... && go run ./cmd/connectorgen validate internal/connectors/defs
BUILD OK; connectorgen validate: 13 connector(s) checked, 0 findings
$ go test -count=1 ./internal/connectors/... ./cmd/...
ok (zero FAIL lines across the entire fleet + cmd/connectorgen + cmd/inventorygen)
$ make lint
0 issues.
```

---

## Item 4 — `docs/migration/conventions.md` updates

§3 "The engine dialect reference" updated:
- Template-references list: added `incremental.lower_bound`.
- Filters list: added `const:<value>`.
- New `when` grammar paragraph cross-reference: `ResolveCheckAuthSpec` now routes `token`/
  `username`/`password`/`value`/`token_url`/`client_id`/`client_secret`/`scopes` through
  `ResolveCheck` and `when` through the new `ResolveCheckWhen`.
- New `**ResolveCheckWhen**` subsection: explains the prior bug (whole `==`/`in` expression treated
  as one dotted reference), the fix (parses the identical `EvalWhen` grammar, validates only the
  LHS via the shared `checkNamespaceRef` helper), and that `conformance/static.go`'s
  `checkInterpolationsResolve` must stay wired the same way as `cmd/connectorgen`'s
  `checkInterpolations`.
- New `**{{ incremental.lower_bound }} in stream.Query**` subsection: the ordering fix in
  `buildInitialQuery`, why the pre-existing `omit_when_absent`/`default` dialect alone couldn't
  express this (state-cursor path has no config/secret key to gate on), and the `const:<value>`
  pairing for a fixed-literal-value param (chargebee's exact shape) vs. a param that needs the
  lower bound's own value.
- `stream.Query` section: extended the `omit_when_absent`/absence-classification prose to name the
  `incremental` namespace alongside `config`/`secrets` (`isUnresolvedConfigSecretOrIncremental`).

---

## Final combined self-verify (whole S3 dispatch)

```
$ go build ./...
(clean)

$ go test -count=1 ./internal/connectors/... ./cmd/...
ok (every package; zero FAIL lines; only expected "[no test files]" notices)

$ go run ./cmd/connectorgen validate internal/connectors/defs
connectorgen validate: 13 connector(s) checked, 0 findings

$ make lint
golangci-lint run ./internal/connectors/engine/... ./internal/connectors/defs/... \
  ./internal/connectors/hooks/... ./internal/connectors/native/... \
  ./internal/connectors/conformance/... ./internal/connectors/certify/... \
  ./cmd/connectorgen/... ./cmd/inventorygen/...
0 issues.
```

No dependency additions, no schema migrations (JSON-Schema bundle "schemas" only), no auth/security
WEAKENING (github's `auth_type` restoration is strictly additive — a second explicit, already
fail-loud-gated opt-in, never a new implicit/silent path), no destructive data actions, no secret
access beyond what existing AuthHook/paritytest fixtures already exercised, no quality-gate
reductions. No commits made (per dispatch instructions).

## Files touched

- `internal/connectors/engine/interpolate.go`, `interpolate_test.go`
- `internal/connectors/engine/read.go`, `read_test.go`
- `internal/connectors/conformance/static.go`, `static_test.go`
- `internal/connectors/defs/chargebee/streams.json`, `docs.md`
- `internal/connectors/defs/github/streams.json`, `spec.json`, `docs.md`,
  `fixtures/writes/close_issue.json`, `fixtures/writes/close_pull_request.json`
- `internal/connectors/defs/monday/spec.json`, `docs.md`
- `internal/connectors/defs/sentry/streams.json`, `docs.md`
- `internal/connectors/hooks/gmail/hooks.go`, `hooks_test.go`
- `internal/connectors/paritytest/chargebee/parity_test.go`
- `internal/connectors/paritytest/github/parity_test.go`
- `internal/connectors/paritytest/monday/parity_test.go`
- `internal/connectors/paritytest/sentry/parity_test.go`
- `cmd/connectorgen/main_test.go` +
  `cmd/connectorgen/testdata/invalid/when-clause-equality-unknown-spec-key/**` (new) +
  `cmd/connectorgen/testdata/valid-extra/when-clause-equality-valid/**` (new)
- `docs/migration/conventions.md`
- `.planning/phases/wave2-fanout-http-sm/traces/s3-engine-ledger.md` (this file)
