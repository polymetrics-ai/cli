# Overview

Yotpo is a read-only declarative bundle migrated from `internal/connectors/yotpo` (the
hand-written legacy connector, which stays registered and unchanged until wave6's registry flip).
It reads Yotpo store products, customers, and orders through the Yotpo Core API v3.

## Auth setup

Provide a Yotpo API access token via the `access_token` secret; it is used only for Bearer auth
(`Authorization: Bearer <access_token>`) and is never logged. `store_id` (config, required) scopes
every stream and the `check` request to a specific Yotpo store.

## Streams notes

All 3 streams (`products`, `customers`, `orders`) share the same shape: `GET
/core/v3/stores/{store_id}/{resource}`, records at the resource's own plural key (`products`/
`customers`/`orders`), primary key `["id"]`, cursor field `updated_at`. `store_id` is substituted
into every path (urlencoded by default per `InterpolatePath`) and is a required config value —
matching legacy's `resolveResource`, which errors when `store_id` is unset. Pagination is
`page_number` (`page`/`limit` query params, `page_size: 100`, 1-based), matching legacy's
`connsdk.PageNumberPaginator{PageParam: "page", SizeParam: "limit", StartPage: 1, PageSize:
pageSize}` with legacy's default `pageSize` of 100.

All 3 streams declare `"projection": "passthrough"`. Legacy's `Read` decodes each page via
`connsdk.Harvest` and emits every record verbatim (`emit(connectors.Record(rec))`, `yotpo.go:117-
119`) with no field-building/filtering — `streamEndpoints[stream].fields` is consumed only by
`Catalog` (`yotpo.go:79-81`), never by `Read`. Every real Yotpo field beyond each stream's narrow
catalog schema (e.g. `sku`, `url`, `image_url`, `price` on `products`; `first_name`, `last_name`,
`accepts_marketing` on `customers`; `email`, `total_price`, `financial_status` on `orders`)
survives to the emitted record exactly as legacy would emit it. Declaring the default `"schema"`
projection mode here would silently narrow every emitted record to the catalog schema's
properties — a silent, undocumented parity deviation from legacy's verbatim passthrough — so
`passthrough` is required, matching conventions.md's projection rule (§3) and the
passthrough-decision rule (§8 rule 1: legacy's raw `emit(connectors.Record(rec))` with no
`mapRecord`/field-building is the mechanical signal to use `passthrough`). Each stream's
`schemas/<stream>.json` remains a narrow documentation surface (the fields legacy's `Catalog`
advertises), not an enforced filter — with `passthrough` set, the engine never uses it to drop
fields.

## Write actions & risks

None — this connector is read-only (`capabilities.write: false`), matching legacy's `Write`
returning `connectors.ErrUnsupportedOperation` unconditionally.

## Known limits

- Legacy's runtime-configurable `page_size` (bounded 1-500, default 100) and `max_pages` config
  keys are **not modeled** in this bundle's `spec.json`: `engine.PaginationSpec.PageSize`/`MaxPages`
  are plain (non-templated) JSON integers set once in `streams.json`, with no mechanism to bind a
  runtime `config.*` value into them (F6, `docs/migration/conventions.md`). `page_size` is fixed at
  the bundle level to legacy's own default (100); a caller that previously overrode
  `page_size`/`max_pages` away from that default loses the override, but any caller that used
  legacy's default is unaffected. `max_pages` is likewise not enforced (unbounded reads).
- Full Yotpo API surface (reviews, loyalty, SMS marketing, subscriptions administration) is out of
  scope for wave2; see `api_surface.json`'s `excluded: {category: out_of_scope}` entries.
- `fixtures/streams/products/{page_1,page_2}.json` is the required 2-page pagination fixture
  (page 1 returns 100 records to trigger a next page per `page_number`'s short-page stop rule;
  page 2 returns 1 record and stops). `customers`/`orders` ship single-page fixtures.
