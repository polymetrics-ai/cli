# Gap-loop cycle-1, Step 2 — calendly repair ledger

Scope: `.planning/phases/wave1-pilot/GAP-LOOP-PLAN.md` Step 2's calendly bullet, per
`REVIEW-B.md`'s calendly findings (2 blockers, 2 majors) and cross-cutting adjudications 1/2.
Files touched: `internal/connectors/defs/calendly/**`, `internal/connectors/paritytest/calendly/**`,
this ledger. No commits made (per dispatch instructions). Repo IS a git repo (env banner said
otherwise); `git diff HEAD` used throughout to separate this agent's edits from sibling gap-loop
repair agents' concurrent in-flight changes (chargebee, github, xkcd, zendesk-support all had
uncommitted changes at session start — confirmed via `git status`, none touched).

Legacy ground truth read in full before editing: `internal/connectors/calendly/calendly.go`,
`internal/connectors/calendly/streams.go`. Step-1 engine additions confirmed already landed before
starting: `last_path_segment` filter (`engine/interpolate.go:283-344`, `docs/migration/
conventions.md` line 193), C3 config-default materialization (`engine/schema.go`'s `Defaults()`/
`DefaultTypeMismatches()`, `engine/read.go`'s `materializeConfigDefaults`).

## Baseline (before this session's edits)

`internal/connectors/defs/calendly/spec.json` at HEAD already declared `page_size`'s
`"default": "100"` (REVIEW-B.md finding 2's fix was already partially landed — confirmed via
`git diff HEAD` showing zero page_size-property change, only max_pages/mode deletion). All other
findings (1, 3, 4) were still unfixed at session start: schemas had `x-primary-key: ["uri"]` with
no `id` property, `parity_test.go` had the `stripDerivedID` workaround REVIEW-B.md's adjudication
1 orders deleted, spec.json still declared dead `max_pages`/`mode`, and `streams.json`'s `users`
stream had no pagination override.

## RED-first evidence (finding 1 — dropped derived `id`)

Per TDD discipline, the test was rewritten to assert the FIXED behavior FIRST (deleting
`stripDerivedID` and asserting raw equality including `id`), confirming it fails against the
UNFIXED bundle, before touching schemas/streams.json:

```
$ go test ./internal/connectors/paritytest/calendly/... -v
=== RUN   TestParityCalendly_StreamRecords/scheduled_events
    parity_test.go:241: stream "scheduled_events" record 0 mismatch:
        engine:  map[... uri:https://api.calendly.com/scheduled_events/E1]  (no "id")
        legacy:  map[... id:E1 uri:https://api.calendly.com/scheduled_events/E1]
--- FAIL: TestParityCalendly_StreamRecords (0.00s)
    --- FAIL: TestParityCalendly_StreamRecords/scheduled_events (0.00s)
    --- FAIL: TestParityCalendly_StreamRecords/event_types (0.00s)
    --- FAIL: TestParityCalendly_StreamRecords/organization_memberships (0.00s)
    --- FAIL: TestParityCalendly_StreamRecords/groups (0.00s)
=== RUN   TestParityCalendly_UsersSingleObject
    parity_test.go:269: users record mismatch: ... (missing "id")
--- FAIL: TestParityCalendly_UsersSingleObject (0.00s)
=== RUN   TestParityCalendly_ManifestSurface
    parity_test.go:612: engine stream "scheduled_events" primary key = [uri], want [id] (legacy)
--- FAIL: TestParityCalendly_ManifestSurface (0.00s)
FAIL
```

Honest RED, captured before any schema/streams.json edit landed.

## GREEN evidence (finding 1)

Fix: every schema (`scheduled_events.json`, `event_types.json`, `organization_memberships.json`,
`groups.json`, `users.json`) gained an `id` property, `x-primary-key: ["id"]` (was `["uri"]`),
`required: ["id", "uri"]` (was `["uri"]`); `streams.json` gained
`"computed_fields": {"id": "{{ record.uri | last_path_segment }}"}` on all 5 streams (added to the
existing `computed_fields` block for `organization_memberships`, new blocks for the other 4).
`stripDerivedID` deleted from `parity_test.go` entirely (not just unused — removed per REVIEW-B.md
adjudication 1's explicit instruction: "the parity strips were removed on resolution"); every call
site (`TestParityCalendly_StreamRecords`, `TestParityCalendly_UsersSingleObject`) now asserts raw
`reflect.DeepEqual` with no field exclusion. `TestParityCalendly_ManifestSurface` extended to
compare primary keys against legacy's `Catalog()` (previously deliberately skipped, per the
old comment, now asserted `["id"] == ["id"]` for every stream — REVIEW-B.md correctly identified
this omission as complicit in the misclassification).

```
$ go test ./internal/connectors/paritytest/calendly/... -v
--- PASS: TestParityCalendly_StreamRecords (0.00s)  (+ all 4 subtests)
--- PASS: TestParityCalendly_UsersSingleObject (0.00s)
--- PASS: TestParityCalendly_ManifestSurface (0.00s)
PASS
```

## RED/GREEN evidence (finding 4 — users stream pagination type)

New test `TestParityCalendly_UsersStreamPaginationExplicitlyNone` written first, confirmed RED
against a scratch copy of `streams.json` with the `users` stream's `pagination` override removed:

```
$ go test ./internal/connectors/paritytest/calendly/... -run TestParityCalendly_UsersStreamPaginationExplicitlyNone -v
    parity_test.go:306: users stream has no stream-level pagination override, want explicit {type: none}
--- FAIL: TestParityCalendly_UsersStreamPaginationExplicitlyNone (0.00s)
```

Fix: `streams.json`'s `users` stream entry gained `"pagination": {"type": "none"}` (a stream-level
`Pagination` entirely replaces `base.pagination`, per `bundle.go`'s `StreamSpec.Pagination` doc
comment — confirmed by reading the struct before relying on this).

```
$ go test ./internal/connectors/paritytest/calendly/... -run TestParityCalendly_UsersStreamPaginationExplicitlyNone -v
--- PASS: TestParityCalendly_UsersStreamPaginationExplicitlyNone (0.00s)
```

## RED/GREEN evidence (finding 3 — dead spec keys `max_pages`/`mode`)

New regression-guard test `TestParityCalendly_SpecHasNoDeadConfigKeys` written first (asserts
`bundle.Spec.Properties()` == the 5 real, wired keys), confirmed RED against a scratch copy of
`spec.json` with `max_pages`/`mode` re-added:

```
$ go test ./internal/connectors/paritytest/calendly/... -run TestParityCalendly_SpecHasNoDeadConfigKeys -v
    parity_test.go:800: spec.json declared properties = [api_key base_url max_pages mode organization_uri page_size start_date], want [api_key base_url organization_uri page_size start_date]
--- FAIL: TestParityCalendly_SpecHasNoDeadConfigKeys (0.00s)
```

Fix: `spec.json`'s `max_pages` and `mode` properties deleted outright (neither is consumed by any
template, engine mechanism, or hook — grepped the whole bundle + paritytest dir first to confirm
zero other references before deleting).

```
$ go test ./internal/connectors/paritytest/calendly/... -run TestParityCalendly_SpecHasNoDeadConfigKeys -v
--- PASS: TestParityCalendly_SpecHasNoDeadConfigKeys (0.00s)
```

## Finding 2 (page_size default) — regression-locking evidence, not RED-first

`spec.json`'s `page_size` `"default": "100"` was ALREADY present at session start (confirmed via
`git diff HEAD` showing no change to that property) — the C3 default-materialization mechanism
this finding calls for was already wired for calendly before this dispatch began (likely landed by
an earlier partial pass alongside Step 1's engine work). Rather than skip verification, two new
tests were added to lock this in as an explicit, asserted behavior rather than an implicit side
effect of an unrelated fixture default:

- `TestParityCalendly_PageSizeDefaultsTo100WhenUnset`: `page_size` deliberately left UNSET in both
  connectors' config (the parity harness's `calendlyRuntimeConfig` helper was also changed to STOP
  hardcoding `"page_size": "100"` into every test's config map, so the suite now genuinely
  exercises the unset/default path instead of masking it, per REVIEW-B.md finding 2's exact
  complaint: "Every parity test masks it by setting page_size: '100'"). Asserts both sides send
  `count=100` on the wire.
- `TestParityCalendly_PageSizeExplicitOverride`: an explicit `page_size: "50"` still overrides the
  spec default on both sides (proves `materializeConfigDefaults` only fills genuinely absent keys,
  never clobbers an explicit value).

Both pass against the current bundle:

```
$ go test ./internal/connectors/paritytest/calendly/... -run 'TestParityCalendly_PageSize' -v
--- PASS: TestParityCalendly_PageSizeDefaultsTo100WhenUnset (0.00s)
--- PASS: TestParityCalendly_PageSizeExplicitOverride (0.00s)
```

Since `calendlyRuntimeConfig`'s hardcoded `page_size: "100"` was removed for ALL callers (not just
the two new tests), every pre-existing parity test in the suite now also exercises the real
default-materialization path implicitly — confirmed by the full-suite GREEN run below.

## docs.md honesty fixes (finding 5)

- Known-limits' `id`-not-reproduced item was FALSE after finding 1's fix (id IS now reproduced) —
  replaced with an accurate item stating `id` IS reproduced via `last_path_segment`, explicitly
  flagging that it corrects (rather than silently drops) the prior stale claim.
- Streams-notes' page_size line previously asserted an unqualified "matching legacy's
  `calendlyPageSize` default of 100" without disclosing the HARD-ERROR-ON-UNSET regression
  REVIEW-B.md finding 2 identified — rewritten to state `page_size` is NOT required, name the C3
  mechanism (`materializeConfigDefaults`) that now makes an unset value resolve to 100 rather than
  erroring, and cite the exact legacy fallback it matches (`calendly.go:363-376`).
  the fix (`materializeConfigDefaults`) that makes it true, and the exact legacy line it matches.
- Added a sentence documenting the `users` stream's explicit `pagination: {type: none}` override
  (finding 4) in the pagination paragraph.
- No `max_pages`/`mode` references existed in docs.md to begin with (verified by grep before and
  after) — no wording fix needed there. The org-scoping Known-limits item and the three-streams
  `x-cursor-field`-without-incremental-block item were re-read and confirmed still accurate; left
  unchanged.

## Full self-verify (final state)

```
$ go build ./...
(clean)

$ go test ./internal/connectors/paritytest/calendly/... -v
17 top-level tests (some with subtests) — ALL PASS, 0 failures.
ok  	polymetrics.ai/internal/connectors/paritytest/calendly	0.37s

$ go test ./internal/connectors/conformance -run 'TestConformance/calendly' -v
--- PASS: TestConformance (0.01s)
    --- PASS: TestConformance/calendly (0.01s)

$ go run ./cmd/connectorgen validate internal/connectors/defs
connectorgen validate: 13 connector(s) checked, 1 finding(s)
  -> the 1 finding is in `github` (auth_type/when unresolved spec key), a DIFFERENT gap-loop Step-2
     repair agent's in-progress file, outside this agent's scope/writable set. Zero calendly
     findings (confirmed via grep).

$ golangci-lint run ./internal/connectors/paritytest/calendly/...
0 issues.

$ make lint   (repo-wide; LINT_PKGS = engine/defs/hooks/native/conformance/certify + cmd/connectorgen + cmd/inventorygen)
0 issues.

$ go test ./internal/connectors/... ./cmd/...
2 unrelated pre-existing failures observed, both OUTSIDE this agent's scope:
  - TestConformance/github (sibling gap-loop agent's in-progress github repair)
  - TestParityGmail_ComputedFieldsStringifyLabelCountFields (unrelated connector, not in Step 2's
    calendly bullet)
calendly-scoped packages (paritytest/calendly, conformance/calendly subtest) both clean.
```

## Files changed

- `internal/connectors/defs/calendly/spec.json` — deleted dead `max_pages`/`mode` properties
  (page_size default already present at HEAD, untouched).
- `internal/connectors/defs/calendly/streams.json` — added `computed_fields.id` (last_path_segment
  filter) to all 5 streams; added `"pagination": {"type": "none"}` to the `users` stream.
- `internal/connectors/defs/calendly/schemas/{scheduled_events,event_types,
  organization_memberships,groups,users}.json` — added `id` property, `x-primary-key: ["id"]`
  (was `["uri"]`), `required` now includes `"id"`.
- `internal/connectors/defs/calendly/docs.md` — corrected the stale `id`-not-reproduced Known
  limits claim; corrected the page_size streams-note to disclose the C3 default mechanism instead
  of an unqualified "matches legacy" claim; documented the `users` pagination-none override.
- `internal/connectors/paritytest/calendly/parity_test.go` — deleted `stripDerivedID` and every
  call site (raw equality now includes `id`); `TestParityCalendly_ManifestSurface` now asserts
  primary-key parity; `calendlyRuntimeConfig` no longer hardcodes `page_size: "100"` (exercises the
  real default path); added `TestParityCalendly_UsersStreamPaginationExplicitlyNone`,
  `TestParityCalendly_PageSizeDefaultsTo100WhenUnset`, `TestParityCalendly_PageSizeExplicitOverride`,
  `TestParityCalendly_SpecHasNoDeadConfigKeys`.

## Blockers

None. All 5 mandated fixes landed; no human gate reached (no dependency additions, no schema
migrations outside this bundle's own JSON schemas, no auth/security changes, no destructive data
actions, no secret access, no quality-gate reductions).
