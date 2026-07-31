# Overview

Notion connector for the official Notion API. The bundle reads users, pages, blocks, data sources,
databases, comments, views, meeting-note query results, search results, page markdown, async tasks,
custom emojis, and file-upload metadata. It also declares typed reverse-ETL actions for supportable
Notion page, block, data-source, database, comment, view, view-query, meeting-note, and file-upload
lifecycle mutations.

Official source audited for this wave: `https://developers.notion.com/openapi.json` (OpenAPI 3.1,
49 official HTTP operations, last-modified `Thu, 30 Jul 2026 23:41:58 GMT`, sha256
`4170a025f155ab721aa2e30451d1143ae79ad069b784b98d4fa420855f5d9d86`). OAuth token lifecycle
endpoints are classified as excluded/not-applicable; they are not connector data operations and no
auth-scope change is approved in this wave.

## Auth setup

Connection fields:

- `token` (required secret string): Notion integration token. Add it from an environment variable or
  stdin; never paste secret values into chat, docs, shell history, or JSON fixtures.
- `base_url` (optional URI): defaults to `https://api.notion.com/v1`.
- `page_size` (optional string): Notion cursor page size, 1-100, default `100`.
- `max_pages` (optional string): hard cap for hook-driven reads. Empty, `all`, `unlimited`, malformed,
  or non-positive values mean unbounded.
- Identifier fields such as `page_id`, `block_id`, `data_source_id`, `database_id`, `comment_id`,
  `view_id`, `query_id`, `file_upload_id`, `user_id`, `property_id`, and `task_id` scope the
  corresponding typed streams or write actions.

Requests send `Authorization: Bearer <token>` through the engine auth layer and `Notion-Version:
2022-06-28`. Connection checks call `GET /users`.

## Streams notes

Notion list responses use `{ results, next_cursor, has_more }`. The Notion StreamHook handles both
query-cursor GET endpoints and JSON-body cursor POST endpoints (notably `/search`, data-source
queries, and meeting-note queries), because Notion carries `start_cursor` and filters in request
bodies for several read operations.

Executable fixture-backed streams:

- Search variants: `databases`, `pages`, `search_results`.
- Users: `users`, `bot_user`, `user`.
- Pages: `page`, `page_property_items`, `page_markdown`.
- Async/block data: `async_task`, `block`, `block_children`.
- Data sources/databases: `data_source`, `data_source_query`, `data_source_templates`, `database`.
- Comments: `comments`, `comment`.
- Files/custom UI: `file_uploads`, `file_upload`, `custom_emojis`, `views`, `view`,
  `view_query_results`, `meeting_notes_query`.

The schemas allow Notion's polymorphic property/block payloads while retaining typed IDs and cursor
fields where the API exposes them. Fixture pages are sanitized and do not require live credentials.

## Write actions & risks

Reverse ETL actions are typed, schema-closed at the top level, and executed only through the existing
plan → preview → explicit approval → execute path. Destructive/archive/delete-like actions declare
`confirm: "destructive"`, path ID redaction, and idempotent 404 handling for fixture-backed replay.

Executable fixture-backed actions:

- Pages: `create_page`, `update_page`, `move_page`, `update_page_markdown`.
- Blocks: `update_block`, `delete_block`, `append_block_children`.
- Data sources/databases: `create_data_source`, `update_data_source`, `create_database`,
  `update_database`.
- Comments: `create_comment`, `update_comment`, `delete_comment`.
- Views/query resources: `create_view`, `update_view`, `delete_view`, `create_view_query`,
  `delete_view_query`.
- Meeting notes: `create_meeting_note`.
- File upload lifecycle: `create_file_upload`, `complete_file_upload`.

The official multipart `upload_file` byte-transfer endpoint remains blocked/planned in the operation
ledger until the shared binary payload approval/conformance runner can provide approved digests and
live-safe redacted artifacts without broad local file exposure. No live upload is performed by
conformance.

## Known limits

- This wave is fixture-only and uncertified. `certification.json` records that no live-safe Notion
  certification artifacts are present.
- The official multipart file byte-transfer endpoint is blocked/planned behind the shared binary
  payload approval/conformance runner. The JSON file-upload create/list/retrieve/complete lifecycle
  surfaces are fixture-backed.
- The official OAuth token, revoke, and introspection endpoints are classified as
  excluded/not-applicable (`disallowed` operation rows) because they are auth lifecycle operations,
  not connector data operations, and would require an approved auth-flow/scope foundation.
- The API surface has 49 official OpenAPI operations plus two connector-surface rows preserving the
  existing `databases` and `pages` object-filtered `/search` streams. Planning/count artifacts count
  official operations once; connector conformance covers both legacy stream variants with fixtures.
- Identifier-scoped streams require the corresponding config ID. Missing IDs fail closed during path
  interpolation rather than issuing broad or generic API calls.
