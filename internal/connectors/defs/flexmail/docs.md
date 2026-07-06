# Overview

Reads Flexmail contacts, custom fields, interests, segments, and sources through the Flexmail REST
API.

Readable streams: `contacts`, `custom_fields`, `interests`, `segments`, `sources`.

This connector is read-only; no write actions are declared.

Service API documentation: https://api.flexmail.eu/documentation/.

## Auth setup

Connection fields:

- `account_id` (required, string); Flexmail account id, used as the HTTP Basic auth username.
- `base_url` (optional, string); default `https://api.flexmail.eu`; format `uri`; Flexmail API base
  URL override for tests or proxies.
- `mode` (optional, string).
- `page_size` (optional, string); default `500`; Records per page for paginated endpoints (contacts,
  sources); 1-500.
- `personal_access_token` (required, secret, string); Flexmail personal access token, used as the
  HTTP Basic auth password. Never logged.

Secret fields are redacted in logs and write previews: `personal_access_token`.

Default configuration values: `base_url=https://api.flexmail.eu`, `page_size=500`.

Authentication behavior:

- HTTP Basic authentication using `config.account_id`, `secrets.personal_access_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/contacts` with query `limit`=`1`; `offset`=`0`.

## Streams notes

Default pagination: single request; no pagination.

Pagination by stream: none: `custom_fields`, `interests`, `segments`; offset_limit: `contacts`,
`sources`.

- `contacts`: GET `/contacts` - records path `_embedded.item`; offset/limit pagination; offset
  parameter `offset`; limit parameter `limit`; page size 500.
- `custom_fields`: GET `/custom-fields` - records path `_embedded.item`.
- `interests`: GET `/interests` - records path `_embedded.item`.
- `segments`: GET `/segments` - records path `_embedded.item`.
- `sources`: GET `/sources` - records path `_embedded.item`; offset/limit pagination; offset
  parameter `offset`; limit parameter `limit`; page size 500.

## Write actions & risks

This connector is read-only. Read behavior: external Flexmail API read of contact and marketing-list
data.

## Known limits

- Batch defaults: read_page_size=500.
- API coverage includes 5 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  destructive_admin=1, out_of_scope=4.
