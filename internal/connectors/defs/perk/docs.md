# Overview

Reads Perk/TravelPerk trips and invoices through read-only REST list endpoints.

Readable streams: `trips`, `invoices`, `invoice_lines`, `invoice_profiles`, `trip_custom_fields`.

This connector is read-only; no write actions are declared.

Service API documentation: https://api.travelperk.com.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Perk/TravelPerk API key, sent as 'Authorization: ApiKey
  <api_key>'. Never logged.
- `base_url` (optional, string); default `https://api.travelperk.com`; format `uri`; Perk/TravelPerk
  API base URL override for tests or proxies.
- `max_pages` (optional, string); default `0`; Maximum pages; use 0, all, or unlimited to exhaust
  the stream.
- `mode` (optional, string).
- `page_size` (optional, string); default `50`; Records per page (1-100).
- `start_date` (optional, string); format `date-time`; RFC3339 lower bound forwarded as each
  stream's own start-date query parameter (trips: modified_gte, invoices: issuing_date_gte).

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.travelperk.com`, `max_pages=0`, `page_size=50`.

Authentication behavior:

- API key authentication in `Authorization` with prefix `ApiKey` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/trips`.

## Streams notes

Default pagination: offset/limit pagination; offset parameter `offset`; limit parameter `limit`;
page size 50.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `trips`: GET `/trips` - records path `trips`; offset/limit pagination; offset parameter `offset`;
  limit parameter `limit`; page size 50; incremental cursor `modified`; sent as `modified_gte`;
  formatted as `rfc3339`; initial lower bound from `start_date`; emits passthrough records.
- `invoices`: GET `/invoices` - records path `invoices`; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 50; incremental cursor `issuing_date`; sent as
  `issuing_date_gte`; formatted as `rfc3339`; initial lower bound from `start_date`; emits
  passthrough records.
- `invoice_lines`: GET `/invoices/lines` - records path `invoice_lines`; offset/limit pagination;
  offset parameter `offset`; limit parameter `limit`; page size 50; incremental cursor
  `issuing_date`; sent as `issuing_date_gte`; formatted as `rfc3339`; initial lower bound from
  `start_date`; emits passthrough records.
- `invoice_profiles`: GET `/profiles` - records path `profiles`; offset/limit pagination; offset
  parameter `offset`; limit parameter `limit`; page size 50; emits passthrough records.
- `trip_custom_fields`: GET `/trips/{{ fanout.id }}/custom-fields` - single-object response; records
  path `.`; offset/limit pagination; offset parameter `offset`; limit parameter `limit`; page size
  50; fan-out; ids from request `/trips`; id-list records path `trips`; id field `id`; id inserted
  into the request path; stamps `trip_id`; emits passthrough records.

## Write actions & risks

This connector is read-only. Read behavior: external Perk/TravelPerk API read of trip and invoice
data.

## Known limits

- Batch defaults: read_page_size=50.
- API coverage includes 5 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=1, duplicate_of=2, out_of_scope=13.
