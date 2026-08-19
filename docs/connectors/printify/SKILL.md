---
name: pm-printify
description: Printify connector knowledge and safe action guide.
---

# pm-printify

## Purpose

Reads and writes Printify shops, catalog resources, products, orders, uploads, and webhooks through the Printify public API.

## Icon

- id: printify
- asset: icons/printify.svg
- source: official
- review_status: official_verified
- review_url: https://developers.printify.com/

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- blueprint_id
- image_id
- order_id
- order_sku
- order_status
- print_provider_id
- product_id
- shop_id
- show_out_of_stock
- webhook_id
- api_token (secret) (required)

## ETL Streams

- shops:
  - primary key: id
  - fields: id(integer), sales_channel(string), title(string)
- products:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(string), id(string), sales_channel(string), status(string), title(string), updated_at(string), visible(boolean)
- orders:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(string), id(string), sales_channel(string), status(string), title(string), updated_at(string), visible(boolean)
- blueprints:
  - primary key: id
  - fields: id(integer), title(string)
- print_providers:
  - primary key: id
  - fields: id(integer), title(string)
- blueprint_detail:
  - primary key: id
  - fields: brand(string), description(string), id(integer), images(array), model(string), title(string)
- blueprint_print_providers:
  - primary key: id
  - fields: decoration_methods(array), id(integer), title(string)
- blueprint_variants:
  - primary key: id
  - fields: blueprint_id(string), decoration_methods(array), id(integer), is_available(boolean), options(object), placeholders(array), print_provider_id(string), title(string)
- shipping_profiles:
  - primary key: variant_ids
  - fields: additional_items(object), blueprint_id(string), countries(array), first_item(object), handling_time(object), print_provider_id(string), variant_ids(array)
- print_provider_detail:
  - primary key: id
  - fields: blueprints(array), id(integer), location(object), title(string)
- product_detail:
  - primary key: id
  - cursor: updated_at
  - fields: blueprint_id(integer), created_at(string), description(string), external(object), id(string), images(array), is_locked(boolean), options(array), print_areas(array), print_provider_id(integer), sales_channel(string), shop_id(integer), tags(array), title(string), updated_at(string), user_id(integer), variants(array), visible(boolean)
- product_gpsr:
  - primary key: title
  - fields: text(string), title(string)
- order_detail:
  - primary key: id
  - cursor: updated_at
  - fields: address_to(object), app_order_id(string), created_at(string), id(string), line_items(array), metadata(object), shipping_method(integer), status(string), updated_at(string)
- uploads:
  - primary key: id
  - cursor: upload_time
  - fields: file_name(string), height(integer), id(string), mime_type(string), preview_url(string), size(integer), upload_time(string), width(integer)
- upload_detail:
  - primary key: id
  - cursor: upload_time
  - fields: file_name(string), height(integer), id(string), mime_type(string), preview_url(string), size(integer), upload_time(string), width(integer)
- webhooks:
  - primary key: id
  - fields: id(string), shop_id(string), topic(string), url(string)
- v2_shipping_methods:
  - primary key: id
  - fields: attributes(object), blueprint_id(string), id(string), print_provider_id(string), type(string)
- v2_shipping_standard:
  - primary key: id
  - fields: attributes(object), blueprint_id(string), id(string), print_provider_id(string), shipping_method(string), type(string)
- v2_shipping_priority:
  - primary key: id
  - fields: attributes(object), blueprint_id(string), id(string), print_provider_id(string), shipping_method(string), type(string)
- v2_shipping_express:
  - primary key: id
  - fields: attributes(object), blueprint_id(string), id(string), print_provider_id(string), shipping_method(string), type(string)
- v2_shipping_economy:
  - primary key: id
  - fields: attributes(object), blueprint_id(string), id(string), print_provider_id(string), shipping_method(string), type(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- disconnect_shop:
  - endpoint: DELETE /v1/shops/{{ config.shop_id }}/connection.json
  - risk: disconnects the configured shop from the Printify account
- create_product:
  - endpoint: POST /v1/shops/{{ config.shop_id }}/products.json
  - required fields: title, blueprint_id, print_provider_id
  - risk: creates a product in the configured shop
- update_product:
  - endpoint: PUT /v1/shops/{{ config.shop_id }}/products/{{ record.product_id }}.json
  - required fields: product_id
  - risk: updates an existing product in the configured shop
- delete_product:
  - endpoint: DELETE /v1/shops/{{ config.shop_id }}/products/{{ record.product_id }}.json
  - required fields: product_id
  - risk: deletes a product from the configured shop
- publish_product:
  - endpoint: POST /v1/shops/{{ config.shop_id }}/products/{{ record.product_id }}/publish.json
  - required fields: product_id
  - risk: publishes a product to the connected sales channel
- mark_product_publishing_succeeded:
  - endpoint: POST /v1/shops/{{ config.shop_id }}/products/{{ record.product_id }}/publishing_succeeded.json
  - required fields: product_id, external
  - risk: marks product publishing as succeeded and stores an external handle
- mark_product_publishing_failed:
  - endpoint: POST /v1/shops/{{ config.shop_id }}/products/{{ record.product_id }}/publishing_failed.json
  - required fields: product_id, reason
  - risk: marks product publishing as failed
- unpublish_product:
  - endpoint: POST /v1/shops/{{ config.shop_id }}/products/{{ record.product_id }}/unpublish.json
  - required fields: product_id
  - risk: notifies Printify that a product has been unpublished
- submit_order:
  - endpoint: POST /v1/shops/{{ config.shop_id }}/orders.json
  - required fields: line_items, address_to
  - risk: submits an order to Printify
- submit_express_order:
  - endpoint: POST /v1/shops/{{ config.shop_id }}/orders/express.json
  - required fields: line_items, address_to
  - risk: submits a Printify Express order
- send_order_to_production:
  - endpoint: POST /v1/shops/{{ config.shop_id }}/orders/{{ record.order_id }}/send_to_production.json
  - required fields: order_id
  - risk: sends an existing order to production
- calculate_order_shipping:
  - endpoint: POST /v1/shops/{{ config.shop_id }}/orders/shipping.json
  - required fields: line_items, address_to
  - risk: calculates shipping costs for a prospective order without submitting it
- cancel_order:
  - endpoint: POST /v1/shops/{{ config.shop_id }}/orders/{{ record.order_id }}/cancel.json
  - required fields: order_id
  - risk: cancels an unpaid order
- upload_image:
  - endpoint: POST /v1/uploads/images.json
  - required fields: file_name
  - risk: uploads an image into the Printify media library
- archive_uploaded_image:
  - endpoint: POST /v1/uploads/{{ record.image_id }}/archive.json
  - required fields: image_id
  - risk: archives an uploaded image
- create_webhook:
  - endpoint: POST /v1/shops/{{ config.shop_id }}/webhooks.json
  - required fields: topic, url
  - risk: creates a webhook subscription for the configured shop
- update_webhook:
  - endpoint: PUT /v1/shops/{{ config.shop_id }}/webhooks/{{ record.webhook_id }}.json
  - required fields: webhook_id
  - risk: updates an existing webhook subscription
- delete_webhook:
  - endpoint: DELETE /v1/shops/{{ config.shop_id }}/webhooks/{{ record.webhook_id }}.json?host={{ record.host }}
  - required fields: webhook_id, host
  - risk: deletes a webhook subscription after host safeguard matching
- simulate_webhook:
  - endpoint: POST /v1/shops/{{ config.shop_id }}/webhooks/{{ record.webhook_id }}/simulate
  - required fields: webhook_id
  - risk: sends a webhook simulation event for testing

## Security

- read risk: external Printify API read of shop, catalog, product, order, upload, and webhook metadata
- write risk: creates, updates, publishes, unpublishes, deletes, archives, disconnects, submits, cancels, and simulates Printify resources depending on the selected write action
- approval: reverse ETL writes require plan preview and approval token; destructive product/order/shop/upload/webhook actions are marked destructive
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect printify
```

### Inspect as structured JSON

```bash
pm connectors inspect printify --json
```

## Agent Rules

- Run pm connectors inspect printify before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
