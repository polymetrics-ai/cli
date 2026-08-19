---
name: pm-dockerhub
description: Docker Hub connector knowledge and safe action guide.
---

# pm-dockerhub

## Purpose

Reads public Docker Hub repositories and image tags, and performs source-declared organization mutations through the approved reverse-ETL write lifecycle.

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

- base_url
- docker_username (required)
- page_size
- repository
- tag
- access_token (secret)
- code (secret)
- login_2fa_token (secret)
- password (secret)
- secret (secret)

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

- update_organization_settings:
  - endpoint: PUT /v2/orgs/{{ record.name }}/settings
  - required fields: name, restricted_images
  - risk: high
- update_repository_immutable_tags:
  - endpoint: PATCH /v2/namespaces/{{ record.namespace }}/repositories/{{ record.repository }}/immutabletags
  - required fields: namespace, repository, immutable_tags, immutable_tags_rules
  - risk: high
- verify_repository_immutable_tags:
  - endpoint: POST /v2/namespaces/{{ record.namespace }}/repositories/{{ record.repository }}/immutabletags/verify
  - required fields: namespace, repository, regex
  - risk: high
- create_repository_group:
  - endpoint: POST /v2/repositories/{{ record.namespace }}/{{ record.repository }}/groups
  - required fields: namespace, repository, group_id, permission
  - risk: high
- create_repository:
  - endpoint: POST /v2/namespaces/{{ record.namespace }}/repositories
  - required fields: namespace, name
  - optional fields: description, full_description, registry, is_private
  - risk: high
- update_organization_member:
  - endpoint: PUT /v2/orgs/{{ record.org_name }}/members/{{ record.username }}
  - required fields: org_name, username, role
  - risk: high
- delete_organization_member:
  - endpoint: DELETE /v2/orgs/{{ record.org_name }}/members/{{ record.username }}
  - required fields: org_name, username
  - risk: high
- create_organization_group:
  - endpoint: POST /v2/orgs/{{ record.org_name }}/groups
  - required fields: org_name, name
  - optional fields: description
  - risk: high
- replace_organization_group:
  - endpoint: PUT /v2/orgs/{{ record.org_name }}/groups/{{ record.group_name }}
  - required fields: org_name, group_name, name
  - optional fields: description
  - risk: high
- update_organization_group:
  - endpoint: PATCH /v2/orgs/{{ record.org_name }}/groups/{{ record.group_name }}
  - required fields: org_name, group_name
  - optional fields: name, description
  - risk: high
- delete_organization_group:
  - endpoint: DELETE /v2/orgs/{{ record.org_name }}/groups/{{ record.group_name }}
  - required fields: org_name, group_name
  - risk: high
- add_organization_group_member:
  - endpoint: POST /v2/orgs/{{ record.org_name }}/groups/{{ record.group_name }}/members
  - required fields: org_name, group_name, member
  - risk: high
- delete_organization_group_member:
  - endpoint: DELETE /v2/orgs/{{ record.org_name }}/groups/{{ record.group_name }}/members/{{ record.username }}
  - required fields: org_name, group_name, username
  - risk: high
- delete_organization_invite:
  - endpoint: DELETE /v2/invites/{{ record.id }}
  - required fields: id
  - risk: high
- resend_organization_invite:
  - endpoint: PATCH /v2/invites/{{ record.id }}/resend
  - required fields: id
  - risk: high
- create_organization_invites:
  - endpoint: POST /v2/invites/bulk
  - required fields: org, invitees
  - optional fields: team, role, dry_run
  - risk: high
- update_personal_access_token:
  - endpoint: PATCH /v2/access-tokens/{{ record.uuid }}
  - required fields: uuid
  - optional fields: token_label, is_active
  - risk: high
- delete_personal_access_token:
  - endpoint: DELETE /v2/access-tokens/{{ record.uuid }}
  - required fields: uuid
  - risk: high
- update_organization_access_token:
  - endpoint: PATCH /v2/orgs/{{ record.org_name }}/access-tokens/{{ record.access_token_id }}
  - required fields: org_name, access_token_id
  - optional fields: label, description, resources, is_active
  - risk: high
- delete_organization_access_token:
  - endpoint: DELETE /v2/orgs/{{ record.org_name }}/access-tokens/{{ record.access_token_id }}
  - required fields: org_name, access_token_id
  - risk: high

## Security

- read risk: external Docker Hub API reads, including public repository and authenticated organization resources
- write risk: source-declared Docker Hub organization mutations require reverse-ETL plan, preview, approval, and execute
- approval: reverse-ETL writes require plan, preview, approval, and execute
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Command Surface

- Run Docker Hub's declared streams and reverse-ETL actions.
- Usage: pm dockerhub <command> [flags]
- Global flags:
  - --approval-token-stdin (boolean): Read the approval token as one bounded line from standard input.
- Read streams
- Other Commands
  - repositories list - Run the repositories ETL stream [intent=etl availability=implemented stream=repositories]
  - repository detail list - Run the repository detail ETL stream [intent=etl availability=implemented stream=repository_detail]
  - tag detail list - Run the tag detail ETL stream [intent=etl availability=implemented stream=tag_detail]
  - tags list - Run the tags ETL stream [intent=etl availability=implemented stream=tags]
  - audit log actions view - List audit log actions [intent=direct_read availability=implemented operation=dockerhub.auditlogs_listauditactions]; approval: none; risk: Docker Hub provider authorization is enforced at runtime.; flags: --account (required), --page, --page-cursor
  - organization settings view - Get organization settings [intent=direct_read availability=implemented operation=dockerhub.get__v2_orgs__name__settings]; approval: none; risk: Docker Hub provider authorization is enforced at runtime.; flags: --name (required), --page, --page-cursor
  - organization members list - List org members [intent=direct_read availability=implemented operation=dockerhub.get__v2_orgs__org_name__members]; approval: none; risk: Docker Hub provider authorization is enforced at runtime.; flags: --org-name (required), --search, --invites, --type, --role, --page, --page-cursor
  - organization invites list - List org invites [intent=direct_read availability=implemented operation=dockerhub.get__v2_orgs__org_name__invites]; approval: none; risk: Docker Hub provider authorization is enforced at runtime.; flags: --org-name (required), --page, --page-cursor
  - organization groups list - Get groups of an organization [intent=direct_read availability=implemented operation=dockerhub.get__v2_orgs__org_name__groups]; approval: none; risk: Docker Hub provider authorization is enforced at runtime.; flags: --org-name (required), --username, --search, --page, --page-cursor
  - organization group view - Get a group of an organization [intent=direct_read availability=implemented operation=dockerhub.get__v2_orgs__org_name__groups__group_name_]; approval: none; risk: Docker Hub provider authorization is enforced at runtime.; flags: --org-name (required), --group-name (required), --page, --page-cursor
  - organization group members list - List members of a group [intent=direct_read availability=implemented operation=dockerhub.get__v2_orgs__org_name__groups__group_name__members]; approval: none; risk: Docker Hub provider authorization is enforced at runtime.; flags: --org-name (required), --group-name (required), --search, --page, --page-cursor
  - scim service-provider-config view - Get service provider config [intent=direct_read availability=implemented operation=dockerhub.get__v2_scim_2.0_serviceproviderconfig]; approval: none; risk: Docker Hub provider authorization is enforced at runtime.; flags: --page, --page-cursor
  - scim resource-types list - List resource types [intent=direct_read availability=implemented operation=dockerhub.get__v2_scim_2.0_resourcetypes]; approval: none; risk: Docker Hub provider authorization is enforced at runtime.; flags: --page, --page-cursor
  - scim resource-type view - Get a resource type [intent=direct_read availability=implemented operation=dockerhub.get__v2_scim_2.0_resourcetypes__name_]; approval: none; risk: Docker Hub provider authorization is enforced at runtime.; flags: --name (required), --page, --page-cursor
  - scim schemas list - List schemas [intent=direct_read availability=implemented operation=dockerhub.get__v2_scim_2.0_schemas]; approval: none; risk: Docker Hub provider authorization is enforced at runtime.; flags: --page, --page-cursor
  - scim schema view - Get a schema [intent=direct_read availability=implemented operation=dockerhub.get__v2_scim_2.0_schemas__id_]; approval: none; risk: Docker Hub provider authorization is enforced at runtime.; flags: --id (required), --page, --page-cursor
  - scim user view - Get a user [intent=direct_read availability=implemented operation=dockerhub.get__v2_scim_2.0_users__id_]; approval: none; risk: Docker Hub provider authorization is enforced at runtime.; flags: --id (required), --page, --page-cursor
  - organization settings update - Update organization settings [intent=reverse_etl availability=implemented write=update_organization_settings]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: Docker Hub mutation; reverse ETL requires plan, preview, approval, and execute.; flags: --restricted-images (required), --name (required)
  - repository immutable-tags update - Update repository immutable tags [intent=reverse_etl availability=implemented write=update_repository_immutable_tags]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: Docker Hub mutation; reverse ETL requires plan, preview, approval, and execute.; flags: --immutable-tags (required), --immutable-tags-rules (required), --namespace (required), --repository (required)
  - repository immutable-tags verify - Verify repository immutable tags [intent=reverse_etl availability=implemented write=verify_repository_immutable_tags]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: Docker Hub mutation; reverse ETL requires plan, preview, approval, and execute.; flags: --regex (required), --namespace (required), --repository (required)
  - repository group create - Assign a group (Team) to a repository for access [intent=reverse_etl availability=implemented write=create_repository_group]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: Docker Hub mutation; reverse ETL requires plan, preview, approval, and execute.; flags: --group-id (required), --permission (required), --namespace (required), --repository (required)
  - repository create - Create a new repository [intent=reverse_etl availability=implemented write=create_repository]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: Docker Hub mutation; reverse ETL requires plan, preview, approval, and execute.; flags: --name (required), --namespace (required), --description, --full-description, --registry, --is-private
  - organization member update - Update org member (role) [intent=reverse_etl availability=implemented write=update_organization_member]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: Docker Hub mutation; reverse ETL requires plan, preview, approval, and execute.; flags: --role (required), --org-name (required), --username (required)
  - organization member delete - Remove member from org [intent=reverse_etl availability=implemented write=delete_organization_member]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: Docker Hub mutation; reverse ETL requires plan, preview, approval, and execute.; flags: --org-name (required), --username (required)
  - organization group create - Create a new group [intent=reverse_etl availability=implemented write=create_organization_group]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: Docker Hub mutation; reverse ETL requires plan, preview, approval, and execute.; flags: --name (required), --description, --org-name (required)
  - organization group replace - Update the details for an organization group [intent=reverse_etl availability=implemented write=replace_organization_group]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: Docker Hub mutation; reverse ETL requires plan, preview, approval, and execute.; flags: --name (required), --description, --org-name (required), --group-name (required)
  - organization group update - Update some details for an organization group [intent=reverse_etl availability=implemented write=update_organization_group]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: Docker Hub mutation; reverse ETL requires plan, preview, approval, and execute.; flags: --name, --description, --org-name (required), --group-name (required)
  - organization group delete - Delete an organization group [intent=reverse_etl availability=implemented write=delete_organization_group]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: Docker Hub mutation; reverse ETL requires plan, preview, approval, and execute.; flags: --org-name (required), --group-name (required)
  - organization group member add - Add a member to a group [intent=reverse_etl availability=implemented write=add_organization_group_member]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: Docker Hub mutation; reverse ETL requires plan, preview, approval, and execute.; flags: --member (required), --org-name (required), --group-name (required)
  - organization group member delete - Remove a user from a group [intent=reverse_etl availability=implemented write=delete_organization_group_member]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: Docker Hub mutation; reverse ETL requires plan, preview, approval, and execute.; flags: --org-name (required), --group-name (required), --username (required)
  - organization invite delete - Cancel an invite [intent=reverse_etl availability=implemented write=delete_organization_invite]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: Docker Hub mutation; reverse ETL requires plan, preview, approval, and execute.; flags: --id (required)
  - organization invite resend - Resend an invite [intent=reverse_etl availability=implemented write=resend_organization_invite]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: Docker Hub mutation; reverse ETL requires plan, preview, approval, and execute.; flags: --id (required)
  - organization invite bulk-create - Bulk create invites [intent=reverse_etl availability=implemented write=create_organization_invites]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: Docker Hub mutation; reverse ETL requires plan, preview, approval, and execute.; flags: --org (required), --team, --role, --invitees (required), --dry-run
  - personal access tokens list - List personal access tokens [intent=direct_read availability=implemented operation=dockerhub.get__v2_access-tokens]; approval: none; risk: Docker Hub provider authorization is enforced at runtime.; flags: --page, --page-cursor
  - personal access token view - Get personal access token [intent=direct_read availability=implemented operation=dockerhub.get__v2_access-tokens__uuid_]; approval: none; risk: Docker Hub provider authorization is enforced at runtime.; flags: --uuid (required), --page, --page-cursor
  - organization access tokens list - List access tokens [intent=direct_read availability=implemented operation=dockerhub.get__v2_orgs__name__access-tokens]; approval: none; risk: Docker Hub provider authorization is enforced at runtime.; flags: --name (required), --page, --page-cursor
  - organization access token view - Get access token [intent=direct_read availability=implemented operation=dockerhub.get__v2_orgs__org_name__access-tokens__access_token_id_]; approval: none; risk: Docker Hub provider authorization is enforced at runtime.; flags: --org-name (required), --access-token-id (required), --page, --page-cursor
  - personal access token update - Update personal access token [intent=reverse_etl availability=implemented write=update_personal_access_token]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: Docker Hub mutation; reverse ETL requires plan, preview, approval, and execute.; flags: --uuid (required), --token-label, --is-active
  - personal access token delete - Delete personal access token [intent=reverse_etl availability=implemented write=delete_personal_access_token]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: Docker Hub mutation; reverse ETL requires plan, preview, approval, and execute.; flags: --uuid (required)
  - organization access token update - Update access token [intent=reverse_etl availability=implemented write=update_organization_access_token]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: Docker Hub mutation; reverse ETL requires plan, preview, approval, and execute.; flags: --label, --description, --resources, --is-active, --org-name (required), --access-token-id (required)
  - organization access token delete - Delete access token [intent=reverse_etl availability=implemented write=delete_organization_access_token]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: Docker Hub mutation; reverse ETL requires plan, preview, approval, and execute.; flags: --org-name (required), --access-token-id (required)

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
