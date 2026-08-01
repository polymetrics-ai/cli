# pm connectors inspect linear

```text
NAME
  pm connectors inspect linear - Linear connector manual

SYNOPSIS
  pm connectors inspect linear
  pm connectors inspect linear --json
  pm credentials add <name> --connector linear [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Linear GraphQL list/connection data and exposes typed fixed GraphQL reverse-ETL mutations where connector-local schemas can do so without raw GraphQL passthrough.

ICON
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics

CAPABILITIES
  check=false catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  auth_type
  base_url
  access_token (secret)
  api_key (secret)

ETL STREAMS
  issues:
    primary key: id
    cursor: updatedAt
    fields: __typename(), archivedAt(), createdAt(), description(), id(), identifier(), number(), title(), updatedAt(), url()
  projects:
    primary key: id
    cursor: updatedAt
    fields: __typename(), archivedAt(), createdAt(), description(), id(), name(), state(), updatedAt(), url()
  teams:
    primary key: id
    cursor: updatedAt
    fields: __typename(), archivedAt(), createdAt(), description(), displayName(), id(), key(), name(), updatedAt()
  users:
    primary key: id
    cursor: updatedAt
    fields: __typename(), archivedAt(), createdAt(), description(), displayName(), email(), id(), name(), title(), updatedAt(), url()
  administrable_teams:
    primary key: id
    cursor: updatedAt
    fields: __typename(), archivedAt(), createdAt(), description(), displayName(), id(), key(), name(), updatedAt()
  agent_activities:
    primary key: id
    cursor: updatedAt
    fields: __typename(), archivedAt(), createdAt(), id(), updatedAt()
  agent_sessions:
    primary key: id
    cursor: updatedAt
    fields: __typename(), archivedAt(), createdAt(), id(), updatedAt(), url()
  agent_skills:
    primary key: id
    cursor: updatedAt
    fields: __typename(), archivedAt(), createdAt(), description(), id(), title(), updatedAt()
  archived_integrations:
    primary key: id
    fields: __typename(), archivedAt(), codeAccess(), enterpriseUrl(), externalOrgId(), id()
  archived_teams:
    primary key: id
    cursor: updatedAt
    fields: __typename(), archivedAt(), createdAt(), description(), displayName(), id(), key(), name(), updatedAt()
  attachments:
    primary key: id
    cursor: updatedAt
    fields: __typename(), archivedAt(), createdAt(), id(), title(), updatedAt(), url()
  attachments_for_url:
    primary key: id
    cursor: updatedAt
    fields: __typename(), archivedAt(), createdAt(), id(), title(), updatedAt(), url()
  audit_entries:
    primary key: id
    cursor: updatedAt
    fields: __typename(), archivedAt(), createdAt(), id(), type(), updatedAt()
  audit_entry_types:
    primary key: __typename
    fields: __typename(), description(), type()
  authentication_sessions:
    primary key: id
    cursor: updatedAt
    fields: __typename(), createdAt(), id(), name(), updatedAt()
  comments:
    primary key: id
    cursor: updatedAt
    fields: __typename(), archivedAt(), createdAt(), id(), updatedAt(), url()
  custom_views:
    primary key: id
    cursor: updatedAt
    fields: __typename(), archivedAt(), createdAt(), description(), id(), name(), updatedAt()
  customer_needs:
    primary key: id
    cursor: updatedAt
    fields: __typename(), archivedAt(), createdAt(), id(), updatedAt(), url()
  customer_statuses:
    primary key: id
    cursor: updatedAt
    fields: __typename(), archivedAt(), createdAt(), description(), displayName(), id(), name(), updatedAt()
  customer_tiers:
    primary key: id
    cursor: updatedAt
    fields: __typename(), archivedAt(), createdAt(), description(), displayName(), id(), name(), updatedAt()
  customers:
    primary key: id
    cursor: updatedAt
    fields: __typename(), archivedAt(), createdAt(), id(), name(), updatedAt(), url()
  cycles:
    primary key: id
    cursor: updatedAt
    fields: __typename(), archivedAt(), createdAt(), description(), id(), name(), number(), updatedAt()
  documents:
    primary key: id
    cursor: updatedAt
    fields: __typename(), archivedAt(), createdAt(), id(), title(), updatedAt(), url()
  emojis:
    primary key: id
    cursor: updatedAt
    fields: __typename(), archivedAt(), createdAt(), id(), name(), updatedAt(), url()
  external_users:
    primary key: id
    cursor: updatedAt
    fields: __typename(), archivedAt(), createdAt(), displayName(), email(), id(), name(), updatedAt()
  failures_for_oauth_webhooks:
    primary key: id
    cursor: createdAt
    fields: __typename(), createdAt(), id(), url()
  favorites:
    primary key: id
    cursor: updatedAt
    fields: __typename(), archivedAt(), createdAt(), id(), title(), type(), updatedAt(), url()
  initiative_labels:
    primary key: id
    cursor: updatedAt
    fields: __typename(), archivedAt(), createdAt(), description(), id(), name(), updatedAt()
  initiative_relations:
    primary key: id
    cursor: updatedAt
    fields: __typename(), archivedAt(), createdAt(), id(), updatedAt()
  initiative_to_projects:
    primary key: id
    cursor: updatedAt
    fields: __typename(), archivedAt(), createdAt(), id(), updatedAt()
  initiative_updates:
    primary key: id
    cursor: updatedAt
    fields: __typename(), archivedAt(), createdAt(), id(), updatedAt(), url()
  initiatives:
    primary key: id
    cursor: updatedAt
    fields: __typename(), archivedAt(), createdAt(), description(), id(), identifier(), name(), updatedAt(), url()
  integration_templates:
    primary key: id
    cursor: updatedAt
    fields: __typename(), archivedAt(), createdAt(), id(), updatedAt()
  integrations:
    primary key: id
    cursor: updatedAt
    fields: __typename(), archivedAt(), createdAt(), id(), updatedAt()
  issue_figma_file_key_search:
    primary key: id
    cursor: updatedAt
    fields: __typename(), archivedAt(), createdAt(), description(), id(), identifier(), number(), title(), updatedAt(), url()
  issue_labels:
    primary key: id
    cursor: updatedAt
    fields: __typename(), archivedAt(), createdAt(), description(), id(), name(), updatedAt()
  issue_priority_values:
    primary key: __typename
    fields: __typename(), label(), priority()
  issue_relations:
    primary key: id
    cursor: updatedAt
    fields: __typename(), archivedAt(), createdAt(), id(), type(), updatedAt()
  issue_search:
    primary key: id
    cursor: updatedAt
    fields: __typename(), archivedAt(), createdAt(), description(), id(), identifier(), number(), title(), updatedAt(), url()
  issue_to_releases:
    primary key: id
    cursor: updatedAt
    fields: __typename(), archivedAt(), createdAt(), id(), updatedAt()
  oauth_applications:
    primary key: id
    cursor: updatedAt
    fields: __typename(), createdAt(), description(), id(), name(), updatedAt()
  organization_invites:
    primary key: id
    cursor: updatedAt
    fields: __typename(), archivedAt(), createdAt(), email(), id(), updatedAt()
  project_labels:
    primary key: id
    cursor: updatedAt
    fields: __typename(), archivedAt(), createdAt(), description(), id(), name(), updatedAt()
  project_milestones:
    primary key: id
    cursor: updatedAt
    fields: __typename(), archivedAt(), createdAt(), description(), id(), name(), updatedAt()
  project_relations:
    primary key: id
    cursor: updatedAt
    fields: __typename(), archivedAt(), createdAt(), id(), type(), updatedAt()
  project_statuses:
    primary key: id
    cursor: updatedAt
    fields: __typename(), archivedAt(), createdAt(), description(), id(), name(), updatedAt()
  project_updates:
    primary key: id
    cursor: updatedAt
    fields: __typename(), archivedAt(), createdAt(), id(), updatedAt(), url()
  recent_releases_by_access_key:
    primary key: id
    cursor: updatedAt
    fields: __typename(), archivedAt(), createdAt(), description(), id(), name(), updatedAt(), url()
  release_notes:
    primary key: id
    cursor: updatedAt
    fields: __typename(), archivedAt(), createdAt(), id(), title(), updatedAt(), url()
  release_pipelines:
    primary key: id
    cursor: updatedAt
    fields: __typename(), archivedAt(), createdAt(), id(), name(), updatedAt(), url()
  release_search:
    primary key: id
    cursor: updatedAt
    fields: __typename(), archivedAt(), createdAt(), description(), id(), name(), updatedAt(), url()
  release_stages:
    primary key: id
    cursor: updatedAt
    fields: __typename(), archivedAt(), createdAt(), id(), name(), updatedAt()
  releases:
    primary key: id
    cursor: updatedAt
    fields: __typename(), archivedAt(), createdAt(), description(), id(), name(), updatedAt(), url()
  roadmap_to_projects:
    primary key: id
    cursor: updatedAt
    fields: __typename(), archivedAt(), createdAt(), id(), updatedAt()
  roadmaps:
    primary key: id
    cursor: updatedAt
    fields: __typename(), archivedAt(), createdAt(), description(), id(), name(), updatedAt(), url()
  sla_configurations:
    primary key: id
    fields: __typename(), conditions(), id(), name(), removesSla(), sla()
  team_memberships:
    primary key: id
    cursor: updatedAt
    fields: __typename(), archivedAt(), createdAt(), id(), updatedAt()
  templates:
    primary key: id
    cursor: updatedAt
    fields: __typename(), archivedAt(), createdAt(), description(), id(), name(), type(), updatedAt()
  templates_for_integration:
    primary key: id
    cursor: updatedAt
    fields: __typename(), archivedAt(), createdAt(), description(), id(), name(), type(), updatedAt()
  time_schedules:
    primary key: id
    cursor: updatedAt
    fields: __typename(), archivedAt(), createdAt(), id(), name(), updatedAt()
  triage_responsibilities:
    primary key: id
    cursor: updatedAt
    fields: __typename(), archivedAt(), createdAt(), id(), updatedAt()
  user_sessions:
    primary key: id
    cursor: updatedAt
    fields: __typename(), createdAt(), id(), name(), updatedAt()
  webhooks:
    primary key: id
    cursor: updatedAt
    fields: __typename(), archivedAt(), createdAt(), id(), updatedAt(), url()
  workflow_states:
    primary key: id
    cursor: updatedAt
    fields: __typename(), archivedAt(), createdAt(), description(), id(), name(), type(), updatedAt()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  agent_skill_delete:
    endpoint: POST
    required fields: id
    risk: destructive Linear GraphQL mutation; requires typed confirmation
  attachment_delete:
    endpoint: POST
    required fields: id
    risk: destructive Linear GraphQL mutation; requires typed confirmation
  attachment_sync_to_slack:
    endpoint: POST
    required fields: id
    risk: typed Linear GraphQL mutation; executes only through reverse ETL approval
  comment_delete:
    endpoint: POST
    required fields: id
    risk: destructive Linear GraphQL mutation; requires typed confirmation
  comment_unresolve:
    endpoint: POST
    required fields: id
    risk: typed Linear GraphQL mutation; executes only through reverse ETL approval
  custom_view_delete:
    endpoint: POST
    required fields: id
    risk: destructive Linear GraphQL mutation; requires typed confirmation
  customer_delete:
    endpoint: POST
    required fields: id
    risk: destructive Linear GraphQL mutation; requires typed confirmation
  customer_merge:
    endpoint: POST
    required fields: sourceCustomerId, targetCustomerId
    risk: typed Linear GraphQL mutation; executes only through reverse ETL approval
  customer_need_archive:
    endpoint: POST
    required fields: id
    risk: destructive Linear GraphQL mutation; requires typed confirmation
  customer_need_unarchive:
    endpoint: POST
    required fields: id
    risk: destructive Linear GraphQL mutation; requires typed confirmation
  customer_status_delete:
    endpoint: POST
    required fields: id
    risk: destructive Linear GraphQL mutation; requires typed confirmation
  customer_tier_delete:
    endpoint: POST
    required fields: id
    risk: destructive Linear GraphQL mutation; requires typed confirmation
  customer_unsync:
    endpoint: POST
    required fields: id
    risk: destructive Linear GraphQL mutation; requires typed confirmation
  cycle_archive:
    endpoint: POST
    required fields: id
    risk: destructive Linear GraphQL mutation; requires typed confirmation
  cycle_start_upcoming_cycle_today:
    endpoint: POST
    required fields: id
    risk: typed Linear GraphQL mutation; executes only through reverse ETL approval
  document_delete:
    endpoint: POST
    required fields: id
    risk: destructive Linear GraphQL mutation; requires typed confirmation
  document_unarchive:
    endpoint: POST
    required fields: id
    risk: destructive Linear GraphQL mutation; requires typed confirmation
  email_intake_address_delete:
    endpoint: POST
    required fields: id
    risk: destructive Linear GraphQL mutation; requires typed confirmation
  email_intake_address_refresh_ses_domain_status:
    endpoint: POST
    required fields: id
    risk: typed Linear GraphQL mutation; executes only through reverse ETL approval
  email_intake_address_rotate:
    endpoint: POST
    required fields: id
    risk: destructive Linear GraphQL mutation; requires typed confirmation
  emoji_delete:
    endpoint: POST
    required fields: id
    risk: destructive Linear GraphQL mutation; requires typed confirmation
  entity_external_link_delete:
    endpoint: POST
    required fields: id
    risk: destructive Linear GraphQL mutation; requires typed confirmation
  favorite_delete:
    endpoint: POST
    required fields: id
    risk: destructive Linear GraphQL mutation; requires typed confirmation
  file_upload_dangerously_delete:
    endpoint: POST
    required fields: assetUrl
    risk: destructive Linear GraphQL mutation; requires typed confirmation
  git_automation_state_delete:
    endpoint: POST
    required fields: id
    risk: destructive Linear GraphQL mutation; requires typed confirmation
  git_automation_target_branch_delete:
    endpoint: POST
    required fields: id
    risk: destructive Linear GraphQL mutation; requires typed confirmation
  image_upload_from_url:
    endpoint: POST
    required fields: url
    risk: typed Linear GraphQL mutation; executes only through reverse ETL approval
  initiative_add_label:
    endpoint: POST
    required fields: id, labelId
    risk: typed Linear GraphQL mutation; executes only through reverse ETL approval
  initiative_archive:
    endpoint: POST
    required fields: id
    risk: destructive Linear GraphQL mutation; requires typed confirmation
  initiative_delete:
    endpoint: POST
    required fields: id
    risk: destructive Linear GraphQL mutation; requires typed confirmation
  initiative_label_delete:
    endpoint: POST
    required fields: id
    risk: destructive Linear GraphQL mutation; requires typed confirmation
  initiative_label_restore:
    endpoint: POST
    required fields: id
    risk: typed Linear GraphQL mutation; executes only through reverse ETL approval
  initiative_label_retire:
    endpoint: POST
    required fields: id
    risk: destructive Linear GraphQL mutation; requires typed confirmation
  initiative_relation_delete:
    endpoint: POST
    required fields: id
    risk: destructive Linear GraphQL mutation; requires typed confirmation
  initiative_remove_label:
    endpoint: POST
    required fields: id, labelId
    risk: destructive Linear GraphQL mutation; requires typed confirmation
  initiative_to_project_delete:
    endpoint: POST
    required fields: id
    risk: destructive Linear GraphQL mutation; requires typed confirmation
  initiative_unarchive:
    endpoint: POST
    required fields: id
    risk: destructive Linear GraphQL mutation; requires typed confirmation
  initiative_update_archive:
    endpoint: POST
    required fields: id
    risk: destructive Linear GraphQL mutation; requires typed confirmation
  initiative_update_unarchive:
    endpoint: POST
    required fields: id
    risk: destructive Linear GraphQL mutation; requires typed confirmation
  integration_archive:
    endpoint: POST
    required fields: id
    risk: destructive Linear GraphQL mutation; requires typed confirmation
  integration_git_hub_enterprise_server_connect:
    endpoint: POST
    required fields: githubUrl, organizationName
    risk: typed Linear GraphQL mutation; executes only through reverse ETL approval
  integration_github_commit_create:
    endpoint: POST
    risk: typed Linear GraphQL mutation; executes only through reverse ETL approval
  integration_github_import_refresh:
    endpoint: POST
    required fields: id
    risk: typed Linear GraphQL mutation; executes only through reverse ETL approval
  integration_github_remove_code_access:
    endpoint: POST
    required fields: integrationId
    risk: destructive Linear GraphQL mutation; requires typed confirmation
  integration_gitlab_test_connection:
    endpoint: POST
    required fields: integrationId
    risk: typed Linear GraphQL mutation; executes only through reverse ETL approval
  integration_intercom_delete:
    endpoint: POST
    risk: destructive Linear GraphQL mutation; requires typed confirmation
  integration_opsgenie_refresh_schedule_mappings:
    endpoint: POST
    risk: typed Linear GraphQL mutation; executes only through reverse ETL approval
  integration_pager_duty_refresh_schedule_mappings:
    endpoint: POST
    risk: typed Linear GraphQL mutation; executes only through reverse ETL approval
  integration_salesforce_metadata_refresh:
    endpoint: POST
    required fields: id
    risk: typed Linear GraphQL mutation; executes only through reverse ETL approval
  integration_slack_or_asks_update_slack_team_name:
    endpoint: POST
    required fields: integrationId
    risk: typed Linear GraphQL mutation; executes only through reverse ETL approval
  integration_template_delete:
    endpoint: POST
    required fields: id
    risk: destructive Linear GraphQL mutation; requires typed confirmation
  issue_add_label:
    endpoint: POST
    required fields: id, labelId
    risk: typed Linear GraphQL mutation; executes only through reverse ETL approval
  issue_description_update_from_front:
    endpoint: POST
    required fields: description, id
    risk: typed Linear GraphQL mutation; executes only through reverse ETL approval
  issue_external_sync_disable:
    endpoint: POST
    required fields: attachmentId
    risk: destructive Linear GraphQL mutation; requires typed confirmation
  issue_import_delete:
    endpoint: POST
    required fields: issueImportId
    risk: destructive Linear GraphQL mutation; requires typed confirmation
  issue_label_delete:
    endpoint: POST
    required fields: id
    risk: destructive Linear GraphQL mutation; requires typed confirmation
  issue_label_restore:
    endpoint: POST
    required fields: id
    risk: typed Linear GraphQL mutation; executes only through reverse ETL approval
  issue_label_retire:
    endpoint: POST
    required fields: id
    risk: destructive Linear GraphQL mutation; requires typed confirmation
  issue_relation_delete:
    endpoint: POST
    required fields: id
    risk: destructive Linear GraphQL mutation; requires typed confirmation
  issue_reminder:
    endpoint: POST
    required fields: id, reminderAt
    risk: typed Linear GraphQL mutation; executes only through reverse ETL approval
  issue_remove_label:
    endpoint: POST
    required fields: id, labelId
    risk: destructive Linear GraphQL mutation; requires typed confirmation
  issue_share:
    endpoint: POST
    required fields: id, userId
    risk: typed Linear GraphQL mutation; executes only through reverse ETL approval
  issue_to_release_delete:
    endpoint: POST
    required fields: id
    risk: destructive Linear GraphQL mutation; requires typed confirmation
  issue_to_release_delete_by_issue_and_release:
    endpoint: POST
    required fields: issueId, releaseId
    risk: destructive Linear GraphQL mutation; requires typed confirmation
  issue_unarchive:
    endpoint: POST
    required fields: id
    risk: destructive Linear GraphQL mutation; requires typed confirmation
  issue_unshare:
    endpoint: POST
    required fields: id, userId
    risk: destructive Linear GraphQL mutation; requires typed confirmation
  leave_organization:
    endpoint: POST
    required fields: organizationId
    risk: destructive Linear GraphQL mutation; requires typed confirmation
  logout_session:
    endpoint: POST
    required fields: sessionId
    risk: destructive Linear GraphQL mutation; requires typed confirmation
  notification_archive:
    endpoint: POST
    required fields: id
    risk: destructive Linear GraphQL mutation; requires typed confirmation
  notification_unarchive:
    endpoint: POST
    required fields: id
    risk: destructive Linear GraphQL mutation; requires typed confirmation
  oauth_application_archive:
    endpoint: POST
    required fields: id
    risk: destructive Linear GraphQL mutation; requires typed confirmation
  oauth_application_rotate_secret:
    endpoint: POST
    required fields: id
    risk: destructive Linear GraphQL mutation; requires typed confirmation
  oauth_application_rotate_webhook_secret:
    endpoint: POST
    required fields: id
    risk: destructive Linear GraphQL mutation; requires typed confirmation
  organization_cancel_delete:
    endpoint: POST
    risk: destructive Linear GraphQL mutation; requires typed confirmation
  organization_delete_challenge:
    endpoint: POST
    risk: destructive Linear GraphQL mutation; requires typed confirmation
  organization_domain_claim:
    endpoint: POST
    required fields: id
    risk: typed Linear GraphQL mutation; executes only through reverse ETL approval
  organization_domain_delete:
    endpoint: POST
    required fields: id
    risk: destructive Linear GraphQL mutation; requires typed confirmation
  organization_invite_delete:
    endpoint: POST
    required fields: id
    risk: destructive Linear GraphQL mutation; requires typed confirmation
  passkey_login_start:
    endpoint: POST
    required fields: authId
    risk: typed Linear GraphQL mutation; executes only through reverse ETL approval
  project_add_label:
    endpoint: POST
    required fields: id, labelId
    risk: typed Linear GraphQL mutation; executes only through reverse ETL approval
  project_delete:
    endpoint: POST
    required fields: id
    risk: destructive Linear GraphQL mutation; requires typed confirmation
  project_label_delete:
    endpoint: POST
    required fields: id
    risk: destructive Linear GraphQL mutation; requires typed confirmation
  project_label_restore:
    endpoint: POST
    required fields: id
    risk: typed Linear GraphQL mutation; executes only through reverse ETL approval
  project_label_retire:
    endpoint: POST
    required fields: id
    risk: destructive Linear GraphQL mutation; requires typed confirmation
  project_milestone_delete:
    endpoint: POST
    required fields: id
    risk: destructive Linear GraphQL mutation; requires typed confirmation
  project_reassign_status:
    endpoint: POST
    required fields: newProjectStatusId, originalProjectStatusId
    risk: typed Linear GraphQL mutation; executes only through reverse ETL approval
  project_relation_delete:
    endpoint: POST
    required fields: id
    risk: destructive Linear GraphQL mutation; requires typed confirmation
  project_remove_label:
    endpoint: POST
    required fields: id, labelId
    risk: destructive Linear GraphQL mutation; requires typed confirmation
  project_status_archive:
    endpoint: POST
    required fields: id
    risk: destructive Linear GraphQL mutation; requires typed confirmation
  project_status_unarchive:
    endpoint: POST
    required fields: id
    risk: destructive Linear GraphQL mutation; requires typed confirmation
  project_unarchive:
    endpoint: POST
    required fields: id
    risk: destructive Linear GraphQL mutation; requires typed confirmation
  project_update_archive:
    endpoint: POST
    required fields: id
    risk: destructive Linear GraphQL mutation; requires typed confirmation
  project_update_unarchive:
    endpoint: POST
    required fields: id
    risk: destructive Linear GraphQL mutation; requires typed confirmation
  push_subscription_delete:
    endpoint: POST
    required fields: id
    risk: destructive Linear GraphQL mutation; requires typed confirmation
  reaction_delete:
    endpoint: POST
    required fields: id
    risk: destructive Linear GraphQL mutation; requires typed confirmation
  release_archive:
    endpoint: POST
    required fields: id
    risk: destructive Linear GraphQL mutation; requires typed confirmation
  release_delete:
    endpoint: POST
    required fields: id
    risk: destructive Linear GraphQL mutation; requires typed confirmation
  release_note_delete:
    endpoint: POST
    required fields: id
    risk: destructive Linear GraphQL mutation; requires typed confirmation
  release_pipeline_archive:
    endpoint: POST
    required fields: id
    risk: destructive Linear GraphQL mutation; requires typed confirmation
  release_pipeline_delete:
    endpoint: POST
    required fields: id
    risk: destructive Linear GraphQL mutation; requires typed confirmation
  release_pipeline_unarchive:
    endpoint: POST
    required fields: id
    risk: destructive Linear GraphQL mutation; requires typed confirmation
  release_stage_archive:
    endpoint: POST
    required fields: id
    risk: destructive Linear GraphQL mutation; requires typed confirmation
  release_stage_unarchive:
    endpoint: POST
    required fields: id
    risk: destructive Linear GraphQL mutation; requires typed confirmation
  release_unarchive:
    endpoint: POST
    required fields: id
    risk: destructive Linear GraphQL mutation; requires typed confirmation
  resend_organization_invite:
    endpoint: POST
    required fields: id
    risk: typed Linear GraphQL mutation; executes only through reverse ETL approval
  resend_organization_invite_by_email:
    endpoint: POST
    required fields: email
    risk: typed Linear GraphQL mutation; executes only through reverse ETL approval
  team_cycles_delete:
    endpoint: POST
    required fields: id
    risk: destructive Linear GraphQL mutation; requires typed confirmation
  team_delete:
    endpoint: POST
    required fields: id
    risk: destructive Linear GraphQL mutation; requires typed confirmation
  team_key_delete:
    endpoint: POST
    required fields: id
    risk: destructive Linear GraphQL mutation; requires typed confirmation
  team_unarchive:
    endpoint: POST
    required fields: id
    risk: destructive Linear GraphQL mutation; requires typed confirmation
  template_delete:
    endpoint: POST
    required fields: id
    risk: destructive Linear GraphQL mutation; requires typed confirmation
  time_schedule_delete:
    endpoint: POST
    required fields: id
    risk: destructive Linear GraphQL mutation; requires typed confirmation
  time_schedule_refresh_integration_schedule:
    endpoint: POST
    required fields: id
    risk: typed Linear GraphQL mutation; executes only through reverse ETL approval
  triage_responsibility_delete:
    endpoint: POST
    required fields: id
    risk: destructive Linear GraphQL mutation; requires typed confirmation
  user_external_user_disconnect:
    endpoint: POST
    required fields: service
    risk: destructive Linear GraphQL mutation; requires typed confirmation
  user_revoke_all_sessions:
    endpoint: POST
    required fields: id
    risk: destructive Linear GraphQL mutation; requires typed confirmation
  user_revoke_session:
    endpoint: POST
    required fields: id, sessionId
    risk: destructive Linear GraphQL mutation; requires typed confirmation
  user_unlink_from_identity_provider:
    endpoint: POST
    required fields: id
    risk: destructive Linear GraphQL mutation; requires typed confirmation
  view_preferences_delete:
    endpoint: POST
    required fields: id
    risk: destructive Linear GraphQL mutation; requires typed confirmation
  webhook_delete:
    endpoint: POST
    required fields: id
    risk: destructive Linear GraphQL mutation; requires typed confirmation
  webhook_rotate_secret:
    endpoint: POST
    required fields: id
    risk: destructive Linear GraphQL mutation; requires typed confirmation
  workflow_state_archive:
    endpoint: POST
    required fields: id
    risk: destructive Linear GraphQL mutation; requires typed confirmation

SECURITY
  read risk: fixed Linear GraphQL documents against documented root Query list/connection fields
  write risk: fixed Linear GraphQL mutations; destructive/delete/archive operations require typed destructive confirmation when implemented
  approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions require typed confirmation
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

COMMAND SURFACE
  Work with Linear through connector-owned fixed GraphQL operations.
  Usage: pm linear <command> [flags]
  Source CLI: Linear GraphQL API (https://developers.linear.app/docs/graphql/working-with-the-graphql-api)
  Global flags:
    --json (boolean): Write machine-readable JSON output.
    --connection (string): Use a saved Linear connector credential.: maps_to=connection
  Read/list commands
    administrable teams list - List Linear administrableTeams. [intent=etl availability=implemented stream=administrable_teams]
    agent activities list - List Linear agentActivities. [intent=etl availability=implemented stream=agent_activities]
    agent sessions list - List Linear agentSessions. [intent=etl availability=implemented stream=agent_sessions]
    agent skills list - List Linear agentSkills. [intent=etl availability=implemented stream=agent_skills]
    agent skill delete - Run Linear mutation agentSkillDelete. [intent=reverse_etl availability=implemented write=agent_skill_delete]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    agent activity events - Planned bounded Linear Subscription.agentActivityCreated command. [intent=direct_read availability=planned]; notes: Planned only: fixed GraphQL direct operation execution is not enabled for Linear in this connector-local slice.
    archived integrations list - List Linear archivedIntegrations. [intent=etl availability=implemented stream=archived_integrations]
    archived teams list - List Linear archivedTeams. [intent=etl availability=implemented stream=archived_teams]
    attachments for url list - List Linear attachmentsForURL. [intent=etl availability=implemented stream=attachments_for_url]; flags: --url
    attachments list - List Linear attachments. [intent=etl availability=implemented stream=attachments]
    audit entries list - List Linear auditEntries. [intent=etl availability=implemented stream=audit_entries]
    audit entry types list - List Linear auditEntryTypes. [intent=etl availability=implemented stream=audit_entry_types]
    authentication sessions list - List Linear authenticationSessions. [intent=etl availability=implemented stream=authentication_sessions]
    comments list - List Linear comments. [intent=etl availability=implemented stream=comments]
    custom views list - List Linear customViews. [intent=etl availability=implemented stream=custom_views]
    custom view delete - Run Linear mutation customViewDelete. [intent=reverse_etl availability=implemented write=custom_view_delete]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    customer needs list - List Linear customerNeeds. [intent=etl availability=implemented stream=customer_needs]
    customer statuses list - List Linear customerStatuses. [intent=etl availability=implemented stream=customer_statuses]
    customer tiers list - List Linear customerTiers. [intent=etl availability=implemented stream=customer_tiers]
    customer delete - Run Linear mutation customerDelete. [intent=reverse_etl availability=implemented write=customer_delete]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    customer merge - Run Linear mutation customerMerge. [intent=reverse_etl availability=implemented write=customer_merge]; approval: reverse ETL plan → preview → explicit approval → execute.; risk: typed Linear GraphQL mutation; executes only through reverse ETL approval; flags: --source-customer-id, --target-customer-id
    customer need archive - Run Linear mutation customerNeedArchive. [intent=reverse_etl availability=implemented write=customer_need_archive]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    customer need unarchive - Run Linear mutation customerNeedUnarchive. [intent=reverse_etl availability=implemented write=customer_need_unarchive]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    customer status delete - Run Linear mutation customerStatusDelete. [intent=reverse_etl availability=implemented write=customer_status_delete]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    customer tier delete - Run Linear mutation customerTierDelete. [intent=reverse_etl availability=implemented write=customer_tier_delete]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    customer unsync - Run Linear mutation customerUnsync. [intent=reverse_etl availability=implemented write=customer_unsync]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    customers list - List Linear customers. [intent=etl availability=implemented stream=customers]
    cycles list - List Linear cycles. [intent=etl availability=implemented stream=cycles]
    documents list - List Linear documents. [intent=etl availability=implemented stream=documents]
    emojis list - List Linear emojis. [intent=etl availability=implemented stream=emojis]
    external users list - List Linear externalUsers. [intent=etl availability=implemented stream=external_users]
    failures for oauth webhooks list - List Linear failuresForOauthWebhooks. [intent=etl availability=implemented stream=failures_for_oauth_webhooks]; flags: --oauthClientId
    favorites list - List Linear favorites. [intent=etl availability=implemented stream=favorites]
    initiative labels list - List Linear initiativeLabels. [intent=etl availability=implemented stream=initiative_labels]
    initiative relations list - List Linear initiativeRelations. [intent=etl availability=implemented stream=initiative_relations]
    initiative to projects list - List Linear initiativeToProjects. [intent=etl availability=implemented stream=initiative_to_projects]
    initiative updates list - List Linear initiativeUpdates. [intent=etl availability=implemented stream=initiative_updates]
    initiative add label - Run Linear mutation initiativeAddLabel. [intent=reverse_etl availability=implemented write=initiative_add_label]; approval: reverse ETL plan → preview → explicit approval → execute.; risk: typed Linear GraphQL mutation; executes only through reverse ETL approval; flags: --id, --label-id
    initiative archive - Run Linear mutation initiativeArchive. [intent=reverse_etl availability=implemented write=initiative_archive]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    initiative delete - Run Linear mutation initiativeDelete. [intent=reverse_etl availability=implemented write=initiative_delete]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    initiative label delete - Run Linear mutation initiativeLabelDelete. [intent=reverse_etl availability=implemented write=initiative_label_delete]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    initiative label restore - Run Linear mutation initiativeLabelRestore. [intent=reverse_etl availability=implemented write=initiative_label_restore]; approval: reverse ETL plan → preview → explicit approval → execute.; risk: typed Linear GraphQL mutation; executes only through reverse ETL approval; flags: --id
    initiative label retire - Run Linear mutation initiativeLabelRetire. [intent=reverse_etl availability=implemented write=initiative_label_retire]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    initiative relation delete - Run Linear mutation initiativeRelationDelete. [intent=reverse_etl availability=implemented write=initiative_relation_delete]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    initiative remove label - Run Linear mutation initiativeRemoveLabel. [intent=reverse_etl availability=implemented write=initiative_remove_label]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id, --label-id
    initiative to project delete - Run Linear mutation initiativeToProjectDelete. [intent=reverse_etl availability=implemented write=initiative_to_project_delete]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    initiative unarchive - Run Linear mutation initiativeUnarchive. [intent=reverse_etl availability=implemented write=initiative_unarchive]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    initiative update archive - Run Linear mutation initiativeUpdateArchive. [intent=reverse_etl availability=implemented write=initiative_update_archive]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    initiative update unarchive - Run Linear mutation initiativeUpdateUnarchive. [intent=reverse_etl availability=implemented write=initiative_update_unarchive]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    initiatives list - List Linear initiatives. [intent=etl availability=implemented stream=initiatives]
    integration templates list - List Linear integrationTemplates. [intent=etl availability=implemented stream=integration_templates]
    integration archive - Run Linear mutation integrationArchive. [intent=reverse_etl availability=implemented write=integration_archive]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    integration git hub enterprise server connect - Run Linear mutation integrationGitHubEnterpriseServerConnect. [intent=reverse_etl availability=implemented write=integration_git_hub_enterprise_server_connect]; approval: reverse ETL plan → preview → explicit approval → execute.; risk: typed Linear GraphQL mutation; executes only through reverse ETL approval; flags: --github-url, --organization-name
    integration github commit create - Run Linear mutation integrationGithubCommitCreate. [intent=reverse_etl availability=implemented write=integration_github_commit_create]; approval: reverse ETL plan → preview → explicit approval → execute.; risk: typed Linear GraphQL mutation; executes only through reverse ETL approval
    integration github import refresh - Run Linear mutation integrationGithubImportRefresh. [intent=reverse_etl availability=implemented write=integration_github_import_refresh]; approval: reverse ETL plan → preview → explicit approval → execute.; risk: typed Linear GraphQL mutation; executes only through reverse ETL approval; flags: --id
    integration github remove code access - Run Linear mutation integrationGithubRemoveCodeAccess. [intent=reverse_etl availability=implemented write=integration_github_remove_code_access]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --integration-id
    integration gitlab test connection - Run Linear mutation integrationGitlabTestConnection. [intent=reverse_etl availability=implemented write=integration_gitlab_test_connection]; approval: reverse ETL plan → preview → explicit approval → execute.; risk: typed Linear GraphQL mutation; executes only through reverse ETL approval; flags: --integration-id
    integration intercom delete - Run Linear mutation integrationIntercomDelete. [intent=reverse_etl availability=implemented write=integration_intercom_delete]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation
    integration opsgenie refresh schedule mappings - Run Linear mutation integrationOpsgenieRefreshScheduleMappings. [intent=reverse_etl availability=implemented write=integration_opsgenie_refresh_schedule_mappings]; approval: reverse ETL plan → preview → explicit approval → execute.; risk: typed Linear GraphQL mutation; executes only through reverse ETL approval
    integration pager duty refresh schedule mappings - Run Linear mutation integrationPagerDutyRefreshScheduleMappings. [intent=reverse_etl availability=implemented write=integration_pager_duty_refresh_schedule_mappings]; approval: reverse ETL plan → preview → explicit approval → execute.; risk: typed Linear GraphQL mutation; executes only through reverse ETL approval
    integration salesforce metadata refresh - Run Linear mutation integrationSalesforceMetadataRefresh. [intent=reverse_etl availability=implemented write=integration_salesforce_metadata_refresh]; approval: reverse ETL plan → preview → explicit approval → execute.; risk: typed Linear GraphQL mutation; executes only through reverse ETL approval; flags: --id
    integration slack or asks update slack team name - Run Linear mutation integrationSlackOrAsksUpdateSlackTeamName. [intent=reverse_etl availability=implemented write=integration_slack_or_asks_update_slack_team_name]; approval: reverse ETL plan → preview → explicit approval → execute.; risk: typed Linear GraphQL mutation; executes only through reverse ETL approval; flags: --integration-id
    integration template delete - Run Linear mutation integrationTemplateDelete. [intent=reverse_etl availability=implemented write=integration_template_delete]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    integrations list - List Linear integrations. [intent=etl availability=implemented stream=integrations]
    issue figma file key search list - List Linear issueFigmaFileKeySearch. [intent=etl availability=implemented stream=issue_figma_file_key_search]; flags: --fileKey
    issue labels list - List Linear issueLabels. [intent=etl availability=implemented stream=issue_labels]
    issue priority values list - List Linear issuePriorityValues. [intent=etl availability=implemented stream=issue_priority_values]
    issue relations list - List Linear issueRelations. [intent=etl availability=implemented stream=issue_relations]
    issue search list - List Linear issueSearch. [intent=etl availability=implemented stream=issue_search]
    issue to releases list - List Linear issueToReleases. [intent=etl availability=implemented stream=issue_to_releases]
    issue add label - Run Linear mutation issueAddLabel. [intent=reverse_etl availability=implemented write=issue_add_label]; approval: reverse ETL plan → preview → explicit approval → execute.; risk: typed Linear GraphQL mutation; executes only through reverse ETL approval; flags: --id, --label-id
    issue description update from front - Run Linear mutation issueDescriptionUpdateFromFront. [intent=reverse_etl availability=implemented write=issue_description_update_from_front]; approval: reverse ETL plan → preview → explicit approval → execute.; risk: typed Linear GraphQL mutation; executes only through reverse ETL approval; flags: --description, --id
    issue external sync disable - Run Linear mutation issueExternalSyncDisable. [intent=reverse_etl availability=implemented write=issue_external_sync_disable]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --attachment-id
    issue import delete - Run Linear mutation issueImportDelete. [intent=reverse_etl availability=implemented write=issue_import_delete]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --issue-import-id
    issue label delete - Run Linear mutation issueLabelDelete. [intent=reverse_etl availability=implemented write=issue_label_delete]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    issue label restore - Run Linear mutation issueLabelRestore. [intent=reverse_etl availability=implemented write=issue_label_restore]; approval: reverse ETL plan → preview → explicit approval → execute.; risk: typed Linear GraphQL mutation; executes only through reverse ETL approval; flags: --id
    issue label retire - Run Linear mutation issueLabelRetire. [intent=reverse_etl availability=implemented write=issue_label_retire]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    issue relation delete - Run Linear mutation issueRelationDelete. [intent=reverse_etl availability=implemented write=issue_relation_delete]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    issue reminder - Run Linear mutation issueReminder. [intent=reverse_etl availability=implemented write=issue_reminder]; approval: reverse ETL plan → preview → explicit approval → execute.; risk: typed Linear GraphQL mutation; executes only through reverse ETL approval; flags: --id, --reminder-at
    issue remove label - Run Linear mutation issueRemoveLabel. [intent=reverse_etl availability=implemented write=issue_remove_label]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id, --label-id
    issue share - Run Linear mutation issueShare. [intent=reverse_etl availability=implemented write=issue_share]; approval: reverse ETL plan → preview → explicit approval → execute.; risk: typed Linear GraphQL mutation; executes only through reverse ETL approval; flags: --id, --user-id
    issue to release delete - Run Linear mutation issueToReleaseDelete. [intent=reverse_etl availability=implemented write=issue_to_release_delete]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    issue to release delete by issue and release - Run Linear mutation issueToReleaseDeleteByIssueAndRelease. [intent=reverse_etl availability=implemented write=issue_to_release_delete_by_issue_and_release]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --issue-id, --release-id
    issue unarchive - Run Linear mutation issueUnarchive. [intent=reverse_etl availability=implemented write=issue_unarchive]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    issue unshare - Run Linear mutation issueUnshare. [intent=reverse_etl availability=implemented write=issue_unshare]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id, --user-id
    issues list - List Linear issues. [intent=etl availability=implemented stream=issues]
    oauth applications list - List Linear oauthApplications. [intent=etl availability=implemented stream=oauth_applications]
    oauth application archive - Run Linear mutation oauthApplicationArchive. [intent=reverse_etl availability=implemented write=oauth_application_archive]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    oauth application rotate secret - Run Linear mutation oauthApplicationRotateSecret. [intent=reverse_etl availability=implemented write=oauth_application_rotate_secret]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    oauth application rotate webhook secret - Run Linear mutation oauthApplicationRotateWebhookSecret. [intent=reverse_etl availability=implemented write=oauth_application_rotate_webhook_secret]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    organization invites list - List Linear organizationInvites. [intent=etl availability=implemented stream=organization_invites]
    organization cancel delete - Run Linear mutation organizationCancelDelete. [intent=reverse_etl availability=implemented write=organization_cancel_delete]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation
    organization delete challenge - Run Linear mutation organizationDeleteChallenge. [intent=reverse_etl availability=implemented write=organization_delete_challenge]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation
    organization domain claim - Run Linear mutation organizationDomainClaim. [intent=reverse_etl availability=implemented write=organization_domain_claim]; approval: reverse ETL plan → preview → explicit approval → execute.; risk: typed Linear GraphQL mutation; executes only through reverse ETL approval; flags: --id
    organization domain delete - Run Linear mutation organizationDomainDelete. [intent=reverse_etl availability=implemented write=organization_domain_delete]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    organization invite delete - Run Linear mutation organizationInviteDelete. [intent=reverse_etl availability=implemented write=organization_invite_delete]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    project labels list - List Linear projectLabels. [intent=etl availability=implemented stream=project_labels]
    project milestones list - List Linear projectMilestones. [intent=etl availability=implemented stream=project_milestones]
    project relations list - List Linear projectRelations. [intent=etl availability=implemented stream=project_relations]
    project statuses list - List Linear projectStatuses. [intent=etl availability=implemented stream=project_statuses]
    project updates list - List Linear projectUpdates. [intent=etl availability=implemented stream=project_updates]
    project add label - Run Linear mutation projectAddLabel. [intent=reverse_etl availability=implemented write=project_add_label]; approval: reverse ETL plan → preview → explicit approval → execute.; risk: typed Linear GraphQL mutation; executes only through reverse ETL approval; flags: --id, --label-id
    project delete - Run Linear mutation projectDelete. [intent=reverse_etl availability=implemented write=project_delete]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    project label delete - Run Linear mutation projectLabelDelete. [intent=reverse_etl availability=implemented write=project_label_delete]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    project label restore - Run Linear mutation projectLabelRestore. [intent=reverse_etl availability=implemented write=project_label_restore]; approval: reverse ETL plan → preview → explicit approval → execute.; risk: typed Linear GraphQL mutation; executes only through reverse ETL approval; flags: --id
    project label retire - Run Linear mutation projectLabelRetire. [intent=reverse_etl availability=implemented write=project_label_retire]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    project milestone delete - Run Linear mutation projectMilestoneDelete. [intent=reverse_etl availability=implemented write=project_milestone_delete]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    project reassign status - Run Linear mutation projectReassignStatus. [intent=reverse_etl availability=implemented write=project_reassign_status]; approval: reverse ETL plan → preview → explicit approval → execute.; risk: typed Linear GraphQL mutation; executes only through reverse ETL approval; flags: --new-project-status-id, --original-project-status-id
    project relation delete - Run Linear mutation projectRelationDelete. [intent=reverse_etl availability=implemented write=project_relation_delete]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    project remove label - Run Linear mutation projectRemoveLabel. [intent=reverse_etl availability=implemented write=project_remove_label]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id, --label-id
    project status archive - Run Linear mutation projectStatusArchive. [intent=reverse_etl availability=implemented write=project_status_archive]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    project status unarchive - Run Linear mutation projectStatusUnarchive. [intent=reverse_etl availability=implemented write=project_status_unarchive]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    project unarchive - Run Linear mutation projectUnarchive. [intent=reverse_etl availability=implemented write=project_unarchive]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    project update archive - Run Linear mutation projectUpdateArchive. [intent=reverse_etl availability=implemented write=project_update_archive]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    project update unarchive - Run Linear mutation projectUpdateUnarchive. [intent=reverse_etl availability=implemented write=project_update_unarchive]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    projects list - List Linear projects. [intent=etl availability=implemented stream=projects]
    recent releases by access key list - List Linear recentReleasesByAccessKey. [intent=etl availability=implemented stream=recent_releases_by_access_key]
    release notes list - List Linear releaseNotes. [intent=etl availability=implemented stream=release_notes]
    release pipelines list - List Linear releasePipelines. [intent=etl availability=implemented stream=release_pipelines]
    release search list - List Linear releaseSearch. [intent=etl availability=implemented stream=release_search]
    release stages list - List Linear releaseStages. [intent=etl availability=implemented stream=release_stages]
    release archive - Run Linear mutation releaseArchive. [intent=reverse_etl availability=implemented write=release_archive]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    release delete - Run Linear mutation releaseDelete. [intent=reverse_etl availability=implemented write=release_delete]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    release note delete - Run Linear mutation releaseNoteDelete. [intent=reverse_etl availability=implemented write=release_note_delete]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    release pipeline archive - Run Linear mutation releasePipelineArchive. [intent=reverse_etl availability=implemented write=release_pipeline_archive]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    release pipeline delete - Run Linear mutation releasePipelineDelete. [intent=reverse_etl availability=implemented write=release_pipeline_delete]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    release pipeline unarchive - Run Linear mutation releasePipelineUnarchive. [intent=reverse_etl availability=implemented write=release_pipeline_unarchive]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    release stage archive - Run Linear mutation releaseStageArchive. [intent=reverse_etl availability=implemented write=release_stage_archive]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    release stage unarchive - Run Linear mutation releaseStageUnarchive. [intent=reverse_etl availability=implemented write=release_stage_unarchive]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    release unarchive - Run Linear mutation releaseUnarchive. [intent=reverse_etl availability=implemented write=release_unarchive]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    releases list - List Linear releases. [intent=etl availability=implemented stream=releases]
    roadmap to projects list - List Linear roadmapToProjects. [intent=etl availability=implemented stream=roadmap_to_projects]
    roadmaps list - List Linear roadmaps. [intent=etl availability=implemented stream=roadmaps]
    sla configurations list - List Linear slaConfigurations. [intent=etl availability=implemented stream=sla_configurations]; flags: --teamId
    team memberships list - List Linear teamMemberships. [intent=etl availability=implemented stream=team_memberships]
    team cycles delete - Run Linear mutation teamCyclesDelete. [intent=reverse_etl availability=implemented write=team_cycles_delete]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    team delete - Run Linear mutation teamDelete. [intent=reverse_etl availability=implemented write=team_delete]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    team key delete - Run Linear mutation teamKeyDelete. [intent=reverse_etl availability=implemented write=team_key_delete]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    team unarchive - Run Linear mutation teamUnarchive. [intent=reverse_etl availability=implemented write=team_unarchive]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    teams list - List Linear teams. [intent=etl availability=implemented stream=teams]
    templates for integration list - List Linear templatesForIntegration. [intent=etl availability=implemented stream=templates_for_integration]; flags: --integrationType
    templates list - List Linear templates. [intent=etl availability=implemented stream=templates]
    time schedules list - List Linear timeSchedules. [intent=etl availability=implemented stream=time_schedules]
    time schedule delete - Run Linear mutation timeScheduleDelete. [intent=reverse_etl availability=implemented write=time_schedule_delete]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    time schedule refresh integration schedule - Run Linear mutation timeScheduleRefreshIntegrationSchedule. [intent=reverse_etl availability=implemented write=time_schedule_refresh_integration_schedule]; approval: reverse ETL plan → preview → explicit approval → execute.; risk: typed Linear GraphQL mutation; executes only through reverse ETL approval; flags: --id
    triage responsibilities list - List Linear triageResponsibilities. [intent=etl availability=implemented stream=triage_responsibilities]
    triage responsibility delete - Run Linear mutation triageResponsibilityDelete. [intent=reverse_etl availability=implemented write=triage_responsibility_delete]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    user sessions list - List Linear userSessions. [intent=etl availability=implemented stream=user_sessions]; flags: --id
    user external user disconnect - Run Linear mutation userExternalUserDisconnect. [intent=reverse_etl availability=implemented write=user_external_user_disconnect]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --service
    user revoke all sessions - Run Linear mutation userRevokeAllSessions. [intent=reverse_etl availability=implemented write=user_revoke_all_sessions]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    user revoke session - Run Linear mutation userRevokeSession. [intent=reverse_etl availability=implemented write=user_revoke_session]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id, --session-id
    user unlink from identity provider - Run Linear mutation userUnlinkFromIdentityProvider. [intent=reverse_etl availability=implemented write=user_unlink_from_identity_provider]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    users list - List Linear users. [intent=etl availability=implemented stream=users]
    webhooks list - List Linear webhooks. [intent=etl availability=implemented stream=webhooks]
  Typed reverse-ETL mutation commands
    agent activities list - List Linear agentActivities. [intent=etl availability=implemented stream=agent_activities]
    agent sessions list - List Linear agentSessions. [intent=etl availability=implemented stream=agent_sessions]
    agent skills list - List Linear agentSkills. [intent=etl availability=implemented stream=agent_skills]
    agent skill delete - Run Linear mutation agentSkillDelete. [intent=reverse_etl availability=implemented write=agent_skill_delete]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    agent activity events - Planned bounded Linear Subscription.agentActivityCreated command. [intent=direct_read availability=planned]; notes: Planned only: fixed GraphQL direct operation execution is not enabled for Linear in this connector-local slice.
    attachment delete - Run Linear mutation attachmentDelete. [intent=reverse_etl availability=implemented write=attachment_delete]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    attachment sync to slack - Run Linear mutation attachmentSyncToSlack. [intent=reverse_etl availability=implemented write=attachment_sync_to_slack]; approval: reverse ETL plan → preview → explicit approval → execute.; risk: typed Linear GraphQL mutation; executes only through reverse ETL approval; flags: --id
    attachment get - Planned bounded Linear Query.attachment command. [intent=direct_read availability=planned]; notes: Planned only: fixed GraphQL direct operation execution is not enabled for Linear in this connector-local slice.
    comment delete - Run Linear mutation commentDelete. [intent=reverse_etl availability=implemented write=comment_delete]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    comment unresolve - Run Linear mutation commentUnresolve. [intent=reverse_etl availability=implemented write=comment_unresolve]; approval: reverse ETL plan → preview → explicit approval → execute.; risk: typed Linear GraphQL mutation; executes only through reverse ETL approval; flags: --id
    custom views list - List Linear customViews. [intent=etl availability=implemented stream=custom_views]
    custom view delete - Run Linear mutation customViewDelete. [intent=reverse_etl availability=implemented write=custom_view_delete]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    customer needs list - List Linear customerNeeds. [intent=etl availability=implemented stream=customer_needs]
    customer statuses list - List Linear customerStatuses. [intent=etl availability=implemented stream=customer_statuses]
    customer tiers list - List Linear customerTiers. [intent=etl availability=implemented stream=customer_tiers]
    customer delete - Run Linear mutation customerDelete. [intent=reverse_etl availability=implemented write=customer_delete]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    customer merge - Run Linear mutation customerMerge. [intent=reverse_etl availability=implemented write=customer_merge]; approval: reverse ETL plan → preview → explicit approval → execute.; risk: typed Linear GraphQL mutation; executes only through reverse ETL approval; flags: --source-customer-id, --target-customer-id
    customer need archive - Run Linear mutation customerNeedArchive. [intent=reverse_etl availability=implemented write=customer_need_archive]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    customer need unarchive - Run Linear mutation customerNeedUnarchive. [intent=reverse_etl availability=implemented write=customer_need_unarchive]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    customer status delete - Run Linear mutation customerStatusDelete. [intent=reverse_etl availability=implemented write=customer_status_delete]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    customer tier delete - Run Linear mutation customerTierDelete. [intent=reverse_etl availability=implemented write=customer_tier_delete]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    customer unsync - Run Linear mutation customerUnsync. [intent=reverse_etl availability=implemented write=customer_unsync]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    cycle archive - Run Linear mutation cycleArchive. [intent=reverse_etl availability=implemented write=cycle_archive]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    cycle start upcoming cycle today - Run Linear mutation cycleStartUpcomingCycleToday. [intent=reverse_etl availability=implemented write=cycle_start_upcoming_cycle_today]; approval: reverse ETL plan → preview → explicit approval → execute.; risk: typed Linear GraphQL mutation; executes only through reverse ETL approval; flags: --id
    document delete - Run Linear mutation documentDelete. [intent=reverse_etl availability=implemented write=document_delete]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    document unarchive - Run Linear mutation documentUnarchive. [intent=reverse_etl availability=implemented write=document_unarchive]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    email intake address delete - Run Linear mutation emailIntakeAddressDelete. [intent=reverse_etl availability=implemented write=email_intake_address_delete]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    email intake address refresh ses domain status - Run Linear mutation emailIntakeAddressRefreshSesDomainStatus. [intent=reverse_etl availability=implemented write=email_intake_address_refresh_ses_domain_status]; approval: reverse ETL plan → preview → explicit approval → execute.; risk: typed Linear GraphQL mutation; executes only through reverse ETL approval; flags: --id
    email intake address rotate - Run Linear mutation emailIntakeAddressRotate. [intent=reverse_etl availability=implemented write=email_intake_address_rotate]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    emoji delete - Run Linear mutation emojiDelete. [intent=reverse_etl availability=implemented write=emoji_delete]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    entity external link delete - Run Linear mutation entityExternalLinkDelete. [intent=reverse_etl availability=implemented write=entity_external_link_delete]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    favorite delete - Run Linear mutation favoriteDelete. [intent=reverse_etl availability=implemented write=favorite_delete]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    file upload dangerously delete - Run Linear mutation fileUploadDangerouslyDelete. [intent=reverse_etl availability=implemented write=file_upload_dangerously_delete]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --asset-url
    git automation state delete - Run Linear mutation gitAutomationStateDelete. [intent=reverse_etl availability=implemented write=git_automation_state_delete]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    git automation target branch delete - Run Linear mutation gitAutomationTargetBranchDelete. [intent=reverse_etl availability=implemented write=git_automation_target_branch_delete]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    initiative labels list - List Linear initiativeLabels. [intent=etl availability=implemented stream=initiative_labels]
    initiative relations list - List Linear initiativeRelations. [intent=etl availability=implemented stream=initiative_relations]
    initiative to projects list - List Linear initiativeToProjects. [intent=etl availability=implemented stream=initiative_to_projects]
    initiative updates list - List Linear initiativeUpdates. [intent=etl availability=implemented stream=initiative_updates]
    initiative add label - Run Linear mutation initiativeAddLabel. [intent=reverse_etl availability=implemented write=initiative_add_label]; approval: reverse ETL plan → preview → explicit approval → execute.; risk: typed Linear GraphQL mutation; executes only through reverse ETL approval; flags: --id, --label-id
    initiative archive - Run Linear mutation initiativeArchive. [intent=reverse_etl availability=implemented write=initiative_archive]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    initiative delete - Run Linear mutation initiativeDelete. [intent=reverse_etl availability=implemented write=initiative_delete]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    initiative label delete - Run Linear mutation initiativeLabelDelete. [intent=reverse_etl availability=implemented write=initiative_label_delete]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    initiative label restore - Run Linear mutation initiativeLabelRestore. [intent=reverse_etl availability=implemented write=initiative_label_restore]; approval: reverse ETL plan → preview → explicit approval → execute.; risk: typed Linear GraphQL mutation; executes only through reverse ETL approval; flags: --id
    initiative label retire - Run Linear mutation initiativeLabelRetire. [intent=reverse_etl availability=implemented write=initiative_label_retire]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    initiative relation delete - Run Linear mutation initiativeRelationDelete. [intent=reverse_etl availability=implemented write=initiative_relation_delete]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    initiative remove label - Run Linear mutation initiativeRemoveLabel. [intent=reverse_etl availability=implemented write=initiative_remove_label]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id, --label-id
    initiative to project delete - Run Linear mutation initiativeToProjectDelete. [intent=reverse_etl availability=implemented write=initiative_to_project_delete]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    initiative unarchive - Run Linear mutation initiativeUnarchive. [intent=reverse_etl availability=implemented write=initiative_unarchive]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    initiative update archive - Run Linear mutation initiativeUpdateArchive. [intent=reverse_etl availability=implemented write=initiative_update_archive]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    initiative update unarchive - Run Linear mutation initiativeUpdateUnarchive. [intent=reverse_etl availability=implemented write=initiative_update_unarchive]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    integration templates list - List Linear integrationTemplates. [intent=etl availability=implemented stream=integration_templates]
    integration archive - Run Linear mutation integrationArchive. [intent=reverse_etl availability=implemented write=integration_archive]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    integration git hub enterprise server connect - Run Linear mutation integrationGitHubEnterpriseServerConnect. [intent=reverse_etl availability=implemented write=integration_git_hub_enterprise_server_connect]; approval: reverse ETL plan → preview → explicit approval → execute.; risk: typed Linear GraphQL mutation; executes only through reverse ETL approval; flags: --github-url, --organization-name
    integration github commit create - Run Linear mutation integrationGithubCommitCreate. [intent=reverse_etl availability=implemented write=integration_github_commit_create]; approval: reverse ETL plan → preview → explicit approval → execute.; risk: typed Linear GraphQL mutation; executes only through reverse ETL approval
    integration github import refresh - Run Linear mutation integrationGithubImportRefresh. [intent=reverse_etl availability=implemented write=integration_github_import_refresh]; approval: reverse ETL plan → preview → explicit approval → execute.; risk: typed Linear GraphQL mutation; executes only through reverse ETL approval; flags: --id
    integration github remove code access - Run Linear mutation integrationGithubRemoveCodeAccess. [intent=reverse_etl availability=implemented write=integration_github_remove_code_access]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --integration-id
    integration gitlab test connection - Run Linear mutation integrationGitlabTestConnection. [intent=reverse_etl availability=implemented write=integration_gitlab_test_connection]; approval: reverse ETL plan → preview → explicit approval → execute.; risk: typed Linear GraphQL mutation; executes only through reverse ETL approval; flags: --integration-id
    integration intercom delete - Run Linear mutation integrationIntercomDelete. [intent=reverse_etl availability=implemented write=integration_intercom_delete]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation
    integration opsgenie refresh schedule mappings - Run Linear mutation integrationOpsgenieRefreshScheduleMappings. [intent=reverse_etl availability=implemented write=integration_opsgenie_refresh_schedule_mappings]; approval: reverse ETL plan → preview → explicit approval → execute.; risk: typed Linear GraphQL mutation; executes only through reverse ETL approval
    integration pager duty refresh schedule mappings - Run Linear mutation integrationPagerDutyRefreshScheduleMappings. [intent=reverse_etl availability=implemented write=integration_pager_duty_refresh_schedule_mappings]; approval: reverse ETL plan → preview → explicit approval → execute.; risk: typed Linear GraphQL mutation; executes only through reverse ETL approval
    integration salesforce metadata refresh - Run Linear mutation integrationSalesforceMetadataRefresh. [intent=reverse_etl availability=implemented write=integration_salesforce_metadata_refresh]; approval: reverse ETL plan → preview → explicit approval → execute.; risk: typed Linear GraphQL mutation; executes only through reverse ETL approval; flags: --id
    integration slack or asks update slack team name - Run Linear mutation integrationSlackOrAsksUpdateSlackTeamName. [intent=reverse_etl availability=implemented write=integration_slack_or_asks_update_slack_team_name]; approval: reverse ETL plan → preview → explicit approval → execute.; risk: typed Linear GraphQL mutation; executes only through reverse ETL approval; flags: --integration-id
    integration template delete - Run Linear mutation integrationTemplateDelete. [intent=reverse_etl availability=implemented write=integration_template_delete]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    issue figma file key search list - List Linear issueFigmaFileKeySearch. [intent=etl availability=implemented stream=issue_figma_file_key_search]; flags: --fileKey
    issue labels list - List Linear issueLabels. [intent=etl availability=implemented stream=issue_labels]
    issue priority values list - List Linear issuePriorityValues. [intent=etl availability=implemented stream=issue_priority_values]
    issue relations list - List Linear issueRelations. [intent=etl availability=implemented stream=issue_relations]
    issue search list - List Linear issueSearch. [intent=etl availability=implemented stream=issue_search]
    issue to releases list - List Linear issueToReleases. [intent=etl availability=implemented stream=issue_to_releases]
    issue add label - Run Linear mutation issueAddLabel. [intent=reverse_etl availability=implemented write=issue_add_label]; approval: reverse ETL plan → preview → explicit approval → execute.; risk: typed Linear GraphQL mutation; executes only through reverse ETL approval; flags: --id, --label-id
    issue description update from front - Run Linear mutation issueDescriptionUpdateFromFront. [intent=reverse_etl availability=implemented write=issue_description_update_from_front]; approval: reverse ETL plan → preview → explicit approval → execute.; risk: typed Linear GraphQL mutation; executes only through reverse ETL approval; flags: --description, --id
    issue external sync disable - Run Linear mutation issueExternalSyncDisable. [intent=reverse_etl availability=implemented write=issue_external_sync_disable]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --attachment-id
    issue import delete - Run Linear mutation issueImportDelete. [intent=reverse_etl availability=implemented write=issue_import_delete]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --issue-import-id
    issue label delete - Run Linear mutation issueLabelDelete. [intent=reverse_etl availability=implemented write=issue_label_delete]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    issue label restore - Run Linear mutation issueLabelRestore. [intent=reverse_etl availability=implemented write=issue_label_restore]; approval: reverse ETL plan → preview → explicit approval → execute.; risk: typed Linear GraphQL mutation; executes only through reverse ETL approval; flags: --id
    issue label retire - Run Linear mutation issueLabelRetire. [intent=reverse_etl availability=implemented write=issue_label_retire]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    issue relation delete - Run Linear mutation issueRelationDelete. [intent=reverse_etl availability=implemented write=issue_relation_delete]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    issue reminder - Run Linear mutation issueReminder. [intent=reverse_etl availability=implemented write=issue_reminder]; approval: reverse ETL plan → preview → explicit approval → execute.; risk: typed Linear GraphQL mutation; executes only through reverse ETL approval; flags: --id, --reminder-at
    issue remove label - Run Linear mutation issueRemoveLabel. [intent=reverse_etl availability=implemented write=issue_remove_label]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id, --label-id
    issue share - Run Linear mutation issueShare. [intent=reverse_etl availability=implemented write=issue_share]; approval: reverse ETL plan → preview → explicit approval → execute.; risk: typed Linear GraphQL mutation; executes only through reverse ETL approval; flags: --id, --user-id
    issue to release delete - Run Linear mutation issueToReleaseDelete. [intent=reverse_etl availability=implemented write=issue_to_release_delete]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    issue to release delete by issue and release - Run Linear mutation issueToReleaseDeleteByIssueAndRelease. [intent=reverse_etl availability=implemented write=issue_to_release_delete_by_issue_and_release]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --issue-id, --release-id
    issue unarchive - Run Linear mutation issueUnarchive. [intent=reverse_etl availability=implemented write=issue_unarchive]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    issue unshare - Run Linear mutation issueUnshare. [intent=reverse_etl availability=implemented write=issue_unshare]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id, --user-id
    leave organization - Run Linear mutation leaveOrganization. [intent=reverse_etl availability=implemented write=leave_organization]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --organization-id
    logout session - Run Linear mutation logoutSession. [intent=reverse_etl availability=implemented write=logout_session]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --session-id
    notification archive - Run Linear mutation notificationArchive. [intent=reverse_etl availability=implemented write=notification_archive]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    notification unarchive - Run Linear mutation notificationUnarchive. [intent=reverse_etl availability=implemented write=notification_unarchive]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    oauth applications list - List Linear oauthApplications. [intent=etl availability=implemented stream=oauth_applications]
    oauth application archive - Run Linear mutation oauthApplicationArchive. [intent=reverse_etl availability=implemented write=oauth_application_archive]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    oauth application rotate secret - Run Linear mutation oauthApplicationRotateSecret. [intent=reverse_etl availability=implemented write=oauth_application_rotate_secret]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    oauth application rotate webhook secret - Run Linear mutation oauthApplicationRotateWebhookSecret. [intent=reverse_etl availability=implemented write=oauth_application_rotate_webhook_secret]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    organization invites list - List Linear organizationInvites. [intent=etl availability=implemented stream=organization_invites]
    organization cancel delete - Run Linear mutation organizationCancelDelete. [intent=reverse_etl availability=implemented write=organization_cancel_delete]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation
    organization delete challenge - Run Linear mutation organizationDeleteChallenge. [intent=reverse_etl availability=implemented write=organization_delete_challenge]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation
    organization domain claim - Run Linear mutation organizationDomainClaim. [intent=reverse_etl availability=implemented write=organization_domain_claim]; approval: reverse ETL plan → preview → explicit approval → execute.; risk: typed Linear GraphQL mutation; executes only through reverse ETL approval; flags: --id
    organization domain delete - Run Linear mutation organizationDomainDelete. [intent=reverse_etl availability=implemented write=organization_domain_delete]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    organization invite delete - Run Linear mutation organizationInviteDelete. [intent=reverse_etl availability=implemented write=organization_invite_delete]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    project labels list - List Linear projectLabels. [intent=etl availability=implemented stream=project_labels]
    project milestones list - List Linear projectMilestones. [intent=etl availability=implemented stream=project_milestones]
    project relations list - List Linear projectRelations. [intent=etl availability=implemented stream=project_relations]
    project statuses list - List Linear projectStatuses. [intent=etl availability=implemented stream=project_statuses]
    project updates list - List Linear projectUpdates. [intent=etl availability=implemented stream=project_updates]
    project add label - Run Linear mutation projectAddLabel. [intent=reverse_etl availability=implemented write=project_add_label]; approval: reverse ETL plan → preview → explicit approval → execute.; risk: typed Linear GraphQL mutation; executes only through reverse ETL approval; flags: --id, --label-id
    project delete - Run Linear mutation projectDelete. [intent=reverse_etl availability=implemented write=project_delete]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    project label delete - Run Linear mutation projectLabelDelete. [intent=reverse_etl availability=implemented write=project_label_delete]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    project label restore - Run Linear mutation projectLabelRestore. [intent=reverse_etl availability=implemented write=project_label_restore]; approval: reverse ETL plan → preview → explicit approval → execute.; risk: typed Linear GraphQL mutation; executes only through reverse ETL approval; flags: --id
    project label retire - Run Linear mutation projectLabelRetire. [intent=reverse_etl availability=implemented write=project_label_retire]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    project milestone delete - Run Linear mutation projectMilestoneDelete. [intent=reverse_etl availability=implemented write=project_milestone_delete]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    project reassign status - Run Linear mutation projectReassignStatus. [intent=reverse_etl availability=implemented write=project_reassign_status]; approval: reverse ETL plan → preview → explicit approval → execute.; risk: typed Linear GraphQL mutation; executes only through reverse ETL approval; flags: --new-project-status-id, --original-project-status-id
    project relation delete - Run Linear mutation projectRelationDelete. [intent=reverse_etl availability=implemented write=project_relation_delete]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    project remove label - Run Linear mutation projectRemoveLabel. [intent=reverse_etl availability=implemented write=project_remove_label]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id, --label-id
    project status archive - Run Linear mutation projectStatusArchive. [intent=reverse_etl availability=implemented write=project_status_archive]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    project status unarchive - Run Linear mutation projectStatusUnarchive. [intent=reverse_etl availability=implemented write=project_status_unarchive]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    project unarchive - Run Linear mutation projectUnarchive. [intent=reverse_etl availability=implemented write=project_unarchive]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    project update archive - Run Linear mutation projectUpdateArchive. [intent=reverse_etl availability=implemented write=project_update_archive]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    project update unarchive - Run Linear mutation projectUpdateUnarchive. [intent=reverse_etl availability=implemented write=project_update_unarchive]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    push subscription delete - Run Linear mutation pushSubscriptionDelete. [intent=reverse_etl availability=implemented write=push_subscription_delete]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    reaction delete - Run Linear mutation reactionDelete. [intent=reverse_etl availability=implemented write=reaction_delete]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    release notes list - List Linear releaseNotes. [intent=etl availability=implemented stream=release_notes]
    release pipelines list - List Linear releasePipelines. [intent=etl availability=implemented stream=release_pipelines]
    release search list - List Linear releaseSearch. [intent=etl availability=implemented stream=release_search]
    release stages list - List Linear releaseStages. [intent=etl availability=implemented stream=release_stages]
    release archive - Run Linear mutation releaseArchive. [intent=reverse_etl availability=implemented write=release_archive]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    release delete - Run Linear mutation releaseDelete. [intent=reverse_etl availability=implemented write=release_delete]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    release note delete - Run Linear mutation releaseNoteDelete. [intent=reverse_etl availability=implemented write=release_note_delete]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    release pipeline archive - Run Linear mutation releasePipelineArchive. [intent=reverse_etl availability=implemented write=release_pipeline_archive]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    release pipeline delete - Run Linear mutation releasePipelineDelete. [intent=reverse_etl availability=implemented write=release_pipeline_delete]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    release pipeline unarchive - Run Linear mutation releasePipelineUnarchive. [intent=reverse_etl availability=implemented write=release_pipeline_unarchive]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    release stage archive - Run Linear mutation releaseStageArchive. [intent=reverse_etl availability=implemented write=release_stage_archive]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    release stage unarchive - Run Linear mutation releaseStageUnarchive. [intent=reverse_etl availability=implemented write=release_stage_unarchive]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    release unarchive - Run Linear mutation releaseUnarchive. [intent=reverse_etl availability=implemented write=release_unarchive]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    team memberships list - List Linear teamMemberships. [intent=etl availability=implemented stream=team_memberships]
    team cycles delete - Run Linear mutation teamCyclesDelete. [intent=reverse_etl availability=implemented write=team_cycles_delete]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    team delete - Run Linear mutation teamDelete. [intent=reverse_etl availability=implemented write=team_delete]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    team key delete - Run Linear mutation teamKeyDelete. [intent=reverse_etl availability=implemented write=team_key_delete]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    team unarchive - Run Linear mutation teamUnarchive. [intent=reverse_etl availability=implemented write=team_unarchive]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    template delete - Run Linear mutation templateDelete. [intent=reverse_etl availability=implemented write=template_delete]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    time schedules list - List Linear timeSchedules. [intent=etl availability=implemented stream=time_schedules]
    time schedule delete - Run Linear mutation timeScheduleDelete. [intent=reverse_etl availability=implemented write=time_schedule_delete]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    time schedule refresh integration schedule - Run Linear mutation timeScheduleRefreshIntegrationSchedule. [intent=reverse_etl availability=implemented write=time_schedule_refresh_integration_schedule]; approval: reverse ETL plan → preview → explicit approval → execute.; risk: typed Linear GraphQL mutation; executes only through reverse ETL approval; flags: --id
    triage responsibilities list - List Linear triageResponsibilities. [intent=etl availability=implemented stream=triage_responsibilities]
    triage responsibility delete - Run Linear mutation triageResponsibilityDelete. [intent=reverse_etl availability=implemented write=triage_responsibility_delete]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    user sessions list - List Linear userSessions. [intent=etl availability=implemented stream=user_sessions]; flags: --id
    user external user disconnect - Run Linear mutation userExternalUserDisconnect. [intent=reverse_etl availability=implemented write=user_external_user_disconnect]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --service
    user revoke all sessions - Run Linear mutation userRevokeAllSessions. [intent=reverse_etl availability=implemented write=user_revoke_all_sessions]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    user revoke session - Run Linear mutation userRevokeSession. [intent=reverse_etl availability=implemented write=user_revoke_session]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id, --session-id
    user unlink from identity provider - Run Linear mutation userUnlinkFromIdentityProvider. [intent=reverse_etl availability=implemented write=user_unlink_from_identity_provider]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    view preferences delete - Run Linear mutation viewPreferencesDelete. [intent=reverse_etl availability=implemented write=view_preferences_delete]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    webhook delete - Run Linear mutation webhookDelete. [intent=reverse_etl availability=implemented write=webhook_delete]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    webhook rotate secret - Run Linear mutation webhookRotateSecret. [intent=reverse_etl availability=implemented write=webhook_rotate_secret]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    workflow states list - List Linear workflowStates. [intent=etl availability=implemented stream=workflow_states]
    workflow state archive - Run Linear mutation workflowStateArchive. [intent=reverse_etl availability=implemented write=workflow_state_archive]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
  Planned direct/binary/changefeed commands
    attachment delete - Run Linear mutation attachmentDelete. [intent=reverse_etl availability=implemented write=attachment_delete]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    attachment sync to slack - Run Linear mutation attachmentSyncToSlack. [intent=reverse_etl availability=implemented write=attachment_sync_to_slack]; approval: reverse ETL plan → preview → explicit approval → execute.; risk: typed Linear GraphQL mutation; executes only through reverse ETL approval; flags: --id
    attachment get - Planned bounded Linear Query.attachment command. [intent=direct_read availability=planned]; notes: Planned only: fixed GraphQL direct operation execution is not enabled for Linear in this connector-local slice.
    attachments for url list - List Linear attachmentsForURL. [intent=etl availability=implemented stream=attachments_for_url]; flags: --url
    attachments list - List Linear attachments. [intent=etl availability=implemented stream=attachments]
    agent activities list - List Linear agentActivities. [intent=etl availability=implemented stream=agent_activities]
    agent sessions list - List Linear agentSessions. [intent=etl availability=implemented stream=agent_sessions]
    agent skills list - List Linear agentSkills. [intent=etl availability=implemented stream=agent_skills]
    agent skill delete - Run Linear mutation agentSkillDelete. [intent=reverse_etl availability=implemented write=agent_skill_delete]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --id
    agent activity events - Planned bounded Linear Subscription.agentActivityCreated command. [intent=direct_read availability=planned]; notes: Planned only: fixed GraphQL direct operation execution is not enabled for Linear in this connector-local slice.
  Other Commands
    image upload from url - Run Linear mutation imageUploadFromUrl. [intent=reverse_etl availability=implemented write=image_upload_from_url]; approval: reverse ETL plan → preview → explicit approval → execute.; risk: typed Linear GraphQL mutation; executes only through reverse ETL approval; flags: --url
    passkey login start - Run Linear mutation passkeyLoginStart. [intent=reverse_etl availability=implemented write=passkey_login_start]; approval: reverse ETL plan → preview → explicit approval → execute.; risk: typed Linear GraphQL mutation; executes only through reverse ETL approval; flags: --auth-id
    resend organization invite - Run Linear mutation resendOrganizationInvite. [intent=reverse_etl availability=implemented write=resend_organization_invite]; approval: reverse ETL plan → preview → explicit approval → execute.; risk: typed Linear GraphQL mutation; executes only through reverse ETL approval; flags: --id
    resend organization invite by email - Run Linear mutation resendOrganizationInviteByEmail. [intent=reverse_etl availability=implemented write=resend_organization_invite_by_email]; approval: reverse ETL plan → preview → explicit approval → execute.; risk: typed Linear GraphQL mutation; executes only through reverse ETL approval; flags: --email
  Help topics:
    linear auth - Configure Linear api_key or access_token without printing secret values.
    linear safety - Linear writes use reverse ETL plan, preview, explicit approval, execute, and typed destructive confirmation where applicable.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect linear

  # Inspect as structured JSON
  pm connectors inspect linear --json

AGENT WORKFLOW
  - Run pm connectors inspect linear before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
