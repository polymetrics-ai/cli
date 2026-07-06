# Overview

Reads n8n workflows, executions, tags, users, variables, projects, data tables, and credential
metadata; writes supported n8n public REST API mutations.

Readable streams: `workflows`, `workflow`, `workflow_version`, `workflow_tags`, `executions`,
`execution`, `execution_tags`, `tags`, `tag`, `users`, `user`, `variables`, `projects`,
`project_members`, `credentials`, `credential`, `credential_schema`, `data_tables`, `data_table`,
`data_table_rows`, `data_table_columns`.

Write actions: `create_workflow`, `update_workflow`, `publish_workflow`, `deactivate_workflow`,
`archive_workflow`, `unarchive_workflow`, `transfer_workflow`, `retry_execution`, `stop_execution`,
`stop_executions`, `create_tag`, `update_tag`, `create_variable`, `update_variable`,
`create_project`, `update_project`, `add_project_users`, `change_project_user_role`,
`create_data_table`, `update_data_table`, `insert_data_table_rows`, `update_data_table_rows`,
`upsert_data_table_row`, `add_data_table_column`, `update_data_table_column`, `test_credential`,
`transfer_credential`, `pull_source_control`, `generate_audit`.

Service API documentation: https://docs.n8n.io/connect/n8n-api/api-reference.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); n8n API key, sent as the X-N8N-API-KEY header. Never logged.
- `base_url` (required, string); format `uri`; Fully-qualified n8n public REST API base URL,
  including the /api/v1 version path (e.g. https://your-instance.n8n.cloud/api/v1).
- `credential_id` (optional, string); Credential id used by credential detail, test, and transfer
  actions.
- `credential_type_name` (optional, string); Credential type name used by the credential_schema
  stream.
- `data_table_column_id` (optional, string); Data table column id used by column write actions.
- `data_table_filter` (optional, string); Optional encoded filter value for data table list/row
  reads.
- `data_table_id` (optional, string); Data table id used by data table detail, row, column, and
  write actions.
- `data_table_search` (optional, string); Optional search value for data table row reads.
- `data_table_sort_by` (optional, string); Optional sortBy value for data table list/row reads.
- `exclude_pinned_data` (optional, string); Optional n8n excludePinnedData flag for workflow reads;
  use true or false.
- `execution_id` (optional, string); Execution id used by execution detail/tag streams and execution
  write actions.
- `execution_status` (optional, string); Optional status filter for the executions list stream.
- `ignore_data_size_limit` (optional, string); Optional ignoreDataSizeLimit flag for execution
  reads; use true or false.
- `include_execution_data` (optional, string); Optional includeData flag for execution reads; use
  true or false.
- `include_role` (optional, string); Optional includeRole flag for user reads; use true or false.
- `max_pages` (optional, string); default `0`; Maximum pages; use 0, all, or unlimited to exhaust
  the stream.
- `mode` (optional, string).
- `page_size` (optional, string); default `100`; Records per page (1-250).
- `project_id` (optional, string); Project id used by project-scoped streams, filters, and project
  write actions.
- `project_user_id` (optional, string); User id used by project membership write actions.
- `redact_execution_data` (optional, string); Optional redactExecutionData flag for execution reads;
  use true or false.
- `tag_id` (optional, string); Tag id used by the tag detail stream and tag write actions.
- `user_id` (optional, string); User id or email used by the user detail stream.
- `variable_id` (optional, string); Variable id used by variable write actions.
- `variable_state` (optional, string); Optional state filter for the variables list stream.
- `workflow_active` (optional, string); Optional active filter for the workflows list stream; use
  true or false.
- `workflow_id` (optional, string); Workflow id used by workflow detail/version/tag streams and
  workflow write actions.
- `workflow_name` (optional, string); Optional workflow name filter for the workflows list stream.
- `workflow_tags` (optional, string); Optional comma-separated tag filter for the workflows list
  stream.
- `workflow_version_id` (optional, string); Workflow version id used by the workflow_version stream.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `max_pages=0`, `page_size=100`.

Authentication behavior:

- API key authentication in `X-N8N-API-KEY` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/workflows`.

## Streams notes

Default pagination: cursor pagination; cursor parameter `cursor`; next token from `nextCursor`.

Pagination by stream: cursor: `workflows`, `executions`, `tags`, `users`, `variables`, `projects`,
`project_members`, `credentials`, `data_tables`, `data_table_rows`; none: `workflow`,
`workflow_version`, `workflow_tags`, `execution`, `execution_tags`, `tag`, `user`, `credential`,
`credential_schema`, `data_table`, `data_table_columns`.

- `workflows`: GET `/workflows` - records path `data`; query `active` from template `{{
  config.workflow_active }}`, omitted when absent; `excludePinnedData` from template `{{
  config.exclude_pinned_data }}`, omitted when absent; `limit` from template `{{ config.page_size
  }}`, default `100`; `name` from template `{{ config.workflow_name }}`, omitted when absent;
  `projectId` from template `{{ config.project_id }}`, omitted when absent; `tags` from template `{{
  config.workflow_tags }}`, omitted when absent; cursor pagination; cursor parameter `cursor`; next
  token from `nextCursor`.
- `workflow`: GET `/workflows/{{ config.workflow_id }}` - single-object response; records path `.`;
  query `excludePinnedData` from template `{{ config.exclude_pinned_data }}`, omitted when absent;
  emits passthrough records.
- `workflow_version`: GET `/workflows/{{ config.workflow_id }}/{{ config.workflow_version_id }}` -
  single-object response; records path `.`; emits passthrough records.
- `workflow_tags`: GET `/workflows/{{ config.workflow_id }}/tags` - records path `.`.
- `executions`: GET `/executions` - records path `data`; query `ignoreDataSizeLimit` from template
  `{{ config.ignore_data_size_limit }}`, omitted when absent; `includeData` from template `{{
  config.include_execution_data }}`, omitted when absent; `limit` from template `{{ config.page_size
  }}`, default `100`; `projectId` from template `{{ config.project_id }}`, omitted when absent;
  `redactExecutionData` from template `{{ config.redact_execution_data }}`, omitted when absent;
  `status` from template `{{ config.execution_status }}`, omitted when absent; `workflowId` from
  template `{{ config.workflow_id }}`, omitted when absent; cursor pagination; cursor parameter
  `cursor`; next token from `nextCursor`.
- `execution`: GET `/executions/{{ config.execution_id }}` - single-object response; records path
  `.`; query `ignoreDataSizeLimit` from template `{{ config.ignore_data_size_limit }}`, omitted when
  absent; `includeData` from template `{{ config.include_execution_data }}`, omitted when absent;
  `redactExecutionData` from template `{{ config.redact_execution_data }}`, omitted when absent;
  emits passthrough records.
- `execution_tags`: GET `/executions/{{ config.execution_id }}/tags` - records path `.`.
- `tags`: GET `/tags` - records path `data`; query `limit` from template `{{ config.page_size }}`,
  default `100`; cursor pagination; cursor parameter `cursor`; next token from `nextCursor`.
- `tag`: GET `/tags/{{ config.tag_id }}` - single-object response; records path `.`.
- `users`: GET `/users` - records path `data`; query `includeRole` from template `{{
  config.include_role }}`, omitted when absent; `limit` from template `{{ config.page_size }}`,
  default `100`; `projectId` from template `{{ config.project_id }}`, omitted when absent; cursor
  pagination; cursor parameter `cursor`; next token from `nextCursor`.
- `user`: GET `/users/{{ config.user_id }}` - single-object response; records path `.`; query
  `includeRole` from template `{{ config.include_role }}`, omitted when absent; emits passthrough
  records.
- `variables`: GET `/variables` - records path `data`; query `limit` from template `{{
  config.page_size }}`, default `100`; `projectId` from template `{{ config.project_id }}`, omitted
  when absent; `state` from template `{{ config.variable_state }}`, omitted when absent; cursor
  pagination; cursor parameter `cursor`; next token from `nextCursor`.
- `projects`: GET `/projects` - records path `data`; query `limit` from template `{{
  config.page_size }}`, default `100`; cursor pagination; cursor parameter `cursor`; next token from
  `nextCursor`.
- `project_members`: GET `/projects/{{ config.project_id }}/users` - records path `data`; query
  `limit` from template `{{ config.page_size }}`, default `100`; cursor pagination; cursor parameter
  `cursor`; next token from `nextCursor`.
- `credentials`: GET `/credentials` - records path `data`; query `limit` from template `{{
  config.page_size }}`, default `100`; cursor pagination; cursor parameter `cursor`; next token from
  `nextCursor`.
- `credential`: GET `/credentials/{{ config.credential_id }}` - single-object response; records path
  `.`; emits passthrough records.
- `credential_schema`: GET `/credentials/schema/{{ config.credential_type_name }}` - single-object
  response; records path `.`; emits passthrough records.
- `data_tables`: GET `/data-tables` - records path `data`; query `filter` from template `{{
  config.data_table_filter }}`, omitted when absent; `limit` from template `{{ config.page_size }}`,
  default `100`; `sortBy` from template `{{ config.data_table_sort_by }}`, omitted when absent;
  cursor pagination; cursor parameter `cursor`; next token from `nextCursor`.
- `data_table`: GET `/data-tables/{{ config.data_table_id }}` - single-object response; records path
  `.`; emits passthrough records.
- `data_table_rows`: GET `/data-tables/{{ config.data_table_id }}/rows` - records path `data`; query
  `filter` from template `{{ config.data_table_filter }}`, omitted when absent; `limit` from
  template `{{ config.page_size }}`, default `100`; `search` from template `{{
  config.data_table_search }}`, omitted when absent; `sortBy` from template `{{
  config.data_table_sort_by }}`, omitted when absent; cursor pagination; cursor parameter `cursor`;
  next token from `nextCursor`; emits passthrough records.
- `data_table_columns`: GET `/data-tables/{{ config.data_table_id }}/columns` - records path `.`.

## Write actions & risks

Overall write risk: external n8n instance API mutation of workflows, executions, tags, variables,
projects, data tables, source-control pull, audit generation, and credential tests/transfers.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_workflow`: POST `/workflows` - kind `create`; body type `json`; required record fields
  `name`, `nodes`, `connections`, `settings`; accepted fields `connections`, `description`, `name`,
  `nodes`, `projectId`, `settings`; risk: external mutation; approval required.
- `update_workflow`: PUT `/workflows/{{ record.workflow_id }}` - kind `update`; body type `json`;
  path fields `workflow_id`; required record fields `workflow_id`, `name`, `nodes`, `connections`,
  `settings`; accepted fields `connections`, `description`, `name`, `nodes`, `settings`,
  `versionId`, `workflow_id`; risk: external mutation; approval required.
- `publish_workflow`: POST `/workflows/{{ record.workflow_id }}/activate` - kind `update`; body type
  `json`; path fields `workflow_id`; required record fields `workflow_id`; accepted fields
  `description`, `name`, `versionId`, `workflow_id`; risk: external mutation; approval required.
- `deactivate_workflow`: POST `/workflows/{{ record.workflow_id }}/deactivate` - kind `update`; body
  type `none`; path fields `workflow_id`; required record fields `workflow_id`; accepted fields
  `workflow_id`; risk: external mutation; approval required.
- `archive_workflow`: POST `/workflows/{{ record.workflow_id }}/archive` - kind `update`; body type
  `none`; path fields `workflow_id`; required record fields `workflow_id`; accepted fields
  `workflow_id`; risk: external mutation; approval required.
- `unarchive_workflow`: POST `/workflows/{{ record.workflow_id }}/unarchive` - kind `update`; body
  type `none`; path fields `workflow_id`; required record fields `workflow_id`; accepted fields
  `workflow_id`; risk: external mutation; approval required.
- `transfer_workflow`: PUT `/workflows/{{ record.workflow_id }}/transfer` - kind `update`; body type
  `json`; path fields `workflow_id`; required record fields `workflow_id`, `destinationProjectId`;
  accepted fields `destinationProjectId`, `workflow_id`; risk: external mutation; approval required.
- `retry_execution`: POST `/executions/{{ record.execution_id }}/retry` - kind `update`; body type
  `json`; path fields `execution_id`; required record fields `execution_id`; accepted fields
  `execution_id`, `loadWorkflow`; risk: external mutation; approval required.
- `stop_execution`: POST `/executions/{{ record.execution_id }}/stop` - kind `update`; body type
  `none`; path fields `execution_id`; required record fields `execution_id`; accepted fields
  `execution_id`; risk: external mutation; approval required.
- `stop_executions`: POST `/executions/stop` - kind `update`; body type `json`; required record
  fields `status`; accepted fields `startedAfter`, `startedBefore`, `status`, `workflowId`; risk:
  external mutation; approval required.
- `create_tag`: POST `/tags` - kind `create`; body type `json`; required record fields `name`;
  accepted fields `name`; risk: external mutation; approval required.
- `update_tag`: PUT `/tags/{{ record.tag_id }}` - kind `update`; body type `json`; path fields
  `tag_id`; required record fields `tag_id`, `name`; accepted fields `name`, `tag_id`; risk:
  external mutation; approval required.
- `create_variable`: POST `/variables` - kind `create`; body type `json`; required record fields
  `key`, `value`; accepted fields `key`, `projectId`, `value`; risk: external mutation; approval
  required.
- `update_variable`: PUT `/variables/{{ record.variable_id }}` - kind `update`; body type `json`;
  path fields `variable_id`; required record fields `variable_id`, `key`, `value`; accepted fields
  `key`, `projectId`, `value`, `variable_id`; risk: external mutation; approval required.
- `create_project`: POST `/projects` - kind `create`; body type `json`; required record fields
  `name`; accepted fields `name`; risk: external mutation; approval required.
- `update_project`: PUT `/projects/{{ record.project_id }}` - kind `update`; body type `json`; path
  fields `project_id`; required record fields `project_id`, `name`; accepted fields `name`,
  `project_id`; risk: external mutation; approval required.
- `add_project_users`: POST `/projects/{{ record.project_id }}/users` - kind `update`; body type
  `json`; path fields `project_id`; required record fields `project_id`, `relations`; accepted
  fields `project_id`, `relations`; risk: external mutation; approval required.
- `change_project_user_role`: PATCH `/projects/{{ record.project_id }}/users/{{
  record.project_user_id }}` - kind `update`; body type `json`; path fields `project_id`,
  `project_user_id`; required record fields `project_id`, `project_user_id`, `role`; accepted fields
  `project_id`, `project_user_id`, `role`; risk: external mutation; approval required.
- `create_data_table`: POST `/data-tables` - kind `create`; body type `json`; required record fields
  `name`, `columns`; accepted fields `columns`, `name`, `projectId`; risk: external mutation;
  approval required.
- `update_data_table`: PATCH `/data-tables/{{ record.data_table_id }}` - kind `update`; body type
  `json`; path fields `data_table_id`; required record fields `data_table_id`, `name`; accepted
  fields `data_table_id`, `name`; risk: external mutation; approval required.
- `insert_data_table_rows`: POST `/data-tables/{{ record.data_table_id }}/rows` - kind `create`;
  body type `json`; path fields `data_table_id`; required record fields `data_table_id`, `data`;
  accepted fields `data`, `data_table_id`, `returnType`; risk: external mutation; approval required.
- `update_data_table_rows`: PATCH `/data-tables/{{ record.data_table_id }}/rows/update` - kind
  `update`; body type `json`; path fields `data_table_id`; required record fields `data_table_id`,
  `filter`, `data`; accepted fields `data`, `data_table_id`, `dryRun`, `filter`, `returnData`; risk:
  external mutation; approval required.
- `upsert_data_table_row`: POST `/data-tables/{{ record.data_table_id }}/rows/upsert` - kind
  `upsert`; body type `json`; path fields `data_table_id`; required record fields `data_table_id`,
  `filter`, `data`; accepted fields `data`, `data_table_id`, `dryRun`, `filter`, `returnData`; risk:
  external mutation; approval required.
- `add_data_table_column`: POST `/data-tables/{{ record.data_table_id }}/columns` - kind `create`;
  body type `json`; path fields `data_table_id`; required record fields `data_table_id`, `name`,
  `type`; accepted fields `data_table_id`, `index`, `name`, `type`; risk: external mutation;
  approval required.
- `update_data_table_column`: PATCH `/data-tables/{{ record.data_table_id }}/columns/{{
  record.data_table_column_id }}` - kind `update`; body type `json`; path fields `data_table_id`,
  `data_table_column_id`; required record fields `data_table_id`, `data_table_column_id`, `name`;
  accepted fields `data_table_column_id`, `data_table_id`, `index`, `name`; risk: external mutation;
  approval required.
- `test_credential`: POST `/credentials/{{ record.credential_id }}/test` - kind `update`; body type
  `none`; path fields `credential_id`; required record fields `credential_id`; accepted fields
  `credential_id`; risk: external mutation; approval required.
- `transfer_credential`: PUT `/credentials/{{ record.credential_id }}/transfer` - kind `update`;
  body type `json`; path fields `credential_id`; required record fields `credential_id`,
  `destinationProjectId`; accepted fields `credential_id`, `destinationProjectId`; risk: external
  mutation; approval required.
- `pull_source_control`: POST `/source-control/pull` - kind `update`; body type `json`; accepted
  fields `autoPublish`, `force`; risk: external mutation; approval required.
- `generate_audit`: POST `/audit` - kind `create`; body type `json`; accepted fields
  `additionalOptions`; risk: external mutation; approval required.

## Known limits

- Batch defaults: read_page_size=100, write_batch_size=100.
- API coverage includes 21 stream-backed endpoint group(s), 29 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  destructive_admin=10, out_of_scope=3, requires_elevated_scope=4.
