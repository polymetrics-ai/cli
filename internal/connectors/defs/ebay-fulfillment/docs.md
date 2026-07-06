# Overview

Reads eBay seller orders, exploded line items, shipping fulfillments, and payment disputes through
the eBay Sell Fulfillment REST API.

Readable streams: `orders`, `order_line_items`, `shipping_fulfillments`, `payment_disputes`.

This connector is read-only; no write actions are declared.

Service API documentation: https://developer.ebay.com/api-docs/sell/fulfillment/overview.html.

## Auth setup

Connection fields:

- `base_url` (optional, string); default `https://api.ebay.com`; format `uri`; eBay API host
  override for tests, sandbox, or proxies.
- `mode` (optional, string).
- `page_size` (optional, string); default `50`; Records per page (1-1000, limit).
- `password` (optional, secret, string); eBay OAuth 2.0 client secret, sent as the Basic-auth
  password on the refresh-token exchange. Never logged.
- `refresh_token` (required, secret, string); Long-lived eBay OAuth 2.0 refresh token. The 3-legged
  consent/acquisition dance is out of scope for this connector (credentials layer already owns it).
- `refresh_token_endpoint` (optional, string); default
  `https://api.ebay.com/identity/v1/oauth2/token`; format `uri`; eBay OAuth 2.0 token endpoint
  override. MUST be https in production; the hook fails closed on a non-https or unparseable value
  to prevent exfiltrating the refresh token to an attacker-chosen endpoint.
- `scope` (optional, string); Space-separated OAuth scopes requested on the token-refresh grant.
- `start_date` (optional, string); format `date-time`; RFC3339 lower bound for
  orders/order_line_items/shipping_fulfillments, sent as an eBay creationdate filter
  (creationdate:[<value>..]). Only used on a fresh sync with no persisted cursor.
- `username` (optional, secret, string); eBay OAuth 2.0 client id (application client id), sent as
  the Basic-auth username on the refresh-token exchange. Never logged.

Secret fields are redacted in logs and write previews: `password`, `refresh_token`, `username`.

Default configuration values: `base_url=https://api.ebay.com`, `page_size=50`,
`refresh_token_endpoint=https://api.ebay.com/identity/v1/oauth2/token`.

Authentication behavior:

- Connector-specific authentication using `secrets.refresh_token`, `config.refresh_token_endpoint`,
  `secrets.username`, `secrets.password`, `config.scope`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/sell/fulfillment/v1/order` with query `limit`=`1`.

## Streams notes

Default pagination: single request; no pagination.

Pagination by stream: none: `order_line_items`, `shipping_fulfillments`; offset_limit: `orders`,
`payment_disputes`.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `orders`: GET `/sell/fulfillment/v1/order` - records path `orders`; query `filter` from template
  `creationdate:[{{ incremental.lower_bound }}..]`, omitted when absent; offset/limit pagination;
  offset parameter `offset`; limit parameter `limit`; page size 2; incremental cursor
  `creation_date`; formatted as `rfc3339`; initial lower bound from `start_date`; computed output
  fields `buyer_username`, `creation_date`, `last_modified_date`, `line_item_count`,
  `order_fulfillment_status`, `order_id`, `order_payment_status`, `sales_record_reference`,
  `seller_id`, `total_currency`, `total_value`.
- `order_line_items`: GET `/sell/fulfillment/v1/order` - records path `orders`.
- `shipping_fulfillments`: GET `/sell/fulfillment/v1/order` - records path `orders`.
- `payment_disputes`: GET `/sell/fulfillment/v1/payment_dispute` - records path
  `paymentDisputeSummaries`; query `filter` from template `creationdate:[{{ incremental.lower_bound
  }}..]`, omitted when absent; offset/limit pagination; offset parameter `offset`; limit parameter
  `limit`; page size 2; incremental cursor `open_date`; formatted as `rfc3339`; initial lower bound
  from `start_date`; computed output fields `amount_currency`, `amount_value`, `buyer_username`,
  `dispute_state`, `dispute_status`, `open_date`, `order_id`, `payment_dispute_id`, `reason`.

## Write actions & risks

This connector is read-only. Read behavior: external eBay Sell Fulfillment API read of a seller's
order, shipment, and dispute data.

## Known limits

- Batch defaults: read_page_size=50.
- API coverage includes 4 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  destructive_admin=2, out_of_scope=5.
