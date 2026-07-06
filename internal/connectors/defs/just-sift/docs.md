# Overview

Reads JustSift people directory profiles and person field definitions through the Sift REST API.

Readable streams: `peoples`, `fields`.

This connector is read-only; no write actions are declared.

Service API documentation: https://sift.com/developers/docs/curl/apis-overview.

## Auth setup

Connection fields:

- `api_token` (required, secret, string); JustSift API token. Used only for Bearer auth
  (Authorization: Bearer <api_token>); never logged.
- `base_url` (optional, string); default `https://api.justsift.com/v1`; format `uri`; JustSift API
  base URL override for tests or proxies.
- `mode` (optional, string).

Secret fields are redacted in logs and write previews: `api_token`.

Default configuration values: `base_url=https://api.justsift.com/v1`.

Authentication behavior:

- Bearer token authentication using `secrets.api_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/fields/person`.

## Streams notes

Default pagination: single request; no pagination.

Pagination by stream: cursor: `fields`; page_number: `peoples`.

- `peoples`: GET `/search/people` - records path `data`; query `limit`=`100`; page-number
  pagination; page parameter `page`; no page-size parameter; starts at 1; page size 100; computed
  output fields `connector`.
- `fields`: GET `/fields/person` - records path `data`; query `limit`=`100`; cursor pagination;
  cursor parameter `cursor`; next token from `links.next`; computed output fields `connector`.

## Write actions & risks

This connector is read-only. Read behavior: external JustSift API read of people directory profiles
and field definitions.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 2 stream-backed endpoint group(s).
