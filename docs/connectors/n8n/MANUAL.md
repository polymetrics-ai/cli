# pm connectors inspect n8n

```text
NAME
  pm connectors inspect n8n - n8n connector manual

SYNOPSIS
  pm connectors inspect n8n
  pm connectors inspect n8n --json
  pm credentials add <name> --connector n8n [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads n8n workflows, executions, tags, users, variables, projects, data tables, and credential metadata; writes supported n8n public REST API mutations.

ICON
  id: n8n
  asset: icons/n8n.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://docs.n8n.io/api/

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url (required)
  credential_id
  credential_type_name
  data_table_column_id
  data_table_filter
  data_table_id
  data_table_search
  data_table_sort_by
  exclude_pinned_data
  execution_id
  execution_status
  ignore_data_size_limit
  include_execution_data
  include_role
  max_pages
  mode
  page_size
  project_id
  project_user_id
  redact_execution_data
  tag_id
  user_id
  variable_id
  variable_state
  workflow_active
  workflow_id
  workflow_name
  workflow_tags
  workflow_version_id
  api_key (secret) (required)

ETL STREAMS
  workflows:
    primary key: id
    cursor: updatedAt
    fields: active(boolean), createdAt(string), id(string), isArchived(boolean), name(string), triggerCount(integer), updatedAt(string), versionId(string)
  workflow:
    primary key: id
    cursor: updatedAt
    fields: active(boolean), connections(object), createdAt(string), description(string), id(string), isArchived(boolean), name(string), nodes(array), settings(object), shared(array), tags(array), triggerCount(integer), updatedAt(string), versionId(string)
  workflow_version:
    primary key: versionId
    cursor: updatedAt
    fields: authors(string), connections(object), createdAt(string), description(string), name(string), nodeGroups(array), nodes(array), updatedAt(string), versionId(string), workflowId(string)
  workflow_tags:
    primary key: id
    cursor: updatedAt
    fields: createdAt(string), id(string), name(string), updatedAt(string)
  executions:
    primary key: id
    cursor: startedAt
    fields: finished(boolean), id(string), mode(string), retryOf(string), startedAt(string), status(string), stoppedAt(string), workflowId(string)
  execution:
    primary key: id
    cursor: startedAt
    fields: customData(object), data(object), finished(boolean), id(string), mode(string), retryOf(string), retrySuccessId(string), startedAt(string), status(string), stoppedAt(string), waitTill(string), workflowId(string)
  execution_tags:
    primary key: id
    cursor: updatedAt
    fields: createdAt(string), id(string), name(string), updatedAt(string)
  tags:
    primary key: id
    cursor: updatedAt
    fields: createdAt(string), id(string), name(string), updatedAt(string)
  tag:
    primary key: id
    cursor: updatedAt
    fields: createdAt(string), id(string), name(string), updatedAt(string)
  users:
    primary key: id
    cursor: updatedAt
    fields: createdAt(string), email(string), firstName(string), id(string), isPending(boolean), lastName(string), role(string), updatedAt(string)
  user:
    primary key: id
    cursor: updatedAt
    fields: createdAt(string), email(string), firstName(string), id(string), isPending(boolean), lastName(string), mfaEnabled(boolean), role(string), updatedAt(string)
  variables:
    primary key: id
    fields: id(string), key(string), project(object), projectId(string), type(string), value(string)
  projects:
    primary key: id
    fields: id(string), name(string), type(string)
  project_members:
    primary key: id
    fields: createdAt(string), email(string), firstName(string), id(string), lastName(string), role(string), updatedAt(string)
  credentials:
    primary key: id
    cursor: updatedAt
    fields: createdAt(string), id(string), isGlobal(boolean), isManaged(boolean), isResolvable(boolean), name(string), resolvableAllowFallback(boolean), resolverId(string), type(string), updatedAt(string)
  credential:
    primary key: id
    cursor: updatedAt
    fields: createdAt(string), id(string), isGlobal(boolean), isManaged(boolean), isResolvable(boolean), name(string), resolvableAllowFallback(boolean), resolverId(string), type(string), updatedAt(string)
  credential_schema:
    fields: displayName(string), documentationUrl(string), name(string), properties(array)
  data_tables:
    primary key: id
    cursor: updatedAt
    fields: columns(array), createdAt(string), id(string), name(string), projectId(string), updatedAt(string)
  data_table:
    primary key: id
    cursor: updatedAt
    fields: columns(array), createdAt(string), id(string), name(string), projectId(string), updatedAt(string)
  data_table_rows:
    primary key: id
    cursor: updatedAt
    fields: createdAt(string), id(string), updatedAt(string)
  data_table_columns:
    primary key: id
    fields: dataTableId(string), id(string), index(integer), name(string), type(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  create_workflow:
    endpoint: POST /workflows
    required fields: name, nodes, connections, settings
    risk: external mutation; approval required
  update_workflow:
    endpoint: PUT /workflows/{{ record.workflow_id }}
    required fields: workflow_id, name, nodes, connections, settings
    risk: external mutation; approval required
  publish_workflow:
    endpoint: POST /workflows/{{ record.workflow_id }}/activate
    required fields: workflow_id
    risk: external mutation; approval required
  deactivate_workflow:
    endpoint: POST /workflows/{{ record.workflow_id }}/deactivate
    required fields: workflow_id
    risk: external mutation; approval required
  archive_workflow:
    endpoint: POST /workflows/{{ record.workflow_id }}/archive
    required fields: workflow_id
    risk: external mutation; approval required
  unarchive_workflow:
    endpoint: POST /workflows/{{ record.workflow_id }}/unarchive
    required fields: workflow_id
    risk: external mutation; approval required
  transfer_workflow:
    endpoint: PUT /workflows/{{ record.workflow_id }}/transfer
    required fields: workflow_id, destinationProjectId
    risk: external mutation; approval required
  retry_execution:
    endpoint: POST /executions/{{ record.execution_id }}/retry
    required fields: execution_id
    risk: external mutation; approval required
  stop_execution:
    endpoint: POST /executions/{{ record.execution_id }}/stop
    required fields: execution_id
    risk: external mutation; approval required
  stop_executions:
    endpoint: POST /executions/stop
    required fields: status
    risk: external mutation; approval required
  create_tag:
    endpoint: POST /tags
    required fields: name
    risk: external mutation; approval required
  update_tag:
    endpoint: PUT /tags/{{ record.tag_id }}
    required fields: tag_id, name
    risk: external mutation; approval required
  create_variable:
    endpoint: POST /variables
    required fields: key, value
    risk: external mutation; approval required
  update_variable:
    endpoint: PUT /variables/{{ record.variable_id }}
    required fields: variable_id, key, value
    risk: external mutation; approval required
  create_project:
    endpoint: POST /projects
    required fields: name
    risk: external mutation; approval required
  update_project:
    endpoint: PUT /projects/{{ record.project_id }}
    required fields: project_id, name
    risk: external mutation; approval required
  add_project_users:
    endpoint: POST /projects/{{ record.project_id }}/users
    required fields: project_id, relations
    risk: external mutation; approval required
  change_project_user_role:
    endpoint: PATCH /projects/{{ record.project_id }}/users/{{ record.project_user_id }}
    required fields: project_id, project_user_id, role
    risk: external mutation; approval required
  create_data_table:
    endpoint: POST /data-tables
    required fields: name, columns
    risk: external mutation; approval required
  update_data_table:
    endpoint: PATCH /data-tables/{{ record.data_table_id }}
    required fields: data_table_id, name
    risk: external mutation; approval required
  insert_data_table_rows:
    endpoint: POST /data-tables/{{ record.data_table_id }}/rows
    required fields: data_table_id, data
    risk: external mutation; approval required
  update_data_table_rows:
    endpoint: PATCH /data-tables/{{ record.data_table_id }}/rows/update
    required fields: data_table_id, filter, data
    risk: external mutation; approval required
  upsert_data_table_row:
    endpoint: POST /data-tables/{{ record.data_table_id }}/rows/upsert
    required fields: data_table_id, filter, data
    risk: external mutation; approval required
  add_data_table_column:
    endpoint: POST /data-tables/{{ record.data_table_id }}/columns
    required fields: data_table_id, name, type
    risk: external mutation; approval required
  update_data_table_column:
    endpoint: PATCH /data-tables/{{ record.data_table_id }}/columns/{{ record.data_table_column_id }}
    required fields: data_table_id, data_table_column_id, name
    risk: external mutation; approval required
  test_credential:
    endpoint: POST /credentials/{{ record.credential_id }}/test
    required fields: credential_id
    risk: external mutation; approval required
  transfer_credential:
    endpoint: PUT /credentials/{{ record.credential_id }}/transfer
    required fields: credential_id, destinationProjectId
    risk: external mutation; approval required
  pull_source_control:
    endpoint: POST /source-control/pull
    risk: external mutation; approval required
  generate_audit:
    endpoint: POST /audit
    risk: external mutation; approval required

SECURITY
  read risk: external n8n instance API read of workflow, execution, tag, user, variable, project, data table, and credential metadata
  write risk: external n8n instance API mutation of workflows, executions, tags, variables, projects, data tables, source-control pull, audit generation, and credential tests/transfers
  approval: required for all write actions
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

COMMAND SURFACE
  Run n8n's declared streams and reverse-ETL actions.
  Usage: pm n8n <command> [flags]
  Global flags:
    --approval-token-stdin (boolean): Read the approval token as one bounded line from standard input.
  Reverse ETL writes
  Other Commands
    add data table column apply - Plan and execute the add data table column reverse-ETL action. [intent=reverse_etl availability=implemented write=add_data_table_column]; approval: requires plan, preview, approval, and execute; risk: external mutation; approval required; flags: --data_table_id (required), --name (required), --type (required)
    add project users apply - Plan and execute the add project users reverse-ETL action. [intent=reverse_etl availability=implemented write=add_project_users]; approval: requires plan, preview, approval, and execute; risk: external mutation; approval required; flags: --project_id (required), --relations (required)
    change project user role apply - Plan and execute the change project user role reverse-ETL action. [intent=reverse_etl availability=implemented write=change_project_user_role]; approval: requires plan, preview, approval, and execute; risk: external mutation; approval required; flags: --project_id (required), --project_user_id (required), --role (required)
    create data table apply - Plan and execute the create data table reverse-ETL action. [intent=reverse_etl availability=implemented write=create_data_table]; approval: requires plan, preview, approval, and execute; risk: external mutation; approval required; flags: --columns (required), --name (required)
    create project apply - Plan and execute the create project reverse-ETL action. [intent=reverse_etl availability=implemented write=create_project]; approval: requires plan, preview, approval, and execute; risk: external mutation; approval required; flags: --name (required)
    create tag apply - Plan and execute the create tag reverse-ETL action. [intent=reverse_etl availability=implemented write=create_tag]; approval: requires plan, preview, approval, and execute; risk: external mutation; approval required; flags: --name (required)
    create variable apply - Plan and execute the create variable reverse-ETL action. [intent=reverse_etl availability=implemented write=create_variable]; approval: requires plan, preview, approval, and execute; risk: external mutation; approval required; flags: --key (required), --value (required)
    create workflow apply - Plan and execute the create workflow reverse-ETL action. [intent=reverse_etl availability=implemented write=create_workflow]; approval: requires plan, preview, approval, and execute; risk: external mutation; approval required; flags: --connections (required), --name (required), --nodes (required), --settings (required)
    deactivate workflow apply - Plan and execute the deactivate workflow reverse-ETL action. [intent=reverse_etl availability=implemented write=deactivate_workflow]; approval: requires plan, preview, approval, and execute; risk: external mutation; approval required; flags: --workflow_id (required)
    generate audit apply - Plan and execute the generate audit reverse-ETL action. [intent=reverse_etl availability=implemented write=generate_audit]; approval: requires plan, preview, approval, and execute; risk: external mutation; approval required
    insert data table rows apply - Plan and execute the insert data table rows reverse-ETL action. [intent=reverse_etl availability=implemented write=insert_data_table_rows]; approval: requires plan, preview, approval, and execute; risk: external mutation; approval required; flags: --data (required), --data_table_id (required)
    publish workflow apply - Plan and execute the publish workflow reverse-ETL action. [intent=reverse_etl availability=implemented write=publish_workflow]; approval: requires plan, preview, approval, and execute; risk: external mutation; approval required; flags: --workflow_id (required)
    pull source control apply - Plan and execute the pull source control reverse-ETL action. [intent=reverse_etl availability=implemented write=pull_source_control]; approval: requires plan, preview, approval, and execute; risk: external mutation; approval required
    retry execution apply - Plan and execute the retry execution reverse-ETL action. [intent=reverse_etl availability=implemented write=retry_execution]; approval: requires plan, preview, approval, and execute; risk: external mutation; approval required; flags: --execution_id (required)
    stop execution apply - Plan and execute the stop execution reverse-ETL action. [intent=reverse_etl availability=implemented write=stop_execution]; approval: requires plan, preview, approval, and execute; risk: external mutation; approval required; flags: --execution_id (required)
    stop executions apply - Plan and execute the stop executions reverse-ETL action. [intent=reverse_etl availability=implemented write=stop_executions]; approval: requires plan, preview, approval, and execute; risk: external mutation; approval required; flags: --status (required)
    test credential apply - Plan and execute the test credential reverse-ETL action. [intent=reverse_etl availability=implemented write=test_credential]; approval: requires plan, preview, approval, and execute; risk: external mutation; approval required; flags: --credential_id (required)
    transfer credential apply - Plan and execute the transfer credential reverse-ETL action. [intent=reverse_etl availability=implemented write=transfer_credential]; approval: requires plan, preview, approval, and execute; risk: external mutation; approval required; flags: --credential_id (required), --destinationProjectId (required)
    update data table apply - Plan and execute the update data table reverse-ETL action. [intent=reverse_etl availability=implemented write=update_data_table]; approval: requires plan, preview, approval, and execute; risk: external mutation; approval required; flags: --data_table_id (required), --name (required)
    update data table column apply - Plan and execute the update data table column reverse-ETL action. [intent=reverse_etl availability=implemented write=update_data_table_column]; approval: requires plan, preview, approval, and execute; risk: external mutation; approval required; flags: --data_table_column_id (required), --data_table_id (required), --name (required)
    update data table rows apply - Plan and execute the update data table rows reverse-ETL action. [intent=reverse_etl availability=implemented write=update_data_table_rows]; approval: requires plan, preview, approval, and execute; risk: external mutation; approval required; flags: --data (required), --data_table_id (required), --filter (required)
    update project apply - Plan and execute the update project reverse-ETL action. [intent=reverse_etl availability=implemented write=update_project]; approval: requires plan, preview, approval, and execute; risk: external mutation; approval required; flags: --name (required), --project_id (required)
    update tag apply - Plan and execute the update tag reverse-ETL action. [intent=reverse_etl availability=implemented write=update_tag]; approval: requires plan, preview, approval, and execute; risk: external mutation; approval required; flags: --name (required), --tag_id (required)
    update variable apply - Plan and execute the update variable reverse-ETL action. [intent=reverse_etl availability=implemented write=update_variable]; approval: requires plan, preview, approval, and execute; risk: external mutation; approval required; flags: --key (required), --value (required), --variable_id (required)
    update workflow apply - Plan and execute the update workflow reverse-ETL action. [intent=reverse_etl availability=implemented write=update_workflow]; approval: requires plan, preview, approval, and execute; risk: external mutation; approval required; flags: --connections (required), --name (required), --nodes (required), --settings (required), --workflow_id (required)
    upsert data table row apply - Plan and execute the upsert data table row reverse-ETL action. [intent=reverse_etl availability=implemented write=upsert_data_table_row]; approval: requires plan, preview, approval, and execute; risk: external mutation; approval required; flags: --data (required), --data_table_id (required), --filter (required)

EXAMPLES
  # Inspect as a manual
  pm connectors inspect n8n

  # Inspect as structured JSON
  pm connectors inspect n8n --json

AGENT WORKFLOW
  - Run pm connectors inspect n8n before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
