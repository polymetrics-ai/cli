# Overview

Reads Zoho Desk tickets, contacts, and accounts through the Zoho Desk REST API.

Readable streams: `tickets`, `contacts`, `accounts`.

This connector is read-only; no write actions are declared.

Service API documentation: https://desk.zoho.com/DeskAPIDocument.

## Auth setup

Connection fields:

- `access_token` (required, secret, string); Zoho OAuth access token. Sent as the Authorization
  header with a 'Zoho-oauthtoken ' prefix; never logged.
- `base_url` (optional, string); default `https://desk.zoho.com/api/v1`; format `uri`; Zoho Desk API
  base URL override for tests or region-specific data centers.
- `max_pages` (optional, string); default `0`; Maximum pages; use 0, all, or unlimited to exhaust
  the stream.
- `mode` (optional, string).
- `org_id` (optional, string); Optional Zoho Desk organization ID; sent as the orgId header on every
  request when set.
- `page_size` (optional, string); default `100`; Records per page (1-100).

Secret fields are redacted in logs and write previews: `access_token`.

Default configuration values: `base_url=https://desk.zoho.com/api/v1`, `max_pages=0`,
`page_size=100`.

Authentication behavior:

- API key authentication in `Authorization` with prefix `Zoho-oauthtoken` using
  `secrets.access_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/tickets`.

## Streams notes

Default pagination: offset/limit pagination; offset parameter `from`; limit parameter `limit`; page
size 100.

- `tickets`: GET `/tickets` - records path `data`; offset/limit pagination; offset parameter `from`;
  limit parameter `limit`; page size 100; computed output fields `id`, `name`, `updated_at`; emits
  passthrough records.
- `contacts`: GET `/contacts` - records path `data`; offset/limit pagination; offset parameter
  `from`; limit parameter `limit`; page size 100; computed output fields `id`, `name`, `updated_at`;
  emits passthrough records.
- `accounts`: GET `/accounts` - records path `data`; offset/limit pagination; offset parameter
  `from`; limit parameter `limit`; page size 100; computed output fields `id`, `name`, `updated_at`;
  emits passthrough records.

## Write actions & risks

This connector is read-only. Read behavior: external Zoho Desk API read of support ticket and
contact data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 3 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  non_data_endpoint=1, out_of_scope=5.
