# Overview

Reads documents, folders, groups, and annotations from the Mendeley reference manager REST API.

Readable streams: `documents`, `folders`, `groups`, `annotations`.

This connector is read-only; no write actions are declared.

Service API documentation: https://dev.mendeley.com/.

## Auth setup

Connection fields:

- `base_url` (optional, string).
- `client_id` (required, secret, string); Could be found at `https://dev.mendeley.com/myapps.html`.
- `client_refresh_token` (required, secret, string); Use cURL or Postman with the OAuth 2.0
  Authorization tab. Set the Auth URL to https://api.mendeley.com/oauth/authorize, the Token URL to
  https://api.mendeley.com/oauth/token, and use all as the scope.
- `client_secret` (required, secret, string); Could be found at
  `https://dev.mendeley.com/myapps.html`.
- `mode` (optional, string).
- `name_for_institution` (required, string); The name parameter for institutions search.
- `query_for_catalog` (required, string); Query for catalog search.
- `start_date` (required, string).

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

- `documents`: GET connector-managed request path - records path `data`; incremental cursor
  `last_modified`; formatted as `rfc3339`; records at or before the lower bound are filtered
  client-side.
- `folders`: GET connector-managed request path - records path `data`; incremental cursor
  `modified`; formatted as `rfc3339`; records at or before the lower bound are filtered client-side.
- `groups`: GET connector-managed request path - records path `data`.
- `annotations`: GET connector-managed request path - records path `data`; incremental cursor
  `last_modified`; formatted as `rfc3339`; records at or before the lower bound are filtered
  client-side.

## Write actions & risks

This connector is read-only; no reverse-ETL write actions are declared.

## Known limits

- API coverage includes 4 stream-backed endpoint group(s).
- Client-side incremental filtering is used for: `documents`, `folders`, `annotations`.
