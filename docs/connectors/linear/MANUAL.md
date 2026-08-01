# pm connectors inspect linear

```text
NAME
  pm connectors inspect linear - Linear connector manual

SYNOPSIS
  pm connectors inspect linear
  pm connectors inspect linear --json
  pm credentials add <name> --connector linear [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Linear GraphQL list/connection data and exposes only fixed GraphQL reverse-ETL mutations whose payloads do not require success:Boolean! assertions.

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
    fields: __typename(), archivedAt(), createdAt(), id(), name(), url()
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
  integration_github_remove_code_access:
    endpoint: POST
    required fields: integrationId
    risk: destructive Linear GraphQL mutation; requires typed confirmation
  leave_organization:
    endpoint: POST
    required fields: organizationId
    risk: destructive Linear GraphQL mutation; requires typed confirmation

SECURITY
  read risk: fixed Linear GraphQL documents against documented root Query list/connection fields
  write risk: fixed Linear GraphQL mutations whose payloads do not require success:Boolean! enforcement; success-payload and destructive mutation shapes remain blocked
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
    agent activity events - Planned bounded Linear Subscription.agentActivityCreated command. [intent=direct_read availability=planned]; notes: Planned only: fixed GraphQL direct operation execution is not enabled for Linear in this connector-local slice.
    archived integrations list - List Linear archivedIntegrations. [intent=etl availability=implemented stream=archived_integrations]
    attachments for url list - List Linear attachmentsForURL. [intent=etl availability=implemented stream=attachments_for_url]; flags: --url
    attachments list - List Linear attachments. [intent=etl availability=implemented stream=attachments]
    audit entries list - List Linear auditEntries. [intent=etl availability=implemented stream=audit_entries]
    audit entry types list - List Linear auditEntryTypes. [intent=etl availability=implemented stream=audit_entry_types]
    authentication sessions list - List Linear authenticationSessions. [intent=etl availability=implemented stream=authentication_sessions]
    comments list - List Linear comments. [intent=etl availability=implemented stream=comments]
    custom views list - List Linear customViews. [intent=etl availability=implemented stream=custom_views]
    customer needs list - List Linear customerNeeds. [intent=etl availability=implemented stream=customer_needs]
    customer statuses list - List Linear customerStatuses. [intent=etl availability=implemented stream=customer_statuses]
    customer tiers list - List Linear customerTiers. [intent=etl availability=implemented stream=customer_tiers]
    customers list - List Linear customers. [intent=etl availability=implemented stream=customers]
    cycles list - List Linear cycles. [intent=etl availability=implemented stream=cycles]
    documents list - List Linear documents. [intent=etl availability=implemented stream=documents]
    emojis list - List Linear emojis. [intent=etl availability=implemented stream=emojis]
    external users list - List Linear externalUsers. [intent=etl availability=implemented stream=external_users]
    favorites list - List Linear favorites. [intent=etl availability=implemented stream=favorites]
    initiative labels list - List Linear initiativeLabels. [intent=etl availability=implemented stream=initiative_labels]
    initiative relations list - List Linear initiativeRelations. [intent=etl availability=implemented stream=initiative_relations]
    initiative to projects list - List Linear initiativeToProjects. [intent=etl availability=implemented stream=initiative_to_projects]
    initiative updates list - List Linear initiativeUpdates. [intent=etl availability=implemented stream=initiative_updates]
    initiatives list - List Linear initiatives. [intent=etl availability=implemented stream=initiatives]
    integration templates list - List Linear integrationTemplates. [intent=etl availability=implemented stream=integration_templates]
    integration github remove code access - Run Linear mutation integrationGithubRemoveCodeAccess. [intent=reverse_etl availability=implemented write=integration_github_remove_code_access]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --integration-id
    integrations list - List Linear integrations. [intent=etl availability=implemented stream=integrations]
    issue figma file key search list - List Linear issueFigmaFileKeySearch. [intent=etl availability=implemented stream=issue_figma_file_key_search]; flags: --fileKey
    issue labels list - List Linear issueLabels. [intent=etl availability=implemented stream=issue_labels]
    issue priority values list - List Linear issuePriorityValues. [intent=etl availability=implemented stream=issue_priority_values]
    issue relations list - List Linear issueRelations. [intent=etl availability=implemented stream=issue_relations]
    issue search list - List Linear issueSearch. [intent=etl availability=implemented stream=issue_search]
    issue to releases list - List Linear issueToReleases. [intent=etl availability=implemented stream=issue_to_releases]
    issues list - List Linear issues. [intent=etl availability=implemented stream=issues]
    oauth applications list - List Linear oauthApplications. [intent=etl availability=implemented stream=oauth_applications]
    organization invites list - List Linear organizationInvites. [intent=etl availability=implemented stream=organization_invites]
    project labels list - List Linear projectLabels. [intent=etl availability=implemented stream=project_labels]
    project milestones list - List Linear projectMilestones. [intent=etl availability=implemented stream=project_milestones]
    project relations list - List Linear projectRelations. [intent=etl availability=implemented stream=project_relations]
    project statuses list - List Linear projectStatuses. [intent=etl availability=implemented stream=project_statuses]
    project updates list - List Linear projectUpdates. [intent=etl availability=implemented stream=project_updates]
    projects list - List Linear projects. [intent=etl availability=implemented stream=projects]
    recent releases by access key list - List Linear recentReleasesByAccessKey. [intent=etl availability=implemented stream=recent_releases_by_access_key]
    release notes list - List Linear releaseNotes. [intent=etl availability=implemented stream=release_notes]
    release pipelines list - List Linear releasePipelines. [intent=etl availability=implemented stream=release_pipelines]
    release search list - List Linear releaseSearch. [intent=etl availability=implemented stream=release_search]
    release stages list - List Linear releaseStages. [intent=etl availability=implemented stream=release_stages]
    releases list - List Linear releases. [intent=etl availability=implemented stream=releases]
    roadmap to projects list - List Linear roadmapToProjects. [intent=etl availability=implemented stream=roadmap_to_projects]
    roadmaps list - List Linear roadmaps. [intent=etl availability=implemented stream=roadmaps]
    sla configurations list - List Linear slaConfigurations. [intent=etl availability=implemented stream=sla_configurations]; flags: --teamId
    team memberships list - List Linear teamMemberships. [intent=etl availability=implemented stream=team_memberships]
    teams list - List Linear teams. [intent=etl availability=implemented stream=teams]
    templates for integration list - List Linear templatesForIntegration. [intent=etl availability=implemented stream=templates_for_integration]; flags: --integrationType
    templates list - List Linear templates. [intent=etl availability=implemented stream=templates]
    time schedules list - List Linear timeSchedules. [intent=etl availability=implemented stream=time_schedules]
    triage responsibilities list - List Linear triageResponsibilities. [intent=etl availability=implemented stream=triage_responsibilities]
    user sessions list - List Linear userSessions. [intent=etl availability=implemented stream=user_sessions]; flags: --id
    users list - List Linear users. [intent=etl availability=implemented stream=users]
    webhooks list - List Linear webhooks. [intent=etl availability=implemented stream=webhooks]
    workflow states list - List Linear workflowStates. [intent=etl availability=implemented stream=workflow_states]
  Typed reverse-ETL mutation commands
    integration templates list - List Linear integrationTemplates. [intent=etl availability=implemented stream=integration_templates]
    integration github remove code access - Run Linear mutation integrationGithubRemoveCodeAccess. [intent=reverse_etl availability=implemented write=integration_github_remove_code_access]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --integration-id
    leave organization - Run Linear mutation leaveOrganization. [intent=reverse_etl availability=implemented write=leave_organization]; approval: reverse ETL plan → preview → explicit approval → execute with typed destructive confirmation.; risk: destructive Linear GraphQL mutation; requires typed confirmation; flags: --organization-id
  Planned direct/binary/changefeed commands
    attachment get - Planned bounded Linear Query.attachment command. [intent=direct_read availability=planned]; notes: Planned only: fixed GraphQL direct operation execution is not enabled for Linear in this connector-local slice.
    attachments for url list - List Linear attachmentsForURL. [intent=etl availability=implemented stream=attachments_for_url]; flags: --url
    attachments list - List Linear attachments. [intent=etl availability=implemented stream=attachments]
    agent activities list - List Linear agentActivities. [intent=etl availability=implemented stream=agent_activities]
    agent sessions list - List Linear agentSessions. [intent=etl availability=implemented stream=agent_sessions]
    agent skills list - List Linear agentSkills. [intent=etl availability=implemented stream=agent_skills]
    agent activity events - Planned bounded Linear Subscription.agentActivityCreated command. [intent=direct_read availability=planned]; notes: Planned only: fixed GraphQL direct operation execution is not enabled for Linear in this connector-local slice.
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
