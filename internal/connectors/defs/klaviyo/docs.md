# Overview

Klaviyo is a wave2 fan-out declarative-HTTP migration. It reads Klaviyo profiles, events, campaigns,
lists, metrics, and segments through Klaviyo's REST (JSON:API) API (default
`https://a.klaviyo.com/api`). This bundle migrates `internal/connectors/klaviyo` (the hand-written
connector) at capability parity; the legacy package stays registered and unchanged until wave6's
registry flip. Klaviyo is read-only here — legacy has no reverse-ETL write set — so
`capabilities.write` is `false` and no `writes.json` is shipped.

## Auth setup

Provide a Klaviyo private API key via the `api_key` secret. It is sent as the `Authorization` header
with the literal `Klaviyo-API-Key ` prefix (`auth: [{"mode": "api_key_header", "header":
"Authorization", "value": "{{ secrets.api_key }}", "prefix": "Klaviyo-API-Key "}]`), matching legacy's
`connsdk.APIKeyHeader("Authorization", secret, "Klaviyo-API-Key ")` (`klaviyo.go:258`) exactly — never
logged. `revision` (the Klaviyo API date-version) defaults to `2024-10-15` (materialized via
`spec.json`'s `default`, matching legacy's `klaviyoDefaultRevision` constant) and is sent as the
`revision` header on every request; it may be overridden. `base_url` defaults to
`https://a.klaviyo.com/api` and may be overridden for tests/proxies.

## Streams notes

All 6 streams share the identical JSON:API shape: `GET /<resource>?page[size]=100`, records at the
top-level `data` array, next page followed via the response's absolute `links.next` URL
(`pagination: {"type": "next_url", "next_url_path": "links.next"}`) — matching legacy's `harvest`
(`klaviyo.go:152-192`) exactly: follow `links.next` verbatim until it is empty, no extra query merged
onto the absolute follow-up URL (the engine's `next_url` paginator has the identical verbatim-follow,
stop-on-empty, loop-guard-against-repeat behavior).

Every JSON:API object exposes a top-level string `id`/`type` (schema-projected automatically) plus a
nested `attributes` object; `computed_fields` flattens the curated attribute subset legacy's own
mappers promote to the top level (e.g. `"email": "{{ record.attributes.email }}"`), matching each of
`klaviyoProfileRecord`/`klaviyoEventRecord`/`klaviyoCampaignRecord`/`klaviyoListRecord`/
`klaviyoMetricRecord`/`klaviyoSegmentRecord` field-for-field. Every `computed_fields` entry here is a
single bare `{{ record.attributes.<field> }}` reference, so the engine's typed extraction applies:
native JSON types (boolean `archived`/`is_active`/`is_processing`, integer `timestamp`) survive without
stringification, matching legacy's raw `map[string]any` passthrough — schemas declare the real wire
type, not a widened string union. `metrics.integration_name` reaches two levels deep
(`{{ record.attributes.integration.name }}`, matching legacy's `integrationName(attr)` helper) and is
silently skipped (absent, not an error) for a metric with no `integration` object, matching legacy's
nil-safe type-assertion fallback.

`page[size]=100` (legacy's `klaviyoDefaultPageSize`/`klaviyoMaxPageSize`, both 100) is sent as a static
per-stream `query` literal, matching stripe's `limit=100` static-query precedent
(`docs/migration/conventions.md` worked example) — see "Known limits" for why this bundle no longer
declares `page_size`/`max_pages` as runtime-configurable, unlike legacy.

None of the 6 streams declare an `incremental` block: legacy's `harvest` never applies a server-side or
client-side cursor filter (there is no `updated[gte]`-style query param sent, ever) — `CursorFields`
(`updated`/`datetime`/`updated_at`) are declared on the legacy `Stream` catalog purely for downstream
state-cursor bookkeeping. This bundle's schemas mirror that exactly: `x-cursor-field` is declared on
every schema (so `incremental_append`-family sync modes stay selectable downstream, per design §B.6),
with no corresponding `streams.json` `incremental` block — the identical "declared for bookkeeping, not
enforced" shape as legacy and as high-level's own migrated bundle.

## Write actions & risks

None. Klaviyo is `capabilities.write: false`; no `writes.json` is shipped, matching legacy's `Write`
always returning `connectors.ErrUnsupportedOperation`.

## Known limits

- **`page_size`/`max_pages` are no longer runtime-configurable, unlike legacy.** Legacy accepts a
  `page_size` config (1-100, default 100) and a `max_pages` config (0/all/unlimited or a positive
  integer cap) and applies both at read time. The engine's `PaginationSpec` fields are plain
  (non-templated) JSON values baked into `streams.json` at bundle-author time — there is no
  config-driven override mechanism for a `next_url` pagination block's implicit per-page size or a
  request-count cap (the same "declared config with nothing to wire it to" gap documented in
  searxng's `docs.md` and conventions.md's read-only/no-auth worked example). Declaring
  `page_size`/`max_pages` as dead `spec.json` config would violate F6 (REVIEW.md) — a property no
  template in the bundle consumes should not be declared. `page[size]=100` is baked in as a static
  per-stream query literal (matching legacy's default exactly); `max_pages` is left unbounded
  (`PaginationSpec.MaxPages` unset — matching legacy's own default-unlimited behavior), so this is a
  narrowing of an operator-facing override knob, never a change to default emitted-record behavior for
  any config a caller did not explicitly override.
- **`contacts`/all 6 streams ship single-page conformance fixtures, per the sanctioned `next_url`
  exception** (`docs/migration/conventions.md` §4): a `next_url` stream's next-page URL is the fixture
  replay server's own address, unknown until the harness picks a port at runtime, so a static fixture
  file cannot embed a correct second-page URL. Every fixture here sets `links.next: null` so
  `pagination_terminates` and `read_fixture_nonempty` pass on a single, real page. Conventions.md's
  exception additionally calls for a live `paritytest/<name>` test proving real 2-page `next_url`
  correctness against an `httptest.Server`; this wave's task scope is JSON+docs.md only (no
  Go/paritytest packages), so that live-parity proof is deferred to a follow-up wave rather than
  fabricated here — the two-page cursor-follow behavior itself is unchanged from the bitly/calendly/
  high-level-proven `next_url` paginator, only this specific connector's own live-parity test is not
  yet written.
