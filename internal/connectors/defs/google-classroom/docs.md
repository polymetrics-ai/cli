# Overview

Reads Google Classroom courses, teachers, students, course work, and announcements through the
Classroom REST API using an OAuth2 refresh token.

Readable streams: `courses`, `teachers`, `students`, `course_work`, `announcements`.

This connector is read-only; no write actions are declared.

Service API documentation: https://developers.google.com/classroom/reference/rest.

## Auth setup

Connection fields:

- `base_url` (optional, string).
- `client_id` (required, secret, string).
- `client_refresh_token` (required, secret, string).
- `client_secret` (required, secret, string).
- `mode` (optional, string).

Secret fields are redacted in logs and write previews: `client_id`, `client_refresh_token`,
`client_secret`.

Provide the secret fields listed above. Authentication is applied by the connector-specific
implementation for this service.

Requests use the configured `base_url` value after applying defaults.

Connection checks use a connector-managed request.

## Streams notes

Default pagination: single request; no pagination.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `courses`: GET connector-managed request path - records path `data`; incremental cursor
  `updateTime`; formatted as `rfc3339`; records at or before the lower bound are filtered
  client-side.
- `teachers`: GET connector-managed request path - records path `data`.
- `students`: GET connector-managed request path - records path `data`.
- `course_work`: GET connector-managed request path - records path `data`; incremental cursor
  `updateTime`; formatted as `rfc3339`; records at or before the lower bound are filtered
  client-side.
- `announcements`: GET connector-managed request path - records path `data`; incremental cursor
  `updateTime`; formatted as `rfc3339`; records at or before the lower bound are filtered
  client-side.

## Write actions & risks

This connector is read-only; no reverse-ETL write actions are declared.

## Known limits

- API coverage includes 5 stream-backed endpoint group(s).
- Client-side incremental filtering is used for: `courses`, `course_work`, `announcements`.
