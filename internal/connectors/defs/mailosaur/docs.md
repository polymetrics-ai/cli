# Overview

Reads Mailosaur virtual servers, message summaries, and account usage transactions through the
Mailosaur REST API.

Readable streams: `servers`, `messages`, `transactions`.

This connector is read-only; no write actions are declared.

Service API documentation: https://mailosaur.com/docs/api/.

## Auth setup

Connection fields:

- `base_url` (optional, string); default `https://mailosaur.com/api`; format `uri`; Mailosaur API
  base URL override for tests or proxies.
- `items_per_page` (optional, string); default `50`; Records per page (1-1000) for the 'messages'
  stream.
- `mode` (optional, string).
- `password` (required, secret, string); Mailosaur API key, sent as the password of HTTP Basic auth
  (username defaults to the literal 'api'). Never logged.
- `received_after` (optional, string); Optional receivedAfter query filter for the 'messages' stream
  (RFC3339 or Mailosaur's accepted date format).
- `server` (optional, string); Mailosaur virtual server id the 'messages' stream is scoped to
  (required for that stream).
- `username` (optional, string); default `api`; HTTP Basic auth username. Defaults to the literal
  'api', matching Mailosaur's own convention.

Secret fields are redacted in logs and write previews: `password`.

Default configuration values: `base_url=https://mailosaur.com/api`, `items_per_page=50`,
`username=api`.

Authentication behavior:

- HTTP Basic authentication using `config.username`, `secrets.password`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/servers`.

## Streams notes

Default pagination: single request; no pagination.

Pagination by stream: none: `servers`, `transactions`; page_number: `messages`.

- `servers`: GET `/servers` - records path `.`.
- `messages`: GET `/messages` - records path `items`; query `receivedAfter` from template `{{
  config.received_after }}`, omitted when absent; `server`=`{{ config.server }}`; page-number
  pagination; page parameter `page`; size parameter `itemsPerPage`; starts at 0; page size 50.
- `transactions`: GET `/usage/transactions` - records path `items`.

## Write actions & risks

This connector is read-only. Read behavior: external Mailosaur API read of virtual-server,
message-summary, and usage-transaction data.

## Known limits

- Batch defaults: read_page_size=50.
- API coverage includes 3 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  destructive_admin=2, non_data_endpoint=1, out_of_scope=3.
