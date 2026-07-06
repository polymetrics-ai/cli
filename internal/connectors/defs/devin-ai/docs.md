# Overview

Reads Devin AI sessions, session child resources, playbooks, knowledge notes, repositories,
schedules, membership, metrics, consumption, and secret metadata through the Devin v3 REST API;
writes documented organization-scoped JSON mutations.

Readable streams: `sessions`, `sessions_insights`, `session_details`, `session_messages`,
`session_attachments`, `session_tags`, `playbooks`, `secrets`, `knowledge_notes`,
`knowledge_folders`, `repositories`, `indexed_repositories`, `schedules`, `organization_users`,
`organization_idp_group_users`, `self`, `org_daily_consumption`, `session_daily_consumption`,
`user_daily_consumption`, `org_usage_metrics`, `org_session_metrics`, `org_active_users_metrics`,
`org_daily_active_users`, `org_monthly_active_users`, `org_pr_metrics`, `org_search_metrics`,
`org_weekly_active_users`.

Write actions: `create_session`, `send_session_message`, `append_session_tags`,
`replace_session_tags`, `archive_session`, `terminate_session`, `generate_session_insights`,
`create_schedule`, `update_schedule`, `delete_schedule`, `create_playbook`, `update_playbook`,
`delete_playbook`, `create_knowledge_note`, `update_knowledge_note`, `delete_knowledge_note`,
`index_repository`, `bulk_index_repositories`, `remove_repository_indexing`,
`bulk_remove_repository_indexing`, `remove_repository_branch_indexing`, `trigger_pr_review`,
`delete_secret`.

Service API documentation: https://docs.devin.ai/api-reference/overview.

## Auth setup

Connection fields:

- `api_token` (required, secret, string); Devin service-user API key. Used only for Bearer auth;
  never logged.
- `base_url` (optional, string); default `https://api.devin.ai`; format `uri`; Devin API base URL
  override for tests or proxies.
- `metrics_time_after` (optional, string); Optional Unix-seconds lower bound sent as time_after on
  metrics and consumption streams.
- `metrics_time_before` (optional, string); Optional Unix-seconds upper bound sent as time_before on
  metrics and consumption streams.
- `mode` (optional, string).
- `org_id` (required, string); Devin organization id; every stream reads
  /v3/organizations/{org_id}/... scoped to this org.
- `page_size` (optional, string); default `100`; Records per page (1-200), sent as the first query
  param on cursor-paginated streams.
- `repository_filter_name` (optional, string); Optional case-insensitive repository-name filter for
  the repositories stream.
- `start_date` (optional, string); format `date-time`; RFC3339 lower bound; only session-derived
  objects created at or after this time are read.
- `user_email` (optional, string); Optional exact email filter for organization user streams.

Secret fields are redacted in logs and write previews: `api_token`.

Default configuration values: `base_url=https://api.devin.ai`, `page_size=100`.

Authentication behavior:

- Bearer token authentication using `secrets.api_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/v3/organizations/{{ config.org_id }}/sessions`.

## Streams notes

Default pagination: cursor pagination; cursor parameter `after`; next token from `end_cursor`; stop
flag `has_next_page`.

Pagination by stream: cursor: `sessions`, `sessions_insights`, `session_details`,
`session_messages`, `session_attachments`, `session_tags`, `playbooks`, `secrets`,
`knowledge_notes`, `repositories`, `indexed_repositories`, `organization_users`,
`organization_idp_group_users`; none: `knowledge_folders`, `self`, `org_daily_consumption`,
`session_daily_consumption`, `user_daily_consumption`, `org_usage_metrics`, `org_session_metrics`,
`org_active_users_metrics`, `org_daily_active_users`, `org_monthly_active_users`, `org_pr_metrics`,
`org_search_metrics`, `org_weekly_active_users`; offset_limit: `schedules`.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `sessions`: GET `/v3/organizations/{{ config.org_id }}/sessions` - records path `items`; query
  `first` from template `{{ config.page_size }}`, default `100`; cursor pagination; cursor parameter
  `after`; next token from `end_cursor`; stop flag `has_next_page`; incremental cursor `created_at`;
  sent as `created_after`; formatted as Unix-seconds timestamp; initial lower bound from
  `start_date`.
- `sessions_insights`: GET `/v3/organizations/{{ config.org_id }}/sessions/insights` - records path
  `items`; query `first` from template `{{ config.page_size }}`, default `100`; cursor pagination;
  cursor parameter `after`; next token from `end_cursor`; stop flag `has_next_page`; incremental
  cursor `created_at`; sent as `created_after`; formatted as Unix-seconds timestamp; initial lower
  bound from `start_date`.
- `session_details`: GET `/v3/organizations/{{ config.org_id }}/sessions/{{ fanout.id }}` -
  single-object response; records path `.`; cursor pagination; cursor parameter `after`; next token
  from `end_cursor`; stop flag `has_next_page`; fan-out; ids from request `/v3/organizations/{{
  config.org_id }}/sessions`; id-list records path `items`; id field `session_id`; id inserted into
  the request path.
- `session_messages`: GET `/v3/organizations/{{ config.org_id }}/sessions/{{ fanout.id }}/messages`
  - records path `items`; query `first` from template `{{ config.page_size }}`, default `100`;
  cursor pagination; cursor parameter `after`; next token from `end_cursor`; stop flag
  `has_next_page`; incremental cursor `created_at`; sent as `created_after`; formatted as
  Unix-seconds timestamp; initial lower bound from `start_date`; computed output fields `content`,
  `message_id`, `role`; fan-out; ids from request `/v3/organizations/{{ config.org_id }}/sessions`;
  id-list records path `items`; id field `session_id`; id inserted into the request path; stamps
  `session_id`.
- `session_attachments`: GET `/v3/organizations/{{ config.org_id }}/sessions/{{ fanout.id
  }}/attachments` - records path `.`; cursor pagination; cursor parameter `after`; next token from
  `end_cursor`; stop flag `has_next_page`; fan-out; ids from request `/v3/organizations/{{
  config.org_id }}/sessions`; id-list records path `items`; id field `session_id`; id inserted into
  the request path; stamps `session_id`.
- `session_tags`: GET `/v3/organizations/{{ config.org_id }}/sessions/{{ fanout.id }}/tags` -
  single-object response; records path `.`; cursor pagination; cursor parameter `after`; next token
  from `end_cursor`; stop flag `has_next_page`; fan-out; ids from request `/v3/organizations/{{
  config.org_id }}/sessions`; id-list records path `items`; id field `session_id`; id inserted into
  the request path; stamps `session_id`.
- `playbooks`: GET `/v3/organizations/{{ config.org_id }}/playbooks` - records path `items`; query
  `first` from template `{{ config.page_size }}`, default `100`; cursor pagination; cursor parameter
  `after`; next token from `end_cursor`; stop flag `has_next_page`; computed output fields
  `description`, `name`.
- `secrets`: GET `/v3/organizations/{{ config.org_id }}/secrets` - records path `items`; query
  `first` from template `{{ config.page_size }}`, default `100`; cursor pagination; cursor parameter
  `after`; next token from `end_cursor`; stop flag `has_next_page`; computed output fields `name`,
  `type`.
- `knowledge_notes`: GET `/v3/organizations/{{ config.org_id }}/knowledge/notes` - records path
  `items`; query `first` from template `{{ config.page_size }}`, default `100`; cursor pagination;
  cursor parameter `after`; next token from `end_cursor`; stop flag `has_next_page`.
- `knowledge_folders`: GET `/v3/organizations/{{ config.org_id }}/knowledge/folders` - records path
  `folders`.
- `repositories`: GET `/v3beta1/organizations/{{ config.org_id }}/repositories` - records path
  `items`; query `filter_name` from template `{{ config.repository_filter_name }}`, omitted when
  absent; `first` from template `{{ config.page_size }}`, default `100`; cursor pagination; cursor
  parameter `after`; next token from `end_cursor`; stop flag `has_next_page`.
- `indexed_repositories`: GET `/v3beta1/organizations/{{ config.org_id }}/repositories/indexing` -
  records path `items`; query `first` from template `{{ config.page_size }}`, default `100`; cursor
  pagination; cursor parameter `after`; next token from `end_cursor`; stop flag `has_next_page`.
- `schedules`: GET `/v3/organizations/{{ config.org_id }}/schedules` - records path `items`;
  offset/limit pagination; offset parameter `offset`; limit parameter `limit`; page size 100.
- `organization_users`: GET `/v3beta1/organizations/{{ config.org_id }}/members/users` - records
  path `items`; query `email` from template `{{ config.user_email }}`, omitted when absent; `first`
  from template `{{ config.page_size }}`, default `100`; cursor pagination; cursor parameter
  `after`; next token from `end_cursor`; stop flag `has_next_page`.
- `organization_idp_group_users`: GET `/v3beta1/organizations/{{ config.org_id }}/members/idp-users`
  - records path `items`; query `email` from template `{{ config.user_email }}`, omitted when
  absent; `first` from template `{{ config.page_size }}`, default `100`; cursor pagination; cursor
  parameter `after`; next token from `end_cursor`; stop flag `has_next_page`.
- `self`: GET `/v3/self` - single-object response; records path `.`.
- `org_daily_consumption`: GET `/v3/organizations/{{ config.org_id }}/consumption/daily` -
  single-object response; records path `.`; query `time_after` from template `{{
  config.metrics_time_after }}`, omitted when absent; `time_before` from template `{{
  config.metrics_time_before }}`, omitted when absent; computed output fields `metric`; emits
  passthrough records.
- `session_daily_consumption`: GET `/v3/organizations/{{ config.org_id
  }}/consumption/daily/sessions/{{ fanout.id }}` - single-object response; records path `.`; query
  `time_after` from template `{{ config.metrics_time_after }}`, omitted when absent; `time_before`
  from template `{{ config.metrics_time_before }}`, omitted when absent; computed output fields
  `metric`; fan-out; ids from request `/v3/organizations/{{ config.org_id }}/sessions`; id-list
  records path `items`; id field `session_id`; id inserted into the request path; stamps
  `session_id`; emits passthrough records.
- `user_daily_consumption`: GET `/v3/organizations/{{ config.org_id }}/consumption/daily/users/{{
  fanout.id }}` - single-object response; records path `.`; query `time_after` from template `{{
  config.metrics_time_after }}`, omitted when absent; `time_before` from template `{{
  config.metrics_time_before }}`, omitted when absent; computed output fields `metric`; fan-out; ids
  from request `/v3beta1/organizations/{{ config.org_id }}/members/users`; id-list records path
  `items`; id field `user_id`; id inserted into the request path; stamps `user_id`; emits
  passthrough records.
- `org_usage_metrics`: GET `/v3/organizations/{{ config.org_id }}/metrics/usage` - single-object
  response; records path `.`; query `time_after` from template `{{ config.metrics_time_after }}`,
  omitted when absent; `time_before` from template `{{ config.metrics_time_before }}`, omitted when
  absent; computed output fields `metric`; emits passthrough records.
- `org_session_metrics`: GET `/v3/organizations/{{ config.org_id }}/metrics/sessions` -
  single-object response; records path `.`; query `time_after` from template `{{
  config.metrics_time_after }}`, omitted when absent; `time_before` from template `{{
  config.metrics_time_before }}`, omitted when absent; computed output fields `metric`; emits
  passthrough records.
- `org_active_users_metrics`: GET `/v3/organizations/{{ config.org_id }}/metrics/active-users` -
  single-object response; records path `.`; query `time_after` from template `{{
  config.metrics_time_after }}`, omitted when absent; `time_before` from template `{{
  config.metrics_time_before }}`, omitted when absent; computed output fields `metric`; emits
  passthrough records.
- `org_daily_active_users`: GET `/v3/organizations/{{ config.org_id }}/metrics/dau` - single-object
  response; records path `.`; query `time_after` from template `{{ config.metrics_time_after }}`,
  omitted when absent; `time_before` from template `{{ config.metrics_time_before }}`, omitted when
  absent; computed output fields `metric`; emits passthrough records.
- `org_monthly_active_users`: GET `/v3/organizations/{{ config.org_id }}/metrics/mau` -
  single-object response; records path `.`; query `time_after` from template `{{
  config.metrics_time_after }}`, omitted when absent; `time_before` from template `{{
  config.metrics_time_before }}`, omitted when absent; computed output fields `metric`; emits
  passthrough records.
- `org_pr_metrics`: GET `/v3/organizations/{{ config.org_id }}/metrics/prs` - single-object
  response; records path `.`; query `time_after` from template `{{ config.metrics_time_after }}`,
  omitted when absent; `time_before` from template `{{ config.metrics_time_before }}`, omitted when
  absent; computed output fields `metric`; emits passthrough records.
- `org_search_metrics`: GET `/v3/organizations/{{ config.org_id }}/metrics/searches` - single-object
  response; records path `.`; query `time_after` from template `{{ config.metrics_time_after }}`,
  omitted when absent; `time_before` from template `{{ config.metrics_time_before }}`, omitted when
  absent; computed output fields `metric`; emits passthrough records.
- `org_weekly_active_users`: GET `/v3/organizations/{{ config.org_id }}/metrics/wau` - single-object
  response; records path `.`; query `time_after` from template `{{ config.metrics_time_after }}`,
  omitted when absent; `time_before` from template `{{ config.metrics_time_before }}`, omitted when
  absent; computed output fields `metric`; emits passthrough records.

## Write actions & risks

Overall write risk: creates or mutates Devin sessions, session tags/messages, schedules, playbooks,
knowledge notes, repository indexing state, and PR reviews; destructive actions can terminate
sessions or delete objects.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_session`: POST `/v3/organizations/{{ config.org_id }}/sessions` - kind `create`; body type
  `json`; required record fields `prompt`; accepted fields `attachments`, `create_as_user_id`,
  `devin_mode`, `idempotent`, `playbook_id`, `prompt`, `snapshot_id`, `structured_output_required`,
  `structured_output_schema`, `tags`, `unlisted`; risk: creates a new Devin session in the
  organization and can consume ACUs.
- `send_session_message`: POST `/v3/organizations/{{ config.org_id }}/sessions/{{ record.devin_id
  }}/messages` - kind `create`; body type `json`; path fields `devin_id`; required record fields
  `devin_id`, `message`; accepted fields `attachment_urls`, `devin_id`, `message`,
  `message_as_user_id`; risk: sends a message to an active or suspended Devin session and may resume
  work.
- `append_session_tags`: POST `/v3/organizations/{{ config.org_id }}/sessions/{{ record.devin_id
  }}/tags` - kind `update`; body type `json`; path fields `devin_id`; required record fields
  `devin_id`, `tags`; accepted fields `devin_id`, `tags`; risk: adds tags to a Devin session.
- `replace_session_tags`: PUT `/v3/organizations/{{ config.org_id }}/sessions/{{ record.devin_id
  }}/tags` - kind `update`; body type `json`; path fields `devin_id`; required record fields
  `devin_id`, `tags`; accepted fields `devin_id`, `tags`; risk: replaces all tags on a Devin
  session.
- `archive_session`: POST `/v3/organizations/{{ config.org_id }}/sessions/{{ record.devin_id
  }}/archive` - kind `update`; body type `none`; path fields `devin_id`; required record fields
  `devin_id`; accepted fields `devin_id`; confirmation `destructive`; risk: archives a Devin session
  and puts it to sleep if currently running.
- `terminate_session`: DELETE `/v3/organizations/{{ config.org_id }}/sessions/{{ record.devin_id }}`
  - kind `delete`; body type `none`; path fields `devin_id`; required record fields `devin_id`;
  accepted fields `devin_id`; confirmation `destructive`; risk: terminates a Devin session.
- `generate_session_insights`: POST `/v3/organizations/{{ config.org_id }}/sessions/{{
  record.devin_id }}/insights/generate` - kind `custom`; body type `none`; path fields `devin_id`;
  required record fields `devin_id`; accepted fields `devin_id`; risk: triggers on-demand generation
  of session insights.
- `create_schedule`: POST `/v3/organizations/{{ config.org_id }}/schedules` - kind `create`; body
  type `json`; required record fields `name`, `prompt`; accepted fields `agent`,
  `create_as_user_id`, `enabled`, `frequency`, `name`, `notify_on`, `playbook_id`, `prompt`,
  `schedule_type`, `scheduled_at`, `tags`; risk: creates a scheduled Devin session that can run
  automatically.
- `update_schedule`: PATCH `/v3/organizations/{{ config.org_id }}/schedules/{{ record.schedule_id
  }}` - kind `update`; body type `json`; path fields `schedule_id`; required record fields
  `schedule_id`; accepted fields `agent`, `enabled`, `frequency`, `name`, `notify_on`,
  `playbook_id`, `prompt`, `run_as_user_id`, `schedule_id`, `schedule_type`, `scheduled_at`, `tags`;
  risk: updates an existing scheduled Devin session.
- `delete_schedule`: DELETE `/v3/organizations/{{ config.org_id }}/schedules/{{ record.schedule_id
  }}` - kind `delete`; body type `none`; path fields `schedule_id`; required record fields
  `schedule_id`; accepted fields `schedule_id`; confirmation `destructive`; risk: soft-deletes a
  schedule.
- `create_playbook`: POST `/v3/organizations/{{ config.org_id }}/playbooks` - kind `create`; body
  type `json`; required record fields `title`, `body`; accepted fields `body`, `macro`,
  `structured_output_schema`, `title`; risk: creates an organization-level Devin playbook.
- `update_playbook`: PUT `/v3/organizations/{{ config.org_id }}/playbooks/{{ record.playbook_id }}`
  - kind `update`; body type `json`; path fields `playbook_id`; required record fields
  `playbook_id`, `title`, `body`; accepted fields `body`, `macro`, `playbook_id`,
  `structured_output_schema`, `title`; risk: replaces an organization-level Devin playbook.
- `delete_playbook`: DELETE `/v3/organizations/{{ config.org_id }}/playbooks/{{ record.playbook_id
  }}` - kind `delete`; body type `none`; path fields `playbook_id`; required record fields
  `playbook_id`; accepted fields `playbook_id`; confirmation `destructive`; risk: deletes an
  organization-level Devin playbook.
- `create_knowledge_note`: POST `/v3/organizations/{{ config.org_id }}/knowledge/notes` - kind
  `create`; body type `json`; required record fields `name`, `body`; accepted fields `body`,
  `folder_id`, `folder_path`, `is_enabled`, `macro`, `name`, `pinned_repo`, `trigger`; risk: creates
  an organization-level Devin knowledge note.
- `update_knowledge_note`: PUT `/v3/organizations/{{ config.org_id }}/knowledge/notes/{{
  record.note_id }}` - kind `update`; body type `json`; path fields `note_id`; required record
  fields `note_id`, `name`, `body`; accepted fields `body`, `folder_id`, `folder_path`,
  `is_enabled`, `macro`, `name`, `note_id`, `pinned_repo`, `trigger`; risk: replaces an
  organization-level Devin knowledge note.
- `delete_knowledge_note`: DELETE `/v3/organizations/{{ config.org_id }}/knowledge/notes/{{
  record.note_id }}` - kind `delete`; body type `none`; path fields `note_id`; required record
  fields `note_id`; accepted fields `note_id`; confirmation `destructive`; risk: deletes an
  organization-level Devin knowledge note.
- `index_repository`: PUT `/v3beta1/organizations/{{ config.org_id }}/repositories/{{
  record.encoded_repository_path }}/indexing` - kind `upsert`; body type `json`; path fields
  `encoded_repository_path`; required record fields `encoded_repository_path`; accepted fields
  `branch_names`, `encoded_repository_path`; risk: enables indexing for a repository and can trigger
  indexing jobs.
- `bulk_index_repositories`: PUT `/v3beta1/organizations/{{ config.org_id }}/repositories/indexing`
  - kind `upsert`; body type `json`; required record fields `repositories`; accepted fields
  `repositories`; risk: enables indexing for multiple repositories and can trigger indexing jobs.
- `remove_repository_indexing`: DELETE `/v3beta1/organizations/{{ config.org_id }}/repositories/{{
  record.encoded_repository_path }}/indexing` - kind `delete`; body type `none`; path fields
  `encoded_repository_path`; required record fields `encoded_repository_path`; accepted fields
  `encoded_repository_path`; confirmation `destructive`; risk: disables indexing and clears
  configured branches for a repository.
- `bulk_remove_repository_indexing`: DELETE `/v3beta1/organizations/{{ config.org_id
  }}/repositories/indexing` - kind `delete`; body type `json`; body fields `repository_paths`;
  required record fields `repository_paths`; accepted fields `repository_paths`; confirmation
  `destructive`; risk: disables indexing and clears configured branches for multiple repositories.
- `remove_repository_branch_indexing`: DELETE `/v3beta1/organizations/{{ config.org_id
  }}/repositories/{{ record.encoded_repository_path }}/indexing/branches/{{
  record.encoded_branch_name }}` - kind `delete`; body type `none`; path fields
  `encoded_repository_path`, `encoded_branch_name`; required record fields
  `encoded_repository_path`, `encoded_branch_name`; accepted fields `encoded_branch_name`,
  `encoded_repository_path`; confirmation `destructive`; risk: removes one branch from repository
  indexing and can disable indexing if no branches remain.
- `trigger_pr_review`: POST `/v3/organizations/{{ config.org_id }}/pr-reviews` - kind `create`; body
  type `json`; required record fields `pr_url`; accepted fields `pr_url`; risk: triggers a Devin
  Review for a pull or merge request.
- `delete_secret`: DELETE `/v3/organizations/{{ config.org_id }}/secrets/{{ record.secret_id }}` -
  kind `delete`; body type `none`; path fields `secret_id`; required record fields `secret_id`;
  accepted fields `secret_id`; confirmation `destructive`; risk: deletes Devin secret metadata and
  its stored value from the organization.

## Known limits

- Batch defaults: read_page_size=100, write_batch_size=1.
- API coverage includes 27 stream-backed endpoint group(s), 23 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=7, deprecated=2, destructive_admin=22, duplicate_of=6, non_data_endpoint=1,
  requires_elevated_scope=103.
