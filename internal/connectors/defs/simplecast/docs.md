# Overview

Reads Simplecast podcasts and episodes through the Simplecast REST API.

Readable streams: `podcasts`, `episodes`.

This connector is read-only; no write actions are declared.

Service API documentation: https://developers.simplecast.com/.

## Auth setup

Connection fields:

- `access_token` (required, secret, string); Simplecast OAuth access token, sent as a Bearer token
  (Authorization: Bearer <access_token>). Never logged.
- `base_url` (optional, string); default `https://api.simplecast.com`; format `uri`; Simplecast API
  base URL override for tests or proxies.

Secret fields are redacted in logs and write previews: `access_token`.

Default configuration values: `base_url=https://api.simplecast.com`.

Authentication behavior:

- Bearer token authentication using `secrets.access_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/podcasts` with query `limit`=`1`; `page`=`1`.

## Streams notes

Default pagination: single request; no pagination.

- `podcasts`: GET `/podcasts` - records path `collection`; query `limit`=`100`; `page`=`1`; follows
  a next-page URL from the response body; URL path `pages.next.href`; next URLs stay on the
  configured API host; computed output fields `title`, `updated_at`.
- `episodes`: GET `/episodes` - records path `collection`; query `limit`=`100`; `page`=`1`; follows
  a next-page URL from the response body; URL path `pages.next.href`; next URLs stay on the
  configured API host; computed output fields `title`, `updated_at`.

## Write actions & risks

This connector is read-only. Read behavior: external Simplecast API read of podcast and episode
data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 2 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=2.
