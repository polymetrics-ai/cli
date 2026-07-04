# Overview

Yotpo is a declarative bundle migrated from `internal/connectors/yotpo` (the hand-written legacy
connector, which stays registered and unchanged until wave6's registry flip). It reads Yotpo store
products, product variants, collections, customers, orders, and webhook targets/filters/
subscriptions, and writes product/variant/order/customer/fulfillment/collection-membership/webhook
mutations, through the Yotpo Core API v3. Pass B full-surface expansion also corrected this
bundle's `docs_url`: the pre-Pass-B bundle pointed at `apidocs.yotpo.com` (Yotpo's separate
Reviews/UGC API reference), which never documented the `/core/v3/...` products/customers/orders
endpoints this bundle actually calls — the real source of truth is `core-api.yotpo.com`.

## Auth setup

Provide a Yotpo API access token via the `access_token` secret; it is sent as the `X-Yotpo-Token`
header on every request (Yotpo Core API v3's own documented auth scheme — see
`https://core-api.yotpo.com/reference/yotpo-authentication`) and is never logged. Generating that
token is a one-time `POST /core/v3/stores/{store_id}/access_tokens` exchange using the store's
API secret; that exchange is a token-generation flow this dialect's declarative `auth` candidates
cannot express (it would require a new Tier-2 `AuthHook`, which this pass is not authorized to add
— no new hook packages), so it is performed once out-of-band and the resulting token is supplied
directly as `secrets.access_token` (see `api_surface.json`'s `out_of_scope` entry for
`/access_tokens`). `store_id` (config, required) scopes every stream and the `check` request to a
specific Yotpo store; `product_id` (config, optional) scopes the `product_variants` stream to one
product.

## Streams notes

8 streams, all under `/core/v3/stores/{store_id}/...` (urlencoded by default per
`InterpolatePath`):

- `products` (`GET .../products`, records at `products`): full product-catalog fields
  (`yotpo_id`/`external_id`/`name`/`price`/`sku`/`gtins`/`custom_properties`/etc, per Yotpo's
  documented response shape). `computed_fields` aliases `id` from the bare `yotpo_id` reference
  (typed extraction preserves its native integer type) — the primary key every stream in this
  bundle now uses is the numeric `yotpo_id`, not a synthetic string, correcting the pre-Pass-B
  bundle's schema (which declared a bare `id: string` field the real API never actually returns).
- `product_variants` (`GET .../products/{product_id}/variants`, records at `variants`): one
  product's variant set (color/size/etc option combinations); same `yotpo_id`→`id` aliasing.
- `collections` (`GET .../collections`, records at `collections`): product-collection groupings.
- `customers` (`GET .../customers`, records at `customers`): customer profile fields. Unlike every
  other stream, Yotpo customers have **no `yotpo_id` at all** — `external_id` (the caller-supplied
  identifier) is the customer's only natural key, so `computed_fields` aliases `id` from
  `external_id` instead, and `updated_at` from the native `account_updated_at` field.
- `orders` (`GET .../orders`, records at `orders`): order fields including embedded
  `customer`/`billing_address`/`shipping_address`/`fulfillments[]`/`line_items[]` nested objects
  (passthrough preserves all of them verbatim — order fulfillments are NOT modeled as a separate
  read stream since they are already present here, per `api_surface.json`). `updated_at` is
  aliased from `order_date` (orders have no separate last-modified timestamp field in Yotpo's
  documented response).
- `webhook_targets` / `webhook_filters` / `webhook_subscriptions` (`GET
  .../webhooks/{targets,filters,subscriptions}`, records at the matching plural envelope key):
  Yotpo's 3-part webhook configuration model — a target (callback URL), a filter (subscribed event
  type list), and a subscription (activates delivery by combining one target + one filter). All 3
  are `pagination: none` (Yotpo's webhook endpoints are not documented as paginated, unlike the
  cursor-paginated catalog/order endpoints).

**Pagination — kept at legacy's existing `page_number` shape, not migrated to Yotpo's real
cursor-based `page_info`/`next_page_info` scheme**: Yotpo's actual v3 API documents
`page_info`/`limit` cursor pagination (an opaque server-issued token, not a page number) for every
list endpoint. This bundle's pre-existing `products`/`customers`/`orders` streams (and the two new
streams that share the same base pagination block, `product_variants`/`collections`) instead
declare `page_number` (`page`/`limit` params, 1-based) — the legacy connector's own
`connsdk.PageNumberPaginator` shape, which the real API does not document supporting. This is a
carried-forward pre-Pass-B parity choice (Pass B's mandate is full-surface expansion of covered
resources, not re-deriving the base request/pagination mechanics of already-migrated streams) —
flagged here as a known limit rather than silently re-verified as correct; see below.

All 8 streams declare `"projection": "passthrough"`, matching legacy's verbatim
`emit(connectors.Record(rec))` behavior for the original 3 streams and this bundle's general
policy of never schema-narrowing an externally-owned, deeply-nested JSON API response without a
field-for-field verified mapping (§8 rule 1). Each stream's `schemas/<stream>.json` is a
documentation surface only.

## Write actions & risks

21 write actions, all requiring approval (`capabilities.write: true`):

- `create_product` / `update_product` (`POST`/`PATCH .../products[/{yotpo_id}]`)
- `create_product_variant` / `update_product_variant` (`POST`/`PATCH
  .../products/{product_yotpo_id}/variants[/{yotpo_id}]`)
- `create_order` / `update_order` (`POST`/`PATCH .../orders[/{yotpo_id}]`): creating an order may
  trigger Yotpo's automatic review-request email flow for the associated customer; Yotpo's own
  docs note this is not possible for orders older than six months.
- `create_customer` (`POST .../customers`): Yotpo's own endpoint is documented as create-or-update
  (upsert keyed on `external_id`) — there is no separate `update_customer` action since the same
  request both creates and upserts.
- `create_order_fulfillment` / `update_order_fulfillment` (`POST`/`PATCH
  .../orders/{order_yotpo_id}/fulfillments[/{yotpo_id}]`): records/updates a shipment event.
- `create_collection` / `update_collection` (`POST`/`PATCH .../collections[/{yotpo_id}]`)
- `add_product_to_collection` / `remove_product_from_collection` (`POST`/`DELETE
  .../collections/{collection_yotpo_id}/products`): collection-membership mutations.
  `remove_product_from_collection` is `kind: delete` with a JSON body (`body_fields: ["product_id"]`
  — Yotpo's DELETE endpoint here takes a request body identifying which product to remove, not a
  path-parameterized single-resource delete, so `body_fields` is used instead of the default
  path-only delete shape).
- `create_webhook_target` / `update_webhook_target` / `delete_webhook_target`,
  `create_webhook_filter` / `update_webhook_filter` / `delete_webhook_filter`,
  `create_webhook_subscription` / `update_webhook_subscription` / `delete_webhook_subscription`
  (full CRUD on all 3 webhook configuration resources, all under `.../webhooks/{targets,filters,
  subscriptions}[/{yotpo_id}]`). Yotpo's own documented constraints (not enforced client-side,
  surfaced as ordinary HTTP errors via `error_map`): an event type cannot be used twice across
  filters; only unused filters/targets can be deleted.

All 21 actions use Yotpo's own top-level wrapper-key request-body convention (`{"product": {...}}`,
`{"variant": {...}}`, `{"order": {...}}`, `{"customer": {...}}`, `{"fulfillment": {...}}`,
`{"collection": {...}}`, `{"webhook_target": {...}}`, `{"webhook_filter": {...}}`,
`{"webhook_subscription": {...}}`) — the engine's write dialect has no nested-wrapper body
construction primitive, so each action's `record_schema` declares the wrapper key itself as a
required nested-object field, and the caller-supplied record already carries that shape (see
teamwork's `create_project`/YNAB's identical sanctioned pattern, documented in
`docs/migration/conventions.md`).

## Known limits

- Pagination is NOT migrated to Yotpo's real documented cursor-based `page_info`/`limit` scheme
  (see "Streams notes" above) — this bundle's `page_number` pagination is a carried-forward
  pre-Pass-B choice matching the original 3-stream migration's parity target (legacy's
  `connsdk.PageNumberPaginator`), not the live API's actual pagination contract. A future increment
  correcting this to genuine `page_info` cursor pagination is a real, documented gap — every stream
  beyond the first page of any real collection larger than 100 items may not paginate correctly
  against the live API today. Flagged, not silently worked around.
- Auth token *generation* (`POST /access_tokens`, store_id + API secret → access token) is out of
  scope — see "Auth setup" above; only the resulting already-generated token is modeled as this
  bundle's `access_token` secret.
- `orders`' embedded `fulfillments[]` array is read-only via the `orders` stream itself (passthrough
  preserves it verbatim); there is no separate `order_fulfillments` read stream, since Yotpo has no
  bulk cross-order fulfillment-listing endpoint — only a per-order one, which would duplicate data
  the `orders` stream already carries (`api_surface.json`: `duplicate_of`).
- The "aggregated order info" simplified-integration batch endpoint (`POST /purchases`) is not
  modeled — `create_order` plus `create_order_fulfillment` already cover the equivalent data
  mutations at their own native per-resource granularity (`api_surface.json`: `duplicate_of`).
- Custom customer events (`PATCH /customer_events`) are excluded: Yotpo's own documentation marks
  the endpoint deprecated and scheduled for removal.
- `fixtures/streams/products/{page_1,page_2}.json` is the required 2-page pagination fixture
  (page 1 returns 100 records to trigger a next page per `page_number`'s short-page stop rule;
  page 2 returns 1 record and stops). `product_variants`/`collections` inherit the same base
  pagination block but ship single-page fixtures (each returns fewer than `page_size` records,
  the natural short-page stop). `customers`/`orders`/`webhook_targets`/`webhook_filters`/
  `webhook_subscriptions` ship single-page fixtures (`webhook_*` streams declare
  `pagination: none` outright).
