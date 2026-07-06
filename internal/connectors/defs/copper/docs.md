# Overview

Reads Copper CRM people, companies, opportunities, leads, and tasks through the Copper REST API.

Readable streams: `people`, `companies`, `opportunities`, `leads`, `tasks`.

This connector is read-only; no write actions are declared.

Service API documentation: https://developer.copper.com/.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Copper API key.
- `base_url` (optional, string).
- `mode` (optional, string).
- `user_email` (required, string); user email used to login in to Copper.

Secret fields are redacted in logs and write previews: `api_key`.

Provide the secret fields listed above. Authentication is applied by the connector-specific
implementation for this service.

Requests use the configured `base_url` value after applying defaults.

Connection checks use a connector-managed request.

## Streams notes

Default pagination: single request; no pagination.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `people`: GET connector-managed request path - records path `data`; incremental cursor
  `date_modified`; formatted as `rfc3339`; records at or before the lower bound are filtered
  client-side.
- `companies`: GET connector-managed request path - records path `data`; incremental cursor
  `date_modified`; formatted as `rfc3339`; records at or before the lower bound are filtered
  client-side.
- `opportunities`: GET connector-managed request path - records path `data`; incremental cursor
  `date_modified`; formatted as `rfc3339`; records at or before the lower bound are filtered
  client-side.
- `leads`: GET connector-managed request path - records path `data`; incremental cursor
  `date_modified`; formatted as `rfc3339`; records at or before the lower bound are filtered
  client-side.
- `tasks`: GET connector-managed request path - records path `data`; incremental cursor
  `date_modified`; formatted as `rfc3339`; records at or before the lower bound are filtered
  client-side.

## Write actions & risks

This connector is read-only; no reverse-ETL write actions are declared.

## Known limits

- API coverage includes 5 stream-backed endpoint group(s).
- Client-side incremental filtering is used for: `people`, `companies`, `opportunities`, `leads`,
  `tasks`.
