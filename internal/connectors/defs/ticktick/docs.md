# Overview

Reads projects and project tasks, and writes task create/complete/delete actions, through the
TickTick Open API.

Readable streams: `projects`, `tasks`.

Write actions: `create_task`, `complete_task`, `delete_task`.

Service API documentation: https://developer.ticktick.com/api.

## Auth setup

Connection fields:

- `access_token` (optional, secret, string); Fallback TickTick OAuth access token, used only when
  neither bearer_token nor client_access_token is configured. Sent as a Bearer token. Never logged.
- `base_url` (optional, string); default `https://api.ticktick.com/open/v1`; format `uri`; TickTick
  Open API base URL override for tests or proxies.
- `bearer_token` (optional, secret, string); TickTick OAuth access token, sent as a Bearer token
  (Authorization: Bearer <token>). Never logged. Takes precedence over client_access_token and
  access_token when more than one is configured.
- `client_access_token` (optional, secret, string); Fallback TickTick OAuth access token, used only
  when bearer_token is not configured. Sent as a Bearer token. Never logged. Takes precedence over
  access_token when both are configured.
- `project_id` (optional, string); TickTick project id the 'tasks' stream reads (GET
  /project/{project_id}/data). Required for the 'tasks' stream.

Secret fields are redacted in logs and write previews: `access_token`, `bearer_token`,
`client_access_token`.

Default configuration values: `base_url=https://api.ticktick.com/open/v1`.

Authentication behavior:

- Bearer token authentication using `secrets.bearer_token` when `{{ secrets.bearer_token }}`.
- Bearer token authentication using `secrets.client_access_token` when `{{
  secrets.client_access_token }}`.
- Bearer token authentication using `secrets.access_token` when `{{ secrets.access_token }}`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/project`.

## Streams notes

Default pagination: single request; no pagination.

- `projects`: GET `/project` - records path `.`; emits passthrough records.
- `tasks`: GET `/project/{{ config.project_id }}/data` - records path `tasks`; emits passthrough
  records.

## Write actions & risks

Overall write risk: external TickTick API mutation: creates a task, marks a task complete, or
permanently deletes a task.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_task`: POST `/task` - kind `create`; body type `json`; required record fields `title`;
  accepted fields `content`, `desc`, `dueDate`, `isAllDay`, `priority`, `projectId`, `reminders`,
  `repeatFlag`, `sortOrder`, `startDate`, `timeZone`, `title`; risk: creates a new task in the
  caller's TickTick account (in the given projectId, or the default Inbox if omitted); low-risk
  external mutation, no approval required.
- `complete_task`: POST `/project/{{ record.projectId }}/task/{{ record.id }}/complete` - kind
  `update`; body type `none`; path fields `projectId`, `id`; required record fields `projectId`,
  `id`; accepted fields `id`, `projectId`; risk: marks an existing task as completed; a completed
  task is removed from active task lists/reminders for every collaborator on the project.
- `delete_task`: DELETE `/project/{{ record.projectId }}/task/{{ record.id }}` - kind `delete`; body
  type `none`; path fields `projectId`, `id`; required record fields `projectId`, `id`; accepted
  fields `id`, `projectId`; missing records treated as success for status `404`; risk: permanently
  removes a task from the given project; irreversible, no undo via the API.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 2 stream-backed endpoint group(s), 3 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  duplicate_of=2, out_of_scope=4, requires_elevated_scope=3.
