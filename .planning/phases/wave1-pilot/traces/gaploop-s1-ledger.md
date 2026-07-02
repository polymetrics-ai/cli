# Gap-loop cycle 1 — Step 1 (engine mini-wave) TDD ledger

Executor: gsd-loop-backend. HEAD at start: d96253a, branch connector-architecture-v2.
Scope: internal/connectors/engine/**, cmd/connectorgen/** (rules+corpus), docs/migration/conventions.md.
Mandated by GAP-LOOP-PLAN.md Step 1 (adjudications A1/A3/C1/C3 in REVIEW-A.md; adjudications 2-3 +
zendesk flag in REVIEW-B.md).

Per-item plan: RED test(s) first (recorded below with failing output), then GREEN behavior change,
then re-run green. No pilot bundle/paritytest file is edited (Step 2 scope) — where item 1's
semantics change could break a pilot's string-lock-in test, this ledger records what WOULD break
for the Step-2 dispatcher instead of editing it here.

---

## Item 1 — typed computed_fields extraction (bare `{{ record.path }}` copies raw typed value)
## Item 2 — wire Config (not Secrets) into applyComputedFields Vars (combined per A3: same function)

Tests added to `internal/connectors/engine/read_test.go`:
- `TestReadComputedFieldsBareRecordPathPreservesNativeType` (item 1)
- `TestReadComputedFieldsFilteredOrMixedTemplateKeepsStringSemantics` (item 1, negative/lock-in half)
- `TestReadComputedFieldsConfigReference` (item 2 / A3 G0)
- `TestReadComputedFieldsSecretsNotAccessible` (item 2 threat-model line: secrets must stay excluded)

### RED

```
$ go test ./internal/connectors/engine -run 'TestReadComputedFieldsBareRecordPathPreservesNativeType|TestReadComputedFieldsFilteredOrMixedTemplateKeepsStringSemantics|TestReadComputedFieldsConfigReference|TestReadComputedFieldsSecretsNotAccessible' -v
=== RUN   TestReadComputedFieldsBareRecordPathPreservesNativeType
    read_test.go:617: count_typed = "42" (string), want a native number (json.Number/float64/int), not a string
--- FAIL: TestReadComputedFieldsBareRecordPathPreservesNativeType (0.00s)
=== RUN   TestReadComputedFieldsFilteredOrMixedTemplateKeepsStringSemantics
--- PASS: TestReadComputedFieldsFilteredOrMixedTemplateKeepsStringSemantics (0.00s)   # already-correct baseline behavior, kept as lock-in
=== RUN   TestReadComputedFieldsConfigReference
    read_test.go:696: Read: acme stream=widgets page=0: engine: computed_fields "repository": interpolate: unresolved key "owner" in config
--- FAIL: TestReadComputedFieldsConfigReference (0.00s)
=== RUN   TestReadComputedFieldsSecretsNotAccessible
--- PASS: TestReadComputedFieldsSecretsNotAccessible (0.00s)   # already-correct baseline (secrets.* already hard-errors); kept as regression guard
FAIL
```

Both item-1 and item-2 positive tests fail exactly as expected (string instead of native type; unresolved
config key). The two negative/lock-in tests already pass, confirming they describe behavior that must
NOT regress once the fix lands.

### GREEN

Implementation (`internal/connectors/engine/read.go`):
- `readDeclarative`'s call site now passes `req.Config.Config` into `applyComputedFields`.
- `applyComputedFields(projected, raw, cfg map[string]string, computed map[string]string) error` —
  new `cfg` parameter; `Vars{Record: raw, Config: cfg}` (Secrets deliberately never populated here —
  a `secrets.*` reference inside computed_fields therefore keeps hard-erroring exactly as before,
  since `vars.Secrets` is always nil in this call path).
- New `bareRecordPathReference(tmpl string) (path string, ok bool)`: detects a computed_fields
  template that is EXACTLY one `{{ record.<path> }}` expression spanning the whole string, no filter
  stage (`|`), no surrounding literal text, no second `{{ }}` occurrence. When true,
  `applyComputedFields` resolves the raw typed value via `resolveRecordPathValue` directly
  (bypassing `Interpolate`'s stringify) and writes it into `projected` unchanged (number/bool/null/
  object/array preserved). Any other shape (filter chain, mixed template, static literal,
  `config.*`/`cursor` bare reference) is UNCHANGED — still routed through `Interpolate` and still
  produces a string.

```
$ go test ./internal/connectors/engine/... -run 'TestReadComputedFields' -v
=== RUN   TestReadComputedFieldsNestedExtraction
--- PASS
=== RUN   TestReadComputedFieldsMissingIntermediateDoesNotPanic
--- PASS
=== RUN   TestReadComputedFieldsStaticLiteralNoTemplate
--- PASS
=== RUN   TestReadComputedFieldsJoinFilterArrayField
--- PASS
=== RUN   TestReadComputedFieldsBareRecordPathPreservesNativeType
--- PASS
=== RUN   TestReadComputedFieldsFilteredOrMixedTemplateKeepsStringSemantics
--- PASS
=== RUN   TestReadComputedFieldsConfigReference
--- PASS
=== RUN   TestReadComputedFieldsSecretsNotAccessible
--- PASS
PASS
```

`go build ./...` clean. `go test ./internal/connectors/engine/... ./cmd/connectorgen/...` all pass.

### Pilot-suite impact (Step-2 handoff — NOT edited here per scope)

`go test ./internal/connectors/...` after this change surfaces exactly the breakage REVIEW-A.md A1
predicted and explicitly scoped to Step 2 ("pilots must be re-tightened once the engine feature
lands"):

- `internal/connectors/paritytest/chargebee`: `TestParityChargebee_ComputedFieldsStringifyNumericAndBooleanFields`
  FAILS — `customers[0].created_at` is now `json.Number` (native), test still asserts the old
  stringified form. chargebee's bare `{{ record.created_at }}`-style computed fields (~30 fields per
  A1) now get typed extraction; schemas still declare `["string","null"]` widened types.
- `internal/connectors/paritytest/gmail`: `TestParityGmail_ComputedFieldsStringifyLabelCountFields`
  FAILS — `labels[0].messages_total` is now `json.Number`, test still asserts the old stringified
  form. Same class (gmail's 4 stringify-widened fields, A1).
- `internal/connectors/conformance` `TestConformance/chargebee`: `records_match_schema` now FAILS for
  the `customers` stream — `/created_at: value does not match type [string null]` — schema.json still
  declares the widened `["string","null"]` type; the emitted value is now a native number.
- `internal/connectors/paritytest/github`: NO breakage. github's nested-id bare computed fields
  (`user_id`/`author_id`/`committer_id`/`workflow_run_id`) also now emit native numbers, but
  `isStringifiedNestedID`'s comparison already uses `fmt.Sprint(engVal) != fmt.Sprint(legacyVal)`
  (string-form comparison on both sides), which tolerates the type change transparently — github's
  suite stays green and its G0b ledger entry becomes stale-but-harmless (Step 2/P-12 doc cleanup, not
  a test break).

Per the dispatch instructions, chargebee/gmail schema + parity-test + docs.md edits are explicitly
Step-2 (pilot repair wave) scope, listed here for the Step-2 dispatcher, not touched by this agent.

---

## Item 3 — optional-query dialect (`stream.Query` entry may be `{template, omit_when_absent, default?}`)

Per REVIEW-B.md cross-cutting adjudication 2: per-entry opt-in on `stream.Query`; string entries
keep today's exact hard-error semantics; object entries get `when`-grammar absent-key-falsy
resolution (omit the param when the resolved value is empty) plus an optional `default` literal.
Static validation stays strict: the referenced key must still be declared in spec.json.

Tests added:
- `internal/connectors/engine/bundle_test.go`: `TestBundleLoadParsesOptionalQueryDialect` (JSON
  decode: plain string entry vs `{template, omit_when_absent, default}` object entry).
- `internal/connectors/engine/read_test.go`: `TestReadOptionalQueryOmittedWhenConfigAbsent`,
  `TestReadOptionalQuerySentWhenConfigPresent`, `TestReadOptionalQueryDefaultAppliedWhenConfigAbsent`,
  `TestReadOptionalQueryDefaultOverriddenByConfig`, `TestReadQueryStringEntryStillHardErrorsOnAbsentConfig`
  (lock-in: string entries keep hard-erroring).

### RED

```
$ go vet ./internal/connectors/engine/...
# polymetrics.ai/internal/connectors/engine [polymetrics.ai/internal/connectors/engine.test]
internal/connectors/engine/bundle_test.go:506:24: staticEntry.Template undefined (type string has no field or method Template)
internal/connectors/engine/bundle_test.go:506:57: staticEntry.OmitWhenAbsent undefined (type string has no field or method OmitWhenAbsent)
internal/connectors/engine/bundle_test.go:510:24: statusEntry.Template undefined (type string has no field or method Template)
internal/connectors/engine/bundle_test.go:510:74: statusEntry.OmitWhenAbsent undefined (type string has no field or method OmitWhenAbsent)
internal/connectors/engine/bundle_test.go:514:23: countEntry.Template undefined (type string has no field or method Template)
internal/connectors/engine/bundle_test.go:514:74: countEntry.Default undefined (type string has no field or method Default)
internal/connectors/engine/read_test.go:117:21: undefined: QueryParam
...
```

Compile-error-shape RED (new type `QueryParam` + `StreamSpec.Query` retype don't exist yet) — valid
TDD evidence for introducing a new dialect type, per this repo's established pattern (see r3-ledger.md
"undefined-symbol compile errors" for the same RED shape on a prior increment).

### GREEN

Implementation:
- `internal/connectors/engine/bundle.go`: new `QueryParam{Template, OmitWhenAbsent, Default}` +
  custom `UnmarshalJSON` (bare string -> `{Template: s}`; object -> full struct); `StreamSpec.Query`
  retyped `map[string]QueryParam`.
- `internal/connectors/engine/read.go`: `buildInitialQuery` resolves `param.Template` via
  `Interpolate`; on an unresolved config/secrets key specifically (`isUnresolvedConfigOrSecret`, new
  helper mirroring `isUnresolvedRecordPath`'s errors.As pattern), `OmitWhenAbsent` skips the param,
  else a non-empty `Default` is sent verbatim, else the error still propagates exactly as before. Any
  OTHER interpolation failure (CRLF, unknown filter/namespace) is never tolerated regardless of
  OmitWhenAbsent/Default.
- `cmd/connectorgen/validate.go` `checkInterpolations`: `check("streams.json", v.Template)` (was
  `v` when Query values were bare strings) — mechanical fixup, no rule/behavior change.

**Unavoidable compile-breakage touch (justified, minimal, mechanical)**:
`internal/connectors/conformance/static.go` line ~188 iterates `s.Query` and called
`check(v)`; `v` is now `engine.QueryParam`, not `string`, so `go build ./...` failed with a
type-mismatch compile error. `conformance/**` is outside this agent's editable file set, but this is
the exact "meta-schema/dialect-type change requires a corpus/harness touch — justify" carve-out named
in the dispatch brief: a one-line `check(v)` -> `check(v.Template)` fix, no behavior/semantic change
to conformance's `interpolations_resolve` check (it still validates the exact same template string,
just reached through the new struct field instead of directly). Without this fix `go build ./...`
cannot pass at all, which would block every downstream self-verify command. No other line in
conformance/** was touched.

```
$ go test ./internal/connectors/engine/... -run 'TestReadOptionalQuery|TestReadQueryStringEntry|TestBundleLoadParsesOptionalQueryDialect|TestReadStaticQuery' -v
--- PASS: TestBundleLoadParsesOptionalQueryDialect
--- PASS: TestReadStaticQuery
--- PASS: TestReadOptionalQueryOmittedWhenConfigAbsent
--- PASS: TestReadOptionalQuerySentWhenConfigPresent
--- PASS: TestReadOptionalQueryDefaultAppliedWhenConfigAbsent
--- PASS: TestReadOptionalQueryDefaultOverriddenByConfig
--- PASS: TestReadQueryStringEntryStillHardErrorsOnAbsentConfig
PASS
```

`go build ./...` clean. `go test ./internal/connectors/engine/... ./cmd/connectorgen/...` all pass.
`go test ./internal/connectors/...` shows the SAME (and only the same) chargebee/gmail breakage
already reported under items 1-2 above — item 3 introduces no new pilot-suite breakage.

---

## Item 4 — `last_path_segment` interpolation filter (calendly id)

Per REVIEW-B.md finding 1 / cross-cutting adjudication 1: unblocks calendly's dropped derived `id`
field (`id = idFromURI(uri)`, legacy's trailing-URI-segment convention) and every HAL/URI-keyed API.

Tests added to `internal/connectors/engine/interpolate_test.go`: `TestApplyFilterLastPathSegment`,
`TestApplyFilterLastPathSegmentTrailingSlashIgnored`, `TestApplyFilterLastPathSegmentNoSlashReturnsWholeValue`,
`TestApplyFilterLastPathSegmentKnownToResolveCheck`.

### RED

```
$ go test ./internal/connectors/engine/... -run 'TestApplyFilterLastPathSegment' -v
--- FAIL: TestApplyFilterLastPathSegment (interpolate: unknown filter "last_path_segment")
--- FAIL: TestApplyFilterLastPathSegmentTrailingSlashIgnored (same)
--- FAIL: TestApplyFilterLastPathSegmentNoSlashReturnsWholeValue (same)
--- FAIL: TestApplyFilterLastPathSegmentKnownToResolveCheck (resolve check: unknown filter "last_path_segment")
```

### GREEN

Implementation (`internal/connectors/engine/interpolate.go`):
- New `lastPathSegment(val string) string`: trims a trailing "/", then returns the text after the
  last remaining "/" (or the whole (trimmed) value if there is no "/" at all, or "" for an empty
  value) — never errors.
- Wired into `applyFilterValue`'s switch as `case filter == "last_path_segment"`.
- Added to `knownFilterNames` so `ResolveCheck`/`connectorgen validate` accept it statically (F9).

```
$ go test ./internal/connectors/engine/... -run 'TestApplyFilterLastPathSegment' -v
--- PASS: TestApplyFilterLastPathSegment
--- PASS: TestApplyFilterLastPathSegmentTrailingSlashIgnored
--- PASS: TestApplyFilterLastPathSegmentNoSlashReturnsWholeValue
--- PASS: TestApplyFilterLastPathSegmentKnownToResolveCheck
PASS
```

`go build ./...` clean. `go test ./internal/connectors/engine/... ./cmd/connectorgen/...` all pass.
No pilot-suite impact (calendly itself is Step-2 scope; not touched here).

---

## Item 5 — token_path cursor paginator: stop_path support + loop guard (zendesk has_more)

Per REVIEW-B.md zendesk-support finding 2: `connsdk.CursorPaginator` (the token_path cursor variant)
stops ONLY on an absent/empty `after_cursor`, ignoring `meta.has_more`; Zendesk's docs instruct
clients to use `has_more` and warn the cursor may be populated even when `has_more` is false. Also
no loop-detection guard (unlike `nextURL`/`linkHeaderPaginator`). `connsdk/paginate.go` is out of
this agent's editable scope (not in the FILES YOU MAY TOUCH list) — implemented as a NEW
engine-local paginator (`tokenPathCursor`, alongside the existing engine-local `lastRecordCursor`/
`nextURL`/`linkHeaderPaginator` pattern in `paginate.go`) that `newCursorPaginator`'s `hasToken`
branch now returns instead of a bare `connsdk.CursorPaginator`.

Tests added to `internal/connectors/engine/paginate_test.go`:
`TestNewPaginatorCursorTokenPathStopPathHonoredEvenWithNonEmptyCursor`,
`TestNewPaginatorCursorTokenPathNoStopPathKeepsPriorBehavior` (lock-in, no regression),
`TestNewPaginatorCursorTokenPathLoopGuardSameTokenTwice`.

### RED

```
$ go test ./internal/connectors/engine/... -run 'TestNewPaginatorCursorTokenPath' -v
--- PASS: TestNewPaginatorCursorTokenPathExhausts
--- FAIL: TestNewPaginatorCursorTokenPathStopPathHonoredEvenWithNonEmptyCursor
    paginate_test.go:378: unexpected cursor (must not be requested): page3
    (followed the non-empty after_cursor into a phantom page 3 — exactly the bug REVIEW-B finding 2
    describes: has_more:false is ignored because stop_path isn't wired for the token_path variant)
--- PASS: TestNewPaginatorCursorTokenPathNoStopPathKeepsPriorBehavior
--- FAIL: TestNewPaginatorCursorTokenPathLoopGuardSameTokenTwice
    paginate_test.go:465: token_path cursor paginator *connsdk.CursorPaginator does not implement Err() error
FAIL
```

### GREEN

Implementation (`internal/connectors/engine/paginate.go`):
- New `tokenPathCursor{cursorParam, tokenPath, stopPath, seen map[string]bool, lastErr error}`,
  engine-local (connsdk/paginate.go is out of scope) — mirrors `lastRecordCursor`'s `stopPath`
  falsy-value-stops semantics (read via `connsdk.StringAt`, any value other than the literal
  `"true"` stops) AND `nextURL`'s/`linkHeaderPaginator`'s `seen`-map loop guard + sticky `Err()`.
  A spec with no `stop_path` set preserves the exact prior stop-on-empty-token behavior.
- `newCursorPaginator`'s `hasToken` branch now returns `&tokenPathCursor{...}` instead of a bare
  `&connsdk.CursorPaginator{...}`.

```
$ go test ./internal/connectors/engine/... -run 'TestNewPaginatorCursorTokenPath' -v
--- PASS: TestNewPaginatorCursorTokenPathExhausts
--- PASS: TestNewPaginatorCursorTokenPathStopPathHonoredEvenWithNonEmptyCursor
--- PASS: TestNewPaginatorCursorTokenPathNoStopPathKeepsPriorBehavior
--- PASS: TestNewPaginatorCursorTokenPathLoopGuardSameTokenTwice
PASS
```

`go build ./...` clean. `go test ./internal/connectors/engine/... ./cmd/connectorgen/...` all pass.
`go test ./internal/connectors/...` shows the SAME (only) chargebee/gmail breakage already reported
— item 5 introduces no new pilot-suite breakage (zendesk-support's own `streams.json` does not yet
declare `stop_path`; wiring it there is Step 2 scope).

---

## Item 6 — C3 decision: engine materializes spec.json `default` values at runtimeConfig build

Per REVIEW-A.md C3 flag: every batch-A bundle sets `base.url: {{ config.base_url }}`; the engine
never materializes spec.json `default` values into RuntimeConfig, so every migrated connector
hard-errors on a config shape legacy accepted (github api.github.com, gmail base+token URLs, monday
api.monday.com/v2, chargebee/sentry site/hostname-derived, etc). Adjudicated fix (a): engine/app
increment materializing spec defaults into config, single mechanism for all legacy base-URL
defaults. Validate rule: default must type-check against the property's declared JSON type.

Tests added to `internal/connectors/engine/read_test.go`:
`TestReadBaseURLDefaultMaterializedWhenConfigAbsent`,
`TestReadConfigDefaultDoesNotOverrideExplicitValue` (lock-in: explicit value always wins),
`TestCheckBaseURLDefaultMaterializedWhenConfigAbsent` (Check path, same mechanism).

### RED

```
$ go test ./internal/connectors/engine/... -run 'TestReadBaseURLDefaultMaterialized|TestReadConfigDefaultDoesNotOverrideExplicitValue|TestCheckBaseURLDefaultMaterialized' -v
--- FAIL: TestReadBaseURLDefaultMaterializedWhenConfigAbsent
    read_test.go:256: Read: engine: resolve base url: interpolate: unresolved key "base_url" in config
--- PASS: TestReadConfigDefaultDoesNotOverrideExplicitValue   (trivially true before the fix exists: nothing materializes yet, so there's nothing to override)
--- FAIL: TestCheckBaseURLDefaultMaterializedWhenConfigAbsent
    read_test.go:319: Check: engine: resolve base url: interpolate: unresolved key "base_url" in config
FAIL
```

### GREEN

Implementation:
- `internal/connectors/engine/schema.go`: `schemaNode` gained `defaultVal any`/`hasDefault bool`,
  populated in `compileNode` from the (previously accepted-but-discarded) `"default"` annotation
  keyword (decoded with `json.Decoder.UseNumber()` for integer fidelity). New accessors:
  `Schema.Defaults() map[string]string` (stringifies each declared default via the same `stringify()`
  helper `interpolate.go` uses for every other config-shaped value) and
  `Schema.DefaultTypeMismatches() []string` (reuses `typeMatches` — the same instance-validation
  type-check `Validate()` uses — against the default value itself; sorted).
- `internal/connectors/engine/read.go`: new `materializeConfigDefaults(b Bundle, cfg
  connectors.RuntimeConfig) connectors.RuntimeConfig` — merges `b.Spec.Defaults()` under (never
  over) `cfg.Config`'s existing keys; a `nil` `b.Spec` or empty `Defaults()` is a no-op (returns
  `cfg` unchanged, no allocation). Wired into BOTH `ReadWithSleeper` (mutates `req.Config` before
  `newRuntime`/`readDeclarative`, so every subsequent template resolution — base URL, headers, query,
  computed_fields' `config.*` references — sees the materialized value) and `Check` (mutates `cfg`
  before `newRuntime`) — the single mechanism C3 asked for, covering both entry points.

```
$ go test ./internal/connectors/engine/... -run 'TestReadBaseURLDefaultMaterialized|TestReadConfigDefaultDoesNotOverrideExplicitValue|TestCheckBaseURLDefaultMaterialized|TestSchemaDefaults|TestSchemaDefaultTypeMismatches' -v
--- PASS: TestReadBaseURLDefaultMaterializedWhenConfigAbsent
--- PASS: TestReadConfigDefaultDoesNotOverrideExplicitValue
--- PASS: TestCheckBaseURLDefaultMaterializedWhenConfigAbsent
--- PASS: TestSchemaDefaults
--- PASS: TestSchemaDefaultTypeMismatches
PASS
```

### RED (validate rule: default must type-check)

Seeded a new `cmd/connectorgen/testdata/invalid/default-type-mismatch/` bundle (spec.json's
`max_pages` declares `"type":"integer"` with `"default":"not-a-number"`; `base_url`'s string default
is well-typed, used as the negative-case control). Tests added to `cmd/connectorgen/main_test.go`:
`TestValidate_DefaultTypeMismatchIsHardFinding`, `TestValidate_WellTypedDefaultDoesNotTriggerMismatchRule`,
plus a new row in `TestValidate_RejectsSeededInvalidBundles`'s table.

```
$ go vet ./cmd/connectorgen/...
# polymetrics.ai/cmd/connectorgen [polymetrics.ai/cmd/connectorgen.test]
vet: cmd/connectorgen/main_test.go:74:29: undefined: ruleDefaultTypeMismatch
```

### GREEN (validate rule)

Implementation (`cmd/connectorgen/validate.go`): new `ruleDefaultTypeMismatch = "default_type_mismatch"`
+ `checkDefaultTypeMismatch(b engine.Bundle) []Finding` (hard Finding, not a warning — a bundle
author's spec.json default/type mismatch is always a fixable authoring defect, unlike N2's
plausibility-heuristic warning below it), wired into `validateBundleDir`'s Findings chain. Reuses
`b.Spec.DefaultTypeMismatches()` (item 6's engine accessor) directly.

```
$ go test ./cmd/connectorgen/... -run 'TestValidate_DefaultTypeMismatch|TestValidate_WellTypedDefault|TestValidate_RejectsSeededInvalidBundles' -v
--- PASS: TestValidate_RejectsSeededInvalidBundles (incl. new default-type-mismatch subtest)
--- PASS: TestValidate_DefaultTypeMismatchIsHardFinding
--- PASS: TestValidate_WellTypedDefaultDoesNotTriggerMismatchRule
PASS
```

`go build ./...` clean. `go test ./cmd/connectorgen/... ./internal/connectors/engine/...` all pass.
`go run ./cmd/connectorgen validate internal/connectors/defs` → **13 connector(s) checked, 0
findings** (every existing pilot spec.json default already type-checks cleanly — no retroactive
findings introduced). `go test ./internal/connectors/...` shows the SAME (only) chargebee/gmail
breakage already reported under items 1-2 — item 6's Read/Check default-materialization is
additive-only (only fills a key that was genuinely absent) and introduces no new pilot-suite
regressions.

---

## Item 7 — connectorgen/meta-schema/conventions.md updates (docs-only; no RED/GREEN test cycle)

Docs-only item; no behavior code, so no TDD RED/GREEN applies — verified instead by re-running the
build/test/validate gate after every doc edit (below) and by spot-checking the meta-schemas
against the new dialect shapes with a real bundle-load test (`TestBundleLoadParsesOptionalQueryDialect`,
already GREEN under item 3).

**Meta-schema (`internal/connectors/engine/schema/{streams,spec}.schema.json`)**: NO changes
required. `streams.schema.json`'s `query` property is already `{"type": "object"}` with no
per-entry-value shape constraint, so both the plain-string and new object-form `QueryParam` JSON
shapes validate against it unchanged (confirmed via `TestBundleLoadParsesOptionalQueryDialect`
round-tripping through the full `Load` -> meta-schema-validate path). Same for `pagination`
(`{"type": "object"}`, no `stop_path` field constraint needed). `spec.schema.json`'s `properties` is
likewise unconstrained at the per-property level, so a `"default"` annotation of any shape already
passes. Verified no meta-schema edit was silently required by re-running
`go run ./cmd/connectorgen validate internal/connectors/defs` after every other item landed (13
connectors, 0 findings throughout).

**`docs/migration/conventions.md` updates** (all verified present in the file after editing):
1. §1 Tier-2 line-cap wording (REVIEW-A.md §C1): replaced the self-contradictory "hard-capped at
   ~300" with "~300 soft target; >300 requires a self-reported trace-ledger justification; 400 is a
   hard ceiling; >400 OR a 3rd hook interface (regardless of line count) escalates to Tier 3."
2. §3 filter table: added `last_path_segment` (item 4).
3. §3 `computed_fields` section: added the "Typed extraction — bare `{{ record.<path> }}`" paragraph
   (item 1/A1) and the "`config.*` in `computed_fields` — Config only, Secrets is EXCLUDED by
   design" paragraph (item 2/A3/G0), including the explicit threat-model line (a computed field must
   never copy a secret into record data).
4. §3 `stream.Query` section: replaced "no such omission tolerance at all" with the full
   optional-query dialect description (item 3/REVIEW-B adjudication 2), worked JSON example,
   explicit "not a blanket absent-key-falsy change" framing.
5. §3 pagination table: `cursor (token_path)` row now documents optional `stop_path` + the loop
   guard (item 5); added the shared `stop_path` falsy-value rule paragraph (`connsdk.StringAt`, any
   value other than literal `"true"` is falsy) generalizing Zendesk-shaped boolean stop signals
   beyond just Stripe's shape.
6. §3 auth section: new "Dual-auth ordering is load-bearing — the golden pattern" paragraph lifting
   zendesk-support's ledger item 3 (dual-auth ordering worked example, both-secrets-present parity
   test requirement) per REVIEW-B.md's fan-out-readiness note.
7. §3 new "`spec.json` `"default"` values ARE now materialized" paragraph (item 6/C3): mechanism
   description, derived-default (sentry hostname/chargebee site) guidance (require base_url + drop
   derivation, documented + ledgered, OR a future computed_fields-style base-URL template — never ad
   hoc Go), and the `default_type_mismatch` validate rule cross-reference.
8. §4 Fixture rules: new bullet formalizing the `next_url` single-page-conformance-fixture +
   live-parity-test exception (bitly REVIEW-B.md finding 3 / fan-out-readiness item 4; calendly's
   identical accepted shape) — explicit "why a static fixture cannot embed the correct URL" reasoning
   and the "prove it live instead" requirement.
9. §5 ledger closing paragraph: added a paragraph naming all six item closures and pointing Step 2 at
   which pilot workarounds are now supersedable (chargebee/gmail/github's A1 stringify-widening;
   calendly's dropped `id`; zendesk-support's invented incremental filter + unguarded `has_more`;
   sentry/chargebee's dead hostname/site config).

### Final self-verify (whole gap-loop Step 1, all 7 items)

```
$ go build ./...
(clean)

$ go test ./internal/connectors/engine ./cmd/connectorgen
ok  	polymetrics.ai/internal/connectors/engine
ok  	polymetrics.ai/cmd/connectorgen

$ make lint
(see below)

$ go run ./cmd/connectorgen validate internal/connectors/defs
connectorgen validate: 13 connector(s) checked, 0 findings

$ make lint
golangci-lint run ./internal/connectors/engine/... ./internal/connectors/defs/... ./internal/connectors/hooks/... ./internal/connectors/native/... ./internal/connectors/conformance/... ./internal/connectors/certify/... ./cmd/connectorgen/... ./cmd/inventorygen/...
0 issues.
```

### Full pilot-suite breakage snapshot (handoff data for Step 2, per dispatch instructions)

```
$ go test ./internal/connectors/... 2>&1 | grep -v '^ok'
--- FAIL: TestConformance/chargebee
    records_match_schema: stream "customers": record failed schema validation: /created_at: value does not match type [string null]
FAIL	polymetrics.ai/internal/connectors/conformance
--- FAIL: TestParityChargebee_ComputedFieldsStringifyNumericAndBooleanFields
    engine customers[0].created_at = "1700000000" (json.Number), want string
FAIL	polymetrics.ai/internal/connectors/paritytest/chargebee
--- FAIL: TestParityGmail_ComputedFieldsStringifyLabelCountFields
    engine labels[0].messages_total = "10" (json.Number), want string
FAIL	polymetrics.ai/internal/connectors/paritytest/gmail
```

Exactly and only the item-1/A1-predicted breakage (chargebee's ~30 stringify-widened fields, gmail's
4) — no other pilot (xkcd, vitally, bitly, calendly, sentry, monday, github, zendesk-support, stripe,
searxng, postgres) regressed. github's own nested-id bare computed fields
(`user_id`/`author_id`/`committer_id`/`workflow_run_id`) ALSO now emit native numbers under item 1,
but its parity test's `isStringifiedNestedID` helper already compares `fmt.Sprint(engVal) !=
fmt.Sprint(legacyVal)` (string-form-only comparison), so it tolerates the type upgrade transparently
— github's G0b ledger entry is now stale-but-harmless (doc cleanup only, not a test break; flagged
for Step 2/P-12, not fixed here since it is a defs/docs.md edit).

**Step-2 dispatcher action items** (chargebee/gmail bundle+parity-test+docs.md edits — explicitly
Step-2 scope per this task's dispatch, NOT edited by this agent):
1. `internal/connectors/defs/chargebee/schemas/customers.json` (+ every other stream's numeric/
   boolean computed_fields property, ~30 total per A1): retighten `["string","null"]`-widened types
   back to the real wire type (`["integer","null"]`/`["boolean","null"]` etc.) now that bare
   `{{ record.<path> }}` computed_fields preserve native types.
2. `internal/connectors/paritytest/chargebee/parity_test.go`:
   `TestParityChargebee_ComputedFieldsStringifyNumericAndBooleanFields` should flip from asserting
   the stringified form to asserting native-type equality (or be deleted if RAW `reflect.DeepEqual`
   now holds record-wide) — conventions.md §5's chargebee ledger row should move to RESOLVED.
3. Same shape for `internal/connectors/defs/gmail/schemas/*.json` (4 stringify-widened fields) +
   `TestParityGmail_ComputedFieldsStringifyLabelCountFields`.
4. `internal/connectors/defs/chargebee/docs.md` + `internal/connectors/defs/gmail/docs.md`: update
   the Known-limits entries describing the stringify deviation to reflect it no longer applies.
5. (Lower priority, non-breaking) `internal/connectors/defs/github/docs.md`
   G0b ledger entry is now stale (describes a stringify behavior that no longer occurs) — cosmetic
   only, github's own test suite is green.

None of the above required any change outside `internal/connectors/engine/**`/`cmd/connectorgen/**`/
`docs/migration/conventions.md` — the breakage is exactly the pilot-bundle re-tightening REVIEW-A.md
A1 already scoped to "once the engine feature lands."



