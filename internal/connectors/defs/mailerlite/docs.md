# Overview

Reads MailerLite subscribers, campaigns, groups, segments, and automations through the MailerLite v2
REST API.

Readable streams: `subscribers`, `campaigns`, `groups`, `segments`, `automations`.

This connector is read-only; no write actions are declared.

Service API documentation: https://developers.mailerlite.com/.

## Auth setup

Connection fields:

- `api_token` (required, secret, string); MailerLite API token. Used only for Bearer auth; never
  logged.
- `base_url` (optional, string); default `https://connect.mailerlite.com/api`; format `uri`;
  MailerLite API base URL override for tests or proxies.
- `mode` (optional, string).

Secret fields are redacted in logs and write previews: `api_token`.

Default configuration values: `base_url=https://connect.mailerlite.com/api`.

Authentication behavior:

- Bearer token authentication using `secrets.api_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/subscribers` with query `limit`=`1`.

## Streams notes

Default pagination: cursor pagination; cursor parameter `cursor`; next token from
`meta.next_cursor`.

- `subscribers`: GET `/subscribers` - records path `data`; query `limit`=`100`; cursor pagination;
  cursor parameter `cursor`; next token from `meta.next_cursor`.
- `campaigns`: GET `/campaigns` - records path `data`; query `limit`=`100`; cursor pagination;
  cursor parameter `cursor`; next token from `meta.next_cursor`.
- `groups`: GET `/groups` - records path `data`; query `limit`=`100`; cursor pagination; cursor
  parameter `cursor`; next token from `meta.next_cursor`.
- `segments`: GET `/segments` - records path `data`; query `limit`=`100`; cursor pagination; cursor
  parameter `cursor`; next token from `meta.next_cursor`.
- `automations`: GET `/automations` - records path `data`; query `limit`=`100`; cursor pagination;
  cursor parameter `cursor`; next token from `meta.next_cursor`.

## Write actions & risks

This connector is read-only. Read behavior: external MailerLite API read of subscriber, campaign,
group, segment, and automation data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 5 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=4.
