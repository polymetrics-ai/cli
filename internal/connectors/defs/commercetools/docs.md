# Overview

Reads commercetools customers, orders, and products through the HTTP API.

Readable streams: `customers`, `orders`, `products`.

This connector is read-only; no write actions are declared.

Service API documentation: https://docs.commercetools.com/api/.

## Auth setup

Connection fields:

- `base_url` (required, string); format `uri`; commercetools API base URL, e.g.
  https://api.<region>.<host>.commercetools.com.
- `client_id` (required, secret, string); commercetools OAuth2 client-credentials client id. Never
  logged.
- `client_secret` (required, secret, string); commercetools OAuth2 client-credentials client secret.
  Never logged.
- `mode` (optional, string).
- `project_key` (required, string); commercetools project key; every stream path is scoped to this
  project.
- `token_url` (required, string); format `uri`; commercetools OAuth2 token endpoint, e.g.
  https://auth.<region>.<host>.commercetools.com/oauth/token. Required for the same reason base_url
  is required (see docs.md Known limits).

Secret fields are redacted in logs and write previews: `client_id`, `client_secret`.

Authentication behavior:

- OAuth 2.0 client credentials authentication using `config.token_url`, `secrets.client_id`,
  `secrets.client_secret`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/{{ config.project_key }}/customers` with query `limit`=`1`.

## Streams notes

Default pagination: offset/limit pagination; offset parameter `offset`; limit parameter `limit`;
page size 100; maximum 100 page(s).

- `customers`: GET `/{{ config.project_key }}/customers` - records path `results`; offset/limit
  pagination; offset parameter `offset`; limit parameter `limit`; page size 100; maximum 100
  page(s); emits passthrough records.
- `orders`: GET `/{{ config.project_key }}/orders` - records path `results`; offset/limit
  pagination; offset parameter `offset`; limit parameter `limit`; page size 100; maximum 100
  page(s); emits passthrough records.
- `products`: GET `/{{ config.project_key }}/products` - records path `results`; offset/limit
  pagination; offset parameter `offset`; limit parameter `limit`; page size 100; maximum 100
  page(s); emits passthrough records.

## Write actions & risks

This connector is read-only. Read behavior: external commercetools API read of customer, order, and
product data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 3 stream-backed endpoint group(s).
