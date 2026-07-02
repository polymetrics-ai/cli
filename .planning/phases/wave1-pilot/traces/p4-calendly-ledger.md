# P-4 calendly — migration trace (wave1-pilot, DW-1)

Connector: calendly. Legacy: `internal/connectors/calendly/calendly.go` + `streams.go` (673 loc,
read entirely before authoring). Bundle: `internal/connectors/defs/calendly/`. Parity suite:
`internal/connectors/paritytest/calendly/parity_test.go`. No hooks (no Tier-2 trigger applies per
conventions.md §6 decision tree — calendly's Bearer auth, `next_url` pagination, and
`min_start_time` incremental filter are all expressible in Tier 1; SPEC wave1-pilot §5.2 also
specifies calendly with no "Extra dirs" column, i.e. plain Tier 1).

## RED-first evidence

`internal/connectors/paritytest/calendly/parity_test.go` was written FIRST, calling
`engine.Load(defs.FS, "calendly")` (initially `engine.LoadAll(defs.FS)`, see note below) before any
bundle file existed. First run output (captured before `internal/connectors/defs/calendly/**`
existed — the empty `internal/connectors/defs/calendly/` directory I had to create for the
`paritytest/calendly` sibling scaffold broke the shared `defs.FS` `//go:embed all:*` directive
outright):

```
$ go test ./internal/connectors/paritytest/calendly/... -v
# polymetrics.ai/internal/connectors/paritytest/calendly
internal/connectors/defs/defs.go:14:12: pattern all:*: cannot embed directory calendly: contains no embeddable files
FAIL	polymetrics.ai/internal/connectors/paritytest/calendly [setup failed]
```

This is the honest RED signal: the calendly bundle did not exist yet. (A second, unrelated
DW-1-fan-out-transient RED was also observed later, mid-authoring, when sibling agents' `github`
bundle directory was momentarily structurally incomplete under the shared `defs.FS` root —
`engine.LoadAll(defs.FS)` failed with `github: missing required file docs.md`. This has nothing to
do with the calendly bundle itself; fixed by switching `loadCalendlyBundle` to
`engine.Load(defs.FS, "calendly")`, which only descends into the `calendly` subtree via `fs.Sub`
and is unaffected by any sibling's in-progress state — both `engine.Load` and `engine.LoadAll` are
named as the production discovery path in SPEC wave1-pilot §6, so this is not a deviation from the
mandated pattern, just the more isolated of the two equally-sanctioned choices for a
parallel-dispatch context.)

## GREEN evidence

After authoring `internal/connectors/defs/calendly/{metadata,spec,streams,api_surface}.json`,
`schemas/{scheduled_events,event_types,organization_memberships,groups,users}.json`, `docs.md`,
`fixtures/{check.json,streams/{scheduled_events,event_types,organization_memberships,groups,users}/page_1.json}`:

```
$ go test ./internal/connectors/paritytest/calendly -v
=== RUN   TestParityCalendly_StreamRecords
=== RUN   TestParityCalendly_StreamRecords/scheduled_events
=== RUN   TestParityCalendly_StreamRecords/event_types
=== RUN   TestParityCalendly_StreamRecords/organization_memberships
=== RUN   TestParityCalendly_StreamRecords/groups
--- PASS: TestParityCalendly_StreamRecords (0.01s)
=== RUN   TestParityCalendly_UsersSingleObject
--- PASS: TestParityCalendly_UsersSingleObject (0.00s)
=== RUN   TestParityCalendly_ScheduledEventsTwoPagePagination
--- PASS: TestParityCalendly_ScheduledEventsTwoPagePagination (0.00s)
=== RUN   TestParityCalendly_NextPageNullTerminates
--- PASS: TestParityCalendly_NextPageNullTerminates (0.00s)
=== RUN   TestParityCalendly_IncrementalMinStartTimeFromStartDate
--- PASS: TestParityCalendly_IncrementalMinStartTimeFromStartDate (0.00s)
=== RUN   TestParityCalendly_IncrementalMinStartTimeFromStateCursor
--- PASS: TestParityCalendly_IncrementalMinStartTimeFromStateCursor (0.00s)
=== RUN   TestParityCalendly_MinStartTimeOnlyAppliesToScheduledEvents
--- PASS: TestParityCalendly_MinStartTimeOnlyAppliesToScheduledEvents (0.00s)
=== RUN   TestParityCalendly_BearerAuthHeaderByteIdentical
--- PASS: TestParityCalendly_BearerAuthHeaderByteIdentical (0.00s)
=== RUN   TestParityCalendly_Non2xxErrorPath
--- PASS: TestParityCalendly_Non2xxErrorPath (0.00s)
=== RUN   TestParityCalendly_ManifestSurface
--- PASS: TestParityCalendly_ManifestSurface (0.00s)
=== RUN   TestParityCalendly_BundleLoadsAndValidates
--- PASS: TestParityCalendly_BundleLoadsAndValidates (0.00s)
PASS
ok  	polymetrics.ai/internal/connectors/paritytest/calendly	0.31s

$ go test ./internal/connectors/conformance -run 'TestConformance/calendly' -v
=== RUN   TestConformance
=== RUN   TestConformance/calendly
--- PASS: TestConformance (0.01s)
    --- PASS: TestConformance/calendly (0.01s)

$ go run ./cmd/connectorgen validate internal/connectors/defs   # calendly-scoped findings
(no calendly findings; full run: "connectorgen validate: 13 connector(s) checked, 0 findings")

$ go build ./internal/connectors/... && go vet ./internal/connectors/...
(clean; calendly-scoped: clean)

$ golangci-lint run ./internal/connectors/paritytest/calendly/...
0 issues.

$ go build ./... && go test ./internal/connectors/... ./cmd/...
(clean at time of this run for calendly; 3 unrelated failures observed transiently in
TestConformance/github, TestConformance/gmail, TestConformance/zendesk-support — different DW-1
sibling connectors, outside this agent's writable/forbidden scope)
```

Per-check conformance detail (via a scratch `engine.Load` + `conformance.RunBundle` probe, not
committed): every applicable check (`spec_schema_valid`, `stream_schemas_valid`, `pk_fields_exist`,
`cursor_fields_exist`, `interpolations_resolve`, `write_schemas_valid`, `surface_complete`,
`docs_present`, `secret_redaction`, `fixtures_present`, `check_fixture`,
`read_fixture_nonempty:<stream>` ×5, `pagination_terminates`, `records_match_schema`,
`cursor_advances`) passed; `delete_semantics` correctly Skipped (calendly has no write actions).

## Design decision: org/user scoping is config-driven, not `/users/me`-auto-discovered

Legacy resolves the `organization` query param DYNAMICALLY on every single read by first calling
`GET /users/me` and reading `resource.current_organization` (`calendly.go`'s
`currentUser`/`scopeQuery`, `Read`'s `user, err := c.currentUser(ctx, r)` at the top of every
non-single-object stream read). The engine's declarative dialect has genuinely no mechanism to
chain one request's response into a later request's query params: `read.go`'s `buildInitialQuery`
only resolves `config.*`/`secrets.*`/`record.*`/`cursor` templates against the READ REQUEST's own
inputs, never a prior response body (verified by reading `interpolate.go`'s `Vars` struct and
`resolveRefValue` end to end — no such namespace exists). Expressing legacy's exact
auto-discovery behavior would require a `StreamHook` (Tier 2, making two HTTP calls per read
inside hand-written Go) — SPEC wave1-pilot §5.2 specifies calendly with no Tier-2 escalation, and
the escape-hatch decision tree (conventions.md §6) only justifies Tier 2 for signature/token-
exchange auth, non-JSON bodies, async polling, sub-resource fan-out, or compound writes — none of
which apply here; this is a query-param-sourcing gap, not one of those triggers.

Decision: the bundle instead declares a REQUIRED `organization_uri` spec.json config value that
the operator configures once (the exact, per-account-invariant URI value legacy would have
discovered via `/users/me` at read time — resolvable by calling that same endpoint once during
setup). Every subsequent request both connectors send is byte-identical given the same
organization URI, which the parity suite proves directly (`calendlyRuntimeConfig` feeds the
engine `organization_uri` while legacy independently discovers the identical value from the SAME
shared httptest server's `/users/me` handler on every subtest).

## Parity-deviation ledger entries (conventions.md §5 candidates)

1. **Organization/user scoping is config-driven (`organization_uri`), not auto-discovered via
   `/users/me` at read time.** See the design-decision section above for the full reasoning.
   **Verdict: ACCEPTABLE.** This never changes any emitted record's DATA for any input legacy
   itself would accept — every organization-scoped list request both sides send is byte-identical
   given the same (real, per-account-invariant) organization URI; it only changes WHEN/HOW that
   URI is supplied (once, at config time, vs. rediscovered on every read). Every parity subtest
   supplies `organization_uri` explicitly and separately proves legacy's own dynamic discovery
   yields the identical value via a shared `/users/me` fixture.

2. **The `id` primary-key convenience field (`idFromURI(uri)`) is not reproduced; schemas declare
   `uri` itself as `x-primary-key` instead.** Legacy derives `id` from `uri`'s trailing path
   segment on every record (`calendly.go`'s `idFromURI`, called from every `calendly*Record`
   mapper in `streams.go`). The engine dialect's closed filter set (`urlencode`, `unix_seconds`,
   `base64`, `join:<sep>` — read in full from `interpolate.go`) has no "take the last path segment
   of a URI" transform, so `computed_fields` cannot reproduce this exact derivation. Filing this as
   a new engine filter was considered and rejected for this pilot: SPEC wave1-pilot §7 says "no
   engine behavior changes in this phase except P-0"; a single-connector-motivated filter addition
   is exactly the kind of per-connector patch the `ENGINE_GAP` recurrence rule (conventions.md §6)
   says to wait for (≥3 occurrences) before promoting to a mini engine increment — and using
   `uri` (Calendly's own stable, always-present, globally-unique resource identifier, and the
   EXACT value `id` is derived FROM) as the primary key instead is a clean, honest, zero-data-loss
   alternative, not a workaround. **Verdict: ACCEPTABLE.** No other field's value changes for any
   input legacy itself would accept; `id` is fully recoverable by any consumer as `uri`'s trailing
   path segment. `stripDerivedID` in the parity suite strips `id` before RAW record comparison
   (proving every OTHER field matches exactly, which a full-test skip would not have proven) and
   `TestParityCalendly_ManifestSurface` compares stream names/cursor fields but deliberately not
   primary keys, with an inline comment explaining why.

3. **`event_types`/`organization_memberships`/`groups` publish an `x-cursor-field` (`updated_at`)
   but receive NO server-side incremental filtering and are NOT `client_filtered`.** This is not a
   bundle-authoring gap — it matches legacy's actual (lack of) filtering behavior for these three
   streams exactly: legacy's `Read` only sets `min_start_time` when
   `endpoint.resource == "scheduled_events"` (`calendly.go:158`), and never applies any other
   filter or client-side drop for `event_types`/`organization_memberships`/`groups`, even though
   `streams.go` publishes `CursorFields: []string{"updated_at"}` for all three (their published
   cursor field is informational/manifest-only in legacy too — a full page is always fetched
   regardless of prior sync state). **Verdict: ACCEPTABLE** (matches legacy byte-for-byte; adding
   `client_filtered: true` here would be NEW behavior legacy never had, and conventions.md's
   meta-rule forbids introducing new deviations under the guise of parity).
   `TestParityCalendly_MinStartTimeOnlyAppliesToScheduledEvents` asserts this directly against
   legacy for all three streams.

## Self-verify summary

| Command | Result |
|---|---|
| `go run ./cmd/connectorgen validate internal/connectors/defs` (calendly-scoped) | 0 findings (full run: 13 connectors, 0 findings) |
| `go build ./internal/connectors/...` | clean |
| `go vet ./internal/connectors/...` | clean |
| `go test ./internal/connectors/conformance -run 'TestConformance/calendly' -v` | PASS |
| `go test ./internal/connectors/paritytest/calendly -v` | PASS (11/11 top-level, incl. subtests) |
| `go build ./...` | clean |
| `golangci-lint run ./internal/connectors/paritytest/calendly/...` | 0 issues |
| `make lint` (repo-wide) | 1 unrelated finding in `internal/connectors/hooks/gmail/hooks.go` (different DW-1 connector, outside this agent's scope) |

## Blockers

None. Status: **migrated**.

## Escape hatches

None. Pure Tier-1 declarative bundle — no hooks package.
