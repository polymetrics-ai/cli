# Overview

Reads 100ms rooms, sessions, recordings, templates, live streams, external streams, recording
assets, and webhook events, and writes room/template/room-code/recording lifecycle mutations,
through the 100ms server-side REST API.

Readable streams: `rooms`, `sessions`, `recordings`, `templates`, `live_streams`,
`external_streams`, `recording_assets`, `webhook_events`.

Write actions: `create_room`, `update_room`, `create_template`, `create_room_code`,
`update_room_code`, `start_recording`, `stop_recording`.

Service API documentation: https://www.100ms.live/docs/server-side/v2/api-reference.

## Auth setup

Connection fields:

- `base_url` (optional, string); default `https://api.100ms.live/v2`; format `uri`; 100ms API base
  URL override for tests or proxies.
- `management_token` (required, secret, string); 100ms management token. Used only for Bearer auth;
  never logged.
- `mode` (optional, string).

Secret fields are redacted in logs and write previews: `management_token`.

Default configuration values: `base_url=https://api.100ms.live/v2`.

Authentication behavior:

- Bearer token authentication using `secrets.management_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/rooms`.

## Streams notes

Default pagination: cursor pagination; cursor parameter `start`; next token from `last`.

Pagination by stream: cursor: `rooms`, `sessions`, `recordings`, `templates`, `live_streams`,
`external_streams`, `recording_assets`; page_number: `webhook_events`.

- `rooms`: GET `/rooms` - records path `data`; query `limit`=`100`; cursor pagination; cursor
  parameter `start`; next token from `last`.
- `sessions`: GET `/sessions` - records path `data`; query `limit`=`100`; cursor pagination; cursor
  parameter `start`; next token from `last`.
- `recordings`: GET `/recordings` - records path `data`; query `limit`=`100`; cursor pagination;
  cursor parameter `start`; next token from `last`.
- `templates`: GET `/templates` - records path `data`; query `limit`=`100`; cursor pagination;
  cursor parameter `start`; next token from `last`.
- `live_streams`: GET `/live-streams` - records path `data`; query `limit`=`100`; cursor pagination;
  cursor parameter `start`; next token from `last`.
- `external_streams`: GET `/external-streams` - records path `data`; query `limit`=`100`; cursor
  pagination; cursor parameter `start`; next token from `last`.
- `recording_assets`: GET `/recording-assets` - records path `data`; query `limit`=`100`; cursor
  pagination; cursor parameter `start`; next token from `last`.
- `webhook_events`: GET `/analytics/webhooks` - records path `events`; page-number pagination; page
  parameter `page`; size parameter `limit`; starts at 1; page size 100.

## Write actions & risks

Overall write risk: external 100ms mutation: creates/updates rooms, creates templates,
creates/updates room join-codes, and starts/stops room recordings; approval required.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_room`: POST `/rooms` - kind `create`; body type `json`; accepted fields `description`,
  `large_room`, `max_duration_seconds`, `name`, `region`, `template_id`; risk: creates a new 100ms
  room, or upserts an existing room's template if the same name is reused (100ms's own documented
  create-with-existing-name behavior); external mutation, approval required.
- `update_room`: POST `/rooms/{{ record.id }}` - kind `update`; body type `json`; path fields `id`;
  required record fields `id`; accepted fields `description`, `enabled`, `id`,
  `max_duration_seconds`, `name`, `region`; risk: mutates an existing room's metadata, or
  disables/re-enables it via the enabled field (100ms's disable/enable API is the same POST
  /rooms/{id} endpoint); disabling blocks all future joins to that room. External mutation, approval
  required.
- `create_template`: POST `/templates` - kind `create`; body type `json`; required record fields
  `name`; accepted fields `name`, `roles`, `settings`; risk: creates a new room-policy template
  (roles/settings); external mutation, approval required.
- `create_room_code`: POST `/room-codes/room/{{ record.room_id }}` - kind `create`; body type
  `none`; path fields `room_id`; required record fields `room_id`; accepted fields `room_id`; risk:
  generates join-authentication room codes for every role in the named room; codes act as join
  credentials, external mutation, approval required.
- `update_room_code`: POST `/room-codes/code` - kind `update`; body type `json`; required record
  fields `code`, `enabled`; accepted fields `code`, `enabled`; risk: enables or disables a specific
  join-credential room code; disabling revokes that code's ability to join. External mutation,
  approval required.
- `start_recording`: POST `/recordings/room/{{ record.room_id }}/start` - kind `create`; body type
  `json`; path fields `room_id`; body fields `meeting_url`, `resolution`; required record fields
  `room_id`; accepted fields `meeting_url`, `resolution`, `room_id`; risk: starts a composite
  recording job for the named room; consumes recording/storage quota. External mutation, approval
  required.
- `stop_recording`: POST `/recordings/room/{{ record.room_id }}/stop` - kind `update`; body type
  `none`; path fields `room_id`; required record fields `room_id`; accepted fields `room_id`; risk:
  stops all recording jobs currently running in the named room; external mutation, approval
  required.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 8 stream-backed endpoint group(s), 7 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=1, destructive_admin=6, duplicate_of=13, non_data_endpoint=4, out_of_scope=19,
  requires_elevated_scope=3.
