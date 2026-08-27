# pm connectors inspect miro

```text
NAME
  pm connectors inspect miro - Miro connector manual

SYNOPSIS
  pm connectors inspect miro
  pm connectors inspect miro --json
  pm credentials add <name> --connector miro [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads and writes documented Miro Platform, Enterprise, SCIM, and experimental REST API resources through the Miro Developer Platform API.

ICON
  id: simple-icons-miro
  asset: icons/simple-icons/miro.svg
  title: Miro
  simple_icon_slug: miro
  simple_icon_hex: 050038
  source: simple-icons
  license: CC0-1.0
  review_status: cc0_with_trademark_caveat
  review_url: https://simpleicons.org/?q=Miro
  match: exact-name-or-slug
  matched_by: miro

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  app_id
  base_url
  board_id
  board_id_platform_containers
  board_id_platform_tags
  board_member_id
  case_id
  connector_id
  created_after
  created_before
  end_date
  from
  group_id
  group_item_id
  id
  item_id
  job_id
  legal_hold_id
  member_id
  org_id
  parent_item_id
  project_id
  resource
  start_date
  tag_id
  team_id
  to
  uri
  api_key (secret) (required)

ETL STREAMS
  boards:
    primary key: id
    fields: created_at(string), description(string), id(string), modified_at(string), name(string), owner_id(string), team_id(string), type(string), view_link(string)
  board_users:
    primary key: id
    fields: board_id(string), id(string), name(string), role(string), type(string)
  board_items:
    primary key: id
    fields: board_id(string), created_at(string), id(string), modified_at(string), type(string)
  board_tags:
    primary key: id
    fields: board_id(string), fill_color(string), id(string), title(string), type(string)
  board_connectors:
    primary key: id
    fields: board_id(string), id(string), shape(string), type(string)
  orgs_org_id_ai_interaction_logs:
    primary key: id
    fields: actor(object), aiFeatureName(string), createdAt(string), details(object), id(string), logType(string), messageId(string), object(object), sessionId(string), storedAt(string)
  audit_logs:
    primary key: id
    fields: category(string), context(object), createdAt(string), createdBy(object), details(object), event(string), id(string), object(object)
  orgs_org_id_data_classification_settings:
    fields: enabled(boolean), labels(array), type(string)
  orgs_org_id_teams_team_id_data_classification_settings:
    fields: defaultLabelId(string), enabled(boolean), type(string)
  orgs_org_id_teams_team_id_boards_board_id_data_classification:
    primary key: id
    fields: color(string), description(string), guidelineUrl(string), id(string), name(string), sharingRecommendation(string), type(string)
  boards_board_id_docs_item_id:
    primary key: id
    fields: createdAt(string), createdBy(object), data(object), id(string), links(object), modifiedAt(string), modifiedBy(object), parent(object), position(object), type(string)
  orgs_org_id_cases:
    primary key: id
    fields: createdAt(string), createdBy(object), description(string), id(string), lastModifiedAt(string), lastModifiedBy(object), name(string), organizationId(string)
  orgs_org_id_cases_case_id:
    primary key: id
    fields: createdAt(string), createdBy(object), description(string), id(string), lastModifiedAt(string), lastModifiedBy(object), name(string), organizationId(string)
  orgs_org_id_cases_case_id_legal_holds:
    primary key: id
    fields: caseId(string), createdAt(string), createdBy(object), description(string), id(string), lastModifiedAt(string), lastModifiedBy(object), name(string), organizationId(string), scope(object), state(string)
  orgs_org_id_cases_case_id_export_jobs:
    primary key: id
    fields: id(string)
  orgs_org_id_cases_case_id_legal_holds_legal_hold_id:
    primary key: id
    fields: caseId(string), createdAt(string), createdBy(object), description(string), id(string), lastModifiedAt(string), lastModifiedBy(object), name(string), organizationId(string), scope(object), state(string)
  orgs_org_id_cases_case_id_legal_holds_legal_hold_id_content_items:
    fields: contentId(string), type(string)
  orgs_org_id_boards_export_jobs:
    primary key: id
    fields: boardFormat(string), createdAt(string), creator(object), id(string), modifiedAt(string), name(string), status(string), tasksCount(object)
  orgs_org_id_boards_export_jobs_job_id:
    fields: jobStatus(string)
  orgs_org_id_boards_export_jobs_job_id_results:
    fields: boardId(string), errorMessage(string), errorType(string), exportLink(string), status(string)
  orgs_org_id_boards_export_jobs_job_id_tasks:
    primary key: id
    fields: artifactExpiredAt(string), board(object), errorMessage(string), errorType(string), id(string), sizeInBytes(integer), status(string)
  orgs_org_id_content_logs_items:
    primary key: id
    fields: actionTime(string), actionType(string), actor(object), contentId(string), id(string), itemId(string), itemType(string), relationships(array), state(object)
  users:
    primary key: id
    fields: active(boolean), displayName(string), emails(array), groups(array), id(string), meta(object), name(object), photos(array), preferredLanguage(string), roles(array), schemas(array), urn:ietf:params:scim:schemas:extension:enterprise:2.0:User(object), userName(string), userType(string)
  users_id:
    primary key: id
    fields: active(boolean), displayName(string), emails(array), groups(array), id(string), meta(object), name(object), photos(array), preferredLanguage(string), roles(array), schemas(array), urn:ietf:params:scim:schemas:extension:enterprise:2.0:User(object), userName(string), userType(string)
  groups:
    primary key: id
    fields: displayName(string), id(string), members(array), meta(object), schemas(array)
  groups_id:
    primary key: id
    fields: displayName(string), id(string), members(array), meta(object), schemas(array)
  service_provider_config:
    fields: authenticationSchemes(array), bulk(object), changePassword(object), documentationUri(string), etag(object), filter(object), patch(object), schemas(array), sort(object)
  resource_types:
    primary key: id
    fields: description(string), endpoint(string), id(string), name(string), schema(string), schemaExtensions(array), schemas(array)
  resource_types_resource:
    primary key: id
    fields: description(string), endpoint(string), id(string), name(string), schema(string), schemaExtensions(array), schemas(array)
  schemas:
    primary key: id
    fields: attributes(array), description(string), id(string), meta(object), name(string), schemas(array)
  schemas_uri:
    primary key: id
    fields: attributes(array), description(string), id(string), meta(object), name(string)
  orgs_org_id:
    primary key: id
    fields: fullLicensesPurchased(integer), id(string), name(string), plan(string), type(string)
  orgs_org_id_members:
    primary key: id
    fields: active(boolean), adminRoles(array), email(string), id(string), lastActivityAt(string), license(string), licenseAssignedAt(string), role(string), type(string)
  orgs_org_id_members_member_id:
    primary key: id
    fields: active(boolean), adminRoles(array), email(string), id(string), lastActivityAt(string), license(string), licenseAssignedAt(string), role(string), type(string)
  boards_board_id:
    primary key: id
    fields: createdAt(string), createdBy(object), currentUserMembership(object), description(string), id(string), lastOpenedAt(string), lastOpenedBy(object), links(object), modifiedAt(string), modifiedBy(object), name(string), owner(object), picture(object), policy(object), project(object), team(object), type(string), viewLink(string)
  boards_board_id_app_cards_item_id:
    primary key: id
    fields: createdAt(string), createdBy(object), data(object), geometry(object), id(string), links(object), modifiedAt(string), modifiedBy(object), parent(object), position(object), style(object), type(string)
  boards_board_id_cards_item_id:
    primary key: id
    fields: createdAt(string), createdBy(object), data(object), geometry(object), id(string), links(object), modifiedAt(string), modifiedBy(object), parent(object), position(object), style(object), type(string)
  boards_board_id_connectors_connector_id:
    primary key: id
    fields: captions(array), createdAt(string), createdBy(object), endItem(object), id(string), isSupported(boolean), links(object), modifiedAt(string), modifiedBy(object), shape(string), startItem(object), style(object), type(string)
  boards_board_id_documents_item_id:
    primary key: id
    fields: createdAt(string), createdBy(object), data(object), geometry(object), id(string), links(object), modifiedAt(string), modifiedBy(object), parent(object), position(object), type(string)
  boards_board_id_embeds_item_id:
    primary key: id
    fields: createdAt(string), createdBy(object), data(object), geometry(object), id(string), links(object), modifiedAt(string), modifiedBy(object), parent(object), position(object), type(string)
  boards_board_id_images_item_id:
    primary key: id
    fields: createdAt(string), createdBy(object), data(object), geometry(object), id(string), links(object), modifiedAt(string), modifiedBy(object), parent(object), position(object), type(string)
  boards_board_id_items_item_id:
    primary key: id
    fields: createdAt(string), createdBy(object), data(object), geometry(object), id(string), modifiedAt(string), modifiedBy(object), parent(object), position(object), type(string)
  boards_board_id_members_board_member_id:
    primary key: id
    fields: id(string), links(object), name(string), role(string), type(string)
  boards_board_id_shapes_item_id:
    primary key: id
    fields: createdAt(string), createdBy(object), data(object), geometry(object), id(string), links(object), modifiedAt(string), modifiedBy(object), parent(object), position(object), style(object), type(string)
  boards_board_id_sticky_notes_item_id:
    primary key: id
    fields: createdAt(string), createdBy(object), data(object), geometry(object), id(string), links(object), modifiedAt(string), modifiedBy(object), parent(object), position(object), style(object), type(string)
  boards_board_id_texts_item_id:
    primary key: id
    fields: createdAt(string), createdBy(object), data(object), geometry(object), id(string), links(object), modifiedAt(string), modifiedBy(object), parent(object), position(object), style(object), type(string)
  boards_board_id_frames_item_id:
    primary key: id
    fields: createdAt(string), createdBy(object), data(object), geometry(object), id(string), links(object), modifiedAt(string), modifiedBy(object), position(object), style(object), type(string)
  boards_board_id_platform_containers_items:
    primary key: id
    fields: createdAt(string), createdBy(object), data(object), geometry(object), id(string), modifiedAt(string), modifiedBy(object), parent(object), position(object), type(string)
  experimental_apps_app_id_metrics:
    fields: installations(integer), periodStart(string), uninstallations(integer), uniqueOrganizations(integer), uniqueRecurringUsers(integer), uniqueUsers(integer)
  experimental_apps_app_id_metrics_total:
    fields: installations(integer), uninstallations(integer), uniqueOrganizations(integer), uniqueRecurringUsers(integer), uniqueUsers(integer)
  experimental_boards_board_id_mindmap_nodes_item_id:
    primary key: id
    fields: createdAt(string), createdBy(object), data(object), id(string), links(object), modifiedAt(string), modifiedBy(object), parent(object), style(object), type(string)
  experimental_boards_board_id_mindmap_nodes:
    primary key: id
    fields: createdAt(string), createdBy(object), data(object), id(string), links(object), modifiedAt(string), modifiedBy(object), parent(object), style(object), type(string)
  experimental_boards_board_id_items:
    primary key: id
    fields: createdAt(string), createdBy(object), data(object), geometry(object), id(string), modifiedAt(string), modifiedBy(object), parent(object), position(object), type(string)
  experimental_boards_board_id_items_item_id:
    primary key: id
    fields: createdAt(string), createdBy(object), data(object), geometry(object), id(string), modifiedAt(string), modifiedBy(object), parent(object), position(object), type(string)
  experimental_boards_board_id_shapes_item_id:
    primary key: id
    fields: createdAt(string), createdBy(object), data(object), geometry(object), id(string), links(object), modifiedAt(string), modifiedBy(object), parent(object), position(object), style(object), type(string)
  experimental_boards_board_id_code_widgets:
    primary key: id
    fields: createdAt(string), createdBy(object), data(object), geometry(object), id(string), links(object), modifiedAt(string), modifiedBy(object), position(object), type(string)
  experimental_boards_board_id_code_widgets_item_id:
    primary key: id
    fields: createdAt(string), createdBy(object), data(object), geometry(object), id(string), links(object), modifiedAt(string), modifiedBy(object), position(object), type(string)
  boards_board_id_groups:
    primary key: id
    fields: data(object), id(string), links(object), type(string)
  boards_board_id_groups_items:
    fields: data(array), limit(integer), links(object), offset(integer), size(integer), total(integer), type(string)
  boards_board_id_groups_group_id:
    primary key: id
    fields: data(object), id(string), links(object), type(string)
  boards_board_id_items_item_id_tags:
    primary key: id
    fields: fillColor(string), id(string), title(string), type(string)
  boards_board_id_tags_tag_id:
    primary key: id
    fields: fillColor(string), id(string), links(object), title(string), type(string)
  boards_board_id_platform_tags_items:
    primary key: id
    fields: createdAt(string), createdBy(object), data(object), geometry(object), id(string), modifiedAt(string), modifiedBy(object), parent(object), position(object), type(string)
  orgs_org_id_teams_team_id_projects:
    primary key: id
    fields: id(string), name(string), type(string)
  orgs_org_id_teams_team_id_projects_project_id:
    primary key: id
    fields: id(string), name(string), type(string)
  orgs_org_id_teams_team_id_projects_project_id_settings:
    fields: sharingPolicySettings(object), type(string)
  orgs_org_id_teams_team_id_projects_project_id_members:
    primary key: id
    fields: email(string), id(string), role(string), type(string)
  orgs_org_id_teams_team_id_projects_project_id_members_member_id:
    primary key: id
    fields: email(string), id(string), role(string), type(string)
  orgs_org_id_teams:
    primary key: id
    fields: id(string), name(string), picture(object), type(string)
  orgs_org_id_teams_team_id:
    primary key: id
    fields: id(string), name(string), picture(object), type(string)
  orgs_org_id_teams_team_id_members:
    primary key: id
    fields: createdAt(string), createdBy(string), id(string), modifiedAt(string), modifiedBy(string), role(string), teamId(string), type(string)
  orgs_org_id_teams_team_id_members_member_id:
    primary key: id
    fields: createdAt(string), createdBy(string), id(string), modifiedAt(string), modifiedBy(string), role(string), teamId(string), type(string)
  orgs_org_id_default_teams_settings:
    fields: organizationId(string), teamAccountDiscoverySettings(object), teamCollaborationSettings(object), teamCopyAccessLevelSettings(object), teamId(string), teamInvitationSettings(object), teamSharingPolicySettings(object), type(string)
  orgs_org_id_teams_team_id_settings:
    fields: organizationId(string), teamAccountDiscoverySettings(object), teamCollaborationSettings(object), teamCopyAccessLevelSettings(object), teamId(string), teamInvitationSettings(object), teamSharingPolicySettings(object), type(string)
  orgs_org_id_groups:
    primary key: id
    fields: description(string), id(string), name(string), type(string)
  orgs_org_id_groups_group_id:
    primary key: id
    fields: description(string), id(string), name(string), type(string)
  orgs_org_id_groups_group_id_members:
    primary key: id
    fields: email(string), id(string), type(string)
  orgs_org_id_groups_group_id_members_member_id:
    primary key: id
    fields: email(string), id(string), type(string)
  orgs_org_id_groups_group_id_teams:
    primary key: id
    fields: id(string), role(string), type(string)
  orgs_org_id_groups_group_id_teams_team_id:
    primary key: id
    fields: id(string), role(string), type(string)
  orgs_org_id_teams_team_id_groups:
    primary key: id
    fields: id(string), role(string), type(string)
  orgs_org_id_teams_team_id_groups_group_id:
    primary key: id
    fields: id(string), role(string), type(string)
  orgs_org_id_boards_board_id_groups:
    primary key: id
    fields: id(string), role(object), type(object)
  orgs_org_id_projects_project_id_groups:
    primary key: id
    fields: id(string), role(object), type(object)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite

REVERSE ETL ACTIONS
  update_orgs_org_id_teams_team_id_data_classification:
    endpoint: PATCH /v2/orgs/{{ record.org_id }}/teams/{{ record.team_id }}/data-classification
    required fields: org_id, team_id
    risk: medium: external Miro API mutation; approval required
  update_orgs_org_id_teams_team_id_data_classification_settings:
    endpoint: PATCH /v2/orgs/{{ record.org_id }}/teams/{{ record.team_id }}/data-classification-settings
    required fields: org_id, team_id
    risk: medium: external Miro API mutation; approval required
  create_orgs_org_id_teams_team_id_boards_board_id_data_classification:
    endpoint: POST /v2/orgs/{{ record.org_id }}/teams/{{ record.team_id }}/boards/{{ record.board_id }}/data-classification
    required fields: org_id, team_id, board_id
    risk: medium: external Miro API mutation; approval required
  create_boards_board_id_docs:
    endpoint: POST /v2/boards/{{ record.board_id }}/docs
    required fields: board_id, data
    risk: medium: external Miro API mutation; approval required
  delete_boards_board_id_docs_item_id:
    endpoint: DELETE /v2/boards/{{ record.board_id }}/docs/{{ record.item_id }}
    required fields: board_id, item_id
    risk: high: external Miro API mutation; approval required
  create_orgs_org_id_cases:
    endpoint: POST /v2/orgs/{{ record.org_id }}/cases
    required fields: org_id, name
    risk: medium: external Miro API mutation; approval required
  update_orgs_org_id_cases_case_id:
    endpoint: PUT /v2/orgs/{{ record.org_id }}/cases/{{ record.case_id }}
    required fields: org_id, case_id, name
    risk: medium: external Miro API mutation; approval required
  delete_orgs_org_id_cases_case_id:
    endpoint: DELETE /v2/orgs/{{ record.org_id }}/cases/{{ record.case_id }}
    required fields: org_id, case_id
    risk: high: external Miro API mutation; approval required
  create_orgs_org_id_cases_case_id_legal_holds:
    endpoint: POST /v2/orgs/{{ record.org_id }}/cases/{{ record.case_id }}/legal-holds
    required fields: org_id, case_id, name, scope
    risk: medium: external Miro API mutation; approval required
  update_orgs_org_id_cases_case_id_legal_holds_legal_hold_id:
    endpoint: PUT /v2/orgs/{{ record.org_id }}/cases/{{ record.case_id }}/legal-holds/{{ record.legal_hold_id }}
    required fields: org_id, case_id, legal_hold_id, name, scope
    risk: medium: external Miro API mutation; approval required
  delete_orgs_org_id_cases_case_id_legal_holds_legal_hold_id:
    endpoint: DELETE /v2/orgs/{{ record.org_id }}/cases/{{ record.case_id }}/legal-holds/{{ record.legal_hold_id }}
    required fields: org_id, case_id, legal_hold_id
    risk: high: external Miro API mutation; approval required
  update_orgs_org_id_boards_export_jobs_job_id_status:
    endpoint: PUT /v2/orgs/{{ record.org_id }}/boards/export/jobs/{{ record.job_id }}/status
    required fields: org_id, job_id, status
    risk: medium: external Miro API mutation; approval required
  create_orgs_org_id_boards_export_jobs_job_id_tasks_task_id_export_link:
    endpoint: POST /v2/orgs/{{ record.org_id }}/boards/export/jobs/{{ record.job_id }}/tasks/{{ record.task_id }}/export-link
    required fields: org_id, job_id, task_id
    risk: medium: external Miro API mutation; approval required
  create_users:
    endpoint: POST /Users
    required fields: userName
    risk: medium: external Miro API mutation; approval required
  update_users_id:
    endpoint: PUT /Users/{{ record.id }}
    required fields: id
    risk: medium: external Miro API mutation; approval required
  update_users_id_2:
    endpoint: PATCH /Users/{{ record.id }}
    required fields: id, schemas, Operations
    risk: medium: external Miro API mutation; approval required
  delete_users_id:
    endpoint: DELETE /Users/{{ record.id }}
    required fields: id
    risk: high: external Miro API mutation; approval required
  update_groups_id:
    endpoint: PATCH /Groups/{{ record.id }}
    required fields: id, schemas, Operations
    risk: medium: external Miro API mutation; approval required
  create_boards:
    endpoint: POST /v2/boards
    risk: medium: external Miro API mutation; approval required
  update_boards_board_id:
    endpoint: PATCH /v2/boards/{{ record.board_id }}
    required fields: board_id
    risk: medium: external Miro API mutation; approval required
  delete_boards_board_id:
    endpoint: DELETE /v2/boards/{{ record.board_id }}
    required fields: board_id
    risk: high: external Miro API mutation; approval required
  create_boards_board_id_app_cards:
    endpoint: POST /v2/boards/{{ record.board_id }}/app_cards
    required fields: board_id
    risk: medium: external Miro API mutation; approval required
  update_boards_board_id_app_cards_item_id:
    endpoint: PATCH /v2/boards/{{ record.board_id }}/app_cards/{{ record.item_id }}
    required fields: board_id, item_id
    risk: medium: external Miro API mutation; approval required
  delete_boards_board_id_app_cards_item_id:
    endpoint: DELETE /v2/boards/{{ record.board_id }}/app_cards/{{ record.item_id }}
    required fields: board_id, item_id
    risk: high: external Miro API mutation; approval required
  create_boards_board_id_cards:
    endpoint: POST /v2/boards/{{ record.board_id }}/cards
    required fields: board_id
    risk: medium: external Miro API mutation; approval required
  update_boards_board_id_cards_item_id:
    endpoint: PATCH /v2/boards/{{ record.board_id }}/cards/{{ record.item_id }}
    required fields: board_id, item_id
    risk: medium: external Miro API mutation; approval required
  delete_boards_board_id_cards_item_id:
    endpoint: DELETE /v2/boards/{{ record.board_id }}/cards/{{ record.item_id }}
    required fields: board_id, item_id
    risk: high: external Miro API mutation; approval required
  create_boards_board_id_connectors:
    endpoint: POST /v2/boards/{{ record.board_id }}/connectors
    required fields: board_id, endItem, startItem
    risk: medium: external Miro API mutation; approval required
  update_boards_board_id_connectors_connector_id:
    endpoint: PATCH /v2/boards/{{ record.board_id }}/connectors/{{ record.connector_id }}
    required fields: board_id, connector_id
    risk: medium: external Miro API mutation; approval required
  delete_boards_board_id_connectors_connector_id:
    endpoint: DELETE /v2/boards/{{ record.board_id }}/connectors/{{ record.connector_id }}
    required fields: board_id, connector_id
    risk: high: external Miro API mutation; approval required
  create_boards_board_id_documents:
    endpoint: POST /v2/boards/{{ record.board_id }}/documents
    required fields: board_id, data
    risk: medium: external Miro API mutation; approval required
  update_boards_board_id_documents_item_id:
    endpoint: PATCH /v2/boards/{{ record.board_id }}/documents/{{ record.item_id }}
    required fields: board_id, item_id
    risk: medium: external Miro API mutation; approval required
  delete_boards_board_id_documents_item_id:
    endpoint: DELETE /v2/boards/{{ record.board_id }}/documents/{{ record.item_id }}
    required fields: board_id, item_id
    risk: high: external Miro API mutation; approval required
  create_boards_board_id_embeds:
    endpoint: POST /v2/boards/{{ record.board_id }}/embeds
    required fields: board_id, data
    risk: medium: external Miro API mutation; approval required
  update_boards_board_id_embeds_item_id:
    endpoint: PATCH /v2/boards/{{ record.board_id }}/embeds/{{ record.item_id }}
    required fields: board_id, item_id
    risk: medium: external Miro API mutation; approval required
  delete_boards_board_id_embeds_item_id:
    endpoint: DELETE /v2/boards/{{ record.board_id }}/embeds/{{ record.item_id }}
    required fields: board_id, item_id
    risk: high: external Miro API mutation; approval required
  create_boards_board_id_images:
    endpoint: POST /v2/boards/{{ record.board_id }}/images
    required fields: board_id, data
    risk: medium: external Miro API mutation; approval required
  update_boards_board_id_images_item_id:
    endpoint: PATCH /v2/boards/{{ record.board_id }}/images/{{ record.item_id }}
    required fields: board_id, item_id
    risk: medium: external Miro API mutation; approval required
  delete_boards_board_id_images_item_id:
    endpoint: DELETE /v2/boards/{{ record.board_id }}/images/{{ record.item_id }}
    required fields: board_id, item_id
    risk: high: external Miro API mutation; approval required
  update_boards_board_id_items_item_id:
    endpoint: PATCH /v2/boards/{{ record.board_id }}/items/{{ record.item_id }}
    required fields: board_id, item_id
    risk: medium: external Miro API mutation; approval required
  delete_boards_board_id_items_item_id:
    endpoint: DELETE /v2/boards/{{ record.board_id }}/items/{{ record.item_id }}
    required fields: board_id, item_id
    risk: high: external Miro API mutation; approval required
  create_boards_board_id_members:
    endpoint: POST /v2/boards/{{ record.board_id }}/members
    required fields: board_id, emails
    risk: medium: external Miro API mutation; approval required
  update_boards_board_id_members_board_member_id:
    endpoint: PATCH /v2/boards/{{ record.board_id }}/members/{{ record.board_member_id }}
    required fields: board_id, board_member_id
    risk: medium: external Miro API mutation; approval required
  delete_boards_board_id_members_board_member_id:
    endpoint: DELETE /v2/boards/{{ record.board_id }}/members/{{ record.board_member_id }}
    required fields: board_id, board_member_id
    risk: high: external Miro API mutation; approval required
  create_boards_board_id_shapes:
    endpoint: POST /v2/boards/{{ record.board_id }}/shapes
    required fields: board_id
    risk: medium: external Miro API mutation; approval required
  update_boards_board_id_shapes_item_id:
    endpoint: PATCH /v2/boards/{{ record.board_id }}/shapes/{{ record.item_id }}
    required fields: board_id, item_id
    risk: medium: external Miro API mutation; approval required
  delete_boards_board_id_shapes_item_id:
    endpoint: DELETE /v2/boards/{{ record.board_id }}/shapes/{{ record.item_id }}
    required fields: board_id, item_id
    risk: high: external Miro API mutation; approval required
  create_boards_board_id_sticky_notes:
    endpoint: POST /v2/boards/{{ record.board_id }}/sticky_notes
    required fields: board_id
    risk: medium: external Miro API mutation; approval required
  update_boards_board_id_sticky_notes_item_id:
    endpoint: PATCH /v2/boards/{{ record.board_id }}/sticky_notes/{{ record.item_id }}
    required fields: board_id, item_id
    risk: medium: external Miro API mutation; approval required
  delete_boards_board_id_sticky_notes_item_id:
    endpoint: DELETE /v2/boards/{{ record.board_id }}/sticky_notes/{{ record.item_id }}
    required fields: board_id, item_id
    risk: high: external Miro API mutation; approval required
  create_boards_board_id_texts:
    endpoint: POST /v2/boards/{{ record.board_id }}/texts
    required fields: board_id, data
    risk: medium: external Miro API mutation; approval required
  update_boards_board_id_texts_item_id:
    endpoint: PATCH /v2/boards/{{ record.board_id }}/texts/{{ record.item_id }}
    required fields: board_id, item_id
    risk: medium: external Miro API mutation; approval required
  delete_boards_board_id_texts_item_id:
    endpoint: DELETE /v2/boards/{{ record.board_id }}/texts/{{ record.item_id }}
    required fields: board_id, item_id
    risk: high: external Miro API mutation; approval required
  create_boards_board_id_frames:
    endpoint: POST /v2/boards/{{ record.board_id }}/frames
    required fields: board_id, data
    risk: medium: external Miro API mutation; approval required
  update_boards_board_id_frames_item_id:
    endpoint: PATCH /v2/boards/{{ record.board_id }}/frames/{{ record.item_id }}
    required fields: board_id, item_id
    risk: medium: external Miro API mutation; approval required
  delete_boards_board_id_frames_item_id:
    endpoint: DELETE /v2/boards/{{ record.board_id }}/frames/{{ record.item_id }}
    required fields: board_id, item_id
    risk: high: external Miro API mutation; approval required
  delete_experimental_boards_board_id_mindmap_nodes_item_id:
    endpoint: DELETE /v2-experimental/boards/{{ record.board_id }}/mindmap_nodes/{{ record.item_id }}
    required fields: board_id, item_id
    risk: high: external Miro API mutation; approval required
  create_experimental_boards_board_id_mindmap_nodes:
    endpoint: POST /v2-experimental/boards/{{ record.board_id }}/mindmap_nodes
    required fields: board_id, data
    risk: medium: external Miro API mutation; approval required
  delete_experimental_boards_board_id_items_item_id:
    endpoint: DELETE /v2-experimental/boards/{{ record.board_id }}/items/{{ record.item_id }}
    required fields: board_id, item_id
    risk: high: external Miro API mutation; approval required
  create_experimental_boards_board_id_shapes:
    endpoint: POST /v2-experimental/boards/{{ record.board_id }}/shapes
    required fields: board_id
    risk: medium: external Miro API mutation; approval required
  update_experimental_boards_board_id_shapes_item_id:
    endpoint: PATCH /v2-experimental/boards/{{ record.board_id }}/shapes/{{ record.item_id }}
    required fields: board_id, item_id
    risk: medium: external Miro API mutation; approval required
  delete_experimental_boards_board_id_shapes_item_id:
    endpoint: DELETE /v2-experimental/boards/{{ record.board_id }}/shapes/{{ record.item_id }}
    required fields: board_id, item_id
    risk: high: external Miro API mutation; approval required
  create_experimental_boards_board_id_code_widgets:
    endpoint: POST /v2-experimental/boards/{{ record.board_id }}/code_widgets
    required fields: board_id
    risk: medium: external Miro API mutation; approval required
  update_experimental_boards_board_id_code_widgets_item_id:
    endpoint: PATCH /v2-experimental/boards/{{ record.board_id }}/code_widgets/{{ record.item_id }}
    required fields: board_id, item_id
    risk: medium: external Miro API mutation; approval required
  delete_experimental_boards_board_id_code_widgets_item_id:
    endpoint: DELETE /v2-experimental/boards/{{ record.board_id }}/code_widgets/{{ record.item_id }}
    required fields: board_id, item_id
    risk: high: external Miro API mutation; approval required
  update_experimental_boards_board_id_code_widgets_item_id_position:
    endpoint: PATCH /v2-experimental/boards/{{ record.board_id }}/code_widgets/{{ record.item_id }}/position
    required fields: board_id, item_id
    risk: medium: external Miro API mutation; approval required
  create_boards_board_id_groups:
    endpoint: POST /v2/boards/{{ record.board_id }}/groups
    required fields: board_id, id, name, type
    risk: medium: external Miro API mutation; approval required
  update_boards_board_id_groups_group_id:
    endpoint: PUT /v2/boards/{{ record.board_id }}/groups/{{ record.group_id }}
    required fields: board_id, group_id, id, name, type
    risk: medium: external Miro API mutation; approval required
  delete_boards_board_id_groups_group_id:
    endpoint: DELETE /v2/boards/{{ record.board_id }}/groups/{{ record.group_id }}
    required fields: board_id, group_id
    risk: high: external Miro API mutation; approval required
  create_boards_board_id_tags:
    endpoint: POST /v2/boards/{{ record.board_id }}/tags
    required fields: board_id, title
    risk: medium: external Miro API mutation; approval required
  update_boards_board_id_tags_tag_id:
    endpoint: PATCH /v2/boards/{{ record.board_id }}/tags/{{ record.tag_id }}
    required fields: board_id, tag_id
    risk: medium: external Miro API mutation; approval required
  delete_boards_board_id_tags_tag_id:
    endpoint: DELETE /v2/boards/{{ record.board_id }}/tags/{{ record.tag_id }}
    required fields: board_id, tag_id
    risk: high: external Miro API mutation; approval required
  create_orgs_org_id_teams_team_id_projects:
    endpoint: POST /v2/orgs/{{ record.org_id }}/teams/{{ record.team_id }}/projects
    required fields: org_id, team_id, name
    risk: medium: external Miro API mutation; approval required
  update_orgs_org_id_teams_team_id_projects_project_id:
    endpoint: PATCH /v2/orgs/{{ record.org_id }}/teams/{{ record.team_id }}/projects/{{ record.project_id }}
    required fields: org_id, team_id, project_id, name
    risk: medium: external Miro API mutation; approval required
  delete_orgs_org_id_teams_team_id_projects_project_id:
    endpoint: DELETE /v2/orgs/{{ record.org_id }}/teams/{{ record.team_id }}/projects/{{ record.project_id }}
    required fields: org_id, team_id, project_id
    risk: high: external Miro API mutation; approval required
  update_orgs_org_id_teams_team_id_projects_project_id_settings:
    endpoint: PATCH /v2/orgs/{{ record.org_id }}/teams/{{ record.team_id }}/projects/{{ record.project_id }}/settings
    required fields: org_id, team_id, project_id
    risk: medium: external Miro API mutation; approval required
  create_orgs_org_id_teams_team_id_projects_project_id_members:
    endpoint: POST /v2/orgs/{{ record.org_id }}/teams/{{ record.team_id }}/projects/{{ record.project_id }}/members
    required fields: org_id, team_id, project_id, email, role
    risk: medium: external Miro API mutation; approval required
  update_orgs_org_id_teams_team_id_projects_project_id_members_member_id:
    endpoint: PATCH /v2/orgs/{{ record.org_id }}/teams/{{ record.team_id }}/projects/{{ record.project_id }}/members/{{ record.member_id }}
    required fields: org_id, team_id, project_id, member_id
    risk: medium: external Miro API mutation; approval required
  delete_orgs_org_id_teams_team_id_projects_project_id_members_member_id:
    endpoint: DELETE /v2/orgs/{{ record.org_id }}/teams/{{ record.team_id }}/projects/{{ record.project_id }}/members/{{ record.member_id }}
    required fields: org_id, team_id, project_id, member_id
    risk: high: external Miro API mutation; approval required
  create_orgs_org_id_teams:
    endpoint: POST /v2/orgs/{{ record.org_id }}/teams
    required fields: org_id, name
    risk: medium: external Miro API mutation; approval required
  update_orgs_org_id_teams_team_id:
    endpoint: PATCH /v2/orgs/{{ record.org_id }}/teams/{{ record.team_id }}
    required fields: org_id, team_id
    risk: medium: external Miro API mutation; approval required
  delete_orgs_org_id_teams_team_id:
    endpoint: DELETE /v2/orgs/{{ record.org_id }}/teams/{{ record.team_id }}
    required fields: org_id, team_id
    risk: high: external Miro API mutation; approval required
  create_orgs_org_id_teams_team_id_members:
    endpoint: POST /v2/orgs/{{ record.org_id }}/teams/{{ record.team_id }}/members
    required fields: org_id, team_id, email
    risk: medium: external Miro API mutation; approval required
  update_orgs_org_id_teams_team_id_members_member_id:
    endpoint: PATCH /v2/orgs/{{ record.org_id }}/teams/{{ record.team_id }}/members/{{ record.member_id }}
    required fields: org_id, team_id, member_id
    risk: medium: external Miro API mutation; approval required
  delete_orgs_org_id_teams_team_id_members_member_id:
    endpoint: DELETE /v2/orgs/{{ record.org_id }}/teams/{{ record.team_id }}/members/{{ record.member_id }}
    required fields: org_id, team_id, member_id
    risk: high: external Miro API mutation; approval required
  update_orgs_org_id_teams_team_id_settings:
    endpoint: PATCH /v2/orgs/{{ record.org_id }}/teams/{{ record.team_id }}/settings
    required fields: org_id, team_id
    risk: medium: external Miro API mutation; approval required
  create_orgs_org_id_groups:
    endpoint: POST /v2/orgs/{{ record.org_id }}/groups
    required fields: org_id, name
    risk: medium: external Miro API mutation; approval required
  update_orgs_org_id_groups_group_id:
    endpoint: PATCH /v2/orgs/{{ record.org_id }}/groups/{{ record.group_id }}
    required fields: org_id, group_id
    risk: medium: external Miro API mutation; approval required
  delete_orgs_org_id_groups_group_id:
    endpoint: DELETE /v2/orgs/{{ record.org_id }}/groups/{{ record.group_id }}
    required fields: org_id, group_id
    risk: high: external Miro API mutation; approval required
  create_orgs_org_id_groups_group_id_members:
    endpoint: POST /v2/orgs/{{ record.org_id }}/groups/{{ record.group_id }}/members
    required fields: org_id, group_id, email
    risk: medium: external Miro API mutation; approval required
  update_orgs_org_id_groups_group_id_members:
    endpoint: PATCH /v2/orgs/{{ record.org_id }}/groups/{{ record.group_id }}/members
    required fields: org_id, group_id
    risk: medium: external Miro API mutation; approval required
  delete_orgs_org_id_groups_group_id_members_member_id:
    endpoint: DELETE /v2/orgs/{{ record.org_id }}/groups/{{ record.group_id }}/members/{{ record.member_id }}
    required fields: org_id, group_id, member_id
    risk: high: external Miro API mutation; approval required
  create_orgs_org_id_teams_team_id_groups:
    endpoint: POST /v2/orgs/{{ record.org_id }}/teams/{{ record.team_id }}/groups
    required fields: org_id, team_id, userGroupId, role
    risk: medium: external Miro API mutation; approval required
  delete_orgs_org_id_teams_team_id_groups_group_id:
    endpoint: DELETE /v2/orgs/{{ record.org_id }}/teams/{{ record.team_id }}/groups/{{ record.group_id }}
    required fields: org_id, team_id, group_id
    risk: high: external Miro API mutation; approval required
  create_orgs_org_id_boards_board_id_groups:
    endpoint: POST /v2/orgs/{{ record.org_id }}/boards/{{ record.board_id }}/groups
    required fields: org_id, board_id, userGroupIds, role
    risk: medium: external Miro API mutation; approval required
  delete_orgs_org_id_boards_board_id_groups_group_id:
    endpoint: DELETE /v2/orgs/{{ record.org_id }}/boards/{{ record.board_id }}/groups/{{ record.group_id }}
    required fields: org_id, board_id, group_id
    risk: high: external Miro API mutation; approval required
  create_orgs_org_id_projects_project_id_groups:
    endpoint: POST /v2/orgs/{{ record.org_id }}/projects/{{ record.project_id }}/groups
    required fields: org_id, project_id, userGroupIds, role
    risk: medium: external Miro API mutation; approval required
  delete_orgs_org_id_projects_project_id_groups_group_id:
    endpoint: DELETE /v2/orgs/{{ record.org_id }}/projects/{{ record.project_id }}/groups/{{ record.group_id }}
    required fields: org_id, project_id, group_id
    risk: high: external Miro API mutation; approval required

SECURITY
  read risk: external Miro API reads across board, enterprise, SCIM, user-group, project, and experimental resources
  write risk: external Miro API mutations including board sharing, item changes, enterprise administration, SCIM provisioning, groups, projects, and deletes
  approval: required for every write action; destructive deletes require destructive confirmation
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

COMMAND SURFACE
  Run Miro's declared streams and reverse-ETL actions.
  Usage: pm miro <command> [flags]
  Global flags:
    --approval-token-stdin (boolean): Read the approval token as one bounded line from standard input.
  Reverse ETL writes
  Other Commands
    create boards apply - Plan and execute the create boards reverse-ETL action. [intent=reverse_etl availability=implemented write=create_boards]; approval: requires plan, preview, approval, and execute; risk: medium: external Miro API mutation; approval required
    create boards board id app cards apply - Plan and execute the create boards board id app cards reverse-ETL action. [intent=reverse_etl availability=implemented write=create_boards_board_id_app_cards]; approval: requires plan, preview, approval, and execute; risk: medium: external Miro API mutation; approval required; flags: --board_id (required)
    create boards board id cards apply - Plan and execute the create boards board id cards reverse-ETL action. [intent=reverse_etl availability=implemented write=create_boards_board_id_cards]; approval: requires plan, preview, approval, and execute; risk: medium: external Miro API mutation; approval required; flags: --board_id (required)
    create boards board id connectors apply - Plan and execute the create boards board id connectors reverse-ETL action. [intent=reverse_etl availability=implemented write=create_boards_board_id_connectors]; approval: requires plan, preview, approval, and execute; risk: medium: external Miro API mutation; approval required; flags: --board_id (required), --endItem (required), --startItem (required)
    create boards board id docs apply - Plan and execute the create boards board id docs reverse-ETL action. [intent=reverse_etl availability=implemented write=create_boards_board_id_docs]; approval: requires plan, preview, approval, and execute; risk: medium: external Miro API mutation; approval required; flags: --board_id (required), --data (required)
    create boards board id documents apply - Plan and execute the create boards board id documents reverse-ETL action. [intent=reverse_etl availability=implemented write=create_boards_board_id_documents]; approval: requires plan, preview, approval, and execute; risk: medium: external Miro API mutation; approval required; flags: --board_id (required), --data (required)
    create boards board id embeds apply - Plan and execute the create boards board id embeds reverse-ETL action. [intent=reverse_etl availability=implemented write=create_boards_board_id_embeds]; approval: requires plan, preview, approval, and execute; risk: medium: external Miro API mutation; approval required; flags: --board_id (required), --data (required)
    create boards board id frames apply - Plan and execute the create boards board id frames reverse-ETL action. [intent=reverse_etl availability=implemented write=create_boards_board_id_frames]; approval: requires plan, preview, approval, and execute; risk: medium: external Miro API mutation; approval required; flags: --board_id (required), --data (required)
    create boards board id groups apply - Plan and execute the create boards board id groups reverse-ETL action. [intent=reverse_etl availability=implemented write=create_boards_board_id_groups]; approval: requires plan, preview, approval, and execute; risk: medium: external Miro API mutation; approval required; flags: --board_id (required), --id (required), --name (required), --type (required)
    create boards board id images apply - Plan and execute the create boards board id images reverse-ETL action. [intent=reverse_etl availability=implemented write=create_boards_board_id_images]; approval: requires plan, preview, approval, and execute; risk: medium: external Miro API mutation; approval required; flags: --board_id (required), --data (required)
    create boards board id members apply - Plan and execute the create boards board id members reverse-ETL action. [intent=reverse_etl availability=implemented write=create_boards_board_id_members]; approval: requires plan, preview, approval, and execute; risk: medium: external Miro API mutation; approval required; flags: --board_id (required), --emails (required)
    create boards board id shapes apply - Plan and execute the create boards board id shapes reverse-ETL action. [intent=reverse_etl availability=implemented write=create_boards_board_id_shapes]; approval: requires plan, preview, approval, and execute; risk: medium: external Miro API mutation; approval required; flags: --board_id (required)
    create boards board id sticky notes apply - Plan and execute the create boards board id sticky notes reverse-ETL action. [intent=reverse_etl availability=implemented write=create_boards_board_id_sticky_notes]; approval: requires plan, preview, approval, and execute; risk: medium: external Miro API mutation; approval required; flags: --board_id (required)
    create boards board id tags apply - Plan and execute the create boards board id tags reverse-ETL action. [intent=reverse_etl availability=implemented write=create_boards_board_id_tags]; approval: requires plan, preview, approval, and execute; risk: medium: external Miro API mutation; approval required; flags: --board_id (required), --title (required)
    create boards board id texts apply - Plan and execute the create boards board id texts reverse-ETL action. [intent=reverse_etl availability=implemented write=create_boards_board_id_texts]; approval: requires plan, preview, approval, and execute; risk: medium: external Miro API mutation; approval required; flags: --board_id (required), --data (required)
    create experimental boards board id code widgets apply - Plan and execute the create experimental boards board id code widgets reverse-ETL action. [intent=reverse_etl availability=implemented write=create_experimental_boards_board_id_code_widgets]; approval: requires plan, preview, approval, and execute; risk: medium: external Miro API mutation; approval required; flags: --board_id (required)
    create experimental boards board id mindmap nodes apply - Plan and execute the create experimental boards board id mindmap nodes reverse-ETL action. [intent=reverse_etl availability=implemented write=create_experimental_boards_board_id_mindmap_nodes]; approval: requires plan, preview, approval, and execute; risk: medium: external Miro API mutation; approval required; flags: --board_id (required), --data (required)
    create experimental boards board id shapes apply - Plan and execute the create experimental boards board id shapes reverse-ETL action. [intent=reverse_etl availability=implemented write=create_experimental_boards_board_id_shapes]; approval: requires plan, preview, approval, and execute; risk: medium: external Miro API mutation; approval required; flags: --board_id (required)
    create orgs org id boards board id groups apply - Plan and execute the create orgs org id boards board id groups reverse-ETL action. [intent=reverse_etl availability=implemented write=create_orgs_org_id_boards_board_id_groups]; approval: requires plan, preview, approval, and execute; risk: medium: external Miro API mutation; approval required; flags: --board_id (required), --org_id (required), --role (required), --userGroupIds (required)
    create orgs org id boards export jobs job id tasks task id export link apply - Plan and execute the create orgs org id boards export jobs job id tasks task id export link reverse-ETL action. [intent=reverse_etl availability=implemented write=create_orgs_org_id_boards_export_jobs_job_id_tasks_task_id_export_link]; approval: requires plan, preview, approval, and execute; risk: medium: external Miro API mutation; approval required; flags: --job_id (required), --org_id (required), --task_id (required)
    create orgs org id cases apply - Plan and execute the create orgs org id cases reverse-ETL action. [intent=reverse_etl availability=implemented write=create_orgs_org_id_cases]; approval: requires plan, preview, approval, and execute; risk: medium: external Miro API mutation; approval required; flags: --name (required), --org_id (required)
    create orgs org id cases case id legal holds apply - Plan and execute the create orgs org id cases case id legal holds reverse-ETL action. [intent=reverse_etl availability=implemented write=create_orgs_org_id_cases_case_id_legal_holds]; approval: requires plan, preview, approval, and execute; risk: medium: external Miro API mutation; approval required; flags: --case_id (required), --name (required), --org_id (required), --scope (required)
    create orgs org id groups apply - Plan and execute the create orgs org id groups reverse-ETL action. [intent=reverse_etl availability=implemented write=create_orgs_org_id_groups]; approval: requires plan, preview, approval, and execute; risk: medium: external Miro API mutation; approval required; flags: --name (required), --org_id (required)
    create orgs org id groups group id members apply - Plan and execute the create orgs org id groups group id members reverse-ETL action. [intent=reverse_etl availability=implemented write=create_orgs_org_id_groups_group_id_members]; approval: requires plan, preview, approval, and execute; risk: medium: external Miro API mutation; approval required; flags: --email (required), --group_id (required), --org_id (required)
    create orgs org id projects project id groups apply - Plan and execute the create orgs org id projects project id groups reverse-ETL action. [intent=reverse_etl availability=implemented write=create_orgs_org_id_projects_project_id_groups]; approval: requires plan, preview, approval, and execute; risk: medium: external Miro API mutation; approval required; flags: --org_id (required), --project_id (required), --role (required), --userGroupIds (required)
    create orgs org id teams apply - Plan and execute the create orgs org id teams reverse-ETL action. [intent=reverse_etl availability=implemented write=create_orgs_org_id_teams]; approval: requires plan, preview, approval, and execute; risk: medium: external Miro API mutation; approval required; flags: --name (required), --org_id (required)
    create orgs org id teams team id boards board id data classification apply - Plan and execute the create orgs org id teams team id boards board id data classification reverse-ETL action. [intent=reverse_etl availability=implemented write=create_orgs_org_id_teams_team_id_boards_board_id_data_classification]; approval: requires plan, preview, approval, and execute; risk: medium: external Miro API mutation; approval required; flags: --board_id (required), --org_id (required), --team_id (required)
    create orgs org id teams team id groups apply - Plan and execute the create orgs org id teams team id groups reverse-ETL action. [intent=reverse_etl availability=implemented write=create_orgs_org_id_teams_team_id_groups]; approval: requires plan, preview, approval, and execute; risk: medium: external Miro API mutation; approval required; flags: --org_id (required), --role (required), --team_id (required), --userGroupId (required)
    create orgs org id teams team id members apply - Plan and execute the create orgs org id teams team id members reverse-ETL action. [intent=reverse_etl availability=implemented write=create_orgs_org_id_teams_team_id_members]; approval: requires plan, preview, approval, and execute; risk: medium: external Miro API mutation; approval required; flags: --email (required), --org_id (required), --team_id (required)
    create orgs org id teams team id projects apply - Plan and execute the create orgs org id teams team id projects reverse-ETL action. [intent=reverse_etl availability=implemented write=create_orgs_org_id_teams_team_id_projects]; approval: requires plan, preview, approval, and execute; risk: medium: external Miro API mutation; approval required; flags: --name (required), --org_id (required), --team_id (required)
    create orgs org id teams team id projects project id members apply - Plan and execute the create orgs org id teams team id projects project id members reverse-ETL action. [intent=reverse_etl availability=implemented write=create_orgs_org_id_teams_team_id_projects_project_id_members]; approval: requires plan, preview, approval, and execute; risk: medium: external Miro API mutation; approval required; flags: --email (required), --org_id (required), --project_id (required), --role (required), --team_id (required)
    create users apply - Plan and execute the create users reverse-ETL action. [intent=reverse_etl availability=implemented write=create_users]; approval: requires plan, preview, approval, and execute; risk: medium: external Miro API mutation; approval required; flags: --userName (required)
    delete boards board id app cards item id apply - Plan and execute the delete boards board id app cards item id reverse-ETL action. [intent=reverse_etl availability=implemented write=delete_boards_board_id_app_cards_item_id]; approval: requires plan, preview, approval, and execute; risk: high: external Miro API mutation; approval required; flags: --board_id (required), --item_id (required)
    delete boards board id apply - Plan and execute the delete boards board id reverse-ETL action. [intent=reverse_etl availability=implemented write=delete_boards_board_id]; approval: requires plan, preview, approval, and execute; risk: high: external Miro API mutation; approval required; flags: --board_id (required)
    delete boards board id cards item id apply - Plan and execute the delete boards board id cards item id reverse-ETL action. [intent=reverse_etl availability=implemented write=delete_boards_board_id_cards_item_id]; approval: requires plan, preview, approval, and execute; risk: high: external Miro API mutation; approval required; flags: --board_id (required), --item_id (required)
    delete boards board id connectors connector id apply - Plan and execute the delete boards board id connectors connector id reverse-ETL action. [intent=reverse_etl availability=implemented write=delete_boards_board_id_connectors_connector_id]; approval: requires plan, preview, approval, and execute; risk: high: external Miro API mutation; approval required; flags: --board_id (required), --connector_id (required)
    delete boards board id docs item id apply - Plan and execute the delete boards board id docs item id reverse-ETL action. [intent=reverse_etl availability=implemented write=delete_boards_board_id_docs_item_id]; approval: requires plan, preview, approval, and execute; risk: high: external Miro API mutation; approval required; flags: --board_id (required), --item_id (required)
    delete boards board id documents item id apply - Plan and execute the delete boards board id documents item id reverse-ETL action. [intent=reverse_etl availability=implemented write=delete_boards_board_id_documents_item_id]; approval: requires plan, preview, approval, and execute; risk: high: external Miro API mutation; approval required; flags: --board_id (required), --item_id (required)
    delete boards board id embeds item id apply - Plan and execute the delete boards board id embeds item id reverse-ETL action. [intent=reverse_etl availability=implemented write=delete_boards_board_id_embeds_item_id]; approval: requires plan, preview, approval, and execute; risk: high: external Miro API mutation; approval required; flags: --board_id (required), --item_id (required)
    delete boards board id frames item id apply - Plan and execute the delete boards board id frames item id reverse-ETL action. [intent=reverse_etl availability=implemented write=delete_boards_board_id_frames_item_id]; approval: requires plan, preview, approval, and execute; risk: high: external Miro API mutation; approval required; flags: --board_id (required), --item_id (required)
    delete boards board id groups group id apply - Plan and execute the delete boards board id groups group id reverse-ETL action. [intent=reverse_etl availability=implemented write=delete_boards_board_id_groups_group_id]; approval: requires plan, preview, approval, and execute; risk: high: external Miro API mutation; approval required; flags: --board_id (required), --group_id (required)
    delete boards board id images item id apply - Plan and execute the delete boards board id images item id reverse-ETL action. [intent=reverse_etl availability=implemented write=delete_boards_board_id_images_item_id]; approval: requires plan, preview, approval, and execute; risk: high: external Miro API mutation; approval required; flags: --board_id (required), --item_id (required)
    delete boards board id items item id apply - Plan and execute the delete boards board id items item id reverse-ETL action. [intent=reverse_etl availability=implemented write=delete_boards_board_id_items_item_id]; approval: requires plan, preview, approval, and execute; risk: high: external Miro API mutation; approval required; flags: --board_id (required), --item_id (required)
    delete boards board id members board member id apply - Plan and execute the delete boards board id members board member id reverse-ETL action. [intent=reverse_etl availability=implemented write=delete_boards_board_id_members_board_member_id]; approval: requires plan, preview, approval, and execute; risk: high: external Miro API mutation; approval required; flags: --board_id (required), --board_member_id (required)
    delete boards board id shapes item id apply - Plan and execute the delete boards board id shapes item id reverse-ETL action. [intent=reverse_etl availability=implemented write=delete_boards_board_id_shapes_item_id]; approval: requires plan, preview, approval, and execute; risk: high: external Miro API mutation; approval required; flags: --board_id (required), --item_id (required)
    delete boards board id sticky notes item id apply - Plan and execute the delete boards board id sticky notes item id reverse-ETL action. [intent=reverse_etl availability=implemented write=delete_boards_board_id_sticky_notes_item_id]; approval: requires plan, preview, approval, and execute; risk: high: external Miro API mutation; approval required; flags: --board_id (required), --item_id (required)
    delete boards board id tags tag id apply - Plan and execute the delete boards board id tags tag id reverse-ETL action. [intent=reverse_etl availability=implemented write=delete_boards_board_id_tags_tag_id]; approval: requires plan, preview, approval, and execute; risk: high: external Miro API mutation; approval required; flags: --board_id (required), --tag_id (required)
    delete boards board id texts item id apply - Plan and execute the delete boards board id texts item id reverse-ETL action. [intent=reverse_etl availability=implemented write=delete_boards_board_id_texts_item_id]; approval: requires plan, preview, approval, and execute; risk: high: external Miro API mutation; approval required; flags: --board_id (required), --item_id (required)
    delete experimental boards board id code widgets item id apply - Plan and execute the delete experimental boards board id code widgets item id reverse-ETL action. [intent=reverse_etl availability=implemented write=delete_experimental_boards_board_id_code_widgets_item_id]; approval: requires plan, preview, approval, and execute; risk: high: external Miro API mutation; approval required; flags: --board_id (required), --item_id (required)
    delete experimental boards board id items item id apply - Plan and execute the delete experimental boards board id items item id reverse-ETL action. [intent=reverse_etl availability=implemented write=delete_experimental_boards_board_id_items_item_id]; approval: requires plan, preview, approval, and execute; risk: high: external Miro API mutation; approval required; flags: --board_id (required), --item_id (required)
    delete experimental boards board id mindmap nodes item id apply - Plan and execute the delete experimental boards board id mindmap nodes item id reverse-ETL action. [intent=reverse_etl availability=implemented write=delete_experimental_boards_board_id_mindmap_nodes_item_id]; approval: requires plan, preview, approval, and execute; risk: high: external Miro API mutation; approval required; flags: --board_id (required), --item_id (required)
    delete experimental boards board id shapes item id apply - Plan and execute the delete experimental boards board id shapes item id reverse-ETL action. [intent=reverse_etl availability=implemented write=delete_experimental_boards_board_id_shapes_item_id]; approval: requires plan, preview, approval, and execute; risk: high: external Miro API mutation; approval required; flags: --board_id (required), --item_id (required)
    delete orgs org id boards board id groups group id apply - Plan and execute the delete orgs org id boards board id groups group id reverse-ETL action. [intent=reverse_etl availability=implemented write=delete_orgs_org_id_boards_board_id_groups_group_id]; approval: requires plan, preview, approval, and execute; risk: high: external Miro API mutation; approval required; flags: --board_id (required), --group_id (required), --org_id (required)
    delete orgs org id cases case id apply - Plan and execute the delete orgs org id cases case id reverse-ETL action. [intent=reverse_etl availability=implemented write=delete_orgs_org_id_cases_case_id]; approval: requires plan, preview, approval, and execute; risk: high: external Miro API mutation; approval required; flags: --case_id (required), --org_id (required)
    delete orgs org id cases case id legal holds legal hold id apply - Plan and execute the delete orgs org id cases case id legal holds legal hold id reverse-ETL action. [intent=reverse_etl availability=implemented write=delete_orgs_org_id_cases_case_id_legal_holds_legal_hold_id]; approval: requires plan, preview, approval, and execute; risk: high: external Miro API mutation; approval required; flags: --case_id (required), --legal_hold_id (required), --org_id (required)
    delete orgs org id groups group id apply - Plan and execute the delete orgs org id groups group id reverse-ETL action. [intent=reverse_etl availability=implemented write=delete_orgs_org_id_groups_group_id]; approval: requires plan, preview, approval, and execute; risk: high: external Miro API mutation; approval required; flags: --group_id (required), --org_id (required)
    delete orgs org id groups group id members member id apply - Plan and execute the delete orgs org id groups group id members member id reverse-ETL action. [intent=reverse_etl availability=implemented write=delete_orgs_org_id_groups_group_id_members_member_id]; approval: requires plan, preview, approval, and execute; risk: high: external Miro API mutation; approval required; flags: --group_id (required), --member_id (required), --org_id (required)
    delete orgs org id projects project id groups group id apply - Plan and execute the delete orgs org id projects project id groups group id reverse-ETL action. [intent=reverse_etl availability=implemented write=delete_orgs_org_id_projects_project_id_groups_group_id]; approval: requires plan, preview, approval, and execute; risk: high: external Miro API mutation; approval required; flags: --group_id (required), --org_id (required), --project_id (required)
    delete orgs org id teams team id apply - Plan and execute the delete orgs org id teams team id reverse-ETL action. [intent=reverse_etl availability=implemented write=delete_orgs_org_id_teams_team_id]; approval: requires plan, preview, approval, and execute; risk: high: external Miro API mutation; approval required; flags: --org_id (required), --team_id (required)
    delete orgs org id teams team id groups group id apply - Plan and execute the delete orgs org id teams team id groups group id reverse-ETL action. [intent=reverse_etl availability=implemented write=delete_orgs_org_id_teams_team_id_groups_group_id]; approval: requires plan, preview, approval, and execute; risk: high: external Miro API mutation; approval required; flags: --group_id (required), --org_id (required), --team_id (required)
    delete orgs org id teams team id members member id apply - Plan and execute the delete orgs org id teams team id members member id reverse-ETL action. [intent=reverse_etl availability=implemented write=delete_orgs_org_id_teams_team_id_members_member_id]; approval: requires plan, preview, approval, and execute; risk: high: external Miro API mutation; approval required; flags: --member_id (required), --org_id (required), --team_id (required)
    delete orgs org id teams team id projects project id apply - Plan and execute the delete orgs org id teams team id projects project id reverse-ETL action. [intent=reverse_etl availability=implemented write=delete_orgs_org_id_teams_team_id_projects_project_id]; approval: requires plan, preview, approval, and execute; risk: high: external Miro API mutation; approval required; flags: --org_id (required), --project_id (required), --team_id (required)
    delete orgs org id teams team id projects project id members member id apply - Plan and execute the delete orgs org id teams team id projects project id members member id reverse-ETL action. [intent=reverse_etl availability=implemented write=delete_orgs_org_id_teams_team_id_projects_project_id_members_member_id]; approval: requires plan, preview, approval, and execute; risk: high: external Miro API mutation; approval required; flags: --member_id (required), --org_id (required), --project_id (required), --team_id (required)
    delete users id apply - Plan and execute the delete users id reverse-ETL action. [intent=reverse_etl availability=implemented write=delete_users_id]; approval: requires plan, preview, approval, and execute; risk: high: external Miro API mutation; approval required; flags: --id (required)
    update boards board id app cards item id apply - Plan and execute the update boards board id app cards item id reverse-ETL action. [intent=reverse_etl availability=implemented write=update_boards_board_id_app_cards_item_id]; approval: requires plan, preview, approval, and execute; risk: medium: external Miro API mutation; approval required; flags: --board_id (required), --item_id (required)
    update boards board id apply - Plan and execute the update boards board id reverse-ETL action. [intent=reverse_etl availability=implemented write=update_boards_board_id]; approval: requires plan, preview, approval, and execute; risk: medium: external Miro API mutation; approval required; flags: --board_id (required)
    update boards board id cards item id apply - Plan and execute the update boards board id cards item id reverse-ETL action. [intent=reverse_etl availability=implemented write=update_boards_board_id_cards_item_id]; approval: requires plan, preview, approval, and execute; risk: medium: external Miro API mutation; approval required; flags: --board_id (required), --item_id (required)
    update boards board id connectors connector id apply - Plan and execute the update boards board id connectors connector id reverse-ETL action. [intent=reverse_etl availability=implemented write=update_boards_board_id_connectors_connector_id]; approval: requires plan, preview, approval, and execute; risk: medium: external Miro API mutation; approval required; flags: --board_id (required), --connector_id (required)
    update boards board id documents item id apply - Plan and execute the update boards board id documents item id reverse-ETL action. [intent=reverse_etl availability=implemented write=update_boards_board_id_documents_item_id]; approval: requires plan, preview, approval, and execute; risk: medium: external Miro API mutation; approval required; flags: --board_id (required), --item_id (required)
    update boards board id embeds item id apply - Plan and execute the update boards board id embeds item id reverse-ETL action. [intent=reverse_etl availability=implemented write=update_boards_board_id_embeds_item_id]; approval: requires plan, preview, approval, and execute; risk: medium: external Miro API mutation; approval required; flags: --board_id (required), --item_id (required)
    update boards board id frames item id apply - Plan and execute the update boards board id frames item id reverse-ETL action. [intent=reverse_etl availability=implemented write=update_boards_board_id_frames_item_id]; approval: requires plan, preview, approval, and execute; risk: medium: external Miro API mutation; approval required; flags: --board_id (required), --item_id (required)
    update boards board id groups group id apply - Plan and execute the update boards board id groups group id reverse-ETL action. [intent=reverse_etl availability=implemented write=update_boards_board_id_groups_group_id]; approval: requires plan, preview, approval, and execute; risk: medium: external Miro API mutation; approval required; flags: --board_id (required), --group_id (required), --id (required), --name (required), --type (required)
    update boards board id images item id apply - Plan and execute the update boards board id images item id reverse-ETL action. [intent=reverse_etl availability=implemented write=update_boards_board_id_images_item_id]; approval: requires plan, preview, approval, and execute; risk: medium: external Miro API mutation; approval required; flags: --board_id (required), --item_id (required)
    update boards board id items item id apply - Plan and execute the update boards board id items item id reverse-ETL action. [intent=reverse_etl availability=implemented write=update_boards_board_id_items_item_id]; approval: requires plan, preview, approval, and execute; risk: medium: external Miro API mutation; approval required; flags: --board_id (required), --item_id (required)
    update boards board id members board member id apply - Plan and execute the update boards board id members board member id reverse-ETL action. [intent=reverse_etl availability=implemented write=update_boards_board_id_members_board_member_id]; approval: requires plan, preview, approval, and execute; risk: medium: external Miro API mutation; approval required; flags: --board_id (required), --board_member_id (required)
    update boards board id shapes item id apply - Plan and execute the update boards board id shapes item id reverse-ETL action. [intent=reverse_etl availability=implemented write=update_boards_board_id_shapes_item_id]; approval: requires plan, preview, approval, and execute; risk: medium: external Miro API mutation; approval required; flags: --board_id (required), --item_id (required)
    update boards board id sticky notes item id apply - Plan and execute the update boards board id sticky notes item id reverse-ETL action. [intent=reverse_etl availability=implemented write=update_boards_board_id_sticky_notes_item_id]; approval: requires plan, preview, approval, and execute; risk: medium: external Miro API mutation; approval required; flags: --board_id (required), --item_id (required)
    update boards board id tags tag id apply - Plan and execute the update boards board id tags tag id reverse-ETL action. [intent=reverse_etl availability=implemented write=update_boards_board_id_tags_tag_id]; approval: requires plan, preview, approval, and execute; risk: medium: external Miro API mutation; approval required; flags: --board_id (required), --tag_id (required)
    update boards board id texts item id apply - Plan and execute the update boards board id texts item id reverse-ETL action. [intent=reverse_etl availability=implemented write=update_boards_board_id_texts_item_id]; approval: requires plan, preview, approval, and execute; risk: medium: external Miro API mutation; approval required; flags: --board_id (required), --item_id (required)
    update experimental boards board id code widgets item id apply - Plan and execute the update experimental boards board id code widgets item id reverse-ETL action. [intent=reverse_etl availability=implemented write=update_experimental_boards_board_id_code_widgets_item_id]; approval: requires plan, preview, approval, and execute; risk: medium: external Miro API mutation; approval required; flags: --board_id (required), --item_id (required)
    update experimental boards board id code widgets item id position apply - Plan and execute the update experimental boards board id code widgets item id position reverse-ETL action. [intent=reverse_etl availability=implemented write=update_experimental_boards_board_id_code_widgets_item_id_position]; approval: requires plan, preview, approval, and execute; risk: medium: external Miro API mutation; approval required; flags: --board_id (required), --item_id (required)
    update experimental boards board id shapes item id apply - Plan and execute the update experimental boards board id shapes item id reverse-ETL action. [intent=reverse_etl availability=implemented write=update_experimental_boards_board_id_shapes_item_id]; approval: requires plan, preview, approval, and execute; risk: medium: external Miro API mutation; approval required; flags: --board_id (required), --item_id (required)
    update groups id apply - Plan and execute the update groups id reverse-ETL action. [intent=reverse_etl availability=implemented write=update_groups_id]; approval: requires plan, preview, approval, and execute; risk: medium: external Miro API mutation; approval required; flags: --Operations (required), --id (required), --schemas (required)
    update orgs org id boards export jobs job id status apply - Plan and execute the update orgs org id boards export jobs job id status reverse-ETL action. [intent=reverse_etl availability=implemented write=update_orgs_org_id_boards_export_jobs_job_id_status]; approval: requires plan, preview, approval, and execute; risk: medium: external Miro API mutation; approval required; flags: --job_id (required), --org_id (required), --status (required)
    update orgs org id cases case id apply - Plan and execute the update orgs org id cases case id reverse-ETL action. [intent=reverse_etl availability=implemented write=update_orgs_org_id_cases_case_id]; approval: requires plan, preview, approval, and execute; risk: medium: external Miro API mutation; approval required; flags: --case_id (required), --name (required), --org_id (required)
    update orgs org id cases case id legal holds legal hold id apply - Plan and execute the update orgs org id cases case id legal holds legal hold id reverse-ETL action. [intent=reverse_etl availability=implemented write=update_orgs_org_id_cases_case_id_legal_holds_legal_hold_id]; approval: requires plan, preview, approval, and execute; risk: medium: external Miro API mutation; approval required; flags: --case_id (required), --legal_hold_id (required), --name (required), --org_id (required), --scope (required)
    update orgs org id groups group id apply - Plan and execute the update orgs org id groups group id reverse-ETL action. [intent=reverse_etl availability=implemented write=update_orgs_org_id_groups_group_id]; approval: requires plan, preview, approval, and execute; risk: medium: external Miro API mutation; approval required; flags: --group_id (required), --org_id (required)
    update orgs org id groups group id members apply - Plan and execute the update orgs org id groups group id members reverse-ETL action. [intent=reverse_etl availability=implemented write=update_orgs_org_id_groups_group_id_members]; approval: requires plan, preview, approval, and execute; risk: medium: external Miro API mutation; approval required; flags: --group_id (required), --org_id (required)
    update orgs org id teams team id apply - Plan and execute the update orgs org id teams team id reverse-ETL action. [intent=reverse_etl availability=implemented write=update_orgs_org_id_teams_team_id]; approval: requires plan, preview, approval, and execute; risk: medium: external Miro API mutation; approval required; flags: --org_id (required), --team_id (required)
    update orgs org id teams team id data classification apply - Plan and execute the update orgs org id teams team id data classification reverse-ETL action. [intent=reverse_etl availability=implemented write=update_orgs_org_id_teams_team_id_data_classification]; approval: requires plan, preview, approval, and execute; risk: medium: external Miro API mutation; approval required; flags: --org_id (required), --team_id (required)
    update orgs org id teams team id data classification settings apply - Plan and execute the update orgs org id teams team id data classification settings reverse-ETL action. [intent=reverse_etl availability=implemented write=update_orgs_org_id_teams_team_id_data_classification_settings]; approval: requires plan, preview, approval, and execute; risk: medium: external Miro API mutation; approval required; flags: --org_id (required), --team_id (required)
    update orgs org id teams team id members member id apply - Plan and execute the update orgs org id teams team id members member id reverse-ETL action. [intent=reverse_etl availability=implemented write=update_orgs_org_id_teams_team_id_members_member_id]; approval: requires plan, preview, approval, and execute; risk: medium: external Miro API mutation; approval required; flags: --member_id (required), --org_id (required), --team_id (required)
    update orgs org id teams team id projects project id apply - Plan and execute the update orgs org id teams team id projects project id reverse-ETL action. [intent=reverse_etl availability=implemented write=update_orgs_org_id_teams_team_id_projects_project_id]; approval: requires plan, preview, approval, and execute; risk: medium: external Miro API mutation; approval required; flags: --name (required), --org_id (required), --project_id (required), --team_id (required)
    update orgs org id teams team id projects project id members member id apply - Plan and execute the update orgs org id teams team id projects project id members member id reverse-ETL action. [intent=reverse_etl availability=implemented write=update_orgs_org_id_teams_team_id_projects_project_id_members_member_id]; approval: requires plan, preview, approval, and execute; risk: medium: external Miro API mutation; approval required; flags: --member_id (required), --org_id (required), --project_id (required), --team_id (required)
    update orgs org id teams team id projects project id settings apply - Plan and execute the update orgs org id teams team id projects project id settings reverse-ETL action. [intent=reverse_etl availability=implemented write=update_orgs_org_id_teams_team_id_projects_project_id_settings]; approval: requires plan, preview, approval, and execute; risk: medium: external Miro API mutation; approval required; flags: --org_id (required), --project_id (required), --team_id (required)
    update orgs org id teams team id settings apply - Plan and execute the update orgs org id teams team id settings reverse-ETL action. [intent=reverse_etl availability=implemented write=update_orgs_org_id_teams_team_id_settings]; approval: requires plan, preview, approval, and execute; risk: medium: external Miro API mutation; approval required; flags: --org_id (required), --team_id (required)
    update users id 2 apply - Plan and execute the update users id 2 reverse-ETL action. [intent=reverse_etl availability=implemented write=update_users_id_2]; approval: requires plan, preview, approval, and execute; risk: medium: external Miro API mutation; approval required; flags: --Operations (required), --id (required), --schemas (required)
    update users id apply - Plan and execute the update users id reverse-ETL action. [intent=reverse_etl availability=implemented write=update_users_id]; approval: requires plan, preview, approval, and execute; risk: medium: external Miro API mutation; approval required; flags: --id (required)

EXAMPLES
  # Inspect as a manual
  pm connectors inspect miro

  # Inspect as structured JSON
  pm connectors inspect miro --json

AGENT WORKFLOW
  - Run pm connectors inspect miro before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
