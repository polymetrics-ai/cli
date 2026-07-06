# Overview

Reads Lokalise project keys, languages, translations, contributors, and comments through the
Lokalise REST API.

Readable streams: `keys`, `languages`, `translations`, `contributors`, `comments`.

This connector is read-only; no write actions are declared.

Service API documentation: https://developers.lokalise.com/reference/api-introduction.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Lokalise API Key with read-access. Available at Profile
  settings > API tokens. See <a
  href="https://docs.lokalise.com/en/articles/1929556-api-tokens">here</a>.
- `base_url` (optional, string).
- `mode` (optional, string).
- `project_id` (required, string); Lokalise project ID. Available at Project Settings > General.

Secret fields are redacted in logs and write previews: `api_key`.

Provide the secret fields listed above. Authentication is applied by the connector-specific
implementation for this service.

Requests use the configured `base_url` value after applying defaults.

Connection checks use a connector-managed request.

## Streams notes

Default pagination: single request; no pagination.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `keys`: GET connector-managed request path - records path `data`; incremental cursor
  `modified_at_timestamp`; formatted as `rfc3339`; records at or before the lower bound are filtered
  client-side.
- `languages`: GET connector-managed request path - records path `data`.
- `translations`: GET connector-managed request path - records path `data`; incremental cursor
  `modified_at_timestamp`; formatted as `rfc3339`; records at or before the lower bound are filtered
  client-side.
- `contributors`: GET connector-managed request path - records path `data`.
- `comments`: GET connector-managed request path - records path `data`.

## Write actions & risks

This connector is read-only; no reverse-ETL write actions are declared.

## Known limits

- API coverage includes 5 stream-backed endpoint group(s).
- Client-side incremental filtering is used for: `keys`, `translations`.
