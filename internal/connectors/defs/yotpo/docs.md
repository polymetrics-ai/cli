# Overview

Reads Yotpo store products, product variants, collections, customers, orders, and webhook
targets/filters/subscriptions, and writes
product/variant/order/customer/fulfillment/collection-membership/webhook mutations through the Yotpo
Core API v3.

Readable streams: `products`, `product_variants`, `collections`, `customers`, `orders`,
`webhook_targets`, `webhook_filters`, `webhook_subscriptions`.

Write actions: `create_product`, `update_product`, `create_product_variant`,
`update_product_variant`, `create_order`, `update_order`, `create_customer`,
`create_order_fulfillment`, `update_order_fulfillment`, `create_collection`, `update_collection`,
`add_product_to_collection`, `remove_product_from_collection`, `create_webhook_target`,
`update_webhook_target`, `delete_webhook_target`, `create_webhook_filter`, `update_webhook_filter`,
`delete_webhook_filter`, `create_webhook_subscription`, `update_webhook_subscription`,
`delete_webhook_subscription`.

Service API documentation: https://core-api.yotpo.com/reference/welcome.

## Auth setup

Connection fields:

- `access_token` (required, secret, string); Yotpo API access token, sent as the X-Yotpo-Token
  header (Yotpo Core API v3's own documented auth scheme; generate via POST
  /core/v3/stores/{store_id}/access_tokens using your store ID and API secret, per
  https://core-api.yotpo.com/reference/yotpo-authentication - that token-exchange step is outside
  this connector's scope and must be performed once out-of-band). Never logged.
- `base_url` (optional, string); default `https://api.yotpo.com`; format `uri`; Yotpo API base URL
  override for tests or proxies.
- `mode` (optional, string).
- `product_id` (optional, string); Yotpo product ID (yotpo_id) scoping the product_variants stream's
  path (/products/{product_id}/variants); required only when reading that stream.
- `store_id` (required, string); Yotpo store ID (app_key); substituted into every stream's
  store-scoped path.

Secret fields are redacted in logs and write previews: `access_token`.

Default configuration values: `base_url=https://api.yotpo.com`.

Authentication behavior:

- API key authentication in `X-Yotpo-Token` using `secrets.access_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/core/v3/stores/{{ config.store_id }}/products`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `limit`; starts at
1; page size 100.

Pagination by stream: none: `webhook_targets`, `webhook_filters`, `webhook_subscriptions`;
page_number: `products`, `product_variants`, `collections`, `customers`, `orders`.

- `products`: GET `/core/v3/stores/{{ config.store_id }}/products` - records path `products`;
  page-number pagination; page parameter `page`; size parameter `limit`; starts at 1; page size 100;
  emits passthrough records.
- `product_variants`: GET `/core/v3/stores/{{ config.store_id }}/products/{{ config.product_id
  }}/variants` - records path `variants`; page-number pagination; page parameter `page`; size
  parameter `limit`; starts at 1; page size 100; computed output fields `id`; emits passthrough
  records.
- `collections`: GET `/core/v3/stores/{{ config.store_id }}/collections` - records path
  `collections`; page-number pagination; page parameter `page`; size parameter `limit`; starts at 1;
  page size 100; computed output fields `id`; emits passthrough records.
- `customers`: GET `/core/v3/stores/{{ config.store_id }}/customers` - records path `customers`;
  page-number pagination; page parameter `page`; size parameter `limit`; starts at 1; page size 100;
  emits passthrough records.
- `orders`: GET `/core/v3/stores/{{ config.store_id }}/orders` - records path `orders`; page-number
  pagination; page parameter `page`; size parameter `limit`; starts at 1; page size 100; emits
  passthrough records.
- `webhook_targets`: GET `/core/v3/stores/{{ config.store_id }}/webhooks/targets` - records path
  `webhook_targets`; computed output fields `id`; emits passthrough records.
- `webhook_filters`: GET `/core/v3/stores/{{ config.store_id }}/webhooks/filters` - records path
  `webhook_filters`; computed output fields `id`; emits passthrough records.
- `webhook_subscriptions`: GET `/core/v3/stores/{{ config.store_id }}/webhooks/subscriptions` -
  records path `webhook_subscriptions`; computed output fields `id`; emits passthrough records.

## Write actions & risks

Overall write risk: external mutation: creates/updates products, variants, orders, customers, order
fulfillments, and collections; manages collection membership and webhook target/filter/subscription
lifecycle.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_product`: POST `/core/v3/stores/{{ config.store_id }}/products` - kind `create`; body type
  `json`; required record fields `product`; accepted fields `product`; risk: external mutation;
  creates a new product in the store's catalog; approval required. Body is wrapped under a top-level
  "product" key (Yotpo Core API v3 convention) - the record itself carries that wrapper, since the
  engine's write dialect sends record fields verbatim as the JSON body with no nested-wrapper
  construction primitive (see teamwork/ynab precedent).
- `update_product`: PATCH `/core/v3/stores/{{ config.store_id }}/products/{{ record.yotpo_id }}` -
  kind `update`; body type `json`; path fields `yotpo_id`; required record fields `yotpo_id`,
  `product`; accepted fields `product`, `yotpo_id`; risk: external mutation; updates an existing
  product's catalog fields; approval required.
- `create_product_variant`: POST `/core/v3/stores/{{ config.store_id }}/products/{{
  record.product_yotpo_id }}/variants` - kind `create`; body type `json`; path fields
  `product_yotpo_id`; required record fields `product_yotpo_id`, `variant`; accepted fields
  `product_yotpo_id`, `variant`; risk: external mutation; creates a new variant under an existing
  product; approval required.
- `update_product_variant`: PATCH `/core/v3/stores/{{ config.store_id }}/products/{{
  record.product_yotpo_id }}/variants/{{ record.yotpo_id }}` - kind `update`; body type `json`; path
  fields `product_yotpo_id`, `yotpo_id`; required record fields `product_yotpo_id`, `yotpo_id`,
  `variant`; accepted fields `product_yotpo_id`, `variant`, `yotpo_id`; risk: external mutation;
  updates an existing product variant's fields; approval required.
- `create_order`: POST `/core/v3/stores/{{ config.store_id }}/orders` - kind `create`; body type
  `json`; required record fields `order`; accepted fields `order`; risk: external mutation; creates
  a new order (may trigger Yotpo's automatic review-request email flow for the associated customer);
  approval required. Not possible to send automatic review-request emails for orders older than six
  months (Yotpo's own documented constraint).
- `update_order`: PATCH `/core/v3/stores/{{ config.store_id }}/orders/{{ record.yotpo_id }}` - kind
  `update`; body type `json`; path fields `yotpo_id`; required record fields `yotpo_id`, `order`;
  accepted fields `order`, `yotpo_id`; risk: external mutation; updates an existing order's
  status/pricing/cancellation fields; approval required.
- `create_customer`: POST `/core/v3/stores/{{ config.store_id }}/customers` - kind `create`; body
  type `json`; required record fields `customer`; accepted fields `customer`; risk: external
  mutation; creates or updates (upsert-by-external_id) a customer profile; approval required.
  Yotpo's own endpoint is documented as create-or-update, keyed on external_id - there is no
  separate update_customer action since the same request both creates and upserts.
- `create_order_fulfillment`: POST `/core/v3/stores/{{ config.store_id }}/orders/{{
  record.order_yotpo_id }}/fulfillments` - kind `create`; body type `json`; path fields
  `order_yotpo_id`; required record fields `order_yotpo_id`, `fulfillment`; accepted fields
  `fulfillment`, `order_yotpo_id`; risk: external mutation; records a shipment/fulfillment event
  against an existing order; approval required.
- `update_order_fulfillment`: PATCH `/core/v3/stores/{{ config.store_id }}/orders/{{
  record.order_yotpo_id }}/fulfillments/{{ record.yotpo_id }}` - kind `update`; body type `json`;
  path fields `order_yotpo_id`, `yotpo_id`; required record fields `order_yotpo_id`, `yotpo_id`,
  `fulfillment`; accepted fields `fulfillment`, `order_yotpo_id`, `yotpo_id`; risk: external
  mutation; updates the shipment status/tracking of an existing order fulfillment; approval
  required.
- `create_collection`: POST `/core/v3/stores/{{ config.store_id }}/collections` - kind `create`;
  body type `json`; required record fields `collection`; accepted fields `collection`; risk:
  external mutation; creates a new product collection; approval required.
- `update_collection`: PATCH `/core/v3/stores/{{ config.store_id }}/collections/{{ record.yotpo_id
  }}` - kind `update`; body type `json`; path fields `yotpo_id`; required record fields `yotpo_id`,
  `collection`; accepted fields `collection`, `yotpo_id`; risk: external mutation; renames an
  existing product collection; approval required.
- `add_product_to_collection`: POST `/core/v3/stores/{{ config.store_id }}/collections/{{
  record.collection_yotpo_id }}/products` - kind `create`; body type `json`; path fields
  `collection_yotpo_id`; required record fields `collection_yotpo_id`, `product_id`; accepted fields
  `collection_yotpo_id`, `product_id`; risk: external mutation; adds a product to an existing
  collection; approval required.
- `remove_product_from_collection`: DELETE `/core/v3/stores/{{ config.store_id }}/collections/{{
  record.collection_yotpo_id }}/products` - kind `delete`; body type `json`; path fields
  `collection_yotpo_id`; body fields `product_id`; required record fields `collection_yotpo_id`,
  `product_id`; accepted fields `collection_yotpo_id`, `product_id`; missing records treated as
  success for status `404`; risk: irreversible external mutation; removes a product from an existing
  collection; approval required.
- `create_webhook_target`: POST `/core/v3/stores/{{ config.store_id }}/webhooks/targets` - kind
  `create`; body type `json`; required record fields `webhook_target`; accepted fields
  `webhook_target`; risk: external mutation; registers a webhook callback URL target; approval
  required.
- `update_webhook_target`: PATCH `/core/v3/stores/{{ config.store_id }}/webhooks/targets/{{
  record.yotpo_id }}` - kind `update`; body type `json`; path fields `yotpo_id`; required record
  fields `yotpo_id`, `webhook_target`; accepted fields `webhook_target`, `yotpo_id`; risk: external
  mutation; changes an existing webhook target's callback URL; approval required.
- `delete_webhook_target`: DELETE `/core/v3/stores/{{ config.store_id }}/webhooks/targets/{{
  record.yotpo_id }}` - kind `delete`; body type `none`; path fields `yotpo_id`; required record
  fields `yotpo_id`; accepted fields `yotpo_id`; missing records treated as success for status
  `404`; risk: irreversible external deletion; removes a registered webhook target (any subscription
  still referencing it becomes inactive); approval required.
- `create_webhook_filter`: POST `/core/v3/stores/{{ config.store_id }}/webhooks/filters` - kind
  `create`; body type `json`; required record fields `webhook_filter`; accepted fields
  `webhook_filter`; risk: external mutation; creates a webhook event filter (an event type cannot be
  used twice across filters, per Yotpo's own constraint); approval required.
- `update_webhook_filter`: PATCH `/core/v3/stores/{{ config.store_id }}/webhooks/filters/{{
  record.yotpo_id }}` - kind `update`; body type `json`; path fields `yotpo_id`; required record
  fields `yotpo_id`, `webhook_filter`; accepted fields `webhook_filter`, `yotpo_id`; risk: external
  mutation; changes an existing webhook filter's subscribed event types; approval required.
- `delete_webhook_filter`: DELETE `/core/v3/stores/{{ config.store_id }}/webhooks/filters/{{
  record.yotpo_id }}` - kind `delete`; body type `none`; path fields `yotpo_id`; required record
  fields `yotpo_id`; accepted fields `yotpo_id`; missing records treated as success for status
  `404`; risk: irreversible external deletion; removes a webhook filter (only unused filters can be
  deleted, per Yotpo's own constraint); approval required.
- `create_webhook_subscription`: POST `/core/v3/stores/{{ config.store_id }}/webhooks/subscriptions`
  - kind `create`; body type `json`; required record fields `webhook_subscription`; accepted fields
  `webhook_subscription`; risk: external mutation; activates webhook event delivery by combining an
  existing target and filter; approval required.
- `update_webhook_subscription`: PATCH `/core/v3/stores/{{ config.store_id
  }}/webhooks/subscriptions/{{ record.yotpo_id }}` - kind `update`; body type `json`; path fields
  `yotpo_id`; required record fields `yotpo_id`, `webhook_subscription`; accepted fields
  `webhook_subscription`, `yotpo_id`; risk: external mutation; retargets or (de)activates an
  existing webhook subscription; approval required.
- `delete_webhook_subscription`: DELETE `/core/v3/stores/{{ config.store_id
  }}/webhooks/subscriptions/{{ record.yotpo_id }}` - kind `delete`; body type `none`; path fields
  `yotpo_id`; required record fields `yotpo_id`; accepted fields `yotpo_id`; missing records treated
  as success for status `404`; risk: irreversible external deletion; stops webhook event delivery
  for an existing target/filter combination; approval required.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 8 stream-backed endpoint group(s), 22 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  deprecated=1, duplicate_of=11, out_of_scope=1.
