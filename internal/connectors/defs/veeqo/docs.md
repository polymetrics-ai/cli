# Overview

Reads orders, products, customers, warehouses, suppliers, purchase orders, sales channels, delivery
methods, and tags from the Veeqo API, and writes orders, products, customers, suppliers, warehouses,
delivery methods, tags, sales channels, product properties, payments, and shipments.

Readable streams: `orders`, `products`, `customers`, `warehouses`, `suppliers`, `purchase_orders`,
`channels`, `delivery_methods`, `tags`.

Write actions: `create_supplier`, `update_supplier`, `delete_supplier`, `create_warehouse`,
`update_warehouse`, `create_delivery_method`, `update_delivery_method`, `delete_delivery_method`,
`create_tag`, `delete_tag`, `create_channel`, `update_channel`, `delete_channel`,
`create_product_property`, `create_customer`, `update_customer`, `create_product`, `update_product`,
`delete_product`, `create_order`, `update_order`, `cancel_order`, `create_payment`,
`create_shipment`.

Service API documentation: https://developers.veeqo.com/.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Veeqo API key, sent as the x-api-key header. Never logged.
- `base_url` (optional, string); default `https://api.veeqo.com`; format `uri`; Veeqo API base URL
  override for tests or proxies.
- `start_date` (optional, string); format `date-time`.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.veeqo.com`.

Authentication behavior:

- API key authentication in `x-api-key` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/orders`.

## Streams notes

Default pagination: single request; no pagination.

Pagination by stream: none: `orders`, `tags`; page_number: `products`, `customers`, `warehouses`,
`suppliers`, `purchase_orders`, `channels`, `delivery_methods`.

- `orders`: GET `/orders` - records path `.`; query `start_date` from template `{{ config.start_date
  }}`, omitted when absent; computed output fields `created_at`, `id`, `number`, `status`.
- `products`: GET `/products` - records path `.`; page-number pagination; page parameter `page`;
  size parameter `page_size`; starts at 1; page size 100; computed output fields `id`.
- `customers`: GET `/customers` - records path `.`; page-number pagination; page parameter `page`;
  size parameter `page_size`; starts at 1; page size 100; computed output fields `id`.
- `warehouses`: GET `/warehouses` - records path `.`; page-number pagination; page parameter `page`;
  size parameter `page_size`; starts at 1; page size 100; computed output fields `id`.
- `suppliers`: GET `/suppliers` - records path `.`; page-number pagination; page parameter `page`;
  size parameter `page_size`; starts at 1; page size 100; computed output fields `id`.
- `purchase_orders`: GET `/purchase_orders` - records path `.`; page-number pagination; page
  parameter `page`; size parameter `page_size`; starts at 1; page size 100; computed output fields
  `id`.
- `channels`: GET `/channels` - records path `.`; page-number pagination; page parameter `page`;
  size parameter `page_size`; starts at 1; page size 100; computed output fields `id`.
- `delivery_methods`: GET `/delivery_methods` - records path `.`; page-number pagination; page
  parameter `page`; size parameter `page_size`; starts at 1; page size 100; computed output fields
  `id`.
- `tags`: GET `/tags` - records path `.`; computed output fields `id`.

## Write actions & risks

Overall write risk: external mutation of Veeqo orders, products, customers, suppliers, warehouses,
delivery methods, tags, sales channels, product properties, payments, and shipments; approval
required.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_supplier`: POST `/suppliers` - kind `create`; body type `json`; required record fields
  `name`; accepted fields `address_line_1`, `city`, `country`, `currency_code`, `name`, `post_code`;
  risk: external mutation; approval required.
- `update_supplier`: PUT `/suppliers/{{ record.id }}` - kind `update`; body type `json`; path fields
  `id`; required record fields `id`; accepted fields `id`, `name`; risk: external mutation; approval
  required.
- `delete_supplier`: DELETE `/suppliers/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `404`; confirmation `destructive`; risk: destructive external mutation; approval
  required.
- `create_warehouse`: POST `/warehouses` - kind `create`; body type `json`; required record fields
  `name`; accepted fields `address_line_1`, `city`, `country`, `name`, `post_code`; risk: external
  mutation; approval required.
- `update_warehouse`: PUT `/warehouses/{{ record.id }}` - kind `update`; body type `json`; path
  fields `id`; required record fields `id`; accepted fields `id`, `name`; risk: external mutation;
  approval required.
- `create_delivery_method`: POST `/delivery_methods` - kind `create`; body type `json`; required
  record fields `name`; accepted fields `cost`, `name`; risk: external mutation; approval required.
- `update_delivery_method`: PUT `/delivery_methods/{{ record.id }}` - kind `update`; body type
  `json`; path fields `id`; required record fields `id`; accepted fields `cost`, `id`, `name`; risk:
  external mutation; approval required.
- `delete_delivery_method`: DELETE `/delivery_methods/{{ record.id }}` - kind `delete`; body type
  `none`; path fields `id`; required record fields `id`; accepted fields `id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: destructive external
  mutation; approval required.
- `create_tag`: POST `/tags` - kind `create`; body type `json`; required record fields `name`;
  accepted fields `colour`, `name`; risk: external mutation; approval required.
- `delete_tag`: DELETE `/tags/{{ record.id }}` - kind `delete`; body type `none`; path fields `id`;
  required record fields `id`; accepted fields `id`; missing records treated as success for status
  `404`; confirmation `destructive`; risk: destructive external mutation; approval required.
- `create_channel`: POST `/channels` - kind `create`; body type `json`; required record fields
  `name`, `type_code`; accepted fields `currency_code`, `email`, `name`, `short_name`, `type_code`;
  risk: external mutation; approval required.
- `update_channel`: PUT `/channels/{{ record.id }}` - kind `update`; body type `json`; path fields
  `id`; required record fields `id`; accepted fields `id`, `name`; risk: external mutation; approval
  required.
- `delete_channel`: DELETE `/channels/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `404`; confirmation `destructive`; risk: destructive external mutation; approval
  required.
- `create_product_property`: POST `/product_properties` - kind `create`; body type `json`; required
  record fields `name`; accepted fields `name`; risk: external mutation; approval required.
- `create_customer`: POST `/customers` - kind `create`; body type `json`; required record fields
  `customer`; accepted fields `customer`; risk: external mutation; approval required.
- `update_customer`: PUT `/customers/{{ record.id }}` - kind `update`; body type `json`; path fields
  `id`; required record fields `id`, `customer`; accepted fields `customer`, `id`; risk: external
  mutation; approval required.
- `create_product`: POST `/products` - kind `create`; body type `json`; required record fields
  `product`; accepted fields `product`; risk: external mutation; approval required.
- `update_product`: PUT `/products/{{ record.id }}` - kind `update`; body type `json`; path fields
  `id`; required record fields `id`, `product`; accepted fields `id`, `product`; risk: external
  mutation; approval required.
- `delete_product`: DELETE `/products/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `404`; confirmation `destructive`; risk: destructive external mutation; approval
  required.
- `create_order`: POST `/orders` - kind `create`; body type `json`; required record fields `order`;
  accepted fields `order`; risk: external mutation; approval required.
- `update_order`: PUT `/orders/{{ record.id }}` - kind `update`; body type `json`; path fields `id`;
  required record fields `id`, `order`; accepted fields `id`, `order`; risk: external mutation;
  approval required.
- `cancel_order`: PUT `/orders/{{ record.id }}/cancel` - kind `custom`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; risk: external mutation (cancels
  an order); approval required.
- `create_payment`: POST `/payments` - kind `create`; body type `json`; required record fields
  `amount`, `payment_attributes`; accepted fields `amount`, `payment_attributes`; risk: external
  mutation; approval required.
- `create_shipment`: POST `/shipments` - kind `create`; body type `json`; required record fields
  `carrier_id`, `notify_customer`, `update_remote_order`, `allocation_id`, `order_id`; accepted
  fields `allocation_id`, `carrier_id`, `notify_customer`, `order_id`, `shipment`,
  `update_remote_order`; risk: external mutation; approval required.

## Known limits

- API coverage includes 9 stream-backed endpoint group(s), 24 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=1, duplicate_of=10, non_data_endpoint=2, out_of_scope=22,
  requires_elevated_scope=2.
