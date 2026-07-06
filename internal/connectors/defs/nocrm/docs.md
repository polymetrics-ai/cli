# Overview

Reads noCRM.io CRM objects and exposes write actions for supported noCRM API v2
mutations.

Readable streams: `leads`, `pipelines`, `users`, `teams`, `prospecting_lists`, `steps`, `step`,
`client_folders`, `client_folder`, `categories`, `predefined_tags`, `fields`, `activities`, `lead`,
`unassigned_leads`, `lead_comments`, `lead_duplicates`, `lead_attachments`, `lead_attachment`,
`lead_action_histories`, `post_sales_tasks`, `spreadsheets`, `spreadsheet`, `prospects`,
`prospects_called`, `user`, `team`, `webhooks`, `webhook_events`, `webhook_event`.

Write actions: `create_client_folder`, `update_client_folder`, `delete_client_folder`,
`create_category`, `create_predefined_tag`, `create_field`, `create_lead`, `duplicate_lead`,
`update_lead`, `assign_lead`, `add_lead_to_client_folder`, `delete_lead`, `delete_multiple_leads`,
`create_lead_comment`, `update_lead_comment`, `delete_lead_comment`, `delete_lead_attachment`,
`send_lead_email_from_template`, `create_lead_follow_up_from_template`, `create_prospecting_list`,
`assign_prospecting_list`, `create_prospecting_list_comment`, `create_prospect_comment`,
`update_prospect_comment`, `create_prospects`, `update_prospect_fields`,
`create_lead_from_prospect`, `delete_prospect`, `create_user`, `disable_user`, `create_team`,
`update_team`, `delete_team`, `add_team_member`, `remove_team_member`, `create_webhook`,
`activate_webhook`, `delete_webhook`.

Service API documentation: https://www.nocrm.io/api.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); noCRM API key, sent as the X-API-KEY header. Never logged.
- `attachment_id` (optional, string); Lead attachment id for the lead_attachment detail stream.
- `base_url` (optional, string); default `https://api.nocrm.io/api/v2`; format `uri`; noCRM API base
  URL. Override with your account's subdomain base URL, e.g. https://yourcompany.nocrm.io/api/v2.
- `client_id` (optional, string); Client folder id for the client_folder detail stream.
- `lead_id` (optional, string); Lead id for lead detail and lead subresource streams.
- `prospecting_list_id` (optional, string); Prospecting list/spreadsheet id for current
  prospecting-list detail and prospect streams.
- `step_id` (optional, string); Step id for the step detail stream.
- `team_id` (optional, string); Team id for the team detail stream.
- `user_id` (optional, string); User id for the user detail stream.
- `webhook_event_id` (optional, string); Webhook event id for the webhook_event detail stream.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.nocrm.io/api/v2`.

Authentication behavior:

- API key authentication in `X-API-KEY` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/pipelines`.

## Streams notes

Default pagination: offset/limit pagination; offset parameter `offset`; limit parameter `limit`;
page size 100.

Pagination by stream: none: `step`, `client_folder`, `lead`, `lead_attachment`, `spreadsheet`,
`user`, `team`, `webhook_event`; offset_limit: `leads`, `pipelines`, `users`, `teams`,
`prospecting_lists`, `steps`, `client_folders`, `categories`, `predefined_tags`, `fields`,
`activities`, `unassigned_leads`, `lead_comments`, `lead_duplicates`, `lead_attachments`,
`lead_action_histories`, `post_sales_tasks`, `spreadsheets`, `prospects`, `prospects_called`,
`webhooks`, `webhook_events`.

- `leads`: GET `/leads` - records at response root; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 100.
- `pipelines`: GET `/pipelines` - records at response root; offset/limit pagination; offset
  parameter `offset`; limit parameter `limit`; page size 100.
- `users`: GET `/users` - records at response root; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 100.
- `teams`: GET `/teams` - records at response root; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 100.
- `prospecting_lists`: GET `/prospecting_lists` - records at response root; offset/limit pagination;
  offset parameter `offset`; limit parameter `limit`; page size 100.
- `steps`: GET `/steps` - records at response root; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 100.
- `step`: GET `/steps/{{ config.step_id }}` - single-object response; records at response root.
- `client_folders`: GET `/clients` - records at response root; offset/limit pagination; offset
  parameter `offset`; limit parameter `limit`; page size 100.
- `client_folder`: GET `/clients/{{ config.client_id }}` - single-object response; records at
  response root.
- `categories`: GET `/categories` - records at response root; offset/limit pagination; offset
  parameter `offset`; limit parameter `limit`; page size 100.
- `predefined_tags`: GET `/predefined_tags` - records at response root; offset/limit pagination;
  offset parameter `offset`; limit parameter `limit`; page size 100.
- `fields`: GET `/fields` - records at response root; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 100.
- `activities`: GET `/activities` - records at response root; offset/limit pagination; offset
  parameter `offset`; limit parameter `limit`; page size 100.
- `lead`: GET `/leads/{{ config.lead_id }}` - single-object response; records at response root.
- `unassigned_leads`: GET `/leads/unassigned` - records at response root; offset/limit pagination;
  offset parameter `offset`; limit parameter `limit`; page size 100.
- `lead_comments`: GET `/leads/{{ config.lead_id }}/comments` - records at response root;
  offset/limit pagination; offset parameter `offset`; limit parameter `limit`; page size 100.
- `lead_duplicates`: GET `/leads/{{ config.lead_id }}/duplicates` - records at response root;
  offset/limit pagination; offset parameter `offset`; limit parameter `limit`; page size 100.
- `lead_attachments`: GET `/leads/{{ config.lead_id }}/attachments` - records at response root;
  offset/limit pagination; offset parameter `offset`; limit parameter `limit`; page size 100.
- `lead_attachment`: GET `/leads/{{ config.lead_id }}/attachments/{{ config.attachment_id }}` -
  single-object response; records at response root.
- `lead_action_histories`: GET `/leads/{{ config.lead_id }}/action_histories` - records at response
  root; offset/limit pagination; offset parameter `offset`; limit parameter `limit`; page size 100.
- `post_sales_tasks`: GET `/follow_ups` - records at response root; offset/limit pagination; offset
  parameter `offset`; limit parameter `limit`; page size 100.
- `spreadsheets`: GET `/spreadsheets` - records at response root; offset/limit pagination; offset
  parameter `offset`; limit parameter `limit`; page size 100.
- `spreadsheet`: GET `/spreadsheets/{{ config.prospecting_list_id }}` - single-object response;
  records at response root.
- `prospects`: GET `/rows` - records at response root; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 100.
- `prospects_called`: GET `/rows/called_from` - records at response root; offset/limit pagination;
  offset parameter `offset`; limit parameter `limit`; page size 100.
- `user`: GET `/users/{{ config.user_id }}` - single-object response; records at response root.
- `team`: GET `/teams/{{ config.team_id }}` - single-object response; records at response root.
- `webhooks`: GET `/webhooks` - records at response root; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 100.
- `webhook_events`: GET `/webhook_events` - records at response root; offset/limit pagination;
  offset parameter `offset`; limit parameter `limit`; page size 100.
- `webhook_event`: GET `/webhook_events/{{ config.webhook_event_id }}` - single-object response;
  records at response root.

## Write actions & risks

Overall write risk: external noCRM API mutations can create, update, assign, email, disable, or
delete live CRM/admin records.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_client_folder`: POST `/clients` - kind `create`; body type `json`; required record fields
  `name`; accepted fields `description`, `name`; risk: creates a noCRM client folder visible to
  account users; external mutation, approval required.
- `update_client_folder`: PUT `/clients/{{ record.id }}` - kind `update`; body type `json`; path
  fields `id`; required record fields `id`; accepted fields `description`, `id`, `name`; risk:
  updates a noCRM client folder; external mutation, approval required.
- `delete_client_folder`: DELETE `/clients/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `404`; confirmation `destructive`; risk: deletes a noCRM client folder; destructive
  external mutation, approval required.
- `create_category`: POST `/category` - kind `create`; body type `json`; required record fields
  `name`; accepted fields `name`; risk: creates a noCRM category/tag grouping; account configuration
  mutation, approval required.
- `create_predefined_tag`: POST `/predefined_tags` - kind `create`; body type `json`; required
  record fields `name`; accepted fields `name`; risk: creates a predefined tag available in noCRM;
  account taxonomy mutation, approval required.
- `create_field`: POST `/fields` - kind `create`; body type `json`; required record fields `name`,
  `type`; accepted fields `is_duplicate`, `name`, `type`; risk: creates a noCRM custom field;
  account schema mutation, approval required.
- `create_lead`: POST `/leads` - kind `create`; body type `json`; required record fields `title`;
  accepted fields `amount`, `description`, `step`, `tags`, `title`, `user_id`; risk: creates a live
  noCRM lead and may assign ownership/tags/step; external CRM mutation, approval required.
- `duplicate_lead`: POST `/leads/{{ record.id }}/duplicate_lead` - kind `create`; body type `none`;
  path fields `id`; required record fields `id`; accepted fields `id`; risk: duplicates a live noCRM
  lead; external CRM mutation, approval required.
- `update_lead`: PUT `/leads/{{ record.id }}` - kind `update`; body type `json`; path fields `id`;
  required record fields `id`; accepted fields `amount`, `description`, `id`, `step`, `tags`,
  `title`; risk: updates a live noCRM lead; external CRM mutation, approval required.
- `assign_lead`: POST `/leads/{{ record.id }}/assign` - kind `update`; body type `json`; path fields
  `id`; required record fields `id`, `user_id`; accepted fields `id`, `user_id`; risk: assigns a
  live noCRM lead to a user and can trigger notifications; external CRM mutation, approval required.
- `add_lead_to_client_folder`: POST `/leads/{{ record.id }}/add_to_client` - kind `update`; body
  type `json`; path fields `id`; required record fields `id`, `client_id`; accepted fields
  `client_id`, `id`; risk: links a live lead to a client folder; external CRM mutation, approval
  required.
- `delete_lead`: DELETE `/leads/{{ record.id }}` - kind `delete`; body type `none`; path fields
  `id`; required record fields `id`; accepted fields `id`; missing records treated as success for
  status `404`; confirmation `destructive`; risk: deletes a live noCRM lead; destructive external
  mutation, approval required.
- `delete_multiple_leads`: DELETE `/leads/delete_multiple` - kind `delete`; body type `json`; body
  fields `ids`; required record fields `ids`; accepted fields `ids`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: bulk-deletes live noCRM leads;
  destructive external mutation, approval required.
- `create_lead_comment`: POST `/leads/{{ record.lead_id }}/comments` - kind `create`; body type
  `json`; path fields `lead_id`; required record fields `lead_id`, `content`; accepted fields
  `activity_id`, `content`, `is_pinned`, `lead_id`; risk: adds a comment to a live noCRM lead;
  external CRM mutation, approval required.
- `update_lead_comment`: PUT `/leads/{{ record.lead_id }}/comments/{{ record.id }}` - kind `update`;
  body type `json`; path fields `lead_id`, `id`; required record fields `lead_id`, `id`; accepted
  fields `activity_id`, `content`, `id`, `is_pinned`, `lead_id`; risk: updates a lead comment in
  noCRM; external CRM mutation, approval required.
- `delete_lead_comment`: DELETE `/leads/{{ record.lead_id }}/comments/{{ record.id }}` - kind
  `delete`; body type `none`; path fields `lead_id`, `id`; required record fields `lead_id`, `id`;
  accepted fields `id`, `lead_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: deletes a noCRM lead comment; destructive external mutation, approval
  required.
- `delete_lead_attachment`: DELETE `/leads/{{ record.lead_id }}/attachments/{{ record.id }}` - kind
  `delete`; body type `none`; path fields `lead_id`, `id`; required record fields `lead_id`, `id`;
  accepted fields `id`, `lead_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: deletes an attachment from a live noCRM lead; destructive external mutation,
  approval required.
- `send_lead_email_from_template`: POST `/leads/{{ record.lead_id
  }}/emails/send_email_from_template` - kind `create`; body type `json`; path fields `lead_id`;
  required record fields `lead_id`, `email_template_id`, `from_user_id`; accepted fields
  `email_template_id`, `from_user_id`, `lead_id`; risk: sends an email from a noCRM template to a
  lead; external communication mutation, approval required.
- `create_lead_follow_up_from_template`: POST `/leads/{{ record.lead_id
  }}/follow_ups/create_from_template` - kind `create`; body type `json`; path fields `lead_id`;
  required record fields `lead_id`, `post_sales_template_id`; accepted fields `lead_id`,
  `post_sales_template_id`; risk: creates post-sales tasks for a live noCRM lead; workflow mutation,
  approval required.
- `create_prospecting_list`: POST `/spreadsheets` - kind `create`; body type `json`; required record
  fields `title`, `content`; accepted fields `content`, `description`, `privacy`, `tags`, `title`,
  `user_id`; risk: creates a noCRM prospecting list and optional rows/tags/owner; external CRM
  mutation, approval required.
- `assign_prospecting_list`: POST `/spreadsheets/{{ record.id }}/assign` - kind `update`; body type
  `json`; path fields `id`; required record fields `id`, `user_id`; accepted fields `id`, `user_id`;
  risk: assigns a prospecting list to a user; external CRM mutation, approval required.
- `create_prospecting_list_comment`: POST `/spreadsheets/{{ record.spreadsheet_id }}/comments` -
  kind `create`; body type `json`; path fields `spreadsheet_id`; required record fields
  `spreadsheet_id`, `content`; accepted fields `activity_id`, `content`, `is_pinned`,
  `spreadsheet_id`; risk: adds a comment to a noCRM prospecting list; external CRM mutation,
  approval required.
- `create_prospect_comment`: POST `/spreadsheets/{{ record.spreadsheet_id }}/rows/{{
  record.prospect_id }}/comments` - kind `create`; body type `json`; path fields `spreadsheet_id`,
  `prospect_id`; required record fields `spreadsheet_id`, `prospect_id`, `content`; accepted fields
  `activity_id`, `content`, `is_pinned`, `prospect_id`, `spreadsheet_id`; risk: adds a comment to a
  prospect row; external CRM mutation, approval required.
- `update_prospect_comment`: PUT `/spreadsheets/{{ record.spreadsheet_id }}/rows/{{
  record.prospect_id }}/comments/{{ record.id }}` - kind `update`; body type `json`; path fields
  `spreadsheet_id`, `prospect_id`, `id`; required record fields `spreadsheet_id`, `prospect_id`,
  `id`; accepted fields `activity_id`, `content`, `id`, `is_pinned`, `prospect_id`,
  `spreadsheet_id`; risk: updates a prospect-row comment; external CRM mutation, approval required.
- `create_prospects`: POST `/spreadsheets/{{ record.spreadsheet_id }}/rows` - kind `create`; body
  type `json`; path fields `spreadsheet_id`; required record fields `spreadsheet_id`, `content`;
  accepted fields `content`, `spreadsheet_id`; risk: adds prospect rows to a noCRM prospecting list;
  external CRM mutation, approval required.
- `update_prospect_fields`: PUT `/spreadsheets/{{ record.spreadsheet_id }}/rows/{{ record.id
  }}/update_fields` - kind `update`; body type `json`; path fields `spreadsheet_id`, `id`; required
  record fields `spreadsheet_id`, `id`, `fields`; accepted fields `fields`, `id`, `spreadsheet_id`;
  risk: updates named fields on a prospect row; external CRM mutation, approval required.
- `create_lead_from_prospect`: POST `/spreadsheets/{{ record.spreadsheet_id }}/rows/{{ record.id
  }}/create_lead` - kind `create`; body type `none`; path fields `spreadsheet_id`, `id`; required
  record fields `spreadsheet_id`, `id`; accepted fields `id`, `spreadsheet_id`; risk: converts a
  prospect row into a live noCRM lead; external CRM mutation, approval required.
- `delete_prospect`: DELETE `/spreadsheets/{{ record.spreadsheet_id }}/rows/{{ record.id }}` - kind
  `delete`; body type `none`; path fields `spreadsheet_id`, `id`; required record fields
  `spreadsheet_id`, `id`; accepted fields `id`, `spreadsheet_id`; missing records treated as success
  for status `404`; confirmation `destructive`; risk: deletes a prospect row from a prospecting
  list; destructive external mutation, approval required.
- `create_user`: POST `/users` - kind `create`; body type `json`; required record fields `lastname`,
  `firstname`, `email`; accepted fields `dont_send_email`, `email`, `firstname`, `is_admin`,
  `lastname`; risk: creates a noCRM user account and can send activation email depending on payload;
  administrative mutation, approval required.
- `disable_user`: PUT `/users/{{ record.id }}/disable` - kind `update`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; confirmation `destructive`; risk:
  disables a noCRM user account; administrative mutation, approval required.
- `create_team`: POST `/teams` - kind `create`; body type `json`; required record fields `name`;
  accepted fields `name`; risk: creates a noCRM team; administrative mutation, approval required.
- `update_team`: PUT `/teams/{{ record.id }}` - kind `update`; body type `json`; path fields `id`;
  required record fields `id`; accepted fields `id`, `name`; risk: updates a noCRM team;
  administrative mutation, approval required.
- `delete_team`: DELETE `/teams/{{ record.id }}` - kind `delete`; body type `none`; path fields
  `id`; required record fields `id`; accepted fields `id`; missing records treated as success for
  status `404`; confirmation `destructive`; risk: deletes a noCRM team; destructive administrative
  mutation, approval required.
- `add_team_member`: POST `/teams/{{ record.id }}/add_member` - kind `update`; body type `json`;
  path fields `id`; required record fields `id`, `user_id`; accepted fields `id`, `is_manager`,
  `user_id`; risk: adds a user to a noCRM team and can change manager status; administrative
  mutation, approval required.
- `remove_team_member`: DELETE `/teams/{{ record.id }}/remove_member` - kind `delete`; body type
  `json`; path fields `id`; body fields `user_id`; required record fields `id`, `user_id`; accepted
  fields `id`, `user_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: removes a user from a noCRM team; administrative mutation, approval required.
- `create_webhook`: POST `/webhooks` - kind `create`; body type `json`; required record fields
  `event`, `target_type`, `target`; accepted fields `event`, `target`, `target_type`; risk: creates
  a noCRM webhook/notification destination; outbound data delivery mutation, approval required.
- `activate_webhook`: PUT `/webhooks/{{ record.id }}/activate` - kind `update`; body type `none`;
  path fields `id`; required record fields `id`; accepted fields `id`; risk: activates a noCRM
  webhook/notification destination; outbound data delivery mutation, approval required.
- `delete_webhook`: DELETE `/webhooks/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `404`; confirmation `destructive`; risk: disables or removes a noCRM
  webhook/notification destination; external mutation, approval required.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 30 stream-backed endpoint group(s), 38 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=2, destructive_admin=1, duplicate_of=1, non_data_endpoint=1,
  requires_elevated_scope=7.
