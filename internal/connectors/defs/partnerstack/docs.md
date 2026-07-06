# Overview

Reads PartnerStack partnerships, customers, transactions, and groups through the REST API.

Readable streams: `partnerships`, `customers`, `transactions`, `groups`.

This connector is read-only; no write actions are declared.

Service API documentation: https://docs.partnerstack.com/docs/api-overview.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); PartnerStack API key. Used only for Bearer auth; never
  logged.
- `base_url` (optional, string); default `https://api.partnerstack.com/api/v2`; format `uri`;
  PartnerStack API base URL override for tests or proxies.
- `limit` (optional, string); default `100`; Records per page (1-250).
- `max_pages` (optional, string); default `0`; Maximum pages; use 0, all, or unlimited to exhaust
  the stream.
- `mode` (optional, string).

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.partnerstack.com/api/v2`, `limit=100`,
`max_pages=0`.

Authentication behavior:

- Bearer token authentication using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/partnerships` with query `limit`=`1`.

## Streams notes

Default pagination: cursor pagination; cursor parameter `cursor`; next token from `pagination.next`.

- `partnerships`: GET `/partnerships` - records path `data`; query `limit`=`{{ config.limit }}`;
  cursor pagination; cursor parameter `cursor`; next token from `pagination.next`.
- `customers`: GET `/customers` - records path `data`; query `limit`=`{{ config.limit }}`; cursor
  pagination; cursor parameter `cursor`; next token from `pagination.next`.
- `transactions`: GET `/transactions` - records path `data`; query `limit`=`{{ config.limit }}`;
  cursor pagination; cursor parameter `cursor`; next token from `pagination.next`.
- `groups`: GET `/groups` - records path `data`; query `limit`=`{{ config.limit }}`; cursor
  pagination; cursor parameter `cursor`; next token from `pagination.next`.

## Write actions & risks

This connector is read-only. Read behavior: external PartnerStack API read of partnership and
referral-customer data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 4 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=3.
