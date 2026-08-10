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
  id: simple-icons-gitbook
  asset: icons/simple-icons/gitbook.svg
  title: GitBook
  simple_icon_slug: gitbook
  simple_icon_hex: BBDDE5
  source: simple-icons
  license: CC0-1.0
  review_status: cc0_with_trademark_caveat
  review_url: https://simpleicons.org/?q=GitBook
  match: exact-name-or-slug
  matched_by: gitbook

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
    fields: display_name(string), email(string), id(string), photo_url(string)
  organizations:
    primary key: id
    fields: created_at(string), id(string), title(string), type(string), url(object)
  org_members:
    primary key: id
    fields: display_name(string), email(string), id(string), role(string)
  content:
    primary key: id
    fields: id(string), kind(string), path(string), slug(string), title(string), type(string)
  get_api_information:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  get_user_by_id:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  get_space_by_id:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  get_embed_by_url_in_space:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  search_space_content:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  get_space_git_info:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  list_user_permissions_in_space:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  list_team_permissions_in_space:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  get_current_revision:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  list_files:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  get_file_by_id:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  list_space_file_backlinks:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  get_page_by_id:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  list_page_links_in_space:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  list_space_page_backlinks:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  list_space_page_meta_links:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  get_page_by_path:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  get_reusable_content_by_id:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  get_document_by_id:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  list_change_requests_for_space:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  get_change_request_by_id:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  get_reviews_by_change_request_id:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  get_change_request_review_by_id:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  get_requested_reviewers_by_change_request_id:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  list_change_request_conversations:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  list_change_request_links:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  list_comments_in_change_request:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  get_comment_in_change_request:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  list_comment_replies_in_change_request:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  get_comment_reply_in_change_request:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  get_contributors_by_change_request_id:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  get_revision_of_change_request_by_id:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  list_pages_in_change_request:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  list_files_in_change_request_by_id:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  get_file_in_change_request_by_id:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  list_change_request_file_backlinks:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  get_page_in_change_request_by_id:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  list_page_links_in_change_request:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  list_change_request_page_backlinks:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  list_change_request_page_meta_links:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  get_reusable_content_in_change_request_by_id:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  get_change_request_changes:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  get_change_request_pdf:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  get_revision_by_id:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  get_revision_semantic_changes:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  list_pages_in_revision_by_id:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  list_files_in_revision_by_id:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  get_file_in_revision_by_id:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  get_page_in_revision_by_id:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  get_page_document_in_revision_by_id:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  get_page_in_revision_by_path:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  list_revision_page_meta_links:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  get_page_in_change_request_by_path:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  get_reusable_content_in_revision_by_id:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  get_reusable_content_document_in_revision_by_id:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  list_comments_in_space:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  get_comment_in_space:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  list_comment_replies_in_space:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  get_comment_reply_in_space:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  list_commenters_in_space:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  list_commenters_in_change_request:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  list_permissions_aggregate_in_space:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  list_space_integrations:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  list_space_integrations_blocks:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  get_space_pdf:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  list_space_links:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  get_collection_by_id:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  list_spaces_in_collection_by_id:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  list_team_permissions_in_collection:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  list_user_permissions_in_collection:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  list_permissions_aggregate_in_collection:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  list_integrations:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  get_integration_by_name:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  list_integration_installations:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  list_integration_events:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  get_integration_event:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  list_integration_space_installations:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  list_integration_site_installations:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  render_integration_ui_with_get:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  get_integration_installation_by_id:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  list_integration_installation_spaces:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  get_integration_space_installation:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  list_integration_installation_sites:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  get_integration_site_installation:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  get_organization_by_id:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  get_member_in_organization_by_id:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  list_spaces_for_organization_member:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  list_teams_for_organization_member:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  list_teams_in_organization_by_id:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  get_team_in_organization_by_id:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  list_team_members_in_organization_by_id:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  list_organization_invite_links:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  get_organization_invite_link:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  search_organization_content:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  list_change_requests_for_organization:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  list_spaces_in_organization_by_id:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  list_collections_in_organization_by_id:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  list_organization_integrations:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  get_organization_integration_status:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  list_organization_installations:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  list_organization_integrations_status:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  list_saml_providers_in_organization_by_id:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  get_organization_saml_provider_by_id:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  list_sso_provider_logins_in_organization:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  get_recommended_questions_in_organization:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  list_open_api_specs:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  get_open_api_spec_by_slug:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  list_open_api_spec_versions:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  get_latest_open_api_spec_version:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  get_latest_open_api_spec_version_content:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  get_open_api_spec_version_by_id:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  get_open_api_spec_version_content_by_id:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  get_organization_agent_instructions:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  list_translations:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  get_translation:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  list_glossary_entries:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  get_glossary_entry:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  list_custom_fonts:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  get_custom_font:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  list_sites:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  get_site_by_id:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  list_site_git_sync_installations:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  get_site_adaptive_schema:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  list_site_adaptive_template_conditions:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  get_published_content_site:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  list_site_share_links:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  get_site_structure:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  get_site_publishing_auth_by_id:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  get_site_publishing_preview_by_id:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  get_site_customization_by_id:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  list_site_integration_scripts:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  list_site_integrations:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  list_site_spaces:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  list_site_section_groups:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  list_site_sections:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  list_site_context_records:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  get_site_context_record_by_id:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  list_site_scans:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  get_site_scan_by_id:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  list_site_findings:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  get_site_finding_by_id:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  list_change_requests_for_site_finding:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  list_pages_for_site_finding:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  list_questions_for_site_finding:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  list_records_for_site_finding:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  list_site_context_connections:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  get_site_context_connection_by_id:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  list_site_topics:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  get_site_topic_by_id:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  list_site_questions:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  get_site_question_by_id:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  list_site_question_sources:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  get_site_question_stats:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  list_site_question_answers:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  get_site_question_answer_by_id:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  get_site_question_answer_thread_by_id:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  list_site_question_answer_sources:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  get_site_space_customization_by_id:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  list_permissions_aggregate_in_site:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  list_user_permissions_in_site:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  list_team_permissions_in_site:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  get_site_agent_settings_by_id:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  list_site_visitor_segments:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  list_site_redirects:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  get_site_redirect_by_source:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  list_site_mcp_servers:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  get_site_mcp_server_by_id:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  list_site_channels:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  get_site_channel_by_id:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  get_subdomain:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  get_custom_hostname:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  get_organizations_for_email_domain:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  ads_list_sites:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  get_content_by_url:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  get_embed_by_url:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  get_published_content_by_url:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  get_git_sync_installation_by_id:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  list_git_hub_repositories_for_git_sync_installation:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  list_git_hub_repo_branches_for_git_sync_installation:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  list_git_lab_projects_for_git_sync_installation:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)
  list_git_lab_project_branches_for_git_sync_installation:
    primary key: id
    fields: id(string), name(string), object(string), operation_id(string), status(string), title(string)

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
    required fields: space_id, url, ref
    risk: POST /spaces/{spaceId}/git/import (Pull content into a space from a connected Git repository) executes a live GitBook API operation.
  export_to_git_repository:
    endpoint: POST /spaces/{{ record.space_id }}/git/export
    required fields: space_id, url, ref, commit_message
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
    required fields: space_id, id
    risk: POST /spaces/{spaceId}/content/template (Apply a content template to populate a space with initial pages) executes a live GitBook API operation.
  get_computed_document:
    endpoint: POST /spaces/{{ record.space_id }}/content/computed/document
    required fields: space_id, source, seed
    risk: POST /spaces/{spaceId}/content/computed/document (Compute and render a document from a structured content source) executes a live GitBook API operation.
  get_computed_revision:
    endpoint: POST /spaces/{{ record.space_id }}/content/computed/revision
    required fields: space_id, source, seed
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
    required fields: space_id, change_request_id, status
    risk: POST /spaces/{spaceId}/change-requests/{changeRequestId}/reviews (Submit an approve or request-changes review for a change request) executes a live GitBook API operation.
  request_reviewers_for_change_request:
    endpoint: POST /spaces/{{ record.space_id }}/change-requests/{{ record.change_request_id }}/requested-reviewers
    required fields: space_id, change_request_id, users
    risk: POST /spaces/{spaceId}/change-requests/{changeRequestId}/requested-reviewers (Send review requests to users for a change request) executes a live GitBook API operation.
  remove_requested_reviewer_from_change_request:
    endpoint: DELETE /spaces/{{ record.space_id }}/change-requests/{{ record.change_request_id }}/requested-reviewers/{{ record.user_id }}
    required fields: space_id, change_request_id, user_id
    risk: DELETE /spaces/{spaceId}/change-requests/{changeRequestId}/requested-reviewers/{userId} (Remove a reviewer from a change request) executes a live GitBook API operation.
  update_change_request_conversation:
    endpoint: PATCH /spaces/{{ record.space_id }}/change-requests/{{ record.change_request_id }}/conversations/{{ record.conversation_id }}
    required fields: space_id, change_request_id, conversation_id, title
    risk: PATCH /spaces/{spaceId}/change-requests/{changeRequestId}/conversations/{conversationId} (Update the title of an AI agent conversation on a change request) executes a live GitBook API operation.
  delete_change_request_conversation:
    endpoint: DELETE /spaces/{{ record.space_id }}/change-requests/{{ record.change_request_id }}/conversations/{{ record.conversation_id }}
    required fields: space_id, change_request_id, conversation_id
    risk: DELETE /spaces/{spaceId}/change-requests/{changeRequestId}/conversations/{conversationId} (Delete an agent conversation) executes a live GitBook API operation.
  post_comment_in_change_request:
    endpoint: POST /spaces/{{ record.space_id }}/change-requests/{{ record.change_request_id }}/comments
    required fields: space_id, change_request_id, body
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
    required fields: space_id, change_request_id, comment_id, body
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
    required fields: space_id, change_request_id, changes
    risk: POST /spaces/{spaceId}/change-requests/{changeRequestId}/content (Apply a batch of content changes to a change request) executes a live GitBook API operation.
  post_comment_in_space:
    endpoint: POST /spaces/{{ record.space_id }}/comments
    required fields: space_id, body
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
    required fields: space_id, comment_id, body
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
    required fields: collection_id, organization
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
    required fields: integration_name, organization, title, description, script, scopes
    risk: POST /integrations/{integrationName} (Publish an integration) executes a live GitBook API operation.
  unpublish_integration:
    endpoint: DELETE /integrations/{{ record.integration_name }}
    required fields: integration_name
    risk: DELETE /integrations/{integrationName} (Unpublish an integration) executes a live GitBook API operation.
  install_integration:
    endpoint: POST /integrations/{{ record.integration_name }}/installations
    required fields: integration_name, organization
    risk: POST /integrations/{integrationName}/installations (Install an integration) executes a live GitBook API operation.
  set_integration_development_mode:
    endpoint: PUT /integrations/{{ record.integration_name }}/dev
    required fields: integration_name, tunnel_url
    risk: PUT /integrations/{integrationName}/dev (Enable integration dev mode) executes a live GitBook API operation.
  disable_integration_development_mode:
    endpoint: DELETE /integrations/{{ record.integration_name }}/dev
    required fields: integration_name
    risk: DELETE /integrations/{integrationName}/dev (Disable integration dev mode) executes a live GitBook API operation.
  render_integration_ui_with_post:
    endpoint: POST /integrations/{{ record.integration_name }}/render
    required fields: integration_name, component_id, props, context
    risk: POST /integrations/{integrationName}/render (Render an integration UI with POST method) executes a live GitBook API operation.
  queue_integration_task:
    endpoint: POST /integrations/{{ record.integration_name }}/tasks
    required fields: integration_name, task
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
    required fields: integration_name, installation_id, space
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
    required fields: integration_name, installation_id, site_id
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
    required fields: organization_id, title
    risk: PUT /orgs/{organizationId}/teams (Create a team) executes a live GitBook API operation.
  update_team_in_organization_by_id:
    endpoint: PATCH /orgs/{{ record.organization_id }}/teams/{{ record.team_id }}
    required fields: organization_id, team_id, title
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
    required fields: organization_id, emails
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
    required fields: organization_id, label
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
    required fields: organization_id, query
    risk: POST /orgs/{organizationId}/ask (Ask a question in an organization) executes a live GitBook API operation.
  create_open_api_spec:
    endpoint: POST /orgs/{{ record.organization_id }}/openapi
    required fields: organization_id, source, slug
    risk: POST /orgs/{organizationId}/openapi (Create an OpenAPI spec) executes a live GitBook API operation.
  create_or_update_open_api_spec_by_slug:
    endpoint: PUT /orgs/{{ record.organization_id }}/openapi/{{ record.spec_slug }}
    required fields: organization_id, spec_slug, source
    risk: PUT /orgs/{organizationId}/openapi/{specSlug} (Create or update an OpenAPI spec) executes a live GitBook API operation.
  update_open_api_spec_by_slug:
    endpoint: PATCH /orgs/{{ record.organization_id }}/openapi/{{ record.spec_slug }}
    required fields: organization_id, spec_slug, visibility
    risk: PATCH /orgs/{organizationId}/openapi/{specSlug} (Update OpenAPI spec visibility) executes a live GitBook API operation.
  delete_open_api_spec_by_slug:
    endpoint: DELETE /orgs/{{ record.organization_id }}/openapi/{{ record.spec_slug }}
    required fields: organization_id, spec_slug
    risk: DELETE /orgs/{organizationId}/openapi/{specSlug} (Delete an OpenAPI spec) executes a live GitBook API operation.
  update_organization_agent_instructions:
    endpoint: PUT /orgs/{{ record.organization_id }}/agent-instructions
    required fields: organization_id, instructions
    risk: PUT /orgs/{organizationId}/agent-instructions (Update Docs agent instructions for an organization) executes a live GitBook API operation.
  create_translation:
    endpoint: POST /orgs/{{ record.organization_id }}/translations
    required fields: organization_id, language, source
    risk: POST /orgs/{organizationId}/translations (Create a translation) executes a live GitBook API operation.
  update_translation:
    endpoint: PUT /orgs/{{ record.organization_id }}/translations/{{ record.translation_id }}
    required fields: organization_id, translation_id, instructions
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
    required fields: organization_id, operations
    risk: PUT /orgs/{organizationId}/translations-glossary (Update glossary entries) executes a live GitBook API operation.
  generate_storage_upload_url:
    endpoint: POST /orgs/{{ record.organization_id }}/storage/upload
    required fields: organization_id, file, kind
    risk: POST /orgs/{organizationId}/storage/upload (Create a signed URL to upload a file) executes a live GitBook API operation.
  create_custom_font:
    endpoint: PUT /orgs/{{ record.organization_id }}/fonts
    required fields: organization_id, font_family, font_faces
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
    required fields: organization_id, source, target
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
    required fields: organization_id, site_id, json_schema
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
    required fields: organization_id, site_id, name
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
    required fields: organization_id, site_id, item, position
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
    required fields: organization_id, site_id, styling, internationalization, favicon, header, footer, themes, feedback, ai, advanced_customization, trademark, external_links, pagination, page_actions, privacy_policy, social_preview, social_accounts, insights
    risk: PUT /orgs/{organizationId}/sites/{siteId}/customization (Update the branding and visual customization settings for a site) executes a live GitBook API operation.
  add_space_to_site:
    endpoint: POST /orgs/{{ record.organization_id }}/sites/{{ record.site_id }}/site-spaces
    required fields: organization_id, site_id, space_id
    risk: POST /orgs/{organizationId}/sites/{siteId}/site-spaces (Add a space to a site as a content source) executes a live GitBook API operation.
  add_section_group_to_site:
    endpoint: POST /orgs/{{ record.organization_id }}/sites/{{ record.site_id }}/section-groups
    required fields: organization_id, site_id, title
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
    required fields: organization_id, site_id, space_id
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
    required fields: organization_id, site_id, query
    risk: POST /orgs/{organizationId}/sites/{siteId}/search (Full-text search across all content in a site) executes a live GitBook API operation.
  stream_ask_in_site:
    endpoint: POST /orgs/{{ record.organization_id }}/sites/{{ record.site_id }}/ask
    required fields: organization_id, site_id, question, scope
    risk: POST /orgs/{organizationId}/sites/{siteId}/ask (Ask a question in a site) executes a live GitBook API operation.
  create_site_scan:
    endpoint: POST /orgs/{{ record.organization_id }}/sites/{{ record.site_id }}/scans
    required fields: organization_id, site_id, topic
    risk: POST /orgs/{organizationId}/sites/{siteId}/scans (Enqueue a new site scan) executes a live GitBook API operation.
  update_site_finding_by_id:
    endpoint: PATCH /orgs/{{ record.organization_id }}/sites/{{ record.site_id }}/findings/{{ record.site_finding_id }}
    required fields: organization_id, site_id, site_finding_id, status
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
    required fields: organization_id, site_id, site_topic_id, usage_settings
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
    required fields: organization_id, site_id, input
    risk: POST /orgs/{organizationId}/sites/{siteId}/ai/response (Generate an AI response in a site) executes a live GitBook API operation.
  update_site_agent_settings_by_id:
    endpoint: PUT /orgs/{{ record.organization_id }}/sites/{{ record.site_id }}/agent-settings
    required fields: organization_id, site_id, scans, findings, editing
    risk: PUT /orgs/{organizationId}/sites/{siteId}/agent-settings (Update the AI agent configuration for a site) executes a live GitBook API operation.
  create_site_styleguide_by_id:
    endpoint: POST /orgs/{{ record.organization_id }}/sites/{{ record.site_id }}/styleguide
    required fields: organization_id, site_id
    risk: POST /orgs/{organizationId}/sites/{siteId}/styleguide (Create or retrieve the styleguide space for a site) executes a live GitBook API operation.
  track_events_in_site_by_id:
    endpoint: POST /orgs/{{ record.organization_id }}/sites/{{ record.site_id }}/insights/events
    required fields: organization_id, site_id, events
    risk: POST /orgs/{organizationId}/sites/{siteId}/insights/events (Track site events) executes a live GitBook API operation.
  aggregate_site_events:
    endpoint: POST /orgs/{{ record.organization_id }}/sites/{{ record.site_id }}/insights/events/aggregate
    required fields: organization_id, site_id, range
    risk: POST /orgs/{organizationId}/sites/{siteId}/insights/events/aggregate (Query site events) executes a live GitBook API operation.
  update_site_ads_by_id:
    endpoint: POST /orgs/{{ record.organization_id }}/sites/{{ record.site_id }}/ads
    required fields: organization_id, site_id
    risk: POST /orgs/{organizationId}/sites/{siteId}/ads (Update the advertising settings for a site) executes a live GitBook API operation.
  create_site_redirect:
    endpoint: POST /orgs/{{ record.organization_id }}/sites/{{ record.site_id }}/redirects
    required fields: organization_id, site_id, source, destination
    risk: POST /orgs/{organizationId}/sites/{siteId}/redirects (Create a URL redirect rule for a site) executes a live GitBook API operation.
  bulk_upsert_site_redirects:
    endpoint: PUT /orgs/{{ record.organization_id }}/sites/{{ record.site_id }}/redirects
    required fields: organization_id, site_id, redirects
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
    required fields: organization_id, site_id, name, url, headers
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
    required fields: url
    risk: POST /urls/published (Resolve a URL of a published content.) executes a live GitBook API operation.
  install_git_sync_provider_on_target:
    endpoint: POST /git/installations
    required fields: provider, target
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

COMMAND SURFACE
  Run GitBook's declared streams and reverse-ETL actions.
  Usage: pm gitbook <command> [flags]
  Read streams
  Reverse ETL writes
  Other Commands
    add member to organization team by id apply - Plan and execute the add member to organization team by id reverse-ETL action [intent=reverse_etl availability=implemented write=add_member_to_organization_team_by_id]; approval: requires plan, preview, approval, and execute; risk: PUT /orgs/{organizationId}/teams/{teamId}/members/{userId} (Add a team member) executes a live GitBook API operation.; flags: --organization_id (required), --team_id (required), --user_id (required)
    add section group to site apply - Plan and execute the add section group to site reverse-ETL action [intent=reverse_etl availability=not_implemented write=add_section_group_to_site]; approval: requires plan, preview, approval, and execute; risk: POST /orgs/{organizationId}/sites/{siteId}/section-groups (Add a section group to a site's navigation structure) executes a live GitBook API operation.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    add section to site apply - Plan and execute the add section to site reverse-ETL action [intent=reverse_etl availability=not_implemented write=add_section_to_site]; approval: requires plan, preview, approval, and execute; risk: POST /orgs/{organizationId}/sites/{siteId}/sections (Add a new navigation section to a site backed by a space) executes a live GitBook API operation.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    add space to site apply - Plan and execute the add space to site reverse-ETL action [intent=reverse_etl availability=not_implemented write=add_space_to_site]; approval: requires plan, preview, approval, and execute; risk: POST /orgs/{organizationId}/sites/{siteId}/site-spaces (Add a space to a site as a content source) executes a live GitBook API operation.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    ads list sites list - Run the ads list sites ETL stream [intent=etl availability=implemented stream=ads_list_sites]
    ads update site apply - Plan and execute the ads update site reverse-ETL action [intent=reverse_etl availability=implemented write=ads_update_site]; approval: requires plan, preview, approval, and execute; risk: PATCH /ads/sites/{siteId} (Update the Ads configuration for a site) executes a live GitBook API operation.; flags: --site_id (required)
    aggregate site events apply - Plan and execute the aggregate site events reverse-ETL action [intent=reverse_etl availability=not_implemented write=aggregate_site_events]; approval: requires plan, preview, approval, and execute; risk: POST /orgs/{organizationId}/sites/{siteId}/insights/events/aggregate (Query site events) executes a live GitBook API operation.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    api get orgs organizationid ask questions stream - Documented GET /orgs/{organizationId}/ask/questions/stream (not implemented) [intent=direct_read availability=not_implemented operation=gitbook.get.orgs-organizationid-ask-questions-stream]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get orgs organizationid ask stream - Documented GET /orgs/{organizationId}/ask/stream (not implemented) [intent=direct_read availability=not_implemented operation=gitbook.get.orgs-organizationid-ask-stream]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get orgs organizationid openapi specslug versions latest content raw - Documented GET /orgs/{organizationId}/openapi/{specSlug}/versions/latest/content/raw (not implemented) [intent=direct_read availability=not_implemented operation=gitbook.get.orgs-organizationid-openapi-specslug-versions-latest-content-raw]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get orgs organizationid sites siteid ask questions - Documented GET /orgs/{organizationId}/sites/{siteId}/ask/questions (not implemented) [intent=direct_read availability=not_implemented operation=gitbook.get.orgs-organizationid-sites-siteid-ask-questions]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get orgs organizationid sites siteid stats findings - Documented GET /orgs/{organizationId}/sites/{siteId}/stats/findings (not implemented) [intent=direct_read availability=not_implemented operation=gitbook.get.orgs-organizationid-sites-siteid-stats-findings]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get orgs organizationid styleguide-templates - Documented GET /orgs/{organizationId}/styleguide-templates (not implemented) [intent=direct_read availability=not_implemented operation=gitbook.get.orgs-organizationid-styleguide-templates]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get orgs organizationid styleguides - Documented GET /orgs/{organizationId}/styleguides (not implemented) [intent=direct_read availability=not_implemented operation=gitbook.get.orgs-organizationid-styleguides]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api post orgs organizationid sites siteid agent-feedback - Documented POST /orgs/{organizationId}/sites/{siteId}/agent-feedback (not implemented) [intent=direct_write availability=not_implemented operation=gitbook.post.orgs-organizationid-sites-siteid-agent-feedback]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post orgs organizationid sites siteid context-connections sitecontextconnectionid test - Documented POST /orgs/{organizationId}/sites/{siteId}/context-connections/{siteContextConnectionId}/test (not implemented) [intent=direct_write availability=not_implemented operation=gitbook.post.orgs-organizationid-sites-siteid-context-connections-sitecontextconnectionid-test]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post orgs organizationid sites siteid site-spaces sitespaceid duplicate - Documented POST /orgs/{organizationId}/sites/{siteId}/site-spaces/{siteSpaceId}/duplicate (not implemented) [intent=direct_write availability=not_implemented operation=gitbook.post.orgs-organizationid-sites-siteid-site-spaces-sitespaceid-duplicate]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post orgs organizationid sites siteid template - Documented POST /orgs/{organizationId}/sites/{siteId}/template (not implemented) [intent=direct_write availability=not_implemented operation=gitbook.post.orgs-organizationid-sites-siteid-template]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put orgs organizationid sites siteid context-records - Documented PUT /orgs/{organizationId}/sites/{siteId}/context-records (not implemented) [intent=direct_write availability=not_implemented operation=gitbook.put.orgs-organizationid-sites-siteid-context-records]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    apply template to space apply - Plan and execute the apply template to space reverse-ETL action [intent=reverse_etl availability=not_implemented write=apply_template_to_space]; approval: requires plan, preview, approval, and execute; risk: POST /spaces/{spaceId}/content/template (Apply a content template to populate a space with initial pages) executes a live GitBook API operation.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    ask in organization apply - Plan and execute the ask in organization reverse-ETL action [intent=reverse_etl availability=not_implemented write=ask_in_organization]; approval: requires plan, preview, approval, and execute; risk: POST /orgs/{organizationId}/ask (Ask a question in an organization) executes a live GitBook API operation.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    bulk upsert site redirects apply - Plan and execute the bulk upsert site redirects reverse-ETL action [intent=reverse_etl availability=not_implemented write=bulk_upsert_site_redirects]; approval: requires plan, preview, approval, and execute; risk: PUT /orgs/{organizationId}/sites/{siteId}/redirects (Create, update, delete, or publish site redirect rules in bulk) executes a live GitBook API operation.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    cancel import run apply - Plan and execute the cancel import run reverse-ETL action [intent=reverse_etl availability=implemented write=cancel_import_run]; approval: requires plan, preview, approval, and execute; risk: POST /org/{organizationId}/imports/{importRunId}/cancel (Cancel an import run) executes a live GitBook API operation.; flags: --import_run_id (required), --organization_id (required)
    content list - Run the content ETL stream [intent=etl availability=implemented stream=content]
    create change request apply - Plan and execute the create change request reverse-ETL action [intent=reverse_etl availability=implemented write=create_change_request]; approval: requires plan, preview, approval, and execute; risk: POST /spaces/{spaceId}/change-requests (Create a new change request in a space) executes a live GitBook API operation.; flags: --space_id (required)
    create collection apply - Plan and execute the create collection reverse-ETL action [intent=reverse_etl availability=implemented write=create_collection]; approval: requires plan, preview, approval, and execute; risk: POST /orgs/{organizationId}/collections (Create a collection) executes a live GitBook API operation.; flags: --organization_id (required)
    create custom font apply - Plan and execute the create custom font reverse-ETL action [intent=reverse_etl availability=not_implemented write=create_custom_font]; approval: requires plan, preview, approval, and execute; risk: PUT /orgs/{organizationId}/fonts (Create a custom font) executes a live GitBook API operation.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    create integration installation token apply - Plan and execute the create integration installation token reverse-ETL action [intent=reverse_etl availability=implemented write=create_integration_installation_token]; approval: requires plan, preview, approval, and execute; risk: POST /integrations/{integrationName}/installations/{installationId}/tokens (Create an integration installation API token) executes a live GitBook API operation.; flags: --installation_id (required), --integration_name (required)
    create open api spec apply - Plan and execute the create open api spec reverse-ETL action [intent=reverse_etl availability=not_implemented write=create_open_api_spec]; approval: requires plan, preview, approval, and execute; risk: POST /orgs/{organizationId}/openapi (Create an OpenAPI spec) executes a live GitBook API operation.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    create or update open api spec by slug apply - Plan and execute the create or update open api spec by slug reverse-ETL action [intent=reverse_etl availability=not_implemented write=create_or_update_open_api_spec_by_slug]; approval: requires plan, preview, approval, and execute; risk: PUT /orgs/{organizationId}/openapi/{specSlug} (Create or update an OpenAPI spec) executes a live GitBook API operation.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    create organization invite apply - Plan and execute the create organization invite reverse-ETL action [intent=reverse_etl availability=implemented write=create_organization_invite]; approval: requires plan, preview, approval, and execute; risk: POST /orgs/{organizationId}/link-invites (Create an organization invite) executes a live GitBook API operation.; flags: --organization_id (required)
    create organization saml provider apply - Plan and execute the create organization saml provider reverse-ETL action [intent=reverse_etl availability=not_implemented write=create_organization_saml_provider]; approval: requires plan, preview, approval, and execute; risk: POST /orgs/{organizationId}/saml (Create a new SAML provider) executes a live GitBook API operation.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    create organization team apply - Plan and execute the create organization team reverse-ETL action [intent=reverse_etl availability=not_implemented write=create_organization_team]; approval: requires plan, preview, approval, and execute; risk: PUT /orgs/{organizationId}/teams (Create a team) executes a live GitBook API operation.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    create site apply - Plan and execute the create site reverse-ETL action [intent=reverse_etl availability=implemented write=create_site]; approval: requires plan, preview, approval, and execute; risk: POST /orgs/{organizationId}/sites (Create a new documentation site in an organization) executes a live GitBook API operation.; flags: --organization_id (required)
    create site channel apply - Plan and execute the create site channel reverse-ETL action [intent=reverse_etl availability=implemented write=create_site_channel]; approval: requires plan, preview, approval, and execute; risk: POST /orgs/{organizationId}/sites/{siteId}/channels (Create a new GitBook Agent channel for a site) executes a live GitBook API operation.; flags: --organization_id (required), --site_id (required)
    create site context connection apply - Plan and execute the create site context connection reverse-ETL action [intent=reverse_etl availability=implemented write=create_site_context_connection]; approval: requires plan, preview, approval, and execute; risk: POST /orgs/{organizationId}/sites/{siteId}/context-connections (Create a context connection) executes a live GitBook API operation.; flags: --organization_id (required), --site_id (required)
    create site mcp server apply - Plan and execute the create site mcp server reverse-ETL action [intent=reverse_etl availability=not_implemented write=create_site_mcp_server]; approval: requires plan, preview, approval, and execute; risk: POST /orgs/{organizationId}/sites/{siteId}/mcp-servers (Add a new MCP server configuration to a site) executes a live GitBook API operation.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    create site redirect apply - Plan and execute the create site redirect reverse-ETL action [intent=reverse_etl availability=not_implemented write=create_site_redirect]; approval: requires plan, preview, approval, and execute; risk: POST /orgs/{organizationId}/sites/{siteId}/redirects (Create a URL redirect rule for a site) executes a live GitBook API operation.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    create site scan apply - Plan and execute the create site scan reverse-ETL action [intent=reverse_etl availability=not_implemented write=create_site_scan]; approval: requires plan, preview, approval, and execute; risk: POST /orgs/{organizationId}/sites/{siteId}/scans (Enqueue a new site scan) executes a live GitBook API operation.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    create site share link apply - Plan and execute the create site share link reverse-ETL action [intent=reverse_etl availability=not_implemented write=create_site_share_link]; approval: requires plan, preview, approval, and execute; risk: POST /orgs/{organizationId}/sites/{siteId}/share-links (Create a private share link for a site) executes a live GitBook API operation.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    create site styleguide by id apply - Plan and execute the create site styleguide by id reverse-ETL action [intent=reverse_etl availability=implemented write=create_site_styleguide_by_id]; approval: requires plan, preview, approval, and execute; risk: POST /orgs/{organizationId}/sites/{siteId}/styleguide (Create or retrieve the styleguide space for a site) executes a live GitBook API operation.; flags: --organization_id (required), --site_id (required)
    create space apply - Plan and execute the create space reverse-ETL action [intent=reverse_etl availability=implemented write=create_space]; approval: requires plan, preview, approval, and execute; risk: POST /orgs/{organizationId}/spaces (Create a new documentation space in an organization) executes a live GitBook API operation.; flags: --organization_id (required)
    create translation apply - Plan and execute the create translation reverse-ETL action [intent=reverse_etl availability=not_implemented write=create_translation]; approval: requires plan, preview, approval, and execute; risk: POST /orgs/{organizationId}/translations (Create a translation) executes a live GitBook API operation.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    create user notifications token apply - Plan and execute the create user notifications token reverse-ETL action [intent=reverse_etl availability=implemented write=create_user_notifications_token]; approval: requires plan, preview, approval, and execute; risk: POST /user/notifications/token (Create a JWT to access the in-app notifications service) executes a live GitBook API operation.
    delete change request conversation apply - Plan and execute the delete change request conversation reverse-ETL action [intent=reverse_etl availability=implemented write=delete_change_request_conversation]; approval: requires plan, preview, approval, and execute; risk: DELETE /spaces/{spaceId}/change-requests/{changeRequestId}/conversations/{conversationId} (Delete an agent conversation) executes a live GitBook API operation.; flags: --change_request_id (required), --conversation_id (required), --space_id (required)
    delete collection by id apply - Plan and execute the delete collection by id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_collection_by_id]; approval: requires plan, preview, approval, and execute; risk: DELETE /collections/{collectionId} (Delete a collection) executes a live GitBook API operation.; flags: --collection_id (required)
    delete comment in change request apply - Plan and execute the delete comment in change request reverse-ETL action [intent=reverse_etl availability=implemented write=delete_comment_in_change_request]; approval: requires plan, preview, approval, and execute; risk: DELETE /spaces/{spaceId}/change-requests/{changeRequestId}/comments/{commentId} (Delete a change request comment) executes a live GitBook API operation.; flags: --change_request_id (required), --comment_id (required), --space_id (required)
    delete comment in space apply - Plan and execute the delete comment in space reverse-ETL action [intent=reverse_etl availability=implemented write=delete_comment_in_space]; approval: requires plan, preview, approval, and execute; risk: DELETE /spaces/{spaceId}/comments/{commentId} (Delete a space comment) executes a live GitBook API operation.; flags: --comment_id (required), --space_id (required)
    delete comment reply in change request apply - Plan and execute the delete comment reply in change request reverse-ETL action [intent=reverse_etl availability=implemented write=delete_comment_reply_in_change_request]; approval: requires plan, preview, approval, and execute; risk: DELETE /spaces/{spaceId}/change-requests/{changeRequestId}/comments/{commentId}/replies/{commentReplyId} (Delete a change request comment reply) executes a live GitBook API operation.; flags: --change_request_id (required), --comment_id (required), --comment_reply_id (required), --space_id (required)
    delete comment reply in space apply - Plan and execute the delete comment reply in space reverse-ETL action [intent=reverse_etl availability=implemented write=delete_comment_reply_in_space]; approval: requires plan, preview, approval, and execute; risk: DELETE /spaces/{spaceId}/comments/{commentId}/replies/{commentReplyId} (Delete a space comment reply) executes a live GitBook API operation.; flags: --comment_id (required), --comment_reply_id (required), --space_id (required)
    delete custom font apply - Plan and execute the delete custom font reverse-ETL action [intent=reverse_etl availability=implemented write=delete_custom_font]; approval: requires plan, preview, approval, and execute; risk: DELETE /orgs/{organizationId}/fonts/{fontId} (Delete a custom font) executes a live GitBook API operation.; flags: --font_id (required), --organization_id (required)
    delete legacy git installation apply - Plan and execute the delete legacy git installation reverse-ETL action [intent=reverse_etl availability=implemented write=delete_legacy_git_installation]; approval: requires plan, preview, approval, and execute; risk: DELETE /spaces/{spaceId}/git/legacy-installation (Remove the legacy Git Sync installation from the space to be able to upgrade it to use the new Git integrations) executes a live GitBook API operation.; flags: --space_id (required)
    delete member from organization team by id apply - Plan and execute the delete member from organization team by id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_member_from_organization_team_by_id]; approval: requires plan, preview, approval, and execute; risk: DELETE /orgs/{organizationId}/teams/{teamId}/members/{userId} (Delete a team member) executes a live GitBook API operation.; flags: --organization_id (required), --team_id (required), --user_id (required)
    delete open api spec by slug apply - Plan and execute the delete open api spec by slug reverse-ETL action [intent=reverse_etl availability=implemented write=delete_open_api_spec_by_slug]; approval: requires plan, preview, approval, and execute; risk: DELETE /orgs/{organizationId}/openapi/{specSlug} (Delete an OpenAPI spec) executes a live GitBook API operation.; flags: --organization_id (required), --spec_slug (required)
    delete organization invite by id apply - Plan and execute the delete organization invite by id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_organization_invite_by_id]; approval: requires plan, preview, approval, and execute; risk: DELETE /orgs/{organizationId}/link-invites/{inviteId} (Deletes an organization invite.) executes a live GitBook API operation.; flags: --invite_id (required), --organization_id (required)
    delete organization saml provider apply - Plan and execute the delete organization saml provider reverse-ETL action [intent=reverse_etl availability=implemented write=delete_organization_saml_provider]; approval: requires plan, preview, approval, and execute; risk: DELETE /orgs/{organizationId}/saml/{samlProviderId} (Delete a SAML provider) executes a live GitBook API operation.; flags: --organization_id (required), --saml_provider_id (required)
    delete site by id apply - Plan and execute the delete site by id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_site_by_id]; approval: requires plan, preview, approval, and execute; risk: DELETE /orgs/{organizationId}/sites/{siteId} (Delete a site) executes a live GitBook API operation.; flags: --organization_id (required), --site_id (required)
    delete site channel by id apply - Plan and execute the delete site channel by id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_site_channel_by_id]; approval: requires plan, preview, approval, and execute; risk: DELETE /orgs/{organizationId}/sites/{siteId}/channels/{siteChannelId} (Delete a GitBook Agent channel from a site) executes a live GitBook API operation.; flags: --organization_id (required), --site_channel_id (required), --site_id (required)
    delete site context connection by id apply - Plan and execute the delete site context connection by id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_site_context_connection_by_id]; approval: requires plan, preview, approval, and execute; risk: DELETE /orgs/{organizationId}/sites/{siteId}/context-connections/{siteContextConnectionId} (Delete a context connection) executes a live GitBook API operation.; flags: --organization_id (required), --site_context_connection_id (required), --site_id (required)
    delete site mcp server by id apply - Plan and execute the delete site mcp server by id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_site_mcp_server_by_id]; approval: requires plan, preview, approval, and execute; risk: DELETE /orgs/{organizationId}/sites/{siteId}/mcp-servers/{siteMcpServerId} (Delete a site MCP server) executes a live GitBook API operation.; flags: --organization_id (required), --site_id (required), --site_mcp_server_id (required)
    delete site redirect by id apply - Plan and execute the delete site redirect by id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_site_redirect_by_id]; approval: requires plan, preview, approval, and execute; risk: DELETE /orgs/{organizationId}/sites/{siteId}/redirects/{siteRedirectId} (Delete a site redirect) executes a live GitBook API operation.; flags: --organization_id (required), --site_id (required), --site_redirect_id (required)
    delete site section by id apply - Plan and execute the delete site section by id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_site_section_by_id]; approval: requires plan, preview, approval, and execute; risk: DELETE /orgs/{organizationId}/sites/{siteId}/sections/{siteSectionId} (Delete a site section) executes a live GitBook API operation.; flags: --organization_id (required), --site_id (required), --site_section_id (required)
    delete site section group by id apply - Plan and execute the delete site section group by id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_site_section_group_by_id]; approval: requires plan, preview, approval, and execute; risk: DELETE /orgs/{organizationId}/sites/{siteId}/section-groups/{siteSectionGroupId} (Delete a site section group) executes a live GitBook API operation.; flags: --organization_id (required), --site_id (required), --site_section_group_id (required)
    delete site share link by id apply - Plan and execute the delete site share link by id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_site_share_link_by_id]; approval: requires plan, preview, approval, and execute; risk: DELETE /orgs/{organizationId}/sites/{siteId}/share-links/{shareLinkId} (Deletes a share link) executes a live GitBook API operation.; flags: --organization_id (required), --share_link_id (required), --site_id (required)
    delete site space by id apply - Plan and execute the delete site space by id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_site_space_by_id]; approval: requires plan, preview, approval, and execute; risk: DELETE /orgs/{organizationId}/sites/{siteId}/site-spaces/{siteSpaceId} (Delete a site space) executes a live GitBook API operation.; flags: --organization_id (required), --site_id (required), --site_space_id (required)
    delete site space customization by id apply - Plan and execute the delete site space customization by id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_site_space_customization_by_id]; approval: requires plan, preview, approval, and execute; risk: DELETE /orgs/{organizationId}/sites/{siteId}/site-spaces/{siteSpaceId}/customization (Delete a site space customization settings) executes a live GitBook API operation.; flags: --organization_id (required), --site_id (required), --site_space_id (required)
    delete site topic findings apply - Plan and execute the delete site topic findings reverse-ETL action [intent=reverse_etl availability=implemented write=delete_site_topic_findings]; approval: requires plan, preview, approval, and execute; risk: DELETE /orgs/{organizationId}/sites/{siteId}/topics/{siteTopicId}/findings (Delete all findings for a topic) executes a live GitBook API operation.; flags: --organization_id (required), --site_id (required), --site_topic_id (required)
    delete space by id apply - Plan and execute the delete space by id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_space_by_id]; approval: requires plan, preview, approval, and execute; risk: DELETE /spaces/{spaceId} (Delete a space) executes a live GitBook API operation.; flags: --space_id (required)
    delete translation apply - Plan and execute the delete translation reverse-ETL action [intent=reverse_etl availability=implemented write=delete_translation]; approval: requires plan, preview, approval, and execute; risk: DELETE /orgs/{organizationId}/translations/{translationId} (Delete a translation) executes a live GitBook API operation.; flags: --organization_id (required), --translation_id (required)
    disable integration development mode apply - Plan and execute the disable integration development mode reverse-ETL action [intent=reverse_etl availability=implemented write=disable_integration_development_mode]; approval: requires plan, preview, approval, and execute; risk: DELETE /integrations/{integrationName}/dev (Disable integration dev mode) executes a live GitBook API operation.; flags: --integration_name (required)
    dns revalidate custom hostname apply - Plan and execute the dns revalidate custom hostname reverse-ETL action [intent=reverse_etl availability=implemented write=dns_revalidate_custom_hostname]; approval: requires plan, preview, approval, and execute; risk: PATCH /custom-hostnames/{hostname} (Revalidate a custom hostname DNS) executes a live GitBook API operation.; flags: --hostname (required)
    duplicate space apply - Plan and execute the duplicate space reverse-ETL action [intent=reverse_etl availability=implemented write=duplicate_space]; approval: requires plan, preview, approval, and execute; risk: POST /spaces/{spaceId}/duplicate (Create a full copy of a space) executes a live GitBook API operation.; flags: --space_id (required)
    export to git repository apply - Plan and execute the export to git repository reverse-ETL action [intent=reverse_etl availability=not_implemented write=export_to_git_repository]; approval: requires plan, preview, approval, and execute; risk: POST /spaces/{spaceId}/git/export (Push space content to a connected Git repository) executes a live GitBook API operation.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    generate storage upload url apply - Plan and execute the generate storage upload url reverse-ETL action [intent=reverse_etl availability=not_implemented write=generate_storage_upload_url]; approval: requires plan, preview, approval, and execute; risk: POST /orgs/{organizationId}/storage/upload (Create a signed URL to upload a file) executes a live GitBook API operation.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    get api information list - Run the get api information ETL stream [intent=etl availability=implemented stream=get_api_information]
    get change request by id list - Run the get change request by id ETL stream [intent=etl availability=implemented stream=get_change_request_by_id]
    get change request changes list - Run the get change request changes ETL stream [intent=etl availability=implemented stream=get_change_request_changes]
    get change request pdf list - Run the get change request pdf ETL stream [intent=etl availability=implemented stream=get_change_request_pdf]
    get change request review by id list - Run the get change request review by id ETL stream [intent=etl availability=implemented stream=get_change_request_review_by_id]
    get collection by id list - Run the get collection by id ETL stream [intent=etl availability=implemented stream=get_collection_by_id]
    get comment in change request list - Run the get comment in change request ETL stream [intent=etl availability=implemented stream=get_comment_in_change_request]
    get comment in space list - Run the get comment in space ETL stream [intent=etl availability=implemented stream=get_comment_in_space]
    get comment reply in change request list - Run the get comment reply in change request ETL stream [intent=etl availability=implemented stream=get_comment_reply_in_change_request]
    get comment reply in space list - Run the get comment reply in space ETL stream [intent=etl availability=implemented stream=get_comment_reply_in_space]
    get computed document apply - Plan and execute the get computed document reverse-ETL action [intent=reverse_etl availability=not_implemented write=get_computed_document]; approval: requires plan, preview, approval, and execute; risk: POST /spaces/{spaceId}/content/computed/document (Compute and render a document from a structured content source) executes a live GitBook API operation.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    get computed revision apply - Plan and execute the get computed revision reverse-ETL action [intent=reverse_etl availability=not_implemented write=get_computed_revision]; approval: requires plan, preview, approval, and execute; risk: POST /spaces/{spaceId}/content/computed/revision (Compute and render a full revision from a structured content source) executes a live GitBook API operation.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    get content by url list - Run the get content by url ETL stream [intent=etl availability=implemented stream=get_content_by_url]
    get contributors by change request id list - Run the get contributors by change request id ETL stream [intent=etl availability=implemented stream=get_contributors_by_change_request_id]
    get current revision list - Run the get current revision ETL stream [intent=etl availability=implemented stream=get_current_revision]
    get custom font list - Run the get custom font ETL stream [intent=etl availability=implemented stream=get_custom_font]
    get custom hostname list - Run the get custom hostname ETL stream [intent=etl availability=implemented stream=get_custom_hostname]
    get document by id list - Run the get document by id ETL stream [intent=etl availability=implemented stream=get_document_by_id]
    get embed by url in space list - Run the get embed by url in space ETL stream [intent=etl availability=implemented stream=get_embed_by_url_in_space]
    get embed by url list - Run the get embed by url ETL stream [intent=etl availability=implemented stream=get_embed_by_url]
    get file by id list - Run the get file by id ETL stream [intent=etl availability=implemented stream=get_file_by_id]
    get file in change request by id list - Run the get file in change request by id ETL stream [intent=etl availability=implemented stream=get_file_in_change_request_by_id]
    get file in revision by id list - Run the get file in revision by id ETL stream [intent=etl availability=implemented stream=get_file_in_revision_by_id]
    get git sync installation by id list - Run the get git sync installation by id ETL stream [intent=etl availability=implemented stream=get_git_sync_installation_by_id]
    get glossary entry list - Run the get glossary entry ETL stream [intent=etl availability=implemented stream=get_glossary_entry]
    get integration by name list - Run the get integration by name ETL stream [intent=etl availability=implemented stream=get_integration_by_name]
    get integration event list - Run the get integration event ETL stream [intent=etl availability=implemented stream=get_integration_event]
    get integration installation by id list - Run the get integration installation by id ETL stream [intent=etl availability=implemented stream=get_integration_installation_by_id]
    get integration site installation list - Run the get integration site installation ETL stream [intent=etl availability=implemented stream=get_integration_site_installation]
    get integration space installation list - Run the get integration space installation ETL stream [intent=etl availability=implemented stream=get_integration_space_installation]
    get latest open api spec version content list - Run the get latest open api spec version content ETL stream [intent=etl availability=implemented stream=get_latest_open_api_spec_version_content]
    get latest open api spec version list - Run the get latest open api spec version ETL stream [intent=etl availability=implemented stream=get_latest_open_api_spec_version]
    get member in organization by id list - Run the get member in organization by id ETL stream [intent=etl availability=implemented stream=get_member_in_organization_by_id]
    get open api spec by slug list - Run the get open api spec by slug ETL stream [intent=etl availability=implemented stream=get_open_api_spec_by_slug]
    get open api spec version by id list - Run the get open api spec version by id ETL stream [intent=etl availability=implemented stream=get_open_api_spec_version_by_id]
    get open api spec version content by id list - Run the get open api spec version content by id ETL stream [intent=etl availability=implemented stream=get_open_api_spec_version_content_by_id]
    get organization agent instructions list - Run the get organization agent instructions ETL stream [intent=etl availability=implemented stream=get_organization_agent_instructions]
    get organization by id list - Run the get organization by id ETL stream [intent=etl availability=implemented stream=get_organization_by_id]
    get organization integration status list - Run the get organization integration status ETL stream [intent=etl availability=implemented stream=get_organization_integration_status]
    get organization invite link list - Run the get organization invite link ETL stream [intent=etl availability=implemented stream=get_organization_invite_link]
    get organization saml provider by id list - Run the get organization saml provider by id ETL stream [intent=etl availability=implemented stream=get_organization_saml_provider_by_id]
    get organizations for email domain list - Run the get organizations for email domain ETL stream [intent=etl availability=implemented stream=get_organizations_for_email_domain]
    get page by id list - Run the get page by id ETL stream [intent=etl availability=implemented stream=get_page_by_id]
    get page by path list - Run the get page by path ETL stream [intent=etl availability=implemented stream=get_page_by_path]
    get page document in revision by id list - Run the get page document in revision by id ETL stream [intent=etl availability=implemented stream=get_page_document_in_revision_by_id]
    get page in change request by id list - Run the get page in change request by id ETL stream [intent=etl availability=implemented stream=get_page_in_change_request_by_id]
    get page in change request by path list - Run the get page in change request by path ETL stream [intent=etl availability=implemented stream=get_page_in_change_request_by_path]
    get page in revision by id list - Run the get page in revision by id ETL stream [intent=etl availability=implemented stream=get_page_in_revision_by_id]
    get page in revision by path list - Run the get page in revision by path ETL stream [intent=etl availability=implemented stream=get_page_in_revision_by_path]
    get published content by url list - Run the get published content by url ETL stream [intent=etl availability=implemented stream=get_published_content_by_url]
    get published content site list - Run the get published content site ETL stream [intent=etl availability=implemented stream=get_published_content_site]
    get recommended questions in organization list - Run the get recommended questions in organization ETL stream [intent=etl availability=implemented stream=get_recommended_questions_in_organization]
    get requested reviewers by change request id list - Run the get requested reviewers by change request id ETL stream [intent=etl availability=implemented stream=get_requested_reviewers_by_change_request_id]
    get reusable content by id list - Run the get reusable content by id ETL stream [intent=etl availability=implemented stream=get_reusable_content_by_id]
    get reusable content document in revision by id list - Run the get reusable content document in revision by id ETL stream [intent=etl availability=implemented stream=get_reusable_content_document_in_revision_by_id]
    get reusable content in change request by id list - Run the get reusable content in change request by id ETL stream [intent=etl availability=implemented stream=get_reusable_content_in_change_request_by_id]
    get reusable content in revision by id list - Run the get reusable content in revision by id ETL stream [intent=etl availability=implemented stream=get_reusable_content_in_revision_by_id]
    get reviews by change request id list - Run the get reviews by change request id ETL stream [intent=etl availability=implemented stream=get_reviews_by_change_request_id]
    get revision by id list - Run the get revision by id ETL stream [intent=etl availability=implemented stream=get_revision_by_id]
    get revision of change request by id list - Run the get revision of change request by id ETL stream [intent=etl availability=implemented stream=get_revision_of_change_request_by_id]
    get revision semantic changes list - Run the get revision semantic changes ETL stream [intent=etl availability=implemented stream=get_revision_semantic_changes]
    get site adaptive schema list - Run the get site adaptive schema ETL stream [intent=etl availability=implemented stream=get_site_adaptive_schema]
    get site agent settings by id list - Run the get site agent settings by id ETL stream [intent=etl availability=implemented stream=get_site_agent_settings_by_id]
    get site by id list - Run the get site by id ETL stream [intent=etl availability=implemented stream=get_site_by_id]
    get site channel by id list - Run the get site channel by id ETL stream [intent=etl availability=implemented stream=get_site_channel_by_id]
    get site context connection by id list - Run the get site context connection by id ETL stream [intent=etl availability=implemented stream=get_site_context_connection_by_id]
    get site context record by id list - Run the get site context record by id ETL stream [intent=etl availability=implemented stream=get_site_context_record_by_id]
    get site customization by id list - Run the get site customization by id ETL stream [intent=etl availability=implemented stream=get_site_customization_by_id]
    get site finding by id list - Run the get site finding by id ETL stream [intent=etl availability=implemented stream=get_site_finding_by_id]
    get site mcp server by id list - Run the get site mcp server by id ETL stream [intent=etl availability=implemented stream=get_site_mcp_server_by_id]
    get site publishing auth by id list - Run the get site publishing auth by id ETL stream [intent=etl availability=implemented stream=get_site_publishing_auth_by_id]
    get site publishing preview by id list - Run the get site publishing preview by id ETL stream [intent=etl availability=implemented stream=get_site_publishing_preview_by_id]
    get site question answer by id list - Run the get site question answer by id ETL stream [intent=etl availability=implemented stream=get_site_question_answer_by_id]
    get site question answer thread by id list - Run the get site question answer thread by id ETL stream [intent=etl availability=implemented stream=get_site_question_answer_thread_by_id]
    get site question by id list - Run the get site question by id ETL stream [intent=etl availability=implemented stream=get_site_question_by_id]
    get site question stats list - Run the get site question stats ETL stream [intent=etl availability=implemented stream=get_site_question_stats]
    get site redirect by source list - Run the get site redirect by source ETL stream [intent=etl availability=implemented stream=get_site_redirect_by_source]
    get site scan by id list - Run the get site scan by id ETL stream [intent=etl availability=implemented stream=get_site_scan_by_id]
    get site space customization by id list - Run the get site space customization by id ETL stream [intent=etl availability=implemented stream=get_site_space_customization_by_id]
    get site structure list - Run the get site structure ETL stream [intent=etl availability=implemented stream=get_site_structure]
    get site topic by id list - Run the get site topic by id ETL stream [intent=etl availability=implemented stream=get_site_topic_by_id]
    get space by id list - Run the get space by id ETL stream [intent=etl availability=implemented stream=get_space_by_id]
    get space git info list - Run the get space git info ETL stream [intent=etl availability=implemented stream=get_space_git_info]
    get space pdf list - Run the get space pdf ETL stream [intent=etl availability=implemented stream=get_space_pdf]
    get subdomain list - Run the get subdomain ETL stream [intent=etl availability=implemented stream=get_subdomain]
    get team in organization by id list - Run the get team in organization by id ETL stream [intent=etl availability=implemented stream=get_team_in_organization_by_id]
    get translation list - Run the get translation ETL stream [intent=etl availability=implemented stream=get_translation]
    get user by id list - Run the get user by id ETL stream [intent=etl availability=implemented stream=get_user_by_id]
    import git repository apply - Plan and execute the import git repository reverse-ETL action [intent=reverse_etl availability=not_implemented write=import_git_repository]; approval: requires plan, preview, approval, and execute; risk: POST /spaces/{spaceId}/git/import (Pull content into a space from a connected Git repository) executes a live GitBook API operation.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    install git sync provider on target apply - Plan and execute the install git sync provider on target reverse-ETL action [intent=reverse_etl availability=not_implemented write=install_git_sync_provider_on_target]; approval: requires plan, preview, approval, and execute; risk: POST /git/installations (Install a Git Sync provider on a target) executes a live GitBook API operation.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    install integration apply - Plan and execute the install integration reverse-ETL action [intent=reverse_etl availability=not_implemented write=install_integration]; approval: requires plan, preview, approval, and execute; risk: POST /integrations/{integrationName}/installations (Install an integration) executes a live GitBook API operation.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    install integration on site apply - Plan and execute the install integration on site reverse-ETL action [intent=reverse_etl availability=not_implemented write=install_integration_on_site]; approval: requires plan, preview, approval, and execute; risk: POST /integrations/{integrationName}/installations/{installationId}/sites (Install an integration on a site) executes a live GitBook API operation.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    install integration on space apply - Plan and execute the install integration on space reverse-ETL action [intent=reverse_etl availability=not_implemented write=install_integration_on_space]; approval: requires plan, preview, approval, and execute; risk: POST /integrations/{integrationName}/installations/{installationId}/spaces (Install an integration on a space) executes a live GitBook API operation.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    invite to collection apply - Plan and execute the invite to collection reverse-ETL action [intent=reverse_etl availability=implemented write=invite_to_collection]; approval: requires plan, preview, approval, and execute; risk: POST /collections/{collectionId}/permissions (Invite to a collection) executes a live GitBook API operation.; flags: --collection_id (required)
    invite to site apply - Plan and execute the invite to site reverse-ETL action [intent=reverse_etl availability=implemented write=invite_to_site]; approval: requires plan, preview, approval, and execute; risk: POST /orgs/{organizationId}/sites/{siteId}/permissions (Invite a user or a team to a site) executes a live GitBook API operation.; flags: --organization_id (required), --site_id (required)
    invite to space apply - Plan and execute the invite to space reverse-ETL action [intent=reverse_etl availability=implemented write=invite_to_space]; approval: requires plan, preview, approval, and execute; risk: POST /spaces/{spaceId}/permissions (Invite a user or a team to a space) executes a live GitBook API operation.; flags: --space_id (required)
    invite users to organization apply - Plan and execute the invite users to organization reverse-ETL action [intent=reverse_etl availability=not_implemented write=invite_users_to_organization]; approval: requires plan, preview, approval, and execute; risk: POST /orgs/{organizationId}/invites (Invite users in an organization) executes a live GitBook API operation.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    join organization apply - Plan and execute the join organization reverse-ETL action [intent=reverse_etl availability=implemented write=join_organization]; approval: requires plan, preview, approval, and execute; risk: POST /orgs/{organizationId}/join (Join an organization) executes a live GitBook API operation.; flags: --organization_id (required)
    join organization with invite apply - Plan and execute the join organization with invite reverse-ETL action [intent=reverse_etl availability=implemented write=join_organization_with_invite]; approval: requires plan, preview, approval, and execute; risk: POST /orgs/{organizationId}/invites/{inviteId} (Join an organization with an invite) executes a live GitBook API operation.; flags: --invite_id (required), --organization_id (required)
    list change request conversations list - Run the list change request conversations ETL stream [intent=etl availability=implemented stream=list_change_request_conversations]
    list change request file backlinks list - Run the list change request file backlinks ETL stream [intent=etl availability=implemented stream=list_change_request_file_backlinks]; notes: discrepancy=present-in-surface-absent-from-artifact
    list change request links list - Run the list change request links ETL stream [intent=etl availability=implemented stream=list_change_request_links]; notes: discrepancy=present-in-surface-absent-from-artifact
    list change request page backlinks list - Run the list change request page backlinks ETL stream [intent=etl availability=implemented stream=list_change_request_page_backlinks]; notes: discrepancy=present-in-surface-absent-from-artifact
    list change request page meta links list - Run the list change request page meta links ETL stream [intent=etl availability=implemented stream=list_change_request_page_meta_links]
    list change requests for organization list - Run the list change requests for organization ETL stream [intent=etl availability=implemented stream=list_change_requests_for_organization]
    list change requests for site finding list - Run the list change requests for site finding ETL stream [intent=etl availability=implemented stream=list_change_requests_for_site_finding]
    list change requests for space list - Run the list change requests for space ETL stream [intent=etl availability=implemented stream=list_change_requests_for_space]
    list collections in organization by id list - Run the list collections in organization by id ETL stream [intent=etl availability=implemented stream=list_collections_in_organization_by_id]
    list comment replies in change request list - Run the list comment replies in change request ETL stream [intent=etl availability=implemented stream=list_comment_replies_in_change_request]
    list comment replies in space list - Run the list comment replies in space ETL stream [intent=etl availability=implemented stream=list_comment_replies_in_space]
    list commenters in change request list - Run the list commenters in change request ETL stream [intent=etl availability=implemented stream=list_commenters_in_change_request]
    list commenters in space list - Run the list commenters in space ETL stream [intent=etl availability=implemented stream=list_commenters_in_space]
    list comments in change request list - Run the list comments in change request ETL stream [intent=etl availability=implemented stream=list_comments_in_change_request]
    list comments in space list - Run the list comments in space ETL stream [intent=etl availability=implemented stream=list_comments_in_space]
    list custom fonts list - Run the list custom fonts ETL stream [intent=etl availability=implemented stream=list_custom_fonts]
    list files in change request by id list - Run the list files in change request by id ETL stream [intent=etl availability=implemented stream=list_files_in_change_request_by_id]
    list files in revision by id list - Run the list files in revision by id ETL stream [intent=etl availability=implemented stream=list_files_in_revision_by_id]
    list files list - Run the list files ETL stream [intent=etl availability=implemented stream=list_files]
    list git hub repo branches for git sync installation list - Run the list git hub repo branches for git sync installation ETL stream [intent=etl availability=implemented stream=list_git_hub_repo_branches_for_git_sync_installation]
    list git hub repositories for git sync installation list - Run the list git hub repositories for git sync installation ETL stream [intent=etl availability=implemented stream=list_git_hub_repositories_for_git_sync_installation]
    list git lab project branches for git sync installation list - Run the list git lab project branches for git sync installation ETL stream [intent=etl availability=implemented stream=list_git_lab_project_branches_for_git_sync_installation]
    list git lab projects for git sync installation list - Run the list git lab projects for git sync installation ETL stream [intent=etl availability=implemented stream=list_git_lab_projects_for_git_sync_installation]
    list glossary entries list - Run the list glossary entries ETL stream [intent=etl availability=implemented stream=list_glossary_entries]
    list integration events list - Run the list integration events ETL stream [intent=etl availability=implemented stream=list_integration_events]
    list integration installation sites list - Run the list integration installation sites ETL stream [intent=etl availability=implemented stream=list_integration_installation_sites]
    list integration installation spaces list - Run the list integration installation spaces ETL stream [intent=etl availability=implemented stream=list_integration_installation_spaces]
    list integration installations list - Run the list integration installations ETL stream [intent=etl availability=implemented stream=list_integration_installations]
    list integration site installations list - Run the list integration site installations ETL stream [intent=etl availability=implemented stream=list_integration_site_installations]
    list integration space installations list - Run the list integration space installations ETL stream [intent=etl availability=implemented stream=list_integration_space_installations]
    list integrations list - Run the list integrations ETL stream [intent=etl availability=implemented stream=list_integrations]
    list open api spec versions list - Run the list open api spec versions ETL stream [intent=etl availability=implemented stream=list_open_api_spec_versions]
    list open api specs list - Run the list open api specs ETL stream [intent=etl availability=implemented stream=list_open_api_specs]
    list organization installations list - Run the list organization installations ETL stream [intent=etl availability=implemented stream=list_organization_installations]
    list organization integrations list - Run the list organization integrations ETL stream [intent=etl availability=implemented stream=list_organization_integrations]
    list organization integrations status list - Run the list organization integrations status ETL stream [intent=etl availability=implemented stream=list_organization_integrations_status]
    list organization invite links list - Run the list organization invite links ETL stream [intent=etl availability=implemented stream=list_organization_invite_links]
    list page links in change request list - Run the list page links in change request ETL stream [intent=etl availability=implemented stream=list_page_links_in_change_request]; notes: discrepancy=present-in-surface-absent-from-artifact
    list page links in space list - Run the list page links in space ETL stream [intent=etl availability=implemented stream=list_page_links_in_space]; notes: discrepancy=present-in-surface-absent-from-artifact
    list pages for site finding list - Run the list pages for site finding ETL stream [intent=etl availability=implemented stream=list_pages_for_site_finding]
    list pages in change request list - Run the list pages in change request ETL stream [intent=etl availability=implemented stream=list_pages_in_change_request]
    list pages in revision by id list - Run the list pages in revision by id ETL stream [intent=etl availability=implemented stream=list_pages_in_revision_by_id]
    list permissions aggregate in collection list - Run the list permissions aggregate in collection ETL stream [intent=etl availability=implemented stream=list_permissions_aggregate_in_collection]
    list permissions aggregate in site list - Run the list permissions aggregate in site ETL stream [intent=etl availability=implemented stream=list_permissions_aggregate_in_site]
    list permissions aggregate in space list - Run the list permissions aggregate in space ETL stream [intent=etl availability=implemented stream=list_permissions_aggregate_in_space]
    list questions for site finding list - Run the list questions for site finding ETL stream [intent=etl availability=implemented stream=list_questions_for_site_finding]
    list records for site finding list - Run the list records for site finding ETL stream [intent=etl availability=implemented stream=list_records_for_site_finding]
    list revision page meta links list - Run the list revision page meta links ETL stream [intent=etl availability=implemented stream=list_revision_page_meta_links]
    list saml providers in organization by id list - Run the list saml providers in organization by id ETL stream [intent=etl availability=implemented stream=list_saml_providers_in_organization_by_id]
    list site adaptive template conditions list - Run the list site adaptive template conditions ETL stream [intent=etl availability=implemented stream=list_site_adaptive_template_conditions]
    list site channels list - Run the list site channels ETL stream [intent=etl availability=implemented stream=list_site_channels]
    list site context connections list - Run the list site context connections ETL stream [intent=etl availability=implemented stream=list_site_context_connections]
    list site context records list - Run the list site context records ETL stream [intent=etl availability=implemented stream=list_site_context_records]
    list site findings list - Run the list site findings ETL stream [intent=etl availability=implemented stream=list_site_findings]
    list site git sync installations list - Run the list site git sync installations ETL stream [intent=etl availability=implemented stream=list_site_git_sync_installations]
    list site integration scripts list - Run the list site integration scripts ETL stream [intent=etl availability=implemented stream=list_site_integration_scripts]
    list site integrations list - Run the list site integrations ETL stream [intent=etl availability=implemented stream=list_site_integrations]
    list site mcp servers list - Run the list site mcp servers ETL stream [intent=etl availability=implemented stream=list_site_mcp_servers]
    list site question answer sources list - Run the list site question answer sources ETL stream [intent=etl availability=implemented stream=list_site_question_answer_sources]
    list site question answers list - Run the list site question answers ETL stream [intent=etl availability=implemented stream=list_site_question_answers]
    list site question sources list - Run the list site question sources ETL stream [intent=etl availability=implemented stream=list_site_question_sources]
    list site questions list - Run the list site questions ETL stream [intent=etl availability=implemented stream=list_site_questions]
    list site redirects list - Run the list site redirects ETL stream [intent=etl availability=implemented stream=list_site_redirects]
    list site scans list - Run the list site scans ETL stream [intent=etl availability=implemented stream=list_site_scans]
    list site section groups list - Run the list site section groups ETL stream [intent=etl availability=implemented stream=list_site_section_groups]
    list site sections list - Run the list site sections ETL stream [intent=etl availability=implemented stream=list_site_sections]
    list site share links list - Run the list site share links ETL stream [intent=etl availability=implemented stream=list_site_share_links]
    list site spaces list - Run the list site spaces ETL stream [intent=etl availability=implemented stream=list_site_spaces]
    list site topics list - Run the list site topics ETL stream [intent=etl availability=implemented stream=list_site_topics]
    list site visitor segments list - Run the list site visitor segments ETL stream [intent=etl availability=implemented stream=list_site_visitor_segments]
    list sites list - Run the list sites ETL stream [intent=etl availability=implemented stream=list_sites]
    list space file backlinks list - Run the list space file backlinks ETL stream [intent=etl availability=implemented stream=list_space_file_backlinks]; notes: discrepancy=present-in-surface-absent-from-artifact
    list space integrations blocks list - Run the list space integrations blocks ETL stream [intent=etl availability=implemented stream=list_space_integrations_blocks]
    list space integrations list - Run the list space integrations ETL stream [intent=etl availability=implemented stream=list_space_integrations]
    list space links list - Run the list space links ETL stream [intent=etl availability=implemented stream=list_space_links]; notes: discrepancy=present-in-surface-absent-from-artifact
    list space page backlinks list - Run the list space page backlinks ETL stream [intent=etl availability=implemented stream=list_space_page_backlinks]; notes: discrepancy=present-in-surface-absent-from-artifact
    list space page meta links list - Run the list space page meta links ETL stream [intent=etl availability=implemented stream=list_space_page_meta_links]
    list spaces for organization member list - Run the list spaces for organization member ETL stream [intent=etl availability=implemented stream=list_spaces_for_organization_member]
    list spaces in collection by id list - Run the list spaces in collection by id ETL stream [intent=etl availability=implemented stream=list_spaces_in_collection_by_id]
    list spaces in organization by id list - Run the list spaces in organization by id ETL stream [intent=etl availability=implemented stream=list_spaces_in_organization_by_id]
    list sso provider logins in organization list - Run the list sso provider logins in organization ETL stream [intent=etl availability=implemented stream=list_sso_provider_logins_in_organization]
    list team members in organization by id list - Run the list team members in organization by id ETL stream [intent=etl availability=implemented stream=list_team_members_in_organization_by_id]
    list team permissions in collection list - Run the list team permissions in collection ETL stream [intent=etl availability=implemented stream=list_team_permissions_in_collection]
    list team permissions in site list - Run the list team permissions in site ETL stream [intent=etl availability=implemented stream=list_team_permissions_in_site]
    list team permissions in space list - Run the list team permissions in space ETL stream [intent=etl availability=implemented stream=list_team_permissions_in_space]
    list teams for organization member list - Run the list teams for organization member ETL stream [intent=etl availability=implemented stream=list_teams_for_organization_member]
    list teams in organization by id list - Run the list teams in organization by id ETL stream [intent=etl availability=implemented stream=list_teams_in_organization_by_id]
    list translations list - Run the list translations ETL stream [intent=etl availability=implemented stream=list_translations]
    list user permissions in collection list - Run the list user permissions in collection ETL stream [intent=etl availability=implemented stream=list_user_permissions_in_collection]
    list user permissions in site list - Run the list user permissions in site ETL stream [intent=etl availability=implemented stream=list_user_permissions_in_site]
    list user permissions in space list - Run the list user permissions in space ETL stream [intent=etl availability=implemented stream=list_user_permissions_in_space]
    merge change request apply - Plan and execute the merge change request reverse-ETL action [intent=reverse_etl availability=implemented write=merge_change_request]; approval: requires plan, preview, approval, and execute; risk: POST /spaces/{spaceId}/change-requests/{changeRequestId}/merge (Merge a change request into the space's live content) executes a live GitBook API operation.; flags: --change_request_id (required), --space_id (required)
    move collection apply - Plan and execute the move collection reverse-ETL action [intent=reverse_etl availability=implemented write=move_collection]; approval: requires plan, preview, approval, and execute; risk: POST /collections/{collectionId}/move (Move a collection to a new position.) executes a live GitBook API operation.; flags: --collection_id (required)
    move site section apply - Plan and execute the move site section reverse-ETL action [intent=reverse_etl availability=implemented write=move_site_section]; approval: requires plan, preview, approval, and execute; risk: POST /orgs/{organizationId}/sites/{siteId}/sections/{siteSectionId}/move (Move a site section to a new position. (Deprecated) use sortSiteStructure instead.) executes a live GitBook API operation.; flags: --organization_id (required), --site_id (required), --site_section_id (required)
    move site section group apply - Plan and execute the move site section group reverse-ETL action [intent=reverse_etl availability=implemented write=move_site_section_group]; approval: requires plan, preview, approval, and execute; risk: POST /orgs/{organizationId}/sites/{siteId}/section-groups/{siteSectionGroupId}/move (Move a site section group to a new position. (Deprecated) use sortSiteStructure instead.) executes a live GitBook API operation.; flags: --organization_id (required), --site_id (required), --site_section_group_id (required)
    move site space apply - Plan and execute the move site space reverse-ETL action [intent=reverse_etl availability=implemented write=move_site_space]; approval: requires plan, preview, approval, and execute; risk: POST /orgs/{organizationId}/sites/{siteId}/site-spaces/{siteSpaceId}/move (Move a site space to a new position. (Deprecated) use sortSiteStructure instead.) executes a live GitBook API operation.; flags: --organization_id (required), --site_id (required), --site_space_id (required)
    move space apply - Plan and execute the move space reverse-ETL action [intent=reverse_etl availability=implemented write=move_space]; approval: requires plan, preview, approval, and execute; risk: POST /spaces/{spaceId}/move (Move a space to a different collection or position) executes a live GitBook API operation.; flags: --space_id (required)
    org members list - Run the org members ETL stream [intent=etl availability=implemented stream=org_members]
    organizations list - Run the organizations ETL stream [intent=etl availability=implemented stream=organizations]
    override site space customization by id apply - Plan and execute the override site space customization by id reverse-ETL action [intent=reverse_etl availability=implemented write=override_site_space_customization_by_id]; approval: requires plan, preview, approval, and execute; risk: PATCH /orgs/{organizationId}/sites/{siteId}/site-spaces/{siteSpaceId}/customization (Override branding and customization settings for a specific site space) executes a live GitBook API operation.; flags: --organization_id (required), --site_id (required), --site_space_id (required)
    post comment in change request apply - Plan and execute the post comment in change request reverse-ETL action [intent=reverse_etl availability=not_implemented write=post_comment_in_change_request]; approval: requires plan, preview, approval, and execute; risk: POST /spaces/{spaceId}/change-requests/{changeRequestId}/comments (Post a new comment on a change request) executes a live GitBook API operation.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    post comment in space apply - Plan and execute the post comment in space reverse-ETL action [intent=reverse_etl availability=not_implemented write=post_comment_in_space]; approval: requires plan, preview, approval, and execute; risk: POST /spaces/{spaceId}/comments (Post a new comment on a space or a specific page) executes a live GitBook API operation.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    post comment reply in change request apply - Plan and execute the post comment reply in change request reverse-ETL action [intent=reverse_etl availability=not_implemented write=post_comment_reply_in_change_request]; approval: requires plan, preview, approval, and execute; risk: POST /spaces/{spaceId}/change-requests/{changeRequestId}/comments/{commentId}/replies (Post a reply to a change request comment) executes a live GitBook API operation.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    post comment reply in space apply - Plan and execute the post comment reply in space reverse-ETL action [intent=reverse_etl availability=not_implemented write=post_comment_reply_in_space]; approval: requires plan, preview, approval, and execute; risk: POST /spaces/{spaceId}/comments/{commentId}/replies (Post a reply to an existing space comment) executes a live GitBook API operation.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    publish integration apply - Plan and execute the publish integration reverse-ETL action [intent=reverse_etl availability=not_implemented write=publish_integration]; approval: requires plan, preview, approval, and execute; risk: POST /integrations/{integrationName} (Publish an integration) executes a live GitBook API operation.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    publish site apply - Plan and execute the publish site reverse-ETL action [intent=reverse_etl availability=implemented write=publish_site]; approval: requires plan, preview, approval, and execute; risk: POST /orgs/{organizationId}/sites/{siteId}/publish (Publish a site to make it publicly accessible) executes a live GitBook API operation.; flags: --organization_id (required), --site_id (required)
    queue integration task apply - Plan and execute the queue integration task reverse-ETL action [intent=reverse_etl availability=not_implemented write=queue_integration_task]; approval: requires plan, preview, approval, and execute; risk: POST /integrations/{integrationName}/tasks (Queue an integration task) executes a live GitBook API operation.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    regenerate site publishing auth by id apply - Plan and execute the regenerate site publishing auth by id reverse-ETL action [intent=reverse_etl availability=implemented write=regenerate_site_publishing_auth_by_id]; approval: requires plan, preview, approval, and execute; risk: POST /orgs/{organizationId}/sites/{siteId}/publishing/auth/regenerate (Regenerate the private key for a site's published content authentication) executes a live GitBook API operation.; flags: --organization_id (required), --site_id (required)
    remove custom hostname apply - Plan and execute the remove custom hostname reverse-ETL action [intent=reverse_etl availability=implemented write=remove_custom_hostname]; approval: requires plan, preview, approval, and execute; risk: DELETE /custom-hostnames/{hostname} (Remove a custom hostname) executes a live GitBook API operation.; flags: --hostname (required)
    remove member from organization by id apply - Plan and execute the remove member from organization by id reverse-ETL action [intent=reverse_etl availability=implemented write=remove_member_from_organization_by_id]; approval: requires plan, preview, approval, and execute; risk: DELETE /orgs/{organizationId}/members/{userId} (Delete an organization member) executes a live GitBook API operation.; flags: --organization_id (required), --user_id (required)
    remove requested reviewer from change request apply - Plan and execute the remove requested reviewer from change request reverse-ETL action [intent=reverse_etl availability=implemented write=remove_requested_reviewer_from_change_request]; approval: requires plan, preview, approval, and execute; risk: DELETE /spaces/{spaceId}/change-requests/{changeRequestId}/requested-reviewers/{userId} (Remove a reviewer from a change request) executes a live GitBook API operation.; flags: --change_request_id (required), --space_id (required), --user_id (required)
    remove team from collection apply - Plan and execute the remove team from collection reverse-ETL action [intent=reverse_etl availability=implemented write=remove_team_from_collection]; approval: requires plan, preview, approval, and execute; risk: DELETE /collections/{collectionId}/permissions/teams/{teamId} (Remove an org team from a collection) executes a live GitBook API operation.; flags: --collection_id (required), --team_id (required)
    remove team from organization by id apply - Plan and execute the remove team from organization by id reverse-ETL action [intent=reverse_etl availability=implemented write=remove_team_from_organization_by_id]; approval: requires plan, preview, approval, and execute; risk: DELETE /orgs/{organizationId}/teams/{teamId} (Delete a team) executes a live GitBook API operation.; flags: --organization_id (required), --team_id (required)
    remove team from site apply - Plan and execute the remove team from site reverse-ETL action [intent=reverse_etl availability=implemented write=remove_team_from_site]; approval: requires plan, preview, approval, and execute; risk: DELETE /orgs/{organizationId}/sites/{siteId}/permissions/teams/{teamId} (Remove an org team from a site) executes a live GitBook API operation.; flags: --organization_id (required), --site_id (required), --team_id (required)
    remove team from space apply - Plan and execute the remove team from space reverse-ETL action [intent=reverse_etl availability=implemented write=remove_team_from_space]; approval: requires plan, preview, approval, and execute; risk: DELETE /spaces/{spaceId}/permissions/teams/{teamId} (Remove an org team from a space) executes a live GitBook API operation.; flags: --space_id (required), --team_id (required)
    remove user from collection apply - Plan and execute the remove user from collection reverse-ETL action [intent=reverse_etl availability=implemented write=remove_user_from_collection]; approval: requires plan, preview, approval, and execute; risk: DELETE /collections/{collectionId}/permissions/users/{userId} (Remove a user from a collection) executes a live GitBook API operation.; flags: --collection_id (required), --user_id (required)
    remove user from site apply - Plan and execute the remove user from site reverse-ETL action [intent=reverse_etl availability=implemented write=remove_user_from_site]; approval: requires plan, preview, approval, and execute; risk: DELETE /orgs/{organizationId}/sites/{siteId}/permissions/users/{userId} (Remove a site user) executes a live GitBook API operation.; flags: --organization_id (required), --site_id (required), --user_id (required)
    remove user from space apply - Plan and execute the remove user from space reverse-ETL action [intent=reverse_etl availability=implemented write=remove_user_from_space]; approval: requires plan, preview, approval, and execute; risk: DELETE /spaces/{spaceId}/permissions/users/{userId} (Remove a space user) executes a live GitBook API operation.; flags: --space_id (required), --user_id (required)
    render integration ui with get list - Run the render integration ui with get ETL stream [intent=etl availability=implemented stream=render_integration_ui_with_get]
    render integration ui with post apply - Plan and execute the render integration ui with post reverse-ETL action [intent=reverse_etl availability=not_implemented write=render_integration_ui_with_post]; approval: requires plan, preview, approval, and execute; risk: POST /integrations/{integrationName}/render (Render an integration UI with POST method) executes a live GitBook API operation.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    request reviewers for change request apply - Plan and execute the request reviewers for change request reverse-ETL action [intent=reverse_etl availability=not_implemented write=request_reviewers_for_change_request]; approval: requires plan, preview, approval, and execute; risk: POST /spaces/{spaceId}/change-requests/{changeRequestId}/requested-reviewers (Send review requests to users for a change request) executes a live GitBook API operation.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    resolve published content by url apply - Plan and execute the resolve published content by url reverse-ETL action [intent=reverse_etl availability=not_implemented write=resolve_published_content_by_url]; approval: requires plan, preview, approval, and execute; risk: POST /urls/published (Resolve a URL of a published content.) executes a live GitBook API operation.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    restore space apply - Plan and execute the restore space reverse-ETL action [intent=reverse_etl availability=implemented write=restore_space]; approval: requires plan, preview, approval, and execute; risk: POST /spaces/{spaceId}/restore (Restore a recently deleted space from the trash) executes a live GitBook API operation.; flags: --space_id (required)
    run translation apply - Plan and execute the run translation reverse-ETL action [intent=reverse_etl availability=implemented write=run_translation]; approval: requires plan, preview, approval, and execute; risk: POST /orgs/{organizationId}/translations/{translationId}/run (Run a translation again) executes a live GitBook API operation.; flags: --organization_id (required), --translation_id (required)
    search organization content list - Run the search organization content ETL stream [intent=etl availability=implemented stream=search_organization_content]
    search site content apply - Plan and execute the search site content reverse-ETL action [intent=reverse_etl availability=not_implemented write=search_site_content]; approval: requires plan, preview, approval, and execute; risk: POST /orgs/{organizationId}/sites/{siteId}/search (Full-text search across all content in a site) executes a live GitBook API operation.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    search space content list - Run the search space content ETL stream [intent=etl availability=implemented stream=search_space_content]
    set integration development mode apply - Plan and execute the set integration development mode reverse-ETL action [intent=reverse_etl availability=not_implemented write=set_integration_development_mode]; approval: requires plan, preview, approval, and execute; risk: PUT /integrations/{integrationName}/dev (Enable integration dev mode) executes a live GitBook API operation.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    set user as sso member for organization apply - Plan and execute the set user as sso member for organization reverse-ETL action [intent=reverse_etl availability=implemented write=set_user_as_sso_member_for_organization]; approval: requires plan, preview, approval, and execute; risk: POST /orgs/{organizationId}/members/{userId}/sso (Set a user as an SSO member of an organization) executes a live GitBook API operation.; flags: --organization_id (required), --user_id (required)
    sort site structure apply - Plan and execute the sort site structure reverse-ETL action [intent=reverse_etl availability=not_implemented write=sort_site_structure]; approval: requires plan, preview, approval, and execute; risk: PATCH /orgs/{organizationId}/sites/{siteId}/structure/sort (Move a site space, section, or section group to a new position) executes a live GitBook API operation.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    start import run apply - Plan and execute the start import run reverse-ETL action [intent=reverse_etl availability=not_implemented write=start_import_run]; approval: requires plan, preview, approval, and execute; risk: POST /org/{organizationId}/imports (Import content into a space from a website) executes a live GitBook API operation.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    stream ai response in site apply - Plan and execute the stream ai response in site reverse-ETL action [intent=reverse_etl availability=not_implemented write=stream_ai_response_in_site]; approval: requires plan, preview, approval, and execute; risk: POST /orgs/{organizationId}/sites/{siteId}/ai/response (Generate an AI response in a site) executes a live GitBook API operation.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    stream ask in site apply - Plan and execute the stream ask in site reverse-ETL action [intent=reverse_etl availability=not_implemented write=stream_ask_in_site]; approval: requires plan, preview, approval, and execute; risk: POST /orgs/{organizationId}/sites/{siteId}/ask (Ask a question in a site) executes a live GitBook API operation.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    submit change request review apply - Plan and execute the submit change request review reverse-ETL action [intent=reverse_etl availability=not_implemented write=submit_change_request_review]; approval: requires plan, preview, approval, and execute; risk: POST /spaces/{spaceId}/change-requests/{changeRequestId}/reviews (Submit an approve or request-changes review for a change request) executes a live GitBook API operation.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    sync site context connection apply - Plan and execute the sync site context connection reverse-ETL action [intent=reverse_etl availability=implemented write=sync_site_context_connection]; approval: requires plan, preview, approval, and execute; risk: POST /orgs/{organizationId}/sites/{siteId}/context-connections/{siteContextConnectionId}/sync (Trigger a sync for a context connection) executes a live GitBook API operation.; flags: --organization_id (required), --site_context_connection_id (required), --site_id (required)
    track events in site by id apply - Plan and execute the track events in site by id reverse-ETL action [intent=reverse_etl availability=not_implemented write=track_events_in_site_by_id]; approval: requires plan, preview, approval, and execute; risk: POST /orgs/{organizationId}/sites/{siteId}/insights/events (Track site events) executes a live GitBook API operation.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    transfer collection apply - Plan and execute the transfer collection reverse-ETL action [intent=reverse_etl availability=not_implemented write=transfer_collection]; approval: requires plan, preview, approval, and execute; risk: POST /collections/{collectionId}/transfer (Transfer a collection) executes a live GitBook API operation.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    trigger change requests for site finding apply - Plan and execute the trigger change requests for site finding reverse-ETL action [intent=reverse_etl availability=implemented write=trigger_change_requests_for_site_finding]; approval: requires plan, preview, approval, and execute; risk: POST /orgs/{organizationId}/sites/{siteId}/findings/{siteFindingId}/change-requests (Process a site finding into change requests) executes a live GitBook API operation.; flags: --organization_id (required), --site_finding_id (required), --site_id (required)
    uninstall git sync installation apply - Plan and execute the uninstall git sync installation reverse-ETL action [intent=reverse_etl availability=implemented write=uninstall_git_sync_installation]; approval: requires plan, preview, approval, and execute; risk: DELETE /git/installations/{installationId} (Uninstall a Git Sync installation) executes a live GitBook API operation.; flags: --installation_id (required)
    uninstall integration apply - Plan and execute the uninstall integration reverse-ETL action [intent=reverse_etl availability=implemented write=uninstall_integration]; approval: requires plan, preview, approval, and execute; risk: DELETE /integrations/{integrationName}/installations/{installationId} (Uninstall an integration) executes a live GitBook API operation.; flags: --installation_id (required), --integration_name (required)
    uninstall integration from site apply - Plan and execute the uninstall integration from site reverse-ETL action [intent=reverse_etl availability=implemented write=uninstall_integration_from_site]; approval: requires plan, preview, approval, and execute; risk: DELETE /integrations/{integrationName}/installations/{installationId}/sites/{siteId} (Uninstall an integration from a site) executes a live GitBook API operation.; flags: --installation_id (required), --integration_name (required), --site_id (required)
    uninstall integration from space apply - Plan and execute the uninstall integration from space reverse-ETL action [intent=reverse_etl availability=implemented write=uninstall_integration_from_space]; approval: requires plan, preview, approval, and execute; risk: DELETE /integrations/{integrationName}/installations/{installationId}/spaces/{spaceId} (Uninstall an integration from a space) executes a live GitBook API operation.; flags: --installation_id (required), --integration_name (required), --space_id (required)
    unpublish integration apply - Plan and execute the unpublish integration reverse-ETL action [intent=reverse_etl availability=implemented write=unpublish_integration]; approval: requires plan, preview, approval, and execute; risk: DELETE /integrations/{integrationName} (Unpublish an integration) executes a live GitBook API operation.; flags: --integration_name (required)
    unpublish site apply - Plan and execute the unpublish site reverse-ETL action [intent=reverse_etl availability=implemented write=unpublish_site]; approval: requires plan, preview, approval, and execute; risk: POST /orgs/{organizationId}/sites/{siteId}/unpublish (Take a site offline by unpublishing it) executes a live GitBook API operation.; flags: --organization_id (required), --site_id (required)
    update change request apply - Plan and execute the update change request reverse-ETL action [intent=reverse_etl availability=implemented write=update_change_request]; approval: requires plan, preview, approval, and execute; risk: POST /spaces/{spaceId}/change-requests/{changeRequestId}/update (Sync a change request with the latest live space content) executes a live GitBook API operation.; flags: --change_request_id (required), --space_id (required)
    update change request by id apply - Plan and execute the update change request by id reverse-ETL action [intent=reverse_etl availability=implemented write=update_change_request_by_id]; approval: requires plan, preview, approval, and execute; risk: PATCH /spaces/{spaceId}/change-requests/{changeRequestId} (Update a change request's subject, description, or status) executes a live GitBook API operation.; flags: --change_request_id (required), --space_id (required)
    update change request content apply - Plan and execute the update change request content reverse-ETL action [intent=reverse_etl availability=not_implemented write=update_change_request_content]; approval: requires plan, preview, approval, and execute; risk: POST /spaces/{spaceId}/change-requests/{changeRequestId}/content (Apply a batch of content changes to a change request) executes a live GitBook API operation.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    update change request conversation apply - Plan and execute the update change request conversation reverse-ETL action [intent=reverse_etl availability=not_implemented write=update_change_request_conversation]; approval: requires plan, preview, approval, and execute; risk: PATCH /spaces/{spaceId}/change-requests/{changeRequestId}/conversations/{conversationId} (Update the title of an AI agent conversation on a change request) executes a live GitBook API operation.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    update collection by id apply - Plan and execute the update collection by id reverse-ETL action [intent=reverse_etl availability=implemented write=update_collection_by_id]; approval: requires plan, preview, approval, and execute; risk: PATCH /collections/{collectionId} (Update a collection) executes a live GitBook API operation.; flags: --collection_id (required)
    update comment in change request apply - Plan and execute the update comment in change request reverse-ETL action [intent=reverse_etl availability=implemented write=update_comment_in_change_request]; approval: requires plan, preview, approval, and execute; risk: PUT /spaces/{spaceId}/change-requests/{changeRequestId}/comments/{commentId} (Update the content or status of a change request comment) executes a live GitBook API operation.; flags: --change_request_id (required), --comment_id (required), --space_id (required)
    update comment in space apply - Plan and execute the update comment in space reverse-ETL action [intent=reverse_etl availability=implemented write=update_comment_in_space]; approval: requires plan, preview, approval, and execute; risk: PUT /spaces/{spaceId}/comments/{commentId} (Update the body or status of a space comment) executes a live GitBook API operation.; flags: --comment_id (required), --space_id (required)
    update comment reply in change request apply - Plan and execute the update comment reply in change request reverse-ETL action [intent=reverse_etl availability=implemented write=update_comment_reply_in_change_request]; approval: requires plan, preview, approval, and execute; risk: PUT /spaces/{spaceId}/change-requests/{changeRequestId}/comments/{commentId}/replies/{commentReplyId} (Update the content of a change request comment reply) executes a live GitBook API operation.; flags: --change_request_id (required), --comment_id (required), --comment_reply_id (required), --space_id (required)
    update comment reply in space apply - Plan and execute the update comment reply in space reverse-ETL action [intent=reverse_etl availability=implemented write=update_comment_reply_in_space]; approval: requires plan, preview, approval, and execute; risk: PUT /spaces/{spaceId}/comments/{commentId}/replies/{commentReplyId} (Update the body of a reply to a space comment) executes a live GitBook API operation.; flags: --comment_id (required), --comment_reply_id (required), --space_id (required)
    update custom font apply - Plan and execute the update custom font reverse-ETL action [intent=reverse_etl availability=implemented write=update_custom_font]; approval: requires plan, preview, approval, and execute; risk: POST /orgs/{organizationId}/fonts/{fontId} (Update a custom font) executes a live GitBook API operation.; flags: --font_id (required), --organization_id (required)
    update git sync installation by id apply - Plan and execute the update git sync installation by id reverse-ETL action [intent=reverse_etl availability=implemented write=update_git_sync_installation_by_id]; approval: requires plan, preview, approval, and execute; risk: PATCH /git/installations/{installationId} (Update a Git Sync installation configuration) executes a live GitBook API operation.; flags: --installation_id (required)
    update glossary entries apply - Plan and execute the update glossary entries reverse-ETL action [intent=reverse_etl availability=not_implemented write=update_glossary_entries]; approval: requires plan, preview, approval, and execute; risk: PUT /orgs/{organizationId}/translations-glossary (Update glossary entries) executes a live GitBook API operation.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    update integration installation apply - Plan and execute the update integration installation reverse-ETL action [intent=reverse_etl availability=implemented write=update_integration_installation]; approval: requires plan, preview, approval, and execute; risk: PATCH /integrations/{integrationName}/installations/{installationId} (Update an integration installation) executes a live GitBook API operation.; flags: --installation_id (required), --integration_name (required)
    update integration site installation apply - Plan and execute the update integration site installation reverse-ETL action [intent=reverse_etl availability=implemented write=update_integration_site_installation]; approval: requires plan, preview, approval, and execute; risk: PATCH /integrations/{integrationName}/installations/{installationId}/sites/{siteId} (Update an integration site installation) executes a live GitBook API operation.; flags: --installation_id (required), --integration_name (required), --site_id (required)
    update integration space installation apply - Plan and execute the update integration space installation reverse-ETL action [intent=reverse_etl availability=implemented write=update_integration_space_installation]; approval: requires plan, preview, approval, and execute; risk: PATCH /integrations/{integrationName}/installations/{installationId}/spaces/{spaceId} (Update an integration space installation) executes a live GitBook API operation.; flags: --installation_id (required), --integration_name (required), --space_id (required)
    update member in organization by id apply - Plan and execute the update member in organization by id reverse-ETL action [intent=reverse_etl availability=implemented write=update_member_in_organization_by_id]; approval: requires plan, preview, approval, and execute; risk: PATCH /orgs/{organizationId}/members/{userId} (Update an organization member) executes a live GitBook API operation.; flags: --organization_id (required), --user_id (required)
    update members in organization team apply - Plan and execute the update members in organization team reverse-ETL action [intent=reverse_etl availability=implemented write=update_members_in_organization_team]; approval: requires plan, preview, approval, and execute; risk: PUT /orgs/{organizationId}/teams/{teamId}/members (Updates members of a team) executes a live GitBook API operation.; flags: --organization_id (required), --team_id (required)
    update open api spec by slug apply - Plan and execute the update open api spec by slug reverse-ETL action [intent=reverse_etl availability=not_implemented write=update_open_api_spec_by_slug]; approval: requires plan, preview, approval, and execute; risk: PATCH /orgs/{organizationId}/openapi/{specSlug} (Update OpenAPI spec visibility) executes a live GitBook API operation.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    update organization agent instructions apply - Plan and execute the update organization agent instructions reverse-ETL action [intent=reverse_etl availability=not_implemented write=update_organization_agent_instructions]; approval: requires plan, preview, approval, and execute; risk: PUT /orgs/{organizationId}/agent-instructions (Update Docs agent instructions for an organization) executes a live GitBook API operation.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    update organization by id apply - Plan and execute the update organization by id reverse-ETL action [intent=reverse_etl availability=implemented write=update_organization_by_id]; approval: requires plan, preview, approval, and execute; risk: PATCH /orgs/{organizationId} (Update an organization) executes a live GitBook API operation.; flags: --organization_id (required)
    update organization invite by id apply - Plan and execute the update organization invite by id reverse-ETL action [intent=reverse_etl availability=implemented write=update_organization_invite_by_id]; approval: requires plan, preview, approval, and execute; risk: PATCH /orgs/{organizationId}/link-invites/{inviteId} (Update an organization invite) executes a live GitBook API operation.; flags: --invite_id (required), --organization_id (required)
    update organization member last seen at apply - Plan and execute the update organization member last seen at reverse-ETL action [intent=reverse_etl availability=implemented write=update_organization_member_last_seen_at]; approval: requires plan, preview, approval, and execute; risk: POST /orgs/{organizationId}/ping (Update an organization member last seen at) executes a live GitBook API operation.; flags: --organization_id (required)
    update organization saml provider apply - Plan and execute the update organization saml provider reverse-ETL action [intent=reverse_etl availability=implemented write=update_organization_saml_provider]; approval: requires plan, preview, approval, and execute; risk: PATCH /orgs/{organizationId}/saml/{samlProviderId} (Update a SAML provider) executes a live GitBook API operation.; flags: --organization_id (required), --saml_provider_id (required)
    update site adaptive schema apply - Plan and execute the update site adaptive schema reverse-ETL action [intent=reverse_etl availability=not_implemented write=update_site_adaptive_schema]; approval: requires plan, preview, approval, and execute; risk: PUT /orgs/{organizationId}/sites/{siteId}/adaptive-schema (Update the visitor attributes JSON schema for an adaptive content site) executes a live GitBook API operation.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    update site ads by id apply - Plan and execute the update site ads by id reverse-ETL action [intent=reverse_etl availability=implemented write=update_site_ads_by_id]; approval: requires plan, preview, approval, and execute; risk: POST /orgs/{organizationId}/sites/{siteId}/ads (Update the advertising settings for a site) executes a live GitBook API operation.; flags: --organization_id (required), --site_id (required)
    update site agent settings by id apply - Plan and execute the update site agent settings by id reverse-ETL action [intent=reverse_etl availability=not_implemented write=update_site_agent_settings_by_id]; approval: requires plan, preview, approval, and execute; risk: PUT /orgs/{organizationId}/sites/{siteId}/agent-settings (Update the AI agent configuration for a site) executes a live GitBook API operation.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    update site by id apply - Plan and execute the update site by id reverse-ETL action [intent=reverse_etl availability=implemented write=update_site_by_id]; approval: requires plan, preview, approval, and execute; risk: PATCH /orgs/{organizationId}/sites/{siteId} (Update the properties of a documentation site) executes a live GitBook API operation.; flags: --organization_id (required), --site_id (required)
    update site channel by id apply - Plan and execute the update site channel by id reverse-ETL action [intent=reverse_etl availability=implemented write=update_site_channel_by_id]; approval: requires plan, preview, approval, and execute; risk: PATCH /orgs/{organizationId}/sites/{siteId}/channels/{siteChannelId} (Update a GitBook Agent channel for a site) executes a live GitBook API operation.; flags: --organization_id (required), --site_channel_id (required), --site_id (required)
    update site context connection by id apply - Plan and execute the update site context connection by id reverse-ETL action [intent=reverse_etl availability=implemented write=update_site_context_connection_by_id]; approval: requires plan, preview, approval, and execute; risk: PATCH /orgs/{organizationId}/sites/{siteId}/context-connections/{siteContextConnectionId} (Update a context connection) executes a live GitBook API operation.; flags: --organization_id (required), --site_context_connection_id (required), --site_id (required)
    update site customization by id apply - Plan and execute the update site customization by id reverse-ETL action [intent=reverse_etl availability=not_implemented write=update_site_customization_by_id]; approval: requires plan, preview, approval, and execute; risk: PUT /orgs/{organizationId}/sites/{siteId}/customization (Update the branding and visual customization settings for a site) executes a live GitBook API operation.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    update site finding by id apply - Plan and execute the update site finding by id reverse-ETL action [intent=reverse_etl availability=not_implemented write=update_site_finding_by_id]; approval: requires plan, preview, approval, and execute; risk: PATCH /orgs/{organizationId}/sites/{siteId}/findings/{siteFindingId} (Update a site finding) executes a live GitBook API operation.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    update site mcp server by id apply - Plan and execute the update site mcp server by id reverse-ETL action [intent=reverse_etl availability=implemented write=update_site_mcp_server_by_id]; approval: requires plan, preview, approval, and execute; risk: PATCH /orgs/{organizationId}/sites/{siteId}/mcp-servers/{siteMcpServerId} (Update an MCP server configuration for a site) executes a live GitBook API operation.; flags: --organization_id (required), --site_id (required), --site_mcp_server_id (required)
    update site publishing auth by id apply - Plan and execute the update site publishing auth by id reverse-ETL action [intent=reverse_etl availability=implemented write=update_site_publishing_auth_by_id]; approval: requires plan, preview, approval, and execute; risk: PATCH /orgs/{organizationId}/sites/{siteId}/publishing/auth (Update the published content authentication configuration for a site) executes a live GitBook API operation.; flags: --organization_id (required), --site_id (required)
    update site redirect by id apply - Plan and execute the update site redirect by id reverse-ETL action [intent=reverse_etl availability=implemented write=update_site_redirect_by_id]; approval: requires plan, preview, approval, and execute; risk: PATCH /orgs/{organizationId}/sites/{siteId}/redirects/{siteRedirectId} (Update a URL redirect rule for a site) executes a live GitBook API operation.; flags: --organization_id (required), --site_id (required), --site_redirect_id (required)
    update site section by id apply - Plan and execute the update site section by id reverse-ETL action [intent=reverse_etl availability=implemented write=update_site_section_by_id]; approval: requires plan, preview, approval, and execute; risk: PATCH /orgs/{organizationId}/sites/{siteId}/sections/{siteSectionId} (Update a navigation section in a site) executes a live GitBook API operation.; flags: --organization_id (required), --site_id (required), --site_section_id (required)
    update site section group by id apply - Plan and execute the update site section group by id reverse-ETL action [intent=reverse_etl availability=implemented write=update_site_section_group_by_id]; approval: requires plan, preview, approval, and execute; risk: PATCH /orgs/{organizationId}/sites/{siteId}/section-groups/{siteSectionGroupId} (Update a section group in a site's navigation structure) executes a live GitBook API operation.; flags: --organization_id (required), --site_id (required), --site_section_group_id (required)
    update site share link by id apply - Plan and execute the update site share link by id reverse-ETL action [intent=reverse_etl availability=implemented write=update_site_share_link_by_id]; approval: requires plan, preview, approval, and execute; risk: PATCH /orgs/{organizationId}/sites/{siteId}/share-links/{shareLinkId} (Update a private share link for a site) executes a live GitBook API operation.; flags: --organization_id (required), --share_link_id (required), --site_id (required)
    update site space by id apply - Plan and execute the update site space by id reverse-ETL action [intent=reverse_etl availability=implemented write=update_site_space_by_id]; approval: requires plan, preview, approval, and execute; risk: PATCH /orgs/{organizationId}/sites/{siteId}/site-spaces/{siteSpaceId} (Update a space linked to a site) executes a live GitBook API operation.; flags: --organization_id (required), --site_id (required), --site_space_id (required)
    update site topic by id apply - Plan and execute the update site topic by id reverse-ETL action [intent=reverse_etl availability=not_implemented write=update_site_topic_by_id]; approval: requires plan, preview, approval, and execute; risk: PATCH /orgs/{organizationId}/sites/{siteId}/topics/{siteTopicId} (Update a topic) executes a live GitBook API operation.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    update space by id apply - Plan and execute the update space by id reverse-ETL action [intent=reverse_etl availability=implemented write=update_space_by_id]; approval: requires plan, preview, approval, and execute; risk: PATCH /spaces/{spaceId} (Update a space's title, icon, or settings) executes a live GitBook API operation.; flags: --space_id (required)
    update team in organization by id apply - Plan and execute the update team in organization by id reverse-ETL action [intent=reverse_etl availability=not_implemented write=update_team_in_organization_by_id]; approval: requires plan, preview, approval, and execute; risk: PATCH /orgs/{organizationId}/teams/{teamId} (Update a team) executes a live GitBook API operation.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    update team permission in collection apply - Plan and execute the update team permission in collection reverse-ETL action [intent=reverse_etl availability=implemented write=update_team_permission_in_collection]; approval: requires plan, preview, approval, and execute; risk: PATCH /collections/{collectionId}/permissions/teams/{teamId} (Update an org team's permission in a collection) executes a live GitBook API operation.; flags: --collection_id (required), --team_id (required)
    update team permission in site apply - Plan and execute the update team permission in site reverse-ETL action [intent=reverse_etl availability=implemented write=update_team_permission_in_site]; approval: requires plan, preview, approval, and execute; risk: PATCH /orgs/{organizationId}/sites/{siteId}/permissions/teams/{teamId} (Update an org team's permission in a site) executes a live GitBook API operation.; flags: --organization_id (required), --site_id (required), --team_id (required)
    update team permission in space apply - Plan and execute the update team permission in space reverse-ETL action [intent=reverse_etl availability=implemented write=update_team_permission_in_space]; approval: requires plan, preview, approval, and execute; risk: PATCH /spaces/{spaceId}/permissions/teams/{teamId} (Update an org team's permission in a space) executes a live GitBook API operation.; flags: --space_id (required), --team_id (required)
    update translation apply - Plan and execute the update translation reverse-ETL action [intent=reverse_etl availability=not_implemented write=update_translation]; approval: requires plan, preview, approval, and execute; risk: PUT /orgs/{organizationId}/translations/{translationId} (Update a translation) executes a live GitBook API operation.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    update user by id apply - Plan and execute the update user by id reverse-ETL action [intent=reverse_etl availability=implemented write=update_user_by_id]; approval: requires plan, preview, approval, and execute; risk: PATCH /users/{userId} (Update a user by its ID) executes a live GitBook API operation.; flags: --user_id (required)
    update user permission in collection apply - Plan and execute the update user permission in collection reverse-ETL action [intent=reverse_etl availability=implemented write=update_user_permission_in_collection]; approval: requires plan, preview, approval, and execute; risk: PATCH /collections/{collectionId}/permissions/users/{userId} (Update a collection user permission) executes a live GitBook API operation.; flags: --collection_id (required), --user_id (required)
    update user permission in site apply - Plan and execute the update user permission in site reverse-ETL action [intent=reverse_etl availability=implemented write=update_user_permission_in_site]; approval: requires plan, preview, approval, and execute; risk: PATCH /orgs/{organizationId}/sites/{siteId}/permissions/users/{userId} (Update site user permissions) executes a live GitBook API operation.; flags: --organization_id (required), --site_id (required), --user_id (required)
    update user permission in space apply - Plan and execute the update user permission in space reverse-ETL action [intent=reverse_etl availability=implemented write=update_user_permission_in_space]; approval: requires plan, preview, approval, and execute; risk: PATCH /spaces/{spaceId}/permissions/users/{userId} (Update space user permissions) executes a live GitBook API operation.; flags: --space_id (required), --user_id (required)
    users list - Run the users ETL stream [intent=etl availability=implemented stream=users]

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
