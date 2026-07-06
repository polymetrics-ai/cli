# Overview

Reads public API directory entries and categories from the api.publicapis.org directory API.
Read-only and credential-free.

Readable streams: `entries`.

This connector is read-only; no write actions are declared.

Service API documentation: https://github.com/davemachado/public-api.

## Auth setup

Connection fields:

- `base_url` (optional, string); default `https://api.publicapis.org`; format `uri`; Public APIs
  directory base URL override for tests or proxies.
- `mode` (optional, string).

Default configuration values: `base_url=https://api.publicapis.org`.

Authentication behavior:

- No authentication.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/entries` with query `limit`=`1`; `offset`=`0`.

## Streams notes

Default pagination: offset/limit pagination; offset parameter `offset`; limit parameter `limit`;
page size 100.

- `entries`: GET `/entries` - records path `entries`; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 100; computed output fields `api`, `auth`,
  `category`, `cors`, `description`, `https`, `id`, `link`.

## Write actions & risks

This connector is read-only. Read behavior: external public-apis.org directory read of API listing
metadata.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 1 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=1.
