# Overview

Reads Ashby applicant-tracking data - candidates, jobs, applications, and users - through the Ashby
REST API.

Readable streams: `candidates`, `jobs`, `applications`, `users`.

This connector is read-only; no write actions are declared.

Service API documentation: https://developers.ashbyhq.com/.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); The Ashby API Key, see <a
  href=\"https://developers.ashbyhq.com/reference/authentication\">doc</a> here.
- `base_url` (optional, string).
- `mode` (optional, string).
- `start_date` (required, string); UTC date and time in the format 2017-01-25T00:00:00Z. Any data
  before this date will not be replicated.

Secret fields are redacted in logs and write previews: `api_key`.

Provide the secret fields listed above. Authentication is applied by the connector-specific
implementation for this service.

Requests use the configured `base_url` value after applying defaults.

Connection checks use a connector-managed request.

## Streams notes

Default pagination: single request; no pagination.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `candidates`: GET connector-managed request path - records path `data`; incremental cursor
  `updatedAt`; formatted as `rfc3339`; records at or before the lower bound are filtered
  client-side.
- `jobs`: GET connector-managed request path - records path `data`; incremental cursor `updatedAt`;
  formatted as `rfc3339`; records at or before the lower bound are filtered client-side.
- `applications`: GET connector-managed request path - records path `data`; incremental cursor
  `updatedAt`; formatted as `rfc3339`; records at or before the lower bound are filtered
  client-side.
- `users`: GET connector-managed request path - records path `data`; incremental cursor `updatedAt`;
  formatted as `rfc3339`; records at or before the lower bound are filtered client-side.

## Write actions & risks

This connector is read-only; no reverse-ETL write actions are declared.

## Known limits

- API coverage includes 4 stream-backed endpoint group(s).
- Client-side incremental filtering is used for: `candidates`, `jobs`, `applications`, `users`.
