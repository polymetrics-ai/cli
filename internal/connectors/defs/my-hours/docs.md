# Overview

Reads My Hours clients, projects, team members, tags, and time log activity through the My Hours
REST API.

Readable streams: `clients`, `projects`, `users`, `tags`, `time_logs`.

This connector is read-only; no write actions are declared.

Service API documentation: https://myhours.com/api.

## Auth setup

Connection fields:

- `base_url` (optional, string).
- `email` (required, string); Your My Hours username.
- `logs_batch_size` (optional, string); Pagination size used for retrieving logs in days.
- `mode` (optional, string).
- `password` (required, secret, string); The password associated to the username.
- `start_date` (required, string); Start date for collecting time logs.

Secret fields are redacted in logs and write previews: `password`.

Provide the secret fields listed above. Authentication is applied by the connector-specific
implementation for this service.

Requests use the configured `base_url` value after applying defaults.

Connection checks use a connector-managed request.

## Streams notes

Default pagination: single request; no pagination.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `clients`: GET connector-managed request path - records path `data`.
- `projects`: GET connector-managed request path - records path `data`.
- `users`: GET connector-managed request path - records path `data`.
- `tags`: GET connector-managed request path - records path `data`.
- `time_logs`: GET connector-managed request path - records path `data`; incremental cursor `date`;
  formatted as `rfc3339`; records at or before the lower bound are filtered client-side.

## Write actions & risks

This connector is read-only; no reverse-ETL write actions are declared.

## Known limits

- API coverage includes 5 stream-backed endpoint group(s).
- Client-side incremental filtering is used for: `time_logs`.
