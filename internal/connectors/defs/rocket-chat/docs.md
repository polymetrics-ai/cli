# Overview

Reads Rocket.Chat users, public channels, private groups, direct messages, and rooms through the
REST API.

Readable streams: `users`, `channels`, `groups`, `direct_messages`, `rooms`.

This connector is read-only; no write actions are declared.

Service API documentation: https://developer.rocket.chat/apidocs.

## Auth setup

Connection fields:

- `auth_token` (required, secret, string); Rocket.Chat personal access token, sent as the
  X-Auth-Token header. Never logged.
- `base_url` (required, string); format `uri`; Rocket.Chat server base URL (e.g.
  https://chat.example.com). The /api/v1 suffix is appended automatically if not already present.
- `fields` (optional, string); Optional Rocket.Chat field-projection filter passed through as the
  'fields' parameter.
- `mode` (optional, string).
- `query` (optional, string); Optional raw Rocket.Chat query-string filter passed through to list
  endpoints as the 'query' parameter.
- `room_id` (optional, string); Optional room id filter passed through as the 'roomId' parameter
  (used by the rooms stream).
- `updated_since` (optional, string); Optional RFC3339 timestamp passed through as the
  'updatedSince' parameter (used by the rooms stream).
- `user_id` (required, secret, string); Rocket.Chat user id paired with auth_token, sent as the
  X-User-Id header. Never logged.

Secret fields are redacted in logs and write previews: `auth_token`, `user_id`.

Provide the secret fields listed above. Authentication is applied by the connector-specific
implementation for this service.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/me`.

## Streams notes

Default pagination: offset/limit pagination; offset parameter `offset`; limit parameter `count`;
page size 100.

- `users`: GET `/users.list` - records path `users`; query `fields` from template `{{ config.fields
  }}`, omitted when absent; `query` from template `{{ config.query }}`, omitted when absent;
  offset/limit pagination; offset parameter `offset`; limit parameter `count`; page size 100;
  computed output fields `id`, `stream`, `updated_at`; emits passthrough records.
- `channels`: GET `/channels.list` - records path `channels`; query `fields` from template `{{
  config.fields }}`, omitted when absent; `query` from template `{{ config.query }}`, omitted when
  absent; offset/limit pagination; offset parameter `offset`; limit parameter `count`; page size
  100; computed output fields `id`, `stream`, `updated_at`; emits passthrough records.
- `groups`: GET `/groups.list` - records path `groups`; query `fields` from template `{{
  config.fields }}`, omitted when absent; `query` from template `{{ config.query }}`, omitted when
  absent; offset/limit pagination; offset parameter `offset`; limit parameter `count`; page size
  100; computed output fields `id`, `stream`, `updated_at`; emits passthrough records.
- `direct_messages`: GET `/im.list` - records path `ims`; query `fields` from template `{{
  config.fields }}`, omitted when absent; `query` from template `{{ config.query }}`, omitted when
  absent; offset/limit pagination; offset parameter `offset`; limit parameter `count`; page size
  100; computed output fields `id`, `stream`, `updated_at`; emits passthrough records.
- `rooms`: GET `/rooms.get` - records path `update`; query `roomId` from template `{{ config.room_id
  }}`, omitted when absent; `updatedSince` from template `{{ config.updated_since }}`, omitted when
  absent; offset/limit pagination; offset parameter `offset`; limit parameter `count`; page size
  100; computed output fields `id`, `stream`, `type`, `updated_at`; emits passthrough records.

## Write actions & risks

This connector is read-only. Read behavior: external Rocket.Chat API read of workspace users, rooms,
and messages metadata.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 5 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  non_data_endpoint=1, out_of_scope=3, requires_elevated_scope=1.
