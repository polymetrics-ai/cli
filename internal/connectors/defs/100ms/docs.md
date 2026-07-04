# Overview

100ms reads 100ms rooms, sessions, recordings, templates, live streams, external (RTMP push)
streams, recording assets, and webhook delivery events through the 100ms server-side REST API
(`https://api.100ms.live/v2/...`), and writes room/template/room-code/recording lifecycle mutations
back. This bundle originally targeted capability parity with `internal/connectors/100ms` (package
`onehms`, the hand-written connector it migrates; the legacy package stays registered and unchanged
until wave6's registry flip) and was expanded in Pass B to cover the full documented server-side v2
REST surface (`api_surface.json`).

## Auth setup

Provide a 100ms management token via the `management_token` secret; it is used only for Bearer
auth (`Authorization: Bearer <management_token>`) and is never logged, matching legacy's
`connsdk.Bearer(token)` (`onehms.go:245`).

## Streams notes

The 4 legacy-parity streams (`rooms`, `sessions`, `recordings`, `templates`) and the 3 new
Pass B streams that share 100ms's `data`/`last` list-envelope convention (`live_streams`,
`external_streams`, `recording_assets`) all follow the same shape: `GET` against the 100ms list
endpoint, records at `data`, primary key `["id"]`. Every 100ms object exposes a string `id` and an
RFC3339 `created_at`/`updated_at` pair; `created_at` is declared as the soft cursor field on every
schema that has one (matching legacy's `streams.go` catalog for the original 4), but **no stream
declares an `incremental` block** — legacy's `harvest` loop has no server-side incremental filter
parameter at all (`onehms.go`'s `Read`/`harvest` always walks the full result set every sync;
`InitialState` always seeds an empty cursor), so this bundle reproduces the identical
full-refresh-only behavior rather than inventing an incremental filter legacy never had.
`recording_assets` has no documented `created_at` field at all (100ms's own reference response
shows `id`/`job_id`/`room_id`/`session_id`/`type`/`path`/`duration`/`size`/`status` only), so its
schema declares no `x-cursor-field`.

Pagination for the shared-envelope streams follows 100ms's own `data`/`last` body-cursor convention
(`pagination.type: cursor` with `token_path: last`, `cursor_param: start`): the next page is
requested with `start=<last>`, and pagination stops when `last` is empty, the page returns no
records, or (the engine's `tokenPathCursor` loop guard) the same token repeats twice in a row —
this is an exact match for legacy's own three-way stop condition (`onehms.go:183`: `last == "" ||
len(records) == 0 || last == start`). No `stop_path` is declared: 100ms has no separate boolean
"has more" signal the way Stripe/Zendesk do, only the token's own emptiness. Every request sends
`limit=100` via each stream's static `query: {"limit": "100"}`.

The new `webhook_events` stream (`GET /analytics/webhooks`) is a genuinely different shape — 100ms's
own docs paginate it with `page`/`limit` query params and no `data`/`last` body-cursor envelope at
all (the records live at the top-level `events` array) — so it declares its own stream-level
`page_number` pagination override (`page_param: page`, `size_param: limit`, `start_page: 1`,
`page_size: 100`), which replaces the base cursor pagination wholesale for this one stream per the
engine's per-stream-override rule (§3). Its primary key is `event_id` and its cursor field is
`event_timestamp`; no `incremental` block is declared for the same full-refresh-only reasoning as
every other stream in this bundle.

`page_size`/`max_pages` are **not modeled as config-driven overrides** here even though legacy
exposes them (`onehms.go`'s `pageSize`/`maxPages` config helpers, defaults 100/0-unlimited): the
`cursor`+`token_path` paginator never reads `PaginationSpec.PageSize`, and `PaginationSpec.MaxPages`
is a fixed integer, not a templated field — there is no mechanism to wire a `config.*` value into
either at read time. This bundle bakes legacy's own defaults (`limit=100`, unbounded pages) as
fixed values instead.

## Write actions & risks

Pass B adds 7 write actions, all newly modeled (legacy shipped none — `onehms.go:250-255`'s "no
safe reverse-ETL surface" applied only to the original migration's scope, not to the full documented
API):

- `create_room` (`POST /rooms`) — creates a room, or (100ms's own documented behavior) re-templates
  an existing room if the same `name` is reused. External mutation, approval required.
- `update_room` (`POST /rooms/{id}`) — updates room metadata; the SAME endpoint also disables/
  re-enables a room via its `enabled` field (100ms's "Disable/Enable a room" doc page is this exact
  endpoint, not a separate one), so both legacy-documented actions are modeled as one write action.
  Disabling blocks all future joins to that room. Approval required.
- `create_template` (`POST /templates`) — creates a new room-policy template. Approval required.
- `create_room_code` (`POST /room-codes/room/{room_id}`) — generates join-authentication room codes
  for every role in a room. Room codes are join credentials; approval required.
- `update_room_code` (`POST /room-codes/code`) — enables/disables a specific room code; disabling
  revokes that code's ability to join. Approval required.
- `start_recording` (`POST /recordings/room/{room_id}/start`) — starts a composite recording job;
  consumes recording/storage quota. Approval required.
- `stop_recording` (`POST /recordings/room/{room_id}/stop`) — stops all recording jobs running in a
  room. Approval required.

`capabilities.write` is now `true`.

## Known limits

- Real-time/ephemeral session-state endpoints (`active-rooms` peer list/get/update/remove/message,
  end-room) are excluded as `non_data_endpoint`/`destructive_admin` — they describe or mutate a
  live, in-progress call rather than syncable catalog data; see `api_surface.json`.
- Room-codes are not modeled as a read stream (`requires_elevated_scope`): listing them at scale
  would fan out over every room and land active join-authentication secrets in warehouse-destined
  record data.
- Poll endpoints are entirely out of scope: 100ms's API exposes no poll-listing endpoint (only
  get-by-id and room-linkage), so no catalog stream can be built without a poll-id source; poll
  creation/response reads are deferred alongside it.
- `analytics/events` (raw track/recording/error telemetry) and `analytics/peer-stats` (call-quality
  time series) are excluded as `out_of_scope`: high-cardinality diagnostic firehoses, not catalog
  data suited to a warehouse sync.
- Live-stream and external-stream (RTMP push) lifecycle mutations (start/stop/pause/resume/
  timed-metadata) are deferred `out_of_scope` pending a dedicated Pass B write-review pass for
  streaming-control actions; only their list-read surface is covered here.
- Full template update (`POST /templates/{id}` and its roles/settings/destinations sub-resources) is
  out of scope; only template creation is modeled, matching the create-first parity most connectors
  of this shape ship.
- RTMP stream-key issuance/retrieval/disable is excluded as `requires_elevated_scope`
  (credential-shaped, not catalog data) or `destructive_admin` (disable revokes live ingest).
- `page_size`/`max_pages` are fixed (not config-driven); see Streams notes above for why the engine
  dialect cannot wire them for the `cursor`+`token_path` paginator shape without changing behavior.
- Legacy's fixture-mode-only fields (`onehms.go`'s `readFixture`, e.g. `previous_cursor` echoing a
  prior sync cursor) are not modeled — they only ever appeared in legacy's own credential-free
  fixture path, never in a live record; this bundle's schemas target the live record shape only,
  and the engine's own conformance/fixture-replay harness is the credential-free test affordance
  for this bundle.
