# Overview

Reads SendOwl orders, products, subscriptions, discounts, bundles, and licenses, and writes
product/subscription/discount/bundle lifecycle mutations and order actions (refund, cancel
subscription, resend email) through the SendOwl API.

Readable streams: `orders`, `products`, `subscriptions`, `discounts`, `bundles`, `licenses`.

Write actions: `create_product`, `update_product`, `delete_product`, `create_subscription`,
`update_subscription`, `delete_subscription`, `create_discount`, `update_discount`,
`delete_discount`, `update_bundle`, `delete_bundle`, `refund_order`, `cancel_order_subscription`,
`resend_order_email`.

Service API documentation: https://www.sendowl.com/developers/.

## Auth setup

Connection fields:

- `base_url` (optional, string); default `https://www.sendowl.com`; format `uri`; SendOwl API host,
  with no /api/v1 (or /api/v1_2, /api/v1_3) suffix -- each stream/write's own path now carries its
  own API version segment, since SendOwl's real documented surface spans v1
  (products/subscriptions/packages/licenses), v1_2 (discounts), and v1_3 (orders list/show/update).
- `mode` (optional, string).
- `password` (required, secret, string); SendOwl API secret, sent as the HTTP Basic password. Never
  logged.
- `username` (required, string); SendOwl API key, sent as the HTTP Basic username.

Secret fields are redacted in logs and write previews: `password`.

Default configuration values: `base_url=https://www.sendowl.com`.

Authentication behavior:

- HTTP Basic authentication using `config.username`, `secrets.password`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/api/v1/orders`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `per_page`; starts
at 1; page size 100; maximum 1 page(s).

- `orders`: GET `/api/v1/orders` - records path `.`; page-number pagination; page parameter `page`;
  size parameter `per_page`; starts at 1; page size 100; maximum 1 page(s); emits passthrough
  records.
- `products`: GET `/api/v1/products` - records path `.`; page-number pagination; page parameter
  `page`; size parameter `per_page`; starts at 1; page size 100; maximum 1 page(s); emits
  passthrough records.
- `subscriptions`: GET `/api/v1/subscriptions` - records path `.`; page-number pagination; page
  parameter `page`; size parameter `per_page`; starts at 1; page size 100; maximum 1 page(s); emits
  passthrough records.
- `discounts`: GET `/api/v1_2/discounts` - records path `.`; page-number pagination; page parameter
  `page`; size parameter `per_page`; starts at 1; page size 100; maximum 1 page(s); emits
  passthrough records.
- `bundles`: GET `/api/v1/packages` - records path `.`; page-number pagination; page parameter
  `page`; size parameter `per_page`; starts at 1; page size 100; maximum 1 page(s); emits
  passthrough records.
- `licenses`: GET `/api/v1/products/{{ fanout.id }}/licenses` - records path `.`; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 100; maximum
  1 page(s); fan-out; ids from request `/api/v1/products`; id-list records path `.`; id field `id`;
  id inserted into the request path; stamps `product_id`; emits passthrough records.

## Write actions & risks

Overall write risk: external SendOwl API mutation, including a real financial refund action
(refund_order).

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_product`: POST `/api/v1/products` - kind `create`; body type `form`; required record
  fields `name`; accepted fields `currency_code`, `name`, `price`, `product_type`; risk: creates a
  new sellable product (no file attachment; SendOwl's file-upload create path is a separate
  multipart-only endpoint this dialect cannot express, see docs.md Known limits); external mutation,
  approval required.
- `update_product`: PUT `/api/v1/products/{{ record.id }}` - kind `update`; body type `form`; path
  fields `id`; required record fields `id`; accepted fields `currency_code`, `id`, `name`, `price`;
  risk: mutates an existing product's name/price/currency; affects future checkouts of this product.
- `delete_product`: DELETE `/api/v1/products/{{ record.id }}` - kind `delete`; body type `none`;
  path fields `id`; required record fields `id`; accepted fields `id`; missing records treated as
  success for status `404`; risk: permanently removes a product; breaks any existing
  order-fulfillment/download links referencing it.
- `create_subscription`: POST `/api/v1/subscriptions` - kind `create`; body type `form`; required
  record fields `name`; accepted fields `currency_code`, `name`, `period`, `price`; risk: creates a
  new recurring-billing subscription product; external mutation, approval required.
- `update_subscription`: PUT `/api/v1/subscriptions/{{ record.id }}` - kind `update`; body type
  `form`; path fields `id`; required record fields `id`; accepted fields `id`, `name`, `price`;
  risk: mutates an existing subscription's name/price; affects future recurring charges for new
  subscribers to this plan.
- `delete_subscription`: DELETE `/api/v1/subscriptions/{{ record.id }}` - kind `delete`; body type
  `none`; path fields `id`; required record fields `id`; accepted fields `id`; missing records
  treated as success for status `404`; risk: permanently removes a subscription product; does not
  itself cancel any buyer's already-active recurring order (see cancel_order_subscription).
- `create_discount`: POST `/api/v1_2/discounts` - kind `create`; body type `form`; required record
  fields `code`; accepted fields `code`, `discount_flat_rate`, `discount_percentage`, `end_at`,
  `max_uses`; risk: creates a new discount code usable at checkout; external mutation, approval
  required.
- `update_discount`: PUT `/api/v1_2/discounts/{{ record.id }}` - kind `update`; body type `form`;
  path fields `id`; required record fields `id`; accepted fields `discount_percentage`, `end_at`,
  `id`, `max_uses`; risk: mutates an existing discount's percentage/usage cap/expiry; affects all
  buyers subsequently applying this code.
- `delete_discount`: DELETE `/api/v1_2/discounts/{{ record.id }}` - kind `delete`; body type `none`;
  path fields `id`; required record fields `id`; accepted fields `id`; missing records treated as
  success for status `404`; risk: permanently removes a discount code; any buyer with the code saved
  can no longer redeem it.
- `update_bundle`: PUT `/api/v1/packages/{{ record.id }}` - kind `update`; body type `form`; path
  fields `id`; required record fields `id`; accepted fields `id`, `name`, `price`.
- `delete_bundle`: DELETE `/api/v1/packages/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `404`; risk: permanently removes a bundle; breaks any existing order-fulfillment links
  referencing it.
- `refund_order`: POST `/api/v1/orders/{{ record.id }}/refund` - kind `update`; body type `form`;
  path fields `id`; body fields `amount`, `cancel_subscription`, `revoke_access`; required record
  fields `id`, `amount`; accepted fields `amount`, `cancel_subscription`, `id`, `revoke_access`;
  risk: issues a real financial refund against the buyer's original payment method; irreversible
  external money movement, approval required.
- `cancel_order_subscription`: PUT `/api/v1/orders/{{ record.id }}/cancel_subscription` - kind
  `update`; body type `none`; path fields `id`; required record fields `id`; accepted fields `id`;
  risk: cancels the buyer's active recurring subscription tied to this order; stops future recurring
  charges, irreversible without the buyer re-subscribing.
- `resend_order_email`: POST `/api/v1/orders/{{ record.id }}/resend_email` - kind `update`; body
  type `form`; path fields `id`; body fields `type`; required record fields `id`; accepted fields
  `id`, `type`; risk: resends the order confirmation/receipt/download email to the buyer's address
  on file; low-risk external side effect, no approval required.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 6 stream-backed endpoint group(s), 14 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=1, duplicate_of=15, non_data_endpoint=1, out_of_scope=11,
  requires_elevated_scope=2.
