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
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics

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
  api_key (secret)

ETL STREAMS
  boards:
    primary key: id
    fields: created_at(), description(), id(), modified_at(), name(), owner_id(), team_id(), type(), view_link()
  board_users:
    primary key: id
    fields: board_id(), id(), name(), role(), type()
  board_items:
    primary key: id
    fields: board_id(), created_at(), id(), modified_at(), type()
  board_tags:
    primary key: id
    fields: board_id(), fill_color(), id(), title(), type()
  board_connectors:
    primary key: id
    fields: board_id(), id(), shape(), type()
  orgs_org_id_ai_interaction_logs:
    primary key: id
    fields: actor(), aiFeatureName(), createdAt(), details(), id(), logType(), messageId(), object(), sessionId(), storedAt()
  audit_logs:
    primary key: id
    fields: category(), context(), createdAt(), createdBy(), details(), event(), id(), object()
  orgs_org_id_data_classification_settings:
    fields: enabled(), labels(), type()
  orgs_org_id_teams_team_id_data_classification_settings:
    fields: defaultLabelId(), enabled(), type()
  orgs_org_id_teams_team_id_boards_board_id_data_classification:
    primary key: id
    fields: color(), description(), guidelineUrl(), id(), name(), sharingRecommendation(), type()
  boards_board_id_docs_item_id:
    primary key: id
    fields: createdAt(), createdBy(), data(), id(), links(), modifiedAt(), modifiedBy(), parent(), position(), type()
  orgs_org_id_cases:
    primary key: id
    fields: createdAt(), createdBy(), description(), id(), lastModifiedAt(), lastModifiedBy(), name(), organizationId()
  orgs_org_id_cases_case_id:
    primary key: id
    fields: createdAt(), createdBy(), description(), id(), lastModifiedAt(), lastModifiedBy(), name(), organizationId()
  orgs_org_id_cases_case_id_legal_holds:
    primary key: id
    fields: caseId(), createdAt(), createdBy(), description(), id(), lastModifiedAt(), lastModifiedBy(), name(), organizationId(), scope(), state()
  orgs_org_id_cases_case_id_export_jobs:
    primary key: id
    fields: id()
  orgs_org_id_cases_case_id_legal_holds_legal_hold_id:
    primary key: id
    fields: caseId(), createdAt(), createdBy(), description(), id(), lastModifiedAt(), lastModifiedBy(), name(), organizationId(), scope(), state()
  orgs_org_id_cases_case_id_legal_holds_legal_hold_id_content_items:
    fields: contentId(), type()
  orgs_org_id_boards_export_jobs:
    primary key: id
    fields: boardFormat(), createdAt(), creator(), id(), modifiedAt(), name(), status(), tasksCount()
  orgs_org_id_boards_export_jobs_job_id:
    fields: jobStatus()
  orgs_org_id_boards_export_jobs_job_id_results:
    fields: boardId(), errorMessage(), errorType(), exportLink(), status()
  orgs_org_id_boards_export_jobs_job_id_tasks:
    primary key: id
    fields: artifactExpiredAt(), board(), errorMessage(), errorType(), id(), sizeInBytes(), status()
  orgs_org_id_content_logs_items:
    primary key: id
    fields: actionTime(), actionType(), actor(), contentId(), id(), itemId(), itemType(), relationships(), state()
  users:
    primary key: id
    fields: active(), displayName(), emails(), groups(), id(), meta(), name(), photos(), preferredLanguage(), roles(), schemas(), urn:ietf:params:scim:schemas:extension:enterprise:2.0:User(), userName(), userType()
  users_id:
    primary key: id
    fields: active(), displayName(), emails(), groups(), id(), meta(), name(), photos(), preferredLanguage(), roles(), schemas(), urn:ietf:params:scim:schemas:extension:enterprise:2.0:User(), userName(), userType()
  groups:
    primary key: id
    fields: displayName(), id(), members(), meta(), schemas()
  groups_id:
    primary key: id
    fields: displayName(), id(), members(), meta(), schemas()
  service_provider_config:
    fields: authenticationSchemes(), bulk(), changePassword(), documentationUri(), etag(), filter(), patch(), schemas(), sort()
  resource_types:
    primary key: id
    fields: description(), endpoint(), id(), name(), schema(), schemaExtensions(), schemas()
  resource_types_resource:
    primary key: id
    fields: description(), endpoint(), id(), name(), schema(), schemaExtensions(), schemas()
  schemas:
    primary key: id
    fields: attributes(), description(), id(), meta(), name(), schemas()
  schemas_uri:
    primary key: id
    fields: attributes(), description(), id(), meta(), name()
  orgs_org_id:
    primary key: id
    fields: fullLicensesPurchased(), id(), name(), plan(), type()
  orgs_org_id_members:
    primary key: id
    fields: active(), adminRoles(), email(), id(), lastActivityAt(), license(), licenseAssignedAt(), role(), type()
  orgs_org_id_members_member_id:
    primary key: id
    fields: active(), adminRoles(), email(), id(), lastActivityAt(), license(), licenseAssignedAt(), role(), type()
  boards_board_id:
    primary key: id
    fields: createdAt(), createdBy(), currentUserMembership(), description(), id(), lastOpenedAt(), lastOpenedBy(), links(), modifiedAt(), modifiedBy(), name(), owner(), picture(), policy(), project(), team(), type(), viewLink()
  boards_board_id_app_cards_item_id:
    primary key: id
    fields: createdAt(), createdBy(), data(), geometry(), id(), links(), modifiedAt(), modifiedBy(), parent(), position(), style(), type()
  boards_board_id_cards_item_id:
    primary key: id
    fields: createdAt(), createdBy(), data(), geometry(), id(), links(), modifiedAt(), modifiedBy(), parent(), position(), style(), type()
  boards_board_id_connectors_connector_id:
    primary key: id
    fields: captions(), createdAt(), createdBy(), endItem(), id(), isSupported(), links(), modifiedAt(), modifiedBy(), shape(), startItem(), style(), type()
  boards_board_id_documents_item_id:
    primary key: id
    fields: createdAt(), createdBy(), data(), geometry(), id(), links(), modifiedAt(), modifiedBy(), parent(), position(), type()
  boards_board_id_embeds_item_id:
    primary key: id
    fields: createdAt(), createdBy(), data(), geometry(), id(), links(), modifiedAt(), modifiedBy(), parent(), position(), type()
  boards_board_id_images_item_id:
    primary key: id
    fields: createdAt(), createdBy(), data(), geometry(), id(), links(), modifiedAt(), modifiedBy(), parent(), position(), type()
  boards_board_id_items_item_id:
    primary key: id
    fields: createdAt(), createdBy(), data(), geometry(), id(), modifiedAt(), modifiedBy(), parent(), position(), type()
  boards_board_id_members_board_member_id:
    primary key: id
    fields: id(), links(), name(), role(), type()
  boards_board_id_shapes_item_id:
    primary key: id
    fields: createdAt(), createdBy(), data(), geometry(), id(), links(), modifiedAt(), modifiedBy(), parent(), position(), style(), type()
  boards_board_id_sticky_notes_item_id:
    primary key: id
    fields: createdAt(), createdBy(), data(), geometry(), id(), links(), modifiedAt(), modifiedBy(), parent(), position(), style(), type()
  boards_board_id_texts_item_id:
    primary key: id
    fields: createdAt(), createdBy(), data(), geometry(), id(), links(), modifiedAt(), modifiedBy(), parent(), position(), style(), type()
  boards_board_id_frames_item_id:
    primary key: id
    fields: createdAt(), createdBy(), data(), geometry(), id(), links(), modifiedAt(), modifiedBy(), position(), style(), type()
  boards_board_id_platform_containers_items:
    primary key: id
    fields: createdAt(), createdBy(), data(), geometry(), id(), modifiedAt(), modifiedBy(), parent(), position(), type()
  experimental_apps_app_id_metrics:
    fields: installations(), periodStart(), uninstallations(), uniqueOrganizations(), uniqueRecurringUsers(), uniqueUsers()
  experimental_apps_app_id_metrics_total:
    fields: installations(), uninstallations(), uniqueOrganizations(), uniqueRecurringUsers(), uniqueUsers()
  experimental_boards_board_id_mindmap_nodes_item_id:
    primary key: id
    fields: createdAt(), createdBy(), data(), id(), links(), modifiedAt(), modifiedBy(), parent(), style(), type()
  experimental_boards_board_id_mindmap_nodes:
    primary key: id
    fields: createdAt(), createdBy(), data(), id(), links(), modifiedAt(), modifiedBy(), parent(), style(), type()
  experimental_boards_board_id_items:
    primary key: id
    fields: createdAt(), createdBy(), data(), geometry(), id(), modifiedAt(), modifiedBy(), parent(), position(), type()
  experimental_boards_board_id_items_item_id:
    primary key: id
    fields: createdAt(), createdBy(), data(), geometry(), id(), modifiedAt(), modifiedBy(), parent(), position(), type()
  experimental_boards_board_id_shapes_item_id:
    primary key: id
    fields: createdAt(), createdBy(), data(), geometry(), id(), links(), modifiedAt(), modifiedBy(), parent(), position(), style(), type()
  experimental_boards_board_id_code_widgets:
    primary key: id
    fields: createdAt(), createdBy(), data(), geometry(), id(), links(), modifiedAt(), modifiedBy(), position(), type()
  experimental_boards_board_id_code_widgets_item_id:
    primary key: id
    fields: createdAt(), createdBy(), data(), geometry(), id(), links(), modifiedAt(), modifiedBy(), position(), type()
  boards_board_id_groups:
    primary key: id
    fields: data(), id(), links(), type()
  boards_board_id_groups_items:
    fields: data(), limit(), links(), offset(), size(), total(), type()
  boards_board_id_groups_group_id:
    primary key: id
    fields: data(), id(), links(), type()
  boards_board_id_items_item_id_tags:
    primary key: id
    fields: fillColor(), id(), title(), type()
  boards_board_id_tags_tag_id:
    primary key: id
    fields: fillColor(), id(), links(), title(), type()
  boards_board_id_platform_tags_items:
    primary key: id
    fields: createdAt(), createdBy(), data(), geometry(), id(), modifiedAt(), modifiedBy(), parent(), position(), type()
  orgs_org_id_teams_team_id_projects:
    primary key: id
    fields: id(), name(), type()
  orgs_org_id_teams_team_id_projects_project_id:
    primary key: id
    fields: id(), name(), type()
  orgs_org_id_teams_team_id_projects_project_id_settings:
    fields: sharingPolicySettings(), type()
  orgs_org_id_teams_team_id_projects_project_id_members:
    primary key: id
    fields: email(), id(), role(), type()
  orgs_org_id_teams_team_id_projects_project_id_members_member_id:
    primary key: id
    fields: email(), id(), role(), type()
  orgs_org_id_teams:
    primary key: id
    fields: id(), name(), picture(), type()
  orgs_org_id_teams_team_id:
    primary key: id
    fields: id(), name(), picture(), type()
  orgs_org_id_teams_team_id_members:
    primary key: id
    fields: createdAt(), createdBy(), id(), modifiedAt(), modifiedBy(), role(), teamId(), type()
  orgs_org_id_teams_team_id_members_member_id:
    primary key: id
    fields: createdAt(), createdBy(), id(), modifiedAt(), modifiedBy(), role(), teamId(), type()
  orgs_org_id_default_teams_settings:
    fields: organizationId(), teamAccountDiscoverySettings(), teamCollaborationSettings(), teamCopyAccessLevelSettings(), teamId(), teamInvitationSettings(), teamSharingPolicySettings(), type()
  orgs_org_id_teams_team_id_settings:
    fields: organizationId(), teamAccountDiscoverySettings(), teamCollaborationSettings(), teamCopyAccessLevelSettings(), teamId(), teamInvitationSettings(), teamSharingPolicySettings(), type()
  orgs_org_id_groups:
    primary key: id
    fields: description(), id(), name(), type()
  orgs_org_id_groups_group_id:
    primary key: id
    fields: description(), id(), name(), type()
  orgs_org_id_groups_group_id_members:
    primary key: id
    fields: email(), id(), type()
  orgs_org_id_groups_group_id_members_member_id:
    primary key: id
    fields: email(), id(), type()
  orgs_org_id_groups_group_id_teams:
    primary key: id
    fields: id(), role(), type()
  orgs_org_id_groups_group_id_teams_team_id:
    primary key: id
    fields: id(), role(), type()
  orgs_org_id_teams_team_id_groups:
    primary key: id
    fields: id(), role(), type()
  orgs_org_id_teams_team_id_groups_group_id:
    primary key: id
    fields: id(), role(), type()
  orgs_org_id_boards_board_id_groups:
    primary key: id
    fields: id(), role(), type()
  orgs_org_id_projects_project_id_groups:
    primary key: id
    fields: id(), role(), type()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

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
    required fields: board_id
    risk: medium: external Miro API mutation; approval required
  delete_boards_board_id_docs_item_id:
    endpoint: DELETE /v2/boards/{{ record.board_id }}/docs/{{ record.item_id }}
    required fields: board_id, item_id
    risk: high: external Miro API mutation; approval required
  create_orgs_org_id_cases:
    endpoint: POST /v2/orgs/{{ record.org_id }}/cases
    required fields: org_id
    risk: medium: external Miro API mutation; approval required
  update_orgs_org_id_cases_case_id:
    endpoint: PUT /v2/orgs/{{ record.org_id }}/cases/{{ record.case_id }}
    required fields: org_id, case_id
    risk: medium: external Miro API mutation; approval required
  delete_orgs_org_id_cases_case_id:
    endpoint: DELETE /v2/orgs/{{ record.org_id }}/cases/{{ record.case_id }}
    required fields: org_id, case_id
    risk: high: external Miro API mutation; approval required
  create_orgs_org_id_cases_case_id_legal_holds:
    endpoint: POST /v2/orgs/{{ record.org_id }}/cases/{{ record.case_id }}/legal-holds
    required fields: org_id, case_id
    risk: medium: external Miro API mutation; approval required
  update_orgs_org_id_cases_case_id_legal_holds_legal_hold_id:
    endpoint: PUT /v2/orgs/{{ record.org_id }}/cases/{{ record.case_id }}/legal-holds/{{ record.legal_hold_id }}
    required fields: org_id, case_id, legal_hold_id
    risk: medium: external Miro API mutation; approval required
  delete_orgs_org_id_cases_case_id_legal_holds_legal_hold_id:
    endpoint: DELETE /v2/orgs/{{ record.org_id }}/cases/{{ record.case_id }}/legal-holds/{{ record.legal_hold_id }}
    required fields: org_id, case_id, legal_hold_id
    risk: high: external Miro API mutation; approval required
  update_orgs_org_id_boards_export_jobs_job_id_status:
    endpoint: PUT /v2/orgs/{{ record.org_id }}/boards/export/jobs/{{ record.job_id }}/status
    required fields: org_id, job_id
    risk: medium: external Miro API mutation; approval required
  create_orgs_org_id_boards_export_jobs_job_id_tasks_task_id_export_link:
    endpoint: POST /v2/orgs/{{ record.org_id }}/boards/export/jobs/{{ record.job_id }}/tasks/{{ record.task_id }}/export-link
    required fields: org_id, job_id, task_id
    risk: medium: external Miro API mutation; approval required
  create_users:
    endpoint: POST /Users
    risk: medium: external Miro API mutation; approval required
  update_users_id:
    endpoint: PUT /Users/{{ record.id }}
    required fields: id
    risk: medium: external Miro API mutation; approval required
  update_users_id_2:
    endpoint: PATCH /Users/{{ record.id }}
    required fields: id
    risk: medium: external Miro API mutation; approval required
  delete_users_id:
    endpoint: DELETE /Users/{{ record.id }}
    required fields: id
    risk: high: external Miro API mutation; approval required
  update_groups_id:
    endpoint: PATCH /Groups/{{ record.id }}
    required fields: id
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
    required fields: board_id
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
    required fields: board_id
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
    required fields: board_id
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
    required fields: board_id
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
    required fields: board_id
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
    required fields: board_id
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
    required fields: board_id
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
    required fields: board_id
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
    required fields: board_id
    risk: medium: external Miro API mutation; approval required
  update_boards_board_id_groups_group_id:
    endpoint: PUT /v2/boards/{{ record.board_id }}/groups/{{ record.group_id }}
    required fields: board_id, group_id
    risk: medium: external Miro API mutation; approval required
  delete_boards_board_id_groups_group_id:
    endpoint: DELETE /v2/boards/{{ record.board_id }}/groups/{{ record.group_id }}
    required fields: board_id, group_id
    risk: high: external Miro API mutation; approval required
  create_boards_board_id_tags:
    endpoint: POST /v2/boards/{{ record.board_id }}/tags
    required fields: board_id
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
    required fields: org_id, team_id
    risk: medium: external Miro API mutation; approval required
  update_orgs_org_id_teams_team_id_projects_project_id:
    endpoint: PATCH /v2/orgs/{{ record.org_id }}/teams/{{ record.team_id }}/projects/{{ record.project_id }}
    required fields: org_id, team_id, project_id
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
    required fields: org_id, team_id, project_id
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
    required fields: org_id
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
    required fields: org_id, team_id
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
    required fields: org_id
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
    required fields: org_id, group_id
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
    required fields: org_id, team_id
    risk: medium: external Miro API mutation; approval required
  delete_orgs_org_id_teams_team_id_groups_group_id:
    endpoint: DELETE /v2/orgs/{{ record.org_id }}/teams/{{ record.team_id }}/groups/{{ record.group_id }}
    required fields: org_id, team_id, group_id
    risk: high: external Miro API mutation; approval required
  create_orgs_org_id_boards_board_id_groups:
    endpoint: POST /v2/orgs/{{ record.org_id }}/boards/{{ record.board_id }}/groups
    required fields: org_id, board_id
    risk: medium: external Miro API mutation; approval required
  delete_orgs_org_id_boards_board_id_groups_group_id:
    endpoint: DELETE /v2/orgs/{{ record.org_id }}/boards/{{ record.board_id }}/groups/{{ record.group_id }}
    required fields: org_id, board_id, group_id
    risk: high: external Miro API mutation; approval required
  create_orgs_org_id_projects_project_id_groups:
    endpoint: POST /v2/orgs/{{ record.org_id }}/projects/{{ record.project_id }}/groups
    required fields: org_id, project_id
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
