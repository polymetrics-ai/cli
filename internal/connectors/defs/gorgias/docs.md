# Overview

Reads Gorgias helpdesk, customer, account configuration, analytics, integration, job, voice, and
widget list resources through the Gorgias REST API (read-only).

Readable streams: `tickets`, `customers`, `messages`, `satisfaction_surveys`, `account_settings`,
`custom_fields`, `events`, `integrations`, `jobs`, `macros`, `metric_cards`, `rules`, `tags`,
`teams`, `users`, `views`, `voice_calls`, `voice_call_events`, `widgets`,
`customer_custom_fields`, `ticket_custom_fields`, `ticket_tags`, `ticket_messages`, and
`view_items`.

This connector is read-only; no write actions are declared.

Service API documentation: https://developers.gorgias.com/reference.

CLI surface metadata is present for provider-inspired Gorgias commands. Stream-backed `list` and
`search` commands are marked implemented only where they map to typed ETL streams; write,
direct-read detail, binary/file, and sensitive/admin commands remain planned for later issue lanes
and are not exposed as raw API tools.

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
`meta.next_cursor`. Stream requests send `limit` from `{{ config.page_size }}`, default `100`.

Top-level streams:

- `tickets`: GET `/tickets` - records path `data`.
- `customers`: GET `/customers` - records path `data`.
- `messages`: GET `/messages` - records path `data`.
- `satisfaction_surveys`: GET `/satisfaction-surveys` - records path `data`.
- `account_settings`: GET `/account/settings` - records path `data`.
- `custom_fields`: GET `/custom-fields` - records path `data`.
- `events`: GET `/events` - records path `data`.
- `integrations`: GET `/integrations` - records path `data`.
- `jobs`: GET `/jobs` - records path `data`.
- `macros`: GET `/macros` - records path `data`.
- `metric_cards`: GET `/metric-cards` - records path `data`.
- `rules`: GET `/rules` - records path `data`.
- `tags`: GET `/tags` - records path `data`.
- `teams`: GET `/teams` - records path `data`.
- `users`: GET `/users` - records path `data`.
- `views`: GET `/views` - records path `data`.
- `voice_calls`: GET `/phone/voice-calls` - records path `data`.
- `voice_call_events`: GET `/phone/voice-call-events` - records path `data`.
- `widgets`: GET `/widgets` - records path `data`.

Fan-out streams:

- `customer_custom_fields`: lists `/customers/{customer_id}/custom-fields` for customer IDs from
  `customers` and stamps `customer_id`.
- `ticket_custom_fields`: lists `/tickets/{ticket_id}/custom-fields` for ticket IDs from `tickets`
  and stamps `ticket_id`.
- `ticket_tags`: lists `/tickets/{ticket_id}/tags` for ticket IDs from `tickets` and stamps
  `ticket_id`.
- `ticket_messages`: lists `/tickets/{ticket_id}/messages` for ticket IDs from `tickets` and stamps
  `ticket_id`.
- `view_items`: lists `/views/{view_id}/items` for view IDs from `views` and stamps `view_id`.

## Write actions & risks

This connector is read-only. Read behavior: external Gorgias API reads of helpdesk, customer,
account configuration, analytics, integration, job, voice, and widget list resources.

## Known limits

- Batch defaults: read_page_size=100.
- API surface metadata accounts for all 114 public operations captured from Gorgias `llms.txt` plus
  linked ReadMe OpenAPI pages.
- Executable API coverage is limited to 24 stream-backed GET endpoints. Detail direct reads, binary
  payloads, advanced POST query/report operations, reverse-ETL writes, admin mutations, destructive
  actions, and product-scope operations remain blocked metadata until later lanes add typed support.
- Fan-out is single-level only; endpoints requiring nested or caller-supplied IDs remain direct-read
  or future-lane candidates rather than raw API passthroughs.
