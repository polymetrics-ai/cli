---
name: pm-yotpo
description: Yotpo connector knowledge and safe action guide.
---

# pm-yotpo

## Purpose

Reads Yotpo store products, product variants, collections, customers, orders, and webhook targets/filters/subscriptions, and writes product/variant/order/customer/fulfillment/collection-membership/webhook mutations through the Yotpo Core API v3.

## Icon

- id: yotpo
- asset: icons/yotpo.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://apidocs.yotpo.com/reference

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- mode
- product_id
- store_id (required)
- access_token (secret) (required)

## ETL Streams

- products:
  - primary key: yotpo_id
  - cursor: updated_at
  - fields: brand(string), compare_at_price(integer), created_at(string), currency(string), custom_properties(object), description(string), external_id(string), group_name(string), gtins(array), image_url(string), inventory_quantity(integer), is_discontinued(boolean), is_valid_url(boolean), mpn(string), name(string), price(integer), sku(string), status(string), updated_at(string), url(string), yotpo_id(integer)
- product_variants:
  - primary key: id
  - cursor: updated_at
  - fields: compare_at_price(integer), created_at(string), currency(string), description(string), external_id(string), gtins(array), id(integer), image_url(string), inventory_quantity(integer), is_discontinued(boolean), is_valid_url(boolean), name(string), options(array), price(integer), sku(string), updated_at(string), url(string), yotpo_id(integer)
- collections:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(string), external_id(string), id(integer), name(string), updated_at(string), yotpo_id(integer)
- customers:
  - primary key: external_id
  - cursor: account_updated_at
  - fields: accepts_email_marketing(boolean), accepts_sms_marketing(boolean), account_created_at(string), account_status(string), account_updated_at(string), address(object), custom_properties(object), default_currency(string), default_language(string), email(string), external_id(string), first_name(string), gender(string), last_name(string), phone_number(string), tags(string)
- orders:
  - primary key: yotpo_id
  - cursor: order_date
  - fields: billing_address(object), cancellation(object), checkout_token(string), currency(string), custom_properties(object), customer(object), customer_locale(string), external_id(string), fulfillments(array), landing_site_url(string), line_items(array), order_date(string), order_name(string), order_number(string), payment_method(string), payment_status(string), shipping_address(object), status(string), subtotal_price(integer), total_price(integer), yotpo_id(integer)
- webhook_targets:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(string), id(integer), updated_at(string), url(string), yotpo_id(integer)
- webhook_filters:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(string), events(array), id(integer), updated_at(string), yotpo_id(integer)
- webhook_subscriptions:
  - primary key: id
  - cursor: updated_at
  - fields: active(boolean), created_at(string), filter_id(integer), id(integer), target_id(integer), updated_at(string), yotpo_id(integer)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- create_product:
  - endpoint: POST /core/v3/stores/{{ config.store_id }}/products
  - required fields: product
  - risk: external mutation; creates a new product in the store's catalog; approval required. Body is wrapped under a top-level "product" key (Yotpo Core API v3 convention) — the record itself carries that wrapper, since the engine's write dialect sends record fields verbatim as the JSON body with no nested-wrapper construction primitive (see teamwork/ynab precedent).
- update_product:
  - endpoint: PATCH /core/v3/stores/{{ config.store_id }}/products/{{ record.yotpo_id }}
  - required fields: yotpo_id, product
  - risk: external mutation; updates an existing product's catalog fields; approval required
- create_product_variant:
  - endpoint: POST /core/v3/stores/{{ config.store_id }}/products/{{ record.product_yotpo_id }}/variants
  - required fields: product_yotpo_id, variant
  - risk: external mutation; creates a new variant under an existing product; approval required
- update_product_variant:
  - endpoint: PATCH /core/v3/stores/{{ config.store_id }}/products/{{ record.product_yotpo_id }}/variants/{{ record.yotpo_id }}
  - required fields: product_yotpo_id, yotpo_id, variant
  - risk: external mutation; updates an existing product variant's fields; approval required
- create_order:
  - endpoint: POST /core/v3/stores/{{ config.store_id }}/orders
  - required fields: order
  - risk: external mutation; creates a new order (may trigger Yotpo's automatic review-request email flow for the associated customer); approval required. Not possible to send automatic review-request emails for orders older than six months (Yotpo's own documented constraint).
- update_order:
  - endpoint: PATCH /core/v3/stores/{{ config.store_id }}/orders/{{ record.yotpo_id }}
  - required fields: yotpo_id, order
  - risk: external mutation; updates an existing order's status/pricing/cancellation fields; approval required
- create_customer:
  - endpoint: POST /core/v3/stores/{{ config.store_id }}/customers
  - required fields: customer
  - risk: external mutation; creates or updates (upsert-by-external_id) a customer profile; approval required. Yotpo's own endpoint is documented as create-or-update, keyed on external_id — there is no separate update_customer action since the same request both creates and upserts.
- create_order_fulfillment:
  - endpoint: POST /core/v3/stores/{{ config.store_id }}/orders/{{ record.order_yotpo_id }}/fulfillments
  - required fields: order_yotpo_id, fulfillment
  - risk: external mutation; records a shipment/fulfillment event against an existing order; approval required
- update_order_fulfillment:
  - endpoint: PATCH /core/v3/stores/{{ config.store_id }}/orders/{{ record.order_yotpo_id }}/fulfillments/{{ record.yotpo_id }}
  - required fields: order_yotpo_id, yotpo_id, fulfillment
  - risk: external mutation; updates the shipment status/tracking of an existing order fulfillment; approval required
- create_collection:
  - endpoint: POST /core/v3/stores/{{ config.store_id }}/collections
  - required fields: collection
  - risk: external mutation; creates a new product collection; approval required
- update_collection:
  - endpoint: PATCH /core/v3/stores/{{ config.store_id }}/collections/{{ record.yotpo_id }}
  - required fields: yotpo_id, collection
  - risk: external mutation; renames an existing product collection; approval required
- add_product_to_collection:
  - endpoint: POST /core/v3/stores/{{ config.store_id }}/collections/{{ record.collection_yotpo_id }}/products
  - required fields: collection_yotpo_id, product_id
  - risk: external mutation; adds a product to an existing collection; approval required
- remove_product_from_collection:
  - endpoint: DELETE /core/v3/stores/{{ config.store_id }}/collections/{{ record.collection_yotpo_id }}/products
  - required fields: collection_yotpo_id, product_id
  - risk: irreversible external mutation; removes a product from an existing collection; approval required
- create_webhook_target:
  - endpoint: POST /core/v3/stores/{{ config.store_id }}/webhooks/targets
  - required fields: webhook_target
  - risk: external mutation; registers a webhook callback URL target; approval required
- update_webhook_target:
  - endpoint: PATCH /core/v3/stores/{{ config.store_id }}/webhooks/targets/{{ record.yotpo_id }}
  - required fields: yotpo_id, webhook_target
  - risk: external mutation; changes an existing webhook target's callback URL; approval required
- delete_webhook_target:
  - endpoint: DELETE /core/v3/stores/{{ config.store_id }}/webhooks/targets/{{ record.yotpo_id }}
  - required fields: yotpo_id
  - risk: irreversible external deletion; removes a registered webhook target (any subscription still referencing it becomes inactive); approval required
- create_webhook_filter:
  - endpoint: POST /core/v3/stores/{{ config.store_id }}/webhooks/filters
  - required fields: webhook_filter
  - risk: external mutation; creates a webhook event filter (an event type cannot be used twice across filters, per Yotpo's own constraint); approval required
- update_webhook_filter:
  - endpoint: PATCH /core/v3/stores/{{ config.store_id }}/webhooks/filters/{{ record.yotpo_id }}
  - required fields: yotpo_id, webhook_filter
  - risk: external mutation; changes an existing webhook filter's subscribed event types; approval required
- delete_webhook_filter:
  - endpoint: DELETE /core/v3/stores/{{ config.store_id }}/webhooks/filters/{{ record.yotpo_id }}
  - required fields: yotpo_id
  - risk: irreversible external deletion; removes a webhook filter (only unused filters can be deleted, per Yotpo's own constraint); approval required
- create_webhook_subscription:
  - endpoint: POST /core/v3/stores/{{ config.store_id }}/webhooks/subscriptions
  - required fields: webhook_subscription
  - risk: external mutation; activates webhook event delivery by combining an existing target and filter; approval required
- update_webhook_subscription:
  - endpoint: PATCH /core/v3/stores/{{ config.store_id }}/webhooks/subscriptions/{{ record.yotpo_id }}
  - required fields: yotpo_id, webhook_subscription
  - risk: external mutation; retargets or (de)activates an existing webhook subscription; approval required
- delete_webhook_subscription:
  - endpoint: DELETE /core/v3/stores/{{ config.store_id }}/webhooks/subscriptions/{{ record.yotpo_id }}
  - required fields: yotpo_id
  - risk: irreversible external deletion; stops webhook event delivery for an existing target/filter combination; approval required

## Security

- read risk: external Yotpo Core API read of product, variant, collection, customer, order, and webhook configuration data
- write risk: external mutation: creates/updates products, variants, orders, customers, order fulfillments, and collections; manages collection membership and webhook target/filter/subscription lifecycle
- approval: required for all write actions; reads require none
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect yotpo
```

### Inspect as structured JSON

```bash
pm connectors inspect yotpo --json
```

## Agent Rules

- Run pm connectors inspect yotpo before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
