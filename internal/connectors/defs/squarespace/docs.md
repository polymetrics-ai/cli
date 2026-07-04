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

### Pass B additions

Four new read streams, added against the real, live OpenAPI spec (fetched directly from
`https://developers.squarespace.com/commerce-apis/latest/schema-processor-version-version-latest.json`
— see `api_surface.json`'s `scope` note):

- **`transactions`** — `GET /1.0/commerce/transactions`, cursor-paginated identically to the
  legacy-parity streams, records at `documents`.
- **`store_pages`** — `GET /1.0/commerce/store_pages`, cursor-paginated, records at `storePages`.
- **`webhook_subscriptions`** — `GET /1.0/webhook_subscriptions`; this list endpoint is genuinely
  non-paginated in the real API (no `cursor` query parameter documented), so the stream declares
  `pagination: {"type": "none"}` to override the base-level cursor pagination block. Records at
  `webhookSubscriptions`.
- **`contacts`** — `GET /v1/contacts` on the separate Contacts API v1 surface. The real API's base
  host is the same (`api.squarespace.com`) but the path prefix is `/v1/contacts`, not `/1.0/...`
  like every Commerce API stream — this bundle's `base_url` default already bakes in the `/1.0`
  Commerce API version segment (an existing, unchanged wave2 decision; see Known limits), so this
  stream's `path` is declared as a full absolute URL (`https://api.squarespace.com/v1/contacts`),
  bypassing `base.url`/`config.base_url` entirely for this one stream — the same sanctioned pattern
  `defillama`'s `stablecoins` stream uses for a differently-hosted resource. Pagination uses the
  same `pagination.nextPageCursor` cursor convention as every other stream (uniform across
  Squarespace's whole API surface), but the page-size query parameter is named `pageSize` here
  (Contacts API's own `PaginationParameters` schema), not `limit` like the Commerce API streams.

## Write actions & risks

Two write actions, both against the real, live OpenAPI spec:

- **`create_webhook_subscription`** (`POST /1.0/webhook_subscriptions`) — registers a new HTTPS
  endpoint (`endpointUrl`) and optional `topics` array to receive live order/contact/address event
  notifications; low risk, no approval required.
- **`delete_webhook_subscription`** (`DELETE /1.0/webhook_subscriptions/{{ record.id }}`) —
  idempotent delete (`missing_ok_status: [404]`); destructive, approval required.

Both resolve against the Commerce API's own `base_url` (`/1.0`-relative paths), so they are fully
exercised by `write_request_shape` conformance replay like any ordinary write action.

`create_contact`/`delete_contact` (Contacts API v1: `POST`/`DELETE /v1/contacts...`) were
evaluated and NOT added as write actions, even though their request bodies are flat enough to
express: the Contacts API v1 surface lives at a different version prefix (`/v1/...`) than
`base_url`'s baked-in `/1.0` Commerce API segment, so reaching it requires the same
absolute-URL `path` override the `contacts` **read** stream uses (see Streams notes) — but unlike
read streams, write actions have no `conformance.skip_dynamic` equivalent: `write_request_shape`'s
capture-server replay always points `b.HTTP.URL` at the test double, and an absolute-URL
`action.path` bypasses that entirely, so a `create_contact`/`delete_contact` action could never be
proven correct by this repo's conformance harness. Shipping an untestable write action was judged
worse than not shipping it; see `api_surface.json`'s `out_of_scope` entries for
`POST`/`DELETE /v1/contacts...` for the full reasoning.

Every other mutation endpoint in the real API (`update_product`, `update_webhook_subscription`,
order fulfillment, inventory adjustments, product/variant create) was also evaluated and excluded —
see `api_surface.json` for the endpoint-by-endpoint reasoning. The unifying theme there: Squarespace's
own partial-update convention wraps every changed field as `{"present": true, "value": ...}`, and its
bulk-operation endpoints (fulfillments, inventory adjustments) take named arrays of nested
`{variantId, quantity}`/`{carrierName, service, shipDate, trackingNumber}` objects — neither shape
is constructible by the engine's default JSON write body (every record field except `path_fields`,
copied verbatim as a flat top-level key) without inventing a Tier-2 `WriteHook`, which this
declarative-only Pass B pass does not add.

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
- **The wave2 `products` stream's path does not match the real API's documented v2 path.** This
  bundle's `products` stream (unchanged from wave2, preserved at parity per the meta-rule against
  altering already-migrated behavior) requests `GET {{ config.base_url }}/commerce/products`
  against a `base_url` default of `https://api.squarespace.com/1.0` — i.e.
  `/1.0/commerce/products`. The real, live Squarespace API serves the Products resource at
  `/v2/commerce/products` (a distinct API version with a different response envelope — `products`
  is the real records key, matching this stream's declared `records.path`, but the real pagination
  block lives under `PaginatedProductListResponseV2`, confirmed identical in shape to what this
  bundle already expects). This is a pre-existing wave2 discrepancy discovered during this Pass B
  research pass, not introduced by it; flagged here for a future capability-expansion or bug-fix
  pass rather than silently corrected, since correcting it would change the resolved request URL
  for an already-migrated, already-tested stream outside this task's scope (`defs/squarespace/`
  streams.json's `orders`/`inventory`/`profiles` streams are unaffected — only `products` uses a
  versioned path on the real API).
- **New streams use the real API's version-correct absolute/relative paths** (`/1.0/commerce/
  transactions`, `/1.0/commerce/store_pages`, `/1.0/webhook_subscriptions` relative to the existing
  `/1.0`-inclusive `base_url`; `contacts` as a full absolute URL bypassing `base_url` for the
  differently-versioned Contacts API) — these were authored directly against the live OpenAPI spec
  and do not inherit the `products` stream's pre-existing version mismatch.
- Full endpoint-by-endpoint accounting, including every excluded mutation and the reasoning for
  each, is in `api_surface.json`.
