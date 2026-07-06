---
name: pm-n8n
description: n8n connector knowledge and safe action guide.
---

# pm-n8n

## Purpose

Reads n8n workflows, executions, tags, users, variables, projects, data tables, and credential metadata; writes supported n8n public REST API mutations.

## Icon

- asset: icons/n8n.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://docs.n8n.io/api/

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- credential_id
- credential_type_name
- data_table_column_id
- data_table_filter
- data_table_id
- data_table_search
- data_table_sort_by
- exclude_pinned_data
- execution_id
- execution_status
- ignore_data_size_limit
- include_execution_data
- include_role
- max_pages
- mode
- page_size
- project_id
- project_user_id
- redact_execution_data
- tag_id
- user_id
- variable_id
- variable_state
- workflow_active
- workflow_id
- workflow_name
- workflow_tags
- workflow_version_id
- api_key (secret)

## ETL Streams

- workflows:
  - primary key: id
  - cursor: updatedAt
  - fields: active(), createdAt(), id(), isArchived(), name(), triggerCount(), updatedAt(), versionId()
- workflow:
  - primary key: id
  - cursor: updatedAt
  - fields: active(), connections(), createdAt(), description(), id(), isArchived(), name(), nodes(), settings(), shared(), tags(), triggerCount(), updatedAt(), versionId()
- workflow_version:
  - primary key: versionId
  - cursor: updatedAt
  - fields: authors(), connections(), createdAt(), description(), name(), nodeGroups(), nodes(), updatedAt(), versionId(), workflowId()
- workflow_tags:
  - primary key: id
  - cursor: updatedAt
  - fields: createdAt(), id(), name(), updatedAt()
- executions:
  - primary key: id
  - cursor: startedAt
  - fields: finished(), id(), mode(), retryOf(), startedAt(), status(), stoppedAt(), workflowId()
- execution:
  - primary key: id
  - cursor: startedAt
  - fields: customData(), data(), finished(), id(), mode(), retryOf(), retrySuccessId(), startedAt(), status(), stoppedAt(), waitTill(), workflowId()
- execution_tags:
  - primary key: id
  - cursor: updatedAt
  - fields: createdAt(), id(), name(), updatedAt()
- tags:
  - primary key: id
  - cursor: updatedAt
  - fields: createdAt(), id(), name(), updatedAt()
- tag:
  - primary key: id
  - cursor: updatedAt
  - fields: createdAt(), id(), name(), updatedAt()
- users:
  - primary key: id
  - cursor: updatedAt
  - fields: createdAt(), email(), firstName(), id(), isPending(), lastName(), role(), updatedAt()
- user:
  - primary key: id
  - cursor: updatedAt
  - fields: createdAt(), email(), firstName(), id(), isPending(), lastName(), mfaEnabled(), role(), updatedAt()
- variables:
  - primary key: id
  - fields: id(), key(), project(), projectId(), type(), value()
- projects:
  - primary key: id
  - fields: id(), name(), type()
- project_members:
  - primary key: id
  - fields: createdAt(), email(), firstName(), id(), lastName(), role(), updatedAt()
- credentials:
  - primary key: id
  - cursor: updatedAt
  - fields: createdAt(), id(), isGlobal(), isManaged(), isResolvable(), name(), resolvableAllowFallback(), resolverId(), type(), updatedAt()
- credential:
  - primary key: id
  - cursor: updatedAt
  - fields: createdAt(), id(), isGlobal(), isManaged(), isResolvable(), name(), resolvableAllowFallback(), resolverId(), type(), updatedAt()
- credential_schema:
  - fields: displayName(), documentationUrl(), name(), properties()
- data_tables:
  - primary key: id
  - cursor: updatedAt
  - fields: columns(), createdAt(), id(), name(), projectId(), updatedAt()
- data_table:
  - primary key: id
  - cursor: updatedAt
  - fields: columns(), createdAt(), id(), name(), projectId(), updatedAt()
- data_table_rows:
  - primary key: id
  - cursor: updatedAt
  - fields: createdAt(), id(), updatedAt()
- data_table_columns:
  - primary key: id
  - fields: dataTableId(), id(), index(), name(), type()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- create_workflow:
  - endpoint: POST /workflows
  - risk: external mutation; approval required
- update_workflow:
  - endpoint: PUT /workflows/{{ record.workflow_id }}
  - required fields: workflow_id
  - risk: external mutation; approval required
- publish_workflow:
  - endpoint: POST /workflows/{{ record.workflow_id }}/activate
  - required fields: workflow_id
  - risk: external mutation; approval required
- deactivate_workflow:
  - endpoint: POST /workflows/{{ record.workflow_id }}/deactivate
  - required fields: workflow_id
  - risk: external mutation; approval required
- archive_workflow:
  - endpoint: POST /workflows/{{ record.workflow_id }}/archive
  - required fields: workflow_id
  - risk: external mutation; approval required
- unarchive_workflow:
  - endpoint: POST /workflows/{{ record.workflow_id }}/unarchive
  - required fields: workflow_id
  - risk: external mutation; approval required
- transfer_workflow:
  - endpoint: PUT /workflows/{{ record.workflow_id }}/transfer
  - required fields: workflow_id
  - risk: external mutation; approval required
- retry_execution:
  - endpoint: POST /executions/{{ record.execution_id }}/retry
  - required fields: execution_id
  - risk: external mutation; approval required
- stop_execution:
  - endpoint: POST /executions/{{ record.execution_id }}/stop
  - required fields: execution_id
  - risk: external mutation; approval required
- stop_executions:
  - endpoint: POST /executions/stop
  - risk: external mutation; approval required
- create_tag:
  - endpoint: POST /tags
  - risk: external mutation; approval required
- update_tag:
  - endpoint: PUT /tags/{{ record.tag_id }}
  - required fields: tag_id
  - risk: external mutation; approval required
- create_variable:
  - endpoint: POST /variables
  - risk: external mutation; approval required
- update_variable:
  - endpoint: PUT /variables/{{ record.variable_id }}
  - required fields: variable_id
  - risk: external mutation; approval required
- create_project:
  - endpoint: POST /projects
  - risk: external mutation; approval required
- update_project:
  - endpoint: PUT /projects/{{ record.project_id }}
  - required fields: project_id
  - risk: external mutation; approval required
- add_project_users:
  - endpoint: POST /projects/{{ record.project_id }}/users
  - required fields: project_id
  - risk: external mutation; approval required
- change_project_user_role:
  - endpoint: PATCH /projects/{{ record.project_id }}/users/{{ record.project_user_id }}
  - required fields: project_id, project_user_id
  - risk: external mutation; approval required
- create_data_table:
  - endpoint: POST /data-tables
  - risk: external mutation; approval required
- update_data_table:
  - endpoint: PATCH /data-tables/{{ record.data_table_id }}
  - required fields: data_table_id
  - risk: external mutation; approval required
- insert_data_table_rows:
  - endpoint: POST /data-tables/{{ record.data_table_id }}/rows
  - required fields: data_table_id
  - risk: external mutation; approval required
- update_data_table_rows:
  - endpoint: PATCH /data-tables/{{ record.data_table_id }}/rows/update
  - required fields: data_table_id
  - risk: external mutation; approval required
- upsert_data_table_row:
  - endpoint: POST /data-tables/{{ record.data_table_id }}/rows/upsert
  - required fields: data_table_id
  - risk: external mutation; approval required
- add_data_table_column:
  - endpoint: POST /data-tables/{{ record.data_table_id }}/columns
  - required fields: data_table_id
  - risk: external mutation; approval required
- update_data_table_column:
  - endpoint: PATCH /data-tables/{{ record.data_table_id }}/columns/{{ record.data_table_column_id }}
  - required fields: data_table_id, data_table_column_id
  - risk: external mutation; approval required
- test_credential:
  - endpoint: POST /credentials/{{ record.credential_id }}/test
  - required fields: credential_id
  - risk: external mutation; approval required
- transfer_credential:
  - endpoint: PUT /credentials/{{ record.credential_id }}/transfer
  - required fields: credential_id
  - risk: external mutation; approval required
- pull_source_control:
  - endpoint: POST /source-control/pull
  - risk: external mutation; approval required
- generate_audit:
  - endpoint: POST /audit
  - risk: external mutation; approval required

## Security

- read risk: external n8n instance API read of workflow, execution, tag, user, variable, project, data table, and credential metadata
- write risk: external n8n instance API mutation of workflows, executions, tags, variables, projects, data tables, source-control pull, audit generation, and credential tests/transfers
- approval: required for all write actions
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect n8n
```

### Inspect as structured JSON

```bash
pm connectors inspect n8n --json
```

## Agent Rules

- Run pm connectors inspect n8n before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
