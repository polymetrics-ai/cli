# Overview

Reads Recharge customers, subscriptions, and orders through the Recharge REST API.

Readable streams: `customers`, `subscriptions`, `orders`.

This connector is read-only; no write actions are declared.

Service API documentation: https://developer.rechargepayments.com/2021-11/.

## Auth setup

Connection fields:

- `access_token` (required, secret, string); Recharge API access token, sent as the
  X-Recharge-Access-Token header. Never logged.
- `api_version` (optional, string); default `2021-11`; Recharge API version, sent as the
  X-Recharge-Version header.
- `base_url` (optional, string); default `https://api.rechargeapps.com`; format `uri`; Recharge API
  base URL override for tests or proxies.
- `mode` (optional, string).

Secret fields are redacted in logs and write previews: `access_token`.

Default configuration values: `api_version=2021-11`, `base_url=https://api.rechargeapps.com`.

Authentication behavior:

- API key authentication in `X-Recharge-Access-Token` using `secrets.access_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/customers` with query `limit`=`1`.

## Streams notes

Default pagination: cursor pagination; cursor parameter `cursor`; next token from `next_cursor`;
page size 250.

- `customers`: GET `/customers` - records path `customers`; query `limit`=`250`; cursor pagination;
  cursor parameter `cursor`; next token from `next_cursor`; page size 250.
- `subscriptions`: GET `/subscriptions` - records path `subscriptions`; query `limit`=`250`; cursor
  pagination; cursor parameter `cursor`; next token from `next_cursor`; page size 250.
- `orders`: GET `/orders` - records path `orders`; query `limit`=`250`; cursor pagination; cursor
  parameter `cursor`; next token from `next_cursor`; page size 250.

## Write actions & risks

This connector is read-only. Read behavior: external Recharge API read of customer, subscription,
and order data.

## Known limits

- Batch defaults: read_page_size=250.
- API coverage includes 3 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=6.
