# Overview

Reads Google Calendar calendar lists, events, settings, and access control rules through the
Calendar API v3 using an OAuth2 refresh token.

Readable streams: `calendar_list`, `events`, `settings`, `acl`.

This connector is read-only; no write actions are declared.

Service API documentation: https://developers.google.com/calendar/api/v3/reference.

## Auth setup

Connection fields:

- `base_url` (optional, string).
- `calendarid` (required, string).
- `client_id` (required, secret, string).
- `client_refresh_token_2` (required, secret, string).
- `client_secret` (required, secret, string).
- `mode` (optional, string).

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

- `calendar_list`: GET connector-managed request path - records path `data`.
- `events`: GET connector-managed request path - records path `data`; incremental cursor `updated`;
  formatted as `rfc3339`; records at or before the lower bound are filtered client-side.
- `settings`: GET connector-managed request path - records path `data`.
- `acl`: GET connector-managed request path - records path `data`.

## Write actions & risks

This connector is read-only; no reverse-ETL write actions are declared.

## Known limits

- API coverage includes 4 stream-backed endpoint group(s).
- Client-side incremental filtering is used for: `events`.
