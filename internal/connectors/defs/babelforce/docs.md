# Overview

Reads Babelforce call reporting, recordings, numbers, and users through the Babelforce v2 REST API.

Readable streams: `calls`, `calls_extended`, `recordings`, `numbers`, `users`.

This connector is read-only; no write actions are declared.

Service API documentation: https://api.babelforce.com/.

## Auth setup

Connection fields:

- `access_key_id` (required, secret, string); The Babelforce access key ID.
- `access_token` (required, secret, string); The Babelforce access token.
- `base_url` (optional, string).
- `date_created_from` (optional, string); Timestamp in Unix the replication from Babelforce API will
  start from. For example 1651363200 which corresponds to 2022-05-01 00:00:00.
- `date_created_to` (optional, string); Timestamp in Unix the replication from Babelforce will be up
  to. For example 1651363200 which corresponds to 2022-05-01 00:00:00.
- `mode` (optional, string).
- `region` (required, string); Babelforce region.

Secret fields are redacted in logs and write previews: `access_key_id`, `access_token`.

Provide the secret fields listed above. Authentication is applied by the connector-specific
implementation for this service.

Requests use the configured `base_url` value after applying defaults.

Connection checks use a connector-managed request.

## Streams notes

Default pagination: single request; no pagination.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `calls`: GET connector-managed request path - records path `data`; incremental cursor
  `dateCreated`; formatted as `rfc3339`; records at or before the lower bound are filtered
  client-side.
- `calls_extended`: GET connector-managed request path - records path `data`; incremental cursor
  `dateCreated`; formatted as `rfc3339`; records at or before the lower bound are filtered
  client-side.
- `recordings`: GET connector-managed request path - records path `data`; incremental cursor
  `dateCreated`; formatted as `rfc3339`; records at or before the lower bound are filtered
  client-side.
- `numbers`: GET connector-managed request path - records path `data`; incremental cursor
  `dateCreated`; formatted as `rfc3339`; records at or before the lower bound are filtered
  client-side.
- `users`: GET connector-managed request path - records path `data`; incremental cursor
  `dateCreated`; formatted as `rfc3339`; records at or before the lower bound are filtered
  client-side.

## Write actions & risks

This connector is read-only; no reverse-ETL write actions are declared.

## Known limits

- API coverage includes 5 stream-backed endpoint group(s).
- Client-side incremental filtering is used for: `calls`, `calls_extended`, `recordings`, `numbers`,
  `users`.
