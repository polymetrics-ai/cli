# Overview

Airtable connector bundle expanded to official Web API parity from the embedded OpenAPI specification. It reads Web API, SCIM, metadata, comments, webhook payloads, enterprise/admin, audit/eDiscovery/change-event data, exposes the HyperDB getRecords direct read, and declares typed reverse-ETL write actions for every supportable JSON/no-body mutation.

Official documentation: https://airtable.com/developers/web/api/introduction. OpenAPI SHA-256: `f1506571034500ffb0887e7244f770c0b060a4205e734881bd3de6fa20d6f6b8`. Fixture-only validation; no live provider calls or certification claims.

## Auth setup

- `api_key` (secret): Airtable Personal Access Token, preferred when both token fields are set.
- `access_token` (secret): Airtable OAuth2 access token fallback.
- `base_url`: default `https://api.airtable.com`; connector paths include `/v0` or `/scim`.
- Scope/config selectors include `base_id`, `table_id`, `record_id`, `webhook_id`, `enterprise_account_id`, `workspace_id`, `group_id`, `user_id`, `view_id`, and related task ids.

Authentication uses Bearer tokens and the check endpoint `GET /v0/meta/bases`. Secret fields are redacted in logs and previews.

## Streams notes

31 fixture-backed streams cover all official GET operations. Default pagination uses `offset`; SCIM/webhook-payload/audit endpoints declare their provider-specific cursor parameters where applicable.

- `scim_groups`: GET `/scim/v2/Groups` -> `Resources`; query count.

- `scim_group`: GET `/scim/v2/Groups/{{ config.group_id }}` -> `.`.

- `scim_users`: GET `/scim/v2/Users` -> `Resources`; query count.

- `scim_user`: GET `/scim/v2/Users/{{ config.user_id }}` -> `.`.

- `webhooks`: GET `/v0/bases/{{ config.base_id }}/webhooks` -> `webhooks`.

- `webhook_payloads`: GET `/v0/bases/{{ config.base_id }}/webhooks/{{ config.webhook_id }}/payloads` -> `payloads`; query limit.

- `bases`: GET `/v0/meta/bases` -> `bases`.

- `base_collaborators`: GET `/v0/meta/bases/{{ config.base_id }}` -> `.`.

- `block_installations`: GET `/v0/meta/bases/{{ config.base_id }}/blockInstallations` -> `.`.

- `interface`: GET `/v0/meta/bases/{{ config.base_id }}/interfaces/{{ config.page_bundle_id }}` -> `.`.

- `shares`: GET `/v0/meta/bases/{{ config.base_id }}/shares` -> `shares`.

- `tables`: GET `/v0/meta/bases/{{ config.base_id }}/tables` -> `tables`.

- `views`: GET `/v0/meta/bases/{{ config.base_id }}/views` -> `views`.

- `view_metadata`: GET `/v0/meta/bases/{{ config.base_id }}/views/{{ config.view_id }}` -> `.`.

- `enterprise`: GET `/v0/meta/enterpriseAccounts/{{ config.enterprise_account_id }}` -> `.`.

- `audit_log_events`: GET `/v0/meta/enterpriseAccounts/{{ config.enterprise_account_id }}/auditLogEvents` -> `events`; query pageSize.

- `audit_log_requests`: GET `/v0/meta/enterpriseAccounts/{{ config.enterprise_account_id }}/auditLogs` -> `auditLogs`; query pageSize.

- `audit_log_request`: GET `/v0/meta/enterpriseAccounts/{{ config.enterprise_account_id }}/auditLogs/{{ config.enterprise_audit_log_task_id }}` -> `.`.

- `change_events`: GET `/v0/meta/enterpriseAccounts/{{ config.enterprise_account_id }}/changeEvents` -> `events`; query pageSize.

- `ediscovery_exports`: GET `/v0/meta/enterpriseAccounts/{{ config.enterprise_account_id }}/exports` -> `exports`; query pageSize.

- `ediscovery_export`: GET `/v0/meta/enterpriseAccounts/{{ config.enterprise_account_id }}/exports/{{ config.enterprise_task_id }}` -> `.`.

- `enterprise_packages`: GET `/v0/meta/enterpriseAccounts/{{ config.enterprise_account_id }}/packages` -> `packages`.

- `enterprise_personal_access_tokens`: GET `/v0/meta/enterpriseAccounts/{{ config.enterprise_account_id }}/personalAccessTokens` -> `personalAccessTokens`.

- `enterprise_users`: GET `/v0/meta/enterpriseAccounts/{{ config.enterprise_account_id }}/users` -> `users`.

- `enterprise_user`: GET `/v0/meta/enterpriseAccounts/{{ config.enterprise_account_id }}/users/{{ config.user_id }}` -> `.`.

- `user_group`: GET `/v0/meta/groups/{{ config.group_id }}` -> `.`.

- `whoami`: GET `/v0/meta/whoami` -> `.`.

- `workspace_collaborators`: GET `/v0/meta/workspaces/{{ config.workspace_id }}` -> `.`.

- `records`: GET `/v0/{{ config.base_id }}/{{ config.table_id }}` -> `records`; query pageSize.

- `record`: GET `/v0/{{ config.base_id }}/{{ config.table_id }}/{{ config.record_id }}` -> `.`.

- `comments`: GET `/v0/{{ config.base_id }}/{{ config.table_id }}/{{ fanout.id }}/comments` -> `comments`; fan-out.

## Direct read

- `hyperdb get-records`: POST `/v0/meta/enterpriseAccounts/{enterpriseAccountId}/hyperdb/dataTables/{dataTableId}/records`, typed JSON body with `primaryKeys`, optional `fields`, `maxRecords`, and `cursor`; output policy `json_redacted`.

## Write actions & risks

70 fixture-backed write actions cover supportable JSON/no-body POST/PATCH/PUT/DELETE operations. Every write must use reverse ETL plan -> preview -> explicit approval -> execute. DELETE actions treat 404 as missing-ok idempotent success where supported; DELETE/PUT/revoke/remove/logout actions require destructive confirmation.

- `create_scim_group`: POST `/scim/v2/Groups`; kind `create`; body `json`; path fields `none`; required `schemas, displayName`.

- `delete_scim_group`: DELETE `/scim/v2/Groups/{{ record.id }}`; kind `delete`; body `none`; path fields `id`; required `id`.

- `patch_scim_group`: PATCH `/scim/v2/Groups/{{ record.id }}`; kind `update`; body `json`; path fields `id`; required `id, schemas, Operations`.

- `put_scim_group`: PUT `/scim/v2/Groups/{{ record.id }}`; kind `update`; body `json`; path fields `id`; required `id, schemas`.

- `create_scim_user`: POST `/scim/v2/Users`; kind `create`; body `json`; path fields `none`; required `schemas, userName`.

- `delete_scim_user`: DELETE `/scim/v2/Users/{{ record.id }}`; kind `delete`; body `none`; path fields `id`; required `id`.

- `patch_scim_user`: PATCH `/scim/v2/Users/{{ record.id }}`; kind `update`; body `json`; path fields `id`; required `id, schemas, Operations`.

- `put_scim_user`: PUT `/scim/v2/Users/{{ record.id }}`; kind `update`; body `json`; path fields `id`; required `id, schemas, userName`.

- `create_webhook`: POST `/v0/bases/{{ config.base_id }}/webhooks`; kind `create`; body `json`; path fields `none`; required `specification`.

- `delete_webhook`: DELETE `/v0/bases/{{ config.base_id }}/webhooks/{{ record.id }}`; kind `delete`; body `none`; path fields `id`; required `id`.

- `set_webhook_notifications`: POST `/v0/bases/{{ config.base_id }}/webhooks/{{ record.id }}/enableNotifications`; kind `update`; body `json`; path fields `id`; required `id, enable`.

- `refresh_webhook`: POST `/v0/bases/{{ config.base_id }}/webhooks/{{ record.id }}/refresh`; kind `update`; body `json`; path fields `id`; required `id`.

- `create_base`: POST `/v0/meta/bases`; kind `create`; body `json`; path fields `none`; required `workspaceId, name, tables`.

- `delete_base`: DELETE `/v0/meta/bases/{{ record.base_id }}`; kind `delete`; body `none`; path fields `base_id`; required `base_id`.

- `delete_block_installation`: DELETE `/v0/meta/bases/{{ config.base_id }}/blockInstallations/{{ record.id }}`; kind `delete`; body `none`; path fields `id`; required `id`.

- `manage_block_installation`: PATCH `/v0/meta/bases/{{ config.base_id }}/blockInstallations/{{ record.id }}`; kind `update`; body `json`; path fields `id`; required `id, state`.

- `add_base_collaborator`: POST `/v0/meta/bases/{{ config.base_id }}/collaborators`; kind `create`; body `json`; path fields `none`; required `collaborators`.

- `delete_base_collaborator`: DELETE `/v0/meta/bases/{{ config.base_id }}/collaborators/{{ record.user_or_group_id }}`; kind `delete`; body `none`; path fields `user_or_group_id`; required `user_or_group_id`.

- `update_collaborator_base_permission`: PATCH `/v0/meta/bases/{{ config.base_id }}/collaborators/{{ record.user_or_group_id }}`; kind `update`; body `json`; path fields `user_or_group_id`; required `user_or_group_id, permissionLevel`.

- `add_interface_collaborator`: POST `/v0/meta/bases/{{ config.base_id }}/interfaces/{{ record.page_bundle_id }}/collaborators`; kind `create`; body `json`; path fields `page_bundle_id`; required `page_bundle_id, collaborators`.

- `delete_interface_collaborator`: DELETE `/v0/meta/bases/{{ config.base_id }}/interfaces/{{ record.page_bundle_id }}/collaborators/{{ record.user_or_group_id }}`; kind `delete`; body `none`; path fields `page_bundle_id, user_or_group_id`; required `page_bundle_id, user_or_group_id`.

- `update_interface_collaborator`: PATCH `/v0/meta/bases/{{ config.base_id }}/interfaces/{{ record.page_bundle_id }}/collaborators/{{ record.user_or_group_id }}`; kind `update`; body `json`; path fields `page_bundle_id, user_or_group_id`; required `page_bundle_id, user_or_group_id, permissionLevel`.

- `delete_interface_invite`: DELETE `/v0/meta/bases/{{ config.base_id }}/interfaces/{{ record.page_bundle_id }}/invites/{{ record.invite_id }}`; kind `delete`; body `none`; path fields `page_bundle_id, invite_id`; required `page_bundle_id, invite_id`.

- `delete_base_invite`: DELETE `/v0/meta/bases/{{ config.base_id }}/invites/{{ record.invite_id }}`; kind `delete`; body `none`; path fields `invite_id`; required `invite_id`.

- `delete_share`: DELETE `/v0/meta/bases/{{ config.base_id }}/shares/{{ record.id }}`; kind `delete`; body `none`; path fields `id`; required `id`.

- `manage_share`: PATCH `/v0/meta/bases/{{ config.base_id }}/shares/{{ record.id }}`; kind `update`; body `json`; path fields `id`; required `id, state`.

- `create_table`: POST `/v0/meta/bases/{{ config.base_id }}/tables`; kind `create`; body `json`; path fields `none`; required `name, fields`.

- `update_table`: PATCH `/v0/meta/bases/{{ config.base_id }}/tables/{{ config.table_id }}`; kind `update`; body `json`; path fields `none`; required `none`.

- `create_field`: POST `/v0/meta/bases/{{ config.base_id }}/tables/{{ record.table_id }}/fields`; kind `create`; body `json`; path fields `table_id`; required `table_id, name, type`.

- `update_field`: PATCH `/v0/meta/bases/{{ config.base_id }}/tables/{{ record.table_id }}/fields/{{ record.column_id }}`; kind `update`; body `json`; path fields `table_id, column_id`; required `table_id, column_id`.

- `delete_view`: DELETE `/v0/meta/bases/{{ config.base_id }}/views/{{ record.id }}`; kind `delete`; body `none`; path fields `id`; required `id`.

- `create_audit_log_request`: POST `/v0/meta/enterpriseAccounts/{{ record.enterprise_account_id }}/auditLogs`; kind `create`; body `json`; path fields `enterprise_account_id`; required `enterprise_account_id, timePeriod`.

- `create_descendant_enterprise`: POST `/v0/meta/enterpriseAccounts/{{ record.enterprise_account_id }}/descendants`; kind `create`; body `json`; path fields `enterprise_account_id`; required `enterprise_account_id, name`.

- `create_ediscovery_export`: POST `/v0/meta/enterpriseAccounts/{{ record.enterprise_account_id }}/exports`; kind `create`; body `json`; path fields `enterprise_account_id`; required `enterprise_account_id, baseId`.

- `move_user_groups`: POST `/v0/meta/enterpriseAccounts/{{ record.enterprise_account_id }}/moveGroups`; kind `update`; body `json`; path fields `enterprise_account_id`; required `enterprise_account_id, groupIds, targetEnterpriseAccountId`.

- `move_workspaces`: POST `/v0/meta/enterpriseAccounts/{{ record.enterprise_account_id }}/moveWorkspaces`; kind `update`; body `json`; path fields `enterprise_account_id`; required `enterprise_account_id, workspaceIds, targetEnterpriseAccountId`.

- `create_base_from_package_enterprise`: POST `/v0/meta/enterpriseAccounts/{{ record.enterprise_account_id }}/packages/{{ record.package_id }}/install`; kind `create`; body `json`; path fields `enterprise_account_id, package_id`; required `enterprise_account_id, package_id, workspaceId, packageReleaseId, name`.

- `revoke_enterprise_personal_access_tokens`: POST `/v0/meta/enterpriseAccounts/{{ record.enterprise_account_id }}/personalAccessTokens/revoke`; kind `update`; body `json`; path fields `enterprise_account_id`; required `enterprise_account_id, tokenIds`.

- `delete_users_by_email`: DELETE `/v0/meta/enterpriseAccounts/{{ record.enterprise_account_id }}/users?email={{ record.email | urlencode }}`; kind `delete`; body `none`; path fields `enterprise_account_id, email`; required `enterprise_account_id, email`.

- `manage_user_batched`: PATCH `/v0/meta/enterpriseAccounts/{{ record.enterprise_account_id }}/users`; kind `update`; body `json`; path fields `enterprise_account_id`; required `enterprise_account_id, users`.

- `manage_user_membership`: POST `/v0/meta/enterpriseAccounts/{{ record.enterprise_account_id }}/users/claim`; kind `update`; body `json`; path fields `enterprise_account_id`; required `enterprise_account_id, users`.

- `grant_admin_access`: POST `/v0/meta/enterpriseAccounts/{{ record.enterprise_account_id }}/users/grantAdminAccess`; kind `create`; body `json`; path fields `enterprise_account_id`; required `enterprise_account_id, users`.

- `revoke_admin_access`: POST `/v0/meta/enterpriseAccounts/{{ record.enterprise_account_id }}/users/revokeAdminAccess`; kind `update`; body `json`; path fields `enterprise_account_id`; required `enterprise_account_id, users`.

- `delete_user_by_id`: DELETE `/v0/meta/enterpriseAccounts/{{ record.enterprise_account_id }}/users/{{ record.id }}`; kind `delete`; body `none`; path fields `enterprise_account_id, id`; required `enterprise_account_id, id`.

- `manage_user`: PATCH `/v0/meta/enterpriseAccounts/{{ record.enterprise_account_id }}/users/{{ record.id }}`; kind `update`; body `json`; path fields `enterprise_account_id, id`; required `enterprise_account_id, id`.

- `logout_user`: POST `/v0/meta/enterpriseAccounts/{{ record.enterprise_account_id }}/users/{{ record.id }}/logout`; kind `update`; body `json`; path fields `enterprise_account_id, id`; required `enterprise_account_id, id`.

- `remove_user_from_enterprise`: POST `/v0/meta/enterpriseAccounts/{{ record.enterprise_account_id }}/users/{{ record.id }}/remove`; kind `update`; body `json`; path fields `enterprise_account_id, id`; required `enterprise_account_id, id`.

- `update_workspace_ai_allowlist`: POST `/v0/meta/enterpriseAccounts/{{ record.enterprise_account_id }}/workspaceAiAllowlist`; kind `update`; body `json`; path fields `enterprise_account_id`; required `enterprise_account_id, workspaces`.

- `create_workspace`: POST `/v0/meta/workspaces`; kind `create`; body `json`; path fields `none`; required `enterpriseAccountId, name`.

- `delete_workspace`: DELETE `/v0/meta/workspaces/{{ record.workspace_id }}`; kind `delete`; body `none`; path fields `workspace_id`; required `workspace_id`.

- `add_workspace_collaborator`: POST `/v0/meta/workspaces/{{ record.workspace_id }}/collaborators`; kind `create`; body `json`; path fields `workspace_id`; required `workspace_id, collaborators`.

- `delete_workspace_collaborator`: DELETE `/v0/meta/workspaces/{{ record.workspace_id }}/collaborators/{{ record.user_or_group_id }}`; kind `delete`; body `none`; path fields `workspace_id, user_or_group_id`; required `workspace_id, user_or_group_id`.

- `update_workspace_collaborator`: PATCH `/v0/meta/workspaces/{{ record.workspace_id }}/collaborators/{{ record.user_or_group_id }}`; kind `update`; body `json`; path fields `workspace_id, user_or_group_id`; required `workspace_id, user_or_group_id, permissionLevel`.

- `delete_workspace_invite`: DELETE `/v0/meta/workspaces/{{ record.workspace_id }}/invites/{{ record.invite_id }}`; kind `delete`; body `none`; path fields `workspace_id, invite_id`; required `workspace_id, invite_id`.

- `move_base`: POST `/v0/meta/workspaces/{{ record.workspace_id }}/moveBase`; kind `update`; body `json`; path fields `workspace_id`; required `workspace_id, baseId, targetWorkspaceId`.

- `update_workspace_restrictions`: POST `/v0/meta/workspaces/{{ record.workspace_id }}/updateRestrictions`; kind `update`; body `json`; path fields `workspace_id`; required `workspace_id`.

- `upload_attachment`: POST `/v0/{{ config.base_id }}/{{ record.record_id }}/{{ record.attachment_field_id_or_name }}/uploadAttachment`; kind `custom`; body `json`; path fields `record_id, attachment_field_id_or_name`; required `record_id, attachment_field_id_or_name, contentType, filename, file`.

- `delete_multiple_records`: DELETE `/v0/{{ config.base_id }}/{{ config.table_id }}?records[]={{ record.records | join:&records[]= }}`; kind `delete`; body `none`; path fields `records`; required `records`.

- `update_multiple_records`: PATCH `/v0/{{ config.base_id }}/{{ config.table_id }}`; kind `update`; body `json`; path fields `none`; required `records`.

- `create_records`: POST `/v0/{{ config.base_id }}/{{ config.table_id }}`; kind `create`; body `json`; path fields `none`; required `records`.

- `update_multiple_records_put`: PUT `/v0/{{ config.base_id }}/{{ config.table_id }}`; kind `update`; body `json`; path fields `none`; required `records`.

- `delete_record`: DELETE `/v0/{{ config.base_id }}/{{ config.table_id }}/{{ record.id }}`; kind `delete`; body `none`; path fields `id`; required `id`.

- `update_record`: PATCH `/v0/{{ config.base_id }}/{{ config.table_id }}/{{ record.id }}`; kind `update`; body `json`; path fields `id`; required `id, fields`.

- `replace_record`: PUT `/v0/{{ config.base_id }}/{{ config.table_id }}/{{ record.id }}`; kind `update`; body `json`; path fields `id`; required `id, fields`.

- `create_comment`: POST `/v0/{{ config.base_id }}/{{ config.table_id }}/{{ record.record_id }}/comments`; kind `create`; body `json`; path fields `record_id`; required `record_id, text`.

- `delete_comment`: DELETE `/v0/{{ config.base_id }}/{{ config.table_id }}/{{ record.record_id }}/comments/{{ record.row_comment_id }}`; kind `delete`; body `none`; path fields `record_id, row_comment_id`; required `record_id, row_comment_id`.

- `update_comment`: PATCH `/v0/{{ config.base_id }}/{{ config.table_id }}/{{ record.record_id }}/comments/{{ record.row_comment_id }}`; kind `update`; body `json`; path fields `record_id, row_comment_id`; required `record_id, row_comment_id, text`.

- `update_date_dependency_metadata`: PATCH `/v0/{{ config.base_id }}/{{ config.table_id }}/{{ record.id }}/dateDependencyMetadata`; kind `update`; body `json`; path fields `id`; required `id, predecessorRecordId, dateDependencyMetadata`.

- `hyperdb_delete_records_by_primary_keys`: POST `/v0/{{ record.enterprise_account_id }}/{{ record.data_table_id }}/deleteRecords`; kind `custom`; body `json`; path fields `enterprise_account_id, data_table_id`; required `enterprise_account_id, data_table_id, primaryKeys`.

- `hyperdb_upsert_records_by_primary_keys`: PUT `/v0/{{ record.enterprise_account_id }}/{{ record.data_table_id }}/upsertRecords`; kind `upsert`; body `json`; path fields `enterprise_account_id, data_table_id`; required `enterprise_account_id, data_table_id, records`.

## Known limits

- Blocked row: `post-sync-api-endpoint` (`POST /v0/{baseId}/{apiEndpointSyncId}/sync`) because the official Sync API import is `text/csv`; the shared declarative write runtime has no typed bounded CSV body dialect and generic raw uploads remain disallowed.

- API-surface ledger rows: 103 total = 31 streams + 70 writes + 1 direct read + 1 blocked operation.

- Certification remains fixture-only/uncertified until separately approved live-safe credentials and provider resources are supplied outside this wave.
