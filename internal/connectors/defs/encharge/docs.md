# Overview

Reads Encharge people, segments, fields, account tags, and schemas through the Encharge REST API.

Readable streams: `peoples`, `segments`, `fields`, `account_tags`, `schemas`.

This connector is read-only; no write actions are declared.

Service API documentation: https://developers.encharge.io/.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Encharge API token, sent as the X-Encharge-Token header.
  Never logged.
- `base_url` (optional, string); default `https://api.encharge.io/v1`; format `uri`; Encharge API
  base URL override for tests or proxies.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.encharge.io/v1`.

Authentication behavior:

- API key authentication in `X-Encharge-Token` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/people/all` with query `limit`=`1`; `offset`=`0`.

## Streams notes

Default pagination: single request; no pagination.

Pagination by stream: none: `segments`, `fields`, `account_tags`, `schemas`; offset_limit:
`peoples`.

- `peoples`: GET `/people/all` - records path `people`; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 100.
- `segments`: GET `/segments` - records path `segments`.
- `fields`: GET `/fields` - records path `items`.
- `account_tags`: GET `/tags-management` - records path `tags`.
- `schemas`: GET `/schemas` - records path `objects`.

## Write actions & risks

This connector is read-only. Read behavior: external Encharge API read of people, segment, field,
and tag data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 5 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=4.
