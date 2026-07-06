# Overview

Reads Canny boards, posts, comments, categories, and companies through the Canny REST API.

Readable streams: `boards`, `posts`, `comments`, `categories`, `companies`.

This connector is read-only; no write actions are declared.

Service API documentation: https://developers.canny.io/api-reference.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); You can find your secret API key in Your Canny Subdomain >
  Settings > API.
- `base_url` (optional, string).
- `mode` (optional, string).

Secret fields are redacted in logs and write previews: `api_key`.

Provide the secret fields listed above. Authentication is applied by the connector-specific
implementation for this service.

Requests use the configured `base_url` value after applying defaults.

Connection checks use a connector-managed request.

## Streams notes

Default pagination: single request; no pagination.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `boards`: GET connector-managed request path - records path `data`; incremental cursor `created`;
  formatted as `rfc3339`; records at or before the lower bound are filtered client-side.
- `posts`: GET connector-managed request path - records path `data`; incremental cursor `created`;
  formatted as `rfc3339`; records at or before the lower bound are filtered client-side.
- `comments`: GET connector-managed request path - records path `data`; incremental cursor
  `created`; formatted as `rfc3339`; records at or before the lower bound are filtered client-side.
- `categories`: GET connector-managed request path - records path `data`; incremental cursor
  `created`; formatted as `rfc3339`; records at or before the lower bound are filtered client-side.
- `companies`: GET connector-managed request path - records path `data`; incremental cursor
  `created`; formatted as `rfc3339`; records at or before the lower bound are filtered client-side.

## Write actions & risks

This connector is read-only; no reverse-ETL write actions are declared.

## Known limits

- API coverage includes 5 stream-backed endpoint group(s).
- Client-side incremental filtering is used for: `boards`, `posts`, `comments`, `categories`,
  `companies`.
