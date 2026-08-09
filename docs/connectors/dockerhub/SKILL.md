---
name: pm-dockerhub
description: Docker Hub connector knowledge and safe action guide.
---

# pm-dockerhub

## Purpose

Reads public Docker Hub repositories and image tags for a configured target namespace; an optional Personal Access Token authenticates as a separately configured Docker Hub username. It also manages access tokens, organizations, groups/teams, invites, audit logs, and repositories via the Docker Hub API.

## Icon

- id: dockerhub
- asset: icons/dockerhub.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://docs.docker.com/docker-hub/api/latest/

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- auth_type
- auth_url
- base_url
- docker_username (required)
- namespace (required)
- page_size
- registry_client_ip
- repository
- tag
- tier
- docker_pat (secret)
- scim_bearer_token (secret)

## ETL Streams

- repositories:
  - primary key: name
  - cursor: last_updated
  - fields: date_registered(string), description(string), is_private(boolean), last_modified(string), last_updated(string), name(string), namespace(string), pull_count(integer), repository_type(string), star_count(integer), status(integer), status_description(string), storage_size(integer)
- tags:
  - primary key: id
  - cursor: last_updated
  - fields: content_type(string), digest(string), full_size(integer), id(integer), last_pushed(string), last_updated(string), last_updater_username(string), media_type(string), name(string), repository(integer), tag_status(string)
- repository_detail:
  - primary key: name
  - fields: collaborator_count(integer), date_registered(string), description(string), full_description(string), has_starred(boolean), hub_user(string), is_automated(boolean), is_private(boolean), last_updated(string), name(string), namespace(string), pull_count(integer), repository_type(string), star_count(integer), status(integer), status_description(string), storage_size(integer)
- tag_detail:
  - primary key: id
  - fields: creator(integer), full_size(integer), id(integer), last_updated(string), last_updater(integer), last_updater_username(string), name(string), repository(integer), status(string), tag_last_pulled(string), tag_last_pushed(string), v2(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- create_repository:
  - endpoint: POST /namespaces/{{ record.namespace }}/repositories
  - required fields: namespace, name
  - risk: medium: mutates Docker Hub account/organization state; requires reverse ETL approval
- update_repository_immutable_tags:
  - endpoint: PATCH /namespaces/{{ record.namespace }}/repositories/{{ record.repository }}/immutabletags
  - required fields: namespace, repository, immutable_tags, immutable_tags_rules
  - risk: medium: mutates Docker Hub account/organization state; requires reverse ETL approval
- assign_repository_group:
  - endpoint: POST /repositories/{{ record.namespace }}/{{ record.repository }}/groups
  - required fields: namespace, repository, group_id, permission
  - risk: medium: mutates Docker Hub account/organization state; requires reverse ETL approval
- create_access_token:
  - endpoint: POST /access-tokens
  - required fields: token_label, scopes
  - risk: high: creates a live Docker Hub credential; provider response is returned unchanged and can include a raw token; requires reverse ETL approval
- update_access_token:
  - endpoint: PATCH /access-tokens/{{ record.uuid }}
  - required fields: uuid
  - risk: medium: mutates Docker Hub account/organization state; requires reverse ETL approval
- delete_access_token:
  - endpoint: DELETE /access-tokens/{{ record.uuid }}
  - required fields: uuid
  - risk: high: revokes a live Docker Hub credential; requires reverse ETL approval and destructive confirmation
- create_org_access_token:
  - endpoint: POST /orgs/{{ record.name }}/access-tokens
  - required fields: name, label
  - risk: high: creates a live Docker Hub organization credential; provider response is returned unchanged and can include a raw token; requires reverse ETL approval
- update_org_access_token:
  - endpoint: PATCH /orgs/{{ record.org_name }}/access-tokens/{{ record.access_token_id }}
  - required fields: org_name, access_token_id
  - risk: medium: mutates Docker Hub account/organization state; requires reverse ETL approval
- delete_org_access_token:
  - endpoint: DELETE /orgs/{{ record.org_name }}/access-tokens/{{ record.access_token_id }}
  - required fields: org_name, access_token_id
  - risk: high: revokes a live Docker Hub organization credential; requires reverse ETL approval and destructive confirmation
- create_group:
  - endpoint: POST /orgs/{{ record.org_name }}/groups
  - required fields: org_name, name
  - risk: medium: mutates Docker Hub account/organization state; requires reverse ETL approval
- replace_group:
  - endpoint: PUT /orgs/{{ record.org_name }}/groups/{{ record.group_name }}
  - required fields: org_name, group_name, name
  - risk: medium: mutates Docker Hub account/organization state; requires reverse ETL approval
- update_group:
  - endpoint: PATCH /orgs/{{ record.org_name }}/groups/{{ record.group_name }}
  - required fields: org_name, group_name
  - risk: medium: mutates Docker Hub account/organization state; requires reverse ETL approval
- delete_group:
  - endpoint: DELETE /orgs/{{ record.org_name }}/groups/{{ record.group_name }}
  - required fields: org_name, group_name
  - risk: high: removes an organization group and its access grants; requires reverse ETL approval and destructive confirmation
- add_group_member:
  - endpoint: POST /orgs/{{ record.org_name }}/groups/{{ record.group_name }}/members
  - required fields: org_name, group_name, member
  - risk: medium: mutates Docker Hub account/organization state; requires reverse ETL approval
- remove_group_member:
  - endpoint: DELETE /orgs/{{ record.org_name }}/groups/{{ record.group_name }}/members/{{ record.username }}
  - required fields: org_name, group_name, username
  - risk: high: removes a user's access via this group; requires reverse ETL approval and destructive confirmation
- bulk_create_invites:
  - endpoint: POST /invites/bulk
  - required fields: org, invitees
  - risk: medium: mutates Docker Hub account/organization state; requires reverse ETL approval
- cancel_invite:
  - endpoint: DELETE /invites/{{ record.id }}
  - required fields: id
  - risk: medium: cancels a pending organization invite; requires reverse ETL approval and destructive confirmation
- resend_invite:
  - endpoint: PATCH /invites/{{ record.id }}/resend
  - required fields: id
  - risk: medium: mutates Docker Hub account/organization state; requires reverse ETL approval
- update_org_settings:
  - endpoint: PUT /orgs/{{ record.name }}/settings
  - required fields: name, restricted_images_enabled, restricted_images_allow_official, restricted_images_allow_verified_publishers
  - risk: medium: mutates Docker Hub account/organization state; requires reverse ETL approval
- update_org_member:
  - endpoint: PUT /orgs/{{ record.org_name }}/members/{{ record.username }}
  - required fields: org_name, username, role
  - risk: high: changes a member's organization role/permissions; requires reverse ETL approval
- remove_org_member:
  - endpoint: DELETE /orgs/{{ record.org_name }}/members/{{ record.username }}
  - required fields: org_name, username
  - risk: high: removes a user's organization membership; requires reverse ETL approval and destructive confirmation
- create_scim_user:
  - endpoint: POST /scim/2.0/Users
  - required fields: schemas, userName
  - risk: high: provisions a new SCIM-managed identity; requires reverse ETL approval
- update_scim_user:
  - endpoint: PUT /scim/2.0/Users/{{ record.id }}
  - required fields: id, schemas, enabled
  - risk: high: replaces a SCIM-managed identity's profile and can deactivate it; requires reverse ETL approval

## Security

- read risk: external Docker Hub API read of public repository/tag data, and, when authenticated, account-scoped access-token/org/group/invite/audit-log metadata
- approval: reverse ETL mutations (repository, access-token, group, invite, org member/settings writes) require plan, preview, explicit approval, and execute; destructive actions require typed destructive confirmation
- For auth credential fields, prefer --from-env or --value-stdin over command-line flags; argv can be observed by other local processes and shell history.
- The approved runtime returns Docker Hub response bodies unchanged, including a newly-created token; handle runtime output as secret material.

## Command Surface

- Run Docker Hub's declared streams, direct reads, and reverse-ETL actions.
- Usage: pm dockerhub <command> [flags]
- Repositories & tags
- Personal access tokens
- Organization access tokens
- Audit logs
- Authentication
- Groups (teams)
- Invites
- Organizations & settings
- SCIM
- Other Commands
  - access-tokens create - Create personal access token [intent=reverse_etl availability=implemented write=create_access_token]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: high: creates a live Docker Hub credential; provider response is returned unchanged and can include a raw token; requires reverse ETL approval; notes: The provider can return the raw token value once in the create response. The approved runtime returns that response unchanged; handle its output as secret material.; flags: --token-label (required), --scopes (required), --expires-at
  - access-tokens delete - Delete personal access token [intent=reverse_etl availability=implemented write=delete_access_token]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: high: revokes a live Docker Hub credential; requires reverse ETL approval and destructive confirmation; flags: --uuid (required)
  - access-tokens get - Get personal access token [intent=direct_read availability=implemented operation=dockerhub.get_access_token]; flags: --uuid (required), --page, --page-cursor
  - access-tokens list - List personal access tokens [intent=direct_read availability=implemented operation=dockerhub.list_access_tokens]; flags: --page, --page-cursor
  - access-tokens update - Update personal access token [intent=reverse_etl availability=implemented write=update_access_token]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium: mutates Docker Hub account/organization state; requires reverse ETL approval; notes: The provider can echo token metadata, including a raw value, back on update. The approved runtime returns that response unchanged; handle its output as secret material.; flags: --uuid (required), --token-label, --is-active
  - audit-logs actions list - List audit log actions [intent=direct_read availability=implemented operation=dockerhub.list_audit_log_actions]; flags: --account (required), --page, --page-cursor
  - audit-logs list - List audit log events [intent=direct_read availability=implemented operation=dockerhub.list_audit_logs]; flags: --account (required), --page, --page-cursor
  - auth 2fa-login create - Second factor authentication (completes a 2FA-challenged login) [intent=direct_write availability=implemented operation=dockerhub.create_2fa_login]; approval: Requires plan, preview, explicit approval, and execute.; risk: high: exchanges a live intermediate login token and second factor code for a session token; notes: The provider response is returned unchanged by the approved runtime output.; flags: --login-2fa-token (required), --code (required)
  - auth login create - Create an authentication token (username+password/PAT session login) [intent=direct_write availability=implemented operation=dockerhub.create_login]; approval: Requires plan, preview, explicit approval, and execute.; risk: high: exchanges a live username and password for a session token; notes: The provider response is returned unchanged by the approved runtime output.; flags: --username (required), --password (required)
  - auth token create - Create access token (identifier+secret to short-lived access token exchange) [intent=direct_write availability=implemented operation=dockerhub.create_auth_token]; approval: Requires plan, preview, explicit approval, and execute.; risk: high: exchanges an identifier and secret for an access token; notes: The provider response is returned unchanged by the approved runtime output.; flags: --identifier (required), --secret (required)
  - groups create - Create a new group [intent=reverse_etl availability=implemented write=create_group]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium: mutates Docker Hub account/organization state; requires reverse ETL approval; flags: --org-name (required), --name (required), --description
  - groups delete - Delete an organization group [intent=reverse_etl availability=implemented write=delete_group]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: high: removes an organization group and its access grants; requires reverse ETL approval and destructive confirmation; flags: --org-name (required), --group-name (required)
  - groups get - Get a group of an organization [intent=direct_read availability=implemented operation=dockerhub.get_group]; flags: --org-name (required), --group-name (required), --page, --page-cursor
  - groups list - Get groups of an organization [intent=direct_read availability=implemented operation=dockerhub.list_groups]; flags: --org-name (required), --page, --page-cursor
  - groups members add - Add a member to a group [intent=reverse_etl availability=implemented write=add_group_member]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium: mutates Docker Hub account/organization state; requires reverse ETL approval; flags: --org-name (required), --group-name (required), --member (required)
  - groups members list - List members of a group [intent=direct_read availability=implemented operation=dockerhub.list_group_members]; flags: --org-name (required), --group-name (required), --page, --page-cursor
  - groups members remove - Remove a user from a group [intent=reverse_etl availability=implemented write=remove_group_member]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: high: removes a user's access via this group; requires reverse ETL approval and destructive confirmation; flags: --org-name (required), --group-name (required), --username (required)
  - groups replace - Update the details for an organization group (full replace) [intent=reverse_etl availability=implemented write=replace_group]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium: mutates Docker Hub account/organization state; requires reverse ETL approval; flags: --org-name (required), --group-name (required), --name (required), --description
  - groups update - Update some details for an organization group (partial update) [intent=reverse_etl availability=implemented write=update_group]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium: mutates Docker Hub account/organization state; requires reverse ETL approval; flags: --org-name (required), --group-name (required), --name, --description
  - invites bulk-create - Bulk create invites [intent=reverse_etl availability=implemented write=bulk_create_invites]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium: mutates Docker Hub account/organization state; requires reverse ETL approval; flags: --org (required), --team, --role, --invitees (required), --dry-run
  - invites cancel - Cancel an invite [intent=reverse_etl availability=implemented write=cancel_invite]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium: cancels a pending organization invite; requires reverse ETL approval and destructive confirmation; flags: --id (required)
  - invites list - List org invites [intent=direct_read availability=implemented operation=dockerhub.list_org_invites]; flags: --org-name (required), --page, --page-cursor
  - invites resend - Resend an invite [intent=reverse_etl availability=implemented write=resend_invite]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium: mutates Docker Hub account/organization state; requires reverse ETL approval; flags: --id (required)
  - org access-tokens create - Create organization access token [intent=reverse_etl availability=implemented write=create_org_access_token]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: high: creates a live Docker Hub organization credential; provider response is returned unchanged and can include a raw token; requires reverse ETL approval; notes: `resources` (the token's repo/org scope grants) is a required-for-meaningful-use array of objects with no typed scalar leaf, so it is not flag-mapped here; supply it via a reverse-ETL source record instead. The provider can return the raw token value once in the create response; the approved runtime returns that response unchanged, so handle its output as secret material.; flags: --name (required), --label (required), --description, --expires-at
  - org access-tokens delete - Delete organization access token [intent=reverse_etl availability=implemented write=delete_org_access_token]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: high: revokes a live Docker Hub organization credential; requires reverse ETL approval and destructive confirmation; flags: --org-name (required), --access-token-id (required)
  - org access-tokens get - Get organization access token [intent=direct_read availability=implemented operation=dockerhub.get_org_access_token]; flags: --org-name (required), --access-token-id (required), --page, --page-cursor
  - org access-tokens list - List organization access tokens [intent=direct_read availability=implemented operation=dockerhub.list_org_access_tokens]; flags: --name (required), --page, --page-cursor
  - org access-tokens update - Update organization access token [intent=reverse_etl availability=implemented write=update_org_access_token]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium: mutates Docker Hub account/organization state; requires reverse ETL approval; notes: `resources` is a required-for-meaningful-use array of objects with no typed scalar leaf, so it is not flag-mapped here; supply it via a reverse-ETL source record instead.; flags: --org-name (required), --access-token-id (required), --label, --description, --is-active
  - org members export - Export org members CSV [intent=binary_download availability=implemented operation=dockerhub.export_org_members]; flags: --org-name (required), --dest-root (required), --file-name, --max-bytes
  - org members list - List org members [intent=direct_read availability=implemented operation=dockerhub.list_org_members]; flags: --org-name (required), --page, --page-cursor
  - org members remove - Remove member from org [intent=reverse_etl availability=implemented write=remove_org_member]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: high: removes a user's organization membership; requires reverse ETL approval and destructive confirmation; flags: --org-name (required), --username (required)
  - org members update - Update org member (role) [intent=reverse_etl availability=implemented write=update_org_member]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: high: changes a member's organization role/permissions; requires reverse ETL approval; flags: --org-name (required), --username (required), --role (required)
  - org settings get - Get organization settings [intent=direct_read availability=implemented operation=dockerhub.get_org_settings]; flags: --name (required), --page, --page-cursor
  - org settings update - Update organization settings [intent=reverse_etl availability=implemented write=update_org_settings]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium: mutates Docker Hub account/organization state; requires reverse ETL approval; notes: The provider's restricted_images body is a nested object ({enabled, allow_official_images, allow_verified_publishers}); this command flattens it into three top-level scalar flags mapped by the write executor's record, matching the object's own three required booleans one-for-one.; flags: --name (required), --restricted-images-enabled (required), --restricted-images-allow-official (required), --restricted-images-allow-verified-publishers (required)
  - repositories list - Run the repositories ETL stream [intent=etl availability=implemented stream=repositories]
  - repository check - Check repository in a namespace [intent=direct_read availability=implemented operation=dockerhub.check_repository]; flags: --namespace (required), --repository (required), --page, --page-cursor
  - repository create - Create a new repository [intent=reverse_etl availability=implemented write=create_repository]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium: mutates Docker Hub account/organization state; requires reverse ETL approval; flags: --name (required), --namespace (required), --description, --full-description, --registry, --is-private
  - repository detail list - Run the repository detail ETL stream [intent=etl availability=implemented stream=repository_detail]
  - repository group assign - Assign a group (Team) to a repository for access [intent=reverse_etl availability=implemented write=assign_repository_group]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium: mutates Docker Hub account/organization state; requires reverse ETL approval; flags: --namespace (required), --repository (required), --group-id (required), --permission (required)
  - repository immutable-tags update - Update repository immutable tags [intent=reverse_etl availability=implemented write=update_repository_immutable_tags]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium: mutates Docker Hub account/organization state; requires reverse ETL approval; flags: --namespace (required), --repository (required), --immutable-tags (required), --immutable-tags-rules (required)
  - repository immutable-tags verify - Verify repository immutable tags [intent=direct_read availability=implemented operation=dockerhub.verify_repository_immutable_tags]; flags: --namespace (required), --repository (required), --regex (required), --page, --page-cursor
  - repository tag check - Check repository tag [intent=direct_read availability=implemented operation=dockerhub.check_repository_tag]; flags: --namespace (required), --repository (required), --tag (required), --page, --page-cursor
  - repository tags check - Check repository tags [intent=direct_read availability=implemented operation=dockerhub.check_repository_tags]; flags: --namespace (required), --repository (required), --page, --page-cursor
  - scim-resource-types get - Get a resource type [intent=direct_read availability=implemented operation=dockerhub.get_scim_resource_type]; flags: --name (required), --page, --page-cursor
  - scim-resource-types list - List resource types [intent=direct_read availability=implemented operation=dockerhub.list_scim_resource_types]; flags: --page, --page-cursor
  - scim-schemas get - Get a schema [intent=direct_read availability=implemented operation=dockerhub.get_scim_schema]; flags: --id (required), --page, --page-cursor
  - scim-schemas list - List schemas [intent=direct_read availability=implemented operation=dockerhub.list_scim_schemas]; flags: --page, --page-cursor
  - scim-service-provider-config get - Get service provider config [intent=direct_read availability=implemented operation=dockerhub.get_scim_service_provider_config]; flags: --page, --page-cursor
  - scim-users create - Create user [intent=reverse_etl availability=implemented write=create_scim_user]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: high: provisions a new SCIM-managed identity; requires reverse ETL approval; notes: The provider documents this endpoint's request/response media type as application/scim+json (RFC 7644); the declared write sends that media type.; flags: --schemas (required), --user-name (required), --given-name, --family-name
  - scim-users get - Get a user [intent=direct_read availability=implemented operation=dockerhub.get_scim_user]; flags: --id (required), --page, --page-cursor
  - scim-users list - List users [intent=direct_read availability=implemented operation=dockerhub.list_scim_users]; flags: --page, --page-cursor
  - scim-users update - Update a user (full replace; enabled is required here even though the provider defaults a missing value to false-deactivated, to prevent an accidental deactivation-by-omission) [intent=reverse_etl availability=implemented write=update_scim_user]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: high: replaces a SCIM-managed identity's profile and can deactivate it; requires reverse ETL approval; notes: The provider documents this endpoint's request/response media type as application/scim+json (RFC 7644); the declared write sends that media type.; flags: --id (required), --schemas (required), --given-name, --family-name, --enabled (required)
  - tag detail list - Run the tag detail ETL stream [intent=etl availability=implemented stream=tag_detail]
  - tags list - Run the tags ETL stream [intent=etl availability=implemented stream=tags]

## Commands

### Inspect as a manual

```bash
pm connectors inspect dockerhub
```

### Inspect as structured JSON

```bash
pm connectors inspect dockerhub --json
```

## Agent Rules

- Run pm connectors inspect dockerhub before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
