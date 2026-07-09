# Overview

Reads Gorgias helpdesk tickets, customers, messages, and satisfaction surveys through the Gorgias
REST API (read-only).

Readable streams: `tickets`, `customers`, `messages`, `satisfaction_surveys`.

This connector is read-only; no write actions are declared.

Service API documentation: https://developers.gorgias.com/reference.

CLI surface metadata is present for provider-inspired Gorgias commands. Only the four stream-backed
`list` commands are marked implemented in this slice; write, direct-read, binary/file, and
sensitive/admin commands are planned for later issue lanes and are not exposed as raw API tools.

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
- API coverage currently includes 4 stream-backed endpoint group(s) and 7 legacy out-of-scope rows.
- The public Gorgias llms.txt plus linked ReadMe OpenAPI pages expose 114 operations. Issue #200
  owns the complete operation ledger; this metadata slice does not claim full 114-operation runtime
  parity.
- Planned CLI entries in `cli_surface.json` are discovery/help metadata only until later lanes add
  streams, direct reads, binary policies, or typed reverse-ETL actions.
