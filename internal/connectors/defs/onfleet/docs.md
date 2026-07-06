# Overview

Reads Onfleet tasks, workers, teams, hubs, and administrators through the Onfleet REST API.

Readable streams: `tasks`, `workers`, `teams`, `hubs`, `administrators`.

This connector is read-only; no write actions are declared.

Service API documentation: https://docs.onfleet.com/reference.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Onfleet API key, sent as the HTTP Basic auth username with
  an empty password. Never logged.
- `base_url` (optional, string); default `https://onfleet.com/api/v2`; format `uri`; Onfleet API
  base URL override for tests or proxies.
- `max_pages` (optional, string); default `0`; Maximum pages for the paginated tasks stream; use 0,
  all, or unlimited to exhaust the stream.
- `mode` (optional, string).

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://onfleet.com/api/v2`, `max_pages=0`.

Authentication behavior:

- HTTP Basic authentication using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/auth/test`.

## Streams notes

Default pagination: single request; no pagination.

Pagination by stream: cursor: `tasks`; none: `workers`, `teams`, `hubs`, `administrators`.

- `tasks`: GET `/tasks/all` - records path `tasks`; cursor pagination; cursor parameter `lastId`;
  next token from `lastId`.
- `workers`: GET `/workers` - records path `.`.
- `teams`: GET `/teams` - records path `.`.
- `hubs`: GET `/hubs` - records path `.`.
- `administrators`: GET `/admins` - records path `.`.

## Write actions & risks

This connector is read-only. Read behavior: external Onfleet API read of delivery task and workforce
data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 5 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  destructive_admin=2, non_data_endpoint=1, out_of_scope=2.
