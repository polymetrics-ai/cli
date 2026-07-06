# Overview

Reads Convex tables and documents through the deployment HTTP API.

Readable streams: `tables`, `documents`.

This connector is read-only; no write actions are declared.

Service API documentation: https://docs.convex.dev/http-api/.

## Auth setup

Connection fields:

- `access_key` (required, secret, string); Convex deployment access key. Used only for Bearer auth
  (Authorization: Bearer <access_key>); never logged.
- `deployment_url` (required, string); format `uri`; URL of the Convex deployment to read from (e.g.
  https://my-deployment.convex.cloud).
- `mode` (optional, string).
- `table` (optional, string); default `data`; Convex table name to read documents from (the
  'documents' stream). Not used by the 'tables' stream.

Secret fields are redacted in logs and write previews: `access_key`.

Default configuration values: `table=data`.

Authentication behavior:

- Bearer token authentication using `secrets.access_key`.

Requests use the configured `deployment_url` value after applying defaults.

Connection checks call GET `api/tables`.

## Streams notes

Default pagination: single request; no pagination.

Pagination by stream: cursor: `documents`; none: `tables`.

- `tables`: GET `api/tables` - records path `tables`; emits passthrough records.
- `documents`: GET `api/tables/{{ config.table }}/documents` - records path `documents`; cursor
  pagination; cursor parameter `cursor`; next token from `cursor`; computed output fields `id`;
  emits passthrough records.

## Write actions & risks

This connector is read-only. Read behavior: external Convex deployment API read of table metadata
and documents.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 2 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=3.
