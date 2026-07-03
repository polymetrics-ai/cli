# Overview

Google Tasks reads the authenticated user's task lists, and every task across every task list,
through the Google Tasks API v1. This bundle is a full capability-parity migration of the legacy
hand-written connector (`internal/connectors/google-tasks`), which stays registered and unchanged
until wave6's registry flip. Read-only: legacy itself sets `Capabilities.Write = false` with no
reverse-ETL write path.

## Auth setup

Provide a Google OAuth access token via the `api_key` secret; it is used only for Bearer auth
(`Authorization: Bearer <api_key>`) and is never logged. `base_url` defaults to
`https://tasks.googleapis.com/tasks/v1` and may be overridden for tests/proxies (validated as an
absolute http/https URL with a host).

## Streams notes

`tasklists` (`GET /users/@me/lists`, `records.path: "items"`) lists every task list belonging to
the authenticated user; primary key `id`.

`tasks` is **task-list-scoped**: legacy's own read path (`readTasks`/`collectTaskListIDs`,
`internal/connectors/google-tasks/google_tasks.go:156-196`) first lists every task list, then reads
that list's tasks (`GET /lists/{tasklistId}/tasks`) once per list id, stamping `tasklist_id` onto
every emitted task. This bundle reproduces that exact pattern with the engine's `stream.fan_out`
dialect: `ids_from.request` issues a preliminary, fully-paginated `GET /users/@me/lists` listing
(the SAME endpoint the `tasklists` stream itself reads, extracting `id` off each record);
`into.path_var` makes the resolved task-list id referenceable in the stream's own `path` as
`{{ fanout.id }}` (`/lists/{{ fanout.id }}/tasks`); `stamp_field: tasklist_id` writes the current
task-list id onto every emitted record after projection, exactly matching legacy's manual
`rec["tasklist_id"] = listID` stamp. Pagination and page size are independent per task list (a
fresh cursor paginator + fresh `maxResults` query per list), mirroring legacy's own per-list
`harvest` call.

Both streams share the same pagination shape: Google's `nextPageToken`/`pageToken` cursor
convention (`pagination.type: cursor`, `cursor_param: pageToken`, `token_path: nextPageToken`), and
both request `maxResults` from the `records_limit` config value (default 50, matching legacy's
`googleTasksDefaultLimit`; legacy also caps it at 100 — `records_limit`'s own bounds are enforced
client-side in legacy but the engine has no per-config numeric-range validation primitive, so an
out-of-range value here is passed through to the live API rather than rejected pre-flight; the
upstream API itself will reject an out-of-bounds `maxResults`).

Both streams publish `updated` as their incremental cursor field (matching legacy's own catalog,
which sets `CursorFields: []string{"updated"}` on both `tasklists` and `tasks`) — but neither
stream sends a server-side `updatedMin`-style filter, since legacy's own read path never applies
one either (§8 rule 2 truth table: bare `cursor_field`, no `request_param`, since legacy publishes
the cursor field in its catalog but sends no server-side filter). Every sync is effectively full
refresh with a downstream-dedupable cursor column, exactly matching legacy's behavior.

`self_link` is a `computed_fields` rename from the raw API's camelCase `selfLink` (plain schema
projection copies by exact key match only). Every other published field name matches the raw
Google Tasks wire shape verbatim (`id`, `title`, `updated`, `parent`, `position`, `notes`,
`status`, `due`, `completed`, `deleted`, `hidden`), mirroring legacy's own `taskListRecord`/
`taskRecord` field selection.

## Write actions & risks

None. `capabilities.write` is `false`; no `writes.json` is shipped. Legacy itself implements no
write path for Google Tasks (`Write` is a stub returning `ErrUnsupportedOperation`).

## Known limits

- `records_limit`'s legacy-enforced bounds (1-100, `googleTasksMaxLimit`) are not statically
  validated by the engine dialect — `spec.json` declares it a plain `integer` with a default; an
  out-of-range value is passed straight to the live API rather than rejected client-side the way
  legacy's `googleTasksLimit` helper does. Not a data-parity issue (the API itself still rejects an
  invalid value), just a shifted validation boundary.
- `fixtures/streams/tasks/page_1.json` ships two fixture task-list ids' worth of preliminary
  `/users/@me/lists` listing data plus one page of `tasks` for the first list, to exercise the
  fan-out path under conformance's replay harness; see cisco-meraki's bundle for the same pattern
  applied to a different API.
