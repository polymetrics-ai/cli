# Overview

Reads Everhour projects, clients, team members, team time records, per-project tasks and sections,
time-off types, time-off allocations, expenses, expense categories, and invoices, and writes
client/project/task/section/time-record/expense mutations, through the Everhour REST API.

Readable streams: `projects`, `clients`, `users`, `time`, `tasks`, `sections`, `time_off_types`,
`allocations`, `expense_categories`, `expenses`, `invoices`.

Write actions: `create_client`, `update_client`, `delete_client`, `create_project`,
`update_project`, `archive_project`, `delete_project`, `create_task`, `update_task`, `delete_task`,
`create_section`, `delete_section`, `create_time_record`, `update_time_record`,
`delete_time_record`, `create_expense`, `delete_expense`.

Service API documentation: https://everhour.docs.apiary.io/.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Everhour API key, sent as the X-Api-Key header. Never
  logged.
- `base_url` (optional, string); default `https://api.everhour.com`; format `uri`; Everhour API base
  URL override for tests or proxies.
- `mode` (optional, string).

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.everhour.com`.

Authentication behavior:

- API key authentication in `X-Api-Key` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/team/users`.

## Streams notes

Default pagination: single request; no pagination.

- `projects`: GET `/projects` - records at response root.
- `clients`: GET `/clients` - records at response root.
- `users`: GET `/team/users` - records at response root.
- `time`: GET `/team/time` - records at response root.
- `tasks`: GET `/projects/{{ fanout.id }}/tasks` - records at response root; fan-out; ids from
  request `/projects`; id field `id`; id inserted into the request path; stamps `project_id`.
- `sections`: GET `/projects/{{ fanout.id }}/sections` - records at response root; fan-out; ids from
  request `/projects`; id field `id`; id inserted into the request path; stamps `project_id`.
- `time_off_types`: GET `/resource-planner/time-off-types` - records at response root.
- `allocations`: GET `/allocations` - records at response root.
- `expense_categories`: GET `/expenses/categories` - records at response root.
- `expenses`: GET `/expenses` - records at response root.
- `invoices`: GET `/invoices` - records at response root.

## Write actions & risks

Overall write risk: external mutation of clients, projects, tasks, sections, time records, and
expenses; deletes are irreversible and time-record/expense mutations can affect client
billing/invoicing, every write ships with an explicit per-action risk string.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_client`: POST `/clients` - kind `create`; body type `json`; required record fields `name`;
  accepted fields `name`; risk: creates a new client record; low-risk external mutation, no approval
  required.
- `update_client`: PUT `/clients/{{ record.id }}` - kind `update`; body type `json`; path fields
  `id`; required record fields `id`; accepted fields `id`, `name`; risk: renames or otherwise
  mutates an existing client's metadata; low-risk external mutation.
- `delete_client`: DELETE `/clients/{{ record.id }}` - kind `delete`; body type `none`; path fields
  `id`; required record fields `id`; accepted fields `id`; missing records treated as success for
  status `404`; risk: permanently removes a client and its association with any linked projects;
  irreversible, approval required.
- `create_project`: POST `/projects` - kind `create`; body type `json`; required record fields
  `name`, `type`; accepted fields `name`, `type`; risk: creates a new project; low-risk external
  mutation, no approval required.
- `update_project`: PUT `/projects/{{ record.id }}` - kind `update`; body type `json`; path fields
  `id`; required record fields `id`; accepted fields `id`, `name`, `type`; risk: renames or
  reconfigures an existing project; low-risk external mutation.
- `archive_project`: PATCH `/projects/{{ record.id }}/archive` - kind `update`; body type `json`;
  path fields `id`; required record fields `id`, `archived`; accepted fields `archived`, `id`; risk:
  archives or unarchives a project, hiding it from active project lists and blocking new time
  entries against it while archived; approval required for archiving a project still in active use.
- `delete_project`: DELETE `/projects/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `404`; risk: permanently removes a project and its tasks/sections/time associations;
  irreversible, approval required.
- `create_task`: POST `/projects/{{ record.project_id }}/tasks` - kind `create`; body type `json`;
  path fields `project_id`; required record fields `project_id`, `name`, `section`; accepted fields
  `dueOn`, `labels`, `name`, `project_id`, `section`; risk: creates a new task under an existing
  project section; low-risk external mutation, no approval required.
- `update_task`: PUT `/tasks/{{ record.id }}` - kind `update`; body type `json`; path fields `id`;
  required record fields `id`; accepted fields `dueOn`, `id`, `name`, `section`, `status`; risk:
  renames or reconfigures an existing task; low-risk external mutation.
- `delete_task`: DELETE `/tasks/{{ record.id }}` - kind `delete`; body type `none`; path fields
  `id`; required record fields `id`; accepted fields `id`; missing records treated as success for
  status `404`; risk: permanently removes a task and its logged time association; irreversible,
  approval required.
- `create_section`: POST `/projects/{{ record.project_id }}/sections` - kind `create`; body type
  `json`; path fields `project_id`; required record fields `project_id`, `name`; accepted fields
  `name`, `position`, `project_id`; risk: creates a new task section within a project; low-risk
  external mutation, no approval required.
- `delete_section`: DELETE `/sections/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `404`; risk: permanently removes a task section; any tasks in it become unsectioned,
  approval required.
- `create_time_record`: POST `/time` - kind `create`; body type `json`; required record fields
  `time`, `date`; accepted fields `comment`, `date`, `task`, `time`, `user`; risk: logs a new time
  entry against a task, which can feed directly into client billing/invoicing; low-risk external
  mutation, no approval required.
- `update_time_record`: PUT `/time/{{ record.id }}` - kind `update`; body type `json`; path fields
  `id`; required record fields `id`; accepted fields `comment`, `date`, `id`, `time`; risk: changes
  the logged duration/date/comment of an existing time entry that may already be invoiced; approval
  required if the entry is locked or billed.
- `delete_time_record`: DELETE `/time/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `404`; risk: permanently removes a logged time entry, which can affect
  billing/invoicing history; irreversible, approval required.
- `create_expense`: POST `/expenses` - kind `create`; body type `json`; required record fields
  `category`, `date`; accepted fields `amount`, `billable`, `category`, `date`, `details`,
  `project`, `quantity`, `user`; risk: logs a new billable/non-billable expense, which can feed
  directly into client invoicing; low-risk external mutation, no approval required.
- `delete_expense`: DELETE `/expenses/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `404`; risk: permanently removes a logged expense, which can affect billing/invoicing
  history; irreversible, approval required.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 11 stream-backed endpoint group(s), 17 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=3, deprecated=2, destructive_admin=2, duplicate_of=10, non_data_endpoint=1,
  out_of_scope=54, requires_elevated_scope=3.
