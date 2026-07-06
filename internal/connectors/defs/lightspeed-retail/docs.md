# Overview

Reads Lightspeed Retail (X-Series) products, customers, sales, outlets, and registers through the
Lightspeed REST API. Read-only.

Readable streams: `products`, `customers`, `sales`, `outlets`, `registers`.

This connector is read-only; no write actions are declared.

Service API documentation: https://developers.lightspeedhq.com/retail/introduction/introduction/.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Lightspeed Retail personal token / OAuth access token. Sent
  as Authorization: Bearer <api_key>; never logged.
- `mode` (optional, string).
- `subdomain` (required, string); Your Lightspeed Retail (X-Series) account subdomain (the
  <subdomain> in https://<subdomain>.retail.lightspeed.app). Used to derive the API base URL.

Secret fields are redacted in logs and write previews: `api_key`.

Authentication behavior:

- Bearer token authentication using `secrets.api_key`.

Requests use base URL `https://{{ config.subdomain }}.retail.lightspeed.app` after applying
configuration defaults.

Connection checks call GET `/api/2.0/outlets` with query `page_size`=`1`.

## Streams notes

Default pagination: cursor pagination; cursor parameter `after`; next token from `version.max`; page
size 100.

- `products`: GET `/api/2.0/products` - records path `data`; query `page_size`=`100`; cursor
  pagination; cursor parameter `after`; next token from `version.max`; page size 100.
- `customers`: GET `/api/2.0/customers` - records path `data`; query `page_size`=`100`; cursor
  pagination; cursor parameter `after`; next token from `version.max`; page size 100.
- `sales`: GET `/api/2.0/sales` - records path `data`; query `page_size`=`100`; cursor pagination;
  cursor parameter `after`; next token from `version.max`; page size 100.
- `outlets`: GET `/api/2.0/outlets` - records path `data`; query `page_size`=`100`; cursor
  pagination; cursor parameter `after`; next token from `version.max`; page size 100.
- `registers`: GET `/api/2.0/registers` - records path `data`; query `page_size`=`100`; cursor
  pagination; cursor parameter `after`; next token from `version.max`; page size 100.

## Write actions & risks

This connector is read-only. Read behavior: external Lightspeed Retail API read of product,
customer, and sales data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 5 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=6.
