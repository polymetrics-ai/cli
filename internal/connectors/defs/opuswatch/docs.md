# Overview

Reads OPUSWatch monitors, incidents, and checks.

Readable streams: `monitors`, `incidents`, `checks`.

This connector is read-only; no write actions are declared.

Service API documentation: https://opuswatch.com/api.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); OPUSWatch API key. Sent as the X-API-Key header; never
  logged.
- `base_url` (optional, string); default `https://opuswatch.com/api`; format `uri`; OPUSWatch API
  base URL override for tests or proxies.
- `mode` (optional, string).
- `page_size` (optional, string); default `100`; Records per page (1-500).

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://opuswatch.com/api`, `page_size=100`.

Authentication behavior:

- API key authentication in `X-API-Key` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/monitors`.

## Streams notes

Default pagination: cursor pagination; cursor parameter `page`; next token from `next_page`.

- `monitors`: GET `/monitors` - records path `data`; query `per_page` from template `{{
  config.page_size }}`, default `100`; cursor pagination; cursor parameter `page`; next token from
  `next_page`.
- `incidents`: GET `/incidents` - records path `data`; query `per_page` from template `{{
  config.page_size }}`, default `100`; cursor pagination; cursor parameter `page`; next token from
  `next_page`.
- `checks`: GET `/checks` - records path `data`; query `per_page` from template `{{ config.page_size
  }}`, default `100`; cursor pagination; cursor parameter `page`; next token from `next_page`.

## Write actions & risks

This connector is read-only. Read behavior: external OPUSWatch API read of monitor, incident, and
check status data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 3 stream-backed endpoint group(s).
