# Overview

Statuspage is a wave2 fan-out declarative-HTTP migration. It reads Statuspage pages, components,
incidents, and subscribers through the Statuspage API (`GET https://api.statuspage.io/v1/...`).
This bundle targets capability parity with `internal/connectors/statuspage` (the hand-written
connector it migrates); the legacy package stays registered and unchanged until wave6's registry
flip.

## Auth setup

Provide a Statuspage API key via the `api_key` secret; it is sent as the `Authorization` header
with an `OAuth ` prefix (`Authorization: OAuth <api_key>`), matching legacy's
`connsdk.APIKeyHeader("Authorization", key, "OAuth ")` (`statuspage.go:119`). Never logged.
`base_url` defaults to `https://api.statuspage.io/v1` and may be overridden for tests/proxies.

## Streams notes

`pages` is a top-level `GET /pages` list with no page scoping; records are the response body's
top-level JSON array (`records.path: "."`). `components`, `incidents`, and `subscribers` are all
scoped to one Statuspage page via the required `page_id` config value, substituted into each
stream's `/pages/{{ config.page_id }}/...` path template (urlencoded by `InterpolatePath`'s
per-segment default, matching legacy's own `url.PathEscape(pageID)` in `resolveResource`); an
absent `page_id` hard-errors on both sides (legacy: `"statuspage stream requires config page_id for
path %q"`; engine: an unresolved `config.page_id` path-template key). All 4 streams' records are
top-level JSON arrays (`records.path: "."`), matching legacy's `recordsPath: "."` for every stream.

`incidents` declares `incremental.cursor_field: created_at`, matching legacy's own
`CursorFields: []string{"created_at"}` declaration; neither this bundle nor legacy sends a
server-side lower-bound filter or performs client-side filtering for this stream (legacy's `Read`
performs no incremental filtering at all) — this bundle matches that exactly (no `request_param`/
`client_filtered` declared), not introducing new filtering under the guise of a migration. `pages`,
`components`, and `subscribers` have no cursor field, matching legacy (full refresh only).

Pagination is page-number (`pagination.type: page_number`, `page_param: page`, `size_param:
per_page`, `start_page: 1`, `page_size: 100`), identical to legacy's
`connsdk.PageNumberPaginator{PageParam: "page", SizeParam: "per_page", StartPage: 1, PageSize:
pageSize}` with legacy's default `pageSize` of 100.

## Write actions & risks

None. Legacy's `Write` always returns `connectors.ErrUnsupportedOperation`; `capabilities.write` is
`false` and this bundle ships no `writes.json`.

## Known limits

- **`page_size`/`max_pages` are not runtime-configurable.** Legacy exposes both as config-driven
  overrides (`statuspage.go`'s `pageSize`/`maxPages`, bounded 1-100 / a non-negative integer or
  `all`/`unlimited`). The engine's `page_number` paginator has no config-driven page-size or
  max-pages knob (`PaginationSpec.PageSize`/`MaxPages` are static bundle JSON, never templated), so
  this bundle uses legacy's own default (`page_size: 100`) as a fixed bundle value and does not
  declare `page_size`/`max_pages` in `spec.json` at all (a declared-but-unwireable config key is
  worse than an absent one, per `docs/migration/conventions.md` F6). Pagination is unbounded by
  default (reads every page until a short page), matching legacy's own default of `maxPages=0`
  (unbounded) when `max_pages` is unset.
- Full Statuspage API surface (incident creation/mutation, metrics, page users, webhooks) is out of
  scope for this wave; see `api_surface.json`'s `excluded: {category: out_of_scope}` entries.
