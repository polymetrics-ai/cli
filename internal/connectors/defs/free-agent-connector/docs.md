# Overview

Reads FreeAgent contacts, invoices, bills, projects, and tasks through the FreeAgent v2 REST API
using OAuth2 refresh-token authentication.

Readable streams: `contacts`, `invoices`, `bills`, `projects`, `tasks`.

This connector is read-only; no write actions are declared.

Service API documentation: https://dev.freeagent.com/.

## Auth setup

Connection fields:

- `base_url` (optional, string).
- `client_id` (required, secret, string).
- `client_refresh_token_2` (required, secret, string).
- `client_secret` (required, secret, string).
- `mode` (optional, string).
- `payroll_year` (optional, string).
- `updated_since` (optional, string).

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

- `contacts`: GET connector-managed request path - records path `data`; incremental cursor
  `updated_at`; sent as `updated_since`; formatted as `rfc3339`.
- `invoices`: GET connector-managed request path - records path `data`; incremental cursor
  `updated_at`; sent as `updated_since`; formatted as `rfc3339`.
- `bills`: GET connector-managed request path - records path `data`; incremental cursor
  `updated_at`; sent as `updated_since`; formatted as `rfc3339`.
- `projects`: GET connector-managed request path - records path `data`; incremental cursor
  `updated_at`; sent as `updated_since`; formatted as `rfc3339`.
- `tasks`: GET connector-managed request path - records path `data`; incremental cursor
  `updated_at`; sent as `updated_since`; formatted as `rfc3339`.

## Write actions & risks

This connector is read-only; no reverse-ETL write actions are declared.

## Known limits

- API coverage includes 5 stream-backed endpoint group(s).
