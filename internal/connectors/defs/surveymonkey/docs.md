# Overview

Reads and writes SurveyMonkey REST v3 and SCIM v2 resources through the documented API surface.

Readable streams: `surveys`, `survey_search`, `survey`, `survey_details`, `survey_share_list`,
`survey_categories`, `survey_templates`, `team_survey_templates`, `survey_languages_catalog`,
`collectors`, `survey_pages`, `survey_page`, `survey_questions`, `survey_question`,
`question_benchmark`, `question_rollups`, `question_trends`, `page_rollups`, `page_trends`,
`survey_response_summaries`, `survey_response`, `survey_response_details`, `survey_responses`,
`survey_rollups`, `survey_trends`, `survey_languages`, `survey_language`, `question_bank_questions`,
`survey_folders`, `users_me`, `user_workgroups`, `user_shared`, `groups`, `group`,
`group_activities`, `group_activities_by_action`, `group_members`, `group_member`, `workgroups`,
`workgroup`, `workgroup_shares`, `workgroup_share`, `workgroup_members`, `workgroup_member`,
`contact_lists`, `contact_list`, `contact_list_contacts`, `contact_list_contacts_bulk`, `contacts`,
`contact`, `contacts_bulk`, `contact_fields`, `contact_field`, `collector`, `collector_messages`,
`collector_message`, `collector_message_recipients`, `collector_recipients`,
`survey_collector_recipient`, `survey_collector_message_stats`, `collector_responses`,
`collector_response`, `collector_response_details`, `collector_bulk_responses`, `collector_stats`,
`webhooks`, `webhook`, `benchmark_bundles`, `benchmark_bundle`, `benchmark_bundle_analysis`,
`organizations`, `organization`, `roles`, `errors`, `error`, `scim_users`, `scim_user`,
`scim_schemas`, `scim_schema`, `scim_resource_types`, `scim_resource_type`,
`scim_service_provider_config`.

Write actions: `update_collector`, `replace_collector`, `create_collector_message`,
`update_collector_message`, `replace_collector_message`, `create_message_recipient`,
`create_message_recipients_bulk`, `send_collector_message`, `create_collector_response`,
`update_collector_response`, `replace_collector_response`, `create_contact_list`,
`update_contact_list`, `replace_contact_list`, `copy_contact_list`, `merge_contact_list`,
`add_contact_to_contact_list`, `add_contacts_to_contact_list_bulk`, `create_contact`,
`create_contacts_bulk`, `update_contact`, `replace_contact`, `update_contact_field`,
`create_organization`, `update_organization`, `create_survey_folder`, `create_survey`,
`update_survey`, `replace_survey`, `share_survey`, `create_survey_page`, `update_survey_page`,
`replace_survey_page`, `create_survey_question`, `create_survey_language`, `update_survey_language`,
`create_survey_collector`, `update_survey_response`, `replace_survey_response`, `create_webhook`,
`update_webhook`, `replace_webhook`, `update_group_member`, `create_workgroup`, `update_workgroup`,
`create_workgroup_share`, `create_workgroup_shares_bulk`, `create_workgroup_member`,
`create_workgroup_members_bulk`, `update_workgroup_member`, `create_scim_user`, `replace_scim_user`,
`update_scim_user`.

Service API documentation: https://developer.surveymonkey.com/api/v3/.

## Auth setup

Connection fields:

- `access_token` (required, secret, string); SurveyMonkey OAuth access token, sent as a Bearer
  token. Never logged.
- `action_type` (optional, string); SurveyMonkey group activity action type for action-scoped
  activity streams.
- `base_url` (optional, string); default `https://api.surveymonkey.com`; format `uri`; SurveyMonkey
  API host root. REST streams use /v3 paths and SCIM streams use /scim/v2 paths.
- `benchmark_bundle_id` (optional, string); SurveyMonkey benchmark bundle ID for benchmark detail
  and analysis streams.
- `collector_id` (optional, string); SurveyMonkey collector ID for collector-scoped streams.
- `contact_field_id` (optional, string); SurveyMonkey contact field ID for contact-field detail
  streams.
- `contact_id` (optional, string); SurveyMonkey contact ID for contact detail streams.
- `contact_list_id` (optional, string); SurveyMonkey contact list ID for contact-list scoped
  streams.
- `error_id` (optional, string); SurveyMonkey error ID for error detail streams.
- `group_id` (optional, string); SurveyMonkey group ID for group-scoped streams.
- `language_code` (optional, string); SurveyMonkey survey language code for translation streams.
- `member_id` (optional, string); SurveyMonkey group/workgroup member ID for member detail streams.
- `message_id` (optional, string); SurveyMonkey collector message ID for message-scoped streams.
- `organization_id` (optional, string); SurveyMonkey organization ID for organization detail
  streams.
- `page_id` (optional, string); SurveyMonkey survey page ID for page/question streams.
- `question_id` (optional, string); SurveyMonkey survey question ID for question detail, benchmark,
  rollup, and trend streams.
- `recipient_id` (optional, string); SurveyMonkey collector recipient ID for recipient detail
  streams.
- `response_id` (optional, string); SurveyMonkey response ID for response detail streams.
- `scim_resource_type_id` (optional, string); SurveyMonkey SCIM ResourceType ID for SCIM
  resource-type detail streams.
- `scim_schema_id` (optional, string); SurveyMonkey SCIM Schema ID for SCIM schema detail streams.
- `scim_user_id` (optional, string); SurveyMonkey SCIM User ID for SCIM user detail streams.
- `share_id` (optional, string); SurveyMonkey workgroup share ID for share detail streams.
- `survey_id` (optional, string); SurveyMonkey survey ID for survey-scoped streams.
- `user_id` (optional, string); SurveyMonkey user ID for user-scoped streams.
- `webhook_id` (optional, string); SurveyMonkey webhook ID for webhook detail streams.
- `workgroup_id` (optional, string); SurveyMonkey workgroup ID for workgroup-scoped streams.

Secret fields are redacted in logs and write previews: `access_token`.

Default configuration values: `base_url=https://api.surveymonkey.com`.

Authentication behavior:

- Bearer token authentication using `secrets.access_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/v3/users/me`.

## Streams notes

Default pagination: single request; no pagination.

Pagination by stream: next_url: `surveys`, `survey_search`, `survey_share_list`,
`survey_categories`, `survey_templates`, `team_survey_templates`, `survey_languages_catalog`,
`collectors`, `survey_pages`, `survey_questions`, `survey_response_summaries`, `survey_responses`,
`survey_languages`, `question_bank_questions`, `survey_folders`, `user_workgroups`, `user_shared`,
`groups`, `group_activities`, `group_activities_by_action`, `group_members`, `workgroups`,
`workgroup_shares`, `workgroup_members`, `contact_lists`, `contact_list_contacts`,
`contact_list_contacts_bulk`, `contacts`, `contacts_bulk`, `contact_fields`, `collector_messages`,
`collector_message_recipients`, `collector_recipients`, `collector_responses`,
`collector_bulk_responses`, `webhooks`, `benchmark_bundles`, `organizations`, `roles`, `errors`;
none: `survey`, `survey_details`, `survey_page`, `survey_question`, `question_benchmark`,
`question_rollups`, `question_trends`, `page_rollups`, `page_trends`, `survey_response`,
`survey_response_details`, `survey_rollups`, `survey_trends`, `survey_language`, `users_me`,
`group`, `group_member`, `workgroup`, `workgroup_share`, `workgroup_member`, `contact_list`,
`contact`, `contact_field`, `collector`, `collector_message`, `survey_collector_recipient`,
`survey_collector_message_stats`, `collector_response`, `collector_response_details`,
`collector_stats`, `webhook`, `benchmark_bundle`, `benchmark_bundle_analysis`, `organization`,
`error`, `scim_users`, `scim_user`, `scim_schemas`, `scim_schema`, `scim_resource_types`,
`scim_resource_type`, `scim_service_provider_config`.

- `surveys`: GET `/v3/surveys` - records path `data`; query `per_page`=`100`; follows a next-page
  URL from the response body; URL path `links.next`; next URLs stay on the configured API host.
- `survey_search`: GET `/v3/surveys/search` - records path `data`; query `per_page`=`100`; follows a
  next-page URL from the response body; URL path `links.next`; next URLs stay on the configured API
  host; emits passthrough records.
- `survey`: GET `/v3/surveys/{{ config.survey_id }}` - single-object response; records path `.`;
  computed output fields `id`; emits passthrough records.
- `survey_details`: GET `/v3/surveys/{{ config.survey_id }}/details` - single-object response;
  records path `.`; computed output fields `id`; emits passthrough records.
- `survey_share_list`: GET `/v3/surveys/{{ config.survey_id }}/share_list` - records path `data`;
  query `per_page`=`100`; follows a next-page URL from the response body; URL path `links.next`;
  next URLs stay on the configured API host; emits passthrough records.
- `survey_categories`: GET `/v3/survey_categories` - records path `data`; query `per_page`=`100`;
  follows a next-page URL from the response body; URL path `links.next`; next URLs stay on the
  configured API host; emits passthrough records.
- `survey_templates`: GET `/v3/survey_templates` - records path `data`; query `per_page`=`100`;
  follows a next-page URL from the response body; URL path `links.next`; next URLs stay on the
  configured API host; emits passthrough records.
- `team_survey_templates`: GET `/v3/team_survey_templates` - records path `data`; query
  `per_page`=`100`; follows a next-page URL from the response body; URL path `links.next`; next URLs
  stay on the configured API host; emits passthrough records.
- `survey_languages_catalog`: GET `/v3/survey_languages` - records path `data`; query
  `per_page`=`100`; follows a next-page URL from the response body; URL path `links.next`; next URLs
  stay on the configured API host; emits passthrough records.
- `collectors`: GET `/v3/surveys/{{ config.survey_id }}/collectors` - records path `data`; query
  `per_page`=`100`; follows a next-page URL from the response body; URL path `links.next`; next URLs
  stay on the configured API host.
- `survey_pages`: GET `/v3/surveys/{{ config.survey_id }}/pages` - records path `data`; query
  `per_page`=`100`; follows a next-page URL from the response body; URL path `links.next`; next URLs
  stay on the configured API host; emits passthrough records.
- `survey_page`: GET `/v3/surveys/{{ config.survey_id }}/pages/{{ config.page_id }}` - single-object
  response; records path `.`; computed output fields `id`, `survey_id`; emits passthrough records.
- `survey_questions`: GET `/v3/surveys/{{ config.survey_id }}/pages/{{ config.page_id }}/questions`
  - records path `data`; query `per_page`=`100`; follows a next-page URL from the response body; URL
  path `links.next`; next URLs stay on the configured API host; emits passthrough records.
- `survey_question`: GET `/v3/surveys/{{ config.survey_id }}/pages/{{ config.page_id }}/questions/{{
  config.question_id }}` - single-object response; records path `.`; computed output fields `id`,
  `page_id`, `survey_id`; emits passthrough records.
- `question_benchmark`: GET `/v3/surveys/{{ config.survey_id }}/pages/{{ config.page_id
  }}/questions/{{ config.question_id }}/benchmark` - single-object response; records path `.`;
  computed output fields `id`, `page_id`, `survey_id`; emits passthrough records.
- `question_rollups`: GET `/v3/surveys/{{ config.survey_id }}/pages/{{ config.page_id
  }}/questions/{{ config.question_id }}/rollups` - single-object response; records path `.`;
  computed output fields `id`, `page_id`, `survey_id`; emits passthrough records.
- `question_trends`: GET `/v3/surveys/{{ config.survey_id }}/pages/{{ config.page_id }}/questions/{{
  config.question_id }}/trends` - single-object response; records path `.`; computed output fields
  `id`, `page_id`, `survey_id`; emits passthrough records.
- `page_rollups`: GET `/v3/surveys/{{ config.survey_id }}/pages/{{ config.page_id }}/rollups` -
  single-object response; records path `.`; computed output fields `id`, `survey_id`; emits
  passthrough records.
- `page_trends`: GET `/v3/surveys/{{ config.survey_id }}/pages/{{ config.page_id }}/trends` -
  single-object response; records path `.`; computed output fields `id`, `survey_id`; emits
  passthrough records.
- `survey_response_summaries`: GET `/v3/surveys/{{ config.survey_id }}/responses` - records path
  `data`; query `per_page`=`100`; follows a next-page URL from the response body; URL path
  `links.next`; next URLs stay on the configured API host; emits passthrough records.
- `survey_response`: GET `/v3/surveys/{{ config.survey_id }}/responses/{{ config.response_id }}` -
  single-object response; records path `.`; computed output fields `id`, `survey_id`; emits
  passthrough records.
- `survey_response_details`: GET `/v3/surveys/{{ config.survey_id }}/responses/{{ config.response_id
  }}/details` - single-object response; records path `.`; computed output fields `id`, `survey_id`;
  emits passthrough records.
- `survey_responses`: GET `/v3/surveys/{{ config.survey_id }}/responses/bulk` - records path `data`;
  query `per_page`=`100`; follows a next-page URL from the response body; URL path `links.next`;
  next URLs stay on the configured API host.
- `survey_rollups`: GET `/v3/surveys/{{ config.survey_id }}/rollups` - single-object response;
  records path `.`; computed output fields `id`; emits passthrough records.
- `survey_trends`: GET `/v3/surveys/{{ config.survey_id }}/trends` - single-object response; records
  path `.`; computed output fields `id`; emits passthrough records.
- `survey_languages`: GET `/v3/surveys/{{ config.survey_id }}/languages` - records path `data`;
  query `per_page`=`100`; follows a next-page URL from the response body; URL path `links.next`;
  next URLs stay on the configured API host; emits passthrough records.
- `survey_language`: GET `/v3/surveys/{{ config.survey_id }}/languages/{{ config.language_code }}` -
  single-object response; records path `.`; computed output fields `id`, `survey_id`; emits
  passthrough records.
- `question_bank_questions`: GET `/v3/question_bank/questions` - records path `data`; query
  `per_page`=`100`; follows a next-page URL from the response body; URL path `links.next`; next URLs
  stay on the configured API host; emits passthrough records.
- `survey_folders`: GET `/v3/survey_folders` - records path `data`; query `per_page`=`100`; follows
  a next-page URL from the response body; URL path `links.next`; next URLs stay on the configured
  API host; emits passthrough records.
- `users_me`: GET `/v3/users/me` - single-object response; records path `.`; emits passthrough
  records.
- `user_workgroups`: GET `/v3/users/{{ config.user_id }}/workgroups` - records path `data`; query
  `per_page`=`100`; follows a next-page URL from the response body; URL path `links.next`; next URLs
  stay on the configured API host; emits passthrough records.
- `user_shared`: GET `/v3/users/{{ config.user_id }}/shared` - records path `data`; query
  `per_page`=`100`; follows a next-page URL from the response body; URL path `links.next`; next URLs
  stay on the configured API host; emits passthrough records.
- `groups`: GET `/v3/groups` - records path `data`; query `per_page`=`100`; follows a next-page URL
  from the response body; URL path `links.next`; next URLs stay on the configured API host; emits
  passthrough records.
- `group`: GET `/v3/groups/{{ config.group_id }}` - single-object response; records path `.`;
  computed output fields `id`; emits passthrough records.
- `group_activities`: GET `/v3/groups/{{ config.group_id }}/activities` - records path `data`; query
  `per_page`=`100`; follows a next-page URL from the response body; URL path `links.next`; next URLs
  stay on the configured API host; emits passthrough records.
- `group_activities_by_action`: GET `/v3/groups/{{ config.group_id }}/activities/{{
  config.action_type }}` - records path `data`; query `per_page`=`100`; follows a next-page URL from
  the response body; URL path `links.next`; next URLs stay on the configured API host; emits
  passthrough records.
- `group_members`: GET `/v3/groups/{{ config.group_id }}/members` - records path `data`; query
  `per_page`=`100`; follows a next-page URL from the response body; URL path `links.next`; next URLs
  stay on the configured API host; emits passthrough records.
- `group_member`: GET `/v3/groups/{{ config.group_id }}/members/{{ config.member_id }}` -
  single-object response; records path `.`; computed output fields `group_id`, `id`; emits
  passthrough records.
- `workgroups`: GET `/v3/workgroups` - records path `data`; query `per_page`=`100`; follows a
  next-page URL from the response body; URL path `links.next`; next URLs stay on the configured API
  host; emits passthrough records.
- `workgroup`: GET `/v3/workgroups/{{ config.workgroup_id }}` - single-object response; records path
  `.`; computed output fields `id`; emits passthrough records.
- `workgroup_shares`: GET `/v3/workgroups/{{ config.workgroup_id }}/shares` - records path `data`;
  query `per_page`=`100`; follows a next-page URL from the response body; URL path `links.next`;
  next URLs stay on the configured API host; emits passthrough records.
- `workgroup_share`: GET `/v3/workgroups/{{ config.workgroup_id }}/shares/{{ config.share_id }}` -
  single-object response; records path `.`; computed output fields `id`, `workgroup_id`; emits
  passthrough records.
- `workgroup_members`: GET `/v3/workgroups/{{ config.workgroup_id }}/members` - records path `data`;
  query `per_page`=`100`; follows a next-page URL from the response body; URL path `links.next`;
  next URLs stay on the configured API host; emits passthrough records.
- `workgroup_member`: GET `/v3/workgroups/{{ config.workgroup_id }}/members/{{ config.member_id }}`
  - single-object response; records path `.`; computed output fields `id`, `workgroup_id`; emits
  passthrough records.
- `contact_lists`: GET `/v3/contact_lists` - records path `data`; query `per_page`=`100`; follows a
  next-page URL from the response body; URL path `links.next`; next URLs stay on the configured API
  host; emits passthrough records.
- `contact_list`: GET `/v3/contact_lists/{{ config.contact_list_id }}` - single-object response;
  records path `.`; computed output fields `id`; emits passthrough records.
- `contact_list_contacts`: GET `/v3/contact_lists/{{ config.contact_list_id }}/contacts` - records
  path `data`; query `per_page`=`100`; follows a next-page URL from the response body; URL path
  `links.next`; next URLs stay on the configured API host; emits passthrough records.
- `contact_list_contacts_bulk`: GET `/v3/contact_lists/{{ config.contact_list_id }}/contacts/bulk` -
  records path `data`; query `per_page`=`100`; follows a next-page URL from the response body; URL
  path `links.next`; next URLs stay on the configured API host; emits passthrough records.
- `contacts`: GET `/v3/contacts` - records path `data`; query `per_page`=`100`; follows a next-page
  URL from the response body; URL path `links.next`; next URLs stay on the configured API host;
  emits passthrough records.
- `contact`: GET `/v3/contacts/{{ config.contact_id }}` - single-object response; records path `.`;
  computed output fields `id`; emits passthrough records.
- `contacts_bulk`: GET `/v3/contacts/bulk` - records path `data`; query `per_page`=`100`; follows a
  next-page URL from the response body; URL path `links.next`; next URLs stay on the configured API
  host; emits passthrough records.
- `contact_fields`: GET `/v3/contact_fields` - records path `data`; query `per_page`=`100`; follows
  a next-page URL from the response body; URL path `links.next`; next URLs stay on the configured
  API host; emits passthrough records.
- `contact_field`: GET `/v3/contact_fields/{{ config.contact_field_id }}` - single-object response;
  records path `.`; computed output fields `id`; emits passthrough records.
- `collector`: GET `/v3/collectors/{{ config.collector_id }}` - single-object response; records path
  `.`; computed output fields `id`; emits passthrough records.
- `collector_messages`: GET `/v3/collectors/{{ config.collector_id }}/messages` - records path
  `data`; query `per_page`=`100`; follows a next-page URL from the response body; URL path
  `links.next`; next URLs stay on the configured API host; emits passthrough records.
- `collector_message`: GET `/v3/collectors/{{ config.collector_id }}/messages/{{ config.message_id
  }}` - single-object response; records path `.`; computed output fields `collector_id`, `id`; emits
  passthrough records.
- `collector_message_recipients`: GET `/v3/collectors/{{ config.collector_id }}/messages/{{
  config.message_id }}/recipients` - records path `data`; query `per_page`=`100`; follows a
  next-page URL from the response body; URL path `links.next`; next URLs stay on the configured API
  host; emits passthrough records.
- `collector_recipients`: GET `/v3/collectors/{{ config.collector_id }}/recipients` - records path
  `data`; query `per_page`=`100`; follows a next-page URL from the response body; URL path
  `links.next`; next URLs stay on the configured API host; emits passthrough records.
- `survey_collector_recipient`: GET `/v3/surveys/{{ config.survey_id }}/collectors/{{
  config.collector_id }}/recipients/{{ config.recipient_id }}` - single-object response; records
  path `.`; computed output fields `collector_id`, `id`, `survey_id`; emits passthrough records.
- `survey_collector_message_stats`: GET `/v3/surveys/{{ config.survey_id }}/collectors/{{
  config.collector_id }}/messages/{{ config.message_id }}/stats` - single-object response; records
  path `.`; computed output fields `collector_id`, `id`, `survey_id`; emits passthrough records.
- `collector_responses`: GET `/v3/collectors/{{ config.collector_id }}/responses` - records path
  `data`; query `per_page`=`100`; follows a next-page URL from the response body; URL path
  `links.next`; next URLs stay on the configured API host; emits passthrough records.
- `collector_response`: GET `/v3/collectors/{{ config.collector_id }}/responses/{{
  config.response_id }}` - single-object response; records path `.`; computed output fields
  `collector_id`, `id`; emits passthrough records.
- `collector_response_details`: GET `/v3/collectors/{{ config.collector_id }}/responses/{{
  config.response_id }}/details` - single-object response; records path `.`; computed output fields
  `collector_id`, `id`; emits passthrough records.
- `collector_bulk_responses`: GET `/v3/collectors/{{ config.collector_id }}/responses/bulk` -
  records path `data`; query `per_page`=`100`; follows a next-page URL from the response body; URL
  path `links.next`; next URLs stay on the configured API host; emits passthrough records.
- `collector_stats`: GET `/v3/collectors/{{ config.collector_id }}/stats` - single-object response;
  records path `.`; computed output fields `id`; emits passthrough records.
- `webhooks`: GET `/v3/webhooks` - records path `data`; query `per_page`=`100`; follows a next-page
  URL from the response body; URL path `links.next`; next URLs stay on the configured API host;
  emits passthrough records.
- `webhook`: GET `/v3/webhooks/{{ config.webhook_id }}` - single-object response; records path `.`;
  computed output fields `id`; emits passthrough records.
- `benchmark_bundles`: GET `/v3/benchmark_bundles` - records path `data`; query `per_page`=`100`;
  follows a next-page URL from the response body; URL path `links.next`; next URLs stay on the
  configured API host; emits passthrough records.
- `benchmark_bundle`: GET `/v3/benchmark_bundles/{{ config.benchmark_bundle_id }}` - single-object
  response; records path `.`; computed output fields `id`; emits passthrough records.
- `benchmark_bundle_analysis`: GET `/v3/benchmark_bundles/{{ config.benchmark_bundle_id }}/analyze`
  - single-object response; records path `.`; computed output fields `id`; emits passthrough
  records.
- `organizations`: GET `/v3/organizations` - records path `data`; query `per_page`=`100`; follows a
  next-page URL from the response body; URL path `links.next`; next URLs stay on the configured API
  host; emits passthrough records.
- `organization`: GET `/v3/organization/{{ config.organization_id }}` - single-object response;
  records path `.`; computed output fields `id`; emits passthrough records.
- `roles`: GET `/v3/roles` - records path `data`; query `per_page`=`100`; follows a next-page URL
  from the response body; URL path `links.next`; next URLs stay on the configured API host; emits
  passthrough records.
- `errors`: GET `/v3/errors` - records path `data`; query `per_page`=`100`; follows a next-page URL
  from the response body; URL path `links.next`; next URLs stay on the configured API host; emits
  passthrough records.
- `error`: GET `/v3/errors/{{ config.error_id }}` - single-object response; records path `.`;
  computed output fields `id`; emits passthrough records.
- `scim_users`: GET `/scim/v2/Users` - records path `Resources`; query `count`=`1000`; emits
  passthrough records.
- `scim_user`: GET `/scim/v2/Users/{{ config.scim_user_id }}` - single-object response; records path
  `.`; computed output fields `id`; emits passthrough records.
- `scim_schemas`: GET `/scim/v2/Schemas` - records path `Resources`; query `count`=`1000`; emits
  passthrough records.
- `scim_schema`: GET `/scim/v2/Schemas/{{ config.scim_schema_id }}` - single-object response;
  records path `.`; computed output fields `id`; emits passthrough records.
- `scim_resource_types`: GET `/scim/v2/ResourceTypes` - records path `Resources`; query
  `count`=`1000`; emits passthrough records.
- `scim_resource_type`: GET `/scim/v2/ResourceTypes/{{ config.scim_resource_type_id }}` -
  single-object response; records path `.`; computed output fields `id`; emits passthrough records.
- `scim_service_provider_config`: GET `/scim/v2/ServiceProviderConfig` - single-object response;
  records path `.`; computed output fields `id`; emits passthrough records.

## Write actions & risks

Overall write risk: external SurveyMonkey API mutations can create or update surveys, collectors,
messages, contacts, contact lists, workgroups, organizations, webhooks, and SCIM users; message
send/share actions can notify real recipients.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `update_collector`: PATCH `/v3/collectors/{{ record.collector_id }}` - kind `update`; body type
  `json`; path fields `collector_id`; required record fields `collector_id`; accepted fields
  `collector_id`; risk: updates SurveyMonkey collector data in the connected account; approval
  required.
- `replace_collector`: PUT `/v3/collectors/{{ record.collector_id }}` - kind `update`; body type
  `json`; path fields `collector_id`; required record fields `collector_id`; accepted fields
  `collector_id`; risk: replaces SurveyMonkey collector data in the connected account; approval
  required.
- `create_collector_message`: POST `/v3/collectors/{{ record.collector_id }}/messages` - kind
  `create`; body type `json`; path fields `collector_id`; required record fields `collector_id`;
  accepted fields `collector_id`; risk: creates SurveyMonkey collector message data in the connected
  account; approval required.
- `update_collector_message`: PATCH `/v3/collectors/{{ record.collector_id }}/messages/{{
  record.message_id }}` - kind `update`; body type `json`; path fields `collector_id`, `message_id`;
  required record fields `collector_id`, `message_id`; accepted fields `collector_id`, `message_id`;
  risk: updates SurveyMonkey collector message data in the connected account; approval required.
- `replace_collector_message`: PUT `/v3/collectors/{{ record.collector_id }}/messages/{{
  record.message_id }}` - kind `update`; body type `json`; path fields `collector_id`, `message_id`;
  required record fields `collector_id`, `message_id`; accepted fields `collector_id`, `message_id`;
  risk: replaces SurveyMonkey collector message data in the connected account; approval required.
- `create_message_recipient`: POST `/v3/collectors/{{ record.collector_id }}/messages/{{
  record.message_id }}/recipients` - kind `create`; body type `json`; path fields `collector_id`,
  `message_id`; required record fields `collector_id`, `message_id`; accepted fields `collector_id`,
  `message_id`; risk: creates SurveyMonkey message recipient data in the connected account; approval
  required.
- `create_message_recipients_bulk`: POST `/v3/collectors/{{ record.collector_id }}/messages/{{
  record.message_id }}/recipients/bulk` - kind `create`; body type `json`; path fields
  `collector_id`, `message_id`; required record fields `collector_id`, `message_id`; accepted fields
  `collector_id`, `message_id`; risk: creates SurveyMonkey bulk message recipients data in the
  connected account; approval required.
- `send_collector_message`: POST `/v3/collectors/{{ record.collector_id }}/messages/{{
  record.message_id }}/send` - kind `custom`; body type `none`; path fields `collector_id`,
  `message_id`; required record fields `collector_id`, `message_id`; accepted fields `collector_id`,
  `message_id`; confirmation `destructive`; risk: sends or shares SurveyMonkey collector message;
  may notify real recipients or change access for real users; approval required.
- `create_collector_response`: POST `/v3/collectors/{{ record.collector_id }}/responses` - kind
  `create`; body type `json`; path fields `collector_id`; required record fields `collector_id`;
  accepted fields `collector_id`; risk: creates SurveyMonkey collector response data in the
  connected account; approval required.
- `update_collector_response`: PATCH `/v3/collectors/{{ record.collector_id }}/responses/{{
  record.response_id }}` - kind `update`; body type `json`; path fields `collector_id`,
  `response_id`; required record fields `collector_id`, `response_id`; accepted fields
  `collector_id`, `response_id`; risk: updates SurveyMonkey collector response data in the connected
  account; approval required.
- `replace_collector_response`: PUT `/v3/collectors/{{ record.collector_id }}/responses/{{
  record.response_id }}` - kind `update`; body type `json`; path fields `collector_id`,
  `response_id`; required record fields `collector_id`, `response_id`; accepted fields
  `collector_id`, `response_id`; risk: replaces SurveyMonkey collector response data in the
  connected account; approval required.
- `create_contact_list`: POST `/v3/contact_lists` - kind `create`; body type `json`; risk: creates
  SurveyMonkey contact list data in the connected account; approval required.
- `update_contact_list`: PATCH `/v3/contact_lists/{{ record.contact_list_id }}` - kind `update`;
  body type `json`; path fields `contact_list_id`; required record fields `contact_list_id`;
  accepted fields `contact_list_id`; risk: updates SurveyMonkey contact list data in the connected
  account; approval required.
- `replace_contact_list`: PUT `/v3/contact_lists/{{ record.contact_list_id }}` - kind `update`; body
  type `json`; path fields `contact_list_id`; required record fields `contact_list_id`; accepted
  fields `contact_list_id`; risk: replaces SurveyMonkey contact list data in the connected account;
  approval required.
- `copy_contact_list`: POST `/v3/contact_lists/{{ record.contact_list_id }}/copy` - kind `create`;
  body type `json`; path fields `contact_list_id`; required record fields `contact_list_id`;
  accepted fields `contact_list_id`; risk: creates SurveyMonkey copied contact list data in the
  connected account; approval required.
- `merge_contact_list`: POST `/v3/contact_lists/{{ record.contact_list_id }}/merge` - kind `custom`;
  body type `json`; path fields `contact_list_id`; required record fields `contact_list_id`;
  accepted fields `contact_list_id`; risk: updates SurveyMonkey contact lists through merge data in
  the connected account; approval required.
- `add_contact_to_contact_list`: POST `/v3/contact_lists/{{ record.contact_list_id }}/contacts` -
  kind `create`; body type `json`; path fields `contact_list_id`; required record fields
  `contact_list_id`; accepted fields `contact_list_id`; risk: creates SurveyMonkey contact-list
  membership data in the connected account; approval required.
- `add_contacts_to_contact_list_bulk`: POST `/v3/contact_lists/{{ record.contact_list_id
  }}/contacts/bulk` - kind `create`; body type `json`; path fields `contact_list_id`; required
  record fields `contact_list_id`; accepted fields `contact_list_id`; risk: creates SurveyMonkey
  bulk contact-list memberships data in the connected account; approval required.
- `create_contact`: POST `/v3/contacts` - kind `create`; body type `json`; risk: creates
  SurveyMonkey contact data in the connected account; approval required.
- `create_contacts_bulk`: POST `/v3/contacts/bulk` - kind `create`; body type `json`; risk: creates
  SurveyMonkey bulk contacts data in the connected account; approval required.
- `update_contact`: PATCH `/v3/contacts/{{ record.contact_id }}` - kind `update`; body type `json`;
  path fields `contact_id`; required record fields `contact_id`; accepted fields `contact_id`; risk:
  updates SurveyMonkey contact data in the connected account; approval required.
- `replace_contact`: PUT `/v3/contacts/{{ record.contact_id }}` - kind `update`; body type `json`;
  path fields `contact_id`; required record fields `contact_id`; accepted fields `contact_id`; risk:
  replaces SurveyMonkey contact data in the connected account; approval required.
- `update_contact_field`: PATCH `/v3/contact_fields/{{ record.contact_field_id }}` - kind `update`;
  body type `json`; path fields `contact_field_id`; required record fields `contact_field_id`;
  accepted fields `contact_field_id`; risk: updates SurveyMonkey contact field data in the connected
  account; approval required.
- `create_organization`: POST `/v3/organizations` - kind `create`; body type `json`; risk: changes
  SurveyMonkey organization administration/provisioning state; approval required.
- `update_organization`: PATCH `/v3/organization/{{ record.organization_id }}` - kind `update`; body
  type `json`; path fields `organization_id`; required record fields `organization_id`; accepted
  fields `organization_id`; risk: changes SurveyMonkey organization administration/provisioning
  state; approval required.
- `create_survey_folder`: POST `/v3/survey_folders` - kind `create`; body type `json`; risk: creates
  SurveyMonkey survey folder data in the connected account; approval required.
- `create_survey`: POST `/v3/surveys` - kind `create`; body type `json`; risk: creates SurveyMonkey
  survey data in the connected account; approval required.
- `update_survey`: PATCH `/v3/surveys/{{ record.survey_id }}` - kind `update`; body type `json`;
  path fields `survey_id`; required record fields `survey_id`; accepted fields `survey_id`; risk:
  updates SurveyMonkey survey data in the connected account; approval required.
- `replace_survey`: PUT `/v3/surveys/{{ record.survey_id }}` - kind `update`; body type `json`; path
  fields `survey_id`; required record fields `survey_id`; accepted fields `survey_id`; risk:
  replaces SurveyMonkey survey data in the connected account; approval required.
- `share_survey`: POST `/v3/surveys/{{ record.survey_id }}/share` - kind `custom`; body type `json`;
  path fields `survey_id`; required record fields `survey_id`; accepted fields `survey_id`;
  confirmation `destructive`; risk: sends or shares SurveyMonkey survey access; may notify real
  recipients or change access for real users; approval required.
- `create_survey_page`: POST `/v3/surveys/{{ record.survey_id }}/pages` - kind `create`; body type
  `json`; path fields `survey_id`; required record fields `survey_id`; accepted fields `survey_id`;
  risk: creates SurveyMonkey survey page data in the connected account; approval required.
- `update_survey_page`: PATCH `/v3/surveys/{{ record.survey_id }}/pages/{{ record.page_id }}` - kind
  `update`; body type `json`; path fields `survey_id`, `page_id`; required record fields
  `survey_id`, `page_id`; accepted fields `page_id`, `survey_id`; risk: updates SurveyMonkey survey
  page data in the connected account; approval required.
- `replace_survey_page`: PUT `/v3/surveys/{{ record.survey_id }}/pages/{{ record.page_id }}` - kind
  `update`; body type `json`; path fields `survey_id`, `page_id`; required record fields
  `survey_id`, `page_id`; accepted fields `page_id`, `survey_id`; risk: replaces SurveyMonkey survey
  page data in the connected account; approval required.
- `create_survey_question`: POST `/v3/surveys/{{ record.survey_id }}/pages/{{ record.page_id
  }}/questions` - kind `create`; body type `json`; path fields `survey_id`, `page_id`; required
  record fields `survey_id`, `page_id`; accepted fields `page_id`, `survey_id`; risk: creates
  SurveyMonkey survey question data in the connected account; approval required.
- `create_survey_language`: POST `/v3/surveys/{{ record.survey_id }}/languages/{{
  record.language_code }}` - kind `create`; body type `json`; path fields `survey_id`,
  `language_code`; required record fields `survey_id`, `language_code`; accepted fields
  `language_code`, `survey_id`; risk: creates SurveyMonkey survey language/translation data in the
  connected account; approval required.
- `update_survey_language`: PATCH `/v3/surveys/{{ record.survey_id }}/languages/{{
  record.language_code }}` - kind `update`; body type `json`; path fields `survey_id`,
  `language_code`; required record fields `survey_id`, `language_code`; accepted fields
  `language_code`, `survey_id`; risk: updates SurveyMonkey survey language/translation data in the
  connected account; approval required.
- `create_survey_collector`: POST `/v3/surveys/{{ record.survey_id }}/collectors` - kind `create`;
  body type `json`; path fields `survey_id`; required record fields `survey_id`; accepted fields
  `survey_id`; risk: creates SurveyMonkey survey collector data in the connected account; approval
  required.
- `update_survey_response`: PATCH `/v3/surveys/{{ record.survey_id }}/responses/{{
  record.response_id }}` - kind `update`; body type `json`; path fields `survey_id`, `response_id`;
  required record fields `survey_id`, `response_id`; accepted fields `response_id`, `survey_id`;
  risk: updates SurveyMonkey survey response data in the connected account; approval required.
- `replace_survey_response`: PUT `/v3/surveys/{{ record.survey_id }}/responses/{{ record.response_id
  }}` - kind `update`; body type `json`; path fields `survey_id`, `response_id`; required record
  fields `survey_id`, `response_id`; accepted fields `response_id`, `survey_id`; risk: replaces
  SurveyMonkey survey response data in the connected account; approval required.
- `create_webhook`: POST `/v3/webhooks` - kind `create`; body type `json`; risk: creates
  SurveyMonkey webhook data in the connected account; approval required.
- `update_webhook`: PATCH `/v3/webhooks/{{ record.webhook_id }}` - kind `update`; body type `json`;
  path fields `webhook_id`; required record fields `webhook_id`; accepted fields `webhook_id`; risk:
  updates SurveyMonkey webhook data in the connected account; approval required.
- `replace_webhook`: PUT `/v3/webhooks/{{ record.webhook_id }}` - kind `update`; body type `json`;
  path fields `webhook_id`; required record fields `webhook_id`; accepted fields `webhook_id`; risk:
  replaces SurveyMonkey webhook data in the connected account; approval required.
- `update_group_member`: PATCH `/v3/groups/{{ record.group_id }}/members/{{ record.member_id }}` -
  kind `update`; body type `json`; path fields `group_id`, `member_id`; required record fields
  `group_id`, `member_id`; accepted fields `group_id`, `member_id`; risk: changes SurveyMonkey group
  membership administration/provisioning state; approval required.
- `create_workgroup`: POST `/v3/workgroups` - kind `create`; body type `json`; risk: changes
  SurveyMonkey workgroup administration/provisioning state; approval required.
- `update_workgroup`: PATCH `/v3/workgroups/{{ record.workgroup_id }}` - kind `update`; body type
  `json`; path fields `workgroup_id`; required record fields `workgroup_id`; accepted fields
  `workgroup_id`; risk: changes SurveyMonkey workgroup administration/provisioning state; approval
  required.
- `create_workgroup_share`: POST `/v3/workgroups/{{ record.workgroup_id }}/shares` - kind `create`;
  body type `json`; path fields `workgroup_id`; required record fields `workgroup_id`; accepted
  fields `workgroup_id`; risk: sends or shares SurveyMonkey workgroup share; may notify real
  recipients or change access for real users; approval required.
- `create_workgroup_shares_bulk`: POST `/v3/workgroups/{{ record.workgroup_id }}/shares/bulk` - kind
  `create`; body type `json`; path fields `workgroup_id`; required record fields `workgroup_id`;
  accepted fields `workgroup_id`; risk: sends or shares SurveyMonkey bulk workgroup shares; may
  notify real recipients or change access for real users; approval required.
- `create_workgroup_member`: POST `/v3/workgroups/{{ record.workgroup_id }}/members` - kind
  `create`; body type `json`; path fields `workgroup_id`; required record fields `workgroup_id`;
  accepted fields `workgroup_id`; risk: changes SurveyMonkey workgroup membership
  administration/provisioning state; approval required.
- `create_workgroup_members_bulk`: POST `/v3/workgroups/{{ record.workgroup_id }}/members/bulk` -
  kind `create`; body type `json`; path fields `workgroup_id`; required record fields
  `workgroup_id`; accepted fields `workgroup_id`; risk: changes SurveyMonkey bulk workgroup
  membership administration/provisioning state; approval required.
- `update_workgroup_member`: PATCH `/v3/workgroups/{{ record.workgroup_id }}/members/{{
  record.member_id }}` - kind `update`; body type `json`; path fields `workgroup_id`, `member_id`;
  required record fields `workgroup_id`, `member_id`; accepted fields `member_id`, `workgroup_id`;
  risk: changes SurveyMonkey workgroup membership administration/provisioning state; approval
  required.
- `create_scim_user`: POST `/scim/v2/Users` - kind `create`; body type `json`; risk: changes
  SurveyMonkey SCIM user provisioning administration/provisioning state; approval required.
- `replace_scim_user`: PUT `/scim/v2/Users/{{ record.scim_user_id }}` - kind `update`; body type
  `json`; path fields `scim_user_id`; required record fields `scim_user_id`; accepted fields
  `scim_user_id`; risk: changes SurveyMonkey SCIM user provisioning administration/provisioning
  state; approval required.
- `update_scim_user`: PATCH `/scim/v2/Users/{{ record.scim_user_id }}` - kind `update`; body type
  `json`; path fields `scim_user_id`; required record fields `scim_user_id`; accepted fields
  `scim_user_id`; risk: changes SurveyMonkey SCIM user provisioning administration/provisioning
  state; approval required.

## Known limits

- Batch defaults: read_page_size=100, write_batch_size=1.
- API coverage includes 82 stream-backed endpoint group(s), 53 write-backed endpoint group(s).
