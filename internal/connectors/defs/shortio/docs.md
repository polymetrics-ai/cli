# Overview

Reads Short.io links and domains through the Short.io REST API.

Readable streams: `links`, `domains`.

This connector is read-only; no write actions are declared.

Service API documentation: https://developers.short.io/.

## Auth setup

Connection fields:

- `api_key` (optional, secret, string); Short.io API key, sent verbatim as the Authorization header
  (no Bearer prefix, matching Short.io's own convention).
- `base_url` (optional, string); default `https://api.short.io`; format `uri`; Short.io API base
  URL.
- `page_size` (optional, integer); default `150`.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.short.io`, `page_size=150`.

Authentication behavior:

- API key authentication in `Authorization` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/api/links`.

## Streams notes

Default pagination: cursor pagination; cursor parameter `nextPageToken`; next token from
`nextPageToken`.

- `links`: GET `/api/links` - records path `links`; query `limit` from template `{{ config.page_size
  }}`, default `150`; cursor pagination; cursor parameter `nextPageToken`; next token from
  `nextPageToken`; computed output fields `id`, `title`, `updated_at`.
- `domains`: GET `/api/domains` - records path `domains`; query `limit` from template `{{
  config.page_size }}`, default `150`; cursor pagination; cursor parameter `nextPageToken`; next
  token from `nextPageToken`; computed output fields `id`, `name`, `title`, `updated_at`.

## Write actions & risks

This connector is read-only. Read behavior: external Short.io API read of link and domain data.

## Known limits

- Batch defaults: read_page_size=150.
- API coverage includes 2 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=3.
