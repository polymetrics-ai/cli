# Overview

Airtable reads bases, tables, and records through the Airtable Web API (`https://api.airtable.com/v0`).
This bundle migrates `internal/connectors/airtable` (the hand-written connector) to a declarative
defs bundle at capability parity; the legacy package stays registered and unchanged until wave6's
registry flip. The connector is read-only — Airtable record writes are intentionally out of scope,
matching legacy.

## Auth setup

Provide either `api_key` (a Personal Access Token) or `access_token` (an OAuth2 access token) as a
secret; both are sent as a Bearer token (`Authorization: Bearer <token>`) and never logged. When
both are configured, `api_key` takes precedence — this matches legacy `airtableSecret`'s lookup
order (`credentials.api_key` checked before `credentials.access_token`).

## Streams notes

Three streams, matching legacy's `airtableStreamEndpoints` table exactly:

- `bases` — `GET /meta/bases`, records at `bases`, no config inputs required.
- `tables` — `GET /meta/bases/{base_id}/tables`, records at `tables`, requires config `base_id`.
- `records` — `GET /{base_id}/{table_id}`, records at `records`, requires config `base_id` and
  `table_id`; sends `pageSize` (default 100, matching legacy `airtableDefaultPageSize`) — the only
  stream that sends a page-size param, matching legacy's `endpoint.needsTable` gate.

Pagination is Airtable's body-offset convention (`pagination.type: cursor` with `token_path: offset`,
`cursor_param: offset`): the next page is requested with `?offset=<value>` when the previous
response's top-level `offset` string is present, and pagination stops when `offset` is absent —
identical to legacy `harvest`'s loop, which has no `stop_path`-style secondary stop signal (Airtable's
`offset` absence is itself the sole stop condition).

Airtable has no incremental/cursor-field concept in this connector (legacy declares no
`CursorFields`); all three streams sync full-refresh only, matching legacy's `Stream` definitions
(no `incremental` block declared).

## Write actions & risks

None. Airtable is `capabilities.write: false`; no `writes.json` is shipped, matching legacy's
`Write` always returning `connectors.ErrUnsupportedOperation`.

## Known limits

- Airtable record writes (`POST`/`PATCH`/`DELETE` on `/v0/{baseId}/{tableId}`), field metadata
  (`/v0/meta/bases/{baseId}/tables/{tableId}/fields`), and record comments are out of scope for this
  wave; see `api_surface.json`'s `excluded: {category: out_of_scope, reason: "Pass B capability
  expansion"}` entries. Only the 3 legacy-parity read streams are implemented.
- `base_id`/`table_id` are plain (non-secret) config values, matching legacy's `configID` validation
  (required, trimmed, no format constraint beyond non-empty).
