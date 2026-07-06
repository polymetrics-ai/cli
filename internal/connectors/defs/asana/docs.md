# Overview

Reads Asana workspaces, projects, tasks, sections, tags, stories, users, teams, custom fields,
project statuses, and team/workspace memberships through the Asana v1 REST API. Writes
task/project/section/tag create-update-delete and task comments.

Readable streams: `workspaces`, `projects`, `tasks`, `users`, `teams`, `tags`, `sections`,
`stories`, `custom_fields`, `project_statuses`, `team_memberships`, `workspace_memberships`.

Write actions: `create_task`, `update_task`, `delete_task`, `create_project`, `update_project`,
`delete_project`, `create_section`, `update_section`, `delete_section`, `create_tag`, `update_tag`,
`delete_tag`, `add_comment`.

Service API documentation: https://developers.asana.com/reference/rest-api-reference.

## Auth setup

Connection fields:

- `access_token` (required, secret, string); Asana personal access token, sent as a Bearer token
  (Authorization: Bearer <access_token>). Never logged.
- `assignee` (optional, string); Optional Asana assignee gid; scopes the 'tasks' stream to this
  assignee when set.
- `base_url` (optional, string); default `https://app.asana.com/api/1.0`; format `uri`; Asana API
  base URL override for tests or proxies.
- `mode` (optional, string).
- `project_id` (optional, string); Optional Asana project gid; scopes the 'tasks' stream to this
  project when set.
- `team_id` (optional, string); Optional Asana team gid; scopes the 'team_memberships' stream to
  this team when set (Asana requires either team, or workspace+user, on that endpoint).
- `workspace_id` (optional, string); Optional Asana workspace gid; scopes the 'projects' and 'tasks'
  streams to this workspace when set.

Secret fields are redacted in logs and write previews: `access_token`.

Default configuration values: `base_url=https://app.asana.com/api/1.0`.

Authentication behavior:

- Bearer token authentication using `secrets.access_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/users/me`.

## Streams notes

Default pagination: follows a next-page URL from the response body; URL path `next_page.uri`; next
URLs stay on the configured API host.

- `workspaces`: GET `/workspaces` - records path `data`; query `limit`=`100`;
  `opt_fields`=`gid,name,resource_type`; follows a next-page URL from the response body; URL path
  `next_page.uri`; next URLs stay on the configured API host.
- `projects`: GET `/projects` - records path `data`; query `limit`=`100`;
  `opt_fields`=`gid,name,resource_type,created_at,modified_at`; `workspace` from template `{{
  config.workspace_id }}`, omitted when absent; follows a next-page URL from the response body; URL
  path `next_page.uri`; next URLs stay on the configured API host.
- `tasks`: GET `/tasks` - records path `data`; query `assignee` from template `{{ config.assignee
  }}`, omitted when absent; `limit`=`100`;
  `opt_fields`=`gid,name,resource_type,created_at,modified_at,completed`; `project` from template
  `{{ config.project_id }}`, omitted when absent; `workspace` from template `{{ config.workspace_id
  }}`, omitted when absent; follows a next-page URL from the response body; URL path
  `next_page.uri`; next URLs stay on the configured API host.
- `users`: GET `/users` - records path `data`; query `opt_fields`=`gid,name,resource_type,email`;
  `workspace` from template `{{ config.workspace_id }}`, omitted when absent; follows a next-page
  URL from the response body; URL path `next_page.uri`; next URLs stay on the configured API host.
- `teams`: GET `/workspaces/{{ config.workspace_id }}/teams` - records path `data`; query
  `opt_fields`=`gid,name,resource_type,description,visibility`; follows a next-page URL from the
  response body; URL path `next_page.uri`; next URLs stay on the configured API host.
- `tags`: GET `/tags` - records path `data`; query
  `opt_fields`=`gid,name,resource_type,color,notes,created_at`; `workspace` from template `{{
  config.workspace_id }}`, omitted when absent; follows a next-page URL from the response body; URL
  path `next_page.uri`; next URLs stay on the configured API host.
- `sections`: GET `/projects/{{ fanout.id }}/sections` - records path `data`; query
  `opt_fields`=`gid,name,resource_type,created_at`; follows a next-page URL from the response body;
  URL path `next_page.uri`; next URLs stay on the configured API host; fan-out; ids from request
  `/projects`; id-list records path `data`; id field `gid`; id inserted into the request path;
  stamps `project_gid`.
- `stories`: GET `/tasks/{{ fanout.id }}/stories` - records path `data`; query
  `opt_fields`=`gid,resource_type,created_at,resource_subtype,text,type,created_by.gid,created_by.name`;
  follows a next-page URL from the response body; URL path `next_page.uri`; next URLs stay on the
  configured API host; fan-out; ids from request `/tasks`; id-list records path `data`; id field
  `gid`; id inserted into the request path; stamps `task_gid`.
- `custom_fields`: GET `/workspaces/{{ config.workspace_id }}/custom_fields` - records path `data`;
  query `opt_fields`=`gid,name,resource_type,type,enabled,description`; follows a next-page URL from
  the response body; URL path `next_page.uri`; next URLs stay on the configured API host.
- `project_statuses`: GET `/projects/{{ fanout.id }}/project_statuses` - records path `data`; query
  `opt_fields`=`gid,resource_type,title,text,color,created_at,modified_at`; follows a next-page URL
  from the response body; URL path `next_page.uri`; next URLs stay on the configured API host;
  fan-out; ids from request `/projects`; id-list records path `data`; id field `gid`; id inserted
  into the request path; stamps `project_gid`.
- `team_memberships`: GET `/team_memberships` - records path `data`; query
  `opt_fields`=`gid,resource_type,is_guest,is_admin,is_limited_access,user.gid,user.name,team.gid,team.name`;
  `team` from template `{{ config.team_id }}`, omitted when absent; `workspace` from template `{{
  config.workspace_id }}`, omitted when absent; follows a next-page URL from the response body; URL
  path `next_page.uri`; next URLs stay on the configured API host.
- `workspace_memberships`: GET `/workspaces/{{ config.workspace_id }}/workspace_memberships` -
  records path `data`; query
  `opt_fields`=`gid,resource_type,is_active,is_admin,is_guest,user.gid,user.name`; follows a
  next-page URL from the response body; URL path `next_page.uri`; next URLs stay on the configured
  API host.

## Write actions & risks

Overall write risk: external mutations: creates/updates/deletes tasks, projects, sections, and tags,
and posts task comments.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_task`: POST `/tasks` - kind `create`; body type `json`; required record fields `data`;
  accepted fields `data`; risk: external mutation; creates a new task visible to the whole
  team/project; low risk, no approval required.
- `update_task`: PUT `/tasks/{{ record.gid }}` - kind `update`; body type `json`; path fields `gid`;
  required record fields `gid`, `data`; accepted fields `data`, `gid`; risk: external mutation;
  overwrites task fields (e.g. can mark completed, reassign, reschedule); approval required.
- `delete_task`: DELETE `/tasks/{{ record.gid }}` - kind `delete`; body type `none`; path fields
  `gid`; required record fields `gid`; accepted fields `gid`; missing records treated as success for
  status `404`; confirmation `destructive`; risk: irreversible external deletion of a task; approval
  required.
- `create_project`: POST `/projects` - kind `create`; body type `json`; required record fields
  `data`; accepted fields `data`; risk: external mutation; creates a new project visible to the
  team/workspace; low risk, no approval required.
- `update_project`: PUT `/projects/{{ record.gid }}` - kind `update`; body type `json`; path fields
  `gid`; required record fields `gid`, `data`; accepted fields `data`, `gid`; risk: external
  mutation; overwrites project fields (can archive, reassign owner); approval required.
- `delete_project`: DELETE `/projects/{{ record.gid }}` - kind `delete`; body type `none`; path
  fields `gid`; required record fields `gid`; accepted fields `gid`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: irreversible external deletion of a
  project and its association with its tasks; approval required.
- `create_section`: POST `/projects/{{ record.project_gid }}/sections` - kind `create`; body type
  `json`; path fields `project_gid`; required record fields `project_gid`, `data`; accepted fields
  `data`, `project_gid`; risk: external mutation; creates a new section in a project's board/list
  view; low risk, no approval required.
- `update_section`: PUT `/sections/{{ record.gid }}` - kind `update`; body type `json`; path fields
  `gid`; required record fields `gid`, `data`; accepted fields `data`, `gid`; risk: external
  mutation; renames a section; low risk, no approval required.
- `delete_section`: DELETE `/sections/{{ record.gid }}` - kind `delete`; body type `none`; path
  fields `gid`; required record fields `gid`; accepted fields `gid`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: irreversible external deletion of a
  section (Asana requires the section be empty of tasks first); approval required.
- `create_tag`: POST `/tags` - kind `create`; body type `json`; required record fields `data`;
  accepted fields `data`; risk: external mutation; creates a new workspace-visible tag; low risk, no
  approval required.
- `update_tag`: PUT `/tags/{{ record.gid }}` - kind `update`; body type `json`; path fields `gid`;
  required record fields `gid`, `data`; accepted fields `data`, `gid`; risk: external mutation;
  renames/recolors a tag visible workspace-wide; low risk, no approval required.
- `delete_tag`: DELETE `/tags/{{ record.gid }}` - kind `delete`; body type `none`; path fields
  `gid`; required record fields `gid`; accepted fields `gid`; missing records treated as success for
  status `404`; confirmation `destructive`; risk: irreversible external deletion of a tag, removed
  from every task that carries it; approval required.
- `add_comment`: POST `/tasks/{{ record.task_gid }}/stories` - kind `create`; body type `json`; path
  fields `task_gid`; required record fields `task_gid`, `data`; accepted fields `data`, `task_gid`;
  risk: external mutation; posts a comment visible to everyone with access to the task; approval
  required.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 12 stream-backed endpoint group(s), 13 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=1, duplicate_of=60, non_data_endpoint=10, out_of_scope=128,
  requires_elevated_scope=25.
