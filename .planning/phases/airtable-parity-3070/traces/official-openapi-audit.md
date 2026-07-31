# Airtable official OpenAPI re-audit

- Source: https://airtable.com/developers/web/api/introduction
- OpenAPI version: 3.1.0
- Extracted OpenAPI SHA-256: f1506571034500ffb0887e7244f770c0b060a4205e734881bd3de6fa20d6f6b8
- Operation count: 103
- Method counts: {'GET': 31, 'POST': 33, 'DELETE': 19, 'PATCH': 15, 'PUT': 5}
- Lane counts: {'etl_read': 27, 'reverse_etl_write': 69, 'cdc_changefeed': 5, 'binary_file': 1, 'direct_read_query_search': 1}

## Operations

| # | Lane | Method | Path | Operation ID | Tags | Summary |
| ---: | --- | --- | --- | --- | --- | --- |
| 1 | `etl_read` | `GET` | `/scim/v2/Groups` | `list-scim-groups` | scim | List groups |
| 2 | `reverse_etl_write` | `POST` | `/scim/v2/Groups` | `create-scim-group` | scim | Create group |
| 3 | `reverse_etl_write` | `DELETE` | `/scim/v2/Groups/{groupId}` | `delete-scim-group` | scim | Delete group |
| 4 | `etl_read` | `GET` | `/scim/v2/Groups/{groupId}` | `get-scim-group` | scim | Get group |
| 5 | `reverse_etl_write` | `PATCH` | `/scim/v2/Groups/{groupId}` | `patch-scim-group` | scim | Patch group |
| 6 | `reverse_etl_write` | `PUT` | `/scim/v2/Groups/{groupId}` | `put-scim-group` | scim | Put group |
| 7 | `etl_read` | `GET` | `/scim/v2/Users` | `list-scim-users` | scim | List users |
| 8 | `reverse_etl_write` | `POST` | `/scim/v2/Users` | `create-scim-user` | scim | Create user |
| 9 | `reverse_etl_write` | `DELETE` | `/scim/v2/Users/{userId}` | `delete-scim-user` | scim | Delete user |
| 10 | `etl_read` | `GET` | `/scim/v2/Users/{userId}` | `get-scim-user` | scim | Get user |
| 11 | `reverse_etl_write` | `PATCH` | `/scim/v2/Users/{userId}` | `patch-scim-user` | scim | Patch user |
| 12 | `reverse_etl_write` | `PUT` | `/scim/v2/Users/{userId}` | `put-scim-user` | scim | Put user |
| 13 | `etl_read` | `GET` | `/v0/bases/{baseId}/webhooks` | `list-webhooks` | webhooks | List webhooks |
| 14 | `reverse_etl_write` | `POST` | `/v0/bases/{baseId}/webhooks` | `create-a-webhook` | webhooks | Create a webhook |
| 15 | `reverse_etl_write` | `DELETE` | `/v0/bases/{baseId}/webhooks/{webhookId}` | `delete-a-webhook` | webhooks | Delete a webhook |
| 16 | `reverse_etl_write` | `POST` | `/v0/bases/{baseId}/webhooks/{webhookId}/enableNotifications` | `enable-disable-webhook-notifications` | webhooks | Enable/disable webhook notifications |
| 17 | `cdc_changefeed` | `GET` | `/v0/bases/{baseId}/webhooks/{webhookId}/payloads` | `list-webhook-payloads` | webhooks | List webhook payloads |
| 18 | `reverse_etl_write` | `POST` | `/v0/bases/{baseId}/webhooks/{webhookId}/refresh` | `refresh-a-webhook` | webhooks | Refresh a webhook |
| 19 | `etl_read` | `GET` | `/v0/meta/bases` | `list-bases` | base | List bases |
| 20 | `reverse_etl_write` | `POST` | `/v0/meta/bases` | `create-base` | base | Create base |
| 21 | `reverse_etl_write` | `DELETE` | `/v0/meta/bases/{baseId}` | `delete-base` | workspace | Delete base |
| 22 | `etl_read` | `GET` | `/v0/meta/bases/{baseId}` | `get-base-collaborators` | base | Get base collaborators |
| 23 | `etl_read` | `GET` | `/v0/meta/bases/{baseId}/blockInstallations` | `list-block-installations` | block | List block installations |
| 24 | `reverse_etl_write` | `DELETE` | `/v0/meta/bases/{baseId}/blockInstallations/{blockInstallationId}` | `delete-block-installation` | block | Delete block installation |
| 25 | `reverse_etl_write` | `PATCH` | `/v0/meta/bases/{baseId}/blockInstallations/{blockInstallationId}` | `manage-block-installation` | block | Manage block installation |
| 26 | `reverse_etl_write` | `POST` | `/v0/meta/bases/{baseId}/collaborators` | `add-base-collaborator` | collaborators | Add base collaborator |
| 27 | `reverse_etl_write` | `DELETE` | `/v0/meta/bases/{baseId}/collaborators/{userOrGroupId}` | `delete-base-collaborator` | collaborators | Delete base collaborator |
| 28 | `reverse_etl_write` | `PATCH` | `/v0/meta/bases/{baseId}/collaborators/{userOrGroupId}` | `update-collaborator-base-permission` | collaborators | Update collaborator base permission |
| 29 | `etl_read` | `GET` | `/v0/meta/bases/{baseId}/interfaces/{pageBundleId}` | `get-interface` | collaborators | Get interface |
| 30 | `reverse_etl_write` | `POST` | `/v0/meta/bases/{baseId}/interfaces/{pageBundleId}/collaborators` | `add-interface-collaborator` | collaborators | Add interface collaborator |
| 31 | `reverse_etl_write` | `DELETE` | `/v0/meta/bases/{baseId}/interfaces/{pageBundleId}/collaborators/{userOrGroupId}` | `delete-interface-collaborator` | collaborators | Delete interface collaborator |
| 32 | `reverse_etl_write` | `PATCH` | `/v0/meta/bases/{baseId}/interfaces/{pageBundleId}/collaborators/{userOrGroupId}` | `update-interface-collaborator` | collaborators | Update interface collaborator |
| 33 | `reverse_etl_write` | `DELETE` | `/v0/meta/bases/{baseId}/interfaces/{pageBundleId}/invites/{inviteId}` | `delete-interface-invite` | invites | Delete interface invite |
| 34 | `reverse_etl_write` | `DELETE` | `/v0/meta/bases/{baseId}/invites/{inviteId}` | `delete-base-invite` | invites | Delete base invite |
| 35 | `etl_read` | `GET` | `/v0/meta/bases/{baseId}/shares` | `list-shares` | shares | List shares |
| 36 | `reverse_etl_write` | `DELETE` | `/v0/meta/bases/{baseId}/shares/{shareId}` | `delete-share` | shares | Delete share |
| 37 | `reverse_etl_write` | `PATCH` | `/v0/meta/bases/{baseId}/shares/{shareId}` | `manage-share` | shares | Manage share |
| 38 | `etl_read` | `GET` | `/v0/meta/bases/{baseId}/tables` | `get-base-schema` | base | Get base schema |
| 39 | `reverse_etl_write` | `POST` | `/v0/meta/bases/{baseId}/tables` | `create-table` | table | Create table |
| 40 | `reverse_etl_write` | `PATCH` | `/v0/meta/bases/{baseId}/tables/{tableIdOrName}` | `update-table` | table | Update table |
| 41 | `reverse_etl_write` | `POST` | `/v0/meta/bases/{baseId}/tables/{tableId}/fields` | `create-field` | column | Create field |
| 42 | `reverse_etl_write` | `PATCH` | `/v0/meta/bases/{baseId}/tables/{tableId}/fields/{columnId}` | `update-field` | column | Update field |
| 43 | `etl_read` | `GET` | `/v0/meta/bases/{baseId}/views` | `list-views` | view | List views |
| 44 | `reverse_etl_write` | `DELETE` | `/v0/meta/bases/{baseId}/views/{viewId}` | `delete-view` | view | Delete view |
| 45 | `etl_read` | `GET` | `/v0/meta/bases/{baseId}/views/{viewId}` | `get-view-metadata` | view | Get view metadata |
| 46 | `etl_read` | `GET` | `/v0/meta/enterpriseAccounts/{enterpriseAccountId}` | `get-enterprise` | enterprise | Get enterprise |
| 47 | `etl_read` | `GET` | `/v0/meta/enterpriseAccounts/{enterpriseAccountId}/auditLogEvents` | `audit-log-events` | auditLogs | Audit log events |
| 48 | `etl_read` | `GET` | `/v0/meta/enterpriseAccounts/{enterpriseAccountId}/auditLogs` | `list-audit-log-requests` | auditing | List audit log requests |
| 49 | `reverse_etl_write` | `POST` | `/v0/meta/enterpriseAccounts/{enterpriseAccountId}/auditLogs` | `create-audit-log-request` | auditing | Create audit log request |
| 50 | `etl_read` | `GET` | `/v0/meta/enterpriseAccounts/{enterpriseAccountId}/auditLogs/{enterpriseAuditLogTaskId}` | `get-audit-log-request` | auditing | Get audit log request |
| 51 | `cdc_changefeed` | `GET` | `/v0/meta/enterpriseAccounts/{enterpriseAccountId}/changeEvents` | `change-events` | changeEvents | Change events |
| 52 | `reverse_etl_write` | `POST` | `/v0/meta/enterpriseAccounts/{enterpriseAccountId}/descendants` | `create-descendant-enterprise` | enterprise | Create descendant enterprise |
| 53 | `cdc_changefeed` | `GET` | `/v0/meta/enterpriseAccounts/{enterpriseAccountId}/exports` | `list-ediscovery-export` | changeEvents | List eDiscovery exports |
| 54 | `cdc_changefeed` | `POST` | `/v0/meta/enterpriseAccounts/{enterpriseAccountId}/exports` | `create-ediscovery-export` | changeEvents | Create eDiscovery export |
| 55 | `cdc_changefeed` | `GET` | `/v0/meta/enterpriseAccounts/{enterpriseAccountId}/exports/{enterpriseTaskId}` | `get-ediscovery-export` | changeEvents | Get eDiscovery export |
| 56 | `reverse_etl_write` | `POST` | `/v0/meta/enterpriseAccounts/{enterpriseAccountId}/moveGroups` | `move-user-groups` | userManagement | Move user groups |
| 57 | `reverse_etl_write` | `POST` | `/v0/meta/enterpriseAccounts/{enterpriseAccountId}/moveWorkspaces` | `move-workspaces` | workspace | Move workspaces |
| 58 | `etl_read` | `GET` | `/v0/meta/enterpriseAccounts/{enterpriseAccountId}/packages` | `list-enterprise-packages` | enterprise | List packages |
| 59 | `reverse_etl_write` | `POST` | `/v0/meta/enterpriseAccounts/{enterpriseAccountId}/packages/{packageId}/install` | `create-base-from-package-enterprise` | enterprise | Create base from package |
| 60 | `etl_read` | `GET` | `/v0/meta/enterpriseAccounts/{enterpriseAccountId}/personalAccessTokens` | `list-enterprise-personal-access-tokens` | enterprise | List personal access tokens |
| 61 | `reverse_etl_write` | `POST` | `/v0/meta/enterpriseAccounts/{enterpriseAccountId}/personalAccessTokens/revoke` | `revoke-enterprise-personal-access-tokens` | enterprise | Revoke personal access tokens |
| 62 | `reverse_etl_write` | `DELETE` | `/v0/meta/enterpriseAccounts/{enterpriseAccountId}/users` | `delete-users-by-email` | user | Delete users by email |
| 63 | `etl_read` | `GET` | `/v0/meta/enterpriseAccounts/{enterpriseAccountId}/users` | `get-users-by-id-or-email` | user | Get users by id or email |
| 64 | `reverse_etl_write` | `PATCH` | `/v0/meta/enterpriseAccounts/{enterpriseAccountId}/users` | `manage-user-batched` | user | Manage user batched |
| 65 | `reverse_etl_write` | `POST` | `/v0/meta/enterpriseAccounts/{enterpriseAccountId}/users/claim` | `manage-user-membership` | user | Manage user membership |
| 66 | `reverse_etl_write` | `POST` | `/v0/meta/enterpriseAccounts/{enterpriseAccountId}/users/grantAdminAccess` | `grant-admin-access` | user | Grant admin access |
| 67 | `reverse_etl_write` | `POST` | `/v0/meta/enterpriseAccounts/{enterpriseAccountId}/users/revokeAdminAccess` | `revoke-admin-access` | user | Revoke admin access |
| 68 | `reverse_etl_write` | `DELETE` | `/v0/meta/enterpriseAccounts/{enterpriseAccountId}/users/{userId}` | `delete-user-by-id` | user | Delete user by id |
| 69 | `etl_read` | `GET` | `/v0/meta/enterpriseAccounts/{enterpriseAccountId}/users/{userId}` | `get-user-by-id` | user | Get user by id |
| 70 | `reverse_etl_write` | `PATCH` | `/v0/meta/enterpriseAccounts/{enterpriseAccountId}/users/{userId}` | `manage-user` | user | Manage user |
| 71 | `reverse_etl_write` | `POST` | `/v0/meta/enterpriseAccounts/{enterpriseAccountId}/users/{userId}/logout` | `logout-user` | user | Logout user |
| 72 | `reverse_etl_write` | `POST` | `/v0/meta/enterpriseAccounts/{enterpriseAccountId}/users/{userId}/remove` | `remove-user-from-enterprise` | enterprise | Remove user from enterprise |
| 73 | `reverse_etl_write` | `POST` | `/v0/meta/enterpriseAccounts/{enterpriseAccountId}/workspaceAiAllowlist` | `update-workspace-ai-allowlist` | enterprise | Update workspace AI allowlist |
| 74 | `etl_read` | `GET` | `/v0/meta/groups/{groupId}` | `get-user-group` | userManagement | Get user group |
| 75 | `etl_read` | `GET` | `/v0/meta/whoami` | `get-user-id-scopes` | whoAmI | Get user info |
| 76 | `reverse_etl_write` | `POST` | `/v0/meta/workspaces` | `create-workspace` | workspace | Create workspace |
| 77 | `reverse_etl_write` | `DELETE` | `/v0/meta/workspaces/{workspaceId}` | `delete-workspace` | workspace | Delete workspace |
| 78 | `etl_read` | `GET` | `/v0/meta/workspaces/{workspaceId}` | `get-workspace-collaborators` | collaborators | Get workspace collaborators |
| 79 | `reverse_etl_write` | `POST` | `/v0/meta/workspaces/{workspaceId}/collaborators` | `add-workspace-collaborator` | collaborators | Add workspace collaborator |
| 80 | `reverse_etl_write` | `DELETE` | `/v0/meta/workspaces/{workspaceId}/collaborators/{userOrGroupId}` | `delete-workspace-collaborator` | collaborators | Delete workspace collaborator |
| 81 | `reverse_etl_write` | `PATCH` | `/v0/meta/workspaces/{workspaceId}/collaborators/{userOrGroupId}` | `update-workspace-collaborator` | collaborators | Update workspace collaborator |
| 82 | `reverse_etl_write` | `DELETE` | `/v0/meta/workspaces/{workspaceId}/invites/{inviteId}` | `delete-workspace-invite` | invites | Delete workspace invite |
| 83 | `reverse_etl_write` | `POST` | `/v0/meta/workspaces/{workspaceId}/moveBase` | `move-base` | workspace | Move base |
| 84 | `reverse_etl_write` | `POST` | `/v0/meta/workspaces/{workspaceId}/updateRestrictions` | `update-workspace-restrictions` | workspace | Update workspace restrictions |
| 85 | `binary_file` | `POST` | `/v0/{baseId}/{recordId}/{attachmentFieldIdOrName}/uploadAttachment` | `upload-attachment` | record | Upload attachment |
| 86 | `reverse_etl_write` | `DELETE` | `/v0/{baseId}/{tableIdOrName}` | `delete-multiple-records` | record | Delete multiple records |
| 87 | `etl_read` | `GET` | `/v0/{baseId}/{tableIdOrName}` | `list-records` | record | List records |
| 88 | `reverse_etl_write` | `PATCH` | `/v0/{baseId}/{tableIdOrName}` | `update-multiple-records` | record | Update multiple records |
| 89 | `reverse_etl_write` | `POST` | `/v0/{baseId}/{tableIdOrName}` | `create-records` | record | Create records |
| 90 | `reverse_etl_write` | `PUT` | `/v0/{baseId}/{tableIdOrName}` | `update-multiple-records-put` | doNotRender | Update multiple records (put) |
| 91 | `reverse_etl_write` | `POST` | `/v0/{baseId}/{tableIdOrName}/sync/{apiEndpointSyncId}` | `post-sync-api-endpoint` | record | Sync CSV data |
| 92 | `reverse_etl_write` | `DELETE` | `/v0/{baseId}/{tableIdOrName}/{recordId}` | `delete-record` | record | Delete record |
| 93 | `etl_read` | `GET` | `/v0/{baseId}/{tableIdOrName}/{recordId}` | `get-record` | record | Get record |
| 94 | `reverse_etl_write` | `PATCH` | `/v0/{baseId}/{tableIdOrName}/{recordId}` | `update-record` | record | Update record |
| 95 | `reverse_etl_write` | `PUT` | `/v0/{baseId}/{tableIdOrName}/{recordId}` | `update-record-put` | doNotRender | Update record (put) |
| 96 | `etl_read` | `GET` | `/v0/{baseId}/{tableIdOrName}/{recordId}/comments` | `list-comments` | comment | List comments |
| 97 | `reverse_etl_write` | `POST` | `/v0/{baseId}/{tableIdOrName}/{recordId}/comments` | `create-comment` | comment | Create comment |
| 98 | `reverse_etl_write` | `DELETE` | `/v0/{baseId}/{tableIdOrName}/{recordId}/comments/{rowCommentId}` | `delete-comment` | comment | Delete comment |
| 99 | `reverse_etl_write` | `PATCH` | `/v0/{baseId}/{tableIdOrName}/{recordId}/comments/{rowCommentId}` | `update-comment` | comment | Update comment |
| 100 | `reverse_etl_write` | `PATCH` | `/v0/{baseId}/{tableIdOrName}/{recordId}/dateDependencyMetadata` | `update-date-dependency-metadata` | record | Update date dependency metadata |
| 101 | `reverse_etl_write` | `POST` | `/v0/{enterpriseAccountId}/{dataTableId}/deleteRecords` | `hyperdb-delete-records-by-primary-keys` | hyperDB, enterprise | Delete records |
| 102 | `direct_read_query_search` | `POST` | `/v0/{enterpriseAccountId}/{dataTableId}/getRecords` | `hyperdb-table-read-records` | hyperDB, enterprise | Read records |
| 103 | `reverse_etl_write` | `PUT` | `/v0/{enterpriseAccountId}/{dataTableId}/upsertRecords` | `hyperdb-upsert-records-by-primary-keys` | hyperDB, enterprise | Update or insert records |
