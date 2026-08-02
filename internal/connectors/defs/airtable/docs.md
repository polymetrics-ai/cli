# Overview

Airtable connector bundle expanded to official Web API parity from the embedded OpenAPI specification. It reads Web API metadata, records, comments, webhook payloads, SCIM detail, enterprise/admin, audit/eDiscovery/change-event data, exposes the HyperDB getRecords direct read, and declares typed reverse-ETL write actions for supportable JSON/no-body mutations whose schemas are enforceable by the current runtime.

Official documentation: https://airtable.com/developers/web/api/introduction. OpenAPI SHA-256: `f1506571034500ffb0887e7244f770c0b060a4205e734881bd3de6fa20d6f6b8`. Fixture-only validation; no live provider calls or certification claims.

## Auth setup

- `api_key` (secret): Airtable Personal Access Token, preferred when both token fields are set.
- `access_token` (secret): Airtable OAuth2 access token fallback.
- `base_url`: default `https://api.airtable.com`; connector paths include `/v0` or `/scim`.
- Scope/config selectors include `base_id`, `table_id`, `record_id`, `webhook_id`, `enterprise_account_id`, `workspace_id`, `group_id`, `user_id`, `view_id`, and related task ids.

Authentication uses Bearer tokens and the check endpoint `GET /v0/meta/bases`. Secret fields are redacted in logs and previews.

## Streams notes

28 fixture-backed streams cover executable GET operations. Default pagination uses `offset`; webhook-payload pagination sends the documented maximum `limit=50` and consumes the documented `cursor` only while `mightHaveMore` is true, and audit endpoints declare provider cursor parameters where the current runtime can consume the documented token directly. SCIM list endpoints are blocked because Airtable SCIM responses expose `Resources`, `startIndex`, `itemsPerPage`, and `totalResults`, not a `nextStartIndex` token. Enterprise user search is blocked because the documented endpoint requires at least one `email[]` or `id[]` query value.

- `scim_group`: GET `/scim/v2/Groups/{{ config.group_id }}` -> `.`.

- `scim_user`: GET `/scim/v2/Users/{{ config.user_id }}` -> `.`.

- `webhooks`: GET `/v0/bases/{{ config.base_id }}/webhooks` -> `webhooks`.

- `webhook_payloads`: GET `/v0/bases/{{ config.base_id }}/webhooks/{{ config.webhook_id }}/payloads` -> `payloads`; query limit=50; cursor cursor; stop flag `mightHaveMore`.

- `bases`: GET `/v0/meta/bases` -> `bases`.

- `base_collaborators`: GET `/v0/meta/bases/{{ config.base_id }}` -> `.`.

- `block_installations`: GET `/v0/meta/bases/{{ config.base_id }}/blockInstallations` -> `.`.

- `interface`: GET `/v0/meta/bases/{{ config.base_id }}/interfaces/{{ config.page_bundle_id }}` -> `.`.

- `shares`: GET `/v0/meta/bases/{{ config.base_id }}/shares` -> `shares`.

- `tables`: GET `/v0/meta/bases/{{ config.base_id }}/tables` -> `tables`.

- `views`: GET `/v0/meta/bases/{{ config.base_id }}/views` -> `views`.

- `view_metadata`: GET `/v0/meta/bases/{{ config.base_id }}/views/{{ config.view_id }}` -> `.`.

- `enterprise`: GET `/v0/meta/enterpriseAccounts/{{ config.enterprise_account_id }}` -> `.`.

- `audit_log_events`: GET `/v0/meta/enterpriseAccounts/{{ config.enterprise_account_id }}/auditLogEvents` -> `events`; query pageSize; cursor pagination.next.

- `audit_log_requests`: GET `/v0/meta/enterpriseAccounts/{{ config.enterprise_account_id }}/auditLogs` -> `auditLogs`; query pageSize.

- `audit_log_request`: GET `/v0/meta/enterpriseAccounts/{{ config.enterprise_account_id }}/auditLogs/{{ config.enterprise_audit_log_task_id }}` -> `.`.

- `change_events`: GET `/v0/meta/enterpriseAccounts/{{ config.enterprise_account_id }}/changeEvents` -> `events`; query pageSize.

- `ediscovery_exports`: GET `/v0/meta/enterpriseAccounts/{{ config.enterprise_account_id }}/exports` -> `exports`; query pageSize.

- `ediscovery_export`: GET `/v0/meta/enterpriseAccounts/{{ config.enterprise_account_id }}/exports/{{ config.enterprise_task_id }}` -> `.`.

- `enterprise_packages`: GET `/v0/meta/enterpriseAccounts/{{ config.enterprise_account_id }}/packages` -> `packages`.

- `enterprise_personal_access_tokens`: GET `/v0/meta/enterpriseAccounts/{{ config.enterprise_account_id }}/personalAccessTokens` -> `personalAccessTokens`.

- `enterprise_user`: GET `/v0/meta/enterpriseAccounts/{{ config.enterprise_account_id }}/users/{{ config.user_id }}` -> `.`.

- `user_group`: GET `/v0/meta/groups/{{ config.group_id }}` -> `.`.

- `whoami`: GET `/v0/meta/whoami` -> `.`.

- `workspace_collaborators`: GET `/v0/meta/workspaces/{{ config.workspace_id }}` -> `.`.

- `records`: GET `/v0/{{ config.base_id }}/{{ config.table_id }}` -> `records`; query pageSize.

- `record`: GET `/v0/{{ config.base_id }}/{{ config.table_id }}/{{ config.record_id }}` -> `.`.

- `comments`: GET `/v0/{{ config.base_id }}/{{ config.table_id }}/{{ config.record_id }}/comments` -> `comments`; exact per-record read; `record_id` is required and stamped onto each emitted comment.

## Direct read

- `hyperdb get-records`: POST `/v0/{enterpriseAccountId}/{dataTableId}/getRecords`, typed JSON body with optional `primaryKeys`, `fields`, `maxRecords`, and `cursor` (including `{}` and cursor-only reads); output policy `json_redacted`.

## Write actions & risks

44 fixture-backed write actions cover supportable JSON/no-body mutations whose closed schemas are enforceable by the current runtime. Every write must use reverse ETL plan -> preview -> explicit approval -> execute. DELETE actions treat 404 as missing-ok idempotent success where supported; DELETE/PUT/revoke/remove/logout actions require destructive confirmation.

- `delete_scim_group`: DELETE `/scim/v2/Groups/{{ record.id }}`; kind `delete`; body `none`; path fields `id`; required `id`.

- `delete_scim_user`: DELETE `/scim/v2/Users/{{ record.id }}`; kind `delete`; body `none`; path fields `id`; required `id`.

- `create_webhook`: POST `/v0/bases/{{ config.base_id }}/webhooks`; kind `create`; body `json`; path fields `none`; required `specification`.

- `delete_webhook`: DELETE `/v0/bases/{{ config.base_id }}/webhooks/{{ record.id }}`; kind `delete`; body `none`; path fields `id`; required `id`.

- `set_webhook_notifications`: POST `/v0/bases/{{ config.base_id }}/webhooks/{{ record.id }}/enableNotifications`; kind `update`; body `json`; path fields `id`; required `id, enable`.

- `refresh_webhook`: POST `/v0/bases/{{ config.base_id }}/webhooks/{{ record.id }}/refresh`; kind `update`; body `json`; path fields `id`; required `id`.

- `delete_base`: DELETE `/v0/meta/bases/{{ record.base_id }}`; kind `delete`; body `none`; path fields `base_id`; required `base_id`.

- `delete_block_installation`: DELETE `/v0/meta/bases/{{ config.base_id }}/blockInstallations/{{ record.id }}`; kind `delete`; body `none`; path fields `id`; required `id`.

- `manage_block_installation`: PATCH `/v0/meta/bases/{{ config.base_id }}/blockInstallations/{{ record.id }}`; kind `update`; body `json`; path fields `id`; required `id, state`.

- `delete_base_collaborator`: DELETE `/v0/meta/bases/{{ config.base_id }}/collaborators/{{ record.user_or_group_id }}`; kind `delete`; body `none`; path fields `user_or_group_id`; required `user_or_group_id`.

- `update_collaborator_base_permission`: PATCH `/v0/meta/bases/{{ config.base_id }}/collaborators/{{ record.user_or_group_id }}`; kind `update`; body `json`; path fields `user_or_group_id`; required `user_or_group_id, permissionLevel`.

- `delete_interface_collaborator`: DELETE `/v0/meta/bases/{{ config.base_id }}/interfaces/{{ record.page_bundle_id }}/collaborators/{{ record.user_or_group_id }}`; kind `delete`; body `none`; path fields `page_bundle_id, user_or_group_id`; required `page_bundle_id, user_or_group_id`.

- `update_interface_collaborator`: PATCH `/v0/meta/bases/{{ config.base_id }}/interfaces/{{ record.page_bundle_id }}/collaborators/{{ record.user_or_group_id }}`; kind `update`; body `json`; path fields `page_bundle_id, user_or_group_id`; required `page_bundle_id, user_or_group_id, permissionLevel`.

- `delete_interface_invite`: DELETE `/v0/meta/bases/{{ config.base_id }}/interfaces/{{ record.page_bundle_id }}/invites/{{ record.invite_id }}`; kind `delete`; body `none`; path fields `page_bundle_id, invite_id`; required `page_bundle_id, invite_id`.

- `delete_base_invite`: DELETE `/v0/meta/bases/{{ config.base_id }}/invites/{{ record.invite_id }}`; kind `delete`; body `none`; path fields `invite_id`; required `invite_id`.

- `delete_share`: DELETE `/v0/meta/bases/{{ config.base_id }}/shares/{{ record.id }}`; kind `delete`; body `none`; path fields `id`; required `id`.

- `manage_share`: PATCH `/v0/meta/bases/{{ config.base_id }}/shares/{{ record.id }}`; kind `update`; body `json`; path fields `id`; required `id, state`.

- `update_table`: PATCH `/v0/meta/bases/{{ config.base_id }}/tables/{{ config.table_id }}`; kind `update`; body `json`; path fields `none`; required `none`.

- `create_field`: POST `/v0/meta/bases/{{ config.base_id }}/tables/{{ record.table_id }}/fields`; kind `create`; body `json`; path fields `table_id`; required `table_id, name, type`.

- `update_field`: PATCH `/v0/meta/bases/{{ config.base_id }}/tables/{{ record.table_id }}/fields/{{ record.column_id }}`; kind `update`; body `json`; path fields `table_id, column_id`; required `table_id, column_id`.

- `delete_view`: DELETE `/v0/meta/bases/{{ config.base_id }}/views/{{ record.id }}`; kind `delete`; body `none`; path fields `id`; required `id`.

- `create_audit_log_request`: POST `/v0/meta/enterpriseAccounts/{{ record.enterprise_account_id }}/auditLogs`; kind `create`; body `json`; path fields `enterprise_account_id`; required `enterprise_account_id, timePeriod`.

- `create_descendant_enterprise`: POST `/v0/meta/enterpriseAccounts/{{ record.enterprise_account_id }}/descendants`; kind `create`; body `json`; path fields `enterprise_account_id`; required `enterprise_account_id, name`.

- `create_ediscovery_export`: POST `/v0/meta/enterpriseAccounts/{{ record.enterprise_account_id }}/exports`; kind `create`; body `json`; path fields `enterprise_account_id`; required `enterprise_account_id, baseId`.

- `create_base_from_package_enterprise`: POST `/v0/meta/enterpriseAccounts/{{ record.enterprise_account_id }}/packages/{{ record.package_id }}/install`; kind `create`; body `json`; path fields `enterprise_account_id, package_id`; required `enterprise_account_id, package_id, workspaceId, packageReleaseId, name`.

- `delete_users_by_email`: DELETE `/v0/meta/enterpriseAccounts/{{ record.enterprise_account_id }}/users?email={{ record.email | urlencode }}`; kind `delete`; body `none`; path fields `enterprise_account_id, email`; required `enterprise_account_id, email`.

- `delete_user_by_id`: DELETE `/v0/meta/enterpriseAccounts/{{ record.enterprise_account_id }}/users/{{ record.id }}`; kind `delete`; body `none`; path fields `enterprise_account_id, id`; required `enterprise_account_id, id`.

- `manage_user`: PATCH `/v0/meta/enterpriseAccounts/{{ record.enterprise_account_id }}/users/{{ record.id }}`; kind `update`; body `json`; path fields `enterprise_account_id, id`; required `enterprise_account_id, id`.

- `logout_user`: POST `/v0/meta/enterpriseAccounts/{{ record.enterprise_account_id }}/users/{{ record.id }}/logout`; kind `update`; body `json`; path fields `enterprise_account_id, id`; required `enterprise_account_id, id`.

- `remove_user_from_enterprise`: POST `/v0/meta/enterpriseAccounts/{{ record.enterprise_account_id }}/users/{{ record.id }}/remove`; kind `update`; body `json`; path fields `enterprise_account_id, id`; required `enterprise_account_id, id`.

- `create_workspace`: POST `/v0/meta/workspaces`; kind `create`; body `json`; path fields `none`; required `enterpriseAccountId, name`.

- `delete_workspace`: DELETE `/v0/meta/workspaces/{{ record.workspace_id }}`; kind `delete`; body `none`; path fields `workspace_id`; required `workspace_id`.

- `delete_workspace_collaborator`: DELETE `/v0/meta/workspaces/{{ record.workspace_id }}/collaborators/{{ record.user_or_group_id }}`; kind `delete`; body `none`; path fields `workspace_id, user_or_group_id`; required `workspace_id, user_or_group_id`.

- `update_workspace_collaborator`: PATCH `/v0/meta/workspaces/{{ record.workspace_id }}/collaborators/{{ record.user_or_group_id }}`; kind `update`; body `json`; path fields `workspace_id, user_or_group_id`; required `workspace_id, user_or_group_id, permissionLevel`.

- `delete_workspace_invite`: DELETE `/v0/meta/workspaces/{{ record.workspace_id }}/invites/{{ record.invite_id }}`; kind `delete`; body `none`; path fields `workspace_id, invite_id`; required `workspace_id, invite_id`.

- `move_base`: POST `/v0/meta/workspaces/{{ record.workspace_id }}/moveBase`; kind `update`; body `json`; path fields `workspace_id`; required `workspace_id, baseId, targetWorkspaceId`.

- `update_workspace_restrictions`: POST `/v0/meta/workspaces/{{ record.workspace_id }}/updateRestrictions`; kind `update`; body `json`; path fields `workspace_id`; required `workspace_id`.

- `delete_record`: DELETE `/v0/{{ config.base_id }}/{{ config.table_id }}/{{ record.id }}`; kind `delete`; body `none`; path fields `id`; required `id`.

- `update_record`: PATCH `/v0/{{ config.base_id }}/{{ config.table_id }}/{{ record.id }}`; kind `update`; body `json`; path fields `id`; required `id, fields`.

- `replace_record`: PUT `/v0/{{ config.base_id }}/{{ config.table_id }}/{{ record.id }}`; kind `update`; body `json`; path fields `id`; required `id, fields`.

- `create_comment`: POST `/v0/{{ config.base_id }}/{{ config.table_id }}/{{ record.record_id }}/comments`; kind `create`; body `json`; path fields `record_id`; required `record_id, text`.

- `delete_comment`: DELETE `/v0/{{ config.base_id }}/{{ config.table_id }}/{{ record.record_id }}/comments/{{ record.row_comment_id }}`; kind `delete`; body `none`; path fields `record_id, row_comment_id`; required `record_id, row_comment_id`.

- `update_comment`: PATCH `/v0/{{ config.base_id }}/{{ config.table_id }}/{{ record.record_id }}/comments/{{ record.row_comment_id }}`; kind `update`; body `json`; path fields `record_id, row_comment_id`; required `record_id, row_comment_id, text`.

- `update_date_dependency_metadata`: PATCH `/v0/{{ config.base_id }}/{{ config.table_id }}/{{ record.id }}/dateDependencyMetadata`; kind `update`; body `json`; path fields `id`; required `id, predecessorRecordId, dateDependencyMetadata`.

## Known limits

- Blocked row: `post-sync-api-endpoint` (`POST /v0/{baseId}/{tableIdOrName}/sync/{apiEndpointSyncId}`) because the official Sync API import is `text/csv`; the shared declarative write runtime has no typed bounded CSV body dialect and generic raw uploads remain disallowed.

- Blocked foundation: `airtable-scim-pagination-foundation` for SCIM list users/groups. Airtable's documented SCIM list response uses `Resources`, `startIndex`, `itemsPerPage`, and `totalResults`; the current declarative cursor paginator cannot compute the next `startIndex` arithmetically and must not use nonexistent `nextStartIndex`.

- Blocked foundation: `airtable-array-cardinality-foundation` for 25 official mutations that require non-empty request arrays. The current schema subset supports `minProperties` but not `minItems`, so these operations stay blocked instead of accepting `[]` no-op/destructive payloads.

- Blocked foundation: `airtable-field-variant-schema-foundation` for `create_field` field types whose official write variants require or accept type-specific `options`. The current schema subset has no conditional `anyOf`/`const` enforcement, so `create_field` only advertises official no-options request variants.

- Blocked foundation: `airtable-required-query-foundation` for enterprise user search. Airtable requires at least one documented `email[]` or `id[]` query value, so the connector does not expose an unfiltered ETL stream or raw query escape.

- Blocked foundation: `airtable-bounded-base64-upload-foundation` for attachment upload. The current Airtable-owned definition path cannot validate base64 decoding and the official decoded-size limit before transmission, so the unbounded write action is not executable.

- API-surface ledger rows: 103 total = 28 streams + 44 writes + 1 direct read + 30 blocked operations.

- Certification remains fixture-only/uncertified until separately approved live-safe credentials and provider resources are supplied outside this wave.
