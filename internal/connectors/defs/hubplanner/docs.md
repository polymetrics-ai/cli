# Overview

Reads Hubplanner resources, projects, clients, events, holidays, bookings, and billing rates through
the Hubplanner REST API.

Readable streams: `resources`, `projects`, `clients`, `events`, `holidays`, `bookings`,
`billing_rates`.

This connector is read-only; no write actions are declared.

Service API documentation: https://github.com/hubplanner/API.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Hubplanner API key, sent verbatim (no Bearer prefix) as the
  Authorization header. Never logged.
- `base_url` (optional, string); default `https://api.hubplanner.com/v1`; format `uri`; Hubplanner
  API base URL override for tests or proxies.
- `max_pages` (optional, string); default `0`; Maximum pages; use 0, all, or unlimited to exhaust
  the stream.
- `mode` (optional, string).
- `page_size` (optional, string); default `200`; Records per page (1-1000).

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.hubplanner.com/v1`, `max_pages=0`,
`page_size=200`.

Authentication behavior:

- API key authentication in `Authorization` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/resource` with query `limit`=`1`; `page`=`0`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `limit`; starts at
0; page size 200.

- `resources`: GET `/resource` - records at response root; page-number pagination; page parameter
  `page`; size parameter `limit`; starts at 0; page size 200.
- `projects`: GET `/project` - records at response root; page-number pagination; page parameter
  `page`; size parameter `limit`; starts at 0; page size 200.
- `clients`: GET `/client` - records at response root; page-number pagination; page parameter
  `page`; size parameter `limit`; starts at 0; page size 200.
- `events`: GET `/event` - records at response root; page-number pagination; page parameter `page`;
  size parameter `limit`; starts at 0; page size 200.
- `holidays`: GET `/holiday` - records at response root; page-number pagination; page parameter
  `page`; size parameter `limit`; starts at 0; page size 200.
- `bookings`: GET `/booking` - records at response root; page-number pagination; page parameter
  `page`; size parameter `limit`; starts at 0; page size 200.
- `billing_rates`: GET `/billingrate` - records at response root; page-number pagination; page
  parameter `page`; size parameter `limit`; starts at 0; page size 200.

## Write actions & risks

This connector is read-only. Read behavior: external Hubplanner API read of scheduling, project, and
billing data.

## Known limits

- Batch defaults: read_page_size=200.
- API coverage includes 7 stream-backed endpoint group(s).
