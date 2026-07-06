# Overview

Reads Lever Hiring opportunities, postings, users, requisitions, and stages through the Lever Data
API. Read-only (full-refresh).

Readable streams: `opportunities`, `postings`, `users`, `requisitions`, `stages`.

This connector is read-only; no write actions are declared.

Service API documentation: https://hire.lever.co/developer/documentation.

## Auth setup

Connection fields:

- `access_token` (optional, secret, string); Lever OAuth2 access token. Sent as Authorization:
  Bearer <access_token>. Never logged.
- `api_key` (optional, secret, string); Lever API key. Sent as the HTTP Basic auth username with a
  blank password (Lever's documented API-key auth convention). Never logged.
- `base_url` (optional, string); default `https://api.lever.co/v1`; format `uri`; Lever Data API
  base URL override for tests, proxies, or a sandbox environment (https://api.sandbox.lever.co/v1).
- `mode` (optional, string).

Secret fields are redacted in logs and write previews: `access_token`, `api_key`.

Default configuration values: `base_url=https://api.lever.co/v1`.

Authentication behavior:

- Bearer token authentication using `secrets.access_token` when `{{ secrets.access_token }}`.
- HTTP Basic authentication using `secrets.api_key` when `{{ secrets.api_key }}`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/postings` with query `limit`=`1`.

## Streams notes

Default pagination: cursor pagination; cursor parameter `offset`; next token from `next`; stop flag
`hasNext`; page size 100.

- `opportunities`: GET `/opportunities` - records path `data`; query `limit`=`100`; cursor
  pagination; cursor parameter `offset`; next token from `next`; stop flag `hasNext`; page size 100.
- `postings`: GET `/postings` - records path `data`; query `limit`=`100`; cursor pagination; cursor
  parameter `offset`; next token from `next`; stop flag `hasNext`; page size 100.
- `users`: GET `/users` - records path `data`; query `limit`=`100`; cursor pagination; cursor
  parameter `offset`; next token from `next`; stop flag `hasNext`; page size 100.
- `requisitions`: GET `/requisitions` - records path `data`; query `limit`=`100`; cursor pagination;
  cursor parameter `offset`; next token from `next`; stop flag `hasNext`; page size 100.
- `stages`: GET `/stages` - records path `data`; query `limit`=`100`; cursor pagination; cursor
  parameter `offset`; next token from `next`; stop flag `hasNext`; page size 100.

## Write actions & risks

This connector is read-only. Read behavior: external Lever API read of candidate and hiring pipeline
data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 5 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=6.
