# pm connectors inspect gitbook

```text
NAME
  pm connectors inspect gitbook - GitBook connector manual

SYNOPSIS
  pm connectors inspect gitbook
  pm connectors inspect gitbook --json
  pm credentials add <name> --connector gitbook [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads 185 GitBook REST resources and executes 170 JSON/no-body GitBook mutations through the GitBook API.

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
  account_name
  base_url
  change_request_id
  collection_id
  comment_id
  comment_reply_id
  conversation_id
  document_id
  email_domain
  event_id
  file_id
  font_id
  glossary_entry_id
  hostname
  import_run_id
  installation_id
  integration_name
  invite_id
  organization_id
  page_id
  page_path
  project_id
  query
  repository_name
  request
  reusable_content_id
  review_id
  revision_id
  saml_provider_id
  share_link_id
  site_channel_id
  site_context_connection_id
  site_context_record_id
  site_finding_id
  site_id
  site_mcp_server_id
  site_question_answer_id
  site_question_id
  site_redirect_id
  site_scan_id
  site_section_group_id
  site_section_id
  site_space_id
  site_topic_id
  source
  space_id
  spec_slug
  subdomain
  team_id
  translation_id
  url
  user_id
  version_id
  access_token (secret)

ETL STREAMS
  users:
    primary key: id
    fields: display_name(), email(), id(), photo_url()
  organizations:
    primary key: id
    fields: created_at(), id(), title(), type(), url()
  org_members:
    primary key: id
    fields: display_name(), email(), id(), role()
  content:
    primary key: id
    fields: id(), kind(), path(), slug(), title(), type()
  get_api_information:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  get_user_by_id:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  get_space_by_id:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  get_embed_by_url_in_space:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  search_space_content:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  get_space_git_info:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  list_user_permissions_in_space:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  list_team_permissions_in_space:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  get_current_revision:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  list_files:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  get_file_by_id:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  list_space_file_backlinks:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  get_page_by_id:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  list_page_links_in_space:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  list_space_page_backlinks:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  list_space_page_meta_links:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  get_page_by_path:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  get_reusable_content_by_id:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  get_document_by_id:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  list_change_requests_for_space:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  get_change_request_by_id:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  get_reviews_by_change_request_id:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  get_change_request_review_by_id:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  get_requested_reviewers_by_change_request_id:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  list_change_request_conversations:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  list_change_request_links:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  list_comments_in_change_request:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  get_comment_in_change_request:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  list_comment_replies_in_change_request:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  get_comment_reply_in_change_request:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  get_contributors_by_change_request_id:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  get_revision_of_change_request_by_id:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  list_pages_in_change_request:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  list_files_in_change_request_by_id:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  get_file_in_change_request_by_id:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  list_change_request_file_backlinks:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  get_page_in_change_request_by_id:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  list_page_links_in_change_request:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  list_change_request_page_backlinks:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  list_change_request_page_meta_links:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  get_reusable_content_in_change_request_by_id:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  get_change_request_changes:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  get_change_request_pdf:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  get_revision_by_id:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  get_revision_semantic_changes:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  list_pages_in_revision_by_id:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  list_files_in_revision_by_id:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  get_file_in_revision_by_id:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  get_page_in_revision_by_id:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  get_page_document_in_revision_by_id:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  get_page_in_revision_by_path:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  list_revision_page_meta_links:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  get_page_in_change_request_by_path:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  get_reusable_content_in_revision_by_id:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  get_reusable_content_document_in_revision_by_id:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  list_comments_in_space:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  get_comment_in_space:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  list_comment_replies_in_space:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  get_comment_reply_in_space:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  list_commenters_in_space:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  list_commenters_in_change_request:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  list_permissions_aggregate_in_space:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  list_space_integrations:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  list_space_integrations_blocks:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  get_space_pdf:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  list_space_links:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  get_collection_by_id:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  list_spaces_in_collection_by_id:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  list_team_permissions_in_collection:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  list_user_permissions_in_collection:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  list_permissions_aggregate_in_collection:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  list_integrations:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  get_integration_by_name:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  list_integration_installations:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  list_integration_events:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  get_integration_event:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  list_integration_space_installations:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  list_integration_site_installations:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  render_integration_ui_with_get:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  get_integration_installation_by_id:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  list_integration_installation_spaces:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  get_integration_space_installation:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  list_integration_installation_sites:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  get_integration_site_installation:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  get_organization_by_id:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  get_member_in_organization_by_id:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  list_spaces_for_organization_member:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  list_teams_for_organization_member:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  list_teams_in_organization_by_id:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  get_team_in_organization_by_id:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  list_team_members_in_organization_by_id:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  list_organization_invite_links:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  get_organization_invite_link:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  search_organization_content:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  list_change_requests_for_organization:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  list_spaces_in_organization_by_id:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  list_collections_in_organization_by_id:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  list_organization_integrations:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  get_organization_integration_status:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  list_organization_installations:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  list_organization_integrations_status:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  list_saml_providers_in_organization_by_id:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  get_organization_saml_provider_by_id:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  list_sso_provider_logins_in_organization:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  get_recommended_questions_in_organization:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  list_open_api_specs:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  get_open_api_spec_by_slug:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  list_open_api_spec_versions:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  get_latest_open_api_spec_version:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  get_latest_open_api_spec_version_content:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  get_open_api_spec_version_by_id:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  get_open_api_spec_version_content_by_id:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  get_organization_agent_instructions:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  list_translations:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  get_translation:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  list_glossary_entries:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  get_glossary_entry:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  list_custom_fonts:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  get_custom_font:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  list_sites:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  get_site_by_id:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  list_site_git_sync_installations:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  get_site_adaptive_schema:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  list_site_adaptive_template_conditions:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  get_published_content_site:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  list_site_share_links:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  get_site_structure:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  get_site_publishing_auth_by_id:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  get_site_publishing_preview_by_id:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  get_site_customization_by_id:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  list_site_integration_scripts:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  list_site_integrations:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  list_site_spaces:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  list_site_section_groups:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  list_site_sections:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  list_site_context_records:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  get_site_context_record_by_id:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  list_site_scans:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  get_site_scan_by_id:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  list_site_findings:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  get_site_finding_by_id:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  list_change_requests_for_site_finding:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  list_pages_for_site_finding:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  list_questions_for_site_finding:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  list_records_for_site_finding:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  list_site_context_connections:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  get_site_context_connection_by_id:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  list_site_topics:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  get_site_topic_by_id:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  list_site_questions:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  get_site_question_by_id:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  list_site_question_sources:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  get_site_question_stats:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  list_site_question_answers:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  get_site_question_answer_by_id:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  get_site_question_answer_thread_by_id:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  list_site_question_answer_sources:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  get_site_space_customization_by_id:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  list_permissions_aggregate_in_site:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  list_user_permissions_in_site:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  list_team_permissions_in_site:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  get_site_agent_settings_by_id:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  list_site_visitor_segments:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  list_site_redirects:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  get_site_redirect_by_source:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  list_site_mcp_servers:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  get_site_mcp_server_by_id:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  list_site_channels:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  get_site_channel_by_id:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  get_subdomain:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  get_custom_hostname:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  get_organizations_for_email_domain:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  ads_list_sites:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  get_content_by_url:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  get_embed_by_url:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  get_published_content_by_url:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  get_git_sync_installation_by_id:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  list_git_hub_repositories_for_git_sync_installation:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  list_git_hub_repo_branches_for_git_sync_installation:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  list_git_lab_projects_for_git_sync_installation:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()
  list_git_lab_project_branches_for_git_sync_installation:
    primary key: id
    fields: id(), name(), object(), operation_id(), status(), title()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  create_user_notifications_token:
    endpoint: POST /user/notifications/token
    risk: POST /user/notifications/token (Create a JWT to access the in-app notifications service) executes a live GitBook API operation.
  update_user_by_id:
    endpoint: PATCH /users/{{ record.user_id }}
    required fields: user_id
    risk: PATCH /users/{userId} (Update a user by its ID) executes a live GitBook API operation.
  update_space_by_id:
    endpoint: PATCH /spaces/{{ record.space_id }}
    required fields: space_id
    risk: PATCH /spaces/{spaceId} (Update a space's title, icon, or settings) executes a live GitBook API operation.
  delete_space_by_id:
    endpoint: DELETE /spaces/{{ record.space_id }}
    required fields: space_id
    risk: DELETE /spaces/{spaceId} (Delete a space) executes a live GitBook API operation.
  duplicate_space:
    endpoint: POST /spaces/{{ record.space_id }}/duplicate
    required fields: space_id
    risk: POST /spaces/{spaceId}/duplicate (Create a full copy of a space) executes a live GitBook API operation.
  restore_space:
    endpoint: POST /spaces/{{ record.space_id }}/restore
    required fields: space_id
    risk: POST /spaces/{spaceId}/restore (Restore a recently deleted space from the trash) executes a live GitBook API operation.
  move_space:
    endpoint: POST /spaces/{{ record.space_id }}/move
    required fields: space_id
    risk: POST /spaces/{spaceId}/move (Move a space to a different collection or position) executes a live GitBook API operation.
  import_git_repository:
    endpoint: POST /spaces/{{ record.space_id }}/git/import
    required fields: space_id
    risk: POST /spaces/{spaceId}/git/import (Pull content into a space from a connected Git repository) executes a live GitBook API operation.
  export_to_git_repository:
    endpoint: POST /spaces/{{ record.space_id }}/git/export
    required fields: space_id
    risk: POST /spaces/{spaceId}/git/export (Push space content to a connected Git repository) executes a live GitBook API operation.
  delete_legacy_git_installation:
    endpoint: DELETE /spaces/{{ record.space_id }}/git/legacy-installation
    required fields: space_id
    risk: DELETE /spaces/{spaceId}/git/legacy-installation (Remove the legacy Git Sync installation from the space to be able to upgrade it to use the new Git integrations) executes a live GitBook API operation.
  invite_to_space:
    endpoint: POST /spaces/{{ record.space_id }}/permissions
    required fields: space_id
    risk: POST /spaces/{spaceId}/permissions (Invite a user or a team to a space) executes a live GitBook API operation.
  update_team_permission_in_space:
    endpoint: PATCH /spaces/{{ record.space_id }}/permissions/teams/{{ record.team_id }}
    required fields: space_id, team_id
    risk: PATCH /spaces/{spaceId}/permissions/teams/{teamId} (Update an org team's permission in a space) executes a live GitBook API operation.
  remove_team_from_space:
    endpoint: DELETE /spaces/{{ record.space_id }}/permissions/teams/{{ record.team_id }}
    required fields: space_id, team_id
    risk: DELETE /spaces/{spaceId}/permissions/teams/{teamId} (Remove an org team from a space) executes a live GitBook API operation.
  update_user_permission_in_space:
    endpoint: PATCH /spaces/{{ record.space_id }}/permissions/users/{{ record.user_id }}
    required fields: space_id, user_id
    risk: PATCH /spaces/{spaceId}/permissions/users/{userId} (Update space user permissions) executes a live GitBook API operation.
  remove_user_from_space:
    endpoint: DELETE /spaces/{{ record.space_id }}/permissions/users/{{ record.user_id }}
    required fields: space_id, user_id
    risk: DELETE /spaces/{spaceId}/permissions/users/{userId} (Remove a space user) executes a live GitBook API operation.
  apply_template_to_space:
    endpoint: POST /spaces/{{ record.space_id }}/content/template
    required fields: space_id
    risk: POST /spaces/{spaceId}/content/template (Apply a content template to populate a space with initial pages) executes a live GitBook API operation.
  get_computed_document:
    endpoint: POST /spaces/{{ record.space_id }}/content/computed/document
    required fields: space_id
    risk: POST /spaces/{spaceId}/content/computed/document (Compute and render a document from a structured content source) executes a live GitBook API operation.
  get_computed_revision:
    endpoint: POST /spaces/{{ record.space_id }}/content/computed/revision
    required fields: space_id
    risk: POST /spaces/{spaceId}/content/computed/revision (Compute and render a full revision from a structured content source) executes a live GitBook API operation.
  create_change_request:
    endpoint: POST /spaces/{{ record.space_id }}/change-requests
    required fields: space_id
    risk: POST /spaces/{spaceId}/change-requests (Create a new change request in a space) executes a live GitBook API operation.
  update_change_request_by_id:
    endpoint: PATCH /spaces/{{ record.space_id }}/change-requests/{{ record.change_request_id }}
    required fields: space_id, change_request_id
    risk: PATCH /spaces/{spaceId}/change-requests/{changeRequestId} (Update a change request's subject, description, or status) executes a live GitBook API operation.
  merge_change_request:
    endpoint: POST /spaces/{{ record.space_id }}/change-requests/{{ record.change_request_id }}/merge
    required fields: space_id, change_request_id
    risk: POST /spaces/{spaceId}/change-requests/{changeRequestId}/merge (Merge a change request into the space's live content) executes a live GitBook API operation.
  update_change_request:
    endpoint: POST /spaces/{{ record.space_id }}/change-requests/{{ record.change_request_id }}/update
    required fields: space_id, change_request_id
    risk: POST /spaces/{spaceId}/change-requests/{changeRequestId}/update (Sync a change request with the latest live space content) executes a live GitBook API operation.
  submit_change_request_review:
    endpoint: POST /spaces/{{ record.space_id }}/change-requests/{{ record.change_request_id }}/reviews
    required fields: space_id, change_request_id
    risk: POST /spaces/{spaceId}/change-requests/{changeRequestId}/reviews (Submit an approve or request-changes review for a change request) executes a live GitBook API operation.
  request_reviewers_for_change_request:
    endpoint: POST /spaces/{{ record.space_id }}/change-requests/{{ record.change_request_id }}/requested-reviewers
    required fields: space_id, change_request_id
    risk: POST /spaces/{spaceId}/change-requests/{changeRequestId}/requested-reviewers (Send review requests to users for a change request) executes a live GitBook API operation.
  remove_requested_reviewer_from_change_request:
    endpoint: DELETE /spaces/{{ record.space_id }}/change-requests/{{ record.change_request_id }}/requested-reviewers/{{ record.user_id }}
    required fields: space_id, change_request_id, user_id
    risk: DELETE /spaces/{spaceId}/change-requests/{changeRequestId}/requested-reviewers/{userId} (Remove a reviewer from a change request) executes a live GitBook API operation.
  update_change_request_conversation:
    endpoint: PATCH /spaces/{{ record.space_id }}/change-requests/{{ record.change_request_id }}/conversations/{{ record.conversation_id }}
    required fields: space_id, change_request_id, conversation_id
    risk: PATCH /spaces/{spaceId}/change-requests/{changeRequestId}/conversations/{conversationId} (Update the title of an AI agent conversation on a change request) executes a live GitBook API operation.
  delete_change_request_conversation:
    endpoint: DELETE /spaces/{{ record.space_id }}/change-requests/{{ record.change_request_id }}/conversations/{{ record.conversation_id }}
    required fields: space_id, change_request_id, conversation_id
    risk: DELETE /spaces/{spaceId}/change-requests/{changeRequestId}/conversations/{conversationId} (Delete an agent conversation) executes a live GitBook API operation.
  post_comment_in_change_request:
    endpoint: POST /spaces/{{ record.space_id }}/change-requests/{{ record.change_request_id }}/comments
    required fields: space_id, change_request_id
    risk: POST /spaces/{spaceId}/change-requests/{changeRequestId}/comments (Post a new comment on a change request) executes a live GitBook API operation.
  update_comment_in_change_request:
    endpoint: PUT /spaces/{{ record.space_id }}/change-requests/{{ record.change_request_id }}/comments/{{ record.comment_id }}
    required fields: space_id, change_request_id, comment_id
    risk: PUT /spaces/{spaceId}/change-requests/{changeRequestId}/comments/{commentId} (Update the content or status of a change request comment) executes a live GitBook API operation.
  delete_comment_in_change_request:
    endpoint: DELETE /spaces/{{ record.space_id }}/change-requests/{{ record.change_request_id }}/comments/{{ record.comment_id }}
    required fields: space_id, change_request_id, comment_id
    risk: DELETE /spaces/{spaceId}/change-requests/{changeRequestId}/comments/{commentId} (Delete a change request comment) executes a live GitBook API operation.
  post_comment_reply_in_change_request:
    endpoint: POST /spaces/{{ record.space_id }}/change-requests/{{ record.change_request_id }}/comments/{{ record.comment_id }}/replies
    required fields: space_id, change_request_id, comment_id
    risk: POST /spaces/{spaceId}/change-requests/{changeRequestId}/comments/{commentId}/replies (Post a reply to a change request comment) executes a live GitBook API operation.
  update_comment_reply_in_change_request:
    endpoint: PUT /spaces/{{ record.space_id }}/change-requests/{{ record.change_request_id }}/comments/{{ record.comment_id }}/replies/{{ record.comment_reply_id }}
    required fields: space_id, change_request_id, comment_id, comment_reply_id
    risk: PUT /spaces/{spaceId}/change-requests/{changeRequestId}/comments/{commentId}/replies/{commentReplyId} (Update the content of a change request comment reply) executes a live GitBook API operation.
  delete_comment_reply_in_change_request:
    endpoint: DELETE /spaces/{{ record.space_id }}/change-requests/{{ record.change_request_id }}/comments/{{ record.comment_id }}/replies/{{ record.comment_reply_id }}
    required fields: space_id, change_request_id, comment_id, comment_reply_id
    risk: DELETE /spaces/{spaceId}/change-requests/{changeRequestId}/comments/{commentId}/replies/{commentReplyId} (Delete a change request comment reply) executes a live GitBook API operation.
  update_change_request_content:
    endpoint: POST /spaces/{{ record.space_id }}/change-requests/{{ record.change_request_id }}/content
    required fields: space_id, change_request_id
    risk: POST /spaces/{spaceId}/change-requests/{changeRequestId}/content (Apply a batch of content changes to a change request) executes a live GitBook API operation.
  post_comment_in_space:
    endpoint: POST /spaces/{{ record.space_id }}/comments
    required fields: space_id
    risk: POST /spaces/{spaceId}/comments (Post a new comment on a space or a specific page) executes a live GitBook API operation.
  update_comment_in_space:
    endpoint: PUT /spaces/{{ record.space_id }}/comments/{{ record.comment_id }}
    required fields: space_id, comment_id
    risk: PUT /spaces/{spaceId}/comments/{commentId} (Update the body or status of a space comment) executes a live GitBook API operation.
  delete_comment_in_space:
    endpoint: DELETE /spaces/{{ record.space_id }}/comments/{{ record.comment_id }}
    required fields: space_id, comment_id
    risk: DELETE /spaces/{spaceId}/comments/{commentId} (Delete a space comment) executes a live GitBook API operation.
  post_comment_reply_in_space:
    endpoint: POST /spaces/{{ record.space_id }}/comments/{{ record.comment_id }}/replies
    required fields: space_id, comment_id
    risk: POST /spaces/{spaceId}/comments/{commentId}/replies (Post a reply to an existing space comment) executes a live GitBook API operation.
  update_comment_reply_in_space:
    endpoint: PUT /spaces/{{ record.space_id }}/comments/{{ record.comment_id }}/replies/{{ record.comment_reply_id }}
    required fields: space_id, comment_id, comment_reply_id
    risk: PUT /spaces/{spaceId}/comments/{commentId}/replies/{commentReplyId} (Update the body of a reply to a space comment) executes a live GitBook API operation.
  delete_comment_reply_in_space:
    endpoint: DELETE /spaces/{{ record.space_id }}/comments/{{ record.comment_id }}/replies/{{ record.comment_reply_id }}
    required fields: space_id, comment_id, comment_reply_id
    risk: DELETE /spaces/{spaceId}/comments/{commentId}/replies/{commentReplyId} (Delete a space comment reply) executes a live GitBook API operation.
  update_collection_by_id:
    endpoint: PATCH /collections/{{ record.collection_id }}
    required fields: collection_id
    risk: PATCH /collections/{collectionId} (Update a collection) executes a live GitBook API operation.
  delete_collection_by_id:
    endpoint: DELETE /collections/{{ record.collection_id }}
    required fields: collection_id
    risk: DELETE /collections/{collectionId} (Delete a collection) executes a live GitBook API operation.
  move_collection:
    endpoint: POST /collections/{{ record.collection_id }}/move
    required fields: collection_id
    risk: POST /collections/{collectionId}/move (Move a collection to a new position.) executes a live GitBook API operation.
  transfer_collection:
    endpoint: POST /collections/{{ record.collection_id }}/transfer
    required fields: collection_id
    risk: POST /collections/{collectionId}/transfer (Transfer a collection) executes a live GitBook API operation.
  invite_to_collection:
    endpoint: POST /collections/{{ record.collection_id }}/permissions
    required fields: collection_id
    risk: POST /collections/{collectionId}/permissions (Invite to a collection) executes a live GitBook API operation.
  update_team_permission_in_collection:
    endpoint: PATCH /collections/{{ record.collection_id }}/permissions/teams/{{ record.team_id }}
    required fields: collection_id, team_id
    risk: PATCH /collections/{collectionId}/permissions/teams/{teamId} (Update an org team's permission in a collection) executes a live GitBook API operation.
  remove_team_from_collection:
    endpoint: DELETE /collections/{{ record.collection_id }}/permissions/teams/{{ record.team_id }}
    required fields: collection_id, team_id
    risk: DELETE /collections/{collectionId}/permissions/teams/{teamId} (Remove an org team from a collection) executes a live GitBook API operation.
  update_user_permission_in_collection:
    endpoint: PATCH /collections/{{ record.collection_id }}/permissions/users/{{ record.user_id }}
    required fields: collection_id, user_id
    risk: PATCH /collections/{collectionId}/permissions/users/{userId} (Update a collection user permission) executes a live GitBook API operation.
  remove_user_from_collection:
    endpoint: DELETE /collections/{{ record.collection_id }}/permissions/users/{{ record.user_id }}
    required fields: collection_id, user_id
    risk: DELETE /collections/{collectionId}/permissions/users/{userId} (Remove a user from a collection) executes a live GitBook API operation.
  publish_integration:
    endpoint: POST /integrations/{{ record.integration_name }}
    required fields: integration_name
    risk: POST /integrations/{integrationName} (Publish an integration) executes a live GitBook API operation.
  unpublish_integration:
    endpoint: DELETE /integrations/{{ record.integration_name }}
    required fields: integration_name
    risk: DELETE /integrations/{integrationName} (Unpublish an integration) executes a live GitBook API operation.
  install_integration:
    endpoint: POST /integrations/{{ record.integration_name }}/installations
    required fields: integration_name
    risk: POST /integrations/{integrationName}/installations (Install an integration) executes a live GitBook API operation.
  set_integration_development_mode:
    endpoint: PUT /integrations/{{ record.integration_name }}/dev
    required fields: integration_name
    risk: PUT /integrations/{integrationName}/dev (Enable integration dev mode) executes a live GitBook API operation.
  disable_integration_development_mode:
    endpoint: DELETE /integrations/{{ record.integration_name }}/dev
    required fields: integration_name
    risk: DELETE /integrations/{integrationName}/dev (Disable integration dev mode) executes a live GitBook API operation.
  render_integration_ui_with_post:
    endpoint: POST /integrations/{{ record.integration_name }}/render
    required fields: integration_name
    risk: POST /integrations/{integrationName}/render (Render an integration UI with POST method) executes a live GitBook API operation.
  queue_integration_task:
    endpoint: POST /integrations/{{ record.integration_name }}/tasks
    required fields: integration_name
    risk: POST /integrations/{integrationName}/tasks (Queue an integration task) executes a live GitBook API operation.
  update_integration_installation:
    endpoint: PATCH /integrations/{{ record.integration_name }}/installations/{{ record.installation_id }}
    required fields: integration_name, installation_id
    risk: PATCH /integrations/{integrationName}/installations/{installationId} (Update an integration installation) executes a live GitBook API operation.
  uninstall_integration:
    endpoint: DELETE /integrations/{{ record.integration_name }}/installations/{{ record.installation_id }}
    required fields: integration_name, installation_id
    risk: DELETE /integrations/{integrationName}/installations/{installationId} (Uninstall an integration) executes a live GitBook API operation.
  create_integration_installation_token:
    endpoint: POST /integrations/{{ record.integration_name }}/installations/{{ record.installation_id }}/tokens
    required fields: integration_name, installation_id
    risk: POST /integrations/{integrationName}/installations/{installationId}/tokens (Create an integration installation API token) executes a live GitBook API operation.
  install_integration_on_space:
    endpoint: POST /integrations/{{ record.integration_name }}/installations/{{ record.installation_id }}/spaces
    required fields: integration_name, installation_id
    risk: POST /integrations/{integrationName}/installations/{installationId}/spaces (Install an integration on a space) executes a live GitBook API operation.
  update_integration_space_installation:
    endpoint: PATCH /integrations/{{ record.integration_name }}/installations/{{ record.installation_id }}/spaces/{{ record.space_id }}
    required fields: integration_name, installation_id, space_id
    risk: PATCH /integrations/{integrationName}/installations/{installationId}/spaces/{spaceId} (Update an integration space installation) executes a live GitBook API operation.
  uninstall_integration_from_space:
    endpoint: DELETE /integrations/{{ record.integration_name }}/installations/{{ record.installation_id }}/spaces/{{ record.space_id }}
    required fields: integration_name, installation_id, space_id
    risk: DELETE /integrations/{integrationName}/installations/{installationId}/spaces/{spaceId} (Uninstall an integration from a space) executes a live GitBook API operation.
  install_integration_on_site:
    endpoint: POST /integrations/{{ record.integration_name }}/installations/{{ record.installation_id }}/sites
    required fields: integration_name, installation_id
    risk: POST /integrations/{integrationName}/installations/{installationId}/sites (Install an integration on a site) executes a live GitBook API operation.
  update_integration_site_installation:
    endpoint: PATCH /integrations/{{ record.integration_name }}/installations/{{ record.installation_id }}/sites/{{ record.site_id }}
    required fields: integration_name, installation_id, site_id
    risk: PATCH /integrations/{integrationName}/installations/{installationId}/sites/{siteId} (Update an integration site installation) executes a live GitBook API operation.
  uninstall_integration_from_site:
    endpoint: DELETE /integrations/{{ record.integration_name }}/installations/{{ record.installation_id }}/sites/{{ record.site_id }}
    required fields: integration_name, installation_id, site_id
    risk: DELETE /integrations/{integrationName}/installations/{installationId}/sites/{siteId} (Uninstall an integration from a site) executes a live GitBook API operation.
  update_organization_by_id:
    endpoint: PATCH /orgs/{{ record.organization_id }}
    required fields: organization_id
    risk: PATCH /orgs/{organizationId} (Update an organization) executes a live GitBook API operation.
  update_member_in_organization_by_id:
    endpoint: PATCH /orgs/{{ record.organization_id }}/members/{{ record.user_id }}
    required fields: organization_id, user_id
    risk: PATCH /orgs/{organizationId}/members/{userId} (Update an organization member) executes a live GitBook API operation.
  remove_member_from_organization_by_id:
    endpoint: DELETE /orgs/{{ record.organization_id }}/members/{{ record.user_id }}
    required fields: organization_id, user_id
    risk: DELETE /orgs/{organizationId}/members/{userId} (Delete an organization member) executes a live GitBook API operation.
  update_organization_member_last_seen_at:
    endpoint: POST /orgs/{{ record.organization_id }}/ping
    required fields: organization_id
    risk: POST /orgs/{organizationId}/ping (Update an organization member last seen at) executes a live GitBook API operation.
  set_user_as_sso_member_for_organization:
    endpoint: POST /orgs/{{ record.organization_id }}/members/{{ record.user_id }}/sso
    required fields: organization_id, user_id
    risk: POST /orgs/{organizationId}/members/{userId}/sso (Set a user as an SSO member of an organization) executes a live GitBook API operation.
  create_organization_team:
    endpoint: PUT /orgs/{{ record.organization_id }}/teams
    required fields: organization_id
    risk: PUT /orgs/{organizationId}/teams (Create a team) executes a live GitBook API operation.
  update_team_in_organization_by_id:
    endpoint: PATCH /orgs/{{ record.organization_id }}/teams/{{ record.team_id }}
    required fields: organization_id, team_id
    risk: PATCH /orgs/{organizationId}/teams/{teamId} (Update a team) executes a live GitBook API operation.
  remove_team_from_organization_by_id:
    endpoint: DELETE /orgs/{{ record.organization_id }}/teams/{{ record.team_id }}
    required fields: organization_id, team_id
    risk: DELETE /orgs/{organizationId}/teams/{teamId} (Delete a team) executes a live GitBook API operation.
  update_members_in_organization_team:
    endpoint: PUT /orgs/{{ record.organization_id }}/teams/{{ record.team_id }}/members
    required fields: organization_id, team_id
    risk: PUT /orgs/{organizationId}/teams/{teamId}/members (Updates members of a team) executes a live GitBook API operation.
  add_member_to_organization_team_by_id:
    endpoint: PUT /orgs/{{ record.organization_id }}/teams/{{ record.team_id }}/members/{{ record.user_id }}
    required fields: organization_id, team_id, user_id
    risk: PUT /orgs/{organizationId}/teams/{teamId}/members/{userId} (Add a team member) executes a live GitBook API operation.
  delete_member_from_organization_team_by_id:
    endpoint: DELETE /orgs/{{ record.organization_id }}/teams/{{ record.team_id }}/members/{{ record.user_id }}
    required fields: organization_id, team_id, user_id
    risk: DELETE /orgs/{organizationId}/teams/{teamId}/members/{userId} (Delete a team member) executes a live GitBook API operation.
  invite_users_to_organization:
    endpoint: POST /orgs/{{ record.organization_id }}/invites
    required fields: organization_id
    risk: POST /orgs/{organizationId}/invites (Invite users in an organization) executes a live GitBook API operation.
  join_organization_with_invite:
    endpoint: POST /orgs/{{ record.organization_id }}/invites/{{ record.invite_id }}
    required fields: organization_id, invite_id
    risk: POST /orgs/{organizationId}/invites/{inviteId} (Join an organization with an invite) executes a live GitBook API operation.
  create_organization_invite:
    endpoint: POST /orgs/{{ record.organization_id }}/link-invites
    required fields: organization_id
    risk: POST /orgs/{organizationId}/link-invites (Create an organization invite) executes a live GitBook API operation.
  update_organization_invite_by_id:
    endpoint: PATCH /orgs/{{ record.organization_id }}/link-invites/{{ record.invite_id }}
    required fields: organization_id, invite_id
    risk: PATCH /orgs/{organizationId}/link-invites/{inviteId} (Update an organization invite) executes a live GitBook API operation.
  delete_organization_invite_by_id:
    endpoint: DELETE /orgs/{{ record.organization_id }}/link-invites/{{ record.invite_id }}
    required fields: organization_id, invite_id
    risk: DELETE /orgs/{organizationId}/link-invites/{inviteId} (Deletes an organization invite.) executes a live GitBook API operation.
  join_organization:
    endpoint: POST /orgs/{{ record.organization_id }}/join
    required fields: organization_id
    risk: POST /orgs/{organizationId}/join (Join an organization) executes a live GitBook API operation.
  create_space:
    endpoint: POST /orgs/{{ record.organization_id }}/spaces
    required fields: organization_id
    risk: POST /orgs/{organizationId}/spaces (Create a new documentation space in an organization) executes a live GitBook API operation.
  create_collection:
    endpoint: POST /orgs/{{ record.organization_id }}/collections
    required fields: organization_id
    risk: POST /orgs/{organizationId}/collections (Create a collection) executes a live GitBook API operation.
  create_organization_saml_provider:
    endpoint: POST /orgs/{{ record.organization_id }}/saml
    required fields: organization_id
    risk: POST /orgs/{organizationId}/saml (Create a new SAML provider) executes a live GitBook API operation.
  update_organization_saml_provider:
    endpoint: PATCH /orgs/{{ record.organization_id }}/saml/{{ record.saml_provider_id }}
    required fields: organization_id, saml_provider_id
    risk: PATCH /orgs/{organizationId}/saml/{samlProviderId} (Update a SAML provider) executes a live GitBook API operation.
  delete_organization_saml_provider:
    endpoint: DELETE /orgs/{{ record.organization_id }}/saml/{{ record.saml_provider_id }}
    required fields: organization_id, saml_provider_id
    risk: DELETE /orgs/{organizationId}/saml/{samlProviderId} (Delete a SAML provider) executes a live GitBook API operation.
  ask_in_organization:
    endpoint: POST /orgs/{{ record.organization_id }}/ask
    required fields: organization_id
    risk: POST /orgs/{organizationId}/ask (Ask a question in an organization) executes a live GitBook API operation.
  create_open_api_spec:
    endpoint: POST /orgs/{{ record.organization_id }}/openapi
    required fields: organization_id
    risk: POST /orgs/{organizationId}/openapi (Create an OpenAPI spec) executes a live GitBook API operation.
  create_or_update_open_api_spec_by_slug:
    endpoint: PUT /orgs/{{ record.organization_id }}/openapi/{{ record.spec_slug }}
    required fields: organization_id, spec_slug
    risk: PUT /orgs/{organizationId}/openapi/{specSlug} (Create or update an OpenAPI spec) executes a live GitBook API operation.
  update_open_api_spec_by_slug:
    endpoint: PATCH /orgs/{{ record.organization_id }}/openapi/{{ record.spec_slug }}
    required fields: organization_id, spec_slug
    risk: PATCH /orgs/{organizationId}/openapi/{specSlug} (Update OpenAPI spec visibility) executes a live GitBook API operation.
  delete_open_api_spec_by_slug:
    endpoint: DELETE /orgs/{{ record.organization_id }}/openapi/{{ record.spec_slug }}
    required fields: organization_id, spec_slug
    risk: DELETE /orgs/{organizationId}/openapi/{specSlug} (Delete an OpenAPI spec) executes a live GitBook API operation.
  update_organization_agent_instructions:
    endpoint: PUT /orgs/{{ record.organization_id }}/agent-instructions
    required fields: organization_id
    risk: PUT /orgs/{organizationId}/agent-instructions (Update Docs agent instructions for an organization) executes a live GitBook API operation.
  create_translation:
    endpoint: POST /orgs/{{ record.organization_id }}/translations
    required fields: organization_id
    risk: POST /orgs/{organizationId}/translations (Create a translation) executes a live GitBook API operation.
  update_translation:
    endpoint: PUT /orgs/{{ record.organization_id }}/translations/{{ record.translation_id }}
    required fields: organization_id, translation_id
    risk: PUT /orgs/{organizationId}/translations/{translationId} (Update a translation) executes a live GitBook API operation.
  delete_translation:
    endpoint: DELETE /orgs/{{ record.organization_id }}/translations/{{ record.translation_id }}
    required fields: organization_id, translation_id
    risk: DELETE /orgs/{organizationId}/translations/{translationId} (Delete a translation) executes a live GitBook API operation.
  run_translation:
    endpoint: POST /orgs/{{ record.organization_id }}/translations/{{ record.translation_id }}/run
    required fields: organization_id, translation_id
    risk: POST /orgs/{organizationId}/translations/{translationId}/run (Run a translation again) executes a live GitBook API operation.
  update_glossary_entries:
    endpoint: PUT /orgs/{{ record.organization_id }}/translations-glossary
    required fields: organization_id
    risk: PUT /orgs/{organizationId}/translations-glossary (Update glossary entries) executes a live GitBook API operation.
  generate_storage_upload_url:
    endpoint: POST /orgs/{{ record.organization_id }}/storage/upload
    required fields: organization_id
    risk: POST /orgs/{organizationId}/storage/upload (Create a signed URL to upload a file) executes a live GitBook API operation.
  create_custom_font:
    endpoint: PUT /orgs/{{ record.organization_id }}/fonts
    required fields: organization_id
    risk: PUT /orgs/{organizationId}/fonts (Create a custom font) executes a live GitBook API operation.
  update_custom_font:
    endpoint: POST /orgs/{{ record.organization_id }}/fonts/{{ record.font_id }}
    required fields: organization_id, font_id
    risk: POST /orgs/{organizationId}/fonts/{fontId} (Update a custom font) executes a live GitBook API operation.
  delete_custom_font:
    endpoint: DELETE /orgs/{{ record.organization_id }}/fonts/{{ record.font_id }}
    required fields: organization_id, font_id
    risk: DELETE /orgs/{organizationId}/fonts/{fontId} (Delete a custom font) executes a live GitBook API operation.
  start_import_run:
    endpoint: POST /org/{{ record.organization_id }}/imports
    required fields: organization_id
    risk: POST /org/{organizationId}/imports (Import content into a space from a website) executes a live GitBook API operation.
  cancel_import_run:
    endpoint: POST /org/{{ record.organization_id }}/imports/{{ record.import_run_id }}/cancel
    required fields: organization_id, import_run_id
    risk: POST /org/{organizationId}/imports/{importRunId}/cancel (Cancel an import run) executes a live GitBook API operation.
  create_site:
    endpoint: POST /orgs/{{ record.organization_id }}/sites
    required fields: organization_id
    risk: POST /orgs/{organizationId}/sites (Create a new documentation site in an organization) executes a live GitBook API operation.
  update_site_by_id:
    endpoint: PATCH /orgs/{{ record.organization_id }}/sites/{{ record.site_id }}
    required fields: organization_id, site_id
    risk: PATCH /orgs/{organizationId}/sites/{siteId} (Update the properties of a documentation site) executes a live GitBook API operation.
  delete_site_by_id:
    endpoint: DELETE /orgs/{{ record.organization_id }}/sites/{{ record.site_id }}
    required fields: organization_id, site_id
    risk: DELETE /orgs/{organizationId}/sites/{siteId} (Delete a site) executes a live GitBook API operation.
  update_site_adaptive_schema:
    endpoint: PUT /orgs/{{ record.organization_id }}/sites/{{ record.site_id }}/adaptive-schema
    required fields: organization_id, site_id
    risk: PUT /orgs/{organizationId}/sites/{siteId}/adaptive-schema (Update the visitor attributes JSON schema for an adaptive content site) executes a live GitBook API operation.
  publish_site:
    endpoint: POST /orgs/{{ record.organization_id }}/sites/{{ record.site_id }}/publish
    required fields: organization_id, site_id
    risk: POST /orgs/{organizationId}/sites/{siteId}/publish (Publish a site to make it publicly accessible) executes a live GitBook API operation.
  unpublish_site:
    endpoint: POST /orgs/{{ record.organization_id }}/sites/{{ record.site_id }}/unpublish
    required fields: organization_id, site_id
    risk: POST /orgs/{organizationId}/sites/{siteId}/unpublish (Take a site offline by unpublishing it) executes a live GitBook API operation.
  create_site_share_link:
    endpoint: POST /orgs/{{ record.organization_id }}/sites/{{ record.site_id }}/share-links
    required fields: organization_id, site_id
    risk: POST /orgs/{organizationId}/sites/{siteId}/share-links (Create a private share link for a site) executes a live GitBook API operation.
  update_site_share_link_by_id:
    endpoint: PATCH /orgs/{{ record.organization_id }}/sites/{{ record.site_id }}/share-links/{{ record.share_link_id }}
    required fields: organization_id, site_id, share_link_id
    risk: PATCH /orgs/{organizationId}/sites/{siteId}/share-links/{shareLinkId} (Update a private share link for a site) executes a live GitBook API operation.
  delete_site_share_link_by_id:
    endpoint: DELETE /orgs/{{ record.organization_id }}/sites/{{ record.site_id }}/share-links/{{ record.share_link_id }}
    required fields: organization_id, site_id, share_link_id
    risk: DELETE /orgs/{organizationId}/sites/{siteId}/share-links/{shareLinkId} (Deletes a share link) executes a live GitBook API operation.
  sort_site_structure:
    endpoint: PATCH /orgs/{{ record.organization_id }}/sites/{{ record.site_id }}/structure/sort
    required fields: organization_id, site_id
    risk: PATCH /orgs/{organizationId}/sites/{siteId}/structure/sort (Move a site space, section, or section group to a new position) executes a live GitBook API operation.
  update_site_publishing_auth_by_id:
    endpoint: PATCH /orgs/{{ record.organization_id }}/sites/{{ record.site_id }}/publishing/auth
    required fields: organization_id, site_id
    risk: PATCH /orgs/{organizationId}/sites/{siteId}/publishing/auth (Update the published content authentication configuration for a site) executes a live GitBook API operation.
  regenerate_site_publishing_auth_by_id:
    endpoint: POST /orgs/{{ record.organization_id }}/sites/{{ record.site_id }}/publishing/auth/regenerate
    required fields: organization_id, site_id
    risk: POST /orgs/{organizationId}/sites/{siteId}/publishing/auth/regenerate (Regenerate the private key for a site's published content authentication) executes a live GitBook API operation.
  update_site_customization_by_id:
    endpoint: PUT /orgs/{{ record.organization_id }}/sites/{{ record.site_id }}/customization
    required fields: organization_id, site_id
    risk: PUT /orgs/{organizationId}/sites/{siteId}/customization (Update the branding and visual customization settings for a site) executes a live GitBook API operation.
  add_space_to_site:
    endpoint: POST /orgs/{{ record.organization_id }}/sites/{{ record.site_id }}/site-spaces
    required fields: organization_id, site_id
    risk: POST /orgs/{organizationId}/sites/{siteId}/site-spaces (Add a space to a site as a content source) executes a live GitBook API operation.
  add_section_group_to_site:
    endpoint: POST /orgs/{{ record.organization_id }}/sites/{{ record.site_id }}/section-groups
    required fields: organization_id, site_id
    risk: POST /orgs/{organizationId}/sites/{siteId}/section-groups (Add a section group to a site's navigation structure) executes a live GitBook API operation.
  update_site_section_group_by_id:
    endpoint: PATCH /orgs/{{ record.organization_id }}/sites/{{ record.site_id }}/section-groups/{{ record.site_section_group_id }}
    required fields: organization_id, site_id, site_section_group_id
    risk: PATCH /orgs/{organizationId}/sites/{siteId}/section-groups/{siteSectionGroupId} (Update a section group in a site's navigation structure) executes a live GitBook API operation.
  delete_site_section_group_by_id:
    endpoint: DELETE /orgs/{{ record.organization_id }}/sites/{{ record.site_id }}/section-groups/{{ record.site_section_group_id }}
    required fields: organization_id, site_id, site_section_group_id
    risk: DELETE /orgs/{organizationId}/sites/{siteId}/section-groups/{siteSectionGroupId} (Delete a site section group) executes a live GitBook API operation.
  add_section_to_site:
    endpoint: POST /orgs/{{ record.organization_id }}/sites/{{ record.site_id }}/sections
    required fields: organization_id, site_id
    risk: POST /orgs/{organizationId}/sites/{siteId}/sections (Add a new navigation section to a site backed by a space) executes a live GitBook API operation.
  update_site_section_by_id:
    endpoint: PATCH /orgs/{{ record.organization_id }}/sites/{{ record.site_id }}/sections/{{ record.site_section_id }}
    required fields: organization_id, site_id, site_section_id
    risk: PATCH /orgs/{organizationId}/sites/{siteId}/sections/{siteSectionId} (Update a navigation section in a site) executes a live GitBook API operation.
  delete_site_section_by_id:
    endpoint: DELETE /orgs/{{ record.organization_id }}/sites/{{ record.site_id }}/sections/{{ record.site_section_id }}
    required fields: organization_id, site_id, site_section_id
    risk: DELETE /orgs/{organizationId}/sites/{siteId}/sections/{siteSectionId} (Delete a site section) executes a live GitBook API operation.
  search_site_content:
    endpoint: POST /orgs/{{ record.organization_id }}/sites/{{ record.site_id }}/search
    required fields: organization_id, site_id
    risk: POST /orgs/{organizationId}/sites/{siteId}/search (Full-text search across all content in a site) executes a live GitBook API operation.
  stream_ask_in_site:
    endpoint: POST /orgs/{{ record.organization_id }}/sites/{{ record.site_id }}/ask
    required fields: organization_id, site_id
    risk: POST /orgs/{organizationId}/sites/{siteId}/ask (Ask a question in a site) executes a live GitBook API operation.
  create_site_scan:
    endpoint: POST /orgs/{{ record.organization_id }}/sites/{{ record.site_id }}/scans
    required fields: organization_id, site_id
    risk: POST /orgs/{organizationId}/sites/{siteId}/scans (Enqueue a new site scan) executes a live GitBook API operation.
  update_site_finding_by_id:
    endpoint: PATCH /orgs/{{ record.organization_id }}/sites/{{ record.site_id }}/findings/{{ record.site_finding_id }}
    required fields: organization_id, site_id, site_finding_id
    risk: PATCH /orgs/{organizationId}/sites/{siteId}/findings/{siteFindingId} (Update a site finding) executes a live GitBook API operation.
  trigger_change_requests_for_site_finding:
    endpoint: POST /orgs/{{ record.organization_id }}/sites/{{ record.site_id }}/findings/{{ record.site_finding_id }}/change-requests
    required fields: organization_id, site_id, site_finding_id
    risk: POST /orgs/{organizationId}/sites/{siteId}/findings/{siteFindingId}/change-requests (Process a site finding into change requests) executes a live GitBook API operation.
  create_site_context_connection:
    endpoint: POST /orgs/{{ record.organization_id }}/sites/{{ record.site_id }}/context-connections
    required fields: organization_id, site_id
    risk: POST /orgs/{organizationId}/sites/{siteId}/context-connections (Create a context connection) executes a live GitBook API operation.
  update_site_context_connection_by_id:
    endpoint: PATCH /orgs/{{ record.organization_id }}/sites/{{ record.site_id }}/context-connections/{{ record.site_context_connection_id }}
    required fields: organization_id, site_id, site_context_connection_id
    risk: PATCH /orgs/{organizationId}/sites/{siteId}/context-connections/{siteContextConnectionId} (Update a context connection) executes a live GitBook API operation.
  delete_site_context_connection_by_id:
    endpoint: DELETE /orgs/{{ record.organization_id }}/sites/{{ record.site_id }}/context-connections/{{ record.site_context_connection_id }}
    required fields: organization_id, site_id, site_context_connection_id
    risk: DELETE /orgs/{organizationId}/sites/{siteId}/context-connections/{siteContextConnectionId} (Delete a context connection) executes a live GitBook API operation.
  sync_site_context_connection:
    endpoint: POST /orgs/{{ record.organization_id }}/sites/{{ record.site_id }}/context-connections/{{ record.site_context_connection_id }}/sync
    required fields: organization_id, site_id, site_context_connection_id
    risk: POST /orgs/{organizationId}/sites/{siteId}/context-connections/{siteContextConnectionId}/sync (Trigger a sync for a context connection) executes a live GitBook API operation.
  update_site_topic_by_id:
    endpoint: PATCH /orgs/{{ record.organization_id }}/sites/{{ record.site_id }}/topics/{{ record.site_topic_id }}
    required fields: organization_id, site_id, site_topic_id
    risk: PATCH /orgs/{organizationId}/sites/{siteId}/topics/{siteTopicId} (Update a topic) executes a live GitBook API operation.
  delete_site_topic_findings:
    endpoint: DELETE /orgs/{{ record.organization_id }}/sites/{{ record.site_id }}/topics/{{ record.site_topic_id }}/findings
    required fields: organization_id, site_id, site_topic_id
    risk: DELETE /orgs/{organizationId}/sites/{siteId}/topics/{siteTopicId}/findings (Delete all findings for a topic) executes a live GitBook API operation.
  update_site_space_by_id:
    endpoint: PATCH /orgs/{{ record.organization_id }}/sites/{{ record.site_id }}/site-spaces/{{ record.site_space_id }}
    required fields: organization_id, site_id, site_space_id
    risk: PATCH /orgs/{organizationId}/sites/{siteId}/site-spaces/{siteSpaceId} (Update a space linked to a site) executes a live GitBook API operation.
  delete_site_space_by_id:
    endpoint: DELETE /orgs/{{ record.organization_id }}/sites/{{ record.site_id }}/site-spaces/{{ record.site_space_id }}
    required fields: organization_id, site_id, site_space_id
    risk: DELETE /orgs/{organizationId}/sites/{siteId}/site-spaces/{siteSpaceId} (Delete a site space) executes a live GitBook API operation.
  override_site_space_customization_by_id:
    endpoint: PATCH /orgs/{{ record.organization_id }}/sites/{{ record.site_id }}/site-spaces/{{ record.site_space_id }}/customization
    required fields: organization_id, site_id, site_space_id
    risk: PATCH /orgs/{organizationId}/sites/{siteId}/site-spaces/{siteSpaceId}/customization (Override branding and customization settings for a specific site space) executes a live GitBook API operation.
  delete_site_space_customization_by_id:
    endpoint: DELETE /orgs/{{ record.organization_id }}/sites/{{ record.site_id }}/site-spaces/{{ record.site_space_id }}/customization
    required fields: organization_id, site_id, site_space_id
    risk: DELETE /orgs/{organizationId}/sites/{siteId}/site-spaces/{siteSpaceId}/customization (Delete a site space customization settings) executes a live GitBook API operation.
  move_site_section_group:
    endpoint: POST /orgs/{{ record.organization_id }}/sites/{{ record.site_id }}/section-groups/{{ record.site_section_group_id }}/move
    required fields: organization_id, site_id, site_section_group_id
    risk: POST /orgs/{organizationId}/sites/{siteId}/section-groups/{siteSectionGroupId}/move (Move a site section group to a new position. (Deprecated) use sortSiteStructure instead.) executes a live GitBook API operation.
  move_site_section:
    endpoint: POST /orgs/{{ record.organization_id }}/sites/{{ record.site_id }}/sections/{{ record.site_section_id }}/move
    required fields: organization_id, site_id, site_section_id
    risk: POST /orgs/{organizationId}/sites/{siteId}/sections/{siteSectionId}/move (Move a site section to a new position. (Deprecated) use sortSiteStructure instead.) executes a live GitBook API operation.
  move_site_space:
    endpoint: POST /orgs/{{ record.organization_id }}/sites/{{ record.site_id }}/site-spaces/{{ record.site_space_id }}/move
    required fields: organization_id, site_id, site_space_id
    risk: POST /orgs/{organizationId}/sites/{siteId}/site-spaces/{siteSpaceId}/move (Move a site space to a new position. (Deprecated) use sortSiteStructure instead.) executes a live GitBook API operation.
  invite_to_site:
    endpoint: POST /orgs/{{ record.organization_id }}/sites/{{ record.site_id }}/permissions
    required fields: organization_id, site_id
    risk: POST /orgs/{organizationId}/sites/{siteId}/permissions (Invite a user or a team to a site) executes a live GitBook API operation.
  update_user_permission_in_site:
    endpoint: PATCH /orgs/{{ record.organization_id }}/sites/{{ record.site_id }}/permissions/users/{{ record.user_id }}
    required fields: organization_id, site_id, user_id
    risk: PATCH /orgs/{organizationId}/sites/{siteId}/permissions/users/{userId} (Update site user permissions) executes a live GitBook API operation.
  remove_user_from_site:
    endpoint: DELETE /orgs/{{ record.organization_id }}/sites/{{ record.site_id }}/permissions/users/{{ record.user_id }}
    required fields: organization_id, site_id, user_id
    risk: DELETE /orgs/{organizationId}/sites/{siteId}/permissions/users/{userId} (Remove a site user) executes a live GitBook API operation.
  update_team_permission_in_site:
    endpoint: PATCH /orgs/{{ record.organization_id }}/sites/{{ record.site_id }}/permissions/teams/{{ record.team_id }}
    required fields: organization_id, site_id, team_id
    risk: PATCH /orgs/{organizationId}/sites/{siteId}/permissions/teams/{teamId} (Update an org team's permission in a site) executes a live GitBook API operation.
  remove_team_from_site:
    endpoint: DELETE /orgs/{{ record.organization_id }}/sites/{{ record.site_id }}/permissions/teams/{{ record.team_id }}
    required fields: organization_id, site_id, team_id
    risk: DELETE /orgs/{organizationId}/sites/{siteId}/permissions/teams/{teamId} (Remove an org team from a site) executes a live GitBook API operation.
  stream_ai_response_in_site:
    endpoint: POST /orgs/{{ record.organization_id }}/sites/{{ record.site_id }}/ai/response
    required fields: organization_id, site_id
    risk: POST /orgs/{organizationId}/sites/{siteId}/ai/response (Generate an AI response in a site) executes a live GitBook API operation.
  update_site_agent_settings_by_id:
    endpoint: PUT /orgs/{{ record.organization_id }}/sites/{{ record.site_id }}/agent-settings
    required fields: organization_id, site_id
    risk: PUT /orgs/{organizationId}/sites/{siteId}/agent-settings (Update the AI agent configuration for a site) executes a live GitBook API operation.
  create_site_styleguide_by_id:
    endpoint: POST /orgs/{{ record.organization_id }}/sites/{{ record.site_id }}/styleguide
    required fields: organization_id, site_id
    risk: POST /orgs/{organizationId}/sites/{siteId}/styleguide (Create or retrieve the styleguide space for a site) executes a live GitBook API operation.
  track_events_in_site_by_id:
    endpoint: POST /orgs/{{ record.organization_id }}/sites/{{ record.site_id }}/insights/events
    required fields: organization_id, site_id
    risk: POST /orgs/{organizationId}/sites/{siteId}/insights/events (Track site events) executes a live GitBook API operation.
  aggregate_site_events:
    endpoint: POST /orgs/{{ record.organization_id }}/sites/{{ record.site_id }}/insights/events/aggregate
    required fields: organization_id, site_id
    risk: POST /orgs/{organizationId}/sites/{siteId}/insights/events/aggregate (Query site events) executes a live GitBook API operation.
  update_site_ads_by_id:
    endpoint: POST /orgs/{{ record.organization_id }}/sites/{{ record.site_id }}/ads
    required fields: organization_id, site_id
    risk: POST /orgs/{organizationId}/sites/{siteId}/ads (Update the advertising settings for a site) executes a live GitBook API operation.
  create_site_redirect:
    endpoint: POST /orgs/{{ record.organization_id }}/sites/{{ record.site_id }}/redirects
    required fields: organization_id, site_id
    risk: POST /orgs/{organizationId}/sites/{siteId}/redirects (Create a URL redirect rule for a site) executes a live GitBook API operation.
  bulk_upsert_site_redirects:
    endpoint: PUT /orgs/{{ record.organization_id }}/sites/{{ record.site_id }}/redirects
    required fields: organization_id, site_id
    risk: PUT /orgs/{organizationId}/sites/{siteId}/redirects (Create, update, delete, or publish site redirect rules in bulk) executes a live GitBook API operation.
  update_site_redirect_by_id:
    endpoint: PATCH /orgs/{{ record.organization_id }}/sites/{{ record.site_id }}/redirects/{{ record.site_redirect_id }}
    required fields: organization_id, site_id, site_redirect_id
    risk: PATCH /orgs/{organizationId}/sites/{siteId}/redirects/{siteRedirectId} (Update a URL redirect rule for a site) executes a live GitBook API operation.
  delete_site_redirect_by_id:
    endpoint: DELETE /orgs/{{ record.organization_id }}/sites/{{ record.site_id }}/redirects/{{ record.site_redirect_id }}
    required fields: organization_id, site_id, site_redirect_id
    risk: DELETE /orgs/{organizationId}/sites/{siteId}/redirects/{siteRedirectId} (Delete a site redirect) executes a live GitBook API operation.
  create_site_mcp_server:
    endpoint: POST /orgs/{{ record.organization_id }}/sites/{{ record.site_id }}/mcp-servers
    required fields: organization_id, site_id
    risk: POST /orgs/{organizationId}/sites/{siteId}/mcp-servers (Add a new MCP server configuration to a site) executes a live GitBook API operation.
  update_site_mcp_server_by_id:
    endpoint: PATCH /orgs/{{ record.organization_id }}/sites/{{ record.site_id }}/mcp-servers/{{ record.site_mcp_server_id }}
    required fields: organization_id, site_id, site_mcp_server_id
    risk: PATCH /orgs/{organizationId}/sites/{siteId}/mcp-servers/{siteMcpServerId} (Update an MCP server configuration for a site) executes a live GitBook API operation.
  delete_site_mcp_server_by_id:
    endpoint: DELETE /orgs/{{ record.organization_id }}/sites/{{ record.site_id }}/mcp-servers/{{ record.site_mcp_server_id }}
    required fields: organization_id, site_id, site_mcp_server_id
    risk: DELETE /orgs/{organizationId}/sites/{siteId}/mcp-servers/{siteMcpServerId} (Delete a site MCP server) executes a live GitBook API operation.
  create_site_channel:
    endpoint: POST /orgs/{{ record.organization_id }}/sites/{{ record.site_id }}/channels
    required fields: organization_id, site_id
    risk: POST /orgs/{organizationId}/sites/{siteId}/channels (Create a new GitBook Agent channel for a site) executes a live GitBook API operation.
  update_site_channel_by_id:
    endpoint: PATCH /orgs/{{ record.organization_id }}/sites/{{ record.site_id }}/channels/{{ record.site_channel_id }}
    required fields: organization_id, site_id, site_channel_id
    risk: PATCH /orgs/{organizationId}/sites/{siteId}/channels/{siteChannelId} (Update a GitBook Agent channel for a site) executes a live GitBook API operation.
  delete_site_channel_by_id:
    endpoint: DELETE /orgs/{{ record.organization_id }}/sites/{{ record.site_id }}/channels/{{ record.site_channel_id }}
    required fields: organization_id, site_id, site_channel_id
    risk: DELETE /orgs/{organizationId}/sites/{siteId}/channels/{siteChannelId} (Delete a GitBook Agent channel from a site) executes a live GitBook API operation.
  dns_revalidate_custom_hostname:
    endpoint: PATCH /custom-hostnames/{{ record.hostname }}
    required fields: hostname
    risk: PATCH /custom-hostnames/{hostname} (Revalidate a custom hostname DNS) executes a live GitBook API operation.
  remove_custom_hostname:
    endpoint: DELETE /custom-hostnames/{{ record.hostname }}
    required fields: hostname
    risk: DELETE /custom-hostnames/{hostname} (Remove a custom hostname) executes a live GitBook API operation.
  ads_update_site:
    endpoint: PATCH /ads/sites/{{ record.site_id }}
    required fields: site_id
    risk: PATCH /ads/sites/{siteId} (Update the Ads configuration for a site) executes a live GitBook API operation.
  resolve_published_content_by_url:
    endpoint: POST /urls/published
    risk: POST /urls/published (Resolve a URL of a published content.) executes a live GitBook API operation.
  install_git_sync_provider_on_target:
    endpoint: POST /git/installations
    risk: POST /git/installations (Install a Git Sync provider on a target) executes a live GitBook API operation.
  update_git_sync_installation_by_id:
    endpoint: PATCH /git/installations/{{ record.installation_id }}
    required fields: installation_id
    risk: PATCH /git/installations/{installationId} (Update a Git Sync installation configuration) executes a live GitBook API operation.
  uninstall_git_sync_installation:
    endpoint: DELETE /git/installations/{{ record.installation_id }}
    required fields: installation_id
    risk: DELETE /git/installations/{installationId} (Uninstall a Git Sync installation) executes a live GitBook API operation.

SECURITY
  read risk: external GitBook API reads across users, organizations, spaces, sites, content, permissions, integrations, analytics, search, and related resources
  write risk: creates, updates, publishes, archives, deletes, imports, exports, invites, permission changes, and content changes in GitBook depending on the selected write action
  approval: reverse ETL writes require plan preview and approval token before execution
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect gitbook

  # Inspect as structured JSON
  pm connectors inspect gitbook --json

AGENT WORKFLOW
  - Run pm connectors inspect gitbook before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
