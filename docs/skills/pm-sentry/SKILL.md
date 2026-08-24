---
name: pm-sentry
description: Sentry connector knowledge and safe action guide.
---

# pm-sentry

## Purpose

Reads Sentry projects, issues, error events, and releases and exposes source-cited, approval-gated REST mutations through the Sentry REST API.

## Icon

- id: sentry
- asset: icons/sentry.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://docs.sentry.io/api/

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url (required)
- max_pages
- organization
- page_size
- project
- auth_token (secret) (required)

## ETL Streams

- projects:
  - primary key: id
  - cursor: dateCreated
  - fields: dateCreated(string), id(string), isBookmarked(boolean), isPublic(boolean), name(string), platform(string), slug(string), status(string)
- issues:
  - primary key: id
  - cursor: lastSeen
  - fields: count(string), culprit(string), firstSeen(string), id(string), lastSeen(string), level(string), shortId(string), status(string), title(string), type(string), userCount(integer)
- events:
  - primary key: id
  - cursor: dateCreated
  - fields: dateCreated(string), eventID(string), groupID(string), id(string), message(string), platform(string), title(string), type(string)
- releases:
  - primary key: version
  - cursor: dateCreated
  - fields: dateCreated(string), dateReleased(string), ref(string), shortVersion(string), status(string), url(string), version(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- sentry_delete_api_0_organizations_organization_id_or_slug_dashboards_dashboard_id:
  - endpoint: DELETE /api/0/organizations/{{ record.organization_id_or_slug }}/dashboards/{{ record.dashboard_id }}/
  - required fields: dashboard_id, organization_id_or_slug
  - risk: Sentry DELETE on /api/0/organizations/{organization_id_or_slug}/dashboards/{dashboard_id}/; changes Sentry data.
- sentry_delete_api_0_organizations_organization_id_or_slug_detectors_detector_id:
  - endpoint: DELETE /api/0/organizations/{{ record.organization_id_or_slug }}/detectors/{{ record.detector_id }}/
  - required fields: detector_id, organization_id_or_slug
  - risk: Sentry DELETE on /api/0/organizations/{organization_id_or_slug}/detectors/{detector_id}/; changes Sentry data.
- sentry_delete_api_0_organizations_organization_id_or_slug_discover_saved_query_id:
  - endpoint: DELETE /api/0/organizations/{{ record.organization_id_or_slug }}/discover/saved/{{ record.query_id }}/
  - required fields: organization_id_or_slug, query_id
  - risk: Sentry DELETE on /api/0/organizations/{organization_id_or_slug}/discover/saved/{query_id}/; changes Sentry data.
- sentry_delete_api_0_organizations_organization_id_or_slug_external_users_external_user_id:
  - endpoint: DELETE /api/0/organizations/{{ record.organization_id_or_slug }}/external-users/{{ record.external_user_id }}/
  - required fields: external_user_id, organization_id_or_slug
  - risk: Sentry DELETE on /api/0/organizations/{organization_id_or_slug}/external-users/{external_user_id}/; changes Sentry data.
- sentry_delete_api_0_organizations_organization_id_or_slug_forwarding_data_forwarder_id:
  - endpoint: DELETE /api/0/organizations/{{ record.organization_id_or_slug }}/forwarding/{{ record.data_forwarder_id }}/
  - required fields: data_forwarder_id, organization_id_or_slug
  - risk: Sentry DELETE on /api/0/organizations/{organization_id_or_slug}/forwarding/{data_forwarder_id}/; changes Sentry data.
- sentry_delete_api_0_organizations_organization_id_or_slug_integrations_integration_id:
  - endpoint: DELETE /api/0/organizations/{{ record.organization_id_or_slug }}/integrations/{{ record.integration_id }}/
  - required fields: integration_id, organization_id_or_slug
  - risk: Sentry DELETE on /api/0/organizations/{organization_id_or_slug}/integrations/{integration_id}/; changes Sentry data.
- sentry_delete_api_0_organizations_organization_id_or_slug_issues_issue_id:
  - endpoint: DELETE /api/0/organizations/{{ record.organization_id_or_slug }}/issues/{{ record.issue_id }}/
  - required fields: issue_id, organization_id_or_slug
  - risk: Sentry DELETE on /api/0/organizations/{organization_id_or_slug}/issues/{issue_id}/; changes Sentry data.
- sentry_delete_api_0_organizations_organization_id_or_slug_issues_issue_id_integrations_integration_id:
  - endpoint: DELETE /api/0/organizations/{{ record.organization_id_or_slug }}/issues/{{ record.issue_id }}/integrations/{{ record.integration_id }}/
  - required fields: integration_id, issue_id, organization_id_or_slug, externalIssue
  - risk: Sentry DELETE on /api/0/organizations/{organization_id_or_slug}/issues/{issue_id}/integrations/{integration_id}/; changes Sentry data.
- sentry_delete_api_0_organizations_organization_id_or_slug_members_member_id:
  - endpoint: DELETE /api/0/organizations/{{ record.organization_id_or_slug }}/members/{{ record.member_id }}/
  - required fields: member_id, organization_id_or_slug
  - risk: Sentry DELETE on /api/0/organizations/{organization_id_or_slug}/members/{member_id}/; changes Sentry data.
- sentry_delete_api_0_organizations_organization_id_or_slug_members_member_id_teams_team_id_or_slug:
  - endpoint: DELETE /api/0/organizations/{{ record.organization_id_or_slug }}/members/{{ record.member_id }}/teams/{{ record.team_id_or_slug }}/
  - required fields: member_id, organization_id_or_slug, team_id_or_slug
  - risk: Sentry DELETE on /api/0/organizations/{organization_id_or_slug}/members/{member_id}/teams/{team_id_or_slug}/; changes Sentry data.
- sentry_delete_api_0_organizations_organization_id_or_slug_notifications_actions_action_id:
  - endpoint: DELETE /api/0/organizations/{{ record.organization_id_or_slug }}/notifications/actions/{{ record.action_id }}/
  - required fields: action_id, organization_id_or_slug
  - risk: Sentry DELETE on /api/0/organizations/{organization_id_or_slug}/notifications/actions/{action_id}/; changes Sentry data.
- sentry_delete_api_0_organizations_organization_id_or_slug_preprodartifacts_snapshots_snapshot_id:
  - endpoint: DELETE /api/0/organizations/{{ record.organization_id_or_slug }}/preprodartifacts/snapshots/{{ record.snapshot_id }}/
  - required fields: organization_id_or_slug, snapshot_id
  - risk: Sentry DELETE on /api/0/organizations/{organization_id_or_slug}/preprodartifacts/snapshots/{snapshot_id}/; changes Sentry data.
- sentry_delete_api_0_organizations_organization_id_or_slug_releases_version:
  - endpoint: DELETE /api/0/organizations/{{ record.organization_id_or_slug }}/releases/{{ record.version }}/
  - required fields: organization_id_or_slug, version
  - risk: Sentry DELETE on /api/0/organizations/{organization_id_or_slug}/releases/{version}/; changes Sentry data.
- sentry_delete_api_0_organizations_organization_id_or_slug_releases_version_files_file_id:
  - endpoint: DELETE /api/0/organizations/{{ record.organization_id_or_slug }}/releases/{{ record.version }}/files/{{ record.file_id }}/
  - required fields: file_id, organization_id_or_slug, version
  - risk: Sentry DELETE on /api/0/organizations/{organization_id_or_slug}/releases/{version}/files/{file_id}/; changes Sentry data.
- sentry_delete_api_0_organizations_organization_id_or_slug_scim_v2_groups_team_id_or_slug:
  - endpoint: DELETE /api/0/organizations/{{ record.organization_id_or_slug }}/scim/v2/Groups/{{ record.team_id_or_slug }}
  - required fields: organization_id_or_slug, team_id_or_slug
  - risk: Sentry DELETE on /api/0/organizations/{organization_id_or_slug}/scim/v2/Groups/{team_id_or_slug}; changes Sentry data.
- sentry_delete_api_0_organizations_organization_id_or_slug_scim_v2_users_member_id:
  - endpoint: DELETE /api/0/organizations/{{ record.organization_id_or_slug }}/scim/v2/Users/{{ record.member_id }}
  - required fields: member_id, organization_id_or_slug
  - risk: Sentry DELETE on /api/0/organizations/{organization_id_or_slug}/scim/v2/Users/{member_id}; changes Sentry data.
- sentry_delete_api_0_organizations_organization_id_or_slug_workflows_workflow_id:
  - endpoint: DELETE /api/0/organizations/{{ record.organization_id_or_slug }}/workflows/{{ record.workflow_id }}/
  - required fields: organization_id_or_slug, workflow_id
  - risk: Sentry DELETE on /api/0/organizations/{organization_id_or_slug}/workflows/{workflow_id}/; changes Sentry data.
- sentry_delete_api_0_projects_organization_id_or_slug_project_id_or_slug:
  - endpoint: DELETE /api/0/projects/{{ record.organization_id_or_slug }}/{{ record.project_id_or_slug }}/
  - required fields: organization_id_or_slug, project_id_or_slug
  - risk: Sentry DELETE on /api/0/projects/{organization_id_or_slug}/{project_id_or_slug}/; changes Sentry data.
- sentry_delete_api_0_projects_organization_id_or_slug_project_id_or_slug_hooks_hook_id:
  - endpoint: DELETE /api/0/projects/{{ record.organization_id_or_slug }}/{{ record.project_id_or_slug }}/hooks/{{ record.hook_id }}/
  - required fields: hook_id, organization_id_or_slug, project_id_or_slug
  - risk: Sentry DELETE on /api/0/projects/{organization_id_or_slug}/{project_id_or_slug}/hooks/{hook_id}/; changes Sentry data.
- sentry_delete_api_0_projects_organization_id_or_slug_project_id_or_slug_issues:
  - endpoint: DELETE /api/0/projects/{{ record.organization_id_or_slug }}/{{ record.project_id_or_slug }}/issues/
  - required fields: organization_id_or_slug, project_id_or_slug
  - risk: Sentry DELETE on /api/0/projects/{organization_id_or_slug}/{project_id_or_slug}/issues/; changes Sentry data.
- sentry_delete_api_0_projects_organization_id_or_slug_project_id_or_slug_keys_key_id:
  - endpoint: DELETE /api/0/projects/{{ record.organization_id_or_slug }}/{{ record.project_id_or_slug }}/keys/{{ record.key_id }}/
  - required fields: key_id, organization_id_or_slug, project_id_or_slug
  - risk: Sentry DELETE on /api/0/projects/{organization_id_or_slug}/{project_id_or_slug}/keys/{key_id}/; changes Sentry data.
- sentry_delete_api_0_projects_organization_id_or_slug_project_id_or_slug_releases_version_files_file_id:
  - endpoint: DELETE /api/0/projects/{{ record.organization_id_or_slug }}/{{ record.project_id_or_slug }}/releases/{{ record.version }}/files/{{ record.file_id }}/
  - required fields: file_id, organization_id_or_slug, project_id_or_slug, version
  - risk: Sentry DELETE on /api/0/projects/{organization_id_or_slug}/{project_id_or_slug}/releases/{version}/files/{file_id}/; changes Sentry data.
- sentry_delete_api_0_projects_organization_id_or_slug_project_id_or_slug_replays_replay_id:
  - endpoint: DELETE /api/0/projects/{{ record.organization_id_or_slug }}/{{ record.project_id_or_slug }}/replays/{{ record.replay_id }}/
  - required fields: organization_id_or_slug, project_id_or_slug, replay_id
  - risk: Sentry DELETE on /api/0/projects/{organization_id_or_slug}/{project_id_or_slug}/replays/{replay_id}/; changes Sentry data.
- sentry_delete_api_0_projects_organization_id_or_slug_project_id_or_slug_symbol_sources:
  - endpoint: DELETE /api/0/projects/{{ record.organization_id_or_slug }}/{{ record.project_id_or_slug }}/symbol-sources/
  - required fields: organization_id_or_slug, project_id_or_slug, id
  - risk: Sentry DELETE on /api/0/projects/{organization_id_or_slug}/{project_id_or_slug}/symbol-sources/; changes Sentry data.
- sentry_delete_api_0_projects_organization_id_or_slug_project_id_or_slug_teams_team_id_or_slug:
  - endpoint: DELETE /api/0/projects/{{ record.organization_id_or_slug }}/{{ record.project_id_or_slug }}/teams/{{ record.team_id_or_slug }}/
  - required fields: organization_id_or_slug, project_id_or_slug, team_id_or_slug
  - risk: Sentry DELETE on /api/0/projects/{organization_id_or_slug}/{project_id_or_slug}/teams/{team_id_or_slug}/; changes Sentry data.
- sentry_delete_api_0_sentry_app_installations_uuid_external_issues_external_issue_id:
  - endpoint: DELETE /api/0/sentry-app-installations/{{ record.uuid }}/external-issues/{{ record.external_issue_id }}/
  - required fields: external_issue_id, uuid
  - risk: Sentry DELETE on /api/0/sentry-app-installations/{uuid}/external-issues/{external_issue_id}/; changes Sentry data.
- sentry_delete_api_0_sentry_apps_sentry_app_id_or_slug:
  - endpoint: DELETE /api/0/sentry-apps/{{ record.sentry_app_id_or_slug }}/
  - required fields: sentry_app_id_or_slug
  - risk: Sentry DELETE on /api/0/sentry-apps/{sentry_app_id_or_slug}/; changes Sentry data.
- sentry_delete_api_0_teams_organization_id_or_slug_team_id_or_slug:
  - endpoint: DELETE /api/0/teams/{{ record.organization_id_or_slug }}/{{ record.team_id_or_slug }}/
  - required fields: organization_id_or_slug, team_id_or_slug
  - risk: Sentry DELETE on /api/0/teams/{organization_id_or_slug}/{team_id_or_slug}/; changes Sentry data.
- sentry_delete_api_0_teams_organization_id_or_slug_team_id_or_slug_external_teams_external_team_id:
  - endpoint: DELETE /api/0/teams/{{ record.organization_id_or_slug }}/{{ record.team_id_or_slug }}/external-teams/{{ record.external_team_id }}/
  - required fields: external_team_id, organization_id_or_slug, team_id_or_slug
  - risk: Sentry DELETE on /api/0/teams/{organization_id_or_slug}/{team_id_or_slug}/external-teams/{external_team_id}/; changes Sentry data.
- sentry_post_api_0_organizations_organization_id_or_slug_members_member_id_teams_team_id_or_slug:
  - endpoint: POST /api/0/organizations/{{ record.organization_id_or_slug }}/members/{{ record.member_id }}/teams/{{ record.team_id_or_slug }}/
  - required fields: member_id, organization_id_or_slug, team_id_or_slug
  - risk: Sentry POST on /api/0/organizations/{organization_id_or_slug}/members/{member_id}/teams/{team_id_or_slug}/; changes Sentry data.
- sentry_post_api_0_projects_organization_id_or_slug_project_id_or_slug_preprodartifacts_snapshots:
  - endpoint: POST /api/0/projects/{{ record.organization_id_or_slug }}/{{ record.project_id_or_slug }}/preprodartifacts/snapshots/
  - required fields: organization_id_or_slug, project_id_or_slug
  - risk: Sentry POST on /api/0/projects/{organization_id_or_slug}/{project_id_or_slug}/preprodartifacts/snapshots/; changes Sentry data.
- sentry_post_api_0_projects_organization_id_or_slug_project_id_or_slug_teams_team_id_or_slug:
  - endpoint: POST /api/0/projects/{{ record.organization_id_or_slug }}/{{ record.project_id_or_slug }}/teams/{{ record.team_id_or_slug }}/
  - required fields: organization_id_or_slug, project_id_or_slug, team_id_or_slug
  - risk: Sentry POST on /api/0/projects/{organization_id_or_slug}/{project_id_or_slug}/teams/{team_id_or_slug}/; changes Sentry data.

## Security

- read risk: external Sentry API read of project, issue, event, and release data
- write risk: approval-gated Sentry REST mutations; the declared actions change Sentry data
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Command Surface

- Run Sentry's declared streams and reverse-ETL actions.
- Usage: pm sentry <command> [flags]
- PM execution policy pm-request-contract-bounds-v1: each max N bytes qualifier is the effective PM request limit, not a provider schema assertion; path/query values are measured after exact wire encoding and rejected rather than truncated.
- Global flags:
  - --approval-token-stdin (boolean): Read the approval token as one bounded line from standard input.
- Read streams
- Other Commands
  - events list - Run the events ETL stream [intent=etl availability=implemented stream=events]
  - issues list - Run the issues ETL stream [intent=etl availability=implemented stream=issues]
  - projects list - Run the projects ETL stream [intent=etl availability=implemented stream=projects]
  - releases list - Run the releases ETL stream [intent=etl availability=implemented stream=releases]
  - api op-44454c455445202f6170692f302f6f7267616e697a6174696f6e732f7b6f7267616e697a6174696f6e5f69645f6f725f736c75677d2f64617368626f617264732f7b64617368626f6172645f69647d2f - DELETE /api/0/organizations/{organization_id_or_slug}/dashboards/{dashboard_id}/ [intent=reverse_etl availability=implemented write=sentry_delete_api_0_organizations_organization_id_or_slug_dashboards_dashboard_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Sentry DELETE on /api/0/organizations/{organization_id_or_slug}/dashboards/{dashboard_id}/; changes Sentry data.; flags: --dashboard-id (required), --organization-id-or-slug (required, max 32768 bytes)
  - api op-44454c455445202f6170692f302f6f7267616e697a6174696f6e732f7b6f7267616e697a6174696f6e5f69645f6f725f736c75677d2f6465746563746f72732f7b6465746563746f725f69647d2f - DELETE /api/0/organizations/{organization_id_or_slug}/detectors/{detector_id}/ [intent=reverse_etl availability=implemented write=sentry_delete_api_0_organizations_organization_id_or_slug_detectors_detector_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Sentry DELETE on /api/0/organizations/{organization_id_or_slug}/detectors/{detector_id}/; changes Sentry data.; flags: --detector-id (required), --organization-id-or-slug (required, max 32768 bytes)
  - api op-44454c455445202f6170692f302f6f7267616e697a6174696f6e732f7b6f7267616e697a6174696f6e5f69645f6f725f736c75677d2f646973636f7665722f73617665642f7b71756572795f69647d2f - DELETE /api/0/organizations/{organization_id_or_slug}/discover/saved/{query_id}/ [intent=reverse_etl availability=implemented write=sentry_delete_api_0_organizations_organization_id_or_slug_discover_saved_query_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Sentry DELETE on /api/0/organizations/{organization_id_or_slug}/discover/saved/{query_id}/; changes Sentry data.; flags: --organization-id-or-slug (required, max 32768 bytes), --query-id (required)
  - api op-44454c455445202f6170692f302f6f7267616e697a6174696f6e732f7b6f7267616e697a6174696f6e5f69645f6f725f736c75677d2f65787465726e616c2d75736572732f7b65787465726e616c5f757365725f69647d2f - DELETE /api/0/organizations/{organization_id_or_slug}/external-users/{external_user_id}/ [intent=reverse_etl availability=implemented write=sentry_delete_api_0_organizations_organization_id_or_slug_external_users_external_user_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Sentry DELETE on /api/0/organizations/{organization_id_or_slug}/external-users/{external_user_id}/; changes Sentry data.; flags: --external-user-id (required), --organization-id-or-slug (required, max 32768 bytes)
  - api op-44454c455445202f6170692f302f6f7267616e697a6174696f6e732f7b6f7267616e697a6174696f6e5f69645f6f725f736c75677d2f666f7277617264696e672f7b646174615f666f727761726465725f69647d2f - DELETE /api/0/organizations/{organization_id_or_slug}/forwarding/{data_forwarder_id}/ [intent=reverse_etl availability=implemented write=sentry_delete_api_0_organizations_organization_id_or_slug_forwarding_data_forwarder_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Sentry DELETE on /api/0/organizations/{organization_id_or_slug}/forwarding/{data_forwarder_id}/; changes Sentry data.; flags: --data-forwarder-id (required), --organization-id-or-slug (required, max 32768 bytes)
  - api op-44454c455445202f6170692f302f6f7267616e697a6174696f6e732f7b6f7267616e697a6174696f6e5f69645f6f725f736c75677d2f696e746567726174696f6e732f7b696e746567726174696f6e5f69647d2f - DELETE /api/0/organizations/{organization_id_or_slug}/integrations/{integration_id}/ [intent=reverse_etl availability=implemented write=sentry_delete_api_0_organizations_organization_id_or_slug_integrations_integration_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Sentry DELETE on /api/0/organizations/{organization_id_or_slug}/integrations/{integration_id}/; changes Sentry data.; flags: --integration-id (required, max 32768 bytes), --organization-id-or-slug (required, max 32768 bytes)
  - api op-44454c455445202f6170692f302f6f7267616e697a6174696f6e732f7b6f7267616e697a6174696f6e5f69645f6f725f736c75677d2f6973737565732f7b69737375655f69647d2f - DELETE /api/0/organizations/{organization_id_or_slug}/issues/{issue_id}/ [intent=reverse_etl availability=implemented write=sentry_delete_api_0_organizations_organization_id_or_slug_issues_issue_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Sentry DELETE on /api/0/organizations/{organization_id_or_slug}/issues/{issue_id}/; changes Sentry data.; flags: --issue-id (required, max 32768 bytes), --organization-id-or-slug (required, max 32768 bytes)
  - api op-44454c455445202f6170692f302f6f7267616e697a6174696f6e732f7b6f7267616e697a6174696f6e5f69645f6f725f736c75677d2f6973737565732f7b69737375655f69647d2f696e746567726174696f6e732f7b696e746567726174696f6e5f69647d2f - DELETE /api/0/organizations/{organization_id_or_slug}/issues/{issue_id}/integrations/{integration_id}/ [intent=reverse_etl availability=implemented write=sentry_delete_api_0_organizations_organization_id_or_slug_issues_issue_id_integrations_integration_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Sentry DELETE on /api/0/organizations/{organization_id_or_slug}/issues/{issue_id}/integrations/{integration_id}/; changes Sentry data.; flags: --externalIssue (required), --integration-id (required, max 32768 bytes), --issue-id (required, max 32768 bytes), --organization-id-or-slug (required, max 32768 bytes)
  - api op-44454c455445202f6170692f302f6f7267616e697a6174696f6e732f7b6f7267616e697a6174696f6e5f69645f6f725f736c75677d2f6d656d626572732f7b6d656d6265725f69647d2f - DELETE /api/0/organizations/{organization_id_or_slug}/members/{member_id}/ [intent=reverse_etl availability=implemented write=sentry_delete_api_0_organizations_organization_id_or_slug_members_member_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Sentry DELETE on /api/0/organizations/{organization_id_or_slug}/members/{member_id}/; changes Sentry data.; flags: --member-id (required, max 32768 bytes), --organization-id-or-slug (required, max 32768 bytes)
  - api op-44454c455445202f6170692f302f6f7267616e697a6174696f6e732f7b6f7267616e697a6174696f6e5f69645f6f725f736c75677d2f6d656d626572732f7b6d656d6265725f69647d2f7465616d732f7b7465616d5f69645f6f725f736c75677d2f - DELETE /api/0/organizations/{organization_id_or_slug}/members/{member_id}/teams/{team_id_or_slug}/ [intent=reverse_etl availability=implemented write=sentry_delete_api_0_organizations_organization_id_or_slug_members_member_id_teams_team_id_or_slug]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Sentry DELETE on /api/0/organizations/{organization_id_or_slug}/members/{member_id}/teams/{team_id_or_slug}/; changes Sentry data.; flags: --member-id (required, max 32768 bytes), --organization-id-or-slug (required, max 32768 bytes), --team-id-or-slug (required, max 32768 bytes)
  - api op-44454c455445202f6170692f302f6f7267616e697a6174696f6e732f7b6f7267616e697a6174696f6e5f69645f6f725f736c75677d2f6e6f74696669636174696f6e732f616374696f6e732f7b616374696f6e5f69647d2f - DELETE /api/0/organizations/{organization_id_or_slug}/notifications/actions/{action_id}/ [intent=reverse_etl availability=implemented write=sentry_delete_api_0_organizations_organization_id_or_slug_notifications_actions_action_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Sentry DELETE on /api/0/organizations/{organization_id_or_slug}/notifications/actions/{action_id}/; changes Sentry data.; flags: --action-id (required), --organization-id-or-slug (required, max 32768 bytes)
  - api op-44454c455445202f6170692f302f6f7267616e697a6174696f6e732f7b6f7267616e697a6174696f6e5f69645f6f725f736c75677d2f70726570726f646172746966616374732f736e617073686f74732f7b736e617073686f745f69647d2f - DELETE /api/0/organizations/{organization_id_or_slug}/preprodartifacts/snapshots/{snapshot_id}/ [intent=reverse_etl availability=implemented write=sentry_delete_api_0_organizations_organization_id_or_slug_preprodartifacts_snapshots_snapshot_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Sentry DELETE on /api/0/organizations/{organization_id_or_slug}/preprodartifacts/snapshots/{snapshot_id}/; changes Sentry data.; flags: --organization-id-or-slug (required, max 32768 bytes), --snapshot-id (required, max 32768 bytes)
  - api op-44454c455445202f6170692f302f6f7267616e697a6174696f6e732f7b6f7267616e697a6174696f6e5f69645f6f725f736c75677d2f72656c65617365732f7b76657273696f6e7d2f - DELETE /api/0/organizations/{organization_id_or_slug}/releases/{version}/ [intent=reverse_etl availability=implemented write=sentry_delete_api_0_organizations_organization_id_or_slug_releases_version]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Sentry DELETE on /api/0/organizations/{organization_id_or_slug}/releases/{version}/; changes Sentry data.; flags: --organization-id-or-slug (required, max 32768 bytes), --version (required, max 32768 bytes)
  - api op-44454c455445202f6170692f302f6f7267616e697a6174696f6e732f7b6f7267616e697a6174696f6e5f69645f6f725f736c75677d2f72656c65617365732f7b76657273696f6e7d2f66696c65732f7b66696c655f69647d2f - DELETE /api/0/organizations/{organization_id_or_slug}/releases/{version}/files/{file_id}/ [intent=reverse_etl availability=implemented write=sentry_delete_api_0_organizations_organization_id_or_slug_releases_version_files_file_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Sentry DELETE on /api/0/organizations/{organization_id_or_slug}/releases/{version}/files/{file_id}/; changes Sentry data.; flags: --file-id (required, max 32768 bytes), --organization-id-or-slug (required, max 32768 bytes), --version (required, max 32768 bytes)
  - api op-44454c455445202f6170692f302f6f7267616e697a6174696f6e732f7b6f7267616e697a6174696f6e5f69645f6f725f736c75677d2f7363696d2f76322f47726f7570732f7b7465616d5f69645f6f725f736c75677d - DELETE /api/0/organizations/{organization_id_or_slug}/scim/v2/Groups/{team_id_or_slug} [intent=reverse_etl availability=implemented write=sentry_delete_api_0_organizations_organization_id_or_slug_scim_v2_groups_team_id_or_slug]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Sentry DELETE on /api/0/organizations/{organization_id_or_slug}/scim/v2/Groups/{team_id_or_slug}; changes Sentry data.; flags: --organization-id-or-slug (required, max 32768 bytes), --team-id-or-slug (required, max 32768 bytes)
  - api op-44454c455445202f6170692f302f6f7267616e697a6174696f6e732f7b6f7267616e697a6174696f6e5f69645f6f725f736c75677d2f7363696d2f76322f55736572732f7b6d656d6265725f69647d - DELETE /api/0/organizations/{organization_id_or_slug}/scim/v2/Users/{member_id} [intent=reverse_etl availability=implemented write=sentry_delete_api_0_organizations_organization_id_or_slug_scim_v2_users_member_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Sentry DELETE on /api/0/organizations/{organization_id_or_slug}/scim/v2/Users/{member_id}; changes Sentry data.; flags: --member-id (required, max 32768 bytes), --organization-id-or-slug (required, max 32768 bytes)
  - api op-44454c455445202f6170692f302f6f7267616e697a6174696f6e732f7b6f7267616e697a6174696f6e5f69645f6f725f736c75677d2f776f726b666c6f77732f7b776f726b666c6f775f69647d2f - DELETE /api/0/organizations/{organization_id_or_slug}/workflows/{workflow_id}/ [intent=reverse_etl availability=implemented write=sentry_delete_api_0_organizations_organization_id_or_slug_workflows_workflow_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Sentry DELETE on /api/0/organizations/{organization_id_or_slug}/workflows/{workflow_id}/; changes Sentry data.; flags: --organization-id-or-slug (required, max 32768 bytes), --workflow-id (required)
  - api op-44454c455445202f6170692f302f70726f6a656374732f7b6f7267616e697a6174696f6e5f69645f6f725f736c75677d2f7b70726f6a6563745f69645f6f725f736c75677d2f - DELETE /api/0/projects/{organization_id_or_slug}/{project_id_or_slug}/ [intent=reverse_etl availability=implemented write=sentry_delete_api_0_projects_organization_id_or_slug_project_id_or_slug]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Sentry DELETE on /api/0/projects/{organization_id_or_slug}/{project_id_or_slug}/; changes Sentry data.; flags: --organization-id-or-slug (required, max 32768 bytes), --project-id-or-slug (required, max 32768 bytes)
  - api op-44454c455445202f6170692f302f70726f6a656374732f7b6f7267616e697a6174696f6e5f69645f6f725f736c75677d2f7b70726f6a6563745f69645f6f725f736c75677d2f686f6f6b732f7b686f6f6b5f69647d2f - DELETE /api/0/projects/{organization_id_or_slug}/{project_id_or_slug}/hooks/{hook_id}/ [intent=reverse_etl availability=implemented write=sentry_delete_api_0_projects_organization_id_or_slug_project_id_or_slug_hooks_hook_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Sentry DELETE on /api/0/projects/{organization_id_or_slug}/{project_id_or_slug}/hooks/{hook_id}/; changes Sentry data.; flags: --hook-id (required, max 32768 bytes), --organization-id-or-slug (required, max 32768 bytes), --project-id-or-slug (required, max 32768 bytes)
  - api op-44454c455445202f6170692f302f70726f6a656374732f7b6f7267616e697a6174696f6e5f69645f6f725f736c75677d2f7b70726f6a6563745f69645f6f725f736c75677d2f6973737565732f - DELETE /api/0/projects/{organization_id_or_slug}/{project_id_or_slug}/issues/ [intent=reverse_etl availability=implemented write=sentry_delete_api_0_projects_organization_id_or_slug_project_id_or_slug_issues]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Sentry DELETE on /api/0/projects/{organization_id_or_slug}/{project_id_or_slug}/issues/; changes Sentry data.; flags: --id, --organization-id-or-slug (required, max 32768 bytes), --project-id-or-slug (required, max 32768 bytes)
  - api op-44454c455445202f6170692f302f70726f6a656374732f7b6f7267616e697a6174696f6e5f69645f6f725f736c75677d2f7b70726f6a6563745f69645f6f725f736c75677d2f6b6579732f7b6b65795f69647d2f - DELETE /api/0/projects/{organization_id_or_slug}/{project_id_or_slug}/keys/{key_id}/ [intent=reverse_etl availability=implemented write=sentry_delete_api_0_projects_organization_id_or_slug_project_id_or_slug_keys_key_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Sentry DELETE on /api/0/projects/{organization_id_or_slug}/{project_id_or_slug}/keys/{key_id}/; changes Sentry data.; flags: --key-id (required, max 32768 bytes), --organization-id-or-slug (required, max 32768 bytes), --project-id-or-slug (required, max 32768 bytes)
  - api op-44454c455445202f6170692f302f70726f6a656374732f7b6f7267616e697a6174696f6e5f69645f6f725f736c75677d2f7b70726f6a6563745f69645f6f725f736c75677d2f72656c65617365732f7b76657273696f6e7d2f66696c65732f7b66696c655f69647d2f - DELETE /api/0/projects/{organization_id_or_slug}/{project_id_or_slug}/releases/{version}/files/{file_id}/ [intent=reverse_etl availability=implemented write=sentry_delete_api_0_projects_organization_id_or_slug_project_id_or_slug_releases_version_files_file_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Sentry DELETE on /api/0/projects/{organization_id_or_slug}/{project_id_or_slug}/releases/{version}/files/{file_id}/; changes Sentry data.; flags: --file-id (required, max 32768 bytes), --organization-id-or-slug (required, max 32768 bytes), --project-id-or-slug (required, max 32768 bytes), --version (required, max 32768 bytes)
  - api op-44454c455445202f6170692f302f70726f6a656374732f7b6f7267616e697a6174696f6e5f69645f6f725f736c75677d2f7b70726f6a6563745f69645f6f725f736c75677d2f7265706c6179732f7b7265706c61795f69647d2f - DELETE /api/0/projects/{organization_id_or_slug}/{project_id_or_slug}/replays/{replay_id}/ [intent=reverse_etl availability=implemented write=sentry_delete_api_0_projects_organization_id_or_slug_project_id_or_slug_replays_replay_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Sentry DELETE on /api/0/projects/{organization_id_or_slug}/{project_id_or_slug}/replays/{replay_id}/; changes Sentry data.; flags: --organization-id-or-slug (required, max 32768 bytes), --project-id-or-slug (required, max 32768 bytes), --replay-id (required, max 32768 bytes)
  - api op-44454c455445202f6170692f302f70726f6a656374732f7b6f7267616e697a6174696f6e5f69645f6f725f736c75677d2f7b70726f6a6563745f69645f6f725f736c75677d2f73796d626f6c2d736f75726365732f - DELETE /api/0/projects/{organization_id_or_slug}/{project_id_or_slug}/symbol-sources/ [intent=reverse_etl availability=implemented write=sentry_delete_api_0_projects_organization_id_or_slug_project_id_or_slug_symbol_sources]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Sentry DELETE on /api/0/projects/{organization_id_or_slug}/{project_id_or_slug}/symbol-sources/; changes Sentry data.; flags: --id (required, max 32768 bytes), --organization-id-or-slug (required, max 32768 bytes), --project-id-or-slug (required, max 32768 bytes)
  - api op-44454c455445202f6170692f302f70726f6a656374732f7b6f7267616e697a6174696f6e5f69645f6f725f736c75677d2f7b70726f6a6563745f69645f6f725f736c75677d2f7465616d732f7b7465616d5f69645f6f725f736c75677d2f - DELETE /api/0/projects/{organization_id_or_slug}/{project_id_or_slug}/teams/{team_id_or_slug}/ [intent=reverse_etl availability=implemented write=sentry_delete_api_0_projects_organization_id_or_slug_project_id_or_slug_teams_team_id_or_slug]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Sentry DELETE on /api/0/projects/{organization_id_or_slug}/{project_id_or_slug}/teams/{team_id_or_slug}/; changes Sentry data.; flags: --organization-id-or-slug (required, max 32768 bytes), --project-id-or-slug (required, max 32768 bytes), --team-id-or-slug (required, max 32768 bytes)
  - api op-44454c455445202f6170692f302f73656e7472792d6170702d696e7374616c6c6174696f6e732f7b757569647d2f65787465726e616c2d6973737565732f7b65787465726e616c5f69737375655f69647d2f - DELETE /api/0/sentry-app-installations/{uuid}/external-issues/{external_issue_id}/ [intent=reverse_etl availability=implemented write=sentry_delete_api_0_sentry_app_installations_uuid_external_issues_external_issue_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Sentry DELETE on /api/0/sentry-app-installations/{uuid}/external-issues/{external_issue_id}/; changes Sentry data.; flags: --external-issue-id (required, max 32768 bytes), --uuid (required, max 32768 bytes)
  - api op-44454c455445202f6170692f302f73656e7472792d617070732f7b73656e7472795f6170705f69645f6f725f736c75677d2f - DELETE /api/0/sentry-apps/{sentry_app_id_or_slug}/ [intent=reverse_etl availability=implemented write=sentry_delete_api_0_sentry_apps_sentry_app_id_or_slug]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Sentry DELETE on /api/0/sentry-apps/{sentry_app_id_or_slug}/; changes Sentry data.; flags: --sentry-app-id-or-slug (required, max 32768 bytes)
  - api op-44454c455445202f6170692f302f7465616d732f7b6f7267616e697a6174696f6e5f69645f6f725f736c75677d2f7b7465616d5f69645f6f725f736c75677d2f - DELETE /api/0/teams/{organization_id_or_slug}/{team_id_or_slug}/ [intent=reverse_etl availability=implemented write=sentry_delete_api_0_teams_organization_id_or_slug_team_id_or_slug]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Sentry DELETE on /api/0/teams/{organization_id_or_slug}/{team_id_or_slug}/; changes Sentry data.; flags: --organization-id-or-slug (required, max 32768 bytes), --team-id-or-slug (required, max 32768 bytes)
  - api op-44454c455445202f6170692f302f7465616d732f7b6f7267616e697a6174696f6e5f69645f6f725f736c75677d2f7b7465616d5f69645f6f725f736c75677d2f65787465726e616c2d7465616d732f7b65787465726e616c5f7465616d5f69647d2f - DELETE /api/0/teams/{organization_id_or_slug}/{team_id_or_slug}/external-teams/{external_team_id}/ [intent=reverse_etl availability=implemented write=sentry_delete_api_0_teams_organization_id_or_slug_team_id_or_slug_external_teams_external_team_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Sentry DELETE on /api/0/teams/{organization_id_or_slug}/{team_id_or_slug}/external-teams/{external_team_id}/; changes Sentry data.; flags: --external-team-id (required), --organization-id-or-slug (required, max 32768 bytes), --team-id-or-slug (required, max 32768 bytes)
  - api op-504f5354202f6170692f302f6f7267616e697a6174696f6e732f7b6f7267616e697a6174696f6e5f69645f6f725f736c75677d2f6d656d626572732f7b6d656d6265725f69647d2f7465616d732f7b7465616d5f69645f6f725f736c75677d2f - POST /api/0/organizations/{organization_id_or_slug}/members/{member_id}/teams/{team_id_or_slug}/ [intent=reverse_etl availability=implemented write=sentry_post_api_0_organizations_organization_id_or_slug_members_member_id_teams_team_id_or_slug]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Sentry POST on /api/0/organizations/{organization_id_or_slug}/members/{member_id}/teams/{team_id_or_slug}/; changes Sentry data.; flags: --member-id (required, max 32768 bytes), --organization-id-or-slug (required, max 32768 bytes), --team-id-or-slug (required, max 32768 bytes)
  - api op-504f5354202f6170692f302f70726f6a656374732f7b6f7267616e697a6174696f6e5f69645f6f725f736c75677d2f7b70726f6a6563745f69645f6f725f736c75677d2f70726570726f646172746966616374732f736e617073686f74732f - POST /api/0/projects/{organization_id_or_slug}/{project_id_or_slug}/preprodartifacts/snapshots/ [intent=reverse_etl availability=implemented write=sentry_post_api_0_projects_organization_id_or_slug_project_id_or_slug_preprodartifacts_snapshots]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Sentry POST on /api/0/projects/{organization_id_or_slug}/{project_id_or_slug}/preprodartifacts/snapshots/; changes Sentry data.; flags: --organization-id-or-slug (required, max 32768 bytes), --project-id-or-slug (required, max 32768 bytes)
  - api op-504f5354202f6170692f302f70726f6a656374732f7b6f7267616e697a6174696f6e5f69645f6f725f736c75677d2f7b70726f6a6563745f69645f6f725f736c75677d2f7465616d732f7b7465616d5f69645f6f725f736c75677d2f - POST /api/0/projects/{organization_id_or_slug}/{project_id_or_slug}/teams/{team_id_or_slug}/ [intent=reverse_etl availability=implemented write=sentry_post_api_0_projects_organization_id_or_slug_project_id_or_slug_teams_team_id_or_slug]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Sentry POST on /api/0/projects/{organization_id_or_slug}/{project_id_or_slug}/teams/{team_id_or_slug}/; changes Sentry data.; flags: --organization-id-or-slug (required, max 32768 bytes), --project-id-or-slug (required, max 32768 bytes), --team-id-or-slug (required, max 32768 bytes)

## Sync Transport

- Source transport: declared
- Destination transport: unsupported
- A declared transport still requires runtime preflight and externally verified conformance; it is not a certification claim.
- Source executor: declarative_api/declarative_stream_source

## Commands

### Inspect as a manual

```bash
pm connectors inspect sentry
```

### Inspect as structured JSON

```bash
pm connectors inspect sentry --json
```

## Agent Rules

- Run pm connectors inspect sentry before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
