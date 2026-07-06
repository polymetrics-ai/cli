# Overview

Reads ClickUp workspaces (teams), spaces, folders, lists, tasks, goals, space tags, and webhooks,
and writes task/folder/list/space/webhook lifecycle mutations, task comments, tags, custom field
values, and goal creation, through the ClickUp v2 REST API using a personal API token.

Readable streams: `tasks`, `teams`, `spaces`, `folders`, `lists`, `goals`, `space_tags`, `webhooks`.

Write actions: `create_task`, `update_task`, `delete_task`, `create_task_comment`,
`add_tag_to_task`, `remove_tag_from_task`, `set_custom_field_value`, `create_goal`, `create_folder`,
`update_folder`, `delete_folder`, `create_list`, `update_list`, `delete_list`, `create_space`,
`update_space`, `delete_space`, `create_webhook`, `update_webhook`, `delete_webhook`.

Service API documentation: https://developer.clickup.com/reference.

## Auth setup

Connection fields:

- `api_token` (required, secret, string); ClickUp personal API token, sent raw (no Bearer prefix) in
  the Authorization header. Never logged.
- `base_url` (optional, string); default `https://api.clickup.com/api/v2`; format `uri`; ClickUp API
  base URL override for tests or proxies.
- `folder_id` (optional, string); ClickUp folder id. Required for the create_list write action
  (Lists are created within a specific Folder).
- `include_archived` (optional, string); When 'true', include archived spaces/folders/lists/tasks
  (sent as archived=true instead of the default archived=false).
- `include_closed_tasks` (optional, string); When 'true', include closed tasks in the 'tasks' stream
  (sent as include_closed=true).
- `list_id` (optional, string); ClickUp list id. Required for the create_task write action (tasks
  are created within a specific List).
- `mode` (optional, string).
- `space_id` (optional, string); ClickUp space id. Required for the 'folders', 'lists', and
  'space_tags' streams and the create_folder write action.
- `team_id` (optional, string); ClickUp team (workspace) id. Required for the 'spaces', 'tasks',
  'goals', and 'webhooks' streams and the create_goal/create_space/create_webhook write actions.

Secret fields are redacted in logs and write previews: `api_token`.

Default configuration values: `base_url=https://api.clickup.com/api/v2`.

Authentication behavior:

- API key authentication in `Authorization` using `secrets.api_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/team`.

## Streams notes

Default pagination: single request; no pagination.

Pagination by stream: none: `teams`, `spaces`, `folders`, `lists`, `goals`, `space_tags`,
`webhooks`; page_number: `tasks`.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `tasks`: GET `/team/{{ config.team_id }}/task` - records path `tasks`; query `archived` from
  template `{{ config.include_archived }}`, default `false`; `include_closed` from template `{{
  config.include_closed_tasks }}`, omitted when absent; page-number pagination; page parameter
  `page`; starts at 0; page size 100; incremental cursor `date_updated`; formatted as `rfc3339`;
  computed output fields `creator_id`, `folder_id`, `list_id`, `space_id`, `status`.
- `teams`: GET `/team` - records path `teams`; query `archived` from template `{{
  config.include_archived }}`, omitted when absent.
- `spaces`: GET `/team/{{ config.team_id }}/space` - records path `spaces`; query `archived` from
  template `{{ config.include_archived }}`, omitted when absent.
- `folders`: GET `/space/{{ config.space_id }}/folder` - records path `folders`; query `archived`
  from template `{{ config.include_archived }}`, omitted when absent; computed output fields
  `space_id`.
- `lists`: GET `/space/{{ config.space_id }}/list` - records path `lists`; query `archived` from
  template `{{ config.include_archived }}`, omitted when absent; computed output fields `space_id`.
- `goals`: GET `/team/{{ config.team_id }}/goal` - records path `goals`.
- `space_tags`: GET `/space/{{ config.space_id }}/tag` - records path `tags`; computed output fields
  `space_id`.
- `webhooks`: GET `/team/{{ config.team_id }}/webhook` - records path `webhooks`.

## Write actions & risks

Overall write risk: external mutation of ClickUp tasks, folders, lists, spaces, webhooks, tags,
custom field values, and goals; delete_task/delete_folder/delete_list/delete_space are irreversible
cascading deletes, and create_webhook/update_webhook register or repoint an outbound event-delivery
URL of the caller's choosing - every write ships with an explicit per-action risk string.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_task`: POST `/list/{{ config.list_id }}/task` - kind `create`; body type `json`; required
  record fields `name`; accepted fields `assignees`, `description`, `due_date`, `due_date_time`,
  `links_to`, `markdown_content`, `name`, `notify_all`, `parent`, `priority`, `start_date`,
  `start_date_time`, `status`, `tags`, `time_estimate`; risk: creates a new ClickUp task in the
  configured list; low-risk (additive).
- `update_task`: PUT `/task/{{ record.id }}` - kind `update`; body type `json`; path fields `id`;
  required record fields `id`; accepted fields `archived`, `description`, `due_date`,
  `due_date_time`, `id`, `markdown_content`, `name`, `priority`, `start_date`, `start_date_time`,
  `status`, `time_estimate`; risk: updates fields on an existing ClickUp task (name, description,
  status, dates, priority, archived); approval required.
- `delete_task`: DELETE `/task/{{ record.id }}` - kind `delete`; body type `none`; path fields `id`;
  required record fields `id`; accepted fields `id`; missing records treated as success for status
  `404`; risk: permanently deletes a ClickUp task; irreversible; approval required.
- `create_task_comment`: POST `/task/{{ record.task_id }}/comment` - kind `create`; body type
  `json`; body fields `comment_text`, `notify_all`, `assignee`, `group_assignee`; required record
  fields `task_id`, `comment_text`, `notify_all`; accepted fields `assignee`, `comment_text`,
  `group_assignee`, `notify_all`, `task_id`; risk: adds a new comment to a ClickUp task, visible to
  all task watchers when notify_all is true; low-risk.
- `add_tag_to_task`: POST `/task/{{ record.task_id }}/tag/{{ record.tag_name }}` - kind `update`;
  body type `none`; path fields `task_id`, `tag_name`; required record fields `task_id`, `tag_name`;
  accepted fields `tag_name`, `task_id`; risk: attaches an existing Space Tag to a task; low-risk.
- `remove_tag_from_task`: DELETE `/task/{{ record.task_id }}/tag/{{ record.tag_name }}` - kind
  `delete`; body type `none`; path fields `task_id`, `tag_name`; required record fields `task_id`,
  `tag_name`; accepted fields `tag_name`, `task_id`; missing records treated as success for status
  `404`; risk: removes a tag from a task (does not delete the tag from the Space); low-risk.
- `set_custom_field_value`: POST `/task/{{ record.task_id }}/field/{{ record.field_id }}` - kind
  `update`; body type `json`; path fields `task_id`, `field_id`; body fields `value`,
  `value_options`; required record fields `task_id`, `field_id`, `value`; accepted fields
  `field_id`, `task_id`, `value_options`; risk: sets a Custom Field value on a task; the accepted
  value shape varies by the field's type
  (text/number/date/dropdown/label/people/task-relationship/manual-progress/location/button);
  approval required since an incorrectly-typed value can silently fail or corrupt a
  differently-typed field.
- `create_goal`: POST `/team/{{ config.team_id }}/goal` - kind `create`; body type `json`; required
  record fields `name`; accepted fields `color`, `description`, `due_date`, `multiple_owners`,
  `name`, `owners`; risk: creates a new ClickUp Goal in the configured team/workspace; low-risk
  (additive).
- `create_folder`: POST `/space/{{ config.space_id }}/folder` - kind `create`; body type `json`;
  required record fields `name`; accepted fields `name`; risk: creates a new Folder in the
  configured space; low-risk (additive).
- `update_folder`: PUT `/folder/{{ record.id }}` - kind `update`; body type `json`; path fields
  `id`; required record fields `id`, `name`; accepted fields `id`, `name`; risk: renames an existing
  ClickUp Folder; approval required.
- `delete_folder`: DELETE `/folder/{{ record.id }}` - kind `delete`; body type `none`; path fields
  `id`; required record fields `id`; accepted fields `id`; missing records treated as success for
  status `404`; risk: permanently deletes a ClickUp Folder and every List/task inside it;
  irreversible; approval required.
- `create_list`: POST `/folder/{{ config.folder_id }}/list` - kind `create`; body type `json`;
  required record fields `name`; accepted fields `assignee`, `content`, `due_date`, `due_date_time`,
  `markdown_content`, `name`, `priority`, `status`; risk: creates a new List in the configured
  Folder; low-risk (additive).
- `update_list`: PUT `/list/{{ record.id }}` - kind `update`; body type `json`; path fields `id`;
  required record fields `id`, `name`; accepted fields `assignee`, `content`, `due_date`,
  `due_date_time`, `id`, `name`, `priority`, `status`, `unset_status`; risk: updates an existing
  ClickUp List's name/description/due date/priority/assignee/color; approval required.
- `delete_list`: DELETE `/list/{{ record.id }}` - kind `delete`; body type `none`; path fields `id`;
  required record fields `id`; accepted fields `id`; missing records treated as success for status
  `404`; risk: permanently deletes a ClickUp List and every task inside it; irreversible; approval
  required.
- `create_space`: POST `/team/{{ config.team_id }}/space` - kind `create`; body type `json`;
  required record fields `name`; accepted fields `features`, `multiple_assignees`, `name`; risk:
  creates a new Space in the configured Workspace; low-risk (additive).
- `update_space`: PUT `/space/{{ record.id }}` - kind `update`; body type `json`; path fields `id`;
  required record fields `id`, `name`; accepted fields `admin_can_manage`, `color`, `features`,
  `id`, `multiple_assignees`, `name`, `private`; risk: updates an existing ClickUp Space's
  name/color/privacy/ClickApp feature toggles; ClickUp's own docs mark every body field required (a
  partial update still needs the full current feature set re-sent to avoid resetting unspecified
  features); approval required.
- `delete_space`: DELETE `/space/{{ record.id }}` - kind `delete`; body type `none`; path fields
  `id`; required record fields `id`; accepted fields `id`; missing records treated as success for
  status `404`; risk: permanently deletes a ClickUp Space and every Folder/List/task inside it;
  irreversible; approval required.
- `create_webhook`: POST `/team/{{ config.team_id }}/webhook` - kind `create`; body type `json`;
  required record fields `endpoint`, `events`; accepted fields `endpoint`, `events`, `folder_id`,
  `list_id`, `space_id`, `task_id`; risk: registers or repoints an outbound event-delivery URL of
  the caller's choosing; approval required.
- `update_webhook`: PUT `/webhook/{{ record.id }}` - kind `update`; body type `json`; path fields
  `id`; required record fields `id`; accepted fields `endpoint`, `events`, `id`, `status`; risk:
  changes which events are delivered to (or repoints) an existing outbound webhook; approval
  required.
- `delete_webhook`: DELETE `/webhook/{{ record.id }}` - kind `delete`; body type `none`; path fields
  `id`; required record fields `id`; accepted fields `id`; missing records treated as success for
  status `404`; risk: stops event delivery to a registered webhook endpoint; approval required
  (irreversible without re-registering).

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 8 stream-backed endpoint group(s), 20 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=2, deprecated=4, destructive_admin=3, duplicate_of=67, non_data_endpoint=4,
  out_of_scope=48, requires_elevated_scope=16.
