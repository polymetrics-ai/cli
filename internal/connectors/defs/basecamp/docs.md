# Overview

Reads Basecamp 3 projects, people, and account activity events through the Basecamp REST API.

Readable streams: `projects`, `people`, `events`.

This connector is read-only; no write actions are declared.

Service API documentation: https://github.com/basecamp/bc3-api.

## Auth setup

Connection fields:

- `account_id` (required, string).
- `base_url` (optional, string).
- `client_id` (required, secret, string).
- `client_refresh_token_2` (required, secret, string).
- `client_secret` (required, secret, string).
- `mode` (optional, string).
- `start_date` (required, string).

Secret fields are redacted in logs and write previews: `client_id`, `client_refresh_token_2`,
`client_secret`.

Provide the secret fields listed above. Authentication is applied by the connector-specific
implementation for this service.

Requests use the configured `base_url` value after applying defaults.

Connection checks use a connector-managed request.

## Streams notes

Default pagination: single request; no pagination.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `projects`: GET connector-managed request path - records path `data`; incremental cursor
  `updated_at`; formatted as `rfc3339`; records at or before the lower bound are filtered
  client-side.
- `people`: GET connector-managed request path - records path `data`; incremental cursor
  `updated_at`; formatted as `rfc3339`; records at or before the lower bound are filtered
  client-side.
- `events`: GET connector-managed request path - records path `data`; incremental cursor
  `created_at`; formatted as `rfc3339`; records at or before the lower bound are filtered
  client-side.

## Write actions & risks

This connector is read-only; no reverse-ETL write actions are declared.

## Known limits

- API coverage includes 3 stream-backed endpoint group(s).
- Client-side incremental filtering is used for: `projects`, `people`, `events`.
