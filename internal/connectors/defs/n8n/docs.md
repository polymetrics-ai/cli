# Overview

n8n is a workflow automation platform. This Pass B bundle covers the current n8n public REST API
reference under `/api/v1`: workflows, executions, tags, users, variables, projects, data tables,
credential metadata/schema, credential test/transfer operations, source-control pull, and audit
generation. The hand-written legacy connector remains registered until wave6's registry flip.

## Auth setup

Provide the n8n API key via the `api_key` secret; it is sent as the `X-N8N-API-KEY` header on every
request and is never logged, matching legacy's `connsdk.APIKeyHeader("X-N8N-API-KEY", ...)` wiring.

`base_url` must be the fully qualified API URL including `/api/v1`, for example
`https://your-instance.n8n.cloud/api/v1`.

## Streams notes

Cursor-paginated collection streams use n8n's `{"data":[...],"nextCursor":"..."}` envelope with
`cursor_param: "cursor"` and `token_path: "nextCursor"`: `workflows`, `executions`, `tags`, `users`,
`variables`, `projects`, `project_members`, `credentials`, `data_tables`, and `data_table_rows`.
Each request sends `limit={{ config.page_size | default 100 }}`.

Detail and array-root streams are non-paginated: `workflow`, `workflow_version`, `workflow_tags`,
`execution`, `execution_tags`, `tag`, `user`, `credential`, `credential_schema`, `data_table`, and
`data_table_columns`. The new detail streams use `projection: "passthrough"` where n8n returns
nested workflow, execution, credential-schema, table, or row objects beyond the narrow legacy
projection.

The four legacy streams keep their existing record DATA shape. New streams follow the documented
n8n response field names directly.

## Write actions & risks

The bundle exposes object-body or no-body mutations that the current write dialect can express:

workflow lifecycle (`create_workflow`, `update_workflow`, `publish_workflow`, `deactivate_workflow`,
`archive_workflow`, `unarchive_workflow`, `transfer_workflow`), execution controls
(`retry_execution`, `stop_execution`, `stop_executions`), tag and variable upserts, project and
project-membership writes, data-table/table-row/column writes, `test_credential`,
`transfer_credential`, `pull_source_control`, and `generate_audit`.

All write actions require operator approval. Credential create/update are intentionally excluded
because their request bodies carry write-only credential data; connector rules prohibit requesting
or storing secret values in ordinary write records or fixtures.

## Known limits

- `base_url` is still required and must include `/api/v1`. Legacy accepted a bare `host` and
  appended `/api/v1`; the declarative engine has no derived-default string concatenation.
- `max_pages` remains declared for legacy compatibility but is not modeled by the declarative
  cursor paginator.
- Full write coverage has two typed gaps recorded in `docs/migration/quarantine.json`: n8n's
  workflow/execution tag update endpoints require a root JSON array body, and data-table row delete
  requires query parameters on a DELETE request. `writes.json` can construct object bodies but has
  no root-array body mode or write-side query field.
- Destructive deletes and elevated user/credential administration writes are excluded in
  `api_surface.json` with closed categories and concrete reasons.
