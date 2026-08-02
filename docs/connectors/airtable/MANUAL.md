
```text
NAME
  pm connectors inspect airtable - Airtable connector manual

SYNOPSIS
  pm connectors inspect airtable
  pm connectors inspect airtable --json
  pm credentials add <name> --connector airtable [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Airtable Web API metadata, records, comments, webhooks, SCIM details, enterprise/admin data, and HyperDB direct reads; executes typed single-resource mutations while batch array-cardinality operations stay blocked until enforceable.

ICON
  asset: icons/airtable.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://airtable.com/developers/web/api/changelog

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_id
  base_url
  enterprise_account_id
  enterprise_audit_log_task_id
  enterprise_task_id
  group_id
  page_bundle_id
  page_size
  record_id
  table_id
  user_id
  view_id
  webhook_id
  workspace_id
  access_token (secret)
  api_key (secret)

ETL STREAMS
  scim_group:
    primary key: id
    fields: displayName(), id(), members(), schemas()
  scim_user:
    primary key: id
    fields: active(), emails(), id(), name(), schemas(), userName()
  webhooks:
    primary key: id
    fields: enabled(), id(), notificationUrl()
  webhook_payloads:
    primary key: id
    fields: baseTransactionNumber(), id(), timestamp()
  bases:
    primary key: id
    fields: id(), name(), permissionLevel()
  base_collaborators:
    primary key: id
    fields: collaborators(), createdTime(), id(), name(), permissionLevel()
  block_installations:
    primary key: id
    fields: blockId(), createdByUserId(), createdTime(), id(), state()
  interface:
    primary key: id
    fields: createdTime(), id(), isPublished(), name(), rootTableId()
  shares:
    primary key: id
    fields: createdByUserId(), createdTime(), id(), state(), type(), url()
  tables:
    primary key: id
    fields: fields(), id(), name()
  views:
    primary key: id
    fields: id(), name(), type(), visibleFieldIds()
  view_metadata:
    primary key: id
    fields: id(), name(), type(), visibleFieldIds()
  enterprise:
    primary key: id
    fields: createdTime(), id(), name(), rootEnterpriseAccountId()
  audit_log_events:
    primary key: id
    fields: createdTime(), eventType(), id()
  audit_log_requests:
    primary key: id
    fields: completedTime(), createdByUserId(), createdTime(), id(), state(), url()
  audit_log_request:
    primary key: id
    fields: completedTime(), createdByUserId(), createdTime(), id(), state(), url()
  change_events:
    primary key: id
    fields: createdTime(), eventType(), id()
  ediscovery_exports:
    primary key: id
    fields: baseId(), completedTime(), createdByUserId(), createdTime(), downloadUrl(), id(), state()
  ediscovery_export:
    primary key: id
    fields: baseId(), completedTime(), createdByUserId(), createdTime(), downloadUrl(), id(), state()
  enterprise_packages:
    primary key: id
    fields: createdByUserId(), createdTime(), id(), name(), version()
  enterprise_personal_access_tokens:
    primary key: id
    fields: createdTime(), id(), lastUsedTime(), scopes(), userId()
  enterprise_users:
    primary key: id
    fields: createdTime(), email(), id(), isAdmin(), lastActivityTime(), name(), state()
  enterprise_user:
    primary key: id
    fields: createdTime(), email(), id(), isAdmin(), lastActivityTime(), name(), state()
  user_group:
    primary key: id
    fields: createdTime(), id(), memberUserIds(), name()
  whoami:
    primary key: id
    fields: email(), id(), scopes()
  workspace_collaborators:
    primary key: id
    fields: collaborators(), createdTime(), id(), name()
  records:
    primary key: id
    fields: createdTime(), fields(), id()
  record:
    primary key: id
    fields: createdTime(), fields(), id()
  comments:
    primary key: id
    fields: id(), record_id(), text()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  delete_scim_group:
    endpoint: DELETE /scim/v2/Groups/{{ record.id }}
    required fields: id
    risk: destructive Airtable mutation; delete/revoke/logout/remove semantics are idempotent where Airtable returns 404 and require explicit approval before execute
  delete_scim_user:
    endpoint: DELETE /scim/v2/Users/{{ record.id }}
    required fields: id
    risk: destructive Airtable mutation; delete/revoke/logout/remove semantics are idempotent where Airtable returns 404 and require explicit approval before execute
  create_webhook:
    endpoint: POST /v0/bases/{{ config.base_id }}/webhooks
    required fields: specification
    risk: Airtable webhook mutation; may start, stop, or change outbound notifications and must be previewed and approved
  delete_webhook:
    endpoint: DELETE /v0/bases/{{ config.base_id }}/webhooks/{{ record.id }}
    required fields: id
    risk: destructive Airtable mutation; delete/revoke/logout/remove semantics are idempotent where Airtable returns 404 and require explicit approval before execute
  set_webhook_notifications:
    endpoint: POST /v0/bases/{{ config.base_id }}/webhooks/{{ record.id }}/enableNotifications
    required fields: id, enable
    risk: Airtable webhook mutation; may start, stop, or change outbound notifications and must be previewed and approved
  refresh_webhook:
    endpoint: POST /v0/bases/{{ config.base_id }}/webhooks/{{ record.id }}/refresh
    required fields: id
    risk: Airtable webhook mutation; may start, stop, or change outbound notifications and must be previewed and approved
  delete_base:
    endpoint: DELETE /v0/meta/bases/{{ record.base_id }}
    required fields: base_id
    risk: destructive Airtable mutation; delete/revoke/logout/remove semantics are idempotent where Airtable returns 404 and require explicit approval before execute
  delete_block_installation:
    endpoint: DELETE /v0/meta/bases/{{ config.base_id }}/blockInstallations/{{ record.id }}
    required fields: id
    risk: destructive Airtable mutation; delete/revoke/logout/remove semantics are idempotent where Airtable returns 404 and require explicit approval before execute
  manage_block_installation:
    endpoint: PATCH /v0/meta/bases/{{ config.base_id }}/blockInstallations/{{ record.id }}
    required fields: id, state
    risk: typed Airtable API mutation; preview and explicit approval required before execute
  delete_base_collaborator:
    endpoint: DELETE /v0/meta/bases/{{ config.base_id }}/collaborators/{{ record.user_or_group_id }}
    required fields: user_or_group_id
    risk: destructive Airtable mutation; delete/revoke/logout/remove semantics are idempotent where Airtable returns 404 and require explicit approval before execute
  update_collaborator_base_permission:
    endpoint: PATCH /v0/meta/bases/{{ config.base_id }}/collaborators/{{ record.user_or_group_id }}
    required fields: user_or_group_id, permissionLevel
    risk: typed Airtable API mutation; preview and explicit approval required before execute
  delete_interface_collaborator:
    endpoint: DELETE /v0/meta/bases/{{ config.base_id }}/interfaces/{{ record.page_bundle_id }}/collaborators/{{ record.user_or_group_id }}
    required fields: page_bundle_id, user_or_group_id
    risk: destructive Airtable mutation; delete/revoke/logout/remove semantics are idempotent where Airtable returns 404 and require explicit approval before execute
  update_interface_collaborator:
    endpoint: PATCH /v0/meta/bases/{{ config.base_id }}/interfaces/{{ record.page_bundle_id }}/collaborators/{{ record.user_or_group_id }}
    required fields: page_bundle_id, user_or_group_id, permissionLevel
    risk: typed Airtable API mutation; preview and explicit approval required before execute
  delete_interface_invite:
    endpoint: DELETE /v0/meta/bases/{{ config.base_id }}/interfaces/{{ record.page_bundle_id }}/invites/{{ record.invite_id }}
    required fields: page_bundle_id, invite_id
    risk: destructive Airtable mutation; delete/revoke/logout/remove semantics are idempotent where Airtable returns 404 and require explicit approval before execute
  delete_base_invite:
    endpoint: DELETE /v0/meta/bases/{{ config.base_id }}/invites/{{ record.invite_id }}
    required fields: invite_id
    risk: destructive Airtable mutation; delete/revoke/logout/remove semantics are idempotent where Airtable returns 404 and require explicit approval before execute
  delete_share:
    endpoint: DELETE /v0/meta/bases/{{ config.base_id }}/shares/{{ record.id }}
    required fields: id
    risk: destructive Airtable mutation; delete/revoke/logout/remove semantics are idempotent where Airtable returns 404 and require explicit approval before execute
  manage_share:
    endpoint: PATCH /v0/meta/bases/{{ config.base_id }}/shares/{{ record.id }}
    required fields: id, state
    risk: typed Airtable API mutation; preview and explicit approval required before execute
  update_table:
    endpoint: PATCH /v0/meta/bases/{{ config.base_id }}/tables/{{ config.table_id }}
    risk: Airtable schema mutation visible to collaborators; preview and approval required
  create_field:
    endpoint: POST /v0/meta/bases/{{ config.base_id }}/tables/{{ record.table_id }}/fields
    required fields: table_id, name, type
    risk: Airtable schema mutation visible to collaborators; preview and approval required
  update_field:
    endpoint: PATCH /v0/meta/bases/{{ config.base_id }}/tables/{{ record.table_id }}/fields/{{ record.column_id }}
    required fields: table_id, column_id
    risk: Airtable schema mutation visible to collaborators; preview and approval required
  delete_view:
    endpoint: DELETE /v0/meta/bases/{{ config.base_id }}/views/{{ record.id }}
    required fields: id
    risk: destructive Airtable mutation; delete/revoke/logout/remove semantics are idempotent where Airtable returns 404 and require explicit approval before execute
  create_audit_log_request:
    endpoint: POST /v0/meta/enterpriseAccounts/{{ record.enterprise_account_id }}/auditLogs
    required fields: enterprise_account_id, timePeriod
    risk: elevated Airtable admin/SCIM mutation; affects enterprise identity, collaborators, packages, workspaces, or access controls and must be previewed and approved
  create_descendant_enterprise:
    endpoint: POST /v0/meta/enterpriseAccounts/{{ record.enterprise_account_id }}/descendants
    required fields: enterprise_account_id, name
    risk: elevated Airtable admin/SCIM mutation; affects enterprise identity, collaborators, packages, workspaces, or access controls and must be previewed and approved
  create_ediscovery_export:
    endpoint: POST /v0/meta/enterpriseAccounts/{{ record.enterprise_account_id }}/exports
    required fields: enterprise_account_id, baseId
    risk: elevated Airtable admin/SCIM mutation; affects enterprise identity, collaborators, packages, workspaces, or access controls and must be previewed and approved
  create_base_from_package_enterprise:
    endpoint: POST /v0/meta/enterpriseAccounts/{{ record.enterprise_account_id }}/packages/{{ record.package_id }}/install
    required fields: enterprise_account_id, package_id, workspaceId, packageReleaseId, name
    risk: elevated Airtable admin/SCIM mutation; affects enterprise identity, collaborators, packages, workspaces, or access controls and must be previewed and approved
  delete_users_by_email:
    endpoint: DELETE /v0/meta/enterpriseAccounts/{{ record.enterprise_account_id }}/users?email={{ record.email | urlencode }}
    required fields: enterprise_account_id, email
    risk: destructive Airtable mutation; delete/revoke/logout/remove semantics are idempotent where Airtable returns 404 and require explicit approval before execute
  delete_user_by_id:
    endpoint: DELETE /v0/meta/enterpriseAccounts/{{ record.enterprise_account_id }}/users/{{ record.id }}
    required fields: enterprise_account_id, id
    risk: destructive Airtable mutation; delete/revoke/logout/remove semantics are idempotent where Airtable returns 404 and require explicit approval before execute
  manage_user:
    endpoint: PATCH /v0/meta/enterpriseAccounts/{{ record.enterprise_account_id }}/users/{{ record.id }}
    required fields: enterprise_account_id, id
    risk: elevated Airtable admin/SCIM mutation; affects enterprise identity, collaborators, packages, workspaces, or access controls and must be previewed and approved
  logout_user:
    endpoint: POST /v0/meta/enterpriseAccounts/{{ record.enterprise_account_id }}/users/{{ record.id }}/logout
    required fields: enterprise_account_id, id
    risk: destructive Airtable mutation; delete/revoke/logout/remove semantics are idempotent where Airtable returns 404 and require explicit approval before execute
  remove_user_from_enterprise:
    endpoint: POST /v0/meta/enterpriseAccounts/{{ record.enterprise_account_id }}/users/{{ record.id }}/remove
    required fields: enterprise_account_id, id
    risk: destructive Airtable mutation; delete/revoke/logout/remove semantics are idempotent where Airtable returns 404 and require explicit approval before execute
  create_workspace:
    endpoint: POST /v0/meta/workspaces
    required fields: enterpriseAccountId, name
    risk: typed Airtable API mutation; preview and explicit approval required before execute
  delete_workspace:
    endpoint: DELETE /v0/meta/workspaces/{{ record.workspace_id }}
    required fields: workspace_id
    risk: destructive Airtable mutation; delete/revoke/logout/remove semantics are idempotent where Airtable returns 404 and require explicit approval before execute
  delete_workspace_collaborator:
    endpoint: DELETE /v0/meta/workspaces/{{ record.workspace_id }}/collaborators/{{ record.user_or_group_id }}
    required fields: workspace_id, user_or_group_id
    risk: destructive Airtable mutation; delete/revoke/logout/remove semantics are idempotent where Airtable returns 404 and require explicit approval before execute
  update_workspace_collaborator:
    endpoint: PATCH /v0/meta/workspaces/{{ record.workspace_id }}/collaborators/{{ record.user_or_group_id }}
    required fields: workspace_id, user_or_group_id, permissionLevel
    risk: typed Airtable API mutation; preview and explicit approval required before execute
  delete_workspace_invite:
    endpoint: DELETE /v0/meta/workspaces/{{ record.workspace_id }}/invites/{{ record.invite_id }}
    required fields: workspace_id, invite_id
    risk: destructive Airtable mutation; delete/revoke/logout/remove semantics are idempotent where Airtable returns 404 and require explicit approval before execute
  move_base:
    endpoint: POST /v0/meta/workspaces/{{ record.workspace_id }}/moveBase
    required fields: workspace_id, baseId, targetWorkspaceId
    risk: typed Airtable API mutation; preview and explicit approval required before execute
  update_workspace_restrictions:
    endpoint: POST /v0/meta/workspaces/{{ record.workspace_id }}/updateRestrictions
    required fields: workspace_id
    risk: typed Airtable API mutation; preview and explicit approval required before execute
  upload_attachment:
    endpoint: POST /v0/{{ config.base_id }}/{{ record.record_id }}/{{ record.attachment_field_id_or_name }}/uploadAttachment
    required fields: record_id, attachment_field_id_or_name, contentType, filename, file
    risk: Airtable schema mutation visible to collaborators; preview and approval required
  delete_record:
    endpoint: DELETE /v0/{{ config.base_id }}/{{ config.table_id }}/{{ record.id }}
    required fields: id
    risk: destructive Airtable mutation; delete/revoke/logout/remove semantics are idempotent where Airtable returns 404 and require explicit approval before execute
  update_record:
    endpoint: PATCH /v0/{{ config.base_id }}/{{ config.table_id }}/{{ record.id }}
    required fields: id, fields
    risk: Airtable schema mutation visible to collaborators; preview and approval required
  replace_record:
    endpoint: PUT /v0/{{ config.base_id }}/{{ config.table_id }}/{{ record.id }}
    required fields: id, fields
    risk: replacement/upsert Airtable mutation; may overwrite existing fields or records and requires explicit approval before execute
  create_comment:
    endpoint: POST /v0/{{ config.base_id }}/{{ config.table_id }}/{{ record.record_id }}/comments
    required fields: record_id, text
    risk: Airtable schema mutation visible to collaborators; preview and approval required
  delete_comment:
    endpoint: DELETE /v0/{{ config.base_id }}/{{ config.table_id }}/{{ record.record_id }}/comments/{{ record.row_comment_id }}
    required fields: record_id, row_comment_id
    risk: destructive Airtable mutation; delete/revoke/logout/remove semantics are idempotent where Airtable returns 404 and require explicit approval before execute
  update_comment:
    endpoint: PATCH /v0/{{ config.base_id }}/{{ config.table_id }}/{{ record.record_id }}/comments/{{ record.row_comment_id }}
    required fields: record_id, row_comment_id, text
    risk: Airtable schema mutation visible to collaborators; preview and approval required
  update_date_dependency_metadata:
    endpoint: PATCH /v0/{{ config.base_id }}/{{ config.table_id }}/{{ record.id }}/dateDependencyMetadata
    required fields: id, predecessorRecordId, dateDependencyMetadata
    risk: Airtable schema mutation visible to collaborators; preview and approval required

SECURITY
  read risk: external Airtable API reads for base, table, record, comment, webhook, SCIM detail, collaborator, enterprise, audit, eDiscovery, and change-event metadata
  write risk: typed Airtable API mutations for single records, schema fields, comments, webhooks, selected admin actions, and attachment upload; destructive/admin actions require plan, preview, explicit approval, and execute while non-empty array batch operations remain blocked
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

COMMAND SURFACE
  Airtable provider-style commands for fixture-tested read, direct HyperDB query, and typed reverse-ETL action discovery.
  Usage: pm airtable <command> [flags]
  Source CLI: Airtable Web API (official OpenAPI operationId ledger)
  Read/query commands
  Reverse ETL/write action references
  Other Commands
    read scim-group - Get group [intent=etl availability=implemented stream=scim_group]; flags: --group-id
    read scim-user - Get user [intent=etl availability=implemented stream=scim_user]; flags: --user-id
    read webhooks - List webhooks [intent=etl availability=implemented stream=webhooks]; flags: --base-id
    read webhook-payloads - List webhook payloads [intent=etl availability=implemented stream=webhook_payloads]; flags: --base-id, --webhook-id, --page-size
    read bases - List bases [intent=etl availability=implemented stream=bases]
    read base-collaborators - Get base collaborators [intent=etl availability=implemented stream=base_collaborators]; flags: --base-id
    read block-installations - List block installations [intent=etl availability=implemented stream=block_installations]; flags: --base-id
    read interface - Get interface [intent=etl availability=implemented stream=interface]; flags: --base-id, --page-bundle-id
    read shares - List shares [intent=etl availability=implemented stream=shares]; flags: --base-id
    read tables - Get base schema [intent=etl availability=implemented stream=tables]; flags: --base-id
    read views - List views [intent=etl availability=implemented stream=views]; flags: --base-id
    read view-metadata - Get view metadata [intent=etl availability=implemented stream=view_metadata]; flags: --base-id, --view-id
    read enterprise - Get enterprise [intent=etl availability=implemented stream=enterprise]; flags: --enterprise-account-id
    read audit-log-events - Audit log events [intent=etl availability=implemented stream=audit_log_events]; flags: --enterprise-account-id, --page-size
    read audit-log-requests - List audit log requests [intent=etl availability=implemented stream=audit_log_requests]; flags: --enterprise-account-id, --page-size
    read audit-log-request - Get audit log request [intent=etl availability=implemented stream=audit_log_request]; flags: --enterprise-account-id, --enterprise-audit-log-task-id
    read change-events - Change events [intent=etl availability=implemented stream=change_events]; flags: --enterprise-account-id, --page-size
    read ediscovery-exports - List eDiscovery exports [intent=etl availability=implemented stream=ediscovery_exports]; flags: --enterprise-account-id, --page-size
    read ediscovery-export - Get eDiscovery export [intent=etl availability=implemented stream=ediscovery_export]; flags: --enterprise-account-id, --enterprise-task-id
    read enterprise-packages - List packages [intent=etl availability=implemented stream=enterprise_packages]; flags: --enterprise-account-id
    read enterprise-personal-access-tokens - List personal access tokens [intent=etl availability=implemented stream=enterprise_personal_access_tokens]; flags: --enterprise-account-id
    read enterprise-users - Get users by id or email [intent=etl availability=implemented stream=enterprise_users]; flags: --enterprise-account-id
    read enterprise-user - Get user by id [intent=etl availability=implemented stream=enterprise_user]; flags: --enterprise-account-id, --user-id
    read user-group - Get user group [intent=etl availability=implemented stream=user_group]; flags: --group-id
    read whoami - Get user info [intent=etl availability=implemented stream=whoami]
    read workspace-collaborators - Get workspace collaborators [intent=etl availability=implemented stream=workspace_collaborators]; flags: --workspace-id
    read records - List records [intent=etl availability=implemented stream=records]; flags: --base-id, --table-id, --page-size
    read record - Get record [intent=etl availability=implemented stream=record]; flags: --base-id, --table-id, --record-id
    read comments - List comments [intent=etl availability=implemented stream=comments]; flags: --base-id, --table-id, --record-id
    hyperdb get-records - Read HyperDB records by primary key [intent=direct_read availability=implemented operation=hyperdb_table_read_records]; approval: read-only; risk: medium; flags: --enterprise-account-id, --data-table-id, --primary-key (required), --field, --max-records, --cursor
    write delete-scim-group - Delete group [intent=reverse_etl availability=planned write=delete_scim_group]; approval: Use reverse ETL plan -> preview -> explicit approval -> execute with typed record input; provider-style flag execution is not exposed for arbitrary object bodies.; risk: high
    write delete-scim-user - Delete user [intent=reverse_etl availability=planned write=delete_scim_user]; approval: Use reverse ETL plan -> preview -> explicit approval -> execute with typed record input; provider-style flag execution is not exposed for arbitrary object bodies.; risk: high
    write create-webhook - Create a webhook [intent=reverse_etl availability=planned write=create_webhook]; approval: Use reverse ETL plan -> preview -> explicit approval -> execute with typed record input; provider-style flag execution is not exposed for arbitrary object bodies.; risk: medium
    write delete-webhook - Delete a webhook [intent=reverse_etl availability=planned write=delete_webhook]; approval: Use reverse ETL plan -> preview -> explicit approval -> execute with typed record input; provider-style flag execution is not exposed for arbitrary object bodies.; risk: high
    write set-webhook-notifications - Enable/disable webhook notifications [intent=reverse_etl availability=planned write=set_webhook_notifications]; approval: Use reverse ETL plan -> preview -> explicit approval -> execute with typed record input; provider-style flag execution is not exposed for arbitrary object bodies.; risk: medium
    write refresh-webhook - Refresh a webhook [intent=reverse_etl availability=planned write=refresh_webhook]; approval: Use reverse ETL plan -> preview -> explicit approval -> execute with typed record input; provider-style flag execution is not exposed for arbitrary object bodies.; risk: medium
    write delete-base - Delete base [intent=reverse_etl availability=planned write=delete_base]; approval: Use reverse ETL plan -> preview -> explicit approval -> execute with typed record input; provider-style flag execution is not exposed for arbitrary object bodies.; risk: high
    write delete-block-installation - Delete block installation [intent=reverse_etl availability=planned write=delete_block_installation]; approval: Use reverse ETL plan -> preview -> explicit approval -> execute with typed record input; provider-style flag execution is not exposed for arbitrary object bodies.; risk: high
    write manage-block-installation - Manage block installation [intent=reverse_etl availability=planned write=manage_block_installation]; approval: Use reverse ETL plan -> preview -> explicit approval -> execute with typed record input; provider-style flag execution is not exposed for arbitrary object bodies.; risk: medium
    write delete-base-collaborator - Delete base collaborator [intent=reverse_etl availability=planned write=delete_base_collaborator]; approval: Use reverse ETL plan -> preview -> explicit approval -> execute with typed record input; provider-style flag execution is not exposed for arbitrary object bodies.; risk: high
    write update-collaborator-base-permission - Update collaborator base permission [intent=reverse_etl availability=planned write=update_collaborator_base_permission]; approval: Use reverse ETL plan -> preview -> explicit approval -> execute with typed record input; provider-style flag execution is not exposed for arbitrary object bodies.; risk: medium
    write delete-interface-collaborator - Delete interface collaborator [intent=reverse_etl availability=planned write=delete_interface_collaborator]; approval: Use reverse ETL plan -> preview -> explicit approval -> execute with typed record input; provider-style flag execution is not exposed for arbitrary object bodies.; risk: high
    write update-interface-collaborator - Update interface collaborator [intent=reverse_etl availability=planned write=update_interface_collaborator]; approval: Use reverse ETL plan -> preview -> explicit approval -> execute with typed record input; provider-style flag execution is not exposed for arbitrary object bodies.; risk: medium
    write delete-interface-invite - Delete interface invite [intent=reverse_etl availability=planned write=delete_interface_invite]; approval: Use reverse ETL plan -> preview -> explicit approval -> execute with typed record input; provider-style flag execution is not exposed for arbitrary object bodies.; risk: high
    write delete-base-invite - Delete base invite [intent=reverse_etl availability=planned write=delete_base_invite]; approval: Use reverse ETL plan -> preview -> explicit approval -> execute with typed record input; provider-style flag execution is not exposed for arbitrary object bodies.; risk: high
    write delete-share - Delete share [intent=reverse_etl availability=planned write=delete_share]; approval: Use reverse ETL plan -> preview -> explicit approval -> execute with typed record input; provider-style flag execution is not exposed for arbitrary object bodies.; risk: high
    write manage-share - Manage share [intent=reverse_etl availability=planned write=manage_share]; approval: Use reverse ETL plan -> preview -> explicit approval -> execute with typed record input; provider-style flag execution is not exposed for arbitrary object bodies.; risk: medium
    write update-table - Update table [intent=reverse_etl availability=planned write=update_table]; approval: Use reverse ETL plan -> preview -> explicit approval -> execute with typed record input; provider-style flag execution is not exposed for arbitrary object bodies.; risk: medium
    write create-field - Create field [intent=reverse_etl availability=planned write=create_field]; approval: Use reverse ETL plan -> preview -> explicit approval -> execute with typed record input; provider-style flag execution is not exposed for arbitrary object bodies.; risk: medium
    write update-field - Update field [intent=reverse_etl availability=planned write=update_field]; approval: Use reverse ETL plan -> preview -> explicit approval -> execute with typed record input; provider-style flag execution is not exposed for arbitrary object bodies.; risk: medium
    write delete-view - Delete view [intent=reverse_etl availability=planned write=delete_view]; approval: Use reverse ETL plan -> preview -> explicit approval -> execute with typed record input; provider-style flag execution is not exposed for arbitrary object bodies.; risk: high
    write create-audit-log-request - Create audit log request [intent=reverse_etl availability=planned write=create_audit_log_request]; approval: Use reverse ETL plan -> preview -> explicit approval -> execute with typed record input; provider-style flag execution is not exposed for arbitrary object bodies.; risk: high
    write create-descendant-enterprise - Create descendant enterprise [intent=reverse_etl availability=planned write=create_descendant_enterprise]; approval: Use reverse ETL plan -> preview -> explicit approval -> execute with typed record input; provider-style flag execution is not exposed for arbitrary object bodies.; risk: high
    write create-ediscovery-export - Create eDiscovery export [intent=reverse_etl availability=planned write=create_ediscovery_export]; approval: Use reverse ETL plan -> preview -> explicit approval -> execute with typed record input; provider-style flag execution is not exposed for arbitrary object bodies.; risk: high
    write create-base-from-package-enterprise - Create base from package [intent=reverse_etl availability=planned write=create_base_from_package_enterprise]; approval: Use reverse ETL plan -> preview -> explicit approval -> execute with typed record input; provider-style flag execution is not exposed for arbitrary object bodies.; risk: high
    write delete-users-by-email - Delete users by email [intent=reverse_etl availability=planned write=delete_users_by_email]; approval: Use reverse ETL plan -> preview -> explicit approval -> execute with typed record input; provider-style flag execution is not exposed for arbitrary object bodies.; risk: high
    write delete-user-by-id - Delete user by id [intent=reverse_etl availability=planned write=delete_user_by_id]; approval: Use reverse ETL plan -> preview -> explicit approval -> execute with typed record input; provider-style flag execution is not exposed for arbitrary object bodies.; risk: high
    write manage-user - Manage user [intent=reverse_etl availability=planned write=manage_user]; approval: Use reverse ETL plan -> preview -> explicit approval -> execute with typed record input; provider-style flag execution is not exposed for arbitrary object bodies.; risk: high
    write logout-user - Logout user [intent=reverse_etl availability=planned write=logout_user]; approval: Use reverse ETL plan -> preview -> explicit approval -> execute with typed record input; provider-style flag execution is not exposed for arbitrary object bodies.; risk: high
    write remove-user-from-enterprise - Remove user from enterprise [intent=reverse_etl availability=planned write=remove_user_from_enterprise]; approval: Use reverse ETL plan -> preview -> explicit approval -> execute with typed record input; provider-style flag execution is not exposed for arbitrary object bodies.; risk: high
    write create-workspace - Create workspace [intent=reverse_etl availability=planned write=create_workspace]; approval: Use reverse ETL plan -> preview -> explicit approval -> execute with typed record input; provider-style flag execution is not exposed for arbitrary object bodies.; risk: medium
    write delete-workspace - Delete workspace [intent=reverse_etl availability=planned write=delete_workspace]; approval: Use reverse ETL plan -> preview -> explicit approval -> execute with typed record input; provider-style flag execution is not exposed for arbitrary object bodies.; risk: high
    write delete-workspace-collaborator - Delete workspace collaborator [intent=reverse_etl availability=planned write=delete_workspace_collaborator]; approval: Use reverse ETL plan -> preview -> explicit approval -> execute with typed record input; provider-style flag execution is not exposed for arbitrary object bodies.; risk: high
    write update-workspace-collaborator - Update workspace collaborator [intent=reverse_etl availability=planned write=update_workspace_collaborator]; approval: Use reverse ETL plan -> preview -> explicit approval -> execute with typed record input; provider-style flag execution is not exposed for arbitrary object bodies.; risk: medium
    write delete-workspace-invite - Delete workspace invite [intent=reverse_etl availability=planned write=delete_workspace_invite]; approval: Use reverse ETL plan -> preview -> explicit approval -> execute with typed record input; provider-style flag execution is not exposed for arbitrary object bodies.; risk: high
    write move-base - Move base [intent=reverse_etl availability=planned write=move_base]; approval: Use reverse ETL plan -> preview -> explicit approval -> execute with typed record input; provider-style flag execution is not exposed for arbitrary object bodies.; risk: medium
    write update-workspace-restrictions - Update workspace restrictions [intent=reverse_etl availability=planned write=update_workspace_restrictions]; approval: Use reverse ETL plan -> preview -> explicit approval -> execute with typed record input; provider-style flag execution is not exposed for arbitrary object bodies.; risk: medium
    write upload-attachment - Upload attachment [intent=reverse_etl availability=planned write=upload_attachment]; approval: Use reverse ETL plan -> preview -> explicit approval -> execute with typed record input; provider-style flag execution is not exposed for arbitrary object bodies.; risk: medium
    write delete-record - Delete record [intent=reverse_etl availability=planned write=delete_record]; approval: Use reverse ETL plan -> preview -> explicit approval -> execute with typed record input; provider-style flag execution is not exposed for arbitrary object bodies.; risk: high
    write update-record - Update record [intent=reverse_etl availability=planned write=update_record]; approval: Use reverse ETL plan -> preview -> explicit approval -> execute with typed record input; provider-style flag execution is not exposed for arbitrary object bodies.; risk: medium
    write replace-record - Update record (put) [intent=reverse_etl availability=planned write=replace_record]; approval: Use reverse ETL plan -> preview -> explicit approval -> execute with typed record input; provider-style flag execution is not exposed for arbitrary object bodies.; risk: high
    write create-comment - Create comment [intent=reverse_etl availability=planned write=create_comment]; approval: Use reverse ETL plan -> preview -> explicit approval -> execute with typed record input; provider-style flag execution is not exposed for arbitrary object bodies.; risk: medium
    write delete-comment - Delete comment [intent=reverse_etl availability=planned write=delete_comment]; approval: Use reverse ETL plan -> preview -> explicit approval -> execute with typed record input; provider-style flag execution is not exposed for arbitrary object bodies.; risk: high
    write update-comment - Update comment [intent=reverse_etl availability=planned write=update_comment]; approval: Use reverse ETL plan -> preview -> explicit approval -> execute with typed record input; provider-style flag execution is not exposed for arbitrary object bodies.; risk: medium
    write update-date-dependency-metadata - Update date dependency metadata [intent=reverse_etl availability=planned write=update_date_dependency_metadata]; approval: Use reverse ETL plan -> preview -> explicit approval -> execute with typed record input; provider-style flag execution is not exposed for arbitrary object bodies.; risk: medium
  Help topics:
    airtable-safety - Writes remain reverse ETL only: plan, preview, explicit approval, execute; no generic raw HTTP, SQL, or CSV write escape hatches.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect airtable

  # Inspect as structured JSON
  pm connectors inspect airtable --json

AGENT WORKFLOW
  - Run pm connectors inspect airtable before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```

