# Overview

Sentry is a wave1-pilot migration (PLAN.md P-5, SPEC.md §5.3). It reads Sentry projects, issues,
error events, and releases through the Sentry REST API (`api/0`), read-only. This bundle is
engine-vs-legacy parity-tested against `internal/connectors/sentry` (the hand-written connector it
migrates); the legacy package stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide a Sentry internal-integration or auth token via the `auth_token` secret; it is used only
for Bearer auth (`Authorization: Bearer <auth_token>`) and is never logged. `organization` and
`project` config values scope the `issues`/`events` (org+project) and `releases` (org) streams;
`base_url` (**required** — e.g. `https://sentry.io` or a self-hosted host) selects the API host.
Legacy instead derives the host from a `hostname` config value (default `sentry.io`,
`sentryBaseURL`: `"https://" + hostname` when `base_url` is unset); this bundle does not reproduce
that derivation — see "Known limits" below for why and the config-surface change this implies for
an operator migrating a legacy-shaped config (a bare `hostname` value must become the fully-formed
`https://<hostname>` `base_url`).

## Streams notes

Four streams: `projects` (org/project-independent), `issues` and `events` (scoped to
`organization`/`project`), `releases` (scoped to `organization`). Every endpoint returns a
top-level JSON array (`records.path: ""`), matching Sentry's real list-endpoint wire shape.

**Pagination — Tier-2 StreamHook, not declarative (SPEC.md §5.3 resolution ladder)**: Sentry's
list pagination is RFC 5988 Link header with a twist legacy hand-rolls precisely
(`sentry/sentry.go:7-9,144-152`): a `rel="next"` Link entry is **ALWAYS present**, even on the
truly-last page, and the real "more pages" signal is that entry's `results="true"/"false"`
attribute. The engine's declarative `link_header` pagination type has no knowledge of `results=`
at all (`engine/paginate.go`'s `linkHeaderPaginator` follows `rel="next"` unconditionally) — ladder
rung 1 (Tier-1 `link_header` + document an "at most one extra trailing request" deviation) was
tested against `conformance`'s `pagination_terminates` check (which requires the fixture replay
server to receive EXACTLY one request per recorded fixture page, 404ing — a hard `connsdk` error,
not a benign no-op — on any request beyond the last page) and **rejected**: a Tier-1 paginator
that always follows Sentry's always-present `rel="next"` link would issue a genuine extra request
that hard-fails against any fixture set sized to the true page count, for every real sync, not
just conformance. This lands on ladder rung 2: `hooks/sentry/hooks.go` implements `StreamHook`,
porting legacy's `harvest`/`nextCursor` Link-header + `results=` handling exactly (same request
shape, same stop condition).

Every stream in this bundle carries an explicit `"conformance": {"skip_dynamic": true, "reason":
"..."}` marker (`internal/connectors/engine/bundle.go`'s `StreamSpec.Conformance`,
`docs/migration/conventions.md` §4/§6): `internal/connectors/conformance/dynamic.go` honors this
marker by Skipping (not attempting) every dynamic fixture-replay check for these streams, since the
StreamHook (always `handled=true`) is what every real `Read()` call actually dispatches through,
and a declarative-only fixture replay cannot exercise it at all. The authoritative substitute this
marker names is `paritytest/sentry`'s dedicated 2-page Link-header +`results=false` test
(`TestParitySentry_IssuesTwoPagePaginationAndResultsFalseStop`, per PLAN.md P-5's requirement) and
`hooks/sentry/hooks_test.go`. `streams.json`'s own `base.pagination` stays declared `{"type":
"none"}` (a single, honest request) since it is never dynamically exercised now — no shaping
needed to satisfy a replay harness that no longer runs against these streams.

No incremental cursor is ever sent as a request parameter anywhere in legacy: `CursorFields` is
published on the catalog/manifest surface only (e.g. `lastSeen`/`dateCreated`), but neither
`sentry.go`'s `Read` nor `harvest` ever forwards a state cursor into a request — there is no
`incrementalLowerBound`-equivalent call anywhere in the legacy package. This bundle matches that
exactly by declaring **no `incremental` block on any stream** (full_refresh only); adding one would
be new, unrequested capability, not a migration.

## Write actions & risks

None. Sentry is read-only in legacy (`Write` returns `connectors.ErrUnsupportedOperation`,
`sentry.go:97-99`); `capabilities.write` is `false` and no `writes.json` is declared.

## Known limits

- Full Sentry API surface (teams, members, alert rules, dashboards, integrations) is out of scope
  for this pilot; see `api_surface.json`'s `excluded: {category: out_of_scope, reason: "Pass B
  capability expansion"}` entries. Only the 4 legacy-parity read streams are implemented.
- **`hostname` config key dropped; `base_url` is now required.** Legacy derives the API host from a
  `hostname` config value (default `sentry.io`) as `https://<hostname>` when `base_url` is unset
  (`sentryBaseURL`, `sentry.go:293-308`). The engine's spec-default materialization (gap-loop
  cycle-1 item 6, C3) only fills in a LITERAL per-key default — it cannot express "derive `base_url`
  from `hostname`", a cross-key template (`chargebee`'s `site` key hit the identical class; see
  `docs/migration/conventions.md`'s derived-default guidance). This bundle therefore drops
  `hostname` entirely and requires `base_url`: an operator migrating a legacy `hostname`-only (or
  default, unset-hostname) config must now supply the fully-formed `https://sentry.io` (or
  `https://<hostname>`) URL as `base_url`. This is a documented config-surface narrowing (every
  legacy-accepted `hostname` value has an operator-reachable `base_url` equivalent; no request/data
  change once configured), not a data-shape regression.
- Pagination is a Tier-2 `StreamHook` (`hooks/sentry/hooks.go`), not declarative `link_header` —
  see "Streams notes" above for the full resolution-ladder reasoning (SPEC.md §5.3). Candidate
  future engine feature: a `link_header` pagination "stop attribute" (a configurable Link-header
  attribute name/value that must hold for `rel="next"` to be followed) would let this connector
  drop the hook; not implemented in this phase per the ENGINE_GAP recurrence rule (conventions.md
  §6) since this is the only pilot connector hitting this shape so far.
- `streams.json`'s declared `base.pagination: {"type": "none"}` is NOT the real production
  pagination behavior — every stream carries a `conformance.skip_dynamic` marker naming
  `paritytest/sentry`/`hooks/sentry/hooks_test.go` as the authoritative substitute; conformance's
  dynamic (fixture replay) checks Skip these streams outright rather than exercising a declarative
  shape that would never match Sentry's real Link-header/`results=` wire behavior. The StreamHook
  (always `handled=true`) is what every real `Read()` call actually dispatches through.
- **RESOLVED — removed the 4 streams' inert static `query: {"per_page": "100"}` entries** (S3 engine
  mini-wave carried minor — SUMMARY.md carried minors: "sentry inert per_page entries"). Every stream
  is `StreamHook`-handled (previous bullet), so `readDeclarative`'s `stream.Query` resolution never
  runs for any of them — `hooks/sentry/hooks.go`'s `ReadStream` builds its OWN `per_page` value from
  `pageSizeFor(req.Config)` (default 100, `config.page_size`-overridable, capped at `maxPageSize`)
  independently of anything in `streams.json`. The removed entries were dead JSON with no effect on
  any real request, ever — this is a docs-honesty/dead-config cleanup, not a behavior change (proven
  unchanged by the full `paritytest/sentry`/`hooks/sentry` suite re-run after removal).
- **OPEN — `Check`'s `per_page=1` is not reproduced.** Legacy's `Check` sends `per_page=1` on its
  bounded `/api/0/projects/` probe (`sentry.go:88`, minimizing response payload for a mere
  connectivity/auth check). This bundle's `base.check` (`{"method":"GET","path":"/api/0/projects/"}`)
  has no query at all: `engine.RequestSpec` (the type backing `base.check`) declares only
  `method`/`path`, no `query` field — `engine.Check`'s `rt.Requester.Do(...)` call passes a literal
  `nil` query unconditionally (`engine/read.go`), so there is no dialect surface to express this
  today without an engine change. Verified BENIGN: Sentry's API returns 200 for the check either way
  regardless of `per_page`, and this bundle's `Check` never reads/validates the response body — the
  only difference is a (harmless, larger) response payload on every `Check` call. Not pursued as an
  engine increment in this pass (a single-connector `RequestSpec.Query` field for one non-functional
  payload-size optimization does not meet the ≥3-occurrence recurrence bar, conventions.md §6);
  documented here as the honest, ledgered residual gap rather than silently left unmentioned.
