# Overview

Reads Insightful workforce-analytics employees, teams, projects, and directory entries through the
Insightful REST API.

Readable streams: `employee`, `team`, `projects`, `directory`.

This connector is read-only; no write actions are declared.

Service API documentation: https://developers.insightful.io/.

## Auth setup

Connection fields:

- `api_token` (required, secret, string); Insightful API token. Used only for Bearer auth; never
  logged.
- `base_url` (optional, string); default `https://app.insightful.io/api/v1`; format `uri`;
  Insightful API base URL override for tests or proxies.
- `mode` (optional, string).

Secret fields are redacted in logs and write previews: `api_token`.

Default configuration values: `base_url=https://app.insightful.io/api/v1`.

Authentication behavior:

- Bearer token authentication using `secrets.api_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/team`.

## Streams notes

Default pagination: single request; no pagination.

Pagination by stream: cursor: `employee`, `projects`, `directory`; none: `team`.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `employee`: GET `/employee` - records path `data`; cursor pagination; cursor parameter `next`;
  next token from `next`; incremental cursor `updatedAt`; sent as `start`; formatted as `rfc3339`.
- `team`: GET `/team` - records at response root.
- `projects`: GET `/project` - records path `data`; cursor pagination; cursor parameter `next`; next
  token from `next`; incremental cursor `updatedAt`; sent as `start`; formatted as `rfc3339`.
- `directory`: GET `/directory` - records path `data`; cursor pagination; cursor parameter `next`;
  next token from `next`; incremental cursor `updatedAt`; sent as `start`; formatted as `rfc3339`.

## Write actions & risks

This connector is read-only. Read behavior: external Insightful API read of workforce-analytics
employees, teams, projects, and directory entries.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 4 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=1, destructive_admin=2, duplicate_of=1, out_of_scope=3, requires_elevated_scope=2.
