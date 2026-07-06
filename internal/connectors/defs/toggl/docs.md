# Overview

Reads and writes time entries, projects, clients, tags, tasks, and users through the Toggl Track
API.

Readable streams: `time_entries`, `projects`, `clients`, `workspace_users`, `organization_users`,
`tags`, `tasks`.

Write actions: `create_time_entry`, `update_time_entry`, `stop_time_entry`, `delete_time_entry`,
`create_project`, `update_project`, `delete_project`, `create_client`, `update_client`,
`delete_client`, `create_tag`, `update_tag`, `delete_tag`, `create_task`, `update_task`,
`delete_task`.

Service API documentation: https://developers.track.toggl.com/docs/.

## Auth setup

Connection fields:

- `api_token` (required, secret, string); Toggl Track API token, sent as HTTP Basic auth username
  with the literal string 'api_token' as password (Toggl's own token-auth convention). Never logged.
- `base_url` (optional, string); default `https://api.track.toggl.com/api/v9`; format `uri`; Toggl
  Track API base URL override for tests or proxies.
- `end_date` (optional, string); Optional end_date filter (Toggl date format) for the 'time_entries'
  stream.
- `organization_id` (optional, string); Toggl organization id required by the 'organization_users'
  stream.
- `start_date` (optional, string); Optional start_date filter (Toggl date format) for the
  'time_entries' stream.
- `workspace_id` (optional, string); Toggl workspace id required by the 'projects', 'clients', and
  'workspace_users' streams.

Secret fields are redacted in logs and write previews: `api_token`.

Default configuration values: `base_url=https://api.track.toggl.com/api/v9`.

Authentication behavior:

- HTTP Basic authentication using `secrets.api_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/me`.

## Streams notes

Default pagination: single request; no pagination.

Pagination by stream: none: `time_entries`, `projects`, `clients`, `workspace_users`,
`organization_users`, `tags`; page_number: `tasks`.

- `time_entries`: GET `/me/time_entries` - records path `.`; query `end_date` from template `{{
  config.end_date }}`, omitted when absent; `start_date` from template `{{ config.start_date }}`,
  omitted when absent; emits passthrough records.
- `projects`: GET `/workspaces/{{ config.workspace_id }}/projects` - records path `.`; emits
  passthrough records.
- `clients`: GET `/workspaces/{{ config.workspace_id }}/clients` - records path `.`; emits
  passthrough records.
- `workspace_users`: GET `/workspaces/{{ config.workspace_id }}/users` - records path `.`; emits
  passthrough records.
- `organization_users`: GET `/organizations/{{ config.organization_id }}/users` - records path `.`;
  emits passthrough records.
- `tags`: GET `/workspaces/{{ config.workspace_id }}/tags` - records path `.`; emits passthrough
  records.
- `tasks`: GET `/workspaces/{{ config.workspace_id }}/tasks` - records path `data`; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 100; emits
  passthrough records.

## Write actions & risks

Overall write risk: external mutation of Toggl time entries, projects, clients, tags, and tasks; no
destructive-admin or elevated-scope actions modeled.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_time_entry`: POST `/workspaces/{{ config.workspace_id }}/time_entries` - kind `create`;
  body type `json`; required record fields `start`, `duration`, `created_with`; accepted fields
  `billable`, `created_with`, `description`, `duration`, `project_id`, `start`, `stop`, `tag_ids`,
  `tags`, `task_id`; risk: creates a new time entry on the caller's account; external mutation, no
  approval required.
- `update_time_entry`: PUT `/workspaces/{{ config.workspace_id }}/time_entries/{{ record.id }}` -
  kind `update`; body type `json`; path fields `id`; body fields `description`, `start`, `stop`,
  `duration`, `project_id`, `task_id`, `tag_ids`, `tags`, `tag_action`, `billable`, and 1 more;
  required record fields `id`; accepted fields `billable`, `created_with`, `description`,
  `duration`, `id`, `project_id`, `start`, `stop`, `tag_action`, `tag_ids`, `tags`, `task_id`; risk:
  mutates an existing time entry's timing, project/task association, tags, or billable flag.
- `stop_time_entry`: PATCH `/workspaces/{{ config.workspace_id }}/time_entries/{{ record.id }}/stop`
  - kind `custom`; body type `none`; path fields `id`; required record fields `id`; accepted fields
  `id`; risk: stops a currently-running time entry by setting its stop time to now; no effect on an
  already-stopped entry beyond the API's own idempotency.
- `delete_time_entry`: DELETE `/workspaces/{{ config.workspace_id }}/time_entries/{{ record.id }}` -
  kind `delete`; body type `none`; path fields `id`; required record fields `id`; accepted fields
  `id`; missing records treated as success for status `404`; risk: permanently deletes a time entry;
  irreversible.
- `create_project`: POST `/workspaces/{{ config.workspace_id }}/projects` - kind `create`; body type
  `json`; required record fields `name`; accepted fields `active`, `billable`, `client_id`, `color`,
  `currency`, `end_date`, `is_private`, `name`, `rate`, `start_date`; risk: creates a new project in
  the target workspace; external mutation, no approval required.
- `update_project`: PUT `/workspaces/{{ config.workspace_id }}/projects/{{ record.id }}` - kind
  `update`; body type `json`; path fields `id`; body fields `name`, `client_id`, `is_private`,
  `active`, `color`, `billable`, `rate`, `currency`, `start_date`, `end_date`; required record
  fields `id`; accepted fields `active`, `billable`, `client_id`, `color`, `currency`, `end_date`,
  `id`, `is_private`, `name`, `rate`, `start_date`; risk: mutates an existing project's name, client
  association, active/private state, or billing settings.
- `delete_project`: DELETE `/workspaces/{{ config.workspace_id }}/projects/{{ record.id }}` - kind
  `delete`; body type `none`; path fields `id`; required record fields `id`; accepted fields `id`;
  missing records treated as success for status `404`; risk: permanently deletes a project; also
  removes its association from any time entries that referenced it.
- `create_client`: POST `/workspaces/{{ config.workspace_id }}/clients` - kind `create`; body type
  `json`; required record fields `name`; accepted fields `external_reference`, `name`, `notes`;
  risk: creates a new client in the target workspace; external mutation, no approval required.
- `update_client`: PUT `/workspaces/{{ config.workspace_id }}/clients/{{ record.id }}` - kind
  `update`; body type `json`; path fields `id`; body fields `name`, `notes`, `external_reference`;
  required record fields `id`; accepted fields `external_reference`, `id`, `name`, `notes`; risk:
  mutates an existing client's name or notes.
- `delete_client`: DELETE `/workspaces/{{ config.workspace_id }}/clients/{{ record.id }}` - kind
  `delete`; body type `none`; path fields `id`; required record fields `id`; accepted fields `id`;
  missing records treated as success for status `404`; risk: permanently deletes a client; projects
  previously associated with it lose that association.
- `create_tag`: POST `/workspaces/{{ config.workspace_id }}/tags` - kind `create`; body type `json`;
  required record fields `name`; accepted fields `name`; risk: creates a new tag in the target
  workspace; external mutation, no approval required.
- `update_tag`: PUT `/workspaces/{{ config.workspace_id }}/tags/{{ record.id }}` - kind `update`;
  body type `json`; path fields `id`; body fields `name`; required record fields `id`, `name`;
  accepted fields `id`, `name`; risk: renames an existing tag; the new name applies retroactively
  everywhere the tag is shown.
- `delete_tag`: DELETE `/workspaces/{{ config.workspace_id }}/tags/{{ record.id }}` - kind `delete`;
  body type `none`; path fields `id`; required record fields `id`; accepted fields `id`; missing
  records treated as success for status `404`; risk: permanently deletes a tag; it is removed from
  every time entry that referenced it.
- `create_task`: POST `/workspaces/{{ config.workspace_id }}/projects/{{ record.project_id }}/tasks`
  - kind `create`; body type `json`; path fields `project_id`; body fields `name`, `active`,
  `estimated_seconds`, `user_id`, `external_reference`; required record fields `project_id`, `name`;
  accepted fields `active`, `estimated_seconds`, `external_reference`, `name`, `project_id`,
  `user_id`; risk: creates a new task under the given project; external mutation, no approval
  required.
- `update_task`: PUT `/workspaces/{{ config.workspace_id }}/projects/{{ record.project_id
  }}/tasks/{{ record.id }}` - kind `update`; body type `json`; path fields `project_id`, `id`; body
  fields `name`, `active`, `estimated_seconds`, `user_id`, `external_reference`; required record
  fields `project_id`, `id`; accepted fields `active`, `estimated_seconds`, `external_reference`,
  `id`, `name`, `project_id`, `user_id`; risk: mutates an existing task's name, active/done state,
  estimate, or assignee; setting active:false marks the task done.
- `delete_task`: DELETE `/workspaces/{{ config.workspace_id }}/projects/{{ record.project_id
  }}/tasks/{{ record.id }}` - kind `delete`; body type `none`; path fields `project_id`, `id`;
  required record fields `project_id`, `id`; accepted fields `id`, `project_id`; missing records
  treated as success for status `404`; risk: permanently deletes a task; time entries previously
  linked to it lose that association.

## Known limits

- API coverage includes 7 stream-backed endpoint group(s), 16 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=12, destructive_admin=4, duplicate_of=20, non_data_endpoint=10, out_of_scope=118,
  requires_elevated_scope=95.
