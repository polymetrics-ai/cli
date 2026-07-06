# Overview

Reads and writes the documented Pylon REST API surface through concrete streams and write actions.

Readable streams: `issues`, `accounts`, `contacts`, `users`, `messages`, `account_relationships`,
`account`, `activity_types`, `audit_logs`, `call_recording`, `contact`, `custom_fields`,
`custom_field`, `custom_objects`, `custom_object`, `feature_request`, `issue_statuses`, `issue`,
`issue_followers`, `issue_messages`, `issue_threads`, `issue_voice_calls`, `knowledge_bases`,
`knowledge_base`, `articles`, `article`, `collections`, `collection`, `macro_groups`, `macros`,
`macro`, `me`, `milestone`, `project`, `surveys`, `survey`, `survey_responses`, `tags`, `tag`,
`tasks`, `task`, `task_comments`, `teams`, `team`, `ticket_forms`, `ticket_form`, `training_data`,
`training_data_detail`, `user_roles`, `user`.

Write actions: `update_accounts`, `create_account`, `merge_accounts`, `search_accounts`,
`create_account_highlight`, `delete_account_highlight`, `update_account_highlight`,
`create_account_relationship`, `delete_account_relationship`, `delete_account`, `update_account`,
`create_activity`, `search_audit_logs`, `search_call_recordings`, `delete_call_recording`,
`update_call_recording`, `create_contact`, `search_contacts`, `delete_contact`, `update_contact`,
`create_custom_field`, `update_custom_field`, `update_custom_objects`, `create_custom_object`,
`search_custom_objects`, `delete_custom_object`, `update_custom_object`, `create_feature_request`,
`search_feature_requests`, `delete_feature_request`, `update_feature_request`,
`set_feature_request_portal_visibility`, `import_contact`, `import_issue`, `import_messages`,
`create_issue`, `search_issues`, `delete_issue`, `update_issue`, `create_issue_ai_response`,
`link_external_issue`, `add_issue_followers`, `delete_message`, `redact_message`,
`create_issue_note`, `create_issue_reply`, `snooze_issue`, `create_issue_thread`, `create_article`,
`delete_article`, `update_article`, `request_article_review`, `create_collection`,
`delete_collection`, `update_collection`, `create_route_redirect`, `create_macro`, `update_macro`,
`create_milestone`, `delete_milestone`, `update_milestone`, `create_project`, `search_projects`,
`delete_project`, `update_project`, `search_surveys`, `create_tag`, `delete_tag`, `update_tag`,
`create_task`, `search_tasks`, `delete_task`, `update_task`, `create_task_comment`,
`delete_task_comment`, `update_task_comment`, `create_team`, `update_team`, `create_training_data`,
`upload_training_data_file_content`, and 3 more.

Service API documentation: https://docs.usepylon.com/pylon-docs/developer/api/api-reference.

## Auth setup

Connection fields:

- `account_id` (optional, string); Path parameter used by Pylon account_relationships streams.
- `api_token` (required, secret, string); Pylon API token, sent as an OAuth-style Bearer token.
  Never logged.
- `article_id` (optional, string); Path parameter used by Pylon article streams.
- `base_url` (optional, string); default `https://api.usepylon.com`; format `uri`; Pylon API base
  URL. Defaults to https://api.usepylon.com.
- `call_recording_id` (optional, string); Path parameter used by Pylon call_recording streams.
- `collection_id` (optional, string); Path parameter used by Pylon collection streams.
- `contact_id` (optional, string); Path parameter used by Pylon contact streams.
- `custom_field_id` (optional, string); Path parameter used by Pylon custom_field streams.
- `custom_fields_object_type` (optional, string); Required object_type query parameter for
  custom_fields.
- `custom_object_id` (optional, string); Path parameter used by Pylon custom_object streams.
- `custom_object_type` (optional, string); Path parameter used by Pylon custom_objects streams.
- `feature_request_id` (optional, string); Path parameter used by Pylon feature_request streams.
- `issue_id` (optional, string); Path parameter used by Pylon issue streams.
- `knowledge_base_id` (optional, string); Path parameter used by Pylon knowledge_base streams.
- `macro_id` (optional, string); Path parameter used by Pylon macro streams.
- `milestone_id` (optional, string); Path parameter used by Pylon milestone streams.
- `mode` (optional, string).
- `project_id` (optional, string); Path parameter used by Pylon project streams.
- `start_date` (optional, string); RFC3339 lower bound sent as the updated_after query parameter on
  every request, when set.
- `survey_id` (optional, string); Path parameter used by Pylon survey streams.
- `tag_id` (optional, string); Path parameter used by Pylon tag streams.
- `task_id` (optional, string); Path parameter used by Pylon task streams.
- `team_id` (optional, string); Path parameter used by Pylon team streams.
- `ticket_form_id` (optional, string); Path parameter used by Pylon ticket_form streams.
- `training_data_id` (optional, string); Path parameter used by Pylon training_data_detail streams.
- `user_id` (optional, string); Path parameter used by Pylon user streams.

Secret fields are redacted in logs and write previews: `api_token`.

Default configuration values: `base_url=https://api.usepylon.com`.

Authentication behavior:

- Bearer token authentication using `secrets.api_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/issues` with query `limit`=`1`.

## Streams notes

Default pagination: cursor pagination; cursor parameter `cursor`; next token from
`pagination.next_cursor`.

Pagination by stream: cursor: `issues`, `accounts`, `contacts`, `users`, `messages`, `audit_logs`,
`contact`, `custom_objects`, `issue_messages`, `articles`, `survey_responses`, `tasks`; none:
`account_relationships`, `account`, `activity_types`, `call_recording`, `custom_fields`,
`custom_field`, `custom_object`, `feature_request`, `issue_statuses`, `issue`, `issue_followers`,
`issue_threads`, `issue_voice_calls`, `knowledge_bases`, `knowledge_base`, `article`, `collections`,
`collection`, `macro_groups`, `macros`, `macro`, `me`, `milestone`, `project`, `surveys`, `survey`,
`tags`, `tag`, `task`, `task_comments`, `teams`, `team`, `ticket_forms`, `ticket_form`,
`training_data`, `training_data_detail`, `user_roles`, `user`.

- `issues`: GET `/issues` - records path `data`; query `limit`=`100`; `updated_after` from template
  `{{ config.start_date }}`, omitted when absent; cursor pagination; cursor parameter `cursor`; next
  token from `pagination.next_cursor`; computed output fields `name`, `state`.
- `accounts`: GET `/accounts` - records path `data`; query `limit`=`100`; `updated_after` from
  template `{{ config.start_date }}`, omitted when absent; cursor pagination; cursor parameter
  `cursor`; next token from `pagination.next_cursor`; computed output fields `name`, `state`.
- `contacts`: GET `/contacts` - records path `data`; query `limit`=`100`; `updated_after` from
  template `{{ config.start_date }}`, omitted when absent; cursor pagination; cursor parameter
  `cursor`; next token from `pagination.next_cursor`; computed output fields `name`, `state`.
- `users`: GET `/users` - records path `data`; query `limit`=`100`; `updated_after` from template
  `{{ config.start_date }}`, omitted when absent; cursor pagination; cursor parameter `cursor`; next
  token from `pagination.next_cursor`; computed output fields `name`, `state`.
- `messages`: GET `/messages` - records path `data`; query `limit`=`100`; `updated_after` from
  template `{{ config.start_date }}`, omitted when absent; cursor pagination; cursor parameter
  `cursor`; next token from `pagination.next_cursor`; computed output fields `name`, `state`.
- `account_relationships`: GET `/accounts/{{ config.account_id }}/relationships` - records path
  `data`; emits passthrough records.
- `account`: GET `/accounts/{{ config.account_id }}` - records path `data`; emits passthrough
  records.
- `activity_types`: GET `/activity-types` - records path `data`; emits passthrough records.
- `audit_logs`: GET `/audit-logs` - records path `data`; query `limit`=`100`; cursor pagination;
  cursor parameter `cursor`; next token from `pagination.cursor`; stop flag
  `pagination.has_next_page`; emits passthrough records.
- `call_recording`: GET `/call-recordings/{{ config.call_recording_id }}` - records path `data`;
  emits passthrough records.
- `contact`: GET `/contacts/{{ config.contact_id }}` - records path `data`; query `limit`=`100`;
  cursor pagination; cursor parameter `cursor`; next token from `pagination.cursor`; stop flag
  `pagination.has_next_page`; emits passthrough records.
- `custom_fields`: GET `/custom-fields` - records path `data`; query `object_type`=`{{
  config.custom_fields_object_type }}`; emits passthrough records.
- `custom_field`: GET `/custom-fields/{{ config.custom_field_id }}` - records path `data`; emits
  passthrough records.
- `custom_objects`: GET `/custom-objects/{{ config.custom_object_type }}` - records path `data`;
  query `limit`=`100`; cursor pagination; cursor parameter `cursor`; next token from
  `pagination.cursor`; stop flag `pagination.has_next_page`; emits passthrough records.
- `custom_object`: GET `/custom-objects/{{ config.custom_object_type }}/{{ config.custom_object_id
  }}` - records path `data`; emits passthrough records.
- `feature_request`: GET `/feature-requests/{{ config.feature_request_id }}` - records path `data`;
  emits passthrough records.
- `issue_statuses`: GET `/issue-statuses` - records path `data`; emits passthrough records.
- `issue`: GET `/issues/{{ config.issue_id }}` - records path `data`; emits passthrough records.
- `issue_followers`: GET `/issues/{{ config.issue_id }}/followers` - records path `data`; emits
  passthrough records.
- `issue_messages`: GET `/issues/{{ config.issue_id }}/messages` - records path `data`; query
  `limit`=`100`; cursor pagination; cursor parameter `cursor`; next token from `pagination.cursor`;
  stop flag `pagination.has_next_page`; emits passthrough records.
- `issue_threads`: GET `/issues/{{ config.issue_id }}/threads` - records path `data`; emits
  passthrough records.
- `issue_voice_calls`: GET `/issues/{{ config.issue_id }}/voice-calls` - records path `data`; emits
  passthrough records.
- `knowledge_bases`: GET `/knowledge-bases` - records path `data`; emits passthrough records.
- `knowledge_base`: GET `/knowledge-bases/{{ config.knowledge_base_id }}` - records path `data`;
  emits passthrough records.
- `articles`: GET `/knowledge-bases/{{ config.knowledge_base_id }}/articles` - records path `data`;
  query `limit`=`100`; cursor pagination; cursor parameter `cursor`; next token from
  `pagination.cursor`; stop flag `pagination.has_next_page`; emits passthrough records.
- `article`: GET `/knowledge-bases/{{ config.knowledge_base_id }}/articles/{{ config.article_id }}`
  - records path `data`; emits passthrough records.
- `collections`: GET `/knowledge-bases/{{ config.knowledge_base_id }}/collections` - records path
  `data`; emits passthrough records.
- `collection`: GET `/knowledge-bases/{{ config.knowledge_base_id }}/collections/{{
  config.collection_id }}` - records path `data`; emits passthrough records.
- `macro_groups`: GET `/macro-groups` - records path `data`; emits passthrough records.
- `macros`: GET `/macros` - records path `data`; emits passthrough records.
- `macro`: GET `/macros/{{ config.macro_id }}` - records path `data`; emits passthrough records.
- `me`: GET `/me` - records path `data`; emits passthrough records.
- `milestone`: GET `/milestones/{{ config.milestone_id }}` - records path `data`; emits passthrough
  records.
- `project`: GET `/projects/{{ config.project_id }}` - records path `data`; emits passthrough
  records.
- `surveys`: GET `/surveys` - records path `data`; emits passthrough records.
- `survey`: GET `/surveys/{{ config.survey_id }}` - records path `data`; emits passthrough records.
- `survey_responses`: GET `/surveys/{{ config.survey_id }}/responses` - records path `data`; query
  `limit`=`100`; cursor pagination; cursor parameter `cursor`; next token from `pagination.cursor`;
  stop flag `pagination.has_next_page`; emits passthrough records.
- `tags`: GET `/tags` - records path `data`; emits passthrough records.
- `tag`: GET `/tags/{{ config.tag_id }}` - records path `data`; emits passthrough records.
- `tasks`: GET `/tasks` - records path `data`; query `limit`=`100`; cursor pagination; cursor
  parameter `cursor`; next token from `pagination.cursor`; stop flag `pagination.has_next_page`;
  emits passthrough records.
- `task`: GET `/tasks/{{ config.task_id }}` - records path `data`; emits passthrough records.
- `task_comments`: GET `/tasks/{{ config.task_id }}/comments` - records path `data`; emits
  passthrough records.
- `teams`: GET `/teams` - records path `data`; emits passthrough records.
- `team`: GET `/teams/{{ config.team_id }}` - records path `data`; emits passthrough records.
- `ticket_forms`: GET `/ticket-forms` - records path `data`; emits passthrough records.
- `ticket_form`: GET `/ticket-forms/{{ config.ticket_form_id }}` - records path `data`; emits
  passthrough records.
- `training_data`: GET `/training-data` - records path `data`; emits passthrough records.
- `training_data_detail`: GET `/training-data/{{ config.training_data_id }}` - records path `data`;
  emits passthrough records.
- `user_roles`: GET `/user-roles` - records path `data`; emits passthrough records.
- `user`: GET `/users/{{ config.user_id }}` - records path `data`; emits passthrough records.

## Write actions & risks

Overall write risk: external Pylon API mutations for support, account, knowledge-base, project,
task, tag, survey, and training-data records; approval required.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `update_accounts`: PATCH `/accounts` - kind `update`; body type `json`; required record fields
  `account_ids`; accepted fields `account_ids`, `account_type`, `custom_fields`, `owner_id`, `tags`,
  `tags_apply_mode`; risk: external Pylon PATCH /accounts; approval required.
- `create_account`: POST `/accounts` - kind `create`; body type `json`; required record fields
  `name`; accepted fields `account_type`, `channels`, `custom_fields`, `domain`, `domains`,
  `external_ids`, `logo_url`, `name`, `owner_id`, `primary_domain`, `subaccount_ids`, `tags`; risk:
  external Pylon POST /accounts; approval required.
- `merge_accounts`: POST `/accounts/merge` - kind `update`; body type `json`; required record fields
  `merge_into_account_id`, `merge_account_ids`; accepted fields `merge_account_ids`,
  `merge_into_account_id`; confirmation `destructive`; risk: destructive external Pylon POST
  /accounts/merge; approval required.
- `search_accounts`: POST `/accounts/search` - kind `custom`; body type `json`; required record
  fields `filter`; accepted fields `cursor`, `filter`, `limit`, `search_text`; risk: external Pylon
  POST /accounts/search; approval required.
- `create_account_highlight`: POST `/accounts/{{ record.account_id }}/highlights` - kind `create`;
  body type `json`; path fields `account_id`; required record fields `content_html`, `account_id`;
  accepted fields `account_id`, `content_html`, `expires_at`; risk: external Pylon POST
  /accounts/{account_id}/highlights; approval required.
- `delete_account_highlight`: DELETE `/accounts/{{ record.account_id }}/highlights/{{
  record.highlight_id }}` - kind `delete`; body type `none`; path fields `account_id`,
  `highlight_id`; required record fields `account_id`, `highlight_id`; accepted fields `account_id`,
  `highlight_id`; missing records treated as success for status `404`; confirmation `destructive`;
  risk: destructive external Pylon DELETE /accounts/{account_id}/highlights/{highlight_id}; approval
  required.
- `update_account_highlight`: PATCH `/accounts/{{ record.account_id }}/highlights/{{
  record.highlight_id }}` - kind `update`; body type `json`; path fields `account_id`,
  `highlight_id`; required record fields `account_id`, `highlight_id`; accepted fields `account_id`,
  `content_html`, `expires_at`, `highlight_id`; risk: external Pylon PATCH
  /accounts/{account_id}/highlights/{highlight_id}; approval required.
- `create_account_relationship`: POST `/accounts/{{ record.account_id }}/relationship` - kind
  `create`; body type `json`; path fields `account_id`; required record fields `related_object_id`,
  `relationship_type`, `account_id`; accepted fields `account_id`, `related_object_id`,
  `relationship_type`; risk: external Pylon POST /accounts/{account_id}/relationship; approval
  required.
- `delete_account_relationship`: DELETE `/accounts/{{ record.account_id }}/relationships/{{
  record.relationship_id }}` - kind `delete`; body type `none`; path fields `account_id`,
  `relationship_id`; required record fields `account_id`, `relationship_id`; accepted fields
  `account_id`, `relationship_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: destructive external Pylon DELETE
  /accounts/{account_id}/relationships/{relationship_id}; approval required.
- `delete_account`: DELETE `/accounts/{{ record.account_id }}` - kind `delete`; body type `none`;
  path fields `account_id`; required record fields `account_id`; accepted fields `account_id`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: destructive
  external Pylon DELETE /accounts/{id}; approval required.
- `update_account`: PATCH `/accounts/{{ record.account_id }}` - kind `update`; body type `json`;
  path fields `account_id`; required record fields `account_id`; accepted fields `account_id`,
  `account_type`, `channels`, `custom_fields`, `domains`, `external_ids`, `is_disabled`, `logo_url`,
  `name`, `owner_id`, `primary_domain`, `subaccount_ids`, `tags`; risk: external Pylon PATCH
  /accounts/{id}; approval required.
- `create_activity`: POST `/accounts/{{ record.account_id }}/activities` - kind `create`; body type
  `json`; path fields `account_id`; required record fields `slug`, `account_id`; accepted fields
  `account_id`, `body_html`, `contact_id`, `happened_at`, `link`, `link_text`, `slug`, `user_id`;
  risk: external Pylon POST /accounts/{id}/activities; approval required.
- `search_audit_logs`: POST `/audit-logs/search` - kind `custom`; body type `json`; accepted fields
  `cursor`, `filter`, `limit`; risk: external Pylon POST /audit-logs/search; approval required.
- `search_call_recordings`: POST `/call-recordings/search` - kind `custom`; body type `json`;
  accepted fields `cursor`, `filter`, `limit`; risk: external Pylon POST /call-recordings/search;
  approval required.
- `delete_call_recording`: DELETE `/call-recordings/{{ record.call_recording_id }}` - kind `delete`;
  body type `none`; path fields `call_recording_id`; required record fields `call_recording_id`;
  accepted fields `call_recording_id`; missing records treated as success for status `404`;
  confirmation `destructive`; risk: destructive external Pylon DELETE /call-recordings/{id};
  approval required.
- `update_call_recording`: PATCH `/call-recordings/{{ record.call_recording_id }}` - kind `update`;
  body type `json`; path fields `call_recording_id`; required record fields `call_recording_id`;
  accepted fields `account_id`, `call_recording_id`, `custom_fields`; risk: external Pylon PATCH
  /call-recordings/{id}; approval required.
- `create_contact`: POST `/contacts` - kind `create`; body type `json`; required record fields
  `name`; accepted fields `account_external_id`, `account_id`, `avatar_url`, `custom_fields`,
  `email`, `external_ids`, `name`, `phone_numbers`, `portal_role`, `portal_role_id`,
  `primary_phone_number`; risk: external Pylon POST /contacts; approval required.
- `search_contacts`: POST `/contacts/search` - kind `custom`; body type `json`; required record
  fields `filter`; accepted fields `cursor`, `filter`, `limit`, `search_text`; risk: external Pylon
  POST /contacts/search; approval required.
- `delete_contact`: DELETE `/contacts/{{ record.contact_id }}` - kind `delete`; body type `none`;
  path fields `contact_id`; required record fields `contact_id`; accepted fields `contact_id`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: destructive
  external Pylon DELETE /contacts/{id}; approval required.
- `update_contact`: PATCH `/contacts/{{ record.contact_id }}` - kind `update`; body type `json`;
  path fields `contact_id`; required record fields `contact_id`; accepted fields
  `account_external_id`, `account_id`, `avatar_url`, `contact_id`, `custom_fields`, `email`,
  `emails`, `external_ids`, `name`, `phone_numbers`, `portal_role`, `portal_role_id`,
  `primary_phone_number`; risk: external Pylon PATCH /contacts/{id}; approval required.
- `create_custom_field`: POST `/custom-fields` - kind `create`; body type `json`; required record
  fields `object_type`, `label`, `type`; accepted fields `default_value`, `default_values`,
  `description`, `label`, `object_type`, `select_options`, `slug`, `type`; risk: external Pylon POST
  /custom-fields; approval required.
- `update_custom_field`: PATCH `/custom-fields/{{ record.custom_field_id }}` - kind `update`; body
  type `json`; path fields `custom_field_id`; required record fields `custom_field_id`; accepted
  fields `custom_field_id`, `default_value`, `default_values`, `description`, `label`,
  `select_options`, `slug`; risk: external Pylon PATCH /custom-fields/{id}; approval required.
- `update_custom_objects`: PATCH `/custom-objects/{{ record.custom_object_type }}` - kind `update`;
  body type `json`; path fields `custom_object_type`; required record fields `ids`,
  `custom_object_type`; accepted fields `custom_fields`, `custom_object_type`, `ids`; risk: external
  Pylon PATCH /custom-objects/{type}; approval required.
- `create_custom_object`: POST `/custom-objects/{{ record.custom_object_type }}` - kind `create`;
  body type `json`; path fields `custom_object_type`; required record fields `name`,
  `custom_object_type`; accepted fields `custom_fields`, `custom_object_type`, `name`; risk:
  external Pylon POST /custom-objects/{type}; approval required.
- `search_custom_objects`: POST `/custom-objects/{{ record.custom_object_type }}/search` - kind
  `custom`; body type `json`; path fields `custom_object_type`; required record fields `filter`,
  `custom_object_type`; accepted fields `cursor`, `custom_object_type`, `filter`, `limit`; risk:
  external Pylon POST /custom-objects/{type}/search; approval required.
- `delete_custom_object`: DELETE `/custom-objects/{{ record.custom_object_type }}/{{
  record.custom_object_id }}` - kind `delete`; body type `none`; path fields `custom_object_type`,
  `custom_object_id`; required record fields `custom_object_type`, `custom_object_id`; accepted
  fields `custom_object_id`, `custom_object_type`; missing records treated as success for status
  `404`; confirmation `destructive`; risk: destructive external Pylon DELETE
  /custom-objects/{type}/{id}; approval required.
- `update_custom_object`: PATCH `/custom-objects/{{ record.custom_object_type }}/{{
  record.custom_object_id }}` - kind `update`; body type `json`; path fields `custom_object_type`,
  `custom_object_id`; required record fields `custom_object_type`, `custom_object_id`; accepted
  fields `custom_fields`, `custom_object_id`, `custom_object_type`, `name`; risk: external Pylon
  PATCH /custom-objects/{type}/{id}; approval required.
- `create_feature_request`: POST `/feature-requests` - kind `update`; body type `json`; required
  record fields `title`; accepted fields `description`, `should_auto_fetch_evidence`, `title`; risk:
  external Pylon POST /feature-requests; approval required.
- `search_feature_requests`: POST `/feature-requests/search` - kind `update`; body type `json`;
  accepted fields `account_ids`, `limit`, `query`, `request_statuses`; risk: external Pylon POST
  /feature-requests/search; approval required.
- `delete_feature_request`: DELETE `/feature-requests/{{ record.feature_request_id }}` - kind
  `delete`; body type `none`; path fields `feature_request_id`; required record fields
  `feature_request_id`; accepted fields `feature_request_id`; missing records treated as success for
  status `404`; confirmation `destructive`; risk: destructive external Pylon DELETE
  /feature-requests/{id}; approval required.
- `update_feature_request`: PATCH `/feature-requests/{{ record.feature_request_id }}` - kind
  `update`; body type `json`; path fields `feature_request_id`; required record fields
  `feature_request_id`; accepted fields `custom_fields`, `feature_request_id`, `request_status`;
  risk: external Pylon PATCH /feature-requests/{id}; approval required.
- `set_feature_request_portal_visibility`: POST `/feature-requests/{{ record.feature_request_id
  }}/set-portal-visibility` - kind `update`; body type `json`; path fields `feature_request_id`;
  required record fields `account_ids`, `visible`, `feature_request_id`; accepted fields
  `account_ids`, `feature_request_id`, `visible`; risk: external Pylon POST
  /feature-requests/{id}/set-portal-visibility; approval required.
- `import_contact`: POST `/import/contacts` - kind `create`; body type `json`; required record
  fields `name`, `email`; accepted fields `account_id`, `avatar_url`, `custom_fields`, `email`,
  `name`, `portal_role`; risk: external Pylon POST /import/contacts; approval required.
- `import_issue`: POST `/import/issues` - kind `create`; body type `json`; required record fields
  `title`, `state`, `messages`; accepted fields `account_id`, `assignee_id`, `attachment_urls`,
  `created_at`, `custom_fields`, `external_issues`, `external_refs`, `first_response_time`,
  `messages`, `requester_id`, `resolution_time`, `state`, `tags`, `team_id`, `title`, `updated_at`;
  risk: external Pylon POST /import/issues; approval required.
- `import_messages`: POST `/import/issues/{{ record.issue_id }}/messages` - kind `create`; body type
  `json`; path fields `issue_id`; required record fields `messages`, `issue_id`; accepted fields
  `issue_id`, `messages`; risk: external Pylon POST /import/issues/{id}/messages; approval required.
- `create_issue`: POST `/issues` - kind `create`; body type `json`; required record fields `title`,
  `body_html`; accepted fields `account_id`, `assignee_id`, `attachment_urls`, `author_unverified`,
  `body_html`, `contact_id`, `created_at`, `custom_fields`, `destination_metadata`, `priority`,
  `requester_avatar_url`, `requester_email`, `requester_id`, `requester_name`, `tags`, `team_id`,
  `title`, `user_id`; risk: external Pylon POST /issues; approval required.
- `search_issues`: POST `/issues/search` - kind `custom`; body type `json`; required record fields
  `filter`; accepted fields `cursor`, `filter`, `limit`, `search_text`; risk: external Pylon POST
  /issues/search; approval required.
- `delete_issue`: DELETE `/issues/{{ record.issue_id }}` - kind `delete`; body type `none`; path
  fields `issue_id`; required record fields `issue_id`; accepted fields `issue_id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: destructive external Pylon
  DELETE /issues/{id}; approval required.
- `update_issue`: PATCH `/issues/{{ record.issue_id }}` - kind `update`; body type `json`; path
  fields `issue_id`; required record fields `issue_id`; accepted fields `account_id`, `assignee_id`,
  `custom_fields`, `customer_portal_visible`, `issue_id`, `requester_id`, `requestor_id`, `state`,
  `tags`, `team_id`, `title`, `type`; risk: external Pylon PATCH /issues/{id}; approval required.
- `create_issue_ai_response`: POST `/issues/{{ record.issue_id }}/ai-response` - kind `create`; body
  type `json`; path fields `issue_id`; required record fields `ai_agent_id`, `issue_id`; accepted
  fields `ai_agent_id`, `issue_id`, `post_as_internal_note`; risk: external Pylon POST
  /issues/{id}/ai-response; approval required.
- `link_external_issue`: POST `/issues/{{ record.issue_id }}/external-issues` - kind `update`; body
  type `json`; path fields `issue_id`; required record fields `source`, `external_issue_id`,
  `issue_id`; accepted fields `external_issue_id`, `issue_id`, `operation`, `source`; risk: external
  Pylon POST /issues/{id}/external-issues; approval required.
- `add_issue_followers`: POST `/issues/{{ record.issue_id }}/followers` - kind `update`; body type
  `json`; path fields `issue_id`; required record fields `issue_id`; accepted fields `contact_ids`,
  `issue_id`, `operation`, `user_ids`; risk: external Pylon POST /issues/{id}/followers; approval
  required.
- `delete_message`: DELETE `/issues/{{ record.issue_id }}/messages/{{ record.message_id }}` - kind
  `delete`; body type `none`; path fields `issue_id`, `message_id`; required record fields
  `issue_id`, `message_id`; accepted fields `issue_id`, `message_id`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: destructive external Pylon DELETE
  /issues/{id}/messages/{message_id}; approval required.
- `redact_message`: POST `/issues/{{ record.issue_id }}/messages/{{ record.message_id }}/redact` -
  kind `update`; body type `none`; path fields `issue_id`, `message_id`; required record fields
  `issue_id`, `message_id`; accepted fields `issue_id`, `message_id`; confirmation `destructive`;
  risk: destructive external Pylon POST /issues/{id}/messages/{message_id}/redact; approval
  required.
- `create_issue_note`: POST `/issues/{{ record.issue_id }}/note` - kind `create`; body type `json`;
  path fields `issue_id`; required record fields `body_html`, `issue_id`; accepted fields
  `attachment_urls`, `body_html`, `issue_id`, `message_id`, `thread_id`, `thread_name`, `user_id`;
  risk: external Pylon POST /issues/{id}/note; approval required.
- `create_issue_reply`: POST `/issues/{{ record.issue_id }}/reply` - kind `create`; body type
  `json`; path fields `issue_id`; required record fields `body_html`, `message_id`, `issue_id`;
  accepted fields `attachment_urls`, `body_html`, `contact_id`, `custom_source`, `email_info`,
  `issue_id`, `message_id`, `user_id`; risk: external Pylon POST /issues/{id}/reply; approval
  required.
- `snooze_issue`: POST `/issues/{{ record.issue_id }}/snooze` - kind `update`; body type `json`;
  path fields `issue_id`; required record fields `snooze_until`, `issue_id`; accepted fields
  `issue_id`, `snooze_until`; risk: external Pylon POST /issues/{id}/snooze; approval required.
- `create_issue_thread`: POST `/issues/{{ record.issue_id }}/threads` - kind `create`; body type
  `json`; path fields `issue_id`; required record fields `issue_id`; accepted fields `issue_id`,
  `name`; risk: external Pylon POST /issues/{id}/threads; approval required.
- `create_article`: POST `/knowledge-bases/{{ record.knowledge_base_id }}/articles` - kind `create`;
  body type `json`; path fields `knowledge_base_id`; required record fields `title`,
  `author_user_id`, `body_html`, `knowledge_base_id`; accepted fields `author_user_id`, `body_html`,
  `collection_id`, `is_published`, `is_unlisted`, `knowledge_base_id`, `slug`, `title`,
  `translations`, `visibility_config`; risk: external Pylon POST /knowledge-bases/{id}/articles;
  approval required.
- `delete_article`: DELETE `/knowledge-bases/{{ record.knowledge_base_id }}/articles/{{
  record.article_id }}` - kind `delete`; body type `none`; path fields `knowledge_base_id`,
  `article_id`; required record fields `knowledge_base_id`, `article_id`; accepted fields
  `article_id`, `knowledge_base_id`; missing records treated as success for status `404`;
  confirmation `destructive`; risk: destructive external Pylon DELETE
  /knowledge-bases/{id}/articles/{article_id}; approval required.
- `update_article`: PATCH `/knowledge-bases/{{ record.knowledge_base_id }}/articles/{{
  record.article_id }}` - kind `update`; body type `json`; path fields `knowledge_base_id`,
  `article_id`; required record fields `knowledge_base_id`, `article_id`; accepted fields
  `article_id`, `body_html`, `is_published`, `is_unlisted`, `knowledge_base_id`, `language`,
  `publish_updated_body_html`, `title`, `visibility_config`; risk: external Pylon PATCH
  /knowledge-bases/{id}/articles/{article_id}; approval required.
- `request_article_review`: POST `/knowledge-bases/{{ record.knowledge_base_id }}/articles/{{
  record.article_id }}/request-review` - kind `update`; body type `json`; path fields
  `knowledge_base_id`, `article_id`; required record fields `knowledge_base_id`, `article_id`;
  accepted fields `article_id`, `knowledge_base_id`, `request_ai_review`, `reviewer_user_ids`; risk:
  external Pylon POST /knowledge-bases/{id}/articles/{article_id}/request-review; approval required.
- `create_collection`: POST `/knowledge-bases/{{ record.knowledge_base_id }}/collections` - kind
  `create`; body type `json`; path fields `knowledge_base_id`; required record fields `title`,
  `knowledge_base_id`; accepted fields `description`, `knowledge_base_id`, `parent_collection_id`,
  `slug`, `title`; risk: external Pylon POST /knowledge-bases/{id}/collections; approval required.
- `delete_collection`: DELETE `/knowledge-bases/{{ record.knowledge_base_id }}/collections/{{
  record.collection_id }}` - kind `delete`; body type `none`; path fields `knowledge_base_id`,
  `collection_id`; required record fields `knowledge_base_id`, `collection_id`; accepted fields
  `collection_id`, `knowledge_base_id`; missing records treated as success for status `404`;
  confirmation `destructive`; risk: destructive external Pylon DELETE
  /knowledge-bases/{id}/collections/{collection_id}; approval required.
- `update_collection`: PATCH `/knowledge-bases/{{ record.knowledge_base_id }}/collections/{{
  record.collection_id }}` - kind `update`; body type `json`; path fields `knowledge_base_id`,
  `collection_id`; required record fields `knowledge_base_id`, `collection_id`; accepted fields
  `collection_id`, `description`, `knowledge_base_id`, `slug`, `title`, `visibility_config`; risk:
  external Pylon PATCH /knowledge-bases/{id}/collections/{collection_id}; approval required.
- `create_route_redirect`: POST `/knowledge-bases/{{ record.knowledge_base_id }}/route-redirects` -
  kind `create`; body type `json`; path fields `knowledge_base_id`; required record fields
  `from_path`, `object_id`, `object_type`, `knowledge_base_id`; accepted fields `from_path`,
  `knowledge_base_id`, `language`, `object_id`, `object_type`; risk: external Pylon POST
  /knowledge-bases/{id}/route-redirects; approval required.
- `create_macro`: POST `/macros` - kind `create`; body type `json`; required record fields `name`,
  `text_html`, `macro_group_id`; accepted fields `conditions`, `macro_group_id`, `name`,
  `text_html`, `text_type`, `visibility`; risk: external Pylon POST /macros; approval required.
- `update_macro`: PATCH `/macros/{{ record.macro_id }}` - kind `update`; body type `json`; path
  fields `macro_id`; required record fields `macro_id`; accepted fields `conditions`,
  `macro_group_id`, `macro_id`, `name`, `text_html`, `text_type`, `visibility`; risk: external Pylon
  PATCH /macros/{id}; approval required.
- `create_milestone`: POST `/milestones` - kind `create`; body type `json`; required record fields
  `name`, `project_id`; accepted fields `account_id`, `due_date`, `name`, `project_id`; risk:
  external Pylon POST /milestones; approval required.
- `delete_milestone`: DELETE `/milestones/{{ record.milestone_id }}` - kind `delete`; body type
  `none`; path fields `milestone_id`; required record fields `milestone_id`; accepted fields
  `milestone_id`; missing records treated as success for status `404`; confirmation `destructive`;
  risk: destructive external Pylon DELETE /milestones/{id}; approval required.
- `update_milestone`: PATCH `/milestones/{{ record.milestone_id }}` - kind `update`; body type
  `json`; path fields `milestone_id`; required record fields `milestone_id`; accepted fields
  `due_date`, `milestone_id`, `name`; risk: external Pylon PATCH /milestones/{id}; approval
  required.
- `create_project`: POST `/projects` - kind `create`; body type `json`; required record fields
  `name`, `account_id`; accepted fields `account_id`, `customer_portal_visible`, `description_html`,
  `end_date`, `name`, `owner_id`, `project_template_id`, `start_date`; risk: external Pylon POST
  /projects; approval required.
- `search_projects`: POST `/projects/search` - kind `custom`; body type `json`; accepted fields
  `cursor`, `filter`, `limit`; risk: external Pylon POST /projects/search; approval required.
- `delete_project`: DELETE `/projects/{{ record.project_id }}` - kind `delete`; body type `none`;
  path fields `project_id`; required record fields `project_id`; accepted fields `project_id`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: destructive
  external Pylon DELETE /projects/{id}; approval required.
- `update_project`: PATCH `/projects/{{ record.project_id }}` - kind `update`; body type `json`;
  path fields `project_id`; required record fields `project_id`; accepted fields `custom_fields`,
  `customer_portal_visible`, `description_html`, `end_date`, `is_archived`, `name`, `owner_id`,
  `project_id`, `start_date`; risk: external Pylon PATCH /projects/{id}; approval required.
- `search_surveys`: POST `/surveys/search` - kind `custom`; body type `json`; accepted fields
  `filter`, `limit`; risk: external Pylon POST /surveys/search; approval required.
- `create_tag`: POST `/tags` - kind `create`; body type `json`; required record fields
  `object_type`, `value`; accepted fields `hex_color`, `object_type`, `value`; risk: external Pylon
  POST /tags; approval required.
- `delete_tag`: DELETE `/tags/{{ record.tag_id }}` - kind `delete`; body type `none`; path fields
  `tag_id`; required record fields `tag_id`; accepted fields `tag_id`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: destructive external Pylon DELETE
  /tags/{id}; approval required.
- `update_tag`: PATCH `/tags/{{ record.tag_id }}` - kind `update`; body type `json`; path fields
  `tag_id`; required record fields `tag_id`; accepted fields `hex_color`, `tag_id`, `value`; risk:
  external Pylon PATCH /tags/{id}; approval required.
- `create_task`: POST `/tasks` - kind `create`; body type `json`; required record fields `title`;
  accepted fields `account_id`, `assignee_id`, `body_html`, `custom_fields`,
  `customer_portal_visible`, `due_date`, `milestone_id`, `parent_task_id`, `project_id`, `status`,
  `title`; risk: external Pylon POST /tasks; approval required.
- `search_tasks`: POST `/tasks/search` - kind `custom`; body type `json`; accepted fields `cursor`,
  `filter`, `limit`; risk: external Pylon POST /tasks/search; approval required.
- `delete_task`: DELETE `/tasks/{{ record.task_id }}` - kind `delete`; body type `none`; path fields
  `task_id`; required record fields `task_id`; accepted fields `task_id`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: destructive external Pylon DELETE
  /tasks/{id}; approval required.
- `update_task`: PATCH `/tasks/{{ record.task_id }}` - kind `update`; body type `json`; path fields
  `task_id`; required record fields `task_id`; accepted fields `assignee_id`, `body_html`,
  `custom_fields`, `customer_portal_visible`, `due_date`, `milestone_id`, `project_id`, `status`,
  `task_id`, `title`; risk: external Pylon PATCH /tasks/{id}; approval required.
- `create_task_comment`: POST `/tasks/{{ record.task_id }}/comments` - kind `create`; body type
  `json`; path fields `task_id`; required record fields `body_html`, `task_id`; accepted fields
  `body_html`, `is_internal`, `task_id`; risk: external Pylon POST /tasks/{id}/comments; approval
  required.
- `delete_task_comment`: DELETE `/tasks/{{ record.task_id }}/comments/{{ record.comment_id }}` -
  kind `delete`; body type `none`; path fields `task_id`, `comment_id`; required record fields
  `task_id`, `comment_id`; accepted fields `comment_id`, `task_id`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: destructive external Pylon DELETE
  /tasks/{id}/comments/{comment_id}; approval required.
- `update_task_comment`: PATCH `/tasks/{{ record.task_id }}/comments/{{ record.comment_id }}` - kind
  `update`; body type `json`; path fields `task_id`, `comment_id`; required record fields
  `body_html`, `task_id`, `comment_id`; accepted fields `body_html`, `comment_id`, `task_id`; risk:
  external Pylon PATCH /tasks/{id}/comments/{comment_id}; approval required.
- `create_team`: POST `/teams` - kind `create`; body type `json`; accepted fields `name`,
  `user_ids`; risk: external Pylon POST /teams; approval required.
- `update_team`: PATCH `/teams/{{ record.team_id }}` - kind `update`; body type `json`; path fields
  `team_id`; required record fields `team_id`; accepted fields `name`, `team_id`, `user_ids`; risk:
  external Pylon PATCH /teams/{id}; approval required.
- `create_training_data`: POST `/training-data` - kind `create`; body type `json`; accepted fields
  `training_data_name`, `visibility`; risk: external Pylon POST /training-data; approval required.
- `upload_training_data_file_content`: POST `/training-data/upload-content` - kind `create`; body
  type `json`; accepted fields `content`, `external_id`, `file_name`, `training_data_id`,
  `training_data_name`, `visibility`; risk: external Pylon POST /training-data/upload-content;
  approval required.
- `delete_training_data_documents`: DELETE `/training-data/{{ record.training_data_id }}/documents`
  - kind `delete`; body type `none`; path fields `training_data_id`; required record fields
  `training_data_id`; accepted fields `training_data_id`; missing records treated as success for
  status `404`; confirmation `destructive`; risk: destructive external Pylon DELETE
  /training-data/{id}/documents; approval required.
- `search_users`: POST `/users/search` - kind `custom`; body type `json`; required record fields
  `filter`; accepted fields `cursor`, `filter`, `limit`; risk: external Pylon POST /users/search;
  approval required.
- `update_user`: PATCH `/users/{{ record.user_id }}` - kind `update`; body type `json`; path fields
  `user_id`; required record fields `user_id`; accepted fields `avatar_url`, `name`, `role_id`,
  `status`, `user_id`; risk: external Pylon PATCH /users/{id}; approval required.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 50 stream-backed endpoint group(s), 83 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=3.
