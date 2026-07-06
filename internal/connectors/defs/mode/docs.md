# Overview

Reads Mode collections (spaces), reports, data sources, groups, and memberships through the Mode
REST API.

Readable streams: `spaces`, `reports`, `data_sources`, `groups`, `memberships`.

This connector is read-only; no write actions are declared.

Service API documentation: https://mode.com/developer/api-reference/.

## Auth setup

Connection fields:

- `api_secret` (required, secret, string); API secret to use as the password for Basic
  Authentication.
- `api_token` (required, secret, string); API token to use as the username for Basic Authentication.
- `base_url` (optional, string).
- `mode` (optional, string).
- `workspace` (required, string).

Secret fields are redacted in logs and write previews: `api_secret`, `api_token`.

Provide the secret fields listed above. Authentication is applied by the connector-specific
implementation for this service.

Requests use the configured `base_url` value after applying defaults.

Connection checks use a connector-managed request.

## Streams notes

Default pagination: single request; no pagination.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `spaces`: GET connector-managed request path - records path `data`; incremental cursor
  `updated_at`; formatted as `rfc3339`; records at or before the lower bound are filtered
  client-side.
- `reports`: GET connector-managed request path - records path `data`; incremental cursor
  `updated_at`; formatted as `rfc3339`; records at or before the lower bound are filtered
  client-side.
- `data_sources`: GET connector-managed request path - records path `data`; incremental cursor
  `updated_at`; formatted as `rfc3339`; records at or before the lower bound are filtered
  client-side.
- `groups`: GET connector-managed request path - records path `data`; incremental cursor
  `updated_at`; formatted as `rfc3339`; records at or before the lower bound are filtered
  client-side.
- `memberships`: GET connector-managed request path - records path `data`; incremental cursor
  `created_at`; formatted as `rfc3339`; records at or before the lower bound are filtered
  client-side.

## Write actions & risks

This connector is read-only; no reverse-ETL write actions are declared.

## Known limits

- API coverage includes 5 stream-backed endpoint group(s).
- Client-side incremental filtering is used for: `spaces`, `reports`, `data_sources`, `groups`,
  `memberships`.
