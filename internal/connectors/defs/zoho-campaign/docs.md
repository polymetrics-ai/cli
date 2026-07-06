# Overview

Reads Zoho Campaigns lists, campaigns, and contacts through the Zoho Campaigns REST API.

Readable streams: `lists`, `campaigns`, `contacts`.

This connector is read-only; no write actions are declared.

Service API documentation: https://www.zoho.com/campaigns/help/developers/api-overview.html.

## Auth setup

Connection fields:

- `access_token` (required, secret, string); Zoho OAuth access token. Sent as the Authorization
  header with a 'Zoho-oauthtoken ' prefix; never logged.
- `base_url` (optional, string); default `https://campaigns.zoho.com/api/v1.1`; format `uri`; Zoho
  Campaigns API base URL override for tests or region-specific data centers.
- `max_pages` (optional, string); default `0`; Maximum pages; use 0, all, or unlimited to exhaust
  the stream.
- `mode` (optional, string).
- `page_size` (optional, string); default `200`; Records per page (1-200).

Secret fields are redacted in logs and write previews: `access_token`.

Default configuration values: `base_url=https://campaigns.zoho.com/api/v1.1`, `max_pages=0`,
`page_size=200`.

Authentication behavior:

- API key authentication in `Authorization` with prefix `Zoho-oauthtoken` using
  `secrets.access_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/lists`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `per_page`; starts
at 1; page size 200.

- `lists`: GET `/lists` - records path `lists`; page-number pagination; page parameter `page`; size
  parameter `per_page`; starts at 1; page size 200; computed output fields `id`, `name`,
  `updated_at`; emits passthrough records.
- `campaigns`: GET `/campaigns` - records path `campaigns`; page-number pagination; page parameter
  `page`; size parameter `per_page`; starts at 1; page size 200; computed output fields `id`,
  `name`, `updated_at`; emits passthrough records.
- `contacts`: GET `/contacts` - records path `contacts`; page-number pagination; page parameter
  `page`; size parameter `per_page`; starts at 1; page size 200; computed output fields `id`,
  `name`, `updated_at`; emits passthrough records.

## Write actions & risks

This connector is read-only. Read behavior: external Zoho Campaigns API read of email marketing
data.

## Known limits

- Batch defaults: read_page_size=200.
- API coverage includes 3 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  non_data_endpoint=1, out_of_scope=4.
