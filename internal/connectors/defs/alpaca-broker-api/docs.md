# Overview

Reads Alpaca Broker API accounts, assets, market calendar, clock, country info, account activities,
journals, and per-account positions/watchlists/orders/documents over the Broker REST API
(read-only).

Readable streams: `accounts`, `assets`, `calendar`, `clock`, `country_info`, `account_activities`,
`journals`, `positions`, `watchlists`, `orders`, `documents`.

This connector is read-only; no write actions are declared.

Service API documentation: https://docs.alpaca.markets/docs/broker-api.

## Auth setup

Connection fields:

- `base_url` (optional, string); default `https://broker-api.sandbox.alpaca.markets/v1`; format
  `uri`; Alpaca Broker API base URL. Set to https://broker-api.alpaca.markets/v1 for production.
- `limit` (optional, string); default `20`.
- `password` (required, secret, string); Alpaca Broker API Secret Key, sent as the HTTP Basic auth
  password. Never logged.
- `username` (required, string); Alpaca Broker API Key ID, sent as the HTTP Basic auth username.

Secret fields are redacted in logs and write previews: `password`.

Default configuration values: `base_url=https://broker-api.sandbox.alpaca.markets/v1`, `limit=20`.

Authentication behavior:

- HTTP Basic authentication using `config.username`, `secrets.password`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/clock`.

## Streams notes

Default pagination: single request; no pagination.

Pagination by stream: cursor: `accounts`, `account_activities`; none: `assets`, `calendar`, `clock`,
`country_info`, `journals`, `positions`, `watchlists`, `orders`, `documents`.

- `accounts`: GET `/accounts` - records path `.`; query `limit`=`{{ config.limit }}`; cursor
  pagination; cursor parameter `page_token`; next cursor from last record field `id`.
- `assets`: GET `/assets` - records path `.`; query `limit`=`{{ config.limit }}`.
- `calendar`: GET `/calendar` - records path `.`; query `limit`=`{{ config.limit }}`.
- `clock`: GET `/clock` - records path `.`.
- `country_info`: GET `/country_info` - records path `.`; query `limit`=`{{ config.limit }}`.
- `account_activities`: GET `/accounts/activities` - records path `.`; query `page_size`=`100`;
  cursor pagination; cursor parameter `page_token`; next cursor from last record field `id`.
- `journals`: GET `/journals` - records path `.`.
- `positions`: GET `/trading/accounts/{{ fanout.id }}/positions` - records path `.`; computed output
  fields `id`; fan-out; ids from request `/accounts`; id field `id`; id inserted into the request
  path; stamps `account_id`.
- `watchlists`: GET `/trading/accounts/{{ fanout.id }}/watchlists` - records path `.`; fan-out; ids
  from request `/accounts`; id field `id`; id inserted into the request path; stamps `account_id`.
- `orders`: GET `/trading/accounts/{{ fanout.id }}/orders` - records path `.`; query `status`=`all`;
  fan-out; ids from request `/accounts`; id field `id`; id inserted into the request path; stamps
  `account_id`.
- `documents`: GET `/accounts/{{ fanout.id }}/documents` - records path `.`; fan-out; ids from
  request `/accounts`; id field `id`; id inserted into the request path; stamps `account_id`.

## Write actions & risks

This connector is read-only. Read behavior: external Alpaca Broker API read of account/asset/market
metadata, plus per-account trading positions, orders, watchlists, and document metadata (financial
PII adjacent; no document content is downloaded, only listing metadata).

## Known limits

- Batch defaults: read_page_size=20.
- API coverage includes 11 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=2, deprecated=3, destructive_admin=16, duplicate_of=9, non_data_endpoint=6,
  out_of_scope=18, requires_elevated_scope=4.
