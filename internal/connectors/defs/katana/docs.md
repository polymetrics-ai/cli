# Overview

Katana is a read-only declarative-HTTP migration of `internal/connectors/katana` (the legacy
hand-written connector, which stays registered and unchanged until wave6's registry flip). It
reads Katana MRP (Cloud Inventory) products, materials, variants, sales orders, and customers
through the Katana REST API. Katana upstream supports full_refresh only; there is no write
capability and no `writes.json`.

## Auth setup

Provide a Katana API key via the `api_key` secret; it is used only for Bearer auth
(`Authorization: Bearer <api_key>`) and is never logged. `base_url` defaults to
`https://api.katanamrp.com/v1` and can be overridden for tests or proxies (matches legacy's
`katanaDefaultBaseURL` fallback).

## Streams notes

All 5 streams (`products`, `materials`, `variants`, `sales_orders`, `customers`) share the
identical shape: `GET` against the Katana list endpoint (`/products`, `/materials`, `/variants`,
`/sales_orders`, `/customers`), records at the top-level `data` array, primary key `["id"]`.
Pagination is `page_number` (`page_param: page`, `size_param: limit`, `start_page: 1`,
`page_size: 50`) — matches legacy's `PageNumberPaginator` with `katanaDefaultPageSize` (50),
stopping on a short/empty page exactly like `connsdk.Harvest`'s stop condition.

Every Katana object publishes an integer `id` and ISO-8601 `created_at`/`updated_at` timestamps
(matches legacy's `katanaStreams()` catalog, which declares `CursorFields: ["updated_at"]` for
every stream). Each stream's schema declares `x-cursor-field: updated_at` and `streams.json`
declares a bare `incremental.cursor_field: updated_at` with no `request_param` and no
`client_filtered` — this is a deliberate, parity-accurate representation, not an oversight:
legacy's own `Read` never consults `req.State["cursor"]` to filter records at all (server-side or
client-side); it always performs a full re-scan of every page on every sync, exactly like this
bundle now does. Declaring the bare `cursor_field` only affects which sync modes the engine
derives (`incremental_append`/`incremental_append_deduped` become available because a
cursor is published), matching legacy's own catalog-published `CursorFields`, without changing
the actual records read or their order for any sync.

`page_size` (default 50, legacy's `katanaDefaultPageSize`) and `max_pages` (default 0/unlimited)
are **no longer configurable at runtime** — this is a documented, deliberate config-surface
narrowing (same class as `searxng`'s F6 finding, see `docs/migration/conventions.md` §3): the
dialect's `page_number` pagination `page_size`/`max_pages` fields are load-time JSON literals with
no `config.*`-driven override mechanism (`paginate.go`'s `newPaginator` takes a plain `int`, never
a template), so legacy's `page_size`/`max_pages` config keys are genuinely dead in this dialect and
are not declared in `spec.json` at all (a declared-but-unwireable key is worse than an absent one,
per F6). The bundle's fixed `page_size: 50` reproduces legacy's own *default* page size exactly
(legacy's default, absent an override, is also 50); `max_pages` is omitted (unbounded), matching
legacy's default (`0` = unlimited).

## Write actions & risks

None. Katana is read-only upstream (`capabilities.write: false`); there is no `writes.json`.

## Known limits

- **`page_size`/`max_pages` runtime overrides are not modeled** — see Streams notes above. Legacy
  allowed a caller to override these via config; the new engine's `page_number` pagination spec has
  no such mechanism. The bundle's fixed defaults (`page_size: 50`, unbounded `max_pages`) match
  legacy's own defaults exactly, so this only removes an override a caller could previously supply,
  never changes default behavior.
- Full Katana API surface (bill of materials, purchase orders, stocktakes, webhooks) is out of
  scope for this migration; see `api_surface.json`'s `excluded: {category: out_of_scope, reason:
  "Pass B capability expansion"}` entries. Only the 5 legacy-parity read streams are implemented.
- Fixtures (`fixtures/streams/**`) use synthetic-but-real-shaped Katana objects; `products` (the
  `pagination_terminates` conformance stream) ships a full 50-record page 1 and a 1-record page 2
  to exercise the short-page stop condition.
