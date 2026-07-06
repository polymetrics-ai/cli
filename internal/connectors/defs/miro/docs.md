# Overview

Reads and writes documented Miro Platform, Enterprise, SCIM, and experimental REST API resources
through the Miro Developer Platform API.

Readable streams: `boards`, `board_users`, `board_items`, `board_tags`, `board_connectors`,
`orgs_org_id_ai_interaction_logs`, `audit_logs`, `orgs_org_id_data_classification_settings`,
`orgs_org_id_teams_team_id_data_classification_settings`,
`orgs_org_id_teams_team_id_boards_board_id_data_classification`, `boards_board_id_docs_item_id`,
`orgs_org_id_cases`, `orgs_org_id_cases_case_id`, `orgs_org_id_cases_case_id_legal_holds`,
`orgs_org_id_cases_case_id_export_jobs`, `orgs_org_id_cases_case_id_legal_holds_legal_hold_id`,
`orgs_org_id_cases_case_id_legal_holds_legal_hold_id_content_items`,
`orgs_org_id_boards_export_jobs`, `orgs_org_id_boards_export_jobs_job_id`,
`orgs_org_id_boards_export_jobs_job_id_results`, `orgs_org_id_boards_export_jobs_job_id_tasks`,
`orgs_org_id_content_logs_items`, `users`, `users_id`, `groups`, `groups_id`,
`service_provider_config`, `resource_types`, `resource_types_resource`, `schemas`, `schemas_uri`,
`orgs_org_id`, `orgs_org_id_members`, `orgs_org_id_members_member_id`, `boards_board_id`,
`boards_board_id_app_cards_item_id`, `boards_board_id_cards_item_id`,
`boards_board_id_connectors_connector_id`, `boards_board_id_documents_item_id`,
`boards_board_id_embeds_item_id`, `boards_board_id_images_item_id`, `boards_board_id_items_item_id`,
`boards_board_id_members_board_member_id`, `boards_board_id_shapes_item_id`,
`boards_board_id_sticky_notes_item_id`, `boards_board_id_texts_item_id`,
`boards_board_id_frames_item_id`, `boards_board_id_platform_containers_items`,
`experimental_apps_app_id_metrics`, `experimental_apps_app_id_metrics_total`,
`experimental_boards_board_id_mindmap_nodes_item_id`, `experimental_boards_board_id_mindmap_nodes`,
`experimental_boards_board_id_items`, `experimental_boards_board_id_items_item_id`,
`experimental_boards_board_id_shapes_item_id`, `experimental_boards_board_id_code_widgets`,
`experimental_boards_board_id_code_widgets_item_id`, `boards_board_id_groups`,
`boards_board_id_groups_items`, `boards_board_id_groups_group_id`,
`boards_board_id_items_item_id_tags`, `boards_board_id_tags_tag_id`,
`boards_board_id_platform_tags_items`, `orgs_org_id_teams_team_id_projects`,
`orgs_org_id_teams_team_id_projects_project_id`,
`orgs_org_id_teams_team_id_projects_project_id_settings`,
`orgs_org_id_teams_team_id_projects_project_id_members`,
`orgs_org_id_teams_team_id_projects_project_id_members_member_id`, `orgs_org_id_teams`,
`orgs_org_id_teams_team_id`, `orgs_org_id_teams_team_id_members`,
`orgs_org_id_teams_team_id_members_member_id`, `orgs_org_id_default_teams_settings`,
`orgs_org_id_teams_team_id_settings`, `orgs_org_id_groups`, `orgs_org_id_groups_group_id`,
`orgs_org_id_groups_group_id_members`, `orgs_org_id_groups_group_id_members_member_id`,
`orgs_org_id_groups_group_id_teams`, `orgs_org_id_groups_group_id_teams_team_id`,
`orgs_org_id_teams_team_id_groups`, `orgs_org_id_teams_team_id_groups_group_id`,
`orgs_org_id_boards_board_id_groups`, `orgs_org_id_projects_project_id_groups`.

Write actions: `update_orgs_org_id_teams_team_id_data_classification`,
`update_orgs_org_id_teams_team_id_data_classification_settings`,
`create_orgs_org_id_teams_team_id_boards_board_id_data_classification`,
`create_boards_board_id_docs`, `delete_boards_board_id_docs_item_id`, `create_orgs_org_id_cases`,
`update_orgs_org_id_cases_case_id`, `delete_orgs_org_id_cases_case_id`,
`create_orgs_org_id_cases_case_id_legal_holds`,
`update_orgs_org_id_cases_case_id_legal_holds_legal_hold_id`,
`delete_orgs_org_id_cases_case_id_legal_holds_legal_hold_id`,
`update_orgs_org_id_boards_export_jobs_job_id_status`,
`create_orgs_org_id_boards_export_jobs_job_id_tasks_task_id_export_link`, `create_users`,
`update_users_id`, `update_users_id_2`, `delete_users_id`, `update_groups_id`, `create_boards`,
`update_boards_board_id`, `delete_boards_board_id`, `create_boards_board_id_app_cards`,
`update_boards_board_id_app_cards_item_id`, `delete_boards_board_id_app_cards_item_id`,
`create_boards_board_id_cards`, `update_boards_board_id_cards_item_id`,
`delete_boards_board_id_cards_item_id`, `create_boards_board_id_connectors`,
`update_boards_board_id_connectors_connector_id`, `delete_boards_board_id_connectors_connector_id`,
`create_boards_board_id_documents`, `update_boards_board_id_documents_item_id`,
`delete_boards_board_id_documents_item_id`, `create_boards_board_id_embeds`,
`update_boards_board_id_embeds_item_id`, `delete_boards_board_id_embeds_item_id`,
`create_boards_board_id_images`, `update_boards_board_id_images_item_id`,
`delete_boards_board_id_images_item_id`, `update_boards_board_id_items_item_id`,
`delete_boards_board_id_items_item_id`, `create_boards_board_id_members`,
`update_boards_board_id_members_board_member_id`, `delete_boards_board_id_members_board_member_id`,
`create_boards_board_id_shapes`, `update_boards_board_id_shapes_item_id`,
`delete_boards_board_id_shapes_item_id`, `create_boards_board_id_sticky_notes`,
`update_boards_board_id_sticky_notes_item_id`, `delete_boards_board_id_sticky_notes_item_id`,
`create_boards_board_id_texts`, `update_boards_board_id_texts_item_id`,
`delete_boards_board_id_texts_item_id`, `create_boards_board_id_frames`,
`update_boards_board_id_frames_item_id`, `delete_boards_board_id_frames_item_id`,
`delete_experimental_boards_board_id_mindmap_nodes_item_id`,
`create_experimental_boards_board_id_mindmap_nodes`,
`delete_experimental_boards_board_id_items_item_id`, `create_experimental_boards_board_id_shapes`,
`update_experimental_boards_board_id_shapes_item_id`,
`delete_experimental_boards_board_id_shapes_item_id`,
`create_experimental_boards_board_id_code_widgets`,
`update_experimental_boards_board_id_code_widgets_item_id`,
`delete_experimental_boards_board_id_code_widgets_item_id`,
`update_experimental_boards_board_id_code_widgets_item_id_position`,
`create_boards_board_id_groups`, `update_boards_board_id_groups_group_id`,
`delete_boards_board_id_groups_group_id`, `create_boards_board_id_tags`,
`update_boards_board_id_tags_tag_id`, `delete_boards_board_id_tags_tag_id`,
`create_orgs_org_id_teams_team_id_projects`, `update_orgs_org_id_teams_team_id_projects_project_id`,
`delete_orgs_org_id_teams_team_id_projects_project_id`,
`update_orgs_org_id_teams_team_id_projects_project_id_settings`,
`create_orgs_org_id_teams_team_id_projects_project_id_members`,
`update_orgs_org_id_teams_team_id_projects_project_id_members_member_id`,
`delete_orgs_org_id_teams_team_id_projects_project_id_members_member_id`,
`create_orgs_org_id_teams`, and 18 more.

Service API documentation: https://developers.miro.com/reference/api-reference.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Miro REST API access token. Sent as a Bearer token and never
  logged.
- `app_id` (optional, string); Path parameter app_id for the experimental_apps_app_id_metrics
  stream.
- `base_url` (optional, string); default `https://api.miro.com`; format `uri`; Miro API base URL
  override for tests or proxies.
- `board_id` (optional, string); Path parameter board_id for Miro board-scoped streams.
- `board_id_platform_containers` (optional, string); Path parameter board_id_PlatformContainers for
  the boards_board_id_platform_containers_items stream.
- `board_id_platform_tags` (optional, string); Path parameter board_id_PlatformTags for the
  boards_board_id_platform_tags_items stream.
- `board_member_id` (optional, string); Path parameter board_member_id for the
  boards_board_id_members_board_member_id stream.
- `case_id` (optional, string); Path parameter case_id for the orgs_org_id_cases_case_id stream.
- `connector_id` (optional, string); Path parameter connector_id for the
  boards_board_id_connectors_connector_id stream.
- `created_after` (optional, string); Required query parameter createdAfter for the audit_logs
  stream.
- `created_before` (optional, string); Required query parameter createdBefore for the audit_logs
  stream.
- `end_date` (optional, string); Required query parameter endDate for the
  experimental_apps_app_id_metrics stream.
- `from` (optional, string); Required query parameter from for the orgs_org_id_ai_interaction_logs
  stream.
- `group_id` (optional, string); Path parameter group_id for the boards_board_id_groups_group_id
  stream.
- `group_item_id` (optional, string); Required query parameter group_item_id for the
  boards_board_id_groups_items stream.
- `id` (optional, string); Path parameter id for the users_id stream.
- `item_id` (optional, string); Path parameter item_id for the boards_board_id_docs_item_id stream.
- `job_id` (optional, string); Path parameter job_id for the orgs_org_id_boards_export_jobs_job_id
  stream.
- `legal_hold_id` (optional, string); Path parameter legal_hold_id for the
  orgs_org_id_cases_case_id_legal_holds_legal_hold_id stream.
- `member_id` (optional, string); Path parameter member_id for the orgs_org_id_members_member_id
  stream.
- `org_id` (optional, string); Path parameter org_id for the orgs_org_id_ai_interaction_logs stream.
- `parent_item_id` (optional, string); Required query parameter parent_item_id for the
  boards_board_id_platform_containers_items stream.
- `project_id` (optional, string); Path parameter project_id for the
  orgs_org_id_teams_team_id_projects_project_id stream.
- `resource` (optional, string); Path parameter resource for the resource_types_resource stream.
- `start_date` (optional, string); Required query parameter startDate for the
  experimental_apps_app_id_metrics stream.
- `tag_id` (optional, string); Path parameter tag_id for the boards_board_id_tags_tag_id stream.
- `team_id` (optional, string); Path parameter team_id for the
  orgs_org_id_teams_team_id_data_classification_settings stream.
- `to` (optional, string); Required query parameter to for the orgs_org_id_ai_interaction_logs
  stream.
- `uri` (optional, string); Path parameter uri for the schemas_uri stream.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.miro.com`.

Authentication behavior:

- Bearer token authentication using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/v2/boards`.

## Streams notes

Default pagination: single request; no pagination.

Pagination by stream: cursor: `orgs_org_id_ai_interaction_logs`, `audit_logs`,
`orgs_org_id_boards_export_jobs`, `orgs_org_id_boards_export_jobs_job_id_tasks`,
`orgs_org_id_content_logs_items`, `orgs_org_id_members`,
`boards_board_id_platform_containers_items`, `experimental_boards_board_id_mindmap_nodes`,
`experimental_boards_board_id_items`, `experimental_boards_board_id_code_widgets`,
`boards_board_id_groups`, `boards_board_id_groups_items`, `orgs_org_id_teams_team_id_projects`,
`orgs_org_id_teams_team_id_projects_project_id_members`, `orgs_org_id_teams`,
`orgs_org_id_teams_team_id_members`; none: `orgs_org_id_data_classification_settings`,
`orgs_org_id_teams_team_id_data_classification_settings`,
`orgs_org_id_teams_team_id_boards_board_id_data_classification`, `boards_board_id_docs_item_id`,
`orgs_org_id_cases`, `orgs_org_id_cases_case_id`, `orgs_org_id_cases_case_id_legal_holds`,
`orgs_org_id_cases_case_id_export_jobs`, `orgs_org_id_cases_case_id_legal_holds_legal_hold_id`,
`orgs_org_id_cases_case_id_legal_holds_legal_hold_id_content_items`,
`orgs_org_id_boards_export_jobs_job_id`, `orgs_org_id_boards_export_jobs_job_id_results`, `users`,
`users_id`, `groups`, `groups_id`, `service_provider_config`, `resource_types`,
`resource_types_resource`, `schemas`, `schemas_uri`, `orgs_org_id`, `orgs_org_id_members_member_id`,
`boards_board_id`, `boards_board_id_app_cards_item_id`, `boards_board_id_cards_item_id`,
`boards_board_id_connectors_connector_id`, `boards_board_id_documents_item_id`,
`boards_board_id_embeds_item_id`, `boards_board_id_images_item_id`, `boards_board_id_items_item_id`,
`boards_board_id_members_board_member_id`, `boards_board_id_shapes_item_id`,
`boards_board_id_sticky_notes_item_id`, `boards_board_id_texts_item_id`,
`boards_board_id_frames_item_id`, `experimental_apps_app_id_metrics`,
`experimental_apps_app_id_metrics_total`, `experimental_boards_board_id_mindmap_nodes_item_id`,
`experimental_boards_board_id_items_item_id`, `experimental_boards_board_id_shapes_item_id`,
`experimental_boards_board_id_code_widgets_item_id`, `boards_board_id_groups_group_id`,
`boards_board_id_items_item_id_tags`, `boards_board_id_tags_tag_id`,
`orgs_org_id_teams_team_id_projects_project_id`,
`orgs_org_id_teams_team_id_projects_project_id_settings`,
`orgs_org_id_teams_team_id_projects_project_id_members_member_id`, `orgs_org_id_teams_team_id`,
`orgs_org_id_teams_team_id_members_member_id`, `orgs_org_id_default_teams_settings`,
`orgs_org_id_teams_team_id_settings`, `orgs_org_id_groups`, `orgs_org_id_groups_group_id`,
`orgs_org_id_groups_group_id_members`, `orgs_org_id_groups_group_id_members_member_id`,
`orgs_org_id_groups_group_id_teams`, `orgs_org_id_groups_group_id_teams_team_id`,
`orgs_org_id_teams_team_id_groups`, `orgs_org_id_teams_team_id_groups_group_id`, and 2 more;
offset_limit: `boards`, `board_users`, `board_items`, `board_tags`, `board_connectors`,
`boards_board_id_platform_tags_items`.

- `boards`: GET `/v2/boards` - records path `data`; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 50; computed output fields `created_at`,
  `modified_at`, `owner_id`, `team_id`, `view_link`.
- `board_users`: GET `/v2/boards/{{ config.board_id }}/members` - records path `data`; offset/limit
  pagination; offset parameter `offset`; limit parameter `limit`; page size 50; computed output
  fields `board_id`.
- `board_items`: GET `/v2/boards/{{ config.board_id }}/items` - records path `data`; offset/limit
  pagination; offset parameter `offset`; limit parameter `limit`; page size 50; computed output
  fields `board_id`, `created_at`, `modified_at`.
- `board_tags`: GET `/v2/boards/{{ config.board_id }}/tags` - records path `data`; offset/limit
  pagination; offset parameter `offset`; limit parameter `limit`; page size 50; computed output
  fields `board_id`, `fill_color`.
- `board_connectors`: GET `/v2/boards/{{ config.board_id }}/connectors` - records path `data`;
  offset/limit pagination; offset parameter `offset`; limit parameter `limit`; page size 50;
  computed output fields `board_id`.
- `orgs_org_id_ai_interaction_logs`: GET `/v2/orgs/{{ config.org_id }}/ai-interaction-logs` -
  records path `data`; query `from`=`{{ config.from }}`; `limit`=`50`; `to`=`{{ config.to }}`;
  cursor pagination; cursor parameter `cursor`; next token from `cursor`; emits passthrough records.
- `audit_logs`: GET `/v2/audit/logs` - records path `data`; query `createdAfter`=`{{
  config.created_after }}`; `createdBefore`=`{{ config.created_before }}`; `limit`=`50`; cursor
  pagination; cursor parameter `cursor`; next token from `cursor`; emits passthrough records.
- `orgs_org_id_data_classification_settings`: GET `/v2/orgs/{{ config.org_id
  }}/data-classification-settings` - records at response root; emits passthrough records.
- `orgs_org_id_teams_team_id_data_classification_settings`: GET `/v2/orgs/{{ config.org_id
  }}/teams/{{ config.team_id }}/data-classification-settings` - records at response root; emits
  passthrough records.
- `orgs_org_id_teams_team_id_boards_board_id_data_classification`: GET `/v2/orgs/{{ config.org_id
  }}/teams/{{ config.team_id }}/boards/{{ config.board_id }}/data-classification` - records at
  response root; emits passthrough records.
- `boards_board_id_docs_item_id`: GET `/v2/boards/{{ config.board_id }}/docs/{{ config.item_id }}` -
  records at response root; emits passthrough records.
- `orgs_org_id_cases`: GET `/v2/orgs/{{ config.org_id }}/cases` - records path `data`; emits
  passthrough records.
- `orgs_org_id_cases_case_id`: GET `/v2/orgs/{{ config.org_id }}/cases/{{ config.case_id }}` -
  records at response root; emits passthrough records.
- `orgs_org_id_cases_case_id_legal_holds`: GET `/v2/orgs/{{ config.org_id }}/cases/{{ config.case_id
  }}/legal-holds` - records path `data`; emits passthrough records.
- `orgs_org_id_cases_case_id_export_jobs`: GET `/v2/orgs/{{ config.org_id }}/cases/{{ config.case_id
  }}/export-jobs` - records path `data`; emits passthrough records.
- `orgs_org_id_cases_case_id_legal_holds_legal_hold_id`: GET `/v2/orgs/{{ config.org_id }}/cases/{{
  config.case_id }}/legal-holds/{{ config.legal_hold_id }}` - records at response root; emits
  passthrough records.
- `orgs_org_id_cases_case_id_legal_holds_legal_hold_id_content_items`: GET `/v2/orgs/{{
  config.org_id }}/cases/{{ config.case_id }}/legal-holds/{{ config.legal_hold_id }}/content-items`
  - records path `data`; emits passthrough records.
- `orgs_org_id_boards_export_jobs`: GET `/v2/orgs/{{ config.org_id }}/boards/export/jobs` - records
  path `data`; query `limit`=`50`; cursor pagination; cursor parameter `cursor`; next token from
  `cursor`; emits passthrough records.
- `orgs_org_id_boards_export_jobs_job_id`: GET `/v2/orgs/{{ config.org_id }}/boards/export/jobs/{{
  config.job_id }}` - records at response root; emits passthrough records.
- `orgs_org_id_boards_export_jobs_job_id_results`: GET `/v2/orgs/{{ config.org_id
  }}/boards/export/jobs/{{ config.job_id }}/results` - records path `results`; emits passthrough
  records.
- `orgs_org_id_boards_export_jobs_job_id_tasks`: GET `/v2/orgs/{{ config.org_id
  }}/boards/export/jobs/{{ config.job_id }}/tasks` - records path `data`; query `limit`=`50`; cursor
  pagination; cursor parameter `cursor`; next token from `cursor`; emits passthrough records.
- `orgs_org_id_content_logs_items`: GET `/v2/orgs/{{ config.org_id }}/content-logs/items` - records
  path `data`; query `from`=`{{ config.from }}`; `limit`=`50`; `to`=`{{ config.to }}`; cursor
  pagination; cursor parameter `cursor`; next token from `cursor`; emits passthrough records.
- `users`: GET `/Users` - records path `Resources`; emits passthrough records.
- `users_id`: GET `/Users/{{ config.id }}` - records at response root; emits passthrough records.
- `groups`: GET `/Groups` - records path `Resources`; emits passthrough records.
- `groups_id`: GET `/Groups/{{ config.id }}` - records at response root; emits passthrough records.
- `service_provider_config`: GET `/ServiceProviderConfig` - records at response root; emits
  passthrough records.
- `resource_types`: GET `/ResourceTypes` - records path `Resources`; emits passthrough records.
- `resource_types_resource`: GET `/ResourceTypes/{{ config.resource }}` - records at response root;
  emits passthrough records.
- `schemas`: GET `/Schemas` - records path `Resources`; emits passthrough records.
- `schemas_uri`: GET `/Schemas/{{ config.uri }}` - records at response root; emits passthrough
  records.
- `orgs_org_id`: GET `/v2/orgs/{{ config.org_id }}` - records at response root; emits passthrough
  records.
- `orgs_org_id_members`: GET `/v2/orgs/{{ config.org_id }}/members` - records path `data`; query
  `limit`=`50`; cursor pagination; cursor parameter `cursor`; next token from `cursor`; emits
  passthrough records.
- `orgs_org_id_members_member_id`: GET `/v2/orgs/{{ config.org_id }}/members/{{ config.member_id }}`
  - records at response root; emits passthrough records.
- `boards_board_id`: GET `/v2/boards/{{ config.board_id }}` - records at response root; emits
  passthrough records.
- `boards_board_id_app_cards_item_id`: GET `/v2/boards/{{ config.board_id }}/app_cards/{{
  config.item_id }}` - records at response root; emits passthrough records.
- `boards_board_id_cards_item_id`: GET `/v2/boards/{{ config.board_id }}/cards/{{ config.item_id }}`
  - records at response root; emits passthrough records.
- `boards_board_id_connectors_connector_id`: GET `/v2/boards/{{ config.board_id }}/connectors/{{
  config.connector_id }}` - records at response root; emits passthrough records.
- `boards_board_id_documents_item_id`: GET `/v2/boards/{{ config.board_id }}/documents/{{
  config.item_id }}` - records at response root; emits passthrough records.
- `boards_board_id_embeds_item_id`: GET `/v2/boards/{{ config.board_id }}/embeds/{{ config.item_id
  }}` - records at response root; emits passthrough records.
- `boards_board_id_images_item_id`: GET `/v2/boards/{{ config.board_id }}/images/{{ config.item_id
  }}` - records at response root; emits passthrough records.
- `boards_board_id_items_item_id`: GET `/v2/boards/{{ config.board_id }}/items/{{ config.item_id }}`
  - records at response root; emits passthrough records.
- `boards_board_id_members_board_member_id`: GET `/v2/boards/{{ config.board_id }}/members/{{
  config.board_member_id }}` - records at response root; emits passthrough records.
- `boards_board_id_shapes_item_id`: GET `/v2/boards/{{ config.board_id }}/shapes/{{ config.item_id
  }}` - records at response root; emits passthrough records.
- `boards_board_id_sticky_notes_item_id`: GET `/v2/boards/{{ config.board_id }}/sticky_notes/{{
  config.item_id }}` - records at response root; emits passthrough records.
- `boards_board_id_texts_item_id`: GET `/v2/boards/{{ config.board_id }}/texts/{{ config.item_id }}`
  - records at response root; emits passthrough records.
- `boards_board_id_frames_item_id`: GET `/v2/boards/{{ config.board_id }}/frames/{{ config.item_id
  }}` - records at response root; emits passthrough records.
- `boards_board_id_platform_containers_items`: GET `/v2/boards/{{
  config.board_id_platform_containers }}/items` - records path `data`; query `limit`=`50`;
  `parent_item_id`=`{{ config.parent_item_id }}`; cursor pagination; cursor parameter `cursor`; next
  token from `cursor`; emits passthrough records.
- `experimental_apps_app_id_metrics`: GET `/v2-experimental/apps/{{ config.app_id }}/metrics` -
  records at response root; query `endDate`=`{{ config.end_date }}`; `startDate`=`{{
  config.start_date }}`; emits passthrough records.
- `experimental_apps_app_id_metrics_total`: GET `/v2-experimental/apps/{{ config.app_id
  }}/metrics-total` - records at response root; emits passthrough records.
- `experimental_boards_board_id_mindmap_nodes_item_id`: GET `/v2-experimental/boards/{{
  config.board_id }}/mindmap_nodes/{{ config.item_id }}` - records at response root; emits
  passthrough records.
- `experimental_boards_board_id_mindmap_nodes`: GET `/v2-experimental/boards/{{ config.board_id
  }}/mindmap_nodes` - records path `data`; query `limit`=`50`; cursor pagination; cursor parameter
  `cursor`; next token from `cursor`; emits passthrough records.
- `experimental_boards_board_id_items`: GET `/v2-experimental/boards/{{ config.board_id }}/items` -
  records path `data`; query `limit`=`50`; cursor pagination; cursor parameter `cursor`; next token
  from `cursor`; emits passthrough records.
- `experimental_boards_board_id_items_item_id`: GET `/v2-experimental/boards/{{ config.board_id
  }}/items/{{ config.item_id }}` - records at response root; emits passthrough records.
- `experimental_boards_board_id_shapes_item_id`: GET `/v2-experimental/boards/{{ config.board_id
  }}/shapes/{{ config.item_id }}` - records at response root; emits passthrough records.
- `experimental_boards_board_id_code_widgets`: GET `/v2-experimental/boards/{{ config.board_id
  }}/code_widgets` - records path `data`; query `limit`=`50`; cursor pagination; cursor parameter
  `cursor`; next token from `cursor`; emits passthrough records.
- `experimental_boards_board_id_code_widgets_item_id`: GET `/v2-experimental/boards/{{
  config.board_id }}/code_widgets/{{ config.item_id }}` - records at response root; emits
  passthrough records.
- `boards_board_id_groups`: GET `/v2/boards/{{ config.board_id }}/groups` - records path `data`;
  query `limit`=`50`; cursor pagination; cursor parameter `cursor`; next token from `cursor`; emits
  passthrough records.
- `boards_board_id_groups_items`: GET `/v2/boards/{{ config.board_id }}/groups/items` - records path
  `data.data`; query `group_item_id`=`{{ config.group_item_id }}`; `limit`=`50`; cursor pagination;
  cursor parameter `cursor`; next token from `cursor`; emits passthrough records.
- `boards_board_id_groups_group_id`: GET `/v2/boards/{{ config.board_id }}/groups/{{ config.group_id
  }}` - records at response root; emits passthrough records.
- `boards_board_id_items_item_id_tags`: GET `/v2/boards/{{ config.board_id }}/items/{{
  config.item_id }}/tags` - records path `tags`; emits passthrough records.
- `boards_board_id_tags_tag_id`: GET `/v2/boards/{{ config.board_id }}/tags/{{ config.tag_id }}` -
  records at response root; emits passthrough records.
- `boards_board_id_platform_tags_items`: GET `/v2/boards/{{ config.board_id_platform_tags }}/items`
  - records path `data`; query `tag_id`=`{{ config.tag_id }}`; offset/limit pagination; offset
  parameter `offset`; limit parameter `limit`; page size 50; emits passthrough records.
- `orgs_org_id_teams_team_id_projects`: GET `/v2/orgs/{{ config.org_id }}/teams/{{ config.team_id
  }}/projects` - records path `data`; query `limit`=`50`; cursor pagination; cursor parameter
  `cursor`; next token from `cursor`; emits passthrough records.
- `orgs_org_id_teams_team_id_projects_project_id`: GET `/v2/orgs/{{ config.org_id }}/teams/{{
  config.team_id }}/projects/{{ config.project_id }}` - records at response root; emits passthrough
  records.
- `orgs_org_id_teams_team_id_projects_project_id_settings`: GET `/v2/orgs/{{ config.org_id
  }}/teams/{{ config.team_id }}/projects/{{ config.project_id }}/settings` - records at response
  root; emits passthrough records.
- `orgs_org_id_teams_team_id_projects_project_id_members`: GET `/v2/orgs/{{ config.org_id
  }}/teams/{{ config.team_id }}/projects/{{ config.project_id }}/members` - records path `data`;
  query `limit`=`50`; cursor pagination; cursor parameter `cursor`; next token from `cursor`; emits
  passthrough records.
- `orgs_org_id_teams_team_id_projects_project_id_members_member_id`: GET `/v2/orgs/{{ config.org_id
  }}/teams/{{ config.team_id }}/projects/{{ config.project_id }}/members/{{ config.member_id }}` -
  records at response root; emits passthrough records.
- `orgs_org_id_teams`: GET `/v2/orgs/{{ config.org_id }}/teams` - records path `data`; query
  `limit`=`50`; cursor pagination; cursor parameter `cursor`; next token from `cursor`; emits
  passthrough records.
- `orgs_org_id_teams_team_id`: GET `/v2/orgs/{{ config.org_id }}/teams/{{ config.team_id }}` -
  records at response root; emits passthrough records.
- `orgs_org_id_teams_team_id_members`: GET `/v2/orgs/{{ config.org_id }}/teams/{{ config.team_id
  }}/members` - records path `data`; query `limit`=`50`; cursor pagination; cursor parameter
  `cursor`; next token from `cursor`; emits passthrough records.
- `orgs_org_id_teams_team_id_members_member_id`: GET `/v2/orgs/{{ config.org_id }}/teams/{{
  config.team_id }}/members/{{ config.member_id }}` - records at response root; emits passthrough
  records.
- `orgs_org_id_default_teams_settings`: GET `/v2/orgs/{{ config.org_id }}/default_teams_settings` -
  records at response root; emits passthrough records.
- `orgs_org_id_teams_team_id_settings`: GET `/v2/orgs/{{ config.org_id }}/teams/{{ config.team_id
  }}/settings` - records at response root; emits passthrough records.
- `orgs_org_id_groups`: GET `/v2/orgs/{{ config.org_id }}/groups` - records path `data`; emits
  passthrough records.
- `orgs_org_id_groups_group_id`: GET `/v2/orgs/{{ config.org_id }}/groups/{{ config.group_id }}` -
  records at response root; emits passthrough records.
- `orgs_org_id_groups_group_id_members`: GET `/v2/orgs/{{ config.org_id }}/groups/{{ config.group_id
  }}/members` - records path `data`; emits passthrough records.
- `orgs_org_id_groups_group_id_members_member_id`: GET `/v2/orgs/{{ config.org_id }}/groups/{{
  config.group_id }}/members/{{ config.member_id }}` - records at response root; emits passthrough
  records.
- `orgs_org_id_groups_group_id_teams`: GET `/v2/orgs/{{ config.org_id }}/groups/{{ config.group_id
  }}/teams` - records path `data`; emits passthrough records.
- `orgs_org_id_groups_group_id_teams_team_id`: GET `/v2/orgs/{{ config.org_id }}/groups/{{
  config.group_id }}/teams/{{ config.team_id }}` - records at response root; emits passthrough
  records.
- `orgs_org_id_teams_team_id_groups`: GET `/v2/orgs/{{ config.org_id }}/teams/{{ config.team_id
  }}/groups` - records path `data`; emits passthrough records.
- `orgs_org_id_teams_team_id_groups_group_id`: GET `/v2/orgs/{{ config.org_id }}/teams/{{
  config.team_id }}/groups/{{ config.group_id }}` - records at response root; emits passthrough
  records.
- `orgs_org_id_boards_board_id_groups`: GET `/v2/orgs/{{ config.org_id }}/boards/{{ config.board_id
  }}/groups` - records path `data`; emits passthrough records.
- `orgs_org_id_projects_project_id_groups`: GET `/v2/orgs/{{ config.org_id }}/projects/{{
  config.project_id }}/groups` - records path `data`; emits passthrough records.

## Write actions & risks

Overall write risk: external Miro API mutations including board sharing, item changes, enterprise
administration, SCIM provisioning, groups, projects, and deletes.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `update_orgs_org_id_teams_team_id_data_classification`: PATCH `/v2/orgs/{{ record.org_id
  }}/teams/{{ record.team_id }}/data-classification` - kind `update`; body type `json`; path fields
  `org_id`, `team_id`; required record fields `org_id`, `team_id`; accepted fields `labelId`,
  `notClassifiedOnly`, `org_id`, `team_id`; risk: medium: external Miro API mutation; approval
  required.
- `update_orgs_org_id_teams_team_id_data_classification_settings`: PATCH `/v2/orgs/{{ record.org_id
  }}/teams/{{ record.team_id }}/data-classification-settings` - kind `update`; body type `json`;
  path fields `org_id`, `team_id`; required record fields `org_id`, `team_id`; accepted fields
  `defaultLabelId`, `enabled`, `org_id`, `team_id`; risk: medium: external Miro API mutation;
  approval required.
- `create_orgs_org_id_teams_team_id_boards_board_id_data_classification`: POST `/v2/orgs/{{
  record.org_id }}/teams/{{ record.team_id }}/boards/{{ record.board_id }}/data-classification` -
  kind `create`; body type `json`; path fields `org_id`, `team_id`, `board_id`; required record
  fields `org_id`, `team_id`, `board_id`; accepted fields `board_id`, `labelId`, `org_id`,
  `team_id`; risk: medium: external Miro API mutation; approval required.
- `create_boards_board_id_docs`: POST `/v2/boards/{{ record.board_id }}/docs` - kind `create`; body
  type `json`; path fields `board_id`; required record fields `board_id`, `data`; accepted fields
  `board_id`, `data`, `parent`, `position`; risk: medium: external Miro API mutation; approval
  required.
- `delete_boards_board_id_docs_item_id`: DELETE `/v2/boards/{{ record.board_id }}/docs/{{
  record.item_id }}` - kind `delete`; body type `none`; path fields `board_id`, `item_id`; required
  record fields `board_id`, `item_id`; accepted fields `board_id`, `item_id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: high: external Miro API
  mutation; approval required.
- `create_orgs_org_id_cases`: POST `/v2/orgs/{{ record.org_id }}/cases` - kind `create`; body type
  `json`; path fields `org_id`; required record fields `org_id`, `name`; accepted fields
  `description`, `name`, `org_id`; risk: medium: external Miro API mutation; approval required.
- `update_orgs_org_id_cases_case_id`: PUT `/v2/orgs/{{ record.org_id }}/cases/{{ record.case_id }}`
  - kind `update`; body type `json`; path fields `org_id`, `case_id`; required record fields
  `org_id`, `case_id`, `name`; accepted fields `case_id`, `description`, `name`, `org_id`; risk:
  medium: external Miro API mutation; approval required.
- `delete_orgs_org_id_cases_case_id`: DELETE `/v2/orgs/{{ record.org_id }}/cases/{{ record.case_id
  }}` - kind `delete`; body type `none`; path fields `org_id`, `case_id`; required record fields
  `org_id`, `case_id`; accepted fields `case_id`, `org_id`; missing records treated as success for
  status `404`; confirmation `destructive`; risk: high: external Miro API mutation; approval
  required.
- `create_orgs_org_id_cases_case_id_legal_holds`: POST `/v2/orgs/{{ record.org_id }}/cases/{{
  record.case_id }}/legal-holds` - kind `create`; body type `json`; path fields `org_id`, `case_id`;
  required record fields `org_id`, `case_id`, `name`, `scope`; accepted fields `case_id`,
  `description`, `name`, `org_id`, `scope`; risk: medium: external Miro API mutation; approval
  required.
- `update_orgs_org_id_cases_case_id_legal_holds_legal_hold_id`: PUT `/v2/orgs/{{ record.org_id
  }}/cases/{{ record.case_id }}/legal-holds/{{ record.legal_hold_id }}` - kind `update`; body type
  `json`; path fields `org_id`, `case_id`, `legal_hold_id`; required record fields `org_id`,
  `case_id`, `legal_hold_id`, `name`, `scope`; accepted fields `case_id`, `description`,
  `legal_hold_id`, `name`, `org_id`, `scope`; risk: medium: external Miro API mutation; approval
  required.
- `delete_orgs_org_id_cases_case_id_legal_holds_legal_hold_id`: DELETE `/v2/orgs/{{ record.org_id
  }}/cases/{{ record.case_id }}/legal-holds/{{ record.legal_hold_id }}` - kind `delete`; body type
  `none`; path fields `org_id`, `case_id`, `legal_hold_id`; required record fields `org_id`,
  `case_id`, `legal_hold_id`; accepted fields `case_id`, `legal_hold_id`, `org_id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: high: external Miro API
  mutation; approval required.
- `update_orgs_org_id_boards_export_jobs_job_id_status`: PUT `/v2/orgs/{{ record.org_id
  }}/boards/export/jobs/{{ record.job_id }}/status` - kind `update`; body type `json`; path fields
  `org_id`, `job_id`; required record fields `org_id`, `job_id`, `status`; accepted fields `job_id`,
  `org_id`, `status`; risk: medium: external Miro API mutation; approval required.
- `create_orgs_org_id_boards_export_jobs_job_id_tasks_task_id_export_link`: POST `/v2/orgs/{{
  record.org_id }}/boards/export/jobs/{{ record.job_id }}/tasks/{{ record.task_id }}/export-link` -
  kind `create`; body type `none`; path fields `org_id`, `job_id`, `task_id`; required record fields
  `org_id`, `job_id`, `task_id`; accepted fields `job_id`, `org_id`, `task_id`; risk: medium:
  external Miro API mutation; approval required.
- `create_users`: POST `/Users` - kind `create`; body type `json`; required record fields
  `userName`; accepted fields `active`, `displayName`, `name`, `photos`, `preferredLanguage`,
  `roles`, `schemas`, `urn:ietf:params:scim:schemas:extension:enterprise:2.0:User`, `userName`,
  `userType`; risk: medium: external Miro API mutation; approval required.
- `update_users_id`: PUT `/Users/{{ record.id }}` - kind `update`; body type `json`; path fields
  `id`; required record fields `id`; accepted fields `active`, `displayName`, `emails`, `groups`,
  `id`, `meta`, `name`, `photos`, `preferredLanguage`, `roles`, `schemas`,
  `urn:ietf:params:scim:schemas:extension:enterprise:2.0:User`, `userName`, `userType`; risk:
  medium: external Miro API mutation; approval required.
- `update_users_id_2`: PATCH `/Users/{{ record.id }}` - kind `update`; body type `json`; path fields
  `id`; required record fields `id`, `schemas`, `Operations`; accepted fields `Operations`, `id`,
  `schemas`; risk: medium: external Miro API mutation; approval required.
- `delete_users_id`: DELETE `/Users/{{ record.id }}` - kind `delete`; body type `none`; path fields
  `id`; required record fields `id`; accepted fields `id`; missing records treated as success for
  status `404`; confirmation `destructive`; risk: high: external Miro API mutation; approval
  required.
- `update_groups_id`: PATCH `/Groups/{{ record.id }}` - kind `update`; body type `json`; path fields
  `id`; required record fields `id`, `schemas`, `Operations`; accepted fields `Operations`, `id`,
  `schemas`; risk: medium: external Miro API mutation; approval required.
- `create_boards`: POST `/v2/boards` - kind `create`; body type `json`; accepted fields
  `description`, `name`, `policy`, `projectId`, `teamId`; risk: medium: external Miro API mutation;
  approval required.
- `update_boards_board_id`: PATCH `/v2/boards/{{ record.board_id }}` - kind `update`; body type
  `json`; path fields `board_id`; required record fields `board_id`; accepted fields `board_id`,
  `description`, `name`, `policy`, `projectId`, `teamId`; risk: medium: external Miro API mutation;
  approval required.
- `delete_boards_board_id`: DELETE `/v2/boards/{{ record.board_id }}` - kind `delete`; body type
  `none`; path fields `board_id`; required record fields `board_id`; accepted fields `board_id`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: high:
  external Miro API mutation; approval required.
- `create_boards_board_id_app_cards`: POST `/v2/boards/{{ record.board_id }}/app_cards` - kind
  `create`; body type `json`; path fields `board_id`; required record fields `board_id`; accepted
  fields `board_id`, `data`, `geometry`, `parent`, `position`, `style`; risk: medium: external Miro
  API mutation; approval required.
- `update_boards_board_id_app_cards_item_id`: PATCH `/v2/boards/{{ record.board_id }}/app_cards/{{
  record.item_id }}` - kind `update`; body type `json`; path fields `board_id`, `item_id`; required
  record fields `board_id`, `item_id`; accepted fields `board_id`, `data`, `geometry`, `item_id`,
  `parent`, `position`, `style`; risk: medium: external Miro API mutation; approval required.
- `delete_boards_board_id_app_cards_item_id`: DELETE `/v2/boards/{{ record.board_id }}/app_cards/{{
  record.item_id }}` - kind `delete`; body type `none`; path fields `board_id`, `item_id`; required
  record fields `board_id`, `item_id`; accepted fields `board_id`, `item_id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: high: external Miro API
  mutation; approval required.
- `create_boards_board_id_cards`: POST `/v2/boards/{{ record.board_id }}/cards` - kind `create`;
  body type `json`; path fields `board_id`; required record fields `board_id`; accepted fields
  `board_id`, `data`, `geometry`, `parent`, `position`, `style`; risk: medium: external Miro API
  mutation; approval required.
- `update_boards_board_id_cards_item_id`: PATCH `/v2/boards/{{ record.board_id }}/cards/{{
  record.item_id }}` - kind `update`; body type `json`; path fields `board_id`, `item_id`; required
  record fields `board_id`, `item_id`; accepted fields `board_id`, `data`, `geometry`, `item_id`,
  `parent`, `position`, `style`; risk: medium: external Miro API mutation; approval required.
- `delete_boards_board_id_cards_item_id`: DELETE `/v2/boards/{{ record.board_id }}/cards/{{
  record.item_id }}` - kind `delete`; body type `none`; path fields `board_id`, `item_id`; required
  record fields `board_id`, `item_id`; accepted fields `board_id`, `item_id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: high: external Miro API
  mutation; approval required.
- `create_boards_board_id_connectors`: POST `/v2/boards/{{ record.board_id }}/connectors` - kind
  `create`; body type `json`; path fields `board_id`; required record fields `board_id`, `endItem`,
  `startItem`; accepted fields `board_id`, `captions`, `endItem`, `shape`, `startItem`, `style`;
  risk: medium: external Miro API mutation; approval required.
- `update_boards_board_id_connectors_connector_id`: PATCH `/v2/boards/{{ record.board_id
  }}/connectors/{{ record.connector_id }}` - kind `update`; body type `json`; path fields
  `board_id`, `connector_id`; required record fields `board_id`, `connector_id`; accepted fields
  `board_id`, `captions`, `connector_id`, `endItem`, `shape`, `startItem`, `style`; risk: medium:
  external Miro API mutation; approval required.
- `delete_boards_board_id_connectors_connector_id`: DELETE `/v2/boards/{{ record.board_id
  }}/connectors/{{ record.connector_id }}` - kind `delete`; body type `none`; path fields
  `board_id`, `connector_id`; required record fields `board_id`, `connector_id`; accepted fields
  `board_id`, `connector_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: high: external Miro API mutation; approval required.
- `create_boards_board_id_documents`: POST `/v2/boards/{{ record.board_id }}/documents` - kind
  `create`; body type `json`; path fields `board_id`; required record fields `board_id`, `data`;
  accepted fields `board_id`, `data`, `parent`, `position`; risk: medium: external Miro API
  mutation; approval required.
- `update_boards_board_id_documents_item_id`: PATCH `/v2/boards/{{ record.board_id }}/documents/{{
  record.item_id }}` - kind `update`; body type `json`; path fields `board_id`, `item_id`; required
  record fields `board_id`, `item_id`; accepted fields `board_id`, `data`, `geometry`, `item_id`,
  `parent`, `position`; risk: medium: external Miro API mutation; approval required.
- `delete_boards_board_id_documents_item_id`: DELETE `/v2/boards/{{ record.board_id }}/documents/{{
  record.item_id }}` - kind `delete`; body type `none`; path fields `board_id`, `item_id`; required
  record fields `board_id`, `item_id`; accepted fields `board_id`, `item_id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: high: external Miro API
  mutation; approval required.
- `create_boards_board_id_embeds`: POST `/v2/boards/{{ record.board_id }}/embeds` - kind `create`;
  body type `json`; path fields `board_id`; required record fields `board_id`, `data`; accepted
  fields `board_id`, `data`, `geometry`, `parent`, `position`; risk: medium: external Miro API
  mutation; approval required.
- `update_boards_board_id_embeds_item_id`: PATCH `/v2/boards/{{ record.board_id }}/embeds/{{
  record.item_id }}` - kind `update`; body type `json`; path fields `board_id`, `item_id`; required
  record fields `board_id`, `item_id`; accepted fields `board_id`, `data`, `geometry`, `item_id`,
  `parent`, `position`; risk: medium: external Miro API mutation; approval required.
- `delete_boards_board_id_embeds_item_id`: DELETE `/v2/boards/{{ record.board_id }}/embeds/{{
  record.item_id }}` - kind `delete`; body type `none`; path fields `board_id`, `item_id`; required
  record fields `board_id`, `item_id`; accepted fields `board_id`, `item_id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: high: external Miro API
  mutation; approval required.
- `create_boards_board_id_images`: POST `/v2/boards/{{ record.board_id }}/images` - kind `create`;
  body type `json`; path fields `board_id`; required record fields `board_id`, `data`; accepted
  fields `board_id`, `data`, `geometry`, `parent`, `position`; risk: medium: external Miro API
  mutation; approval required.
- `update_boards_board_id_images_item_id`: PATCH `/v2/boards/{{ record.board_id }}/images/{{
  record.item_id }}` - kind `update`; body type `json`; path fields `board_id`, `item_id`; required
  record fields `board_id`, `item_id`; accepted fields `board_id`, `data`, `geometry`, `item_id`,
  `parent`, `position`; risk: medium: external Miro API mutation; approval required.
- `delete_boards_board_id_images_item_id`: DELETE `/v2/boards/{{ record.board_id }}/images/{{
  record.item_id }}` - kind `delete`; body type `none`; path fields `board_id`, `item_id`; required
  record fields `board_id`, `item_id`; accepted fields `board_id`, `item_id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: high: external Miro API
  mutation; approval required.
- `update_boards_board_id_items_item_id`: PATCH `/v2/boards/{{ record.board_id }}/items/{{
  record.item_id }}` - kind `update`; body type `json`; path fields `board_id`, `item_id`; required
  record fields `board_id`, `item_id`; accepted fields `board_id`, `item_id`, `parent`, `position`;
  risk: medium: external Miro API mutation; approval required.
- `delete_boards_board_id_items_item_id`: DELETE `/v2/boards/{{ record.board_id }}/items/{{
  record.item_id }}` - kind `delete`; body type `none`; path fields `board_id`, `item_id`; required
  record fields `board_id`, `item_id`; accepted fields `board_id`, `item_id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: high: external Miro API
  mutation; approval required.
- `create_boards_board_id_members`: POST `/v2/boards/{{ record.board_id }}/members` - kind `create`;
  body type `json`; path fields `board_id`; required record fields `board_id`, `emails`; accepted
  fields `board_id`, `emails`, `message`, `role`; risk: medium: external Miro API mutation; approval
  required.
- `update_boards_board_id_members_board_member_id`: PATCH `/v2/boards/{{ record.board_id
  }}/members/{{ record.board_member_id }}` - kind `update`; body type `json`; path fields
  `board_id`, `board_member_id`; required record fields `board_id`, `board_member_id`; accepted
  fields `board_id`, `board_member_id`, `role`; risk: medium: external Miro API mutation; approval
  required.
- `delete_boards_board_id_members_board_member_id`: DELETE `/v2/boards/{{ record.board_id
  }}/members/{{ record.board_member_id }}` - kind `delete`; body type `none`; path fields
  `board_id`, `board_member_id`; required record fields `board_id`, `board_member_id`; accepted
  fields `board_id`, `board_member_id`; missing records treated as success for status `404`;
  confirmation `destructive`; risk: high: external Miro API mutation; approval required.
- `create_boards_board_id_shapes`: POST `/v2/boards/{{ record.board_id }}/shapes` - kind `create`;
  body type `json`; path fields `board_id`; required record fields `board_id`; accepted fields
  `board_id`, `data`, `geometry`, `parent`, `position`, `style`; risk: medium: external Miro API
  mutation; approval required.
- `update_boards_board_id_shapes_item_id`: PATCH `/v2/boards/{{ record.board_id }}/shapes/{{
  record.item_id }}` - kind `update`; body type `json`; path fields `board_id`, `item_id`; required
  record fields `board_id`, `item_id`; accepted fields `board_id`, `data`, `geometry`, `item_id`,
  `parent`, `position`, `style`; risk: medium: external Miro API mutation; approval required.
- `delete_boards_board_id_shapes_item_id`: DELETE `/v2/boards/{{ record.board_id }}/shapes/{{
  record.item_id }}` - kind `delete`; body type `none`; path fields `board_id`, `item_id`; required
  record fields `board_id`, `item_id`; accepted fields `board_id`, `item_id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: high: external Miro API
  mutation; approval required.
- `create_boards_board_id_sticky_notes`: POST `/v2/boards/{{ record.board_id }}/sticky_notes` - kind
  `create`; body type `json`; path fields `board_id`; required record fields `board_id`; accepted
  fields `board_id`, `data`, `geometry`, `parent`, `position`, `style`; risk: medium: external Miro
  API mutation; approval required.
- `update_boards_board_id_sticky_notes_item_id`: PATCH `/v2/boards/{{ record.board_id
  }}/sticky_notes/{{ record.item_id }}` - kind `update`; body type `json`; path fields `board_id`,
  `item_id`; required record fields `board_id`, `item_id`; accepted fields `board_id`, `data`,
  `geometry`, `item_id`, `parent`, `position`, `style`; risk: medium: external Miro API mutation;
  approval required.
- `delete_boards_board_id_sticky_notes_item_id`: DELETE `/v2/boards/{{ record.board_id
  }}/sticky_notes/{{ record.item_id }}` - kind `delete`; body type `none`; path fields `board_id`,
  `item_id`; required record fields `board_id`, `item_id`; accepted fields `board_id`, `item_id`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: high:
  external Miro API mutation; approval required.
- `create_boards_board_id_texts`: POST `/v2/boards/{{ record.board_id }}/texts` - kind `create`;
  body type `json`; path fields `board_id`; required record fields `board_id`, `data`; accepted
  fields `board_id`, `data`, `geometry`, `parent`, `position`, `style`; risk: medium: external Miro
  API mutation; approval required.
- `update_boards_board_id_texts_item_id`: PATCH `/v2/boards/{{ record.board_id }}/texts/{{
  record.item_id }}` - kind `update`; body type `json`; path fields `board_id`, `item_id`; required
  record fields `board_id`, `item_id`; accepted fields `board_id`, `data`, `geometry`, `item_id`,
  `parent`, `position`, `style`; risk: medium: external Miro API mutation; approval required.
- `delete_boards_board_id_texts_item_id`: DELETE `/v2/boards/{{ record.board_id }}/texts/{{
  record.item_id }}` - kind `delete`; body type `none`; path fields `board_id`, `item_id`; required
  record fields `board_id`, `item_id`; accepted fields `board_id`, `item_id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: high: external Miro API
  mutation; approval required.
- `create_boards_board_id_frames`: POST `/v2/boards/{{ record.board_id }}/frames` - kind `create`;
  body type `json`; path fields `board_id`; required record fields `board_id`, `data`; accepted
  fields `board_id`, `data`, `geometry`, `position`, `style`; risk: medium: external Miro API
  mutation; approval required.
- `update_boards_board_id_frames_item_id`: PATCH `/v2/boards/{{ record.board_id }}/frames/{{
  record.item_id }}` - kind `update`; body type `json`; path fields `board_id`, `item_id`; required
  record fields `board_id`, `item_id`; accepted fields `board_id`, `data`, `geometry`, `item_id`,
  `position`, `style`; risk: medium: external Miro API mutation; approval required.
- `delete_boards_board_id_frames_item_id`: DELETE `/v2/boards/{{ record.board_id }}/frames/{{
  record.item_id }}` - kind `delete`; body type `none`; path fields `board_id`, `item_id`; required
  record fields `board_id`, `item_id`; accepted fields `board_id`, `item_id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: high: external Miro API
  mutation; approval required.
- `delete_experimental_boards_board_id_mindmap_nodes_item_id`: DELETE `/v2-experimental/boards/{{
  record.board_id }}/mindmap_nodes/{{ record.item_id }}` - kind `delete`; body type `none`; path
  fields `board_id`, `item_id`; required record fields `board_id`, `item_id`; accepted fields
  `board_id`, `item_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: high: external Miro API mutation; approval required.
- `create_experimental_boards_board_id_mindmap_nodes`: POST `/v2-experimental/boards/{{
  record.board_id }}/mindmap_nodes` - kind `create`; body type `json`; path fields `board_id`;
  required record fields `board_id`, `data`; accepted fields `board_id`, `data`, `geometry`,
  `parent`, `position`; risk: medium: external Miro API mutation; approval required.
- `delete_experimental_boards_board_id_items_item_id`: DELETE `/v2-experimental/boards/{{
  record.board_id }}/items/{{ record.item_id }}` - kind `delete`; body type `none`; path fields
  `board_id`, `item_id`; required record fields `board_id`, `item_id`; accepted fields `board_id`,
  `item_id`; missing records treated as success for status `404`; confirmation `destructive`; risk:
  high: external Miro API mutation; approval required.
- `create_experimental_boards_board_id_shapes`: POST `/v2-experimental/boards/{{ record.board_id
  }}/shapes` - kind `create`; body type `json`; path fields `board_id`; required record fields
  `board_id`; accepted fields `board_id`, `data`, `geometry`, `parent`, `position`, `style`; risk:
  medium: external Miro API mutation; approval required.
- `update_experimental_boards_board_id_shapes_item_id`: PATCH `/v2-experimental/boards/{{
  record.board_id }}/shapes/{{ record.item_id }}` - kind `update`; body type `json`; path fields
  `board_id`, `item_id`; required record fields `board_id`, `item_id`; accepted fields `board_id`,
  `data`, `geometry`, `item_id`, `parent`, `position`, `style`; risk: medium: external Miro API
  mutation; approval required.
- `delete_experimental_boards_board_id_shapes_item_id`: DELETE `/v2-experimental/boards/{{
  record.board_id }}/shapes/{{ record.item_id }}` - kind `delete`; body type `none`; path fields
  `board_id`, `item_id`; required record fields `board_id`, `item_id`; accepted fields `board_id`,
  `item_id`; missing records treated as success for status `404`; confirmation `destructive`; risk:
  high: external Miro API mutation; approval required.
- `create_experimental_boards_board_id_code_widgets`: POST `/v2-experimental/boards/{{
  record.board_id }}/code_widgets` - kind `create`; body type `json`; path fields `board_id`;
  required record fields `board_id`; accepted fields `board_id`, `data`, `geometry`, `parent`,
  `position`; risk: medium: external Miro API mutation; approval required.
- `update_experimental_boards_board_id_code_widgets_item_id`: PATCH `/v2-experimental/boards/{{
  record.board_id }}/code_widgets/{{ record.item_id }}` - kind `update`; body type `json`; path
  fields `board_id`, `item_id`; required record fields `board_id`, `item_id`; accepted fields
  `board_id`, `data`, `geometry`, `item_id`, `parent`, `position`; risk: medium: external Miro API
  mutation; approval required.
- `delete_experimental_boards_board_id_code_widgets_item_id`: DELETE `/v2-experimental/boards/{{
  record.board_id }}/code_widgets/{{ record.item_id }}` - kind `delete`; body type `none`; path
  fields `board_id`, `item_id`; required record fields `board_id`, `item_id`; accepted fields
  `board_id`, `item_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: high: external Miro API mutation; approval required.
- `update_experimental_boards_board_id_code_widgets_item_id_position`: PATCH
  `/v2-experimental/boards/{{ record.board_id }}/code_widgets/{{ record.item_id }}/position` - kind
  `update`; body type `json`; path fields `board_id`, `item_id`; required record fields `board_id`,
  `item_id`; accepted fields `board_id`, `item_id`, `x`, `y`; risk: medium: external Miro API
  mutation; approval required.
- `create_boards_board_id_groups`: POST `/v2/boards/{{ record.board_id }}/groups` - kind `create`;
  body type `json`; path fields `board_id`; required record fields `board_id`, `id`, `name`, `type`;
  accepted fields `board_id`, `description`, `id`, `name`, `type`; risk: medium: external Miro API
  mutation; approval required.
- `update_boards_board_id_groups_group_id`: PUT `/v2/boards/{{ record.board_id }}/groups/{{
  record.group_id }}` - kind `update`; body type `json`; path fields `board_id`, `group_id`;
  required record fields `board_id`, `group_id`, `id`, `name`, `type`; accepted fields `board_id`,
  `description`, `group_id`, `id`, `name`, `type`; risk: medium: external Miro API mutation;
  approval required.
- `delete_boards_board_id_groups_group_id`: DELETE `/v2/boards/{{ record.board_id }}/groups/{{
  record.group_id }}` - kind `delete`; body type `none`; path fields `board_id`, `group_id`;
  required record fields `board_id`, `group_id`; accepted fields `board_id`, `group_id`; missing
  records treated as success for status `404`; confirmation `destructive`; risk: high: external Miro
  API mutation; approval required.
- `create_boards_board_id_tags`: POST `/v2/boards/{{ record.board_id }}/tags` - kind `create`; body
  type `json`; path fields `board_id`; required record fields `board_id`, `title`; accepted fields
  `board_id`, `fillColor`, `title`; risk: medium: external Miro API mutation; approval required.
- `update_boards_board_id_tags_tag_id`: PATCH `/v2/boards/{{ record.board_id }}/tags/{{
  record.tag_id }}` - kind `update`; body type `json`; path fields `board_id`, `tag_id`; required
  record fields `board_id`, `tag_id`; accepted fields `board_id`, `fillColor`, `tag_id`, `title`;
  risk: medium: external Miro API mutation; approval required.
- `delete_boards_board_id_tags_tag_id`: DELETE `/v2/boards/{{ record.board_id }}/tags/{{
  record.tag_id }}` - kind `delete`; body type `none`; path fields `board_id`, `tag_id`; required
  record fields `board_id`, `tag_id`; accepted fields `board_id`, `tag_id`; missing records treated
  as success for status `404`; confirmation `destructive`; risk: high: external Miro API mutation;
  approval required.
- `create_orgs_org_id_teams_team_id_projects`: POST `/v2/orgs/{{ record.org_id }}/teams/{{
  record.team_id }}/projects` - kind `create`; body type `json`; path fields `org_id`, `team_id`;
  required record fields `org_id`, `team_id`, `name`; accepted fields `name`, `org_id`, `team_id`;
  risk: medium: external Miro API mutation; approval required.
- `update_orgs_org_id_teams_team_id_projects_project_id`: PATCH `/v2/orgs/{{ record.org_id
  }}/teams/{{ record.team_id }}/projects/{{ record.project_id }}` - kind `update`; body type `json`;
  path fields `org_id`, `team_id`, `project_id`; required record fields `org_id`, `team_id`,
  `project_id`, `name`; accepted fields `name`, `org_id`, `project_id`, `team_id`; risk: medium:
  external Miro API mutation; approval required.
- `delete_orgs_org_id_teams_team_id_projects_project_id`: DELETE `/v2/orgs/{{ record.org_id
  }}/teams/{{ record.team_id }}/projects/{{ record.project_id }}` - kind `delete`; body type `none`;
  path fields `org_id`, `team_id`, `project_id`; required record fields `org_id`, `team_id`,
  `project_id`; accepted fields `org_id`, `project_id`, `team_id`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: high: external Miro API mutation;
  approval required.
- `update_orgs_org_id_teams_team_id_projects_project_id_settings`: PATCH `/v2/orgs/{{ record.org_id
  }}/teams/{{ record.team_id }}/projects/{{ record.project_id }}/settings` - kind `update`; body
  type `json`; path fields `org_id`, `team_id`, `project_id`; required record fields `org_id`,
  `team_id`, `project_id`; accepted fields `org_id`, `project_id`, `sharingPolicySettings`,
  `team_id`; risk: medium: external Miro API mutation; approval required.
- `create_orgs_org_id_teams_team_id_projects_project_id_members`: POST `/v2/orgs/{{ record.org_id
  }}/teams/{{ record.team_id }}/projects/{{ record.project_id }}/members` - kind `create`; body type
  `json`; path fields `org_id`, `team_id`, `project_id`; required record fields `org_id`, `team_id`,
  `project_id`, `email`, `role`; accepted fields `email`, `org_id`, `project_id`, `role`, `team_id`;
  risk: medium: external Miro API mutation; approval required.
- `update_orgs_org_id_teams_team_id_projects_project_id_members_member_id`: PATCH `/v2/orgs/{{
  record.org_id }}/teams/{{ record.team_id }}/projects/{{ record.project_id }}/members/{{
  record.member_id }}` - kind `update`; body type `json`; path fields `org_id`, `team_id`,
  `project_id`, `member_id`; required record fields `org_id`, `team_id`, `project_id`, `member_id`;
  accepted fields `member_id`, `org_id`, `project_id`, `role`, `team_id`; risk: medium: external
  Miro API mutation; approval required.
- `delete_orgs_org_id_teams_team_id_projects_project_id_members_member_id`: DELETE `/v2/orgs/{{
  record.org_id }}/teams/{{ record.team_id }}/projects/{{ record.project_id }}/members/{{
  record.member_id }}` - kind `delete`; body type `none`; path fields `org_id`, `team_id`,
  `project_id`, `member_id`; required record fields `org_id`, `team_id`, `project_id`, `member_id`;
  accepted fields `member_id`, `org_id`, `project_id`, `team_id`; missing records treated as success
  for status `404`; confirmation `destructive`; risk: high: external Miro API mutation; approval
  required.
- `create_orgs_org_id_teams`: POST `/v2/orgs/{{ record.org_id }}/teams` - kind `create`; body type
  `json`; path fields `org_id`; required record fields `org_id`, `name`; accepted fields `name`,
  `org_id`; risk: medium: external Miro API mutation; approval required.
- `update_orgs_org_id_teams_team_id`: PATCH `/v2/orgs/{{ record.org_id }}/teams/{{ record.team_id
  }}` - kind `update`; body type `json`; path fields `org_id`, `team_id`; required record fields
  `org_id`, `team_id`; accepted fields `name`, `org_id`, `team_id`; risk: medium: external Miro API
  mutation; approval required.
- `delete_orgs_org_id_teams_team_id`: DELETE `/v2/orgs/{{ record.org_id }}/teams/{{ record.team_id
  }}` - kind `delete`; body type `none`; path fields `org_id`, `team_id`; required record fields
  `org_id`, `team_id`; accepted fields `org_id`, `team_id`; missing records treated as success for
  status `404`; confirmation `destructive`; risk: high: external Miro API mutation; approval
  required.
- `create_orgs_org_id_teams_team_id_members`: POST `/v2/orgs/{{ record.org_id }}/teams/{{
  record.team_id }}/members` - kind `create`; body type `json`; path fields `org_id`, `team_id`;
  required record fields `org_id`, `team_id`, `email`; accepted fields `email`, `org_id`, `role`,
  `team_id`; risk: medium: external Miro API mutation; approval required.
- `update_orgs_org_id_teams_team_id_members_member_id`: PATCH `/v2/orgs/{{ record.org_id }}/teams/{{
  record.team_id }}/members/{{ record.member_id }}` - kind `update`; body type `json`; path fields
  `org_id`, `team_id`, `member_id`; required record fields `org_id`, `team_id`, `member_id`;
  accepted fields `member_id`, `org_id`, `role`, `team_id`; risk: medium: external Miro API
  mutation; approval required.
- `delete_orgs_org_id_teams_team_id_members_member_id`: DELETE `/v2/orgs/{{ record.org_id
  }}/teams/{{ record.team_id }}/members/{{ record.member_id }}` - kind `delete`; body type `none`;
  path fields `org_id`, `team_id`, `member_id`; required record fields `org_id`, `team_id`,
  `member_id`; accepted fields `member_id`, `org_id`, `team_id`; missing records treated as success
  for status `404`; confirmation `destructive`; risk: high: external Miro API mutation; approval
  required.
- `update_orgs_org_id_teams_team_id_settings`: PATCH `/v2/orgs/{{ record.org_id }}/teams/{{
  record.team_id }}/settings` - kind `update`; body type `json`; path fields `org_id`, `team_id`;
  required record fields `org_id`, `team_id`; accepted fields `org_id`,
  `teamAccountDiscoverySettings`, `teamCollaborationSettings`, `teamCopyAccessLevelSettings`,
  `teamInvitationSettings`, `teamSharingPolicySettings`, `team_id`; risk: medium: external Miro API
  mutation; approval required.
- `create_orgs_org_id_groups`: POST `/v2/orgs/{{ record.org_id }}/groups` - kind `create`; body type
  `json`; path fields `org_id`; required record fields `org_id`, `name`; accepted fields
  `description`, `name`, `org_id`; risk: medium: external Miro API mutation; approval required.
- `update_orgs_org_id_groups_group_id`: PATCH `/v2/orgs/{{ record.org_id }}/groups/{{
  record.group_id }}` - kind `update`; body type `json`; path fields `org_id`, `group_id`; required
  record fields `org_id`, `group_id`; accepted fields `description`, `group_id`, `name`, `org_id`;
  risk: medium: external Miro API mutation; approval required.
- `delete_orgs_org_id_groups_group_id`: DELETE `/v2/orgs/{{ record.org_id }}/groups/{{
  record.group_id }}` - kind `delete`; body type `none`; path fields `org_id`, `group_id`; required
  record fields `org_id`, `group_id`; accepted fields `group_id`, `org_id`; missing records treated
  as success for status `404`; confirmation `destructive`; risk: high: external Miro API mutation;
  approval required.
- `create_orgs_org_id_groups_group_id_members`: POST `/v2/orgs/{{ record.org_id }}/groups/{{
  record.group_id }}/members` - kind `create`; body type `json`; path fields `org_id`, `group_id`;
  required record fields `org_id`, `group_id`, `email`; accepted fields `email`, `group_id`,
  `org_id`; risk: medium: external Miro API mutation; approval required.
- `update_orgs_org_id_groups_group_id_members`: PATCH `/v2/orgs/{{ record.org_id }}/groups/{{
  record.group_id }}/members` - kind `update`; body type `json`; path fields `org_id`, `group_id`;
  required record fields `org_id`, `group_id`; accepted fields `group_id`, `membersToAdd`,
  `membersToRemove`, `org_id`; risk: medium: external Miro API mutation; approval required.
- `delete_orgs_org_id_groups_group_id_members_member_id`: DELETE `/v2/orgs/{{ record.org_id
  }}/groups/{{ record.group_id }}/members/{{ record.member_id }}` - kind `delete`; body type `none`;
  path fields `org_id`, `group_id`, `member_id`; required record fields `org_id`, `group_id`,
  `member_id`; accepted fields `group_id`, `member_id`, `org_id`; missing records treated as success
  for status `404`; confirmation `destructive`; risk: high: external Miro API mutation; approval
  required.
- `create_orgs_org_id_teams_team_id_groups`: POST `/v2/orgs/{{ record.org_id }}/teams/{{
  record.team_id }}/groups` - kind `create`; body type `json`; path fields `org_id`, `team_id`;
  required record fields `org_id`, `team_id`, `userGroupId`, `role`; accepted fields `org_id`,
  `role`, `team_id`, `userGroupId`; risk: medium: external Miro API mutation; approval required.
- `delete_orgs_org_id_teams_team_id_groups_group_id`: DELETE `/v2/orgs/{{ record.org_id }}/teams/{{
  record.team_id }}/groups/{{ record.group_id }}` - kind `delete`; body type `none`; path fields
  `org_id`, `team_id`, `group_id`; required record fields `org_id`, `team_id`, `group_id`; accepted
  fields `group_id`, `org_id`, `team_id`; missing records treated as success for status `404`;
  confirmation `destructive`; risk: high: external Miro API mutation; approval required.
- `create_orgs_org_id_boards_board_id_groups`: POST `/v2/orgs/{{ record.org_id }}/boards/{{
  record.board_id }}/groups` - kind `create`; body type `json`; path fields `org_id`, `board_id`;
  required record fields `org_id`, `board_id`, `userGroupIds`, `role`; accepted fields `board_id`,
  `org_id`, `role`, `userGroupIds`; risk: medium: external Miro API mutation; approval required.
- `delete_orgs_org_id_boards_board_id_groups_group_id`: DELETE `/v2/orgs/{{ record.org_id
  }}/boards/{{ record.board_id }}/groups/{{ record.group_id }}` - kind `delete`; body type `none`;
  path fields `org_id`, `board_id`, `group_id`; required record fields `org_id`, `board_id`,
  `group_id`; accepted fields `board_id`, `group_id`, `org_id`; missing records treated as success
  for status `404`; confirmation `destructive`; risk: high: external Miro API mutation; approval
  required.
- `create_orgs_org_id_projects_project_id_groups`: POST `/v2/orgs/{{ record.org_id }}/projects/{{
  record.project_id }}/groups` - kind `create`; body type `json`; path fields `org_id`,
  `project_id`; required record fields `org_id`, `project_id`, `userGroupIds`, `role`; accepted
  fields `org_id`, `project_id`, `role`, `userGroupIds`; risk: medium: external Miro API mutation;
  approval required.
- `delete_orgs_org_id_projects_project_id_groups_group_id`: DELETE `/v2/orgs/{{ record.org_id
  }}/projects/{{ record.project_id }}/groups/{{ record.group_id }}` - kind `delete`; body type
  `none`; path fields `org_id`, `project_id`, `group_id`; required record fields `org_id`,
  `project_id`, `group_id`; accepted fields `group_id`, `org_id`, `project_id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: high: external Miro API
  mutation; approval required.

## Known limits

- Batch defaults: read_page_size=50, write_batch_size=1.
- API coverage includes 84 stream-backed endpoint group(s), 98 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=5, deprecated=1, non_data_endpoint=2, out_of_scope=7.
