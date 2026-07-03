# Overview

Retently is a wave2 fan-out declarative-HTTP migration. It reads Retently customers, survey
responses, surveys, and campaigns through the Retently REST API v2
(`GET https://app.retently.com/api/v2/...`). This bundle targets capability parity with
`internal/connectors/retently` (the hand-written connector it migrates); the legacy package stays
registered and unchanged until wave6's registry flip. Read-only (`capabilities.write` is `false`,
matching legacy's `Write` returning `connectors.ErrUnsupportedOperation`).

## Auth setup

Provide a Retently API v2 key via the `api_key` secret. It is sent as an
`Authorization: api_key=<api_key>` header, matching legacy's
`connsdk.APIKeyHeader("Authorization", key, "api_key=")` exactly (`retently.go:170`) — the
declarative `api_key_header` auth mode with `header: "Authorization"` and `prefix: "api_key="`
reproduces this byte-for-byte. `base_url` defaults to `https://app.retently.com/api/v2` and may be
overridden for tests/proxies.

## Streams notes

All 4 streams (`customers`, `responses`, `surveys`, `campaigns`) share the identical shape: `GET`
against the Retently v2 list endpoint (`/customers`, `/responses`, `/surveys`, `/campaigns`),
records at the response body's top-level `data` array — legacy's `recordsAt` tries a fallback
candidate list (`path, "data", "items", "records", "results", ""`) per stream, but every one of
those streams declares the SAME endpoint-level `recordsPath: "data"` as its primary candidate
(`retently.go:107-111`), so `records.path: "data"` reproduces the actual first-match behavior
exactly; the remaining fallback candidates only matter for a differently-shaped response body,
which the primary `data` envelope never triggers. Pagination is `page_number` (`page`/`limit`,
`page_size: 100`), stopping on a short page exactly as legacy's `connsdk.PageNumberPaginator` does.

Legacy applies four passthrough filters (`updated_after`, `created_after`, `email`,
`campaign_id`) identically to every stream's request (`retently.go:87-92`'s loop iterates a fixed
key list regardless of which stream is being read) — this bundle reproduces that exact blanket
behavior via the identical four `omit_when_absent` query entries declared on EACH of the four
streams' own `query` block (`HTTPBase` has no `query` field in the engine dialect, so this is
per-stream duplicated rather than a single shared declaration), sent only when the corresponding
config value is set, matching legacy's own `strings.TrimSpace(...) != ""` gate before adding each
to the base `url.Values{}`.
`computed_fields` stamps a static `stream` marker on every record (`"customers"`/`"responses"`/
etc.), matching legacy's `mapRecord`'s `out["stream"] = stream`.

`updated_at`/`created_at` are declared as `x-cursor-field` on each schema, matching legacy's own
`CursorFields` Catalog declarations. No `incremental` block is declared: legacy's `Read` never
reads a persisted sync cursor back into `updated_after`/`created_after` (`harvest` reads only
`req.Config.Config[key]`, never `req.State["cursor"]`) — it always resends the exact same raw
config value on every sync, with no forward advancement. Declaring an `incremental` block instead
would introduce new, behavior-changing state-driven filtering legacy never had. Full refresh (and
`_deduped` sync modes, since `x-primary-key` is present) are what this bundle actually supports,
matching legacy exactly.

## Write actions & risks

None. Legacy `retently.go`'s `Write` returns `connectors.ErrUnsupportedOperation`
unconditionally; `capabilities.write` is `false` and this bundle ships no `writes.json`.

## Known limits

- **`page_size`/`max_pages` are not runtime-configurable.** Legacy exposes both as config-driven
  overrides (`page_size` bounded 1-100, default 100; `max_pages` 0/all/unlimited for unbounded).
  The engine's `page_number` paginator reads `PaginationSpec.PageSize`/`MaxPages` as static
  bundle-authored integers, not config templates — there is no mechanism to wire a `spec.json`
  property into either field. This bundle sends `page_size: 100` (legacy's own default) as a
  static value in `streams.json`'s `base.pagination` block; neither `page_size` nor `max_pages` is
  declared in `spec.json` (F6: dead config is worse than absent config). Pagination is otherwise
  unbounded (matches legacy's `max_pages: 0` = unlimited default) other than the short-page stop
  signal. The `customers` stream's `fixtures/streams/customers/page_1.json` carries a full
  100-record page (matching the declared `page_size` exactly) specifically so the conformance
  harness's `pagination_terminates` check can prove real 2-page continuation without altering the
  bundle's production page size, matching repairshopr's identical documented precedent.
- **Legacy's `id` fallback (`uuid`/`email`/`name`) is not modeled.** Legacy's `mapRecord` falls
  back to a record's `uuid`, `email`, or `name` field when `id` is absent. Every Retently resource
  this bundle reads always carries an `id` in its real wire shape (legacy's own `Catalog`/
  `PrimaryKey` declarations assume `id` unconditionally for all 4 streams), so this fallback is
  defensive dead code against the real API — not exercised by any input legacy itself would
  realistically receive. Documented here for completeness, not implemented via a hook.
- The full Retently API surface (batch feedback import, transactional survey queueing, templates,
  aggregate NPS/CSAT/CES score endpoints, campaign reports) is out of scope for this wave; see
  `api_surface.json`'s `excluded: {category: out_of_scope}` entries.
