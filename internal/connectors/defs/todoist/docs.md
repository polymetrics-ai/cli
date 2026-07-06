# Overview

Reads projects, sections, tasks, comments, labels, and project collaborators, and writes
project/section/task/comment/label create, update, and delete actions (plus task close/reopen),
through the Todoist REST API.

Readable streams: `projects`, `sections`, `tasks`, `comments`, `labels`, `collaborators`.

Write actions: `create_project`, `update_project`, `delete_project`, `create_section`,
`update_section`, `delete_section`, `create_task`, `update_task`, `close_task`, `reopen_task`,
`delete_task`, `create_comment`, `update_comment`, `delete_comment`, `create_label`, `update_label`,
`delete_label`.

Service API documentation: https://developer.todoist.com/rest/v2/.

## Auth setup

Connection fields:

- `base_url` (optional, string); default `https://api.todoist.com/rest/v2`; format `uri`; Todoist
  REST API base URL override for tests or proxies.
- `bearer_token` (optional, secret, string); Fallback Todoist personal API token, used only when
  token is not configured. Sent as a Bearer token. Never logged.
- `project_id` (optional, string); Optional Todoist project id used to scope the 'comments' stream
  to a single project.
- `task_id` (optional, string); Optional Todoist task id used to scope the 'comments' stream to a
  single task.
- `token` (optional, secret, string); Todoist personal API token, sent as a Bearer token
  (Authorization: Bearer <token>). Never logged. Takes precedence over bearer_token when both are
  configured.

Secret fields are redacted in logs and write previews: `bearer_token`, `token`.

Default configuration values: `base_url=https://api.todoist.com/rest/v2`.

Authentication behavior:

- Bearer token authentication using `secrets.token` when `{{ secrets.token }}`.
- Bearer token authentication using `secrets.bearer_token` when `{{ secrets.bearer_token }}`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/projects`.

## Streams notes

Default pagination: single request; no pagination.

- `projects`: GET `/projects` - records path `.`; emits passthrough records.
- `sections`: GET `/sections` - records path `.`; emits passthrough records.
- `tasks`: GET `/tasks` - records path `.`; emits passthrough records.
- `comments`: GET `/comments` - records path `.`; query `project_id` from template `{{
  config.project_id }}`, omitted when absent; `task_id` from template `{{ config.task_id }}`,
  omitted when absent; emits passthrough records.
- `labels`: GET `/labels` - records path `.`; emits passthrough records.
- `collaborators`: GET `/projects/{{ fanout.id }}/collaborators` - records path `.`; fan-out; ids
  from request `/projects`; id field `id`; id inserted into the request path; stamps `project_id`;
  emits passthrough records.

## Write actions & risks

Overall write risk: external Todoist API mutation: creates, updates, or deletes
projects/sections/tasks/comments/labels, and closes/reopens tasks.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_project`: POST `/projects` - kind `create`; body type `json`; required record fields
  `name`; accepted fields `color`, `is_favorite`, `name`, `parent_id`, `view_style`; risk: creates a
  new project in the caller's Todoist account; low-risk external mutation, no approval required.
- `update_project`: POST `/projects/{{ record.id }}` - kind `update`; body type `json`; path fields
  `id`; required record fields `id`; accepted fields `color`, `id`, `is_favorite`, `name`,
  `view_style`; risk: mutates an existing project's name, color, favorite flag, or display style.
- `delete_project`: DELETE `/projects/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `404`; risk: permanently removes a project and everything in it (its sections, tasks,
  and comments); irreversible.
- `create_section`: POST `/sections` - kind `create`; body type `json`; required record fields
  `project_id`, `name`; accepted fields `name`, `order`, `project_id`; risk: creates a new section
  within an existing project; low-risk external mutation, no approval required.
- `update_section`: POST `/sections/{{ record.id }}` - kind `update`; body type `json`; path fields
  `id`; required record fields `id`, `name`; accepted fields `id`, `name`; risk: renames an existing
  section.
- `delete_section`: DELETE `/sections/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `404`; risk: permanently removes a section and every task in it; irreversible.
- `create_task`: POST `/tasks` - kind `create`; body type `json`; required record fields `content`;
  accepted fields `assignee_id`, `content`, `description`, `due_date`, `due_datetime`, `due_lang`,
  `due_string`, `duration`, `duration_unit`, `labels`, `order`, `parent_id`, `priority`,
  `project_id`, `section_id`; risk: creates a new task in the caller's Todoist account (in the given
  project, or Inbox if omitted); low-risk external mutation, no approval required.
- `update_task`: POST `/tasks/{{ record.id }}` - kind `update`; body type `json`; path fields `id`;
  required record fields `id`; accepted fields `assignee_id`, `content`, `description`, `due_date`,
  `due_datetime`, `due_lang`, `due_string`, `duration`, `duration_unit`, `id`, `labels`, `priority`;
  risk: mutates an existing task's content, description, labels, priority, due date, assignee, or
  duration.
- `close_task`: POST `/tasks/{{ record.id }}/close` - kind `update`; body type `none`; path fields
  `id`; required record fields `id`; accepted fields `id`; risk: marks an existing task as completed
  (mirrors clicking the checkbox in the Todoist UI); recurring tasks advance to their next
  occurrence instead of disappearing.
- `reopen_task`: POST `/tasks/{{ record.id }}/reopen` - kind `update`; body type `none`; path fields
  `id`; required record fields `id`; accepted fields `id`; risk: reopens a previously completed
  task, returning it to the active task list.
- `delete_task`: DELETE `/tasks/{{ record.id }}` - kind `delete`; body type `none`; path fields
  `id`; required record fields `id`; accepted fields `id`; missing records treated as success for
  status `404`; risk: permanently removes a task; irreversible.
- `create_comment`: POST `/comments` - kind `create`; body type `json`; required record fields
  `content`; accepted fields `attachment`, `content`, `project_id`, `task_id`; risk: posts a new
  comment on a task or project; low-risk external mutation, no approval required.
- `update_comment`: POST `/comments/{{ record.id }}` - kind `update`; body type `json`; path fields
  `id`; required record fields `id`, `content`; accepted fields `content`, `id`; risk: edits the
  content of an existing comment.
- `delete_comment`: DELETE `/comments/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `404`; risk: permanently removes a comment; irreversible.
- `create_label`: POST `/labels` - kind `create`; body type `json`; required record fields `name`;
  accepted fields `color`, `is_favorite`, `name`, `order`; risk: creates a new personal label;
  low-risk external mutation, no approval required.
- `update_label`: POST `/labels/{{ record.id }}` - kind `update`; body type `json`; path fields
  `id`; required record fields `id`; accepted fields `color`, `id`, `is_favorite`, `name`, `order`;
  risk: renames an existing label or changes its order/color/favorite flag; renaming changes how it
  appears on every task already using it.
- `delete_label`: DELETE `/labels/{{ record.id }}` - kind `delete`; body type `none`; path fields
  `id`; required record fields `id`; accepted fields `id`; missing records treated as success for
  status `404`; risk: permanently deletes a personal label; it is removed from every task that used
  it.

## Known limits

- API coverage includes 6 stream-backed endpoint group(s), 17 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  duplicate_of=5, non_data_endpoint=1.
