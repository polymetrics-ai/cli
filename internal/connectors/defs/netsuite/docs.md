# Overview

Reads selected NetSuite REST Record API resources (customers, vendors, items, sales orders),
authenticating with OAuth 1.0a Token-Based Authentication (HMAC-SHA256 request signing). Read-only.

Readable streams: `customers`, `vendors`, `items`, `sales_orders`.

This connector is read-only; no write actions are declared.

Service API documentation:
https://docs.oracle.com/en/cloud/saas/netsuite/ns-online-help/chapter_1540391670.html.

## Auth setup

Connection fields:

- `base_url` (required, string); format `uri`; NetSuite REST Record API base URL, e.g.
  https://<account-id>.suitetalk.api.netsuite.com/services/rest/record/v1.
- `consumer_key` (required, secret, string); NetSuite integration record consumer key (OAuth 1.0a
  oauth_consumer_key). Never logged.
- `consumer_secret` (required, secret, string); NetSuite integration record consumer secret. Used
  only to compute the OAuth 1.0a HMAC-SHA256 request signature; never sent on the wire itself, never
  logged.
- `max_pages` (optional, string); default `0`; Maximum pages; use 0, all, or unlimited to exhaust
  the stream.
- `mode` (optional, string).
- `page_size` (optional, string); default `100`; Records per page (1-1000, limit query param).
- `realm` (required, string); NetSuite account id (OAuth 1.0a realm), e.g. 123456 or 123456_SB1.
  Sent verbatim as the OAuth Authorization header's realm parameter.
- `token_key` (required, secret, string); NetSuite access token (OAuth 1.0a oauth_token). Never
  logged.
- `token_secret` (required, secret, string); NetSuite token secret. Used only to compute the OAuth
  1.0a HMAC-SHA256 request signature; never sent on the wire itself, never logged.

Secret fields are redacted in logs and write previews: `consumer_key`, `consumer_secret`,
`token_key`, `token_secret`.

Default configuration values: `max_pages=0`, `page_size=100`.

Authentication behavior:

- Connector-specific authentication.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/customer` with query `limit`=`1`.

## Streams notes

Default pagination: offset/limit pagination; offset parameter `offset`; limit parameter `limit`;
page size 100.

- `customers`: GET `/customer` - records path `items`; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 100; computed output fields `entity_id`,
  `last_modified_date`, `name`, `status`.
- `vendors`: GET `/vendor` - records path `items`; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 100; computed output fields `entity_id`,
  `last_modified_date`, `name`, `status`.
- `items`: GET `/inventoryItem` - records path `items`; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 100; computed output fields `entity_id`,
  `last_modified_date`, `name`, `status`.
- `sales_orders`: GET `/salesOrder` - records path `items`; offset/limit pagination; offset
  parameter `offset`; limit parameter `limit`; page size 100; computed output fields `entity_id`,
  `last_modified_date`, `name`, `status`.

## Write actions & risks

This connector is read-only. Read behavior: external NetSuite REST Record API read of customer,
vendor, item, and sales order data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 4 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=5.
