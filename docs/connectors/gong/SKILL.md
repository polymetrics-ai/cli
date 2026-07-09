---
name: pm-gong
description: Gong connector knowledge and safe action guide.
---

# pm-gong

## Purpose

Reads Gong users, calls, scorecards, settings, flows, and related public API resources; models Gong mutations as typed reverse-ETL actions.

## Icon

- asset: icons/gong.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://us-66463.app.gong.io/settings/api/documentation

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- max_pages
- mode
- page_size
- start_date
- access_key (secret)
- access_key_secret (secret)

## ETL Streams

- users:
  - primary key: id
  - cursor: created
  - fields: active(), created(), email_address(), first_name(), id(), last_name(), manager_id(), phone_number(), title()
- calls:
  - primary key: id
  - cursor: started
  - fields: direction(), duration(), id(), is_private(), language(), media(), scheduled(), scope(), started(), system(), title(), url()
- scorecards:
  - primary key: scorecardId
  - cursor: updated
  - fields: created(), enabled(), scorecardId(), scorecardName(), updated(), workspaceId()
- crm_integrations:
  - fields: created(), id(), name(), title(), updated()
- workspaces:
  - fields: created(), id(), name(), title(), updated()
- trackers:
  - fields: created(), id(), name(), title(), updated()
- briefs:
  - fields: created(), id(), name(), title(), updated()
- library_folders:
  - fields: created(), id(), name(), title(), updated()
- flows:
  - fields: created(), id(), name(), title(), updated()
- flow_folders:
  - fields: created(), id(), name(), title(), updated()
- call_outcomes:
  - fields: created(), id(), name(), title(), updated()
- permission_profiles:
  - fields: created(), id(), name(), title(), updated()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Reverse ETL Actions

- add_call:
  - endpoint: POST /calls
  - risk: high: sends call, engagement, or digital interaction content to Gong; requires reverse ETL approval
- create_permission_profile:
  - endpoint: POST /permission-profile?workspaceId={{ record.workspaceId }}
  - required fields: workspaceId
  - risk: high: administrative Gong settings or permissions mutation; requires reverse ETL approval and destructive confirmation
- update_permission_profile:
  - endpoint: PUT /permission-profile?profileId={{ record.profileId }}
  - required fields: profileId
  - risk: high: administrative Gong settings or permissions mutation; requires reverse ETL approval and destructive confirmation
- update_meeting:
  - endpoint: PUT /meetings/{meetingId}
  - required fields: meetingId
  - risk: medium: mutates Gong API state; requires reverse ETL approval
- delete_meeting:
  - endpoint: DELETE /meetings/{meetingId}
  - required fields: meetingId
  - risk: high: removes Gong access, integration, meeting, or flow assignment state; requires reverse ETL approval and destructive confirmation
- content_viewed:
  - endpoint: PUT /customer-engagement/content/viewed
  - risk: high: sends call, engagement, or digital interaction content to Gong; requires reverse ETL approval
- content_shared:
  - endpoint: PUT /customer-engagement/content/shared
  - risk: high: sends call, engagement, or digital interaction content to Gong; requires reverse ETL approval
- custom_action:
  - endpoint: PUT /customer-engagement/action
  - risk: high: sends call, engagement, or digital interaction content to Gong; requires reverse ETL approval
- register_crm_integration:
  - endpoint: PUT /crm/integrations
  - risk: medium: mutates Gong API state; requires reverse ETL approval
- delete_crm_integration:
  - endpoint: DELETE /crm/integrations?integrationId={{ record.integrationId }}&clientRequestId={{ record.clientRequestId }}
  - required fields: integrationId, clientRequestId
  - risk: high: removes Gong access, integration, meeting, or flow assignment state; requires reverse ETL approval and destructive confirmation
- add_calls_users_access:
  - endpoint: PUT /calls/users-access
  - risk: medium: mutates Gong API state; requires reverse ETL approval
- delete_calls_users_access:
  - endpoint: DELETE /calls/users-access
  - risk: high: removes Gong access, integration, meeting, or flow assignment state; requires reverse ETL approval and destructive confirmation
- create_meeting:
  - endpoint: POST /meetings
  - risk: medium: mutates Gong API state; requires reverse ETL approval
- integration_settings:
  - endpoint: POST /integration-settings
  - risk: high: administrative Gong settings or permissions mutation; requires reverse ETL approval and destructive confirmation
- unassign_flows_by_instance_id:
  - endpoint: POST /flows/prospects/unassign-flows-by-instance-id
  - risk: high: removes Gong access, integration, meeting, or flow assignment state; requires reverse ETL approval and destructive confirmation
- unassign_flows_by_crm_id:
  - endpoint: POST /flows/prospects/unassign-flows-by-crm-id
  - risk: high: removes Gong access, integration, meeting, or flow assignment state; requires reverse ETL approval and destructive confirmation
- submit_flow_prospects_bulk_assignment:
  - endpoint: POST /flows/prospects/bulk-assignments
  - risk: medium: mutates Gong API state; requires reverse ETL approval
- assign_prospects:
  - endpoint: POST /flows/prospects/assign
  - risk: medium: mutates Gong API state; requires reverse ETL approval
- assign_prospects_cool_off_override:
  - endpoint: POST /flows/prospects/assign/cool-off-override
  - risk: medium: mutates Gong API state; requires reverse ETL approval
- add_digital_interaction:
  - endpoint: POST /digital-interaction
  - risk: high: sends call, engagement, or digital interaction content to Gong; requires reverse ETL approval
- purge_phone_number:
  - endpoint: POST /data-privacy/erase-data-for-phone-number?phoneNumber={{ record.phoneNumber }}
  - required fields: phoneNumber
  - risk: critical: destructive Gong data privacy erasure; requires reverse ETL plan, preview, approval, and destructive confirmation
- purge_email_address:
  - endpoint: POST /data-privacy/erase-data-for-email-address?emailAddress={{ record.emailAddress }}
  - required fields: emailAddress
  - risk: critical: destructive Gong data privacy erasure; requires reverse ETL plan, preview, approval, and destructive confirmation
- update_task:
  - endpoint: PATCH /tasks/{taskId}
  - required fields: taskId
  - risk: medium: mutates Gong API state; requires reverse ETL approval

## Security

- read risk: external Gong API read of call, user, CRM, settings, flow, and activity data; direct reads are bounded and redacted
- write risk: typed Gong reverse ETL mutations for calls, meetings, CRM, permissions, flows, engagement, and data privacy erasure
- approval: reverse ETL writes require plan, preview, approval, execute; destructive/admin actions require --confirm destructive
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Command Surface

- Inspect, read, and safely plan typed Gong operations.
- Usage: pm gong <command> [flags]
- Source CLI: Gong API (Public OpenAPI 3.0.1)
- Global flags:
  - --credential (string): Credential name to use for the Gong request.
  - --connection (string): Alias for --credential.
  - --config (string_array): Connector config override as key=value.
  - --json (boolean): Emit machine-readable JSON output.
  - --limit (integer): Maximum ETL records to emit for stream commands.
  - --max-bytes (integer): Maximum direct-read response bytes, capped at 1 MiB.
  - --plan (string): Execute an approved reverse-ETL plan by id.
  - --preview (boolean): Preview a reverse-ETL write command without making a network mutation.
  - --approve (boolean): Approve a reverse-ETL plan before execution.
  - --confirm (string): Typed confirmation challenge for destructive reverse-ETL writes.
- Calls
  - calls list - List Gong calls as ETL records. [intent=etl availability=implemented stream=calls]
  - calls get - Retrieve data for a specific call (/v2/calls/{id}) [intent=direct_read availability=implemented]; risk: bounded Gong JSON read; response is limited to 1 MiB and secret/download/content-shaped fields are redacted; flags: --id
  - calls create - sends call, engagement, or digital interaction content to Gong; requires reverse ETL approval [intent=reverse_etl availability=partial write=add_call]; approval: Use reverse ETL plan -> preview -> approval -> execute. Connector command execution is metadata-only for complex object/array records; use typed reverse ETL records.; risk: high: sends call, engagement, or digital interaction content to Gong; requires reverse ETL approval; notes: No raw HTTP body is accepted. Object and array payloads must come from typed reverse-ETL records validated by writes.json.; flags: --actualStart, --clientUniqueId, --direction, --parties, --primaryUser
  - calls users-access add - mutates Gong API state; requires reverse ETL approval [intent=reverse_etl availability=partial write=add_calls_users_access]; approval: Use reverse ETL plan -> preview -> approval -> execute. Connector command execution is metadata-only for complex object/array records; use typed reverse ETL records.; risk: medium: mutates Gong API state; requires reverse ETL approval; notes: No raw HTTP body is accepted. Object and array payloads must come from typed reverse-ETL records validated by writes.json.
  - calls users-access delete - removes Gong access, integration, meeting, or flow assignment state; requires reverse ETL approval and destructive confirmation [intent=reverse_etl availability=partial write=delete_calls_users_access]; approval: Use reverse ETL plan -> preview -> approval -> execute. Connector command execution is metadata-only for complex object/array records; use typed reverse ETL records.; risk: high: removes Gong access, integration, meeting, or flow assignment state; requires reverse ETL approval and destructive confirmation; notes: No raw HTTP body is accepted. Object and array payloads must come from typed reverse-ETL records validated by writes.json.
  - calls extensive - Retrieve detailed call data by various filters (/v2/calls/extensive) [intent=direct_read availability=planned]; approval: none until executor support exists; no mutation is modeled; risk: bounded read-query planned; typed request body schema is declared in operations.json; notes: Blocked by operation executor support for fixed JSON POST read-query bodies, not by endpoint sensitivity alone.
  - calls users-access get - Retrieve users that have specific individual access to calls (/v2/calls/users-access) [intent=direct_read availability=planned]; approval: none until executor support exists; no mutation is modeled; risk: bounded read-query planned; typed request body schema is declared in operations.json; notes: Blocked by operation executor support for fixed JSON POST read-query bodies, not by endpoint sensitivity alone.
  - calls transcript - Retrieve transcripts of calls (/v2/calls/transcript) [intent=direct_read availability=planned]; approval: none until executor support exists; no mutation is modeled; risk: bounded read-query planned; typed request body schema is declared in operations.json; notes: Blocked by operation executor support for fixed JSON POST read-query bodies, not by endpoint sensitivity alone.
  - calls upload-media - Add call media (/v2/calls/{id}/media) [intent=reverse_etl availability=planned]; approval: reverse ETL plan -> preview -> approval -> execute; executor support for multipart/top-level array bodies is required before execution; risk: bounded sensitive/admin payload operation; typed schema and redaction policy are declared in operations.json; notes: No generic file upload or raw body command is exposed.
- Users
  - users list - List Gong users as ETL records. [intent=etl availability=implemented stream=users]
  - users get - Retrieve user (/v2/users/{id}) [intent=direct_read availability=implemented]; risk: bounded Gong JSON read; response is limited to 1 MiB and secret/download/content-shaped fields are redacted; flags: --id
  - users settings-history - Retrieve user settings history (/v2/users/{id}/settings-history) [intent=direct_read availability=implemented]; risk: bounded Gong JSON read; response is limited to 1 MiB and secret/download/content-shaped fields are redacted; flags: --id
  - users extensive - List users by filter (/v2/users/extensive) [intent=direct_read availability=planned]; approval: none until executor support exists; no mutation is modeled; risk: bounded read-query planned; typed request body schema is declared in operations.json; notes: Blocked by operation executor support for fixed JSON POST read-query bodies, not by endpoint sensitivity alone.
- Scorecards
  - scorecards list - List Gong scorecards as ETL records. [intent=etl availability=implemented stream=scorecards]
- Workspaces
  - workspaces list - List Gong workspaces as ETL records. [intent=etl availability=implemented stream=workspaces]
- Settings
  - settings trackers list - List Gong tracker settings as ETL records. [intent=etl availability=implemented stream=trackers]
  - settings briefs list - List Gong AI brief settings as ETL records. [intent=etl availability=implemented stream=briefs]
- Permissions
  - permissions profiles list - List Gong permission profiles as ETL records. [intent=etl availability=implemented stream=permission_profiles]
  - permissions profile get - Permission profile for a given profile Id (/v2/permission-profile) [intent=direct_read availability=implemented]; risk: bounded Gong JSON read; response is limited to 1 MiB and secret/download/content-shaped fields are redacted; flags: --profileId
  - permissions profile-users list - List all users attached to a given permission profile (/v2/permission-profile/users) [intent=direct_read availability=implemented]; risk: bounded Gong JSON read; response is limited to 1 MiB and secret/download/content-shaped fields are redacted; flags: --profileId
  - permissions profile create - administrative Gong settings or permissions mutation; requires reverse ETL approval and destructive confirmation [intent=reverse_etl availability=partial write=create_permission_profile]; approval: Use reverse ETL plan -> preview -> approval -> execute. Connector command execution is metadata-only for complex object/array records; use typed reverse ETL records.; risk: high: administrative Gong settings or permissions mutation; requires reverse ETL approval and destructive confirmation; notes: No raw HTTP body is accepted. Object and array payloads must come from typed reverse-ETL records validated by writes.json.; flags: --workspaceId
  - permissions profile update - administrative Gong settings or permissions mutation; requires reverse ETL approval and destructive confirmation [intent=reverse_etl availability=partial write=update_permission_profile]; approval: Use reverse ETL plan -> preview -> approval -> execute. Connector command execution is metadata-only for complex object/array records; use typed reverse ETL records.; risk: high: administrative Gong settings or permissions mutation; requires reverse ETL approval and destructive confirmation; notes: No raw HTTP body is accepted. Object and array payloads must come from typed reverse-ETL records validated by writes.json.; flags: --profileId
- CRM
  - crm integrations list - List Generic CRM integrations as ETL records. [intent=etl availability=implemented stream=crm_integrations]
  - crm entity-schema get - List Schema Fields (/v2/crm/entity-schema) [intent=direct_read availability=implemented]; risk: bounded Gong JSON read; response is limited to 1 MiB and secret/download/content-shaped fields are redacted; flags: --integrationId, --objectType
  - crm entities get - Get CRM objects (/v2/crm/entities) [intent=direct_read availability=implemented]; risk: bounded Gong JSON read; response is limited to 1 MiB and secret/download/content-shaped fields are redacted; flags: --integrationId, --objectType, --objectsCrmIds
  - crm request-status get - Get Request Status (/v2/crm/request-status) [intent=direct_read availability=implemented]; risk: bounded Gong JSON read; response is limited to 1 MiB and secret/download/content-shaped fields are redacted; flags: --integrationId, --clientRequestId
  - crm integrations register - mutates Gong API state; requires reverse ETL approval [intent=reverse_etl availability=partial write=register_crm_integration]; approval: Use reverse ETL plan -> preview -> approval -> execute. Connector command execution is metadata-only for complex object/array records; use typed reverse ETL records.; risk: medium: mutates Gong API state; requires reverse ETL approval; notes: No raw HTTP body is accepted. Object and array payloads must come from typed reverse-ETL records validated by writes.json.; flags: --name, --ownerEmail
  - crm integrations delete - removes Gong access, integration, meeting, or flow assignment state; requires reverse ETL approval and destructive confirmation [intent=reverse_etl availability=partial write=delete_crm_integration]; approval: Use reverse ETL plan -> preview -> approval -> execute. Connector command execution is metadata-only for complex object/array records; use typed reverse ETL records.; risk: high: removes Gong access, integration, meeting, or flow assignment state; requires reverse ETL approval and destructive confirmation; notes: No raw HTTP body is accepted. Object and array payloads must come from typed reverse-ETL records validated by writes.json.; flags: --clientRequestId, --integrationId
  - crm upload-entities - Upload CRM objects (/v2/crm/entities) [intent=reverse_etl availability=planned]; approval: reverse ETL plan -> preview -> approval -> execute; executor support for multipart/top-level array bodies is required before execution; risk: bounded sensitive/admin payload operation; typed schema and redaction policy are declared in operations.json; notes: No generic file upload or raw body command is exposed.
  - crm upload-entity-schema - Upload Object Schema (/v2/crm/entity-schema) [intent=reverse_etl availability=planned]; approval: reverse ETL plan -> preview -> approval -> execute; executor support for multipart/top-level array bodies is required before execution; risk: bounded sensitive/admin payload operation; typed schema and redaction policy are declared in operations.json; notes: No generic file upload or raw body command is exposed.
- Meetings
  - meetings update - mutates Gong API state; requires reverse ETL approval [intent=reverse_etl availability=partial write=update_meeting]; approval: Use reverse ETL plan -> preview -> approval -> execute. Connector command execution is metadata-only for complex object/array records; use typed reverse ETL records.; risk: medium: mutates Gong API state; requires reverse ETL approval; notes: No raw HTTP body is accepted. Object and array payloads must come from typed reverse-ETL records validated by writes.json.; flags: --endTime, --invitees, --meetingId, --organizerEmail, --startTime
  - meetings delete - removes Gong access, integration, meeting, or flow assignment state; requires reverse ETL approval and destructive confirmation [intent=reverse_etl availability=partial write=delete_meeting]; approval: Use reverse ETL plan -> preview -> approval -> execute. Connector command execution is metadata-only for complex object/array records; use typed reverse ETL records.; risk: high: removes Gong access, integration, meeting, or flow assignment state; requires reverse ETL approval and destructive confirmation; notes: No raw HTTP body is accepted. Object and array payloads must come from typed reverse-ETL records validated by writes.json.; flags: --meetingId
  - meetings create - mutates Gong API state; requires reverse ETL approval [intent=reverse_etl availability=partial write=create_meeting]; approval: Use reverse ETL plan -> preview -> approval -> execute. Connector command execution is metadata-only for complex object/array records; use typed reverse ETL records.; risk: medium: mutates Gong API state; requires reverse ETL approval; notes: No raw HTTP body is accepted. Object and array payloads must come from typed reverse-ETL records validated by writes.json.; flags: --endTime, --invitees, --organizerEmail, --startTime
  - meetings integration-status - Validate Gong meeting Integration (/v2/meetings/integration/status) [intent=direct_read availability=planned]; approval: none until executor support exists; no mutation is modeled; risk: bounded read-query planned; typed request body schema is declared in operations.json; notes: Blocked by operation executor support for fixed JSON POST read-query bodies, not by endpoint sensitivity alone.
- Engagement
  - engagement content-viewed - sends call, engagement, or digital interaction content to Gong; requires reverse ETL approval [intent=reverse_etl availability=partial write=content_viewed]; approval: Use reverse ETL plan -> preview -> approval -> execute. Connector command execution is metadata-only for complex object/array records; use typed reverse ETL records.; risk: high: sends call, engagement, or digital interaction content to Gong; requires reverse ETL approval; notes: No raw HTTP body is accepted. Object and array payloads must come from typed reverse-ETL records validated by writes.json.; flags: --contentId, --contentTitle, --contentUrl, --eventTimestamp, --reportingSystem
  - engagement content-shared - sends call, engagement, or digital interaction content to Gong; requires reverse ETL approval [intent=reverse_etl availability=partial write=content_shared]; approval: Use reverse ETL plan -> preview -> approval -> execute. Connector command execution is metadata-only for complex object/array records; use typed reverse ETL records.; risk: high: sends call, engagement, or digital interaction content to Gong; requires reverse ETL approval; notes: No raw HTTP body is accepted. Object and array payloads must come from typed reverse-ETL records validated by writes.json.; flags: --contentId, --contentTitle, --contentUrl, --eventTimestamp, --reportingSystem
  - engagement custom-action - sends call, engagement, or digital interaction content to Gong; requires reverse ETL approval [intent=reverse_etl availability=partial write=custom_action]; approval: Use reverse ETL plan -> preview -> approval -> execute. Connector command execution is metadata-only for complex object/array records; use typed reverse ETL records.; risk: high: sends call, engagement, or digital interaction content to Gong; requires reverse ETL approval; notes: No raw HTTP body is accepted. Object and array payloads must come from typed reverse-ETL records validated by writes.json.; flags: --eventTimestamp, --reportingSystem
- Stats
  - stats interaction - Retrieve interaction stats for applicable users by date (/v2/stats/interaction) [intent=direct_read availability=planned]; approval: none until executor support exists; no mutation is modeled; risk: bounded read-query planned; typed request body schema is declared in operations.json; notes: Blocked by operation executor support for fixed JSON POST read-query bodies, not by endpoint sensitivity alone.
  - stats activity-scorecards - Retrieve answered scorecards for applicable reviewed users or scorecards for a date range (/v2/stats/activity/scorecards) [intent=direct_read availability=planned]; approval: none until executor support exists; no mutation is modeled; risk: bounded read-query planned; typed request body schema is declared in operations.json; notes: Blocked by operation executor support for fixed JSON POST read-query bodies, not by endpoint sensitivity alone.
  - stats activity-day-by-day - Retrieve daily activity for applicable users for a date range (/v2/stats/activity/day-by-day) [intent=direct_read availability=planned]; approval: none until executor support exists; no mutation is modeled; risk: bounded read-query planned; typed request body schema is declared in operations.json; notes: Blocked by operation executor support for fixed JSON POST read-query bodies, not by endpoint sensitivity alone.
  - stats activity-aggregate - Retrieve aggregated activity for defined users by date (/v2/stats/activity/aggregate) [intent=direct_read availability=planned]; approval: none until executor support exists; no mutation is modeled; risk: bounded read-query planned; typed request body schema is declared in operations.json; notes: Blocked by operation executor support for fixed JSON POST read-query bodies, not by endpoint sensitivity alone.
  - stats activity-aggregate-by-period - Retrieve aggregated activity for defined users by a date range with grouping in time periods (/v2/stats/activity/aggregate-by-period) [intent=direct_read availability=planned]; approval: none until executor support exists; no mutation is modeled; risk: bounded read-query planned; typed request body schema is declared in operations.json; notes: Blocked by operation executor support for fixed JSON POST read-query bodies, not by endpoint sensitivity alone.
- Engage Flows
  - flows list - List Gong Engage flows as ETL records. [intent=etl availability=implemented stream=flows]
  - flows folders list - List Gong Engage flow folders as ETL records. [intent=etl availability=implemented stream=flow_folders]
  - flows bulk-assignment get - Get the results of a bulk assignment of prospects to a flow (/v2/flows/prospects/bulk-assignments/{id}) [intent=direct_read availability=implemented]; risk: bounded Gong JSON read; response is limited to 1 MiB and secret/download/content-shaped fields are redacted; flags: --id
  - flows prospects unassign-by-instance - removes Gong access, integration, meeting, or flow assignment state; requires reverse ETL approval and destructive confirmation [intent=reverse_etl availability=partial write=unassign_flows_by_instance_id]; approval: Use reverse ETL plan -> preview -> approval -> execute. Connector command execution is metadata-only for complex object/array records; use typed reverse ETL records.; risk: high: removes Gong access, integration, meeting, or flow assignment state; requires reverse ETL approval and destructive confirmation; notes: No raw HTTP body is accepted. Object and array payloads must come from typed reverse-ETL records validated by writes.json.; flags: --flowInstanceIds
  - flows prospects unassign-by-crm - removes Gong access, integration, meeting, or flow assignment state; requires reverse ETL approval and destructive confirmation [intent=reverse_etl availability=partial write=unassign_flows_by_crm_id]; approval: Use reverse ETL plan -> preview -> approval -> execute. Connector command execution is metadata-only for complex object/array records; use typed reverse ETL records.; risk: high: removes Gong access, integration, meeting, or flow assignment state; requires reverse ETL approval and destructive confirmation; notes: No raw HTTP body is accepted. Object and array payloads must come from typed reverse-ETL records validated by writes.json.; flags: --crmProspectId
  - flows prospects bulk-assign - mutates Gong API state; requires reverse ETL approval [intent=reverse_etl availability=partial write=submit_flow_prospects_bulk_assignment]; approval: Use reverse ETL plan -> preview -> approval -> execute. Connector command execution is metadata-only for complex object/array records; use typed reverse ETL records.; risk: medium: mutates Gong API state; requires reverse ETL approval; notes: No raw HTTP body is accepted. Object and array payloads must come from typed reverse-ETL records validated by writes.json.; flags: --flowId, --flowInstanceOwnerEmail, --prospects
  - flows prospects assign - mutates Gong API state; requires reverse ETL approval [intent=reverse_etl availability=partial write=assign_prospects]; approval: Use reverse ETL plan -> preview -> approval -> execute. Connector command execution is metadata-only for complex object/array records; use typed reverse ETL records.; risk: medium: mutates Gong API state; requires reverse ETL approval; notes: No raw HTTP body is accepted. Object and array payloads must come from typed reverse-ETL records validated by writes.json.; flags: --crmProspectsIds, --flowId, --flowInstanceOwnerEmail
  - flows prospects assign-cool-off-override - mutates Gong API state; requires reverse ETL approval [intent=reverse_etl availability=partial write=assign_prospects_cool_off_override]; approval: Use reverse ETL plan -> preview -> approval -> execute. Connector command execution is metadata-only for complex object/array records; use typed reverse ETL records.; risk: medium: mutates Gong API state; requires reverse ETL approval; notes: No raw HTTP body is accepted. Object and array payloads must come from typed reverse-ETL records validated by writes.json.; flags: --crmProspectsIds, --flowId, --flowInstanceOwnerEmail
  - flows steps - Get flow details and steps for one or more Engage flows (/v2/flows/steps) [intent=direct_read availability=planned]; approval: none until executor support exists; no mutation is modeled; risk: bounded read-query planned; typed request body schema is declared in operations.json; notes: Blocked by operation executor support for fixed JSON POST read-query bodies, not by endpoint sensitivity alone.
  - flows prospects - List prospects assigned flows (/v2/flows/prospects) [intent=direct_read availability=planned]; approval: none until executor support exists; no mutation is modeled; risk: bounded read-query planned; typed request body schema is declared in operations.json; notes: Blocked by operation executor support for fixed JSON POST read-query bodies, not by endpoint sensitivity alone.
- Library
  - library folders list - List Gong library folders as ETL records. [intent=etl availability=implemented stream=library_folders]
  - library folder-content list - List of calls in a specific folder (/v2/library/folder-content) [intent=direct_read availability=implemented]; risk: bounded Gong JSON read; response is limited to 1 MiB and secret/download/content-shaped fields are redacted; flags: --folderId
- Entities
  - entities get-brief - Generate a brief for a CRM entity (/v2/entities/get-brief) [intent=direct_read availability=implemented]; risk: bounded Gong JSON read; response is limited to 1 MiB and secret/download/content-shaped fields are redacted; flags: --workspaceId, --briefName, --crmEntityType, --crmEntityId, --timePeriod, --fromDateTime, --toDateTime
  - entities ask - Ask about an entity (/v2/entities/ask-entity) [intent=direct_read availability=implemented]; risk: bounded Gong JSON read; response is limited to 1 MiB and secret/download/content-shaped fields are redacted; flags: --workspaceId, --crmEntityType, --crmEntityId, --timePeriod, --fromDateTime, --toDateTime, --question
- Data Privacy
  - privacy find-phone - Retrieve all references to a phone number. (/v2/data-privacy/data-for-phone-number) [intent=direct_read availability=implemented]; risk: bounded Gong JSON read; response is limited to 1 MiB and secret/download/content-shaped fields are redacted; flags: --phoneNumber
  - privacy find-email - Retrieve all references to an email address. (/v2/data-privacy/data-for-email-address) [intent=direct_read availability=implemented]; risk: bounded Gong JSON read; response is limited to 1 MiB and secret/download/content-shaped fields are redacted; flags: --emailAddress
  - privacy erase-phone - destructive Gong data privacy erasure; requires reverse ETL plan, preview, approval, and destructive confirmation [intent=reverse_etl availability=partial write=purge_phone_number]; approval: Use reverse ETL plan -> preview -> approval -> execute. Connector command execution is metadata-only for complex object/array records; use typed reverse ETL records.; risk: critical: destructive Gong data privacy erasure; requires reverse ETL plan, preview, approval, and destructive confirmation; notes: No raw HTTP body is accepted. Object and array payloads must come from typed reverse-ETL records validated by writes.json.; flags: --phoneNumber
  - privacy erase-email - destructive Gong data privacy erasure; requires reverse ETL plan, preview, approval, and destructive confirmation [intent=reverse_etl availability=partial write=purge_email_address]; approval: Use reverse ETL plan -> preview -> approval -> execute. Connector command execution is metadata-only for complex object/array records; use typed reverse ETL records.; risk: critical: destructive Gong data privacy erasure; requires reverse ETL plan, preview, approval, and destructive confirmation; notes: No raw HTTP body is accepted. Object and array payloads must come from typed reverse-ETL records validated by writes.json.; flags: --emailAddress
- Tasks
  - tasks update - mutates Gong API state; requires reverse ETL approval [intent=reverse_etl availability=partial write=update_task]; approval: Use reverse ETL plan -> preview -> approval -> execute. Connector command execution is metadata-only for complex object/array records; use typed reverse ETL records.; risk: medium: mutates Gong API state; requires reverse ETL approval; notes: No raw HTTP body is accepted. Object and array payloads must come from typed reverse-ETL records validated by writes.json.; flags: --taskId, --userId
  - tasks list - List user's tasks (/v2/tasks) [intent=direct_read availability=planned]; approval: none until executor support exists; no mutation is modeled; risk: bounded read-query planned; typed request body schema is declared in operations.json; notes: Blocked by operation executor support for fixed JSON POST read-query bodies, not by endpoint sensitivity alone.
- Digital Interactions
  - digital-interaction create - sends call, engagement, or digital interaction content to Gong; requires reverse ETL approval [intent=reverse_etl availability=partial write=add_digital_interaction]; approval: Use reverse ETL plan -> preview -> approval -> execute. Connector command execution is metadata-only for complex object/array records; use typed reverse ETL records.; risk: high: sends call, engagement, or digital interaction content to Gong; requires reverse ETL approval; notes: No raw HTTP body is accepted. Object and array payloads must come from typed reverse-ETL records validated by writes.json.; flags: --content, --eventId, --eventType, --timestamp
- Logs
  - logs list - Retrieve logs data by type and time range (/v2/logs) [intent=direct_read availability=implemented]; risk: bounded Gong JSON read; response is limited to 1 MiB and secret/download/content-shaped fields are redacted; flags: --logType, --fromDateTime, --toDateTime, --cursor
- Coaching
  - coaching list - List all coaching metrics (/v2/coaching) [intent=direct_read availability=implemented]; risk: bounded Gong JSON read; response is limited to 1 MiB and secret/download/content-shaped fields are redacted; flags: --workspace-id, --manager-id, --from, --to
- Outcomes
  - call-outcomes list - List Gong call outcomes as ETL records. [intent=etl availability=implemented stream=call_outcomes]
- Integration Settings
  - integration-settings update - administrative Gong settings or permissions mutation; requires reverse ETL approval and destructive confirmation [intent=reverse_etl availability=partial write=integration_settings]; approval: Use reverse ETL plan -> preview -> approval -> execute. Connector command execution is metadata-only for complex object/array records; use typed reverse ETL records.; risk: high: administrative Gong settings or permissions mutation; requires reverse ETL approval and destructive confirmation; notes: No raw HTTP body is accepted. Object and array payloads must come from typed reverse-ETL records validated by writes.json.; flags: --integrationTypeSettings
- Help topics:
  - gong-auth - Use Gong access key and access key secret via credentials; never pass secrets in command text.
  - gong-writes - Gong mutations are typed reverse-ETL actions with plan, preview, approval, execute gates.
  - gong-direct-read - Gong direct reads are bounded JSON GET operations with redaction for secret/download/content-shaped fields.

## Commands

### Inspect as a manual

```bash
pm connectors inspect gong
```

### Inspect as structured JSON

```bash
pm connectors inspect gong --json
```

## Agent Rules

- Run pm connectors inspect gong before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
