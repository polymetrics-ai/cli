# Overview

High Level (Go HighLevel / LeadConnector) is a wave2 fan-out declarative-HTTP migration. It reads
HighLevel contacts, opportunities, pipelines, custom fields, and form submissions for one location
through the HighLevel REST API, proxied at `{{ config.base_url }}/upstream/<resource>` (default
`https://api.leadconnectorpro.co`). This bundle is engine-vs-legacy parity-intended against
`internal/connectors/high-level` (the hand-written connector it migrates); the legacy package stays
registered and unchanged until wave6's registry flip.

## Auth setup

Provide a HighLevel API key via the `api_key` secret; it is sent as the `x-api-key` header on every
request (`auth: [{"mode": "api_key_header", "header": "x-api-key", "value": "{{ secrets.api_key }}"}]`),
matching legacy's `connsdk.APIKeyHeader("x-api-key", token, "")` (`high_level.go:290`) exactly — never
logged. `location_id` is a required config value scoped to a single HighLevel location and is sent as
the `locationId` query param on every request. `api_version` defaults to `2021-07-28` (materialized via
`spec.json`'s `default`, matching legacy's `defaultAPIVersion` constant) and is sent as the `Version`
header on every request; it may be overridden for forward compatibility exactly as legacy allows.
`base_url` defaults to `https://api.leadconnectorpro.co` and may be overridden for tests/proxies.

## Streams notes

`pipelines` and `custom_fields` are single-request, non-paginated endpoints (`pagination: none`) —
records live at the `pipelines`/`customFields` top-level keys respectively, matching legacy's
`styleNone` endpoints exactly (one GET, no follow-up request).

`contacts`, `opportunities`, and `form_submissions` are cursor-paginated: legacy's `harvest` follows
the response's absolute `meta.nextPageUrl` (falling back to a top-level `nextPageUrl`) until it is
absent, or a page returns zero records (`high_level.go:182-190`). This bundle declares
`pagination: {"type": "next_url", "next_url_path": "meta.nextPageUrl"}` for all three — the engine's
`next_url` paginator follows the identical absolute-URL convention, stops on an absent/empty value the
same way (`connsdk.StringAt` stringifies a JSON `null` to `""`), and loop-guards against the same URL
being requested twice. `limit=100` (legacy's `defaultPageSize`) is sent as a static per-stream `query`
literal on the FIRST request of each of these three streams, matching stripe's `limit=100`
static-query precedent (`docs/migration/conventions.md` worked example). None of the 5 streams declare
an `incremental` block: legacy's `harvest` never applies a server-side or client-side cursor filter —
`CursorFields` on `contacts`/`opportunities`/`form_submissions` are declared on the legacy `Stream`
catalog entries purely for downstream state-cursor bookkeeping, never used to filter a request. This
bundle's schemas mirror that: `x-cursor-field` is declared on those 3 schemas (so
`incremental_append`-family sync modes remain selectable downstream, per design §B.6's
schema-shape-driven sync-mode derivation) without a corresponding `streams.json` `incremental` block —
the identical "declared for bookkeeping, not enforced" shape as legacy.

Every stream's records selector matches legacy's `recordsPath` field-for-field: `contacts` ->
`contacts`, `opportunities` -> `opportunities`, `pipelines` -> `pipelines`, `custom_fields` ->
`customFields`, `form_submissions` -> `submissions`.

## Write actions & risks

None. HighLevel is exposed read-only by legacy (`Capabilities{..., Write: false}`,
`Write` returns `connectors.ErrUnsupportedOperation`); `capabilities.write` is `false` and this bundle
ships no `writes.json`.

## Known limits

- **`query` is re-sent on every followed page, unlike legacy.** The engine's `readDeclarative` loop
  merges `stream.Query` (`locationId`, `limit`) into EVERY page request via `mergeQuery`, including
  when following an absolute `next_url` (`engine/read.go`), whereas legacy explicitly resets to an
  empty `url.Values{}` once it follows the absolute next URL (`high_level.go:188-189`), sending
  `locationId`/`limit` only on the first request. This is the identical wire-request-shape divergence
  documented for bitly (`docs/migration/conventions.md` bitly worked example / bitly's own `docs.md`):
  verified benign in DATA terms only, because HighLevel's own `meta.nextPageUrl` already carries the
  identical `locationId` value this bundle's merge re-applies (the Del+Add replace is idempotent for
  that param), and `limit` does not affect which records a given cursor-addressed page returns. If
  HighLevel's `nextPageUrl` ever diverged from the first request's `locationId`, this bundle's request
  would differ from legacy's; today it does not.
- **`contacts`/`opportunities`/`form_submissions` ship single-page conformance fixtures, per the
  sanctioned `next_url` exception** (`docs/migration/conventions.md` §4): a `next_url` stream's
  next-page URL is the fixture replay server's own address, unknown until the harness picks a port at
  runtime, so a static fixture file cannot embed a correct second-page URL. `streams.json` orders
  `pipelines` (a `pagination: none` stream) first so `pagination_terminates` exercises a real
  multi-record, non-paginated read instead. Conventions.md's exception additionally calls for a live
  `paritytest/<name>` test proving real 2-page `next_url` correctness against an `httptest.Server`;
  this wave's task scope is JSON+docs.md only (no Go/paritytest packages), so that live-parity proof
  is deferred to a follow-up wave rather than fabricated here — the two-page cursor-follow behavior
  itself is unchanged from the bitly/calendly-proven `next_url` paginator, only this specific
  connector's own live-parity test is not yet written.
- **`pipelines.stages` is typed `["array","object","null"]`, not a single fixed type.** Legacy's own
  `Field.Type` catalog entry says `"object"`, but legacy's own `readFixture` stamps an empty array
  literal (`high_level.go:258`, `"stages": []any{}`), and a real HighLevel pipeline's `stages` value is
  documented as an array of stage objects — the schema is widened to accept both shapes actually
  observed rather than picking one and silently narrowing the other away.
