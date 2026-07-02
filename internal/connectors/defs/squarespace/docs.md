# Overview

Squarespace is a wave2 fan-out declarative-HTTP migration. It reads Squarespace Commerce orders,
products, inventory items, and profiles through the Squarespace API
(`GET https://api.squarespace.com/1.0/...`). This bundle targets capability parity with
`internal/connectors/squarespace` (the hand-written connector it migrates); the legacy package stays
registered and unchanged until wave6's registry flip.

## Auth setup

Provide a Squarespace API key via the `api_key` secret; it is sent as a Bearer token
(`Authorization: Bearer <api_key>`) and is never logged, matching legacy's `connsdk.Bearer(key)`
(`squarespace.go:115`). `base_url` defaults to `https://api.squarespace.com/1.0` and may be
overridden for tests/proxies.

## Streams notes

All 4 streams (`orders`, `products`, `inventory`, `profiles`) are `GET` list endpoints. `orders` and
`products` read from `commerce/orders`/`commerce/products` with records at the `result` key;
`inventory` reads `commerce/inventory` (also `result`); `profiles` reads the top-level `profiles`
resource with records at the `profiles` key. Every request sends `limit=100` (matches legacy's
`defaultPageSize`) via each stream's static `query: {"limit": "100"}`.

Pagination follows Squarespace's own `pagination.nextPageCursor` token convention
(`pagination.type: cursor` with `token_path: pagination.nextPageCursor`, `cursor_param: cursor`) —
identical to legacy's `connsdk.CursorPaginator{CursorParam: "cursor", TokenPath:
"pagination.nextPageCursor"}`. No `stop_path` is declared: legacy's paginator has no analogous
falsy-body-value stop signal either (it stops purely when the cursor token itself is empty or a page
yields no records), so the engine's default stop-on-empty-token-only behavior matches exactly.

`orders` and `products` declare `incremental.cursor_field: modifiedOn` to expose the
`incremental_append` sync mode (matching legacy's own `CursorFields: []string{"modifiedOn"}`
declaration on those two streams), but neither this bundle nor legacy ever sends a server-side
lower-bound filter or performs client-side filtering for these streams — both sides read the full
list on every sync. No `request_param`/`client_filtered` is declared, matching legacy's real
(lack of) incremental filtering behavior exactly, not introducing new filtering under the guise of a
migration. `inventory` and `profiles` have no cursor field, matching legacy (full refresh only).

`profiles`' record mapping (`id`, `name`, `createdOn`, `modifiedOn`) mirrors legacy's own
`mapRecord: copyRecord("id", "name", "createdOn", "modifiedOn")` for that stream exactly — legacy
reuses the same field list as `products` for `profiles` (rather than mapping the Profiles API's own
`firstName`/`lastName`/`email` shape), and this bundle preserves that exact legacy behavior rather
than "fixing" it, per the meta-rule that legacy is ground truth over any doc.

## Write actions & risks

None. Legacy's `Write` always returns `connectors.ErrUnsupportedOperation`; `capabilities.write` is
`false` and this bundle ships no `writes.json`.

## Known limits

- **`page_size`/`max_pages` are not runtime-configurable.** Legacy exposes both as config-driven
  overrides (`squarespace.go`'s `pageSize`/`maxPages`, bounded 1-200 / a non-negative integer or
  `all`/`unlimited`). The engine's `cursor`+`token_path` paginator has no config-driven page-size or
  max-pages knob (`PaginationSpec.PageSize`/`MaxPages` are static bundle JSON, never templated), so
  this bundle sends legacy's own default (`limit=100`) as a static per-stream query literal
  (matching stripe's `limit=100` static-query precedent) and does not declare `page_size`/`max_pages`
  in `spec.json` at all (a declared-but-unwireable config key is worse than an absent one, per
  `docs/migration/conventions.md` F6). Pagination is unbounded by default (reads every page until a
  short/empty page or the cursor token stops advancing), matching legacy's own default of
  `maxPages=0` (unbounded) when `max_pages` is unset.
- Full Squarespace Commerce API surface (fulfillments, transactions, store pages, webhooks) is out of
  scope for this wave; see `api_surface.json`'s `excluded: {category: out_of_scope}` entries.
