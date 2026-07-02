# Overview

Leadfeeder is a wave2 fan-out declarative-HTTP migration. It reads Leadfeeder accounts and their
leads, visits, and custom feeds through Leadfeeder's JSON:API (default `https://api.leadfeeder.com`).
This bundle migrates `internal/connectors/leadfeeder` (the hand-written connector) at capability
parity; the legacy package stays registered and unchanged until wave6's registry flip. Leadfeeder is
read-only here — legacy has no reverse-ETL write set — so `capabilities.write` is `false` and no
`writes.json` is shipped.

## Auth setup

Provide a Leadfeeder API token via the `api_token` secret. It is sent as the `Authorization` header
with the literal `Token token=` prefix (`auth: [{"mode": "api_key_header", "header": "Authorization",
"value": "{{ secrets.api_token }}", "prefix": "Token token="}]`), matching legacy's
`connsdk.APIKeyHeader("Authorization", secret, leadfeederAuthPrefix)` (`leadfeeder.go:276`) exactly —
never logged. `account_id` is an optional config value required only by the `leads`, `visits`, and
`custom_feeds` streams (nested under `/accounts/{account_id}/...`); those three streams' `path`
templates `{{ config.account_id }}` and hard-error at read time if it is unset, matching legacy's
`endpointPath` (`leadfeeder.go:283-292`) exactly — `accounts` never references it. `base_url` defaults
to `https://api.leadfeeder.com` and may be overridden for tests/proxies.

## Streams notes

All 4 streams share Leadfeeder's JSON:API envelope: records at the top-level `data` array, next page
followed via the response's absolute `links.next` URL (`pagination: {"type": "next_url",
"next_url_path": "links.next"}`) — matching legacy's `harvest` (`leadfeeder.go:172-214`) exactly:
follow `links.next` verbatim (treating both an empty string and the literal `"null"` as absent, which
the engine's `next_url` paginator's `connsdk.StringAt`-then-empty-check already handles identically for
a JSON `null`) until absent, no extra query merged onto the absolute follow-up URL. Every JSON:API
object exposes a top-level string `id`/`type` (schema-projected automatically) plus a nested
`attributes` object; `computed_fields` flattens the curated attribute subset legacy's own mappers
promote to the top level (e.g. `"name": "{{ record.attributes.name }}"`), matching each of
`leadfeederAccountRecord`/`leadfeederLeadRecord`/`leadfeederVisitRecord`/`leadfeederCustomFeedRecord`
field-for-field. Every `computed_fields` entry is a single bare `{{ record.attributes.<field> }}`
reference, so the engine's typed extraction applies: native JSON types (integer `quality`/`visits`/
`visit_length`/`pageviews`) survive without stringification, matching legacy's raw `map[string]any`
passthrough.

`leads` and `visits` are Leadfeeder's `dateScoped` streams: legacy's `dateWindow` sends `start_date`/
`end_date` (yyyy-mm-dd, normalized from RFC3339 or bare date input) whenever a lower bound resolves
(the incremental state cursor, else the `start_date` config), defaulting `end_date` to today (UTC)
when only `start_date` resolves — Leadfeeder requires both together. This bundle expresses the
config-driven half of that with the optional-query dialect (`query.start_date`/`query.end_date`, both
`omit_when_absent: true`, referencing `{{ config.start_date }}`/`{{ config.end_date }}` directly) —
see "Known limits" for the one piece of legacy behavior this does NOT reproduce (the auto-default of
`end_date` to today, and driving `start_date` from the incremental state cursor rather than only the
static config value).

`page[size]=100` (legacy's `leadfeederDefaultPageSize`/`leadfeederMaxPageSize`, both 100) is sent as a
static per-stream `query` literal, matching stripe's `limit=100` static-query precedent
(`docs/migration/conventions.md` worked example) — see "Known limits" for why this bundle no longer
declares `page_size`/`max_pages` as runtime-configurable.

## Write actions & risks

None. Leadfeeder is `capabilities.write: false`; no `writes.json` is shipped, matching legacy's
`Write` always returning `connectors.ErrUnsupportedOperation`.

## Known limits

- **`start_date`/`end_date` are sourced from `config` only, not from the incremental state cursor, and
  `end_date` does NOT auto-default to today.** Legacy's `dateWindow` (`leadfeeder.go:297-319`) prefers
  the incremental state cursor (`connsdk.Cursor(req.State)`) over the `start_date` config for the lower
  bound, and defaults `end_date` to today (UTC) whenever a lower bound resolves but no `end_date`
  config is set. The engine's `stream.Query` optional-query dialect resolves `config.*`/`secrets.*`/
  `incremental.lower_bound` references, but there is no "compute today's date" template primitive
  (`docs/migration/conventions.md` §3's template-reference list has no `now()`/date-arithmetic
  function), so an auto-default-to-today value cannot be expressed without inventing ad hoc Go. This
  bundle intentionally does NOT declare a `leads`/`visits` `incremental` block (which would let the
  state cursor drive `{{ incremental.lower_bound }}` for `start_date`) specifically because pairing it
  with a static "run to end of time" `end_date` would silently diverge from legacy's paired-window
  semantics — instead, `start_date`/`end_date` are both plain optional `config.*`-sourced query
  params: a caller must set BOTH explicitly for a date-scoped read; the automatic "run from last sync
  cursor to today" repeat-sync behavior legacy provides is not reproduced. This is a scope-narrowing
  simplification (an operator must now do the incremental bookkeeping/end-date defaulting themselves,
  e.g. via an orchestration-layer scheduled config update), not a change to the DATA emitted for any
  request this bundle successfully sends — a caller who does supply matching `start_date`/`end_date`
  values gets byte-identical query parameters to what legacy would have sent for that same window.
- **`page_size`/`max_pages` are no longer runtime-configurable.** Legacy exposes both as config
  overrides (`leadfeederPageSize`/`leadfeederMaxPages`, `leadfeeder.go:365-393`, default 100, max 100).
  The engine's `next_url` paginator has no page-size/request-count config-driven override mechanism
  (the same gap documented in klaviyo's and searxng's `docs.md`); `page[size]=100` is baked in as a
  static per-stream query literal (matching legacy's default exactly), and `max_pages` is left
  unbounded (matching legacy's own default-unlimited behavior).
- **All 4 streams ship single-page conformance fixtures, per the sanctioned `next_url` exception**
  (`docs/migration/conventions.md` §4): a `next_url` stream's next-page URL is the fixture replay
  server's own address, unknown until the harness picks a port at runtime, so a static fixture file
  cannot embed a correct second-page URL. Every fixture here sets `links.next: null` so
  `pagination_terminates` and `read_fixture_nonempty` pass on a single, real page. Conventions.md's
  exception additionally calls for a live `paritytest/<name>` test proving real 2-page `next_url`
  correctness against an `httptest.Server`; this wave's task scope is JSON+docs.md only (no
  Go/paritytest packages), so that live-parity proof is deferred to a follow-up wave rather than
  fabricated here.
