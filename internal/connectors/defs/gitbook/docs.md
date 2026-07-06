# Overview

Reads 185 GitBook REST resources and executes 170 JSON/no-body GitBook mutations through the GitBook
API.

Readable streams: `users`, `organizations`, `org_members`, `content`, `get_api_information`,
`get_user_by_id`, `get_space_by_id`, `get_embed_by_url_in_space`, `search_space_content`,
`get_space_git_info`, `list_user_permissions_in_space`, `list_team_permissions_in_space`,
`get_current_revision`, `list_files`, `get_file_by_id`, `list_space_file_backlinks`,
`get_page_by_id`, `list_page_links_in_space`, `list_space_page_backlinks`,
`list_space_page_meta_links`, `get_page_by_path`, `get_reusable_content_by_id`,
`get_document_by_id`, `list_change_requests_for_space`, `get_change_request_by_id`,
`get_reviews_by_change_request_id`, `get_change_request_review_by_id`,
`get_requested_reviewers_by_change_request_id`, `list_change_request_conversations`,
`list_change_request_links`, `list_comments_in_change_request`, `get_comment_in_change_request`,
`list_comment_replies_in_change_request`, `get_comment_reply_in_change_request`,
`get_contributors_by_change_request_id`, `get_revision_of_change_request_by_id`,
`list_pages_in_change_request`, `list_files_in_change_request_by_id`,
`get_file_in_change_request_by_id`, `list_change_request_file_backlinks`,
`get_page_in_change_request_by_id`, `list_page_links_in_change_request`,
`list_change_request_page_backlinks`, `list_change_request_page_meta_links`,
`get_reusable_content_in_change_request_by_id`, `get_change_request_changes`,
`get_change_request_pdf`, `get_revision_by_id`, `get_revision_semantic_changes`,
`list_pages_in_revision_by_id`, `list_files_in_revision_by_id`, `get_file_in_revision_by_id`,
`get_page_in_revision_by_id`, `get_page_document_in_revision_by_id`, `get_page_in_revision_by_path`,
`list_revision_page_meta_links`, `get_page_in_change_request_by_path`,
`get_reusable_content_in_revision_by_id`, `get_reusable_content_document_in_revision_by_id`,
`list_comments_in_space`, `get_comment_in_space`, `list_comment_replies_in_space`,
`get_comment_reply_in_space`, `list_commenters_in_space`, `list_commenters_in_change_request`,
`list_permissions_aggregate_in_space`, `list_space_integrations`, `list_space_integrations_blocks`,
`get_space_pdf`, `list_space_links`, `get_collection_by_id`, `list_spaces_in_collection_by_id`,
`list_team_permissions_in_collection`, `list_user_permissions_in_collection`,
`list_permissions_aggregate_in_collection`, `list_integrations`, `get_integration_by_name`,
`list_integration_installations`, `list_integration_events`, `get_integration_event`,
`list_integration_space_installations`, `list_integration_site_installations`,
`render_integration_ui_with_get`, `get_integration_installation_by_id`,
`list_integration_installation_spaces`, `get_integration_space_installation`,
`list_integration_installation_sites`, `get_integration_site_installation`,
`get_organization_by_id`, `get_member_in_organization_by_id`, `list_spaces_for_organization_member`,
`list_teams_for_organization_member`, `list_teams_in_organization_by_id`,
`get_team_in_organization_by_id`, `list_team_members_in_organization_by_id`,
`list_organization_invite_links`, `get_organization_invite_link`, `search_organization_content`,
`list_change_requests_for_organization`, `list_spaces_in_organization_by_id`,
`list_collections_in_organization_by_id`, `list_organization_integrations`,
`get_organization_integration_status`, `list_organization_installations`,
`list_organization_integrations_status`, `list_saml_providers_in_organization_by_id`,
`get_organization_saml_provider_by_id`, `list_sso_provider_logins_in_organization`,
`get_recommended_questions_in_organization`, `list_open_api_specs`, `get_open_api_spec_by_slug`,
`list_open_api_spec_versions`, `get_latest_open_api_spec_version`,
`get_latest_open_api_spec_version_content`, `get_open_api_spec_version_by_id`,
`get_open_api_spec_version_content_by_id`, `get_organization_agent_instructions`,
`list_translations`, `get_translation`, `list_glossary_entries`, `get_glossary_entry`,
`list_custom_fonts`, `get_custom_font`, `list_sites`, `get_site_by_id`,
`list_site_git_sync_installations`, `get_site_adaptive_schema`,
`list_site_adaptive_template_conditions`, `get_published_content_site`, `list_site_share_links`,
`get_site_structure`, `get_site_publishing_auth_by_id`, `get_site_publishing_preview_by_id`,
`get_site_customization_by_id`, `list_site_integration_scripts`, `list_site_integrations`,
`list_site_spaces`, `list_site_section_groups`, `list_site_sections`, `list_site_context_records`,
`get_site_context_record_by_id`, `list_site_scans`, `get_site_scan_by_id`, `list_site_findings`,
`get_site_finding_by_id`, `list_change_requests_for_site_finding`, `list_pages_for_site_finding`,
`list_questions_for_site_finding`, `list_records_for_site_finding`, `list_site_context_connections`,
`get_site_context_connection_by_id`, `list_site_topics`, `get_site_topic_by_id`,
`list_site_questions`, `get_site_question_by_id`, `list_site_question_sources`,
`get_site_question_stats`, `list_site_question_answers`, `get_site_question_answer_by_id`,
`get_site_question_answer_thread_by_id`, `list_site_question_answer_sources`,
`get_site_space_customization_by_id`, `list_permissions_aggregate_in_site`,
`list_user_permissions_in_site`, `list_team_permissions_in_site`, `get_site_agent_settings_by_id`,
`list_site_visitor_segments`, `list_site_redirects`, `get_site_redirect_by_source`,
`list_site_mcp_servers`, `get_site_mcp_server_by_id`, `list_site_channels`,
`get_site_channel_by_id`, `get_subdomain`, `get_custom_hostname`,
`get_organizations_for_email_domain`, `ads_list_sites`, `get_content_by_url`, `get_embed_by_url`,
`get_published_content_by_url`, `get_git_sync_installation_by_id`,
`list_git_hub_repositories_for_git_sync_installation`,
`list_git_hub_repo_branches_for_git_sync_installation`,
`list_git_lab_projects_for_git_sync_installation`,
`list_git_lab_project_branches_for_git_sync_installation`.

Write actions: `create_user_notifications_token`, `update_user_by_id`, `update_space_by_id`,
`delete_space_by_id`, `duplicate_space`, `restore_space`, `move_space`, `import_git_repository`,
`export_to_git_repository`, `delete_git_installation`, `invite_to_space`,
`update_team_permission_in_space`, `remove_team_from_space`, `update_user_permission_in_space`,
`remove_user_from_space`, `apply_template_to_space`, `get_computed_document`,
`get_computed_revision`, `create_change_request`, `update_change_request_by_id`,
`merge_change_request`, `update_change_request`, `submit_change_request_review`,
`request_reviewers_for_change_request`, `remove_requested_reviewer_from_change_request`,
`update_change_request_conversation`, `delete_change_request_conversation`,
`post_comment_in_change_request`, `update_comment_in_change_request`,
`delete_comment_in_change_request`, `post_comment_reply_in_change_request`,
`update_comment_reply_in_change_request`, `delete_comment_reply_in_change_request`,
`update_change_request_content`, `post_comment_in_space`, `update_comment_in_space`,
`delete_comment_in_space`, `post_comment_reply_in_space`, `update_comment_reply_in_space`,
`delete_comment_reply_in_space`, `update_collection_by_id`, `delete_collection_by_id`,
`move_collection`, `transfer_collection`, `invite_to_collection`,
`update_team_permission_in_collection`, `remove_team_from_collection`,
`update_user_permission_in_collection`, `remove_user_from_collection`, `publish_integration`,
`unpublish_integration`, `install_integration`, `set_integration_development_mode`,
`disable_integration_development_mode`, `render_integration_ui_with_post`, `queue_integration_task`,
`update_integration_installation`, `uninstall_integration`, `create_integration_installation_token`,
`install_integration_on_space`, `update_integration_space_installation`,
`uninstall_integration_from_space`, `install_integration_on_site`,
`update_integration_site_installation`, `uninstall_integration_from_site`,
`update_organization_by_id`, `update_member_in_organization_by_id`,
`remove_member_from_organization_by_id`, `update_organization_member_last_seen_at`,
`set_user_as_sso_member_for_organization`, `create_organization_team`,
`update_team_in_organization_by_id`, `remove_team_from_organization_by_id`,
`update_members_in_organization_team`, `add_member_to_organization_team_by_id`,
`delete_member_from_organization_team_by_id`, `invite_users_to_organization`,
`join_organization_with_invite`, `create_organization_invite`, `update_organization_invite_by_id`,
and 90 more.

Service API documentation: https://gitbook.com/docs/developers/gitbook-api/api-reference.

## Auth setup

Connection fields:

- `access_token` (required, secret, string); GitBook API access token. Used only for Bearer auth;
  never logged.
- `account_name` (optional, string); GitBook path parameter accountName; required when reading
  streams scoped by accountName.
- `base_url` (optional, string); default `https://api.gitbook.com/v1`; format `uri`; GitBook API
  base URL override for tests or proxies.
- `change_request_id` (optional, string); GitBook path parameter changeRequestId; required when
  reading streams scoped by changeRequestId.
- `collection_id` (optional, string); GitBook path parameter collectionId; required when reading
  streams scoped by collectionId.
- `comment_id` (optional, string); GitBook path parameter commentId; required when reading streams
  scoped by commentId.
- `comment_reply_id` (optional, string); GitBook path parameter commentReplyId; required when
  reading streams scoped by commentReplyId.
- `conversation_id` (optional, string); GitBook path parameter conversationId; required when reading
  streams scoped by conversationId.
- `document_id` (optional, string); GitBook path parameter documentId; required when reading streams
  scoped by documentId.
- `email_domain` (optional, string); GitBook path parameter emailDomain; required when reading
  streams scoped by emailDomain.
- `event_id` (optional, string); GitBook path parameter eventId; required when reading streams
  scoped by eventId.
- `file_id` (optional, string); GitBook path parameter fileId; required when reading streams scoped
  by fileId.
- `font_id` (optional, string); GitBook path parameter fontId; required when reading streams scoped
  by fontId.
- `glossary_entry_id` (optional, string); GitBook path parameter glossaryEntryId; required when
  reading streams scoped by glossaryEntryId.
- `hostname` (optional, string); GitBook path parameter hostname; required when reading streams
  scoped by hostname.
- `import_run_id` (optional, string); GitBook path parameter importRunId; required when reading
  streams scoped by importRunId.
- `installation_id` (optional, string); GitBook path parameter installationId; required when reading
  streams scoped by installationId.
- `integration_name` (optional, string); GitBook path parameter integrationName; required when
  reading streams scoped by integrationName.
- `invite_id` (optional, string); GitBook path parameter inviteId; required when reading streams
  scoped by inviteId.
- `organization_id` (optional, string); GitBook path parameter organizationId; required when reading
  streams scoped by organizationId.
- `page_id` (optional, string); GitBook path parameter pageId; required when reading streams scoped
  by pageId.
- `page_path` (optional, string); GitBook path parameter pagePath; required when reading streams
  scoped by pagePath.
- `project_id` (optional, string); GitBook path parameter projectId; required when reading streams
  scoped by projectId.
- `query` (optional, string); GitBook required query parameter query; required for streams that use
  it.
- `repository_name` (optional, string); GitBook path parameter repositoryName; required when reading
  streams scoped by repositoryName.
- `request` (optional, string); GitBook required query parameter request; required for streams that
  use it.
- `reusable_content_id` (optional, string); GitBook path parameter reusableContentId; required when
  reading streams scoped by reusableContentId.
- `review_id` (optional, string); GitBook path parameter reviewId; required when reading streams
  scoped by reviewId.
- `revision_id` (optional, string); GitBook path parameter revisionId; required when reading streams
  scoped by revisionId.
- `saml_provider_id` (optional, string); GitBook path parameter samlProviderId; required when
  reading streams scoped by samlProviderId.
- `share_link_id` (optional, string); GitBook path parameter shareLinkId; required when reading
  streams scoped by shareLinkId.
- `site_channel_id` (optional, string); GitBook path parameter siteChannelId; required when reading
  streams scoped by siteChannelId.
- `site_context_connection_id` (optional, string); GitBook path parameter siteContextConnectionId;
  required when reading streams scoped by siteContextConnectionId.
- `site_context_record_id` (optional, string); GitBook path parameter siteContextRecordId; required
  when reading streams scoped by siteContextRecordId.
- `site_finding_id` (optional, string); GitBook path parameter siteFindingId; required when reading
  streams scoped by siteFindingId.
- `site_id` (optional, string); GitBook path parameter siteId; required when reading streams scoped
  by siteId.
- `site_mcp_server_id` (optional, string); GitBook path parameter siteMcpServerId; required when
  reading streams scoped by siteMcpServerId.
- `site_question_answer_id` (optional, string); GitBook path parameter siteQuestionAnswerId;
  required when reading streams scoped by siteQuestionAnswerId.
- `site_question_id` (optional, string); GitBook path parameter siteQuestionId; required when
  reading streams scoped by siteQuestionId.
- `site_redirect_id` (optional, string); GitBook path parameter siteRedirectId; required when
  reading streams scoped by siteRedirectId.
- `site_scan_id` (optional, string); GitBook path parameter siteScanId; required when reading
  streams scoped by siteScanId.
- `site_section_group_id` (optional, string); GitBook path parameter siteSectionGroupId; required
  when reading streams scoped by siteSectionGroupId.
- `site_section_id` (optional, string); GitBook path parameter siteSectionId; required when reading
  streams scoped by siteSectionId.
- `site_space_id` (optional, string); GitBook path parameter siteSpaceId; required when reading
  streams scoped by siteSpaceId.
- `site_topic_id` (optional, string); GitBook path parameter siteTopicId; required when reading
  streams scoped by siteTopicId.
- `source` (optional, string); GitBook required query parameter source; required for streams that
  use it.
- `space_id` (optional, string); GitBook path parameter spaceId; required when reading streams
  scoped by spaceId.
- `spec_slug` (optional, string); GitBook path parameter specSlug; required when reading streams
  scoped by specSlug.
- `subdomain` (optional, string); GitBook path parameter subdomain; required when reading streams
  scoped by subdomain.
- `team_id` (optional, string); GitBook path parameter teamId; required when reading streams scoped
  by teamId.
- `translation_id` (optional, string); GitBook path parameter translationId; required when reading
  streams scoped by translationId.
- `url` (optional, string); GitBook required query parameter url; required for streams that use it.
- `user_id` (optional, string); GitBook path parameter userId; required when reading streams scoped
  by userId.
- `version_id` (optional, string); GitBook path parameter versionId; required when reading streams
  scoped by versionId.

Secret fields are redacted in logs and write previews: `access_token`.

Default configuration values: `base_url=https://api.gitbook.com/v1`.

Authentication behavior:

- Bearer token authentication using `secrets.access_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/user`.

## Streams notes

Default pagination: cursor pagination; cursor parameter `page`; next token from `next.page`; page
size 50.

Pagination by stream: cursor: `organizations`, `org_members`, `content`, `search_space_content`,
`list_user_permissions_in_space`, `list_team_permissions_in_space`, `get_current_revision`,
`list_files`, `list_space_file_backlinks`, `get_page_by_id`, `list_page_links_in_space`,
`list_space_page_backlinks`, `list_space_page_meta_links`, `get_page_by_path`, `get_document_by_id`,
`list_change_requests_for_space`, `get_change_request_by_id`, `get_reviews_by_change_request_id`,
`get_requested_reviewers_by_change_request_id`, `list_change_request_conversations`,
`list_change_request_links`, `list_comments_in_change_request`, `get_comment_in_change_request`,
`list_comment_replies_in_change_request`, `get_comment_reply_in_change_request`,
`get_contributors_by_change_request_id`, `get_revision_of_change_request_by_id`,
`list_pages_in_change_request`, `list_files_in_change_request_by_id`,
`list_change_request_file_backlinks`, `get_page_in_change_request_by_id`,
`list_page_links_in_change_request`, `list_change_request_page_backlinks`,
`list_change_request_page_meta_links`, `get_change_request_changes`, `get_revision_by_id`,
`get_revision_semantic_changes`, `list_pages_in_revision_by_id`, `list_files_in_revision_by_id`,
`get_page_in_revision_by_id`, `get_page_document_in_revision_by_id`, `get_page_in_revision_by_path`,
`list_revision_page_meta_links`, `get_page_in_change_request_by_path`,
`get_reusable_content_document_in_revision_by_id`, `list_comments_in_space`, `get_comment_in_space`,
`list_comment_replies_in_space`, `get_comment_reply_in_space`, `list_commenters_in_space`,
`list_commenters_in_change_request`, `list_permissions_aggregate_in_space`,
`list_space_integrations`, `list_space_links`, `list_spaces_in_collection_by_id`,
`list_team_permissions_in_collection`, `list_user_permissions_in_collection`,
`list_permissions_aggregate_in_collection`, `list_integrations`, `get_integration_by_name`, and 74
more; none: `users`, `get_api_information`, `get_user_by_id`, `get_space_by_id`,
`get_embed_by_url_in_space`, `get_space_git_info`, `get_file_by_id`, `get_reusable_content_by_id`,
`get_change_request_review_by_id`, `get_file_in_change_request_by_id`,
`get_reusable_content_in_change_request_by_id`, `get_change_request_pdf`,
`get_file_in_revision_by_id`, `get_reusable_content_in_revision_by_id`,
`list_space_integrations_blocks`, `get_space_pdf`, `get_collection_by_id`, `get_integration_event`,
`render_integration_ui_with_get`, `get_member_in_organization_by_id`,
`get_team_in_organization_by_id`, `get_organization_invite_link`,
`get_organization_integration_status`, `get_organization_saml_provider_by_id`,
`get_latest_open_api_spec_version`, `get_latest_open_api_spec_version_content`,
`get_open_api_spec_version_by_id`, `get_open_api_spec_version_content_by_id`,
`get_organization_agent_instructions`, `get_translation`, `get_glossary_entry`,
`get_site_adaptive_schema`, `get_site_publishing_auth_by_id`, `get_site_publishing_preview_by_id`,
`list_site_integration_scripts`, `get_site_scan_by_id`, `get_site_finding_by_id`,
`get_site_context_connection_by_id`, `get_site_topic_by_id`, `get_site_question_by_id`,
`get_site_question_answer_thread_by_id`, `get_site_agent_settings_by_id`,
`get_site_redirect_by_source`, `get_site_mcp_server_by_id`, `get_site_channel_by_id`,
`get_subdomain`, `get_custom_hostname`, `get_content_by_url`, `get_embed_by_url`,
`get_published_content_by_url`, `get_git_sync_installation_by_id`.

- `users`: GET `/user` - records at response root; computed output fields `display_name`,
  `photo_url`.
- `organizations`: GET `/orgs` - records path `items`; query `limit`=`50`; cursor pagination; cursor
  parameter `page`; next token from `next.page`; page size 50; computed output fields `created_at`,
  `url`.
- `org_members`: GET `/orgs/{{ config.organization_id }}/members` - records path `items`; query
  `limit`=`50`; cursor pagination; cursor parameter `page`; next token from `next.page`; page size
  50; computed output fields `display_name`, `email`, `id`.
- `content`: GET `/spaces/{{ config.space_id }}/content/pages` - records path `pages`; query
  `limit`=`50`; cursor pagination; cursor parameter `page`; next token from `next.page`; page size
  50.
- `get_api_information`: GET `/` - records path `.`; emits passthrough records.
- `get_user_by_id`: GET `/users/{{ config.user_id }}` - records path `.`; emits passthrough records.
- `get_space_by_id`: GET `/spaces/{{ config.space_id }}` - records path `.`; emits passthrough
  records.
- `get_embed_by_url_in_space`: GET `/spaces/{{ config.space_id }}/embed` - records path `.`; query
  `url`=`{{ config.url }}`; emits passthrough records.
- `search_space_content`: GET `/spaces/{{ config.space_id }}/search` - records path `items`; query
  `limit`=`50`; `query`=`{{ config.query }}`; cursor pagination; cursor parameter `page`; next token
  from `next.page`; page size 50; emits passthrough records.
- `get_space_git_info`: GET `/spaces/{{ config.space_id }}/git/info` - records path `.`; emits
  passthrough records.
- `list_user_permissions_in_space`: GET `/spaces/{{ config.space_id }}/permissions/users` - records
  path `items`; query `limit`=`50`; cursor pagination; cursor parameter `page`; next token from
  `next.page`; page size 50; emits passthrough records.
- `list_team_permissions_in_space`: GET `/spaces/{{ config.space_id }}/permissions/teams` - records
  path `items`; query `limit`=`50`; cursor pagination; cursor parameter `page`; next token from
  `next.page`; page size 50; emits passthrough records.
- `get_current_revision`: GET `/spaces/{{ config.space_id }}/content` - records path `pages`; cursor
  pagination; cursor parameter `page`; next token from `next.page`; page size 50; emits passthrough
  records.
- `list_files`: GET `/spaces/{{ config.space_id }}/content/files` - records path `items`; query
  `limit`=`50`; cursor pagination; cursor parameter `page`; next token from `next.page`; page size
  50; emits passthrough records.
- `get_file_by_id`: GET `/spaces/{{ config.space_id }}/content/files/{{ config.file_id }}` - records
  path `.`; emits passthrough records.
- `list_space_file_backlinks`: GET `/spaces/{{ config.space_id }}/content/files/{{ config.file_id
  }}/backlinks` - records path `items`; query `limit`=`50`; cursor pagination; cursor parameter
  `page`; next token from `next.page`; page size 50; emits passthrough records.
- `get_page_by_id`: GET `/spaces/{{ config.space_id }}/content/page/{{ config.page_id }}` - records
  path `pages`; cursor pagination; cursor parameter `page`; next token from `next.page`; page size
  50; emits passthrough records.
- `list_page_links_in_space`: GET `/spaces/{{ config.space_id }}/content/page/{{ config.page_id
  }}/links` - records path `items`; query `limit`=`50`; cursor pagination; cursor parameter `page`;
  next token from `next.page`; page size 50; emits passthrough records.
- `list_space_page_backlinks`: GET `/spaces/{{ config.space_id }}/content/page/{{ config.page_id
  }}/backlinks` - records path `items`; query `limit`=`50`; cursor pagination; cursor parameter
  `page`; next token from `next.page`; page size 50; emits passthrough records.
- `list_space_page_meta_links`: GET `/spaces/{{ config.space_id }}/content/page/{{ config.page_id
  }}/meta-links` - records path `alternates`; cursor pagination; cursor parameter `page`; next token
  from `next.page`; page size 50; emits passthrough records.
- `get_page_by_path`: GET `/spaces/{{ config.space_id }}/content/path/{{ config.page_path }}` -
  records path `pages`; cursor pagination; cursor parameter `page`; next token from `next.page`;
  page size 50; emits passthrough records.
- `get_reusable_content_by_id`: GET `/spaces/{{ config.space_id }}/content/reusable-contents/{{
  config.reusable_content_id }}` - records path `.`; emits passthrough records.
- `get_document_by_id`: GET `/spaces/{{ config.space_id }}/documents/{{ config.document_id }}` -
  records path `nodes`; cursor pagination; cursor parameter `page`; next token from `next.page`;
  page size 50; emits passthrough records.
- `list_change_requests_for_space`: GET `/spaces/{{ config.space_id }}/change-requests` - records
  path `items`; query `limit`=`50`; cursor pagination; cursor parameter `page`; next token from
  `next.page`; page size 50; emits passthrough records.
- `get_change_request_by_id`: GET `/spaces/{{ config.space_id }}/change-requests/{{
  config.change_request_id }}` - records path `links`; cursor pagination; cursor parameter `page`;
  next token from `next.page`; page size 50; emits passthrough records.
- `get_reviews_by_change_request_id`: GET `/spaces/{{ config.space_id }}/change-requests/{{
  config.change_request_id }}/reviews` - records path `items`; query `limit`=`50`; cursor
  pagination; cursor parameter `page`; next token from `next.page`; page size 50; emits passthrough
  records.
- `get_change_request_review_by_id`: GET `/spaces/{{ config.space_id }}/change-requests/{{
  config.change_request_id }}/reviews/{{ config.review_id }}` - records path `.`; emits passthrough
  records.
- `get_requested_reviewers_by_change_request_id`: GET `/spaces/{{ config.space_id
  }}/change-requests/{{ config.change_request_id }}/requested-reviewers` - records path `items`;
  query `limit`=`50`; cursor pagination; cursor parameter `page`; next token from `next.page`; page
  size 50; emits passthrough records.
- `list_change_request_conversations`: GET `/spaces/{{ config.space_id }}/change-requests/{{
  config.change_request_id }}/conversations` - records path `items`; query `limit`=`50`; cursor
  pagination; cursor parameter `page`; next token from `next.page`; page size 50; emits passthrough
  records.
- `list_change_request_links`: GET `/spaces/{{ config.space_id }}/change-requests/{{
  config.change_request_id }}/links` - records path `items`; query `limit`=`50`; cursor pagination;
  cursor parameter `page`; next token from `next.page`; page size 50; emits passthrough records.
- `list_comments_in_change_request`: GET `/spaces/{{ config.space_id }}/change-requests/{{
  config.change_request_id }}/comments` - records path `items`; query `limit`=`50`; cursor
  pagination; cursor parameter `page`; next token from `next.page`; page size 50; emits passthrough
  records.
- `get_comment_in_change_request`: GET `/spaces/{{ config.space_id }}/change-requests/{{
  config.change_request_id }}/comments/{{ config.comment_id }}` - records path `reactions`; cursor
  pagination; cursor parameter `page`; next token from `next.page`; page size 50; emits passthrough
  records.
- `list_comment_replies_in_change_request`: GET `/spaces/{{ config.space_id }}/change-requests/{{
  config.change_request_id }}/comments/{{ config.comment_id }}/replies` - records path `items`;
  query `limit`=`50`; cursor pagination; cursor parameter `page`; next token from `next.page`; page
  size 50; emits passthrough records.
- `get_comment_reply_in_change_request`: GET `/spaces/{{ config.space_id }}/change-requests/{{
  config.change_request_id }}/comments/{{ config.comment_id }}/replies/{{ config.comment_reply_id
  }}` - records path `reactions`; cursor pagination; cursor parameter `page`; next token from
  `next.page`; page size 50; emits passthrough records.
- `get_contributors_by_change_request_id`: GET `/spaces/{{ config.space_id }}/change-requests/{{
  config.change_request_id }}/contributors` - records path `items`; cursor pagination; cursor
  parameter `page`; next token from `next.page`; page size 50; emits passthrough records.
- `get_revision_of_change_request_by_id`: GET `/spaces/{{ config.space_id }}/change-requests/{{
  config.change_request_id }}/content` - records path `pages`; cursor pagination; cursor parameter
  `page`; next token from `next.page`; page size 50; emits passthrough records.
- `list_pages_in_change_request`: GET `/spaces/{{ config.space_id }}/change-requests/{{
  config.change_request_id }}/content/pages` - records path `pages`; cursor pagination; cursor
  parameter `page`; next token from `next.page`; page size 50; emits passthrough records.
- `list_files_in_change_request_by_id`: GET `/spaces/{{ config.space_id }}/change-requests/{{
  config.change_request_id }}/content/files` - records path `items`; query `limit`=`50`; cursor
  pagination; cursor parameter `page`; next token from `next.page`; page size 50; emits passthrough
  records.
- `get_file_in_change_request_by_id`: GET `/spaces/{{ config.space_id }}/change-requests/{{
  config.change_request_id }}/content/files/{{ config.file_id }}` - records path `.`; emits
  passthrough records.
- `list_change_request_file_backlinks`: GET `/spaces/{{ config.space_id }}/change-requests/{{
  config.change_request_id }}/content/files/{{ config.file_id }}/backlinks` - records path `items`;
  query `limit`=`50`; cursor pagination; cursor parameter `page`; next token from `next.page`; page
  size 50; emits passthrough records.
- `get_page_in_change_request_by_id`: GET `/spaces/{{ config.space_id }}/change-requests/{{
  config.change_request_id }}/content/page/{{ config.page_id }}` - records path `pages`; cursor
  pagination; cursor parameter `page`; next token from `next.page`; page size 50; emits passthrough
  records.
- `list_page_links_in_change_request`: GET `/spaces/{{ config.space_id }}/change-requests/{{
  config.change_request_id }}/content/page/{{ config.page_id }}/links` - records path `items`; query
  `limit`=`50`; cursor pagination; cursor parameter `page`; next token from `next.page`; page size
  50; emits passthrough records.
- `list_change_request_page_backlinks`: GET `/spaces/{{ config.space_id }}/change-requests/{{
  config.change_request_id }}/content/page/{{ config.page_id }}/backlinks` - records path `items`;
  query `limit`=`50`; cursor pagination; cursor parameter `page`; next token from `next.page`; page
  size 50; emits passthrough records.
- `list_change_request_page_meta_links`: GET `/spaces/{{ config.space_id }}/change-requests/{{
  config.change_request_id }}/content/page/{{ config.page_id }}/meta-links` - records path
  `alternates`; cursor pagination; cursor parameter `page`; next token from `next.page`; page size
  50; emits passthrough records.
- `get_reusable_content_in_change_request_by_id`: GET `/spaces/{{ config.space_id
  }}/change-requests/{{ config.change_request_id }}/content/reusable-contents/{{
  config.reusable_content_id }}` - records path `.`; emits passthrough records.
- `get_change_request_changes`: GET `/spaces/{{ config.space_id }}/change-requests/{{
  config.change_request_id }}/changes` - records path `changes`; query `limit`=`50`; cursor
  pagination; cursor parameter `page`; next token from `next.page`; page size 50; emits passthrough
  records.
- `get_change_request_pdf`: GET `/spaces/{{ config.space_id }}/change-requests/{{
  config.change_request_id }}/pdf` - records path `.`; emits passthrough records.
- `get_revision_by_id`: GET `/spaces/{{ config.space_id }}/revisions/{{ config.revision_id }}` -
  records path `pages`; cursor pagination; cursor parameter `page`; next token from `next.page`;
  page size 50; emits passthrough records.
- `get_revision_semantic_changes`: GET `/spaces/{{ config.space_id }}/revisions/{{
  config.revision_id }}/changes` - records path `changes`; query `limit`=`50`; cursor pagination;
  cursor parameter `page`; next token from `next.page`; page size 50; emits passthrough records.
- `list_pages_in_revision_by_id`: GET `/spaces/{{ config.space_id }}/revisions/{{ config.revision_id
  }}/pages` - records path `pages`; cursor pagination; cursor parameter `page`; next token from
  `next.page`; page size 50; emits passthrough records.
- `list_files_in_revision_by_id`: GET `/spaces/{{ config.space_id }}/revisions/{{ config.revision_id
  }}/files` - records path `items`; query `limit`=`50`; cursor pagination; cursor parameter `page`;
  next token from `next.page`; page size 50; emits passthrough records.
- `get_file_in_revision_by_id`: GET `/spaces/{{ config.space_id }}/revisions/{{ config.revision_id
  }}/files/{{ config.file_id }}` - records path `.`; emits passthrough records.
- `get_page_in_revision_by_id`: GET `/spaces/{{ config.space_id }}/revisions/{{ config.revision_id
  }}/page/{{ config.page_id }}` - records path `pages`; cursor pagination; cursor parameter `page`;
  next token from `next.page`; page size 50; emits passthrough records.
- `get_page_document_in_revision_by_id`: GET `/spaces/{{ config.space_id }}/revisions/{{
  config.revision_id }}/page/{{ config.page_id }}/document` - records path `nodes`; cursor
  pagination; cursor parameter `page`; next token from `next.page`; page size 50; emits passthrough
  records.
- `get_page_in_revision_by_path`: GET `/spaces/{{ config.space_id }}/revisions/{{ config.revision_id
  }}/path/{{ config.page_path }}` - records path `pages`; cursor pagination; cursor parameter
  `page`; next token from `next.page`; page size 50; emits passthrough records.
- `list_revision_page_meta_links`: GET `/spaces/{{ config.space_id }}/revisions/{{
  config.revision_id }}/page/{{ config.page_id }}/meta-links` - records path `alternates`; cursor
  pagination; cursor parameter `page`; next token from `next.page`; page size 50; emits passthrough
  records.
- `get_page_in_change_request_by_path`: GET `/spaces/{{ config.space_id }}/change-requests/{{
  config.change_request_id }}/content/path/{{ config.page_path }}` - records path `pages`; cursor
  pagination; cursor parameter `page`; next token from `next.page`; page size 50; emits passthrough
  records.
- `get_reusable_content_in_revision_by_id`: GET `/spaces/{{ config.space_id }}/revisions/{{
  config.revision_id }}/reusable-contents/{{ config.reusable_content_id }}` - records path `.`;
  emits passthrough records.
- `get_reusable_content_document_in_revision_by_id`: GET `/spaces/{{ config.space_id }}/revisions/{{
  config.revision_id }}/reusable-contents/{{ config.reusable_content_id }}/document` - records path
  `nodes`; cursor pagination; cursor parameter `page`; next token from `next.page`; page size 50;
  emits passthrough records.
- `list_comments_in_space`: GET `/spaces/{{ config.space_id }}/comments` - records path `items`;
  query `limit`=`50`; cursor pagination; cursor parameter `page`; next token from `next.page`; page
  size 50; emits passthrough records.
- `get_comment_in_space`: GET `/spaces/{{ config.space_id }}/comments/{{ config.comment_id }}` -
  records path `reactions`; cursor pagination; cursor parameter `page`; next token from `next.page`;
  page size 50; emits passthrough records.
- `list_comment_replies_in_space`: GET `/spaces/{{ config.space_id }}/comments/{{ config.comment_id
  }}/replies` - records path `items`; query `limit`=`50`; cursor pagination; cursor parameter
  `page`; next token from `next.page`; page size 50; emits passthrough records.
- `get_comment_reply_in_space`: GET `/spaces/{{ config.space_id }}/comments/{{ config.comment_id
  }}/replies/{{ config.comment_reply_id }}` - records path `reactions`; cursor pagination; cursor
  parameter `page`; next token from `next.page`; page size 50; emits passthrough records.
- `list_commenters_in_space`: GET `/spaces/{{ config.space_id }}/commenters` - records path `items`;
  query `limit`=`50`; cursor pagination; cursor parameter `page`; next token from `next.page`; page
  size 50; emits passthrough records.
- `list_commenters_in_change_request`: GET `/spaces/{{ config.space_id }}/change-requests/{{
  config.change_request_id }}/commenters` - records path `items`; query `limit`=`50`; cursor
  pagination; cursor parameter `page`; next token from `next.page`; page size 50; emits passthrough
  records.
- `list_permissions_aggregate_in_space`: GET `/spaces/{{ config.space_id }}/permissions/aggregate` -
  records path `items`; query `limit`=`50`; cursor pagination; cursor parameter `page`; next token
  from `next.page`; page size 50; emits passthrough records.
- `list_space_integrations`: GET `/spaces/{{ config.space_id }}/integrations` - records path
  `items`; cursor pagination; cursor parameter `page`; next token from `next.page`; page size 50;
  emits passthrough records.
- `list_space_integrations_blocks`: GET `/spaces/{{ config.space_id }}/integration-blocks` - records
  path `.`; emits passthrough records.
- `get_space_pdf`: GET `/spaces/{{ config.space_id }}/pdf` - records path `.`; emits passthrough
  records.
- `list_space_links`: GET `/spaces/{{ config.space_id }}/links` - records path `items`; query
  `limit`=`50`; cursor pagination; cursor parameter `page`; next token from `next.page`; page size
  50; emits passthrough records.
- `get_collection_by_id`: GET `/collections/{{ config.collection_id }}` - records path `.`; emits
  passthrough records.
- `list_spaces_in_collection_by_id`: GET `/collections/{{ config.collection_id }}/spaces` - records
  path `items`; query `limit`=`50`; cursor pagination; cursor parameter `page`; next token from
  `next.page`; page size 50; emits passthrough records.
- `list_team_permissions_in_collection`: GET `/collections/{{ config.collection_id
  }}/permissions/teams` - records path `items`; query `limit`=`50`; cursor pagination; cursor
  parameter `page`; next token from `next.page`; page size 50; emits passthrough records.
- `list_user_permissions_in_collection`: GET `/collections/{{ config.collection_id
  }}/permissions/users` - records path `items`; query `limit`=`50`; cursor pagination; cursor
  parameter `page`; next token from `next.page`; page size 50; emits passthrough records.
- `list_permissions_aggregate_in_collection`: GET `/collections/{{ config.collection_id
  }}/permissions/aggregate` - records path `items`; query `limit`=`50`; cursor pagination; cursor
  parameter `page`; next token from `next.page`; page size 50; emits passthrough records.
- `list_integrations`: GET `/integrations` - records path `items`; query `limit`=`50`; cursor
  pagination; cursor parameter `page`; next token from `next.page`; page size 50; emits passthrough
  records.
- `get_integration_by_name`: GET `/integrations/{{ config.integration_name }}` - records path
  `previewImages`; cursor pagination; cursor parameter `page`; next token from `next.page`; page
  size 50; emits passthrough records.
- `list_integration_installations`: GET `/integrations/{{ config.integration_name }}/installations`
  - records path `items`; query `limit`=`50`; cursor pagination; cursor parameter `page`; next token
  from `next.page`; page size 50; emits passthrough records.
- `list_integration_events`: GET `/integrations/{{ config.integration_name }}/events` - records path
  `items`; query `limit`=`50`; cursor pagination; cursor parameter `page`; next token from
  `next.page`; page size 50; emits passthrough records.
- `get_integration_event`: GET `/integrations/{{ config.integration_name }}/events/{{
  config.event_id }}` - records path `.`; emits passthrough records.
- `list_integration_space_installations`: GET `/integrations/{{ config.integration_name }}/spaces` -
  records path `items`; query `limit`=`50`; cursor pagination; cursor parameter `page`; next token
  from `next.page`; page size 50; emits passthrough records.
- `list_integration_site_installations`: GET `/integrations/{{ config.integration_name }}/sites` -
  records path `items`; query `limit`=`50`; cursor pagination; cursor parameter `page`; next token
  from `next.page`; page size 50; emits passthrough records.
- `render_integration_ui_with_get`: GET `/integrations/{{ config.integration_name }}/render` -
  records path `.`; query `request`=`{{ config.request }}`; emits passthrough records.
- `get_integration_installation_by_id`: GET `/integrations/{{ config.integration_name
  }}/installations/{{ config.installation_id }}` - records path `externalIds`; cursor pagination;
  cursor parameter `page`; next token from `next.page`; page size 50; emits passthrough records.
- `list_integration_installation_spaces`: GET `/integrations/{{ config.integration_name
  }}/installations/{{ config.installation_id }}/spaces` - records path `items`; query `limit`=`50`;
  cursor pagination; cursor parameter `page`; next token from `next.page`; page size 50; emits
  passthrough records.
- `get_integration_space_installation`: GET `/integrations/{{ config.integration_name
  }}/installations/{{ config.installation_id }}/spaces/{{ config.space_id }}` - records path
  `externalIds`; cursor pagination; cursor parameter `page`; next token from `next.page`; page size
  50; emits passthrough records.
- `list_integration_installation_sites`: GET `/integrations/{{ config.integration_name
  }}/installations/{{ config.installation_id }}/sites` - records path `items`; query `limit`=`50`;
  cursor pagination; cursor parameter `page`; next token from `next.page`; page size 50; emits
  passthrough records.
- `get_integration_site_installation`: GET `/integrations/{{ config.integration_name
  }}/installations/{{ config.installation_id }}/sites/{{ config.site_id }}` - records path
  `externalIds`; cursor pagination; cursor parameter `page`; next token from `next.page`; page size
  50; emits passthrough records.
- `get_organization_by_id`: GET `/orgs/{{ config.organization_id }}` - records path `emailDomains`;
  cursor pagination; cursor parameter `page`; next token from `next.page`; page size 50; emits
  passthrough records.
- `get_member_in_organization_by_id`: GET `/orgs/{{ config.organization_id }}/members/{{
  config.user_id }}` - records path `.`; emits passthrough records.
- `list_spaces_for_organization_member`: GET `/orgs/{{ config.organization_id }}/members/{{
  config.user_id }}/spaces` - records path `items`; query `limit`=`50`; cursor pagination; cursor
  parameter `page`; next token from `next.page`; page size 50; emits passthrough records.
- `list_teams_for_organization_member`: GET `/orgs/{{ config.organization_id }}/members/{{
  config.user_id }}/teams` - records path `items`; query `limit`=`50`; cursor pagination; cursor
  parameter `page`; next token from `next.page`; page size 50; emits passthrough records.
- `list_teams_in_organization_by_id`: GET `/orgs/{{ config.organization_id }}/teams` - records path
  `items`; query `limit`=`50`; cursor pagination; cursor parameter `page`; next token from
  `next.page`; page size 50; emits passthrough records.
- `get_team_in_organization_by_id`: GET `/orgs/{{ config.organization_id }}/teams/{{ config.team_id
  }}` - records path `.`; emits passthrough records.
- `list_team_members_in_organization_by_id`: GET `/orgs/{{ config.organization_id }}/teams/{{
  config.team_id }}/members` - records path `items`; query `limit`=`50`; cursor pagination; cursor
  parameter `page`; next token from `next.page`; page size 50; emits passthrough records.
- `list_organization_invite_links`: GET `/orgs/{{ config.organization_id }}/link-invites` - records
  path `items`; query `limit`=`50`; cursor pagination; cursor parameter `page`; next token from
  `next.page`; page size 50; emits passthrough records.
- `get_organization_invite_link`: GET `/orgs/{{ config.organization_id }}/link-invites/{{
  config.invite_id }}` - records path `.`; emits passthrough records.
- `search_organization_content`: GET `/orgs/{{ config.organization_id }}/search` - records path
  `items`; query `limit`=`50`; `query`=`{{ config.query }}`; cursor pagination; cursor parameter
  `page`; next token from `next.page`; page size 50; emits passthrough records.
- `list_change_requests_for_organization`: GET `/orgs/{{ config.organization_id }}/change-requests`
  - records path `items`; query `limit`=`50`; cursor pagination; cursor parameter `page`; next token
  from `next.page`; page size 50; emits passthrough records.
- `list_spaces_in_organization_by_id`: GET `/orgs/{{ config.organization_id }}/spaces` - records
  path `items`; query `limit`=`50`; cursor pagination; cursor parameter `page`; next token from
  `next.page`; page size 50; emits passthrough records.
- `list_collections_in_organization_by_id`: GET `/orgs/{{ config.organization_id }}/collections` -
  records path `items`; query `limit`=`50`; cursor pagination; cursor parameter `page`; next token
  from `next.page`; page size 50; emits passthrough records.
- `list_organization_integrations`: GET `/orgs/{{ config.organization_id }}/integrations` - records
  path `items`; query `limit`=`50`; cursor pagination; cursor parameter `page`; next token from
  `next.page`; page size 50; emits passthrough records.
- `get_organization_integration_status`: GET `/orgs/{{ config.organization_id }}/integrations/{{
  config.integration_name }}/installation_status` - records path `.`; emits passthrough records.
- `list_organization_installations`: GET `/orgs/{{ config.organization_id }}/installations` -
  records path `items`; query `limit`=`50`; cursor pagination; cursor parameter `page`; next token
  from `next.page`; page size 50; emits passthrough records.
- `list_organization_integrations_status`: GET `/orgs/{{ config.organization_id
  }}/integrations/installations-status` - records path `items`; query `limit`=`50`; cursor
  pagination; cursor parameter `page`; next token from `next.page`; page size 50; emits passthrough
  records.
- `list_saml_providers_in_organization_by_id`: GET `/orgs/{{ config.organization_id }}/saml` -
  records path `items`; query `limit`=`50`; cursor pagination; cursor parameter `page`; next token
  from `next.page`; page size 50; emits passthrough records.
- `get_organization_saml_provider_by_id`: GET `/orgs/{{ config.organization_id }}/saml/{{
  config.saml_provider_id }}` - records path `.`; emits passthrough records.
- `list_sso_provider_logins_in_organization`: GET `/orgs/{{ config.organization_id }}/sso` - records
  path `items`; cursor pagination; cursor parameter `page`; next token from `next.page`; page size
  50; emits passthrough records.
- `get_recommended_questions_in_organization`: GET `/orgs/{{ config.organization_id
  }}/ask/questions` - records path `questions`; cursor pagination; cursor parameter `page`; next
  token from `next.page`; page size 50; emits passthrough records.
- `list_open_api_specs`: GET `/orgs/{{ config.organization_id }}/openapi` - records path `items`;
  query `limit`=`50`; cursor pagination; cursor parameter `page`; next token from `next.page`; page
  size 50; emits passthrough records.
- `get_open_api_spec_by_slug`: GET `/orgs/{{ config.organization_id }}/openapi/{{ config.spec_slug
  }}` - records path `lastProcessedErrors`; cursor pagination; cursor parameter `page`; next token
  from `next.page`; page size 50; emits passthrough records.
- `list_open_api_spec_versions`: GET `/orgs/{{ config.organization_id }}/openapi/{{ config.spec_slug
  }}/versions` - records path `items`; query `limit`=`50`; cursor pagination; cursor parameter
  `page`; next token from `next.page`; page size 50; emits passthrough records.
- `get_latest_open_api_spec_version`: GET `/orgs/{{ config.organization_id }}/openapi/{{
  config.spec_slug }}/versions/latest` - records path `.`; emits passthrough records.
- `get_latest_open_api_spec_version_content`: GET `/orgs/{{ config.organization_id }}/openapi/{{
  config.spec_slug }}/versions/latest/content` - records path `.`; emits passthrough records.
- `get_open_api_spec_version_by_id`: GET `/orgs/{{ config.organization_id }}/openapi/{{
  config.spec_slug }}/versions/{{ config.version_id }}` - records path `.`; emits passthrough
  records.
- `get_open_api_spec_version_content_by_id`: GET `/orgs/{{ config.organization_id }}/openapi/{{
  config.spec_slug }}/versions/{{ config.version_id }}/content` - records path `.`; emits
  passthrough records.
- `get_organization_agent_instructions`: GET `/orgs/{{ config.organization_id }}/agent-instructions`
  - records path `.`; emits passthrough records.
- `list_translations`: GET `/orgs/{{ config.organization_id }}/translations` - records path `items`;
  query `limit`=`50`; cursor pagination; cursor parameter `page`; next token from `next.page`; page
  size 50; emits passthrough records.
- `get_translation`: GET `/orgs/{{ config.organization_id }}/translations/{{ config.translation_id
  }}` - records path `.`; emits passthrough records.
- `list_glossary_entries`: GET `/orgs/{{ config.organization_id }}/translations-glossary` - records
  path `items`; query `limit`=`50`; cursor pagination; cursor parameter `page`; next token from
  `next.page`; page size 50; emits passthrough records.
- `get_glossary_entry`: GET `/orgs/{{ config.organization_id }}/translations-glossary/{{
  config.glossary_entry_id }}` - records path `.`; emits passthrough records.
- `list_custom_fonts`: GET `/orgs/{{ config.organization_id }}/fonts` - records path `items`; cursor
  pagination; cursor parameter `page`; next token from `next.page`; page size 50; emits passthrough
  records.
- `get_custom_font`: GET `/orgs/{{ config.organization_id }}/fonts/{{ config.font_id }}` - records
  path `fontFaces`; cursor pagination; cursor parameter `page`; next token from `next.page`; page
  size 50; emits passthrough records.
- `list_sites`: GET `/orgs/{{ config.organization_id }}/sites` - records path `items`; query
  `limit`=`50`; cursor pagination; cursor parameter `page`; next token from `next.page`; page size
  50; emits passthrough records.
- `get_site_by_id`: GET `/orgs/{{ config.organization_id }}/sites/{{ config.site_id }}` - records
  path `features`; cursor pagination; cursor parameter `page`; next token from `next.page`; page
  size 50; emits passthrough records.
- `list_site_git_sync_installations`: GET `/orgs/{{ config.organization_id }}/sites/{{
  config.site_id }}/spaces/git/installations` - records path `items`; cursor pagination; cursor
  parameter `page`; next token from `next.page`; page size 50; emits passthrough records.
- `get_site_adaptive_schema`: GET `/orgs/{{ config.organization_id }}/sites/{{ config.site_id
  }}/adaptive-schema` - records path `.`; emits passthrough records.
- `list_site_adaptive_template_conditions`: GET `/orgs/{{ config.organization_id }}/sites/{{
  config.site_id }}/adaptive-schema/template-conditions` - records path `items`; cursor pagination;
  cursor parameter `page`; next token from `next.page`; page size 50; emits passthrough records.
- `get_published_content_site`: GET `/orgs/{{ config.organization_id }}/sites/{{ config.site_id
  }}/published` - records path `scripts`; cursor pagination; cursor parameter `page`; next token
  from `next.page`; page size 50; emits passthrough records.
- `list_site_share_links`: GET `/orgs/{{ config.organization_id }}/sites/{{ config.site_id
  }}/share-links` - records path `items`; query `limit`=`50`; cursor pagination; cursor parameter
  `page`; next token from `next.page`; page size 50; emits passthrough records.
- `get_site_structure`: GET `/orgs/{{ config.organization_id }}/sites/{{ config.site_id
  }}/structure` - records path `structure`; cursor pagination; cursor parameter `page`; next token
  from `next.page`; page size 50; emits passthrough records.
- `get_site_publishing_auth_by_id`: GET `/orgs/{{ config.organization_id }}/sites/{{ config.site_id
  }}/publishing/auth` - records path `.`; emits passthrough records.
- `get_site_publishing_preview_by_id`: GET `/orgs/{{ config.organization_id }}/sites/{{
  config.site_id }}/publishing/preview` - records path `.`; emits passthrough records.
- `get_site_customization_by_id`: GET `/orgs/{{ config.organization_id }}/sites/{{ config.site_id
  }}/customization` - records path `socialAccounts`; cursor pagination; cursor parameter `page`;
  next token from `next.page`; page size 50; emits passthrough records.
- `list_site_integration_scripts`: GET `/orgs/{{ config.organization_id }}/sites/{{ config.site_id
  }}/integration-scripts` - records path `.`; emits passthrough records.
- `list_site_integrations`: GET `/orgs/{{ config.organization_id }}/sites/{{ config.site_id
  }}/integrations` - records path `items`; cursor pagination; cursor parameter `page`; next token
  from `next.page`; page size 50; emits passthrough records.
- `list_site_spaces`: GET `/orgs/{{ config.organization_id }}/sites/{{ config.site_id
  }}/site-spaces` - records path `items`; query `limit`=`50`; cursor pagination; cursor parameter
  `page`; next token from `next.page`; page size 50; emits passthrough records.
- `list_site_section_groups`: GET `/orgs/{{ config.organization_id }}/sites/{{ config.site_id
  }}/section-groups` - records path `items`; query `limit`=`50`; cursor pagination; cursor parameter
  `page`; next token from `next.page`; page size 50; emits passthrough records.
- `list_site_sections`: GET `/orgs/{{ config.organization_id }}/sites/{{ config.site_id }}/sections`
  - records path `items`; query `limit`=`50`; cursor pagination; cursor parameter `page`; next token
  from `next.page`; page size 50; emits passthrough records.
- `list_site_context_records`: GET `/orgs/{{ config.organization_id }}/sites/{{ config.site_id
  }}/context-records` - records path `items`; query `limit`=`50`; cursor pagination; cursor
  parameter `page`; next token from `next.page`; page size 50; emits passthrough records.
- `get_site_context_record_by_id`: GET `/orgs/{{ config.organization_id }}/sites/{{ config.site_id
  }}/context-records/{{ config.site_context_record_id }}` - records path `topics`; cursor
  pagination; cursor parameter `page`; next token from `next.page`; page size 50; emits passthrough
  records.
- `list_site_scans`: GET `/orgs/{{ config.organization_id }}/sites/{{ config.site_id }}/scans` -
  records path `items`; query `limit`=`50`; cursor pagination; cursor parameter `page`; next token
  from `next.page`; page size 50; emits passthrough records.
- `get_site_scan_by_id`: GET `/orgs/{{ config.organization_id }}/sites/{{ config.site_id }}/scans/{{
  config.site_scan_id }}` - records path `.`; emits passthrough records.
- `list_site_findings`: GET `/orgs/{{ config.organization_id }}/sites/{{ config.site_id }}/findings`
  - records path `items`; query `limit`=`50`; cursor pagination; cursor parameter `page`; next token
  from `next.page`; page size 50; emits passthrough records.
- `get_site_finding_by_id`: GET `/orgs/{{ config.organization_id }}/sites/{{ config.site_id
  }}/findings/{{ config.site_finding_id }}` - records path `.`; emits passthrough records.
- `list_change_requests_for_site_finding`: GET `/orgs/{{ config.organization_id }}/sites/{{
  config.site_id }}/findings/{{ config.site_finding_id }}/change-requests` - records path `items`;
  query `limit`=`50`; cursor pagination; cursor parameter `page`; next token from `next.page`; page
  size 50; emits passthrough records.
- `list_pages_for_site_finding`: GET `/orgs/{{ config.organization_id }}/sites/{{ config.site_id
  }}/findings/{{ config.site_finding_id }}/pages` - records path `items`; query `limit`=`50`; cursor
  pagination; cursor parameter `page`; next token from `next.page`; page size 50; emits passthrough
  records.
- `list_questions_for_site_finding`: GET `/orgs/{{ config.organization_id }}/sites/{{ config.site_id
  }}/findings/{{ config.site_finding_id }}/questions` - records path `items`; query `limit`=`50`;
  cursor pagination; cursor parameter `page`; next token from `next.page`; page size 50; emits
  passthrough records.
- `list_records_for_site_finding`: GET `/orgs/{{ config.organization_id }}/sites/{{ config.site_id
  }}/findings/{{ config.site_finding_id }}/records` - records path `items`; query `limit`=`50`;
  cursor pagination; cursor parameter `page`; next token from `next.page`; page size 50; emits
  passthrough records.
- `list_site_context_connections`: GET `/orgs/{{ config.organization_id }}/sites/{{ config.site_id
  }}/context-connections` - records path `items`; query `limit`=`50`; cursor pagination; cursor
  parameter `page`; next token from `next.page`; page size 50; emits passthrough records.
- `get_site_context_connection_by_id`: GET `/orgs/{{ config.organization_id }}/sites/{{
  config.site_id }}/context-connections/{{ config.site_context_connection_id }}` - records path `.`;
  emits passthrough records.
- `list_site_topics`: GET `/orgs/{{ config.organization_id }}/sites/{{ config.site_id }}/topics` -
  records path `items`; cursor pagination; cursor parameter `page`; next token from `next.page`;
  page size 50; emits passthrough records.
- `get_site_topic_by_id`: GET `/orgs/{{ config.organization_id }}/sites/{{ config.site_id
  }}/topics/{{ config.site_topic_id }}` - records path `.`; emits passthrough records.
- `list_site_questions`: GET `/orgs/{{ config.organization_id }}/sites/{{ config.site_id
  }}/questions` - records path `items`; query `limit`=`50`; cursor pagination; cursor parameter
  `page`; next token from `next.page`; page size 50; emits passthrough records.
- `get_site_question_by_id`: GET `/orgs/{{ config.organization_id }}/sites/{{ config.site_id
  }}/questions/{{ config.site_question_id }}` - records path `.`; emits passthrough records.
- `list_site_question_sources`: GET `/orgs/{{ config.organization_id }}/sites/{{ config.site_id
  }}/questions/{{ config.site_question_id }}/sources` - records path `items`; query `limit`=`50`;
  cursor pagination; cursor parameter `page`; next token from `next.page`; page size 50; emits
  passthrough records.
- `get_site_question_stats`: GET `/orgs/{{ config.organization_id }}/sites/{{ config.site_id
  }}/question-stats` - records path `weeklyPulse`; cursor pagination; cursor parameter `page`; next
  token from `next.page`; page size 50; emits passthrough records.
- `list_site_question_answers`: GET `/orgs/{{ config.organization_id }}/sites/{{ config.site_id
  }}/answers` - records path `items`; query `limit`=`50`; cursor pagination; cursor parameter
  `page`; next token from `next.page`; page size 50; emits passthrough records.
- `get_site_question_answer_by_id`: GET `/orgs/{{ config.organization_id }}/sites/{{ config.site_id
  }}/answers/{{ config.site_question_answer_id }}` - records path `topics`; cursor pagination;
  cursor parameter `page`; next token from `next.page`; page size 50; emits passthrough records.
- `get_site_question_answer_thread_by_id`: GET `/orgs/{{ config.organization_id }}/sites/{{
  config.site_id }}/answers/{{ config.site_question_answer_id }}/thread` - records path `.`; emits
  passthrough records.
- `list_site_question_answer_sources`: GET `/orgs/{{ config.organization_id }}/sites/{{
  config.site_id }}/answers/{{ config.site_question_answer_id }}/sources` - records path `items`;
  query `limit`=`50`; cursor pagination; cursor parameter `page`; next token from `next.page`; page
  size 50; emits passthrough records.
- `get_site_space_customization_by_id`: GET `/orgs/{{ config.organization_id }}/sites/{{
  config.site_id }}/site-spaces/{{ config.site_space_id }}/customization` - records path
  `socialAccounts`; cursor pagination; cursor parameter `page`; next token from `next.page`; page
  size 50; emits passthrough records.
- `list_permissions_aggregate_in_site`: GET `/orgs/{{ config.organization_id }}/sites/{{
  config.site_id }}/permissions/aggregate` - records path `items`; query `limit`=`50`; cursor
  pagination; cursor parameter `page`; next token from `next.page`; page size 50; emits passthrough
  records.
- `list_user_permissions_in_site`: GET `/orgs/{{ config.organization_id }}/sites/{{ config.site_id
  }}/permissions/users` - records path `items`; query `limit`=`50`; cursor pagination; cursor
  parameter `page`; next token from `next.page`; page size 50; emits passthrough records.
- `list_team_permissions_in_site`: GET `/orgs/{{ config.organization_id }}/sites/{{ config.site_id
  }}/permissions/teams` - records path `items`; query `limit`=`50`; cursor pagination; cursor
  parameter `page`; next token from `next.page`; page size 50; emits passthrough records.
- `get_site_agent_settings_by_id`: GET `/orgs/{{ config.organization_id }}/sites/{{ config.site_id
  }}/agent-settings` - records path `.`; emits passthrough records.
- `list_site_visitor_segments`: GET `/orgs/{{ config.organization_id }}/sites/{{ config.site_id
  }}/insights/visitor-segments` - records path `items`; cursor pagination; cursor parameter `page`;
  next token from `next.page`; page size 50; emits passthrough records.
- `list_site_redirects`: GET `/orgs/{{ config.organization_id }}/sites/{{ config.site_id
  }}/redirects` - records path `items`; query `limit`=`50`; cursor pagination; cursor parameter
  `page`; next token from `next.page`; page size 50; emits passthrough records.
- `get_site_redirect_by_source`: GET `/orgs/{{ config.organization_id }}/sites/{{ config.site_id
  }}/redirect` - records path `.`; query `source`=`{{ config.source }}`; emits passthrough records.
- `list_site_mcp_servers`: GET `/orgs/{{ config.organization_id }}/sites/{{ config.site_id
  }}/mcp-servers` - records path `items`; query `limit`=`50`; cursor pagination; cursor parameter
  `page`; next token from `next.page`; page size 50; emits passthrough records.
- `get_site_mcp_server_by_id`: GET `/orgs/{{ config.organization_id }}/sites/{{ config.site_id
  }}/mcp-servers/{{ config.site_mcp_server_id }}` - records path `.`; emits passthrough records.
- `list_site_channels`: GET `/orgs/{{ config.organization_id }}/sites/{{ config.site_id }}/channels`
  - records path `items`; query `limit`=`50`; cursor pagination; cursor parameter `page`; next token
  from `next.page`; page size 50; emits passthrough records.
- `get_site_channel_by_id`: GET `/orgs/{{ config.organization_id }}/sites/{{ config.site_id
  }}/channels/{{ config.site_channel_id }}` - records path `.`; emits passthrough records.
- `get_subdomain`: GET `/subdomains/{{ config.subdomain }}` - records path `.`; emits passthrough
  records.
- `get_custom_hostname`: GET `/custom-hostnames/{{ config.hostname }}` - records path `.`; emits
  passthrough records.
- `get_organizations_for_email_domain`: GET `/email-domains/{{ config.email_domain }}/orgs` -
  records path `organizations`; cursor pagination; cursor parameter `page`; next token from
  `next.page`; page size 50; emits passthrough records.
- `ads_list_sites`: GET `/ads/sites` - records path `items`; query `limit`=`50`; cursor pagination;
  cursor parameter `page`; next token from `next.page`; page size 50; emits passthrough records.
- `get_content_by_url`: GET `/urls/content` - records path `.`; query `url`=`{{ config.url }}`;
  emits passthrough records.
- `get_embed_by_url`: GET `/urls/embed` - records path `.`; query `url`=`{{ config.url }}`; emits
  passthrough records.
- `get_published_content_by_url`: GET `/urls/published` - records path `.`; query `url`=`{{
  config.url }}`; emits passthrough records.
- `get_git_sync_installation_by_id`: GET `/git/installations/{{ config.installation_id }}` - records
  path `.`; emits passthrough records.
- `list_git_hub_repositories_for_git_sync_installation`: GET `/git/installations/{{
  config.installation_id }}/github/repos` - records path `items`; cursor pagination; cursor
  parameter `page`; next token from `next.page`; page size 50; emits passthrough records.
- `list_git_hub_repo_branches_for_git_sync_installation`: GET `/git/installations/{{
  config.installation_id }}/github/repos/{{ config.account_name }}/{{ config.repository_name
  }}/branches` - records path `items`; cursor pagination; cursor parameter `page`; next token from
  `next.page`; page size 50; emits passthrough records.
- `list_git_lab_projects_for_git_sync_installation`: GET `/git/installations/{{
  config.installation_id }}/gitlab/projects` - records path `items`; cursor pagination; cursor
  parameter `page`; next token from `next.page`; page size 50; emits passthrough records.
- `list_git_lab_project_branches_for_git_sync_installation`: GET `/git/installations/{{
  config.installation_id }}/gitlab/projects/{{ config.project_id }}/branches` - records path
  `items`; cursor pagination; cursor parameter `page`; next token from `next.page`; page size 50;
  emits passthrough records.

## Write actions & risks

Overall write risk: creates, updates, publishes, archives, deletes, imports, exports, invites,
permission changes, and content changes in GitBook depending on the selected write action.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_user_notifications_token`: POST `/user/notifications/token` - kind `create`; body type
  `none`; risk: POST /user/notifications/token (Create a JWT to access the in-app notifications
  service) executes a live GitBook API operation.
- `update_user_by_id`: PATCH `/users/{{ record.user_id }}` - kind `update`; body type `json`; path
  fields `user_id`; required record fields `user_id`; accepted fields `display_name`, `photo_url`,
  `user_id`; risk: PATCH /users/{userId} (Update a user by its ID) executes a live GitBook API
  operation.
- `update_space_by_id`: PATCH `/spaces/{{ record.space_id }}` - kind `update`; body type `json`;
  path fields `space_id`; required record fields `space_id`; accepted fields `default_level`,
  `edit_mode`, `emoji`, `icon`, `language`, `merge_rules`, `space_id`, `title`; risk: PATCH
  /spaces/{spaceId} (Update a space's title, icon, or settings) executes a live GitBook API
  operation.
- `delete_space_by_id`: DELETE `/spaces/{{ record.space_id }}` - kind `delete`; body type `none`;
  path fields `space_id`; required record fields `space_id`; accepted fields `space_id`; missing
  records treated as success for status `404`; confirmation `destructive`; risk: DELETE
  /spaces/{spaceId} (Delete a space) executes a live GitBook API operation.
- `duplicate_space`: POST `/spaces/{{ record.space_id }}/duplicate` - kind `custom`; body type
  `none`; path fields `space_id`; required record fields `space_id`; accepted fields `space_id`;
  risk: POST /spaces/{spaceId}/duplicate (Create a full copy of a space) executes a live GitBook API
  operation.
- `restore_space`: POST `/spaces/{{ record.space_id }}/restore` - kind `custom`; body type `none`;
  path fields `space_id`; required record fields `space_id`; accepted fields `space_id`; risk: POST
  /spaces/{spaceId}/restore (Restore a recently deleted space from the trash) executes a live
  GitBook API operation.
- `move_space`: POST `/spaces/{{ record.space_id }}/move` - kind `custom`; body type `json`; path
  fields `space_id`; required record fields `space_id`; accepted fields `parent`, `position`,
  `space_id`; risk: POST /spaces/{spaceId}/move (Move a space to a different collection or position)
  executes a live GitBook API operation.
- `import_git_repository`: POST `/spaces/{{ record.space_id }}/git/import` - kind `custom`; body
  type `json`; path fields `space_id`; required record fields `space_id`, `url`, `ref`; accepted
  fields `force`, `git_info`, `ref`, `repo_cache_id`, `repo_commit_url`, `repo_project_directory`,
  `repo_tree_url`, `space_id`, `standalone`, `timestamp`, `url`; risk: POST
  /spaces/{spaceId}/git/import (Pull content into a space from a connected Git repository) executes
  a live GitBook API operation.
- `export_to_git_repository`: POST `/spaces/{{ record.space_id }}/git/export` - kind `custom`; body
  type `json`; path fields `space_id`; required record fields `space_id`, `url`, `ref`,
  `commit_message`; accepted fields `commit_message`, `force`, `git_info`, `ref`, `repo_cache_id`,
  `repo_commit_url`, `repo_project_directory`, `repo_tree_url`, `space_id`, `timestamp`, `url`;
  risk: POST /spaces/{spaceId}/git/export (Push space content to a connected Git repository)
  executes a live GitBook API operation.
- `delete_git_installation`: DELETE connector-managed endpoint - kind `delete`; body type `none`;
  path fields `space_id`; required record fields `space_id`; accepted fields `space_id`; missing
  records treated as success for status `404`; confirmation `destructive`.
- `invite_to_space`: POST `/spaces/{{ record.space_id }}/permissions` - kind `create`; body type
  `json`; path fields `space_id`; required record fields `space_id`; accepted fields `space_id`,
  `teams`, `users`; risk: POST /spaces/{spaceId}/permissions (Invite a user or a team to a space)
  executes a live GitBook API operation.
- `update_team_permission_in_space`: PATCH `/spaces/{{ record.space_id }}/permissions/teams/{{
  record.team_id }}` - kind `update`; body type `json`; path fields `space_id`, `team_id`; required
  record fields `space_id`, `team_id`; accepted fields `role`, `space_id`, `team_id`; risk: PATCH
  /spaces/{spaceId}/permissions/teams/{teamId} (Update an org team's permission in a space) executes
  a live GitBook API operation.
- `remove_team_from_space`: DELETE `/spaces/{{ record.space_id }}/permissions/teams/{{
  record.team_id }}` - kind `delete`; body type `none`; path fields `space_id`, `team_id`; required
  record fields `space_id`, `team_id`; accepted fields `space_id`, `team_id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: DELETE
  /spaces/{spaceId}/permissions/teams/{teamId} (Remove an org team from a space) executes a live
  GitBook API operation.
- `update_user_permission_in_space`: PATCH `/spaces/{{ record.space_id }}/permissions/users/{{
  record.user_id }}` - kind `update`; body type `json`; path fields `space_id`, `user_id`; required
  record fields `space_id`, `user_id`; accepted fields `role`, `space_id`, `user_id`; risk: PATCH
  /spaces/{spaceId}/permissions/users/{userId} (Update space user permissions) executes a live
  GitBook API operation.
- `remove_user_from_space`: DELETE `/spaces/{{ record.space_id }}/permissions/users/{{
  record.user_id }}` - kind `delete`; body type `none`; path fields `space_id`, `user_id`; required
  record fields `space_id`, `user_id`; accepted fields `space_id`, `user_id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: DELETE
  /spaces/{spaceId}/permissions/users/{userId} (Remove a space user) executes a live GitBook API
  operation.
- `apply_template_to_space`: POST `/spaces/{{ record.space_id }}/content/template` - kind `custom`;
  body type `json`; path fields `space_id`; required record fields `space_id`, `id`; accepted fields
  `change_request_id`, `id`, `params`, `space_id`; risk: POST /spaces/{spaceId}/content/template
  (Apply a content template to populate a space with initial pages) executes a live GitBook API
  operation.
- `get_computed_document`: POST `/spaces/{{ record.space_id }}/content/computed/document` - kind
  `custom`; body type `json`; path fields `space_id`; required record fields `space_id`, `source`,
  `seed`; accepted fields `seed`, `source`, `space_id`; risk: POST
  /spaces/{spaceId}/content/computed/document (Compute and render a document from a structured
  content source) executes a live GitBook API operation.
- `get_computed_revision`: POST `/spaces/{{ record.space_id }}/content/computed/revision` - kind
  `custom`; body type `json`; path fields `space_id`; required record fields `space_id`, `source`,
  `seed`; accepted fields `seed`, `source`, `space_id`; risk: POST
  /spaces/{spaceId}/content/computed/revision (Compute and render a full revision from a structured
  content source) executes a live GitBook API operation.
- `create_change_request`: POST `/spaces/{{ record.space_id }}/change-requests` - kind `create`;
  body type `json`; path fields `space_id`; required record fields `space_id`; accepted fields
  `space_id`, `subject`, `template`; risk: POST /spaces/{spaceId}/change-requests (Create a new
  change request in a space) executes a live GitBook API operation.
- `update_change_request_by_id`: PATCH `/spaces/{{ record.space_id }}/change-requests/{{
  record.change_request_id }}` - kind `update`; body type `json`; path fields `space_id`,
  `change_request_id`; required record fields `space_id`, `change_request_id`; accepted fields
  `change_request_id`, `description`, `links`, `space_id`, `status`, `subject`; risk: PATCH
  /spaces/{spaceId}/change-requests/{changeRequestId} (Update a change request's subject,
  description, or status) executes a live GitBook API operation.
- `merge_change_request`: POST `/spaces/{{ record.space_id }}/change-requests/{{
  record.change_request_id }}/merge` - kind `custom`; body type `none`; path fields `space_id`,
  `change_request_id`; required record fields `space_id`, `change_request_id`; accepted fields
  `change_request_id`, `space_id`; risk: POST
  /spaces/{spaceId}/change-requests/{changeRequestId}/merge (Merge a change request into the space's
  live content) executes a live GitBook API operation.
- `update_change_request`: POST `/spaces/{{ record.space_id }}/change-requests/{{
  record.change_request_id }}/update` - kind `update`; body type `none`; path fields `space_id`,
  `change_request_id`; required record fields `space_id`, `change_request_id`; accepted fields
  `change_request_id`, `space_id`; risk: POST
  /spaces/{spaceId}/change-requests/{changeRequestId}/update (Sync a change request with the latest
  live space content) executes a live GitBook API operation.
- `submit_change_request_review`: POST `/spaces/{{ record.space_id }}/change-requests/{{
  record.change_request_id }}/reviews` - kind `custom`; body type `json`; path fields `space_id`,
  `change_request_id`; required record fields `space_id`, `change_request_id`, `status`; accepted
  fields `change_request_id`, `comment`, `space_id`, `status`; risk: POST
  /spaces/{spaceId}/change-requests/{changeRequestId}/reviews (Submit an approve or request-changes
  review for a change request) executes a live GitBook API operation.
- `request_reviewers_for_change_request`: POST `/spaces/{{ record.space_id }}/change-requests/{{
  record.change_request_id }}/requested-reviewers` - kind `custom`; body type `json`; path fields
  `space_id`, `change_request_id`; required record fields `space_id`, `change_request_id`, `users`;
  accepted fields `change_request_id`, `description`, `space_id`, `subject`, `users`; risk: POST
  /spaces/{spaceId}/change-requests/{changeRequestId}/requested-reviewers (Send review requests to
  users for a change request) executes a live GitBook API operation.
- `remove_requested_reviewer_from_change_request`: DELETE `/spaces/{{ record.space_id
  }}/change-requests/{{ record.change_request_id }}/requested-reviewers/{{ record.user_id }}` - kind
  `delete`; body type `none`; path fields `space_id`, `change_request_id`, `user_id`; required
  record fields `space_id`, `change_request_id`, `user_id`; accepted fields `change_request_id`,
  `space_id`, `user_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: DELETE
  /spaces/{spaceId}/change-requests/{changeRequestId}/requested-reviewers/{userId} (Remove a
  reviewer from a change request) executes a live GitBook API operation.
- `update_change_request_conversation`: PATCH `/spaces/{{ record.space_id }}/change-requests/{{
  record.change_request_id }}/conversations/{{ record.conversation_id }}` - kind `update`; body type
  `json`; path fields `space_id`, `change_request_id`, `conversation_id`; required record fields
  `space_id`, `change_request_id`, `conversation_id`, `title`; accepted fields `change_request_id`,
  `conversation_id`, `space_id`, `title`; risk: PATCH
  /spaces/{spaceId}/change-requests/{changeRequestId}/conversations/{conversationId} (Update the
  title of an AI agent conversation on a change request) executes a live GitBook API operation.
- `delete_change_request_conversation`: DELETE `/spaces/{{ record.space_id }}/change-requests/{{
  record.change_request_id }}/conversations/{{ record.conversation_id }}` - kind `delete`; body type
  `none`; path fields `space_id`, `change_request_id`, `conversation_id`; required record fields
  `space_id`, `change_request_id`, `conversation_id`; accepted fields `change_request_id`,
  `conversation_id`, `space_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: DELETE
  /spaces/{spaceId}/change-requests/{changeRequestId}/conversations/{conversationId} (Delete an
  agent conversation) executes a live GitBook API operation.
- `post_comment_in_change_request`: POST `/spaces/{{ record.space_id }}/change-requests/{{
  record.change_request_id }}/comments` - kind `create`; body type `json`; path fields `space_id`,
  `change_request_id`; required record fields `space_id`, `change_request_id`, `body`; accepted
  fields `body`, `change_request_id`, `node`, `page`, `space_id`, `table_cell`; risk: POST
  /spaces/{spaceId}/change-requests/{changeRequestId}/comments (Post a new comment on a change
  request) executes a live GitBook API operation.
- `update_comment_in_change_request`: PUT `/spaces/{{ record.space_id }}/change-requests/{{
  record.change_request_id }}/comments/{{ record.comment_id }}` - kind `update`; body type `json`;
  path fields `space_id`, `change_request_id`, `comment_id`; required record fields `space_id`,
  `change_request_id`, `comment_id`; accepted fields `added_reactions`, `body`, `change_request_id`,
  `comment_id`, `removed_reactions`, `resolved`, `space_id`; risk: PUT
  /spaces/{spaceId}/change-requests/{changeRequestId}/comments/{commentId} (Update the content or
  status of a change request comment) executes a live GitBook API operation.
- `delete_comment_in_change_request`: DELETE `/spaces/{{ record.space_id }}/change-requests/{{
  record.change_request_id }}/comments/{{ record.comment_id }}` - kind `delete`; body type `none`;
  path fields `space_id`, `change_request_id`, `comment_id`; required record fields `space_id`,
  `change_request_id`, `comment_id`; accepted fields `change_request_id`, `comment_id`, `space_id`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: DELETE
  /spaces/{spaceId}/change-requests/{changeRequestId}/comments/{commentId} (Delete a change request
  comment) executes a live GitBook API operation.
- `post_comment_reply_in_change_request`: POST `/spaces/{{ record.space_id }}/change-requests/{{
  record.change_request_id }}/comments/{{ record.comment_id }}/replies` - kind `create`; body type
  `json`; path fields `space_id`, `change_request_id`, `comment_id`; required record fields
  `space_id`, `change_request_id`, `comment_id`, `body`; accepted fields `body`,
  `change_request_id`, `comment_id`, `space_id`; risk: POST
  /spaces/{spaceId}/change-requests/{changeRequestId}/comments/{commentId}/replies (Post a reply to
  a change request comment) executes a live GitBook API operation.
- `update_comment_reply_in_change_request`: PUT `/spaces/{{ record.space_id }}/change-requests/{{
  record.change_request_id }}/comments/{{ record.comment_id }}/replies/{{ record.comment_reply_id
  }}` - kind `update`; body type `json`; path fields `space_id`, `change_request_id`, `comment_id`,
  `comment_reply_id`; required record fields `space_id`, `change_request_id`, `comment_id`,
  `comment_reply_id`; accepted fields `added_reactions`, `body`, `change_request_id`, `comment_id`,
  `comment_reply_id`, `removed_reactions`, `resolved`, `space_id`; risk: PUT
  /spaces/{spaceId}/change-requests/{changeRequestId}/comments/{commentId}/replies/{commentReplyId}
  (Update the content of a change request comment reply) executes a live GitBook API operation.
- `delete_comment_reply_in_change_request`: DELETE `/spaces/{{ record.space_id }}/change-requests/{{
  record.change_request_id }}/comments/{{ record.comment_id }}/replies/{{ record.comment_reply_id
  }}` - kind `delete`; body type `none`; path fields `space_id`, `change_request_id`, `comment_id`,
  `comment_reply_id`; required record fields `space_id`, `change_request_id`, `comment_id`,
  `comment_reply_id`; accepted fields `change_request_id`, `comment_id`, `comment_reply_id`,
  `space_id`; missing records treated as success for status `404`; confirmation `destructive`; risk:
  DELETE
  /spaces/{spaceId}/change-requests/{changeRequestId}/comments/{commentId}/replies/{commentReplyId}
  (Delete a change request comment reply) executes a live GitBook API operation.
- `update_change_request_content`: POST `/spaces/{{ record.space_id }}/change-requests/{{
  record.change_request_id }}/content` - kind `update`; body type `json`; path fields `space_id`,
  `change_request_id`; required record fields `space_id`, `change_request_id`, `changes`; accepted
  fields `change_request_id`, `changes`, `space_id`; risk: POST
  /spaces/{spaceId}/change-requests/{changeRequestId}/content (Apply a batch of content changes to a
  change request) executes a live GitBook API operation.
- `post_comment_in_space`: POST `/spaces/{{ record.space_id }}/comments` - kind `create`; body type
  `json`; path fields `space_id`; required record fields `space_id`, `body`; accepted fields `body`,
  `node`, `page`, `space_id`, `table_cell`; risk: POST /spaces/{spaceId}/comments (Post a new
  comment on a space or a specific page) executes a live GitBook API operation.
- `update_comment_in_space`: PUT `/spaces/{{ record.space_id }}/comments/{{ record.comment_id }}` -
  kind `update`; body type `json`; path fields `space_id`, `comment_id`; required record fields
  `space_id`, `comment_id`; accepted fields `added_reactions`, `body`, `comment_id`,
  `removed_reactions`, `resolved`, `space_id`; risk: PUT /spaces/{spaceId}/comments/{commentId}
  (Update the body or status of a space comment) executes a live GitBook API operation.
- `delete_comment_in_space`: DELETE `/spaces/{{ record.space_id }}/comments/{{ record.comment_id }}`
  - kind `delete`; body type `none`; path fields `space_id`, `comment_id`; required record fields
  `space_id`, `comment_id`; accepted fields `comment_id`, `space_id`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: DELETE
  /spaces/{spaceId}/comments/{commentId} (Delete a space comment) executes a live GitBook API
  operation.
- `post_comment_reply_in_space`: POST `/spaces/{{ record.space_id }}/comments/{{ record.comment_id
  }}/replies` - kind `create`; body type `json`; path fields `space_id`, `comment_id`; required
  record fields `space_id`, `comment_id`, `body`; accepted fields `body`, `comment_id`, `space_id`;
  risk: POST /spaces/{spaceId}/comments/{commentId}/replies (Post a reply to an existing space
  comment) executes a live GitBook API operation.
- `update_comment_reply_in_space`: PUT `/spaces/{{ record.space_id }}/comments/{{ record.comment_id
  }}/replies/{{ record.comment_reply_id }}` - kind `update`; body type `json`; path fields
  `space_id`, `comment_id`, `comment_reply_id`; required record fields `space_id`, `comment_id`,
  `comment_reply_id`; accepted fields `added_reactions`, `body`, `comment_id`, `comment_reply_id`,
  `removed_reactions`, `space_id`; risk: PUT
  /spaces/{spaceId}/comments/{commentId}/replies/{commentReplyId} (Update the body of a reply to a
  space comment) executes a live GitBook API operation.
- `delete_comment_reply_in_space`: DELETE `/spaces/{{ record.space_id }}/comments/{{
  record.comment_id }}/replies/{{ record.comment_reply_id }}` - kind `delete`; body type `none`;
  path fields `space_id`, `comment_id`, `comment_reply_id`; required record fields `space_id`,
  `comment_id`, `comment_reply_id`; accepted fields `comment_id`, `comment_reply_id`, `space_id`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: DELETE
  /spaces/{spaceId}/comments/{commentId}/replies/{commentReplyId} (Delete a space comment reply)
  executes a live GitBook API operation.
- `update_collection_by_id`: PATCH `/collections/{{ record.collection_id }}` - kind `update`; body
  type `json`; path fields `collection_id`; required record fields `collection_id`; accepted fields
  `collection_id`, `default_level`, `description`, `title`; risk: PATCH /collections/{collectionId}
  (Update a collection) executes a live GitBook API operation.
- `delete_collection_by_id`: DELETE `/collections/{{ record.collection_id }}` - kind `delete`; body
  type `none`; path fields `collection_id`; required record fields `collection_id`; accepted fields
  `collection_id`; missing records treated as success for status `404`; confirmation `destructive`;
  risk: DELETE /collections/{collectionId} (Delete a collection) executes a live GitBook API
  operation.
- `move_collection`: POST `/collections/{{ record.collection_id }}/move` - kind `custom`; body type
  `json`; path fields `collection_id`; required record fields `collection_id`; accepted fields
  `collection_id`, `parent`, `position`; risk: POST /collections/{collectionId}/move (Move a
  collection to a new position.) executes a live GitBook API operation.
- `transfer_collection`: POST `/collections/{{ record.collection_id }}/transfer` - kind `custom`;
  body type `json`; path fields `collection_id`; required record fields `collection_id`,
  `organization`; accepted fields `collection_id`, `organization`; risk: POST
  /collections/{collectionId}/transfer (Transfer a collection) executes a live GitBook API
  operation.
- `invite_to_collection`: POST `/collections/{{ record.collection_id }}/permissions` - kind
  `create`; body type `json`; path fields `collection_id`; required record fields `collection_id`;
  accepted fields `collection_id`, `teams`, `users`; risk: POST
  /collections/{collectionId}/permissions (Invite to a collection) executes a live GitBook API
  operation.
- `update_team_permission_in_collection`: PATCH `/collections/{{ record.collection_id
  }}/permissions/teams/{{ record.team_id }}` - kind `update`; body type `json`; path fields
  `collection_id`, `team_id`; required record fields `collection_id`, `team_id`; accepted fields
  `collection_id`, `role`, `team_id`; risk: PATCH
  /collections/{collectionId}/permissions/teams/{teamId} (Update an org team's permission in a
  collection) executes a live GitBook API operation.
- `remove_team_from_collection`: DELETE `/collections/{{ record.collection_id
  }}/permissions/teams/{{ record.team_id }}` - kind `delete`; body type `none`; path fields
  `collection_id`, `team_id`; required record fields `collection_id`, `team_id`; accepted fields
  `collection_id`, `team_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: DELETE /collections/{collectionId}/permissions/teams/{teamId} (Remove an org
  team from a collection) executes a live GitBook API operation.
- `update_user_permission_in_collection`: PATCH `/collections/{{ record.collection_id
  }}/permissions/users/{{ record.user_id }}` - kind `update`; body type `json`; path fields
  `collection_id`, `user_id`; required record fields `collection_id`, `user_id`; accepted fields
  `collection_id`, `role`, `user_id`; risk: PATCH
  /collections/{collectionId}/permissions/users/{userId} (Update a collection user permission)
  executes a live GitBook API operation.
- `remove_user_from_collection`: DELETE `/collections/{{ record.collection_id
  }}/permissions/users/{{ record.user_id }}` - kind `delete`; body type `none`; path fields
  `collection_id`, `user_id`; required record fields `collection_id`, `user_id`; accepted fields
  `collection_id`, `user_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: DELETE /collections/{collectionId}/permissions/users/{userId} (Remove a user
  from a collection) executes a live GitBook API operation.
- `publish_integration`: POST `/integrations/{{ record.integration_name }}` - kind `create`; body
  type `json`; path fields `integration_name`; required record fields `integration_name`,
  `organization`, `title`, `description`, `script`, `scopes`; accepted fields `blocks`,
  `categories`, `configurations`, `content_security_policy`, `content_sources`, `description`,
  `external_links`, `icon`, `integration_name`, `organization`, `preview_images`, `runtime`,
  `scopes`, `script`, `secrets`, `summary`, `target`, `title`, and 1 more; risk: POST
  /integrations/{integrationName} (Publish an integration) executes a live GitBook API operation.
- `unpublish_integration`: DELETE `/integrations/{{ record.integration_name }}` - kind `delete`;
  body type `none`; path fields `integration_name`; required record fields `integration_name`;
  accepted fields `integration_name`; missing records treated as success for status `404`;
  confirmation `destructive`; risk: DELETE /integrations/{integrationName} (Unpublish an
  integration) executes a live GitBook API operation.
- `install_integration`: POST `/integrations/{{ record.integration_name }}/installations` - kind
  `create`; body type `json`; path fields `integration_name`; required record fields
  `integration_name`, `organization`; accepted fields `integration_name`, `organization`; risk: POST
  /integrations/{integrationName}/installations (Install an integration) executes a live GitBook API
  operation.
- `set_integration_development_mode`: PUT `/integrations/{{ record.integration_name }}/dev` - kind
  `update`; body type `json`; path fields `integration_name`; required record fields
  `integration_name`, `tunnel_url`; accepted fields `all`, `integration_name`, `tunnel_url`; risk:
  PUT /integrations/{integrationName}/dev (Enable integration dev mode) executes a live GitBook API
  operation.
- `disable_integration_development_mode`: DELETE `/integrations/{{ record.integration_name }}/dev` -
  kind `delete`; body type `none`; path fields `integration_name`; required record fields
  `integration_name`; accepted fields `integration_name`; missing records treated as success for
  status `404`; confirmation `destructive`; risk: DELETE /integrations/{integrationName}/dev
  (Disable integration dev mode) executes a live GitBook API operation.
- `render_integration_ui_with_post`: POST `/integrations/{{ record.integration_name }}/render` -
  kind `custom`; body type `json`; path fields `integration_name`; required record fields
  `integration_name`, `component_id`, `props`, `context`; accepted fields `action`, `component_id`,
  `context`, `integration_name`, `props`, `state`; risk: POST /integrations/{integrationName}/render
  (Render an integration UI with POST method) executes a live GitBook API operation.
- `queue_integration_task`: POST `/integrations/{{ record.integration_name }}/tasks` - kind
  `custom`; body type `json`; path fields `integration_name`; required record fields
  `integration_name`, `task`; accepted fields `integration_name`, `schedule`, `task`; risk: POST
  /integrations/{integrationName}/tasks (Queue an integration task) executes a live GitBook API
  operation.
- `update_integration_installation`: PATCH `/integrations/{{ record.integration_name
  }}/installations/{{ record.installation_id }}` - kind `update`; body type `json`; path fields
  `integration_name`, `installation_id`; required record fields `integration_name`,
  `installation_id`; accepted fields `configuration`, `external_ids`, `installation_id`,
  `integration_name`, `site_selection`, `space_selection`; risk: PATCH
  /integrations/{integrationName}/installations/{installationId} (Update an integration
  installation) executes a live GitBook API operation.
- `uninstall_integration`: DELETE `/integrations/{{ record.integration_name }}/installations/{{
  record.installation_id }}` - kind `delete`; body type `none`; path fields `integration_name`,
  `installation_id`; required record fields `integration_name`, `installation_id`; accepted fields
  `installation_id`, `integration_name`; missing records treated as success for status `404`;
  confirmation `destructive`; risk: DELETE
  /integrations/{integrationName}/installations/{installationId} (Uninstall an integration) executes
  a live GitBook API operation.
- `create_integration_installation_token`: POST `/integrations/{{ record.integration_name
  }}/installations/{{ record.installation_id }}/tokens` - kind `create`; body type `none`; path
  fields `integration_name`, `installation_id`; required record fields `integration_name`,
  `installation_id`; accepted fields `installation_id`, `integration_name`; risk: POST
  /integrations/{integrationName}/installations/{installationId}/tokens (Create an integration
  installation API token) executes a live GitBook API operation.
- `install_integration_on_space`: POST `/integrations/{{ record.integration_name }}/installations/{{
  record.installation_id }}/spaces` - kind `create`; body type `json`; path fields
  `integration_name`, `installation_id`; required record fields `integration_name`,
  `installation_id`, `space`; accepted fields `installation_id`, `integration_name`, `space`; risk:
  POST /integrations/{integrationName}/installations/{installationId}/spaces (Install an integration
  on a space) executes a live GitBook API operation.
- `update_integration_space_installation`: PATCH `/integrations/{{ record.integration_name
  }}/installations/{{ record.installation_id }}/spaces/{{ record.space_id }}` - kind `update`; body
  type `json`; path fields `integration_name`, `installation_id`, `space_id`; required record fields
  `integration_name`, `installation_id`, `space_id`; accepted fields `configuration`,
  `external_ids`, `installation_id`, `integration_name`, `space_id`; risk: PATCH
  /integrations/{integrationName}/installations/{installationId}/spaces/{spaceId} (Update an
  integration space installation) executes a live GitBook API operation.
- `uninstall_integration_from_space`: DELETE `/integrations/{{ record.integration_name
  }}/installations/{{ record.installation_id }}/spaces/{{ record.space_id }}` - kind `delete`; body
  type `none`; path fields `integration_name`, `installation_id`, `space_id`; required record fields
  `integration_name`, `installation_id`, `space_id`; accepted fields `installation_id`,
  `integration_name`, `space_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: DELETE
  /integrations/{integrationName}/installations/{installationId}/spaces/{spaceId} (Uninstall an
  integration from a space) executes a live GitBook API operation.
- `install_integration_on_site`: POST `/integrations/{{ record.integration_name }}/installations/{{
  record.installation_id }}/sites` - kind `create`; body type `json`; path fields
  `integration_name`, `installation_id`; required record fields `integration_name`,
  `installation_id`, `site_id`; accepted fields `installation_id`, `integration_name`, `site_id`;
  risk: POST /integrations/{integrationName}/installations/{installationId}/sites (Install an
  integration on a site) executes a live GitBook API operation.
- `update_integration_site_installation`: PATCH `/integrations/{{ record.integration_name
  }}/installations/{{ record.installation_id }}/sites/{{ record.site_id }}` - kind `update`; body
  type `json`; path fields `integration_name`, `installation_id`, `site_id`; required record fields
  `integration_name`, `installation_id`, `site_id`; accepted fields `configuration`, `external_ids`,
  `installation_id`, `integration_name`, `site_id`; risk: PATCH
  /integrations/{integrationName}/installations/{installationId}/sites/{siteId} (Update an
  integration site installation) executes a live GitBook API operation.
- `uninstall_integration_from_site`: DELETE `/integrations/{{ record.integration_name
  }}/installations/{{ record.installation_id }}/sites/{{ record.site_id }}` - kind `delete`; body
  type `none`; path fields `integration_name`, `installation_id`, `site_id`; required record fields
  `integration_name`, `installation_id`, `site_id`; accepted fields `installation_id`,
  `integration_name`, `site_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: DELETE
  /integrations/{integrationName}/installations/{installationId}/sites/{siteId} (Uninstall an
  integration from a site) executes a live GitBook API operation.
- `update_organization_by_id`: PATCH `/orgs/{{ record.organization_id }}` - kind `update`; body type
  `json`; path fields `organization_id`; required record fields `organization_id`; accepted fields
  `ai`, `default_content`, `default_role`, `email_domains`, `hostname`, `invite_links`, `logo`,
  `merge_rules`, `organization_id`, `sso`, `title`; risk: PATCH /orgs/{organizationId} (Update an
  organization) executes a live GitBook API operation.
- `update_member_in_organization_by_id`: PATCH `/orgs/{{ record.organization_id }}/members/{{
  record.user_id }}` - kind `update`; body type `json`; path fields `organization_id`, `user_id`;
  required record fields `organization_id`, `user_id`; accepted fields `organization_id`, `role`,
  `user_id`; risk: PATCH /orgs/{organizationId}/members/{userId} (Update an organization member)
  executes a live GitBook API operation.
- `remove_member_from_organization_by_id`: DELETE `/orgs/{{ record.organization_id }}/members/{{
  record.user_id }}` - kind `delete`; body type `none`; path fields `organization_id`, `user_id`;
  required record fields `organization_id`, `user_id`; accepted fields `organization_id`, `user_id`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: DELETE
  /orgs/{organizationId}/members/{userId} (Delete an organization member) executes a live GitBook
  API operation.
- `update_organization_member_last_seen_at`: POST `/orgs/{{ record.organization_id }}/ping` - kind
  `update`; body type `none`; path fields `organization_id`; required record fields
  `organization_id`; accepted fields `organization_id`; risk: POST /orgs/{organizationId}/ping
  (Update an organization member last seen at) executes a live GitBook API operation.
- `set_user_as_sso_member_for_organization`: POST `/orgs/{{ record.organization_id }}/members/{{
  record.user_id }}/sso` - kind `update`; body type `none`; path fields `organization_id`,
  `user_id`; required record fields `organization_id`, `user_id`; accepted fields `organization_id`,
  `user_id`; risk: POST /orgs/{organizationId}/members/{userId}/sso (Set a user as an SSO member of
  an organization) executes a live GitBook API operation.
- `create_organization_team`: PUT `/orgs/{{ record.organization_id }}/teams` - kind `create`; body
  type `json`; path fields `organization_id`; required record fields `organization_id`, `title`;
  accepted fields `members`, `organization_id`, `title`; risk: PUT /orgs/{organizationId}/teams
  (Create a team) executes a live GitBook API operation.
- `update_team_in_organization_by_id`: PATCH `/orgs/{{ record.organization_id }}/teams/{{
  record.team_id }}` - kind `update`; body type `json`; path fields `organization_id`, `team_id`;
  required record fields `organization_id`, `team_id`, `title`; accepted fields `organization_id`,
  `team_id`, `title`; risk: PATCH /orgs/{organizationId}/teams/{teamId} (Update a team) executes a
  live GitBook API operation.
- `remove_team_from_organization_by_id`: DELETE `/orgs/{{ record.organization_id }}/teams/{{
  record.team_id }}` - kind `delete`; body type `none`; path fields `organization_id`, `team_id`;
  required record fields `organization_id`, `team_id`; accepted fields `organization_id`, `team_id`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: DELETE
  /orgs/{organizationId}/teams/{teamId} (Delete a team) executes a live GitBook API operation.
- `update_members_in_organization_team`: PUT `/orgs/{{ record.organization_id }}/teams/{{
  record.team_id }}/members` - kind `update`; body type `json`; path fields `organization_id`,
  `team_id`; required record fields `organization_id`, `team_id`; accepted fields `add`,
  `memberships`, `organization_id`, `remove`, `team_id`; risk: PUT
  /orgs/{organizationId}/teams/{teamId}/members (Updates members of a team) executes a live GitBook
  API operation.
- `add_member_to_organization_team_by_id`: PUT `/orgs/{{ record.organization_id }}/teams/{{
  record.team_id }}/members/{{ record.user_id }}` - kind `create`; body type `json`; path fields
  `organization_id`, `team_id`, `user_id`; required record fields `organization_id`, `team_id`,
  `user_id`; accepted fields `organization_id`, `role`, `team_id`, `user_id`; risk: PUT
  /orgs/{organizationId}/teams/{teamId}/members/{userId} (Add a team member) executes a live GitBook
  API operation.
- `delete_member_from_organization_team_by_id`: DELETE `/orgs/{{ record.organization_id }}/teams/{{
  record.team_id }}/members/{{ record.user_id }}` - kind `delete`; body type `none`; path fields
  `organization_id`, `team_id`, `user_id`; required record fields `organization_id`, `team_id`,
  `user_id`; accepted fields `organization_id`, `team_id`, `user_id`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: DELETE
  /orgs/{organizationId}/teams/{teamId}/members/{userId} (Delete a team member) executes a live
  GitBook API operation.
- `invite_users_to_organization`: POST `/orgs/{{ record.organization_id }}/invites` - kind `create`;
  body type `json`; path fields `organization_id`; required record fields `organization_id`,
  `emails`; accepted fields `emails`, `organization_id`, `role`, `sso`; risk: POST
  /orgs/{organizationId}/invites (Invite users in an organization) executes a live GitBook API
  operation.
- `join_organization_with_invite`: POST `/orgs/{{ record.organization_id }}/invites/{{
  record.invite_id }}` - kind `custom`; body type `none`; path fields `organization_id`,
  `invite_id`; required record fields `organization_id`, `invite_id`; accepted fields `invite_id`,
  `organization_id`; risk: POST /orgs/{organizationId}/invites/{inviteId} (Join an organization with
  an invite) executes a live GitBook API operation.
- `create_organization_invite`: POST `/orgs/{{ record.organization_id }}/link-invites` - kind
  `create`; body type `json`; path fields `organization_id`; required record fields
  `organization_id`; accepted fields `collection`, `level`, `organization_id`, `role`, `space`;
  risk: POST /orgs/{organizationId}/link-invites (Create an organization invite) executes a live
  GitBook API operation.
- `update_organization_invite_by_id`: PATCH `/orgs/{{ record.organization_id }}/link-invites/{{
  record.invite_id }}` - kind `update`; body type `json`; path fields `organization_id`,
  `invite_id`; required record fields `organization_id`, `invite_id`; accepted fields `invite_id`,
  `level`, `organization_id`, `role`; risk: PATCH /orgs/{organizationId}/link-invites/{inviteId}
  (Update an organization invite) executes a live GitBook API operation.
- `delete_organization_invite_by_id`: DELETE `/orgs/{{ record.organization_id }}/link-invites/{{
  record.invite_id }}` - kind `delete`; body type `none`; path fields `organization_id`,
  `invite_id`; required record fields `organization_id`, `invite_id`; accepted fields `invite_id`,
  `organization_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: DELETE /orgs/{organizationId}/link-invites/{inviteId} (Deletes an
  organization invite.) executes a live GitBook API operation.
- `join_organization`: POST `/orgs/{{ record.organization_id }}/join` - kind `custom`; body type
  `none`; path fields `organization_id`; required record fields `organization_id`; accepted fields
  `organization_id`; risk: POST /orgs/{organizationId}/join (Join an organization) executes a live
  GitBook API operation.
- `create_space`: POST `/orgs/{{ record.organization_id }}/spaces` - kind `create`; body type
  `json`; path fields `organization_id`; required record fields `organization_id`; accepted fields
  `computed_source`, `edit_mode`, `emoji`, `empty`, `language`, `organization_id`, `parent`,
  `template`, `title`; risk: POST /orgs/{organizationId}/spaces (Create a new documentation space in
  an organization) executes a live GitBook API operation.
- `create_collection`: POST `/orgs/{{ record.organization_id }}/collections` - kind `create`; body
  type `json`; path fields `organization_id`; required record fields `organization_id`; accepted
  fields `organization_id`, `parent`, `title`; risk: POST /orgs/{organizationId}/collections (Create
  a collection) executes a live GitBook API operation.
- `create_organization_saml_provider`: POST `/orgs/{{ record.organization_id }}/saml` - kind
  `create`; body type `json`; path fields `organization_id`; required record fields
  `organization_id`, `label`; accepted fields `certificate`, `default_role`, `default_team`,
  `entity_id`, `label`, `organization_id`, `sso_url`; risk: POST /orgs/{organizationId}/saml (Create
  a new SAML provider) executes a live GitBook API operation.
- `update_organization_saml_provider`: PATCH `/orgs/{{ record.organization_id }}/saml/{{
  record.saml_provider_id }}` - kind `update`; body type `json`; path fields `organization_id`,
  `saml_provider_id`; required record fields `organization_id`, `saml_provider_id`; accepted fields
  `certificate`, `default_role`, `default_team`, `entity_id`, `label`, `organization_id`,
  `saml_provider_id`, `sso_url`; risk: PATCH /orgs/{organizationId}/saml/{samlProviderId} (Update a
  SAML provider) executes a live GitBook API operation.
- `delete_organization_saml_provider`: DELETE `/orgs/{{ record.organization_id }}/saml/{{
  record.saml_provider_id }}` - kind `delete`; body type `none`; path fields `organization_id`,
  `saml_provider_id`; required record fields `organization_id`, `saml_provider_id`; accepted fields
  `organization_id`, `saml_provider_id`; missing records treated as success for status `404`;
  confirmation `destructive`; risk: DELETE /orgs/{organizationId}/saml/{samlProviderId} (Delete a
  SAML provider) executes a live GitBook API operation.
- `ask_in_organization`: POST `/orgs/{{ record.organization_id }}/ask` - kind `custom`; body type
  `json`; path fields `organization_id`; required record fields `organization_id`, `query`; accepted
  fields `organization_id`, `previous_queries`, `query`; risk: POST /orgs/{organizationId}/ask (Ask
  a question in an organization) executes a live GitBook API operation.
- `create_open_api_spec`: POST `/orgs/{{ record.organization_id }}/openapi` - kind `create`; body
  type `json`; path fields `organization_id`; required record fields `organization_id`, `source`,
  `slug`; accepted fields `organization_id`, `slug`, `source`; risk: POST
  /orgs/{organizationId}/openapi (Create an OpenAPI spec) executes a live GitBook API operation.
- `create_or_update_open_api_spec_by_slug`: PUT `/orgs/{{ record.organization_id }}/openapi/{{
  record.spec_slug }}` - kind `upsert`; body type `json`; path fields `organization_id`,
  `spec_slug`; required record fields `organization_id`, `spec_slug`, `source`; accepted fields
  `organization_id`, `source`, `spec_slug`; risk: PUT /orgs/{organizationId}/openapi/{specSlug}
  (Create or update an OpenAPI spec) executes a live GitBook API operation.
- `update_open_api_spec_by_slug`: PATCH `/orgs/{{ record.organization_id }}/openapi/{{
  record.spec_slug }}` - kind `update`; body type `json`; path fields `organization_id`,
  `spec_slug`; required record fields `organization_id`, `spec_slug`, `visibility`; accepted fields
  `organization_id`, `spec_slug`, `visibility`; risk: PATCH
  /orgs/{organizationId}/openapi/{specSlug} (Update OpenAPI spec visibility) executes a live GitBook
  API operation.
- `delete_open_api_spec_by_slug`: DELETE `/orgs/{{ record.organization_id }}/openapi/{{
  record.spec_slug }}` - kind `delete`; body type `none`; path fields `organization_id`,
  `spec_slug`; required record fields `organization_id`, `spec_slug`; accepted fields
  `organization_id`, `spec_slug`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: DELETE /orgs/{organizationId}/openapi/{specSlug} (Delete an OpenAPI spec)
  executes a live GitBook API operation.
- `update_organization_agent_instructions`: PUT `/orgs/{{ record.organization_id
  }}/agent-instructions` - kind `update`; body type `json`; path fields `organization_id`; required
  record fields `organization_id`, `instructions`; accepted fields `instructions`,
  `organization_id`; risk: PUT /orgs/{organizationId}/agent-instructions (Update Docs agent
  instructions for an organization) executes a live GitBook API operation.
- `create_translation`: POST `/orgs/{{ record.organization_id }}/translations` - kind `create`; body
  type `json`; path fields `organization_id`; required record fields `organization_id`, `language`,
  `source`; accepted fields `instructions`, `language`, `organization_id`, `source`; risk: POST
  /orgs/{organizationId}/translations (Create a translation) executes a live GitBook API operation.
- `update_translation`: PUT `/orgs/{{ record.organization_id }}/translations/{{
  record.translation_id }}` - kind `update`; body type `json`; path fields `organization_id`,
  `translation_id`; required record fields `organization_id`, `translation_id`, `instructions`;
  accepted fields `instructions`, `organization_id`, `translation_id`; risk: PUT
  /orgs/{organizationId}/translations/{translationId} (Update a translation) executes a live GitBook
  API operation.
- `delete_translation`: DELETE `/orgs/{{ record.organization_id }}/translations/{{
  record.translation_id }}` - kind `delete`; body type `none`; path fields `organization_id`,
  `translation_id`; required record fields `organization_id`, `translation_id`; accepted fields
  `organization_id`, `translation_id`; missing records treated as success for status `404`;
  confirmation `destructive`; risk: DELETE /orgs/{organizationId}/translations/{translationId}
  (Delete a translation) executes a live GitBook API operation.
- `run_translation`: POST `/orgs/{{ record.organization_id }}/translations/{{ record.translation_id
  }}/run` - kind `custom`; body type `none`; path fields `organization_id`, `translation_id`;
  required record fields `organization_id`, `translation_id`; accepted fields `organization_id`,
  `translation_id`; risk: POST /orgs/{organizationId}/translations/{translationId}/run (Run a
  translation again) executes a live GitBook API operation.
- `update_glossary_entries`: PUT `/orgs/{{ record.organization_id }}/translations-glossary` - kind
  `update`; body type `json`; path fields `organization_id`; required record fields
  `organization_id`, `operations`; accepted fields `operations`, `organization_id`; risk: PUT
  /orgs/{organizationId}/translations-glossary (Update glossary entries) executes a live GitBook API
  operation.
- `generate_storage_upload_url`: POST `/orgs/{{ record.organization_id }}/storage/upload` - kind
  `custom`; body type `json`; path fields `organization_id`; required record fields
  `organization_id`, `file`, `kind`; accepted fields `file`, `kind`, `organization_id`; risk: POST
  /orgs/{organizationId}/storage/upload (Create a signed URL to upload a file) executes a live
  GitBook API operation.
- `create_custom_font`: PUT `/orgs/{{ record.organization_id }}/fonts` - kind `create`; body type
  `json`; path fields `organization_id`; required record fields `organization_id`, `font_family`,
  `font_faces`; accepted fields `font_faces`, `font_family`, `organization_id`; risk: PUT
  /orgs/{organizationId}/fonts (Create a custom font) executes a live GitBook API operation.
- `update_custom_font`: POST `/orgs/{{ record.organization_id }}/fonts/{{ record.font_id }}` - kind
  `update`; body type `json`; path fields `organization_id`, `font_id`; required record fields
  `organization_id`, `font_id`; accepted fields `font_faces`, `font_family`, `font_id`,
  `organization_id`; risk: POST /orgs/{organizationId}/fonts/{fontId} (Update a custom font)
  executes a live GitBook API operation.
- `delete_custom_font`: DELETE `/orgs/{{ record.organization_id }}/fonts/{{ record.font_id }}` -
  kind `delete`; body type `none`; path fields `organization_id`, `font_id`; required record fields
  `organization_id`, `font_id`; accepted fields `font_id`, `organization_id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: DELETE
  /orgs/{organizationId}/fonts/{fontId} (Delete a custom font) executes a live GitBook API
  operation.
- `start_import_run`: POST `/org/{{ record.organization_id }}/imports` - kind `create`; body type
  `json`; path fields `organization_id`; required record fields `organization_id`, `source`,
  `target`; accepted fields `enhance`, `organization_id`, `source`, `target`; risk: POST
  /org/{organizationId}/imports (Import content into a space from a website) executes a live GitBook
  API operation.
- `cancel_import_run`: POST `/org/{{ record.organization_id }}/imports/{{ record.import_run_id
  }}/cancel` - kind `custom`; body type `none`; path fields `organization_id`, `import_run_id`;
  required record fields `organization_id`, `import_run_id`; accepted fields `import_run_id`,
  `organization_id`; risk: POST /org/{organizationId}/imports/{importRunId}/cancel (Cancel an import
  run) executes a live GitBook API operation.
- `create_site`: POST `/orgs/{{ record.organization_id }}/sites` - kind `create`; body type `json`;
  path fields `organization_id`; required record fields `organization_id`; accepted fields
  `organization_id`, `spaces`, `title`, `type`, `visibility`; risk: POST
  /orgs/{organizationId}/sites (Create a new documentation site in an organization) executes a live
  GitBook API operation.
- `update_site_by_id`: PATCH `/orgs/{{ record.organization_id }}/sites/{{ record.site_id }}` - kind
  `update`; body type `json`; path fields `organization_id`, `site_id`; required record fields
  `organization_id`, `site_id`; accepted fields `adaptive_content`, `basename`, `default_level`,
  `default_site_section`, `default_site_space`, `organization_id`, `permissions_model`, `proxy`,
  `site_id`, `styleguide`, `title`, `visibility`; risk: PATCH /orgs/{organizationId}/sites/{siteId}
  (Update the properties of a documentation site) executes a live GitBook API operation.
- `delete_site_by_id`: DELETE `/orgs/{{ record.organization_id }}/sites/{{ record.site_id }}` - kind
  `delete`; body type `none`; path fields `organization_id`, `site_id`; required record fields
  `organization_id`, `site_id`; accepted fields `organization_id`, `site_id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: DELETE
  /orgs/{organizationId}/sites/{siteId} (Delete a site) executes a live GitBook API operation.
- `update_site_adaptive_schema`: PUT `/orgs/{{ record.organization_id }}/sites/{{ record.site_id
  }}/adaptive-schema` - kind `update`; body type `json`; path fields `organization_id`, `site_id`;
  required record fields `organization_id`, `site_id`, `json_schema`; accepted fields `json_schema`,
  `organization_id`, `site_id`; risk: PUT /orgs/{organizationId}/sites/{siteId}/adaptive-schema
  (Update the visitor attributes JSON schema for an adaptive content site) executes a live GitBook
  API operation.
- `publish_site`: POST `/orgs/{{ record.organization_id }}/sites/{{ record.site_id }}/publish` -
  kind `create`; body type `none`; path fields `organization_id`, `site_id`; required record fields
  `organization_id`, `site_id`; accepted fields `organization_id`, `site_id`; risk: POST
  /orgs/{organizationId}/sites/{siteId}/publish (Publish a site to make it publicly accessible)
  executes a live GitBook API operation.
- `unpublish_site`: POST `/orgs/{{ record.organization_id }}/sites/{{ record.site_id }}/unpublish` -
  kind `delete`; body type `none`; path fields `organization_id`, `site_id`; required record fields
  `organization_id`, `site_id`; accepted fields `organization_id`, `site_id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: POST
  /orgs/{organizationId}/sites/{siteId}/unpublish (Take a site offline by unpublishing it) executes
  a live GitBook API operation.
- `create_site_share_link`: POST `/orgs/{{ record.organization_id }}/sites/{{ record.site_id
  }}/share-links` - kind `create`; body type `json`; path fields `organization_id`, `site_id`;
  required record fields `organization_id`, `site_id`, `name`; accepted fields `name`,
  `organization_id`, `site_id`; risk: POST /orgs/{organizationId}/sites/{siteId}/share-links (Create
  a private share link for a site) executes a live GitBook API operation.
- `update_site_share_link_by_id`: PATCH `/orgs/{{ record.organization_id }}/sites/{{ record.site_id
  }}/share-links/{{ record.share_link_id }}` - kind `update`; body type `json`; path fields
  `organization_id`, `site_id`, `share_link_id`; required record fields `organization_id`,
  `site_id`, `share_link_id`; accepted fields `active`, `name`, `organization_id`, `share_link_id`,
  `site_id`; risk: PATCH /orgs/{organizationId}/sites/{siteId}/share-links/{shareLinkId} (Update a
  private share link for a site) executes a live GitBook API operation.
- `delete_site_share_link_by_id`: DELETE `/orgs/{{ record.organization_id }}/sites/{{ record.site_id
  }}/share-links/{{ record.share_link_id }}` - kind `delete`; body type `none`; path fields
  `organization_id`, `site_id`, `share_link_id`; required record fields `organization_id`,
  `site_id`, `share_link_id`; accepted fields `organization_id`, `share_link_id`, `site_id`; missing
  records treated as success for status `404`; confirmation `destructive`; risk: DELETE
  /orgs/{organizationId}/sites/{siteId}/share-links/{shareLinkId} (Deletes a share link) executes a
  live GitBook API operation.
- `sort_site_structure`: PATCH `/orgs/{{ record.organization_id }}/sites/{{ record.site_id
  }}/structure/sort` - kind `update`; body type `json`; path fields `organization_id`, `site_id`;
  required record fields `organization_id`, `site_id`, `item`, `position`; accepted fields `item`,
  `organization_id`, `position`, `site_id`; risk: PATCH
  /orgs/{organizationId}/sites/{siteId}/structure/sort (Move a site space, section, or section group
  to a new position) executes a live GitBook API operation.
- `update_site_publishing_auth_by_id`: PATCH `/orgs/{{ record.organization_id }}/sites/{{
  record.site_id }}/publishing/auth` - kind `update`; body type `json`; path fields
  `organization_id`, `site_id`; required record fields `organization_id`, `site_id`; accepted fields
  `backend`, `fallback_url`, `logout_url`, `organization_id`, `site_id`; risk: PATCH
  /orgs/{organizationId}/sites/{siteId}/publishing/auth (Update the published content authentication
  configuration for a site) executes a live GitBook API operation.
- `regenerate_site_publishing_auth_by_id`: POST `/orgs/{{ record.organization_id }}/sites/{{
  record.site_id }}/publishing/auth/regenerate` - kind `custom`; body type `none`; path fields
  `organization_id`, `site_id`; required record fields `organization_id`, `site_id`; accepted fields
  `organization_id`, `site_id`; risk: POST
  /orgs/{organizationId}/sites/{siteId}/publishing/auth/regenerate (Regenerate the private key for a
  site's published content authentication) executes a live GitBook API operation.
- `update_site_customization_by_id`: PUT `/orgs/{{ record.organization_id }}/sites/{{ record.site_id
  }}/customization` - kind `update`; body type `json`; path fields `organization_id`, `site_id`;
  required record fields `organization_id`, `site_id`, `styling`, `internationalization`, `favicon`,
  `header`, `footer`, `themes`, `feedback`, `ai`, `advanced_customization`, `trademark`, and 7 more;
  accepted fields `advanced_customization`, `ai`, `announcement`, `external_links`, `favicon`,
  `feedback`, `footer`, `header`, `insights`, `internationalization`, `localized_title`,
  `organization_id`, `page_actions`, `pagination`, `privacy_policy`, `site_id`, `social_accounts`,
  `social_preview`, and 4 more; risk: PUT /orgs/{organizationId}/sites/{siteId}/customization
  (Update the branding and visual customization settings for a site) executes a live GitBook API
  operation.
- `add_space_to_site`: POST `/orgs/{{ record.organization_id }}/sites/{{ record.site_id
  }}/site-spaces` - kind `create`; body type `json`; path fields `organization_id`, `site_id`;
  required record fields `organization_id`, `site_id`, `space_id`; accepted fields `draft`,
  `organization_id`, `section_id`, `site_id`, `space_id`; risk: POST
  /orgs/{organizationId}/sites/{siteId}/site-spaces (Add a space to a site as a content source)
  executes a live GitBook API operation.
- `add_section_group_to_site`: POST `/orgs/{{ record.organization_id }}/sites/{{ record.site_id
  }}/section-groups` - kind `create`; body type `json`; path fields `organization_id`, `site_id`;
  required record fields `organization_id`, `site_id`, `title`; accepted fields `draft`, `icon`,
  `organization_id`, `parent`, `sections`, `site_id`, `title`; risk: POST
  /orgs/{organizationId}/sites/{siteId}/section-groups (Add a section group to a site's navigation
  structure) executes a live GitBook API operation.
- `update_site_section_group_by_id`: PATCH `/orgs/{{ record.organization_id }}/sites/{{
  record.site_id }}/section-groups/{{ record.site_section_group_id }}` - kind `update`; body type
  `json`; path fields `organization_id`, `site_id`, `site_section_group_id`; required record fields
  `organization_id`, `site_id`, `site_section_group_id`; accepted fields `draft`, `icon`,
  `localized_title`, `organization_id`, `site_id`, `site_section_group_id`, `title`; risk: PATCH
  /orgs/{organizationId}/sites/{siteId}/section-groups/{siteSectionGroupId} (Update a section group
  in a site's navigation structure) executes a live GitBook API operation.
- `delete_site_section_group_by_id`: DELETE `/orgs/{{ record.organization_id }}/sites/{{
  record.site_id }}/section-groups/{{ record.site_section_group_id }}` - kind `delete`; body type
  `none`; path fields `organization_id`, `site_id`, `site_section_group_id`; required record fields
  `organization_id`, `site_id`, `site_section_group_id`; accepted fields `organization_id`,
  `site_id`, `site_section_group_id`; missing records treated as success for status `404`;
  confirmation `destructive`; risk: DELETE
  /orgs/{organizationId}/sites/{siteId}/section-groups/{siteSectionGroupId} (Delete a site section
  group) executes a live GitBook API operation.
- `add_section_to_site`: POST `/orgs/{{ record.organization_id }}/sites/{{ record.site_id
  }}/sections` - kind `create`; body type `json`; path fields `organization_id`, `site_id`; required
  record fields `organization_id`, `site_id`, `space_id`; accepted fields `draft`, `icon`,
  `organization_id`, `site_id`, `site_section_group_id`, `space_id`, `title`; risk: POST
  /orgs/{organizationId}/sites/{siteId}/sections (Add a new navigation section to a site backed by a
  space) executes a live GitBook API operation.
- `update_site_section_by_id`: PATCH `/orgs/{{ record.organization_id }}/sites/{{ record.site_id
  }}/sections/{{ record.site_section_id }}` - kind `update`; body type `json`; path fields
  `organization_id`, `site_id`, `site_section_id`; required record fields `organization_id`,
  `site_id`, `site_section_id`; accepted fields `condition`, `default_site_space`, `description`,
  `draft`, `icon`, `localized_description`, `localized_title`, `organization_id`, `path`, `site_id`,
  `site_section_group_id`, `site_section_id`, `title`; risk: PATCH
  /orgs/{organizationId}/sites/{siteId}/sections/{siteSectionId} (Update a navigation section in a
  site) executes a live GitBook API operation.
- `delete_site_section_by_id`: DELETE `/orgs/{{ record.organization_id }}/sites/{{ record.site_id
  }}/sections/{{ record.site_section_id }}` - kind `delete`; body type `none`; path fields
  `organization_id`, `site_id`, `site_section_id`; required record fields `organization_id`,
  `site_id`, `site_section_id`; accepted fields `organization_id`, `site_id`, `site_section_id`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: DELETE
  /orgs/{organizationId}/sites/{siteId}/sections/{siteSectionId} (Delete a site section) executes a
  live GitBook API operation.
- `search_site_content`: POST `/orgs/{{ record.organization_id }}/sites/{{ record.site_id }}/search`
  - kind `custom`; body type `json`; path fields `organization_id`, `site_id`; required record
  fields `organization_id`, `site_id`, `query`; accepted fields `mode`, `organization_id`, `query`,
  `scope`, `site_id`, `site_space_id`, `site_space_ids`; risk: POST
  /orgs/{organizationId}/sites/{siteId}/search (Full-text search across all content in a site)
  executes a live GitBook API operation.
- `stream_ask_in_site`: POST `/orgs/{{ record.organization_id }}/sites/{{ record.site_id }}/ask` -
  kind `custom`; body type `json`; path fields `organization_id`, `site_id`; required record fields
  `organization_id`, `site_id`, `question`, `scope`; accepted fields `context`, `organization_id`,
  `question`, `scope`, `session`, `site_id`; risk: POST /orgs/{organizationId}/sites/{siteId}/ask
  (Ask a question in a site) executes a live GitBook API operation.
- `create_site_scan`: POST `/orgs/{{ record.organization_id }}/sites/{{ record.site_id }}/scans` -
  kind `create`; body type `json`; path fields `organization_id`, `site_id`; required record fields
  `organization_id`, `site_id`, `topic`; accepted fields `organization_id`, `site_id`, `topic`;
  risk: POST /orgs/{organizationId}/sites/{siteId}/scans (Enqueue a new site scan) executes a live
  GitBook API operation.
- `update_site_finding_by_id`: PATCH `/orgs/{{ record.organization_id }}/sites/{{ record.site_id
  }}/findings/{{ record.site_finding_id }}` - kind `update`; body type `json`; path fields
  `organization_id`, `site_id`, `site_finding_id`; required record fields `organization_id`,
  `site_id`, `site_finding_id`, `status`; accepted fields `organization_id`, `site_finding_id`,
  `site_id`, `status`; risk: PATCH /orgs/{organizationId}/sites/{siteId}/findings/{siteFindingId}
  (Update a site finding) executes a live GitBook API operation.
- `trigger_change_requests_for_site_finding`: POST `/orgs/{{ record.organization_id }}/sites/{{
  record.site_id }}/findings/{{ record.site_finding_id }}/change-requests` - kind `custom`; body
  type `none`; path fields `organization_id`, `site_id`, `site_finding_id`; required record fields
  `organization_id`, `site_id`, `site_finding_id`; accepted fields `organization_id`,
  `site_finding_id`, `site_id`; risk: POST
  /orgs/{organizationId}/sites/{siteId}/findings/{siteFindingId}/change-requests (Process a site
  finding into change requests) executes a live GitBook API operation.
- `create_site_context_connection`: POST `/orgs/{{ record.organization_id }}/sites/{{ record.site_id
  }}/context-connections` - kind `create`; body type `json`; path fields `organization_id`,
  `site_id`; required record fields `organization_id`, `site_id`; accepted fields `connector`,
  `organization_id`, `setup_settings`, `site_id`, `usage_settings`; risk: POST
  /orgs/{organizationId}/sites/{siteId}/context-connections (Create a context connection) executes a
  live GitBook API operation.
- `update_site_context_connection_by_id`: PATCH `/orgs/{{ record.organization_id }}/sites/{{
  record.site_id }}/context-connections/{{ record.site_context_connection_id }}` - kind `update`;
  body type `json`; path fields `organization_id`, `site_id`, `site_context_connection_id`; required
  record fields `organization_id`, `site_id`, `site_context_connection_id`; accepted fields
  `connector`, `label`, `organization_id`, `setup_settings`, `site_context_connection_id`,
  `site_id`, `usage_settings`; risk: PATCH
  /orgs/{organizationId}/sites/{siteId}/context-connections/{siteContextConnectionId} (Update a
  context connection) executes a live GitBook API operation.
- `delete_site_context_connection_by_id`: DELETE `/orgs/{{ record.organization_id }}/sites/{{
  record.site_id }}/context-connections/{{ record.site_context_connection_id }}` - kind `delete`;
  body type `none`; path fields `organization_id`, `site_id`, `site_context_connection_id`; required
  record fields `organization_id`, `site_id`, `site_context_connection_id`; accepted fields
  `organization_id`, `site_context_connection_id`, `site_id`; missing records treated as success for
  status `404`; confirmation `destructive`; risk: DELETE
  /orgs/{organizationId}/sites/{siteId}/context-connections/{siteContextConnectionId} (Delete a
  context connection) executes a live GitBook API operation.
- `sync_site_context_connection`: POST `/orgs/{{ record.organization_id }}/sites/{{ record.site_id
  }}/context-connections/{{ record.site_context_connection_id }}/sync` - kind `custom`; body type
  `none`; path fields `organization_id`, `site_id`, `site_context_connection_id`; required record
  fields `organization_id`, `site_id`, `site_context_connection_id`; accepted fields
  `organization_id`, `site_context_connection_id`, `site_id`; risk: POST
  /orgs/{organizationId}/sites/{siteId}/context-connections/{siteContextConnectionId}/sync (Trigger
  a sync for a context connection) executes a live GitBook API operation.
- `update_site_topic_by_id`: PATCH `/orgs/{{ record.organization_id }}/sites/{{ record.site_id
  }}/topics/{{ record.site_topic_id }}` - kind `update`; body type `json`; path fields
  `organization_id`, `site_id`, `site_topic_id`; required record fields `organization_id`,
  `site_id`, `site_topic_id`, `usage_settings`; accepted fields `organization_id`, `site_id`,
  `site_topic_id`, `usage_settings`; risk: PATCH
  /orgs/{organizationId}/sites/{siteId}/topics/{siteTopicId} (Update a topic) executes a live
  GitBook API operation.
- `delete_site_topic_findings`: DELETE `/orgs/{{ record.organization_id }}/sites/{{ record.site_id
  }}/topics/{{ record.site_topic_id }}/findings` - kind `delete`; body type `none`; path fields
  `organization_id`, `site_id`, `site_topic_id`; required record fields `organization_id`,
  `site_id`, `site_topic_id`; accepted fields `organization_id`, `site_id`, `site_topic_id`; missing
  records treated as success for status `404`; confirmation `destructive`; risk: DELETE
  /orgs/{organizationId}/sites/{siteId}/topics/{siteTopicId}/findings (Delete all findings for a
  topic) executes a live GitBook API operation.
- `update_site_space_by_id`: PATCH `/orgs/{{ record.organization_id }}/sites/{{ record.site_id
  }}/site-spaces/{{ record.site_space_id }}` - kind `update`; body type `json`; path fields
  `organization_id`, `site_id`, `site_space_id`; required record fields `organization_id`,
  `site_id`, `site_space_id`; accepted fields `condition`, `draft`, `hidden`, `organization_id`,
  `path`, `site_id`, `site_space_id`, `space_id`; risk: PATCH
  /orgs/{organizationId}/sites/{siteId}/site-spaces/{siteSpaceId} (Update a space linked to a site)
  executes a live GitBook API operation.
- `delete_site_space_by_id`: DELETE `/orgs/{{ record.organization_id }}/sites/{{ record.site_id
  }}/site-spaces/{{ record.site_space_id }}` - kind `delete`; body type `none`; path fields
  `organization_id`, `site_id`, `site_space_id`; required record fields `organization_id`,
  `site_id`, `site_space_id`; accepted fields `organization_id`, `site_id`, `site_space_id`; missing
  records treated as success for status `404`; confirmation `destructive`; risk: DELETE
  /orgs/{organizationId}/sites/{siteId}/site-spaces/{siteSpaceId} (Delete a site space) executes a
  live GitBook API operation.
- `override_site_space_customization_by_id`: PATCH `/orgs/{{ record.organization_id }}/sites/{{
  record.site_id }}/site-spaces/{{ record.site_space_id }}/customization` - kind `update`; body type
  `json`; path fields `organization_id`, `site_id`, `site_space_id`; required record fields
  `organization_id`, `site_id`, `site_space_id`; accepted fields `announcement`, `external_links`,
  `favicon`, `feedback`, `footer`, `header`, `internationalization`, `localized_title`,
  `organization_id`, `pagination`, `privacy_policy`, `site_id`, `site_space_id`, `social_preview`,
  `styling`, `themes`, `title`, `trademark`; risk: PATCH
  /orgs/{organizationId}/sites/{siteId}/site-spaces/{siteSpaceId}/customization (Override branding
  and customization settings for a specific site space) executes a live GitBook API operation.
- `delete_site_space_customization_by_id`: DELETE `/orgs/{{ record.organization_id }}/sites/{{
  record.site_id }}/site-spaces/{{ record.site_space_id }}/customization` - kind `delete`; body type
  `none`; path fields `organization_id`, `site_id`, `site_space_id`; required record fields
  `organization_id`, `site_id`, `site_space_id`; accepted fields `organization_id`, `site_id`,
  `site_space_id`; missing records treated as success for status `404`; confirmation `destructive`;
  risk: DELETE /orgs/{organizationId}/sites/{siteId}/site-spaces/{siteSpaceId}/customization (Delete
  a site space customization settings) executes a live GitBook API operation.
- `move_site_section_group`: POST `/orgs/{{ record.organization_id }}/sites/{{ record.site_id
  }}/section-groups/{{ record.site_section_group_id }}/move` - kind `custom`; body type `json`; path
  fields `organization_id`, `site_id`, `site_section_group_id`; required record fields
  `organization_id`, `site_id`, `site_section_group_id`; accepted fields `organization_id`,
  `position`, `site_id`, `site_section_group_id`; risk: POST
  /orgs/{organizationId}/sites/{siteId}/section-groups/{siteSectionGroupId}/move (Move a site
  section group to a new position. (Deprecated) use sortSiteStructure instead.) executes a live
  GitBook API operation.
- `move_site_section`: POST `/orgs/{{ record.organization_id }}/sites/{{ record.site_id
  }}/sections/{{ record.site_section_id }}/move` - kind `custom`; body type `json`; path fields
  `organization_id`, `site_id`, `site_section_id`; required record fields `organization_id`,
  `site_id`, `site_section_id`; accepted fields `organization_id`, `position`, `site_id`,
  `site_section_id`; risk: POST /orgs/{organizationId}/sites/{siteId}/sections/{siteSectionId}/move
  (Move a site section to a new position. (Deprecated) use sortSiteStructure instead.) executes a
  live GitBook API operation.
- `move_site_space`: POST `/orgs/{{ record.organization_id }}/sites/{{ record.site_id
  }}/site-spaces/{{ record.site_space_id }}/move` - kind `custom`; body type `json`; path fields
  `organization_id`, `site_id`, `site_space_id`; required record fields `organization_id`,
  `site_id`, `site_space_id`; accepted fields `organization_id`, `position`, `site_id`,
  `site_space_id`; risk: POST /orgs/{organizationId}/sites/{siteId}/site-spaces/{siteSpaceId}/move
  (Move a site space to a new position. (Deprecated) use sortSiteStructure instead.) executes a live
  GitBook API operation.
- `invite_to_site`: POST `/orgs/{{ record.organization_id }}/sites/{{ record.site_id }}/permissions`
  - kind `create`; body type `json`; path fields `organization_id`, `site_id`; required record
  fields `organization_id`, `site_id`; accepted fields `organization_id`, `site_id`, `teams`,
  `users`; risk: POST /orgs/{organizationId}/sites/{siteId}/permissions (Invite a user or a team to
  a site) executes a live GitBook API operation.
- `update_user_permission_in_site`: PATCH `/orgs/{{ record.organization_id }}/sites/{{
  record.site_id }}/permissions/users/{{ record.user_id }}` - kind `update`; body type `json`; path
  fields `organization_id`, `site_id`, `user_id`; required record fields `organization_id`,
  `site_id`, `user_id`; accepted fields `organization_id`, `role`, `site_id`, `user_id`; risk: PATCH
  /orgs/{organizationId}/sites/{siteId}/permissions/users/{userId} (Update site user permissions)
  executes a live GitBook API operation.
- `remove_user_from_site`: DELETE `/orgs/{{ record.organization_id }}/sites/{{ record.site_id
  }}/permissions/users/{{ record.user_id }}` - kind `delete`; body type `none`; path fields
  `organization_id`, `site_id`, `user_id`; required record fields `organization_id`, `site_id`,
  `user_id`; accepted fields `organization_id`, `site_id`, `user_id`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: DELETE
  /orgs/{organizationId}/sites/{siteId}/permissions/users/{userId} (Remove a site user) executes a
  live GitBook API operation.
- `update_team_permission_in_site`: PATCH `/orgs/{{ record.organization_id }}/sites/{{
  record.site_id }}/permissions/teams/{{ record.team_id }}` - kind `update`; body type `json`; path
  fields `organization_id`, `site_id`, `team_id`; required record fields `organization_id`,
  `site_id`, `team_id`; accepted fields `organization_id`, `role`, `site_id`, `team_id`; risk: PATCH
  /orgs/{organizationId}/sites/{siteId}/permissions/teams/{teamId} (Update an org team's permission
  in a site) executes a live GitBook API operation.
- `remove_team_from_site`: DELETE `/orgs/{{ record.organization_id }}/sites/{{ record.site_id
  }}/permissions/teams/{{ record.team_id }}` - kind `delete`; body type `none`; path fields
  `organization_id`, `site_id`, `team_id`; required record fields `organization_id`, `site_id`,
  `team_id`; accepted fields `organization_id`, `site_id`, `team_id`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: DELETE
  /orgs/{organizationId}/sites/{siteId}/permissions/teams/{teamId} (Remove an org team from a site)
  executes a live GitBook API operation.
- `stream_ai_response_in_site`: POST `/orgs/{{ record.organization_id }}/sites/{{ record.site_id
  }}/ai/response` - kind `custom`; body type `json`; path fields `organization_id`, `site_id`;
  required record fields `organization_id`, `site_id`, `input`; accepted fields `channel`, `input`,
  `model`, `organization_id`, `previous_response_id`, `session`, `site_id`, `tool_call`, `tools`;
  risk: POST /orgs/{organizationId}/sites/{siteId}/ai/response (Generate an AI response in a site)
  executes a live GitBook API operation.
- `update_site_agent_settings_by_id`: PUT `/orgs/{{ record.organization_id }}/sites/{{
  record.site_id }}/agent-settings` - kind `update`; body type `json`; path fields
  `organization_id`, `site_id`; required record fields `organization_id`, `site_id`, `scans`,
  `findings`, `editing`; accepted fields `editing`, `findings`, `organization_id`, `scans`,
  `site_id`; risk: PUT /orgs/{organizationId}/sites/{siteId}/agent-settings (Update the AI agent
  configuration for a site) executes a live GitBook API operation.
- `create_site_styleguide_by_id`: POST `/orgs/{{ record.organization_id }}/sites/{{ record.site_id
  }}/styleguide` - kind `create`; body type `json`; path fields `organization_id`, `site_id`;
  required record fields `organization_id`, `site_id`; accepted fields `organization_id`, `site_id`,
  `template`; risk: POST /orgs/{organizationId}/sites/{siteId}/styleguide (Create or retrieve the
  styleguide space for a site) executes a live GitBook API operation.
- `track_events_in_site_by_id`: POST `/orgs/{{ record.organization_id }}/sites/{{ record.site_id
  }}/insights/events` - kind `custom`; body type `json`; path fields `organization_id`, `site_id`;
  required record fields `organization_id`, `site_id`, `events`; accepted fields `events`,
  `organization_id`, `site_id`; risk: POST /orgs/{organizationId}/sites/{siteId}/insights/events
  (Track site events) executes a live GitBook API operation.
- `aggregate_site_events`: POST `/orgs/{{ record.organization_id }}/sites/{{ record.site_id
  }}/insights/events/aggregate` - kind `custom`; body type `json`; path fields `organization_id`,
  `site_id`; required record fields `organization_id`, `site_id`, `range`; accepted fields
  `group_by`, `limit`, `order`, `organization_id`, `range`, `select`, `site_id`, `where`; risk: POST
  /orgs/{organizationId}/sites/{siteId}/insights/events/aggregate (Query site events) executes a
  live GitBook API operation.
- `update_site_ads_by_id`: POST `/orgs/{{ record.organization_id }}/sites/{{ record.site_id }}/ads`
  - kind `update`; body type `json`; path fields `organization_id`, `site_id`; required record
  fields `organization_id`, `site_id`; accepted fields `organization_id`, `site_id`, `status`,
  `topic`; risk: POST /orgs/{organizationId}/sites/{siteId}/ads (Update the advertising settings for
  a site) executes a live GitBook API operation.
- `create_site_redirect`: POST `/orgs/{{ record.organization_id }}/sites/{{ record.site_id
  }}/redirects` - kind `create`; body type `json`; path fields `organization_id`, `site_id`;
  required record fields `organization_id`, `site_id`, `source`, `destination`; accepted fields
  `capture_wildcard`, `destination`, `draft`, `organization_id`, `site_id`, `source`; risk: POST
  /orgs/{organizationId}/sites/{siteId}/redirects (Create a URL redirect rule for a site) executes a
  live GitBook API operation.
- `bulk_upsert_site_redirects`: PUT `/orgs/{{ record.organization_id }}/sites/{{ record.site_id
  }}/redirects` - kind `upsert`; body type `json`; path fields `organization_id`, `site_id`;
  required record fields `organization_id`, `site_id`, `redirects`; accepted fields
  `organization_id`, `redirects`, `site_id`; risk: PUT
  /orgs/{organizationId}/sites/{siteId}/redirects (Create, update, delete, or publish site redirect
  rules in bulk) executes a live GitBook API operation.
- `update_site_redirect_by_id`: PATCH `/orgs/{{ record.organization_id }}/sites/{{ record.site_id
  }}/redirects/{{ record.site_redirect_id }}` - kind `update`; body type `json`; path fields
  `organization_id`, `site_id`, `site_redirect_id`; required record fields `organization_id`,
  `site_id`, `site_redirect_id`; accepted fields `capture_wildcard`, `destination`, `draft`,
  `organization_id`, `site_id`, `site_redirect_id`, `source`; risk: PATCH
  /orgs/{organizationId}/sites/{siteId}/redirects/{siteRedirectId} (Update a URL redirect rule for a
  site) executes a live GitBook API operation.
- `delete_site_redirect_by_id`: DELETE `/orgs/{{ record.organization_id }}/sites/{{ record.site_id
  }}/redirects/{{ record.site_redirect_id }}` - kind `delete`; body type `none`; path fields
  `organization_id`, `site_id`, `site_redirect_id`; required record fields `organization_id`,
  `site_id`, `site_redirect_id`; accepted fields `organization_id`, `site_id`, `site_redirect_id`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: DELETE
  /orgs/{organizationId}/sites/{siteId}/redirects/{siteRedirectId} (Delete a site redirect) executes
  a live GitBook API operation.
- `create_site_mcp_server`: POST `/orgs/{{ record.organization_id }}/sites/{{ record.site_id
  }}/mcp-servers` - kind `create`; body type `json`; path fields `organization_id`, `site_id`;
  required record fields `organization_id`, `site_id`, `name`, `url`, `headers`; accepted fields
  `condition`, `headers`, `name`, `organization_id`, `site_id`, `transport`, `url`; risk: POST
  /orgs/{organizationId}/sites/{siteId}/mcp-servers (Add a new MCP server configuration to a site)
  executes a live GitBook API operation.
- `update_site_mcp_server_by_id`: PATCH `/orgs/{{ record.organization_id }}/sites/{{ record.site_id
  }}/mcp-servers/{{ record.site_mcp_server_id }}` - kind `update`; body type `json`; path fields
  `organization_id`, `site_id`, `site_mcp_server_id`; required record fields `organization_id`,
  `site_id`, `site_mcp_server_id`; accepted fields `condition`, `headers`, `name`,
  `organization_id`, `site_id`, `site_mcp_server_id`, `transport`, `url`; risk: PATCH
  /orgs/{organizationId}/sites/{siteId}/mcp-servers/{siteMcpServerId} (Update an MCP server
  configuration for a site) executes a live GitBook API operation.
- `delete_site_mcp_server_by_id`: DELETE `/orgs/{{ record.organization_id }}/sites/{{ record.site_id
  }}/mcp-servers/{{ record.site_mcp_server_id }}` - kind `delete`; body type `none`; path fields
  `organization_id`, `site_id`, `site_mcp_server_id`; required record fields `organization_id`,
  `site_id`, `site_mcp_server_id`; accepted fields `organization_id`, `site_id`,
  `site_mcp_server_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: DELETE /orgs/{organizationId}/sites/{siteId}/mcp-servers/{siteMcpServerId}
  (Delete a site MCP server) executes a live GitBook API operation.
- `create_site_channel`: POST `/orgs/{{ record.organization_id }}/sites/{{ record.site_id
  }}/channels` - kind `create`; body type `json`; path fields `organization_id`, `site_id`; required
  record fields `organization_id`, `site_id`; accepted fields `organization_id`, `role`,
  `setup_settings`, `site_id`, `type`; risk: POST /orgs/{organizationId}/sites/{siteId}/channels
  (Create a new GitBook Agent channel for a site) executes a live GitBook API operation.
- `update_site_channel_by_id`: PATCH `/orgs/{{ record.organization_id }}/sites/{{ record.site_id
  }}/channels/{{ record.site_channel_id }}` - kind `update`; body type `json`; path fields
  `organization_id`, `site_id`, `site_channel_id`; required record fields `organization_id`,
  `site_id`, `site_channel_id`; accepted fields `organization_id`, `role`, `setup_settings`,
  `site_channel_id`, `site_id`, `type`; risk: PATCH
  /orgs/{organizationId}/sites/{siteId}/channels/{siteChannelId} (Update a GitBook Agent channel for
  a site) executes a live GitBook API operation.
- `delete_site_channel_by_id`: DELETE `/orgs/{{ record.organization_id }}/sites/{{ record.site_id
  }}/channels/{{ record.site_channel_id }}` - kind `delete`; body type `none`; path fields
  `organization_id`, `site_id`, `site_channel_id`; required record fields `organization_id`,
  `site_id`, `site_channel_id`; accepted fields `organization_id`, `site_channel_id`, `site_id`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: DELETE
  /orgs/{organizationId}/sites/{siteId}/channels/{siteChannelId} (Delete a GitBook Agent channel
  from a site) executes a live GitBook API operation.
- `dns_revalidate_custom_hostname`: PATCH `/custom-hostnames/{{ record.hostname }}` - kind `update`;
  body type `none`; path fields `hostname`; required record fields `hostname`; accepted fields
  `hostname`; risk: PATCH /custom-hostnames/{hostname} (Revalidate a custom hostname DNS) executes a
  live GitBook API operation.
- `remove_custom_hostname`: DELETE `/custom-hostnames/{{ record.hostname }}` - kind `delete`; body
  type `none`; path fields `hostname`; required record fields `hostname`; accepted fields
  `hostname`; missing records treated as success for status `404`; confirmation `destructive`; risk:
  DELETE /custom-hostnames/{hostname} (Remove a custom hostname) executes a live GitBook API
  operation.
- `ads_update_site`: PATCH `/ads/sites/{{ record.site_id }}` - kind `update`; body type `json`; path
  fields `site_id`; required record fields `site_id`; accepted fields `reason`, `reporting_id`,
  `site_id`, `status`, `zone_id`; risk: PATCH /ads/sites/{siteId} (Update the Ads configuration for
  a site) executes a live GitBook API operation.
- `resolve_published_content_by_url`: POST `/urls/published` - kind `custom`; body type `json`;
  required record fields `url`; accepted fields `redirect_on_error`, `url`, `visitor`; risk: POST
  /urls/published (Resolve a URL of a published content.) executes a live GitBook API operation.
- `install_git_sync_provider_on_target`: POST `/git/installations` - kind `create`; body type
  `json`; required record fields `provider`, `target`; accepted fields `provider`, `target`; risk:
  POST /git/installations (Install a Git Sync provider on a target) executes a live GitBook API
  operation.
- `update_git_sync_installation_by_id`: PATCH `/git/installations/{{ record.installation_id }}` -
  kind `update`; body type `json`; path fields `installation_id`; required record fields
  `installation_id`; accepted fields `commit_message_template`, `installation_id`,
  `preview_external_branches`, `priority`, `project_directory`, `provider`,
  `provider_configuration`; risk: PATCH /git/installations/{installationId} (Update a Git Sync
  installation configuration) executes a live GitBook API operation.
- `uninstall_git_sync_installation`: DELETE `/git/installations/{{ record.installation_id }}` - kind
  `delete`; body type `none`; path fields `installation_id`; required record fields
  `installation_id`; accepted fields `installation_id`; missing records treated as success for
  status `404`; confirmation `destructive`; risk: DELETE /git/installations/{installationId}
  (Uninstall a Git Sync installation) executes a live GitBook API operation.

## Known limits

- Batch defaults: read_page_size=50, write_batch_size=1.
- API coverage includes 185 stream-backed endpoint group(s), 169 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=3, out_of_scope=1.
