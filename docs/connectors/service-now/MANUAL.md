# pm connectors inspect service-now

```text
NAME
  pm connectors inspect service-now - ServiceNow connector manual

SYNOPSIS
  pm connectors inspect service-now
  pm connectors inspect service-now --json
  pm credentials add <name> --connector service-now [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads and writes ServiceNow incident, user, and group table data through the ServiceNow Table API.

ICON
  id: pm-sample
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics
  review_url: https://github.com/polymetrics-ai/cli

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  mode
  username
  password (secret)

ETL STREAMS
  incidents:
    primary key: sys_id
    cursor: updated_on
    fields: name(string), number(string), priority(string), short_description(string), state(string), sys_created_on(string), sys_id(string), updated_on(string)
  users:
    primary key: sys_id
    cursor: updated_on
    fields: active(string), email(string), name(string), number(string), sys_id(string), updated_on(string), user_name(string)
  groups:
    primary key: sys_id
    cursor: updated_on
    fields: active(string), description(string), name(string), number(string), sys_id(string), updated_on(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  create_incident:
    endpoint: POST /api/now/table/incident
    risk: creates a new incident record; low-risk external mutation (a new ticket), no approval required
  create_user:
    endpoint: POST /api/now/table/sys_user
    required fields: user_name
    risk: creates a new ServiceNow user account record; a new user account granted whatever role/ACL defaults the instance applies is a higher-scrutiny mutation than an incident/group create
  create_group:
    endpoint: POST /api/now/table/sys_user_group
    required fields: name
    risk: creates a new user group record; low-risk external mutation, no approval required
  update_incident:
    endpoint: PATCH /api/now/table/incident/{{ record.sys_id }}
    required fields: sys_id
    risk: mutates an existing incident's recorded fields (only fields present in the submitted record are changed; ServiceNow's Table API PATCH/PUT both modify only the submitted fields, never the whole record) by sys_id
  update_user:
    endpoint: PATCH /api/now/table/sys_user/{{ record.sys_id }}
    required fields: sys_id
    risk: mutates an existing user account's profile fields by sys_id, including active (deactivating a user's account revokes their instance access); higher-scrutiny than incident/group updates
  update_group:
    endpoint: PATCH /api/now/table/sys_user_group/{{ record.sys_id }}
    required fields: sys_id
    risk: mutates an existing group's recorded fields by sys_id, including active/manager; can change who is considered the group's membership owner

SECURITY
  read risk: external ServiceNow API read of incident, user, and group table data
  write risk: creates incident/user/group records and updates their fields by sys_id (ServiceNow Table API PATCH, which modifies only submitted fields); creating/deactivating a user account is a higher-scrutiny mutation than incident/group create-update
  approval: none for incident/group create-update (low-risk ticketing/CRM-style data); review user create/update before enabling in a caller with untrusted input, since it can grant or revoke ServiceNow instance access
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

COMMAND SURFACE
  Run ServiceNow's declared streams and reverse-ETL actions.
  Usage: pm service-now <command> [flags]
  Read streams
  Reverse ETL writes
  Other Commands
    api delete api now table incident sys-id - Documented DELETE /api/now/table/incident/{sys_id} (not implemented) [intent=direct_write availability=not_implemented operation=service-now.delete.api-now-table-incident-sys-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete api now table sys-user sys-id - Documented DELETE /api/now/table/sys_user/{sys_id} (not implemented) [intent=direct_write availability=not_implemented operation=service-now.delete.api-now-table-sys-user-sys-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete api now table sys-user-group sys-id - Documented DELETE /api/now/table/sys_user_group/{sys_id} (not implemented) [intent=direct_write availability=not_implemented operation=service-now.delete.api-now-table-sys-user-group-sys-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete now table tablename sys-id - Documented DELETE /now/table/{tableName}/{sys_id} (not implemented) [intent=direct_write availability=not_implemented operation=service-now.delete.now-table-tablename-sys-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api get api now attachment - Documented GET /api/now/attachment (not implemented) [intent=direct_read availability=not_implemented operation=service-now.get.api-now-attachment]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get api now stats tablename - Documented GET /api/now/stats/{tableName} (not implemented) [intent=direct_read availability=not_implemented operation=service-now.get.api-now-stats-tablename]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get api now table incident sys-id - Documented GET /api/now/table/incident/{sys_id} (not implemented) [intent=direct_read availability=not_implemented operation=service-now.get.api-now-table-incident-sys-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get api now table sys-user sys-id - Documented GET /api/now/table/sys_user/{sys_id} (not implemented) [intent=direct_read availability=not_implemented operation=service-now.get.api-now-table-sys-user-sys-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get api now table sys-user-grmember - Documented GET /api/now/table/sys_user_grmember (not implemented) [intent=direct_read availability=not_implemented operation=service-now.get.api-now-table-sys-user-grmember]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get api now table sys-user-group sys-id - Documented GET /api/now/table/sys_user_group/{sys_id} (not implemented) [intent=direct_read availability=not_implemented operation=service-now.get.api-now-table-sys-user-group-sys-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get now table tablename - Documented GET /now/table/{tableName} (not implemented) [intent=direct_read availability=not_implemented operation=service-now.get.now-table-tablename]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get now table tablename sys-id - Documented GET /now/table/{tableName}/{sys_id} (not implemented) [intent=direct_read availability=not_implemented operation=service-now.get.now-table-tablename-sys-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api patch now table tablename sys-id - Documented PATCH /now/table/{tableName}/{sys_id} (not implemented) [intent=direct_write availability=not_implemented operation=service-now.patch.now-table-tablename-sys-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post api now import importsettablename - Documented POST /api/now/import/{importSetTableName} (not implemented) [intent=direct_write availability=not_implemented operation=service-now.post.api-now-import-importsettablename]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post now table tablename - Documented POST /now/table/{tableName} (not implemented) [intent=direct_write availability=not_implemented operation=service-now.post.now-table-tablename]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put api now table incident sys-id - Documented PUT /api/now/table/incident/{sys_id} (not implemented) [intent=direct_write availability=not_implemented operation=service-now.put.api-now-table-incident-sys-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put api now table sys-user sys-id - Documented PUT /api/now/table/sys_user/{sys_id} (not implemented) [intent=direct_write availability=not_implemented operation=service-now.put.api-now-table-sys-user-sys-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put api now table sys-user-group sys-id - Documented PUT /api/now/table/sys_user_group/{sys_id} (not implemented) [intent=direct_write availability=not_implemented operation=service-now.put.api-now-table-sys-user-group-sys-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put now table tablename sys-id - Documented PUT /now/table/{tableName}/{sys_id} (not implemented) [intent=direct_write availability=not_implemented operation=service-now.put.now-table-tablename-sys-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    create group apply - Plan and execute the create group reverse-ETL action [intent=reverse_etl availability=implemented write=create_group]; approval: requires plan, preview, approval, and execute; risk: creates a new user group record; low-risk external mutation, no approval required; flags: --name (required)
    create incident apply - Plan and execute the create incident reverse-ETL action [intent=reverse_etl availability=implemented write=create_incident]; approval: requires plan, preview, approval, and execute; risk: creates a new incident record; low-risk external mutation (a new ticket), no approval required
    create user apply - Plan and execute the create user reverse-ETL action [intent=reverse_etl availability=implemented write=create_user]; approval: requires plan, preview, approval, and execute; risk: creates a new ServiceNow user account record; a new user account granted whatever role/ACL defaults the instance applies is a higher-scrutiny mutation than an incident/group create; flags: --user_name (required)
    groups list - Run the groups ETL stream [intent=etl availability=implemented stream=groups]; notes: discrepancy=present-in-surface-absent-from-artifact
    incidents list - Run the incidents ETL stream [intent=etl availability=implemented stream=incidents]; notes: discrepancy=present-in-surface-absent-from-artifact
    update group apply - Plan and execute the update group reverse-ETL action [intent=reverse_etl availability=implemented write=update_group]; approval: requires plan, preview, approval, and execute; risk: mutates an existing group's recorded fields by sys_id, including active/manager; can change who is considered the group's membership owner; flags: --sys_id (required)
    update incident apply - Plan and execute the update incident reverse-ETL action [intent=reverse_etl availability=implemented write=update_incident]; approval: requires plan, preview, approval, and execute; risk: mutates an existing incident's recorded fields (only fields present in the submitted record are changed; ServiceNow's Table API PATCH/PUT both modify only the submitted fields, never the whole record) by sys_id; flags: --sys_id (required)
    update user apply - Plan and execute the update user reverse-ETL action [intent=reverse_etl availability=implemented write=update_user]; approval: requires plan, preview, approval, and execute; risk: mutates an existing user account's profile fields by sys_id, including active (deactivating a user's account revokes their instance access); higher-scrutiny than incident/group updates; flags: --sys_id (required)
    users list - Run the users ETL stream [intent=etl availability=implemented stream=users]; notes: discrepancy=present-in-surface-absent-from-artifact

EXAMPLES
  # Inspect as a manual
  pm connectors inspect service-now

  # Inspect as structured JSON
  pm connectors inspect service-now --json

AGENT WORKFLOW
  - Run pm connectors inspect service-now before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
