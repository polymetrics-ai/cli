# Overview

Airtable reads bases, tables, records, webhooks, and record comments, and writes record/table/
field/comment/webhook mutations, through the Airtable Web API (`https://api.airtable.com/v0`). This
bundle originally migrated `internal/connectors/airtable` (the hand-written connector, itself
read-only) to a declarative defs bundle at capability parity; Pass B expands it to the full
documented Airtable surface (airtable.com/developers/web/api/). The legacy package stays registered
and unchanged until wave6's registry flip, so this bundle's write surface and 2 new streams are a
genuine capability expansion beyond legacy, not a parity port.

## Auth setup

Provide either `api_key` (a Personal Access Token) or `access_token` (an OAuth2 access token) as a
secret; both are sent as a Bearer token (`Authorization: Bearer <token>`) and never logged. When
both are configured, `api_key` takes precedence — this matches legacy `airtableSecret`'s lookup
order (`credentials.api_key` checked before `credentials.access_token`).

## Streams notes

Five streams:

- `bases` — `GET /meta/bases`, records at `bases`, no config inputs required.
- `tables` — `GET /meta/bases/{base_id}/tables`, records at `tables`, requires config `base_id`.
  Airtable's per-table `fields` array is embedded directly in each table object (get-base-schema IS
  this same endpoint) — there is no separate list-fields GET endpoint, so `fields` is not modeled as
  its own stream.
- `records` — `GET /{base_id}/{table_id}`, records at `records`, requires config `base_id` and
  `table_id`; sends `pageSize` (default 100, matching legacy `airtableDefaultPageSize`).
- `webhooks` (Pass B addition) — `GET /bases/{base_id}/webhooks`, records at `webhooks`. Airtable's
  webhook list has no pagination fields at all (no `offset`, no cursor); the base `cursor`/`offset`
  pagination spec still applies uniformly to every stream, but simply finds no token and stops after
  one page — the correct behavior for an unpaginated endpoint, no stream-level override needed.
- `comments` (Pass B addition) — Airtable's comments endpoint
  (`GET /{base_id}/{table_id}/{record_id}/comments`) is a per-record sub-resource, not a bulk list;
  this stream uses the engine's `fan_out` dialect (`ids_from.request`) to first list every record id
  in the configured table (reusing the exact same `/{base_id}/{table_id}` request the `records`
  stream itself makes, paginated to exhaustion via the base's own `cursor`/`offset` spec — note this
  preliminary id-listing request does NOT carry the `records` stream's own `pageSize` query
  override, since `fan_out.ids_from.request` declares its own path with no query block and the
  engine reuses only the pagination SPEC, not the sibling stream's declared `query`), then re-runs
  the full per-id `GET .../{{ fanout.id }}/comments` request once per record, stamping the source
  record id onto every emitted comment's `record_id` field via `stamp_field`. This is real,
  practical coverage of a genuinely documented resource, but is O(records) requests per sync — a
  large table means one comments-list request per record on every `comments` stream read.

Pagination is Airtable's body-offset convention (`pagination.type: cursor` with `token_path: offset`,
`cursor_param: offset`): the next page is requested with `?offset=<value>` when the previous
response's top-level `offset` string is present, and pagination stops when `offset` is absent —
identical to legacy `harvest`'s loop, which has no `stop_path`-style secondary stop signal (Airtable's
`offset` absence is itself the sole stop condition).

Airtable has no incremental/cursor-field concept in this connector (legacy declares no
`CursorFields`); all streams sync full-refresh only, matching legacy's `Stream` definitions (no
`incremental` block declared) — `webhooks`/`comments` follow the same truth table (§8): neither
Airtable resource publishes a filterable `updatedTime`-shaped field this dialect's `incremental`
block could target with a server-side filter, so neither declares one.

## Write actions & risks

Pass B flips `capabilities.write` to `true` (a genuine capability expansion beyond legacy, which was
fully read-only). 12 actions in `writes.json`:

- **Records**: `create_record` (`POST /{base_id}/{table_id}`, body `{fields: {...}}`),
  `update_record` (`PATCH /{base_id}/{table_id}/{{ record.id }}`, non-destructive — only submitted
  fields change), `delete_record` (`DELETE /{base_id}/{table_id}/{{ record.id }}`). The bulk
  multi-record `PATCH`/`POST`/`DELETE` forms (up to 10 records per call, `records[]` array/query
  params) are not separately modeled — the engine's write path is one-request-per-record already
  (conventions.md §3), so the single-record forms cover the same outcome. The destructive `PUT`
  forms (clear every unincluded cell) are excluded as `destructive_admin` in favor of the
  non-destructive `PATCH`.
- **Tables/fields** (schema mutations, visible to every base collaborator): `create_table`/
  `update_table` (`POST`/`PATCH /meta/bases/{base_id}/tables(/:id)`), `create_field`/`update_field`
  (`POST`/`PATCH /meta/bases/{base_id}/tables/{{ record.table_id }}/fields(/:id)` — `table_id` is a
  `path_fields`-excluded record field, not a `config.*` reference, since a caller may target any
  table in the base, not only the `spec.json`-configured default `table_id`).
- **Comments**: `create_comment`/`update_comment`/`delete_comment`
  (`POST`/`PATCH`/`DELETE /{base_id}/{table_id}/{{ record.record_id }}/comments(/:id)`).
- **Webhooks**: `create_webhook` (`POST /bases/{base_id}/webhooks`, body `{notificationUrl,
  specification}`), `delete_webhook` (`DELETE /bases/{base_id}/webhooks/{{ record.id }}`) —
  `create_webhook` registers a live outbound HTTP callback; verify the target `notificationUrl`
  before enabling. Webhook update has no dedicated PATCH endpoint in Airtable's API (only enable/
  refresh operational actions, excluded per `api_surface.json` as `out_of_scope`); a webhook's
  `notificationUrl`/`specification` can only be changed by delete-then-recreate.

Every action uses `body_type: "json"` (default JSON body construction, `path_fields` excluding the
id(s) already in the path) except the three bodyless deletes (`delete_record`/`delete_comment`/
`delete_webhook`, `body_type: "none"`, no `body_fields` declared, so no body is sent — a pure
path-parameterized DELETE). No action needs a hook: every one of these operations is a single
JSON/bodyless HTTP request with no signature auth, multipart body, or compound follow-up call.

## Known limits

- **Base creation is out of scope.** `POST /v0/meta/bases` requires a `workspaceId` this connector's
  `spec.json` has no config surface for (the bundle is scoped to an already-existing `base_id`); see
  `api_surface.json`'s `requires_elevated_scope` entry.
- **Bulk multi-record write forms (up to 10 records/request) are not modeled.** The engine's write
  path is one-request-per-record by design (conventions.md §3); the single-record `create_record`/
  `update_record`/`delete_record` actions cover the same reachable outcome, just at a 1:1 HTTP-request
  ratio rather than Airtable's up-to-10-per-call batching. A caller writing many records issues many
  requests, subject to Airtable's per-base rate limit (5 req/sec) — no different in kind from any
  other bundle's per-record write semantics.
- **`comments` stream cost scales with table size.** See Streams notes above: one comments-list
  request per record, every read. This is real, documented Airtable behavior (no bulk
  "all comments across all records" endpoint exists) — not a bundle-authoring shortcut.
- **Webhook update is not modeled** (see Write actions & risks): Airtable's API has no
  update-webhook-configuration endpoint; only create/delete/enableNotifications/refresh exist, and
  the latter two are `out_of_scope` operational actions, not configuration mutations.
- `base_id`/`table_id` are plain (non-secret) config values, matching legacy's `configID` validation
  (required, trimmed, no format constraint beyond non-empty).
