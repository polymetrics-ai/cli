# pm connectors inspect linear

```text
NAME
  pm connectors inspect linear - Linear connector manual

SYNOPSIS
  pm connectors inspect linear
  pm connectors inspect linear --json
  pm credentials add <name> --connector linear [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads and writes Linear issues, teams, projects, users, and approved common mutations through fixed Linear GraphQL operations.

ICON
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  auth_type
  base_url
  max_pages
  access_token (secret)
  api_key (secret)

ETL STREAMS
  issues:
    primary key: id
    cursor: updated_at
    fields: assignee_email(), assignee_id(), branch_name(), canceled_at(), completed_at(), createdAt(), created_at(), creator_id(), description(), estimate(), id(), identifier(), priority(), state_id(), state_name(), state_type(), team_id(), team_key(), title(), updatedAt(), updated_at(), url()
  teams:
    primary key: id
    cursor: updated_at
    fields: createdAt(), created_at(), description(), id(), key(), name(), private(), updatedAt(), updated_at()
  projects:
    primary key: id
    cursor: updated_at
    fields: canceled_at(), completed_at(), createdAt(), created_at(), description(), id(), name(), progress(), started_at(), state(), updatedAt(), updated_at(), url()
  users:
    primary key: id
    cursor: updated_at
    fields: active(), admin(), createdAt(), created_at(), display_name(), email(), id(), name(), updatedAt(), updated_at()
  issue:
    primary key: id
    cursor: updated_at
    fields: assignee_email(), assignee_id(), branch_name(), canceled_at(), completed_at(), createdAt(), created_at(), creator_id(), description(), estimate(), id(), identifier(), priority(), state_id(), state_name(), state_type(), team_id(), team_key(), title(), updatedAt(), updated_at(), url()
  team:
    primary key: id
    cursor: updated_at
    fields: createdAt(), created_at(), description(), id(), key(), name(), private(), updatedAt(), updated_at()
  project:
    primary key: id
    cursor: updated_at
    fields: canceled_at(), completed_at(), createdAt(), created_at(), description(), id(), name(), progress(), started_at(), state(), updatedAt(), updated_at(), url()
  user:
    primary key: id
    cursor: updated_at
    fields: active(), admin(), createdAt(), created_at(), display_name(), email(), id(), name(), updatedAt(), updated_at()
  administrable_teams:
    primary key: id
    fields: id(), updated_at()
  agent_activities:
    primary key: id
    fields: id(), updated_at()
  agent_activity:
    primary key: id
    fields: id(), updated_at()
  agent_session:
    primary key: id
    fields: id(), updated_at()
  agent_sessions:
    primary key: id
    fields: id(), updated_at()
  agent_skill:
    primary key: id
    fields: id(), updated_at()
  agent_skills:
    primary key: id
    fields: id(), updated_at()
  application_info:
    primary key: id
    fields: id(), updated_at()
  archived_integrations:
    primary key: id
    fields: id(), updated_at()
  attachment:
    primary key: id
    fields: id(), updated_at()
  attachment_issue:
    primary key: id
    fields: id(), updated_at()
  attachments:
    primary key: id
    fields: id(), updated_at()
  attachments_for_url:
    primary key: id
    fields: id(), updated_at()
  audit_entries:
    primary key: id
    fields: id(), updated_at()
  audit_entry_types:
    primary key: id
    fields: id(), updated_at()
  authentication_sessions:
    primary key: id
    fields: id(), updated_at()
  available_users:
    primary key: id
    fields: id(), updated_at()
  comment:
    primary key: id
    fields: id(), updated_at()
  comments:
    primary key: id
    fields: id(), updated_at()
  custom_view:
    primary key: id
    fields: id(), updated_at()
  custom_view_has_subscribers:
    primary key: id
    fields: id(), updated_at()
  custom_views:
    primary key: id
    fields: id(), updated_at()
  customer:
    primary key: id
    fields: id(), updated_at()
  customer_need:
    primary key: id
    fields: id(), updated_at()
  customer_needs:
    primary key: id
    fields: id(), updated_at()
  customer_status:
    primary key: id
    fields: id(), updated_at()
  customer_statuses:
    primary key: id
    fields: id(), updated_at()
  customer_tier:
    primary key: id
    fields: id(), updated_at()
  customer_tiers:
    primary key: id
    fields: id(), updated_at()
  customers:
    primary key: id
    fields: id(), updated_at()
  cycle:
    primary key: id
    fields: id(), updated_at()
  cycles:
    primary key: id
    fields: id(), updated_at()
  document:
    primary key: id
    fields: id(), updated_at()
  document_content_history:
    primary key: id
    fields: id(), updated_at()
  documents:
    primary key: id
    fields: id(), updated_at()
  email_intake_address:
    primary key: id
    fields: id(), updated_at()
  emoji:
    primary key: id
    fields: id(), updated_at()
  emojis:
    primary key: id
    fields: id(), updated_at()
  entity_external_link:
    primary key: id
    fields: id(), updated_at()
  external_user:
    primary key: id
    fields: id(), updated_at()
  external_users:
    primary key: id
    fields: id(), updated_at()
  favorite:
    primary key: id
    fields: id(), updated_at()
  favorites:
    primary key: id
    fields: id(), updated_at()
  initiative:
    primary key: id
    fields: id(), updated_at()
  initiative_filter_suggestion:
    primary key: id
    fields: id(), updated_at()
  initiative_label:
    primary key: id
    fields: id(), updated_at()
  initiative_labels:
    primary key: id
    fields: id(), updated_at()
  initiative_relation:
    primary key: id
    fields: id(), updated_at()
  initiative_relations:
    primary key: id
    fields: id(), updated_at()
  initiative_to_project:
    primary key: id
    fields: id(), updated_at()
  initiative_to_projects:
    primary key: id
    fields: id(), updated_at()
  initiative_update:
    primary key: id
    fields: id(), updated_at()
  initiative_updates:
    primary key: id
    fields: id(), updated_at()
  initiatives:
    primary key: id
    fields: id(), updated_at()
  integration:
    primary key: id
    fields: id(), updated_at()
  integration_has_scopes:
    primary key: id
    fields: id(), updated_at()
  integration_template:
    primary key: id
    fields: id(), updated_at()
  integration_templates:
    primary key: id
    fields: id(), updated_at()
  integrations:
    primary key: id
    fields: id(), updated_at()
  integrations_settings:
    primary key: id
    fields: id(), updated_at()
  issue_figma_file_key_search:
    primary key: id
    fields: id(), updated_at()
  issue_filter_suggestion:
    primary key: id
    fields: id(), updated_at()
  issue_import_check_csv:
    primary key: id
    fields: id(), updated_at()
  issue_import_check_sync:
    primary key: id
    fields: id(), updated_at()
  issue_import_jql_check:
    primary key: id
    fields: id(), updated_at()
  issue_label:
    primary key: id
    fields: id(), updated_at()
  issue_labels:
    primary key: id
    fields: id(), updated_at()
  issue_priority_values:
    primary key: id
    fields: id(), updated_at()
  issue_relation:
    primary key: id
    fields: id(), updated_at()
  issue_relations:
    primary key: id
    fields: id(), updated_at()
  issue_repository_suggestions:
    primary key: id
    fields: id(), updated_at()
  issue_search:
    primary key: id
    fields: id(), updated_at()
  issue_title_suggestion_from_customer_request:
    primary key: id
    fields: id(), updated_at()
  issue_to_release:
    primary key: id
    fields: id(), updated_at()
  issue_to_releases:
    primary key: id
    fields: id(), updated_at()
  issue_vcs_branch_search:
    primary key: id
    fields: id(), updated_at()
  latest_release_by_access_key:
    primary key: id
    fields: id(), updated_at()
  notification:
    primary key: id
    fields: id(), updated_at()
  notification_subscription:
    primary key: id
    fields: id(), updated_at()
  notification_subscriptions:
    primary key: id
    fields: id(), updated_at()
  notifications:
    primary key: id
    fields: id(), updated_at()
  organization:
    primary key: id
    fields: id(), updated_at()
  organization_exists:
    primary key: id
    fields: id(), updated_at()
  organization_invite:
    primary key: id
    fields: id(), updated_at()
  organization_invites:
    primary key: id
    fields: id(), updated_at()
  project_filter_suggestion:
    primary key: id
    fields: id(), updated_at()
  project_label:
    primary key: id
    fields: id(), updated_at()
  project_labels:
    primary key: id
    fields: id(), updated_at()
  project_milestone:
    primary key: id
    fields: id(), updated_at()
  project_milestones:
    primary key: id
    fields: id(), updated_at()
  project_relation:
    primary key: id
    fields: id(), updated_at()
  project_relations:
    primary key: id
    fields: id(), updated_at()
  project_status:
    primary key: id
    fields: id(), updated_at()
  project_statuses:
    primary key: id
    fields: id(), updated_at()
  project_update:
    primary key: id
    fields: id(), updated_at()
  project_updates:
    primary key: id
    fields: id(), updated_at()
  push_subscription_test:
    primary key: id
    fields: id(), updated_at()
  rate_limit_status:
    primary key: id
    fields: id(), updated_at()
  recent_releases_by_access_key:
    primary key: id
    fields: id(), updated_at()
  release:
    primary key: id
    fields: id(), updated_at()
  release_note:
    primary key: id
    fields: id(), updated_at()
  release_notes:
    primary key: id
    fields: id(), updated_at()
  release_pipeline:
    primary key: id
    fields: id(), updated_at()
  release_pipeline_by_access_key:
    primary key: id
    fields: id(), updated_at()
  release_pipelines:
    primary key: id
    fields: id(), updated_at()
  release_search:
    primary key: id
    fields: id(), updated_at()
  release_stage:
    primary key: id
    fields: id(), updated_at()
  release_stages:
    primary key: id
    fields: id(), updated_at()
  releases:
    primary key: id
    fields: id(), updated_at()
  roadmap:
    primary key: id
    fields: id(), updated_at()
  roadmap_to_project:
    primary key: id
    fields: id(), updated_at()
  roadmap_to_projects:
    primary key: id
    fields: id(), updated_at()
  roadmaps:
    primary key: id
    fields: id(), updated_at()
  search_documents:
    primary key: id
    fields: id(), updated_at()
  search_issues:
    primary key: id
    fields: id(), updated_at()
  search_projects:
    primary key: id
    fields: id(), updated_at()
  semantic_search:
    primary key: id
    fields: id(), updated_at()
  sla_configurations:
    primary key: id
    fields: id(), updated_at()
  sso_url_from_email:
    primary key: id
    fields: id(), updated_at()
  team_membership:
    primary key: id
    fields: id(), updated_at()
  team_memberships:
    primary key: id
    fields: id(), updated_at()
  template:
    primary key: id
    fields: id(), updated_at()
  templates:
    primary key: id
    fields: id(), updated_at()
  templates_for_integration:
    primary key: id
    fields: id(), updated_at()
  time_schedule:
    primary key: id
    fields: id(), updated_at()
  time_schedules:
    primary key: id
    fields: id(), updated_at()
  triage_responsibilities:
    primary key: id
    fields: id(), updated_at()
  triage_responsibility:
    primary key: id
    fields: id(), updated_at()
  user_sessions:
    primary key: id
    fields: id(), updated_at()
  user_settings:
    primary key: id
    fields: id(), updated_at()
  verify_git_hub_enterprise_server_installation:
    primary key: id
    fields: id(), updated_at()
  viewer:
    primary key: id
    fields: id(), updated_at()
  webhook:
    primary key: id
    fields: id(), updated_at()
  webhooks:
    primary key: id
    fields: id(), updated_at()
  workflow_state:
    primary key: id
    fields: id(), updated_at()
  workflow_states:
    primary key: id
    fields: id(), updated_at()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

REVERSE ETL ACTIONS
  create_issue:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `issueCreate` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  update_issue:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `issueUpdate` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  comment_issue:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `commentCreate` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  create_project:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `projectCreate` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  agent_activity_create:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `agentActivityCreate` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  agent_session_create_on_comment:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `agentSessionCreateOnComment` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  agent_session_create_on_issue:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `agentSessionCreateOnIssue` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  agent_session_update:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `agentSessionUpdate` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  agent_session_update_external_url:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `agentSessionUpdateExternalUrl` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  agent_skill_create:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `agentSkillCreate` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  agent_skill_delete:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `agentSkillDelete` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  agent_skill_update:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `agentSkillUpdate` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  airbyte_integration_connect:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `airbyteIntegrationConnect` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  attachment_create:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `attachmentCreate` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  attachment_delete:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `attachmentDelete` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  attachment_link_discord:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `attachmentLinkDiscord` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  attachment_link_front:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `attachmentLinkFront` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  attachment_link_git_hub_issue:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `attachmentLinkGitHubIssue` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  attachment_link_git_hub_pr:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `attachmentLinkGitHubPR` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  attachment_link_git_lab_mr:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `attachmentLinkGitLabMR` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  attachment_link_intercom:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `attachmentLinkIntercom` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  attachment_link_jira_issue:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `attachmentLinkJiraIssue` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  attachment_link_salesforce:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `attachmentLinkSalesforce` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  attachment_link_slack:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `attachmentLinkSlack` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  attachment_link_url:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `attachmentLinkURL` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  attachment_link_zendesk:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `attachmentLinkZendesk` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  attachment_sync_to_slack:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `attachmentSyncToSlack` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  attachment_update:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `attachmentUpdate` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  comment_delete:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `commentDelete` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  comment_resolve:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `commentResolve` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  comment_unresolve:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `commentUnresolve` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  comment_update:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `commentUpdate` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  contact_create:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `contactCreate` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  create_csv_export_report:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `createCsvExportReport` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  create_initiative_update_reminder:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `createInitiativeUpdateReminder` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  create_project_update_reminder:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `createProjectUpdateReminder` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  custom_view_create:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `customViewCreate` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  custom_view_delete:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `customViewDelete` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  custom_view_update:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `customViewUpdate` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  customer_create:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `customerCreate` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  customer_delete:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `customerDelete` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  customer_merge:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `customerMerge` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  customer_need_archive:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `customerNeedArchive` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  customer_need_create:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `customerNeedCreate` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  customer_need_create_from_attachment:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `customerNeedCreateFromAttachment` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  customer_need_delete:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `customerNeedDelete` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  customer_need_unarchive:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `customerNeedUnarchive` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  customer_need_update:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `customerNeedUpdate` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  customer_status_create:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `customerStatusCreate` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  customer_status_delete:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `customerStatusDelete` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  customer_status_update:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `customerStatusUpdate` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  customer_tier_create:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `customerTierCreate` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  customer_tier_delete:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `customerTierDelete` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  customer_tier_update:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `customerTierUpdate` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  customer_unsync:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `customerUnsync` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  customer_update:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `customerUpdate` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  customer_upsert:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `customerUpsert` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  cycle_archive:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `cycleArchive` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  cycle_create:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `cycleCreate` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  cycle_shift_all:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `cycleShiftAll` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  cycle_start_upcoming_cycle_today:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `cycleStartUpcomingCycleToday` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  cycle_update:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `cycleUpdate` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  document_create:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `documentCreate` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  document_delete:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `documentDelete` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  document_unarchive:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `documentUnarchive` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  document_update:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `documentUpdate` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  email_intake_address_create:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `emailIntakeAddressCreate` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  email_intake_address_delete:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `emailIntakeAddressDelete` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  email_intake_address_refresh_ses_domain_status:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `emailIntakeAddressRefreshSesDomainStatus` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  email_intake_address_rotate:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `emailIntakeAddressRotate` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  email_intake_address_update:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `emailIntakeAddressUpdate` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  email_token_user_account_auth:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `emailTokenUserAccountAuth` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  email_unsubscribe:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `emailUnsubscribe` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  email_user_account_auth_challenge:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `emailUserAccountAuthChallenge` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  emoji_create:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `emojiCreate` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  emoji_delete:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `emojiDelete` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  entity_external_link_create:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `entityExternalLinkCreate` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  entity_external_link_delete:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `entityExternalLinkDelete` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  entity_external_link_update:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `entityExternalLinkUpdate` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  favorite_create:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `favoriteCreate` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  favorite_delete:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `favoriteDelete` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  favorite_update:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `favoriteUpdate` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  file_upload:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `fileUpload` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  git_automation_state_create:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `gitAutomationStateCreate` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  git_automation_state_delete:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `gitAutomationStateDelete` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  git_automation_state_update:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `gitAutomationStateUpdate` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  git_automation_target_branch_create:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `gitAutomationTargetBranchCreate` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  git_automation_target_branch_delete:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `gitAutomationTargetBranchDelete` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  git_automation_target_branch_update:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `gitAutomationTargetBranchUpdate` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  google_user_account_auth:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `googleUserAccountAuth` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  image_upload_from_url:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `imageUploadFromUrl` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  import_file_upload:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `importFileUpload` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  initiative_archive:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `initiativeArchive` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  initiative_create:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `initiativeCreate` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  initiative_delete:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `initiativeDelete` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  initiative_label_create:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `initiativeLabelCreate` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  initiative_label_delete:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `initiativeLabelDelete` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  initiative_label_restore:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `initiativeLabelRestore` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  initiative_label_retire:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `initiativeLabelRetire` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  initiative_label_update:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `initiativeLabelUpdate` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  initiative_relation_create:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `initiativeRelationCreate` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  initiative_relation_delete:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `initiativeRelationDelete` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  initiative_relation_update:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `initiativeRelationUpdate` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  initiative_to_project_create:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `initiativeToProjectCreate` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  initiative_to_project_delete:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `initiativeToProjectDelete` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  initiative_to_project_update:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `initiativeToProjectUpdate` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  initiative_unarchive:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `initiativeUnarchive` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  initiative_update:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `initiativeUpdate` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  initiative_update_archive:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `initiativeUpdateArchive` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  initiative_update_create:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `initiativeUpdateCreate` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  initiative_update_unarchive:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `initiativeUpdateUnarchive` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  initiative_update_update:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `initiativeUpdateUpdate` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  integration_archive:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `integrationArchive` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  integration_asks_connect_channel:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `integrationAsksConnectChannel` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  integration_delete:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `integrationDelete` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  integration_discord:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `integrationDiscord` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  integration_figma:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `integrationFigma` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  integration_front:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `integrationFront` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  integration_git_hub_enterprise_server_connect:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `integrationGitHubEnterpriseServerConnect` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  integration_git_hub_personal:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `integrationGitHubPersonal` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  integration_github_commit_create:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `integrationGithubCommitCreate` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  integration_github_connect:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `integrationGithubConnect` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  integration_github_import_connect:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `integrationGithubImportConnect` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  integration_github_import_refresh:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `integrationGithubImportRefresh` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  integration_github_remove_code_access:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `integrationGithubRemoveCodeAccess` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  integration_gitlab_connect:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `integrationGitlabConnect` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  integration_gitlab_test_connection:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `integrationGitlabTestConnection` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  integration_gong:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `integrationGong` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  integration_google_sheets:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `integrationGoogleSheets` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  integration_intercom:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `integrationIntercom` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  integration_intercom_delete:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `integrationIntercomDelete` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  integration_intercom_settings_update:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `integrationIntercomSettingsUpdate` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  integration_jira_personal:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `integrationJiraPersonal` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  integration_loom:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `integrationLoom` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  integration_microsoft_personal_connect:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `integrationMicrosoftPersonalConnect` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  integration_microsoft_teams:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `integrationMicrosoftTeams` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  integration_request:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `integrationRequest` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  integration_salesforce:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `integrationSalesforce` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  integration_sentry_connect:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `integrationSentryConnect` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  integration_slack:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `integrationSlack` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  integration_slack_asks:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `integrationSlackAsks` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  integration_slack_custom_view_notifications:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `integrationSlackCustomViewNotifications` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  integration_slack_customer_channel_link:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `integrationSlackCustomerChannelLink` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  integration_slack_import_emojis:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `integrationSlackImportEmojis` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  integration_slack_or_asks_update_slack_team_name:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `integrationSlackOrAsksUpdateSlackTeamName` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  integration_slack_org_project_updates_post:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `integrationSlackOrgProjectUpdatesPost` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  integration_slack_personal:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `integrationSlackPersonal` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  integration_slack_post:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `integrationSlackPost` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  integration_slack_project_post:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `integrationSlackProjectPost` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  integration_template_create:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `integrationTemplateCreate` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  integration_template_delete:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `integrationTemplateDelete` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  integration_zendesk:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `integrationZendesk` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  integrations_settings_create:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `integrationsSettingsCreate` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  integrations_settings_update:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `integrationsSettingsUpdate` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  issue_add_label:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `issueAddLabel` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  issue_archive:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `issueArchive` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  issue_batch_create:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `issueBatchCreate` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  issue_batch_update:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `issueBatchUpdate` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  issue_delete:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `issueDelete` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  issue_external_sync_disable:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `issueExternalSyncDisable` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  issue_import_create_asana:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `issueImportCreateAsana` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  issue_import_create_csv_jira:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `issueImportCreateCSVJira` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  issue_import_create_clubhouse:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `issueImportCreateClubhouse` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  issue_import_create_github:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `issueImportCreateGithub` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  issue_import_create_jira:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `issueImportCreateJira` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  issue_import_delete:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `issueImportDelete` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  issue_import_process:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `issueImportProcess` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  issue_import_update:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `issueImportUpdate` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  issue_label_create:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `issueLabelCreate` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  issue_label_delete:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `issueLabelDelete` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  issue_label_restore:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `issueLabelRestore` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  issue_label_retire:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `issueLabelRetire` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  issue_label_update:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `issueLabelUpdate` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  issue_relation_create:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `issueRelationCreate` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  issue_relation_delete:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `issueRelationDelete` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  issue_relation_update:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `issueRelationUpdate` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  issue_reminder:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `issueReminder` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  issue_remove_label:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `issueRemoveLabel` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  issue_share:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `issueShare` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  issue_subscribe:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `issueSubscribe` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  issue_to_release_create:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `issueToReleaseCreate` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  issue_to_release_delete:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `issueToReleaseDelete` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  issue_to_release_delete_by_issue_and_release:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `issueToReleaseDeleteByIssueAndRelease` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  issue_unarchive:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `issueUnarchive` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  issue_unshare:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `issueUnshare` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  issue_unsubscribe:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `issueUnsubscribe` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  logout:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `logout` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  logout_all_sessions:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `logoutAllSessions` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  logout_other_sessions:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `logoutOtherSessions` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  logout_session:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `logoutSession` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  notification_archive:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `notificationArchive` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  notification_archive_all:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `notificationArchiveAll` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  notification_category_channel_subscription_update:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `notificationCategoryChannelSubscriptionUpdate` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  notification_mark_read_all:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `notificationMarkReadAll` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  notification_mark_unread_all:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `notificationMarkUnreadAll` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  notification_snooze_all:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `notificationSnoozeAll` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  notification_subscription_create:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `notificationSubscriptionCreate` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  notification_subscription_delete:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `notificationSubscriptionDelete` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  notification_subscription_update:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `notificationSubscriptionUpdate` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  notification_unarchive:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `notificationUnarchive` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  notification_unsnooze_all:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `notificationUnsnoozeAll` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  notification_update:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `notificationUpdate` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  organization_cancel_delete:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `organizationCancelDelete` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  organization_delete:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `organizationDelete` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  organization_delete_challenge:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `organizationDeleteChallenge` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  organization_domain_delete:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `organizationDomainDelete` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  organization_invite_create:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `organizationInviteCreate` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  organization_invite_delete:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `organizationInviteDelete` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  organization_invite_update:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `organizationInviteUpdate` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  organization_start_trial:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `organizationStartTrial` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  organization_start_trial_for_plan:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `organizationStartTrialForPlan` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  organization_update:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `organizationUpdate` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  project_add_label:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `projectAddLabel` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  project_archive:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `projectArchive` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  project_delete:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `projectDelete` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  project_external_sync_disable:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `projectExternalSyncDisable` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  project_label_create:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `projectLabelCreate` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  project_label_delete:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `projectLabelDelete` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  project_label_restore:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `projectLabelRestore` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  project_label_retire:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `projectLabelRetire` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  project_label_update:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `projectLabelUpdate` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  project_milestone_create:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `projectMilestoneCreate` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  project_milestone_delete:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `projectMilestoneDelete` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  project_milestone_update:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `projectMilestoneUpdate` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  project_relation_create:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `projectRelationCreate` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  project_relation_delete:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `projectRelationDelete` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  project_relation_update:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `projectRelationUpdate` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  project_remove_label:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `projectRemoveLabel` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  project_status_archive:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `projectStatusArchive` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  project_status_create:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `projectStatusCreate` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  project_status_unarchive:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `projectStatusUnarchive` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  project_status_update:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `projectStatusUpdate` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  project_unarchive:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `projectUnarchive` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  project_update:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `projectUpdate` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  project_update_archive:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `projectUpdateArchive` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  project_update_create:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `projectUpdateCreate` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  project_update_delete:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `projectUpdateDelete` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  project_update_unarchive:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `projectUpdateUnarchive` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  project_update_update:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `projectUpdateUpdate` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  push_subscription_create:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `pushSubscriptionCreate` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  push_subscription_delete:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `pushSubscriptionDelete` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  reaction_create:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `reactionCreate` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  reaction_delete:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `reactionDelete` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  refresh_google_sheets_data:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `refreshGoogleSheetsData` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  release_archive:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `releaseArchive` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  release_complete:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `releaseComplete` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  release_complete_by_access_key:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `releaseCompleteByAccessKey` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  release_create:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `releaseCreate` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  release_delete:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `releaseDelete` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  release_note_create:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `releaseNoteCreate` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  release_note_delete:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `releaseNoteDelete` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  release_note_update:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `releaseNoteUpdate` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  release_pipeline_archive:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `releasePipelineArchive` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  release_pipeline_create:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `releasePipelineCreate` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  release_pipeline_delete:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `releasePipelineDelete` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  release_pipeline_unarchive:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `releasePipelineUnarchive` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  release_pipeline_update:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `releasePipelineUpdate` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  release_stage_archive:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `releaseStageArchive` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  release_stage_create:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `releaseStageCreate` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  release_stage_unarchive:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `releaseStageUnarchive` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  release_stage_update:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `releaseStageUpdate` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  release_sync:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `releaseSync` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  release_sync_by_access_key:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `releaseSyncByAccessKey` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  release_unarchive:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `releaseUnarchive` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  release_update:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `releaseUpdate` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  release_update_by_pipeline:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `releaseUpdateByPipeline` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  release_update_by_pipeline_by_access_key:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `releaseUpdateByPipelineByAccessKey` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  resend_organization_invite:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `resendOrganizationInvite` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  resend_organization_invite_by_email:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `resendOrganizationInviteByEmail` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  roadmap_archive:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `roadmapArchive` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  roadmap_create:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `roadmapCreate` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  roadmap_delete:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `roadmapDelete` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  roadmap_to_project_create:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `roadmapToProjectCreate` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  roadmap_to_project_delete:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `roadmapToProjectDelete` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  roadmap_to_project_update:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `roadmapToProjectUpdate` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  roadmap_unarchive:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `roadmapUnarchive` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  roadmap_update:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `roadmapUpdate` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  saml_token_user_account_auth:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `samlTokenUserAccountAuth` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  team_create:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `teamCreate` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  team_cycles_delete:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `teamCyclesDelete` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  team_delete:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `teamDelete` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  team_key_delete:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `teamKeyDelete` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  team_membership_create:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `teamMembershipCreate` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  team_membership_delete:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `teamMembershipDelete` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  team_membership_update:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `teamMembershipUpdate` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  team_unarchive:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `teamUnarchive` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  team_update:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `teamUpdate` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  template_create:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `templateCreate` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  template_delete:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `templateDelete` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  template_update:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `templateUpdate` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  time_schedule_create:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `timeScheduleCreate` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  time_schedule_delete:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `timeScheduleDelete` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  time_schedule_refresh_integration_schedule:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `timeScheduleRefreshIntegrationSchedule` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  time_schedule_update:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `timeScheduleUpdate` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  time_schedule_upsert_external:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `timeScheduleUpsertExternal` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  track_anonymous_event:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `trackAnonymousEvent` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  triage_responsibility_create:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `triageResponsibilityCreate` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  triage_responsibility_delete:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `triageResponsibilityDelete` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  triage_responsibility_update:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `triageResponsibilityUpdate` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  user_change_role:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `userChangeRole` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  user_discord_connect:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `userDiscordConnect` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  user_external_user_disconnect:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `userExternalUserDisconnect` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  user_flag_update:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `userFlagUpdate` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  user_revoke_all_sessions:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `userRevokeAllSessions` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  user_revoke_session:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `userRevokeSession` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  user_settings_flags_reset:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `userSettingsFlagsReset` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  user_settings_update:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `userSettingsUpdate` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  user_suspend:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `userSuspend` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  user_unlink_from_identity_provider:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `userUnlinkFromIdentityProvider` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  user_unsuspend:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `userUnsuspend` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  user_update:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `userUpdate` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  view_preferences_create:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `viewPreferencesCreate` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  view_preferences_delete:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `viewPreferencesDelete` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  view_preferences_update:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `viewPreferencesUpdate` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  webhook_create:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `webhookCreate` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  webhook_delete:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `webhookDelete` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  webhook_rotate_secret:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `webhookRotateSecret` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  webhook_update:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `webhookUpdate` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  workflow_state_archive:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `workflowStateArchive` through reverse ETL (high risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  workflow_state_create:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `workflowStateCreate` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.
  workflow_state_update:
    endpoint: POST /graphql
    risk: Executes the fixed Linear GraphQL mutation `workflowStateUpdate` through reverse ETL (medium risk); changes may affect Linear workspace data and require plan, preview, approval, and execute.

SECURITY
  read risk: external Linear GraphQL API read of approved fixed documents
  write risk: approved Linear GraphQL mutations through reverse ETL plan, preview, approval, execute
  approval: writes require connector command plan/preview/approval; sensitive/admin/destructive operations remain blocked by default
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

COMMAND SURFACE
  Work with Linear issues, teams, projects, and users from the command line.
  Usage: pm linear <command> <subcommand> [flags]
  Source CLI: Linear app and GraphQL API (https://developers.linear.app/docs/graphql/working-with-the-graphql-api)
  Global flags:
    --json (boolean): Write machine-readable JSON output.
    --connection (string): Use a saved Linear connector credential.: maps_to=connection
  Core Linear Commands
    issue list - List Linear issues through the implemented ETL stream. [intent=etl availability=implemented stream=issues]
    issue view - View one Linear issue. [intent=direct_read availability=implemented stream=issue]; flags: --issue-id
    issue create - Create a Linear issue. [intent=reverse_etl availability=implemented write=create_issue]; approval: reverse ETL writes require plan, preview, approval token, execute, and typed confirmation when sensitive or destructive.; risk: Creates a visible issue in the configured Linear workspace.; flags: --team-id, --title, --description, --assignee-id, --project-id, --state-id
    issue update - Update a Linear issue. [intent=reverse_etl availability=implemented write=update_issue]; approval: reverse ETL writes require plan, preview, approval token, execute, and typed confirmation when sensitive or destructive.; risk: Mutates an existing Linear issue.; flags: --issue-id, --title, --description, --assignee-id, --project-id, --state-id
    project list - List Linear projects through the implemented ETL stream. [intent=etl availability=implemented stream=projects]
    project view - View one Linear project. [intent=direct_read availability=implemented stream=project]; flags: --project-id
    project create - Create a Linear project. [intent=reverse_etl availability=implemented write=create_project]; approval: reverse ETL writes require plan, preview, approval token, execute, and typed confirmation when sensitive or destructive.; risk: Creates a visible project in Linear.; flags: --team-id, --name, --description
    team list - List Linear teams through the implemented ETL stream. [intent=etl availability=implemented stream=teams]
    team view - View one Linear team. [intent=direct_read availability=implemented stream=team]; flags: --team-id
    user list - List Linear users through the implemented ETL stream. [intent=etl availability=implemented stream=users]
    user view - View one Linear user. [intent=direct_read availability=implemented stream=user]; flags: --user-id
  Collaboration Commands
    comment create - Create a comment on a Linear issue. [intent=reverse_etl availability=implemented write=comment_issue]; approval: reverse ETL writes require plan, preview, approval token, execute, and typed confirmation when sensitive or destructive.; risk: Creates a visible comment in Linear.; flags: --issue-id, --body
    cycle list - List Linear cycles. [intent=etl availability=planned]; notes: Planned ETL candidate from the official GraphQL query surface; not part of the current four-stream baseline.
    label list - List Linear issue labels. [intent=etl availability=planned]; notes: Planned ETL candidate from the official GraphQL query surface; not part of the current four-stream baseline.
    workflow-state list - List Linear workflow states. [intent=etl availability=planned]; notes: Planned ETL candidate from the official GraphQL query surface; not part of the current four-stream baseline.
  Administrative And Integration Commands
    workspace invite - Invite a user to a Linear workspace. [intent=reverse_etl availability=unsafe_or_disallowed]; approval: Blocked by default pending sensitive/admin policy, preview, approval token, and typed confirmation.; risk: Changes workspace membership and may grant access to private Linear data.; notes: Administrative membership mutation; not exposed as an executable command in this connector surface.
    webhook create - Create a Linear webhook. [intent=reverse_etl availability=unsafe_or_disallowed]; approval: Blocked by default pending sensitive/admin policy, explicit destination allow-listing, preview, approval token, and typed confirmation.; risk: Creates an outbound integration endpoint and may expose workspace events to external systems.; notes: Administrative integration mutation; not exposed as an executable command in this connector surface.
    webhook delete - Delete a Linear webhook. [intent=reverse_etl availability=unsafe_or_disallowed]; approval: Blocked by default pending sensitive/admin policy, preview, approval token, and typed confirmation.; risk: Deletes an integration endpoint and can disrupt external workflows.; notes: Destructive administrative mutation; not exposed as an executable command in this connector surface.
    api graphql - Run an arbitrary Linear GraphQL operation. [intent=raw_api availability=unsafe_or_disallowed]; approval: Disallowed. Use fixed reviewed stream, direct-read, or reverse-ETL actions instead.; risk: A raw GraphQL surface could bypass stream/write review, redaction, and approval policy.; notes: Generic GraphQL query or mutation execution is intentionally not exposed.
    auth login - Authenticate the Linear connector. [intent=auth availability=unsupported_local unsupported local workflow]; notes: Credentials are managed through Polymetrics credential commands and environment/stdin flows; secrets are never accepted in prompt text.
    config set - Configure Linear command defaults. [intent=config availability=unsupported_local unsupported local workflow]; notes: Connector configuration is handled by saved connections and runtime flags, not provider-specific config mutation commands.
  Help topics:
    authentication - Use `api_key` or `access_token` through saved credentials or environment/stdin flows; never put secrets in command text.
    graphql-safety - Linear GraphQL operations are exposed only as reviewed fixed streams, direct reads, or reverse-ETL actions; Raw arbitrary GraphQL is disallowed.
    writes - Implemented Linear mutations use fixed GraphQL documents and reverse ETL plan → preview → approval → execute; sensitive/admin/destructive mutations remain blocked by default.

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
