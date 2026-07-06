# Overview

Reads Teamwork projects, people, companies, tags, time entries, tasklists, milestones, and tasks,
and writes approved project/tasklist/task/milestone/company/time-entry mutations through the
Teamwork API.

Readable streams: `projects`, `people`, `companies`, `tags`, `time_entries`, `tasklists`,
`milestones`, `tasks`.

Write actions: `create_project`, `update_project`, `create_tasklist`, `create_task`, `update_task`,
`complete_task`, `create_milestone`, `create_company`, `create_time_entry`.

Service API documentation: https://apidocs.teamwork.com/.

## Auth setup

Connection fields:

- `base_url` (optional, string); default `https://api.teamwork.com`; format `uri`; Teamwork API base
  URL override for tests or proxies.
- `password` (required, secret, string); Teamwork API token/password, sent as the Basic auth
  password. Never logged.
- `username` (required, string); Teamwork username (email), sent as the Basic auth username.

Secret fields are redacted in logs and write previews: `password`.

Default configuration values: `base_url=https://api.teamwork.com`.

Authentication behavior:

- HTTP Basic authentication using `config.username`, `secrets.password`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/projects.json`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `pageSize`; starts
at 1; page size 100.

- `projects`: GET `/projects.json` - records path `projects`; page-number pagination; page parameter
  `page`; size parameter `pageSize`; starts at 1; page size 100; computed output fields
  `created_at`.
- `people`: GET `/people.json` - records path `people`; page-number pagination; page parameter
  `page`; size parameter `pageSize`; starts at 1; page size 100; computed output fields
  `first_name`, `last_name`.
- `companies`: GET `/companies.json` - records path `companies`; page-number pagination; page
  parameter `page`; size parameter `pageSize`; starts at 1; page size 100.
- `tags`: GET `/tags.json` - records path `tags`; page-number pagination; page parameter `page`;
  size parameter `pageSize`; starts at 1; page size 100.
- `time_entries`: GET `/time_entries.json` - records path `time-entries`; page-number pagination;
  page parameter `page`; size parameter `pageSize`; starts at 1; page size 100; computed output
  fields `created_at`, `person_id`, `project_id`, `todo_item_id`.
- `tasklists`: GET `/projects/{{ fanout.id }}/tasklists.json` - records path `todo-lists`;
  page-number pagination; page parameter `page`; size parameter `pageSize`; starts at 1; page size
  100; fan-out; ids from request `/projects.json`; id-list records path `projects`; id field `id`;
  id inserted into the request path; stamps `project_id`.
- `milestones`: GET `/projects/{{ fanout.id }}/milestones.json` - records path `milestones`;
  page-number pagination; page parameter `page`; size parameter `pageSize`; starts at 1; page size
  100; computed output fields `created_at`; fan-out; ids from request `/projects.json`; id-list
  records path `projects`; id field `id`; id inserted into the request path; stamps `project_id`.
- `tasks`: GET `/tasks.json` - records path `todo-items`; page-number pagination; page parameter
  `page`; size parameter `pageSize`; starts at 1; page size 100; computed output fields
  `created_at`.

## Write actions & risks

Overall write risk: external Teamwork API mutation (create/update projects, tasklists, tasks,
milestones, companies, time entries; complete tasks).

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_project`: POST `/projects.json` - kind `create`; body type `json`; required record fields
  `project`; accepted fields `project`; risk: creates a new project; low-risk external mutation, no
  approval required. Body is wrapped under a top-level "project" key (Teamwork's V1 API convention)
  - the record itself must carry that wrapper, since the engine's write dialect sends record fields
  verbatim as the JSON body with no nested-wrapper construction primitive.
- `update_project`: PUT `/projects/{{ record.id }}.json` - kind `update`; body type `json`; path
  fields `id`; required record fields `id`, `project`; accepted fields `id`, `project`; risk:
  mutates an existing project's name or description; visible to every project member.
- `create_tasklist`: POST `/projects/{{ record.project_id }}/tasklists.json` - kind `create`; body
  type `json`; path fields `project_id`; required record fields `project_id`, `todo-list`; accepted
  fields `project_id`, `todo-list`; risk: creates a new task list under the target project; low-risk
  external mutation, no approval required.
- `create_task`: POST `/tasklists/{{ record.tasklist_id }}/tasks.json` - kind `create`; body type
  `json`; path fields `tasklist_id`; required record fields `tasklist_id`, `todo-item`; accepted
  fields `tasklist_id`, `todo-item`; risk: creates a new task in the target task list; low-risk
  external mutation, no approval required.
- `update_task`: PUT `/tasks/{{ record.id }}.json` - kind `update`; body type `json`; path fields
  `id`; required record fields `id`, `todo-item`; accepted fields `id`, `todo-item`; risk: mutates
  an existing task's content, description, or priority.
- `complete_task`: PUT `/tasks/{{ record.id }}/complete.json` - kind `update`; body type `none`;
  path fields `id`; required record fields `id`; accepted fields `id`; risk: marks an existing task
  as complete; a visible, notifiable state change for every task follower.
- `create_milestone`: POST `/projects/{{ record.project_id }}/milestones.json` - kind `create`; body
  type `json`; path fields `project_id`; required record fields `project_id`, `milestone`; accepted
  fields `milestone`, `project_id`; risk: creates a new milestone under the target project; low-risk
  external mutation, no approval required.
- `create_company`: POST `/companies.json` - kind `create`; body type `json`; required record fields
  `company`; accepted fields `company`; risk: creates a new company record; low-risk external
  mutation, no approval required.
- `create_time_entry`: POST `/projects/{{ record.project_id }}/time_entries.json` - kind `create`;
  body type `json`; path fields `project_id`; required record fields `project_id`, `time-entry`;
  accepted fields `project_id`, `time-entry`; risk: logs a new time entry against the target
  project; contributes to billable-hours totals and any linked invoice.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 8 stream-backed endpoint group(s), 9 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=10, destructive_admin=29, duplicate_of=40, non_data_endpoint=14, out_of_scope=105,
  requires_elevated_scope=12.
