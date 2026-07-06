# Overview

Reads FactorialHR employees, teams, time-off leaves, leave types, and locations through the
Factorial REST API.

Readable streams: `employees`, `teams`, `leaves`, `leave_types`, `locations`.

This connector is read-only; no write actions are declared.

Service API documentation: https://apidoc.factorialhr.com/.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Factorial API key, sent as the X-API-KEY header. Never
  logged.
- `base_url` (optional, string); default `https://api.factorialhr.com/api/v2/resources`; format
  `uri`; Factorial API base URL override for tests or proxies.
- `mode` (optional, string).

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.factorialhr.com/api/v2/resources`.

Authentication behavior:

- API key authentication in `X-API-KEY` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/api_public/credentials`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `limit`; starts at
1; page size 50.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `employees`: GET `/employees/employees` - records path `data`; page-number pagination; page
  parameter `page`; size parameter `limit`; starts at 1; page size 50; incremental cursor
  `updated_at`; formatted as `rfc3339`.
- `teams`: GET `/teams/teams` - records path `data`; page-number pagination; page parameter `page`;
  size parameter `limit`; starts at 1; page size 50.
- `leaves`: GET `/timeoff/leaves` - records path `data`; page-number pagination; page parameter
  `page`; size parameter `limit`; starts at 1; page size 50; incremental cursor `updated_at`;
  formatted as `rfc3339`.
- `leave_types`: GET `/timeoff/leave_types` - records path `data`; page-number pagination; page
  parameter `page`; size parameter `limit`; starts at 1; page size 50.
- `locations`: GET `/locations/locations` - records path `data`; page-number pagination; page
  parameter `page`; size parameter `limit`; starts at 1; page size 50.

## Write actions & risks

This connector is read-only. Read behavior: external Factorial API read of employee, team, and
time-off data.

## Known limits

- Batch defaults: read_page_size=50.
- API coverage includes 5 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  non_data_endpoint=1.
