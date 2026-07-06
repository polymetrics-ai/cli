# Overview

Reads Dixa conversations (and their queue, rating, and assignment projections) from the Dixa
conversation_export API.

Readable streams: `conversations`, `conversation_queue`, `conversation_rating`,
`conversation_assignment`.

This connector is read-only; no write actions are declared.

Service API documentation: https://docs.dixa.io/openapi/.

## Auth setup

Connection fields:

- `api_token` (required, secret, string); Dixa API token.
- `base_url` (optional, string).
- `batch_size` (optional, string); Number of days to batch into one request. Max 31.
- `mode` (optional, string).
- `start_date` (required, string); The connector pulls records updated from this date onwards.

Secret fields are redacted in logs and write previews: `api_token`.

Provide the secret fields listed above. Authentication is applied by the connector-specific
implementation for this service.

Requests use the configured `base_url` value after applying defaults.

Connection checks use a connector-managed request.

## Streams notes

Default pagination: single request; no pagination.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `conversations`: GET connector-managed request path - records path `data`; incremental cursor
  `updated_at`; formatted as `rfc3339`; records at or before the lower bound are filtered
  client-side.
- `conversation_queue`: GET connector-managed request path - records path `data`; incremental cursor
  `updated_at`; formatted as `rfc3339`; records at or before the lower bound are filtered
  client-side.
- `conversation_rating`: GET connector-managed request path - records path `data`; incremental
  cursor `updated_at`; formatted as `rfc3339`; records at or before the lower bound are filtered
  client-side.
- `conversation_assignment`: GET connector-managed request path - records path `data`; incremental
  cursor `updated_at`; formatted as `rfc3339`; records at or before the lower bound are filtered
  client-side.

## Write actions & risks

This connector is read-only; no reverse-ETL write actions are declared.

## Known limits

- API coverage includes 4 stream-backed endpoint group(s).
- Client-side incremental filtering is used for: `conversations`, `conversation_queue`,
  `conversation_rating`, `conversation_assignment`.
