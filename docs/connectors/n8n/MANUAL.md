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
  base_url
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
  api_key (secret)

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
  Read streams
  Reverse ETL writes
  Other Commands
    add data table column apply - Plan and execute the add data table column reverse-ETL action [intent=reverse_etl availability=implemented write=add_data_table_column]; approval: requires plan, preview, approval, and execute; risk: external mutation; approval required; flags: --data_table_id (required), --name (required), --type (required)
    add project users apply - Plan and execute the add project users reverse-ETL action [intent=reverse_etl availability=not_implemented write=add_project_users]; approval: requires plan, preview, approval, and execute; risk: external mutation; approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    api delete api v1 credentials id - Documented DELETE /api/v1/credentials/{id} (not implemented) [intent=direct_write availability=not_implemented operation=n8n.delete.api-v1-credentials-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete api v1 data-tables datatableid - Documented DELETE /api/v1/data-tables/{dataTableId} (not implemented) [intent=direct_write availability=not_implemented operation=n8n.delete.api-v1-data-tables-datatableid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete api v1 data-tables datatableid columns columnid - Documented DELETE /api/v1/data-tables/{dataTableId}/columns/{columnId} (not implemented) [intent=direct_write availability=not_implemented operation=n8n.delete.api-v1-data-tables-datatableid-columns-columnid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete api v1 data-tables datatableid rows delete - Documented DELETE /api/v1/data-tables/{dataTableId}/rows/delete (not implemented) [intent=direct_write availability=not_implemented operation=n8n.delete.api-v1-data-tables-datatableid-rows-delete]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete api v1 executions id - Documented DELETE /api/v1/executions/{id} (not implemented) [intent=direct_write availability=not_implemented operation=n8n.delete.api-v1-executions-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete api v1 projects projectid - Documented DELETE /api/v1/projects/{projectId} (not implemented) [intent=direct_write availability=not_implemented operation=n8n.delete.api-v1-projects-projectid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete api v1 projects projectid users userid - Documented DELETE /api/v1/projects/{projectId}/users/{userId} (not implemented) [intent=direct_write availability=not_implemented operation=n8n.delete.api-v1-projects-projectid-users-userid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete api v1 tags id - Documented DELETE /api/v1/tags/{id} (not implemented) [intent=direct_write availability=not_implemented operation=n8n.delete.api-v1-tags-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete api v1 users id - Documented DELETE /api/v1/users/{id} (not implemented) [intent=direct_write availability=not_implemented operation=n8n.delete.api-v1-users-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete api v1 variables id - Documented DELETE /api/v1/variables/{id} (not implemented) [intent=direct_write availability=not_implemented operation=n8n.delete.api-v1-variables-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete api v1 workflows id - Documented DELETE /api/v1/workflows/{id} (not implemented) [intent=direct_write availability=not_implemented operation=n8n.delete.api-v1-workflows-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete community-packages name - Documented DELETE /community-packages/{name} (not implemented) [intent=direct_write availability=not_implemented operation=n8n.delete.community-packages-name]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete credentials id - Documented DELETE /credentials/{id} (not implemented) [intent=direct_write availability=not_implemented operation=n8n.delete.credentials-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete data-tables datatableid - Documented DELETE /data-tables/{dataTableId} (not implemented) [intent=direct_write availability=not_implemented operation=n8n.delete.data-tables-datatableid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete data-tables datatableid columns columnid - Documented DELETE /data-tables/{dataTableId}/columns/{columnId} (not implemented) [intent=direct_write availability=not_implemented operation=n8n.delete.data-tables-datatableid-columns-columnid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete data-tables datatableid rows clear - Documented DELETE /data-tables/{dataTableId}/rows/clear (not implemented) [intent=direct_write availability=not_implemented operation=n8n.delete.data-tables-datatableid-rows-clear]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete data-tables datatableid rows delete - Documented DELETE /data-tables/{dataTableId}/rows/delete (not implemented) [intent=direct_write availability=not_implemented operation=n8n.delete.data-tables-datatableid-rows-delete]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete executions id - Documented DELETE /executions/{id} (not implemented) [intent=direct_write availability=not_implemented operation=n8n.delete.executions-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete projects projectid - Documented DELETE /projects/{projectId} (not implemented) [intent=direct_write availability=not_implemented operation=n8n.delete.projects-projectid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete projects projectid folders folderid - Documented DELETE /projects/{projectId}/folders/{folderId} (not implemented) [intent=direct_write availability=not_implemented operation=n8n.delete.projects-projectid-folders-folderid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete projects projectid users userid - Documented DELETE /projects/{projectId}/users/{userId} (not implemented) [intent=direct_write availability=not_implemented operation=n8n.delete.projects-projectid-users-userid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete settings log-streaming destinations id - Documented DELETE /settings/log-streaming/destinations/{id} (not implemented) [intent=direct_write availability=not_implemented operation=n8n.delete.settings-log-streaming-destinations-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete tags id - Documented DELETE /tags/{id} (not implemented) [intent=direct_write availability=not_implemented operation=n8n.delete.tags-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete users id - Documented DELETE /users/{id} (not implemented) [intent=direct_write availability=not_implemented operation=n8n.delete.users-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete variables id - Documented DELETE /variables/{id} (not implemented) [intent=direct_write availability=not_implemented operation=n8n.delete.variables-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete workflows id - Documented DELETE /workflows/{id} (not implemented) [intent=direct_write availability=not_implemented operation=n8n.delete.workflows-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api get community-packages - Documented GET /community-packages (not implemented) [intent=direct_read availability=not_implemented operation=n8n.get.community-packages]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get credentials - Documented GET /credentials (not implemented) [intent=direct_read availability=not_implemented operation=n8n.get.credentials]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get credentials id - Documented GET /credentials/{id} (not implemented) [intent=direct_read availability=not_implemented operation=n8n.get.credentials-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get credentials schema credentialtypename - Documented GET /credentials/schema/{credentialTypeName} (not implemented) [intent=direct_read availability=not_implemented operation=n8n.get.credentials-schema-credentialtypename]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get data-tables - Documented GET /data-tables (not implemented) [intent=direct_read availability=not_implemented operation=n8n.get.data-tables]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get data-tables datatableid - Documented GET /data-tables/{dataTableId} (not implemented) [intent=direct_read availability=not_implemented operation=n8n.get.data-tables-datatableid]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get data-tables datatableid columns - Documented GET /data-tables/{dataTableId}/columns (not implemented) [intent=direct_read availability=not_implemented operation=n8n.get.data-tables-datatableid-columns]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get data-tables datatableid rows - Documented GET /data-tables/{dataTableId}/rows (not implemented) [intent=direct_read availability=not_implemented operation=n8n.get.data-tables-datatableid-rows]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get discover - Documented GET /discover (not implemented) [intent=direct_read availability=not_implemented operation=n8n.get.discover]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get executions - Documented GET /executions (not implemented) [intent=direct_read availability=not_implemented operation=n8n.get.executions]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get executions id - Documented GET /executions/{id} (not implemented) [intent=direct_read availability=not_implemented operation=n8n.get.executions-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get executions id tags - Documented GET /executions/{id}/tags (not implemented) [intent=direct_read availability=not_implemented operation=n8n.get.executions-id-tags]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get insights summary - Documented GET /insights/summary (not implemented) [intent=direct_read availability=not_implemented operation=n8n.get.insights-summary]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get projects - Documented GET /projects (not implemented) [intent=direct_read availability=not_implemented operation=n8n.get.projects]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get projects projectid folders - Documented GET /projects/{projectId}/folders (not implemented) [intent=direct_read availability=not_implemented operation=n8n.get.projects-projectid-folders]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get projects projectid folders folderid - Documented GET /projects/{projectId}/folders/{folderId} (not implemented) [intent=direct_read availability=not_implemented operation=n8n.get.projects-projectid-folders-folderid]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get projects projectid users - Documented GET /projects/{projectId}/users (not implemented) [intent=direct_read availability=not_implemented operation=n8n.get.projects-projectid-users]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get settings ldap - Documented GET /settings/ldap (not implemented) [intent=direct_read availability=not_implemented operation=n8n.get.settings-ldap]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get settings ldap sync - Documented GET /settings/ldap/sync (not implemented) [intent=direct_read availability=not_implemented operation=n8n.get.settings-ldap-sync]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get settings log-streaming destinations - Documented GET /settings/log-streaming/destinations (not implemented) [intent=direct_read availability=not_implemented operation=n8n.get.settings-log-streaming-destinations]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get settings log-streaming destinations id - Documented GET /settings/log-streaming/destinations/{id} (not implemented) [intent=direct_read availability=not_implemented operation=n8n.get.settings-log-streaming-destinations-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get settings log-streaming event-types - Documented GET /settings/log-streaming/event-types (not implemented) [intent=direct_read availability=not_implemented operation=n8n.get.settings-log-streaming-event-types]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get settings otel - Documented GET /settings/otel (not implemented) [intent=direct_read availability=not_implemented operation=n8n.get.settings-otel]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get settings security-policy - Documented GET /settings/security-policy (not implemented) [intent=direct_read availability=not_implemented operation=n8n.get.settings-security-policy]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get settings sso oidc - Documented GET /settings/sso/oidc (not implemented) [intent=direct_read availability=not_implemented operation=n8n.get.settings-sso-oidc]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get settings sso saml - Documented GET /settings/sso/saml (not implemented) [intent=direct_read availability=not_implemented operation=n8n.get.settings-sso-saml]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get tags - Documented GET /tags (not implemented) [intent=direct_read availability=not_implemented operation=n8n.get.tags]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get tags id - Documented GET /tags/{id} (not implemented) [intent=direct_read availability=not_implemented operation=n8n.get.tags-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get users - Documented GET /users (not implemented) [intent=direct_read availability=not_implemented operation=n8n.get.users]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get users id - Documented GET /users/{id} (not implemented) [intent=direct_read availability=not_implemented operation=n8n.get.users-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get variables - Documented GET /variables (not implemented) [intent=direct_read availability=not_implemented operation=n8n.get.variables]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get workflows - Documented GET /workflows (not implemented) [intent=direct_read availability=not_implemented operation=n8n.get.workflows]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get workflows id - Documented GET /workflows/{id} (not implemented) [intent=direct_read availability=not_implemented operation=n8n.get.workflows-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get workflows id history - Documented GET /workflows/{id}/history (not implemented) [intent=direct_read availability=not_implemented operation=n8n.get.workflows-id-history]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get workflows id tags - Documented GET /workflows/{id}/tags (not implemented) [intent=direct_read availability=not_implemented operation=n8n.get.workflows-id-tags]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get workflows id test-runs - Documented GET /workflows/{id}/test-runs (not implemented) [intent=direct_read availability=not_implemented operation=n8n.get.workflows-id-test-runs]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get workflows id test-runs runid - Documented GET /workflows/{id}/test-runs/{runId} (not implemented) [intent=direct_read availability=not_implemented operation=n8n.get.workflows-id-test-runs-runid]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get workflows id test-runs runid test-cases - Documented GET /workflows/{id}/test-runs/{runId}/test-cases (not implemented) [intent=direct_read availability=not_implemented operation=n8n.get.workflows-id-test-runs-runid-test-cases]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get workflows id versionid - Documented GET /workflows/{id}/{versionId} (not implemented) [intent=direct_read availability=not_implemented operation=n8n.get.workflows-id-versionid]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api patch api v1 credentials id - Documented PATCH /api/v1/credentials/{id} (not implemented) [intent=direct_write availability=not_implemented operation=n8n.patch.api-v1-credentials-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api patch api v1 users id role - Documented PATCH /api/v1/users/{id}/role (not implemented) [intent=direct_write availability=not_implemented operation=n8n.patch.api-v1-users-id-role]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api patch community-packages name - Documented PATCH /community-packages/{name} (not implemented) [intent=direct_write availability=not_implemented operation=n8n.patch.community-packages-name]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api patch credentials id - Documented PATCH /credentials/{id} (not implemented) [intent=direct_write availability=not_implemented operation=n8n.patch.credentials-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api patch data-tables datatableid - Documented PATCH /data-tables/{dataTableId} (not implemented) [intent=direct_write availability=not_implemented operation=n8n.patch.data-tables-datatableid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api patch data-tables datatableid columns columnid - Documented PATCH /data-tables/{dataTableId}/columns/{columnId} (not implemented) [intent=direct_write availability=not_implemented operation=n8n.patch.data-tables-datatableid-columns-columnid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api patch data-tables datatableid rows update - Documented PATCH /data-tables/{dataTableId}/rows/update (not implemented) [intent=direct_write availability=not_implemented operation=n8n.patch.data-tables-datatableid-rows-update]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api patch projects projectid folders folderid - Documented PATCH /projects/{projectId}/folders/{folderId} (not implemented) [intent=direct_write availability=not_implemented operation=n8n.patch.projects-projectid-folders-folderid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api patch projects projectid users userid - Documented PATCH /projects/{projectId}/users/{userId} (not implemented) [intent=direct_write availability=not_implemented operation=n8n.patch.projects-projectid-users-userid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api patch users id role - Documented PATCH /users/{id}/role (not implemented) [intent=direct_write availability=not_implemented operation=n8n.patch.users-id-role]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post api v1 credentials - Documented POST /api/v1/credentials (not implemented) [intent=direct_write availability=not_implemented operation=n8n.post.api-v1-credentials]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post api v1 users - Documented POST /api/v1/users (not implemented) [intent=direct_write availability=not_implemented operation=n8n.post.api-v1-users]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post audit - Documented POST /audit (not implemented) [intent=direct_write availability=not_implemented operation=n8n.post.audit]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post community-packages - Documented POST /community-packages (not implemented) [intent=direct_write availability=not_implemented operation=n8n.post.community-packages]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post credentials - Documented POST /credentials (not implemented) [intent=direct_write availability=not_implemented operation=n8n.post.credentials]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post credentials id test - Documented POST /credentials/{id}/test (not implemented) [intent=direct_write availability=not_implemented operation=n8n.post.credentials-id-test]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post data-tables - Documented POST /data-tables (not implemented) [intent=direct_write availability=not_implemented operation=n8n.post.data-tables]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post data-tables datatableid columns - Documented POST /data-tables/{dataTableId}/columns (not implemented) [intent=direct_write availability=not_implemented operation=n8n.post.data-tables-datatableid-columns]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post data-tables datatableid rows - Documented POST /data-tables/{dataTableId}/rows (not implemented) [intent=direct_write availability=not_implemented operation=n8n.post.data-tables-datatableid-rows]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post data-tables datatableid rows upsert - Documented POST /data-tables/{dataTableId}/rows/upsert (not implemented) [intent=direct_write availability=not_implemented operation=n8n.post.data-tables-datatableid-rows-upsert]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post executions id retry - Documented POST /executions/{id}/retry (not implemented) [intent=direct_write availability=not_implemented operation=n8n.post.executions-id-retry]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post executions id stop - Documented POST /executions/{id}/stop (not implemented) [intent=direct_write availability=not_implemented operation=n8n.post.executions-id-stop]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post executions stop - Documented POST /executions/stop (not implemented) [intent=direct_write availability=not_implemented operation=n8n.post.executions-stop]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post n8n-packages export - Documented POST /n8n-packages/export (not implemented) [intent=direct_write availability=not_implemented operation=n8n.post.n8n-packages-export]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post n8n-packages import - Documented POST /n8n-packages/import (not implemented) [intent=direct_write availability=not_implemented operation=n8n.post.n8n-packages-import]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post projects - Documented POST /projects (not implemented) [intent=direct_write availability=not_implemented operation=n8n.post.projects]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post projects projectid folders - Documented POST /projects/{projectId}/folders (not implemented) [intent=direct_write availability=not_implemented operation=n8n.post.projects-projectid-folders]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post projects projectid users - Documented POST /projects/{projectId}/users (not implemented) [intent=direct_write availability=not_implemented operation=n8n.post.projects-projectid-users]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post settings ldap sync - Documented POST /settings/ldap/sync (not implemented) [intent=direct_write availability=not_implemented operation=n8n.post.settings-ldap-sync]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post settings log-streaming destinations - Documented POST /settings/log-streaming/destinations (not implemented) [intent=direct_write availability=not_implemented operation=n8n.post.settings-log-streaming-destinations]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post settings log-streaming destinations id test - Documented POST /settings/log-streaming/destinations/{id}/test (not implemented) [intent=direct_write availability=not_implemented operation=n8n.post.settings-log-streaming-destinations-id-test]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post settings otel test-trace - Documented POST /settings/otel/test-trace (not implemented) [intent=direct_write availability=not_implemented operation=n8n.post.settings-otel-test-trace]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post source-control pull - Documented POST /source-control/pull (not implemented) [intent=direct_write availability=not_implemented operation=n8n.post.source-control-pull]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post tags - Documented POST /tags (not implemented) [intent=direct_write availability=not_implemented operation=n8n.post.tags]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post users - Documented POST /users (not implemented) [intent=direct_write availability=not_implemented operation=n8n.post.users]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post variables - Documented POST /variables (not implemented) [intent=direct_write availability=not_implemented operation=n8n.post.variables]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post workflows - Documented POST /workflows (not implemented) [intent=direct_write availability=not_implemented operation=n8n.post.workflows]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post workflows id activate - Documented POST /workflows/{id}/activate (not implemented) [intent=direct_write availability=not_implemented operation=n8n.post.workflows-id-activate]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post workflows id archive - Documented POST /workflows/{id}/archive (not implemented) [intent=direct_write availability=not_implemented operation=n8n.post.workflows-id-archive]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post workflows id deactivate - Documented POST /workflows/{id}/deactivate (not implemented) [intent=direct_write availability=not_implemented operation=n8n.post.workflows-id-deactivate]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post workflows id publish - Documented POST /workflows/{id}/publish (not implemented) [intent=direct_write availability=not_implemented operation=n8n.post.workflows-id-publish]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post workflows id test-runs - Documented POST /workflows/{id}/test-runs (not implemented) [intent=direct_write availability=not_implemented operation=n8n.post.workflows-id-test-runs]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post workflows id test-runs runid cancel - Documented POST /workflows/{id}/test-runs/{runId}/cancel (not implemented) [intent=direct_write availability=not_implemented operation=n8n.post.workflows-id-test-runs-runid-cancel]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post workflows id unarchive - Documented POST /workflows/{id}/unarchive (not implemented) [intent=direct_write availability=not_implemented operation=n8n.post.workflows-id-unarchive]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post workflows id unpublish - Documented POST /workflows/{id}/unpublish (not implemented) [intent=direct_write availability=not_implemented operation=n8n.post.workflows-id-unpublish]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put api v1 executions id tags - Documented PUT /api/v1/executions/{id}/tags (not implemented) [intent=direct_write availability=not_implemented operation=n8n.put.api-v1-executions-id-tags]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put api v1 workflows id tags - Documented PUT /api/v1/workflows/{id}/tags (not implemented) [intent=direct_write availability=not_implemented operation=n8n.put.api-v1-workflows-id-tags]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put credentials id transfer - Documented PUT /credentials/{id}/transfer (not implemented) [intent=direct_write availability=not_implemented operation=n8n.put.credentials-id-transfer]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put executions id tags - Documented PUT /executions/{id}/tags (not implemented) [intent=direct_write availability=not_implemented operation=n8n.put.executions-id-tags]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put projects projectid - Documented PUT /projects/{projectId} (not implemented) [intent=direct_write availability=not_implemented operation=n8n.put.projects-projectid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put settings ldap - Documented PUT /settings/ldap (not implemented) [intent=direct_write availability=not_implemented operation=n8n.put.settings-ldap]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put settings log-streaming destinations id - Documented PUT /settings/log-streaming/destinations/{id} (not implemented) [intent=direct_write availability=not_implemented operation=n8n.put.settings-log-streaming-destinations-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put settings otel - Documented PUT /settings/otel (not implemented) [intent=direct_write availability=not_implemented operation=n8n.put.settings-otel]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put settings security-policy - Documented PUT /settings/security-policy (not implemented) [intent=direct_write availability=not_implemented operation=n8n.put.settings-security-policy]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put settings sso oidc - Documented PUT /settings/sso/oidc (not implemented) [intent=direct_write availability=not_implemented operation=n8n.put.settings-sso-oidc]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put settings sso saml - Documented PUT /settings/sso/saml (not implemented) [intent=direct_write availability=not_implemented operation=n8n.put.settings-sso-saml]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put tags id - Documented PUT /tags/{id} (not implemented) [intent=direct_write availability=not_implemented operation=n8n.put.tags-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put variables id - Documented PUT /variables/{id} (not implemented) [intent=direct_write availability=not_implemented operation=n8n.put.variables-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put workflows id - Documented PUT /workflows/{id} (not implemented) [intent=direct_write availability=not_implemented operation=n8n.put.workflows-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put workflows id tags - Documented PUT /workflows/{id}/tags (not implemented) [intent=direct_write availability=not_implemented operation=n8n.put.workflows-id-tags]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put workflows id transfer - Documented PUT /workflows/{id}/transfer (not implemented) [intent=direct_write availability=not_implemented operation=n8n.put.workflows-id-transfer]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    archive workflow apply - Plan and execute the archive workflow reverse-ETL action [intent=reverse_etl availability=implemented write=archive_workflow]; approval: requires plan, preview, approval, and execute; risk: external mutation; approval required; flags: --workflow_id (required)
    change project user role apply - Plan and execute the change project user role reverse-ETL action [intent=reverse_etl availability=implemented write=change_project_user_role]; approval: requires plan, preview, approval, and execute; risk: external mutation; approval required; flags: --project_id (required), --project_user_id (required), --role (required)
    create data table apply - Plan and execute the create data table reverse-ETL action [intent=reverse_etl availability=not_implemented write=create_data_table]; approval: requires plan, preview, approval, and execute; risk: external mutation; approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    create project apply - Plan and execute the create project reverse-ETL action [intent=reverse_etl availability=implemented write=create_project]; approval: requires plan, preview, approval, and execute; risk: external mutation; approval required; flags: --name (required)
    create tag apply - Plan and execute the create tag reverse-ETL action [intent=reverse_etl availability=implemented write=create_tag]; approval: requires plan, preview, approval, and execute; risk: external mutation; approval required; flags: --name (required)
    create variable apply - Plan and execute the create variable reverse-ETL action [intent=reverse_etl availability=implemented write=create_variable]; approval: requires plan, preview, approval, and execute; risk: external mutation; approval required; flags: --key (required), --value (required)
    create workflow apply - Plan and execute the create workflow reverse-ETL action [intent=reverse_etl availability=not_implemented write=create_workflow]; approval: requires plan, preview, approval, and execute; risk: external mutation; approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    credential list - Run the credential ETL stream [intent=etl availability=implemented stream=credential]; notes: discrepancy=present-in-surface-absent-from-artifact
    credential schema list - Run the credential schema ETL stream [intent=etl availability=implemented stream=credential_schema]; notes: discrepancy=present-in-surface-absent-from-artifact
    credentials list - Run the credentials ETL stream [intent=etl availability=implemented stream=credentials]; notes: discrepancy=present-in-surface-absent-from-artifact
    data table columns list - Run the data table columns ETL stream [intent=etl availability=implemented stream=data_table_columns]; notes: discrepancy=present-in-surface-absent-from-artifact
    data table list - Run the data table ETL stream [intent=etl availability=implemented stream=data_table]; notes: discrepancy=present-in-surface-absent-from-artifact
    data table rows list - Run the data table rows ETL stream [intent=etl availability=implemented stream=data_table_rows]; notes: discrepancy=present-in-surface-absent-from-artifact
    data tables list - Run the data tables ETL stream [intent=etl availability=implemented stream=data_tables]; notes: discrepancy=present-in-surface-absent-from-artifact
    deactivate workflow apply - Plan and execute the deactivate workflow reverse-ETL action [intent=reverse_etl availability=implemented write=deactivate_workflow]; approval: requires plan, preview, approval, and execute; risk: external mutation; approval required; flags: --workflow_id (required)
    execution list - Run the execution ETL stream [intent=etl availability=implemented stream=execution]; notes: discrepancy=present-in-surface-absent-from-artifact
    execution tags list - Run the execution tags ETL stream [intent=etl availability=implemented stream=execution_tags]; notes: discrepancy=present-in-surface-absent-from-artifact
    executions list - Run the executions ETL stream [intent=etl availability=implemented stream=executions]; notes: discrepancy=present-in-surface-absent-from-artifact
    generate audit apply - Plan and execute the generate audit reverse-ETL action [intent=reverse_etl availability=implemented write=generate_audit]; approval: requires plan, preview, approval, and execute; risk: external mutation; approval required
    insert data table rows apply - Plan and execute the insert data table rows reverse-ETL action [intent=reverse_etl availability=not_implemented write=insert_data_table_rows]; approval: requires plan, preview, approval, and execute; risk: external mutation; approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    project members list - Run the project members ETL stream [intent=etl availability=implemented stream=project_members]; notes: discrepancy=present-in-surface-absent-from-artifact
    projects list - Run the projects ETL stream [intent=etl availability=implemented stream=projects]; notes: discrepancy=present-in-surface-absent-from-artifact
    publish workflow apply - Plan and execute the publish workflow reverse-ETL action [intent=reverse_etl availability=implemented write=publish_workflow]; approval: requires plan, preview, approval, and execute; risk: external mutation; approval required; flags: --workflow_id (required)
    pull source control apply - Plan and execute the pull source control reverse-ETL action [intent=reverse_etl availability=implemented write=pull_source_control]; approval: requires plan, preview, approval, and execute; risk: external mutation; approval required
    retry execution apply - Plan and execute the retry execution reverse-ETL action [intent=reverse_etl availability=implemented write=retry_execution]; approval: requires plan, preview, approval, and execute; risk: external mutation; approval required; flags: --execution_id (required)
    stop execution apply - Plan and execute the stop execution reverse-ETL action [intent=reverse_etl availability=implemented write=stop_execution]; approval: requires plan, preview, approval, and execute; risk: external mutation; approval required; flags: --execution_id (required)
    stop executions apply - Plan and execute the stop executions reverse-ETL action [intent=reverse_etl availability=not_implemented write=stop_executions]; approval: requires plan, preview, approval, and execute; risk: external mutation; approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    tag list - Run the tag ETL stream [intent=etl availability=implemented stream=tag]; notes: discrepancy=present-in-surface-absent-from-artifact
    tags list - Run the tags ETL stream [intent=etl availability=implemented stream=tags]; notes: discrepancy=present-in-surface-absent-from-artifact
    test credential apply - Plan and execute the test credential reverse-ETL action [intent=reverse_etl availability=implemented write=test_credential]; approval: requires plan, preview, approval, and execute; risk: external mutation; approval required; flags: --credential_id (required)
    transfer credential apply - Plan and execute the transfer credential reverse-ETL action [intent=reverse_etl availability=implemented write=transfer_credential]; approval: requires plan, preview, approval, and execute; risk: external mutation; approval required; flags: --credential_id (required), --destinationProjectId (required)
    transfer workflow apply - Plan and execute the transfer workflow reverse-ETL action [intent=reverse_etl availability=implemented write=transfer_workflow]; approval: requires plan, preview, approval, and execute; risk: external mutation; approval required; flags: --destinationProjectId (required), --workflow_id (required)
    unarchive workflow apply - Plan and execute the unarchive workflow reverse-ETL action [intent=reverse_etl availability=implemented write=unarchive_workflow]; approval: requires plan, preview, approval, and execute; risk: external mutation; approval required; flags: --workflow_id (required)
    update data table apply - Plan and execute the update data table reverse-ETL action [intent=reverse_etl availability=implemented write=update_data_table]; approval: requires plan, preview, approval, and execute; risk: external mutation; approval required; flags: --data_table_id (required), --name (required)
    update data table column apply - Plan and execute the update data table column reverse-ETL action [intent=reverse_etl availability=implemented write=update_data_table_column]; approval: requires plan, preview, approval, and execute; risk: external mutation; approval required; flags: --data_table_column_id (required), --data_table_id (required), --name (required)
    update data table rows apply - Plan and execute the update data table rows reverse-ETL action [intent=reverse_etl availability=not_implemented write=update_data_table_rows]; approval: requires plan, preview, approval, and execute; risk: external mutation; approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    update project apply - Plan and execute the update project reverse-ETL action [intent=reverse_etl availability=implemented write=update_project]; approval: requires plan, preview, approval, and execute; risk: external mutation; approval required; flags: --name (required), --project_id (required)
    update tag apply - Plan and execute the update tag reverse-ETL action [intent=reverse_etl availability=implemented write=update_tag]; approval: requires plan, preview, approval, and execute; risk: external mutation; approval required; flags: --name (required), --tag_id (required)
    update variable apply - Plan and execute the update variable reverse-ETL action [intent=reverse_etl availability=implemented write=update_variable]; approval: requires plan, preview, approval, and execute; risk: external mutation; approval required; flags: --key (required), --value (required), --variable_id (required)
    update workflow apply - Plan and execute the update workflow reverse-ETL action [intent=reverse_etl availability=not_implemented write=update_workflow]; approval: requires plan, preview, approval, and execute; risk: external mutation; approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    upsert data table row apply - Plan and execute the upsert data table row reverse-ETL action [intent=reverse_etl availability=not_implemented write=upsert_data_table_row]; approval: requires plan, preview, approval, and execute; risk: external mutation; approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    user list - Run the user ETL stream [intent=etl availability=implemented stream=user]; notes: discrepancy=present-in-surface-absent-from-artifact
    users list - Run the users ETL stream [intent=etl availability=implemented stream=users]; notes: discrepancy=present-in-surface-absent-from-artifact
    variables list - Run the variables ETL stream [intent=etl availability=implemented stream=variables]; notes: discrepancy=present-in-surface-absent-from-artifact
    workflow list - Run the workflow ETL stream [intent=etl availability=implemented stream=workflow]; notes: discrepancy=present-in-surface-absent-from-artifact
    workflow tags list - Run the workflow tags ETL stream [intent=etl availability=implemented stream=workflow_tags]; notes: discrepancy=present-in-surface-absent-from-artifact
    workflow version list - Run the workflow version ETL stream [intent=etl availability=implemented stream=workflow_version]; notes: discrepancy=present-in-surface-absent-from-artifact
    workflows list - Run the workflows ETL stream [intent=etl availability=implemented stream=workflows]; notes: discrepancy=present-in-surface-absent-from-artifact

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
