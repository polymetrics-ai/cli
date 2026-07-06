# Overview

Reads Planhat companies, end users, and licenses through the Planhat REST API.

Readable streams: `companies`, `endusers`, `licenses`.

This connector is read-only; no write actions are declared.

Service API documentation: https://docs.planhat.com/.

## Auth setup

Connection fields:

- `api_token` (required, secret, string); Planhat API token, sent as a Bearer token. Never logged.
- `base_url` (optional, string); default `https://api.planhat.com`; format `uri`; Planhat API base
  URL override for tests or proxies.
- `max_pages` (optional, string); default `3`; Maximum pages; use 0, all, or unlimited to exhaust
  the stream.
- `mode` (optional, string).
- `page_size` (optional, string); default `100`; Records per page (1-500).

Secret fields are redacted in logs and write previews: `api_token`.

Default configuration values: `base_url=https://api.planhat.com`, `max_pages=3`, `page_size=100`.

Authentication behavior:

- Bearer token authentication using `secrets.api_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/companies` with query `limit`=`1`; `offset`=`0`.

## Streams notes

Default pagination: offset/limit pagination; offset parameter `offset`; limit parameter `limit`;
page size 100; maximum 3 page(s).

- `companies`: GET `/companies` - records path `.`; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 100; maximum 3 page(s); computed output fields `id`,
  `updated_at`.
- `endusers`: GET `/endusers` - records path `.`; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 100; maximum 3 page(s); computed output fields `id`,
  `phase`, `updated_at`.
- `licenses`: GET `/licenses` - records path `.`; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 100; maximum 3 page(s); computed output fields `id`,
  `name`, `phase`, `updated_at`.

## Write actions & risks

This connector is read-only. Read behavior: external Planhat API read of customer success data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 3 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=6.
