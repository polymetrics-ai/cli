# Overview

Reads Oveit events, orders, and attendees.

Readable streams: `events`, `orders`, `attendees`.

This connector is read-only; no write actions are declared.

Service API documentation: https://l.oveit.com/api-documentation/.

## Auth setup

Connection fields:

- `base_url` (optional, string); default `https://oveit.com/api`; format `uri`; Oveit API root.
  Defaults to https://oveit.com/api when unset.
- `email` (required, string); Oveit account email, paired with password for HTTP Basic auth.
- `page_size` (optional, integer); default `100`; Records requested per page (per_page query param).
- `password` (required, secret, string); Oveit account password. Sent via HTTP Basic auth as
  email:password.

Secret fields are redacted in logs and write previews: `password`.

Default configuration values: `base_url=https://oveit.com/api`, `page_size=100`.

Authentication behavior:

- HTTP Basic authentication using `config.email`, `secrets.password`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/events`.

## Streams notes

Default pagination: cursor pagination; cursor parameter `page`; next token from `next_page`.

- `events`: GET `/events` - records path `data`; query `per_page` from template `{{ config.page_size
  }}`, default `100`; cursor pagination; cursor parameter `page`; next token from `next_page`.
- `orders`: GET `/orders` - records path `data`; query `per_page` from template `{{ config.page_size
  }}`, default `100`; cursor pagination; cursor parameter `page`; next token from `next_page`.
- `attendees`: GET `/attendees` - records path `data`; query `per_page` from template `{{
  config.page_size }}`, default `100`; cursor pagination; cursor parameter `page`; next token from
  `next_page`.

## Write actions & risks

This connector is read-only. Read behavior: external Oveit API read of event, order, and attendee
data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 3 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  duplicate_of=5, requires_elevated_scope=18.
