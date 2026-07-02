# Overview

ShopWired is a wave2 fan-out declarative-HTTP migration. It reads ShopWired products, orders,
customers, and categories through the ShopWired REST API (`GET
https://api.shopwired.co.uk/...`). This bundle targets capability parity with
`internal/connectors/shopwired` (the hand-written connector it migrates); the legacy package stays
registered and unchanged until wave6's registry flip. ShopWired is read-only (`capabilities.write`
is `false`, matching legacy's `Write` returning `connectors.ErrUnsupportedOperation`).

## Auth setup

Provide a ShopWired API key via the `api_key` secret; it is sent as the `X-API-Key` header (no
prefix) and never logged, matching legacy's `connsdk.APIKeyHeader("X-API-Key", token, "")`
(`shopwired.go:130`). `base_url` defaults to `https://api.shopwired.co.uk` (legacy's
`shopwiredDefaultBaseURL`) and may be overridden for tests/proxies.

## Streams notes

All four streams (`products`, `orders`, `customers`, `categories`) share the identical shape: `GET
/<resource>?page=<n>&per_page=<size>` (legacy's `harvest`, `shopwired.go:95-119`), returning a
top-level JSON array (`records.path: "."`, matching legacy's `recordsPath: ""` /
`connsdk.RecordsAt(resp.Body, "")` root-array extraction), 1-based page-number pagination
(`pagination.type: page_number`, `page_param: page`, `size_param: per_page`, `start_page: 1`,
`page_size: 100` — legacy's own `shopwiredDefaultPageSize`); a page shorter than 100 records
signals the end, matching legacy's `len(records) < pageSize` stop condition exactly — the engine's
`PageNumberPaginator.Next` uses the identical `recordCount < PageSize` rule.

Every stream shares legacy's single generic `shopwiredRecord` mapper (`shopwired.go:146-148`),
which tries `id` before a defensive `order_id` fallback, `name` before a defensive `title`
fallback, and `updated_at` before a defensive `modified_at` fallback; this bundle's
`computed_fields` reference the FIRST-tried key on every stream (`id`, `name`, `updated_at`) — see
Known limits for the untaken fallback keys. `sku` (products), `email` (customers), and `status`
(orders) pass straight through via schema projection since legacy reads them directly
(`item["sku"]`/`item["email"]`/`item["status"]`, no fallback); `status`/`sku`/`email` are declared
only on the schema(s) whose legacy `Fields` catalog lists them (matching legacy's per-stream
`fields()` catalog exactly, even though the shared mapper function itself always sets all of
`sku`/`email`/`status` regardless of stream — projection mode drops the ones the schema omits).

None of the four streams expose a request-time incremental filter parameter in legacy — legacy's
own `Stream.CursorFields` catalog metadata (`updated_at` on all four) is purely descriptive
cataloging, never wired into an actual request-narrowing filter anywhere in `harvest`. This bundle
mirrors that exactly: `x-cursor-field` is declared on each schema (matching legacy's catalog
metadata) but no stream declares an `incremental` block, so every read is full refresh, identical to
legacy.

## Write actions & risks

None. ShopWired's product/order/customer/category read endpoints have no obviously-safe
reverse-ETL writes in legacy (legacy's own package doc: "read-only native ShopWired connector");
`capabilities.write` is `false` and this bundle ships no `writes.json`, matching legacy's `Write`
returning `connectors.ErrUnsupportedOperation`.

## Known limits

- **Legacy's `order_id`/`title`/`modified_at` fallback keys are not modeled.** Legacy's
  `first(item, "id", "order_id")`, `first(item, "name", "title")`, and `first(item, "updated_at",
  "modified_at")` try a fallback key SECOND, only when the first-choice key is absent on a given
  record. The dialect has no coalesce/fallback filter (only a single template per computed field),
  so the untaken second-choice keys cannot be expressed without a `RecordHook` (a Tier-2 escalation
  this wave's JSON-only mandate forbids). Per conventions.md §5's meta-rule, this is a documented
  scope narrowing: ShopWired's own current API reference (`help.shopwired.co.uk`) confirms `id` and
  a product's `title` (not `name`) as the real field names, so legacy's own `name`-first,
  `title`-second choice is itself already a divergence from the current live API for products —
  this bundle intentionally preserves LEGACY's exact fallback ORDER (per the mission's "legacy is
  ground truth over any doc" rule) rather than "fixing" it to the newer documented shape, and simply
  cannot express the untaken second key at all.
- **Legacy's `status` passthrough assumes a scalar; ShopWired's current documented `orders.status`
  is a nested `{id, name, type}` object.** Legacy's `mapRecord` does `item["status"]` verbatim with
  no type assumption, so this bundle's plain schema-projection passthrough (`"status": { "type":
  ["string", "object", "null"] }`) is equally type-agnostic and reproduces legacy's behavior exactly
  either way (both sides copy whatever value is present, scalar or object).
- **`page`/`per_page`/`max_pages` config overrides are not modeled.** Legacy exposes `page_size`/
  `limit` (1-250, default 100) and `max_pages` (0/all/unlimited or a positive integer cap) as
  config-driven overrides (`pageSize`/`maxPages`, `shopwired.go:215-239`). The engine's
  `page_number` paginator's `PageSize` is a fixed bundle-declared value (`streams.json`'s
  `base.pagination.page_size: 100`, matching legacy's own default) with no per-request config
  override wired in this bundle; `page_size`/`limit`/`max_pages` are therefore not declared in
  `spec.json`. `max_pages` is likewise not modeled; pagination is bounded only by the short-page
  stop signal.
