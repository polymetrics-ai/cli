# Overview

SendOwl is a digital-product delivery and e-commerce platform. This bundle is a Pass B
full-surface expansion: it reads SendOwl orders, products, subscriptions, discounts, bundles
(packages), and per-product license keys, and writes product/subscription/discount/bundle
lifecycle mutations plus order-level actions (refund, cancel subscription, resend email) through
the SendOwl API (v1 / v1_2 / v1_3, per resource — see Auth setup). It supersedes the earlier
read-only wave2 shape (`Capabilities{Write: false}` in legacy `internal/connectors/sendowl`, which
stays registered and unchanged until wave6's registry flip); legacy itself never wrote anything, so
every write action here is genuinely new surface, not a legacy behavior port.

## Auth setup

Provide a SendOwl API key/secret pair via the `username` config value (API key) and the `password`
secret (API secret). They are sent as HTTP Basic credentials via `base.auth`'s `mode: basic`,
identical to legacy's `connsdk.Basic(username, password)`. The password is never logged.

`base_url` now defaults to the bare host `https://www.sendowl.com`, not
`https://www.sendowl.com/api/v1` as in the prior (read-only) wave2 shape: SendOwl's real documented
API spans 3 path-versioned segments (`/api/v1` for products/subscriptions/packages/licenses,
`/api/v1_2` for discounts, `/api/v1_3` for the orders list/show/update endpoints), so every
stream/write's own `path` now carries its own version segment explicitly instead of relying on a
single fixed prefix baked into `base_url`. The final resolved URL for every pre-existing
(`orders`/`products`/`subscriptions`) stream and the `check` request is byte-identical to the prior
shape — this is a pure config-vs-path split, not a behavior change.

## Streams notes

`orders`, `products`, `subscriptions` (the original 3 legacy-parity streams): unchanged shape,
`GET` against the SendOwl list endpoint, records at the top-level JSON array (`records.path:
"."`), `projection: passthrough` (SendOwl list endpoints emit every raw field; passthrough avoids
silently dropping any field not enumerated in `schemas/*.json`). Only `orders` declares
`x-cursor-field: created_at`, matching legacy's `CursorFields: ["created_at"]`; no stream declares
an `incremental` block, since neither legacy nor the real SendOwl list endpoints support a
server-side updated-since filter for these resources.

`discounts` (`GET /api/v1_2/discounts`) and `bundles` (`GET /api/v1/packages`, SendOwl's
"packages" resource — displayed to users as "bundles") are new top-level list streams, same shape
as the original 3: `records.path: "."`, `projection: passthrough`, primary key `["id"]`.

`licenses` (`GET /api/v1/products/{{ fanout.id }}/licenses`) is a sub-resource fan-out over
`products`: SendOwl's real API only exposes license keys scoped to a single product id (or a
single order id — see Known limits), with no top-level `/licenses` list endpoint. `fan_out.ids_from
.request` issues the SAME paginated `GET /api/v1/products` request the `products` stream itself
uses (`records_path: "."`, `id_field: "id"`, reusing this stream's effective pagination — the base
`page_number`/`max_pages: 1` spec, since `licenses` declares no pagination override of its own),
then `into.path_var: "parent_id"` threads each discovered product id into
`/api/v1/products/{{ fanout.id }}/licenses`, and `stamp_field: "product_id"` writes the source
product id onto every emitted license record after projection. `stamp_field` always writes the
fan-out id as a STRING (`fanoutContext.id` is always `string` — `read.go`'s fan-out resolution
extracts every id as text regardless of the parent record's own JSON type), so `licenses.json`
declares `product_id: ["string", "null"]` even though SendOwl's real product ids are numeric — this
is a fan_out-dialect-wide constraint, not a sendowl-specific choice.

Pagination is `page_number` (`page_param: page`, `size_param: per_page`, `start_page: 1`,
`page_size: 100`, `max_pages: 1`) for every stream, identical to legacy's
`connsdk.PageNumberPaginator{PageParam: "page", SizeParam: "per_page", StartPage: 1, PageSize:
pageSize}` at legacy's own default `pageSize = 100`/`maxPages = 1`. See Known limits (carried over
from the prior wave2 shape) for why `page_size`/`max_pages` are not runtime-configurable.

## Write actions & risks

13 write actions, none of which existed in legacy (legacy's `Write` always returned
`connectors.ErrUnsupportedOperation`; this is genuinely new capability surface, not a ported
behavior):

- **`create_product`/`update_product`/`delete_product`** (`POST`/`PUT`/`DELETE
  /api/v1/products[/{{ record.id }}]`, `body_type: form`): product lifecycle. `create_product`
  intentionally covers ONLY the metadata-only create shape (name/price/currency_code/product_type)
  — SendOwl's real product-create ALSO accepts a `multipart/form-data` file attachment (the actual
  digital asset), which this dialect's `body_type` (`json`/`form`/`none`) cannot express (see Known
  limits). `delete_product` is `idempotent`/`missing_ok_status: [404]`.
- **`create_subscription`/`update_subscription`/`delete_subscription`** (same shape as products,
  against `/api/v1/subscriptions`): recurring-billing product lifecycle.
- **`create_discount`/`update_discount`/`delete_discount`** (against `/api/v1_2/discounts`):
  discount-code lifecycle (percentage/flat-rate, usage cap, expiry).
- **`update_bundle`/`delete_bundle`** (against `/api/v1/packages/{{ record.id }}`): bundle
  (package) mutation/removal. There is deliberately no `create_bundle`: SendOwl's real bundle-create
  endpoint requires selecting component products in a `multipart/form-data` body, the same
  binary-payload gap that narrows `create_product` (see Known limits) — `update_bundle` and
  `delete_bundle` remain fully covered since they operate on an EXISTING bundle by id with no file
  payload.
- **`refund_order`** (`POST /api/v1/orders/{{ record.id }}/refund`, `body_fields:
  ["amount","cancel_subscription","revoke_access"]`): issues a REAL financial refund against the
  buyer's original payment method. Highest-risk action in this bundle; irreversible external money
  movement.
- **`cancel_order_subscription`** (`PUT /api/v1/orders/{{ record.id }}/cancel_subscription`,
  `body_type: none`): cancels the buyer's active recurring subscription tied to this specific
  order. Distinct from `delete_subscription` (which removes the subscription PRODUCT definition
  itself, not a buyer's individual active subscription).
- **`resend_order_email`** (`POST /api/v1/orders/{{ record.id }}/resend_email`, `body_fields:
  ["type"]`): resends the order confirmation/receipt/download email; low external side effect, no
  approval required (all other actions in this bundle require reverse-ETL plan approval per
  `metadata.json`'s `risk.approval`).

## Known limits

- **Multipart file-upload creation is out of scope (`binary_payload`, `api_surface.json`).**
  SendOwl's real `POST /api/v1/products` and `POST /api/v1/packages` both accept an optional
  `multipart/form-data` file attachment (the sellable digital asset itself) alongside ordinary
  metadata fields; this dialect's `write.go` `body_type` set (`json`/`form`/`none`) has no
  multipart support at all. `create_product` covers ONLY the metadata-only create shape (SendOwl
  does accept a fileless product create for e.g. license-key-only or externally-hosted products);
  bundle creation has no fileless equivalent in the same way, so `create_bundle` is entirely
  excluded rather than partially modeled.
- **`licenses` has no top-level list endpoint; only product-scoped and order-scoped variants
  exist.** This bundle fans out over `products` (`GET /api/v1/products/{id}/licenses`); the
  order-scoped variant (`GET /api/v1/orders/{order_id}/licenses`) is excluded as `duplicate_of` —
  it returns the identical license record shape, just reachable through a different parent
  resource, and modeling BOTH fan-outs would double-count no new fields, only a different join key.
  A caller who specifically needs order-scoped license lookup is out of scope for this wave.
- **`licenses`' fan-out id-listing request reuses the base `page_number`/`max_pages: 1` spec**, so
  only the FIRST page (up to 100) of `products` is walked for ids to fan out over — a product
  catalog with more than 100 products has license records for products beyond the 100th silently
  unreached by this stream. This is the same `max_pages: 1`-baked-in limitation the pre-existing
  streams already carry (see below), now also affecting which PARENT ids the fan-out discovers, not
  just how many pages of a single resource are read.
- **`base_url`'s default changed shape (not value) from the prior wave2 read-only bundle**: it is
  now the bare host `https://www.sendowl.com` with each stream/write path carrying its own
  `/api/v1`, `/api/v1_2`, or `/api/v1_3` segment, rather than a single fixed `/api/v1` baked into
  `base_url` with every path relative to it. The final resolved URL for every pre-existing
  stream/check request is unchanged; only a caller who had explicitly overridden `base_url` to a
  custom value INCLUDING a version suffix (e.g. a test proxy at `https://proxy.example.com/api/v1`)
  would need to update that override to the bare host instead. Not exercised by any fixture (all
  default to the real SendOwl host).
- **`max_pages` is not a runtime config override.** Legacy accepted a `max_pages` config value
  (default `1`) and threaded it through to `connsdk.Harvest`'s page-count cap for its 3 read-only
  streams. The engine's declarative pagination has no per-read config-driven override mechanism for
  `PaginationSpec.MaxPages` — it is a fixed value baked into `streams.json`'s `base.pagination`
  block (the same "no runtime override" limitation `docs/migration/conventions.md` documents for
  searxng's `page_size`/`max_pages`). This bundle bakes in `max_pages: 1`, matching legacy's own
  default exactly for the 3 original streams (a documented, pre-existing config-surface narrowing);
  the 3 new list streams (`discounts`/`bundles`) and the `licenses` fan-out inherit the identical
  base spec since SendOwl has no legacy precedent to diverge from for these.
- **`page_size` is not a runtime config override either, and is not declared in `spec.json`** — see
  Streams notes: the `page_number` paginator's own per-page query entry unconditionally overwrites
  any same-keyed stream-level `query` entry, so a `spec.json` property feeding it would be dead
  config (F6, `docs/migration/conventions.md`).
- **Fixtures are single-page, matching the `max_pages: 1` cap**, for every stream including the two
  legs of the `licenses` fan-out (the id-listing request and the per-product-id licenses request).
  This mirrors the identical, already-accepted `searxng`/pre-Pass-B-sendowl precedent — see
  `docs/migration/conventions.md` §4.
- Full SendOwl API surface still has documented exclusions beyond what's covered here (search/
  shopify_lookup aliases, discount sub-collections, order access-grant/revoke, bulk license
  import, single license-key validation) — see `api_surface.json`'s per-endpoint `excluded`
  entries, each with a specific closed-vocabulary category and reason (no blanket "Pass B
  capability expansion" placeholders remain).
