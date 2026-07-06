# Overview

Reads Clockify workspaces, clients, projects, tags, users, tasks, time entries, custom fields, user
groups, holidays, expense categories, and time-off policies, and writes clients/projects/tags/tasks
through the Clockify REST API v1.

Readable streams: `workspaces`, `clients`, `projects`, `tags`, `users`, `current_user`,
`custom_fields`, `user_groups`, `holidays`, `expense_categories`, `time_off_policies`, `tasks`,
`time_entries`.

Write actions: `create_client`, `update_client`, `delete_client`, `create_project`,
`update_project`, `delete_project`, `create_tag`, `update_tag`, `delete_tag`, `create_task`,
`update_task`, `delete_task`.

Service API documentation: https://docs.clockify.me/.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Clockify API key, sent as the X-Api-Key header. Used only
  for auth; never logged.
- `base_url` (optional, string); default `https://api.clockify.me/api`; format `uri`; Clockify API
  base URL override for tests or proxies.
- `workspace_id` (optional, string); Clockify workspace id; required for the
  clients/projects/tags/users streams, which are scoped under /v1/workspaces/{workspace_id}/.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.clockify.me/api`.

Authentication behavior:

- API key authentication in `X-Api-Key` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/v1/workspaces`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `page-size`;
starts at 1; page size 50.

Pagination by stream: none: `current_user`, `custom_fields`, `user_groups`, `expense_categories`;
page_number: `workspaces`, `clients`, `projects`, `tags`, `users`, `holidays`, `time_off_policies`,
`tasks`, `time_entries`.

- `workspaces`: GET `/v1/workspaces` - records path `.`; page-number pagination; page parameter
  `page`; size parameter `page-size`; starts at 1; page size 50.
- `clients`: GET `/v1/workspaces/{{ config.workspace_id }}/clients` - records path `.`; page-number
  pagination; page parameter `page`; size parameter `page-size`; starts at 1; page size 50.
- `projects`: GET `/v1/workspaces/{{ config.workspace_id }}/projects` - records path `.`;
  page-number pagination; page parameter `page`; size parameter `page-size`; starts at 1; page size
  50.
- `tags`: GET `/v1/workspaces/{{ config.workspace_id }}/tags` - records path `.`; page-number
  pagination; page parameter `page`; size parameter `page-size`; starts at 1; page size 50.
- `users`: GET `/v1/workspaces/{{ config.workspace_id }}/users` - records path `.`; page-number
  pagination; page parameter `page`; size parameter `page-size`; starts at 1; page size 50.
- `current_user`: GET `/v1/user` - records path `.`.
- `custom_fields`: GET `/v1/workspaces/{{ config.workspace_id }}/custom-fields` - records path `.`.
- `user_groups`: GET `/v1/workspaces/{{ config.workspace_id }}/user-groups` - records path `.`.
- `holidays`: GET `/v1/workspaces/{{ config.workspace_id }}/holidays` - records path `.`;
  page-number pagination; page parameter `page`; size parameter `page-size`; starts at 1; page size
  50.
- `expense_categories`: GET `/v1/workspaces/{{ config.workspace_id }}/expenses/categories` - records
  path `categories`.
- `time_off_policies`: GET `/v1/workspaces/{{ config.workspace_id }}/time-off/policies` - records
  path `.`; page-number pagination; page parameter `page`; size parameter `page-size`; starts at 1;
  page size 50.
- `tasks`: GET `/v1/workspaces/{{ config.workspace_id }}/projects/{{ fanout.id }}/tasks` - records
  path `.`; page-number pagination; page parameter `page`; size parameter `page-size`; starts at 1;
  page size 50; fan-out; ids from request `/v1/workspaces/{{ config.workspace_id }}/projects`;
  id-list records path `.`; id field `id`; id inserted into the request path; stamps `projectId`.
- `time_entries`: GET `/v1/workspaces/{{ config.workspace_id }}/user/{{ fanout.id }}/time-entries` -
  records path `.`; page-number pagination; page parameter `page`; size parameter `page-size`;
  starts at 1; page size 50; fan-out; ids from request `/v1/workspaces/{{ config.workspace_id
  }}/users`; id-list records path `.`; id field `id`; id inserted into the request path; stamps
  `userId`.

## Write actions & risks

Overall write risk: external mutation; creates/updates/deletes live Clockify clients, projects,
tags, and tasks.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_client`: POST `/v1/workspaces/{{ config.workspace_id }}/clients` - kind `create`; body
  type `json`; required record fields `name`; accepted fields `address`, `email`, `name`, `note`;
  risk: external mutation; creates a live Clockify client; approval required.
- `update_client`: PUT `/v1/workspaces/{{ config.workspace_id }}/clients/{{ record.id }}` - kind
  `update`; body type `json`; path fields `id`; required record fields `id`, `name`; accepted fields
  `address`, `archived`, `ccEmails`, `currencyId`, `email`, `id`, `name`, `note`; risk: external
  mutation; overwrites a live Clockify client's fields; approval required.
- `delete_client`: DELETE `/v1/workspaces/{{ config.workspace_id }}/clients/{{ record.id }}` - kind
  `delete`; body type `none`; path fields `id`; required record fields `id`; accepted fields `id`;
  risk: external mutation; irreversibly deletes a live Clockify client; approval required.
- `create_project`: POST `/v1/workspaces/{{ config.workspace_id }}/projects` - kind `create`; body
  type `json`; required record fields `name`; accepted fields `billable`, `clientId`, `color`,
  `estimate`, `isPublic`, `name`, `note`; risk: external mutation; creates a live Clockify project;
  approval required.
- `update_project`: PUT `/v1/workspaces/{{ config.workspace_id }}/projects/{{ record.id }}` - kind
  `update`; body type `json`; path fields `id`; required record fields `id`, `name`; accepted fields
  `archived`, `billable`, `clientId`, `color`, `id`, `isPublic`, `name`, `note`; risk: external
  mutation; overwrites a live Clockify project's fields; approval required.
- `delete_project`: DELETE `/v1/workspaces/{{ config.workspace_id }}/projects/{{ record.id }}` -
  kind `delete`; body type `none`; path fields `id`; required record fields `id`; accepted fields
  `id`; risk: external mutation; irreversibly deletes a live Clockify project; approval required.
- `create_tag`: POST `/v1/workspaces/{{ config.workspace_id }}/tags` - kind `create`; body type
  `json`; required record fields `name`; accepted fields `name`; risk: external mutation; creates a
  live Clockify tag; approval required.
- `update_tag`: PUT `/v1/workspaces/{{ config.workspace_id }}/tags/{{ record.id }}` - kind `update`;
  body type `json`; path fields `id`; required record fields `id`, `name`; accepted fields
  `archived`, `id`, `name`; risk: external mutation; overwrites a live Clockify tag's fields;
  approval required.
- `delete_tag`: DELETE `/v1/workspaces/{{ config.workspace_id }}/tags/{{ record.id }}` - kind
  `delete`; body type `none`; path fields `id`; required record fields `id`; accepted fields `id`;
  risk: external mutation; irreversibly deletes a live Clockify tag; approval required.
- `create_task`: POST `/v1/workspaces/{{ config.workspace_id }}/projects/{{ record.projectId
  }}/tasks` - kind `create`; body type `json`; path fields `projectId`; required record fields
  `projectId`, `name`; accepted fields `assigneeId`, `assigneeIds`, `budgetEstimate`, `estimate`,
  `name`, `projectId`, `status`; risk: external mutation; creates a live Clockify task on a project;
  approval required.
- `update_task`: PUT `/v1/workspaces/{{ config.workspace_id }}/projects/{{ record.projectId
  }}/tasks/{{ record.id }}` - kind `update`; body type `json`; path fields `projectId`, `id`;
  required record fields `projectId`, `id`, `name`; accepted fields `assigneeId`, `assigneeIds`,
  `billable`, `estimate`, `id`, `name`, `projectId`, `status`; risk: external mutation; overwrites a
  live Clockify task's fields; approval required.
- `delete_task`: DELETE `/v1/workspaces/{{ config.workspace_id }}/projects/{{ record.projectId
  }}/tasks/{{ record.id }}` - kind `delete`; body type `none`; path fields `projectId`, `id`;
  required record fields `projectId`, `id`; accepted fields `id`, `projectId`; risk: external
  mutation; irreversibly deletes a live Clockify task; approval required.

## Known limits

- Batch defaults: read_page_size=50.
- API coverage includes 13 stream-backed endpoint group(s), 12 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=4, destructive_admin=13, duplicate_of=12, non_data_endpoint=5, out_of_scope=90,
  requires_elevated_scope=19.
