# Overview

Reads Gorgias helpdesk tickets, customers, messages, and satisfaction surveys through the Gorgias
REST API (read-only).

Readable streams: `tickets`, `customers`, `messages`, `satisfaction_surveys`.

This connector is read-only; no write actions are declared.

Service API documentation: https://developers.gorgias.com/reference.

## Auth setup

Connection fields:

- `base_url` (required, string); format `uri`; Your Gorgias account's API base URL, e.g.
- `mode` (optional, string).
- `page_size` (optional, string); default `100`; Records per page (1-100).
- `password` (required, secret, string); Gorgias API key used for HTTP Basic auth (sent as the Basic
  auth password); never logged.
- `username` (required, string); Gorgias account email used for HTTP Basic auth (sent as the Basic
  auth username).

Secret fields are redacted in logs and write previews: `password`.

Default configuration values: `page_size=100`.

Authentication behavior:

- HTTP Basic authentication using `config.username`, `secrets.password`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/tickets` with query `limit`=`1`.

## Streams notes

Default pagination: cursor pagination; cursor parameter `cursor`; next token from
`meta.next_cursor`.

- `tickets`: GET `/tickets` - records path `data`; query `limit` from template `{{ config.page_size
  }}`, default `100`; cursor pagination; cursor parameter `cursor`; next token from
  `meta.next_cursor`.
- `customers`: GET `/customers` - records path `data`; query `limit` from template `{{
  config.page_size }}`, default `100`; cursor pagination; cursor parameter `cursor`; next token from
  `meta.next_cursor`.
- `messages`: GET `/messages` - records path `data`; query `limit` from template `{{
  config.page_size }}`, default `100`; cursor pagination; cursor parameter `cursor`; next token from
  `meta.next_cursor`.
- `satisfaction_surveys`: GET `/satisfaction-surveys` - records path `data`; query `limit` from
  template `{{ config.page_size }}`, default `100`; cursor pagination; cursor parameter `cursor`;
  next token from `meta.next_cursor`.

## Write actions & risks

This connector is read-only. Read behavior: external Gorgias API read of helpdesk tickets,
customers, messages, and satisfaction surveys.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 4 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=7.
