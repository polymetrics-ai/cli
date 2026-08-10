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
- api_token (secret)

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

## Command Surface

- Run Printify's declared streams and reverse-ETL actions.
- Usage: pm printify <command> [flags]
- Read streams
- Reverse ETL writes
- Other Commands
  - api get app oauth accept - Documented GET /app/oauth/accept (not implemented) [intent=direct_read availability=not_implemented operation=printify.get.app-oauth-accept]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api post app oauth tokens - Documented POST /app/oauth/tokens (not implemented) [intent=direct_write availability=not_implemented operation=printify.post.app-oauth-tokens]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post app oauth tokens refresh - Documented POST /app/oauth/tokens/refresh (not implemented) [intent=direct_write availability=not_implemented operation=printify.post.app-oauth-tokens-refresh]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post v1 shops shop-id express-json - Documented POST /v1/shops/{shop_id}/express.json (not implemented) [intent=direct_write availability=not_implemented operation=printify.post.v1-shops-shop-id-express-json]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - archive uploaded image apply - Plan and execute the archive uploaded image reverse-ETL action [intent=reverse_etl availability=implemented write=archive_uploaded_image]; approval: requires plan, preview, approval, and execute; risk: archives an uploaded image; flags: --image_id (required)
  - blueprint detail list - Run the blueprint detail ETL stream [intent=etl availability=implemented stream=blueprint_detail]
  - blueprint print providers list - Run the blueprint print providers ETL stream [intent=etl availability=implemented stream=blueprint_print_providers]
  - blueprint variants list - Run the blueprint variants ETL stream [intent=etl availability=implemented stream=blueprint_variants]
  - blueprints list - Run the blueprints ETL stream [intent=etl availability=implemented stream=blueprints]
  - calculate order shipping apply - Plan and execute the calculate order shipping reverse-ETL action [intent=reverse_etl availability=not_implemented write=calculate_order_shipping]; approval: requires plan, preview, approval, and execute; risk: calculates shipping costs for a prospective order without submitting it; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - cancel order apply - Plan and execute the cancel order reverse-ETL action [intent=reverse_etl availability=implemented write=cancel_order]; approval: requires plan, preview, approval, and execute; risk: cancels an unpaid order; flags: --order_id (required)
  - create product apply - Plan and execute the create product reverse-ETL action [intent=reverse_etl availability=implemented write=create_product]; approval: requires plan, preview, approval, and execute; risk: creates a product in the configured shop; flags: --blueprint_id (required), --print_provider_id (required), --title (required)
  - create webhook apply - Plan and execute the create webhook reverse-ETL action [intent=reverse_etl availability=implemented write=create_webhook]; approval: requires plan, preview, approval, and execute; risk: creates a webhook subscription for the configured shop; flags: --topic (required), --url (required)
  - delete product apply - Plan and execute the delete product reverse-ETL action [intent=reverse_etl availability=implemented write=delete_product]; approval: requires plan, preview, approval, and execute; risk: deletes a product from the configured shop; flags: --product_id (required)
  - delete webhook apply - Plan and execute the delete webhook reverse-ETL action [intent=reverse_etl availability=implemented write=delete_webhook]; approval: requires plan, preview, approval, and execute; risk: deletes a webhook subscription after host safeguard matching; flags: --host (required), --webhook_id (required)
  - disconnect shop apply - Plan and execute the disconnect shop reverse-ETL action [intent=reverse_etl availability=implemented write=disconnect_shop]; approval: requires plan, preview, approval, and execute; risk: disconnects the configured shop from the Printify account
  - mark product publishing failed apply - Plan and execute the mark product publishing failed reverse-ETL action [intent=reverse_etl availability=implemented write=mark_product_publishing_failed]; approval: requires plan, preview, approval, and execute; risk: marks product publishing as failed; flags: --product_id (required), --reason (required)
  - mark product publishing succeeded apply - Plan and execute the mark product publishing succeeded reverse-ETL action [intent=reverse_etl availability=not_implemented write=mark_product_publishing_succeeded]; approval: requires plan, preview, approval, and execute; risk: marks product publishing as succeeded and stores an external handle; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - order detail list - Run the order detail ETL stream [intent=etl availability=implemented stream=order_detail]
  - orders list - Run the orders ETL stream [intent=etl availability=implemented stream=orders]
  - print provider detail list - Run the print provider detail ETL stream [intent=etl availability=implemented stream=print_provider_detail]
  - print providers list - Run the print providers ETL stream [intent=etl availability=implemented stream=print_providers]
  - product detail list - Run the product detail ETL stream [intent=etl availability=implemented stream=product_detail]
  - product gpsr list - Run the product gpsr ETL stream [intent=etl availability=implemented stream=product_gpsr]
  - products list - Run the products ETL stream [intent=etl availability=implemented stream=products]
  - publish product apply - Plan and execute the publish product reverse-ETL action [intent=reverse_etl availability=implemented write=publish_product]; approval: requires plan, preview, approval, and execute; risk: publishes a product to the connected sales channel; flags: --product_id (required)
  - send order to production apply - Plan and execute the send order to production reverse-ETL action [intent=reverse_etl availability=implemented write=send_order_to_production]; approval: requires plan, preview, approval, and execute; risk: sends an existing order to production; flags: --order_id (required)
  - shipping profiles list - Run the shipping profiles ETL stream [intent=etl availability=implemented stream=shipping_profiles]
  - shops list - Run the shops ETL stream [intent=etl availability=implemented stream=shops]
  - simulate webhook apply - Plan and execute the simulate webhook reverse-ETL action [intent=reverse_etl availability=implemented write=simulate_webhook]; approval: requires plan, preview, approval, and execute; risk: sends a webhook simulation event for testing; flags: --webhook_id (required)
  - submit express order apply - Plan and execute the submit express order reverse-ETL action [intent=reverse_etl availability=not_implemented write=submit_express_order]; approval: requires plan, preview, approval, and execute; risk: submits a Printify Express order; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - submit order apply - Plan and execute the submit order reverse-ETL action [intent=reverse_etl availability=not_implemented write=submit_order]; approval: requires plan, preview, approval, and execute; risk: submits an order to Printify; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - unpublish product apply - Plan and execute the unpublish product reverse-ETL action [intent=reverse_etl availability=implemented write=unpublish_product]; approval: requires plan, preview, approval, and execute; risk: notifies Printify that a product has been unpublished; flags: --product_id (required)
  - update product apply - Plan and execute the update product reverse-ETL action [intent=reverse_etl availability=implemented write=update_product]; approval: requires plan, preview, approval, and execute; risk: updates an existing product in the configured shop; flags: --product_id (required)
  - update webhook apply - Plan and execute the update webhook reverse-ETL action [intent=reverse_etl availability=implemented write=update_webhook]; approval: requires plan, preview, approval, and execute; risk: updates an existing webhook subscription; flags: --webhook_id (required)
  - upload detail list - Run the upload detail ETL stream [intent=etl availability=implemented stream=upload_detail]
  - upload image apply - Plan and execute the upload image reverse-ETL action [intent=reverse_etl availability=implemented write=upload_image]; approval: requires plan, preview, approval, and execute; risk: uploads an image into the Printify media library; flags: --file_name (required)
  - uploads list - Run the uploads ETL stream [intent=etl availability=implemented stream=uploads]
  - v2 shipping economy list - Run the v2 shipping economy ETL stream [intent=etl availability=implemented stream=v2_shipping_economy]
  - v2 shipping express list - Run the v2 shipping express ETL stream [intent=etl availability=implemented stream=v2_shipping_express]
  - v2 shipping methods list - Run the v2 shipping methods ETL stream [intent=etl availability=implemented stream=v2_shipping_methods]
  - v2 shipping priority list - Run the v2 shipping priority ETL stream [intent=etl availability=implemented stream=v2_shipping_priority]
  - v2 shipping standard list - Run the v2 shipping standard ETL stream [intent=etl availability=implemented stream=v2_shipping_standard]
  - webhooks list - Run the webhooks ETL stream [intent=etl availability=implemented stream=webhooks]

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
