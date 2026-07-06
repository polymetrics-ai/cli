# Overview

Reads and writes Printify shops, catalog resources, products, orders, uploads, and webhooks through
the Printify public API.

Readable streams: `shops`, `products`, `orders`, `blueprints`, `print_providers`,
`blueprint_detail`, `blueprint_print_providers`, `blueprint_variants`, `shipping_profiles`,
`print_provider_detail`, `product_detail`, `product_gpsr`, `order_detail`, `uploads`,
`upload_detail`, `webhooks`, `v2_shipping_methods`, `v2_shipping_standard`, `v2_shipping_priority`,
`v2_shipping_express`, `v2_shipping_economy`.

Write actions: `disconnect_shop`, `create_product`, `update_product`, `delete_product`,
`publish_product`, `mark_product_publishing_succeeded`, `mark_product_publishing_failed`,
`unpublish_product`, `submit_order`, `submit_express_order`, `send_order_to_production`,
`calculate_order_shipping`, `cancel_order`, `upload_image`, `archive_uploaded_image`,
`create_webhook`, `update_webhook`, `delete_webhook`, `simulate_webhook`.

Service API documentation: https://developers.printify.com/.

## Auth setup

Connection fields:

- `api_token` (required, secret, string); Printify personal access token, sent as Authorization:
  Bearer <api_token>. Never logged.
- `base_url` (optional, string); default `https://api.printify.com`; format `uri`; Printify API root
  URL override for tests or proxies. Paths include /v1 or /v2.
- `blueprint_id` (optional, string); Catalog blueprint ID for blueprint detail, print-provider,
  variant, and shipping streams.
- `image_id` (optional, string); Uploaded image ID for upload detail streams.
- `order_id` (optional, string); Order ID for order detail streams.
- `order_sku` (optional, string); Optional order list SKU filter.
- `order_status` (optional, string); Optional order list status filter.
- `print_provider_id` (optional, string); Print provider ID for provider detail and
  blueprint/provider subresource streams.
- `product_id` (optional, string); Product ID for product detail streams.
- `shop_id` (optional, string); Printify shop ID for shop-scoped streams and writes.
- `show_out_of_stock` (optional, string); Optional catalog variants out-of-stock flag; use the API
  documented value 0 or 1.
- `webhook_id` (optional, string); Webhook ID for webhook write actions.

Secret fields are redacted in logs and write previews: `api_token`.

Default configuration values: `base_url=https://api.printify.com`.

Authentication behavior:

- Bearer token authentication using `secrets.api_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/v1/shops.json`.

## Streams notes

Default pagination: single request; no pagination.

Pagination by stream: next_url: `products`, `orders`, `uploads`; none: `shops`, `blueprints`,
`print_providers`, `blueprint_detail`, `blueprint_print_providers`, `blueprint_variants`,
`shipping_profiles`, `print_provider_detail`, `product_detail`, `product_gpsr`, `order_detail`,
`upload_detail`, `webhooks`, `v2_shipping_methods`, `v2_shipping_standard`, `v2_shipping_priority`,
`v2_shipping_express`, `v2_shipping_economy`.

- `shops`: GET `/v1/shops.json` - records path `.`.
- `products`: GET `/v1/shops/{{ config.shop_id }}/products.json` - records path `data`; query
  `limit`=`100`; follows a next-page URL from the response body; URL path `next_page_url`; next URLs
  stay on the configured API host.
- `orders`: GET `/v1/shops/{{ config.shop_id }}/orders.json` - records path `data`; query
  `limit`=`100`; `sku` from template `{{ config.order_sku }}`, omitted when absent; `status` from
  template `{{ config.order_status }}`, omitted when absent; follows a next-page URL from the
  response body; URL path `next_page_url`; next URLs stay on the configured API host.
- `blueprints`: GET `/v1/catalog/blueprints.json` - records path `.`.
- `print_providers`: GET `/v1/catalog/print_providers.json` - records path `.`.
- `blueprint_detail`: GET `/v1/catalog/blueprints/{{ config.blueprint_id }}.json` - records path
  `.`.
- `blueprint_print_providers`: GET `/v1/catalog/blueprints/{{ config.blueprint_id
  }}/print_providers.json` - records path `.`.
- `blueprint_variants`: GET `/v1/catalog/blueprints/{{ config.blueprint_id }}/print_providers/{{
  config.print_provider_id }}/variants.json` - records path `variants`; query `show_out_of_stock`
  from template `{{ config.show_out_of_stock }}`, omitted when absent; computed output fields
  `blueprint_id`, `print_provider_id`.
- `shipping_profiles`: GET `/v1/catalog/blueprints/{{ config.blueprint_id }}/print_providers/{{
  config.print_provider_id }}/shipping.json` - records path `profiles`; computed output fields
  `blueprint_id`, `print_provider_id`.
- `print_provider_detail`: GET `/v1/catalog/print_providers/{{ config.print_provider_id }}.json` -
  records path `.`.
- `product_detail`: GET `/v1/shops/{{ config.shop_id }}/products/{{ config.product_id }}.json` -
  records path `.`.
- `product_gpsr`: GET `/v1/shops/{{ config.shop_id }}/products/{{ config.product_id }}/gpsr.json` -
  records path `.`.
- `order_detail`: GET `/v1/shops/{{ config.shop_id }}/orders/{{ config.order_id }}.json` - records
  path `.`.
- `uploads`: GET `/v1/uploads.json` - records path `data`; query `limit`=`100`; follows a next-page
  URL from the response body; URL path `next_page_url`; next URLs stay on the configured API host.
- `upload_detail`: GET `/v1/uploads/{{ config.image_id }}.json` - records path `.`.
- `webhooks`: GET `/v1/shops/{{ config.shop_id }}/webhooks.json` - records path `.`.
- `v2_shipping_methods`: GET `/v2/catalog/blueprints/{{ config.blueprint_id }}/print_providers/{{
  config.print_provider_id }}/shipping.json` - records path `data`; computed output fields
  `blueprint_id`, `print_provider_id`.
- `v2_shipping_standard`: GET `/v2/catalog/blueprints/{{ config.blueprint_id }}/print_providers/{{
  config.print_provider_id }}/shipping/standard.json` - records path `data`; computed output fields
  `blueprint_id`, `print_provider_id`, `shipping_method`.
- `v2_shipping_priority`: GET `/v2/catalog/blueprints/{{ config.blueprint_id }}/print_providers/{{
  config.print_provider_id }}/shipping/priority.json` - records path `data`; computed output fields
  `blueprint_id`, `print_provider_id`, `shipping_method`.
- `v2_shipping_express`: GET `/v2/catalog/blueprints/{{ config.blueprint_id }}/print_providers/{{
  config.print_provider_id }}/shipping/express.json` - records path `data`; computed output fields
  `blueprint_id`, `print_provider_id`, `shipping_method`.
- `v2_shipping_economy`: GET `/v2/catalog/blueprints/{{ config.blueprint_id }}/print_providers/{{
  config.print_provider_id }}/shipping/economy.json` - records path `data`; computed output fields
  `blueprint_id`, `print_provider_id`, `shipping_method`.

## Write actions & risks

Overall write risk: creates, updates, publishes, unpublishes, deletes, archives, disconnects,
submits, cancels, and simulates Printify resources depending on the selected write action.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `disconnect_shop`: DELETE `/v1/shops/{{ config.shop_id }}/connection.json` - kind `delete`; body
  type `none`; confirmation `destructive`; risk: disconnects the configured shop from the Printify
  account.
- `create_product`: POST `/v1/shops/{{ config.shop_id }}/products.json` - kind `create`; body type
  `json`; required record fields `title`, `blueprint_id`, `print_provider_id`; accepted fields
  `blueprint_id`, `description`, `print_areas`, `print_provider_id`, `tags`, `title`, `variants`;
  risk: creates a product in the configured shop.
- `update_product`: PUT `/v1/shops/{{ config.shop_id }}/products/{{ record.product_id }}.json` -
  kind `update`; body type `json`; path fields `product_id`; required record fields `product_id`;
  accepted fields `description`, `print_areas`, `product_id`, `tags`, `title`, `variants`,
  `visible`; risk: updates an existing product in the configured shop.
- `delete_product`: DELETE `/v1/shops/{{ config.shop_id }}/products/{{ record.product_id }}.json` -
  kind `delete`; body type `none`; path fields `product_id`; required record fields `product_id`;
  accepted fields `product_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: deletes a product from the configured shop.
- `publish_product`: POST `/v1/shops/{{ config.shop_id }}/products/{{ record.product_id
  }}/publish.json` - kind `custom`; body type `json`; path fields `product_id`; required record
  fields `product_id`; accepted fields `description`, `images`, `keyFeatures`, `product_id`,
  `shipping_template`, `tags`, `title`, `variants`; risk: publishes a product to the connected sales
  channel.
- `mark_product_publishing_succeeded`: POST `/v1/shops/{{ config.shop_id }}/products/{{
  record.product_id }}/publishing_succeeded.json` - kind `custom`; body type `json`; path fields
  `product_id`; required record fields `product_id`, `external`; accepted fields `external`,
  `product_id`; risk: marks product publishing as succeeded and stores an external handle.
- `mark_product_publishing_failed`: POST `/v1/shops/{{ config.shop_id }}/products/{{
  record.product_id }}/publishing_failed.json` - kind `custom`; body type `json`; path fields
  `product_id`; required record fields `product_id`, `reason`; accepted fields `product_id`,
  `reason`; risk: marks product publishing as failed.
- `unpublish_product`: POST `/v1/shops/{{ config.shop_id }}/products/{{ record.product_id
  }}/unpublish.json` - kind `custom`; body type `none`; path fields `product_id`; required record
  fields `product_id`; accepted fields `product_id`; confirmation `destructive`; risk: notifies
  Printify that a product has been unpublished.
- `submit_order`: POST `/v1/shops/{{ config.shop_id }}/orders.json` - kind `create`; body type
  `json`; required record fields `line_items`, `address_to`; accepted fields `address_to`,
  `external_id`, `label`, `line_items`, `send_shipping_notification`, `shipping_method`; risk:
  submits an order to Printify.
- `submit_express_order`: POST `/v1/shops/{{ config.shop_id }}/orders/express.json` - kind `create`;
  body type `json`; required record fields `line_items`, `address_to`; accepted fields `address_to`,
  `external_id`, `line_items`, `shipping_method`; risk: submits a Printify Express order.
- `send_order_to_production`: POST `/v1/shops/{{ config.shop_id }}/orders/{{ record.order_id
  }}/send_to_production.json` - kind `custom`; body type `none`; path fields `order_id`; required
  record fields `order_id`; accepted fields `order_id`; confirmation `destructive`; risk: sends an
  existing order to production.
- `calculate_order_shipping`: POST `/v1/shops/{{ config.shop_id }}/orders/shipping.json` - kind
  `custom`; body type `json`; required record fields `line_items`, `address_to`; accepted fields
  `address_to`, `line_items`; risk: calculates shipping costs for a prospective order without
  submitting it.
- `cancel_order`: POST `/v1/shops/{{ config.shop_id }}/orders/{{ record.order_id }}/cancel.json` -
  kind `custom`; body type `none`; path fields `order_id`; required record fields `order_id`;
  accepted fields `order_id`; confirmation `destructive`; risk: cancels an unpaid order.
- `upload_image`: POST `/v1/uploads/images.json` - kind `create`; body type `json`; required record
  fields `file_name`; accepted fields `contents`, `file_name`, `url`; risk: uploads an image into
  the Printify media library.
- `archive_uploaded_image`: POST `/v1/uploads/{{ record.image_id }}/archive.json` - kind `custom`;
  body type `none`; path fields `image_id`; required record fields `image_id`; accepted fields
  `image_id`; confirmation `destructive`; risk: archives an uploaded image.
- `create_webhook`: POST `/v1/shops/{{ config.shop_id }}/webhooks.json` - kind `create`; body type
  `json`; required record fields `topic`, `url`; accepted fields `topic`, `url`; risk: creates a
  webhook subscription for the configured shop.
- `update_webhook`: PUT `/v1/shops/{{ config.shop_id }}/webhooks/{{ record.webhook_id }}.json` -
  kind `update`; body type `json`; path fields `webhook_id`; required record fields `webhook_id`;
  accepted fields `topic`, `url`, `webhook_id`; risk: updates an existing webhook subscription.
- `delete_webhook`: DELETE `/v1/shops/{{ config.shop_id }}/webhooks/{{ record.webhook_id
  }}.json?host={{ record.host }}` - kind `delete`; body type `none`; path fields `webhook_id`,
  `host`; required record fields `webhook_id`, `host`; accepted fields `host`, `webhook_id`; missing
  records treated as success for status `404`; confirmation `destructive`; risk: deletes a webhook
  subscription after host safeguard matching.
- `simulate_webhook`: POST `/v1/shops/{{ config.shop_id }}/webhooks/{{ record.webhook_id
  }}/simulate` - kind `custom`; body type `json`; path fields `webhook_id`; required record fields
  `webhook_id`; accepted fields `anything`, `resource`, `webhook_id`; risk: sends a webhook
  simulation event for testing.

## Known limits

- Batch defaults: read_page_size=100, write_batch_size=1.
- API coverage includes 21 stream-backed endpoint group(s), 19 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  duplicate_of=1, non_data_endpoint=3.
